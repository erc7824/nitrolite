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

func (s *ChannelService) RequestCreate(wallet common.Address, params *CreateChannelParams, rpcSigners map[string]struct{}, logger Logger) (CreateChannelResponse, error) {
	_, ok := rpcSigners[wallet.Hex()]
	if !ok {
		return CreateChannelResponse{}, RPCErrorf("invalid signature")
	}

	existingOpenChannel, err := CheckExistingChannels(s.db, wallet.Hex(), params.Token, params.ChainID)
	if err != nil {
		return CreateChannelResponse{}, RPCErrorf("failed to check existing channels")
	}
	if existingOpenChannel != nil {
		return CreateChannelResponse{}, RPCErrorf("an open channel with broker already exists: %s", existingOpenChannel.ChannelID)
	}

	if _, err := GetAssetByToken(s.db, params.Token, params.ChainID); err != nil {
		return CreateChannelResponse{}, RPCErrorf("token not supported: %s", params.Token)
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
		return CreateChannelResponse{}, RPCErrorf("unsupported chain ID: %d", params.ChainID)
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
		return CreateChannelResponse{}, RPCErrorf("failed to get channel ID")
	}

	stateDataHex := "0x"
	stateDataBytes, err := hexutil.Decode(stateDataHex)
	if err != nil {
		logger.Error("failed to decode state data hex", "error", err)
		return CreateChannelResponse{}, RPCErrorf("failed to decode state data hex")
	}

	state := nitrolite.State{
		Intent:      uint8(nitrolite.IntentINITIALIZE),
		Version:     big.NewInt(0),
		Data:        stateDataBytes,
		Allocations: allocations,
	}

	encodedState, err := nitrolite.EncodeState(channelID, state)
	if err != nil {
		logger.Error("error encoding state hash", "error", err)
		return CreateChannelResponse{}, RPCErrorf("error encoding state hash")
	}

	stateHash := crypto.Keccak256Hash(encodedState).Hex()
	sig, err := s.signer.Sign(encodedState)
	if err != nil {
		logger.Error("failed to sign state", "error", err)
		return CreateChannelResponse{}, RPCErrorf("failed to sign state")
	}

	resp := CreateChannelResponse{
		ChannelID: channelID.Hex(),
		StateHash: stateHash,
		State: State{
			Intent:  uint8(nitrolite.IntentINITIALIZE),
			Version: 0,
			Data:    stateDataBytes,
			Sigs:    []Signature{sig},
		},
	}

	for _, alloc := range allocations {
		resp.State.Allocations = append(resp.State.Allocations, Allocation{
			Participant:  alloc.Destination.Hex(),
			TokenAddress: alloc.Token.Hex(),
			RawAmount:    decimal.NewFromBigInt(alloc.Amount, 0),
		})
	}

	return resp, nil
}

func (s *ChannelService) RequestResize(params *ResizeChannelParams, rpcSigners map[string]struct{}, logger Logger) (ResizeChannelResponse, error) {
	channel, err := GetChannelByID(s.db, params.ChannelID)
	if err != nil {
		logger.Error("failed to find channel", "error", err)
		return ResizeChannelResponse{}, RPCErrorf("failed to find channel: %s", params.ChannelID)
	}
	if channel == nil {
		return ResizeChannelResponse{}, RPCErrorf("channel %s not found", params.ChannelID)
	}

	if err = checkChallengedChannels(s.db, channel.Wallet); err != nil {
		logger.Error("failed to check challenged channels", "error", err)
		return ResizeChannelResponse{}, err
	}

	if channel.Status != ChannelStatusOpen {
		return ResizeChannelResponse{}, RPCErrorf("channel %s is not open: %s", params.ChannelID, channel.Status)
	}

	_, ok := rpcSigners[channel.Wallet]
	if !ok {
		return ResizeChannelResponse{}, RPCErrorf("invalid signature")
	}

	asset, err := GetAssetByToken(s.db, channel.Token, channel.ChainID)
	if err != nil {
		logger.Error("failed to find asset", "error", err)
		return ResizeChannelResponse{}, RPCErrorf("failed to find asset for token %s on chain %d", channel.Token, channel.ChainID)
	}

	if params.ResizeAmount == nil {
		params.ResizeAmount = &decimal.Zero
	}
	if params.AllocateAmount == nil {
		params.AllocateAmount = &decimal.Zero
	}

	// Prevent no-op resize operations
	if params.ResizeAmount.Cmp(decimal.Zero) == 0 && params.AllocateAmount.Cmp(decimal.Zero) == 0 {
		return ResizeChannelResponse{}, RPCErrorf("resize operation requires non-zero ResizeAmount or AllocateAmount")
	}

	userAddress := common.HexToAddress(channel.Wallet)
	userAccountID := NewAccountID(channel.Wallet)
	ledger := GetWalletLedger(s.db, userAddress)
	balance, err := ledger.Balance(userAccountID, asset.Symbol)
	if err != nil {
		logger.Error("failed to check participant balance", "error", err)
		return ResizeChannelResponse{}, RPCErrorf("failed to check participant balance for asset %s", asset.Symbol)
	}

	rawBalance := balance.Shift(int32(asset.Decimals))
	newChannelRawAmount := channel.RawAmount.Add(*params.AllocateAmount)

	if rawBalance.Cmp(newChannelRawAmount) < 0 {
		return ResizeChannelResponse{}, RPCErrorf("insufficient unified balance for channel %s: required %s, available %s", channel.ChannelID, newChannelRawAmount.String(), rawBalance.String())
	}
	newChannelRawAmount = newChannelRawAmount.Add(*params.ResizeAmount)
	if newChannelRawAmount.Cmp(decimal.Zero) < 0 {
		return ResizeChannelResponse{}, RPCErrorf("new channel amount must be positive: %s", newChannelRawAmount.String())
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
		return ResizeChannelResponse{}, RPCErrorf("failed to create intention type")
	}
	intentionArgs := abi.Arguments{{Type: intentionType}}
	encodedIntentions, err := intentionArgs.Pack(resizeAmounts)
	if err != nil {
		logger.Error("failed to pack resize amounts", "error", err)
		return ResizeChannelResponse{}, RPCErrorf("failed to pack resize amounts")
	}

	// 6) Encode & sign the new state
	state := nitrolite.State{
		Intent:      uint8(nitrolite.IntentRESIZE),
		Version:     big.NewInt(int64(channel.Version) + 1),
		Data:        encodedIntentions,
		Allocations: allocations,
	}

	channelIDHash := common.HexToHash(channel.ChannelID)
	encodedState, err := nitrolite.EncodeState(channelIDHash, state)
	if err != nil {
		logger.Error("failed to encode state hash", "error", err)
		return ResizeChannelResponse{}, RPCErrorf("failed to encode state hash")
	}
	stateHash := crypto.Keccak256Hash(encodedState).Hex()
	sig, err := s.signer.Sign(encodedState)
	if err != nil {
		logger.Error("failed to sign state", "error", err)
		return ResizeChannelResponse{}, RPCErrorf("failed to sign state")
	}

	resp := ResizeChannelResponse{
		ChannelID: channel.ChannelID,
		Intent:    uint8(nitrolite.IntentRESIZE),
		Version:   channel.Version + 1,
		StateData: hexutil.Encode(encodedIntentions),
		StateHash: stateHash,
		Signature: sig,
	}

	for _, alloc := range allocations {
		resp.Allocations = append(resp.Allocations, Allocation{
			Participant:  alloc.Destination.Hex(),
			TokenAddress: alloc.Token.Hex(),
			RawAmount:    decimal.NewFromBigInt(alloc.Amount, 0),
		})
	}
	return resp, nil
}

func (s *ChannelService) RequestClose(params *CloseChannelParams, rpcSigners map[string]struct{}, logger Logger) (CloseChannelResponse, error) {
	channel, err := GetChannelByID(s.db, params.ChannelID)
	if err != nil {
		logger.Error("failed to find channel", "error", err)
		return CloseChannelResponse{}, RPCErrorf("failed to find channel")
	}
	if channel == nil {
		return CloseChannelResponse{}, RPCErrorf("channel not found")
	}

	if err = checkChallengedChannels(s.db, channel.Wallet); err != nil {
		logger.Error("failed to check challenged channels", "error", err)
		return CloseChannelResponse{}, err
	}

	if channel.Status != ChannelStatusOpen {
		return CloseChannelResponse{}, RPCErrorf("channel %s is not open: %s", params.ChannelID, channel.Status)
	}

	_, ok := rpcSigners[channel.Wallet]
	if !ok {
		return CloseChannelResponse{}, RPCErrorf("invalid signature")
	}

	asset, err := GetAssetByToken(s.db, channel.Token, channel.ChainID)
	if err != nil {
		logger.Error("failed to find asset", "error", err)
		return CloseChannelResponse{}, RPCErrorf("failed to find asset for token %s on chain %d", channel.Token, channel.ChainID)
	}

	userAddress := common.HexToAddress(channel.Wallet)
	userAccountID := NewAccountID(channel.Wallet)
	ledger := GetWalletLedger(s.db, userAddress)
	balance, err := ledger.Balance(userAccountID, asset.Symbol)
	if err != nil {
		logger.Error("failed to check participant balance", "error", err)
		return CloseChannelResponse{}, RPCErrorf("failed to check participant balance")
	}
	if balance.IsNegative() {
		logger.Error("negative balance", "balance", balance.String())
		return CloseChannelResponse{}, RPCErrorf("negative balance")
	}

	rawBalance := balance.Shift(int32(asset.Decimals)).BigInt()
	channelRawAmount := channel.RawAmount.BigInt()
	if channelRawAmount.Cmp(rawBalance) < 0 {
		return CloseChannelResponse{}, RPCErrorf("resize this channel first")
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
		return CloseChannelResponse{}, RPCErrorf("failed to decode state data hex")
	}

	state := nitrolite.State{
		Intent:      uint8(nitrolite.IntentFINALIZE),
		Version:     big.NewInt(int64(channel.Version) + 1),
		Data:        stateDataBytes,
		Allocations: allocations,
	}

	encodedState, err := nitrolite.EncodeState(common.HexToHash(channel.ChannelID), state)
	if err != nil {
		logger.Error("failed to encode state hash", "error", err)
		return CloseChannelResponse{}, RPCErrorf("failed to encode state hash")
	}
	stateHash := crypto.Keccak256Hash(encodedState).Hex()
	sig, err := s.signer.Sign(encodedState)
	if err != nil {
		logger.Error("failed to sign state", "error", err)
		return CloseChannelResponse{}, RPCErrorf("failed to sign state")
	}

	resp := CloseChannelResponse{
		ChannelID: channel.ChannelID,
		Intent:    uint8(nitrolite.IntentFINALIZE),
		Version:   channel.Version + 1,
		StateData: stateDataHex,
		StateHash: stateHash,
		Signature: sig,
	}

	for _, alloc := range allocations {
		resp.FinalAllocations = append(resp.FinalAllocations, Allocation{
			Participant:  alloc.Destination.Hex(),
			TokenAddress: alloc.Token.Hex(),
			RawAmount:    decimal.NewFromBigInt(alloc.Amount, 0),
		})
	}

	return resp, nil
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
