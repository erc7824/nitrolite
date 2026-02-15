package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
)

// databaseChannelToCore converts database.Channel to core.Channel
func databaseChannelToCore(dbChannel *Channel) *core.Channel {
	return &core.Channel{
		ChannelID:             dbChannel.ChannelID,
		UserWallet:            dbChannel.UserWallet,
		Type:                  dbChannel.Type,
		BlockchainID:          dbChannel.BlockchainID,
		TokenAddress:          dbChannel.Token,
		ApprovedSigValidators: dbChannel.ApprovedSigValidators,
		ChallengeDuration:     dbChannel.ChallengeDuration,
		ChallengeExpiresAt:    dbChannel.ChallengeExpiresAt,
		Nonce:                 dbChannel.Nonce,
		Status:                dbChannel.Status,
		StateVersion:          dbChannel.StateVersion,
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

// stateRowToCore converts database.StateRow (from unified queries) to core.State
func stateRowToCore(row *StateRow) (*core.State, error) {
	// Build home ledger with blockchain ID and token address from joined channels
	homeLedger := core.Ledger{
		UserBalance: row.HomeUserBalance,
		UserNetFlow: row.HomeUserNetFlow,
		NodeBalance: row.HomeNodeBalance,
		NodeNetFlow: row.HomeNodeNetFlow,
	}

	// If home channel ID exists, blockchain ID and token address must be present
	if row.HomeChannelID != nil {
		if row.HomeBlockchainID == nil || row.HomeTokenAddress == nil {
			return nil, fmt.Errorf("home channel %s exists but blockchain ID or token address is missing", *row.HomeChannelID)
		}
		homeLedger.BlockchainID = *row.HomeBlockchainID
		homeLedger.TokenAddress = *row.HomeTokenAddress
	}

	transition := core.Transition{
		Type:      core.TransitionType(row.TransitionType),
		TxID:      row.TransitionTxID,
		AccountID: row.TransitionAccountID,
		Amount:    row.TransitionAmount,
	}

	state := &core.State{
		ID:              row.ID,
		Transition:      transition,
		Asset:           row.Asset,
		UserWallet:      row.UserWallet,
		Epoch:           row.Epoch,
		Version:         row.Version,
		HomeChannelID:   row.HomeChannelID,
		EscrowChannelID: row.EscrowChannelID,
		HomeLedger:      homeLedger,
	}

	// If escrow channel ID exists, blockchain ID and token address must be present
	if row.EscrowChannelID != nil {
		if row.EscrowBlockchainID == nil || row.EscrowTokenAddress == nil {
			return nil, fmt.Errorf("escrow channel %s exists but blockchain ID or token address is missing", *row.EscrowChannelID)
		}
		state.EscrowLedger = &core.Ledger{
			BlockchainID: *row.EscrowBlockchainID,
			TokenAddress: *row.EscrowTokenAddress,
			UserBalance:  row.EscrowUserBalance,
			UserNetFlow:  row.EscrowUserNetFlow,
			NodeBalance:  row.EscrowNodeBalance,
			NodeNetFlow:  row.EscrowNodeNetFlow,
		}
	}

	if row.UserSig != nil {
		state.UserSig = row.UserSig
	}
	if row.NodeSig != nil {
		state.NodeSig = row.NodeSig
	}

	return state, nil
}

// coreStateToDB converts core.State to database.State
func coreStateToDB(state *core.State) (*State, error) {
	dbState := &State{
		ID:                  strings.ToLower(state.ID),
		TransitionType:      uint8(state.Transition.Type),
		TransitionTxID:      strings.ToLower(state.Transition.TxID),
		TransitionAccountID: strings.ToLower(state.Transition.AccountID),
		TransitionAmount:    state.Transition.Amount,
		Asset:               state.Asset,
		UserWallet:          strings.ToLower(state.UserWallet),
		Epoch:               state.Epoch,
		Version:             state.Version,
		HomeUserBalance:     state.HomeLedger.UserBalance,
		HomeUserNetFlow:     state.HomeLedger.UserNetFlow,
		HomeNodeBalance:     state.HomeLedger.NodeBalance,
		HomeNodeNetFlow:     state.HomeLedger.NodeNetFlow,
		CreatedAt:           time.Now(),
	}
	if state.HomeChannelID != nil {
		lowered := strings.ToLower(*state.HomeChannelID)
		dbState.HomeChannelID = &lowered
	}
	if state.EscrowChannelID != nil {
		lowered := strings.ToLower(*state.EscrowChannelID)
		dbState.EscrowChannelID = &lowered
	}

	if state.EscrowLedger != nil {
		dbState.EscrowUserBalance = state.EscrowLedger.UserBalance
		dbState.EscrowUserNetFlow = state.EscrowLedger.UserNetFlow
		dbState.EscrowNodeBalance = state.EscrowLedger.NodeBalance
		dbState.EscrowNodeNetFlow = state.EscrowLedger.NodeNetFlow
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

// stateRowSelectColumns returns the SELECT clause for state row queries,
// with appropriate aliasing based on whether it's from the head table or history table.
func stateRowSelectColumns(tableAlias string, fromHead bool) string {
	idColumn := "id"
	if fromHead {
		idColumn = "history_id"
	}
	return fmt.Sprintf(`
		%s.%s AS id,
		%s.user_wallet, %s.asset, %s.epoch, %s.version,
		%s.transition_type, %s.transition_tx_id, %s.transition_account_id, %s.transition_amount,
		%s.home_channel_id, %s.escrow_channel_id,
		%s.home_user_balance, %s.home_user_net_flow, %s.home_node_balance, %s.home_node_net_flow,
		%s.escrow_user_balance, %s.escrow_user_net_flow, %s.escrow_node_balance, %s.escrow_node_net_flow,
		%s.user_sig, %s.node_sig`,
		tableAlias, idColumn,
		tableAlias, tableAlias, tableAlias, tableAlias, // user_wallet, asset, epoch, version
		tableAlias, tableAlias, tableAlias, tableAlias, // transition fields
		tableAlias, tableAlias, // channel IDs
		tableAlias, tableAlias, tableAlias, tableAlias, // home balances
		tableAlias, tableAlias, tableAlias, tableAlias, // escrow balances
		tableAlias, tableAlias, // signatures
	)
}

// channelJoinsFragment returns the LEFT JOIN clauses for channels table.
// sourceAlias is the table alias containing home_channel_id and escrow_channel_id.
func channelJoinsFragment(sourceAlias string) string {
	return fmt.Sprintf(`
		LEFT JOIN channels hc ON %s.home_channel_id = hc.channel_id
		LEFT JOIN channels ec ON %s.escrow_channel_id = ec.channel_id`,
		sourceAlias, sourceAlias,
	)
}

// channelSelectColumns returns the blockchain_id and token columns from joined channels.
func channelSelectColumns() string {
	return `
		hc.blockchain_id AS home_blockchain_id, hc.token AS home_token_address,
		ec.blockchain_id AS escrow_blockchain_id, ec.token AS escrow_token_address`
}
