package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gorilla/websocket"
)

const (
	pingRequestID = 100             // Request ID for ping messages
	pingInterval  = 5 * time.Second // Interval for ping messages
)

type ClearnodeClient struct {
	conn   *websocket.Conn
	signer *Signer // User's Signer

	printEvents   bool
	responseSinks map[uint64]chan *RPCResponse // Map of request IDs to response channels
	mu            sync.RWMutex                 // Mutex to protect access to responseSinks
}

type NetworkInfo struct {
	Name               string `json:"name"`
	ChainID            uint32 `json:"chain_id"`
	CustodyAddress     string `json:"custody_address"`
	AdjudicatorAddress string `json:"adjudicator_address"`
}

type BrokerConfig struct {
	BrokerAddress string        `json:"broker_address"`
	Networks      []NetworkInfo `json:"networks"`
}

func NewClearnodeClient(wsURL string) (*ClearnodeClient, error) {
	dialer := websocket.Dialer{
		HandshakeTimeout:  5 * time.Second,
		EnableCompression: true,
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return nil, err
	}

	client := &ClearnodeClient{
		conn:          conn,
		responseSinks: make(map[uint64]chan *RPCResponse),
	}
	go client.readMessages()
	go client.pingPeriodically()

	return client, nil
}

func (c *ClearnodeClient) Signer() *Signer {
	return c.signer
}

func (c *ClearnodeClient) GetConfig() (*BrokerConfig, error) {
	res, err := c.request("get_config", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch config: %w", err)
	}
	if res.Res.Method != "get_config" || len(res.Res.Params) < 1 {
		return nil, fmt.Errorf("unexpected response to config request: %v", res.Res)
	}

	configJSON, err := json.Marshal(res.Res.Params[0])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config data: %w", err)
	}

	var config BrokerConfig
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return nil, fmt.Errorf("failed to parse broker config: %w", err)
	}

	return &config, nil
}

func (c *ClearnodeClient) GetSupportedAssets() ([]Asset, error) {
	res, err := c.request("get_assets", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch assets: %w", err)
	}
	if res.Res.Method != "get_assets" || len(res.Res.Params) < 1 {
		return nil, fmt.Errorf("unexpected response to assets request: %v", res.Res)
	}

	assetsJSON, err := json.Marshal(res.Res.Params[0])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal assets data: %w", err)
	}

	var assets []Asset
	if err := json.Unmarshal(assetsJSON, &assets); err != nil {
		return nil, fmt.Errorf("failed to parse assets: %w", err)
	}

	return assets, nil
}

func (c *ClearnodeClient) GetChannels(participant, status string) ([]Channel, error) {
	params := map[string]any{
		"participant": participant,
		"status":      status,
	}

	res, err := c.request("get_channels", nil, params)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channels: %w", err)
	}
	if res.Res.Method != "get_channels" || len(res.Res.Params) < 1 {
		return nil, fmt.Errorf("unexpected response to channels request: %v", res.Res)
	}

	channelsJSON, err := json.Marshal(res.Res.Params[0])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal channels data: %w", err)
	}

	var channels []Channel
	if err := json.Unmarshal(channelsJSON, &channels); err != nil {
		return nil, fmt.Errorf("failed to parse channels: %w", err)
	}

	return channels, nil
}

func (c *ClearnodeClient) Authenticate(wallet, signer *Signer) error {
	if c.signer != nil {
		return nil // Already authenticated
	}

	ch := AuthChallenge{
		Wallet:      wallet.Address().Hex(),
		Participant: signer.Address().Hex(), // Using address as session key for simplicity
		AppName:     "Yellow Bridge",
		Allowances:  []any{},                // No allowances for now
		Expire:      "",                     // No expiration for now
		Scope:       "",                     // No specific scope for now
		AppAddress:  wallet.Address().Hex(), // Using address as app address for simplicity
	}
	res, err := c.request("auth_request", nil,
		ch.Wallet,
		ch.Participant,
		ch.AppName,
		ch.Allowances,
		ch.Expire,
		ch.Scope,
		ch.AppAddress,
	)
	if err != nil {
		return fmt.Errorf("authentication request failed: %w", err)
	}
	if res.Res.Method != "auth_challenge" || len(res.Res.Params) < 1 {
		return fmt.Errorf("unexpected response to auth_request: %v", res.Res)
	}

	challengeMap, ok := res.Res.Params[0].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid auth_challenge response format: %v", res.Res.Params[0])
	}
	challengeToken, ok := challengeMap["challenge_message"].(string)
	if !ok {
		return fmt.Errorf("challenge_message not found in auth_challenge response: %v", challengeMap)
	}

	ch.Token = challengeToken
	chSig, err := signChallenge(wallet, ch)
	if err != nil {
		return fmt.Errorf("failed to sign challenge: %w", err)
	}
	authVerifyChallenge := map[string]any{
		"challenge": challengeToken,
	}
	res, err = c.request("auth_verify", []string{hexutil.Encode(chSig)}, authVerifyChallenge)
	if err != nil {
		return fmt.Errorf("authentication verification failed: %w", err)
	}
	if res.Res.Method != "auth_verify" || len(res.Res.Params) < 1 {
		return fmt.Errorf("unexpected response to auth_verify: %v", res.Res)
	}

	verifyMap, ok := res.Res.Params[0].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid auth_verify response format: %v", res.Res.Params[0])
	}
	if authSuccess, _ := verifyMap["success"].(bool); !authSuccess {
		return fmt.Errorf("authentication failed: %v", verifyMap)
	}

	c.signer = signer
	return nil
}

func (c *ClearnodeClient) readMessages() {
	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			fmt.Printf("Error reading message: %s\n", err.Error())
			return
		}

		var msg RPCResponse
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			fmt.Printf("Malformed message: %s\n", string(messageBytes))
			continue
		}

		c.mu.Lock()
		responseSink, exists := c.responseSinks[msg.Res.RequestID]
		c.mu.Unlock()
		if !exists {
			c.handleEvent(msg.Res) // Handle response as an event if no response sink exists
			continue
		}
		responseSink <- &msg // Send the response to the channel
	}
}

func (c *ClearnodeClient) pingPeriodically() {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for range ticker.C {
		res, err := c.request("ping", nil)
		if err != nil {
			fmt.Printf("Error sending ping: %s\n", err.Error())
			return
		}

		if res.Res.Method != "pong" {
			fmt.Printf("Unexpected response to ping: %s\n", res.Res.Method)
			continue
		}
	}
}

func (c *ClearnodeClient) request(method string, sigs []string, params ...any) (*RPCResponse, error) {
	if params == nil {
		params = []any{} // Ensure params is never nil
	}

	if sigs == nil {
		sigs = []string{} // Ensure sigs is never nil
	}

	reqID := uint64(time.Now().UnixMilli())
	req := RPCRequest{
		Req: RPCData{
			RequestID: reqID,
			Method:    method,
			Params:    params,
			Timestamp: uint64(time.Now().UnixMilli()),
		},
		Sig: sigs,
	}

	responseSink := make(chan *RPCResponse, 1) // Create a channel for this request
	c.mu.Lock()
	c.responseSinks[reqID] = responseSink
	c.mu.Unlock()

	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	if err := c.conn.WriteMessage(websocket.TextMessage, reqJSON); err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	res := <-responseSink // Wait for the response
	c.mu.Lock()
	delete(c.responseSinks, reqID) // Remove the response sink after receiving
	c.mu.Unlock()

	if res == nil {
		return nil, fmt.Errorf("no response received for request %d", reqID)
	}

	return res, nil
}

type RPCRequest struct {
	Req RPCData  `json:"req"`
	Sig []string `json:"sig"`
}

type RPCResponse struct {
	Res RPCData  `json:"res"`
	Sig []string `json:"sig"`
}

// RPCData represents the common structure for both requests and responses
// Format: [request_id, method, params, ts]
type RPCData struct {
	RequestID uint64 `json:"request_id" validate:"required"`
	Method    string `json:"method" validate:"required"`
	Params    []any  `json:"params" validate:"required"`
	Timestamp uint64 `json:"ts" validate:"required"`
}

// UnmarshalJSON implements the json.Unmarshaler interface for RPCMessage
func (m *RPCData) UnmarshalJSON(data []byte) error {
	var rawArr []json.RawMessage
	if err := json.Unmarshal(data, &rawArr); err != nil {
		return fmt.Errorf("error reading RPCData as array: %w", err)
	}
	if len(rawArr) != 4 {
		return fmt.Errorf("invalid RPCData: expected 4 elements in array")
	}

	// Element 0: uint64 RequestID
	if err := json.Unmarshal(rawArr[0], &m.RequestID); err != nil {
		return fmt.Errorf("invalid request_id: %w", err)
	}
	// Element 1: string Method
	if err := json.Unmarshal(rawArr[1], &m.Method); err != nil {
		return fmt.Errorf("invalid method: %w", err)
	}
	// Element 2: []any Params
	if err := json.Unmarshal(rawArr[2], &m.Params); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	// Element 3: uint64 Timestamp
	if err := json.Unmarshal(rawArr[3], &m.Timestamp); err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}

	return nil
}

// MarshalJSON for RPCData always emits the array‐form [RequestID, Method, Params, Timestamp].
func (m RPCData) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{
		m.RequestID,
		m.Method,
		m.Params,
		m.Timestamp,
	})
}
