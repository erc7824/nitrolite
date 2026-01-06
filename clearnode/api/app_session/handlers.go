package app_session

// TODO: merge service and handlers

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
