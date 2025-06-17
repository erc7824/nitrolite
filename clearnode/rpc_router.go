package main

import (
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
	r.Node.Handle("auth_request", r.HandleAuthRequest)
	r.Node.Handle("auth_verify", r.HandleAuthVerify)

	privGroup := r.Node.NewGroup("private")
	privGroup.Use(r.AuthMiddleware)

	return r
}

func (r *RPCRouter) LoggerMiddleware(c *RPCContext) {
	c.Context = SetContextLogger(c.Context, r.lg)
}

func (r *RPCRouter) HandlePing(c *RPCContext) {
	c.Succeed("pong")
}

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
