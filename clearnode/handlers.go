package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/erc7824/nitrolite/clearnode/nitrolite"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type AppDefinition struct {
	Protocol           string   `json:"protocol"`
	ParticipantWallets []string `json:"participants"`
	Weights            []int64  `json:"weights"` // Signature weight for each participant.
	Quorum             uint64   `json:"quorum"`
	Challenge          uint64   `json:"challenge"`
	Nonce              uint64   `json:"nonce"`
}

type CreateAppSessionParams struct {
	Definition  AppDefinition   `json:"definition"`
	Allocations []AppAllocation `json:"allocations"`
}

type SubmitStateParams struct {
	AppSessionID string          `json:"app_session_id"`
	Allocations  []AppAllocation `json:"allocations"`
}

type CloseAppSessionParams struct {
	AppSessionID string          `json:"app_session_id"`
	Allocations  []AppAllocation `json:"allocations"`
}

type AppAllocation struct {
	ParticipantWallet string          `json:"participant"`
	AssetSymbol       string          `json:"asset"`
	Amount            decimal.Decimal `json:"amount"`
}

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
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
}

type ResizeChannelParams struct {
	ChannelID        string   `json:"channel_id"                          validate:"required"`
	AllocateAmount   *big.Int `json:"allocate_amount,omitempty"           validate:"required_without=ResizeAmount"`
	ResizeAmount     *big.Int `json:"resize_amount,omitempty"             validate:"required_without=AllocateAmount"`
	FundsDestination string   `json:"funds_destination"                   validate:"required"`
}

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

type CloseChannelParams struct {
	ChannelID        string `json:"channel_id"`
	FundsDestination string `json:"funds_destination"`
}

type CloseChannelResponse struct {
	ChannelID        string       `json:"channel_id"`
	Intent           uint8        `json:"intent"`
	Version          uint64       `json:"version"`
	StateData        string       `json:"state_data"`
	FinalAllocations []Allocation `json:"allocations"`
	StateHash        string       `json:"state_hash"`
	Signature        Signature    `json:"server_signature"`
}

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

type AssetResponse struct {
	Token    string `json:"token"`
	ChainID  uint32 `json:"chain_id"`
	Symbol   string `json:"symbol"`
	Decimals uint8  `json:"decimals"`
}

// HandleGetConfig returns the broker configuration
func HandleGetConfig(rpc *RPCMessage, config *Config, signer *Signer) (*RPCMessage, error) {
	supportedNetworks := make([]NetworkInfo, 0, len(config.networks))

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
		if paramsJSON, err := json.Marshal(rpc.Req.Params[0]); err == nil {
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
		if paramsJSON, err := json.Marshal(rpc.Req.Params[0]); err == nil {
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

	resp := make([]LedgerEntryResponse, len(entries))
	for i, entry := range entries {
		resp[i] = LedgerEntryResponse{
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

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{resp}), nil
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

	rpcSigners, err := getWallets(rpc)
	if err != nil {
		return nil, err
	}

	// Generate a unique ID for the virtual application
	appBytes, err := json.Marshal(params.Definition)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app session ID: %w", err)
	}
	appSessionID := crypto.Keccak256Hash(appBytes).Hex()

	err = db.Transaction(func(tx *gorm.DB) error {
		for _, alloc := range params.Allocations {
			if alloc.Amount.IsNegative() {
				return fmt.Errorf("invalid allocation: negative amount")
			}
			if alloc.Amount.IsPositive() {
				if _, ok := rpcSigners[alloc.ParticipantWallet]; !ok {
					return fmt.Errorf("missing signature for participant %s", alloc.ParticipantWallet)
				}
			}

			walletAddress := alloc.ParticipantWallet
			if wallet := GetWalletBySigner(alloc.ParticipantWallet); wallet != "" {
				walletAddress = wallet
			}

			challenged, err := getChannelsByWallet(tx, walletAddress, string(ChannelStatusChallenged))
			if err != nil {
				return fmt.Errorf("failed to check challenged channels for %s: %w", walletAddress, err)
			}
			if len(challenged) > 0 {
				return fmt.Errorf("participant %s has challenged channels, cannot create application session", walletAddress)
			}

			ledger := GetWalletLedger(tx, walletAddress)
			balance, err := ledger.Balance(walletAddress, alloc.AssetSymbol)
			if err != nil {
				return fmt.Errorf("failed to check participant balance: %w", err)
			}

			if alloc.Amount.GreaterThan(balance) {
				return fmt.Errorf("insufficient funds: %s for asset %s", alloc.ParticipantWallet, alloc.AssetSymbol)
			}
			if err := ledger.Record(walletAddress, alloc.AssetSymbol, alloc.Amount.Neg()); err != nil {
				return fmt.Errorf("failed to debit participant: %w", err)
			}
			if err := ledger.Record(appSessionID, alloc.AssetSymbol, alloc.Amount); err != nil {
				return fmt.Errorf("failed to credit virtual app: %w", err)
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
			Version:            1,
		}).Error
	})

	if err != nil {
		return nil, err
	}

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{
		&AppSessionResponse{
			AppSessionID: appSessionID,
			Version:      1,
			Status:       string(ChannelStatusOpen),
		},
	}), nil
}

// HandleSubmitState updates funds allocations distribution a virtual app session
func HandleSubmitState(policy *Policy, rpc *RPCMessage, db *gorm.DB) (*RPCMessage, error) {
	var params SubmitStateParams
	if err := parseParams(rpc.Req.Params, &params); err != nil {
		return nil, err
	}
	if params.AppSessionID == "" || len(params.Allocations) == 0 {
		return nil, errors.New("missing required parameters: app_id or allocations")
	}

	rpcSigners, err := getWallets(rpc)
	if err != nil {
		return nil, err
	}

	var newVersion uint64
	err = db.Transaction(func(tx *gorm.DB) error {
		appSession, participantWeights, err := verifyQuorum(tx, params.AppSessionID, rpcSigners)
		if err != nil {
			return err
		}

		appSessionBalance, err := getAppSessionBalances(tx, appSession.SessionID)
		if err != nil {
			return err
		}

		allocationSum := map[string]decimal.Decimal{}
		for _, alloc := range params.Allocations {
			walletAddress := GetWalletBySigner(alloc.ParticipantWallet)
			if walletAddress == "" {
				walletAddress = alloc.ParticipantWallet
			}

			if _, ok := participantWeights[walletAddress]; !ok {
				return fmt.Errorf("allocation to non-participant %s", walletAddress)
			}

			ledger := GetWalletLedger(tx, walletAddress)
			balance, err := ledger.Balance(appSession.SessionID, alloc.AssetSymbol)
			if err != nil {
				return fmt.Errorf("failed to get participant balance: %w", err)
			}

			// Reset participant allocation in app session to the new amount
			if err := ledger.Record(appSession.SessionID, alloc.AssetSymbol, balance.Neg()); err != nil {
				return fmt.Errorf("failed to debit session: %w", err)
			}
			if err := ledger.Record(appSession.SessionID, alloc.AssetSymbol, alloc.Amount); err != nil {
				return fmt.Errorf("failed to credit participant: %w", err)
			}

			allocationSum[alloc.AssetSymbol] = allocationSum[alloc.AssetSymbol].Add(alloc.Amount)
		}

		if err := verifyAllocations(appSessionBalance, allocationSum); err != nil {
			return err
		}

		newVersion = appSession.Version + 1

		return tx.Model(&appSession).Updates(map[string]any{
			"version": newVersion,
		}).Error
	})

	if err != nil {
		return nil, err
	}

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{
		&AppSessionResponse{
			AppSessionID: params.AppSessionID,
			Version:      newVersion,
			Status:       string(ChannelStatusOpen),
		},
	}), nil
}

func verifyAllocations(appSessionBalance, allocationSum map[string]decimal.Decimal) error {
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
	return nil
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

	rpcSigners, err := getWallets(rpc)
	if err != nil {
		return nil, err
	}

	var newVersion uint64
	err = db.Transaction(func(tx *gorm.DB) error {
		appSession, participantWeights, err := verifyQuorum(tx, params.AppSessionID, rpcSigners)
		if err != nil {
			return err
		}

		appSessionBalance, err := getAppSessionBalances(tx, appSession.SessionID)
		if err != nil {
			return err
		}

		allocationSum := map[string]decimal.Decimal{}
		for _, alloc := range params.Allocations {
			walletAddress := GetWalletBySigner(alloc.ParticipantWallet)
			if walletAddress == "" {
				walletAddress = alloc.ParticipantWallet
			}

			if _, ok := participantWeights[walletAddress]; !ok {
				return fmt.Errorf("allocation to non-participant %s", walletAddress)
			}

			ledger := GetWalletLedger(tx, walletAddress)
			balance, err := ledger.Balance(appSession.SessionID, alloc.AssetSymbol)
			if err != nil {
				return fmt.Errorf("failed to get participant balance: %w", err)
			}

			// Debit session, credit participant
			if err := ledger.Record(appSession.SessionID, alloc.AssetSymbol, balance.Neg()); err != nil {
				return fmt.Errorf("failed to debit session: %w", err)
			}
			if err := ledger.Record(walletAddress, alloc.AssetSymbol, alloc.Amount); err != nil {
				return fmt.Errorf("failed to credit participant: %w", err)
			}

			allocationSum[alloc.AssetSymbol] = allocationSum[alloc.AssetSymbol].Add(alloc.Amount)
		}

		if err := verifyAllocations(appSessionBalance, allocationSum); err != nil {
			return err
		}

		newVersion = appSession.Version + 1

		return tx.Model(&appSession).Updates(map[string]any{
			"status":  ChannelStatusClosed,
			"version": newVersion,
		}).Error
	})

	if err != nil {
		return nil, err
	}

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{
		&AppSessionResponse{
			AppSessionID: params.AppSessionID,
			Version:      newVersion,
			Status:       string(ChannelStatusClosed),
		},
	}), nil
}

// HandleGetAppDefinition returns the application definition for a ledger account
func HandleGetAppDefinition(rpc *RPCMessage, db *gorm.DB) (*RPCMessage, error) {
	var sessionID string

	if len(rpc.Req.Params) > 0 {
		if paramsJSON, err := json.Marshal(rpc.Req.Params[0]); err == nil {
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
		if paramsJSON, err := json.Marshal(rpc.Req.Params[0]); err == nil {
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

	resp := make([]AppSessionResponse, len(sessions))
	for i, session := range sessions {
		resp[i] = AppSessionResponse{
			AppSessionID:       session.SessionID,
			Status:             string(session.Status),
			ParticipantWallets: session.ParticipantWallets,
			Protocol:           session.Protocol,
			Challenge:          session.Challenge,
			Weights:            session.Weights,
			Quorum:             session.Quorum,
			Version:            session.Version,
			Nonce:              session.Nonce,
			CreatedAt:          session.CreatedAt.Format(time.RFC3339),
			UpdatedAt:          session.UpdatedAt.Format(time.RFC3339),
		}
	}

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{resp}), nil
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
	if channel == nil {
		return nil, fmt.Errorf("channel with id %s not found", params.ChannelID)
	}

	if channel.Status != ChannelStatusOpen {
		return nil, fmt.Errorf("channel %s must be open and not in challenge to resize, current status: %s", channel.ChannelID, channel.Status)
	}

	if err := verifySigner(rpc, channel.Wallet); err != nil {
		return nil, err
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

	// Prevent no-op resize operations
	if params.ResizeAmount.Cmp(big.NewInt(0)) == 0 && params.AllocateAmount.Cmp(big.NewInt(0)) == 0 {
		return nil, fmt.Errorf("resize operation requires non-zero ResizeAmount or AllocateAmount")
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
	intentionArgs := abi.Arguments{{Type: intentionType}}
	encodedIntentions, err := intentionArgs.Pack(resizeAmounts)
	if err != nil {
		return nil, fmt.Errorf("failed to pack intentions: %w", err)
	}

	// 6) Encode & sign the new state
	channelIDHash := common.HexToHash(channel.ChannelID)
	encodedState, err := nitrolite.EncodeState(channelIDHash, nitrolite.IntentRESIZE, big.NewInt(int64(channel.Version)+1), encodedIntentions, allocations)
	if err != nil {
		return nil, fmt.Errorf("failed to encode state hash: %w", err)
	}
	stateHash := crypto.Keccak256Hash(encodedState).Hex()
	sig, err := signer.NitroSign(encodedState)
	if err != nil {
		return nil, fmt.Errorf("failed to sign state: %w", err)
	}

	resp := ResizeChannelResponse{
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
		resp.Allocations = append(resp.Allocations, Allocation{
			Participant:  alloc.Destination.Hex(),
			TokenAddress: alloc.Token.Hex(),
			Amount:       alloc.Amount,
		})
	}

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{resp}), nil
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
	if channel == nil {
		return nil, fmt.Errorf("channel with id %s not found", params.ChannelID)
	}

	if channel.Status != ChannelStatusOpen {
		return nil, fmt.Errorf("channel %s must be open and not in challenge to close, current status: %s", channel.ChannelID, channel.Status)
	}

	if err := verifySigner(rpc, channel.Wallet); err != nil {
		return nil, err
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
		return nil, fmt.Errorf("insufficient funds for participant: %s", channel.Token)
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

	stateDataHex := "0x"
	stateDataBytes, err := hexutil.Decode(stateDataHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode state data: %w", err)
	}
	encodedState, err := nitrolite.EncodeState(common.HexToHash(channel.ChannelID), nitrolite.IntentFINALIZE, big.NewInt(int64(channel.Version)+1), stateDataBytes, allocations)
	if err != nil {
		return nil, fmt.Errorf("failed to encode state hash: %w", err)
	}
	stateHash := crypto.Keccak256Hash(encodedState).Hex()
	sig, err := signer.NitroSign(encodedState)
	if err != nil {
		return nil, fmt.Errorf("failed to sign state: %w", err)
	}

	resp := CloseChannelResponse{
		ChannelID: channel.ChannelID,
		Intent:    uint8(nitrolite.IntentFINALIZE),
		Version:   channel.Version + 1,
		StateData: stateDataHex,
		StateHash: stateHash,
		Signature: Signature{
			V: sig.V,
			R: hexutil.Encode(sig.R[:]),
			S: hexutil.Encode(sig.S[:]),
		},
	}

	for _, alloc := range allocations {
		resp.FinalAllocations = append(resp.FinalAllocations, Allocation{
			Participant:  alloc.Destination.Hex(),
			TokenAddress: alloc.Token.Hex(),
			Amount:       alloc.Amount,
		})
	}

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{resp}), nil
}

// HandleGetChannels returns a list of channels for a given account
func HandleGetChannels(rpc *RPCMessage, db *gorm.DB) (*RPCMessage, error) {
	var participant, status string

	if len(rpc.Req.Params) > 0 {
		if paramsJSON, err := json.Marshal(rpc.Req.Params[0]); err == nil {
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
	} else {
		channels, err = getChannelsByWallet(db, participant, status)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}

	response := make([]ChannelResponse, 0, len(channels))
	for _, channel := range channels {
		response = append(response, ChannelResponse{
			ChannelID:   channel.ChannelID,
			Participant: channel.Participant,
			Status:      channel.Status,
			Token:       channel.Token,
			Wallet:      channel.Wallet,
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
		if paramsJSON, err := json.Marshal(rpc.Req.Params[0]); err == nil {
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

	resp := make([]AssetResponse, 0, len(assets))
	for _, asset := range assets {
		resp = append(resp, AssetResponse{
			Token:    asset.Token,
			ChainID:  asset.ChainID,
			Symbol:   asset.Symbol,
			Decimals: asset.Decimals,
		})
	}

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{resp}), nil
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

// getWallets retrieves the set of wallet addresses (keys) from RPC request signers.
func getWallets(rpc *RPCMessage) (map[string]struct{}, error) {
	rpcSigners, err := rpc.GetRequestSignersMap()
	if err != nil {
		return nil, err
	}

	result := make(map[string]struct{})
	for addr := range rpcSigners {
		walletAddress := GetWalletBySigner(addr)
		if walletAddress != "" {
			result[walletAddress] = struct{}{}
		} else {
			result[addr] = struct{}{}
		}
	}
	return result, nil
}

// verifySigner checks that the recovered signer matches the channel's wallet.
func verifySigner(rpc *RPCMessage, channelWallet string) error {
	if len(rpc.Sig) < 1 {
		return errors.New("missing participant signature")
	}
	recovered, err := RecoverAddress(rpc.Req.rawBytes, rpc.Sig[0])
	if err != nil {
		return err
	}
	if mapped := GetWalletBySigner(recovered); mapped != "" {
		recovered = mapped
	}
	if recovered != channelWallet {
		return errors.New("invalid signature")
	}
	return nil
}
