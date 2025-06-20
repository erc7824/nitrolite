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

type Transfer struct {
	Destination string               `json:"destination"`
	Allocations []TransferAllocation `json:"allocations"`
}

type TransferAllocation struct {
	AssetSymbol string          `json:"asset"`
	Amount      decimal.Decimal `json:"amount"`
}

type TransferResponse struct {
	From        string               `json:"from"`
	To          string               `json:"to"`
	Allocations []TransferAllocation `json:"allocations"`
	CreatedAt   time.Time            `json:"created_at"`
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

func (r *RPCRouter) BalanceUpdateMiddleware(c *RPCContext) {
	logger := LoggerFromContext(c.Context)
	walletAddress := c.UserID

	c.Next()

	// Send new balances
	balances, err := GetWalletLedger(r.DB, walletAddress).GetBalances(walletAddress)
	if err != nil {
		logger.Error("error getting balances", "sender", walletAddress, "error", err)
		return
	}
	r.Node.Notify(c.UserID, "bu", balances)

	// TODO: notify other participants
}

// HandleGetLedgerBalances returns a list of participants and their balances for a ledger account
func (r *RPCRouter) HandleGetLedgerBalances(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req
	walletAddress := c.UserID

	var account string
	if len(req.Params) > 0 {
		if paramsJSON, err := json.Marshal(req.Params[0]); err == nil {
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

	ledger := GetWalletLedger(r.DB, walletAddress)
	balances, err := ledger.GetBalances(account)
	if err != nil {
		logger.Error("failed to get ledger balances", "error", err)
		c.Fail("failed to get ledger balances")
		return
	}

	c.Succeed(req.Method, balances)
}

// HandleTransfer unified balance funds to the specified account
func HandleRPCRouterTransfer(policy *Policy, rpc *RPCMessage, db *gorm.DB) (*RPCMessage, error) {
	var params Transfer
	if err := parseParams(rpc.Req.Params, &params); err != nil {
		return nil, err
	}
	if params.Destination == "" || params.Destination == policy.Wallet {
		return nil, errors.New("invalid destination")
	}

	// Allow only ledger accounts as destination at the current stage. In the future we'll unlock application accounts.
	if !common.IsHexAddress(params.Destination) {
		return nil, fmt.Errorf("invalid destination account: %s", params.Destination)
	}

	if len(params.Allocations) == 0 {
		return nil, errors.New("empty allocations")
	}

	if err := verifySigner(rpc, policy.Wallet); err != nil {
		return nil, err
	}

	fromWallet := policy.Wallet
	err := db.Transaction(func(tx *gorm.DB) error {
		if wallet := GetWalletBySigner(policy.Wallet); wallet != "" {
			fromWallet = wallet
		}

		if err := checkChallengedChannels(tx, fromWallet); err != nil {
			return err
		}

		for _, alloc := range params.Allocations {
			if alloc.Amount.IsZero() || alloc.Amount.IsNegative() {
				return fmt.Errorf("invalid allocation: %s for asset %s", alloc.Amount, alloc.AssetSymbol)
			}
			ledger := GetWalletLedger(tx, fromWallet)
			balance, err := ledger.Balance(fromWallet, alloc.AssetSymbol)
			if err != nil {
				return fmt.Errorf("failed to check participant balance: %w", err)
			}

			if alloc.Amount.GreaterThan(balance) {
				return fmt.Errorf("insufficient funds: %s for asset %s", fromWallet, alloc.AssetSymbol)
			}

			if err = ledger.Record(fromWallet, alloc.AssetSymbol, alloc.Amount.Neg()); err != nil {
				return fmt.Errorf("failed to debit source account: %w", err)
			}
			ledger = GetWalletLedger(tx, params.Destination)
			if err = ledger.Record(params.Destination, alloc.AssetSymbol, alloc.Amount); err != nil {
				return fmt.Errorf("failed to credit destination account: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{
		&TransferResponse{
			From:        fromWallet,
			To:          params.Destination,
			Allocations: params.Allocations,
			CreatedAt:   time.Now(),
		},
	}), nil
}

// HandleCreateApplication creates a virtual application between participants
func (r *RPCRouter) HandleCreateApplication(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var params CreateAppSessionParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err.Error())
		return
	}
	if len(params.Definition.ParticipantWallets) < 2 {
		c.Fail("invalid number of participants")
		return
	}
	if len(params.Definition.Weights) != len(params.Definition.ParticipantWallets) {
		c.Fail("number of weights must be equal to participants")
		return
	}
	if params.Definition.Nonce == 0 {
		c.Fail("nonce is zero or not provided")
		return
	}

	rpcSigners, err := getWallets(&c.Message)
	if err != nil {
		logger.Error("failed to get wallets from RPC message", "error", err)
		c.Fail("failed to get wallets from RPC message")
		return
	}

	// Generate a unique ID for the virtual application
	appBytes, err := json.Marshal(params.Definition)
	if err != nil {
		logger.Error("failed to generate app session ID", "error", err)
		c.Fail("failed to generate app session ID")
		return
	}
	appSessionID := crypto.Keccak256Hash(appBytes).Hex()

	err = r.DB.Transaction(func(tx *gorm.DB) error {
		for _, alloc := range params.Allocations {
			if alloc.Amount.IsPositive() {
				if _, ok := rpcSigners[alloc.ParticipantWallet]; !ok {
					return fmt.Errorf("missing signature for participant %s", alloc.ParticipantWallet)
				}
			}
			if alloc.Amount.IsNegative() {
				return fmt.Errorf("invalid allocation: %s for asset %s", alloc.Amount, alloc.AssetSymbol)
			}
			walletAddress := alloc.ParticipantWallet
			if wallet := GetWalletBySigner(alloc.ParticipantWallet); wallet != "" {
				walletAddress = wallet
			}

			if err := checkChallengedChannels(tx, walletAddress); err != nil {
				return err
			}

			ledger := GetWalletLedger(tx, walletAddress)
			balance, err := ledger.Balance(walletAddress, alloc.AssetSymbol)
			if err != nil {
				return fmt.Errorf("failed to check participant balance: %w", err)
			}

			if alloc.Amount.GreaterThan(balance) {
				return fmt.Errorf("insufficient funds: %s for asset %s", walletAddress, alloc.AssetSymbol)
			}
			if err := ledger.Record(walletAddress, alloc.AssetSymbol, alloc.Amount.Neg()); err != nil {
				return fmt.Errorf("failed to debit source account: %w", err)
			}
			if err := ledger.Record(appSessionID, alloc.AssetSymbol, alloc.Amount); err != nil {
				return fmt.Errorf("failed to credit destination account: %w", err)
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
		logger.Error("failed to create application session", "error", err)
		c.Fail(err.Error())
		return
	}

	c.Succeed(req.Method, AppSessionResponse{
		AppSessionID: appSessionID,
		Version:      1,
		Status:       string(ChannelStatusOpen),
	})
}

// HandleSubmitState updates funds allocations distribution a virtual app session
func (r *RPCRouter) HandleSubmitState(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var params SubmitStateParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err.Error())
		return
	}
	if params.AppSessionID == "" || len(params.Allocations) == 0 {
		c.Fail("missing required parameters: app_session_id or allocations")
		return
	}

	rpcSigners, err := getWallets(&c.Message)
	if err != nil {
		logger.Error("failed to get wallets from RPC message", "error", err)
		c.Fail("failed to get wallets from RPC message")
		return
	}

	var newVersion uint64
	err = r.DB.Transaction(func(tx *gorm.DB) error {
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
		logger.Error("failed to submit state", "error", err)
		c.Fail("failed to submit state")
		return
	}

	c.Succeed(req.Method, AppSessionResponse{
		AppSessionID: params.AppSessionID,
		Version:      newVersion,
		Status:       string(ChannelStatusOpen),
	})
}

// HandleCloseApplication closes a virtual app session and redistributes funds to participants
func (r *RPCRouter) HandleCloseApplication(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var params CloseAppSessionParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err.Error())
		return
	}
	if params.AppSessionID == "" || len(params.Allocations) == 0 {
		c.Fail("missing required parameters: app_session_id or allocations")
		return
	}

	rpcSigners, err := getWallets(&c.Message)
	if err != nil {
		logger.Error("failed to get wallets from RPC message", "error", err)
		c.Fail("failed to get wallets from RPC message")
		return
	}

	var newVersion uint64
	err = r.DB.Transaction(func(tx *gorm.DB) error {
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
		logger.Error("failed to close application session", "error", err)
		c.Fail("failed to close application session")
		return
	}

	c.Succeed(req.Method, AppSessionResponse{
		AppSessionID: params.AppSessionID,
		Version:      newVersion,
		Status:       string(ChannelStatusClosed),
	})
}

// HandleResizeChannel processes a request to resize a payment channel
func (r *RPCRouter) HandleResizeChannel(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var params ResizeChannelParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err.Error())
		return
	}
	if err := validate.Struct(&params); err != nil {
		c.Fail(err.Error())
		return
	}

	channel, err := GetChannelByID(r.DB, params.ChannelID)
	if err != nil {
		logger.Error("failed to find channel", "error", err)
		c.Fail(fmt.Sprintf("failed to find channel: %s", params.ChannelID))
		return
	}
	if channel == nil {
		c.Fail(fmt.Sprintf("channel %s not found", params.ChannelID))
		return
	}

	if err = checkChallengedChannels(r.DB, channel.Wallet); err != nil {
		logger.Error("failed to check challenged channels", "error", err)
		c.Fail(err.Error())
		return
	}

	if channel.Status != ChannelStatusOpen {
		c.Fail(fmt.Sprintf("channel %s is not open: %s", params.ChannelID, channel.Status))
		return
	}

	if err := verifySigner(&c.Message, channel.Wallet); err != nil {
		logger.Error("failed to verify signer", "error", err)
		c.Fail(err.Error())
		return
	}

	asset, err := GetAssetByToken(r.DB, channel.Token, channel.ChainID)
	if err != nil {
		logger.Error("failed to find asset", "error", err)
		c.Fail(fmt.Sprintf("failed to find asset for token %s on chain %d", channel.Token, channel.ChainID))
		return
	}

	if params.ResizeAmount == nil {
		params.ResizeAmount = big.NewInt(0)
	}
	if params.AllocateAmount == nil {
		params.AllocateAmount = big.NewInt(0)
	}

	// Prevent no-op resize operations
	if params.ResizeAmount.Cmp(big.NewInt(0)) == 0 && params.AllocateAmount.Cmp(big.NewInt(0)) == 0 {
		c.Fail("resize operation requires non-zero ResizeAmount or AllocateAmount")
		return
	}

	ledger := GetWalletLedger(r.DB, channel.Wallet)
	balance, err := ledger.Balance(channel.Wallet, asset.Symbol)
	if err != nil {
		logger.Error("failed to check participant balance", "error", err)
		c.Fail(fmt.Sprintf("failed to check participant balance for asset %s", asset.Symbol))
		return
	}

	rawBalance := balance.Shift(int32(asset.Decimals)).BigInt()
	newChannelAmount := new(big.Int).Add(new(big.Int).SetUint64(channel.Amount), params.AllocateAmount)

	if rawBalance.Cmp(newChannelAmount) < 0 {
		c.Fail("insufficient unified balance")
		return
	}
	newChannelAmount.Add(newChannelAmount, params.ResizeAmount)
	if newChannelAmount.Cmp(big.NewInt(0)) < 0 {
		c.Fail("new channel amount must be positive")
		return
	}

	allocations := []nitrolite.Allocation{
		{
			Destination: common.HexToAddress(params.FundsDestination),
			Token:       common.HexToAddress(channel.Token),
			Amount:      newChannelAmount,
		},
		{
			Destination: r.Signer.GetAddress(),
			Token:       common.HexToAddress(channel.Token),
			Amount:      big.NewInt(0),
		},
	}

	resizeAmounts := []*big.Int{params.ResizeAmount, params.AllocateAmount}

	intentionType, err := abi.NewType("int256[]", "", nil)
	if err != nil {
		logger.Fatal("failed to create intention type", "error", err)
		return
	}
	intentionArgs := abi.Arguments{{Type: intentionType}}
	encodedIntentions, err := intentionArgs.Pack(resizeAmounts)
	if err != nil {
		logger.Error("failed to pack resize amounts", "error", err)
		c.Fail("failed to pack resize amounts")
		return
	}

	// 6) Encode & sign the new state
	channelIDHash := common.HexToHash(channel.ChannelID)
	encodedState, err := nitrolite.EncodeState(channelIDHash, nitrolite.IntentRESIZE, big.NewInt(int64(channel.Version)+1), encodedIntentions, allocations)
	if err != nil {
		logger.Error("failed to encode state hash", "error", err)
		c.Fail("failed to encode state hash")
		return
	}
	stateHash := crypto.Keccak256Hash(encodedState).Hex()
	sig, err := r.Signer.NitroSign(encodedState)
	if err != nil {
		logger.Error("failed to sign state", "error", err)
		c.Fail("failed to sign state")
		return
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

	c.Succeed(req.Method, resp)
}

// HandleCloseChannel processes a request to close a payment channel
func (r *RPCRouter) HandleCloseChannel(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var params CloseChannelParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err.Error())
		return
	}

	channel, err := GetChannelByID(r.DB, params.ChannelID)
	if err != nil {
		logger.Error("failed to find channel", "error", err)
		c.Fail("failed to find channel")
		return
	}
	if channel == nil {
		c.Fail(fmt.Sprintf("channel %s not found", params.ChannelID))
		return
	}

	if err = checkChallengedChannels(r.DB, channel.Wallet); err != nil {
		logger.Error("failed to check challenged channels", "error", err)
		c.Fail(err.Error())
		return
	}

	if channel.Status != ChannelStatusOpen {
		c.Fail(fmt.Sprintf("channel %s is not open: %s", params.ChannelID, channel.Status))
		return
	}

	if err := verifySigner(&c.Message, channel.Wallet); err != nil {
		logger.Error("failed to verify signer", "error", err)
		c.Fail(err.Error())
		return
	}

	asset, err := GetAssetByToken(r.DB, channel.Token, channel.ChainID)
	if err != nil {
		logger.Error("failed to find asset", "error", err)
		c.Fail(fmt.Sprintf("failed to find asset for token %s on chain %d", channel.Token, channel.ChainID))
		return
	}

	ledger := GetWalletLedger(r.DB, channel.Wallet)
	balance, err := ledger.Balance(channel.Wallet, asset.Symbol)
	if err != nil {
		logger.Error("failed to check participant balance", "error", err)
		c.Fail("failed to check participant balance")
		return
	}
	if balance.IsNegative() {
		logger.Error("negative balance", "balance", balance.String())
		c.Fail("negative balance")
		return
	}

	rawBalance := balance.Shift(int32(asset.Decimals)).BigInt()
	channelAmount := new(big.Int).SetUint64(channel.Amount)
	if channelAmount.Cmp(rawBalance) < 0 {
		c.Fail("resize this channel first")
	}

	allocations := []nitrolite.Allocation{
		{
			Destination: common.HexToAddress(params.FundsDestination),
			Token:       common.HexToAddress(channel.Token),
			Amount:      rawBalance,
		},
		{
			Destination: r.Signer.GetAddress(),
			Token:       common.HexToAddress(channel.Token),
			Amount:      new(big.Int).Sub(channelAmount, rawBalance),
		},
	}

	stateDataHex := "0x"
	stateDataBytes, err := hexutil.Decode(stateDataHex)
	if err != nil {
		logger.Error("failed to decode state data hex", "error", err)
		c.Fail("failed to decode state data hex")
		return
	}
	encodedState, err := nitrolite.EncodeState(common.HexToHash(channel.ChannelID), nitrolite.IntentFINALIZE, big.NewInt(int64(channel.Version)+1), stateDataBytes, allocations)
	if err != nil {
		logger.Error("failed to encode state hash", "error", err)
		c.Fail("failed to encode state hash")
		return
	}
	stateHash := crypto.Keccak256Hash(encodedState).Hex()
	sig, err := r.Signer.NitroSign(encodedState)
	if err != nil {
		logger.Error("failed to sign state", "error", err)
		c.Fail("failed to sign state")
		return
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

	c.Succeed(req.Method, resp)
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

func checkChallengedChannels(tx *gorm.DB, wallet string) error {
	challengedChannels, err := getChannelsByWallet(tx, wallet, string(ChannelStatusChallenged))
	if err != nil {
		return fmt.Errorf("failed to check challenged channels: %w", err)
	}
	if len(challengedChannels) > 0 {
		return fmt.Errorf("participant %s has challenged channels, cannot execute operation", wallet)
	}
	return nil
}
