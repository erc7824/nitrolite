package clearnet

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/erc7824/nitrolite/clearnode/pkg/rpc"
	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
)

type ClearnodeClient struct {
	rpcDialer rpc.Dialer
	rpcClient *rpc.Client
	signer    sign.Signer // User's Session Key

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

func (c *ClearnodeClient) GetLedgerBalances() (rpc.GetLedgerBalancesResponse, error) {
	res, _, err := c.rpcClient.GetLedgerBalances(context.Background(), rpc.GetLedgerBalancesRequest{})
	if err != nil {
		return rpc.GetLedgerBalancesResponse{}, fmt.Errorf("failed to fetch ledger balances: %w", err)
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

func (c *ClearnodeClient) GetUserTag() (rpc.GetUserTagResponse, error) {
	res, _, err := c.rpcClient.GetUserTag(context.Background())
	if err != nil {
		return rpc.GetUserTagResponse{}, fmt.Errorf("failed to fetch user tag: %w", err)
	}
	return res, nil
}

func (c *ClearnodeClient) RequestChannelCreation(chainID uint32, assetAddress string) (rpc.CreateChannelResponse, error) {
	if c.signer == nil {
		return rpc.CreateChannelResponse{}, fmt.Errorf("client not authenticated")
	}

	sessionKey := c.signer.PublicKey().Address().String()
	amount := decimal.NewFromInt(0)
	params := rpc.CreateChannelRequest{
		ChainID:    chainID,
		SessionKey: &sessionKey,
		Token:      assetAddress,
		Amount:     &amount,
	}
	payload, err := c.rpcClient.PreparePayload(rpc.CreateChannelMethod, params)
	if err != nil {
		return rpc.CreateChannelResponse{}, fmt.Errorf("failed to prepare payload: %w", err)
	}

	hash, err := payload.Hash()
	if err != nil {
		return rpc.CreateChannelResponse{}, fmt.Errorf("failed to hash payload: %w", err)
	}

	sig, err := c.signer.Sign(hash)
	if err != nil {
		return rpc.CreateChannelResponse{}, fmt.Errorf("failed to sign payload: %w", err)
	}

	req := rpc.NewRequest(payload, sig)
	res, _, err := c.rpcClient.CreateChannel(context.Background(), &req)
	if err != nil {
		return rpc.CreateChannelResponse{}, fmt.Errorf("failed to create channel: %w", err)
	}

	return res, nil
}

func (c *ClearnodeClient) RequestChannelClosure(walletAddress sign.Address, channelID string) (rpc.CloseChannelResponse, error) {
	if c.signer == nil {
		return rpc.CloseChannelResponse{}, fmt.Errorf("client not authenticated")
	}

	params := rpc.CloseChannelRequest{
		FundsDestination: walletAddress.String(),
		ChannelID:        channelID,
	}
	payload, err := c.rpcClient.PreparePayload(rpc.CloseChannelMethod, params)
	if err != nil {
		return rpc.CloseChannelResponse{}, fmt.Errorf("failed to prepare payload: %w", err)
	}

	hash, err := payload.Hash()
	if err != nil {
		return rpc.CloseChannelResponse{}, fmt.Errorf("failed to hash payload: %w", err)
	}

	sig, err := c.signer.Sign(hash)
	if err != nil {
		return rpc.CloseChannelResponse{}, fmt.Errorf("failed to sign payload: %w", err)
	}

	req := rpc.NewRequest(payload, sig)
	res, _, err := c.rpcClient.CloseChannel(context.Background(), &req)
	if err != nil {
		return rpc.CloseChannelResponse{}, fmt.Errorf("failed to close channel: %w", err)
	}

	return res, nil
}

func (c *ClearnodeClient) RequestChannelResize(walletAddress sign.Address, channelID string, allocateAmount, resizeAmount decimal.Decimal) (rpc.ResizeChannelResponse, error) {
	if c.signer == nil {
		return rpc.ResizeChannelResponse{}, fmt.Errorf("client not authenticated")
	}

	params := rpc.ResizeChannelRequest{
		ChannelID:        channelID,
		FundsDestination: walletAddress.String(),
		AllocateAmount:   &allocateAmount,
		ResizeAmount:     &resizeAmount,
	}
	payload, err := c.rpcClient.PreparePayload(rpc.ResizeChannelMethod, params)
	if err != nil {
		return rpc.ResizeChannelResponse{}, fmt.Errorf("failed to prepare payload: %w", err)
	}

	hash, err := payload.Hash()
	if err != nil {
		return rpc.ResizeChannelResponse{}, fmt.Errorf("failed to hash payload: %w", err)
	}

	sig, err := c.signer.Sign(hash)
	if err != nil {
		return rpc.ResizeChannelResponse{}, fmt.Errorf("failed to sign payload: %w", err)
	}

	req := rpc.NewRequest(payload, sig)
	res, _, err := c.rpcClient.ResizeChannel(context.Background(), &req)
	if err != nil {
		return rpc.ResizeChannelResponse{}, fmt.Errorf("failed to resize channel: %w", err)
	}

	return res, nil
}

func (c *ClearnodeClient) Transfer(transferByTag bool, destinationValue string, assetSymbol string, amount decimal.Decimal) (rpc.TransferResponse, error) {
	if c.signer == nil {
		return rpc.TransferResponse{}, fmt.Errorf("client not authenticated")
	}

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
	payload, err := c.rpcClient.PreparePayload(rpc.TransferMethod, params)
	if err != nil {
		return rpc.TransferResponse{}, fmt.Errorf("failed to prepare payload: %w", err)
	}

	hash, err := payload.Hash()
	if err != nil {
		return rpc.TransferResponse{}, fmt.Errorf("failed to hash payload: %w", err)
	}

	sig, err := c.signer.Sign(hash)
	if err != nil {
		return rpc.TransferResponse{}, fmt.Errorf("failed to sign payload: %w", err)
	}

	req := rpc.NewRequest(payload, sig)
	res, err := c.rpcDialer.Call(context.Background(), &req)
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

func (c *ClearnodeClient) WaitCh() <-chan struct{} {
	return c.exitCh
}

func (c *ClearnodeClient) exit() {
	close(c.exitCh)
}
