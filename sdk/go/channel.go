package sdk

import (
	"context"
	"fmt"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
)

// ============================================================================
// Channel Query Methods
// ============================================================================

// GetHomeChannel retrieves home channel information for a user's asset.
//
// Parameters:
//   - wallet: The user's wallet address
//   - asset: The asset symbol
//
// Returns:
//   - Channel information for the home channel
//   - Error if the request fails
//
// Example:
//
//	channel, err := client.GetHomeChannel(ctx, "0x1234...", "usdc")
//	fmt.Printf("Home Channel: %s (Version: %d)\n", channel.ChannelID, channel.StateVersion)
func (c *Client) GetHomeChannel(ctx context.Context, wallet, asset string) (*core.Channel, error) {
	req := rpc.ChannelsV1GetHomeChannelRequest{
		Wallet: wallet,
		Asset:  asset,
	}
	resp, err := c.rpcClient.ChannelsV1GetHomeChannel(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get home channel: %w", err)
	}
	channel := transformChannel(resp.Channel)
	return &channel, nil
}

// GetEscrowChannel retrieves escrow channel information for a specific channel ID.
//
// Parameters:
//   - escrowChannelID: The escrow channel ID to query
//
// Returns:
//   - Channel information for the escrow channel
//   - Error if the request fails
//
// Example:
//
//	channel, err := client.GetEscrowChannel(ctx, "0x1234...")
//	fmt.Printf("Escrow Channel: %s (Version: %d)\n", channel.ChannelID, channel.StateVersion)
func (c *Client) GetEscrowChannel(ctx context.Context, escrowChannelID string) (*core.Channel, error) {
	req := rpc.ChannelsV1GetEscrowChannelRequest{
		EscrowChannelID: escrowChannelID,
	}
	resp, err := c.rpcClient.ChannelsV1GetEscrowChannel(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get escrow channel: %w", err)
	}
	channel := transformChannel(resp.Channel)
	return &channel, nil
}

// ============================================================================
// State Management Methods
// ============================================================================

// GetLatestState retrieves the latest state for a user's asset.
//
// Parameters:
//   - wallet: The user's wallet address
//   - asset: The asset symbol (e.g., "usdc")
//   - onlySigned: If true, returns only the latest signed state
//
// Returns:
//   - core.State containing all state information
//   - Error if the request fails
//
// Example:
//
//	state, err := client.GetLatestState(ctx, "0x1234...", "usdc", false)
//	fmt.Printf("State Version: %d, Balance: %s\n", state.Version, state.HomeLedger.UserBalance)
func (c *Client) GetLatestState(ctx context.Context, wallet, asset string, onlySigned bool) (*core.State, error) {
	req := rpc.ChannelsV1GetLatestStateRequest{
		Wallet:     wallet,
		Asset:      asset,
		OnlySigned: onlySigned,
	}
	resp, err := c.rpcClient.ChannelsV1GetLatestState(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest state: %w", err)
	}
	state, err := transformState(resp.State)
	if err != nil {
		return nil, fmt.Errorf("failed to transform state: %w", err)
	}
	return &state, nil
}

// SubmitState submits a signed state update to the node.
// The state must be properly signed by the user before submission.
//
// Parameters:
//   - state: The state to submit (must include valid signatures and transitions)
//
// Returns:
//   - Node's signature of the state
//   - Error if the request fails
//
// Example:
//
//	nodeSig, err := client.SubmitState(ctx, myState)
//	fmt.Printf("State submitted, node signature: %s\n", nodeSig)
func (c *Client) SubmitState(ctx context.Context, state core.State) (string, error) {
	req := rpc.ChannelsV1SubmitStateRequest{
		State: transformStateToRPC(state),
	}
	resp, err := c.rpcClient.ChannelsV1SubmitState(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to submit state: %w", err)
	}
	return resp.Signature, nil
}

// RequestChannelCreation requests the node to sign a channel creation.
// This is typically the first step when creating a new payment channel.
//
// Parameters:
//   - state: The initial state for the channel
//   - channelDef: The channel definition with nonce and challenge period
//
// Returns:
//   - Node's signature for the channel creation
//   - Error if the request fails
//
// Example:
//
//	channelDef := core.ChannelDefinition{
//	    Nonce: 1,
//	    Challenge: 3600,
//	}
//	nodeSig, err := client.RequestChannelCreation(ctx, initialState, channelDef)
func (c *Client) RequestChannelCreation(ctx context.Context, state core.State, channelDef core.ChannelDefinition) (string, error) {
	req := rpc.ChannelsV1RequestCreationRequest{
		State:             transformStateToRPC(state),
		ChannelDefinition: transformChannelDefinitionToRPC(channelDef),
	}
	resp, err := c.rpcClient.ChannelsV1RequestCreation(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to request channel creation: %w", err)
	}
	return resp.Signature, nil
}
