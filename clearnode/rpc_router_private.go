package main

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type GetLedgerBalancesParams struct {
	Participant string `json:"participant,omitempty"` // Optional participant address to filter balances
	AccountID   string `json:"account_id,omitempty"`  // Optional account ID to filter balances
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
	AppSessionID string          `json:"app_session_id"`
	Allocations  []AppAllocation `json:"allocations"`
	SessionData  *string         `json:"session_data"`
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
	AppSessionID string   `json:"app_session_id"`
	Status       string   `json:"status"`
	Participants []string `json:"participants"`
	SessionData  string   `json:"session_data,omitempty"`
	Protocol     string   `json:"protocol"`
	Challenge    uint64   `json:"challenge"`
	Weights      []int64  `json:"weights"`
	Quorum       uint64   `json:"quorum"`
	Version      uint64   `json:"version"`
	Nonce        uint64   `json:"nonce"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
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
	RawAmount    *big.Int `json:"amount,string"`
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
	RawAmount   *big.Int      `json:"amount"` // Total amount in the channel (user + broker)
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
	R string `json:"r"`
	S string `json:"s"`
}

type Balance struct {
	Asset  string          `json:"asset"`
	Amount decimal.Decimal `json:"amount"`
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
	r.Node.Notify(c.UserID, "bu", balances)

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
	} else if params.Participant != "" {
		userAccountID = NewAccountID(params.Participant)
	}

	ledger := GetWalletLedger(r.DB, userAddress)
	balances, err := ledger.GetBalances(userAccountID)
	if err != nil {
		logger.Error("failed to get ledger balances", "error", err)
		c.Fail(err, "failed to get ledger balances")
		return
	}

	c.Succeed(req.Method, balances)
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

	// Allow only ledger accounts as destination at the current stage. In the future we'll unlock application accounts.
	switch {
	case params.Destination == "" && params.DestinationUserTag == "":
		r.Metrics.TransferAttemptsFail.Inc()
		c.Fail(nil, "destination or destination_tag must be provided")
		return
	case params.Destination != "" && !common.IsHexAddress(params.Destination) && !isAppSessionID(params.Destination):
		r.Metrics.TransferAttemptsFail.Inc()
		c.Fail(nil, fmt.Sprintf("invalid destination account: %s", params.Destination))
		return
	case len(params.Allocations) == 0:
		r.Metrics.TransferAttemptsFail.Inc()
		c.Fail(nil, "allocations cannot be empty")
		return
	}

	if err := verifySigner(&c.Message, c.UserID); err != nil {
		r.Metrics.TransferAttemptsFail.Inc()
		logger.Error("failed to verify signer", "error", err)
		c.Fail(err, "failed to verify signer")
		return
	}

	toAccountTag := params.DestinationUserTag
	fromAccountTag := ""

	destinationAccount := params.Destination

	if destinationAccount == "" {
		// Retrieve the destination address by Tag
		destinationWallet, err := GetWalletByTag(r.DB, params.DestinationUserTag)
		if err != nil {
			r.Metrics.TransferAttemptsFail.Inc()
			logger.Error("failed to get wallet by tag", "tag", params.DestinationUserTag, "error", err)
			c.Fail(err, fmt.Sprintf("failed to get wallet by tag: %s", params.DestinationUserTag))
			return
		}

		destinationAccount = destinationWallet.Wallet
		toAccountTag = destinationWallet.Tag
	}
	if toAccountTag == "" {
		// Even if destination tag is not specified, it should be included in the returned transaction in case it exists
		tag, err := GetUserTagByWallet(r.DB, destinationAccount)
		if err != nil && err != gorm.ErrRecordNotFound {
			r.Metrics.TransferAttemptsFail.Inc()
			logger.Error("failed to get user tag by wallet", "wallet", destinationAccount, "error", err)
			c.Fail(err, fmt.Sprintf("failed to get user tag for wallet: %s", destinationAccount))
			return
		}
		toAccountTag = tag
	}

	if destinationAccount == c.UserID {
		r.Metrics.TransferAttemptsFail.Inc()
		c.Fail(nil, "cannot transfer to self")
		return
	}

	var appSession AppSession
	fromWallet := c.UserID
	var err error
	// Sender tag should be included in the returned transaction in case it exists
	fromAccountTag, err = GetUserTagByWallet(r.DB, fromWallet)
	if err != nil && err != gorm.ErrRecordNotFound {
		r.Metrics.TransferAttemptsFail.Inc()
		logger.Error("failed to get user tag by wallet", "wallet", fromWallet, "error", err)
		c.Fail(err, fmt.Sprintf("failed to get user tag for wallet: %s", fromWallet))
		return
	}

	var resp []TransactionResponse
	err = r.DB.Transaction(func(tx *gorm.DB) error {
		if wallet := GetWalletBySigner(fromWallet); wallet != "" {
			fromWallet = wallet
		}

		if err := checkChallengedChannels(tx, fromWallet); err != nil {
			return err
		}

		var transactions []TransactionWithTags
		participantWallets := make(map[string]bool)
		isAppDeposit := isAppSessionID(params.Destination)

		// If user is depositing into app session, perform app-specific validation first.
		if isAppDeposit {
			if err := tx.Where("session_id = ? AND status = ?", params.Destination, ChannelStatusOpen).First(&appSession).Error; err != nil {
				return fmt.Errorf("app session is closed or not found %s", params.Destination)
			}
			// Validate that the user is a participant.
			for _, participant := range appSession.Participants {
				participantWallet := participant
				if wallet := GetWalletBySigner(participant); wallet != "" {
					participantWallet = wallet
				}
				participantWallets[participantWallet] = true // Store resolved wallet addresses
			}
			if !participantWallets[fromWallet] {
				return fmt.Errorf("user is not a participant in this app session")
			}

			if err := tx.Model(&appSession).Updates(map[string]any{
				"version": appSession.Version + 1,
			}).Error; err != nil {
				return fmt.Errorf("failed to update app session version: %w", err)
			}
		}

		for _, alloc := range params.Allocations {
			if alloc.Amount.IsZero() || alloc.Amount.IsNegative() {
				return RPCErrorf("invalid allocation: %s for asset %s", alloc.Amount, alloc.AssetSymbol)
			}

			fromAddress := common.HexToAddress(fromWallet)
			fromAccountID := NewAccountID(fromWallet)
			ledger := GetWalletLedger(tx, fromAddress)

			// Debit from source
			balance, err := ledger.Balance(fromAccountID, alloc.AssetSymbol)
			if err != nil {
				return RPCErrorf("failed to check participant balance: %w", err)
			}
			if alloc.Amount.GreaterThan(balance) {
				return RPCErrorf("insufficient funds: %s for asset %s", fromWallet, alloc.AssetSymbol)
			}
			if err = ledger.Record(fromAccountID, alloc.AssetSymbol, alloc.Amount.Neg()); err != nil {
				return RPCErrorf("failed to debit source account: %w", err)
			}

			// Credit to destination
			toAccountID := NewAccountID(destinationAccount)
			txType := TransactionTypeTransfer
			if isAppDeposit {
				// For app deposits, the credit happens within the sender's own ledger to an account representing the app session
				if err = ledger.Record(toAccountID, alloc.AssetSymbol, alloc.Amount); err != nil {
					return fmt.Errorf("failed to credit destination app account: %w", err)
				}
				txType = TransactionTypeAppDeposit
			} else {
				// For direct transfers, credit the recipient's ledger
				destLedger := GetWalletLedger(tx, common.HexToAddress(destinationAccount))
				if err = destLedger.Record(toAccountID, alloc.AssetSymbol, alloc.Amount); err != nil {
					return fmt.Errorf("failed to credit destination account: %w", err)
				}
			}

			// Record the transaction
			transaction, err := RecordLedgerTransaction(tx, txType, fromAccountID, toAccountID, alloc.AssetSymbol, alloc.Amount)
			if err != nil {
				return fmt.Errorf("failed to record transaction: %w", err)
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
		resp = formattedTransactions

		return nil
	})
	if err != nil {
		r.Metrics.TransferAttemptsFail.Inc()
		logger.Error("failed to process transfer", "error", err)
		c.Fail(err, "failed to process transfer")
		return
	}

	r.SendBalanceUpdate(fromWallet)
	r.SendTransferNotification(fromWallet, resp)
	if common.IsHexAddress(destinationAccount) {
		r.SendBalanceUpdate(destinationAccount)
		r.SendTransferNotification(destinationAccount, resp)
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

	rpcSigners, err := getWallets(&c.Message)
	if err != nil {
		logger.Error("failed to get wallets from RPC message", "error", err)
		c.Fail(err, "failed to get wallets from RPC message")
		return
	}

	appSession, err := r.AppSessionService.CreateApplication(&params, rpcSigners)
	if err != nil {
		logger.Error("failed to create application session", "error", err)
		c.Fail(err, "failed to create application session")
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
		"participants", params.Definition.Participants,
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

	rpcSigners, err := getWallets(&c.Message)
	if err != nil {
		logger.Error("failed to get wallets from RPC message", "error", err)
		c.Fail(err, "failed to get wallets from RPC message")
		return
	}

	newVersion, err := r.AppSessionService.SubmitAppState(&params, rpcSigners)
	if err != nil {
		logger.Error("failed to submit app state", "error", err)
		c.Fail(err, "failed to submit app state")
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
		c.Fail(err, "failed to parse parameters")
		return
	}

	rpcSigners, err := getWallets(&c.Message)
	if err != nil {
		logger.Error("failed to get wallets from RPC message", "error", err)
		c.Fail(err, "failed to get wallets from RPC message")
		return
	}

	finalVersion, err := r.AppSessionService.CloseApplication(&params, rpcSigners)
	if err != nil {
		logger.Error("failed to close application session", "error", err)
		c.Fail(err, "failed to close application session")
		return
	}

	c.Succeed(req.Method, AppSessionResponse{
		AppSessionID: params.AppSessionID,
		Version:      finalVersion,
		Status:       string(ChannelStatusClosed),
	})
	logger.Info("application session closed",
		"userID", c.UserID,
		"sessionID", params.AppSessionID,
		"finalVersion", finalVersion,
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
		c.Fail(err, "failed to parse parameters")
		return
	}

	rpcSigners, err := getWallets(&c.Message)
	if err != nil {
		logger.Error("failed to get wallets from RPC message", "error", err)
		c.Fail(err, "failed to get wallets from RPC message")
		return
	}

	resp, err := r.ChannelService.RequestResize(logger, &params, rpcSigners)
	if err != nil {
		logger.Error("failed to request channel resize", "error", err)
		c.Fail(err, "failed to request channel resize")
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

	resp, err := r.ChannelService.RequestClose(logger, &params, rpcSigners)
	if err != nil {
		logger.Error("failed to request channel closure", "error", err)
		c.Fail(err, "failed to request channel closure")
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
			return RPCErrorf("asset %s not fully redistributed", asset)
		}
	}
	for asset := range allocationSum {
		if _, ok := appSessionBalance[asset]; !ok {
			return RPCErrorf("allocation references unknown asset %s", asset)
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
		return RPCErrorf("missing participant signature")
	}
	recovered, err := RecoverAddress(rpc.Req.rawBytes, rpc.Sig[0])
	if err != nil {
		return err
	}
	if mapped := GetWalletBySigner(recovered); mapped != "" {
		recovered = mapped
	}
	if recovered != channelWallet {
		return RPCErrorf("invalid signature")
	}
	return nil
}
