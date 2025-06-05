// main.go

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
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// AppDefinition represents the definition of an application on the ledger
type AppDefinition struct {
	Protocol           string   `json:"protocol"`
	ParticipantWallets []string `json:"participants"`
	Weights            []int64  `json:"weights"` // Signature weight for each participant.
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

// AppSessionResponse represents response data for application operations
type AppSessionResponse struct {
	AppSessionID       string   `json:"app_session_id"`
	Status             string   `json:"status"`
	ParticipantWallets []string `json:"participants"`
	Protocol           string   `json:"protocol"`
	Challenge          uint64   `json:"challenge"`
	Weights            []int64  `json:"weights"`
	Quorum             uint64   `json:"quorum"`
	Version            uint64   `json:"version"`
	Nonce              uint64   `json:"nonce"`
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

type Allocation struct {
	Participant  string   `json:"destination"`
	TokenAddress string   `json:"token"`
	Amount       *big.Int `json:"amount,string"`
}

// LedgerEntryResponse is used by HandleGetLedgerEntries
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
	Wallet      string        `json:"wallet"`
	Amount      *big.Int      `json:"amount"` // Total amount in the channel (user + broker)
	ChainID     uint32        `json:"chain_id"`
	Adjudicator string        `json:"adjudicator"`
	Challenge   uint64        `json:"challenge"`
	Nonce       uint64        `json:"nonce"`
	Version     uint64        `json:"version"`
	CreatedAt   string        `json:"created_at"`
	UpdatedAt   string        `json:"updated_at"`
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
	Name               string `json:"name"`
	ChainID            uint32 `json:"chain_id"`
	CustodyAddress     string `json:"custody_address"`
	AdjudicatorAddress string `json:"adjudicator_address"`
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

// AssetResponse represents an asset in the response
type AssetResponse struct {
	Token    string `json:"token"`
	ChainID  uint32 `json:"chain_id"`
	Symbol   string `json:"symbol"`
	Decimals uint8  `json:"decimals"`
}

// HandleGetConfig returns the broker configuration
func HandleGetConfig(rpc *RPCMessage, config *Config, signer *Signer) (*RPCMessage, error) {
	supportedNetworks := []NetworkInfo{}

	for name, networkConfig := range config.networks {
		supportedNetworks = append(supportedNetworks, NetworkInfo{
			Name:               name,
			ChainID:            networkConfig.ChainID,
			CustodyAddress:     networkConfig.CustodyAddress,
			AdjudicatorAddress: networkConfig.AdjudicatorAddress,
		})
	}

	brokerConfig := BrokerConfig{
		BrokerAddress: signer.GetAddress().Hex(),
		Networks:      supportedNetworks,
	}

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{brokerConfig}), nil
}

// HandlePing responds to a ping request
func HandlePing(rpc *RPCMessage) (*RPCMessage, error) {
	return CreateResponse(rpc.Req.RequestID, "pong", []any{}), nil
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
				if id, ok := params["account_id"]; ok {
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

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{balances}), nil
}

// HandleGetLedgerEntries returns ledger entries for an account
func HandleGetLedgerEntries(rpc *RPCMessage, walletAddress string, db *gorm.DB) (*RPCMessage, error) {
	var accountID, asset, wallet string

	if len(rpc.Req.Params) > 0 {
		paramsJSON, err := json.Marshal(rpc.Req.Params[0])
		if err == nil {
			var params map[string]string
			if err := json.Unmarshal(paramsJSON, &params); err == nil {
				accountID = params["account_id"]
				asset = params["asset"]
				if w, ok := params["wallet"]; ok {
					wallet = w
				}
			}
		}
	}

	if wallet != "" {
		walletAddress = wallet
	}

	ledger := GetWalletLedger(db, walletAddress)
	entries, err := ledger.GetEntries(accountID, asset)
	if err != nil {
		return nil, fmt.Errorf("failed to find ledger entries: %w", err)
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

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{response}), nil
}

// HandleCreateApplication creates a virtual application between participants
func HandleCreateApplication(policy *Policy, rpc *RPCMessage, db *gorm.DB) (*RPCMessage, error) {
	var params CreateAppSessionParams
	if err := parseParams(rpc.Req.Params, &params); err != nil {
		return nil, err
	}
	if len(params.Definition.ParticipantWallets) < 2 {
		return nil, errors.New("invalid number of participants")
	}
	if len(params.Definition.Weights) != len(params.Definition.ParticipantWallets) {
		return nil, errors.New("number of weights must be equal to participants")
	}
	if params.Definition.Nonce == 0 {
		return nil, errors.New("nonce is zero or not provided")
	}

	recoveredAddresses := map[string]bool{}
	for _, sig := range rpc.Sig {
		addr, err := RecoverAddress(rpc.ReqRaw, sig)
		if err != nil {
			return nil, errors.New("invalid signature")
		}

		walletAddress, err := GetWalletBySigner(addr)
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
	b, err := json.Marshal(params.Definition)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app session ID: %w", err)
	}
	appSessionID := crypto.Keccak256Hash(b).Hex()

	err = db.Transaction(func(tx *gorm.DB) error {
		for _, allocation := range params.Allocations {
			if allocation.Amount.IsNegative() {
				return fmt.Errorf("invalid allocation: negative amount")
			}
			if allocation.Amount.IsPositive() && !recoveredAddresses[allocation.ParticipantWallet] {
				return fmt.Errorf("missing signature for participant %s", allocation.ParticipantWallet)
			}

			participantWallet := GetWalletLedger(tx, allocation.ParticipantWallet)
			balance, err := participantWallet.Balance(allocation.ParticipantWallet, allocation.AssetSymbol)
			if err != nil {
				return fmt.Errorf("failed to check participant balance: %w", err)
			}
			if allocation.Amount.GreaterThan(balance) {
				return fmt.Errorf("insufficient funds: %s for asset %s", allocation.ParticipantWallet, allocation.AssetSymbol)
			}
			if err := participantWallet.Record(allocation.ParticipantWallet, allocation.AssetSymbol, allocation.Amount.Neg()); err != nil {
				return fmt.Errorf("failed to transfer funds from participant: %w", err)
			}
			if err := participantWallet.Record(appSessionID, allocation.AssetSymbol, allocation.Amount); err != nil {
				return fmt.Errorf("failed to transfer funds to virtual app: %w", err)
			}
		}

		return tx.Create(&AppSession{
			Protocol:           params.Definition.Protocol,
			SessionID:          appSessionID,
			ParticipantWallets: params.Definition.ParticipantWallets,
			Status:             ChannelStatusOpen,
			Challenge:          params.Definition.Challenge,
			Weights:            params.Definition.Weights,
			Quorum:             params.Definition.Quorum,
			Nonce:              params.Definition.Nonce,
			Version:            rpc.Req.Timestamp,
		}).Error
	})

	if err != nil {
		return nil, err
	}

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{
		&AppSessionResponse{
			AppSessionID: appSessionID,
			Status:       string(ChannelStatusOpen),
		},
	}), nil
}

// HandleCloseApplication closes a virtual app session and redistributes funds to participants
func HandleCloseApplication(policy *Policy, rpc *RPCMessage, db *gorm.DB) (*RPCMessage, error) {
	var params CloseAppSessionParams
	if err := parseParams(rpc.Req.Params, &params); err != nil {
		return nil, err
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

	var recoveredAddresses = map[string]bool{}
	for _, sigHex := range rpc.Sig {
		recovered, err := RecoverAddress(rpc.ReqRaw, sigHex)
		if err != nil {
			return nil, err
		}
		recoveredAddresses[recovered] = true
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		var appSession AppSession
		if err := tx.Where("session_id = ? AND status = ?", params.AppSessionID, ChannelStatusOpen).
			Order("nonce DESC").
			First(&appSession).Error; err != nil {
			return fmt.Errorf("virtual app not found or not open: %w", err)
		}

		participantWeights := map[string]int64{}
		for i, addr := range appSession.ParticipantWallets {
			participantWeights[addr] = appSession.Weights[i]
		}

		var totalWeight int64
		for address := range recoveredAddresses {
			addr := address
			if walletAddress, _ := GetWalletBySigner(address); walletAddress != "" {
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

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{
		&AppSessionResponse{
			AppSessionID: params.AppSessionID,
			Status:       string(ChannelStatusClosed),
		},
	}), nil
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

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{
		AppDefinition{
			Protocol:           vApp.Protocol,
			ParticipantWallets: vApp.ParticipantWallets,
			Weights:            vApp.Weights,
			Quorum:             vApp.Quorum,
			Challenge:          vApp.Challenge,
			Nonce:              vApp.Nonce,
		},
	}), nil
}

// HandleGetAppSessions returns a list of app sessions
func HandleGetAppSessions(rpc *RPCMessage, db *gorm.DB) (*RPCMessage, error) {
	var participant, status string

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

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{response}), nil
}

// HandleResizeChannel processes a request to resize a payment channel
func HandleResizeChannel(policy *Policy, rpc *RPCMessage, db *gorm.DB, signer *Signer) (*RPCMessage, error) {
	var params ResizeChannelParams
	if err := parseParams(rpc.Req.Params, &params); err != nil {
		return nil, err
	}
	if err := validate.Struct(&params); err != nil {
		return nil, err
	}

	channel, err := GetChannelByID(db, params.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to find channel: %w", err)
	}

	recoveredAddress, err := RecoverAddress(rpc.ReqRaw, rpc.Sig[0])
	if err != nil {
		return nil, err
	}

	walletAddress, _ := GetWalletBySigner(recoveredAddress)
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

	if params.ResizeAmount == nil {
		params.ResizeAmount = big.NewInt(0)
	}
	if params.AllocateAmount == nil {
		params.AllocateAmount = big.NewInt(0)
	}

	ledger := GetWalletLedger(db, channel.Wallet)
	balance, err := ledger.Balance(channel.Wallet, asset.Symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to check participant balance: %w", err)
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

	resizeAmounts := []*big.Int{params.ResizeAmount, params.AllocateAmount}

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

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{response}), nil
}

// HandleCloseChannel processes a request to close a payment channel
func HandleCloseChannel(policy *Policy, rpc *RPCMessage, db *gorm.DB, signer *Signer) (*RPCMessage, error) {
	var params CloseChannelParams
	if err := parseParams(rpc.Req.Params, &params); err != nil {
		return nil, err
	}

	channel, err := GetChannelByID(db, params.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to find channel: %w", err)
	}

	recoveredAddress, err := RecoverAddress(rpc.ReqRaw, rpc.Sig[0])
	if err != nil {
		return nil, err
	}

	walletAddress, _ := GetWalletBySigner(recoveredAddress)
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

	ledger := GetWalletLedger(db, channel.Wallet)
	balance, err := ledger.Balance(channel.Wallet, asset.Symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to check participant balance: %w", err)
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
			Amount:      new(big.Int).Sub(channelAmount, rawBalance),
		},
	}

	stateDataStr := "0x"
	stateData, err := hexutil.Decode(stateDataStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode state data: %w", err)
	}

	encodedState, err := nitrolite.EncodeState(common.HexToHash(channel.ChannelID), nitrolite.IntentFINALIZE, big.NewInt(int64(channel.Version)+1), stateData, allocations)
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

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{response}), nil
}

// HandleGetChannels returns a list of channels for a given account
func HandleGetChannels(rpc *RPCMessage, db *gorm.DB) (*RPCMessage, error) {
	var participant, status string

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

	var channels []Channel
	var err error

	if participant == "" {
		query := db
		if status != "" {
			query = query.Where("status = ?", status)
		}
		err = query.Order("created_at DESC").Find(&channels).Error
		if err != nil {
			return nil, fmt.Errorf("failed to get all channels: %w", err)
		}
	} else {
		channels, err = getChannelsByWallet(db, participant, status)
		if err != nil {
			return nil, fmt.Errorf("failed to get channels: %w", err)
		}
	}

	response := make([]ChannelResponse, 0, len(channels))
	for _, channel := range channels {
		response = append(response, ChannelResponse{
			ChannelID:   channel.ChannelID,
			Participant: channel.Participant,
			Wallet:      channel.Wallet,
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

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{response}), nil
}

// HandleGetRPCHistory returns past RPC calls for a given participant
func HandleGetRPCHistory(policy *Policy, rpc *RPCMessage, store *RPCStore) (*RPCMessage, error) {
	participant := policy.Wallet
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

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{response}), nil
}

// HandleGetAssets returns all supported assets
func HandleGetAssets(rpc *RPCMessage, db *gorm.DB) (*RPCMessage, error) {
	var chainID *uint32

	if len(rpc.Req.Params) > 0 {
		paramsJSON, err := json.Marshal(rpc.Req.Params[0])
		if err == nil {
			var params map[string]interface{}
			if err := json.Unmarshal(paramsJSON, &params); err == nil {
				if cid, ok := params["chain_id"]; ok {
					if chainIDFloat, ok := cid.(float64); ok {
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

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{response}), nil
}

func parseParams(params []any, unmarshalTo any) error {
	if len(params) == 0 {
		return errors.New("missing parameters")
	}
	paramsJSON, err := json.Marshal(params[0])
	if err != nil {
		return fmt.Errorf("failed to parse parameters: %w", err)
	}
	return json.Unmarshal(paramsJSON, &unmarshalTo)
}
