package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/erc7824/nitrolite/clearnode/nitrolite"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// AppDefinition represents the definition of an application on the ledger
type AppDefinition struct {
	Protocol           string   `json:"protocol"`
	ParticipantWallets []string `json:"participants"` // Participants from channels with broker.
	Weights            []uint64 `json:"weights"`      // Signature weight for each participant.
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

// AppSessionResponse represents response data for application operations
type AppSessionResponse struct {
	AppSessionID       string   `json:"app_session_id"`
	Status             string   `json:"status"`
	ParticipantWallets []string `json:"participants,omitempty"`
	Protocol           string   `json:"protocol,omitempty"`
	Challenge          uint64   `json:"challenge,omitempty"`
	Weights            []int64  `json:"weights,omitempty"`
	Quorum             uint64   `json:"quorum,omitempty"`
	Version            uint64   `json:"version,omitempty"`
	Nonce              uint64   `json:"nonce,omitempty"`
}

// ResizeChannelParams represents parameters needed for resizing a channel
type ResizeChannelParams struct {
	ChannelID        string   `json:"channel_id"                          validate:"required"`
	AllocateAmount   *big.Int `json:"allocate_amount,omitempty"           validate:"required_without=ResizeAmount"`
	ResizeAmount     *big.Int `json:"resize_amount,omitempty"             validate:"required_without=AllocateAmount"`
	FundsDestination string   `json:"funds_destination"                   validate:"required"`
}

// ResizeChannelResponse represents the response for resizing a channel
type ResizeChannelResponse struct {
	ChannelID   string       `json:"channel_id"`
	StateData   string       `json:"state_data"`
	Intent      uint8        `json:"intent"`
	Version     uint64       `json:"version"`
	Allocations []Allocation `json:"allocations"`
	StateHash   string       `json:"state_hash"`
	Signature   Signature    `json:"server_signature"`
}

// Allocation represents a token allocation for a specific participant
type Allocation struct {
	Participant  string   `json:"destination"`
	TokenAddress string   `json:"token"`
	Amount       *big.Int `json:"amount,string"`
}

type ResizeChannelSignData struct {
	RequestID uint64
	Method    string
	Params    []ResizeChannelParams
	Timestamp uint64
}

type LedgerEntryResponse struct {
	ID          uint            `json:"id"`
	AccountID   string          `json:"account_id"`
	AccountType AccountType     `json:"account_type"`
	Asset       string          `json:"asset"`
	Participant string          `json:"participant"`
	Credit      decimal.Decimal `json:"credit"`
	Debit       decimal.Decimal `json:"debit"`
	CreatedAt   time.Time       `json:"created_at"`
}

func (r ResizeChannelSignData) MarshalJSON() ([]byte, error) {
	arr := []interface{}{r.RequestID, r.Method, r.Params, r.Timestamp}
	return json.Marshal(arr)
}

// CloseChannelParams represents parameters needed for channel closure
type CloseChannelParams struct {
	ChannelID        string `json:"channel_id"`
	FundsDestination string `json:"funds_destination"`
}

// CloseChannelResponse represents the response for closing a channel
type CloseChannelResponse struct {
	ChannelID        string       `json:"channel_id"`
	Intent           uint8        `json:"intent"`
	Version          uint64       `json:"version"`
	StateData        string       `json:"state_data"`
	FinalAllocations []Allocation `json:"allocations"`
	StateHash        string       `json:"state_hash"`
	Signature        Signature    `json:"server_signature"`
}

// ChannelResponse represents a channel's details in the response
type ChannelResponse struct {
	ChannelID   string        `json:"channel_id"`
	Participant string        `json:"participant"`
	Status      ChannelStatus `json:"status"`
	Token       string        `json:"token"`
	// Total amount in the channel (user + broker)
	Amount      *big.Int `json:"amount"`
	ChainID     uint32   `json:"chain_id"`
	Adjudicator string   `json:"adjudicator"`
	Challenge   uint64   `json:"challenge"`
	Nonce       uint64   `json:"nonce"`
	Version     uint64   `json:"version"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

type Signature struct {
	V uint8  `json:"v,string"`
	R string `json:"r,string"`
	S string `json:"s,string"`
}

type Balance struct {
	Asset  string          `json:"asset"`
	Amount decimal.Decimal `json:"amount"`
}

// NetworkInfo represents information about a supported network
type NetworkInfo struct {
	Name           string `json:"name"`
	ChainID        uint32 `json:"chain_id"`
	CustodyAddress string `json:"custody_address"`
}

// BrokerConfig represents the broker configuration information
type BrokerConfig struct {
	BrokerAddress string        `json:"broker_address"`
	Networks      []NetworkInfo `json:"networks"`
}

// RPCEntry represents an RPC record from history.
type RPCEntry struct {
	ID        uint     `json:"id"`
	Sender    string   `json:"sender"`
	ReqID     uint64   `json:"req_id"`
	Method    string   `json:"method"`
	Params    string   `json:"params"`
	Timestamp uint64   `json:"timestamp"`
	ReqSig    []string `json:"req_sig"`
	Result    string   `json:"response"`
	ResSig    []string `json:"res_sig"`
}

// HandleGetConfig returns the broker configuration
func HandleGetConfig(rpc *RPCMessage, config *Config, signer *Signer) (*RPCMessage, error) {
	supportedNetworks := []NetworkInfo{}

	// Populate the supported networks from the config
	for name, networkConfig := range config.networks {
		supportedNetworks = append(supportedNetworks, NetworkInfo{
			Name:           name,
			ChainID:        networkConfig.ChainID,
			CustodyAddress: networkConfig.CustodyAddress,
		})
	}

	brokerConfig := BrokerConfig{
		BrokerAddress: signer.GetAddress().Hex(),
		Networks:      supportedNetworks,
	}

	rpcResponse := CreateResponse(rpc.Req.RequestID, "get_config", []any{brokerConfig}, time.Now())
	return rpcResponse, nil
}

// HandlePing responds to a ping request with a pong response in RPC format
func HandlePing(rpc *RPCMessage) (*RPCMessage, error) {
	return CreateResponse(rpc.Req.RequestID, "pong", []any{}, time.Now()), nil
}

// HandleGetLedgerBalances returns a list of participants and their balances for a ledger account
func HandleGetLedgerBalances(rpc *RPCMessage, walletAddress string, db *gorm.DB) (*RPCMessage, error) {
	var account string

	if len(rpc.Req.Params) > 0 {
		paramsJSON, err := json.Marshal(rpc.Req.Params[0])
		if err == nil {
			var params map[string]string
			if err := json.Unmarshal(paramsJSON, &params); err == nil {
				account = params["participant"]
				id, ok := params["account_id"]
				if ok {
					account = id
				}
			}
		}
	}

	if account == "" {
		account = walletAddress
	}

	ledger := GetWalletLedger(db, walletAddress)
	balances, err := ledger.GetBalances(account)
	if err != nil {
		return nil, fmt.Errorf("failed to find account: %w", err)
	}

	rpcResponse := CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{balances}, time.Now())
	return rpcResponse, nil
}

func HandleGetLedgerEntries(rpc *RPCMessage, walletAddress string, db *gorm.DB) (*RPCMessage, error) {
	var accountID string
	var asset string

	if len(rpc.Req.Params) > 0 {
		paramsJSON, err := json.Marshal(rpc.Req.Params[0])
		if err == nil {
			var params map[string]string
			if err := json.Unmarshal(paramsJSON, &params); err == nil {
				accountID = params["account_id"]
				asset = params["asset"]
			}
		}
	}

	if accountID == "" {
		return nil, errors.New("missing account_id")
	}

	ledger := GetWalletLedger(db, walletAddress)

	entries, err := ledger.GetEntries(accountID, asset)
	if err != nil {
		return nil, fmt.Errorf("failed to find account: %w", err)
	}

	response := make([]LedgerEntryResponse, len(entries))
	for i, entry := range entries {
		response[i] = LedgerEntryResponse{
			ID:          entry.ID,
			AccountID:   entry.AccountID,
			AccountType: entry.AccountType,
			Asset:       entry.AssetSymbol,
			Participant: entry.Wallet,
			Credit:      entry.Credit,
			Debit:       entry.Debit,
			CreatedAt:   entry.CreatedAt,
		}
	}

	rpcResponse := CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{response}, time.Now())
	return rpcResponse, nil
}

// HandleCreateApplication creates a virtual application between participants
func HandleCreateApplication(rpc *RPCMessage, db *gorm.DB) (*RPCMessage, error) {
	if len(rpc.Req.Params) < 1 {
		return nil, errors.New("missing parameters")
	}

	var createApp CreateAppSessionParams
	paramsJSON, err := json.Marshal(rpc.Req.Params[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	if err := json.Unmarshal(paramsJSON, &createApp); err != nil {
		return nil, fmt.Errorf("invalid parameters format: %w", err)
	}

	if len(createApp.Definition.ParticipantWallets) < 2 {
		return nil, errors.New("invalid number of participants")
	}

	// Allocation should be specified for each participant even if it is zero.
	if len(createApp.Allocations) != len(createApp.Definition.ParticipantWallets) {
		return nil, errors.New("number of allocations must be equal to participants")
	}

	if len(createApp.Definition.Weights) != len(createApp.Definition.ParticipantWallets) {
		return nil, errors.New("number of weights must be equal to participants")
	}

	if createApp.Definition.Nonce == 0 {
		return nil, errors.New("invalid nonce")
	}

	var participantsAddresses []common.Address
	for _, participant := range createApp.Definition.ParticipantWallets {
		participantsAddresses = append(participantsAddresses, common.HexToAddress(participant))
	}

	req := CreateAppSignData{
		RequestID: rpc.Req.RequestID,
		Method:    rpc.Req.Method,
		Params:    []CreateAppSessionParams{{Definition: createApp.Definition, Allocations: createApp.Allocations}},
		Timestamp: rpc.Req.Timestamp,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, errors.New("error serializing message")
	}

	recoveredAddresses := map[string]bool{}
	for _, sig := range rpc.Sig {
		addr, err := RecoverAddress(reqBytes, sig)
		if err != nil {
			return nil, errors.New("invalid signature")
		}

		walletAddress, err := GetWalletBySigner(db, addr)
		if err != nil {
			continue
		}
		if walletAddress != "" {
			recoveredAddresses[walletAddress] = true
		} else {
			recoveredAddresses[addr] = true
		}
	}

	// Generate a unique ID for the virtual application
	b, _ := json.Marshal(createApp.Definition)
	appSessionID := crypto.Keccak256Hash(b)

	// Use a transaction to ensure atomicity for the entire operation
	err = db.Transaction(func(tx *gorm.DB) error {
		for _, allocation := range createApp.Allocations {
			if allocation.Amount.IsNegative() {
				return errors.New("invalid allocation")
			}
			if allocation.Amount.IsPositive() {
				if !recoveredAddresses[allocation.ParticipantWallet] {
					return fmt.Errorf("missing signature for participant %s", allocation.ParticipantWallet)
				}
			}

			participantWalletLedger := GetWalletLedger(tx, allocation.ParticipantWallet)
			balance, err := participantWalletLedger.Balance(allocation.ParticipantWallet, allocation.AssetSymbol)
			if err != nil {
				return fmt.Errorf("failed to check participant balance: %w", err)
			}
			if allocation.Amount.GreaterThan(balance) {
				return errors.New("insufficient funds")
			}
			if err := participantWalletLedger.Record(allocation.ParticipantWallet, allocation.AssetSymbol, allocation.Amount.Neg()); err != nil {
				return fmt.Errorf("failed to transfer funds from participant: %w", err)
			}
			if err := participantWalletLedger.Record(appSessionID.Hex(), allocation.AssetSymbol, allocation.Amount); err != nil {
				return fmt.Errorf("failed to transfer funds to virtual app: %w", err)
			}
		}

		weights := pq.Int64Array{}
		for _, v := range createApp.Definition.Weights {
			weights = append(weights, int64(v))
		}

		// Record the virtual app creation in state
		appSession := &AppSession{
			Protocol:           createApp.Definition.Protocol,
			SessionID:          appSessionID.Hex(),
			ParticipantWallets: createApp.Definition.ParticipantWallets,
			Status:             ChannelStatusOpen,
			Challenge:          createApp.Definition.Challenge,
			Weights:            weights,
			Quorum:             createApp.Definition.Quorum,
			Nonce:              createApp.Definition.Nonce,
			Version:            rpc.Req.Timestamp,
		}

		if err := tx.Create(appSession).Error; err != nil {
			return fmt.Errorf("failed to record virtual app: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	response := &AppSessionResponse{
		AppSessionID: appSessionID.Hex(),
		Status:       string(ChannelStatusOpen),
	}

	rpcResponse := CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{response}, time.Now())
	return rpcResponse, nil
}

// HandleCloseApplication closes a virtual app session and redistributes funds to participants
func HandleCloseApplication(rpc *RPCMessage, db *gorm.DB) (*RPCMessage, error) {
	if len(rpc.Req.Params) == 0 {
		return nil, errors.New("missing parameters")
	}

	var params CloseAppSessionParams
	paramsJSON, err := json.Marshal(rpc.Req.Params[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	if err := json.Unmarshal(paramsJSON, &params); err != nil {
		return nil, fmt.Errorf("invalid parameters format: %w", err)
	}

	if params.AppSessionID == "" || len(params.Allocations) == 0 {
		return nil, errors.New("missing required parameters: app_id or allocations")
	}

	assets := map[string]struct{}{}
	for _, a := range params.Allocations {
		if a.ParticipantWallet == "" || a.AssetSymbol == "" || a.Amount.IsNegative() {
			return nil, errors.New("invalid allocation row")
		}
		assets[a.AssetSymbol] = struct{}{}
	}

	req := CloseAppSignData{
		RequestID: rpc.Req.RequestID,
		Method:    rpc.Req.Method,
		Params:    []CloseAppSessionParams{{AppSessionID: params.AppSessionID, Allocations: params.Allocations}},
		Timestamp: rpc.Req.Timestamp,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, errors.New("error serializing message")
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		var appSession AppSession
		if err := tx.Where("session_id = ? AND status = ?", params.AppSessionID, ChannelStatusOpen).Order("nonce DESC").
			First(&appSession).Error; err != nil {
			return fmt.Errorf("virtual app not found or not open: %w", err)
		}

		participantWeights := map[string]int64{}
		for i, addr := range appSession.ParticipantWallets {
			participantWeights[addr] = appSession.Weights[i]
		}

		recoveredAddresses := map[string]bool{}
		var totalWeight int64
		for _, sigHex := range rpc.Sig {
			recovered, err := RecoverAddress(reqBytes, sigHex)
			if err != nil {
				return err
			}
			recoveredAddresses[recovered] = true
		}

		for address := range recoveredAddresses {
			addr := address
			walletAddress, _ := GetWalletBySigner(db, address)
			if walletAddress != "" {
				addr = walletAddress
			}

			weight, ok := participantWeights[addr]
			if !ok {
				return fmt.Errorf("signature from unknown participant wallet %s", addr)
			}
			if weight <= 0 {
				return fmt.Errorf("zero weight for signer %s", addr)
			}
			totalWeight += weight
		}

		if totalWeight < int64(appSession.Quorum) {
			return fmt.Errorf("quorum not met: %d / %d", totalWeight, appSession.Quorum)
		}

		appSessionBalance := map[string]decimal.Decimal{}
		for _, p := range appSession.ParticipantWallets {
			ledger := GetWalletLedger(tx, p)
			for asset := range assets {
				bal, err := ledger.Balance(appSession.SessionID, asset)
				if err != nil {
					return fmt.Errorf("failed to read balance for %s:%s: %w", p, asset, err)
				}
				appSessionBalance[asset] = appSessionBalance[asset].Add(bal)
			}
		}

		allocationSum := map[string]decimal.Decimal{}

		for _, alloc := range params.Allocations {
			if _, ok := participantWeights[alloc.ParticipantWallet]; !ok {
				return fmt.Errorf("allocation to non-participant %s", alloc.ParticipantWallet)
			}

			ledger := GetWalletLedger(tx, alloc.ParticipantWallet)
			balance, err := ledger.Balance(appSession.SessionID, alloc.AssetSymbol)
			if err != nil {
				return fmt.Errorf("failed to get participant balance: %w", err)
			}

			// Debit session, credit participant
			if err := ledger.Record(appSession.SessionID, alloc.AssetSymbol, balance.Neg()); err != nil {
				return fmt.Errorf("failed to debit session: %w", err)
			}
			if err := ledger.Record(alloc.ParticipantWallet, alloc.AssetSymbol, alloc.Amount); err != nil {
				return fmt.Errorf("failed to credit participant: %w", err)
			}

			allocationSum[alloc.AssetSymbol] = allocationSum[alloc.AssetSymbol].Add(alloc.Amount)
		}

		for asset, bal := range appSessionBalance {
			if alloc, ok := allocationSum[asset]; !ok || !bal.Equal(alloc) {
				return fmt.Errorf("asset %s not fully redistributed", asset)
			}
		}
		for asset := range allocationSum {
			if _, ok := appSessionBalance[asset]; !ok {
				return fmt.Errorf("allocation references unknown asset %s", asset)
			}
		}

		return tx.Model(&appSession).Updates(map[string]any{
			"status": ChannelStatusClosed,
		}).Error
	})

	if err != nil {
		return nil, err
	}

	response := &AppSessionResponse{
		AppSessionID: params.AppSessionID,
		Status:       string(ChannelStatusClosed),
	}

	rpcResponse := CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{response}, time.Now())
	return rpcResponse, nil
}

// HandleGetAppDefinition returns the application definition for a ledger account
func HandleGetAppDefinition(rpc *RPCMessage, db *gorm.DB) (*RPCMessage, error) {
	var sessionID string

	if len(rpc.Req.Params) > 0 {
		paramsJSON, err := json.Marshal(rpc.Req.Params[0])
		if err == nil {
			var params map[string]string
			if err := json.Unmarshal(paramsJSON, &params); err == nil {
				sessionID = params["app_session_id"]
			}
		}
	}

	if sessionID == "" {
		return nil, errors.New("missing account ID")
	}

	var vApp AppSession
	if err := db.Where("session_id = ?", sessionID).First(&vApp).Error; err != nil {
		return nil, fmt.Errorf("failed to find application: %w", err)
	}

	appDef := AppDefinition{
		Protocol:           vApp.Protocol,
		ParticipantWallets: vApp.ParticipantWallets,
		Weights:            make([]uint64, len(vApp.ParticipantWallets)), // Default weights to 0 for now
		Quorum:             vApp.Quorum,                                  // Default quorum to 100 for now
		Challenge:          vApp.Challenge,
		Nonce:              vApp.Nonce,
	}

	for i := range vApp.Weights {
		appDef.Weights[i] = uint64(vApp.Weights[i])
	}

	rpcResponse := CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{appDef}, time.Now())
	return rpcResponse, nil
}

func HandleGetAppSessions(rpc *RPCMessage, db *gorm.DB) (*RPCMessage, error) {
	var participant string
	var status string

	if len(rpc.Req.Params) > 0 {
		paramsJSON, err := json.Marshal(rpc.Req.Params[0])
		if err == nil {
			var params map[string]string
			if err := json.Unmarshal(paramsJSON, &params); err == nil {
				participant = params["participant"]
				status = params["status"]
			}
		}
	}

	if participant == "" {
		return nil, errors.New("missing participant")
	}

	sessions, err := getAppSessions(db, participant, status)
	if err != nil {
		return nil, fmt.Errorf("failed to find application sessions: %w", err)
	}
	response := make([]AppSessionResponse, len(sessions))
	for i, session := range sessions {
		response[i] = AppSessionResponse{
			AppSessionID:       session.SessionID,
			Status:             string(session.Status),
			ParticipantWallets: session.ParticipantWallets,
			Protocol:           session.Protocol,
			Challenge:          session.Challenge,
			Weights:            session.Weights,
			Quorum:             session.Quorum,
			Version:            session.Version,
			Nonce:              session.Nonce,
		}
	}

	rpcResponse := CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{response}, time.Now())
	return rpcResponse, nil
}

// HandleResizeChannel processes a request to resize a payment channel
func HandleResizeChannel(rpc *RPCMessage, db *gorm.DB, signer *Signer) (*RPCMessage, error) {
	if len(rpc.Req.Params) < 1 {
		return nil, errors.New("missing parameters")
	}

	var params ResizeChannelParams
	paramsJSON, err := json.Marshal(rpc.Req.Params[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	if err := json.Unmarshal(paramsJSON, &params); err != nil {
		return nil, fmt.Errorf("invalid parameters format: %w", err)
	}

	if err := validate.Struct(&params); err != nil {
		return nil, err
	}

	channel, err := GetChannelByID(db, params.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to find channel: %w", err)
	}
	if channel == nil {
		return nil, errors.New("channel not found")
	}

	req := ResizeChannelSignData{
		RequestID: rpc.Req.RequestID,
		Method:    rpc.Req.Method,
		Params:    []ResizeChannelParams{{ChannelID: params.ChannelID, ResizeAmount: params.ResizeAmount, AllocateAmount: params.AllocateAmount, FundsDestination: params.FundsDestination}},
		Timestamp: rpc.Req.Timestamp,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, errors.New("error serializing message")
	}

	recoveredAddress, err := RecoverAddress(reqBytes, rpc.Sig[0])
	if err != nil {
		return nil, err
	}

	walletAddress, _ := GetWalletBySigner(db, recoveredAddress)
	if walletAddress != "" {
		recoveredAddress = walletAddress
	}

	if !strings.EqualFold(recoveredAddress, channel.Wallet) {
		return nil, errors.New("invalid signature")
	}

	asset, err := GetAssetByToken(db, channel.Token, channel.ChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to find asset: %w", err)
	}
	if asset == nil {
		return nil, fmt.Errorf("asset not found: %s", channel.Token)
	}

	if params.ResizeAmount == nil {
		params.ResizeAmount = big.NewInt(0)
	}

	if params.AllocateAmount == nil {
		params.AllocateAmount = big.NewInt(0)
	}

	ledger := GetWalletLedger(db, channel.Wallet)
	balance, err := ledger.Balance(channel.Wallet, asset.Symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to check participant A balance: %w", err)
	}

	rawBalance := balance.Shift(int32(asset.Decimals)).BigInt()

	newChannelAmount := new(big.Int).Add(new(big.Int).SetUint64(channel.Amount), params.AllocateAmount)
	if rawBalance.Cmp(newChannelAmount) < 0 {
		return nil, errors.New("insufficient unified balance")
	}

	newChannelAmount.Add(newChannelAmount, params.ResizeAmount)
	if newChannelAmount.Cmp(big.NewInt(0)) < 0 {
		return nil, errors.New("new channel amount must be positive")
	}
	allocations := []nitrolite.Allocation{
		{
			Destination: common.HexToAddress(params.FundsDestination),
			Token:       common.HexToAddress(channel.Token),
			Amount:      newChannelAmount,
		},
		{
			Destination: signer.GetAddress(),
			Token:       common.HexToAddress(channel.Token),
			Amount:      big.NewInt(0),
		},
	}

	resizeAmounts := []*big.Int{new(big.Int).Neg(params.ResizeAmount), params.AllocateAmount}

	intentionType, err := abi.NewType("int256[]", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create ABI type for intentions: %w", err)
	}

	intentionsArgs := abi.Arguments{
		{Type: intentionType},
	}

	encodedIntentions, err := intentionsArgs.Pack(resizeAmounts)
	if err != nil {
		return nil, fmt.Errorf("failed to pack intentions: %w", err)
	}

	// Encode the channel ID and state for signing
	channelID := common.HexToHash(channel.ChannelID)
	encodedState, err := nitrolite.EncodeState(channelID, nitrolite.IntentRESIZE, big.NewInt(int64(channel.Version)+1), encodedIntentions, allocations)
	if err != nil {
		return nil, fmt.Errorf("failed to encode state hash: %w", err)
	}

	// Generate state hash and sign it
	stateHash := crypto.Keccak256Hash(encodedState).Hex()
	sig, err := signer.NitroSign(encodedState)
	if err != nil {
		return nil, fmt.Errorf("failed to sign state: %w", err)
	}

	response := ResizeChannelResponse{
		ChannelID: channel.ChannelID,
		Intent:    uint8(nitrolite.IntentRESIZE),
		Version:   channel.Version + 1,
		StateData: hexutil.Encode(encodedIntentions),
		StateHash: stateHash,
		Signature: Signature{
			V: sig.V,
			R: hexutil.Encode(sig.R[:]),
			S: hexutil.Encode(sig.S[:]),
		},
	}

	for _, alloc := range allocations {
		response.Allocations = append(response.Allocations, Allocation{
			Participant:  alloc.Destination.Hex(),
			TokenAddress: alloc.Token.Hex(),
			Amount:       alloc.Amount,
		})
	}

	rpcResponse := CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{response}, time.Now())
	return rpcResponse, nil
}

// HandleCloseChannel processes a request to close a payment channel
func HandleCloseChannel(rpc *RPCMessage, db *gorm.DB, signer *Signer) (*RPCMessage, error) {
	if len(rpc.Req.Params) < 1 {
		return nil, errors.New("missing parameters")
	}

	var params CloseChannelParams
	paramsJSON, err := json.Marshal(rpc.Req.Params[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	if err := json.Unmarshal(paramsJSON, &params); err != nil {
		return nil, fmt.Errorf("invalid parameters format: %w", err)
	}

	channel, err := GetChannelByID(db, params.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to find channel: %w", err)
	}
	if channel == nil {
		return nil, errors.New("channel not found")
	}

	reqBytes, err := json.Marshal(rpc.Req)
	if err != nil {
		return nil, errors.New("error serializing message")
	}

	recoveredAddress, err := RecoverAddress(reqBytes, rpc.Sig[0])
	if err != nil {
		return nil, err
	}

	walletAddress, _ := GetWalletBySigner(db, recoveredAddress)
	if walletAddress != "" {
		recoveredAddress = walletAddress
	}

	if !strings.EqualFold(recoveredAddress, channel.Wallet) {
		return nil, errors.New("invalid signature")
	}

	asset, err := GetAssetByToken(db, channel.Token, channel.ChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to find asset: %w", err)
	}
	if asset == nil {
		return nil, fmt.Errorf("asset not found: %s", channel.Token)
	}

	ledger := GetWalletLedger(db, channel.Wallet)
	balance, err := ledger.Balance(channel.Wallet, asset.Symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to check participant A balance: %w", err)
	}

	if balance.IsNegative() {
		return nil, errors.New("insufficient funds for participant: " + channel.Token)
	}

	rawBalance := balance.Shift(int32(asset.Decimals)).BigInt()

	channelAmount := new(big.Int).SetUint64(channel.Amount)
	if channelAmount.Cmp(rawBalance) < 0 {
		return nil, errors.New("resize this channel first")
	}

	allocations := []nitrolite.Allocation{
		{
			Destination: common.HexToAddress(params.FundsDestination),
			Token:       common.HexToAddress(channel.Token),
			Amount:      rawBalance,
		},
		{
			Destination: signer.GetAddress(),
			Token:       common.HexToAddress(channel.Token),
			Amount:      new(big.Int).Sub(channelAmount, rawBalance), // Broker receives the remaining amount
		},
	}

	stateDataStr := "0x"
	stateData, err := hexutil.Decode(stateDataStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode state data: %w", err)
	}

	channelID := common.HexToHash(channel.ChannelID)
	encodedState, err := nitrolite.EncodeState(channelID, nitrolite.IntentFINALIZE, big.NewInt(int64(channel.Version)+1), stateData, allocations)
	if err != nil {
		return nil, fmt.Errorf("failed to encode state hash: %w", err)
	}

	stateHash := crypto.Keccak256Hash(encodedState).Hex()
	sig, err := signer.NitroSign(encodedState)
	if err != nil {
		return nil, fmt.Errorf("failed to sign state: %w", err)
	}

	response := CloseChannelResponse{
		ChannelID: channel.ChannelID,
		Intent:    uint8(nitrolite.IntentFINALIZE),
		Version:   channel.Version + 1,
		StateData: stateDataStr,
		StateHash: stateHash,
		Signature: Signature{
			V: sig.V,
			R: hexutil.Encode(sig.R[:]),
			S: hexutil.Encode(sig.S[:]),
		},
	}

	for _, alloc := range allocations {
		response.FinalAllocations = append(response.FinalAllocations, Allocation{
			Participant:  alloc.Destination.Hex(),
			TokenAddress: alloc.Token.Hex(),
			Amount:       alloc.Amount,
		})
	}

	rpcResponse := CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{response}, time.Now())
	return rpcResponse, nil
}

// HandleGetChannels returns a list of channels for a given account
// TODO: add filters, pagination, etc.
func HandleGetChannels(rpc *RPCMessage, db *gorm.DB) (*RPCMessage, error) {
	var participant string
	var status string

	if len(rpc.Req.Params) > 0 {
		paramsJSON, err := json.Marshal(rpc.Req.Params[0])
		if err == nil {
			var params map[string]string
			if err := json.Unmarshal(paramsJSON, &params); err == nil {
				participant = params["participant"]
				status = params["status"]
			}
		}
	}

	if participant == "" {
		return nil, errors.New("missing participant participant")
	}

	channels, err := getChannelsByWallet(db, participant, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}

	var channelResponses []ChannelResponse
	for _, channel := range channels {
		channelResponses = append(channelResponses, ChannelResponse{
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
		})
	}

	rpcResponse := CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{channelResponses}, time.Now())
	return rpcResponse, nil
}

func HandleGetRPCHistory(participant string, rpc *RPCMessage, store *RPCStore) (*RPCMessage, error) {
	if participant == "" {
		return nil, errors.New("missing participant parameter")
	}

	var rpcHistory []RPCRecord
	if err := store.db.Where("sender = ?", participant).Order("timestamp DESC").Find(&rpcHistory).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve RPC history: %w", err)
	}

	response := make([]RPCEntry, 0, len(rpcHistory))
	for _, record := range rpcHistory {
		response = append(response, RPCEntry{
			ID:        record.ID,
			Sender:    record.Sender,
			ReqID:     record.ReqID,
			Method:    record.Method,
			Params:    string(record.Params),
			Timestamp: record.Timestamp,
			ReqSig:    record.ReqSig,
			ResSig:    record.ResSig,
			Result:    string(record.Response),
		})
	}

	rpcResponse := CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{response}, time.Now())
	return rpcResponse, nil
}

// AssetResponse represents an asset in the response
type AssetResponse struct {
	Token    string `json:"token"`    // Token address
	ChainID  uint32 `json:"chain_id"` // Chain ID
	Symbol   string `json:"symbol"`   // Symbol of the asset (e.g., "usdc")
	Decimals uint8  `json:"decimals"` // Number of decimals for the asset
}

// HandleGetAssets returns all supported assets
func HandleGetAssets(rpc *RPCMessage, db *gorm.DB) (*RPCMessage, error) {
	var chainID *uint32

	if len(rpc.Req.Params) > 0 {
		paramsJSON, err := json.Marshal(rpc.Req.Params[0])
		if err == nil {
			var params map[string]interface{}
			if err := json.Unmarshal(paramsJSON, &params); err == nil {
				if chainIDValue, ok := params["chain_id"]; ok {
					// Convert the chain_id to uint32
					if chainIDFloat, ok := chainIDValue.(float64); ok {
						chainIDUint := uint32(chainIDFloat)
						chainID = &chainIDUint
					}
				}
			}
		}
	}

	assets, err := GetAllAssets(db, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve assets: %w", err)
	}

	response := make([]AssetResponse, 0, len(assets))
	for _, asset := range assets {
		response = append(response, AssetResponse{
			Token:    asset.Token,
			ChainID:  asset.ChainID,
			Symbol:   asset.Symbol,
			Decimals: asset.Decimals,
		})
	}

	rpcResponse := CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{response}, time.Now())
	return rpcResponse, nil
}
