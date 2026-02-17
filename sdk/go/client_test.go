package sdk

import (
	"context"
	"testing"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetHomeChannel(t *testing.T) {
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockResp := rpc.ChannelsV1GetHomeChannelResponse{
		Channel: rpc.ChannelV1{
			ChannelID:    "0xChannelID",
			UserWallet:   "0xWallet",
			Type:         "home",
			BlockchainID: "137",
			Status:       "open",
			StateVersion: "1",
			Nonce:        "1",
		},
	}
	mockDialer.RegisterResponse(rpc.ChannelsV1GetHomeChannelMethod.String(), mockResp)

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	ch, err := client.GetHomeChannel(context.Background(), "0xWallet", "USDC")
	require.NoError(t, err)
	assert.Equal(t, "0xChannelID", ch.ChannelID)
	assert.Equal(t, core.ChannelTypeHome, ch.Type)
}

func TestClient_GetEscrowChannel(t *testing.T) {
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockResp := rpc.ChannelsV1GetEscrowChannelResponse{
		Channel: rpc.ChannelV1{
			ChannelID:    "0xEscrowID",
			UserWallet:   "0xWallet",
			Type:         "escrow",
			BlockchainID: "137",
			Status:       "open",
			StateVersion: "1",
			Nonce:        "1",
		},
	}
	mockDialer.RegisterResponse(rpc.ChannelsV1GetEscrowChannelMethod.String(), mockResp)

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	ch, err := client.GetEscrowChannel(context.Background(), "0xEscrowID")
	require.NoError(t, err)
	assert.Equal(t, "0xEscrowID", ch.ChannelID)
	assert.Equal(t, core.ChannelTypeEscrow, ch.Type)
}

func TestClient_GetLatestState(t *testing.T) {
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockResp := rpc.ChannelsV1GetLatestStateResponse{
		State: rpc.StateV1{
			ID:         "0xStateID",
			Epoch:      "1",
			Version:    "1",
			UserWallet: "0xWallet",
			Asset:      "USDC",
			Transition: rpc.TransitionV1{
				Type:   core.TransitionTypeTransferSend,
				Amount: "10.0",
			},
			HomeLedger: rpc.LedgerV1{
				BlockchainID: "137",
				UserBalance:  "100.0",
				UserNetFlow:  "0",
				NodeBalance:  "200.0",
				NodeNetFlow:  "0",
			},
		},
	}
	mockDialer.RegisterResponse(rpc.ChannelsV1GetLatestStateMethod.String(), mockResp)

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	state, err := client.GetLatestState(context.Background(), "0xWallet", "USDC", false)
	require.NoError(t, err)
	assert.Equal(t, "0xStateID", state.ID)
	assert.Equal(t, uint64(1), state.Version)
}

func TestClient_GetBalances(t *testing.T) {
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockResp := rpc.UserV1GetBalancesResponse{
		Balances: []rpc.BalanceEntryV1{
			{Asset: "USDC", Amount: "100.0"},
		},
	}
	mockDialer.RegisterResponse(rpc.UserV1GetBalancesMethod.String(), mockResp)

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	bals, err := client.GetBalances(context.Background(), "0xWallet")
	require.NoError(t, err)
	assert.Len(t, bals, 1)
	assert.Equal(t, "USDC", bals[0].Asset)
	assert.Equal(t, "100", bals[0].Balance.String())
}

func TestClient_GetTransactions(t *testing.T) {
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockResp := rpc.UserV1GetTransactionsResponse{
		Transactions: []rpc.TransactionV1{
			{ID: "0xTxID", Asset: "USDC", Amount: "50.0", CreatedAt: "2023-01-01T00:00:00Z"},
		},
		Metadata: rpc.PaginationMetadataV1{
			TotalCount: 1,
		},
	}
	mockDialer.RegisterResponse(rpc.UserV1GetTransactionsMethod.String(), mockResp)

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	txs, meta, err := client.GetTransactions(context.Background(), "0xWallet", nil)
	require.NoError(t, err)
	assert.Len(t, txs, 1)
	assert.Equal(t, "0xTxID", txs[0].ID)
	assert.Equal(t, uint32(1), meta.TotalCount)
}

func TestClient_GetAppSessions(t *testing.T) {
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockResp := rpc.AppSessionsV1GetAppSessionsResponse{
		AppSessions: []rpc.AppSessionInfoV1{
			{
				AppSessionID: "0xSessionID",
				Participants: []rpc.AppParticipantV1{},
				Allocations:  []rpc.AppAllocationV1{},
				Status:       "open",
				Nonce:        "1",
				Version:      "1",
			},
		},
		Metadata: rpc.PaginationMetadataV1{TotalCount: 1},
	}
	mockDialer.RegisterResponse(rpc.AppSessionsV1GetAppSessionsMethod.String(), mockResp)

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	sessions, meta, err := client.GetAppSessions(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Equal(t, "0xSessionID", sessions[0].AppSessionID)
	assert.Equal(t, uint32(1), meta.TotalCount)
}

func TestClient_GetAppDefinition(t *testing.T) {
	mockDialer := NewMockDialer()
	mockDialer.Dial(context.Background(), "", nil)

	mockResp := rpc.AppSessionsV1GetAppDefinitionResponse{
		Definition: rpc.AppDefinitionV1{
			Application:  "0xApp",
			Participants: []rpc.AppParticipantV1{},
			Nonce:        "1",
			Quorum:       1,
		},
	}
	mockDialer.RegisterResponse(rpc.AppSessionsV1GetAppDefinitionMethod.String(), mockResp)

	client := &Client{
		rpcClient: rpc.NewClient(mockDialer),
	}

	def, err := client.GetAppDefinition(context.Background(), "0xSessionID")
	require.NoError(t, err)
	assert.Equal(t, "0xApp", def.Application)
	assert.Equal(t, uint64(1), def.Nonce)
}
