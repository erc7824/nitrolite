package clearnet

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/erc7824/nitrolite/clearnode/pkg/rpc"
	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
)

type ClearnodeClient struct {
	rpcDialer  rpc.Dialer
	rpcClient  *rpc.Client
	sessionKey sign.Signer // User's Session Key

	exitCh chan struct{} // Channel to signal client exit
}

func NewClearnodeClient(wsURL string) (*ClearnodeClient, error) {
	dialer := rpc.NewWebsocketDialer(rpc.DefaultWebsocketDialerConfig)
	rpcClient := rpc.NewClient(dialer)

	client := &ClearnodeClient{
		rpcDialer: dialer,
		rpcClient: rpcClient,
		exitCh:    make(chan struct{}),
	}

	handleError := func(err error) {
		fmt.Printf("Clearnode RPC error: %s\n", err.Error())
		client.exit()
	}

	err := rpcClient.Start(context.Background(), wsURL, handleError)
	if err != nil {
		return nil, fmt.Errorf("failed to start RPC client: %w", err)
	}

	return client, nil
}

func (c *ClearnodeClient) GetConfig() (rpc.GetConfigResponse, error) {
	res, _, err := c.rpcClient.GetConfig(context.Background())
	if err != nil {
		return rpc.GetConfigResponse{}, fmt.Errorf("failed to fetch config: %w", err)
	}
	return res, nil
}

func (c *ClearnodeClient) GetSupportedAssets() (rpc.GetAssetsResponse, error) {
	res, _, err := c.rpcClient.GetAssets(context.Background(), rpc.GetAssetsRequest{})
	if err != nil {
		return rpc.GetAssetsResponse{}, fmt.Errorf("failed to fetch supported assets: %w", err)
	}
	return res, nil
}

func (c *ClearnodeClient) GetLedgerBalances(accountID string) (rpc.GetLedgerBalancesResponse, error) {
	res, _, err := c.rpcClient.GetLedgerBalances(context.Background(), rpc.GetLedgerBalancesRequest{
		AccountID: accountID,
	})
	if err != nil {
		return rpc.GetLedgerBalancesResponse{}, fmt.Errorf("failed to fetch ledger balances: %w", err)
	}
	return res, nil
}

func (c *ClearnodeClient) GetLedgerEntries(wallet, accountID, assetSymbol string, offset, limit uint32) (rpc.GetLedgerEntriesResponse, error) {
	res, _, err := c.rpcClient.GetLedgerEntries(context.Background(), rpc.GetLedgerEntriesRequest{
		ListOptions: rpc.ListOptions{
			Offset: offset,
			Limit:  limit,
		},
		Wallet:    wallet,
		AccountID: accountID,
		Asset:     assetSymbol,
	})
	if err != nil {
		return rpc.GetLedgerEntriesResponse{}, fmt.Errorf("failed to fetch ledger entries: %w", err)
	}
	return res, nil
}

func (c *ClearnodeClient) GetLedgerTransactions(accountID, assetSymbol string, offset, limit uint32) (rpc.GetLedgerTransactionsResponse, error) {
	res, _, err := c.rpcClient.GetLedgerTransactions(context.Background(), rpc.GetLedgerTransactionsRequest{
		ListOptions: rpc.ListOptions{
			Offset: offset,
			Limit:  limit,
		},
		AccountID: accountID,
		Asset:     assetSymbol,
	})
	if err != nil {
		return rpc.GetLedgerTransactionsResponse{}, fmt.Errorf("failed to fetch ledger transactions: %w", err)
	}
	return res, nil
}

func (c *ClearnodeClient) GetChannels(participant, status string) (rpc.GetChannelsResponse, error) {
	res, _, err := c.rpcClient.GetChannels(context.Background(), rpc.GetChannelsRequest{
		Participant: participant,
		Status:      status,
	})
	if err != nil {
		return rpc.GetChannelsResponse{}, fmt.Errorf("failed to fetch channels: %w", err)
	}
	return res, nil
}

func (c *ClearnodeClient) GetAppSessions(participant, status string, offset, limit uint32) (rpc.GetAppSessionsResponse, error) {
	res, _, err := c.rpcClient.GetAppSessions(context.Background(), rpc.GetAppSessionsRequest{
		ListOptions: rpc.ListOptions{
			Offset: offset,
			Limit:  limit,
		},
		Participant: participant,
		Status:      status,
	})
	if err != nil {
		return rpc.GetAppSessionsResponse{}, fmt.Errorf("failed to fetch app sessions: %w", err)
	}
	return res, nil
}

func (c *ClearnodeClient) GetUserTag() (rpc.GetUserTagResponse, error) {
	res, _, err := c.rpcClient.GetUserTag(context.Background())
	if err != nil {
		return rpc.GetUserTagResponse{}, fmt.Errorf("failed to fetch user tag: %w", err)
	}
	return res, nil
}

func (c *ClearnodeClient) RequestChannelCreation(chainID uint32, assetAddress string) (rpc.CreateChannelResponse, error) {
	sessionKeyAddress := c.sessionKey.PublicKey().Address().String()
	params := rpc.CreateChannelRequest{
		ChainID:    chainID,
		SessionKey: &sessionKeyAddress,
		Token:      assetAddress,
	}

	req, err := c.prepareSignedRequest(rpc.CreateChannelMethod, params)
	if err != nil {
		return rpc.CreateChannelResponse{}, fmt.Errorf("failed to prepare create channel request: %w", err)
	}

	res, _, err := c.rpcClient.CreateChannel(context.Background(), req)
	if err != nil {
		return rpc.CreateChannelResponse{}, fmt.Errorf("failed to create channel: %w", err)
	}

	return res, nil
}

func (c *ClearnodeClient) RequestChannelClosure(walletAddress sign.Address, channelID string) (rpc.CloseChannelResponse, error) {
	params := rpc.CloseChannelRequest{
		FundsDestination: walletAddress.String(),
		ChannelID:        channelID,
	}

	req, err := c.prepareSignedRequest(rpc.CloseChannelMethod, params)
	if err != nil {
		return rpc.CloseChannelResponse{}, fmt.Errorf("failed to prepare close channel request: %w", err)
	}

	res, _, err := c.rpcClient.CloseChannel(context.Background(), req)
	if err != nil {
		return rpc.CloseChannelResponse{}, fmt.Errorf("failed to close channel: %w", err)
	}

	return res, nil
}

func (c *ClearnodeClient) RequestChannelResize(walletAddress sign.Address, channelID string, allocateAmount, resizeAmount decimal.Decimal) (rpc.ResizeChannelResponse, error) {
	params := rpc.ResizeChannelRequest{
		ChannelID:        channelID,
		FundsDestination: walletAddress.String(),
		AllocateAmount:   &allocateAmount,
		ResizeAmount:     &resizeAmount,
	}

	req, err := c.prepareSignedRequest(rpc.ResizeChannelMethod, params)
	if err != nil {
		return rpc.ResizeChannelResponse{}, fmt.Errorf("failed to prepare resize channel request: %w", err)
	}

	res, _, err := c.rpcClient.ResizeChannel(context.Background(), req)
	if err != nil {
		return rpc.ResizeChannelResponse{}, fmt.Errorf("failed to resize channel: %w", err)
	}

	return res, nil
}

func (c *ClearnodeClient) Transfer(transferByTag bool, destinationValue string, assetSymbol string, amount decimal.Decimal) (rpc.TransferResponse, error) {
	destination := ""
	destinationUserTag := ""
	if transferByTag {
		destinationUserTag = destinationValue
	} else {
		destination = destinationValue
	}
	params := rpc.TransferRequest{
		Destination:        destination,
		DestinationUserTag: destinationUserTag,
		Allocations: []rpc.TransferAllocation{
			{
				AssetSymbol: assetSymbol,
				Amount:      amount,
			},
		},
	}

	req, err := c.prepareSignedRequest(rpc.TransferMethod, params)
	if err != nil {
		return rpc.TransferResponse{}, fmt.Errorf("failed to prepare transfer request: %w", err)
	}

	res, err := c.rpcDialer.Call(context.Background(), req)
	if err != nil {
		return rpc.TransferResponse{}, fmt.Errorf("failed to transfer funds: %w", err)
	}

	if err := res.Res.Params.Error(); err != nil {
		return rpc.TransferResponse{}, fmt.Errorf("failed to transfer funds: %w", err)
	}

	var resParams rpc.TransferResponse
	if err := res.Res.Params.Translate(&resParams); err != nil {
		return resParams, err
	}

	return resParams, nil
}

func (c *ClearnodeClient) prepareSignedRequest(method rpc.Method, params any) (*rpc.Request, error) {
	if c.sessionKey == nil {
		return nil, fmt.Errorf("client not authenticated")
	}

	payload, err := c.rpcClient.PreparePayload(method, params)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare payload: %w", err)
	}

	hash, err := payload.Hash()
	if err != nil {
		return nil, fmt.Errorf("failed to hash payload: %w", err)
	}

	sig, err := c.sessionKey.Sign(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign payload: %w", err)
	}

	req := rpc.NewRequest(payload, sig)
	return &req, nil
}

func (c *ClearnodeClient) WaitCh() <-chan struct{} {
	return c.exitCh
}

func (c *ClearnodeClient) exit() {
	close(c.exitCh)
}
