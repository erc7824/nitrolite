package main

import (
	"math/big"

	"github.com/erc7824/nitrolite/clearnode/nitrolite"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"gorm.io/gorm"
)

// ChannelService handles the business logic for funding channels.
type ChannelService struct {
	db     *gorm.DB
	signer *Signer
}

// NewAppSessionService creates a new AppSessionService.
func NewChannelService(db *gorm.DB, signer *Signer) *ChannelService {
	return &ChannelService{db: db, signer: signer}
}

func (s *ChannelService) RequestResize(logger Logger, params *ResizeChannelParams, rpcSigners map[string]struct{}) (ResizeChannelResponse, error) {
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
		params.ResizeAmount = big.NewInt(0)
	}
	if params.AllocateAmount == nil {
		params.AllocateAmount = big.NewInt(0)
	}

	// Prevent no-op resize operations
	if params.ResizeAmount.Cmp(big.NewInt(0)) == 0 && params.AllocateAmount.Cmp(big.NewInt(0)) == 0 {
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

	rawBalance := balance.Shift(int32(asset.Decimals)).BigInt()
	newChannelRawAmount := new(big.Int).Add(channel.RawAmount.BigInt(), params.AllocateAmount)

	if rawBalance.Cmp(newChannelRawAmount) < 0 {
		return ResizeChannelResponse{}, RPCErrorf("insufficient unified balance for channel %s: required %s, available %s", channel.ChannelID, newChannelRawAmount.String(), rawBalance.String())
	}
	newChannelRawAmount.Add(newChannelRawAmount, params.ResizeAmount)
	if newChannelRawAmount.Cmp(big.NewInt(0)) < 0 {
		return ResizeChannelResponse{}, RPCErrorf("new channel amount must be positive: %s", newChannelRawAmount.String())
	}

	allocations := []nitrolite.Allocation{
		{
			Destination: common.HexToAddress(params.FundsDestination),
			Token:       common.HexToAddress(channel.Token),
			Amount:      newChannelRawAmount,
		},
		{
			Destination: s.signer.GetAddress(),
			Token:       common.HexToAddress(channel.Token),
			Amount:      big.NewInt(0),
		},
	}

	resizeAmounts := []*big.Int{params.ResizeAmount, params.AllocateAmount}

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
	channelIDHash := common.HexToHash(channel.ChannelID)
	encodedState, err := nitrolite.EncodeState(channelIDHash, nitrolite.IntentRESIZE, big.NewInt(int64(channel.Version)+1), encodedIntentions, allocations)
	if err != nil {
		logger.Error("failed to encode state hash", "error", err)
		return ResizeChannelResponse{}, RPCErrorf("failed to encode state hash")
	}
	stateHash := crypto.Keccak256Hash(encodedState).Hex()
	sig, err := s.signer.NitroSign(encodedState)
	if err != nil {
		logger.Error("failed to sign state", "error", err)
		return ResizeChannelResponse{}, RPCErrorf("failed to sign state")
	}

	resp := ResizeChannelResponse{
		ChannelID: channel.ChannelID,
		Intent:    uint8(nitrolite.IntentRESIZE),
		Version:   channel.Version + 1,
		StateData: hexutil.Encode(encodedIntentions),
		StateHash: Hex(stateHash),
		Signature: Signature{
			V: sig.V,
			R: hexutil.Encode(sig.R[:]),
			S: hexutil.Encode(sig.S[:]),
		},
	}

	for _, alloc := range allocations {
		resp.Allocations = append(resp.Allocations, Allocation{
			Participant:  alloc.Destination.Hex(),
			TokenAddress: alloc.Token.Hex(),
			RawAmount:    alloc.Amount,
		})
	}
	return resp, nil
}

func (s *ChannelService) RequestClose(logger Logger, params *CloseChannelParams, rpcSigners map[string]struct{}) (CloseChannelResponse, error) {
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
	encodedState, err := nitrolite.EncodeState(common.HexToHash(channel.ChannelID), nitrolite.IntentFINALIZE, big.NewInt(int64(channel.Version)+1), stateDataBytes, allocations)
	if err != nil {
		logger.Error("failed to encode state hash", "error", err)
		return CloseChannelResponse{}, RPCErrorf("failed to encode state hash")
	}
	stateHash := crypto.Keccak256Hash(encodedState).Hex()
	sig, err := s.signer.NitroSign(encodedState)
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
		Signature: Signature{
			V: sig.V,
			R: hexutil.Encode(sig.R[:]),
			S: hexutil.Encode(sig.S[:]),
		},
	}

	for _, alloc := range allocations {
		resp.FinalAllocations = append(resp.FinalAllocations, Allocation{
			Participant:  alloc.Destination.Hex(),
			TokenAddress: alloc.Token.Hex(),
			RawAmount:    alloc.Amount,
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
