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
	logger.Info("ledger balances retrieved", "userID", c.UserID, "accountID", account)
}

// HandleTransfer unified balance funds to the specified account
func (r *RPCRouter) HandleTransfer(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var params Transfer
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
		"sessionID", appSessionID,
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
		c.Fail("failed to submit state")
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
		c.Fail("failed to close application session")
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

	newVersion := channel.Version + 1
	resp := ResizeChannelResponse{
		ChannelID: channel.ChannelID,
		Intent:    uint8(nitrolite.IntentRESIZE),
		Version:   newVersion,
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
	logger.Info("channel resized",
		"userID", c.UserID,
		"channelID", channel.ChannelID,
		"newVersion", newVersion,
		"fundsDestination", params.FundsDestination,
		"resizeAmount", params.ResizeAmount.String(),
		"allocateAmount", params.AllocateAmount.String(),
		"newChannelAmount", newChannelAmount.String(),
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

	finalBrokerAllocation := new(big.Int).Sub(channelAmount, rawBalance)
	allocations := []nitrolite.Allocation{
		{
			Destination: common.HexToAddress(params.FundsDestination),
			Token:       common.HexToAddress(channel.Token),
			Amount:      rawBalance,
		},
		{
			Destination: r.Signer.GetAddress(),
			Token:       common.HexToAddress(channel.Token),
			Amount:      finalBrokerAllocation,
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

	newVersion := channel.Version + 1
	resp := CloseChannelResponse{
		ChannelID: channel.ChannelID,
		Intent:    uint8(nitrolite.IntentFINALIZE),
		Version:   newVersion,
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
	logger.Info("channel closed",
		"userID", c.UserID,
		"channelID", channel.ChannelID,
		"newVersion", newVersion,
		"fundsDestination", params.FundsDestination,
		"finalUserAllocation", rawBalance.String(),
		"finalBrokerAllocation", finalBrokerAllocation.String(),
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
