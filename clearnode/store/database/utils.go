package database

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
)

// databaseChannelToCore converts database.Channel to core.Channel
func databaseChannelToCore(dbChannel *Channel) *core.Channel {
	return &core.Channel{
		ChannelID:          dbChannel.ChannelID,
		UserWallet:         dbChannel.UserWallet,
		Type:               dbChannel.Type,
		BlockchainID:       dbChannel.BlockchainID,
		TokenAddress:       dbChannel.Token,
		ChallengeDuration:  dbChannel.ChallengeDuration,
		ChallengeExpiresAt: dbChannel.ChallengeExpiresAt,
		Nonce:              dbChannel.Nonce,
		Status:             dbChannel.Status,
		StateVersion:       dbChannel.StateVersion,
	}
}

// databaseAppSessionToCore converts database.AppSessionV1 to app.AppSessionV1
func databaseAppSessionToCore(dbSession *AppSessionV1) *app.AppSessionV1 {
	participants := make([]app.AppParticipantV1, len(dbSession.Participants))
	for i, p := range dbSession.Participants {
		participants[i] = app.AppParticipantV1{
			WalletAddress:   p.WalletAddress,
			SignatureWeight: p.SignatureWeight,
		}
	}

	return &app.AppSessionV1{
		SessionID:    dbSession.ID,
		Application:  dbSession.Application,
		Participants: participants,
		Quorum:       dbSession.Quorum,
		Nonce:        dbSession.Nonce,
		Status:       dbSession.Status,
		Version:      dbSession.Version,
		SessionData:  dbSession.SessionData,
		CreatedAt:    dbSession.CreatedAt,
		UpdatedAt:    dbSession.UpdatedAt,
	}
}

// databaseStateToCore converts database.State to core.State
func databaseStateToCore(dbState *State) (*core.State, error) {
	var transitions []core.Transition
	if err := json.Unmarshal([]byte(dbState.Transitions), &transitions); err != nil {
		return nil, fmt.Errorf("cannot unmarshal transitions: %w", err)
	}

	state := &core.State{
		ID:              dbState.ID,
		Transitions:     transitions,
		Asset:           dbState.Asset,
		UserWallet:      dbState.UserWallet,
		Epoch:           dbState.Epoch,
		Version:         dbState.Version,
		HomeChannelID:   dbState.HomeChannelID,
		EscrowChannelID: dbState.EscrowChannelID,
		HomeLedger: core.Ledger{
			UserBalance: dbState.HomeUserBalance,
			UserNetFlow: decimal.NewFromInt(dbState.HomeUserNetFlow),
			NodeBalance: dbState.HomeNodeBalance,
			NodeNetFlow: decimal.NewFromInt(dbState.HomeNodeNetFlow),
		},
	}

	if dbState.EscrowChannelID != nil {
		state.EscrowLedger = &core.Ledger{
			UserBalance: dbState.EscrowUserBalance,
			UserNetFlow: decimal.NewFromInt(dbState.EscrowUserNetFlow),
			NodeBalance: dbState.EscrowNodeBalance,
			NodeNetFlow: decimal.NewFromInt(dbState.EscrowNodeNetFlow),
		}
	}

	if dbState.UserSig != nil {
		state.UserSig = dbState.UserSig
	}
	if dbState.NodeSig != nil {
		state.NodeSig = dbState.NodeSig
	}

	return state, nil
}

// coreStateToDB converts core.State to database.State
func coreStateToDB(state *core.State) (*State, error) {
	bytes, err := json.Marshal(state.Transitions)
	if err != nil {
		return nil, fmt.Errorf("marshal checkpoint data: %w", err)
	}

	dbState := &State{
		ID:              state.ID,
		Transitions:     bytes,
		Asset:           state.Asset,
		UserWallet:      state.UserWallet,
		Epoch:           state.Epoch,
		Version:         state.Version,
		HomeChannelID:   state.HomeChannelID,
		EscrowChannelID: state.EscrowChannelID,
		HomeUserBalance: state.HomeLedger.UserBalance,
		HomeUserNetFlow: state.HomeLedger.UserNetFlow.IntPart(),
		HomeNodeBalance: state.HomeLedger.NodeBalance,
		HomeNodeNetFlow: state.HomeLedger.NodeNetFlow.IntPart(),
		CreatedAt:       time.Now(),
	}

	if state.EscrowLedger != nil {
		dbState.EscrowUserBalance = state.EscrowLedger.UserBalance
		dbState.EscrowUserNetFlow = state.EscrowLedger.UserNetFlow.IntPart()
		dbState.EscrowNodeBalance = state.EscrowLedger.NodeBalance
		dbState.EscrowNodeNetFlow = state.EscrowLedger.NodeNetFlow.IntPart()
	}

	if state.UserSig != nil {
		dbState.UserSig = state.UserSig
	}
	if state.NodeSig != nil {
		dbState.NodeSig = state.NodeSig
	}

	return dbState, nil
}

func toCoreTransaction(dbTx *Transaction) *core.Transaction {
	return &core.Transaction{
		ID:                 dbTx.ID,
		Asset:              dbTx.AssetSymbol,
		TxType:             dbTx.Type,
		FromAccount:        dbTx.FromAccount,
		ToAccount:          dbTx.ToAccount,
		SenderNewStateID:   dbTx.SenderNewStateID,
		ReceiverNewStateID: dbTx.ReceiverNewStateID,
		Amount:             dbTx.Amount,
		CreatedAt:          dbTx.CreatedAt,
	}
}

// calculatePaginationMetadata computes pagination metadata from total count, offset, and limit
func calculatePaginationMetadata(totalCount int64, offset, limit uint32) core.PaginationMetadata {
	pageCount := uint32((totalCount + int64(limit) - 1) / int64(limit))
	currentPage := uint32(1)
	if limit > 0 {
		currentPage = offset/limit + 1
	}

	return core.PaginationMetadata{
		Page:       currentPage,
		PerPage:    limit,
		TotalCount: uint32(totalCount),
		PageCount:  pageCount,
	}
}
