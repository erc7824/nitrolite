package channel_api

import "github.com/ethereum/go-ethereum/common"

// TODO: merge service and handlers

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
