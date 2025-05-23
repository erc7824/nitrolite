package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"flag"
	"fmt"

	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/erc7824/nitrolite/clearnode/nitrolite"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/websocket"
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
	// Create array representation
	return json.Marshal([]any{
		m.RequestID,
		m.Method,
		m.Params,
		m.Timestamp,
	})
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
	// Remove "0x" prefix
	if len(keyHex) >= 2 && keyHex[:2] == "0x" {
		keyHex = keyHex[2:]
	}

	// Create directory if it doesn't exist
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

// getOrCreatePrivateKey gets an existing private key or creates a new one
func getOrCreatePrivateKey(keyPath string) (*ecdsa.PrivateKey, error) {
	if _, err := os.Stat(keyPath); err == nil {
		// Key file exists, load it
		key, err := loadPrivateKey(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load existing key: %w", err)
		}
		return key, nil
	}

	// Generate new key
	key, err := generatePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new key: %w", err)
	}

	// Save the key
	if err := savePrivateKey(key, keyPath); err != nil {
		return nil, fmt.Errorf("failed to save new key: %w", err)
	}

	return key, nil
}

// Client handles websocket connection and RPC messaging
type Client struct {
	conn         *websocket.Conn
	signers      []*Signer
	address      string // Primary address (for backward compatibility)
	addresses    []string
	authSigner   *Signer // Signer used for authentication
	noSignatures bool    // Flag to indicate if signatures should be added
}

// NewClient creates a new websocket client
func NewClient(serverURL string, authSigner *Signer, noSignatures bool, signers ...*Signer) (*Client, error) {
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
		// Set auth signer if not specified
		if authSigner == nil {
			authSigner = signers[0]
		}

		// We'll use the auth signer's address as the primary address for auth
		primaryAddress = authSigner.GetAddress()

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
		conn:         conn,
		signers:      signers,
		address:      primaryAddress,
		addresses:    addresses,
		authSigner:   authSigner,
		noSignatures: noSignatures,
	}, nil
}

// SendMessage sends an RPC message to the server
func (c *Client) SendMessage(rpcMsg RPCMessage) error {
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
	// If noSignatures flag is set, return empty signature array
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
	fmt.Println("Starting authentication...")

	// If no auth signer is provided, we can't authenticate
	if c.authSigner == nil {
		return fmt.Errorf("no authentication signer provided")
	}

	// Step 1: Auth request
	authReq := RPCMessage{
		Req: &RPCData{
			RequestID: 1,
			Method:    "auth_request",
			Params:    []any{c.address},
			Timestamp: uint64(time.Now().UnixMilli()),
		},
		Sig: []string{},
	}

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

	// Send auth request
	if err := c.SendMessage(authReq); err != nil {
		return fmt.Errorf("failed to send auth request: %w", err)
	}

	// Step 2: Receive challenge
	fmt.Println("Waiting for challenge...")
	_, challengeMsg, err := c.conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read challenge response: %w", err)
	}

	var challengeResp map[string]any
	if err := json.Unmarshal(challengeMsg, &challengeResp); err != nil {
		return fmt.Errorf("failed to parse challenge response: %w", err)
	}

	// Extract challenge from response
	resArray, ok := challengeResp["res"].([]any)
	if !ok || len(resArray) < 3 {
		return fmt.Errorf("invalid challenge response format")
	}

	paramsArray, ok := resArray[2].([]any)
	if !ok || len(paramsArray) < 1 {
		return fmt.Errorf("invalid challenge parameters")
	}

	challengeObj, ok := paramsArray[0].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid challenge object")
	}

	challengeStr, ok := challengeObj["challenge_message"].(string)
	if !ok {
		return fmt.Errorf("missing challenge message")
	}

	// Step 3: Send auth verify
	fmt.Println("Sending challenge verification...")
	verifyReq := RPCMessage{
		Req: &RPCData{
			RequestID: 2,
			Method:    "auth_verify",
			Params: []any{map[string]any{
				"address":   c.address,
				"challenge": challengeStr,
			}},
			Timestamp: uint64(time.Now().UnixMilli()),
		},
		Sig: []string{},
	}

	// Sign the verify request with auth signer
	verifyData, err := json.Marshal(verifyReq.Req)
	if err != nil {
		return fmt.Errorf("failed to marshal verify request: %w", err)
	}

	// For authentication, we always need a signature regardless of noSignatures setting
	verifySignature, err := c.authSigner.Sign(verifyData)
	if err != nil {
		return fmt.Errorf("failed to sign verify request: %w", err)
	}
	verifyReq.Sig = []string{hexutil.Encode(verifySignature)}

	// Send verify request
	if err := c.SendMessage(verifyReq); err != nil {
		return fmt.Errorf("failed to send verify request: %w", err)
	}

	// Receive auth verify response
	_, verifyMsg, err := c.conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read verify response: %w", err)
	}

	var verifyResp map[string]any
	if err := json.Unmarshal(verifyMsg, &verifyResp); err != nil {
		return fmt.Errorf("failed to parse verify response: %w", err)
	}

	// Check if auth was successful
	resVerifyArray, ok := verifyResp["res"].([]any)
	if !ok || len(resVerifyArray) < 3 {
		return fmt.Errorf("invalid verify response format")
	}

	verifyParamsArray, ok := resVerifyArray[2].([]any)
	if !ok || len(verifyParamsArray) < 1 {
		return fmt.Errorf("invalid verify parameters")
	}

	verifyObj, ok := verifyParamsArray[0].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid verify object")
	}

	success, ok := verifyObj["success"].(bool)
	if !ok || !success {
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
		var keyPath string
		var keyType string

		// Try to parse as a signer number
		var signerNum int
		if _, err := fmt.Sscanf(*genKeyFlag, "%d", &signerNum); err != nil {
			log.Fatalf("Invalid genkey value. Use a signer number (e.g., '1', '2', '3'): %v", err)
		}

		if signerNum < 1 {
			log.Fatalf("Signer number must be at least 1")
		}

		// Generate signer key
		keyPath = filepath.Join(currentDir, fmt.Sprintf("signer_key_%d.hex", signerNum))
		keyType = fmt.Sprintf("signer #%d", signerNum)

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

	// Look for all signer keys in the directory
	files, err := os.ReadDir(currentDir)
	if err != nil {
		log.Fatalf("Error reading directory: %v", err)
	}

	// Load all available signers
	allSigners := make([]*Signer, 0)
	signerMapping := make(map[int]*Signer)

	// Find any signer key files (format: signer_key_X.hex)
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

	if len(allSigners) == 0 {
		log.Fatalf("No signers found. Generate at least one key.")
	}

	// Determine which signers to use based on the signers flag
	var signers []*Signer
	if *signersFlag != "" {
		// Parse the comma-separated list of signer numbers
		signerNumsStr := strings.Split(*signersFlag, ",")
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

	// Create RPC data
	rpcData := RPCData{
		RequestID: *idFlag,
		Method:    *methodFlag,
		Params:    params,
		Timestamp: uint64(time.Now().UnixMilli()),
	}

	// Serialize RPC data for signing
	dataBytes, err := json.Marshal(rpcData)
	if err != nil {
		log.Fatalf("Error marshaling RPC data: %v", err)
	}

	// Initialize signatures with an empty array (not null)
	signatures := []string{}

	// Only collect signatures if nosign flag is not set
	if !*noSignFlag {
		// Create a temporary client to collect signatures
		tempClient := &Client{
			signers: signers,
		}

		// Collect signatures from all signers
		var err error
		signatures, err = tempClient.collectSignatures(dataBytes)
		if err != nil {
			log.Fatalf("Error signing data: %v", err)
		}
	}

	// Create final RPC message with signatures (or empty array if nosign is set)
	rpcMessage := RPCMessage{
		Req: &rpcData,
		Sig: signatures,
	}

	// Validate auth signer if specified, even if not sending
	var authSigner *Signer
	if *authFlag != "" {
		var authNum int
		if _, err := fmt.Sscanf(*authFlag, "%d", &authNum); err != nil {
			log.Fatalf("Error parsing auth signer number '%s': %v", *authFlag, err)
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
		if *sendFlag {
			fmt.Printf("Using signer #%d for authentication: %s\n", signerNum, authSigner.GetAddress())
		}
	}

	fmt.Println("\nPayload:")

	// Format the output differently based on whether we're sending
	output, err := json.MarshalIndent(rpcMessage, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling final message: %v", err)
	}

	// Always show the JSON payload
	fmt.Println(string(output))

	// For non-send mode, also show the detailed plan
	if !*sendFlag {
		fmt.Println("\nDescription:")

		// Parameters
		if len(rpcData.Params) > 0 {
			paramsJSON, _ := json.MarshalIndent(rpcData.Params, "", "  ")
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

		if *noSignFlag {
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
		if authSigner != nil {
			fmt.Printf("\nAuthentication: Using %s for authentication\n", authSigner.GetAddress())
		} else if *noSignFlag {
			fmt.Println("\nAuthentication: None (--nosign flag)")
		}

		// Server info
		fmt.Printf("\nTarget server: %s\n", *serverFlag)
		fmt.Println("\nTo execute this plan, run with the --send flag")
		fmt.Println()
	}

	// If send flag is set, send the message to the server
	if *sendFlag {

		// Create the client with the specified settings
		client, err := NewClient(*serverFlag, authSigner, *noSignFlag, signers...)
		if err != nil {
			log.Fatalf("Error creating client: %v", err)
		}
		defer client.Close()

		// Skip authentication if noSign flag is set and no auth signer is provided
		if authSigner == nil && *noSignFlag {
			fmt.Println("Skipping authentication as no auth signer is provided and nosign flag is set")
		} else {
			// Authenticate with the server
			if err := client.Authenticate(); err != nil {
				log.Fatalf("Authentication failed: %v", err)
			}
		}

		// Send the message
		if err := client.SendMessage(rpcMessage); err != nil {
			log.Fatalf("Error sending message: %v", err)
		}

		// Keep reading responses until there's an error
		fmt.Println("\nServer responses:")
		responseCount := 0

		for {
			// Set a read deadline to avoid waiting indefinitely
			client.conn.SetReadDeadline(time.Now().Add(2 * time.Second))

			_, respMsg, err := client.conn.ReadMessage()
			if err != nil {
				// Check if this is just a timeout (no more messages)
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) || websocket.IsUnexpectedCloseError(err) || err.Error() == "websocket: close 1000 (normal)" {
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
}
