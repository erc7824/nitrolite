package main

import (
	"encoding/json"
	"math/big"
	"time"

	"gorm.io/gorm"
)

var (
	ConnectionStoragePolicyKey = "connection_auth_policy"
)

type RPCRouter struct {
	Node        *RPCNode
	Config      *Config
	Signer      *Signer
	DB          *gorm.DB
	AuthManager *AuthManager
	Metrics     *Metrics
	RPCStore    *RPCStore

	lg Logger
}

func NewRPCRouter(
	node *RPCNode,
	conf *Config,
	signer *Signer,
	db *gorm.DB,
	authManager *AuthManager,
	metrics *Metrics,
	rpcStore *RPCStore,
	logger Logger,
) *RPCRouter {
	r := &RPCRouter{
		Node:     node,
		Config:   conf,
		Signer:   signer,
		DB:       db,
		Metrics:  metrics,
		RPCStore: rpcStore,
		lg:       logger.NewSystem("rpc-router"),
	}

	r.Node.OnConnect(r.HandleConnect)
	r.Node.OnDisconnect(r.HandleDisconnect)
	r.Node.OnAuthenticated(r.HandleAuthenticated)
	r.Node.OnMessageSent(r.HandleMessageSent)

	r.Node.Use(r.LoggerMiddleware)
	r.Node.Handle("ping", r.HandlePing)
	r.Node.Handle("get_config", r.HandleGetConfig)
	r.Node.Handle("get_assets", r.HandleGetAssets)
	r.Node.Handle("get_app_definition", r.HandleGetAppDefinition)
	r.Node.Handle("get_app_sessions", r.HandleGetAppSessions)
	r.Node.Handle("auth_request", r.HandleAuthRequest)
	r.Node.Handle("auth_verify", r.HandleAuthVerify)

	privGroup := r.Node.NewGroup("private")
	privGroup.Use(r.AuthMiddleware)
	privGroup.Use(r.HistoryMiddleware)
	privGroup.Handle("resize_channel", r.HandleResizeChannel)
	privGroup.Handle("close_channel", r.HandleCloseChannel)

	appSessionGroup := privGroup.NewGroup("app_session")
	appSessionGroup.Use(r.BalanceUpdateMiddleware)
	// appSessionGroup.Handle("transfer", r.HandleTransfer)
	appSessionGroup.Handle("create_app_session", r.HandleCreateApplication)
	appSessionGroup.Handle("submit_state", r.HandleSubmitState)
	appSessionGroup.Handle("close_app_session", r.HandleCloseApplication)

	return r
}

func (r *RPCRouter) HandleConnect(send SendRPCMessageFunc) {
	// Increment connection metrics
	r.Metrics.ConnectionsTotal.Inc()
	r.Metrics.ConnectedClients.Inc()

	// Get all assets from the database
	assets, err := GetAllAssets(r.DB, nil)
	if err != nil {
		r.lg.Error("failed to get all assets", "error", err)
		return
	}

	// Convert to AssetResponse format
	response := make([]AssetResponse, 0, len(assets))
	for _, asset := range assets {
		response = append(response, AssetResponse(asset))
	}

	send("assets", response)
}

func (r *RPCRouter) HandleDisconnect(userID string) {
	// Decrement connection metrics
	r.Metrics.ConnectedClients.Dec()
}

func (r *RPCRouter) HandleAuthenticated(userID string, send SendRPCMessageFunc) {
	walletAddress := userID

	channels, err := getChannelsByWallet(r.DB, walletAddress, string(ChannelStatusOpen))
	if err != nil {
		r.lg.Error("error retrieving channels for participant", "error", err)
	}

	resp := []ChannelResponse{}
	for _, ch := range channels {
		resp = append(resp, ChannelResponse{
			ChannelID:   ch.ChannelID,
			Participant: ch.Participant,
			Status:      ch.Status,
			Token:       ch.Token,
			Amount:      big.NewInt(int64(ch.Amount)),
			ChainID:     ch.ChainID,
			Adjudicator: ch.Adjudicator,
			Challenge:   ch.Challenge,
			Nonce:       ch.Nonce,
			Version:     ch.Version,
			CreatedAt:   ch.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   ch.UpdatedAt.Format(time.RFC3339),
		})
	}

	// Send channel updates
	send("channels", resp)

	// Send initial balances
	balances, err := GetWalletLedger(r.DB, walletAddress).GetBalances(walletAddress)
	if err != nil {
		r.lg.Error("error getting balances", "sender", walletAddress, "error", err)
		return
	}
	send("bu", balances)
}

func (r *RPCRouter) HandleMessageSent() {
	// Increment sent message counter
	r.Metrics.MessageSent.Inc()
}

func (r *RPCRouter) LoggerMiddleware(c *RPCContext) {
	logger := r.lg.With("requestID", c.Message.Req.RequestID)
	c.Context = SetContextLogger(c.Context, logger)
	logger = LoggerFromContext(c.Context)
	logger.Info("handling RPC request",
		"method", c.Message.Req.Method,
		"userID", c.UserID)

	c.Next()
}

func (r *RPCRouter) MetricsMiddleware(c *RPCContext) {
	// Increment received message counter
	r.Metrics.MessageReceived.Inc()

	reqMethod := c.Message.Req.Method
	c.Next()

	status := "success"
	if c.Message.Res.Method == "error" {
		status = "failure"
	}

	r.Metrics.RPCRequests.WithLabelValues(reqMethod, status).Inc()
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

func (r *RPCRouter) HistoryMiddleware(c *RPCContext) {
	logger := LoggerFromContext(c.Context)

	req := c.Message.Req
	reqSig := c.Message.Sig
	c.Next()

	resRaw, err := json.Marshal(c.Message.Res)
	if err != nil {
		logger.Error("failed to marshal response", "error", err)
		return
	}
	resSig := c.Message.Sig

	// Store the request in history
	if err := r.RPCStore.StoreMessage(c.UserID, req, reqSig, resRaw, resSig); err != nil {
		logger.Error("failed to store RPC message", "error", err)
	}
}

// HandleGetRPCHistory returns past RPC calls for a given participant
func (r *RPCRouter) HandleGetRPCHistory(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)

	var rpcHistory []RPCRecord
	if err := r.RPCStore.db.Where("sender = ?", c.UserID).Order("timestamp DESC").Find(&rpcHistory).Error; err != nil {
		logger.Error("failed to retrieve RPC history", "error", err)
		c.Fail("failed to retrieve RPC history")
		return
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

	c.Succeed(c.Message.Req.Method, response)
}

func (r *RPCRouter) HandlePing(c *RPCContext) {
	c.Succeed("pong")
}
