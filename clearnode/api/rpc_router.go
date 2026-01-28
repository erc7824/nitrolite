package api

import (
	"time"

	"github.com/erc7824/nitrolite/clearnode/api/app_session_v1"
	"github.com/erc7824/nitrolite/clearnode/api/channel_v1"
	"github.com/erc7824/nitrolite/clearnode/api/node_v1"
	"github.com/erc7824/nitrolite/clearnode/api/user_v1"
	"github.com/erc7824/nitrolite/clearnode/store/database"
	"github.com/erc7824/nitrolite/clearnode/store/memory"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/log"
	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/erc7824/nitrolite/pkg/sign"
)

type RPCRouter struct {
	Node rpc.Node
	lg   log.Logger
}

func NewRPCRouter(
	nodeVersion string,
	minChallenge uint32,
	node rpc.Node,
	signer sign.Signer,
	dbStore database.DatabaseStore,
	memoryStore memory.MemoryStore,
	logger log.Logger,
) *RPCRouter {
	r := &RPCRouter{
		Node: node,
		lg:   logger.WithName("rpc-router"),
	}

	r.Node.Use(r.LoggerMiddleware)

	// Transaction wrapper helpers for each store type
	wrapInTx := func(handler func(database.DatabaseStore) error) error {
		return dbStore.ExecuteInTransaction(handler)
	}
	useChannelV1StoreInTx := func(h channel_v1.StoreTxHandler) error {
		return wrapInTx(func(s database.DatabaseStore) error { return h(s) })
	}
	useAppSessionV1StoreInTx := func(h app_session_v1.StoreTxHandler) error {
		return wrapInTx(func(s database.DatabaseStore) error { return h(s) })
	}
	useUserV1StoreInTx := func(h user_v1.StoreTxHandler) error {
		return wrapInTx(func(s database.DatabaseStore) error { return h(s) })
	}

	nodeAddress := signer.PublicKey().Address().String()

	statePacker := core.NewStatePackerV1(memoryStore)
	stateAdvancer := core.NewStateAdvancerV1(memoryStore)

	validator := sign.NewECDSASigValidator()
	channelV1Handler := channel_v1.NewHandler(useChannelV1StoreInTx, memoryStore, signer, stateAdvancer, statePacker, map[channel_v1.SigValidatorType]channel_v1.SigValidator{
		channel_v1.EcdsaSigValidatorType: validator,
	}, nodeAddress, minChallenge)
	appSessionV1Handler := app_session_v1.NewHandler(useAppSessionV1StoreInTx, memoryStore, signer, stateAdvancer, statePacker, map[app_session_v1.SigType]app_session_v1.SigValidator{
		app_session_v1.EcdsaSigType: validator,
	}, nodeAddress)
	nodeV1Handler := node_v1.NewHandler(memoryStore, nodeAddress, nodeVersion)
	userV1Handler := user_v1.NewHandler(useUserV1StoreInTx)

	appSessionV1Group := r.Node.NewGroup(rpc.AppSessionsV1Group.String())
	appSessionV1Group.Handle(rpc.AppSessionsV1SubmitDepositStateMethod.String(), appSessionV1Handler.SubmitDepositState)
	appSessionV1Group.Handle(rpc.AppSessionsV1SubmitAppStateMethod.String(), appSessionV1Handler.SubmitAppState)
	appSessionV1Group.Handle(rpc.AppSessionsV1RebalanceAppSessionsMethod.String(), appSessionV1Handler.RebalanceAppSessions)
	appSessionV1Group.Handle(rpc.AppSessionsV1CreateAppSessionMethod.String(), appSessionV1Handler.CreateAppSession)
	appSessionV1Group.Handle(rpc.AppSessionsV1GetAppDefinitionMethod.String(), appSessionV1Handler.GetAppDefinition)
	appSessionV1Group.Handle(rpc.AppSessionsV1GetAppSessionsMethod.String(), appSessionV1Handler.GetAppSessions)

	channelV1Group := r.Node.NewGroup(rpc.ChannelV1Group.String())
	channelV1Group.Handle(rpc.ChannelsV1GetEscrowChannelMethod.String(), channelV1Handler.GetEscrowChannel)
	channelV1Group.Handle(rpc.ChannelsV1GetHomeChannelMethod.String(), channelV1Handler.GetHomeChannel)
	channelV1Group.Handle(rpc.ChannelsV1GetLatestStateMethod.String(), channelV1Handler.GetLatestState)
	channelV1Group.Handle(rpc.ChannelsV1RequestCreationMethod.String(), channelV1Handler.RequestCreation)
	channelV1Group.Handle(rpc.ChannelsV1SubmitStateMethod.String(), channelV1Handler.SubmitState)

	nodeV1Group := r.Node.NewGroup(rpc.NodeV1Group.String())
	nodeV1Group.Handle(rpc.NodeV1PingMethod.String(), nodeV1Handler.Ping)
	nodeV1Group.Handle(rpc.NodeV1GetAssetsMethod.String(), nodeV1Handler.GetAssets)
	nodeV1Group.Handle(rpc.NodeV1GetConfigMethod.String(), nodeV1Handler.GetConfig)

	userV1Group := r.Node.NewGroup(rpc.UserV1Group.String())
	userV1Group.Handle(rpc.UserV1GetBalancesMethod.String(), userV1Handler.GetBalances)
	userV1Group.Handle(rpc.UserV1GetTransactionsMethod.String(), userV1Handler.GetTransactions)

	return r
}

func (r *RPCRouter) LoggerMiddleware(c *rpc.Context) {
	logger := r.lg.WithKV("requestID", c.Request.RequestID)
	c.Context = log.SetContextLogger(c.Context, logger)
	logger = log.FromContext(c.Context)

	startTime := time.Now()

	c.Next()

	logger.Info("handled RPC request",
		"method", c.Request.Method,
		"success", c.Response.Type == rpc.MsgTypeResp,
		"duration", time.Since(startTime).String())
}
