package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/erc7824/nitrolite/clearnode/pkg/log"
	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	defaultNodeErrorMessage = "an error occurred while processing the request"
)

const (
	// nodeGroupHandlerPrefix is the prefix used for all handler group IDs
	nodeGroupHandlerPrefix = "group."
	// nodeGroupRoot is the identifier for the root handler group
	nodeGroupRoot = "root"
)

type Node interface {
	Handle(method string, handler Handler)
	Notify(userID string, method string, params Params)
	Use(middleware Handler)
	NewGroup(name string) HandlerGroup
}

type HandlerGroup interface {
	Handle(method string, handler Handler)
	Use(middleware Handler)
	NewGroup(name string) HandlerGroup
}

var (
	_ Node         = &WebsocketNode{}
	_ http.Handler = &WebsocketNode{}

	_ HandlerGroup = &WebsocketHandlerGroup{}
)

// WebsocketNode is a WebSocket-based RPC server that handles incoming connections,
// routes messages to registered handlers and signs all responses.
// It supports middleware chains and handler groups for organizing endpoints.
type WebsocketNode struct {
	// upgrader handles the HTTP to WebSocket protocol upgrade
	upgrader websocket.Upgrader
	// cfg contains configuration for the node
	cfg WebsocketNodeConfig
	// groupId identifies this node's handler group (defaults to "group.root")
	groupId string
	// handlerChain maps handler IDs to their middleware/handler chains
	handlerChain map[string][]Handler
	// routes maps RPC method names to their handler chain path (e.g., ["group.root", "group.private", "method"])
	routes map[string][]string
	// connHub manages all active WebSocket connections
	connHub *ConnectionHub
}

type WebsocketNodeConfig struct {
	// Signer is used to sign all outgoing messages
	Signer sign.Signer
	// Logger is used for structured logging
	Logger log.Logger

	// OnConnectHandler is called when a new WebSocket connection is established
	OnConnectHandler func(send SendResponseFunc)
	// OnDisconnectHandler is called when a WebSocket connection is closed
	OnDisconnectHandler func(userID string)
	// OnMessageSentHandler is called after a message is sent to a client
	OnMessageSentHandler func([]byte)
	// OnAuthenticatedHandler is called when a connection successfully authenticates
	OnAuthenticatedHandler func(userID string, send SendResponseFunc)

	// WsUpgraderReadBufferSize is the size of the read buffer for WebSocket connections
	WsUpgraderReadBufferSize int
	// WsUpgraderWriteBufferSize is the size of the write buffer for WebSocket connections
	WsUpgraderWriteBufferSize int
	// WsUpgraderCheckOrigin is a function to check the origin of incoming WebSocket requests
	WsUpgraderCheckOrigin func(r *http.Request) bool

	// WsConnWriteTimeout is the timeout for writing messages to the WebSocket connection
	WsConnWriteTimeout time.Duration
	// WsConnWriteBufferSize is the size of the write buffer for WebSocket connections
	WsConnWriteBufferSize int
	// WsConnProcessBufferSize is the size of the process buffer for WebSocket connections
	WsConnProcessBufferSize int
}

// NewWebsocketNode creates a new RPC node instance with the provided configuration.
// The signer is used to sign all outgoing messages, ensuring message authenticity.
func NewWebsocketNode(config WebsocketNodeConfig) (*WebsocketNode, error) {
	if config.Signer == nil {
		return nil, fmt.Errorf("signer cannot be nil")
	}
	if config.Logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	config.Logger = config.Logger.WithName("rpc-node")

	if config.OnConnectHandler == nil {
		config.OnConnectHandler = func(send SendResponseFunc) {}
	}
	if config.OnDisconnectHandler == nil {
		config.OnDisconnectHandler = func(userID string) {}
	}
	if config.OnMessageSentHandler == nil {
		config.OnMessageSentHandler = func([]byte) {}
	}
	if config.OnAuthenticatedHandler == nil {
		config.OnAuthenticatedHandler = func(userID string, send SendResponseFunc) {}
	}
	if config.WsUpgraderReadBufferSize <= 0 {
		config.WsUpgraderReadBufferSize = 1024
	}
	if config.WsUpgraderWriteBufferSize <= 0 {
		config.WsUpgraderWriteBufferSize = 1024
	}
	if config.WsUpgraderCheckOrigin == nil {
		config.WsUpgraderCheckOrigin = func(r *http.Request) bool {
			return true // Allow all origins by default
		}
	}

	node := &WebsocketNode{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  config.WsUpgraderReadBufferSize,
			WriteBufferSize: config.WsUpgraderWriteBufferSize,
			CheckOrigin:     config.WsUpgraderCheckOrigin,
		},
		cfg:          config,
		groupId:      nodeGroupHandlerPrefix + nodeGroupRoot,
		handlerChain: make(map[string][]Handler),
		routes:       make(map[string][]string),
		connHub:      NewConnectionHub(),
	}

	node.Handle(PingMethod.String(), node.handlePing) // Built-in ping handler

	return node, nil
}

// ServeHTTP is the main entry point for WebSocket connections.
// It upgrades the HTTP connection to WebSocket, manages concurrent read/write operations,
// processes incoming RPC messages, and handles connection lifecycle events.
// This method blocks until the connection is closed.
func (wn *WebsocketNode) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wsConnection, err := wn.upgrader.Upgrade(w, r, nil)
	if err != nil {
		wn.cfg.Logger.Error("failed to upgrade connection to WebSocket", "error", err)
		return
	}
	defer wsConnection.Close()

	connectionID := uuid.NewString()

	connConfig := WebsocketConnectionConfig{
		ConnectionID:         connectionID,
		WebsocketConn:        wsConnection,
		Logger:               wn.cfg.Logger,
		OnMessageSentHandler: wn.cfg.OnMessageSentHandler,
	}
	connection, err := NewWebsocketConnection(connConfig)
	if err != nil {
		wn.cfg.Logger.Error("failed to create WebSocket connection", "error", err, "connectionID", connectionID)
		return
	}
	if err := wn.connHub.Add(connection); err != nil {
		wn.cfg.Logger.Error("failed to add connection to hub", "error", err, "connectionID", connectionID)
		return
	}

	wn.cfg.OnConnectHandler(wn.getSendResponseFunc(connection))
	wn.cfg.Logger.Info("new WebSocket connection established", "connectionID", connectionID, "userID", connection.UserID())

	// Cleanup function executed when connection closes
	defer func() {
		userID := connection.UserID()
		wn.connHub.Remove(connectionID)

		wn.cfg.OnDisconnectHandler(userID)
		wn.cfg.Logger.Info("connection closed", "connectionID", connectionID, "userID", userID)
	}()

	parentCtx, cancel := context.WithCancel(r.Context())
	wg := &sync.WaitGroup{}
	wg.Add(2)
	childHandleClosure := func(_ error) {
		cancel()  // Trigger exit on other goroutines
		wg.Done() // Decrement the wait group counter
	}

	go connection.Serve(parentCtx, childHandleClosure)
	go wn.processRequests(connection, parentCtx, childHandleClosure)

	wg.Wait()
}

// processRequests handles incoming requests from the WebsocketConnection.
// It validates messages, routes them to appropriate handlers, and manages authentication.
func (wn *WebsocketNode) processRequests(conn Connection, parentCtx context.Context, handleClosure func(error)) {
	defer handleClosure(nil) // Stop other goroutines when done
	safeStorage := NewSafeStorage()

	for {
		var messageBytes []byte
		select {
		case <-parentCtx.Done():
			wn.cfg.Logger.Debug("context done, stopping message processing")
			return
		case messageBytes = <-conn.RawRequests():
			if len(messageBytes) == 0 {
				return // Exit if the message is empty (connection closed)
			}
		}

		req := Request{}
		if err := json.Unmarshal(messageBytes, &req); err != nil {
			wn.cfg.Logger.Debug("invalid message format", "error", err, "message", string(messageBytes))
			wn.sendErrorResponse(conn, req.Req.RequestID, "invalid message format")
			continue
		}

		methodRoute, ok := wn.routes[req.Req.Method]
		if !ok || len(methodRoute) == 0 {
			wn.cfg.Logger.Debug("no handlers' route found for method", "method", req.Req.Method)
			wn.sendErrorResponse(conn, req.Req.RequestID, fmt.Sprintf("unknown method: %s", req.Req.Method))
			continue
		}

		var routeHandlers []Handler
		for _, handlersId := range methodRoute {
			handlers, exists := wn.handlerChain[handlersId]
			if !exists || len(handlers) == 0 {
				routeHandlers = nil
				wn.cfg.Logger.Error("no handlers found for id", "id", handlersId)
				break
			}

			routeHandlers = append(routeHandlers, handlers...)
		}
		if len(routeHandlers) == 0 {
			wn.sendErrorResponse(conn, req.Req.RequestID, fmt.Sprintf("unknown method: %s", req.Req.Method))
			continue
		}

		wn.cfg.Logger.Info("processing message",
			"requestID", req.Req.RequestID,
			"userID", conn.UserID(),
			"method", req.Req.Method,
			"route", methodRoute)

		ctx := &Context{
			Context:  parentCtx,
			UserID:   conn.UserID(),
			Signer:   wn.cfg.Signer,
			Request:  req,
			handlers: routeHandlers,
			Storage:  safeStorage,
		}
		ctx.Next() // Start processing the handlers

		responseBytes, err := ctx.GetRawResponse()
		if err != nil {
			wn.sendErrorResponse(conn, req.Req.RequestID, defaultNodeErrorMessage)
			wn.cfg.Logger.Error("failed to prepare response", "error", err, "method", req.Req.Method)
			continue
		}
		conn.WriteRawResponse(responseBytes)

		// Handle re-authentication
		if conn.UserID() != ctx.UserID {
			// If the user ID changed during processing, do the re-authentication
			wn.connHub.Reauthenticate(conn.ConnectionID(), ctx.UserID)
			wn.cfg.OnAuthenticatedHandler(ctx.UserID, wn.getSendResponseFunc(conn))
		}
	}
}

// NewGroup creates a new handler group with the given name.
// Groups allow organizing handlers with shared middleware.
// Example: privGroup := node.NewGroup("private"); privGroup.Use(authMiddleware)
func (wn *WebsocketNode) NewGroup(name string) HandlerGroup {
	return &WebsocketHandlerGroup{
		groupId:     nodeGroupHandlerPrefix + name,
		routePrefix: []string{wn.groupId},
		root:        wn,
	}
}

// Handle registers a handler for the specified RPC method.
// The handler will be called when a message with the matching method is received.
func (wn *WebsocketNode) Handle(method string, handler Handler) {
	wn.handle(method, handler)
	wn.routes[method] = []string{wn.groupId, method}
}

// handle is the internal method for registering handlers.
// It validates inputs and stores the handler in the handler chain.
func (wn *WebsocketNode) handle(method string, handler Handler) {
	if method == "" {
		panic("Websocket method cannot be empty")
	}
	if handler == nil {
		panic(fmt.Sprintf("Websocket handler cannot be nil for method %s", method))
	}

	wn.handlerChain[method] = []Handler{handler}
}

// Use adds middleware to the root handler group.
// Middleware will be executed for all requests before reaching the final handler.
func (wn *WebsocketNode) Use(middleware Handler) {
	wn.use(wn.groupId, middleware)
}

// use is the internal method for adding middleware to a specific group.
// Middleware is appended to the group's handler chain.
func (wn *WebsocketNode) use(groupId string, middleware Handler) {
	if middleware == nil {
		panic("Websocket middleware handler cannot be nil for group")
	}

	if _, exists := wn.handlerChain[groupId]; !exists {
		wn.handlerChain[groupId] = []Handler{}
	}

	wn.handlerChain[groupId] = append(wn.handlerChain[groupId], middleware)
}

// Notify sends a server-initiated notification to a specific authenticated user.
// If the user is not connected, the notification is silently dropped.
func (wn *WebsocketNode) Notify(userID, method string, params Params) {
	message, err := prepareRawNotification(wn.cfg.Signer, method, params)
	if err != nil {
		wn.cfg.Logger.Error("failed to prepare notification message", "error", err, "userID", userID, "method", method)
		return
	}

	wn.connHub.Publish(userID, message)
}

// getSendResponseFunc creates a SendResponseFunc for a specific connection.
// The returned function can be used to send notifications to that connection.
func (wn *WebsocketNode) getSendResponseFunc(conn Connection) SendResponseFunc {
	return func(method string, params Params) {
		responseBytes, err := prepareRawNotification(wn.cfg.Signer, method, params)
		if err != nil {
			wn.cfg.Logger.Error("failed to prepare notification message", "error", err, "method", method)
			return
		}

		if conn == nil {
			wn.cfg.Logger.Error("RPCConnection is nil, cannot send message", "method", method)
			return
		}

		conn.WriteRawResponse(responseBytes)
	}
}

// sendErrorResponse sends an error response to a connection.
// It's used for protocol-level errors before request processing.
func (wn *WebsocketNode) sendErrorResponse(conn Connection, requestID uint64, message string) {
	if conn == nil {
		wn.cfg.Logger.Error("connection is nil, cannot send error response", "requestID", requestID)
		return
	}

	res := NewErrorResponse(requestID, message)
	responseBytes, err := prepareRawResponse(wn.cfg.Signer, res.Res)
	if err != nil {
		wn.cfg.Logger.Error("failed to prepare error response", "error", err)
		return
	}

	conn.WriteRawResponse(responseBytes)
}

// handlePing is a built-in handler for the "ping" method.
// It responds with a "pong" message and can be used for keep-alive checks.
func (wn *WebsocketNode) handlePing(ctx *Context) {
	ctx.Next() // Call any middleware first
	ctx.Succeed(PongMethod.String(), nil)
}

// prepareRawNotification creates a signed server-initiated notification message.
// Unlike responses, notifications don't correspond to a specific request.
func prepareRawNotification(signer sign.Signer, method string, params Params) ([]byte, error) {
	payload := NewPayload(0, method, params) // RequestID=0 for notifications

	responseBytes, err := prepareRawResponse(signer, payload)
	if err != nil {
		return nil, err
	}

	return responseBytes, nil
}

// WebsocketHandlerGroup represents a collection of handlers with shared middleware.
// Groups can be nested to create hierarchical middleware chains.
type WebsocketHandlerGroup struct {
	// groupId is the unique identifier for this group
	groupId string
	// routePrefix contains the chain of group IDs leading to this group
	routePrefix []string
	// root is a reference to the Node this group belongs to
	root *WebsocketNode
}

// NewGroup creates a nested handler group within this group.
// The new group inherits all middleware from parent groups.
func (hg *WebsocketHandlerGroup) NewGroup(name string) HandlerGroup {
	return &WebsocketHandlerGroup{
		groupId:     fmt.Sprintf("%s.%s", hg.groupId, name),
		routePrefix: append(hg.routePrefix, hg.groupId),
		root:        hg.root,
	}
}

// Handle registers a handler for the specified RPC method within this group.
// The handler will execute after all group middleware in the chain.
func (hg *WebsocketHandlerGroup) Handle(method string, handler Handler) {
	hg.root.routes[method] = append(hg.routePrefix, hg.groupId, method)
	hg.root.handle(method, handler)
}

// Use adds middleware to this handler group.
// The middleware will execute for all handlers registered in this group.
func (hg *WebsocketHandlerGroup) Use(middleware Handler) {
	hg.root.use(hg.groupId, middleware)
}
