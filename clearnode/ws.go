package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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
}

func NewUnifiedWSHandler(
	signer *Signer,
	db *gorm.DB,
	metrics *Metrics,
	rpcStore *RPCStore,
	config *Config,
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
	}, nil
}

// HandleConnection handles the WebSocket connection lifecycle.
func (h *UnifiedWSHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// Increment connection metrics
	h.metrics.ConnectionsTotal.Inc()
	h.metrics.ConnectedClients.Inc()
	defer h.metrics.ConnectedClients.Dec()

	var signerAddress string
	var authenticated bool

	// Read messages until authentication completes
	for !authenticated {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			return
		}

		// Increment received message counter
		h.metrics.MessageReceived.Inc()

		var rpcMsg RPCMessage
		if err := json.Unmarshal(message, &rpcMsg); err != nil {
			log.Printf("Invalid message format: %v", err)
			h.sendErrorResponse("", nil, conn, "Invalid message format")
			return
		}

		if err := validate.Struct(&rpcMsg); err != nil {
			log.Printf("Invalid message format: %v", err)
			h.sendErrorResponse("", nil, conn, "Invalid message format")
			return
		}

		// Handle message based on the method
		switch rpcMsg.Req.Method {
		case "auth_request":
			// Track auth request metrics
			h.metrics.AuthRequests.Inc()

			// Client is initiating authentication
			err := HandleAuthRequest(h.signer, conn, &rpcMsg, h.authManager)
			if err != nil {
				log.Printf("Auth initialization failed: %v", err)
				h.sendErrorResponse("", nil, conn, err.Error())
				h.metrics.AuthFailure.Inc()
			}
			continue

		case "auth_verify":
			// Client is responding to a challenge
			authAddr, _, err := HandleAuthVerify(conn, &rpcMsg, h.authManager, h.signer, h.db)
			if err != nil {
				log.Printf("Authentication verification failed: %v", err)
				h.sendErrorResponse("", nil, conn, err.Error())
				h.metrics.AuthFailure.Inc()
				continue
			}

			// Authentication successful
			signerAddress = authAddr
			authenticated = true
			h.metrics.AuthSuccess.Inc()

		default:
			// Reject any other messages before authentication
			log.Printf("Unexpected message method during authentication: %s", rpcMsg.Req.Method)
			h.sendErrorResponse("", nil, conn, "Authentication required. Please send auth_request first.")
		}
	}

	log.Printf("Authentication successful for: %s", signerAddress)

	walletAddress, err := GetWalletBySigner(signerAddress)
	if err != nil {
		log.Printf("Error retrieving wallet address for participant %s: %v", signerAddress, err)
		h.sendErrorResponse("", nil, conn, "Failed to retrieve wallet address")
		return
	}
	if walletAddress == "" {
		walletAddress = signerAddress
	}

	// Store connection for authenticated user
	h.connectionsMu.Lock()
	// Currently, only one connection per wallet is allowed
	h.connections[walletAddress] = conn
	h.connectionsMu.Unlock()

	defer func() {
		h.connectionsMu.Lock()
		delete(h.connections, walletAddress)
		h.connectionsMu.Unlock()
		log.Printf("Connection closed for wallet: %s", walletAddress)
		// TODO: Remove signer from DB and cache
	}()

	log.Printf("Participant authenticated: %s", walletAddress)

	// Send initial balance and channels information in form of balance and channel updates
	channels, err := getChannelsByWallet(h.db, walletAddress, string(ChannelStatusOpen))
	if err != nil {
		log.Printf("Error retrieving channels for participant %s: %v", walletAddress, err)
	}

	h.sendChannelsUpdate(walletAddress, channels)
	h.sendBalanceUpdate(walletAddress)

	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket unexpected close error: %v", err)
			} else {
				log.Printf("Error reading message: %v", err)
			}
			break
		}

		// Increment received message counter
		h.metrics.MessageReceived.Inc()

		// Check if session is still valid
		if !h.authManager.ValidateSession(signerAddress) {
			log.Printf("Session expired for wallet: %s", signerAddress)
			h.sendErrorResponse(signerAddress, nil, conn, "Session expired. Please re-authenticate.")
			break
		}

		// Update session activity timestamp
		h.authManager.UpdateSession(signerAddress)

		// Forward request or response for internal vApp communication.
		var msg RPCMessage
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			h.sendErrorResponse(walletAddress, nil, conn, "Invalid message format")
			continue
		}

		if err := validate.Struct(&msg); err != nil {
			log.Printf("Invalid message format: %v", err)
			h.sendErrorResponse(walletAddress, nil, conn, "Invalid message format")
			return
		}

		if msg.AppSessionID != "" {
			if err := forwardMessage(&msg, messageBytes, walletAddress, h); err != nil {
				log.Printf("Error forwarding message: %v", err)
				h.sendErrorResponse(walletAddress, nil, conn, "Failed to forward message: "+err.Error())
				continue
			}
			continue
		}

		if msg.Req == nil {
			continue
		}

		if err = ValidateTimestamp(msg.Req.Timestamp, h.config.msgExpiryTime); err != nil {
			log.Printf("Message timestamp validation failed: %v", err)
			h.sendErrorResponse(walletAddress, &msg, conn, fmt.Sprintf("Message timestamp validation failed: %v", err))
			continue
		}

		var rpcResponse = &RPCMessage{}
		var handlerErr error
		var recordHistory = false

		// Track RPC request by method
		h.metrics.RPCRequests.WithLabelValues(msg.Req.Method).Inc()

		switch msg.Req.Method {
		case "ping":
			rpcResponse, handlerErr = HandlePing(&msg)
			if handlerErr != nil {
				log.Printf("Error handling ping: %v", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to process ping: "+handlerErr.Error())
				continue
			}

		case "get_config":
			rpcResponse, handlerErr = HandleGetConfig(&msg, h.config, h.signer)
			if handlerErr != nil {
				log.Printf("Error handling get_config: %v", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to get config: "+handlerErr.Error())
				continue
			}

		case "get_assets":
			rpcResponse, handlerErr = HandleGetAssets(&msg, h.db)
			if handlerErr != nil {
				log.Printf("Error handling get_assets: %v", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to get assets: "+handlerErr.Error())
				continue
			}

		case "get_ledger_balances":
			rpcResponse, handlerErr = HandleGetLedgerBalances(&msg, walletAddress, h.db)
			if handlerErr != nil {
				log.Printf("Error handling get_ledger_balances: %v", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to get ledger balances: "+handlerErr.Error())
				continue
			}

		case "get_ledger_entries":
			rpcResponse, handlerErr = HandleGetLedgerEntries(&msg, walletAddress, h.db)
			if handlerErr != nil {
				log.Printf("Error handling get_ledger_entries: %v", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to get ledger entries: "+handlerErr.Error())
				continue
			}

		case "get_app_definition":
			rpcResponse, handlerErr = HandleGetAppDefinition(&msg, h.db)
			if handlerErr != nil {
				log.Printf("Error handling get_app_definition: %v", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to get app definition: "+handlerErr.Error())
				continue
			}

		case "create_app_session":
			rpcResponse, handlerErr = HandleCreateApplication(&msg, h.db)
			if handlerErr != nil {
				log.Printf("Error handling create_app_session: %v", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to create application: "+handlerErr.Error())
				continue
			}
			h.sendBalanceUpdate(walletAddress)
			recordHistory = true
		case "close_app_session":
			rpcResponse, handlerErr = HandleCloseApplication(&msg, h.db)
			if handlerErr != nil {
				log.Printf("Error handling close_app_session: %v", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to close application: "+handlerErr.Error())
				continue
			}
			h.sendBalanceUpdate(walletAddress)
			recordHistory = true
		case "get_app_sessions":
			rpcResponse, handlerErr = HandleGetAppSessions(&msg, h.db)
			if handlerErr != nil {
				log.Printf("Error handling get_app_sessions: %v", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to get app sessions: "+handlerErr.Error())
				continue
			}

		case "resize_channel":
			rpcResponse, handlerErr = HandleResizeChannel(&msg, h.db, h.signer)
			if handlerErr != nil {
				log.Printf("Error handling resize_channel: %v", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to resize channel: "+handlerErr.Error())
				continue
			}
			recordHistory = true
		case "close_channel":
			rpcResponse, handlerErr = HandleCloseChannel(&msg, h.db, h.signer)
			if handlerErr != nil {
				log.Printf("Error handling close_channel: %v", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to close channel: "+handlerErr.Error())
				continue
			}
			recordHistory = true
		case "get_channels":
			rpcResponse, handlerErr = HandleGetChannels(&msg, h.db)
			if handlerErr != nil {
				log.Printf("Error handling get_channels: %v", handlerErr)
				h.sendErrorResponse(walletAddress, &msg, conn, "Failed to get channels: "+handlerErr.Error())
				continue
			}

		case "get_rpc_history":
			rpcResponse, handlerErr = HandleGetRPCHistory(walletAddress, &msg, h.rpcStore)
			if handlerErr != nil {
				log.Printf("Error handling get_rpc_history: %v", handlerErr)
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
				log.Printf("Failed to store RPC message: %v", err)
				// continue processing even if storage fails
			}
		}

		// Use NextWriter for safer message delivery
		w, err := conn.NextWriter(websocket.TextMessage)
		if err != nil {
			log.Printf("Error getting writer for response: %v", err)
			continue
		}

		if _, err := w.Write(wsResponseData); err != nil {
			log.Printf("Error writing response: %v", err)
			w.Close()
			continue
		}

		if err := w.Close(); err != nil {
			log.Printf("Error closing writer for response: %v", err)
			continue
		}

		// Increment sent message counter
		h.metrics.MessageSent.Inc()
	}
}

// forwardMessage forwards an RPC message to all recipients in a virtual app
func forwardMessage(rpc *RPCMessage, msg []byte, fromAddress string, h *UnifiedWSHandler) error {
	var data *RPCData
	if rpc.Req != nil {
		data = rpc.Req
	} else {
		data = rpc.Res
	}

	reqBytes, err := json.Marshal(data)
	if err != nil {
		return errors.New("Error validating signature: " + err.Error())
	}

	recoveredAddresses := map[string]bool{}
	for _, sig := range rpc.Sig {
		addr, err := RecoverAddress(reqBytes, sig)
		if err != nil {
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
			// Use NextWriter for safer message delivery
			w, err := recipientConn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("Error getting writer for forwarded message to %s: %v", recipient, err)
				continue
			}

			if _, err := w.Write(msg); err != nil {
				log.Printf("Error writing forwarded message to %s: %v", recipient, err)
				w.Close()
				continue
			}

			if err := w.Close(); err != nil {
				log.Printf("Error closing writer for forwarded message to %s: %v", recipient, err)
				continue
			}

			// Increment sent message counter for each forwarded message
			h.metrics.MessageSent.Inc()

			log.Printf("Successfully forwarded message to %s", recipient)
		} else {
			log.Printf("Recipient %s not connected", recipient)
			continue
		}
	}

	return nil
}

// sendErrorResponse creates and sends an error response to the client
func (h *UnifiedWSHandler) sendErrorResponse(sender string, rpc *RPCMessage, conn *websocket.Conn, errMsg string) {
	reqID := uint64(time.Now().UnixMilli())
	if rpc != nil && rpc.Req != nil {
		reqID = rpc.Req.RequestID
	}
	response := CreateResponse(reqID, "error", []any{map[string]any{
		"error": errMsg,
	}}, time.Now())

	byteData, _ := json.Marshal(response.Req)
	signature, _ := h.signer.Sign(byteData)
	response.Sig = []string{hexutil.Encode(signature)}

	responseData, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling error response: %v", err)
		return
	}

	if rpc != nil && rpc.Req != nil {
		if err := h.rpcStore.StoreMessage(sender, rpc.Req, rpc.Sig, byteData, response.Sig); err != nil {
			log.Printf("Failed to store RPC message: %v", err)
			// continue processing even if storage fails
		}
	}

	// Set a short write deadline to prevent blocking on unresponsive clients
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	// Use NextWriter for safer message delivery
	w, err := conn.NextWriter(websocket.TextMessage)
	if err != nil {
		log.Printf("Error getting writer for error response: %v", err)
		return
	}

	if _, err := w.Write(responseData); err != nil {
		log.Printf("Error writing error response: %v", err)
		w.Close()
		return
	}

	if err := w.Close(); err != nil {
		log.Printf("Error closing writer for error response: %v", err)
	}

	// Increment sent message counter
	h.metrics.MessageSent.Inc()

	// Reset the write deadline
	conn.SetWriteDeadline(time.Time{})
}

// sendResponse sends a response with a given method and payload to a recipient
func (h *UnifiedWSHandler) sendResponse(recipient string, method string, payload []any, updateType string) {
	response := CreateResponse(uint64(time.Now().UnixMilli()), method, payload, time.Now())

	byteData, _ := json.Marshal(response.Req)
	signature, _ := h.signer.Sign(byteData)
	response.Sig = []string{hexutil.Encode(signature)}

	responseData, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling %s response: %v", updateType, err)
		return
	}

	h.connectionsMu.RLock()
	recipientConn, exists := h.connections[recipient]
	h.connectionsMu.RUnlock()
	if exists {
		// Use NextWriter for safer message delivery
		w, err := recipientConn.NextWriter(websocket.TextMessage)
		if err != nil {
			log.Printf("Error getting writer for %s update to %s: %v", updateType, recipient, err)
			return
		}

		if _, err := w.Write(responseData); err != nil {
			log.Printf("Error writing %s update to %s: %v", updateType, recipient, err)
			w.Close()
			return
		}

		if err := w.Close(); err != nil {
			log.Printf("Error closing writer for %s update to %s: %v", updateType, recipient, err)
			return
		}

		// Increment sent message counter
		h.metrics.MessageSent.Inc()

		log.Printf("Successfully sent %s update to %s", updateType, recipient)
	} else {
		log.Printf("Recipient %s not connected", recipient)
		return
	}
}

// sendBalanceUpdate sends balance updates to the client
func (h *UnifiedWSHandler) sendBalanceUpdate(sender string) {
	balances, err := GetWalletLedger(h.db, sender).GetBalances(sender)
	if err != nil {
		log.Printf("Error getting balances for %s: %v", sender, err)
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

// CloseAllConnections closes all open WebSocket connections during shutdown
func (h *UnifiedWSHandler) CloseAllConnections() {
	h.connectionsMu.RLock()
	defer h.connectionsMu.RUnlock()

	for userID, conn := range h.connections {
		log.Printf("Closing connection for participant: %s", userID)
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
func HandleAuthRequest(signer *Signer, conn *websocket.Conn, rpc *RPCMessage, authManager *AuthManager) error {
	// Parse the parameters
	if len(rpc.Req.Params) < 4 {
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
	// Generate a challenge for this address
	token, err := authManager.GenerateChallenge(addr, sessionKey, appName, allowances)
	if err != nil {
		return fmt.Errorf("failed to generate challenge: %w", err)
	}

	// Create challenge response
	challengeRes := AuthResponse{
		ChallengeMessage: token,
	}

	// Create RPC response with the challenge
	response := CreateResponse(rpc.Req.RequestID, "auth_challenge", []any{challengeRes}, time.Now())

	// Sign the response with the server's key
	resBytes, _ := json.Marshal(response.Req)
	signature, _ := signer.Sign(resBytes)
	response.Sig = []string{hexutil.Encode(signature)}

	// Send the challenge response
	responseData, _ := json.Marshal(response)
	return conn.WriteMessage(websocket.TextMessage, responseData)
}

// HandleAuthVerify verifies an authentication response to a challenge
func HandleAuthVerify(conn *websocket.Conn, rpc *RPCMessage, authManager *AuthManager, signer *Signer, db *gorm.DB) (string, string, error) {
	if len(rpc.Req.Params) < 1 {
		return "", "", errors.New("missing parameters")
	}

	var authParams AuthVerifyParams
	paramsJSON, err := json.Marshal(rpc.Req.Params[0])
	if err != nil {
		return "", "", fmt.Errorf("failed to parse parameters: %w", err)

	}

	if err := json.Unmarshal(paramsJSON, &authParams); err != nil {
		return "", "", fmt.Errorf("invalid parameters format: %w", err)
	}

	// If JWT was provided - validate and skip all other checks
	if authParams.JWT != "" {
		claims, err := authManager.VerifyJWT(authParams.JWT)
		if err != nil {
			return "", "", err
		}

		response := CreateResponse(rpc.Req.RequestID, "auth_verify", []any{map[string]any{
			"address":     claims.Policy.Wallet,
			"session_key": claims.Policy.Participant,
			// "jwt_token":   newJwtToken, TODO: add refresh token
			"success": true,
		}}, time.Now())

		if err = sendMessage(conn, signer, response); err != nil {
			log.Printf("Error sending auth success: %v", err)
			return "", "", err
		}

		return claims.Policy.Wallet, claims.Policy.Participant, nil
	}

	// Validate the request signature
	if len(rpc.Sig) == 0 {
		return "", "", errors.New("missing signature in request")
	}

	challenge, err := authManager.GetChallenge(authParams.Challenge)
	if err != nil {
		return "", "", err
	}
	recoveredAddress, err := RecoverAddressFromEip712Signature(challenge.Address, challenge.Token.String(), challenge.SessionKey, challenge.AppName, challenge.Allowances, rpc.Sig[0])
	if err != nil {
		return "", "", errors.New("invalid signature")
	}

	err = authManager.ValidateChallenge(authParams.Challenge, recoveredAddress)
	if err != nil {
		log.Printf("Challenge verification failed: %v", err)
		return "", "", err
	}

	// Store signer
	err = AddSigner(db, challenge.Address, challenge.SessionKey)
	if err != nil {
		log.Printf("Failed to create signer in db: %v", err)
		return "", "", err
	}

	jwtToken, err := authManager.GenerateJWT(challenge.Address, challenge.SessionKey, "", "", challenge.Allowances)
	if err != nil {
		log.Printf("Failed to generate JWT token: %v", err)
		return "", "", err
	}

	response := CreateResponse(rpc.Req.RequestID, "auth_verify", []any{map[string]any{
		"address":     challenge.Address,
		"session_key": challenge.SessionKey,
		"jwt_token":   jwtToken,
		"success":     true,
	}}, time.Now())

	if err = sendMessage(conn, signer, response); err != nil {
		log.Printf("Error sending auth success: %v", err)
		return "", "", err
	}

	return challenge.Address, challenge.SessionKey, nil
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
