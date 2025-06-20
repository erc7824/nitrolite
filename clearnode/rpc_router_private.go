package main

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type GetLedgerBalancesParams struct {
	Participant string `json:"participant,omitempty"` // Optional participant address to filter balances
	AccountID   string `json:"account_id,omitempty"`  // Optional account ID to filter balances
}

type TransferParams struct {
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

	var params GetLedgerBalancesParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err.Error())
		return
	}

	if params.AccountID == "" {
		params.AccountID = params.Participant
		if params.AccountID == "" {
			params.AccountID = walletAddress
		}
	}

	ledger := GetWalletLedger(r.DB, walletAddress)
	balances, err := ledger.GetBalances(params.AccountID)
	if err != nil {
		logger.Error("failed to get ledger balances", "error", err)
		c.Fail("failed to get ledger balances")
		return
	}

	c.Succeed(req.Method, balances)
	logger.Info("ledger balances retrieved", "userID", c.UserID, "accountID", params.AccountID)
}

// HandleTransfer unified balance funds to the specified account
func (r *RPCRouter) HandleTransfer(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var params TransferParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err.Error())
		return
	}

	// Allow only ledger accounts as destination at the current stage. In the future we'll unlock application accounts.
	if params.Destination == "" || params.Destination == c.UserID || !common.IsHexAddress(params.Destination) {
		c.Fail(fmt.Sprintf("invalid destination account: %s", params.Destination))
		return
	}

	if len(params.Allocations) == 0 {
		c.Fail("empty allocations")
		return
	}

	if err := verifySigner(&c.Message, c.UserID); err != nil {
		logger.Error("failed to verify signer", "error", err)
		c.Fail(err.Error())
		return
	}

	fromWallet := c.UserID
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		if wallet := GetWalletBySigner(fromWallet); wallet != "" {
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
		logger.Error("failed to process transfer", "error", err)
		c.Fail(err.Error())
		return
	}

	c.Succeed(req.Method, TransferResponse{
		From:        fromWallet,
		To:          params.Destination,
		Allocations: params.Allocations,
		CreatedAt:   time.Now(),
	})
	logger.Info("transfer completed", "userID", c.UserID, "transferTo", params.Destination, "allocations", params.Allocations)
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

	rpcSigners, err := getWallets(&c.Message)
	if err != nil {
		logger.Error("failed to get wallets from RPC message", "error", err)
		c.Fail("failed to get wallets from RPC message")
		return
	}

	appSession, err := r.AppSessionService.CreateApplication(&params, rpcSigners)
	if err != nil {
		logger.Error("failed to create application session", "error", err)
		c.Fail(err.Error())
		return
	}

	c.Succeed(req.Method, AppSessionResponse{
		AppSessionID: appSession.SessionID,
		Version:      appSession.Version,
		Status:       string(ChannelStatusOpen),
	})
	logger.Info("application session created",
		"userID", c.UserID,
		"sessionID", appSession.SessionID,
		"protocol", params.Definition.Protocol,
		"participants", params.Definition.ParticipantWallets,
		"challenge", params.Definition.Challenge,
		"nonce", params.Definition.Nonce,
		"allocations", params.Allocations,
	)
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

	rpcSigners, err := getWallets(&c.Message)
	if err != nil {
		logger.Error("failed to get wallets from RPC message", "error", err)
		c.Fail("failed to get wallets from RPC message")
		return
	}

	newVersion, err := r.AppSessionService.SubmitState(&params, rpcSigners)
	if err != nil {
		logger.Error("failed to submit state", "error", err)
		c.Fail(err.Error())
		return
	}

	c.Succeed(req.Method, AppSessionResponse{
		AppSessionID: params.AppSessionID,
		Version:      newVersion,
		Status:       string(ChannelStatusOpen),
	})
	logger.Info("application session state submitted",
		"userID", c.UserID,
		"sessionID", params.AppSessionID,
		"newVersion", newVersion,
		"allocations", params.Allocations,
	)
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

	rpcSigners, err := getWallets(&c.Message)
	if err != nil {
		logger.Error("failed to get wallets from RPC message", "error", err)
		c.Fail("failed to get wallets from RPC message")
		return
	}

	newVersion, err := r.AppSessionService.CloseApplication(&params, rpcSigners)
	if err != nil {
		logger.Error("failed to close application session", "error", err)
		c.Fail(err.Error())
		return
	}

	c.Succeed(req.Method, AppSessionResponse{
		AppSessionID: params.AppSessionID,
		Version:      newVersion,
		Status:       string(ChannelStatusClosed),
	})
	logger.Info("application session closed",
		"userID", c.UserID,
		"sessionID", params.AppSessionID,
		"newVersion", newVersion,
		"allocations", params.Allocations,
	)
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

	rpcSigners, err := getWallets(&c.Message)
	if err != nil {
		logger.Error("failed to get wallets from RPC message", "error", err)
		c.Fail("failed to get wallets from RPC message")
		return
	}

	resp, err := r.ChannelService.RequestResize(logger, &params, rpcSigners)
	if err != nil {
		logger.Error("failed to initiate resize channel", "error", err)
		c.Fail(err.Error())
		return
	}

	c.Succeed(req.Method, resp)
	logger.Info("channel resize requested",
		"userID", c.UserID,
		"channelID", resp.ChannelID,
		"newVersion", resp.Version,
		"fundsDestination", params.FundsDestination,
		"resizeAmount", params.ResizeAmount.String(),
		"allocateAmount", params.AllocateAmount.String(),
	)
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

	rpcSigners, err := getWallets(&c.Message)
	if err != nil {
		logger.Error("failed to get wallets from RPC message", "error", err)
		c.Fail("failed to get wallets from RPC message")
		return
	}

	resp, err := r.ChannelService.RequestClose(logger, &params, rpcSigners)
	if err != nil {
		logger.Error("failed to initiate close channel", "error", err)
		c.Fail(err.Error())
		return
	}

	c.Succeed(req.Method, resp)
	logger.Info("channel close requested",
		"userID", c.UserID,
		"channelID", resp.ChannelID,
		"newVersion", resp.Version,
		"fundsDestination", params.FundsDestination,
	)
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
