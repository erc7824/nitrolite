package main

import (
	"time"

	"github.com/erc7824/nitrolite/clearnode/pkg/rpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

type GetAssetsParams struct {
	ChainID *uint32 `json:"chain_id,omitempty"` // Optional chain ID to filter assets
}

type AssetResponse struct {
	Token    string `json:"token"`
	ChainID  uint32 `json:"chain_id"`
	Symbol   string `json:"symbol"`
	Decimals uint8  `json:"decimals"`
}

type GetAssetsResponse struct {
	Assets []AssetResponse `json:"assets"`
}

type GetChannelsParams struct {
	ListOptions
	Participant string `json:"participant,omitempty"` // Optional participant wallet to filter channels
	Status      string `json:"status,omitempty"`      // Optional status to filter channels
}

type GetAppDefinitionParams struct {
	AppSessionID string `json:"app_session_id"` // The application session ID to get the definition for
}

type AppDefinition struct {
	Application        string      `json:"application"`
	Protocol           rpc.Version `json:"protocol"`
	ParticipantWallets []string    `json:"participants"`
	Weights            []int64     `json:"weights"` // Signature weight for each participant.
	Quorum             uint64      `json:"quorum"`
	Challenge          uint64      `json:"challenge"`
	Nonce              uint64      `json:"nonce"`
}

type GetAppSessionParams struct {
	ListOptions
	Participant string `json:"participant,omitempty"` // Optional participant wallet to filter sessions
	Status      string `json:"status,omitempty"`      // Optional status to filter sessions
}
type GetLedgerEntriesParams struct {
	ListOptions
	AccountID string `json:"account_id,omitempty"` // Optional account ID to filter entries
	Asset     string `json:"asset,omitempty"`      // Optional asset to filter entries
	Wallet    string `json:"wallet,omitempty"`     // Optional wallet address to filter entries
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

type GetLedgerTransactionsParams struct {
	ListOptions
	AccountID string `json:"account_id,omitempty"` // Optional account ID to filter transactions
	Asset     string `json:"asset,omitempty"`      // Optional asset to filter transactions
	TxType    string `json:"tx_type,omitempty"`    // Optional transaction type to filter transactions
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

type GetChannelsResponse struct {
	Channels []ChannelResponse `json:"channels"`
}

type GetAppSessionsResponse struct {
	AppSessions []AppSessionResponse `json:"app_sessions"`
}

type GetLedgerEntriesResponse struct {
	LedgerEntries []LedgerEntryResponse `json:"ledger_entries"`
}

type GetLedgerTransactionsResponse struct {
	LedgerTransactions []TransactionResponse `json:"ledger_transactions"`
}

type ErrorResponse struct {
	Error string `json:"error"` // The error message to send back to the client
}

func (r *RPCRouter) HandlePing(c *RPCContext) {
	c.Succeed("pong", nil)
}

// HandleGetConfig returns the broker configuration
func (r *RPCRouter) HandleGetConfig(c *RPCContext) {
	supportedBlockchains := make([]rpc.BlockchainInfo, 0, len(r.Config.blockchains))

	for _, blockchain := range r.Config.blockchains {
		supportedBlockchains = append(supportedBlockchains, rpc.BlockchainInfo{
			ID:                 blockchain.ID,
			Name:               blockchain.Name,
			CustodyAddress:     blockchain.ContractAddresses.Custody,
			AdjudicatorAddress: blockchain.ContractAddresses.Adjudicator,
		})
	}

	brokerConfig := rpc.BrokerConfig{
		BrokerAddress: r.Signer.GetAddress().Hex(),
		Blockchains:   supportedBlockchains,
	}

	c.Succeed(c.Message.Req.Method, brokerConfig)
}

// HandleGetAssets returns all supported assets
func (r *RPCRouter) HandleGetAssets(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var params GetAssetsParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	respAssets := []AssetResponse{}
	for _, asset := range r.Config.assets.Assets {
		for _, token := range asset.Tokens {
			if params.ChainID != nil && token.BlockchainID != *params.ChainID {
				continue
			}

			respAssets = append(respAssets, AssetResponse{
				Symbol:   asset.Symbol,
				ChainID:  token.BlockchainID,
				Token:    token.Address,
				Decimals: token.Decimals,
			})
		}
	}

	resp := GetAssetsResponse{
		Assets: respAssets,
	}

	c.Succeed(req.Method, resp)
	logger.Info("assets retrieved", "chainID", params.ChainID)
}

// HandleGetChannels returns a list of channels for a given account
func (r *RPCRouter) HandleGetChannels(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var params GetChannelsParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	var channels []Channel
	var err error

	query := applyListOptions(r.DB, "created_at", SortTypeDescending, &params.ListOptions)
	channels, err = getChannelsByWallet(query, params.Participant, params.Status)
	if err != nil {
		logger.Error("failed to get channels", "error", err)
		c.Fail(err, "failed to get channels")
		return
	}

	respChannels := make([]ChannelResponse, 0, len(channels))
	for _, channel := range channels {
		respChannels = append(respChannels, ChannelResponse{
			ChannelID:   channel.ChannelID,
			Participant: channel.Participant,
			Status:      channel.Status,
			Token:       channel.Token,
			Wallet:      channel.Wallet,
			RawAmount:   channel.RawAmount,
			ChainID:     channel.ChainID,
			Adjudicator: channel.Adjudicator,
			Challenge:   channel.Challenge,
			Nonce:       channel.Nonce,
			Version:     channel.State.Version,
			CreatedAt:   channel.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   channel.UpdatedAt.Format(time.RFC3339),
		})
	}

	resp := GetChannelsResponse{
		Channels: respChannels,
	}

	c.Succeed(req.Method, resp)
	logger.Info("channels retrieved", "participant", params.Participant, "status", params.Status)
}

// HandleGetAppDefinition returns the application definition for a ledger account
func (r *RPCRouter) HandleGetAppDefinition(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var params GetAppDefinitionParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}
	if params.AppSessionID == "" {
		c.Fail(nil, "missing account ID")
		return
	}

	var vApp AppSession
	if err := r.DB.Where("session_id = ?", params.AppSessionID).First(&vApp).Error; err != nil {
		logger.Error("failed to get application session", "sessionID", params.AppSessionID, "error", err)
		c.Fail(err, "failed to get application session")
		return
	}

	c.Succeed(req.Method, AppDefinition{
		Protocol:           vApp.Protocol,
		ParticipantWallets: vApp.ParticipantWallets,
		Weights:            vApp.Weights,
		Quorum:             vApp.Quorum,
		Challenge:          vApp.Challenge,
		Nonce:              vApp.Nonce,
	})
	logger.Info("application definition retrieved", "sessionID", params.AppSessionID)
}

// HandleGetAppSessions returns a list of app sessions
func (r *RPCRouter) HandleGetAppSessions(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var params GetAppSessionParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	sessions, err := r.AppSessionService.GetAppSessions(params.Participant, params.Status, &params.ListOptions)
	if err != nil {
		logger.Error("failed to get application sessions", "error", err)
		c.Fail(err, "failed to get application sessions")
		return
	}

	// TODO: update response format accordingly to create struct
	respAppSessions := make([]AppSessionResponse, len(sessions))
	for i, session := range sessions {
		respAppSessions[i] = AppSessionResponse{
			AppSessionID:       session.SessionID,
			Application:        session.Application,
			Status:             string(session.Status),
			ParticipantWallets: session.ParticipantWallets,
			SessionData:        session.SessionData,
			Protocol:           session.Protocol.String(),
			Challenge:          session.Challenge,
			Weights:            session.Weights,
			Quorum:             session.Quorum,
			Version:            session.Version,
			Nonce:              session.Nonce,
			CreatedAt:          session.CreatedAt.Format(time.RFC3339),
			UpdatedAt:          session.UpdatedAt.Format(time.RFC3339),
		}
	}

	resp := GetAppSessionsResponse{
		AppSessions: respAppSessions,
	}

	c.Succeed(req.Method, resp)
	logger.Info("application sessions retrieved", "participant", params.Participant, "status", params.Status)
}

// HandleGetLedgerEntries returns ledger entries for an account
func (r *RPCRouter) HandleGetLedgerEntries(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var params GetLedgerEntriesParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	wallet := c.UserID
	if params.Wallet != "" {
		wallet = params.Wallet
	}
	userAddress := common.HexToAddress(wallet)
	userAccountID := NewAccountID(params.AccountID)

	query := applyListOptions(r.DB, "created_at", SortTypeDescending, &params.ListOptions)
	ledger := GetWalletLedger(query, userAddress)
	entries, err := ledger.GetEntries(&userAccountID, params.Asset)
	if err != nil {
		logger.Error("failed to get ledger entries", "error", err)
		c.Fail(err, "failed to get ledger entries")
		return
	}

	respLedgerEntries := make([]LedgerEntryResponse, len(entries))
	for i, entry := range entries {
		respLedgerEntries[i] = LedgerEntryResponse{
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

	resp := GetLedgerEntriesResponse{
		LedgerEntries: respLedgerEntries,
	}

	c.Succeed(req.Method, resp)
	logger.Info("ledger entries retrieved", "accountID", userAccountID, "asset", params.Asset, "wallet", userAddress)
}

func (r *RPCRouter) HandleGetLedgerTransactions(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var params GetLedgerTransactionsParams
	if err := parseParams(req.Params, &params); err != nil {
		c.Fail(err, "failed to parse parameters")
		return
	}

	var txType *TransactionType
	if params.TxType != "" {
		parsedType, err := parseLedgerTransactionType(params.TxType)
		if err != nil {
			c.Fail(err, "failed to parse transaction type")
			return
		}
		txType = &parsedType
	}

	userAccountID := NewAccountID(params.AccountID)
	query := applyListOptions(r.DB, "created_at", SortTypeDescending, &params.ListOptions)
	transactions, err := GetLedgerTransactionsWithTags(query, userAccountID, params.Asset, txType)
	if err != nil {
		logger.Error("failed to get transactions", "error", err)
		c.Fail(err, "failed to get transactions")
		return
	}

	respLedgerTransactions, err := FormatTransactions(r.DB, transactions)
	if err != nil {
		logger.Error("failed to format transactions", "error", err)
		c.Fail(err, "failed to return transactions")
		return
	}

	resp := GetLedgerTransactionsResponse{
		LedgerTransactions: respLedgerTransactions,
	}

	c.Succeed(req.Method, resp)
	logger.Info("transactions retrieved", "count", len(transactions), "accountID", params.AccountID, "asset", params.Asset, "txType", params.TxType)
}
