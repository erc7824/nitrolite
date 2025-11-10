package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/erc7824/nitrolite/clearnode/nitrolite"
	"github.com/erc7824/nitrolite/clearnode/pkg/rpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type GetLedgerBalancesParams struct {
	AccountID string `json:"account_id,omitempty"` // Optional account ID to filter balances
}

type GetRPCHistoryParams struct {
	ListOptions
}

type GetSessionKeysParams struct {
	ListOptions
}

type RevokeSessionKeyParams struct {
	SessionKey string `json:"session_key"`
}

type TransferParams struct {
	Destination        string               `json:"destination"`
	DestinationUserTag string               `json:"destination_user_tag"`
	Allocations        []TransferAllocation `json:"allocations"`
}

type GetUserTagResponse struct {
	Tag string `json:"tag"`
}

type TransferAllocation struct {
	AssetSymbol string          `json:"asset"`
	Amount      decimal.Decimal `json:"amount"`
}

type CreateAppSessionParams struct {
	Definition  AppDefinition   `json:"definition"`
	Allocations []AppAllocation `json:"allocations"`
	SessionData *string         `json:"session_data"`
}

type SubmitAppStateParams struct {
	AppSessionID string               `json:"app_session_id"`
	Intent       rpc.AppSessionIntent `json:"intent"`
	Version      uint64               `json:"version"`
	Allocations  []AppAllocation      `json:"allocations"`
	SessionData  *string              `json:"session_data"`
}

type CloseAppSessionParams struct {
	AppSessionID string          `json:"app_session_id"`
	SessionData  *string         `json:"session_data"`
	Allocations  []AppAllocation `json:"allocations"`
}

type AppAllocation struct {
	Participant string          `json:"participant"`
	AssetSymbol string          `json:"asset"`
	Amount      decimal.Decimal `json:"amount"`
}

type AppSessionResponse struct {
	AppSessionID       string   `json:"app_session_id"`
	Application        string   `json:"application"`
	Status             string   `json:"status"`
	ParticipantWallets []string `json:"participants"`
	SessionData        string   `json:"session_data,omitempty"`
	Protocol           string   `json:"protocol"`
	Challenge          uint64   `json:"challenge"`
	Weights            []int64  `json:"weights"`
	Quorum             uint64   `json:"quorum"`
	Version            uint64   `json:"version"`
	Nonce              uint64   `json:"nonce"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
}

type CreateChannelParams struct {
	ChainID uint32 `json:"chain_id" validate:"required"`
	Token   string `json:"token" validate:"required"`
}

type ResizeChannelParams struct {
	ChannelID        string           `json:"channel_id"                          validate:"required"`
	AllocateAmount   *decimal.Decimal `json:"allocate_amount,omitempty"           validate:"omitempty,required_without=ResizeAmount,bigint"`
	ResizeAmount     *decimal.Decimal `json:"resize_amount,omitempty"             validate:"omitempty,required_without=AllocateAmount,bigint"`
	FundsDestination string           `json:"funds_destination"                   validate:"required"`
}

type CloseChannelParams struct {
	ChannelID        string `json:"channel_id"`
	FundsDestination string `json:"funds_destination"`
}

type ChannelOperationResponse struct {
	ChannelID string `json:"channel_id"`
	Channel   *struct {
		Participants [2]string `json:"participants"`
		Adjudicator  string    `json:"adjudicator"`
		Challenge    uint64    `json:"challenge"`
		Nonce        uint64    `json:"nonce"`
	} `json:"channel,omitempty"`
	State          UnsignedState `json:"state"`
	StateSignature Signature     `json:"server_signature"`
}

type ChannelResponse struct {
	ChannelID   string          `json:"channel_id"`
	Participant string          `json:"participant"`
	Status      ChannelStatus   `json:"status"`
	Token       string          `json:"token"`
	Wallet      string          `json:"wallet"`
	RawAmount   decimal.Decimal `json:"amount"` // Total amount in the channel (user + broker)
	ChainID     uint32          `json:"chain_id"`
	Adjudicator string          `json:"adjudicator"`
	Challenge   uint64          `json:"challenge"`
	Nonce       uint64          `json:"nonce"`
	Version     uint64          `json:"version"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

type Balance struct {
	Asset  string          `json:"asset"`
	Amount decimal.Decimal `json:"amount"`
}

type GetLedgerBalancesResponse struct {
	LedgerBalances []Balance `json:"ledger_balances"`
}

type TransferResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
}

type GetRPCHistoryResponse struct {
	RPCEntries []RPCEntry `json:"rpc_entries"`
}

type GetSessionKeysResponse struct {
	SessionKeys []SessionKeyResponse `json:"session_keys"`
}

type AllowanceUsage struct {
	Asset     string          `json:"asset"`
	Allowance decimal.Decimal `json:"allowance"`
	Used      decimal.Decimal `json:"used"`
}

type SessionKeyResponse struct {
	ID          uint             `json:"id"`
	SessionKey  string           `json:"session_key"`
	Application string           `json:"application"`
	Allowances  []AllowanceUsage `json:"allowances"`
	Scope       string           `json:"scope,omitempty"`
	ExpiresAt   time.Time        `json:"expires_at"`
	CreatedAt   time.Time        `json:"created_at"`
}

func (r *RPCRouter) BalanceUpdateMiddleware(c *RPCContext) {
	logger := LoggerFromContext(c.Context)
	userAddress := common.HexToAddress(c.UserID)
	userAccountID := NewAccountID(c.UserID)

	c.Next()

	// Send new balances
	balances, err := GetWalletLedger(r.DB, userAddress).GetBalances(userAccountID)
	if err != nil {
		logger.Error("error getting balances", "sender", userAddress.Hex(), "error", err)
		return
	}
	r.Node.Notify(c.UserID, "bu", BalanceUpdatesResponse{BalanceUpdates: balances})

	// TODO: notify other participants
}

// HandleGetLedgerBalances returns a list of participants and their balances for a ledger account
func (r *RPCRouter) HandleGetLedgerBalances(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req
	userAddress := common.HexToAddress(c.UserID)

	var params GetLedgerBalancesParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	userAccountID := NewAccountID(c.UserID)
	if params.AccountID != "" {
		userAccountID = NewAccountID(params.AccountID)
	}

	ledger := GetWalletLedger(r.DB, userAddress)
	balances, err := ledger.GetBalances(userAccountID)
	if err != nil {
		logger.Error("failed to get ledger balances", "error", err)
		c.Fail(err, "failed to get ledger balances")
		return
	}

	resp := GetLedgerBalancesResponse{
		LedgerBalances: balances,
	}

	c.Succeed(req.Method, resp)
	logger.Info("ledger balances retrieved", "userID", c.UserID, "accountID", userAccountID)
}

// HandleTransfer unified balance funds to the specified account
func (r *RPCRouter) HandleTransfer(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	r.Metrics.TransferAttemptsTotal.Inc()

	var params TransferParams
	if err := parseParams(req.Params, &params); err != nil {
		r.Metrics.TransferAttemptsFail.Inc()
		c.Fail(err, "failed to parse parameters")
		return
	}

	// Check for duplicate transfer by hashing the RPC message
	messageHash := HashMessage(&c.Message)

	if r.MessageCache.Exists(messageHash) {
		r.Metrics.TransferAttemptsFail.Inc()
		c.Fail(nil, "operation denied: the request has already been processed")
		return
	}

	// Allow only ledger accounts as destination at the current stage. In the future we'll unlock application accounts.
	switch {
	case params.Destination == "" && params.DestinationUserTag == "":
		r.Metrics.TransferAttemptsFail.Inc()
		c.Fail(nil, "destination or destination_tag must be provided")
		return
	case params.Destination != "" && !common.IsHexAddress(params.Destination):
		r.Metrics.TransferAttemptsFail.Inc()
		c.Fail(nil, fmt.Sprintf("invalid destination account: %s", params.Destination))
		return
	case len(params.Allocations) == 0:
		r.Metrics.TransferAttemptsFail.Inc()
		c.Fail(nil, "allocations cannot be empty")
		return
	}

	signerAddress, err := verifySigner(&c.Message, c.UserID)
	if err != nil {
		r.Metrics.TransferAttemptsFail.Inc()
		logger.Error("failed to verify signer", "error", err)
		c.Fail(err, "failed to verify signer")
		return
	}

	toAccountTag := params.DestinationUserTag
	fromAccountTag := ""

	destinationAddress := params.Destination
	if destinationAddress == "" {
		// Retrieve the destination address by Tag
		destinationWallet, err := GetWalletByTag(r.DB, params.DestinationUserTag)
		if err != nil {
			r.Metrics.TransferAttemptsFail.Inc()
			logger.Error("failed to get wallet by tag", "tag", params.DestinationUserTag, "error", err)
			c.Fail(err, fmt.Sprintf("failed to get wallet by tag: %s", params.DestinationUserTag))
			return
		}

		destinationAddress = destinationWallet.Wallet
		toAccountTag = destinationWallet.Tag
	}
	if toAccountTag == "" {
		// Even if destination tag is not specified, it should be included in the returned transaction in case it exists
		tag, err := GetUserTagByWallet(r.DB, destinationAddress)
		if err != nil && err != gorm.ErrRecordNotFound {
			r.Metrics.TransferAttemptsFail.Inc()
			logger.Error("failed to get user tag by wallet", "wallet", destinationAddress, "error", err)
			c.Fail(err, fmt.Sprintf("failed to get user tag for wallet: %s", destinationAddress))
			return
		}
		toAccountTag = tag
	}

	if destinationAddress == c.UserID {
		r.Metrics.TransferAttemptsFail.Inc()
		c.Fail(nil, "cannot transfer to self")
		return
	}

	fromWallet := c.UserID
	// Sender tag should be included in the returned transaction in case it exists
	fromAccountTag, err = GetUserTagByWallet(r.DB, fromWallet)
	if err != nil && err != gorm.ErrRecordNotFound {
		r.Metrics.TransferAttemptsFail.Inc()
		logger.Error("failed to get user tag by wallet", "wallet", fromWallet, "error", err)
		c.Fail(err, fmt.Sprintf("failed to get user tag for wallet: %s", fromWallet))
		return
	}

	var respTransactions []TransactionResponse
	err = r.DB.Transaction(func(tx *gorm.DB) error {
		var sessionKeyAddress *string
		if signerAddress != fromWallet {
			sessionKeyAddress = &signerAddress
		}

		if err := checkChallengedChannels(tx, fromWallet); err != nil {
			return err
		}

		if err := ensureWalletHasAllAllocationsEmpty(tx, fromWallet); err != nil {
			return err
		}

		var transactions []TransactionWithTags
		for _, alloc := range params.Allocations {
			if alloc.Amount.IsZero() || alloc.Amount.IsNegative() {
				return RPCErrorf("invalid allocation: %s for asset %s", alloc.Amount, alloc.AssetSymbol)
			}

			// Validate session key spending cap only when wallet didn't sign
			if sessionKeyAddress != nil {
				sessionKey, err := GetSessionKeyIfActive(tx, *sessionKeyAddress)
				if err != nil {
					return RPCErrorf("session key validation failed: %w", err)
				}

				if err := ValidateSessionKeySpending(tx, sessionKey, alloc.AssetSymbol, alloc.Amount); err != nil {
					return RPCErrorf("session key spending validation failed: %w", err)
				}
			}

			fromAddress := common.HexToAddress(fromWallet)
			fromAccountID := NewAccountID(fromWallet)
			ledger := GetWalletLedger(tx, fromAddress)
			balance, err := ledger.Balance(fromAccountID, alloc.AssetSymbol)
			if err != nil {
				return RPCErrorf(ErrGetAccountBalance+": %w", err)
			}
			if alloc.Amount.GreaterThan(balance) {
				return RPCErrorf("insufficient funds: %s for asset %s", fromWallet, alloc.AssetSymbol)
			}
			if err = ledger.Record(fromAccountID, alloc.AssetSymbol, alloc.Amount.Neg(), sessionKeyAddress); err != nil {
				return err
			}

			toAddress := common.HexToAddress(destinationAddress)
			toAccountID := NewAccountID(destinationAddress)
			ledger = GetWalletLedger(tx, toAddress)
			if err = ledger.Record(toAccountID, alloc.AssetSymbol, alloc.Amount, nil); err != nil {
				return err
			}
			transaction, err := RecordLedgerTransaction(tx, TransactionTypeTransfer, fromAccountID, toAccountID, alloc.AssetSymbol, alloc.Amount)
			if err != nil {
				return err
			}
			transactions = append(transactions, TransactionWithTags{
				LedgerTransaction: *transaction,
				FromAccountTag:    fromAccountTag,
				ToAccountTag:      toAccountTag,
			})
		}

		formattedTransactions, err := FormatTransactions(tx, transactions)
		if err != nil {
			return fmt.Errorf("failed to format transactions: %w", err)
		}
		respTransactions = formattedTransactions

		return nil
	})
	if err != nil {
		r.Metrics.TransferAttemptsFail.Inc()
		logger.Error("failed to process transfer", "error", err)
		c.Fail(err, "failed to process transfer")
		return
	}

	// Add message to cache after successful processing to prevent duplicates
	r.MessageCache.Add(messageHash)

	resp := TransferResponse{
		Transactions: respTransactions,
	}

	r.wsNotifier.Notify(
		NewBalanceNotification(fromWallet, r.DB),
		NewTransferNotification(fromWallet, resp),
	)
	if common.IsHexAddress(destinationAddress) {
		r.wsNotifier.Notify(
			NewBalanceNotification(destinationAddress, r.DB),
			NewTransferNotification(destinationAddress, resp),
		)
	}

	r.Metrics.TransferAttemptsSuccess.Inc()
	c.Succeed(req.Method, resp)
	logger.Info("transfer completed", "userID", c.UserID, "transferTo", params.Destination, "allocations", params.Allocations)
}

// HandleCreateApplication creates a virtual application between participants
func (r *RPCRouter) HandleCreateApplication(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var params CreateAppSessionParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	rpcSigners, err := c.Message.GetRequestSignersMap()
	if err != nil {
		logger.Error("failed to get signers from RPC message", "error", err)
		c.Fail(err, "failed to get signers from RPC message")
		return
	}

	resp, err := r.AppSessionService.CreateAppSession(&params, rpcSigners)
	if err != nil {
		logger.Error("failed to create application session", "error", err)
		c.Fail(err, "failed to create application session")
		return
	}

	c.Succeed(req.Method, resp)
	logger.Info("application session created",
		"userID", c.UserID,
		"sessionID", resp.AppSessionID,
		"protocol", params.Definition.Protocol,
		"participants", params.Definition.ParticipantWallets,
		"challenge", params.Definition.Challenge,
		"nonce", params.Definition.Nonce,
		"allocations", params.Allocations,
	)
}

// HandleSubmitAppState updates funds allocations distribution a virtual app session
func (r *RPCRouter) HandleSubmitAppState(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var params SubmitAppStateParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	rpcWallets, err := getWallets(&c.Message)
	if err != nil {
		logger.Error("failed to get wallets from RPC message", "error", err)
		c.Fail(err, "failed to get wallets from RPC message")
		return
	}

	rpcSigners, err := c.Message.GetRequestSignersMap()
	if err != nil {
		logger.Error("failed to get signers from RPC message", "error", err)
		c.Fail(err, "failed to get signers from RPC message")
		return
	}

	resp, err := r.AppSessionService.SubmitAppState(ctx, &params, rpcWallets, rpcSigners)
	if err != nil {
		logger.Error("failed to submit app state", "error", err)
		c.Fail(err, "failed to submit app state")
		return
	}

	c.Succeed(req.Method, resp)
	logger.Info("application session state submitted",
		"userID", c.UserID,
		"sessionID", params.AppSessionID,
		"newVersion", resp.Version,
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
		c.Fail(err, "failed to parse parameters")
		return
	}

	rpcWallets, err := getWallets(&c.Message)
	if err != nil {
		logger.Error("failed to get wallets from RPC message", "error", err)
		c.Fail(err, "failed to get wallets from RPC message")
		return
	}

	rpcSigners, err := c.Message.GetRequestSignersMap()
	if err != nil {
		logger.Error("failed to get signers from RPC message", "error", err)
		c.Fail(err, "failed to get signers from RPC message")
		return
	}

	resp, err := r.AppSessionService.CloseApplication(&params, rpcWallets, rpcSigners)
	if err != nil {
		logger.Error("failed to close application session", "error", err)
		c.Fail(err, "failed to close application session")
		return
	}

	c.Succeed(req.Method, resp)
	logger.Info("application session closed",
		"userID", c.UserID,
		"sessionID", params.AppSessionID,
		"finalVersion", resp.Version,
		"allocations", params.Allocations,
	)
}

// HandleCreateChannel processes a request to create a payment channel with broker
func (r *RPCRouter) HandleCreateChannel(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var params CreateChannelParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	rpcSigners, err := getWallets(&c.Message)
	if err != nil {
		logger.Error("failed to get wallets from RPC message", "error", err)
		c.Fail(err, "failed to get wallets from RPC message")
		return
	}

	resp, err := r.ChannelService.RequestCreate(common.HexToAddress(c.UserID), &params, rpcSigners, logger)
	if err != nil {
		logger.Error("failed to request channel create", "error", err)
		c.Fail(err, "failed to request channel create")
		return
	}

	c.Succeed(req.Method, resp)
	logger.Info("channel create requested",
		"userID", c.UserID,
		"channelID", resp.ChannelID,
	)
}

// HandleResizeChannel processes a request to resize a payment channel
func (r *RPCRouter) HandleResizeChannel(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var params ResizeChannelParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	rpcSigners, err := getWallets(&c.Message)
	if err != nil {
		logger.Error("failed to get wallets from RPC message", "error", err)
		c.Fail(err, "failed to get wallets from RPC message")
		return
	}

	resp, err := r.ChannelService.RequestResize(&params, rpcSigners, logger)
	if err != nil {
		logger.Error("failed to request channel resize", "error", err)
		c.Fail(err, "failed to request channel resize")
		return
	}

	c.Succeed(req.Method, resp)
	logger.Info("channel resize requested",
		"userID", c.UserID,
		"channelID", resp.ChannelID,
		"fundsDestination", params.FundsDestination,
		"resizeAmount", params.ResizeAmount.String(),
		"allocateAmount", params.AllocateAmount.String(),
	)
}

func (r *RPCRouter) HandleGetUserTag(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	tag, err := GetUserTagByWallet(r.DB, c.UserID)
	if err != nil {
		logger.Error("failed to get user tag", "error", err, "wallet", c.UserID)
		c.Fail(err, "failed to get user tag")
		return
	}

	response := GetUserTagResponse{
		Tag: tag,
	}

	c.Succeed(req.Method, response)
}

// HandleCloseChannel processes a request to close a payment channel
func (r *RPCRouter) HandleCloseChannel(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var params CloseChannelParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	rpcSigners, err := getWallets(&c.Message)
	if err != nil {
		logger.Error("failed to get wallets from RPC message", "error", err)
		c.Fail(err, "failed to get wallets from RPC message")
		return
	}

	resp, err := r.ChannelService.RequestClose(&params, rpcSigners, logger)
	if err != nil {
		logger.Error("failed to request channel closure", "error", err)
		c.Fail(err, "failed to request channel closure")
		return
	}

	c.Succeed(req.Method, resp)
	logger.Info("channel close requested",
		"userID", c.UserID,
		"channelID", resp.ChannelID,
		"fundsDestination", params.FundsDestination,
	)
}

// HandleGetRPCHistory returns past RPC calls for a given participant
func (r *RPCRouter) HandleGetRPCHistory(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var params GetRPCHistoryParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	rpcHistory, err := r.RPCStore.GetRPCHistory(c.UserID, &params.ListOptions)
	if err != nil {
		logger.Error("failed to retrieve RPC history", "error", err)
		c.Fail(nil, "failed to retrieve RPC history")
		return
	}

	respRPCEntries := make([]RPCEntry, 0, len(rpcHistory))
	for _, record := range rpcHistory {
		reqSigs, err := nitrolite.SignaturesFromStrings(record.ReqSig)
		if err != nil {
			logger.Error("failed to decode request signature", "error", err, "recordID", record.ID)
			c.Fail(err, "failed to decode request signature")
			return
		}

		resSigs, err := nitrolite.SignaturesFromStrings(record.ResSig)
		if err != nil {
			logger.Error("failed to decode response signature", "error", err, "recordID", record.ID)
			c.Fail(err, "failed to decode response signature")
			return
		}

		respRPCEntries = append(respRPCEntries, RPCEntry{
			ID:        record.ID,
			Sender:    record.Sender,
			ReqID:     record.ReqID,
			Method:    record.Method,
			Params:    string(record.Params),
			Timestamp: record.Timestamp,
			ReqSig:    reqSigs,
			ResSig:    resSigs,
			Result:    string(record.Response),
		})
	}

	resp := GetRPCHistoryResponse{
		RPCEntries: respRPCEntries,
	}

	c.Succeed(req.Method, resp)
	logger.Info("RPC history retrieved", "userID", c.UserID, "entryCount", len(respRPCEntries))
}

// HandleGetSessionKeys returns a list of session keys for the authenticated user
func (r *RPCRouter) HandleGetSessionKeys(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var params GetSessionKeysParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	sessionKeys, err := GetActiveSessionKeysByWallet(r.DB, c.UserID, &params.ListOptions)
	if err != nil {
		logger.Error("failed to retrieve active session keys", "error", err, "wallet", c.UserID)
		c.Fail(err, "failed to retrieve session keys")
		return
	}

	respSessionKeys := make([]SessionKeyResponse, 0, len(sessionKeys))
	for _, sk := range sessionKeys {
		var allowances []Allowance

		// Parse the allowances from the session key
		if sk.Allowance != nil {
			if err := json.Unmarshal([]byte(*sk.Allowance), &allowances); err != nil {
				logger.Error("failed to unmarshal spending cap", "error", err, "sessionKey", sk.Address)
				c.Fail(err, "failed to parse session key spending cap")
				return
			}
		}

		allowanceUsages := make([]AllowanceUsage, 0, len(allowances))
		for _, allowance := range allowances {
			allowanceAmount, err := decimal.NewFromString(allowance.Amount)
			if err != nil {
				logger.Error("failed to parse allowance amount", "error", err, "sessionKey", sk.Address, "asset", allowance.Asset)
				c.Fail(err, "failed to parse allowance amount")
				return
			}

			usedAmount, err := CalculateSessionKeySpending(r.DB, sk.Address, allowance.Asset)
			if err != nil {
				logger.Error("failed to calculate session key spending", "error", err, "sessionKey", sk.Address, "asset", allowance.Asset)
				c.Fail(err, "failed to calculate session key usage")
				return
			}

			allowanceUsages = append(allowanceUsages, AllowanceUsage{
				Asset:     allowance.Asset,
				Allowance: allowanceAmount,
				Used:      usedAmount,
			})
		}

		respSessionKeys = append(respSessionKeys, SessionKeyResponse{
			ID:          sk.ID,
			SessionKey:  sk.Address,
			Application: sk.Application,
			Allowances:  allowanceUsages,
			Scope:       sk.Scope,
			ExpiresAt:   sk.ExpiresAt,
			CreatedAt:   sk.CreatedAt,
		})
	}

	resp := GetSessionKeysResponse{
		SessionKeys: respSessionKeys,
	}

	c.Succeed(req.Method, resp)
}

// HandleRevokeSessionKey revokes a session key
func (r *RPCRouter) HandleRevokeSessionKey(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var params RevokeSessionKeyParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	if params.SessionKey == "" {
		c.Fail(nil, "session_key parameter is required")
		return
	}

	signerAddress, err := verifySigner(&c.Message, c.UserID)
	if err != nil {
		logger.Error("failed to verify signer", "error", err)
		c.Fail(err, "failed to verify signer")
		return
	}

	var activeSessionKeyAddress *string
	if signerAddress != c.UserID {
		activeSessionKeyAddress = &signerAddress
	}

	err = r.DB.Transaction(func(tx *gorm.DB) error {
		_, err := GetActiveSessionKeyForWallet(tx, params.SessionKey, c.UserID)
		if err != nil {
			return RPCErrorf("operation denied: provided address is not an active session key of this user")
		}

		if activeSessionKeyAddress != nil {
			// If trying to revoke a different session key (not self-revocation)
			if *activeSessionKeyAddress != params.SessionKey {
				activeSessionKey, err := GetSessionKeyIfActive(tx, *activeSessionKeyAddress)
				if err != nil {
					return RPCErrorf("operation denied: active session key is invalid")
				}

				// Only "clearnode" session keys can revoke other session keys
				if activeSessionKey.Application != AppNameClearnode {
					return RPCErrorf("operation denied: insufficient permissions for the active session key")
				}
			}
		}

		if err := RevokeSessionKeyFromDB(tx, params.SessionKey); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logger.Error("failed to revoke session key", "error", err, "sessionKey", params.SessionKey)
		c.Fail(err, "failed to revoke session key")
		return
	}

	sessionKeyCache.Delete(params.SessionKey)

	resp := rpc.RevokeSessionKeyResponse{
		SessionKey: params.SessionKey,
	}
	c.Succeed(req.Method, resp)

	authorizedBy := c.UserID
	if activeSessionKeyAddress != nil {
		authorizedBy = *activeSessionKeyAddress
	}
	logger.Info("session key revoked", "userID", c.UserID, "revokedSessionKey", params.SessionKey, "operationAuthorizedBy", authorizedBy)
}

func verifyAllocations(appSessionBalance, allocationSum map[string]decimal.Decimal) error {
	for asset, bal := range appSessionBalance {
		if bal.IsZero() {
			if alloc, ok := allocationSum[asset]; ok && !alloc.IsZero() {
				return RPCErrorf("asset %s is not deposited into the app session. Please deposit the asset first", asset)
			}
			continue
		}
		if alloc, ok := allocationSum[asset]; !ok || !bal.Equal(alloc) {
			return RPCErrorf("asset %s not fully redistributed", asset)
		}
	}
	for asset, allocSum := range allocationSum {
		if _, ok := appSessionBalance[asset]; !ok {
			if !allocSum.IsZero() {
				return RPCErrorf("allocation references unknown asset %s", asset)
			}
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
		walletAddress := GetWalletBySessionKey(addr)
		if walletAddress != "" {
			result[walletAddress] = struct{}{}
		} else {
			result[addr] = struct{}{}
		}
	}
	return result, nil
}

// verifySigner checks that the recovered signer matches the channel's wallet.
func verifySigner(rpc *RPCMessage, channelWallet string) (string, error) {
	if len(rpc.Sig) < 1 {
		return "", RPCErrorf("missing participant signature")
	}
	recovered, err := RecoverAddress(rpc.Req.rawBytes, rpc.Sig[0])
	if err != nil {
		return "", err
	}
	signerAddress := recovered
	if mapped := GetWalletBySessionKey(recovered); mapped != "" {
		recovered = mapped
	}
	if recovered != channelWallet {
		return "", RPCErrorf("invalid signature")
	}
	return signerAddress, nil
}

func ensureWalletHasAllAllocationsEmpty(tx *gorm.DB, wallet string) error {
	channelAmountSum, err := GetChannelAmountSumByWallet(tx, wallet)
	if err != nil {
		return err
	}
	if !channelAmountSum.Sum.IsZero() {
		return RPCErrorf("operation denied: non-zero allocation in %d channel(s) detected owned by wallet %s", channelAmountSum.Count, wallet)
	}
	return nil
}
