package main

import (
	"encoding/json"
	"math/big"
	"time"
)

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

	var chainID *uint32
	if len(req.Params) > 0 {
		if paramsJSON, err := json.Marshal(req.Params[0]); err == nil {
			var params map[string]interface{}
			if err := json.Unmarshal(paramsJSON, &params); err == nil {
				if cid, ok := params["chain_id"]; ok {
					if chainIDFloat, ok := cid.(float64); ok {
						chainIDUint := uint32(chainIDFloat)
						chainID = &chainIDUint
					}
				}
			}
		}
	}

	assets, err := GetAllAssets(r.DB, chainID)
	if err != nil {
		logger.Error("failed to get assets", "error", err)
		c.Fail("failed to get assets")
		return
	}

	resp := make([]AssetResponse, 0, len(assets))
	for _, asset := range assets {
		resp = append(resp, AssetResponse(asset))
	}

	c.Succeed(req.Method, resp)
}

// HandleGetAppDefinition returns the application definition for a ledger account
func (r *RPCRouter) HandleGetAppDefinition(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var sessionID string
	if len(req.Params) > 0 {
		if paramsJSON, err := json.Marshal(req.Params[0]); err == nil {
			var params map[string]string
			if err := json.Unmarshal(paramsJSON, &params); err == nil {
				sessionID = params["app_session_id"]
			}
		}
	}

	if sessionID == "" {
		c.Fail("missing account ID")
		return
	}

	var vApp AppSession
	if err := r.DB.Where("session_id = ?", sessionID).First(&vApp).Error; err != nil {
		logger.Error("failed to get application session", "sessionID", sessionID, "error", err)
		c.Fail("failed to get application session")
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
}

// HandleGetAppSessions returns a list of app sessions
func (r *RPCRouter) HandleGetAppSessions(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var participant, status string
	if len(req.Params) > 0 {
		if paramsJSON, err := json.Marshal(req.Params[0]); err == nil {
			var params map[string]string
			if err := json.Unmarshal(paramsJSON, &params); err == nil {
				participant = params["participant"]
				status = params["status"]
			}
		}
	}

	sessions, err := getAppSessions(r.DB, participant, status)
	if err != nil {
		logger.Error("failed to get application sessions", "error", err)
		c.Fail("failed to get application sessions")
		return
	}

	resp := make([]AppSessionResponse, len(sessions))
	for i, session := range sessions {
		resp[i] = AppSessionResponse{
			AppSessionID:       session.SessionID,
			Status:             string(session.Status),
			ParticipantWallets: session.ParticipantWallets,
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
}

// HandleGetChannels returns a list of channels for a given account
func (r *RPCRouter) HandleGetChannels(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var participant, status string
	if len(req.Params) > 0 {
		if paramsJSON, err := json.Marshal(req.Params[0]); err == nil {
			var params map[string]string
			if err := json.Unmarshal(paramsJSON, &params); err == nil {
				participant = params["participant"]
				status = params["status"]
			}
		}
	}

	var channels []Channel
	var err error

	if participant == "" {
		query := r.DB
		if status != "" {
			query = query.Where("status = ?", status)
		}
		err = query.Order("created_at DESC").Find(&channels).Error
	} else {
		channels, err = getChannelsByWallet(r.DB, participant, status)
	}
	if err != nil {
		logger.Error("failed to get channels", "error", err)
		c.Fail("failed to get channels")
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
			Amount:      big.NewInt(int64(channel.Amount)),
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
}

// HandleGetLedgerEntries returns ledger entries for an account
func (r *RPCRouter) HandleGetLedgerEntries(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var accountID, asset, wallet string
	if len(req.Params) > 0 {
		if paramsJSON, err := json.Marshal(req.Params[0]); err == nil {
			var params map[string]string
			if err := json.Unmarshal(paramsJSON, &params); err == nil {
				accountID = params["account_id"]
				asset = params["asset"]
				if w, ok := params["wallet"]; ok {
					wallet = w
				}
			}
		}
	}

	walletAddress := c.UserID
	if wallet != "" {
		walletAddress = wallet
	}

	ledger := GetWalletLedger(r.DB, walletAddress)
	entries, err := ledger.GetEntries(accountID, asset)
	if err != nil {
		logger.Error("failed to get ledger entries", "error", err)
		c.Fail("failed to get ledger entries")
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
}
