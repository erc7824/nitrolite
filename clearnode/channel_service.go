package main

import (
	"fmt"
	"math/big"
	"time"

	"github.com/erc7824/nitrolite/clearnode/nitrolite"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ChannelService handles the business logic for funding channels.
type ChannelService struct {
	db       *gorm.DB
	networks map[string]*NetworkConfig
	signer   *Signer
}

// NewAppSessionService creates a new AppSessionService.
func NewChannelService(db *gorm.DB, networks map[string]*NetworkConfig, signer *Signer) *ChannelService {
	return &ChannelService{db: db, networks: networks, signer: signer}
}

func (s *ChannelService) RequestCreate(wallet common.Address, params *CreateChannelParams, rpcSigners map[string]struct{}, logger Logger) (ChannelOperationResponse, error) {
	_, ok := rpcSigners[wallet.Hex()]
	if !ok {
		return ChannelOperationResponse{}, RPCErrorf("invalid signature")
	}

	existingOpenChannel, err := CheckExistingChannels(s.db, wallet.Hex(), params.Token, params.ChainID)
	if err != nil {
		return ChannelOperationResponse{}, RPCErrorf("failed to check existing channels")
	}
	if existingOpenChannel != nil {
		return ChannelOperationResponse{}, RPCErrorf("an open channel with broker already exists: %s", existingOpenChannel.ChannelID)
	}

	if _, err := GetAssetByToken(s.db, params.Token, params.ChainID); err != nil {
		return ChannelOperationResponse{}, RPCErrorf("token not supported: %s", params.Token)
	}

	allocations := []nitrolite.Allocation{
		{
			Destination: wallet,
			Token:       common.HexToAddress(params.Token),
			Amount:      params.Amount.BigInt(),
		},
		{
			Destination: s.signer.GetAddress(),
			Token:       common.HexToAddress(params.Token),
			Amount:      big.NewInt(0),
		},
	}

	networkConfig, ok := s.networks[fmt.Sprintf("%d", params.ChainID)]
	if !ok {
		return ChannelOperationResponse{}, RPCErrorf("unsupported chain ID: %d", params.ChainID)
	}

	channel := nitrolite.Channel{
		Participants: []common.Address{wallet, s.signer.GetAddress()},
		Adjudicator:  common.HexToAddress(networkConfig.AdjudicatorAddress),
		Challenge:    3600,
		Nonce:        uint64(time.Now().UnixMilli()),
	}

	channelID, err := nitrolite.GetChannelID(channel, params.ChainID)
	if err != nil {
		logger.Error("failed to get channel ID", "error", err)
		return ChannelOperationResponse{}, RPCErrorf("failed to get channel ID")
	}

	stateDataHex := "0x"
	stateDataBytes, err := hexutil.Decode(stateDataHex)
	if err != nil {
		logger.Error("failed to decode state data hex", "error", err)
		return ChannelOperationResponse{}, RPCErrorf("failed to decode state data hex")
	}

	state := nitrolite.State{
		Intent:      uint8(nitrolite.IntentINITIALIZE),
		Version:     big.NewInt(0),
		Data:        stateDataBytes,
		Allocations: allocations,
	}

	stateData, err := nitrolite.EncodeStateData(channelID, state)
	if err != nil {
		logger.Error("error encoding state hash", "error", err)
		return ChannelOperationResponse{}, RPCErrorf("error encoding state hash")
	}

	sig, err := s.signer.Sign(stateData)
	if err != nil {
		logger.Error("failed to sign state", "error", err)
		return ChannelOperationResponse{}, RPCErrorf("failed to sign state")
	}

	return createChannelOperationResponse(channelID.Hex(), crypto.Keccak256Hash(stateData).Hex(), state, sig), nil
}

func (s *ChannelService) RequestResize(params *ResizeChannelParams, rpcSigners map[string]struct{}, logger Logger) (ChannelOperationResponse, error) {
	channel, err := GetChannelByID(s.db, params.ChannelID)
	if err != nil {
		logger.Error("failed to find channel", "error", err)
		return ChannelOperationResponse{}, RPCErrorf("failed to find channel: %s", params.ChannelID)
	}
	if channel == nil {
		return ChannelOperationResponse{}, RPCErrorf("channel %s not found", params.ChannelID)
	}

	if err = checkChallengedChannels(s.db, channel.Wallet); err != nil {
		logger.Error("failed to check challenged channels", "error", err)
		return ChannelOperationResponse{}, err
	}

	if channel.Status != ChannelStatusOpen {
		return ChannelOperationResponse{}, RPCErrorf("channel %s is not open: %s", params.ChannelID, channel.Status)
	}

	_, ok := rpcSigners[channel.Wallet]
	if !ok {
		return ChannelOperationResponse{}, RPCErrorf("invalid signature")
	}

	asset, err := GetAssetByToken(s.db, channel.Token, channel.ChainID)
	if err != nil {
		logger.Error("failed to find asset", "error", err)
		return ChannelOperationResponse{}, RPCErrorf("failed to find asset for token %s on chain %d", channel.Token, channel.ChainID)
	}

	if params.ResizeAmount == nil {
		params.ResizeAmount = &decimal.Zero
	}
	if params.AllocateAmount == nil {
		params.AllocateAmount = &decimal.Zero
	}

	// Prevent no-op resize operations
	if params.ResizeAmount.Cmp(decimal.Zero) == 0 && params.AllocateAmount.Cmp(decimal.Zero) == 0 {
		return ChannelOperationResponse{}, RPCErrorf("resize operation requires non-zero ResizeAmount or AllocateAmount")
	}

	userAddress := common.HexToAddress(channel.Wallet)
	userAccountID := NewAccountID(channel.Wallet)
	ledger := GetWalletLedger(s.db, userAddress)
	balance, err := ledger.Balance(userAccountID, asset.Symbol)
	if err != nil {
		logger.Error("failed to check participant balance", "error", err)
		return ChannelOperationResponse{}, RPCErrorf("failed to check participant balance for asset %s", asset.Symbol)
	}

	rawBalance := balance.Shift(int32(asset.Decimals))
	newChannelRawAmount := channel.RawAmount.Add(*params.AllocateAmount)

	if rawBalance.Cmp(newChannelRawAmount) < 0 {
		return ChannelOperationResponse{}, RPCErrorf("insufficient unified balance for channel %s: required %s, available %s", channel.ChannelID, newChannelRawAmount.String(), rawBalance.String())
	}
	newChannelRawAmount = newChannelRawAmount.Add(*params.ResizeAmount)
	if newChannelRawAmount.Cmp(decimal.Zero) < 0 {
		return ChannelOperationResponse{}, RPCErrorf("new channel amount must be positive: %s", newChannelRawAmount.String())
	}

	allocations := []nitrolite.Allocation{
		{
			Destination: common.HexToAddress(params.FundsDestination),
			Token:       common.HexToAddress(channel.Token),
			Amount:      newChannelRawAmount.BigInt(),
		},
		{
			Destination: s.signer.GetAddress(),
			Token:       common.HexToAddress(channel.Token),
			Amount:      big.NewInt(0),
		},
	}

	resizeAmounts := []*big.Int{params.ResizeAmount.BigInt(), params.AllocateAmount.BigInt()}

	intentionType, err := abi.NewType("int256[]", "", nil)
	if err != nil {
		logger.Fatal("failed to create intention type", "error", err)
		return ChannelOperationResponse{}, RPCErrorf("failed to create intention type")
	}
	intentionArgs := abi.Arguments{{Type: intentionType}}
	encodedIntentions, err := intentionArgs.Pack(resizeAmounts)
	if err != nil {
		logger.Error("failed to pack resize amounts", "error", err)
		return ChannelOperationResponse{}, RPCErrorf("failed to pack resize amounts")
	}

	// Encode & sign the new state
	state := nitrolite.State{
		Intent:      uint8(nitrolite.IntentRESIZE),
		Version:     big.NewInt(int64(channel.Version) + 1),
		Data:        encodedIntentions,
		Allocations: allocations,
	}

	channelIDHash := common.HexToHash(channel.ChannelID)
	stateData, err := nitrolite.EncodeStateData(channelIDHash, state)
	if err != nil {
		logger.Error("failed to encode state hash", "error", err)
		return ChannelOperationResponse{}, RPCErrorf("failed to encode state hash")
	}

	sig, err := s.signer.Sign(stateData)
	if err != nil {
		logger.Error("failed to sign state", "error", err)
		return ChannelOperationResponse{}, RPCErrorf("failed to sign state")
	}

	return createChannelOperationResponse(channel.ChannelID, crypto.Keccak256Hash(stateData).Hex(), state, sig), nil
}

func (s *ChannelService) RequestClose(params *CloseChannelParams, rpcSigners map[string]struct{}, logger Logger) (ChannelOperationResponse, error) {
	channel, err := GetChannelByID(s.db, params.ChannelID)
	if err != nil {
		logger.Error("failed to find channel", "error", err)
		return ChannelOperationResponse{}, RPCErrorf("failed to find channel")
	}
	if channel == nil {
		return ChannelOperationResponse{}, RPCErrorf("channel not found")
	}

	if err = checkChallengedChannels(s.db, channel.Wallet); err != nil {
		logger.Error("failed to check challenged channels", "error", err)
		return ChannelOperationResponse{}, err
	}

	if channel.Status != ChannelStatusOpen {
		return ChannelOperationResponse{}, RPCErrorf("channel %s is not open: %s", params.ChannelID, channel.Status)
	}

	_, ok := rpcSigners[channel.Wallet]
	if !ok {
		return ChannelOperationResponse{}, RPCErrorf("invalid signature")
	}

	asset, err := GetAssetByToken(s.db, channel.Token, channel.ChainID)
	if err != nil {
		logger.Error("failed to find asset", "error", err)
		return ChannelOperationResponse{}, RPCErrorf("failed to find asset for token %s on chain %d", channel.Token, channel.ChainID)
	}

	userAddress := common.HexToAddress(channel.Wallet)
	userAccountID := NewAccountID(channel.Wallet)
	ledger := GetWalletLedger(s.db, userAddress)
	balance, err := ledger.Balance(userAccountID, asset.Symbol)
	if err != nil {
		logger.Error("failed to check participant balance", "error", err)
		return ChannelOperationResponse{}, RPCErrorf("failed to check participant balance")
	}
	if balance.IsNegative() {
		logger.Error("negative balance", "balance", balance.String())
		return ChannelOperationResponse{}, RPCErrorf("negative balance")
	}

	rawBalance := balance.Shift(int32(asset.Decimals)).BigInt()
	channelRawAmount := channel.RawAmount.BigInt()
	if channelRawAmount.Cmp(rawBalance) < 0 {
		return ChannelOperationResponse{}, RPCErrorf("resize this channel first")
	}

	allocations := []nitrolite.Allocation{
		{
			Destination: common.HexToAddress(params.FundsDestination),
			Token:       common.HexToAddress(channel.Token),
			Amount:      rawBalance,
		},
		{
			Destination: s.signer.GetAddress(),
			Token:       common.HexToAddress(channel.Token),
			Amount:      new(big.Int).Sub(channelRawAmount, rawBalance),
		},
	}

	stateDataHex := "0x"
	stateDataBytes, err := hexutil.Decode(stateDataHex)
	if err != nil {
		logger.Error("failed to decode state data hex", "error", err)
		return ChannelOperationResponse{}, RPCErrorf("failed to decode state data hex")
	}

	state := nitrolite.State{
		Intent:      uint8(nitrolite.IntentFINALIZE),
		Version:     big.NewInt(int64(channel.Version) + 1),
		Data:        stateDataBytes,
		Allocations: allocations,
	}

	stateData, err := nitrolite.EncodeStateData(common.HexToHash(channel.ChannelID), state)
	if err != nil {
		logger.Error("failed to encode state hash", "error", err)
		return ChannelOperationResponse{}, RPCErrorf("failed to encode state hash")
	}

	sig, err := s.signer.Sign(stateData)
	if err != nil {
		logger.Error("failed to sign state", "error", err)
		return ChannelOperationResponse{}, RPCErrorf("failed to sign state")
	}

	return createChannelOperationResponse(channel.ChannelID, crypto.Keccak256Hash(stateData).Hex(), state, sig), nil
}

// checkChallengedChannels checks if the participant has any channels in the challenged state.
func checkChallengedChannels(tx *gorm.DB, wallet string) error {
	// As this method is also used other handlers as part of transactions, it stays separate from the channel service.
	challengedChannels, err := getChannelsByWallet(tx, wallet, string(ChannelStatusChallenged))
	if err != nil {
		return RPCErrorf("failed to check challenged channels")
	}
	if len(challengedChannels) > 0 {
		return RPCErrorf("participant %s has challenged channels, cannot execute operation", wallet)
	}
	return nil
}

func createChannelOperationResponse(channelID string, stateHash string, state nitrolite.State, signature Signature) ChannelOperationResponse {
	resp := ChannelOperationResponse{
		ChannelID: channelID,
		StateHash: stateHash,
		State: UnsignedState{
			Intent:  uint8(state.Intent),
			Version: state.Version.Uint64(),
			Data:    state.Data,
		},
		StateSignature: signature,
	}
	for _, alloc := range state.Allocations {
		resp.State.Allocations = append(resp.State.Allocations, Allocation{
			Participant:  alloc.Destination.Hex(),
			TokenAddress: alloc.Token.Hex(),
			RawAmount:    decimal.NewFromBigInt(alloc.Amount, 0),
		})
	}
	return resp
}
