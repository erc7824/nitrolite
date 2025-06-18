package main

import (
	"encoding/json"

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
	privGroup.Handle("create_app_session", r.HandleCreateApplication)
	privGroup.Handle("submit_state", r.HandleSubmitState)
	privGroup.Handle("close_app_session", r.HandleCloseApplication)
	privGroup.Handle("resize_channel", r.HandleResizeChannel)
	privGroup.Handle("close_channel", r.HandleCloseChannel)

	return r
}

// TODO:
// ON_CONNECT
// // Send assets immediately upon connection (before authentication)
// h.sendAssets(conn)

// // Increment connection metrics
// h.metrics.ConnectionsTotal.Inc()
// h.metrics.ConnectedClients.Inc()

// ON_AUTHENTICATE
// // Send initial balance and channels information in form of balance and channel updates
// channels, err := getChannelsByWallet(h.db, walletAddress, string(ChannelStatusOpen))
// if err != nil {
// 	logger.Error("error retrieving channels for participant", "error", err)
// }

// h.sendChannelsUpdate(walletAddress, channels)
// h.sendBalanceUpdate(walletAddress)

// ON_MESSAGE_SENT
// h.metrics.MessageSent.Inc()

// ON_DISCONNECT
// h.metrics.ConnectedClients.Dec()

// ON_FORWARD
// if msg.AppSessionID != "" {
// 	if err := forwardMessage(ctx, &msg, messageBytes, walletAddress, h); err != nil {
// 		h.sendErrorResponse(walletAddress, nil, conn, "Failed to forward message: "+err.Error())
// 		continue
// 	}
// 	continue
// }

func (r *RPCRouter) LoggerMiddleware(c *RPCContext) {
	logger := r.lg.With("requestID", c.Message.Req.RequestID)
	c.Context = SetContextLogger(c.Context, logger)
	logger = LoggerFromContext(c.Context)
	logger.Info("handling RPC request",
		"method", c.Message.Req.Method,
		"userID", c.UserID)
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
