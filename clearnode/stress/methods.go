package stress

import (
	"context"
	"fmt"
	"strconv"

	"github.com/erc7824/nitrolite/pkg/core"
	sdk "github.com/erc7824/nitrolite/sdk/go"
)

// MethodRegistry returns all available stress test methods.
func MethodRegistry() map[string]Factory {
	return map[string]Factory{
		"ping":                   stressPing,
		"get-config":             stressGetConfig,
		"get-blockchains":        stressGetBlockchains,
		"get-assets":             stressGetAssets,
		"get-balances":           stressGetBalances,
		"get-transactions":       stressGetTransactions,
		"get-home-channel":       stressGetHomeChannel,
		"get-escrow-channel":     stressGetEscrowChannel,
		"get-latest-state":       stressGetLatestState,
		"get-channel-key-states": stressGetLastChannelKeyStates,
		"get-app-sessions":       stressGetAppSessions,
		"get-app-key-states":     stressGetLastAppKeyStates,
	}
}

func stressPing(_ []string, _ string) (MethodFunc, error) {
	return func(ctx context.Context, client *sdk.Client) error {
		return client.Ping(ctx)
	}, nil
}

func stressGetConfig(_ []string, _ string) (MethodFunc, error) {
	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetConfig(ctx)
		return err
	}, nil
}

func stressGetBlockchains(_ []string, _ string) (MethodFunc, error) {
	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetBlockchains(ctx)
		return err
	}, nil
}

func stressGetAssets(args []string, _ string) (MethodFunc, error) {
	var chainID *uint64
	if len(args) >= 1 {
		parsed, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid chain_id: %s", args[0])
		}
		chainID = &parsed
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetAssets(ctx, chainID)
		return err
	}, nil
}

func stressGetBalances(args []string, walletAddress string) (MethodFunc, error) {
	wallet := walletAddress
	if len(args) >= 1 {
		wallet = args[0]
	}
	if wallet == "" {
		return nil, fmt.Errorf("wallet address required: provide as extra param or set STRESS_PRIVATE_KEY")
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetBalances(ctx, wallet)
		return err
	}, nil
}

func stressGetTransactions(args []string, walletAddress string) (MethodFunc, error) {
	wallet := walletAddress
	if len(args) >= 1 {
		wallet = args[0]
	}
	if wallet == "" {
		return nil, fmt.Errorf("wallet address required: provide as extra param or set STRESS_PRIVATE_KEY")
	}

	limit := uint32(20)
	opts := &sdk.GetTransactionsOptions{
		Pagination: &core.PaginationParams{
			Limit: &limit,
		},
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, _, err := client.GetTransactions(ctx, wallet, opts)
		return err
	}, nil
}

func stressGetHomeChannel(args []string, walletAddress string) (MethodFunc, error) {
	var wallet, asset string

	switch len(args) {
	case 2:
		wallet = args[0]
		asset = args[1]
	case 1:
		wallet = walletAddress
		if wallet == "" {
			return nil, fmt.Errorf("wallet address required: provide as extra param or set STRESS_PRIVATE_KEY")
		}
		asset = args[0]
	default:
		return nil, fmt.Errorf("usage: get-home-channel requires asset param, e.g. get-home-channel:1000:10:usdc")
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetHomeChannel(ctx, wallet, asset)
		return err
	}, nil
}

func stressGetEscrowChannel(args []string, _ string) (MethodFunc, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("usage: get-escrow-channel requires channel_id param")
	}
	channelID := args[0]

	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetEscrowChannel(ctx, channelID)
		return err
	}, nil
}

func stressGetLatestState(args []string, walletAddress string) (MethodFunc, error) {
	var wallet, asset string

	switch len(args) {
	case 2:
		wallet = args[0]
		asset = args[1]
	case 1:
		wallet = walletAddress
		if wallet == "" {
			return nil, fmt.Errorf("wallet address required: provide as extra param or set STRESS_PRIVATE_KEY")
		}
		asset = args[0]
	default:
		return nil, fmt.Errorf("usage: get-latest-state requires asset param, e.g. get-latest-state:1000:10:usdc")
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetLatestState(ctx, wallet, asset, false)
		return err
	}, nil
}

func stressGetLastChannelKeyStates(args []string, walletAddress string) (MethodFunc, error) {
	wallet := walletAddress
	if len(args) >= 1 {
		wallet = args[0]
	}
	if wallet == "" {
		return nil, fmt.Errorf("wallet address required: provide as extra param or set STRESS_PRIVATE_KEY")
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetLastChannelKeyStates(ctx, wallet, nil)
		return err
	}, nil
}

func stressGetAppSessions(args []string, walletAddress string) (MethodFunc, error) {
	wallet := walletAddress
	if len(args) >= 1 {
		wallet = args[0]
	}

	limit := uint32(20)
	opts := &sdk.GetAppSessionsOptions{
		Pagination: &core.PaginationParams{
			Limit: &limit,
		},
	}
	if wallet != "" {
		opts.Participant = &wallet
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, _, err := client.GetAppSessions(ctx, opts)
		return err
	}, nil
}

func stressGetLastAppKeyStates(args []string, walletAddress string) (MethodFunc, error) {
	wallet := walletAddress
	if len(args) >= 1 {
		wallet = args[0]
	}
	if wallet == "" {
		return nil, fmt.Errorf("wallet address required: provide as extra param or set STRESS_PRIVATE_KEY")
	}

	return func(ctx context.Context, client *sdk.Client) error {
		_, err := client.GetLastAppKeyStates(ctx, wallet, nil)
		return err
	}, nil
}
