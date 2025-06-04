package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/erc7824/nitrolite/clearnode/nitrolite"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

// Signer handles signing operations using a private key
type Signer struct {
	privateKey *ecdsa.PrivateKey
}

// RPCMessage represents a complete message in the RPC protocol, including data and signatures
type RPCMessage struct {
	Req          *RPCData `json:"req,omitempty"`
	Sig          []string `json:"sig"`
	AppSessionID string   `json:"sid,omitempty"`
}

// RPCData represents the common structure for both requests and responses
// Format: [request_id, method, params, ts]
type RPCData struct {
	RequestID uint64 `json:"id"`
	Method    string `json:"method"`
	Params    []any  `json:"params"`
	Timestamp uint64 `json:"ts"`
}

// MarshalJSON implements the json.Marshaler interface for RPCData
func (m RPCData) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{
		m.RequestID,
		m.Method,
		m.Params,
		m.Timestamp,
	})
}

// AppDefinition represents the definition of an application on the ledger
type AppDefinition struct {
	Protocol           string   `json:"protocol"`
	ParticipantWallets []string `json:"participants"`
	Weights            []uint64 `json:"weights"`
	Quorum             uint64   `json:"quorum"`
	Challenge          uint64   `json:"challenge"`
	Nonce              uint64   `json:"nonce"`
}

// CreateAppSessionParams represents parameters needed for virtual app creation
type CreateAppSessionParams struct {
	Definition  AppDefinition   `json:"definition"`
	Allocations []AppAllocation `json:"allocations"`
}

type AppAllocation struct {
	ParticipantWallet string          `json:"participant"`
	AssetSymbol       string          `json:"asset"`
	Amount            decimal.Decimal `json:"amount"`
}

type CreateAppSignData struct {
	RequestID uint64
	Method    string
	Params    []CreateAppSessionParams
	Timestamp uint64
}

func (r CreateAppSignData) MarshalJSON() ([]byte, error) {
	arr := []interface{}{r.RequestID, r.Method, r.Params, r.Timestamp}
	return json.Marshal(arr)
}

// CloseAppSessionParams represents parameters needed for virtual app closure
type CloseAppSessionParams struct {
	AppSessionID string          `json:"app_session_id"`
	Allocations  []AppAllocation `json:"allocations"`
}

type CloseAppSignData struct {
	RequestID uint64
	Method    string
	Params    []CloseAppSessionParams
	Timestamp uint64
}

func (r CloseAppSignData) MarshalJSON() ([]byte, error) {
	arr := []interface{}{r.RequestID, r.Method, r.Params, r.Timestamp}
	return json.Marshal(arr)
}

// ResizeChannelParams represents parameters needed for resizing a channel
type ResizeChannelParams struct {
	ChannelID        string   `json:"channel_id"`
	AllocateAmount   *big.Int `json:"allocate_amount,omitempty"`
	ResizeAmount     *big.Int `json:"resize_amount,omitempty"`
	FundsDestination string   `json:"funds_destination"`
}

type ResizeChannelSignData struct {
	RequestID uint64
	Method    string
	Params    []ResizeChannelParams
	Timestamp uint64
}

func (r ResizeChannelSignData) MarshalJSON() ([]byte, error) {
	arr := []interface{}{r.RequestID, r.Method, r.Params, r.Timestamp}
	return json.Marshal(arr)
}

// NewSigner creates a new signer from a hex-encoded private key
func NewSigner(privateKeyHex string) (*Signer, error) {
	if len(privateKeyHex) >= 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, err
	}

	return &Signer{privateKey: privateKey}, nil
}

// Sign creates an ECDSA signature for the provided data
func (s *Signer) Sign(data []byte) ([]byte, error) {
	sig, err := nitrolite.Sign(data, s.privateKey)
	if err != nil {
		return nil, err
	}

	signature := make([]byte, 65)
	copy(signature[0:32], sig.R[:])
	copy(signature[32:64], sig.S[:])

	if sig.V >= 27 {
		signature[64] = sig.V - 27
	}
	return signature, nil
}

// GetAddress returns the address derived from the signer's public key
func (s *Signer) GetAddress() string {
	publicKey := s.privateKey.Public().(*ecdsa.PublicKey)
	return crypto.PubkeyToAddress(*publicKey).Hex()
}

// generatePrivateKey generates a new private key
func generatePrivateKey() (*ecdsa.PrivateKey, error) {
	return crypto.GenerateKey()
}

// savePrivateKey saves a private key to file
func savePrivateKey(key *ecdsa.PrivateKey, filePath string) error {
	keyBytes := crypto.FromECDSA(key)
	keyHex := hexutil.Encode(keyBytes)
	if len(keyHex) >= 2 && keyHex[:2] == "0x" {
		keyHex = keyHex[2:]
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	return os.WriteFile(filePath, []byte(keyHex), 0600)
}

// loadPrivateKey loads a private key from file
func loadPrivateKey(filePath string) (*ecdsa.PrivateKey, error) {
	keyHex, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return crypto.HexToECDSA(string(keyHex))
}

// Client handles websocket connection and RPC messaging
type Client struct {
	conn          *websocket.Conn
	signers       []*Signer
	address       string // Primary address (for backward compatibility)
	addresses     []string
	authSigner    *Signer // Signer used for authentication
	noSignatures  bool    // Flag to indicate if signatures should be added
	noAuth        bool    // Flag to indicate if authentication should be skipped
	jwt           string  // JWT token received after authentication
	serverURL     string  // Server URL for reconnection
	nextRequestID uint64  // Counter for request IDs
}

// NewClient creates a new websocket client
func NewClient(serverURL string, authSigner *Signer, noSignatures bool, noAuth bool, signers ...*Signer) (*Client, error) {
	if len(signers) == 0 && !noSignatures {
		return nil, fmt.Errorf("at least one signer is required unless noSignatures is enabled")
	}

	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	var primaryAddress string
	var addresses []string

	if len(signers) > 0 {
		// Set auth signer if not specified and auth is required
		if authSigner == nil && !noAuth {
			authSigner = signers[0]
		}

		// We'll use the auth signer's address as the primary address for auth
		if authSigner != nil {
			primaryAddress = authSigner.GetAddress()
		}

		// Collect all addresses
		addresses = make([]string, len(signers))
		for i, signer := range signers {
			addresses[i] = signer.GetAddress()
		}
	} else if authSigner != nil {
		// If we have no signers but have auth signer, use its address
		primaryAddress = authSigner.GetAddress()
		addresses = []string{primaryAddress}
	}

	return &Client{
		conn:          conn,
		signers:       signers,
		address:       primaryAddress,
		addresses:     addresses,
		authSigner:    authSigner,
		noSignatures:  noSignatures,
		noAuth:        noAuth,
		serverURL:     serverURL,
		nextRequestID: 1,
	}, nil
}

// SendMessage sends an RPC message to the server
func (c *Client) SendMessage(rpcMsg RPCMessage) error {
	// If we have a JWT token and it's not already set, add it to the message
	if c.jwt != "" && rpcMsg.AppSessionID == "" {
		rpcMsg.AppSessionID = c.jwt
	}

	// Marshal the message to JSON
	data, err := json.Marshal(rpcMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Send the message
	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// collectSignatures gathers signatures from all signers for the given data
func (c *Client) collectSignatures(data []byte) ([]string, error) {
	if c.noSignatures {
		return []string{}, nil
	}

	signatures := make([]string, len(c.signers))

	for i, signer := range c.signers {
		signature, err := signer.Sign(data)
		if err != nil {
			return nil, fmt.Errorf("failed to sign with signer %d: %w", i, err)
		}
		signatures[i] = hexutil.Encode(signature)
	}

	return signatures, nil
}

// Authenticate performs the authentication flow with the server
func (c *Client) Authenticate() error {
	if c.noAuth {
		fmt.Println("Authentication skipped (noAuth mode)")
		return nil
	}

	fmt.Println("Starting authentication...")

	if c.authSigner == nil {
		return fmt.Errorf("no authentication signer provided")
	}

	// Step 1: Auth request - Request a challenge
	authReq := RPCMessage{
		Req: &RPCData{
			RequestID: c.nextRequestID,
			Method:    "auth_request",
			Params:    []any{c.address, c.addresses[0], "test-app", []interface{}{[]interface{}{"usdc", "10000"}}, "3600", "all", "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"},
			Timestamp: uint64(time.Now().UnixMilli()),
		},
		Sig: []string{},
	}
	c.nextRequestID++

	// Sign the request with auth signer
	reqData, err := json.Marshal(authReq.Req)
	if err != nil {
		return fmt.Errorf("failed to marshal auth request: %w", err)
	}

	// For authentication, we always need a signature regardless of noSignatures setting
	signature, err := c.authSigner.Sign(reqData)
	if err != nil {
		return fmt.Errorf("failed to sign auth request: %w", err)
	}
	authReq.Sig = []string{hexutil.Encode(signature)}

	if err := c.SendMessage(authReq); err != nil {
		return fmt.Errorf("failed to send auth request: %w", err)
	}

	// Step 2: Wait for challenge, skipping non-auth related messages
	fmt.Println("Waiting for challenge...")
	var challengeStr string

	// Set a deadline to avoid hanging if no challenge is received
	challengeDeadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(challengeDeadline) {
		// Read a message from the server
		c.conn.SetReadDeadline(challengeDeadline)
		_, challengeMsg, err := c.conn.ReadMessage()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return fmt.Errorf("timed out waiting for challenge")
			}
			return fmt.Errorf("failed to read challenge response: %w", err)
		}

		// Parse the message
		var challengeResp map[string]any
		if err := json.Unmarshal(challengeMsg, &challengeResp); err != nil {
			return fmt.Errorf("failed to parse challenge response: %w", err)
		}

		// Check if we have a response array
		if resArray, ok := challengeResp["res"].([]any); ok {
			// Check if this is a non-auth message
			if len(resArray) > 1 {
				if method, ok := resArray[1].(string); ok {
					// Skip non-auth related messages like assets, etc.
					if method == "assets" {
						fmt.Printf("Skipping non-auth message of type: %s\n", method)
						continue
					}
					fmt.Println(challengeResp)
					// If it's auth_challenge, process it
					if method == "auth_challenge" {
						fmt.Println("Received auth challenge")
						if paramsArray, ok := resArray[2].([]any); ok && len(paramsArray) >= 1 {
							if challengeObj, ok := paramsArray[0].(map[string]any); ok {
								if msg, ok := challengeObj["challenge_message"].(string); ok {
									challengeStr = msg
								}

								if challengeStr != "" {
									break
								}
							}
						}
					}
				}
			}
		}

		// Check alternative locations
		if challengeStr == "" {
			// Look for challenge directly in the response
			if params, ok := challengeResp["params"].([]any); ok && len(params) > 0 {
				if challengeObj, ok := params[0].(map[string]any); ok {
					if msg, ok := challengeObj["challenge_message"].(string); ok {
						challengeStr = msg
					}

					if challengeStr != "" {
						break
					}
				}
			}
		}
	}

	// Reset read deadline
	c.conn.SetReadDeadline(time.Time{})

	// If we didn't find a challenge, check if we need one
	if challengeStr == "" {
		fmt.Println("No auth challenge received. Server may not require auth.")
		fmt.Println("Skipping auth challenge/verify steps.")
		return nil // Skip the rest of the auth flow
	}

	fmt.Printf("Found challenge message: %s\n", challengeStr)

	// Step 3: Send auth verify
	fmt.Println("Sending challenge verification...")
	verifyReq := RPCMessage{
		Req: &RPCData{
			RequestID: c.nextRequestID,
			Method:    "auth_verify",
			Params: []any{map[string]any{
				"challenge": challengeStr,
			}},
			Timestamp: uint64(time.Now().UnixMilli()),
		},
		Sig: []string{},
	}
	c.nextRequestID++

	privKey := c.authSigner.privateKey
	convertedAllowances := convertAllowances([]Allowance{{Asset: "usdc", Amount: "10000"}})

	// Build the EIP-712 TypedData
	td := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {{Name: "name", Type: "string"}},
			"Policy": {
				{Name: "challenge", Type: "string"},
				{Name: "scope", Type: "string"},
				{Name: "wallet", Type: "address"},
				{Name: "application", Type: "address"},
				{Name: "participant", Type: "address"},
				{Name: "expire", Type: "uint256"},
				{Name: "allowances", Type: "Allowance[]"},
			},
			"Allowance": {
				{Name: "asset", Type: "string"},
				{Name: "amount", Type: "uint256"},
			},
		},
		PrimaryType: "Policy",
		Domain:      apitypes.TypedDataDomain{Name: "test-app"},
		Message: map[string]interface{}{
			"challenge":   challengeStr,
			"scope":       "all",
			"wallet":      c.address,
			"application": "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
			"participant": c.addresses[0],
			"expire":      "3600",
			"allowances":  convertedAllowances,
		},
	}

	// Hash according to EIP-712
	hash, _, err := apitypes.TypedDataAndHash(td)
	if err != nil {
		return err
	}

	// Sign the hash
	sigBytes, err := crypto.Sign(hash, privKey)
	if err != nil {
		return err
	}

	verifyReq.Sig = []string{hexutil.Encode(sigBytes)}

	// Send verify request
	if err := c.SendMessage(verifyReq); err != nil {
		return fmt.Errorf("failed to send verify request: %w", err)
	}

	// Wait for auth verify response, skipping non-auth related messages
	fmt.Println("Waiting for verification response...")
	verifyDeadline := time.Now().Add(5 * time.Second)
	var success bool
	var foundVerifyResponse bool

	for time.Now().Before(verifyDeadline) {
		// Read a message from the server
		c.conn.SetReadDeadline(verifyDeadline)
		_, verifyMsg, err := c.conn.ReadMessage()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return fmt.Errorf("timed out waiting for verification response")
			}
			return fmt.Errorf("failed to read verify response: %w", err)
		}

		// Parse the message
		var verifyResp map[string]any
		if err := json.Unmarshal(verifyMsg, &verifyResp); err != nil {
			return fmt.Errorf("failed to parse verify response: %w", err)
		}

		// Check if we have a response array
		resVerifyArray, ok := verifyResp["res"].([]any)
		if !ok || len(resVerifyArray) < 3 {
			fmt.Println("Skipping non-auth message (invalid format)")
			continue
		}

		// Check if this is a non-auth message like assets or ping
		if len(resVerifyArray) > 1 {
			if method, ok := resVerifyArray[1].(string); ok {
				// Skip non-auth related messages
				if method == "assets" || method == "error" || method == "pong" {
					fmt.Printf("Skipping non-auth message of type: %s\n", method)
					continue
				}

				// If it's auth_verify, process it
				if method == "auth_verify" {
					foundVerifyResponse = true
				}
			}
		}

		// Extract verification parameters
		verifyParamsArray, ok := resVerifyArray[2].([]any)
		if !ok || len(verifyParamsArray) < 1 {
			fmt.Println("Skipping message with invalid parameters")
			continue
		}

		verifyObj, ok := verifyParamsArray[0].(map[string]any)
		if !ok {
			fmt.Println("Skipping message with invalid verification object")
			continue
		}

		// Check if auth was successful
		if successValue, ok := verifyObj["success"].(bool); ok {
			success = successValue
			foundVerifyResponse = true

			// Extract JWT token if available
			if token, ok := verifyObj["token"].(string); ok {
				c.jwt = token
				fmt.Println("JWT token received!")
			}

			// If we found the verify response, break out of the loop
			break
		}
	}

	// Reset read deadline
	c.conn.SetReadDeadline(time.Time{})

	// Check if we found a verification response
	if !foundVerifyResponse {
		fmt.Println("No verification response received. Server may not require auth.")
		fmt.Println("Proceeding anyway...")
		return nil
	}

	// Check if authentication was successful
	if !success {
		return fmt.Errorf("authentication failed")
	}

	fmt.Println("Authentication successful!")
	return nil
}

// Close closes the websocket connection
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// Allowance represents an asset allowance for authentication
type Allowance struct {
	Asset  string `json:"asset"`
	Amount string `json:"amount"`
}

// convertAllowances converts allowances to the format needed for EIP-712 signing
func convertAllowances(input []Allowance) []map[string]interface{} {
	out := make([]map[string]interface{}, len(input))
	for i, a := range input {
		amountInt, ok := new(big.Int).SetString(a.Amount, 10)
		if !ok {
			log.Printf("Invalid amount in allowance: %s", a.Amount)
			continue
		}
		out[i] = map[string]interface{}{
			"asset":  a.Asset,
			"amount": amountInt,
		}
	}
	return out
}

func main() {
	// Define flags
	var (
		methodFlag  = flag.String("method", "", "RPC method name")
		idFlag      = flag.Uint64("id", 1, "Request ID")
		paramsFlag  = flag.String("params", "[]", "JSON array of parameters")
		sendFlag    = flag.Bool("send", false, "Send the message to the server")
		serverFlag  = flag.String("server", "ws://localhost:8000/ws", "WebSocket server URL (can also be set via SERVER environment variable)")
		genKeyFlag  = flag.String("genkey", "", "Generate a new key and exit. Use a signer number (e.g., '1', '2', '3').")
		signersFlag = flag.String("signers", "", "Comma-separated list of signer numbers to use (e.g., '1,2,3'). If not specified, all available signers will be used.")
		authFlag    = flag.String("auth", "", "Specify which signer to authenticate with (e.g., '1'). If not specified, first signer is used.")
		noSignFlag  = flag.Bool("nosign", false, "Send request without signatures")
		noAuthFlag  = flag.Bool("noauth", false, "Skip authentication (for public endpoints)")
	)

	flag.Parse()

	// Check if SERVER environment variable is set
	if serverEnv := os.Getenv("SERVER"); serverEnv != "" {
		*serverFlag = serverEnv
	}

	// Get current directory for key storage
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current directory: %v", err)
	}

	// If genkey flag is set, generate a key and exit
	if *genKeyFlag != "" {
		generateKey(*genKeyFlag, currentDir)
		os.Exit(0)
	}

	// For normal operation, method is required
	if *methodFlag == "" {
		fmt.Println("Error: method is required")
		flag.Usage()
		os.Exit(1)
	}

	// Parse params
	var params []any
	if err := json.Unmarshal([]byte(*paramsFlag), &params); err != nil {
		log.Fatalf("Error parsing params JSON: %v", err)
	}

	// Find and load signers
	allSigners, signerMapping := findSigners(currentDir)

	if len(allSigners) == 0 {
		log.Fatalf("No signers found. Generate at least one key with --genkey.")
	}

	// Determine which signers to use
	signers := selectSigners(allSigners, signerMapping, *signersFlag)

	// Get auth signer
	authSigner := getAuthSigner(signers, signerMapping, *authFlag, *sendFlag)

	// Create RPC data and prepare message
	rpcMessage, signatures := prepareRPCMessage(*methodFlag, *idFlag, params, signers, *noSignFlag)

	// Display message info
	printMessageInfo(rpcMessage, *sendFlag, params, signatures, signers, authSigner, *noSignFlag, *noAuthFlag, *serverFlag)

	// If send flag is set, send the message to the server
	if *sendFlag {
		// Create the client and send the message
		client, err := NewClient(*serverFlag, authSigner, *noSignFlag, *noAuthFlag, signers...)
		if err != nil {
			log.Fatalf("Error creating client: %v", err)
		}
		defer client.Close()

		// Authenticate with the server
		if err := client.Authenticate(); err != nil {
			log.Fatalf("Authentication failed: %v", err)
		}

		// Send the message
		if err := client.SendMessage(rpcMessage); err != nil {
			log.Fatalf("Error sending message: %v", err)
		}

		// Read and display responses
		readResponses(client)
	}
}

// generateKey creates a new key and displays its information
func generateKey(genKeyFlag string, currentDir string) {
	var signerNum int
	if _, err := fmt.Sscanf(genKeyFlag, "%d", &signerNum); err != nil {
		log.Fatalf("Invalid genkey value. Use a signer number (e.g., '1', '2', '3'): %v", err)
	}

	if signerNum < 1 {
		log.Fatalf("Signer number must be at least 1")
	}

	keyPath := filepath.Join(currentDir, fmt.Sprintf("signer_key_%d.hex", signerNum))
	keyType := fmt.Sprintf("signer #%d", signerNum)

	// Generate new key
	key, err := generatePrivateKey()
	if err != nil {
		log.Fatalf("Error generating private key: %v", err)
	}

	// Save the key
	if err := savePrivateKey(key, keyPath); err != nil {
		log.Fatalf("Error saving private key: %v", err)
	}

	// Create signer to display address
	signer, err := NewSigner(hexutil.Encode(crypto.FromECDSA(key)))
	if err != nil {
		log.Fatalf("Error creating signer: %v", err)
	}

	fmt.Printf("Generated new %s key at: %s\n", keyType, keyPath)
	fmt.Printf("Ethereum Address: %s\n", signer.GetAddress())

	// Read and display the key for convenience
	keyHex, err := os.ReadFile(keyPath)
	if err != nil {
		log.Fatalf("Error reading key file: %v", err)
	}
	fmt.Printf("Private Key (add 0x prefix for MetaMask): %s\n", string(keyHex))
}

// findSigners locates and loads all signer keys in the directory
func findSigners(currentDir string) ([]*Signer, map[int]*Signer) {
	files, err := os.ReadDir(currentDir)
	if err != nil {
		log.Fatalf("Error reading directory: %v", err)
	}

	allSigners := make([]*Signer, 0)
	signerMapping := make(map[int]*Signer)

	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "signer_key_") && strings.HasSuffix(file.Name(), ".hex") {
			keyPath := filepath.Join(currentDir, file.Name())

			// Extract the signer number
			numStr := strings.TrimPrefix(file.Name(), "signer_key_")
			numStr = strings.TrimSuffix(numStr, ".hex")

			var signerNum int
			if _, err := fmt.Sscanf(numStr, "%d", &signerNum); err != nil {
				log.Printf("Warning: Could not parse signer number from %s: %v", file.Name(), err)
				continue
			}

			key, err := loadPrivateKey(keyPath)
			if err != nil {
				log.Printf("Warning: Error loading key %s: %v", file.Name(), err)
				continue
			}

			signer, err := NewSigner(hexutil.Encode(crypto.FromECDSA(key)))
			if err != nil {
				log.Printf("Warning: Error creating signer from %s: %v", file.Name(), err)
				continue
			}

			allSigners = append(allSigners, signer)
			signerMapping[signerNum] = signer
			fmt.Printf("Found signer #%d: %s from %s\n", signerNum, signer.GetAddress(), file.Name())
		}
	}

	return allSigners, signerMapping
}

// selectSigners determines which signers to use based on the signers flag
func selectSigners(allSigners []*Signer, signerMapping map[int]*Signer, signersFlag string) []*Signer {
	var signers []*Signer

	if signersFlag != "" {
		// Parse the comma-separated list of signer numbers
		signerNumsStr := strings.Split(signersFlag, ",")
		for _, numStr := range signerNumsStr {
			numStr = strings.TrimSpace(numStr)
			var num int
			if _, err := fmt.Sscanf(numStr, "%d", &num); err != nil {
				log.Fatalf("Error parsing signer number '%s': %v", numStr, err)
			}

			if signer, ok := signerMapping[num]; ok {
				signers = append(signers, signer)
				fmt.Printf("Using signer #%d: %s\n", num, signer.GetAddress())
			} else {
				log.Fatalf("Signer #%d not found", num)
			}
		}

		if len(signers) == 0 {
			log.Fatalf("No valid signers specified")
		}
	} else {
		// Use all available signers
		signers = allSigners
		for i := 0; i < len(signers); i++ {
			// Find the signer number
			var signerNum int
			for num, s := range signerMapping {
				if s == signers[i] {
					signerNum = num
					break
				}
			}

			fmt.Printf("Using signer #%d: %s\n", signerNum, signers[i].GetAddress())
		}
	}

	return signers
}

// getAuthSigner determines which signer to use for authentication
func getAuthSigner(signers []*Signer, signerMapping map[int]*Signer, authFlag string, sendFlag bool) *Signer {
	var authSigner *Signer

	if authFlag != "" {
		var authNum int
		if _, err := fmt.Sscanf(authFlag, "%d", &authNum); err != nil {
			log.Fatalf("Error parsing auth signer number '%s': %v", authFlag, err)
		}

		if signer, ok := signerMapping[authNum]; ok {
			authSigner = signer
			fmt.Printf("Using signer #%d for authentication: %s\n", authNum, signer.GetAddress())
		} else {
			log.Fatalf("Auth signer #%d not found", authNum)
		}
	} else if len(signers) > 0 {
		// Default to first signer if not specified
		authSigner = signers[0]

		// Find the signer number for display
		var signerNum int
		for num, s := range signerMapping {
			if s == authSigner {
				signerNum = num
				break
			}
		}
		if sendFlag {
			fmt.Printf("Using signer #%d for authentication: %s\n", signerNum, authSigner.GetAddress())
		}
	}

	return authSigner
}

// prepareRPCMessage creates and signs an RPC message
func prepareRPCMessage(methodFlag string, idFlag uint64, params []any, signers []*Signer, noSignFlag bool) (RPCMessage, []string) {
	// Create RPC data
	rpcData := RPCData{
		RequestID: idFlag,
		Method:    methodFlag,
		Params:    params,
		Timestamp: uint64(time.Now().UnixMilli()),
	}

	// Initialize signatures with an empty array (not null)
	signatures := []string{}

	// Only collect signatures if nosign flag is not set
	if !noSignFlag {
		// Create a temporary client to collect signatures
		tempClient := &Client{
			signers: signers,
		}

		// Determine signing method based on the RPC method
		var dataToSign []byte
		var err error

		switch rpcData.Method {
		case "create_app_session":
			// Special handling for create_app_session
			var createParams CreateAppSessionParams
			paramsJSON, err := json.Marshal(rpcData.Params[0])
			if err != nil {
				log.Fatalf("Error marshaling create app session params: %v", err)
			}
			if err := json.Unmarshal(paramsJSON, &createParams); err != nil {
				log.Fatalf("Error unmarshaling create app session params: %v", err)
			}

			// Create the special sign data structure
			signData := CreateAppSignData{
				RequestID: rpcData.RequestID,
				Method:    rpcData.Method,
				Params:    []CreateAppSessionParams{createParams},
				Timestamp: rpcData.Timestamp,
			}

			// Marshal using the custom MarshalJSON method
			dataToSign, err = signData.MarshalJSON()
			if err != nil {
				log.Fatalf("Error marshaling sign data: %v", err)
			}

		case "close_app_session":
			// Special handling for close_app_session
			var closeParams CloseAppSessionParams
			paramsJSON, err := json.Marshal(rpcData.Params[0])
			if err != nil {
				log.Fatalf("Error marshaling close app session params: %v", err)
			}
			if err := json.Unmarshal(paramsJSON, &closeParams); err != nil {
				log.Fatalf("Error unmarshaling close app session params: %v", err)
			}

			// Create the special sign data structure
			signData := CloseAppSignData{
				RequestID: rpcData.RequestID,
				Method:    rpcData.Method,
				Params:    []CloseAppSessionParams{closeParams},
				Timestamp: rpcData.Timestamp,
			}

			// Marshal using the custom MarshalJSON method
			dataToSign, err = signData.MarshalJSON()
			if err != nil {
				log.Fatalf("Error marshaling sign data: %v", err)
			}

		case "resize_channel":
			// Special handling for resize_channel
			var resizeParams ResizeChannelParams
			paramsJSON, err := json.Marshal(rpcData.Params[0])
			if err != nil {
				log.Fatalf("Error marshaling resize channel params: %v", err)
			}
			if err := json.Unmarshal(paramsJSON, &resizeParams); err != nil {
				log.Fatalf("Error unmarshaling resize channel params: %v", err)
			}

			// Create the special sign data structure
			signData := ResizeChannelSignData{
				RequestID: rpcData.RequestID,
				Method:    rpcData.Method,
				Params:    []ResizeChannelParams{resizeParams},
				Timestamp: rpcData.Timestamp,
			}

			// Marshal using the custom MarshalJSON method
			dataToSign, err = signData.MarshalJSON()
			if err != nil {
				log.Fatalf("Error marshaling sign data: %v", err)
			}

		default:
			// Standard marshaling for other methods
			dataToSign, err = json.Marshal(rpcData)
			if err != nil {
				log.Fatalf("Error marshaling RPC data: %v", err)
			}
		}

		// Collect signatures from all signers
		signatures, err = tempClient.collectSignatures(dataToSign)
		if err != nil {
			log.Fatalf("Error signing data: %v", err)
		}
	}

	// Create final RPC message with signatures
	rpcMessage := RPCMessage{
		Req: &rpcData,
		Sig: signatures,
	}

	return rpcMessage, signatures
}

// printMessageInfo displays information about the message to be sent
func printMessageInfo(rpcMessage RPCMessage, sendFlag bool, params []any, signatures []string,
	signers []*Signer, authSigner *Signer, noSignFlag, noAuthFlag bool, serverFlag string) {
	fmt.Println("\nPayload:")

	// Format the output
	output, err := json.MarshalIndent(rpcMessage, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling final message: %v", err)
	}

	// Show the JSON payload
	fmt.Println(string(output))

	// For non-send mode, show the detailed plan
	if !sendFlag {
		fmt.Println("\nDescription:")

		// Parameters
		if len(params) > 0 {
			paramsJSON, _ := json.MarshalIndent(params, "", "  ")
			fmt.Println("\nParameters:")
			fmt.Println(string(paramsJSON))
		} else {
			fmt.Println("\nParameters: []")
		}

		// Signature info
		signerAddresses := []string{}
		for _, s := range signers {
			signerAddresses = append(signerAddresses, s.GetAddress())
		}

		if noSignFlag {
			fmt.Println("\nSignatures: No signatures will be included (--nosign flag)")
		} else if len(signatures) == 0 {
			fmt.Println("\nSignatures: Empty signature array")
		} else {
			fmt.Printf("\nSignatures: Message will be signed by %d signers\n", len(signatures))
			for i, addr := range signerAddresses {
				fmt.Printf("  - Signer #%d: %s\n", i+1, addr)
			}
		}

		// Auth signer info
		if noAuthFlag {
			fmt.Println("\nAuthentication: None (--noauth flag)")
		} else if authSigner != nil {
			fmt.Printf("\nAuthentication: Using %s for authentication\n", authSigner.GetAddress())
		} else if noSignFlag {
			fmt.Println("\nAuthentication: None (--nosign flag)")
		}

		// Server info
		fmt.Printf("\nTarget server: %s\n", serverFlag)
		fmt.Println("\nTo execute this plan, run with the --send flag")
		fmt.Println()
	}
}

// readResponses reads and displays responses from the server
func readResponses(client *Client) {
	fmt.Println("\nServer responses:")
	responseCount := 0

	for {
		// Set a read deadline to avoid waiting indefinitely
		client.conn.SetReadDeadline(time.Now().Add(2 * time.Second))

		_, respMsg, err := client.conn.ReadMessage()
		if err != nil {
			// Check if this is just a timeout (no more messages)
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) ||
				websocket.IsUnexpectedCloseError(err) ||
				err.Error() == "websocket: close 1000 (normal)" {
				fmt.Println("Connection closed by server.")
				break
			} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// This is a timeout, likely no more messages
				if responseCount > 0 {
					fmt.Println("No more messages received.")
				} else {
					fmt.Println("No response received within timeout period.")
				}
				break
			}

			log.Fatalf("Error reading response: %v", err)
		}

		// Pretty print the response
		var respObj map[string]any
		if err := json.Unmarshal(respMsg, &respObj); err != nil {
			log.Fatalf("Error parsing response: %v", err)
		}

		respOut, err := json.MarshalIndent(respObj, "", "  ")
		if err != nil {
			log.Fatalf("Error marshaling response: %v", err)
		}

		fmt.Printf("\nResponse #%d:\n", responseCount+1)
		fmt.Println(string(respOut))
		responseCount++
	}
}
