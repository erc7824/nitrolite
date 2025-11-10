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

// CloseAppSessionParams represents parameters needed for virtual app closure
type CloseAppSessionParams struct {
	AppSessionID string          `json:"app_session_id"`
	Allocations  []AppAllocation `json:"allocations"`
}

// ResizeChannelParams represents parameters needed for resizing a channel
type ResizeChannelParams struct {
	ChannelID        string   `json:"channel_id"`
	AllocateAmount   *big.Int `json:"allocate_amount,omitempty"`
	ResizeAmount     *big.Int `json:"resize_amount,omitempty"`
	FundsDestination string   `json:"funds_destination"`
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
	return nitrolite.Sign(data, s.privateKey)
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
func NewClient(
	serverURL string,
	authSigner *Signer,
	noSignatures bool,
	noAuth bool,
	signers ...*Signer,
) (*Client, error) {
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
		// If authSigner is nil but we need auth, pick the first
		if authSigner == nil && !noAuth {
			authSigner = signers[0]
		}

		if authSigner != nil {
			primaryAddress = authSigner.GetAddress()
		}

		addresses = make([]string, len(signers))
		for i, signer := range signers {
			addresses[i] = signer.GetAddress()
		}
	} else if authSigner != nil {
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
	// If we have a JWT token and no sid was already set, add it
	if c.jwt != "" && rpcMsg.AppSessionID == "" {
		rpcMsg.AppSessionID = c.jwt
	}

	data, err := json.Marshal(rpcMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

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
		sig, err := signer.Sign(data)
		if err != nil {
			return nil, fmt.Errorf("failed to sign with signer %d: %w", i, err)
		}
		signatures[i] = hexutil.Encode(sig)
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

	// Step 1: send "auth_request"
	authReq := RPCMessage{
		Req: &RPCData{
			RequestID: c.nextRequestID,
			Method:    "auth_request",
			Params: []any{
				c.address,
				c.addresses[0],
				"test-app",
				[]interface{}{[]interface{}{"usdc", "10000"}},
				"3600",
				"all",
				"0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
			},
			Timestamp: uint64(time.Now().UnixMilli()),
		},
		Sig: []string{},
	}
	c.nextRequestID++

	// Sign the auth_request itself
	reqData, err := json.Marshal(authReq.Req)
	if err != nil {
		return fmt.Errorf("failed to marshal auth request: %w", err)
	}

	signature, err := c.authSigner.Sign(reqData)
	if err != nil {
		return fmt.Errorf("failed to sign auth request: %w", err)
	}
	authReq.Sig = []string{hexutil.Encode(signature)}

	if err := c.SendMessage(authReq); err != nil {
		return fmt.Errorf("failed to send auth request: %w", err)
	}

	fmt.Println("Waiting for challenge...")
	var challengeStr string
	challengeDeadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(challengeDeadline) {
		c.conn.SetReadDeadline(challengeDeadline)
		_, challengeMsg, err := c.conn.ReadMessage()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return fmt.Errorf("timed out waiting for challenge")
			}
			return fmt.Errorf("failed to read challenge response: %w", err)
		}

		var challengeResp map[string]any
		if err := json.Unmarshal(challengeMsg, &challengeResp); err != nil {
			return fmt.Errorf("failed to parse challenge response: %w", err)
		}

		if resArray, ok := challengeResp["res"].([]any); ok {
			if len(resArray) > 1 {
				if method, ok := resArray[1].(string); ok {
					if method == "assets" {
						fmt.Printf("Skipping non-auth message: %s\n", method)
						continue
					}
					if method == "auth_challenge" {
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

		// Alternative location
		if challengeStr == "" {
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

	c.conn.SetReadDeadline(time.Time{})

	if challengeStr == "" {
		fmt.Println("No auth challenge received; skipping auth flow.")
		return nil
	}

	fmt.Printf("Found challenge: %s\n", challengeStr)

	// Step 3: send "auth_verify"
	verifyReq := RPCMessage{
		Req: &RPCData{
			RequestID: c.nextRequestID,
			Method:    "auth_verify",
			Params: []any{
				map[string]any{"challenge": challengeStr},
			},
			Timestamp: uint64(time.Now().UnixMilli()),
		},
		Sig: []string{},
	}
	c.nextRequestID++

	privKey := c.authSigner.privateKey
	convertedAllowances := convertAllowances([]Allowance{{Asset: "usdc", Amount: "10000"}})

	td := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {{Name: "name", Type: "string"}},
			"Policy": {
				{Name: "challenge", Type: "string"},
				{Name: "scope", Type: "string"},
				{Name: "wallet", Type: "address"},
				{Name: "session_key", Type: "address"},
				{Name: "expires_at", Type: "uint64"},
				{Name: "allowances", Type: "Allowance[]"},
			},
			"Allowance": {
				{Name: "asset", Type: "string"},
				{Name: "amount", Type: "string"},
			},
		},
		PrimaryType: "Policy",
		Domain:      apitypes.TypedDataDomain{Name: "test-app"},
		Message: map[string]interface{}{
			"challenge":   challengeStr,
			"scope":       "all",
			"wallet":      c.address,
			"session_key": c.addresses[0],
			"expires_at":  big.NewInt(3600),
			"allowances":  convertedAllowances,
		},
	}

	hash, _, err := apitypes.TypedDataAndHash(td)
	if err != nil {
		return err
	}

	sigBytes, err := crypto.Sign(hash, privKey)
	if err != nil {
		return err
	}
	verifyReq.Sig = []string{hexutil.Encode(sigBytes)}

	if err := c.SendMessage(verifyReq); err != nil {
		return fmt.Errorf("failed to send verify request: %w", err)
	}

	fmt.Println("Waiting for verification response...")
	verifyDeadline := time.Now().Add(5 * time.Second)
	var success bool
	var foundVerifyResponse bool

	for time.Now().Before(verifyDeadline) {
		c.conn.SetReadDeadline(verifyDeadline)
		_, verifyMsg, err := c.conn.ReadMessage()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return fmt.Errorf("timed out waiting for verification response")
			}
			return fmt.Errorf("failed to read verify response: %w", err)
		}

		var verifyResp map[string]any
		if err := json.Unmarshal(verifyMsg, &verifyResp); err != nil {
			return fmt.Errorf("failed to parse verify response: %w", err)
		}

		resVerifyArray, ok := verifyResp["res"].([]any)
		if !ok || len(resVerifyArray) < 3 {
			fmt.Println("Skipping non-auth message (invalid format)")
			continue
		}

		if len(resVerifyArray) > 1 {
			if method, ok := resVerifyArray[1].(string); ok {
				if method == "assets" || method == "error" || method == "pong" {
					fmt.Printf("Skipping non-auth message: %s\n", method)
					continue
				}
				if method == "auth_verify" {
					foundVerifyResponse = true
				}
			}
		}

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

		if successValue, ok := verifyObj["success"].(bool); ok {
			success = successValue
			foundVerifyResponse = true
			if token, ok := verifyObj["token"].(string); ok {
				c.jwt = token
				fmt.Println("JWT token received!")
			}
			break
		}
	}

	c.conn.SetReadDeadline(time.Time{})

	if !foundVerifyResponse {
		fmt.Println("No verification response received; proceeding anyway.")
		return nil
	}

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
	var (
		methodFlag  = flag.String("method", "", "RPC method name")
		idFlag      = flag.Uint64("id", 1, "Request ID")
		paramsFlag  = flag.String("params", "[]", "JSON array of parameters")
		sendFlag    = flag.Bool("send", false, "Send the message to the server")
		serverFlag  = flag.String("server", "ws://localhost:8000/ws", "WebSocket server URL (or set SERVER env)")
		genKeyFlag  = flag.String("genkey", "", "Generate a new key and exit. Use a signer number (e.g., '1', '2').")
		signersFlag = flag.String("signers", "", "Comma-separated signer numbers (e.g., '1,2,3'). If empty, all found signers are used.")
		authFlag    = flag.String("auth", "", "Signer number to authenticate with (e.g., '1'). Defaults to first signer if omitted.")
		noSignFlag  = flag.Bool("nosign", false, "Send request without signatures")
		noAuthFlag  = flag.Bool("noauth", false, "Skip authentication flow")
	)

	flag.Parse()

	if serverEnv := os.Getenv("SERVER"); serverEnv != "" {
		*serverFlag = serverEnv
	}

	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting working directory: %v", err)
	}

	if *genKeyFlag != "" {
		generateKey(*genKeyFlag, currentDir)
		os.Exit(0)
	}

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

	// Find and load all signers from working directory
	allSigners, signerMapping := findSigners(currentDir)
	if len(allSigners) == 0 {
		log.Fatalf("No signers found. Generate at least one key with --genkey.")
	}

	// Determine which signers to use
	signers := selectSigners(allSigners, signerMapping, *signersFlag)
	authSigner := getAuthSigner(signers, signerMapping, *authFlag, *sendFlag)

	// Build & sign the RPC message
	rpcMessage, signatures := prepareRPCMessage(*methodFlag, *idFlag, params, signers, *noSignFlag)

	// Display payload (and, if not sending, a descriptive “plan”)
	printMessageInfo(rpcMessage, *sendFlag, params, signatures, signers, authSigner, *noSignFlag, *noAuthFlag, *serverFlag)

	if *sendFlag {
		client, err := NewClient(*serverFlag, authSigner, *noSignFlag, *noAuthFlag, signers...)
		if err != nil {
			log.Fatalf("Error creating client: %v", err)
		}
		defer client.Close()

		if err := client.Authenticate(); err != nil {
			log.Fatalf("Authentication failed: %v", err)
		}

		if err := client.SendMessage(rpcMessage); err != nil {
			log.Fatalf("Error sending message: %v", err)
		}

		readResponses(client)
	}
}

// generateKey creates a new key and displays its information
func generateKey(genKeyFlag string, currentDir string) {
	var signerNum int
	if _, err := fmt.Sscanf(genKeyFlag, "%d", &signerNum); err != nil {
		log.Fatalf("Invalid genkey value. Use a signer number (e.g., '1'): %v", err)
	}
	if signerNum < 1 {
		log.Fatalf("Signer number must be at least 1")
	}

	keyPath := filepath.Join(currentDir, fmt.Sprintf("signer_key_%d.hex", signerNum))
	key, err := generatePrivateKey()
	if err != nil {
		log.Fatalf("Error generating private key: %v", err)
	}
	if err := savePrivateKey(key, keyPath); err != nil {
		log.Fatalf("Error saving private key: %v", err)
	}

	signer, err := NewSigner(hexutil.Encode(crypto.FromECDSA(key)))
	if err != nil {
		log.Fatalf("Error creating signer: %v", err)
	}

	fmt.Printf("Generated signer #%d key at: %s\n", signerNum, keyPath)
	fmt.Printf("Ethereum Address: %s\n", signer.GetAddress())

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
		for _, numStr := range strings.Split(signersFlag, ",") {
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
		signers = allSigners
		for i := 0; i < len(signers); i++ {
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
		authSigner = signers[0]
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
func prepareRPCMessage(
	methodFlag string,
	idFlag uint64,
	params []any,
	signers []*Signer,
	noSignFlag bool,
) (RPCMessage, []string) {
	// Build the RPCData
	rpcData := RPCData{
		RequestID: idFlag,
		Method:    methodFlag,
		Params:    params,
		Timestamp: uint64(time.Now().UnixMilli()),
	}

	// If noSignFlag is true, we leave Sig empty
	var signatures []string
	if !noSignFlag {
		// Always marshal the raw RPCData to a 4-element array
		dataToSign, err := json.Marshal(rpcData)
		if err != nil {
			log.Fatalf("Error marshaling RPC data for signing: %v", err)
		}

		// Temporarily make a client just to collect signatures
		tempClient := &Client{signers: signers}
		signatures, err = tempClient.collectSignatures(dataToSign)
		if err != nil {
			log.Fatalf("Error signing data: %v", err)
		}
	}

	rpcMessage := RPCMessage{
		Req: &rpcData,
		Sig: signatures,
	}
	return rpcMessage, signatures
}

// printMessageInfo displays information about the message to be sent
func printMessageInfo(
	rpcMessage RPCMessage,
	sendFlag bool,
	params []any,
	signatures []string,
	signers []*Signer,
	authSigner *Signer,
	noSignFlag, noAuthFlag bool,
	serverFlag string,
) {
	fmt.Println("\nPayload:")
	output, err := json.MarshalIndent(rpcMessage, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling final message: %v", err)
	}
	fmt.Println(string(output))

	if !sendFlag {
		fmt.Println("\nDescription:")

		if len(params) > 0 {
			paramsJSON, _ := json.MarshalIndent(params, "", "  ")
			fmt.Println("\nParameters:")
			fmt.Println(string(paramsJSON))
		} else {
			fmt.Println("\nParameters: []")
		}

		if noSignFlag {
			fmt.Println("\nSignatures: No signatures will be included (--nosign flag)")
		} else if len(signatures) == 0 {
			fmt.Println("\nSignatures: Empty signature array")
		} else {
			fmt.Printf("\nSignatures: Message will be signed by %d signers\n", len(signatures))
			for i, s := range signers {
				fmt.Printf("  - Signer #%d: %s\n", i+1, s.GetAddress())
			}
		}

		if noAuthFlag {
			fmt.Println("\nAuthentication: None (--noauth flag)")
		} else if authSigner != nil {
			fmt.Printf("\nAuthentication: Using %s for authentication\n", authSigner.GetAddress())
		} else if noSignFlag {
			fmt.Println("\nAuthentication: None (--nosign flag)")
		}

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
		client.conn.SetReadDeadline(time.Now().Add(2 * time.Second))

		_, respMsg, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) ||
				websocket.IsUnexpectedCloseError(err) ||
				err.Error() == "websocket: close 1000 (normal)" {
				fmt.Println("Connection closed by server.")
				break
			} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				if responseCount > 0 {
					fmt.Println("No more messages received.")
				} else {
					fmt.Println("No response received within timeout period.")
				}
				break
			}
			log.Fatalf("Error reading response: %v", err)
		}

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
