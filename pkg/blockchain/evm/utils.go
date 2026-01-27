package evm

import (
	"encoding/hex"
	"time"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// waitForBackOffTimeout implements exponential backoff between retries
func waitForBackOffTimeout(logger log.Logger, backOffCount int, originator string) {
	if backOffCount > maxBackOffCount {
		logger.Fatal("back off limit reached, exiting", "originator", originator, "backOffCollisionCount", backOffCount)
		return
	}

	if backOffCount > 0 {
		logger.Info("backing off", "originator", originator, "backOffCollisionCount", backOffCount)
		<-time.After(time.Duration(2^backOffCount-1) * time.Second)
	}
}

// ========= Client Helper Functions =========

func hexToBytes32(s string) ([32]byte, error) {
	var arr [32]byte
	b, err := hex.DecodeString(s)
	if err != nil {
		return arr, errors.Wrap(err, "failed to decode hex string")
	}
	if len(b) != 32 {
		return arr, errors.Errorf("invalid length: expected 32 bytes, got %d", len(b))
	}
	copy(arr[:], b)
	return arr, nil
}

func coreDefToContractDef(def core.ChannelDefinition, asset, userWallet string, nodeAddress common.Address) (ChannelDefinition, error) {
	return ChannelDefinition{
		ChallengeDuration: def.Challenge,
		User:              common.HexToAddress(userWallet),
		Node:              nodeAddress,
		Nonce:             def.Nonce,
		Metadata:          core.GenerateChannelMetadata(asset),
	}, nil
}

func coreStateToContractState(state core.State, tokenGetter func(blockchainID uint64, tokenAddress string) (uint8, error)) (State, error) {
	homeDecimals, err := tokenGetter(state.HomeLedger.BlockchainID, state.HomeLedger.TokenAddress)
	if err != nil {
		return State{}, errors.Wrap(err, "failed to get home token decimals")
	}

	homeLedger, err := coreLedgerToContractLedger(state.HomeLedger, homeDecimals)
	if err != nil {
		return State{}, errors.Wrap(err, "failed to convert home ledger")
	}

	var nonHomeLedger Ledger
	if state.EscrowLedger != nil {
		escrowDecimals, err := tokenGetter(state.EscrowLedger.BlockchainID, state.EscrowLedger.TokenAddress)
		if err != nil {
			return State{}, errors.Wrap(err, "failed to get escrow token decimals")
		}

		nonHomeLedger, err = coreLedgerToContractLedger(*state.EscrowLedger, escrowDecimals)
		if err != nil {
			return State{}, errors.Wrap(err, "failed to convert escrow ledger")
		}
	}

	var userSig, nodeSig []byte
	if state.UserSig != nil {
		userSig, err = hex.DecodeString(*state.UserSig)
		if err != nil {
			return State{}, errors.Wrap(err, "failed to decode user signature")
		}
	}
	if state.NodeSig != nil {
		nodeSig, err = hex.DecodeString(*state.NodeSig)
		if err != nil {
			return State{}, errors.Wrap(err, "failed to decode node signature")
		}
	}

	lastTransition := state.GetLastTransition()
	intent, err := core.TransitionToIntent(lastTransition)
	if err != nil {
		return State{}, err
	}

	metadata, err := core.GetStateTransitionsHash(state.Transitions)
	if err != nil {
		return State{}, errors.Wrap(err, "failed to compute state transitions hash")
	}

	return State{
		Version:      state.Version,
		Intent:       intent,
		Metadata:     metadata,
		HomeState:    homeLedger,
		NonHomeState: nonHomeLedger,
		UserSig:      userSig,
		NodeSig:      nodeSig,
	}, nil
}

func coreLedgerToContractLedger(ledger core.Ledger, decimals uint8) (Ledger, error) {
	tokenAddr := common.HexToAddress(ledger.TokenAddress)

	userAllocation, err := core.DecimalToBigInt(ledger.UserBalance, decimals)
	if err != nil {
		return Ledger{}, errors.Wrap(err, "failed to convert user balance to big.Int")
	}

	userNetFlow, err := core.DecimalToBigInt(ledger.UserNetFlow, decimals)
	if err != nil {
		return Ledger{}, errors.Wrap(err, "failed to convert user net flow to big.Int")
	}

	nodeAllocation, err := core.DecimalToBigInt(ledger.NodeBalance, decimals)
	if err != nil {
		return Ledger{}, errors.Wrap(err, "failed to convert node balance to big.Int")
	}

	nodeNetFlow, err := core.DecimalToBigInt(ledger.NodeNetFlow, decimals)
	if err != nil {
		return Ledger{}, errors.Wrap(err, "failed to convert node net flow to big.Int")
	}

	return Ledger{
		ChainId:        ledger.BlockchainID,
		Token:          tokenAddr,
		UserAllocation: userAllocation,
		UserNetFlow:    userNetFlow,
		NodeAllocation: nodeAllocation,
		NodeNetFlow:    nodeNetFlow,
	}, nil
}

func contractStateToCoreState(contractState State, homeChannelID string, escrowChannelID *string) (*core.State, error) {
	homeLedger := contractLedgerToCoreLedger(contractState.HomeState)

	var escrowLedger *core.Ledger
	if contractState.NonHomeState.ChainId != 0 {
		el := contractLedgerToCoreLedger(contractState.NonHomeState)
		escrowLedger = &el
	}

	var homeChannelIDPtr *string
	if homeChannelID != "" {
		homeChannelIDPtr = &homeChannelID
	}

	var userSig, nodeSig *string
	if len(contractState.UserSig) > 0 {
		sig := hex.EncodeToString(contractState.UserSig)
		userSig = &sig
	}
	if len(contractState.NodeSig) > 0 {
		sig := hex.EncodeToString(contractState.NodeSig)
		nodeSig = &sig
	}

	return &core.State{
		Version:         contractState.Version,
		HomeChannelID:   homeChannelIDPtr,
		EscrowChannelID: escrowChannelID,
		HomeLedger:      homeLedger,
		EscrowLedger:    escrowLedger,
		UserSig:         userSig,
		NodeSig:         nodeSig,
		// Note: ID, Transitions, Asset, UserWallet, Epoch are not available from contract state
		// These may need to be populated separately or passed as parameters
	}, nil
}

func contractLedgerToCoreLedger(ledger Ledger) core.Ledger {
	// NOTE: consider YN decimals when using
	exp := -int32(ledger.Decimals)
	return core.Ledger{
		BlockchainID: ledger.ChainId,
		TokenAddress: ledger.Token.Hex(),
		UserBalance:  decimal.NewFromBigInt(ledger.UserAllocation, exp),
		UserNetFlow:  decimal.NewFromBigInt(ledger.UserNetFlow, exp),
		NodeBalance:  decimal.NewFromBigInt(ledger.NodeAllocation, exp),
		NodeNetFlow:  decimal.NewFromBigInt(ledger.NodeNetFlow, exp),
	}
}
