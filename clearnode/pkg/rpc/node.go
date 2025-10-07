package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/erc7824/nitrolite/clearnode/pkg/log"
	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	nodeDefaultErrorMessage = "an error occurred while processing the request"
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
	OnConnect(handler func(send SendResponseFunc))
	OnDisconnect(handler func(userID string))
	OnMessageSent(handler func())
	OnAuthenticated(handler func(userID string, send SendResponseFunc))
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

	// groupId identifies this node's handler group (defaults to "group.root")
	groupId string
	// handlerChain maps handler IDs to their middleware/handler chains
	handlerChain map[string][]Handler
	// routes maps RPC method names to their handler chain path (e.g., ["group.root", "group.private", "method"])
	routes map[string][]string

	// signer is used to cryptographically sign all outgoing messages
	signer sign.Signer
	// connHub manages all active WebSocket connections and user mappings
	connHub *ConnectionHub
	// logger for structured logging
	logger log.Logger

	// Event handlers for connection lifecycle
	onConnectHandlers       []func(send SendResponseFunc)
	onDisconnectHandlers    []func(userID string)
	onMessageSentHandlers   []func()
	onAuthenticatedHandlers []func(userID string, send SendResponseFunc)
}

// NewWebsocketNode creates a new RPC node instance with the provided signer and logger.
// The signer is used to sign all outgoing messages, ensuring message authenticity.
func NewWebsocketNode(signer sign.Signer, logger log.Logger) *WebsocketNode {
	return &WebsocketNode{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for simplicity
			},
		},

		groupId:      nodeGroupHandlerPrefix + nodeGroupRoot,
		handlerChain: make(map[string][]Handler),
		routes:       make(map[string][]string),

		signer:  signer,
		connHub: NewConnectionHub(),
		logger:  logger.WithName("rpc-node"),

		onConnectHandlers:       []func(send SendResponseFunc){},
		onDisconnectHandlers:    []func(userID string){},
		onMessageSentHandlers:   []func(){},
		onAuthenticatedHandlers: []func(userID string, send SendResponseFunc){},
	}
}

// ServeHTTP is the main entry point for WebSocket connections.
// It upgrades the HTTP connection to WebSocket, manages concurrent read/write operations,
// processes incoming RPC messages, and handles connection lifecycle events.
// This method blocks until the connection is closed.
func (n *WebsocketNode) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wsConnection, err := n.upgrader.Upgrade(w, r, nil)
	if err != nil {
		n.logger.Error("failed to upgrade connection to WebSocket", "error", err)
		return
	}
	defer wsConnection.Close()

	connectionID := uuid.NewString()
	connection := NewWebsocketConnection(connectionID, "", wsConnection, n.logger, n.onMessageSentHandlers...)
	if err := n.connHub.Add(connection); err != nil {
		n.logger.Error("failed to add connection to hub", "error", err, "connectionID", connectionID)
		return
	}

	// Notify all onConnect handlers about the new connection
	for _, handler := range n.onConnectHandlers {
		handler(n.getSendResponseFunc(connection))
	}

	// Cleanup function executed when connection closes
	defer func() {
		userID := connection.UserID()
		n.connHub.Remove(connectionID)

		// Notify all onDisconnect handlers about the closed connection
		for _, handler := range n.onDisconnectHandlers {
			handler(userID)
		}

		n.logger.Info("connection closed", "connectionID", connectionID, "userID", userID)
	}()

	parentCtx, cancel := context.WithCancel(r.Context())
	wg := &sync.WaitGroup{}
	wg.Add(2)
	childHandleClosure := func(_ error) {
		cancel()  // Trigger exit on other goroutines
		wg.Done() // Decrement the wait group counter
	}

	go connection.Serve(parentCtx, childHandleClosure)
	go n.processRequests(connection, parentCtx, childHandleClosure)

	wg.Wait()
}

// processRequests handles incoming requests from the WebsocketConnection.
// It validates messages, routes them to appropriate handlers, and manages authentication.
func (n *WebsocketNode) processRequests(conn Connection, ctx context.Context, handleClosure func(error)) {
	defer handleClosure(nil) // Stop other goroutines when done
	safeStorage := NewSafeStorage()

read_loop:
	for {
		var messageBytes []byte
		select {
		case <-ctx.Done():
			n.logger.Debug("context done, stopping message processing")
			return
		case messageBytes = <-conn.RawRequests():
			if len(messageBytes) == 0 {
				return // Exit if the message is empty (connection closed)
			}
		}

		req := Request{}
		if err := json.Unmarshal(messageBytes, &req); err != nil {
			n.logger.Debug("invalid message format", "error", err, "message", string(messageBytes))
			n.sendErrorResponse(conn, req.Req.RequestID, "invalid message format")
			continue
		}

		methodRoute, ok := n.routes[req.Req.Method]
		if !ok || len(methodRoute) == 0 {
			n.logger.Debug("no handler found for method", "method", req.Req.Method)
			n.sendErrorResponse(conn, req.Req.RequestID, fmt.Sprintf("unknown method: %s", req.Req.Method))
			continue
		}

		var routeHandlers []Handler
		for _, handlersId := range methodRoute {
			handlers, exists := n.handlerChain[handlersId]
			if !exists || len(handlers) == 0 {
				n.logger.Error("no handlers found for id", "id", handlersId)
				n.sendErrorResponse(conn, req.Req.RequestID, fmt.Sprintf("unknown method: %s", req.Req.Method))
				continue read_loop
			}

			routeHandlers = append(routeHandlers, handlers...)
		}
		n.logger.Info("processing message",
			"requestID", req.Req.RequestID,
			"userID", conn.UserID(),
			"method", req.Req.Method,
			"route", methodRoute)

		ctx := &Context{
			Context:  context.Background(),
			UserID:   conn.UserID(),
			Signer:   n.signer,
			Request:  req,
			handlers: routeHandlers,
			Storage:  safeStorage,
		}
		ctx.Next() // Start processing the handlers

		responseBytes, err := ctx.GetRawResponse()
		if err != nil {
			n.logger.Error("failed to prepare response", "error", err, "method", req.Req.Method)
			continue
		}
		conn.WriteRawResponse(responseBytes)

		// Handle re-authentication
		if conn.UserID() != ctx.UserID {
			// If the user ID changed during processing, do the re-authentication
			n.connHub.Reauthenticate(conn.ConnectionID(), ctx.UserID)

			// Notify authenticated handlers about the new user ID
			for _, handler := range n.onAuthenticatedHandlers {
				handler(ctx.UserID, n.getSendResponseFunc(conn))
			}
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

// OnConnect registers a handler to be called when a new WebSocket connection is established.
// The handler receives a send function for sending messages to the new connection.
func (wn *WebsocketNode) OnConnect(handler func(send SendResponseFunc)) {
	wn.onConnectHandlers = append(wn.onConnectHandlers, handler)
}

// OnDisconnect registers a handler to be called when a WebSocket connection is closed.
// The handler receives the user ID if the connection was authenticated.
func (wn *WebsocketNode) OnDisconnect(handler func(userID string)) {
	wn.onDisconnectHandlers = append(wn.onDisconnectHandlers, handler)
}

// OnMessageSent registers a handler to be called after a message is sent to a client.
// This can be used for metrics, logging, or other post-send operations.
func (wn *WebsocketNode) OnMessageSent(handler func()) {
	wn.onMessageSentHandlers = append(wn.onMessageSentHandlers, handler)
}

// OnAuthenticated registers a handler to be called when a connection successfully authenticates.
// The handler receives the user ID and a send function for the authenticated connection.
func (wn *WebsocketNode) OnAuthenticated(handler func(userID string, send SendResponseFunc)) {
	wn.onAuthenticatedHandlers = append(wn.onAuthenticatedHandlers, handler)
}

// Notify sends a server-initiated notification to a specific authenticated user.
// If the user is not connected, the notification is silently dropped.
func (wn *WebsocketNode) Notify(userID, method string, params Params) {
	message, err := prepareRawNotification(wn.signer, method, params)
	if err != nil {
		wn.logger.Error("failed to prepare notification message", "error", err, "userID", userID, "method", method)
		return
	}

	wn.connHub.Publish(userID, message)
}

// getSendResponseFunc creates a SendResponseFunc for a specific connection.
// The returned function can be used to send notifications to that connection.
func (wn *WebsocketNode) getSendResponseFunc(conn Connection) SendResponseFunc {
	return func(method string, params Params) {
		responseBytes, err := prepareRawNotification(wn.signer, method, params)
		if err != nil {
			wn.logger.Error("failed to prepare notification message", "error", err, "method", method)
			return
		}

		if conn == nil {
			wn.logger.Error("RPCConnection is nil, cannot send message", "method", method)
			return
		}

		conn.WriteRawResponse(responseBytes)
	}
}

// sendErrorResponse sends an error response to a connection.
// It's used for protocol-level errors before request processing.
func (wn *WebsocketNode) sendErrorResponse(conn Connection, requestID uint64, message string) {
	if conn == nil {
		wn.logger.Error("connection is nil, cannot send error response", "requestID", requestID)
		return
	}

	res := NewErrorResponse(requestID, message)
	responseBytes, err := prepareRawResponse(wn.signer, res.Res)
	if err != nil {
		wn.logger.Error("failed to prepare error response", "error", err)
		return
	}

	conn.WriteRawResponse(responseBytes)
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
		groupId:     name,
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
