package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

var validate = validator.New()

// UnifiedWSHandler manages WebSocket connections with authentication
type UnifiedWSHandler struct {
	signer        *Signer
	db            *gorm.DB
	upgrader      websocket.Upgrader
	connections   map[string]*websocket.Conn
	connectionsMu sync.RWMutex
	authManager   *AuthManager
	metrics       *Metrics
	rpcStore      *RPCStore
	config        *Config
	logger        Logger
}

func NewUnifiedWSHandler(
	signer *Signer,
	db *gorm.DB,
	metrics *Metrics,
	rpcStore *RPCStore,
	config *Config,
	logger Logger,
) (*UnifiedWSHandler, error) {
	authManager, err := NewAuthManager(signer.GetPrivateKey())

	if err != nil {
		return nil, err
	}

	return &UnifiedWSHandler{
		signer: signer,
		db:     db,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for testing; should be restricted in production
			},
		},
		connections: make(map[string]*websocket.Conn),
		authManager: authManager,
		metrics:     metrics,
		rpcStore:    rpcStore,
		config:      config,
		logger:      logger.NewSystem("ws-handler"),
	}, nil
}

// HandleConnection handles the WebSocket connection lifecycle.
func (h *UnifiedWSHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("failed to upgrade connection to WebSocket", "error", err)
		return
	}
	defer conn.Close()

	// Increment connection metrics
	h.metrics.ConnectionsTotal.Inc()
	h.metrics.ConnectedClients.Inc()
	defer h.metrics.ConnectedClients.Dec()

	var signerAddress string
	var policy *Policy
	var authenticated bool

	// Send assets immediately upon connection (before authentication)
	h.sendAssets(conn)

	// Read messages until authentication completes
	for !authenticated {
		ctx := context.Background()
		ctx = SetContextLogger(ctx, h.logger)
		logger := LoggerFromContext(ctx)

		_, message, err := conn.ReadMessage()
		if err != nil {
			logger.Error("failed to read message", "error", err)
			return
		}

		// Increment received message counter
		h.metrics.MessageReceived.Inc()

		var rpcMsg RPCMessage
		if err := json.Unmarshal(message, &rpcMsg); err != nil {
			logger.Debug("invalid message format", "error", err, "message", string(message))
			h.sendErrorResponse("", nil, conn, "Invalid message format")
			return
		}

		if err := validate.Struct(&rpcMsg); err != nil {
			logger.Debug("message validation failed", "error", err, "message", string(message))
			h.sendErrorResponse("", nil, conn, "Invalid message format")
			return
		}

		// Handle message based on the method
		switch rpcMsg.Req.Method {
		// Public endpoints
		case "ping", "get_config", "get_assets", "get_app_definition", "get_app_sessions", "get_channels", "get_ledger_entries":
			var rpcResponse *RPCMessage
			var handlerErr error

			switch rpcMsg.Req.Method {
			case "ping":
				rpcResponse, handlerErr = HandlePing(&rpcMsg)
			case "get_config":
				rpcResponse, handlerErr = HandleGetConfig(&rpcMsg, h.config, h.signer)
			case "get_assets":
				rpcResponse, handlerErr = HandleGetAssets(&rpcMsg, h.db)
			case "get_app_definition":
				rpcResponse, handlerErr = HandleGetAppDefinition(&rpcMsg, h.db)
			case "get_app_sessions":
				rpcResponse, handlerErr = HandleGetAppSessions(&rpcMsg, h.db)
			case "get_channels":
				rpcResponse, handlerErr = HandleGetChannels(&rpcMsg, h.db)
			case "get_ledger_entries":
				rpcResponse, handlerErr = HandleGetLedgerEntries(&rpcMsg, "", h.db)
			}

			if handlerErr != nil {
				logger.Error("failed to handle public method", "method", rpcMsg.Req.Method, "error", handlerErr)
				h.sendErrorResponse("", nil, conn, fmt.Sprintf("Failed to process %s: %v", rpcMsg.Req.Method, handlerErr))
			} else {
				byteData, _ := json.Marshal(rpcResponse.Res)
				signature, _ := h.signer.Sign(byteData)
				rpcResponse.Sig = []string{hexutil.Encode(signature)}
				wsResponseData, _ := json.Marshal(rpcResponse)

				if err := h.writeWSResponse(conn, wsResponseData); err != nil {
					// Track RPC request failure by method
					h.incrementRPCRequestCount(rpcMsg.Req.Method, "failure")
					continue
				}

				// Track RPC request success by method
				h.incrementRPCRequestCount(rpcMsg.Req.Method, "success")
			}
			continue

		case "auth_request":
			// Track auth request metrics
			h.metrics.AuthRequests.Inc()

			// Client is initiating authentication
			err := HandleAuthRequest(ctx, h.signer, conn, &rpcMsg, h.authManager)
			if err != nil {
				logger.Debug("failed to handle auth request", "error", err)
				h.sendErrorResponse("", nil, conn, err.Error())
			}
			continue

		case "auth_verify":
			// Client is responding to a challenge
			authPolicy, authMethod, err := HandleAuthVerify(ctx, conn, &rpcMsg, h.authManager, h.signer, h.db)

			// Record metrics
			h.metrics.AuthAttemptsTotal.With(prometheus.Labels{
				"auth_method": authMethod,
			}).Inc()

			if err != nil {
				logger.Debug("failed to verify authentication", "error", err, "method", authMethod)
				h.sendErrorResponse("", nil, conn, err.Error())
				h.metrics.AuthAttempsFail.With(prometheus.Labels{
					"auth_method": authMethod,
				}).Inc()
				continue
			}

			h.metrics.AuthAttempsSuccess.With(prometheus.Labels{
				"auth_method": authMethod,
			}).Inc()

			// Authentication successful
			policy = authPolicy
			signerAddress = authPolicy.Wallet
			authenticated = true

		default:
			// Reject methods except for public endpoints and auth methods
			logger.Debug("unexpected method call within unauthenticated connection", "method", rpcMsg.Req.Method)
			h.sendErrorResponse("", nil, conn, "Unexpected method call within unauthenticated connection")
		}
	}

	walletAddress := GetWalletBySigner(signerAddress)
	if walletAddress == "" {
		walletAddress = signerAddress
	}
	logger := h.logger.With("walletAddress", walletAddress)
	logger.Info("connection authentication successful", "signerAddress", signerAddress)

	// Store connection for authenticated user
	h.connectionsMu.Lock()
	// Currently, only one connection per wallet is allowed
	h.connections[walletAddress] = conn
	h.connectionsMu.Unlock()

	defer func() {
		h.connectionsMu.Lock()
		delete(h.connections, walletAddress)
		h.connectionsMu.Unlock()
		logger.Info("connection closed", "signerAddress", signerAddress)
		// TODO: Remove signer from DB and cache
	}()

	// Send initial balance and channels information in form of balance and channel updates
	channels, err := getChannelsByWallet(h.db, walletAddress, string(ChannelStatusOpen))
	if err != nil {
		logger.Error("error retrieving channels for participant", "error", err)
	}

	h.sendChannelsUpdate(walletAddress, channels)
	h.sendBalanceUpdate(walletAddress)

	for {
		ctx := context.Background()
		ctx = SetContextLogger(ctx, logger)
		logger = LoggerFromContext(ctx)

		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket connection closed with unexpected reason", "error", err)
			} else {
				logger.Error("failed to read message", "error", err)
			}
			break
		}

		// Increment received message counter
		h.metrics.MessageReceived.Inc()

		// Check if session is still valid
		if !h.authManager.ValidateSession(signerAddress) {
			logger.Debug("session expired", "signerAddress", signerAddress)
			h.sendErrorResponse(signerAddress, nil, conn, "Session expired. Please re-authenticate.")
			break
		}

		// Update session activity timestamp
		h.authManager.UpdateSession(signerAddress)

		// Forward request or response for internal vApp communication.
		var msg RPCMessage
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			logger.Debug("invalid message format", "error", err, "message", string(messageBytes))
			h.sendErrorResponse(walletAddress, nil, conn, "Invalid message format")
			continue
		}

		if err := validate.Struct(&msg); err != nil {
			logger.Debug("message validation failed", "error", err, "message", string(messageBytes))
			h.sendErrorResponse(walletAddress, nil, conn, "Invalid message format")
			return
		}

		if msg.AppSessionID != "" {
			if err := forwardMessage(ctx, &msg, messageBytes, walletAddress, h); err != nil {
				h.sendErrorResponse(walletAddress, nil, conn, "Failed to forward message: "+err.Error())
				continue
			}
			continue
		}

		if msg.Req == nil {
			continue
		}

		if err = ValidateTimestamp(msg.Req.Timestamp, h.config.msgExpiryTime); err != nil {
			logger.Debug("invalid message timestamp", "error", err)
			h.sendErrorResponse(walletAddress, &msg, conn, fmt.Sprintf("Message timestamp validation failed: %v", err))
			continue
		}

		var rpcResponse = &RPCMessage{}
		var handlerErr error
		var recordHistory = false

		if policy == nil {
			h.sendErrorResponse(walletAddress, &msg, conn, "Policy not found for the user")
			continue
		}

		logger.Debug("handling RPC request", "method", msg.Req.Method, "message", string(messageBytes))

		switch msg.Req.Method {
		case "ping":
			rpcResponse, handlerErr = HandlePing(&msg)
			if handlerErr != nil {
				logger.Error("error handling ping", "error", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to process ping: "+handlerErr.Error())
				continue
			}

		case "get_config":
			rpcResponse, handlerErr = HandleGetConfig(&msg, h.config, h.signer)
			if handlerErr != nil {
				logger.Error("error handling get_config", "error", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to get config: "+handlerErr.Error())
				continue
			}

		case "get_assets":
			rpcResponse, handlerErr = HandleGetAssets(&msg, h.db)
			if handlerErr != nil {
				logger.Error("error handling get_assets", "error", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to get assets: "+handlerErr.Error())
				continue
			}

		case "get_ledger_balances":
			rpcResponse, handlerErr = HandleGetLedgerBalances(&msg, walletAddress, h.db)
			if handlerErr != nil {
				logger.Error("error handling get_ledger_balances", "error", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to get ledger balances: "+handlerErr.Error())
				continue
			}

		case "get_ledger_entries":
			rpcResponse, handlerErr = HandleGetLedgerEntries(&msg, walletAddress, h.db)
			if handlerErr != nil {
				logger.Error("error handling get_ledger_entries", "error", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to get ledger entries: "+handlerErr.Error())
				continue
			}

		case "get_app_definition":
			rpcResponse, handlerErr = HandleGetAppDefinition(&msg, h.db)
			if handlerErr != nil {
				logger.Error("error handling get_app_definition", "error", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to get app definition: "+handlerErr.Error())
				continue
			}
		case "get_app_sessions":
			rpcResponse, handlerErr = HandleGetAppSessions(&msg, h.db)
			if handlerErr != nil {
				logger.Error("error handling get_app_sessions", "error", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to get app sessions: "+handlerErr.Error())
				continue
			}
		case "get_channels":
			rpcResponse, handlerErr = HandleGetChannels(&msg, h.db)
			if handlerErr != nil {
				logger.Error("error handling get_channels", "error", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to get channels: "+handlerErr.Error())
				continue
			}
		case "create_app_session":
			rpcResponse, handlerErr = HandleCreateApplication(policy, &msg, h.db)
			if handlerErr != nil {
				logger.Warn("error handling create_app_session", "error", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to create application: "+handlerErr.Error())
				continue
			}
			h.sendBalanceUpdate(walletAddress)
			recordHistory = true
		case "submit_state":
			rpcResponse, handlerErr = HandleSubmitState(policy, &msg, h.db)
			if handlerErr != nil {
				logger.Warn("Error handling submit_state", "error", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to update application state: "+handlerErr.Error())
				continue
			}
			h.sendBalanceUpdate(walletAddress)
			recordHistory = true
		case "close_app_session":
			rpcResponse, handlerErr = HandleCloseApplication(policy, &msg, h.db)
			if handlerErr != nil {
				logger.Warn("Error handling close_app_session", "error", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to close application: "+handlerErr.Error())
				continue
			}
			h.sendBalanceUpdate(walletAddress)
			recordHistory = true

		case "resize_channel":
			rpcResponse, handlerErr = HandleResizeChannel(policy, &msg, h.db, h.signer)
			if handlerErr != nil {
				logger.Warn("error handling resize_channel", "error", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to resize channel: "+handlerErr.Error())
				continue
			}
			recordHistory = true
		case "close_channel":
			rpcResponse, handlerErr = HandleCloseChannel(policy, &msg, h.db, h.signer)
			if handlerErr != nil {
				logger.Warn("error handling close_channel", "error", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to close channel: "+handlerErr.Error())
				continue
			}
			recordHistory = true

		case "get_rpc_history":
			rpcResponse, handlerErr = HandleGetRPCHistory(policy, &msg, h.rpcStore)
			if handlerErr != nil {
				logger.Error("error handling get_rpc_history", "error", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to get RPC history: "+handlerErr.Error())
				continue
			}

		default:
			h.sendErrorResponse(walletAddress, &msg, conn, "Unsupported method")
			continue
		}

		// For broker methods, send back a signed RPC response.
		byteData, _ := json.Marshal(rpcResponse.Res)
		signature, _ := h.signer.Sign(byteData)
		rpcResponse.Sig = []string{hexutil.Encode(signature)}
		wsResponseData, _ := json.Marshal(rpcResponse)

		if recordHistory {
			if err := h.rpcStore.StoreMessage(walletAddress, msg.Req, msg.Sig, byteData, rpcResponse.Sig); err != nil {
				logger.Error("failed to store RPC message", "error", err)
				// continue processing even if storage fails
			}
		}

		if err := h.writeWSResponse(conn, wsResponseData); err != nil {
			// Track RPC request failure by method
			h.incrementRPCRequestCount(msg.Req.Method, "failure")
			continue
		}

		// Track RPC request success by method
		h.incrementRPCRequestCount(msg.Req.Method, "success")
	}
}

// forwardMessage forwards an RPC message to all recipients in a virtual app
func forwardMessage(ctx context.Context, rpc *RPCMessage, msg []byte, fromAddress string, h *UnifiedWSHandler) error {
	logger := LoggerFromContext(ctx)

	var data *RPCData
	if rpc.Req != nil {
		data = rpc.Req
	} else {
		data = rpc.Res
	}

	reqBytes, err := json.Marshal(data)
	if err != nil {
		h.incrementRPCRequestCount(data.Method, "failure")
		return errors.New("Error validating signature: " + err.Error())
	}

	recoveredAddresses := map[string]bool{}
	for _, sig := range rpc.Sig {
		addr, err := RecoverAddress(reqBytes, sig)
		if err != nil {
			h.incrementRPCRequestCount(data.Method, "failure")
			return errors.New("invalid signature: " + err.Error())
		}
		recoveredAddresses[addr] = true
	}

	if !recoveredAddresses[fromAddress] {
		return errors.New("unauthorized: invalid signature or sender is not a participant of this vApp")
	}

	var vApp AppSession
	if err := h.db.Where("session_id = ?", rpc.AppSessionID).First(&vApp).Error; err != nil {
		return errors.New("failed to find virtual app session: " + err.Error())
	}

	// Iterate over all recipients in a virtual app and send the message
	for _, recipient := range vApp.ParticipantWallets {
		if recipient == fromAddress {
			continue
		}

		h.connectionsMu.RLock()
		recipientConn, exists := h.connections[recipient]
		h.connectionsMu.RUnlock()
		if exists {
			// Send the message
			if err := h.writeWSResponse(recipientConn, msg); err != nil {
				logger.Error("failed to forward message", "recipient", recipient, "error", err)
				continue
			}

			logger.Debug("successfully forwarded message", "recipient", recipient)
		} else {
			logger.Debug("recipient not connected", "recipient", recipient)
			continue
		}
	}

	h.incrementRPCRequestCount(data.Method, "success")
	return nil
}

// sendErrorResponse creates and sends an error response to the client
func (h *UnifiedWSHandler) sendErrorResponse(sender string, rpc *RPCMessage, conn *websocket.Conn, errMsg string) {
	reqMethod := "unknown"
	reqID := uint64(time.Now().UnixMilli())
	if rpc != nil && rpc.Req != nil {
		reqID = rpc.Req.RequestID
		reqMethod = rpc.Req.Method
	}

	// Track RPC request by method
	defer h.incrementRPCRequestCount(reqMethod, "failure")

	response := CreateResponse(reqID, "error", []any{map[string]any{
		"error": errMsg,
	}})

	byteData, _ := json.Marshal(response.Req)
	signature, _ := h.signer.Sign(byteData)
	response.Sig = []string{hexutil.Encode(signature)}

	responseData, err := json.Marshal(response)
	if err != nil {
		h.logger.Error("failed to marshal error response", "error", err)
		return
	}

	if rpc != nil && rpc.Req != nil {
		if err := h.rpcStore.StoreMessage(sender, rpc.Req, rpc.Sig, byteData, response.Sig); err != nil {
			h.logger.Error("failed to store RPC message", "error", err)
			// continue processing even if storage fails
		}
	}

	// Set a short write deadline to prevent blocking on unresponsive clients
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	// Write the response
	if err := h.writeWSResponse(conn, responseData); err != nil {
		return
	}

	// Reset the write deadline
	conn.SetWriteDeadline(time.Time{})
}

// sendResponse sends a response with a given method and payload to a recipient
func (h *UnifiedWSHandler) sendResponse(recipient string, method string, payload []any, updateType string) {
	logger := h.logger.With("updateType", updateType)

	response := CreateResponse(uint64(time.Now().UnixMilli()), method, payload)

	byteData, _ := json.Marshal(response.Req)
	signature, _ := h.signer.Sign(byteData)
	response.Sig = []string{hexutil.Encode(signature)}

	responseData, err := json.Marshal(response)
	if err != nil {
		logger.Error("error marshaling response", "error", err)
		h.incrementRPCRequestCount(method, "failure")
		return
	}

	h.connectionsMu.RLock()
	recipientConn, exists := h.connections[recipient]
	h.connectionsMu.RUnlock()
	if exists {
		// Write the response
		if err := h.writeWSResponse(recipientConn, responseData); err != nil {
			logger.Error("error writing update", "recipient", recipient, "error", err)
			h.incrementRPCRequestCount(method, "failure")
			return
		}

		logger.Debug("successfully sent update", "recipient", recipient)
	} else {
		logger.Debug("recipient not connected", "recipient", recipient)
		return
	}

	h.incrementRPCRequestCount(method, "success")
}

// incrementRPCRequestCount increments the count of RPC requests for a specific method and status
func (h *UnifiedWSHandler) incrementRPCRequestCount(method string, status string) {
	if method == "" || status == "" {
		return
	}

	h.metrics.RPCRequests.WithLabelValues(method, status).Inc()
}

// sendBalanceUpdate sends balance updates to the client
func (h *UnifiedWSHandler) sendBalanceUpdate(sender string) {
	balances, err := GetWalletLedger(h.db, sender).GetBalances(sender)
	if err != nil {
		h.logger.Error("error getting balances", "sender", sender, "error", err)
		return
	}
	h.sendResponse(sender, "bu", []any{balances}, "balance")
}

// sendChannelsUpdate sends multiple channels updates to the client
func (h *UnifiedWSHandler) sendChannelsUpdate(address string, channels []Channel) {
	resp := []ChannelResponse{}
	for _, ch := range channels {
		resp = append(resp, ChannelResponse{
			ChannelID:   ch.ChannelID,
			Participant: ch.Participant,
			Status:      ch.Status,
			Token:       ch.Token,
			Amount:      big.NewInt(int64(ch.Amount)),
			ChainID:     ch.ChainID,
			Adjudicator: ch.Adjudicator,
			Challenge:   ch.Challenge,
			Nonce:       ch.Nonce,
			Version:     ch.Version,
			CreatedAt:   ch.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   ch.UpdatedAt.Format(time.RFC3339),
		})
	}
	h.sendResponse(address, "channels", []any{resp}, "channels")
}

// sendChannelUpdate sends a single channel update to the client
func (h *UnifiedWSHandler) sendChannelUpdate(channel Channel) {
	channelResponse := ChannelResponse{
		ChannelID:   channel.ChannelID,
		Participant: channel.Participant,
		Status:      channel.Status,
		Token:       channel.Token,
		Amount:      big.NewInt(int64(channel.Amount)),
		ChainID:     channel.ChainID,
		Adjudicator: channel.Adjudicator,
		Challenge:   channel.Challenge,
		Nonce:       channel.Nonce,
		Version:     channel.Version,
		CreatedAt:   channel.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   channel.UpdatedAt.Format(time.RFC3339),
	}
	h.sendResponse(channel.Wallet, "cu", []any{channelResponse}, "channel")
}

// sendAssetsUpdate sends all assets to the client immediately upon connection
func (h *UnifiedWSHandler) sendAssets(conn *websocket.Conn) {
	assets, err := GetAllAssets(h.db, nil) // Get all assets without chain filter
	if err != nil {
		h.logger.Error("error getting assets", "error", err)
		return
	}

	// Convert to AssetResponse format
	response := make([]AssetResponse, 0, len(assets))
	for _, asset := range assets {
		response = append(response, AssetResponse{
			Token:    asset.Token,
			ChainID:  asset.ChainID,
			Symbol:   asset.Symbol,
			Decimals: asset.Decimals,
		})
	}

	// Create RPC response
	rpcResponse := CreateResponse(uint64(time.Now().UnixMilli()), "assets", []any{response})

	sendMessage(conn, h.signer, rpcResponse)
	h.logger.Debug("successfully sent welcome message with assets")
}

// CloseAllConnections closes all open WebSocket connections during shutdown
func (h *UnifiedWSHandler) CloseAllConnections() {
	h.connectionsMu.RLock()
	defer h.connectionsMu.RUnlock()

	for userID, conn := range h.connections {
		h.logger.Debug("closing connection", "userID", userID)
		conn.Close()
	}
}

// AuthResponse represents the server's challenge response
type AuthResponse struct {
	ChallengeMessage uuid.UUID `json:"challenge_message"` // The message to sign
}

// AuthVerifyParams represents parameters for completing authentication
type AuthVerifyParams struct {
	Challenge uuid.UUID `json:"challenge"` // The challenge token
	JWT       string    `json:"jwt"`       // Optional JWT to use for logging in
}

// Allowance represents allowances for connection
type Allowance struct {
	Asset  string `json:"asset"`
	Amount string `json:"amount"`
}

// HandleAuthRequest initializes the authentication process by generating a challenge
func HandleAuthRequest(ctx context.Context, signer *Signer, conn *websocket.Conn, rpc *RPCMessage, authManager *AuthManager) error {
	logger := LoggerFromContext(ctx)

	// Parse the parameters
	if len(rpc.Req.Params) < 7 {
		return errors.New("missing parameters")
	}

	addr, ok := rpc.Req.Params[0].(string)
	if !ok || addr == "" {
		return errors.New("invalid address")
	}

	sessionKey, ok := rpc.Req.Params[1].(string)
	if !ok || sessionKey == "" {
		return errors.New("invalid session key")
	}

	appName, ok := rpc.Req.Params[2].(string)
	if !ok || sessionKey == "" {
		return errors.New("invalid app name")
	}

	rawAllowances := rpc.Req.Params[3]
	allowances, err := parseAllowances(rawAllowances)
	if err != nil {
		return err
	}

	expire, ok := rpc.Req.Params[4].(string)
	if !ok {
		return errors.New("invalid expire")
	}

	scope, ok := rpc.Req.Params[5].(string)
	if !ok {
		return errors.New("invalid scope")
	}

	applicationAddress, ok := rpc.Req.Params[6].(string)
	if !ok {
		return errors.New("invalid application address")
	}

	logger.Debug("incoming auth request",
		"addr", addr,
		"sessionKey", sessionKey,
		"appName", appName,
		"rawAllowances", rawAllowances,
		"scope", scope,
		"expire", expire,
		"applicationAddress", applicationAddress)

	// Generate a challenge for this address
	token, err := authManager.GenerateChallenge(
		addr,
		sessionKey,
		appName,
		allowances,
		scope,
		expire,
		applicationAddress,
	)
	if err != nil {
		return fmt.Errorf("failed to generate challenge: %w", err)
	}

	// Create challenge response
	challengeRes := AuthResponse{
		ChallengeMessage: token,
	}

	// Create RPC response with the challenge
	response := CreateResponse(rpc.Req.RequestID, "auth_challenge", []any{challengeRes})

	// Sign the response with the server's key
	resBytes, _ := json.Marshal(response.Req)
	signature, _ := signer.Sign(resBytes)
	response.Sig = []string{hexutil.Encode(signature)}

	// Send the challenge response
	responseData, _ := json.Marshal(response)
	return conn.WriteMessage(websocket.TextMessage, responseData)
}

// HandleAuthVerify verifies an authentication response to a challenge
// It returns policy, auth method and error
func HandleAuthVerify(ctx context.Context, conn *websocket.Conn, rpc *RPCMessage, authManager *AuthManager, signer *Signer, db *gorm.DB) (*Policy, string, error) {
	logger := LoggerFromContext(ctx)

	authMethod := "unknown"
	if len(rpc.Req.Params) < 1 {
		return nil, authMethod, errors.New("missing parameters")
	}

	var authParams AuthVerifyParams
	paramsJSON, err := json.Marshal(rpc.Req.Params[0])
	if err != nil {
		return nil, authMethod, fmt.Errorf("failed to parse parameters: %w", err)

	}

	if err := json.Unmarshal(paramsJSON, &authParams); err != nil {
		return nil, authMethod, fmt.Errorf("invalid parameters format: %w", err)
	}

	// If JWT was provided - validate and skip all other checks
	if authParams.JWT != "" {
		authMethod = "jwt"

		claims, err := authManager.VerifyJWT(authParams.JWT)
		if err != nil {
			return nil, authMethod, err
		}

		response := CreateResponse(rpc.Req.RequestID, "auth_verify", []any{map[string]any{
			"address":     claims.Policy.Wallet,
			"session_key": claims.Policy.Participant,
			// "jwt_token":   newJwtToken, TODO: add refresh token
			"success": true,
		}})

		if err = sendMessage(conn, signer, response); err != nil {
			logger.Error("failed to send auth success", "error", err)
			return nil, authMethod, err
		}

		return &claims.Policy, authMethod, nil
	}
	authMethod = "signature"

	// Validate the request signature
	if len(rpc.Sig) == 0 {
		return nil, authMethod, errors.New("missing signature in request")
	}

	challenge, err := authManager.GetChallenge(authParams.Challenge)
	if err != nil {
		return nil, authMethod, err
	}
	recoveredAddress, err := RecoverAddressFromEip712Signature(
		challenge.Address,
		challenge.Token.String(),
		challenge.SessionKey,
		challenge.AppName,
		challenge.Allowances,
		challenge.Scope,
		challenge.ApplicationAddress,
		challenge.Expire,
		rpc.Sig[0])
	if err != nil {
		return nil, authMethod, errors.New("invalid signature")
	}

	err = authManager.ValidateChallenge(authParams.Challenge, recoveredAddress)
	if err != nil {
		logger.Debug("challenge verification failed", "error", err)
		return nil, authMethod, err
	}

	// Store signer
	err = AddSigner(db, challenge.Address, challenge.SessionKey)
	if err != nil {
		logger.Error("failed to create signer in db", "error", err)
		return nil, authMethod, err
	}

	claims, jwtToken, err := authManager.GenerateJWT(challenge.Address, challenge.SessionKey, "", "", challenge.Allowances)
	if err != nil {
		logger.Error("failed to generate JWT token", "error", err)
		return nil, authMethod, err
	}

	response := CreateResponse(rpc.Req.RequestID, "auth_verify", []any{map[string]any{
		"address":     challenge.Address,
		"session_key": challenge.SessionKey,
		"jwt_token":   jwtToken,
		"success":     true,
	}})

	if err = sendMessage(conn, signer, response); err != nil {
		logger.Error("error sending auth success", "error", err)
		return nil, authMethod, err
	}

	return &claims.Policy, authMethod, nil
}

func ValidateTimestamp(ts uint64, expirySeconds int) error {
	if ts < 1_000_000_000_000 || ts > 9_999_999_999_999 {
		return fmt.Errorf("invalid timestamp %d: must be 13-digit Unix ms", ts)
	}
	t := time.UnixMilli(int64(ts)).UTC()
	if time.Since(t) > time.Duration(expirySeconds)*time.Second {
		return fmt.Errorf("timestamp expired: %s older than %d s", t.Format(time.RFC3339Nano), expirySeconds)
	}
	return nil
}

func parseAllowances(rawAllowances any) ([]Allowance, error) {
	outerSlice, ok := rawAllowances.([]interface{})
	if !ok {
		return nil, fmt.Errorf("input is not a list of allowances")
	}

	result := make([]Allowance, len(outerSlice))

	for i, item := range outerSlice {
		innerSlice, ok := item.([]interface{})
		if !ok {
			return nil, fmt.Errorf("allowance at index %d is not a list", i)
		}
		if len(innerSlice) != 2 {
			return nil, fmt.Errorf("allowance at index %d must have exactly 2 elements (asset, amount)", i)
		}

		asset, ok1 := innerSlice[0].(string)
		amount, ok2 := innerSlice[1].(string)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("allowance at index %d has non-string asset or amount", i)
		}

		result[i] = Allowance{
			Asset:  asset,
			Amount: amount,
		}
	}

	return result, nil
}

// writeWSResponse writes a response to the WebSocket connection and increments metrics
func (h *UnifiedWSHandler) writeWSResponse(conn *websocket.Conn, responseData []byte) error {
	w, err := conn.NextWriter(websocket.TextMessage)
	if err != nil {
		h.logger.Error("error getting writer for response", "error", err)
		return err
	}

	if _, err := w.Write(responseData); err != nil {
		h.logger.Error("error writing response", "error", err)
		w.Close()
		return err
	}

	if err := w.Close(); err != nil {
		h.logger.Error("error closing writer for response", "error", err)
		return err
	}

	h.metrics.MessageSent.Inc()
	return nil
}

func sendMessage(conn *websocket.Conn, signer *Signer, msg *RPCMessage) error {
	// Sign the response with the server's key
	resBytes, _ := json.Marshal(msg.Req)
	signature, _ := signer.Sign(resBytes)
	msg.Sig = []string{hexutil.Encode(signature)}

	responseData, _ := json.Marshal(msg)
	if err := conn.WriteMessage(websocket.TextMessage, responseData); err != nil {
		return err
	}

	return nil
}
