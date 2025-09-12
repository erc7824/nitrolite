package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func getValidator() *validator.Validate {
	validate := validator.New()

	if err := validate.RegisterValidation("bigint", func(fl validator.FieldLevel) bool {
		n := new(big.Int)
		_, ok := n.SetString(fmt.Sprint(fl.Field()), 10)
		return ok
	}); err != nil {
		panic(fmt.Sprintf("failed to register bigint validation: %v", err))
	}
	return validate
}

const (
	defaultRPCErrorMessage = "an error occurred while processing the request"
)

const (
	// rpcNodeGroupHandlerPrefix is the prefix used for all handler group IDs
	rpcNodeGroupHandlerPrefix = "group."
	// rpcNodeGroupRoot is the identifier for the root handler group
	rpcNodeGroupRoot = "root"
)

var (
	defaultRPCMessageWriteDuration = 5 * time.Second // Default timeout for writing messages to WebSocket
)

// RPCNode is a WebSocket-based RPC server that handles incoming connections,
// routes messages to registered handlers and signs all responses.
// It supports middleware chains and handler groups for organizing endpoints.
type RPCNode struct {
	// upgrader handles the HTTP to WebSocket protocol upgrade
	upgrader websocket.Upgrader

	// groupId identifies this node's handler group (defaults to "group.root")
	groupId string
	// handlerChain maps handler IDs to their middleware/handler chains
	handlerChain map[string][]RPCHandler
	// routes maps RPC method names to their handler chain path (e.g., ["group.root", "group.private", "method"])
	routes map[string][]string

	// signer is used to cryptographically sign all outgoing messages
	signer *Signer
	// connHub manages all active WebSocket connections and user mappings
	connHub *rpcConnectionHub
	// logger for structured logging
	logger Logger

	// Event handlers for connection lifecycle
	onConnectHandlers       []func(send SendRPCMessageFunc)
	onDisconnectHandlers    []func(userID string)
	onMessageSentHandlers   []func()
	onAuthenticatedHandlers []func(userID string, send SendRPCMessageFunc)
}

// NewRPCNode creates a new RPC node instance with the provided signer and logger.
// The signer is used to sign all outgoing messages, ensuring message authenticity.
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

// HandleConnection is the main entry point for WebSocket connections.
// It upgrades the HTTP connection to WebSocket, manages concurrent read/write operations,
// processes incoming RPC messages, and handles connection lifecycle events.
// This method blocks until the connection is closed.
func (n *RPCNode) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := n.upgrader.Upgrade(w, r, nil)
	if err != nil {
		n.logger.Error("failed to upgrade connection to WebSocket", "error", err)
		return
	}
	defer conn.Close()

	connectionID := uuid.NewString()
	rpcConnection := NewRPCConnection(connectionID, "", conn, n.logger, n.onMessageSentHandlers...)
	if err := n.connHub.Add(rpcConnection); err != nil {
		n.logger.Error("failed to add connection to hub", "error", err, "connectionID", connectionID)
		return
	}

	// Notify all onConnect handlers about the new connection
	for _, handler := range n.onConnectHandlers {
		handler(n.getSendMessageFunc(rpcConnection))
	}

	// Cleanup function executed when connection closes
	defer func() {
		userID := rpcConnection.UserID()
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
	abortOthers := func() {
		cancel()  // Trigger exit on other goroutines
		wg.Done() // Decrement the wait group counter
	}

	go rpcConnection.Serve(parentCtx, abortOthers)
	go n.processMessages(rpcConnection, parentCtx, abortOthers)

	wg.Wait()
}

// processMesages handles incoming messages from the RPCConnection.
// It validates messages, routes them to appropriate handlers, and manages authentication.
func (n *RPCNode) processMessages(rpcConn *RPCConnection, ctx context.Context, abortOthers context.CancelFunc) {
	defer abortOthers() // Stop other goroutines when done
	safeStorage := NewSafeStorage()

read_loop:
	for {
		var messageBytes []byte
		select {
		case <-ctx.Done():
			n.logger.Debug("context done, stopping message processing")
			return
		case messageBytes = <-rpcConn.ProcessSink():
			if len(messageBytes) == 0 {
				return // Exit if the message is empty (connection closed)
			}
		}

		msg := RPCMessage{Req: &RPCData{}}
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			n.logger.Debug("invalid message format", "error", err, "message", string(messageBytes))
			n.sendErrorResponse(rpcConn, msg.Req.RequestID, "invalid message format")
			continue
		}

		if err := getValidator().Struct(&msg); err != nil {
			n.logger.Debug("message validation failed", "error", err, "message", string(messageBytes))
			n.sendErrorResponse(rpcConn, 0, "message validation failed")
			continue
		}
		if msg.Req == nil {
			n.logger.Debug("message request is empty", "message", string(messageBytes))
			n.sendErrorResponse(rpcConn, 0, "message request is empty")
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
			"userID", rpcConn.UserID(),
			"method", msg.Req.Method,
			"route", methodRoute)

		ctx := &RPCContext{
			Context:  context.Background(),
			UserID:   rpcConn.UserID(),
			Signer:   n.signer,
			Message:  msg,
			handlers: routeHandlers,
			Storage:  safeStorage,
		}
		ctx.Next() // Start processing the handlers

		responseBytes, err := ctx.GetRawResponse()
		if err != nil {
			n.logger.Error("failed to prepare response", "error", err, "method", msg.Req.Method)
			continue
		}
		rpcConn.Write(responseBytes)

		// Handle re-authentication
		if rpcConn.UserID() != ctx.UserID {
			// If the user ID changed during processing, do the re-authentication
			n.connHub.Reauthenticate(rpcConn.ConnectionID(), ctx.UserID)

			// Notify authenticated handlers about the new user ID
			for _, handler := range n.onAuthenticatedHandlers {
				handler(ctx.UserID, n.getSendMessageFunc(rpcConn))
			}
		}
	}
}

// RPCHandler is a function that processes an RPC request.
// Handlers can call c.Next() to pass control to the next handler in the chain.
type RPCHandler func(c *RPCContext)

// SendRPCMessageFunc is a function type for sending RPC notifications to a connection.
// It's provided to event handlers to allow server-initiated messages.
type SendRPCMessageFunc func(method string, params RPCDataParams)

// RPCContext contains all the information about an RPC request and provides
// methods for handlers to process and respond to the request.
type RPCContext struct {
	// Context is the standard Go context for the request
	Context context.Context
	// UserID is the authenticated user's identifier (empty if not authenticated)
	UserID string
	// Signer is used to sign the response message
	Signer *Signer
	// Message contains the incoming request and will hold the response
	Message RPCMessage
	// Storage provides per-connection storage for session data
	Storage *SafeStorage

	// handlers is the remaining handler chain to execute
	handlers []RPCHandler
}

// Next executes the next handler in the middleware chain.
// If there are no more handlers, it returns without doing anything.
func (c *RPCContext) Next() {
	if len(c.handlers) == 0 {
		return
	}

	handler := c.handlers[0]
	c.handlers = c.handlers[1:]
	handler(c)
}

// Succeed sets a successful response with the given method and parameters.
// This should be called by handlers to indicate successful processing.
func (c *RPCContext) Succeed(method string, params RPCDataParams) {
	c.Message.Res = &RPCData{
		RequestID: c.Message.Req.RequestID,
		Method:    method,
		Params:    params,
		Timestamp: uint64(time.Now().UnixMilli()),
	}
}

// Fail sets an error response for the RPC request. This method should be called by handlers
// when an error occurs during request processing.
//
// Error handling behavior:
//   - If err is an RPCError: The exact error message is sent to the client
//   - If err is any other error type: The fallbackMessage is sent to the client
//   - If both err is nil/non-RPCError AND fallbackMessage is empty: A generic error message is sent
//
// This design allows handlers to control what error information is exposed to clients:
//   - Use RPCError for client-safe, descriptive error messages
//   - Use regular errors with a fallbackMessage to hide internal error details
//
// Usage examples:
//
//	// Hide internal error details from client
//	balance, err := ledger.GetBalance(account)
//	if err != nil {
//		c.Fail(err, "failed to retrieve balance")
//		return
//	}
//
//	// Validation error with no internal error
//	if len(params) < 3 {
//		c.Fail(nil, "invalid parameters: expected at least 3")
//		return
//	}
//
// The response will have Method="error" and Params containing the error message.
func (c *RPCContext) Fail(err error, fallbackMessage string) {
	message := fallbackMessage
	if _, ok := err.(RPCError); ok {
		message = err.Error()
		fmt.Println(err)
	}
	if message == "" {
		message = defaultRPCErrorMessage
	}

	c.Message.Res = &RPCData{
		RequestID: c.Message.Req.RequestID,
		Method:    "error",
		Params:    ErrorResponse{Error: message},
		Timestamp: uint64(time.Now().UnixMilli()),
	}
}

// GetRawResponse returns the signed response message as raw bytes.
// This is called internally after handler processing to prepare the response.
func (c *RPCContext) GetRawResponse() ([]byte, error) {
	return prepareRawRPCResponse(c.Signer, c.Message.Res)
}

// prepareRawRPCResponse creates a signed RPC response message from the given data.
// It marshals the data, signs it, and returns the complete message as bytes.
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
		Sig: []Signature{signature},
	}
	resMessageBytes, err := json.Marshal(responseMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response message: %w", err)
	}

	return resMessageBytes, nil
}

// prepareRawNotification creates a signed server-initiated notification message.
// Unlike responses, notifications don't correspond to a specific request.
func prepareRawNotification(signer *Signer, method string, params RPCDataParams) ([]byte, error) {
	if params == nil {
		params = struct{}{}
	}

	data := &RPCData{
		RequestID: 0, // Notifications have no request ID
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

// NewGroup creates a new handler group with the given name.
// Groups allow organizing handlers with shared middleware.
// Example: privGroup := node.NewGroup("private"); privGroup.Use(authMiddleware)
func (wn *RPCNode) NewGroup(name string) *RPCHandlerGroup {
	return &RPCHandlerGroup{
		groupId:     rpcNodeGroupHandlerPrefix + name,
		routePrefix: []string{wn.groupId},
		root:        wn,
	}
}

// Handle registers a handler for the specified RPC method.
// The handler will be called when a message with the matching method is received.
func (wn *RPCNode) Handle(method string, handler RPCHandler) {
	wn.handle(method, handler)
	wn.routes[method] = []string{wn.groupId, method}
}

// handle is the internal method for registering handlers.
// It validates inputs and stores the handler in the handler chain.
func (wn *RPCNode) handle(method string, handler RPCHandler) {
	if method == "" {
		panic("Websocket method cannot be empty")
	}
	if handler == nil {
		panic(fmt.Sprintf("Websocket handler cannot be nil for method %s", method))
	}

	wn.handlerChain[method] = []RPCHandler{handler}
}

// Use adds middleware to the root handler group.
// Middleware will be executed for all requests before reaching the final handler.
func (wn *RPCNode) Use(middleware RPCHandler) {
	wn.use(wn.groupId, middleware)
}

// use is the internal method for adding middleware to a specific group.
// Middleware is appended to the group's handler chain.
func (wn *RPCNode) use(groupId string, middleware RPCHandler) {
	if middleware == nil {
		panic("Websocket middleware handler cannot be nil for group")
	}

	if _, exists := wn.handlerChain[groupId]; !exists {
		wn.handlerChain[groupId] = []RPCHandler{}
	}

	wn.handlerChain[groupId] = append(wn.handlerChain[groupId], middleware)
}

// OnConnect registers a handler to be called when a new WebSocket connection is established.
// The handler receives a send function for sending messages to the new connection.
func (wn *RPCNode) OnConnect(handler func(send SendRPCMessageFunc)) {
	wn.onConnectHandlers = append(wn.onConnectHandlers, handler)
}

// OnDisconnect registers a handler to be called when a WebSocket connection is closed.
// The handler receives the user ID if the connection was authenticated.
func (wn *RPCNode) OnDisconnect(handler func(userID string)) {
	wn.onDisconnectHandlers = append(wn.onDisconnectHandlers, handler)
}

// OnMessageSent registers a handler to be called after a message is sent to a client.
// This can be used for metrics, logging, or other post-send operations.
func (wn *RPCNode) OnMessageSent(handler func()) {
	wn.onMessageSentHandlers = append(wn.onMessageSentHandlers, handler)
}

// OnAuthenticated registers a handler to be called when a connection successfully authenticates.
// The handler receives the user ID and a send function for the authenticated connection.
func (wn *RPCNode) OnAuthenticated(handler func(userID string, send SendRPCMessageFunc)) {
	wn.onAuthenticatedHandlers = append(wn.onAuthenticatedHandlers, handler)
}

// Notify sends a server-initiated notification to a specific authenticated user.
// If the user is not connected, the notification is silently dropped.
func (wn *RPCNode) Notify(userID, method string, params RPCDataParams) {
	message, err := prepareRawNotification(wn.signer, method, params)
	if err != nil {
		wn.logger.Error("failed to prepare notification message", "error", err, "userID", userID, "method", method)
		return
	}

	wn.connHub.Publish(userID, message)
}

// getSendMessageFunc creates a SendRPCMessageFunc for a specific connection.
// The returned function can be used to send notifications to that connection.
func (wn *RPCNode) getSendMessageFunc(conn *RPCConnection) SendRPCMessageFunc {
	return func(method string, params RPCDataParams) {
		message, err := prepareRawNotification(wn.signer, method, params)
		if err != nil {
			wn.logger.Error("failed to prepare notification message", "error", err, "method", method)
			return
		}

		if conn == nil {
			wn.logger.Error("RPCConnection is nil, cannot send message", "method", method)
			return
		}

		conn.Write(message)
	}
}

// sendErrorResponse sends an error response to a connection.
// It's used for protocol-level errors before request processing.
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
		Params:    ErrorResponse{Error: message},
		Timestamp: uint64(time.Now().UnixMilli()),
	}

	responseBytes, err := prepareRawRPCResponse(wn.signer, data)
	if err != nil {
		wn.logger.Error("failed to prepare error response", "error", err)
		return
	}

	conn.Write(responseBytes)
}

// RPCHandlerGroup represents a collection of handlers with shared middleware.
// Groups can be nested to create hierarchical middleware chains.
type RPCHandlerGroup struct {
	// groupId is the unique identifier for this group
	groupId string
	// routePrefix contains the chain of group IDs leading to this group
	routePrefix []string
	// root is a reference to the RPCNode this group belongs to
	root *RPCNode
}

// NewGroup creates a nested handler group within this group.
// The new group inherits all middleware from parent groups.
func (hg *RPCHandlerGroup) NewGroup(name string) *RPCHandlerGroup {
	return &RPCHandlerGroup{
		groupId:     name,
		routePrefix: append(hg.routePrefix, hg.groupId),
		root:        hg.root,
	}
}

// Handle registers a handler for the specified RPC method within this group.
// The handler will execute after all group middleware in the chain.
func (hg *RPCHandlerGroup) Handle(method string, handler RPCHandler) {
	hg.root.routes[method] = append(hg.routePrefix, hg.groupId, method)
	hg.root.handle(method, handler)
}

// Use adds middleware to this handler group.
// The middleware will execute for all handlers registered in this group.
func (hg *RPCHandlerGroup) Use(middleware RPCHandler) {
	hg.root.use(hg.groupId, middleware)
}

// SafeStorage provides thread-safe key-value storage for connection-specific data.
// It's used to store session information, authentication policies, and other per-connection state.
type SafeStorage struct {
	// mu protects concurrent access to the storage map
	mu sync.RWMutex
	// storage holds the key-value pairs
	storage map[string]any
}

// NewSafeStorage creates a new thread-safe storage instance.
func NewSafeStorage() *SafeStorage {
	return &SafeStorage{
		storage: make(map[string]any),
	}
}

// Set stores a value with the given key.
// If the key already exists, its value is overwritten.
func (s *SafeStorage) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.storage[key] = value
}

// Get retrieves a value by key.
// Returns the value and true if found, or nil and false if not found.
func (s *SafeStorage) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.storage[key], s.storage[key] != nil
}
