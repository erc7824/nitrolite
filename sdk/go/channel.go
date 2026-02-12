package sdk

import (
	"context"
	"fmt"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/shopspring/decimal"
)

const (
	// DefaultChallengePeriod is the default challenge period for channels (1 day in seconds)
	DefaultChallengePeriod = 86400
)

// Deposit adds funds to the user's channel by depositing from the blockchain.
// This method handles two scenarios automatically:
//  1. If no channel exists: Creates a new channel with the initial deposit
//  2. If channel exists: Checkpoints the deposit to the existing channel
//
// Parameters:
//   - ctx: Context for the operation
//   - blockchainID: The blockchain network ID (e.g., 80002 for Polygon Amoy)
//   - asset: The asset symbol to deposit (e.g., "usdc")
//   - amount: The amount to deposit
//
// Returns:
//   - Transaction hash of the blockchain transaction
//   - Error if the operation fails
//
// Requirements:
//   - Blockchain RPC must be configured for the specified chain (use WithBlockchainRPC)
//   - User must have approved the token spend to the contract address
//   - User must have sufficient token balance in their wallet
//
// Example:
//
//	txHash, err := client.Deposit(ctx, 80002, "usdc", decimal.NewFromInt(100))
//	fmt.Printf("Deposit transaction: %s\n", txHash)
func (c *Client) Deposit(ctx context.Context, blockchainID uint64, asset string, amount decimal.Decimal) (string, error) {
	userWallet := c.GetUserAddress()

	// Initialize blockchain client if needed
	if err := c.initializeBlockchainClient(ctx, blockchainID); err != nil {
		return "", err
	}

	blockchainClient := c.blockchainClients[blockchainID]

	// Get node address
	nodeAddress, err := c.getNodeAddress(ctx)
	if err != nil {
		return "", err
	}

	// Get token address for this asset on this blockchain
	tokenAddress, err := c.getTokenAddress(ctx, blockchainID, asset)
	if err != nil {
		return "", err
	}

	// Try to get latest state to determine if channel exists
	state, err := c.GetLatestState(ctx, userWallet, asset, false)

	// Scenario A: Channel doesn't exist - create it
	if err != nil || state.HomeChannelID == nil {
		// Create channel definition
		channelDef := core.ChannelDefinition{
			Nonce:     generateNonce(),
			Challenge: DefaultChallengePeriod,
		}

		if state == nil {
			state = core.NewVoidState(asset, userWallet)
		}
		newState := state.NextState()

		_, err := newState.ApplyChannelCreation(channelDef, blockchainID, tokenAddress, nodeAddress)
		if err != nil {
			return "", fmt.Errorf("failed to apply channel creation: %w", err)
		}

		// Apply deposit transition
		_, err = newState.ApplyHomeDepositTransition(amount)
		if err != nil {
			return "", fmt.Errorf("failed to apply deposit transition: %w", err)
		}

		// Sign state
		sig, err := c.SignState(newState)
		if err != nil {
			return "", fmt.Errorf("failed to sign state: %w", err)
		}
		newState.UserSig = &sig

		// Request channel creation from node
		nodeSig, err := c.requestChannelCreation(ctx, *newState, channelDef)
		if err != nil {
			return "", fmt.Errorf("failed to request channel creation: %w", err)
		}
		newState.NodeSig = &nodeSig

		// Create channel on blockchain
		txHash, err := blockchainClient.Create(channelDef, *newState)
		if err != nil {
			return "", fmt.Errorf("failed to create channel on blockchain: %w", err)
		}

		return txHash, nil
	}

	// Scenario B: Channel exists - checkpoint deposit
	// Create next state
	nextState := state.NextState()

	// Apply deposit transition
	_, err = nextState.ApplyHomeDepositTransition(amount)
	if err != nil {
		return "", fmt.Errorf("failed to apply deposit transition: %w", err)
	}

	// Sign and submit state to node
	_, err = c.signAndSubmitState(ctx, nextState)
	if err != nil {
		return "", err
	}

	// Checkpoint on blockchain
	txHash, err := blockchainClient.Checkpoint(*nextState, nil)
	if err != nil {
		return "", fmt.Errorf("failed to checkpoint on blockchain: %w", err)
	}

	return txHash, nil
}

// Withdraw removes funds from the user's channel and returns them to the blockchain wallet.
// This operation handles two scenarios automatically:
//  1. If no channel exists: Creates a new channel and executes the withdrawal in one transaction
//  2. If channel exists: Checkpoints the withdrawal to the existing channel
//
// Parameters:
//   - ctx: Context for the operation
//   - blockchainID: The blockchain network ID (e.g., 80002 for Polygon Amoy)
//   - asset: The asset symbol to withdraw (e.g., "usdc")
//   - amount: The amount to withdraw
//
// Returns:
//   - Transaction hash of the blockchain transaction
//   - Error if the operation fails
//
// Requirements:
//   - Channel must exist (user must have deposited first)
//   - Blockchain RPC must be configured for the specified chain (use WithBlockchainRPC)
//   - User must have sufficient balance in the channel
//
// Example:
//
//	txHash, err := client.Withdraw(ctx, 80002, "usdc", decimal.NewFromInt(25))
//	fmt.Printf("Withdrawal transaction: %s\n", txHash)
func (c *Client) Withdraw(ctx context.Context, blockchainID uint64, asset string, amount decimal.Decimal) (string, error) {
	userWallet := c.GetUserAddress()

	// Initialize blockchain client if needed
	if err := c.initializeBlockchainClient(ctx, blockchainID); err != nil {
		return "", err
	}

	blockchainClient := c.blockchainClients[blockchainID]

	// Get node address (Required for channel creation flow)
	nodeAddress, err := c.getNodeAddress(ctx)
	if err != nil {
		return "", err
	}

	// Get token address for this asset on this blockchain (Required for channel creation flow)
	tokenAddress, err := c.getTokenAddress(ctx, blockchainID, asset)
	if err != nil {
		return "", err
	}

	// Try to get latest state to determine if channel exists
	state, err := c.GetLatestState(ctx, userWallet, asset, false)

	// Channel doesn't exist - create it and withdraw
	if err != nil || state.HomeChannelID == nil {
		// Create channel definition
		channelDef := core.ChannelDefinition{
			Nonce:     generateNonce(),
			Challenge: DefaultChallengePeriod,
		}

		if state == nil {
			state = core.NewVoidState(asset, userWallet)
		}
		newState := state.NextState()

		_, err := newState.ApplyChannelCreation(channelDef, blockchainID, tokenAddress, nodeAddress)
		if err != nil {
			return "", fmt.Errorf("failed to apply channel creation: %w", err)
		}

		// Apply withdrawal transition
		// Note: Ensure your core logic allows withdrawal on a fresh state
		// (assuming the smart contract handles the net balance check)
		_, err = newState.ApplyHomeWithdrawalTransition(amount)
		if err != nil {
			return "", fmt.Errorf("failed to apply withdrawal transition: %w", err)
		}

		// Sign state
		sig, err := c.SignState(newState)
		if err != nil {
			return "", fmt.Errorf("failed to sign state: %w", err)
		}
		newState.UserSig = &sig

		// Request channel creation from node
		nodeSig, err := c.requestChannelCreation(ctx, *newState, channelDef)
		if err != nil {
			return "", fmt.Errorf("failed to request channel creation: %w", err)
		}
		newState.NodeSig = &nodeSig

		// Create channel on blockchain (Smart Contract handles Creation + Withdrawal)
		txHash, err := blockchainClient.Create(channelDef, *newState)
		if err != nil {
			return "", fmt.Errorf("failed to create channel on blockchain: %w", err)
		}

		return txHash, nil
	}

	// Create next state
	nextState := state.NextState()

	// Apply withdrawal transition
	_, err = nextState.ApplyHomeWithdrawalTransition(amount)
	if err != nil {
		return "", fmt.Errorf("failed to apply withdrawal transition: %w", err)
	}

	// Sign and submit state to node
	_, err = c.signAndSubmitState(ctx, nextState)
	if err != nil {
		return "", err
	}

	// Checkpoint on blockchain
	txHash, err := blockchainClient.Checkpoint(*nextState, nil)
	if err != nil {
		return "", fmt.Errorf("failed to checkpoint withdrawal on blockchain: %w", err)
	}

	return txHash, nil
}

// Transfer sends funds from the user to another wallet address.
// This is the simplest operation as it doesn't require any blockchain interaction.
//
// Parameters:
//   - ctx: Context for the operation
//   - recipientWallet: The recipient's wallet address (e.g., "0x1234...")
//   - asset: The asset symbol to transfer (e.g., "usdc")
//   - amount: The amount to transfer
//
// Returns:
//   - Transaction ID for tracking
//   - Error if the operation fails
//
// Errors:
//   - Returns error if channel doesn't exist (user must deposit first)
//   - Returns error if insufficient balance
//   - Returns error if state submission fails
//
// Example:
//
//	txID, err := client.Transfer(ctx, "0xRecipient...", "usdc", decimal.NewFromInt(50))
//	fmt.Printf("Transfer successful: %s\n", txID)
func (c *Client) Transfer(ctx context.Context, recipientWallet string, asset string, amount decimal.Decimal) (string, error) {
	// Get sender's latest state
	senderWallet := c.GetUserAddress()
	state, err := c.GetLatestState(ctx, senderWallet, asset, false)
	if err != nil || state.HomeChannelID == nil {
		// Create channel definition
		channelDef := core.ChannelDefinition{
			Nonce:     generateNonce(),
			Challenge: DefaultChallengePeriod,
		}

		if state == nil {
			state = core.NewVoidState(asset, senderWallet)
		}
		newState := state.NextState()

		blockchainID, ok := c.homeBlockchains[asset]
		if !ok {
			if state.HomeLedger.BlockchainID != 0 {
				blockchainID = state.HomeLedger.BlockchainID
			} else {
				blockchainID, err = c.assetStore.GetSuggestedBlockchainID(asset)
				if err != nil {
					return "", err
				}
			}
		}

		// Get node address (Required for channel creation flow)
		nodeAddress, err := c.getNodeAddress(ctx)
		if err != nil {
			return "", err
		}

		// Get token address for this asset on this blockchain
		tokenAddress, err := c.getTokenAddress(ctx, blockchainID, asset)
		if err != nil {
			return "", err
		}

		// Initialize blockchain client if needed
		if err := c.initializeBlockchainClient(ctx, blockchainID); err != nil {
			return "", err
		}

		blockchainClient := c.blockchainClients[blockchainID]

		_, err = newState.ApplyChannelCreation(channelDef, blockchainID, tokenAddress, nodeAddress)
		if err != nil {
			return "", fmt.Errorf("failed to apply channel creation: %w", err)
		}

		// Apply transfer send transition
		_, err = newState.ApplyTransferSendTransition(recipientWallet, amount)
		if err != nil {
			return "", fmt.Errorf("failed to apply transfer transition: %w", err)
		}

		sig, err := c.SignState(newState)
		if err != nil {
			return "", fmt.Errorf("failed to sign state: %w", err)
		}
		newState.UserSig = &sig

		// Request channel creation from node
		nodeSig, err := c.requestChannelCreation(ctx, *newState, channelDef)
		if err != nil {
			return "", fmt.Errorf("failed to request channel creation: %w", err)
		}
		newState.NodeSig = &nodeSig

		// Create channel on blockchain (Smart Contract handles Creation + Withdrawal)
		txHash, err := blockchainClient.Create(channelDef, *newState)
		if err != nil {
			return "", fmt.Errorf("failed to create channel on blockchain: %w", err)
		}

		return txHash, nil
	}

	// Create next state
	nextState := state.NextState()

	// Apply transfer send transition
	transition, err := nextState.ApplyTransferSendTransition(recipientWallet, amount)
	if err != nil {
		return "", fmt.Errorf("failed to apply transfer transition: %w", err)
	}

	// Sign and submit state
	_, err = c.signAndSubmitState(ctx, nextState)
	if err != nil {
		return "", err
	}

	// Return transaction ID from the transition
	return transition.TxID, nil
}

// CloseChannel finalizes and closes the user's channel for a specific asset.
// This operation creates a final state and submits it to the node.
//
// Parameters:
//   - ctx: Context for the operation
//   - asset: The asset symbol to transfer (e.g., "usdc")
//
// Returns:
//   - Transaction ID for tracking
//   - Error if the operation fails
//
// Errors:
//   - Returns error if channel doesn't exist (user must deposit first)
//   - Returns error if state submission fails
//
// Example:
//
//	txID, err := client.CloseHomeChannel(ctx, "usdc")
//	fmt.Printf("CloseHomeChannel successful: %s\n", txID)
func (c *Client) CloseHomeChannel(ctx context.Context, asset string) (string, error) {
	// Get sender's latest state
	senderWallet := c.GetUserAddress()

	state, err := c.GetLatestState(ctx, senderWallet, asset, false)
	if err != nil {
		return "", err
	}

	if state.HomeChannelID == nil {
		return "", fmt.Errorf("no channel exists for asset %s", asset)
	}
	blockchainID := state.HomeLedger.BlockchainID

	// Initialize blockchain client if needed
	if err := c.initializeBlockchainClient(ctx, blockchainID); err != nil {
		return "", err
	}

	blockchainClient := c.blockchainClients[blockchainID]

	// Create next state
	nextState := state.NextState()

	// Apply finalize transition
	_, err = nextState.ApplyFinalizeTransition()
	if err != nil {
		return "", fmt.Errorf("failed to apply finalize transition: %w", err)
	}

	// Sign and submit state
	_, err = c.signAndSubmitState(ctx, nextState)
	if err != nil {
		return "", err
	}

	// Checkpoint on blockchain
	txHash, err := blockchainClient.Close(*nextState, nil)
	if err != nil {
		return "", fmt.Errorf("failed to close channel on blockchain: %w", err)
	}

	return txHash, nil
}

// Acknowledge sends an acknowledgement transition for the given asset.
// This is used when a user receives a transfer but hasn't yet acknowledged the state,
// or to acknowledge channel creation without a deposit.
//
// This method handles two scenarios automatically:
//  1. If no channel exists: Creates a new channel with the acknowledgement transition
//  2. If channel exists: Submits the acknowledgement transition to the existing channel
//
// Parameters:
//   - ctx: Context for the operation
//   - asset: The asset symbol to acknowledge (e.g., "usdc")
//
// Returns:
//   - Error if the operation fails
//
// Requirements:
//   - Home blockchain must be set for the asset (use SetHomeBlockchain) when no channel exists
//
// Example:
//
//	err := client.Acknowledge(ctx, "usdc")
func (c *Client) Acknowledge(ctx context.Context, asset string) error {
	userWallet := c.GetUserAddress()

	// Try to get latest state to determine if channel exists
	state, err := c.GetLatestState(ctx, userWallet, asset, false)

	// No channel path - create channel with acknowledgement
	if err != nil || state.HomeChannelID == nil {
		channelDef := core.ChannelDefinition{
			Nonce:     generateNonce(),
			Challenge: DefaultChallengePeriod,
		}

		if state == nil {
			state = core.NewVoidState(asset, userWallet)
		}
		newState := state.NextState()

		blockchainID, ok := c.homeBlockchains[asset]
		if !ok {
			return fmt.Errorf("home blockchain not set for asset %s", asset)
		}

		nodeAddress, err := c.getNodeAddress(ctx)
		if err != nil {
			return err
		}

		tokenAddress, err := c.getTokenAddress(ctx, blockchainID, asset)
		if err != nil {
			return err
		}

		_, err = newState.ApplyChannelCreation(channelDef, blockchainID, tokenAddress, nodeAddress)
		if err != nil {
			return fmt.Errorf("failed to apply channel creation: %w", err)
		}

		_, err = newState.ApplyAcknowledgementTransition()
		if err != nil {
			return fmt.Errorf("failed to apply acknowledgement transition: %w", err)
		}

		sig, err := c.SignState(newState)
		if err != nil {
			return fmt.Errorf("failed to sign state: %w", err)
		}
		newState.UserSig = &sig

		_, err = c.requestChannelCreation(ctx, *newState, channelDef)
		if err != nil {
			return fmt.Errorf("failed to request channel creation: %w", err)
		}

		return nil
	}

	if state.UserSig != nil {
		return fmt.Errorf("state already acknowledged by user")
	}

	// Has channel path - submit acknowledgement
	nextState := state.NextState()

	_, err = nextState.ApplyAcknowledgementTransition()
	if err != nil {
		return fmt.Errorf("failed to apply acknowledgement transition: %w", err)
	}

	_, err = c.signAndSubmitState(ctx, nextState)
	if err != nil {
		return err
	}

	return nil
}

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

	channel, err := transformChannel(resp.Channel)
	if err != nil {
		return nil, fmt.Errorf("failed to transform channel: %w", err)
	}
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

	channel, err := transformChannel(resp.Channel)
	if err != nil {
		return nil, fmt.Errorf("failed to transform channel: %w", err)
	}
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

// submitState submits a signed state update to the node.
// The state must be properly signed by the user before submission.
// This is an internal method used by high-level operations.
func (c *Client) submitState(ctx context.Context, state core.State) (string, error) {
	req := rpc.ChannelsV1SubmitStateRequest{
		State: transformStateToRPC(state),
	}
	resp, err := c.rpcClient.ChannelsV1SubmitState(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to submit state: %w", err)
	}
	return resp.Signature, nil
}

// requestChannelCreation requests the node to sign a channel creation.
// This is typically the first step when creating a new payment channel.
// This is an internal method used by the Deposit operation.
func (c *Client) requestChannelCreation(ctx context.Context, state core.State, channelDef core.ChannelDefinition) (string, error) {
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
