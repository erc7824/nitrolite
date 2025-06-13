package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

type RPCNode struct {
	upgrader websocket.Upgrader

	groupId      string
	handlerChain map[string][]RPCHandler
	routes       map[string][]string

	signer  *Signer
	connHub *rpcConnectionHub
	logger  Logger
}

func NewRPCNode(signer *Signer, logger Logger) *RPCNode {
	return &RPCNode{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for simplicity
			},
		},
		groupId:      "root",
		handlerChain: make(map[string][]RPCHandler),
		routes:       make(map[string][]string),
		signer:       signer,
		connHub:      newRPCConnectionHub(),
		logger:       logger,
	}
}

// HandleConnection handles the WebSocket connection lifecycle.
func (n *RPCNode) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := n.upgrader.Upgrade(w, r, nil)
	if err != nil {
		n.logger.Error("failed to upgrade connection to WebSocket", "error", err)
		return
	}
	defer conn.Close()

	errorGroup, parentCtx := errgroup.WithContext(r.Context())
	connectionID := uuid.NewString()
	messageSink := make(chan []byte, 10)
	n.connHub.Set(&RPCConnection{
		ConnectionID: connectionID,
		WriteSink:    messageSink,
		Storage:      make(map[string]any),
	})

	defer func() {
		n.connHub.Remove(connectionID)
		n.logger.Info("connection closed", "connectionID", connectionID)
	}()

	readMesages := func() error {
		for {
			rpcConn := n.connHub.Get(connectionID)
			if conn == nil {
				n.logger.Error("connection not found in hub", "connectionID", connectionID)
				return fmt.Errorf("connection not found in hub for ID %s", connectionID)
			}

			select {
			case <-parentCtx.Done():
				n.logger.Info("context done, stopping message processing")
				close(rpcConn.WriteSink)
				return nil
			default:
			}

			_, messageBytes, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					n.logger.Error("WebSocket connection closed with unexpected reason", "error", err)
				} else {
					n.logger.Error("failed to read message", "error", err)
				}
				return err
			}

			var msg RPCMessage
			if err := json.Unmarshal(messageBytes, &msg); err != nil {
				n.logger.Debug("invalid message format", "error", err, "message", string(messageBytes))
				n.sendErrorResponse(rpcConn, 0, "invalid message format")
				continue
			}

			if err := validate.Struct(&msg); err != nil {
				n.logger.Debug("message validation failed", "error", err, "message", string(messageBytes))
				n.sendErrorResponse(rpcConn, msg.Req.RequestID, "message validation failed")
				continue
			}

			methodRoute, ok := n.routes[msg.Req.Method]
			if !ok || len(methodRoute) == 0 {
				n.logger.Debug("no handler found for method", "method", msg.Req.Method)
				n.sendErrorResponse(rpcConn, msg.Req.RequestID, fmt.Sprintf("unknown method: %s", msg.Req.Method))
				continue
			}

			var routeHandlers []RPCHandler
			for _, handlersId := range methodRoute {
				handlers, exists := n.handlerChain[handlersId]
				if !exists || len(handlers) == 0 {
					n.logger.Fatal("no handlers found for id", "id", handlersId)
					return fmt.Errorf("no handlers found for id %s", handlersId)
				}

				routeHandlers = append(routeHandlers, handlers...)
			}
			n.logger.Info("processing message",
				"requestID", msg.Req.RequestID,
				"method", msg.Req.Method,
				"route", methodRoute)

			ctx := &RPCContext{
				Context:  context.Background(),
				UserID:   rpcConn.UserID,
				Signer:   n.signer,
				Message:  msg,
				handlers: routeHandlers,
				Storage:  rpcConn.Storage,
			}
			ctx.Next() // Start processing the handlers

			responseBytes, err := ctx.GetRawResponse()
			if err != nil {
				n.logger.Error("failed to prepare response", "error", err, "method", msg.Req.Method)
				continue
			}
			rpcConn.WriteSink <- responseBytes

			n.connHub.Set(&RPCConnection{
				ConnectionID: rpcConn.ConnectionID,
				UserID:       ctx.UserID,
				WriteSink:    rpcConn.WriteSink,
				Storage:      ctx.Storage,
			})
		}
	}

	writeMessages := func() error {
		for messageBytes := range messageSink {
			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				n.logger.Error("error getting writer for response", "error", err)
				continue
			}

			if _, err := w.Write(messageBytes); err != nil {
				n.logger.Error("error writing response", "error", err)
				w.Close()
				continue
			}

			if err := w.Close(); err != nil {
				n.logger.Error("error closing writer for response", "error", err)
				continue
			}
		}

		return nil
	}

	errorGroup.Go(readMesages)
	errorGroup.Go(writeMessages)
	if err := errorGroup.Wait(); err != nil && parentCtx.Err() == nil {
		n.logger.Error("error in WebSocket message handling", "error", err)
	}
}

type RPCHandler func(c *RPCContext)

type RPCContext struct {
	Context context.Context
	UserID  string
	Signer  *Signer
	Message RPCMessage
	Storage map[string]any

	handlers []RPCHandler
}

func (c *RPCContext) Next() {
	if len(c.handlers) == 0 {
		return
	}

	handler := c.handlers[0]
	c.handlers = c.handlers[1:]
	handler(c)
}

func (c *RPCContext) Succeed(params ...any) {
	c.Message.Res = &RPCData{
		RequestID: c.Message.Req.RequestID,
		Method:    c.Message.Req.Method,
		Params:    params,
		Timestamp: uint64(time.Now().UnixMilli()),
	}
}

func (c *RPCContext) Fail(message string) {
	c.Message.Res = &RPCData{
		RequestID: c.Message.Req.RequestID,
		Method:    "error",
		Params:    []any{message},
		Timestamp: uint64(time.Now().UnixMilli()),
	}
}

func (c *RPCContext) GetRawResponse() ([]byte, error) {
	return prepareRawRPCResponse(c.Signer, c.Message.Res)
}

func prepareRawRPCResponse(signer *Signer, data *RPCData) ([]byte, error) {
	if data == nil {
		return nil, fmt.Errorf("response data is nil")
	}

	resDataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response data: %w", err)
	}

	signature, err := signer.Sign(resDataBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to sign response data: %w", err)
	}

	responseMessage := &RPCMessage{
		Res: data,
		Sig: []string{hexutil.Encode(signature)},
	}
	resMessageBytes, err := json.Marshal(responseMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response message: %w", err)
	}

	return resMessageBytes, nil
}

func (wn *RPCNode) NewGroup(name string) *RPCHandlerGroup {
	return &RPCHandlerGroup{
		groupId:     name,
		routePrefix: []string{wn.groupId},
		root:        wn,
	}
}

func (wn *RPCNode) Handle(method string, handler RPCHandler) {
	wn.handle(method, handler)
	wn.routes[method] = []string{wn.groupId, method}
}

func (wn *RPCNode) handle(method string, handler RPCHandler) {
	if method == "" {
		wn.logger.Fatal("Websocket method cannot be empty")
	}
	if handler == nil {
		wn.logger.Fatal("Websocket handler cannot be nil", "method", method)
	}

	wn.handlerChain[method] = []RPCHandler{handler}
}

func (wn *RPCNode) Use(middleware RPCHandler) {
	wn.use(wn.groupId, middleware)
}

func (wn *RPCNode) use(groupId string, middleware RPCHandler) {
	if middleware == nil {
		wn.logger.Fatal("Websocket middleware handler cannot be nil for group")
	}

	if _, exists := wn.handlerChain[groupId]; !exists {
		wn.handlerChain[groupId] = []RPCHandler{}
	}

	wn.handlerChain[groupId] = append(wn.handlerChain[groupId], middleware)
}

func (wn *RPCNode) sendErrorResponse(conn *RPCConnection, requestID uint64, message string) {
	if requestID == 0 {
		requestID = uint64(time.Now().UnixMilli())
	}
	if conn == nil {
		wn.logger.Error("connection is nil, cannot send error response", "requestID", requestID)
		return
	}

	data := &RPCData{
		RequestID: requestID,
		Method:    "error",
		Params:    []any{message},
		Timestamp: uint64(time.Now().UnixMilli()),
	}

	responseBytes, err := prepareRawRPCResponse(wn.signer, data)
	if err != nil {
		wn.logger.Error("failed to prepare error response", "error", err)
		return
	}

	conn.WriteSink <- responseBytes
}

type RPCHandlerGroup struct {
	groupId     string
	routePrefix []string
	root        *RPCNode
}

func (hg *RPCHandlerGroup) NewGroup(name string) *RPCHandlerGroup {
	return &RPCHandlerGroup{
		groupId:     name,
		routePrefix: append(hg.routePrefix, hg.groupId),
		root:        hg.root,
	}
}

func (hg *RPCHandlerGroup) Handle(method string, handler RPCHandler) {
	hg.root.routes[method] = append(hg.routePrefix, hg.groupId, method)
	hg.root.handle(method, handler)
}

func (hg *RPCHandlerGroup) Use(middleware RPCHandler) {
	hg.root.use(hg.groupId, middleware)
}

type RPCConnection struct {
	ConnectionID string
	UserID       string
	WriteSink    chan<- []byte
	Storage      map[string]any
}

type rpcConnectionHub struct {
	connections map[string]*RPCConnection
	authMapping map[string]string // Maps user IDs to connection IDs
	mu          sync.RWMutex
}

// NewRPCConnectionHub creates a new instance of rpcConnectionHub
func newRPCConnectionHub() *rpcConnectionHub {
	return &rpcConnectionHub{
		connections: make(map[string]*RPCConnection),
		authMapping: make(map[string]string),
	}
}

func (hub *rpcConnectionHub) Set(conn *RPCConnection) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	hub.connections[conn.ConnectionID] = conn

	if conn.UserID != "" {
		hub.authMapping[conn.UserID] = conn.ConnectionID
	}
}

func (hub *rpcConnectionHub) Get(connID string) *RPCConnection {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	conn, ok := hub.connections[connID]
	if !ok {
		return nil
	}

	return conn
}

func (hub *rpcConnectionHub) Remove(connID string) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	conn, ok := hub.connections[connID]
	if !ok {
		return
	}

	delete(hub.connections, connID)
	delete(hub.authMapping, conn.UserID)
}

func (hub *rpcConnectionHub) Publish(userID string, message []byte) {
	hub.mu.RLock()
	defer hub.mu.RUnlock()
	connID, ok := hub.authMapping[userID]
	if !ok {
		return
	}

	conn, ok := hub.connections[connID]
	if !ok {
		return
	}

	conn.WriteSink <- message
}
