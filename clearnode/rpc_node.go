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

const (
	rpcNodeGroupHandlerPrefix = "group."
	rpcNodeGroupRoot          = "root"
)

type RPCNode struct {
	upgrader websocket.Upgrader

	groupId      string
	handlerChain map[string][]RPCHandler
	routes       map[string][]string

	signer  *Signer
	connHub *rpcConnectionHub
	logger  Logger

	onConnectHandlers       []func(send SendRPCMessageFunc)
	onDisconnectHandlers    []func(userID string)
	onMessageSentHandlers   []func()
	onAuthenticatedHandlers []func(userID string, send SendRPCMessageFunc)
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

		groupId:      rpcNodeGroupHandlerPrefix + rpcNodeGroupRoot,
		handlerChain: make(map[string][]RPCHandler),
		routes:       make(map[string][]string),

		signer:  signer,
		connHub: newRPCConnectionHub(),
		logger:  logger.NewSystem("rpc-node"),

		onConnectHandlers:       []func(send SendRPCMessageFunc){},
		onDisconnectHandlers:    []func(userID string){},
		onMessageSentHandlers:   []func(){},
		onAuthenticatedHandlers: []func(userID string, send SendRPCMessageFunc){},
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
		Storage:      NewSafeStorage(),
	})

	for _, handler := range n.onConnectHandlers {
		handler(n.getSendMessageFunc(messageSink))
	}

	defer func() {
		n.connHub.Remove(connectionID)

		userID := ""
		if rpcConn := n.connHub.Get(connectionID); rpcConn != nil {
			userID = rpcConn.UserID
		}

		for _, handler := range n.onDisconnectHandlers {
			handler(userID)
		}

		n.logger.Info("connection closed", "connectionID", connectionID, "userID", userID)
	}()

	readMesages := func() error {
		authHandlerExecuted := false
		handleAuthenticated := func(userID string) {
			if authHandlerExecuted {
				return
			}
			authHandlerExecuted = true

			for _, handler := range n.onAuthenticatedHandlers {
				handler(userID, n.getSendMessageFunc(messageSink))
			}
		}
		processContext := func(c *RPCContext) {
			if c.UserID != "" {
				handleAuthenticated(c.UserID)
			}
		}

	read_loop:
		for {
			rpcConn := n.connHub.Get(connectionID)
			if rpcConn == nil {
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
					err = nil
				}

				close(rpcConn.WriteSink)
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
					n.logger.Error("no handlers found for id", "id", handlersId)
					n.sendErrorResponse(rpcConn, msg.Req.RequestID, fmt.Sprintf("unknown method: %s", msg.Req.Method))
					continue read_loop
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
			processContext(ctx)

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

			for _, handler := range n.onMessageSentHandlers {
				handler()
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
type SendRPCMessageFunc func(method string, params ...any)

type RPCContext struct {
	Context context.Context
	UserID  string
	Signer  *Signer
	Message RPCMessage
	Storage *SafeStorage

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

func (c *RPCContext) Succeed(method string, params ...any) {
	c.Message.Res = &RPCData{
		RequestID: c.Message.Req.RequestID,
		Method:    method,
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

func prepareRawNotification(signer *Signer, method string, params ...any) ([]byte, error) {
	if params == nil {
		params = []any{}
	}

	data := &RPCData{
		RequestID: uint64(time.Now().UnixMilli()),
		Method:    method,
		Params:    params,
		Timestamp: uint64(time.Now().UnixMilli()),
	}

	responseBytes, err := prepareRawRPCResponse(signer, data)
	if err != nil {
		return nil, err
	}

	return responseBytes, nil
}

func (wn *RPCNode) NewGroup(name string) *RPCHandlerGroup {
	return &RPCHandlerGroup{
		groupId:     rpcNodeGroupHandlerPrefix + name,
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
		panic("Websocket method cannot be empty")
	}
	if handler == nil {
		panic(fmt.Sprintf("Websocket handler cannot be nil for method %s", method))
	}

	wn.handlerChain[method] = []RPCHandler{handler}
}

func (wn *RPCNode) Use(middleware RPCHandler) {
	wn.use(wn.groupId, middleware)
}

func (wn *RPCNode) use(groupId string, middleware RPCHandler) {
	if middleware == nil {
		panic("Websocket middleware handler cannot be nil for group")
	}

	if _, exists := wn.handlerChain[groupId]; !exists {
		wn.handlerChain[groupId] = []RPCHandler{}
	}

	wn.handlerChain[groupId] = append(wn.handlerChain[groupId], middleware)
}

func (wn *RPCNode) OnConnect(handler func(send SendRPCMessageFunc)) {
	wn.onConnectHandlers = append(wn.onConnectHandlers, handler)
}

func (wn *RPCNode) OnDisconnect(handler func(userID string)) {
	wn.onDisconnectHandlers = append(wn.onDisconnectHandlers, handler)
}

func (wn *RPCNode) OnMessageSent(handler func()) {
	wn.onMessageSentHandlers = append(wn.onMessageSentHandlers, handler)
}

func (wn *RPCNode) OnAuthenticated(handler func(userID string, send SendRPCMessageFunc)) {
	wn.onAuthenticatedHandlers = append(wn.onAuthenticatedHandlers, handler)
}

func (wn *RPCNode) Notify(userID, method string, params ...any) {
	message, err := prepareRawNotification(wn.signer, method, params...)
	if err != nil {
		wn.logger.Error("failed to prepare notification message", "error", err, "userID", userID, "method", method)
		return
	}

	wn.connHub.Publish(userID, message)
}

func (wn *RPCNode) getSendMessageFunc(writeSink chan<- []byte) SendRPCMessageFunc {
	return func(method string, params ...any) {
		message, err := prepareRawNotification(wn.signer, method, params...)
		if err != nil {
			wn.logger.Error("failed to prepare notification message", "error", err, "method", method)
			return
		}

		if writeSink == nil {
			wn.logger.Error("write sink is nil, cannot send message", "method", method)
			return
		}

		writeSink <- message
	}
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
	Storage      *SafeStorage
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

type SafeStorage struct {
	mu      sync.RWMutex
	storage map[string]any
}

func NewSafeStorage() *SafeStorage {
	return &SafeStorage{
		storage: make(map[string]any),
	}
}

func (s *SafeStorage) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.storage[key] = value
}

func (s *SafeStorage) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.storage[key], s.storage[key] != nil
}
