package clearnet

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"

	"github.com/erc7824/nitrolite/examples/cerebro/unisig"
)

const (
	pingRequestID = 100             // Request ID for ping messages
	pingInterval  = 5 * time.Second // Interval for ping messages
)

type ClearnodeClient struct {
	conn   *websocket.Conn
	signer unisig.Signer // User's Signer

	printEvents   bool
	responseSinks map[uint64]chan *RPCResponse // Map of request IDs to response channels
	mu            sync.RWMutex                 // Mutex to protect access to responseSinks
	exitCh        chan struct{}                // Channel to signal client exit
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
		exitCh:        make(chan struct{}),
	}
	go client.readMessages()
	go client.pingPeriodically()

	return client, nil
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

func (c *ClearnodeClient) GetSupportedAssets() ([]AssetRes, error) {
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

	var assets []AssetRes
	if err := json.Unmarshal(assetsJSON, &assets); err != nil {
		return nil, fmt.Errorf("failed to parse assets: %w", err)
	}

	return assets, nil
}

type BalanceRes struct {
	Asset  string          `json:"asset"`
	Amount decimal.Decimal `json:"amount"`
}

func (c *ClearnodeClient) GetLedgerBalances() ([]BalanceRes, error) {
	res, err := c.request("get_ledger_balances", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch balances: %w", err)
	}
	if res.Res.Method != "get_ledger_balances" || len(res.Res.Params) < 1 {
		return nil, fmt.Errorf("unexpected response to get_ledger_balances request: %v", res.Res)
	}

	assetsJSON, err := json.Marshal(res.Res.Params[0])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal assets data: %w", err)
	}

	var balances []BalanceRes
	if err := json.Unmarshal(assetsJSON, &balances); err != nil {
		return nil, fmt.Errorf("failed to parse assets: %w", err)
	}

	return balances, nil
}

func (c *ClearnodeClient) GetChannels(participant, status string) ([]ChannelRes, error) {
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

	var channels []ChannelRes
	if err := json.Unmarshal(channelsJSON, &channels); err != nil {
		return nil, fmt.Errorf("failed to parse channels: %w", err)
	}

	return channels, nil
}

type ChannelClosureRes struct {
	ChannelID        string          `json:"channel_id"`
	Intent           uint8           `json:"intent"`
	Version          uint64          `json:"version"`
	StateData        string          `json:"state_data"`
	FinalAllocations []AllocationRes `json:"allocations"`
	StateHash        string          `json:"state_hash"`
	Signature        SignatureRes    `json:"server_signature"`
}

type AllocationRes struct {
	Destination string          `json:"destination"`
	Token       string          `json:"token"`
	Amount      decimal.Decimal `json:"amount"`
}

type SignatureRes struct {
	V uint8  `json:"v,string"`
	R string `json:"r"`
	S string `json:"s"`
}

func (c *ClearnodeClient) RequestChannelClosure(walletAddress common.Address, channelID string) (*ChannelClosureRes, error) {
	if c.signer == nil {
		return nil, fmt.Errorf("client not authenticated")
	}

	params := map[string]any{
		"funds_destination": walletAddress.Hex(),
		"channel_id":        channelID,
	}

	res, err := c.request("close_channel", nil, params)
	if err != nil {
		return nil, fmt.Errorf("failed to request channel closure: %w", err)
	}
	if res.Res.Method != "close_channel" || len(res.Res.Params) < 1 {
		return nil, fmt.Errorf("unexpected response to close_channel: %v", res.Res)
	}

	closureResJSON, err := json.Marshal(res.Res.Params[0])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal closure response: %w", err)
	}

	var closureRes ChannelClosureRes
	if err := json.Unmarshal(closureResJSON, &closureRes); err != nil {
		return nil, fmt.Errorf("failed to parse channels: %w", err)
	}

	return &closureRes, nil
}

type ChannelResizeRes struct {
	ChannelID   string          `json:"channel_id"`
	Intent      uint8           `json:"intent"`
	Version     uint64          `json:"version"`
	StateData   string          `json:"state_data"`
	Allocations []AllocationRes `json:"allocations"`
	StateHash   string          `json:"state_hash"`
	Signature   SignatureRes    `json:"server_signature"`
}

func (c *ClearnodeClient) RequestChannelResize(walletAddress common.Address, channelID string, allocateAmount, resizeAmount decimal.Decimal) (*ChannelResizeRes, error) {
	if c.signer == nil {
		return nil, fmt.Errorf("client not authenticated")
	}

	params := map[string]any{
		"funds_destination": walletAddress.Hex(),
		"channel_id":        channelID,
		"allocate_amount":   allocateAmount,
		"resize_amount":     resizeAmount,
	}

	res, err := c.request("resize_channel", nil, params)
	if err != nil {
		return nil, fmt.Errorf("failed to request channel resize: %w", err)
	}
	if res.Res.Method != "resize_channel" || len(res.Res.Params) < 1 {
		return nil, fmt.Errorf("unexpected response to resize_channel: %v", res.Res)
	}

	resizeResJSON, err := json.Marshal(res.Res.Params[0])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal closure response: %w", err)
	}

	var resizeRes ChannelResizeRes
	if err := json.Unmarshal(resizeResJSON, &resizeRes); err != nil {
		return nil, fmt.Errorf("failed to parse channels: %w", err)
	}

	return &resizeRes, nil
}

type TransferReq struct {
	Destination        string               `json:"destination"`
	DestinationUserTag string               `json:"destination_user_tag"`
	Allocations        []TransferAllocation `json:"allocations"`
}

type TransferAllocation struct {
	AssetSymbol string          `json:"asset"`
	Amount      decimal.Decimal `json:"amount"`
}

type TransactionResponse struct {
	Id             uint            `json:"id"`
	TxType         string          `json:"tx_type"`
	FromAccount    string          `json:"from_account"`
	FromAccountTag string          `json:"from_account_tag,omitempty"` // Optional tag for the source account
	ToAccount      string          `json:"to_account"`
	ToAccountTag   string          `json:"to_account_tag,omitempty"` // Optional tag for the destination account
	Asset          string          `json:"asset"`
	Amount         decimal.Decimal `json:"amount"`
	CreatedAt      time.Time       `json:"created_at"`
}

func (c *ClearnodeClient) Transfer(destinationTag, assetSymbol string, amount decimal.Decimal) (*TransactionResponse, error) {
	if c.signer == nil {
		return nil, fmt.Errorf("client not authenticated")
	}

	params := TransferReq{
		DestinationUserTag: destinationTag,
		Allocations: []TransferAllocation{
			{
				AssetSymbol: assetSymbol,
				Amount:      amount,
			},
		},
	}

	res, err := c.request("transfer", nil, params)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer: %w", err)
	}
	if res.Res.Method != "transfer" || len(res.Res.Params) < 1 {
		return nil, fmt.Errorf("unexpected response to transfer: %v", res.Res)
	}

	resizeResJSON, err := json.Marshal(res.Res.Params[0])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal closure response: %w", err)
	}

	var txs []TransactionResponse
	if err := json.Unmarshal(resizeResJSON, &txs); err != nil {
		return nil, fmt.Errorf("failed to parse channels: %w", err)
	}
	if len(txs) == 0 {
		return nil, fmt.Errorf("no transactions returned from transfer request")
	}

	return &txs[0], nil
}

func (c *ClearnodeClient) readMessages() {
	defer c.exit() // Ensure exit channel is closed when done

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if _, ok := err.(net.Error); ok {
			fmt.Println("Websocket connection timeout")
			return
		} else if err != nil {
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
	defer c.exit() // Ensure exit channel is closed when done

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

	reqID := uint64(time.Now().UnixMilli())
	rpcData := RPCData{
		RequestID: reqID,
		Method:    method,
		Params:    params,
		Timestamp: uint64(time.Now().UnixMilli()),
	}

	if len(sigs) == 0 && c.signer != nil {
		sig, err := signRPCData(c.signer, rpcData)
		if err != nil {
			return nil, fmt.Errorf("error signing RPC data: %w", err)
		}
		sigs = []string{hexutil.Encode(sig)}
	}

	req := RPCRequest{
		Req: rpcData,
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

func (c *ClearnodeClient) WaitCh() <-chan struct{} {
	return c.exitCh
}

func (c *ClearnodeClient) exit() {
	close(c.exitCh)
}
