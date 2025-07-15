package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/invopop/jsonschema"
	"github.com/shopspring/decimal"
)

type GetAssetsParams struct {
	ChainID *uint32 `json:"chain_id,omitempty"` // Optional chain ID to filter assets
}

type GetAssetsResponse struct {
	Token    string `json:"token"`
	ChainID  uint32 `json:"chain_id"`
	Symbol   string `json:"symbol"`
	Decimals uint8  `json:"decimals"`
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
	Protocol           string   `json:"protocol"`
	ParticipantWallets []string `json:"participants"`
	Weights            []int64  `json:"weights"` // Signature weight for each participant.
	Quorum             uint64   `json:"quorum"`
	Challenge          uint64   `json:"challenge"`
	Nonce              uint64   `json:"nonce"`
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
	AccountID Address           `json:"account_id,omitempty"` // Optional account ID to filter transactions
	Asset     string            `json:"asset,omitempty"`      // Optional asset to filter transactions
	TxType    *TransactionType `json:"tx_type,omitempty"`    // Optional transaction type to filter transactions
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

type TransactionResponse struct {
	Id             uint            `json:"id"`
	TxType         TransactionType `json:"tx_type"`
	FromAccount    Address         `json:"from_account"`
	FromAccountTag string          `json:"from_account_tag,omitempty"` // Optional tag for the source account
	ToAccount      Address         `json:"to_account"`
	ToAccountTag   string          `json:"to_account_tag,omitempty"` // Optional tag for the destination account
	Asset          string          `json:"asset"`
	Amount         BigNumber       `json:"amount"`
	CreatedAt      time.Time       `json:"created_at"`
}

type Address string

func (a Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(a))
}

func (a *Address) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	if !common.IsHexAddress(str) {
		return fmt.Errorf("not an address: %s", str)
	}

	*a = Address(common.HexToAddress(str).Hex())
	return nil
}

func (Address) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{Type: "string", Format: "address"}
}

type Hex string

func (h Hex) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(h))
}

func (h *Hex) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*h = Hex(str)
	return nil
}

func (Hex) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{Type: "string", Format: "hex"}
}

type BigNumber string

func NewBigNumber(v decimal.Decimal) BigNumber {
	return BigNumber(v.String())
}

func (BigNumber) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{Type: "string", Format: "bignumber"}
}

func (r *RPCRouter) HandlePing(c *RPCContext) {
	c.Succeed("pong")
}

// HandleGetConfig returns the broker configuration
func (r *RPCRouter) HandleGetConfig(c *RPCContext) {
	supportedNetworks := make([]NetworkInfo, 0, len(r.Config.networks))

	for name, networkConfig := range r.Config.networks {
		supportedNetworks = append(supportedNetworks, NetworkInfo{
			Name:               name,
			ChainID:            networkConfig.ChainID,
			CustodyAddress:     networkConfig.CustodyAddress,
			AdjudicatorAddress: networkConfig.AdjudicatorAddress,
		})
	}

	brokerConfig := BrokerConfig{
		BrokerAddress: r.Signer.GetAddress().Hex(),
		Networks:      supportedNetworks,
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

	query := applySort(r.DB, "symbol", SortTypeAscending, nil)
	assets, err := GetAllAssets(query, params.ChainID)
	if err != nil {
		logger.Error("failed to get assets", "error", err)
		c.Fail(err, "failed to get assets")
		return
	}

	resp := make([]GetAssetsResponse, 0, len(assets))
	for _, asset := range assets {
		resp = append(resp, GetAssetsResponse(asset))
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

	response := make([]ChannelResponse, 0, len(channels))
	for _, channel := range channels {
		response = append(response, ChannelResponse{
			ChannelID:   channel.ChannelID,
			Participant: channel.Participant,
			Status:      channel.Status,
			Token:       channel.Token,
			Wallet:      channel.Wallet,
			RawAmount:   channel.RawAmount.BigInt(),
			ChainID:     channel.ChainID,
			Adjudicator: channel.Adjudicator,
			Challenge:   channel.Challenge,
			Nonce:       channel.Nonce,
			Version:     channel.Version,
			CreatedAt:   channel.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   channel.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.Succeed(req.Method, response)
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
	resp := make([]AppSessionResponse, len(sessions))
	for i, session := range sessions {
		resp[i] = AppSessionResponse{
			AppSessionID:       session.SessionID,
			Status:             string(session.Status),
			ParticipantWallets: session.ParticipantWallets,
			SessionData:        session.SessionData,
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

	userAccountID := NewAccountID(string(params.AccountID))
	query := applyListOptions(r.DB, "created_at", SortTypeDescending, &params.ListOptions)

	transactions, err := GetLedgerTransactionsWithTags(query, userAccountID, params.Asset, params.TxType)
	if err != nil {
		logger.Error("failed to get transactions", "error", err)
		c.Fail(err, "failed to get transactions")
		return
	}

	resp, err := FormatTransactions(r.DB, transactions)
	if err != nil {
		logger.Error("failed to format transactions", "error", err)
		c.Fail(err, "failed to return transactions")
		return
	}

	c.Succeed(req.Method, resp)
	logger.Info("transactions retrieved", "count", len(transactions), "accountID", params.AccountID, "asset", params.Asset, "txType", params.TxType)
}
