// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package evm

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// ChannelDefinition is an auto generated low-level Go binding around an user-defined struct.
type ChannelDefinition struct {
	ChallengeDuration           uint32
	User                        common.Address
	Node                        common.Address
	Nonce                       uint64
	ApprovedSignatureValidators *big.Int
	Metadata                    [32]byte
}

// Ledger is an auto generated low-level Go binding around an user-defined struct.
type Ledger struct {
	ChainId        uint64
	Token          common.Address
	Decimals       uint8
	UserAllocation *big.Int
	UserNetFlow    *big.Int
	NodeAllocation *big.Int
	NodeNetFlow    *big.Int
}

// State is an auto generated low-level Go binding around an user-defined struct.
type State struct {
	Version       uint64
	Intent        uint8
	Metadata      [32]byte
	HomeLedger    Ledger
	NonHomeLedger Ledger
	UserSig       []byte
	NodeSig       []byte
}

// ChannelHubMetaData contains all meta data concerning the ChannelHub contract.
var ChannelHubMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_defaultSigValidator\",\"type\":\"address\",\"internalType\":\"contractISignatureValidator\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"DEFAULT_SIG_VALIDATOR\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractISignatureValidator\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"ESCROW_DEPOSIT_UNLOCK_DELAY\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MAX_DEPOSIT_ESCROW_PURGE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MIN_CHALLENGE_DURATION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"TRANSFER_GAS_LIMIT\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"challengeChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengerSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"challengerIdx\",\"type\":\"uint8\",\"internalType\":\"enumParticipantIndex\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"challengeEscrowDeposit\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"challengerSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"challengerIdx\",\"type\":\"uint8\",\"internalType\":\"enumParticipantIndex\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"challengeEscrowWithdrawal\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"challengerSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"challengerIdx\",\"type\":\"uint8\",\"internalType\":\"enumParticipantIndex\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"checkpointChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"claimFunds\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"destination\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"closeChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"createChannel\",\"inputs\":[{\"name\":\"def\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"approvedSignatureValidators\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"initState\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"depositToChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"depositToVault\",\"inputs\":[{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"escrowHead\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"finalizeEscrowDeposit\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"finalizeEscrowWithdrawal\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"finalizeMigration\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getAccountBalance\",\"inputs\":[{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getChannelData\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumChannelStatus\"},{\"name\":\"definition\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"approvedSignatureValidators\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"lastState\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengeExpiry\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"lockedFunds\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getChannelIds\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEscrowDepositData\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumEscrowStatus\"},{\"name\":\"unlockAt\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"challengeExpiry\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"lockedAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"initState\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEscrowDepositIds\",\"inputs\":[{\"name\":\"page\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"pageSize\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"ids\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEscrowWithdrawalData\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumEscrowStatus\"},{\"name\":\"challengeExpiry\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"lockedAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"initState\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getNodeValidator\",\"inputs\":[{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"validatorId\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractISignatureValidator\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOpenChannels\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"openChannels\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getReclaimBalance\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initiateEscrowDeposit\",\"inputs\":[{\"name\":\"def\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"approvedSignatureValidators\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"initiateEscrowWithdrawal\",\"inputs\":[{\"name\":\"def\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"approvedSignatureValidators\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"initiateMigration\",\"inputs\":[{\"name\":\"def\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"approvedSignatureValidators\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"purgeEscrowDeposits\",\"inputs\":[{\"name\":\"maxToPurge\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"registerNodeValidator\",\"inputs\":[{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"validatorId\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"validator\",\"type\":\"address\",\"internalType\":\"contractISignatureValidator\"},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"withdrawFromChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"withdrawFromVault\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"ChannelChallenged\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengeExpireAt\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelCheckpointed\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelClosed\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"finalState\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelCreated\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"user\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"definition\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"approvedSignatureValidators\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"initialState\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelDeposited\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelWithdrawn\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Deposited\",\"inputs\":[{\"name\":\"wallet\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositChallenged\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengeExpireAt\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositFinalized\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositFinalizedOnHome\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositInitiated\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositInitiatedOnHome\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositsPurged\",\"inputs\":[{\"name\":\"purgedCount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalChallenged\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengeExpireAt\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalFinalized\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalFinalizedOnHome\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalInitiated\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalInitiatedOnHome\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"FundsClaimed\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"destination\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MigrationInFinalized\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MigrationInInitiated\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MigrationOutFinalized\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MigrationOutInitiated\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TransferFailed\",\"inputs\":[{\"name\":\"recipient\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ValidatorRegistered\",\"inputs\":[{\"name\":\"node\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"validatorId\",\"type\":\"uint8\",\"indexed\":true,\"internalType\":\"uint8\"},{\"name\":\"validator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"contractISignatureValidator\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Withdrawn\",\"inputs\":[{\"name\":\"wallet\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AddressCollision\",\"inputs\":[{\"name\":\"collision\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ChallengerVersionTooLow\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureLength\",\"inputs\":[{\"name\":\"length\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureS\",\"inputs\":[{\"name\":\"s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"EmptySignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectAmount\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectChallengeDuration\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectChannelStatus\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectStateIntent\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectValue\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InsufficientBalance\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAddress\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidValidatorId\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NativeTransferFailed\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"OnlyNonHomeEscrowsCanBeChallenged\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"SafeCastOverflowedIntToUint\",\"inputs\":[{\"name\":\"value\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"type\":\"error\",\"name\":\"SafeERC20FailedOperation\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ValidatorAlreadyRegistered\",\"inputs\":[{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"validatorId\",\"type\":\"uint8\",\"internalType\":\"uint8\"}]},{\"type\":\"error\",\"name\":\"ValidatorNotApproved\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ValidatorNotRegistered\",\"inputs\":[{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"validatorId\",\"type\":\"uint8\",\"internalType\":\"uint8\"}]}]",
	Bin: "0x60a0346100aa57601f61604f38819003918201601f19168301916001600160401b038311848410176100ae578084926020946040528339810103126100aa57516001600160a01b0381168082036100aa5760017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f00551561009b57608052604051615f8c90816100c3823960805181818161114f0152614ac40152f35b63e6c4247b60e01b5f5260045ffd5b5f80fd5b634e487b7160e01b5f52604160045260245ffdfe60806040526004361015610011575f80fd5b5f3560e01c806313c380ed1461024457806316b390b11461023f57806317536c061461023a578063187576d8146102355780633115f6301461023057806338a66be21461022b57806341b660ef1461022657806347de477a14610221578063532691981461021c578063587675e8146102175780635a0745b4146102125780635b9acbf91461020d5780635dc46a74146102085780636898234b146102035780636af820bd146101fe57806371a47141146101f9578063735181f0146101f45780637e7985f9146101ef57806382d3e15d146101ea5780638d0b12a5146101e557806394191051146101e05780639691b468146101db578063a5c82680146101d6578063b00b6fd6146101d1578063b25a1d38146101cc578063beed9d5f146101c7578063c74a2d10146101c2578063d888ccae146101bd578063dc23f29e146101b8578063dd73d494146101b3578063e617208c146101ae578063ecf3d7e8146101a9578063f4ac51f5146101a4578063f766f8d61461019f5763ffa1ad741461019a575f80fd5b612680565b6124e1565b612466565b612354565b6122b5565b612137565b611f3b565b611dfb565b611bb7565b611a11565b611701565b611698565b6114fd565b6113a3565b611386565b6111ef565b6111d2565b6111bb565b61117e565b61113a565b61111f565b611033565b611021565b610fff565b610fe3565b610f9d565b610d3b565b610b82565b6108a6565b610852565b6106c6565b610640565b610532565b610326565b61028f565b90816102609103126102585790565b5f80fd5b90604060031983011261025857600435916024359067ffffffffffffffff82116102585761028c91600401610249565b90565b34610258576102a66102a03661025c565b90612c38565b005b9181601f840112156102585782359167ffffffffffffffff8311610258576020838186019501011161025857565b60643590600282101561025857565b90606060031983011261025857600435916024359067ffffffffffffffff821161025857610315916004016102a8565b909160443560028110156102585790565b3461025857610334366102e5565b61040061034c859493945f52600260205260405f2090565b92835461036761036261035e83613f07565b1590565b612e9a565b6103c2600286019461039261038387546001600160a01b031690565b948560038a019a8b5492614aab565b9591600160068b019a01966103b288546001600160a01b039060081c1690565b926103bc8c61300e565b88614bfb565b60c06103cd886142d5565b604051809581927f6666e4c000000000000000000000000000000000000000000000000000000000835260048301613092565b038173__$682d6198b4eca5bc7e038b912a26498e7e$__5af49081156104f0577fba075bd445233f7cad862c72f0343b3503aad9c8e704a2295f122b82abf8e8019661049e956080955f946104b3575b508361047c61046e61048f9697546001600160a01b039060081c1690565b92546001600160a01b031690565b9254936104888a61300e565b908c6146aa565b015167ffffffffffffffff1690565b906104ae604051928392836130a3565b0390a2005b61048f94509461047c6104e061046e9760c03d60c0116104e9575b6104d8818361292e565b8101906129a5565b95505094610450565b503d6104ce565b612a9f565b6001600160a01b0381160361025857565b60031960609101126102585760043561051e816104f5565b9060243561052b816104f5565b9060443590565b6001600160a01b0361054336610506565b9290911690610553821515613200565b61055e83151561322f565b815f5260066020526105848160405f20906001600160a01b03165f5260205260405f2090565b805491848301809311610602577f8752a472e571a816aea92eec8dae9baf628e840f4929fbcc2d155e6233ff68a7926001600160a01b0392556105c5615951565b6105d0858233614d31565b60017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f00556040519485521692602090a3005b61325e565b60206040818301928281528451809452019201905f5b81811061062a5750505090565b825184526020938401939092019160010161061d565b34610258576020600319360112610258576001600160a01b03600435610665816104f5565b165f52600160205260405f206040519081602082549182815201915f5260205f20905f5b8181106106b0576106ac856106a08187038261292e565b60405191829182610607565b0390f35b8254845260209093019260019283019201610689565b3461025857602060031936011261025857600354600480549190355f5b82841080610849575b1561083c5761071861070a61070086613496565b90549060031b1c90565b5f52600260205260405f2090565b600181016003610729825460ff1690565b61073281611c46565b1461082a5761074082615bbd565b156107e657915f826107d16107dd9560056107d79601926107c26107ba8554926107a5600d61079361077c60028501546001600160a01b031690565b6001600160a01b03165f52600660205260405f2090565b92015460401c6001600160a01b031690565b6001600160a01b03165f5260205260405f2090565b918254613272565b9055600360ff19825416179055565b55613982565b93613982565b915b91926106e3565b50509290506107f59150600455565b806107fc57005b6040519081527f61815f4b11c6ea4e14a2e448a010bed8efdc3e53a15efbf183d16a31085cd14590602090a1005b50509261083690613982565b916107df565b9290506107f59150600455565b508181106106ec565b34610258575f600319360112610258576020604051620186a08152f35b90600319820160e081126102585760c0136102585760049160c4359067ffffffffffffffff82116102585761028c91600401610249565b6108af3661086f565b90602082019160026108c08461287e565b6108c981611c55565b148015610b67575b8015610b49575b6108e190612888565b6109896108f66108f1368561328d565b614f1d565b9161090084614fdb565b602084019061090e826128b7565b9561092d6040870197610920896128b7565b608089013591858961428c565b60e08261095561094e61094261077c8c6128b7565b6107a5608085016128b7565b54886150e6565b60405196879283927fa8b4483c000000000000000000000000000000000000000000000000000000008452600484016133d1565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49384156104f057610a846001600160a01b0394610a9c93610a057fb00e209e275d0e1892f1982b34d3f545d1628aebd95322d7ce3585c558f638b498610a8a955f91610b1a575b506109f4368d61328d565b6109fe368a612b8a565b908c61527e565b610a3189610a2c610a15856128b7565b6001600160a01b03165f52600160205260405f2090565b615bf1565b506002610a3d8261287e565b610a4681611c55565b03610aa15750877f6085f5128b19e0d3cc37524413de47259383f0f75265d5d66f4177869620669660405180610a7c898261286d565b0390a26128b7565b976128b7565b918360405194859416981696836133e8565b0390a4005b610aac60039161287e565b610ab581611c55565b03610aea57877f188e0ade7d115cc397426774adb960ae3e8c83e72f0a6cad4b7085e1d60bf98660405180610a7c898261286d565b877f567044ba1cdd4671ac3979c114241e1e3b56c9e9051f63f2f234f7a2795019cc60405180610a7c898261286d565b610b3c915060e03d60e011610b42575b610b34818361292e565b810190613301565b5f6109e9565b503d610b2a565b506108e1610b568461287e565b610b5f81611c55565b1590506108d8565b506003610b738461287e565b610b7c81611c55565b146108d1565b610b8b3661086f565b90610bac6004610b9d6020850161287e565b610ba681611c55565b14612888565b610bb96108f1368361328d565b9160208201610bc7816128b7565b90610be860408501926080610bdb856128b7565b960135958691868961428c565b610bfa610bf484613478565b86615391565b93610c0486613f07565b15610c4c57505050610c4781610c3b7f471c4ebe4e57d25ef7117e141caac31c6b98f067b8098a7a7bbd38f637c2f98093866153c1565b6040519182918261286d565b0390a3005b610c949060c085610c6088979596976142d5565b60405194859283927fbbc42f3400000000000000000000000000000000000000000000000000000000845260048401612a7a565b038173__$682d6198b4eca5bc7e038b912a26498e7e$__5af480156104f0577fede7867afa7cdb9c443667efd8244d98bf9df1dce68e60dc94dca6605125ca7695610c4795610d09945f93610d12575b50610cf1610cf7916128b7565b916128b7565b91610d023687612b8a565b8a8a6146aa565b610c3b846134e5565b610cf7919350610d33610cf19160c03d60c0116104e9576104d8818361292e565b939150610ce4565b3461025857610d493661025c565b610d5a6009610b9d6020840161287e565b610d766001610d70845f525f60205260405f2090565b01613533565b610e55610d8d60208301516001600160a01b031690565b9161077c60e060408301610db9610dab82516001600160a01b031690565b608086015190888a8c61428c565b610e21610e1a610e03610dcc368b612b8a565b9586946101408c018d8d610ddf83613478565b67ffffffffffffffff1646149d8e610f0f575b50505050516001600160a01b031690565b6060840151602001516001600160a01b03166107a5565b54896150e6565b60405195869283927fa8b4483c000000000000000000000000000000000000000000000000000000008452600484016135bf565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49182156104f057610e87935f93610eee575b508661527e565b15610ebd576104ae7f9a6f675cc94b83b55f1ecc0876affd4332a30c92e6faa2aca0199b1b6df922c3916040519182918261286d565b6104ae7f7b20773c41402791c5f18914dbbeacad38b1ebcc4c55d8eb3bfe0a4cde26c826916040519182918261286d565b610f0891935060e03d60e011610b4257610b34818361292e565b915f610e80565b610f33610f6992610f24610f6e963690612aaa565b60608d01526060369101612aaa565b60808b0152610f406135ab565b60a08b0152610f4d6135ab565b60c08b01526001600160a01b03165f52600160205260405f2090565b615c9a565b505f8d8d82610df2565b600319604091011261025857600435610f90816104f5565b9060243561028c816104f5565b34610258576020610fda6001600160a01b03610fb836610f78565b91165f526006835260405f20906001600160a01b03165f5260205260405f2090565b54604051908152f35b34610258575f600319360112610258576020604051612a308152f35b34610258576040600319360112610258576106ac6106a0602435600435613669565b6102a661102d3661025c565b90613723565b34610258576020600319360112610258576001600160a01b03600435611058816104f5565b165f52600160205261106c60405f20615b75565b5f905f5b815181101561110c5761109e6110976110898385613655565b515f525f60205260405f2090565b5460ff1690565b6110a781612208565b600381141590816110f7575b506110c1575b600101611070565b916110d48184600193106110dc57613982565b9290506110b9565b6110e68585613655565b516110f18286613655565b52613982565b6005915061110481612208565b14155f6110b3565b506106ac91815260405191829182610607565b34610258575f60031936011261025857602060405160408152f35b34610258575f600319360112610258576040517f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03168152602090f35b34610258576020610fda6001600160a01b0361119936610f78565b91165f526008835260405f20906001600160a01b03165f5260205260405f2090565b34610258576102a66111cc3661025c565b90613a5e565b34610258575f600319360112610258576020600454604051908152f35b346102585761127a611200366102e5565b92939190611216855f52600560205260405f2090565b9161122761036261035e8554613f07565b6002830160a061124761124183546001600160a01b031690565b896153dc565b604051809781927f24063eba00000000000000000000000000000000000000000000000000000000835260048301613c4b565b038173__$b69fb814c294bfc16f92e50d7aeced4bde$__5af49485156104f0575f95611355575b508354600185015490949060081c6001600160a01b0316968792546112cc906001600160a01b031690565b809581956003850154976112e1928992614aab565b9a9190946006019a6112f28c61300e565b956112fd968b614bfb565b846113078761300e565b6113119589615444565b6060015167ffffffffffffffff1660405191829161132f91836130a3565b037fb8568a1f475f3c76759a620e08a653d28348c5c09e2e0bc91d533339801fefd891a2005b61137891955060a03d60a01161137f575b611370818361292e565b810190613990565b935f6112a1565b503d611366565b34610258575f600319360112610258576020604051620151808152f35b6114796113af3661025c565b6113d06113c16020839594950161287e565b6113ca81611c55565b15612888565b6113e66001610d70855f525f60205260405f2090565b9060e08161144561143e61094261077c61140a60208901516001600160a01b031690565b6114318b8a604081019389608061142887516001600160a01b031690565b9301519361428c565b516001600160a01b031690565b54876150e6565b60405195869283927fa8b4483c000000000000000000000000000000000000000000000000000000008452600484016133d1565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49283156104f0577f567044ba1cdd4671ac3979c114241e1e3b56c9e9051f63f2f234f7a2795019cc936104ae93610c3b925f926114dc575b506114d53685612b8a565b908761527e565b6114f691925060e03d60e011610b4257610b34818361292e565b905f6114ca565b346102585761150b3661086f565b9061151d6006610b9d6020850161287e565b61152a6108f1368361328d565b9160208201611538816128b7565b9061154c60408501926080610bdb856128b7565b611558610bf484613478565b9361156286613f07565b1561159957505050610c4781610c3b7f587faad1bcd589ce902468251883e1976a645af8563c773eed7356d78433210c93866153c1565b6115e29060a0856115af611241879896976128b7565b60405194859283927eea54e700000000000000000000000000000000000000000000000000000000845260048401613a47565b038173__$b69fb814c294bfc16f92e50d7aeced4bde$__5af480156104f0577f17eb0a6bd5a0de45d1029ce3444941070e149df35b22176fc439f930f73c09f795610c4795610c3b945f93611651575b50610cf161163f916128b7565b9161164a3687612b8a565b8a8a615444565b61163f919350611672610cf19160a03d60a01161137f57611370818361292e565b939150611632565b6024359060ff8216820361025857565b359060ff8216820361025857565b34610258576040600319360112610258576001600160a01b036116e66004356116c0816104f5565b826116c961167a565b91165f52600760205260405f209060ff165f5260205260405f2090565b541660405180916001600160a01b0360208301911682520390f35b60806003193601126102585760043560243567ffffffffffffffff811161025857611730903690600401610249565b60443567ffffffffffffffff8111610258576117509036906004016102a8565b919061175a6102d6565b9261176c855f525f60205260405f2090565b9161177960018401613533565b916117a3611788855460ff1690565b61179181612208565b600181149081156119fd575b50613c5c565b866117b06005860161300e565b926117f16117bd88613478565b67ffffffffffffffff6117e86117db885167ffffffffffffffff1690565b67ffffffffffffffff1690565b91161015613c8b565b60208501516001600160a01b0316926040860161181581516001600160a01b031690565b9567ffffffffffffffff61183d6117db61182e8d613478565b935167ffffffffffffffff1690565b911611611903575b50946118a77f07b9206d5a6026d3bd2a8f9a9b79f6fa4bfbd6a016975829fbaf07488019f28a99866118d5966118a16118c5978d61189560149f9e9c996118bc9a6118f49f60808c015192614aab565b93919490923690612b8a565b90614bfb565b845460ff191660021785555163ffffffff1690565b63ffffffff1690565b67ffffffffffffffff4216613cba565b9301805467ffffffffffffffff191667ffffffffffffffff8516179055565b6104ae60405192839283613cdc565b8861144561193f61094261077c60e0956114318c9f9e9c9a6119469b8d9f9a9b9d6119336113c160208c0161287e565b8960808d01519361428c565b548d6150e6565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49384156104f0577f07b9206d5a6026d3bd2a8f9a9b79f6fa4bfbd6a016975829fbaf07488019f28a99896118f4986118a18e8c61189560149f976118a7986118bc9b6119c66118d59f6118c59f5f916119de575b508d6119bf3689612b8a565b9089615581565b9a9f5050995050509750509650969899509950611845565b6119f7915060e03d60e011610b4257610b34818361292e565b5f6119b3565b60049150611a0a81612208565b145f61179d565b3461025857608060031936011261025857600435611a2e816104f5565b611a3661167a565b90604435611a43816104f5565b60643567ffffffffffffffff811161025857611b906001600160a01b0392611b5d611a9e96611b42611b3d611a7d899736906004016102a8565b60ff85169a91611b3790611a928d1515613d04565b8b89169d8e1515613200565b611afa8785611af4611ae8611ae8611adb85611acb866001600160a01b03165f52600760205260405f2090565b9060ff165f5260205260405f2090565b546001600160a01b031690565b6001600160a01b031690565b15613d33565b6040805160ff891660208201526001600160a01b038b169181019190915246606080830191909152815292611b3060808561292e565b3691612b39565b9061578a565b613d78565b611acb856001600160a01b03165f52600760205260405f2090565b906001600160a01b03167fffffffffffffffffffffffff0000000000000000000000000000000000000000825416179055565b167f2366b94a706a0cfc2dca2fe8be9410b6fba2db75e3e9d3f03b3c2fb0b051efad5f80a4005b611bd7611bc33661025c565b6113d06003610b9d6020849695960161287e565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49283156104f0577f188e0ade7d115cc397426774adb960ae3e8c83e72f0a6cad4b7085e1d60bf986936104ae93610c3b925f926114dc57506114d53685612b8a565b634e487b7160e01b5f52602160045260245ffd5b60041115611c5057565b611c32565b600a1115611c5057565b90600a821015611c505752565b90601f19601f602080948051918291828752018686015e5f8582860101520116010190565b61028c9167ffffffffffffffff8251168152611cb560208301516020830190611c5f565b60408201516040820152611d236060830151606083019060c0809167ffffffffffffffff81511684526001600160a01b03602082015116602085015260ff6040820151166040850152606081015160608501526080810151608085015260a081015160a08501520151910152565b608082810151805167ffffffffffffffff1661014084015260208101516001600160a01b0316610160840152604081015160ff1661018084015260608101516101a0840152908101516101c083015260a08101516101e083015260c0015161020082015260c0611da560a0840151610260610220850152610260840190611c6c565b92015190610240818403910152611c6c565b929367ffffffffffffffff60c09561028c98979482948752611dd881611c46565b602087015216604085015216606083015260808201528160a08201520190611c91565b3461025857602060031936011261025857600435611e17613ddd565b505f52600260205260405f20611e2b612951565b90805482526106ac600182015491611e77611e67611e498560ff1690565b94611e58602088019687613e21565b60081c6001600160a01b031690565b6001600160a01b03166040860152565b611e9e611e8e60028301546001600160a01b031690565b6001600160a01b03166060860152565b60038101546080850152600481015467ffffffffffffffff811660a086019081529490611ed6905b60401c67ffffffffffffffff1690565b67ffffffffffffffff1660c0820190815291611f2a61182e611f06600660058501549460e087019586520161300e565b93610100810194855251965197611f1c89611c46565b5167ffffffffffffffff1690565b905191519260405196879687611db7565b3461025857611f493661086f565b611f5a6008610b9d6020840161287e565b80611fcf611f6b6108f1368661328d565b936020810160e0611f7b826128b7565b91611f9a6040850193611f8d856128b7565b608087013591898c61428c565b610e21610e1a610e0361077c611fb0368b612b8a565b9687958d8a611fbe82613f07565b9d8e15612098575b505050506128b7565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49182156104f05761200c935f93612073575b5061200690369061328d565b8661527e565b15612042576104ae7f3142fb397e715d80415dff7b527bf1c451def4675da6e1199ee1b4588e3f630a916040519182918261286d565b6104ae7f26afbcb9eb52c21f42eb9cfe8f263718ffb65afbf84abe8ad8cce2acfb2242b8916040519182918261286d565b6120069193506120919060e03d60e011610b4257610b34818361292e565b9290611ffa565b6120f1936120cd610a15926120af610a2c95614fdb565b8c60606120c0366101408501612aaa565b9101526060369101612aaa565b60808c01526120da6135ab565b60a08c01526120e76135ab565b60c08c01526128b7565b505f8d8a8e611fc6565b9160a09367ffffffffffffffff9161028c979693855261211a81611c46565b602085015216604083015260608201528160808201520190611c91565b3461025857602060031936011261025857600435612153613ddd565b505f52600560205260405f20612167612961565b90805482526106ac600182015491612185611e67611e498560ff1690565b61219c611e8e60028301546001600160a01b031690565b60038101546080850152600481015467ffffffffffffffff1667ffffffffffffffff1660a08501908152936121f76121e2600660058501549460c085019586520161300e565b9160e0810192835251945195611f1c87611c46565b9151905191604051958695866120fb565b60061115611c5057565b906006821015611c505752565b919261229761012094612239856122aa959a99989a612212565b602085019060a0809163ffffffff81511684526001600160a01b0360208201511660208501526001600160a01b03604082015116604085015267ffffffffffffffff6060820151166060850152608081015160808501520151910152565b61014060e0840152610140830190611c91565b946101008201520152565b34610258576020600319360112610258576004355f60a06040516122d8816128d5565b82815282602082015282604082015282606082015282608082015201526122fd613ddd565b505f525f60205261231060405f20613e39565b805161231b81612208565b6106ac60208301519260408101519060606123446117db608084015167ffffffffffffffff1690565b910151916040519586958661221f565b346102585761236236610506565b90916123786001600160a01b0382161515613200565b61238382151561322f565b335f5260066020526123a98360405f20906001600160a01b03165f5260205260405f2090565b549082821061243e57828203918211610602578383916123fe936123f8836123e2336001600160a01b03165f52600660205260405f2090565b906001600160a01b03165f5260205260405f2090565b55613f64565b7fd1c19fbcd4551a5edfb66d43d2e337c04837afda3482b42bdf569a8fccdae5fb6001600160a01b0360405193169280610c473394829190602083019252565b7ff4d678b8000000000000000000000000000000000000000000000000000000005f5260045ffd5b6124866124723661025c565b6113d06002610b9d6020849695960161287e565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49283156104f0577f6085f5128b19e0d3cc37524413de47259383f0f75265d5d66f41778696206696936104ae93610c3b925f926114dc57506114d53685612b8a565b34610258576124ef36610f78565b6124f7615951565b6001600160a01b0381169161250d831515613200565b6001600160a01b03612534826123e2336001600160a01b03165f52600860205260405f2090565b549161254183151561322f565b5f612561826123e2336001600160a01b03165f52600860205260405f2090565b55169181836125db57612584915f808080858a5af161257e613e97565b50613ec6565b60405190815233907f7b8d70738154be94a9a068a6d2f5dd8cfc65c52855859dc8f47de1ff185f8b5590602090a46102a660017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0055565b50506040517fa9059cbb000000000000000000000000000000000000000000000000000000005f52836004528160245260205f60448180875af160015f5114811615612661575b604091909152612584577f5274afe7000000000000000000000000000000000000000000000000000000005f526001600160a01b03821660045260245ffd5b6001811516612677573d15843b15151616612622565b503d5f823e3d90fd5b34610258575f60031936011261025857602060405160018152f35b906126a581611c46565b60ff60ff198354169116179055565b67ffffffffffffffff81160361025857565b35906126d1826126b4565b565b600a111561025857565b35906126d1826126d3565b60c0809167ffffffffffffffff8135612700816126b4565b1684526001600160a01b036020820135612719816104f5565b16602085015260ff61272d6040830161168a565b166040850152606081013560608501526080810135608085015260a081013560a08501520135910152565b90357fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe18236030181121561025857016020813591019167ffffffffffffffff821161025857813603831361025857565b601f8260209493601f1993818652868601375f8582860101520116010190565b61028c9167ffffffffffffffff82356127e0816126b4565b1681526127fe60208301356127f4816126d3565b6020830190611c5f565b6040820135604082015261281860608201606084016126e8565b61282a610140820161014084016126e8565b61285e61285261283e610220850185612758565b6102606102208601526102608501916127a8565b92610240810190612758565b916102408185039101526127a8565b90602061028c9281815201906127c8565b3561028c816126d3565b1561288f57565b7fc898513c000000000000000000000000000000000000000000000000000000005f5260045ffd5b3561028c816104f5565b634e487b7160e01b5f52604160045260245ffd5b60c0810190811067ffffffffffffffff8211176128f157604052565b6128c1565b60e0810190811067ffffffffffffffff8211176128f157604052565b60a0810190811067ffffffffffffffff8211176128f157604052565b90601f601f19910116810190811067ffffffffffffffff8211176128f157604052565b604051906126d16101208361292e565b604051906126d16101008361292e565b604051906126d160e08361292e565b5190600482101561025857565b51906126d1826126b4565b5190811515820361025857565b908160c091031261025857612a0d60a0604051926129c2846128d5565b80518452602081015160208501526129dc60408201612980565b604085015260608101516129ef816126b4565b60608501526080810151612a02816126b4565b608085015201612998565b60a082015290565b908151612a2181611c46565b815260a080612a3f602085015160c0602086015260c0850190611c91565b936040810151604085015267ffffffffffffffff606082015116606085015267ffffffffffffffff6080820151166080850152015191015290565b9091612a9161028c93604084526040840190612a15565b9160208184039101526127c8565b6040513d5f823e3d90fd5b91908260e091031261025857604051612ac2816128f6565b60c08082948035612ad2816126b4565b84526020810135612ae2816104f5565b6020850152612af36040820161168a565b6040850152606081013560608501526080810135608085015260a081013560a08501520135910152565b67ffffffffffffffff81116128f157601f01601f191660200190565b929192612b4582612b1d565b91612b53604051938461292e565b829481845281830111610258578281602093845f960137010152565b9080601f830112156102585781602061028c93359101612b39565b9190916102608184031261025857612ba0612971565b92612baa826126c6565b8452612bb8602083016126dd565b602085015260408201356040850152612bd48160608401612aaa565b6060850152612be7816101408401612aaa565b608085015261022082013567ffffffffffffffff81116102585781612c0d918401612b6f565b60a085015261024082013567ffffffffffffffff811161025857612c319201612b6f565b60c0830152565b612c4a815f52600260205260405f2090565b60018101805490600882901c6001600160a01b031690612c7460028501546001600160a01b031690565b92612c89612c828654613f07565b9160ff1690565b90159081612e86575b5080612e63575b612ddd5750612caf6005610b9d6020880161287e565b8254612cba81613f07565b15612d05575091612cf2612d00927f32e24720f56fd5a7f4cb219d7ff3278ae95196e79c85b5801395894a6f53466c948784546149af565b54936040519182918261286d565b0390a3565b612d5690612d1e6003869496019185878a85549361428c565b60c087610c60612d4f612d42886001600160a01b03165f52600660205260405f2090565b6107a561016085016128b7565b548961437a565b038173__$682d6198b4eca5bc7e038b912a26498e7e$__5af49384156104f0577f1b92e8ef67d8a7c0d29c99efcd180a5e0d98d60ac41d52abbbb5950882c78e4e95612d0095612cf2945f91612dbe575b508554935493612db7368c612b8a565b908a6146aa565b612dd7915060c03d60c0116104e9576104d8818361292e565b5f612da7565b805460ff19166003179055507f1b92e8ef67d8a7c0d29c99efcd180a5e0d98d60ac41d52abbbb5950882c78e4e91612d0091612cf29060058301905f82549255612e4c600485017fffffffffffffffffffffffffffffffff0000000000000000ffffffffffffffff8154169055565b600d84015460401c6001600160a01b031690613f64565b506004840154612e7f9060401c67ffffffffffffffff166117db565b4211612c99565b60029150612e9381611c46565b145f612c92565b15612ea157565b7ff64df366000000000000000000000000000000000000000000000000000000005f5260045ffd5b90604051612ed6816128f6565b60c060048294612f1460ff825467ffffffffffffffff811687526001600160a01b03808260401c1616602088015260e01c16604086019060ff169052565b6001810154606085015260028101546080850152600381015460a08501520154910152565b90600182811c92168015612f67575b6020831014612f5357565b634e487b7160e01b5f52602260045260245ffd5b91607f1691612f48565b5f9291815491612f8083612f39565b8083529260018116908115612fd55750600114612f9c57505050565b5f9081526020812093945091925b838310612fbb575060209250010190565b600181602092949394548385870101520191019190612faa565b9050602094955060ff1991509291921683830152151560051b010190565b906126d16130079260405193848092612f71565b038361292e565b9060405161301b816128f6565b809260ff815467ffffffffffffffff8116845260401c1690600a821015611c5057600d61308d9160c09360208601526001810154604086015261306060028201612ec9565b606086015261307160078201612ec9565b6080860152613082600c8201612ff3565b60a086015201612ff3565b910152565b90602061028c928181520190612a15565b929160206131ef6126d193604087526130d7815467ffffffffffffffff811660408a015260ff60608a019160401c16611c5f565b60018101546080880152600281015467ffffffffffffffff811660a0890152604081901c6001600160a01b031660c089015260e090811c60ff16908801526003810154610100880152600481015461012088015260058101546101408801526006810154610160880152600781015467ffffffffffffffff8116610180890152604081901c6001600160a01b03166101a089015260e01c60ff166101c088015260088101546101e08801526009810154610200880152600a810154610220880152600b81015461024088015261026080880152600d6131bd6102a08901600c8401612f71565b917fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc0898403016102808a015201612f71565b94019067ffffffffffffffff169052565b1561320757565b7fe6c4247b000000000000000000000000000000000000000000000000000000005f5260045ffd5b1561323657565b7f69640e72000000000000000000000000000000000000000000000000000000005f5260045ffd5b634e487b7160e01b5f52601160045260245ffd5b9190820180921161060257565b63ffffffff81160361025857565b91908260c0910312610258576040516132a5816128d5565b60a080829480356132b58161327f565b845260208101356132c5816104f5565b602085015260408101356132d8816104f5565b604085015260608101356132eb816126b4565b6060850152608081013560808501520135910152565b908160e09103126102585760405190613319826128f6565b8051825260208101516020830152604081015160068110156102585761337a9160c091604085015261334d6060820161298d565b606085015261335e60808201612998565b608085015261336f60a08201612998565b60a085015201612998565b60c082015290565b9061338e818351612212565b608067ffffffffffffffff816133b3602086015160a0602087015260a0860190611c91565b94604081015160408601526060810151606086015201511691015290565b9091612a9161028c93604084526040840190613382565b60e09060a061028c949363ffffffff81356134028161327f565b1683526001600160a01b03602082013561341b816104f5565b1660208401526001600160a01b036040820135613437816104f5565b16604084015267ffffffffffffffff6060820135613454816126b4565b16606084015260808101356080840152013560a08201528160c082015201906127c8565b3561028c816126b4565b634e487b7160e01b5f52603260045260245ffd5b6003548110156134ae5760035f5260205f2001905f90565b613482565b80548210156134ae575f5260205f2001905f90565b916134e1918354905f199060031b92831b921b19161790565b9055565b600354680100000000000000008110156128f157600181016003556003548110156134ae5760035f527fc2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b0155565b90604051613540816128d5565b60a0600382946001600160a01b03815463ffffffff8116865260201c16602085015261359a67ffffffffffffffff60018301546001600160a01b03808216166040880152851c16606086019067ffffffffffffffff169052565b600281015460808501520154910152565b604051906135ba60208361292e565b5f8252565b90916135d661028c93604084526040840190613382565b916020818403910152611c91565b67ffffffffffffffff81116128f15760051b60200190565b6040519061360b60208361292e565b5f808352366020840137565b90613621826135e4565b61362e604051918261292e565b828152601f1961363e82946135e4565b0190602036910137565b9190820391821161060257565b80518210156134ae5760209160051b010190565b9190600354908084029380850482149015171561060257818410156136ed57830190818411610602578082116136e5575b506136ad6136a88483613648565b613617565b92805b8281106136bc57505050565b806136cb610700600193613496565b6136de6136d88584613648565b88613655565b52016136b0565b90505f61369a565b5050905061028c6135fc565b906006811015611c505760ff60ff198354169116179055565b90602061028c928181520190611c91565b906137356001610b9d6020840161287e565b613746825f525f60205260405f2090565b61375260018201613533565b9161375e825460ff1690565b918461376c6005830161300e565b91604086019261378384516001600160a01b031690565b91600261379a60208a01516001600160a01b031690565b976137a481612208565b1480613962575b6138a75750506137d661094e61094261077c61380a9661143160e0978a978c898f608001519361428c565b60405193849283927fa8b4483c000000000000000000000000000000000000000000000000000000008452600484016133d1565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af480156104f057610f6961388194610a1588937f04cd8c68bf83e7bc531ca5a5d75c34e36513c2acf81e07e6470ba79e29da13a898613874965f92613886575b5061386d3689612b8a565b908661527e565b506040519182918261286d565b0390a2565b6138a091925060e03d60e011610b4257610b34818361292e565b905f613862565b7f04cd8c68bf83e7bc531ca5a5d75c34e36513c2acf81e07e6470ba79e29da13a89750613881969195506139559450610f699261390c6014836138f4610a1595600360ff19825416179055565b5f60138201550167ffffffffffffffff198154169055565b606087016139358151606061392b60208301516001600160a01b031690565b9101519086613f64565b5160a061394c60208301516001600160a01b031690565b91015191613f64565b5060405191829182613712565b5061397b6117db601483015467ffffffffffffffff1690565b42116137ab565b5f1981146106025760010190565b908160a0910312610258576139e56080604051926139ad84612912565b80518452602081015160208501526139c760408201612980565b604085015260608101516139da816126b4565b606085015201612998565b608082015290565b9081516139f981611c46565b815260806001600160a01b0381613a1f602086015160a0602087015260a0860190611c91565b946040810151604086015267ffffffffffffffff606082015116606086015201511691015290565b9091612a9161028c936040845260408401906139ed565b613a70815f52600560205260405f2090565b8054600182018054919492600883901c6001600160a01b031691613a9e60028301546001600160a01b031690565b93613aab612c8289613f07565b90159081613c37575b5080613c17575b613bb65750613ad16007610b9d6020870161287e565b613ada86613f07565b15613b11575091610c3b612d00927f6d0cf3d243d63f08f50db493a8af34b27d4e3bc9ec4098e82700abfeffe2d4989483886149af565b6003613b37919492940192613b2a84548287868b61428c565b60a0836115af83896153dc565b038173__$b69fb814c294bfc16f92e50d7aeced4bde$__5af49081156104f0577f2fdac1380dbe23ae259b6871582b7f33e34461547f400bdd20d74991250317d195612d0095610c3b945f94613b95575b50549261164a3687612b8a565b613baf91945060a03d60a01161137f57611370818361292e565b925f613b88565b7f2fdac1380dbe23ae259b6871582b7f33e34461547f400bdd20d74991250317d1949650612cf291925092613bf5612d0094600360ff19825416179055565b60058301905f82549255612e4c6004850167ffffffffffffffff198154169055565b50613c306117db600484015467ffffffffffffffff1690565b4211613abb565b60029150613c4481611c46565b145f613ab4565b90602061028c9281815201906139ed565b15613c6357565b7ff2056b18000000000000000000000000000000000000000000000000000000005f5260045ffd5b15613c9257565b7f7d957361000000000000000000000000000000000000000000000000000000005f5260045ffd5b9067ffffffffffffffff8091169116019067ffffffffffffffff821161060257565b9067ffffffffffffffff613cfd6020929594956040855260408501906127c8565b9416910152565b15613d0b57565b7f06ee4dcd000000000000000000000000000000000000000000000000000000005f5260045ffd5b15613d3c575050565b906001600160a01b0360ff927f0bcc40f3000000000000000000000000000000000000000000000000000000005f52166004521660245260445ffd5b15613d7f57565b7fc1606c2f000000000000000000000000000000000000000000000000000000005f5260045ffd5b60405190613db4826128f6565b5f60c0838281528260208201528260408201528260608201528260808201528260a08201520152565b60405190613dea826128f6565b606060c0835f81525f60208201525f6040820152613e06613da7565b83820152613e12613da7565b60808201528260a08201520152565b613e2a82611c46565b52565b6006821015611c505752565b90604051613e4681612912565b608067ffffffffffffffff60148395613e6360ff82541686613e2d565b613e6f60018201613533565b6020860152613e806005820161300e565b604086015260138101546060860152015416910152565b3d15613ec1573d90613ea882612b1d565b91613eb6604051938461292e565b82523d5f602084013e565b606090565b15613ecf575050565b6001600160a01b03907fa5b05eec000000000000000000000000000000000000000000000000000000005f521660045260245260445ffd5b805f525f60205260ff60405f2054166006811015611c50578015908115613f50575b50613f4b575f525f60205267ffffffffffffffff600760405f20015416461490565b505f90565b60059150613f5d81612208565b145f613f29565b90613f779291613f72615951565b613fac565b60017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0055565b90816020910312610258575190565b9190918115614236576001600160a01b0383169283614050576001600160a01b038216925f8080808488620186a0f1613fe3613e97565b5015613ff0575050505050565b614033612d00926123e27fbf182be802245e8ed88e4b8d3e4344c0863dd2a70334f089fd07265389306fcf956001600160a01b03165f52600860205260405f2090565b61403e828254613272565b90556040519081529081906020820190565b6040517f70a08231000000000000000000000000000000000000000000000000000000008152306004820152602081602481885afa9081156104f0575f91614217575b506040517fa9059cbb00000000000000000000000000000000000000000000000000000000602082019081526001600160a01b0385166024830152604480830187905282525f918291906140e860648261292e565b51908286620186a0f1906140fa613e97565b506040517f70a082310000000000000000000000000000000000000000000000000000000081523060048201526020816024818a5afa9182156104f05786915f936141e6575b50836141db575b836141c7575b5050501561415c575b50505050565b816141a56001600160a01b03926123e27fbf182be802245e8ed88e4b8d3e4344c0863dd2a70334f089fd07265389306fcf956001600160a01b03165f52600860205260405f2090565b6141b0858254613272565b90556040519384521691602090a35f808080614156565b6141d2929350613648565b145f848161414d565b818110159350614147565b61420991935060203d602011614210575b614201818361292e565b810190613f9d565b915f614140565b503d6141f7565b614230915060203d60201161421057614201818361292e565b5f614093565b505050565b9035907fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe181360301821215610258570180359067ffffffffffffffff82116102585760200191813603831361025857565b92916126d1946142bd6142cc926142b283876142ac61022089018961423b565b90614aab565b90878a9493946159c8565b836142ac61024085018561423b565b929091946159c8565b5f604051916142e3836128d5565b818352614376602084016142f5613ddd565b81526143686040860191858352611ec66004606089019288845260808a019589875261433060a08c01998b8b525f52600260205260405f2090565b9160ff60018401541661434281611c46565b8c526143506006840161300e565b90526005820154905201549167ffffffffffffffff83165b67ffffffffffffffff169052565b5290565b9060405191614388836128d5565b5f83526143766020840161439a613ddd565b815261436860408601915f8352611ec6600460608901925f845260808a01955f875261433060a08c01995f8b525f52600260205260405f2090565b9060c060049161440867ffffffffffffffff825116859067ffffffffffffffff1667ffffffffffffffff19825416179055565b602081015184546040808401517fffffff000000000000000000000000000000000000000000ffffffffffffffff90921692901b7bffffffffffffffffffffffffffffffffffffffff0000000000000000169190911760e09190911b7cff0000000000000000000000000000000000000000000000000000000016178455606081015160018501556080810151600285015560a081015160038501550151910155565b601f82116144b857505050565b5f5260205f20906020601f840160051c830193106144f0575b601f0160051c01905b8181106144e5575050565b5f81556001016144da565b90915081906144d1565b919091825167ffffffffffffffff81116128f1576145228161451c8454612f39565b846144ab565b6020601f821160011461455c5781906134e19394955f92614551575b50505f198260011b9260031b1c19161790565b015190505f8061453e565b601f1982169061456f845f5260205f2090565b915f5b8181106145a957509583600195969710614591575b505050811b019055565b01515f1960f88460031b161c191690555f8080614587565b9192602060018192868b015181550194019201614572565b8151815467ffffffffffffffff191667ffffffffffffffff91909116178155602082015191600a831015611c5057600a831015611c505760c0600d916126d1947fffffffffffffffffffffffffffffffffffffffffffffff00ffffffffffffffff68ff000000000000000086549260401b169116178455604081015160018501556146536060820151600286016143d5565b6146646080820151600786016143d5565b61467560a0820151600c86016144fa565b015191016144fa565b7f80000000000000000000000000000000000000000000000000000000000000008114610602575f0390565b60209392916146c761472e929897985f52600260205260405f2090565b966040850180516146d781611c46565b6146e081611c46565b614992575b5060a08501916146f58351151590565b6148fb575b5050606084015167ffffffffffffffff16806148d0575b50608084015167ffffffffffffffff1680614881575b5051151590565b1561486857608001518201516001600160a01b031680935b8251905f8213156148285761476891506147608451615a25565b928391614cef565b61477760058601918254613272565b90555b0180515f8113156147d65750916147b56005926123e261479d6147cb9651615a25565b966001600160a01b03165f52600660205260405f2090565b6147c0858254613648565b905501918254613272565b90555b6126d1614e1b565b9290505f83126147ea575b505050506147ce565b6148076005926123e261479d61480261481d9761467e565b615a25565b614812858254613272565b905501918254613648565b90555f8080806147e1565b5f8212614838575b50505061477a565b61484761480261484f9361467e565b928391613f64565b61485e60058601918254613648565b9055825f80614830565b50600d84015460401c6001600160a01b03168093614746565b6148ca9060048901907fffffffffffffffffffffffffffffffff0000000000000000ffffffffffffffff6fffffffffffffffff000000000000000083549260401b169116179055565b5f614727565b6148f590600489019067ffffffffffffffff1667ffffffffffffffff19825416179055565b5f614711565b6149088560068b016145c1565b88556001880180547fffffffffffffffffffffff0000000000000000000000000000000000000000ff16600889901b74ffffffffffffffffffffffffffffffffffffffff00161790556002880180547fffffffffffffffffffffffff0000000000000000000000000000000000000000166001600160a01b038b1617905560038801555f806146fa565b6149a990516149a081611c46565b60018a0161269b565b5f6146e5565b9091926149dd614a04946149cf6001610d70865f525f60205260405f2090565b92608084015191868661428c565b60e0836109556149fd61094261077c60408701516001600160a01b031690565b54856150e6565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49283156104f0576126d1945f94614a41575b50614a3b903690612b8a565b9161527e565b614a3b919450614a5f9060e03d60e011610b4257610b34818361292e565b9390614a2f565b15614a6f575050565b906001600160a01b0360ff927f577f5940000000000000000000000000000000000000000000000000000000005f52166004521660245260445ffd5b93929190918215614b7e57843560f81c9081614af857507f000000000000000000000000000000000000000000000000000000000000000094600101925f19019150614af49050565b9091565b600180915f97939594975060ff86161c1603614b5657614b4983614b37611adb614af496611acb8a6001600160a01b03165f52600760205260405f2090565b966001600160a01b0388161515614a66565b600101915f199190910190565b7f1a9073b4000000000000000000000000000000000000000000000000000000005f5260045ffd5b7fac241e11000000000000000000000000000000000000000000000000000000005f5260045ffd5b805191908290602001825e015f815290565b60021115611c5057565b9392606093614bed6001600160a01b0394613cfd949998998852608060208901526080880190611c6c565b9186830360408801526127a8565b906001600160a01b039295602097614c2f9599614c66614c1d614cab95615a5a565b614c58604051998a928e840190614ba6565b7f6368616c6c656e67650000000000000000000000000000000000000000000000815260090190565b03601f19810189528861292e565b614c6f81614bb8565b614ce857505b604051988997889687957f600109bb00000000000000000000000000000000000000000000000000000000875260048701614bc2565b0392165afa80156104f0576126d1915f91614cc9575b501515613d78565b614ce2915060203d60201161421057614201818361292e565b5f614cc1565b9050614c75565b90613f779291614cfd615951565b614d31565b15614d0957565b7fd2ade556000000000000000000000000000000000000000000000000000000005f5260045ffd5b908215614236576001600160a01b03169182158015614e0c57614d55823414614d02565b15614d5f57505050565b6001600160a01b03604051927f23b872dd000000000000000000000000000000000000000000000000000000005f52166004523060245260445260205f60648180865af160015f5114811615614df6575b6040919091525f60605215614dc25750565b7f5274afe7000000000000000000000000000000000000000000000000000000005f526001600160a01b031660045260245ffd5b6001811516612677573d15833b15151616614db0565b614e163415614d02565b614d55565b6003546004545f5b82821080614f13575b15614f0857614e4061070a61070084613496565b600181016003614e51825460ff1690565b614e5a81611c46565b14614ef657614e6882615bbd565b15614eb357915f826107d1614eaa956005614ea49601926107c26107ba8554926107a5600d61079361077c60028501546001600160a01b031690565b91613982565b915b9190614e23565b50509150614ec090600455565b80614ec85750565b6040519081527f61815f4b11c6ea4e14a2e448a010bed8efdc3e53a15efbf183d16a31085cd14590602090a1565b505090614f0290613982565b91614eac565b9150614ec090600455565b5060408110614e2c565b7effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff7f010000000000000000000000000000000000000000000000000000000000000091604051614fc460208201809360a0809163ffffffff81511684526001600160a01b0360208201511660208501526001600160a01b03604082015116604085015267ffffffffffffffff6060820151166060850152608081015160808501520151910152565b60c08152614fd360e08261292e565b519020161790565b6020810135614fe9816104f5565b6001600160a01b03811690614fff821515613200565b60408301359061500e826104f5565b6150356001600160a01b03831692615027841515613200565b615030836104f5565b6104f5565b61503e816104f5565b50811461508a575063ffffffff62015180913561505a8161327f565b161061506257565b7f0596b15b000000000000000000000000000000000000000000000000000000005f5260045ffd5b7fabfa558d000000000000000000000000000000000000000000000000000000005f5260045260245ffd5b604051906150c282612912565b5f6080838281526150d1613ddd565b60208201528260408201528260608201520152565b90601467ffffffffffffffff916150fb6150b5565b935f525f60205260405f209061511560ff83541686613e2d565b6151216005830161300e565b6020860152601382015460408601526060850152015416608082015290565b9060a060039163ffffffff8151167fffffffffffffffffffffffffffffffffffffffffffffffffffffffff000000008554161784556001600160a01b036020820151167fffffffffffffffff0000000000000000000000000000000000000000ffffffff77ffffffffffffffffffffffffffffffffffffffff0000000086549260201b16911617845561526d6001850161521b6151e760408501516001600160a01b031690565b82906001600160a01b03167fffffffffffffffffffffffff0000000000000000000000000000000000000000825416179055565b606083015181547fffffffff0000000000000000ffffffffffffffffffffffffffffffffffffffff1660a09190911b7bffffffffffffffff000000000000000000000000000000000000000016179055565b608081015160028501550151910155565b926152ba816152f19460c09461529b885f525f60205260405f2090565b976152a7895460ff1690565b6152b081612208565b1561537f57615581565b6040810180516152c981612208565b6152d281612208565b151580615354575b61533a575b5060a0810151615312575b0151151590565b6152f85750565b5f6013820155601401805467ffffffffffffffff19169055565b835460ff191660011784556153356014850167ffffffffffffffff198154169055565b6152ea565b61534e905161534881612208565b856136f9565b5f6152df565b50845460ff1681519061536682612208565b61536f82612208565b61537881612208565b14156152da565b61538c8260018b01615140565b615581565b9067ffffffffffffffff6040519160208301938452166040820152604081526153bb60608261292e565b51902090565b90614a04916149dd6001610d70835f525f60205260405f2090565b906001600160a01b03906153ee6150b5565b925f52600560205267ffffffffffffffff600460405f2060ff60018201541661541681611c46565b86526154246006820161300e565b602087015260058101546040870152015416606084015216608082015290565b602093929161546161472e929897985f52600560205260405f2090565b9660408501805161547181611c46565b61547a81611c46565b61556d575b50608085019161548f8351151590565b6154d6575b5050606084015167ffffffffffffffff16806154b1575051151590565b6148ca90600489019067ffffffffffffffff1667ffffffffffffffff19825416179055565b6154e38560068b016145c1565b88556001880180547fffffffffffffffffffffff0000000000000000000000000000000000000000ff16600889901b74ffffffffffffffffffffffffffffffffffffffff00161790556002880180547fffffffffffffffffffffffff0000000000000000000000000000000000000000166001600160a01b038b1617905560038801555f80615494565b61557b90516149a081611c46565b5f61547f565b6155976060919493945f525f60205260405f2090565b936155a56080850151151590565b615778575b01916155fd60c06155c66020865101516001600160a01b031690565b9280515f811361573a575b506020810180515f81136156f2575b5081515f81126156b1575b50515f8112615666575b500151151590565b80615658575b615614575b505050506126d1614e1b565b61564d9261564060a092615634604060139601516001600160a01b031690565b90848451015191613f64565b5101519201918254613648565b90555f808080615608565b5060a0835101511515615603565b6148026156729161467e565b61568d856123e261077c60408a01516001600160a01b031690565b615698828254613272565b90556156a960138901918254613648565b90555f6155f5565b6148026156bd9161467e565b6156db81876156d660208b01516001600160a01b031690565b613f64565b6156ea60138a01918254613648565b90555f6155eb565b6156fb90615a25565b615716866123e261077c60408b01516001600160a01b031690565b615721828254613648565b905561573260138a01918254613272565b90555f6155e0565b61574390615a25565b615761818661575c60208a01516001600160a01b031690565b614cef565b61577060138901918254613272565b90555f6155d1565b61578581600587016145c1565b6155aa565b805192835f947a184f03e93ff9f4daa797ed6e38ed64bf6a1f010000000000000000821015615929575b806d04ee2d6d415b85acef8100000000600a92101561590d575b662386f26fc100008110156158f8575b6305f5e1008110156158e6575b6127108110156158d6575b60648110156158c7575b10156158bc575b6158505f19602161581a60018901615dac565b978801015b01917f3031323334353637383961626364656600000000000000000000000000000000600a82061a8353600a900490565b90811561586357615850905f199061581f565b50506001600160a01b036158888461587c858498615d40565b60208151910120615d96565b9116931683146158b4576158a691816020611ae89351910120615d96565b146158af575f90565b600190565b505050600190565b600190940193615807565b60029060649004960195615800565b60049061271090049601956157f6565b6008906305f5e10090049601956157eb565b601090662386f26fc1000090049601956157de565b6020906d04ee2d6d415b85acef810000000090049601956157ce565b50604094507a184f03e93ff9f4daa797ed6e38ed64bf6a1f01000000000000000081046157b4565b60027f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0054146159a05760027f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0055565b7f3ee5aeb5000000000000000000000000000000000000000000000000000000005f5260045ffd5b6001600160a01b0390614cab6159ee6159e960209895999697993690612b8a565b615a5a565b93604051988997889687957f600109bb00000000000000000000000000000000000000000000000000000000875260048701614bc2565b5f8112615a2f5790565b7fa8ce4432000000000000000000000000000000000000000000000000000000005f5260045260245ffd5b67ffffffffffffffff815116906020810151600a811015611c5057615b03826040615b64940151615aa260806060840151930151946040519760208901526040880190611c5f565b6060860152608085019060c0809167ffffffffffffffff81511684526001600160a01b03602082015116602085015260ff6040820151166040850152606081015160608501526080810151608085015260a081015160a08501520151910152565b805167ffffffffffffffff1661016084015260208101516001600160a01b0316610180840152604081015160ff166101a084015260608101516101c084015260808101516101e084015260a081015161020084015260c00151610220830152565b610220815261028c6102408261292e565b90604051918281549182825260208201905f5260205f20925f5b818110615ba45750506126d19250038361292e565b8454835260019485019487945060209093019201615b8f565b67ffffffffffffffff6004820154164210159081615bd9575090565b600180925060ff91015416615bed81611c46565b1490565b6001810190825f528160205260405f2054155f14615c59578054680100000000000000008110156128f157615c46615c308260018794018555846134b3565b81939154905f199060031b92831b921b19161790565b905554915f5260205260405f2055600190565b5050505f90565b80548015615c86575f190190615c7682826134b3565b5f1982549160031b1b1916905555565b634e487b7160e01b5f52603160045260245ffd5b6001810191805f528260205260405f2054928315155f14615d38575f198401848111610602578354935f198501948511610602575f958583615cf597615ce89503615cfb575b505050615c60565b905f5260205260405f2090565b55600190565b615d21615d1b91615d12610700615d2f95886134b3565b928391876134b3565b906134c8565b85905f5260205260405f2090565b555f8080615ce0565b505050505f90565b6126d190615d88615d8294936040519586937f19457468657265756d205369676e6564204d6573736167653a0a0000000000006020860152603a850190614ba6565b90614ba6565b03601f19810184528361292e565b61028c91615da391615dd3565b90929192615e0d565b90615db682612b1d565b615dc3604051918261292e565b828152601f1961363e8294612b1d565b8151919060418303615e0357615dfc9250602082015190606060408401519301515f1a90615ed4565b9192909190565b50505f9160029190565b615e1681611c46565b80615e1f575050565b615e2881611c46565b60018103615e58577ff645eedf000000000000000000000000000000000000000000000000000000005f5260045ffd5b615e6181611c46565b60028103615e9557507ffce698f7000000000000000000000000000000000000000000000000000000005f5260045260245ffd5b80615ea1600392611c46565b14615ea95750565b7fd78bce0c000000000000000000000000000000000000000000000000000000005f5260045260245ffd5b91907f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a08411615f4b579160209360809260ff5f9560405194855216868401526040830152606082015282805260015afa156104f0575f516001600160a01b03811615615f4157905f905f90565b505f906001905f90565b5050505f916003919056fea2646970667358221220d6322a6f79dfa260b6143002f6411140acf47599f685bd2c3432147eed00120464736f6c634300081e0033",
}

// ChannelHubABI is the input ABI used to generate the binding from.
// Deprecated: Use ChannelHubMetaData.ABI instead.
var ChannelHubABI = ChannelHubMetaData.ABI

// ChannelHubBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ChannelHubMetaData.Bin instead.
var ChannelHubBin = ChannelHubMetaData.Bin

// DeployChannelHub deploys a new Ethereum contract, binding an instance of ChannelHub to it.
func DeployChannelHub(auth *bind.TransactOpts, backend bind.ContractBackend, _defaultSigValidator common.Address) (common.Address, *types.Transaction, *ChannelHub, error) {
	parsed, err := ChannelHubMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ChannelHubBin), backend, _defaultSigValidator)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ChannelHub{ChannelHubCaller: ChannelHubCaller{contract: contract}, ChannelHubTransactor: ChannelHubTransactor{contract: contract}, ChannelHubFilterer: ChannelHubFilterer{contract: contract}}, nil
}

// ChannelHub is an auto generated Go binding around an Ethereum contract.
type ChannelHub struct {
	ChannelHubCaller     // Read-only binding to the contract
	ChannelHubTransactor // Write-only binding to the contract
	ChannelHubFilterer   // Log filterer for contract events
}

// ChannelHubCaller is an auto generated read-only Go binding around an Ethereum contract.
type ChannelHubCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ChannelHubTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ChannelHubTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ChannelHubFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ChannelHubFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ChannelHubSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ChannelHubSession struct {
	Contract     *ChannelHub       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ChannelHubCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ChannelHubCallerSession struct {
	Contract *ChannelHubCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// ChannelHubTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ChannelHubTransactorSession struct {
	Contract     *ChannelHubTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// ChannelHubRaw is an auto generated low-level Go binding around an Ethereum contract.
type ChannelHubRaw struct {
	Contract *ChannelHub // Generic contract binding to access the raw methods on
}

// ChannelHubCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ChannelHubCallerRaw struct {
	Contract *ChannelHubCaller // Generic read-only contract binding to access the raw methods on
}

// ChannelHubTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ChannelHubTransactorRaw struct {
	Contract *ChannelHubTransactor // Generic write-only contract binding to access the raw methods on
}

// NewChannelHub creates a new instance of ChannelHub, bound to a specific deployed contract.
func NewChannelHub(address common.Address, backend bind.ContractBackend) (*ChannelHub, error) {
	contract, err := bindChannelHub(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ChannelHub{ChannelHubCaller: ChannelHubCaller{contract: contract}, ChannelHubTransactor: ChannelHubTransactor{contract: contract}, ChannelHubFilterer: ChannelHubFilterer{contract: contract}}, nil
}

// NewChannelHubCaller creates a new read-only instance of ChannelHub, bound to a specific deployed contract.
func NewChannelHubCaller(address common.Address, caller bind.ContractCaller) (*ChannelHubCaller, error) {
	contract, err := bindChannelHub(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ChannelHubCaller{contract: contract}, nil
}

// NewChannelHubTransactor creates a new write-only instance of ChannelHub, bound to a specific deployed contract.
func NewChannelHubTransactor(address common.Address, transactor bind.ContractTransactor) (*ChannelHubTransactor, error) {
	contract, err := bindChannelHub(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ChannelHubTransactor{contract: contract}, nil
}

// NewChannelHubFilterer creates a new log filterer instance of ChannelHub, bound to a specific deployed contract.
func NewChannelHubFilterer(address common.Address, filterer bind.ContractFilterer) (*ChannelHubFilterer, error) {
	contract, err := bindChannelHub(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ChannelHubFilterer{contract: contract}, nil
}

// bindChannelHub binds a generic wrapper to an already deployed contract.
func bindChannelHub(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ChannelHubABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ChannelHub *ChannelHubRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ChannelHub.Contract.ChannelHubCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ChannelHub *ChannelHubRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChannelHubTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ChannelHub *ChannelHubRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChannelHubTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ChannelHub *ChannelHubCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ChannelHub.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ChannelHub *ChannelHubTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ChannelHub.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ChannelHub *ChannelHubTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ChannelHub.Contract.contract.Transact(opts, method, params...)
}

// DEFAULTSIGVALIDATOR is a free data retrieval call binding the contract method 0x71a47141.
//
// Solidity: function DEFAULT_SIG_VALIDATOR() view returns(address)
func (_ChannelHub *ChannelHubCaller) DEFAULTSIGVALIDATOR(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "DEFAULT_SIG_VALIDATOR")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// DEFAULTSIGVALIDATOR is a free data retrieval call binding the contract method 0x71a47141.
//
// Solidity: function DEFAULT_SIG_VALIDATOR() view returns(address)
func (_ChannelHub *ChannelHubSession) DEFAULTSIGVALIDATOR() (common.Address, error) {
	return _ChannelHub.Contract.DEFAULTSIGVALIDATOR(&_ChannelHub.CallOpts)
}

// DEFAULTSIGVALIDATOR is a free data retrieval call binding the contract method 0x71a47141.
//
// Solidity: function DEFAULT_SIG_VALIDATOR() view returns(address)
func (_ChannelHub *ChannelHubCallerSession) DEFAULTSIGVALIDATOR() (common.Address, error) {
	return _ChannelHub.Contract.DEFAULTSIGVALIDATOR(&_ChannelHub.CallOpts)
}

// ESCROWDEPOSITUNLOCKDELAY is a free data retrieval call binding the contract method 0x5a0745b4.
//
// Solidity: function ESCROW_DEPOSIT_UNLOCK_DELAY() view returns(uint32)
func (_ChannelHub *ChannelHubCaller) ESCROWDEPOSITUNLOCKDELAY(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "ESCROW_DEPOSIT_UNLOCK_DELAY")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// ESCROWDEPOSITUNLOCKDELAY is a free data retrieval call binding the contract method 0x5a0745b4.
//
// Solidity: function ESCROW_DEPOSIT_UNLOCK_DELAY() view returns(uint32)
func (_ChannelHub *ChannelHubSession) ESCROWDEPOSITUNLOCKDELAY() (uint32, error) {
	return _ChannelHub.Contract.ESCROWDEPOSITUNLOCKDELAY(&_ChannelHub.CallOpts)
}

// ESCROWDEPOSITUNLOCKDELAY is a free data retrieval call binding the contract method 0x5a0745b4.
//
// Solidity: function ESCROW_DEPOSIT_UNLOCK_DELAY() view returns(uint32)
func (_ChannelHub *ChannelHubCallerSession) ESCROWDEPOSITUNLOCKDELAY() (uint32, error) {
	return _ChannelHub.Contract.ESCROWDEPOSITUNLOCKDELAY(&_ChannelHub.CallOpts)
}

// MAXDEPOSITESCROWPURGE is a free data retrieval call binding the contract method 0x6af820bd.
//
// Solidity: function MAX_DEPOSIT_ESCROW_PURGE() view returns(uint32)
func (_ChannelHub *ChannelHubCaller) MAXDEPOSITESCROWPURGE(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "MAX_DEPOSIT_ESCROW_PURGE")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// MAXDEPOSITESCROWPURGE is a free data retrieval call binding the contract method 0x6af820bd.
//
// Solidity: function MAX_DEPOSIT_ESCROW_PURGE() view returns(uint32)
func (_ChannelHub *ChannelHubSession) MAXDEPOSITESCROWPURGE() (uint32, error) {
	return _ChannelHub.Contract.MAXDEPOSITESCROWPURGE(&_ChannelHub.CallOpts)
}

// MAXDEPOSITESCROWPURGE is a free data retrieval call binding the contract method 0x6af820bd.
//
// Solidity: function MAX_DEPOSIT_ESCROW_PURGE() view returns(uint32)
func (_ChannelHub *ChannelHubCallerSession) MAXDEPOSITESCROWPURGE() (uint32, error) {
	return _ChannelHub.Contract.MAXDEPOSITESCROWPURGE(&_ChannelHub.CallOpts)
}

// MINCHALLENGEDURATION is a free data retrieval call binding the contract method 0x94191051.
//
// Solidity: function MIN_CHALLENGE_DURATION() view returns(uint32)
func (_ChannelHub *ChannelHubCaller) MINCHALLENGEDURATION(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "MIN_CHALLENGE_DURATION")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// MINCHALLENGEDURATION is a free data retrieval call binding the contract method 0x94191051.
//
// Solidity: function MIN_CHALLENGE_DURATION() view returns(uint32)
func (_ChannelHub *ChannelHubSession) MINCHALLENGEDURATION() (uint32, error) {
	return _ChannelHub.Contract.MINCHALLENGEDURATION(&_ChannelHub.CallOpts)
}

// MINCHALLENGEDURATION is a free data retrieval call binding the contract method 0x94191051.
//
// Solidity: function MIN_CHALLENGE_DURATION() view returns(uint32)
func (_ChannelHub *ChannelHubCallerSession) MINCHALLENGEDURATION() (uint32, error) {
	return _ChannelHub.Contract.MINCHALLENGEDURATION(&_ChannelHub.CallOpts)
}

// TRANSFERGASLIMIT is a free data retrieval call binding the contract method 0x38a66be2.
//
// Solidity: function TRANSFER_GAS_LIMIT() view returns(uint256)
func (_ChannelHub *ChannelHubCaller) TRANSFERGASLIMIT(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "TRANSFER_GAS_LIMIT")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TRANSFERGASLIMIT is a free data retrieval call binding the contract method 0x38a66be2.
//
// Solidity: function TRANSFER_GAS_LIMIT() view returns(uint256)
func (_ChannelHub *ChannelHubSession) TRANSFERGASLIMIT() (*big.Int, error) {
	return _ChannelHub.Contract.TRANSFERGASLIMIT(&_ChannelHub.CallOpts)
}

// TRANSFERGASLIMIT is a free data retrieval call binding the contract method 0x38a66be2.
//
// Solidity: function TRANSFER_GAS_LIMIT() view returns(uint256)
func (_ChannelHub *ChannelHubCallerSession) TRANSFERGASLIMIT() (*big.Int, error) {
	return _ChannelHub.Contract.TRANSFERGASLIMIT(&_ChannelHub.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint8)
func (_ChannelHub *ChannelHubCaller) VERSION(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "VERSION")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint8)
func (_ChannelHub *ChannelHubSession) VERSION() (uint8, error) {
	return _ChannelHub.Contract.VERSION(&_ChannelHub.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint8)
func (_ChannelHub *ChannelHubCallerSession) VERSION() (uint8, error) {
	return _ChannelHub.Contract.VERSION(&_ChannelHub.CallOpts)
}

// EscrowHead is a free data retrieval call binding the contract method 0x82d3e15d.
//
// Solidity: function escrowHead() view returns(uint256)
func (_ChannelHub *ChannelHubCaller) EscrowHead(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "escrowHead")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// EscrowHead is a free data retrieval call binding the contract method 0x82d3e15d.
//
// Solidity: function escrowHead() view returns(uint256)
func (_ChannelHub *ChannelHubSession) EscrowHead() (*big.Int, error) {
	return _ChannelHub.Contract.EscrowHead(&_ChannelHub.CallOpts)
}

// EscrowHead is a free data retrieval call binding the contract method 0x82d3e15d.
//
// Solidity: function escrowHead() view returns(uint256)
func (_ChannelHub *ChannelHubCallerSession) EscrowHead() (*big.Int, error) {
	return _ChannelHub.Contract.EscrowHead(&_ChannelHub.CallOpts)
}

// GetAccountBalance is a free data retrieval call binding the contract method 0x587675e8.
//
// Solidity: function getAccountBalance(address node, address token) view returns(uint256)
func (_ChannelHub *ChannelHubCaller) GetAccountBalance(opts *bind.CallOpts, node common.Address, token common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getAccountBalance", node, token)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAccountBalance is a free data retrieval call binding the contract method 0x587675e8.
//
// Solidity: function getAccountBalance(address node, address token) view returns(uint256)
func (_ChannelHub *ChannelHubSession) GetAccountBalance(node common.Address, token common.Address) (*big.Int, error) {
	return _ChannelHub.Contract.GetAccountBalance(&_ChannelHub.CallOpts, node, token)
}

// GetAccountBalance is a free data retrieval call binding the contract method 0x587675e8.
//
// Solidity: function getAccountBalance(address node, address token) view returns(uint256)
func (_ChannelHub *ChannelHubCallerSession) GetAccountBalance(node common.Address, token common.Address) (*big.Int, error) {
	return _ChannelHub.Contract.GetAccountBalance(&_ChannelHub.CallOpts, node, token)
}

// GetChannelData is a free data retrieval call binding the contract method 0xe617208c.
//
// Solidity: function getChannelData(bytes32 channelId) view returns(uint8 status, (uint32,address,address,uint64,uint256,bytes32) definition, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) lastState, uint256 challengeExpiry, uint256 lockedFunds)
func (_ChannelHub *ChannelHubCaller) GetChannelData(opts *bind.CallOpts, channelId [32]byte) (struct {
	Status          uint8
	Definition      ChannelDefinition
	LastState       State
	ChallengeExpiry *big.Int
	LockedFunds     *big.Int
}, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getChannelData", channelId)

	outstruct := new(struct {
		Status          uint8
		Definition      ChannelDefinition
		LastState       State
		ChallengeExpiry *big.Int
		LockedFunds     *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Status = *abi.ConvertType(out[0], new(uint8)).(*uint8)
	outstruct.Definition = *abi.ConvertType(out[1], new(ChannelDefinition)).(*ChannelDefinition)
	outstruct.LastState = *abi.ConvertType(out[2], new(State)).(*State)
	outstruct.ChallengeExpiry = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.LockedFunds = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetChannelData is a free data retrieval call binding the contract method 0xe617208c.
//
// Solidity: function getChannelData(bytes32 channelId) view returns(uint8 status, (uint32,address,address,uint64,uint256,bytes32) definition, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) lastState, uint256 challengeExpiry, uint256 lockedFunds)
func (_ChannelHub *ChannelHubSession) GetChannelData(channelId [32]byte) (struct {
	Status          uint8
	Definition      ChannelDefinition
	LastState       State
	ChallengeExpiry *big.Int
	LockedFunds     *big.Int
}, error) {
	return _ChannelHub.Contract.GetChannelData(&_ChannelHub.CallOpts, channelId)
}

// GetChannelData is a free data retrieval call binding the contract method 0xe617208c.
//
// Solidity: function getChannelData(bytes32 channelId) view returns(uint8 status, (uint32,address,address,uint64,uint256,bytes32) definition, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) lastState, uint256 challengeExpiry, uint256 lockedFunds)
func (_ChannelHub *ChannelHubCallerSession) GetChannelData(channelId [32]byte) (struct {
	Status          uint8
	Definition      ChannelDefinition
	LastState       State
	ChallengeExpiry *big.Int
	LockedFunds     *big.Int
}, error) {
	return _ChannelHub.Contract.GetChannelData(&_ChannelHub.CallOpts, channelId)
}

// GetChannelIds is a free data retrieval call binding the contract method 0x187576d8.
//
// Solidity: function getChannelIds(address user) view returns(bytes32[])
func (_ChannelHub *ChannelHubCaller) GetChannelIds(opts *bind.CallOpts, user common.Address) ([][32]byte, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getChannelIds", user)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetChannelIds is a free data retrieval call binding the contract method 0x187576d8.
//
// Solidity: function getChannelIds(address user) view returns(bytes32[])
func (_ChannelHub *ChannelHubSession) GetChannelIds(user common.Address) ([][32]byte, error) {
	return _ChannelHub.Contract.GetChannelIds(&_ChannelHub.CallOpts, user)
}

// GetChannelIds is a free data retrieval call binding the contract method 0x187576d8.
//
// Solidity: function getChannelIds(address user) view returns(bytes32[])
func (_ChannelHub *ChannelHubCallerSession) GetChannelIds(user common.Address) ([][32]byte, error) {
	return _ChannelHub.Contract.GetChannelIds(&_ChannelHub.CallOpts, user)
}

// GetEscrowDepositData is a free data retrieval call binding the contract method 0xd888ccae.
//
// Solidity: function getEscrowDepositData(bytes32 escrowId) view returns(bytes32 channelId, uint8 status, uint64 unlockAt, uint64 challengeExpiry, uint256 lockedAmount, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState)
func (_ChannelHub *ChannelHubCaller) GetEscrowDepositData(opts *bind.CallOpts, escrowId [32]byte) (struct {
	ChannelId       [32]byte
	Status          uint8
	UnlockAt        uint64
	ChallengeExpiry uint64
	LockedAmount    *big.Int
	InitState       State
}, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getEscrowDepositData", escrowId)

	outstruct := new(struct {
		ChannelId       [32]byte
		Status          uint8
		UnlockAt        uint64
		ChallengeExpiry uint64
		LockedAmount    *big.Int
		InitState       State
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.ChannelId = *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	outstruct.Status = *abi.ConvertType(out[1], new(uint8)).(*uint8)
	outstruct.UnlockAt = *abi.ConvertType(out[2], new(uint64)).(*uint64)
	outstruct.ChallengeExpiry = *abi.ConvertType(out[3], new(uint64)).(*uint64)
	outstruct.LockedAmount = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.InitState = *abi.ConvertType(out[5], new(State)).(*State)

	return *outstruct, err

}

// GetEscrowDepositData is a free data retrieval call binding the contract method 0xd888ccae.
//
// Solidity: function getEscrowDepositData(bytes32 escrowId) view returns(bytes32 channelId, uint8 status, uint64 unlockAt, uint64 challengeExpiry, uint256 lockedAmount, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState)
func (_ChannelHub *ChannelHubSession) GetEscrowDepositData(escrowId [32]byte) (struct {
	ChannelId       [32]byte
	Status          uint8
	UnlockAt        uint64
	ChallengeExpiry uint64
	LockedAmount    *big.Int
	InitState       State
}, error) {
	return _ChannelHub.Contract.GetEscrowDepositData(&_ChannelHub.CallOpts, escrowId)
}

// GetEscrowDepositData is a free data retrieval call binding the contract method 0xd888ccae.
//
// Solidity: function getEscrowDepositData(bytes32 escrowId) view returns(bytes32 channelId, uint8 status, uint64 unlockAt, uint64 challengeExpiry, uint256 lockedAmount, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState)
func (_ChannelHub *ChannelHubCallerSession) GetEscrowDepositData(escrowId [32]byte) (struct {
	ChannelId       [32]byte
	Status          uint8
	UnlockAt        uint64
	ChallengeExpiry uint64
	LockedAmount    *big.Int
	InitState       State
}, error) {
	return _ChannelHub.Contract.GetEscrowDepositData(&_ChannelHub.CallOpts, escrowId)
}

// GetEscrowDepositIds is a free data retrieval call binding the contract method 0x5b9acbf9.
//
// Solidity: function getEscrowDepositIds(uint256 page, uint256 pageSize) view returns(bytes32[] ids)
func (_ChannelHub *ChannelHubCaller) GetEscrowDepositIds(opts *bind.CallOpts, page *big.Int, pageSize *big.Int) ([][32]byte, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getEscrowDepositIds", page, pageSize)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetEscrowDepositIds is a free data retrieval call binding the contract method 0x5b9acbf9.
//
// Solidity: function getEscrowDepositIds(uint256 page, uint256 pageSize) view returns(bytes32[] ids)
func (_ChannelHub *ChannelHubSession) GetEscrowDepositIds(page *big.Int, pageSize *big.Int) ([][32]byte, error) {
	return _ChannelHub.Contract.GetEscrowDepositIds(&_ChannelHub.CallOpts, page, pageSize)
}

// GetEscrowDepositIds is a free data retrieval call binding the contract method 0x5b9acbf9.
//
// Solidity: function getEscrowDepositIds(uint256 page, uint256 pageSize) view returns(bytes32[] ids)
func (_ChannelHub *ChannelHubCallerSession) GetEscrowDepositIds(page *big.Int, pageSize *big.Int) ([][32]byte, error) {
	return _ChannelHub.Contract.GetEscrowDepositIds(&_ChannelHub.CallOpts, page, pageSize)
}

// GetEscrowWithdrawalData is a free data retrieval call binding the contract method 0xdd73d494.
//
// Solidity: function getEscrowWithdrawalData(bytes32 escrowId) view returns(bytes32 channelId, uint8 status, uint64 challengeExpiry, uint256 lockedAmount, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState)
func (_ChannelHub *ChannelHubCaller) GetEscrowWithdrawalData(opts *bind.CallOpts, escrowId [32]byte) (struct {
	ChannelId       [32]byte
	Status          uint8
	ChallengeExpiry uint64
	LockedAmount    *big.Int
	InitState       State
}, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getEscrowWithdrawalData", escrowId)

	outstruct := new(struct {
		ChannelId       [32]byte
		Status          uint8
		ChallengeExpiry uint64
		LockedAmount    *big.Int
		InitState       State
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.ChannelId = *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	outstruct.Status = *abi.ConvertType(out[1], new(uint8)).(*uint8)
	outstruct.ChallengeExpiry = *abi.ConvertType(out[2], new(uint64)).(*uint64)
	outstruct.LockedAmount = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.InitState = *abi.ConvertType(out[4], new(State)).(*State)

	return *outstruct, err

}

// GetEscrowWithdrawalData is a free data retrieval call binding the contract method 0xdd73d494.
//
// Solidity: function getEscrowWithdrawalData(bytes32 escrowId) view returns(bytes32 channelId, uint8 status, uint64 challengeExpiry, uint256 lockedAmount, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState)
func (_ChannelHub *ChannelHubSession) GetEscrowWithdrawalData(escrowId [32]byte) (struct {
	ChannelId       [32]byte
	Status          uint8
	ChallengeExpiry uint64
	LockedAmount    *big.Int
	InitState       State
}, error) {
	return _ChannelHub.Contract.GetEscrowWithdrawalData(&_ChannelHub.CallOpts, escrowId)
}

// GetEscrowWithdrawalData is a free data retrieval call binding the contract method 0xdd73d494.
//
// Solidity: function getEscrowWithdrawalData(bytes32 escrowId) view returns(bytes32 channelId, uint8 status, uint64 challengeExpiry, uint256 lockedAmount, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState)
func (_ChannelHub *ChannelHubCallerSession) GetEscrowWithdrawalData(escrowId [32]byte) (struct {
	ChannelId       [32]byte
	Status          uint8
	ChallengeExpiry uint64
	LockedAmount    *big.Int
	InitState       State
}, error) {
	return _ChannelHub.Contract.GetEscrowWithdrawalData(&_ChannelHub.CallOpts, escrowId)
}

// GetNodeValidator is a free data retrieval call binding the contract method 0xb00b6fd6.
//
// Solidity: function getNodeValidator(address node, uint8 validatorId) view returns(address)
func (_ChannelHub *ChannelHubCaller) GetNodeValidator(opts *bind.CallOpts, node common.Address, validatorId uint8) (common.Address, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getNodeValidator", node, validatorId)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetNodeValidator is a free data retrieval call binding the contract method 0xb00b6fd6.
//
// Solidity: function getNodeValidator(address node, uint8 validatorId) view returns(address)
func (_ChannelHub *ChannelHubSession) GetNodeValidator(node common.Address, validatorId uint8) (common.Address, error) {
	return _ChannelHub.Contract.GetNodeValidator(&_ChannelHub.CallOpts, node, validatorId)
}

// GetNodeValidator is a free data retrieval call binding the contract method 0xb00b6fd6.
//
// Solidity: function getNodeValidator(address node, uint8 validatorId) view returns(address)
func (_ChannelHub *ChannelHubCallerSession) GetNodeValidator(node common.Address, validatorId uint8) (common.Address, error) {
	return _ChannelHub.Contract.GetNodeValidator(&_ChannelHub.CallOpts, node, validatorId)
}

// GetOpenChannels is a free data retrieval call binding the contract method 0x6898234b.
//
// Solidity: function getOpenChannels(address user) view returns(bytes32[] openChannels)
func (_ChannelHub *ChannelHubCaller) GetOpenChannels(opts *bind.CallOpts, user common.Address) ([][32]byte, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getOpenChannels", user)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetOpenChannels is a free data retrieval call binding the contract method 0x6898234b.
//
// Solidity: function getOpenChannels(address user) view returns(bytes32[] openChannels)
func (_ChannelHub *ChannelHubSession) GetOpenChannels(user common.Address) ([][32]byte, error) {
	return _ChannelHub.Contract.GetOpenChannels(&_ChannelHub.CallOpts, user)
}

// GetOpenChannels is a free data retrieval call binding the contract method 0x6898234b.
//
// Solidity: function getOpenChannels(address user) view returns(bytes32[] openChannels)
func (_ChannelHub *ChannelHubCallerSession) GetOpenChannels(user common.Address) ([][32]byte, error) {
	return _ChannelHub.Contract.GetOpenChannels(&_ChannelHub.CallOpts, user)
}

// GetReclaimBalance is a free data retrieval call binding the contract method 0x735181f0.
//
// Solidity: function getReclaimBalance(address account, address token) view returns(uint256)
func (_ChannelHub *ChannelHubCaller) GetReclaimBalance(opts *bind.CallOpts, account common.Address, token common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getReclaimBalance", account, token)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetReclaimBalance is a free data retrieval call binding the contract method 0x735181f0.
//
// Solidity: function getReclaimBalance(address account, address token) view returns(uint256)
func (_ChannelHub *ChannelHubSession) GetReclaimBalance(account common.Address, token common.Address) (*big.Int, error) {
	return _ChannelHub.Contract.GetReclaimBalance(&_ChannelHub.CallOpts, account, token)
}

// GetReclaimBalance is a free data retrieval call binding the contract method 0x735181f0.
//
// Solidity: function getReclaimBalance(address account, address token) view returns(uint256)
func (_ChannelHub *ChannelHubCallerSession) GetReclaimBalance(account common.Address, token common.Address) (*big.Int, error) {
	return _ChannelHub.Contract.GetReclaimBalance(&_ChannelHub.CallOpts, account, token)
}

// ChallengeChannel is a paid mutator transaction binding the contract method 0xb25a1d38.
//
// Solidity: function challengeChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate, bytes challengerSig, uint8 challengerIdx) payable returns()
func (_ChannelHub *ChannelHubTransactor) ChallengeChannel(opts *bind.TransactOpts, channelId [32]byte, candidate State, challengerSig []byte, challengerIdx uint8) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "challengeChannel", channelId, candidate, challengerSig, challengerIdx)
}

// ChallengeChannel is a paid mutator transaction binding the contract method 0xb25a1d38.
//
// Solidity: function challengeChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate, bytes challengerSig, uint8 challengerIdx) payable returns()
func (_ChannelHub *ChannelHubSession) ChallengeChannel(channelId [32]byte, candidate State, challengerSig []byte, challengerIdx uint8) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChallengeChannel(&_ChannelHub.TransactOpts, channelId, candidate, challengerSig, challengerIdx)
}

// ChallengeChannel is a paid mutator transaction binding the contract method 0xb25a1d38.
//
// Solidity: function challengeChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate, bytes challengerSig, uint8 challengerIdx) payable returns()
func (_ChannelHub *ChannelHubTransactorSession) ChallengeChannel(channelId [32]byte, candidate State, challengerSig []byte, challengerIdx uint8) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChallengeChannel(&_ChannelHub.TransactOpts, channelId, candidate, challengerSig, challengerIdx)
}

// ChallengeEscrowDeposit is a paid mutator transaction binding the contract method 0x16b390b1.
//
// Solidity: function challengeEscrowDeposit(bytes32 escrowId, bytes challengerSig, uint8 challengerIdx) returns()
func (_ChannelHub *ChannelHubTransactor) ChallengeEscrowDeposit(opts *bind.TransactOpts, escrowId [32]byte, challengerSig []byte, challengerIdx uint8) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "challengeEscrowDeposit", escrowId, challengerSig, challengerIdx)
}

// ChallengeEscrowDeposit is a paid mutator transaction binding the contract method 0x16b390b1.
//
// Solidity: function challengeEscrowDeposit(bytes32 escrowId, bytes challengerSig, uint8 challengerIdx) returns()
func (_ChannelHub *ChannelHubSession) ChallengeEscrowDeposit(escrowId [32]byte, challengerSig []byte, challengerIdx uint8) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChallengeEscrowDeposit(&_ChannelHub.TransactOpts, escrowId, challengerSig, challengerIdx)
}

// ChallengeEscrowDeposit is a paid mutator transaction binding the contract method 0x16b390b1.
//
// Solidity: function challengeEscrowDeposit(bytes32 escrowId, bytes challengerSig, uint8 challengerIdx) returns()
func (_ChannelHub *ChannelHubTransactorSession) ChallengeEscrowDeposit(escrowId [32]byte, challengerSig []byte, challengerIdx uint8) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChallengeEscrowDeposit(&_ChannelHub.TransactOpts, escrowId, challengerSig, challengerIdx)
}

// ChallengeEscrowWithdrawal is a paid mutator transaction binding the contract method 0x8d0b12a5.
//
// Solidity: function challengeEscrowWithdrawal(bytes32 escrowId, bytes challengerSig, uint8 challengerIdx) returns()
func (_ChannelHub *ChannelHubTransactor) ChallengeEscrowWithdrawal(opts *bind.TransactOpts, escrowId [32]byte, challengerSig []byte, challengerIdx uint8) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "challengeEscrowWithdrawal", escrowId, challengerSig, challengerIdx)
}

// ChallengeEscrowWithdrawal is a paid mutator transaction binding the contract method 0x8d0b12a5.
//
// Solidity: function challengeEscrowWithdrawal(bytes32 escrowId, bytes challengerSig, uint8 challengerIdx) returns()
func (_ChannelHub *ChannelHubSession) ChallengeEscrowWithdrawal(escrowId [32]byte, challengerSig []byte, challengerIdx uint8) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChallengeEscrowWithdrawal(&_ChannelHub.TransactOpts, escrowId, challengerSig, challengerIdx)
}

// ChallengeEscrowWithdrawal is a paid mutator transaction binding the contract method 0x8d0b12a5.
//
// Solidity: function challengeEscrowWithdrawal(bytes32 escrowId, bytes challengerSig, uint8 challengerIdx) returns()
func (_ChannelHub *ChannelHubTransactorSession) ChallengeEscrowWithdrawal(escrowId [32]byte, challengerSig []byte, challengerIdx uint8) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChallengeEscrowWithdrawal(&_ChannelHub.TransactOpts, escrowId, challengerSig, challengerIdx)
}

// CheckpointChannel is a paid mutator transaction binding the contract method 0x9691b468.
//
// Solidity: function checkpointChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubTransactor) CheckpointChannel(opts *bind.TransactOpts, channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "checkpointChannel", channelId, candidate)
}

// CheckpointChannel is a paid mutator transaction binding the contract method 0x9691b468.
//
// Solidity: function checkpointChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubSession) CheckpointChannel(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.CheckpointChannel(&_ChannelHub.TransactOpts, channelId, candidate)
}

// CheckpointChannel is a paid mutator transaction binding the contract method 0x9691b468.
//
// Solidity: function checkpointChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubTransactorSession) CheckpointChannel(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.CheckpointChannel(&_ChannelHub.TransactOpts, channelId, candidate)
}

// ClaimFunds is a paid mutator transaction binding the contract method 0xf766f8d6.
//
// Solidity: function claimFunds(address token, address destination) returns()
func (_ChannelHub *ChannelHubTransactor) ClaimFunds(opts *bind.TransactOpts, token common.Address, destination common.Address) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "claimFunds", token, destination)
}

// ClaimFunds is a paid mutator transaction binding the contract method 0xf766f8d6.
//
// Solidity: function claimFunds(address token, address destination) returns()
func (_ChannelHub *ChannelHubSession) ClaimFunds(token common.Address, destination common.Address) (*types.Transaction, error) {
	return _ChannelHub.Contract.ClaimFunds(&_ChannelHub.TransactOpts, token, destination)
}

// ClaimFunds is a paid mutator transaction binding the contract method 0xf766f8d6.
//
// Solidity: function claimFunds(address token, address destination) returns()
func (_ChannelHub *ChannelHubTransactorSession) ClaimFunds(token common.Address, destination common.Address) (*types.Transaction, error) {
	return _ChannelHub.Contract.ClaimFunds(&_ChannelHub.TransactOpts, token, destination)
}

// CloseChannel is a paid mutator transaction binding the contract method 0x5dc46a74.
//
// Solidity: function closeChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubTransactor) CloseChannel(opts *bind.TransactOpts, channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "closeChannel", channelId, candidate)
}

// CloseChannel is a paid mutator transaction binding the contract method 0x5dc46a74.
//
// Solidity: function closeChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubSession) CloseChannel(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.CloseChannel(&_ChannelHub.TransactOpts, channelId, candidate)
}

// CloseChannel is a paid mutator transaction binding the contract method 0x5dc46a74.
//
// Solidity: function closeChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubTransactorSession) CloseChannel(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.CloseChannel(&_ChannelHub.TransactOpts, channelId, candidate)
}

// CreateChannel is a paid mutator transaction binding the contract method 0x41b660ef.
//
// Solidity: function createChannel((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState) payable returns()
func (_ChannelHub *ChannelHubTransactor) CreateChannel(opts *bind.TransactOpts, def ChannelDefinition, initState State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "createChannel", def, initState)
}

// CreateChannel is a paid mutator transaction binding the contract method 0x41b660ef.
//
// Solidity: function createChannel((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState) payable returns()
func (_ChannelHub *ChannelHubSession) CreateChannel(def ChannelDefinition, initState State) (*types.Transaction, error) {
	return _ChannelHub.Contract.CreateChannel(&_ChannelHub.TransactOpts, def, initState)
}

// CreateChannel is a paid mutator transaction binding the contract method 0x41b660ef.
//
// Solidity: function createChannel((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState) payable returns()
func (_ChannelHub *ChannelHubTransactorSession) CreateChannel(def ChannelDefinition, initState State) (*types.Transaction, error) {
	return _ChannelHub.Contract.CreateChannel(&_ChannelHub.TransactOpts, def, initState)
}

// DepositToChannel is a paid mutator transaction binding the contract method 0xf4ac51f5.
//
// Solidity: function depositToChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubTransactor) DepositToChannel(opts *bind.TransactOpts, channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "depositToChannel", channelId, candidate)
}

// DepositToChannel is a paid mutator transaction binding the contract method 0xf4ac51f5.
//
// Solidity: function depositToChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubSession) DepositToChannel(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.DepositToChannel(&_ChannelHub.TransactOpts, channelId, candidate)
}

// DepositToChannel is a paid mutator transaction binding the contract method 0xf4ac51f5.
//
// Solidity: function depositToChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubTransactorSession) DepositToChannel(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.DepositToChannel(&_ChannelHub.TransactOpts, channelId, candidate)
}

// DepositToVault is a paid mutator transaction binding the contract method 0x17536c06.
//
// Solidity: function depositToVault(address node, address token, uint256 amount) payable returns()
func (_ChannelHub *ChannelHubTransactor) DepositToVault(opts *bind.TransactOpts, node common.Address, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "depositToVault", node, token, amount)
}

// DepositToVault is a paid mutator transaction binding the contract method 0x17536c06.
//
// Solidity: function depositToVault(address node, address token, uint256 amount) payable returns()
func (_ChannelHub *ChannelHubSession) DepositToVault(node common.Address, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ChannelHub.Contract.DepositToVault(&_ChannelHub.TransactOpts, node, token, amount)
}

// DepositToVault is a paid mutator transaction binding the contract method 0x17536c06.
//
// Solidity: function depositToVault(address node, address token, uint256 amount) payable returns()
func (_ChannelHub *ChannelHubTransactorSession) DepositToVault(node common.Address, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ChannelHub.Contract.DepositToVault(&_ChannelHub.TransactOpts, node, token, amount)
}

// FinalizeEscrowDeposit is a paid mutator transaction binding the contract method 0x13c380ed.
//
// Solidity: function finalizeEscrowDeposit(bytes32 escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactor) FinalizeEscrowDeposit(opts *bind.TransactOpts, escrowId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "finalizeEscrowDeposit", escrowId, candidate)
}

// FinalizeEscrowDeposit is a paid mutator transaction binding the contract method 0x13c380ed.
//
// Solidity: function finalizeEscrowDeposit(bytes32 escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubSession) FinalizeEscrowDeposit(escrowId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.FinalizeEscrowDeposit(&_ChannelHub.TransactOpts, escrowId, candidate)
}

// FinalizeEscrowDeposit is a paid mutator transaction binding the contract method 0x13c380ed.
//
// Solidity: function finalizeEscrowDeposit(bytes32 escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactorSession) FinalizeEscrowDeposit(escrowId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.FinalizeEscrowDeposit(&_ChannelHub.TransactOpts, escrowId, candidate)
}

// FinalizeEscrowWithdrawal is a paid mutator transaction binding the contract method 0x7e7985f9.
//
// Solidity: function finalizeEscrowWithdrawal(bytes32 escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactor) FinalizeEscrowWithdrawal(opts *bind.TransactOpts, escrowId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "finalizeEscrowWithdrawal", escrowId, candidate)
}

// FinalizeEscrowWithdrawal is a paid mutator transaction binding the contract method 0x7e7985f9.
//
// Solidity: function finalizeEscrowWithdrawal(bytes32 escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubSession) FinalizeEscrowWithdrawal(escrowId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.FinalizeEscrowWithdrawal(&_ChannelHub.TransactOpts, escrowId, candidate)
}

// FinalizeEscrowWithdrawal is a paid mutator transaction binding the contract method 0x7e7985f9.
//
// Solidity: function finalizeEscrowWithdrawal(bytes32 escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactorSession) FinalizeEscrowWithdrawal(escrowId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.FinalizeEscrowWithdrawal(&_ChannelHub.TransactOpts, escrowId, candidate)
}

// FinalizeMigration is a paid mutator transaction binding the contract method 0x53269198.
//
// Solidity: function finalizeMigration(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactor) FinalizeMigration(opts *bind.TransactOpts, channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "finalizeMigration", channelId, candidate)
}

// FinalizeMigration is a paid mutator transaction binding the contract method 0x53269198.
//
// Solidity: function finalizeMigration(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubSession) FinalizeMigration(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.FinalizeMigration(&_ChannelHub.TransactOpts, channelId, candidate)
}

// FinalizeMigration is a paid mutator transaction binding the contract method 0x53269198.
//
// Solidity: function finalizeMigration(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactorSession) FinalizeMigration(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.FinalizeMigration(&_ChannelHub.TransactOpts, channelId, candidate)
}

// InitiateEscrowDeposit is a paid mutator transaction binding the contract method 0x47de477a.
//
// Solidity: function initiateEscrowDeposit((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubTransactor) InitiateEscrowDeposit(opts *bind.TransactOpts, def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "initiateEscrowDeposit", def, candidate)
}

// InitiateEscrowDeposit is a paid mutator transaction binding the contract method 0x47de477a.
//
// Solidity: function initiateEscrowDeposit((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubSession) InitiateEscrowDeposit(def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.InitiateEscrowDeposit(&_ChannelHub.TransactOpts, def, candidate)
}

// InitiateEscrowDeposit is a paid mutator transaction binding the contract method 0x47de477a.
//
// Solidity: function initiateEscrowDeposit((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubTransactorSession) InitiateEscrowDeposit(def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.InitiateEscrowDeposit(&_ChannelHub.TransactOpts, def, candidate)
}

// InitiateEscrowWithdrawal is a paid mutator transaction binding the contract method 0xa5c82680.
//
// Solidity: function initiateEscrowWithdrawal((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactor) InitiateEscrowWithdrawal(opts *bind.TransactOpts, def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "initiateEscrowWithdrawal", def, candidate)
}

// InitiateEscrowWithdrawal is a paid mutator transaction binding the contract method 0xa5c82680.
//
// Solidity: function initiateEscrowWithdrawal((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubSession) InitiateEscrowWithdrawal(def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.InitiateEscrowWithdrawal(&_ChannelHub.TransactOpts, def, candidate)
}

// InitiateEscrowWithdrawal is a paid mutator transaction binding the contract method 0xa5c82680.
//
// Solidity: function initiateEscrowWithdrawal((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactorSession) InitiateEscrowWithdrawal(def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.InitiateEscrowWithdrawal(&_ChannelHub.TransactOpts, def, candidate)
}

// InitiateMigration is a paid mutator transaction binding the contract method 0xdc23f29e.
//
// Solidity: function initiateMigration((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactor) InitiateMigration(opts *bind.TransactOpts, def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "initiateMigration", def, candidate)
}

// InitiateMigration is a paid mutator transaction binding the contract method 0xdc23f29e.
//
// Solidity: function initiateMigration((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubSession) InitiateMigration(def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.InitiateMigration(&_ChannelHub.TransactOpts, def, candidate)
}

// InitiateMigration is a paid mutator transaction binding the contract method 0xdc23f29e.
//
// Solidity: function initiateMigration((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactorSession) InitiateMigration(def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.InitiateMigration(&_ChannelHub.TransactOpts, def, candidate)
}

// PurgeEscrowDeposits is a paid mutator transaction binding the contract method 0x3115f630.
//
// Solidity: function purgeEscrowDeposits(uint256 maxToPurge) returns()
func (_ChannelHub *ChannelHubTransactor) PurgeEscrowDeposits(opts *bind.TransactOpts, maxToPurge *big.Int) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "purgeEscrowDeposits", maxToPurge)
}

// PurgeEscrowDeposits is a paid mutator transaction binding the contract method 0x3115f630.
//
// Solidity: function purgeEscrowDeposits(uint256 maxToPurge) returns()
func (_ChannelHub *ChannelHubSession) PurgeEscrowDeposits(maxToPurge *big.Int) (*types.Transaction, error) {
	return _ChannelHub.Contract.PurgeEscrowDeposits(&_ChannelHub.TransactOpts, maxToPurge)
}

// PurgeEscrowDeposits is a paid mutator transaction binding the contract method 0x3115f630.
//
// Solidity: function purgeEscrowDeposits(uint256 maxToPurge) returns()
func (_ChannelHub *ChannelHubTransactorSession) PurgeEscrowDeposits(maxToPurge *big.Int) (*types.Transaction, error) {
	return _ChannelHub.Contract.PurgeEscrowDeposits(&_ChannelHub.TransactOpts, maxToPurge)
}

// RegisterNodeValidator is a paid mutator transaction binding the contract method 0xbeed9d5f.
//
// Solidity: function registerNodeValidator(address node, uint8 validatorId, address validator, bytes signature) returns()
func (_ChannelHub *ChannelHubTransactor) RegisterNodeValidator(opts *bind.TransactOpts, node common.Address, validatorId uint8, validator common.Address, signature []byte) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "registerNodeValidator", node, validatorId, validator, signature)
}

// RegisterNodeValidator is a paid mutator transaction binding the contract method 0xbeed9d5f.
//
// Solidity: function registerNodeValidator(address node, uint8 validatorId, address validator, bytes signature) returns()
func (_ChannelHub *ChannelHubSession) RegisterNodeValidator(node common.Address, validatorId uint8, validator common.Address, signature []byte) (*types.Transaction, error) {
	return _ChannelHub.Contract.RegisterNodeValidator(&_ChannelHub.TransactOpts, node, validatorId, validator, signature)
}

// RegisterNodeValidator is a paid mutator transaction binding the contract method 0xbeed9d5f.
//
// Solidity: function registerNodeValidator(address node, uint8 validatorId, address validator, bytes signature) returns()
func (_ChannelHub *ChannelHubTransactorSession) RegisterNodeValidator(node common.Address, validatorId uint8, validator common.Address, signature []byte) (*types.Transaction, error) {
	return _ChannelHub.Contract.RegisterNodeValidator(&_ChannelHub.TransactOpts, node, validatorId, validator, signature)
}

// WithdrawFromChannel is a paid mutator transaction binding the contract method 0xc74a2d10.
//
// Solidity: function withdrawFromChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubTransactor) WithdrawFromChannel(opts *bind.TransactOpts, channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "withdrawFromChannel", channelId, candidate)
}

// WithdrawFromChannel is a paid mutator transaction binding the contract method 0xc74a2d10.
//
// Solidity: function withdrawFromChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubSession) WithdrawFromChannel(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.WithdrawFromChannel(&_ChannelHub.TransactOpts, channelId, candidate)
}

// WithdrawFromChannel is a paid mutator transaction binding the contract method 0xc74a2d10.
//
// Solidity: function withdrawFromChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubTransactorSession) WithdrawFromChannel(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.WithdrawFromChannel(&_ChannelHub.TransactOpts, channelId, candidate)
}

// WithdrawFromVault is a paid mutator transaction binding the contract method 0xecf3d7e8.
//
// Solidity: function withdrawFromVault(address to, address token, uint256 amount) returns()
func (_ChannelHub *ChannelHubTransactor) WithdrawFromVault(opts *bind.TransactOpts, to common.Address, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "withdrawFromVault", to, token, amount)
}

// WithdrawFromVault is a paid mutator transaction binding the contract method 0xecf3d7e8.
//
// Solidity: function withdrawFromVault(address to, address token, uint256 amount) returns()
func (_ChannelHub *ChannelHubSession) WithdrawFromVault(to common.Address, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ChannelHub.Contract.WithdrawFromVault(&_ChannelHub.TransactOpts, to, token, amount)
}

// WithdrawFromVault is a paid mutator transaction binding the contract method 0xecf3d7e8.
//
// Solidity: function withdrawFromVault(address to, address token, uint256 amount) returns()
func (_ChannelHub *ChannelHubTransactorSession) WithdrawFromVault(to common.Address, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ChannelHub.Contract.WithdrawFromVault(&_ChannelHub.TransactOpts, to, token, amount)
}

// ChannelHubChannelChallengedIterator is returned from FilterChannelChallenged and is used to iterate over the raw logs and unpacked data for ChannelChallenged events raised by the ChannelHub contract.
type ChannelHubChannelChallengedIterator struct {
	Event *ChannelHubChannelChallenged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubChannelChallengedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubChannelChallenged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubChannelChallenged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubChannelChallengedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubChannelChallengedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubChannelChallenged represents a ChannelChallenged event raised by the ChannelHub contract.
type ChannelHubChannelChallenged struct {
	ChannelId         [32]byte
	Candidate         State
	ChallengeExpireAt uint64
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterChannelChallenged is a free log retrieval operation binding the contract event 0x07b9206d5a6026d3bd2a8f9a9b79f6fa4bfbd6a016975829fbaf07488019f28a.
//
// Solidity: event ChannelChallenged(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate, uint64 challengeExpireAt)
func (_ChannelHub *ChannelHubFilterer) FilterChannelChallenged(opts *bind.FilterOpts, channelId [][32]byte) (*ChannelHubChannelChallengedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "ChannelChallenged", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubChannelChallengedIterator{contract: _ChannelHub.contract, event: "ChannelChallenged", logs: logs, sub: sub}, nil
}

// WatchChannelChallenged is a free log subscription operation binding the contract event 0x07b9206d5a6026d3bd2a8f9a9b79f6fa4bfbd6a016975829fbaf07488019f28a.
//
// Solidity: event ChannelChallenged(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate, uint64 challengeExpireAt)
func (_ChannelHub *ChannelHubFilterer) WatchChannelChallenged(opts *bind.WatchOpts, sink chan<- *ChannelHubChannelChallenged, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "ChannelChallenged", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubChannelChallenged)
				if err := _ChannelHub.contract.UnpackLog(event, "ChannelChallenged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseChannelChallenged is a log parse operation binding the contract event 0x07b9206d5a6026d3bd2a8f9a9b79f6fa4bfbd6a016975829fbaf07488019f28a.
//
// Solidity: event ChannelChallenged(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate, uint64 challengeExpireAt)
func (_ChannelHub *ChannelHubFilterer) ParseChannelChallenged(log types.Log) (*ChannelHubChannelChallenged, error) {
	event := new(ChannelHubChannelChallenged)
	if err := _ChannelHub.contract.UnpackLog(event, "ChannelChallenged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubChannelCheckpointedIterator is returned from FilterChannelCheckpointed and is used to iterate over the raw logs and unpacked data for ChannelCheckpointed events raised by the ChannelHub contract.
type ChannelHubChannelCheckpointedIterator struct {
	Event *ChannelHubChannelCheckpointed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubChannelCheckpointedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubChannelCheckpointed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubChannelCheckpointed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubChannelCheckpointedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubChannelCheckpointedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubChannelCheckpointed represents a ChannelCheckpointed event raised by the ChannelHub contract.
type ChannelHubChannelCheckpointed struct {
	ChannelId [32]byte
	Candidate State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterChannelCheckpointed is a free log retrieval operation binding the contract event 0x567044ba1cdd4671ac3979c114241e1e3b56c9e9051f63f2f234f7a2795019cc.
//
// Solidity: event ChannelCheckpointed(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate)
func (_ChannelHub *ChannelHubFilterer) FilterChannelCheckpointed(opts *bind.FilterOpts, channelId [][32]byte) (*ChannelHubChannelCheckpointedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "ChannelCheckpointed", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubChannelCheckpointedIterator{contract: _ChannelHub.contract, event: "ChannelCheckpointed", logs: logs, sub: sub}, nil
}

// WatchChannelCheckpointed is a free log subscription operation binding the contract event 0x567044ba1cdd4671ac3979c114241e1e3b56c9e9051f63f2f234f7a2795019cc.
//
// Solidity: event ChannelCheckpointed(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate)
func (_ChannelHub *ChannelHubFilterer) WatchChannelCheckpointed(opts *bind.WatchOpts, sink chan<- *ChannelHubChannelCheckpointed, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "ChannelCheckpointed", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubChannelCheckpointed)
				if err := _ChannelHub.contract.UnpackLog(event, "ChannelCheckpointed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseChannelCheckpointed is a log parse operation binding the contract event 0x567044ba1cdd4671ac3979c114241e1e3b56c9e9051f63f2f234f7a2795019cc.
//
// Solidity: event ChannelCheckpointed(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate)
func (_ChannelHub *ChannelHubFilterer) ParseChannelCheckpointed(log types.Log) (*ChannelHubChannelCheckpointed, error) {
	event := new(ChannelHubChannelCheckpointed)
	if err := _ChannelHub.contract.UnpackLog(event, "ChannelCheckpointed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubChannelClosedIterator is returned from FilterChannelClosed and is used to iterate over the raw logs and unpacked data for ChannelClosed events raised by the ChannelHub contract.
type ChannelHubChannelClosedIterator struct {
	Event *ChannelHubChannelClosed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubChannelClosedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubChannelClosed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubChannelClosed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubChannelClosedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubChannelClosedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubChannelClosed represents a ChannelClosed event raised by the ChannelHub contract.
type ChannelHubChannelClosed struct {
	ChannelId  [32]byte
	FinalState State
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterChannelClosed is a free log retrieval operation binding the contract event 0x04cd8c68bf83e7bc531ca5a5d75c34e36513c2acf81e07e6470ba79e29da13a8.
//
// Solidity: event ChannelClosed(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) finalState)
func (_ChannelHub *ChannelHubFilterer) FilterChannelClosed(opts *bind.FilterOpts, channelId [][32]byte) (*ChannelHubChannelClosedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "ChannelClosed", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubChannelClosedIterator{contract: _ChannelHub.contract, event: "ChannelClosed", logs: logs, sub: sub}, nil
}

// WatchChannelClosed is a free log subscription operation binding the contract event 0x04cd8c68bf83e7bc531ca5a5d75c34e36513c2acf81e07e6470ba79e29da13a8.
//
// Solidity: event ChannelClosed(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) finalState)
func (_ChannelHub *ChannelHubFilterer) WatchChannelClosed(opts *bind.WatchOpts, sink chan<- *ChannelHubChannelClosed, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "ChannelClosed", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubChannelClosed)
				if err := _ChannelHub.contract.UnpackLog(event, "ChannelClosed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseChannelClosed is a log parse operation binding the contract event 0x04cd8c68bf83e7bc531ca5a5d75c34e36513c2acf81e07e6470ba79e29da13a8.
//
// Solidity: event ChannelClosed(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) finalState)
func (_ChannelHub *ChannelHubFilterer) ParseChannelClosed(log types.Log) (*ChannelHubChannelClosed, error) {
	event := new(ChannelHubChannelClosed)
	if err := _ChannelHub.contract.UnpackLog(event, "ChannelClosed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubChannelCreatedIterator is returned from FilterChannelCreated and is used to iterate over the raw logs and unpacked data for ChannelCreated events raised by the ChannelHub contract.
type ChannelHubChannelCreatedIterator struct {
	Event *ChannelHubChannelCreated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubChannelCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubChannelCreated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubChannelCreated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubChannelCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubChannelCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubChannelCreated represents a ChannelCreated event raised by the ChannelHub contract.
type ChannelHubChannelCreated struct {
	ChannelId    [32]byte
	User         common.Address
	Node         common.Address
	Definition   ChannelDefinition
	InitialState State
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterChannelCreated is a free log retrieval operation binding the contract event 0xb00e209e275d0e1892f1982b34d3f545d1628aebd95322d7ce3585c558f638b4.
//
// Solidity: event ChannelCreated(bytes32 indexed channelId, address indexed user, address indexed node, (uint32,address,address,uint64,uint256,bytes32) definition, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initialState)
func (_ChannelHub *ChannelHubFilterer) FilterChannelCreated(opts *bind.FilterOpts, channelId [][32]byte, user []common.Address, node []common.Address) (*ChannelHubChannelCreatedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}
	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "ChannelCreated", channelIdRule, userRule, nodeRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubChannelCreatedIterator{contract: _ChannelHub.contract, event: "ChannelCreated", logs: logs, sub: sub}, nil
}

// WatchChannelCreated is a free log subscription operation binding the contract event 0xb00e209e275d0e1892f1982b34d3f545d1628aebd95322d7ce3585c558f638b4.
//
// Solidity: event ChannelCreated(bytes32 indexed channelId, address indexed user, address indexed node, (uint32,address,address,uint64,uint256,bytes32) definition, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initialState)
func (_ChannelHub *ChannelHubFilterer) WatchChannelCreated(opts *bind.WatchOpts, sink chan<- *ChannelHubChannelCreated, channelId [][32]byte, user []common.Address, node []common.Address) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}
	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "ChannelCreated", channelIdRule, userRule, nodeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubChannelCreated)
				if err := _ChannelHub.contract.UnpackLog(event, "ChannelCreated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseChannelCreated is a log parse operation binding the contract event 0xb00e209e275d0e1892f1982b34d3f545d1628aebd95322d7ce3585c558f638b4.
//
// Solidity: event ChannelCreated(bytes32 indexed channelId, address indexed user, address indexed node, (uint32,address,address,uint64,uint256,bytes32) definition, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initialState)
func (_ChannelHub *ChannelHubFilterer) ParseChannelCreated(log types.Log) (*ChannelHubChannelCreated, error) {
	event := new(ChannelHubChannelCreated)
	if err := _ChannelHub.contract.UnpackLog(event, "ChannelCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubChannelDepositedIterator is returned from FilterChannelDeposited and is used to iterate over the raw logs and unpacked data for ChannelDeposited events raised by the ChannelHub contract.
type ChannelHubChannelDepositedIterator struct {
	Event *ChannelHubChannelDeposited // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubChannelDepositedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubChannelDeposited)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubChannelDeposited)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubChannelDepositedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubChannelDepositedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubChannelDeposited represents a ChannelDeposited event raised by the ChannelHub contract.
type ChannelHubChannelDeposited struct {
	ChannelId [32]byte
	Candidate State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterChannelDeposited is a free log retrieval operation binding the contract event 0x6085f5128b19e0d3cc37524413de47259383f0f75265d5d66f41778696206696.
//
// Solidity: event ChannelDeposited(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate)
func (_ChannelHub *ChannelHubFilterer) FilterChannelDeposited(opts *bind.FilterOpts, channelId [][32]byte) (*ChannelHubChannelDepositedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "ChannelDeposited", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubChannelDepositedIterator{contract: _ChannelHub.contract, event: "ChannelDeposited", logs: logs, sub: sub}, nil
}

// WatchChannelDeposited is a free log subscription operation binding the contract event 0x6085f5128b19e0d3cc37524413de47259383f0f75265d5d66f41778696206696.
//
// Solidity: event ChannelDeposited(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate)
func (_ChannelHub *ChannelHubFilterer) WatchChannelDeposited(opts *bind.WatchOpts, sink chan<- *ChannelHubChannelDeposited, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "ChannelDeposited", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubChannelDeposited)
				if err := _ChannelHub.contract.UnpackLog(event, "ChannelDeposited", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseChannelDeposited is a log parse operation binding the contract event 0x6085f5128b19e0d3cc37524413de47259383f0f75265d5d66f41778696206696.
//
// Solidity: event ChannelDeposited(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate)
func (_ChannelHub *ChannelHubFilterer) ParseChannelDeposited(log types.Log) (*ChannelHubChannelDeposited, error) {
	event := new(ChannelHubChannelDeposited)
	if err := _ChannelHub.contract.UnpackLog(event, "ChannelDeposited", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubChannelWithdrawnIterator is returned from FilterChannelWithdrawn and is used to iterate over the raw logs and unpacked data for ChannelWithdrawn events raised by the ChannelHub contract.
type ChannelHubChannelWithdrawnIterator struct {
	Event *ChannelHubChannelWithdrawn // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubChannelWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubChannelWithdrawn)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubChannelWithdrawn)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubChannelWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubChannelWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubChannelWithdrawn represents a ChannelWithdrawn event raised by the ChannelHub contract.
type ChannelHubChannelWithdrawn struct {
	ChannelId [32]byte
	Candidate State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterChannelWithdrawn is a free log retrieval operation binding the contract event 0x188e0ade7d115cc397426774adb960ae3e8c83e72f0a6cad4b7085e1d60bf986.
//
// Solidity: event ChannelWithdrawn(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate)
func (_ChannelHub *ChannelHubFilterer) FilterChannelWithdrawn(opts *bind.FilterOpts, channelId [][32]byte) (*ChannelHubChannelWithdrawnIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "ChannelWithdrawn", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubChannelWithdrawnIterator{contract: _ChannelHub.contract, event: "ChannelWithdrawn", logs: logs, sub: sub}, nil
}

// WatchChannelWithdrawn is a free log subscription operation binding the contract event 0x188e0ade7d115cc397426774adb960ae3e8c83e72f0a6cad4b7085e1d60bf986.
//
// Solidity: event ChannelWithdrawn(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate)
func (_ChannelHub *ChannelHubFilterer) WatchChannelWithdrawn(opts *bind.WatchOpts, sink chan<- *ChannelHubChannelWithdrawn, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "ChannelWithdrawn", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubChannelWithdrawn)
				if err := _ChannelHub.contract.UnpackLog(event, "ChannelWithdrawn", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseChannelWithdrawn is a log parse operation binding the contract event 0x188e0ade7d115cc397426774adb960ae3e8c83e72f0a6cad4b7085e1d60bf986.
//
// Solidity: event ChannelWithdrawn(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate)
func (_ChannelHub *ChannelHubFilterer) ParseChannelWithdrawn(log types.Log) (*ChannelHubChannelWithdrawn, error) {
	event := new(ChannelHubChannelWithdrawn)
	if err := _ChannelHub.contract.UnpackLog(event, "ChannelWithdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubDepositedIterator is returned from FilterDeposited and is used to iterate over the raw logs and unpacked data for Deposited events raised by the ChannelHub contract.
type ChannelHubDepositedIterator struct {
	Event *ChannelHubDeposited // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubDepositedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubDeposited)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubDeposited)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubDepositedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubDepositedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubDeposited represents a Deposited event raised by the ChannelHub contract.
type ChannelHubDeposited struct {
	Wallet common.Address
	Token  common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDeposited is a free log retrieval operation binding the contract event 0x8752a472e571a816aea92eec8dae9baf628e840f4929fbcc2d155e6233ff68a7.
//
// Solidity: event Deposited(address indexed wallet, address indexed token, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) FilterDeposited(opts *bind.FilterOpts, wallet []common.Address, token []common.Address) (*ChannelHubDepositedIterator, error) {

	var walletRule []interface{}
	for _, walletItem := range wallet {
		walletRule = append(walletRule, walletItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "Deposited", walletRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubDepositedIterator{contract: _ChannelHub.contract, event: "Deposited", logs: logs, sub: sub}, nil
}

// WatchDeposited is a free log subscription operation binding the contract event 0x8752a472e571a816aea92eec8dae9baf628e840f4929fbcc2d155e6233ff68a7.
//
// Solidity: event Deposited(address indexed wallet, address indexed token, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) WatchDeposited(opts *bind.WatchOpts, sink chan<- *ChannelHubDeposited, wallet []common.Address, token []common.Address) (event.Subscription, error) {

	var walletRule []interface{}
	for _, walletItem := range wallet {
		walletRule = append(walletRule, walletItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "Deposited", walletRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubDeposited)
				if err := _ChannelHub.contract.UnpackLog(event, "Deposited", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDeposited is a log parse operation binding the contract event 0x8752a472e571a816aea92eec8dae9baf628e840f4929fbcc2d155e6233ff68a7.
//
// Solidity: event Deposited(address indexed wallet, address indexed token, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) ParseDeposited(log types.Log) (*ChannelHubDeposited, error) {
	event := new(ChannelHubDeposited)
	if err := _ChannelHub.contract.UnpackLog(event, "Deposited", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowDepositChallengedIterator is returned from FilterEscrowDepositChallenged and is used to iterate over the raw logs and unpacked data for EscrowDepositChallenged events raised by the ChannelHub contract.
type ChannelHubEscrowDepositChallengedIterator struct {
	Event *ChannelHubEscrowDepositChallenged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowDepositChallengedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowDepositChallenged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowDepositChallenged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowDepositChallengedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowDepositChallengedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowDepositChallenged represents a EscrowDepositChallenged event raised by the ChannelHub contract.
type ChannelHubEscrowDepositChallenged struct {
	EscrowId          [32]byte
	State             State
	ChallengeExpireAt uint64
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterEscrowDepositChallenged is a free log retrieval operation binding the contract event 0xba075bd445233f7cad862c72f0343b3503aad9c8e704a2295f122b82abf8e801.
//
// Solidity: event EscrowDepositChallenged(bytes32 indexed escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state, uint64 challengeExpireAt)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowDepositChallenged(opts *bind.FilterOpts, escrowId [][32]byte) (*ChannelHubEscrowDepositChallengedIterator, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowDepositChallenged", escrowIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowDepositChallengedIterator{contract: _ChannelHub.contract, event: "EscrowDepositChallenged", logs: logs, sub: sub}, nil
}

// WatchEscrowDepositChallenged is a free log subscription operation binding the contract event 0xba075bd445233f7cad862c72f0343b3503aad9c8e704a2295f122b82abf8e801.
//
// Solidity: event EscrowDepositChallenged(bytes32 indexed escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state, uint64 challengeExpireAt)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowDepositChallenged(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowDepositChallenged, escrowId [][32]byte) (event.Subscription, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowDepositChallenged", escrowIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowDepositChallenged)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositChallenged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowDepositChallenged is a log parse operation binding the contract event 0xba075bd445233f7cad862c72f0343b3503aad9c8e704a2295f122b82abf8e801.
//
// Solidity: event EscrowDepositChallenged(bytes32 indexed escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state, uint64 challengeExpireAt)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowDepositChallenged(log types.Log) (*ChannelHubEscrowDepositChallenged, error) {
	event := new(ChannelHubEscrowDepositChallenged)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositChallenged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowDepositFinalizedIterator is returned from FilterEscrowDepositFinalized and is used to iterate over the raw logs and unpacked data for EscrowDepositFinalized events raised by the ChannelHub contract.
type ChannelHubEscrowDepositFinalizedIterator struct {
	Event *ChannelHubEscrowDepositFinalized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowDepositFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowDepositFinalized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowDepositFinalized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowDepositFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowDepositFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowDepositFinalized represents a EscrowDepositFinalized event raised by the ChannelHub contract.
type ChannelHubEscrowDepositFinalized struct {
	EscrowId  [32]byte
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEscrowDepositFinalized is a free log retrieval operation binding the contract event 0x1b92e8ef67d8a7c0d29c99efcd180a5e0d98d60ac41d52abbbb5950882c78e4e.
//
// Solidity: event EscrowDepositFinalized(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowDepositFinalized(opts *bind.FilterOpts, escrowId [][32]byte, channelId [][32]byte) (*ChannelHubEscrowDepositFinalizedIterator, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowDepositFinalized", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowDepositFinalizedIterator{contract: _ChannelHub.contract, event: "EscrowDepositFinalized", logs: logs, sub: sub}, nil
}

// WatchEscrowDepositFinalized is a free log subscription operation binding the contract event 0x1b92e8ef67d8a7c0d29c99efcd180a5e0d98d60ac41d52abbbb5950882c78e4e.
//
// Solidity: event EscrowDepositFinalized(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowDepositFinalized(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowDepositFinalized, escrowId [][32]byte, channelId [][32]byte) (event.Subscription, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowDepositFinalized", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowDepositFinalized)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositFinalized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowDepositFinalized is a log parse operation binding the contract event 0x1b92e8ef67d8a7c0d29c99efcd180a5e0d98d60ac41d52abbbb5950882c78e4e.
//
// Solidity: event EscrowDepositFinalized(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowDepositFinalized(log types.Log) (*ChannelHubEscrowDepositFinalized, error) {
	event := new(ChannelHubEscrowDepositFinalized)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowDepositFinalizedOnHomeIterator is returned from FilterEscrowDepositFinalizedOnHome and is used to iterate over the raw logs and unpacked data for EscrowDepositFinalizedOnHome events raised by the ChannelHub contract.
type ChannelHubEscrowDepositFinalizedOnHomeIterator struct {
	Event *ChannelHubEscrowDepositFinalizedOnHome // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowDepositFinalizedOnHomeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowDepositFinalizedOnHome)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowDepositFinalizedOnHome)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowDepositFinalizedOnHomeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowDepositFinalizedOnHomeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowDepositFinalizedOnHome represents a EscrowDepositFinalizedOnHome event raised by the ChannelHub contract.
type ChannelHubEscrowDepositFinalizedOnHome struct {
	EscrowId  [32]byte
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEscrowDepositFinalizedOnHome is a free log retrieval operation binding the contract event 0x32e24720f56fd5a7f4cb219d7ff3278ae95196e79c85b5801395894a6f53466c.
//
// Solidity: event EscrowDepositFinalizedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowDepositFinalizedOnHome(opts *bind.FilterOpts, escrowId [][32]byte, channelId [][32]byte) (*ChannelHubEscrowDepositFinalizedOnHomeIterator, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowDepositFinalizedOnHome", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowDepositFinalizedOnHomeIterator{contract: _ChannelHub.contract, event: "EscrowDepositFinalizedOnHome", logs: logs, sub: sub}, nil
}

// WatchEscrowDepositFinalizedOnHome is a free log subscription operation binding the contract event 0x32e24720f56fd5a7f4cb219d7ff3278ae95196e79c85b5801395894a6f53466c.
//
// Solidity: event EscrowDepositFinalizedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowDepositFinalizedOnHome(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowDepositFinalizedOnHome, escrowId [][32]byte, channelId [][32]byte) (event.Subscription, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowDepositFinalizedOnHome", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowDepositFinalizedOnHome)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositFinalizedOnHome", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowDepositFinalizedOnHome is a log parse operation binding the contract event 0x32e24720f56fd5a7f4cb219d7ff3278ae95196e79c85b5801395894a6f53466c.
//
// Solidity: event EscrowDepositFinalizedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowDepositFinalizedOnHome(log types.Log) (*ChannelHubEscrowDepositFinalizedOnHome, error) {
	event := new(ChannelHubEscrowDepositFinalizedOnHome)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositFinalizedOnHome", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowDepositInitiatedIterator is returned from FilterEscrowDepositInitiated and is used to iterate over the raw logs and unpacked data for EscrowDepositInitiated events raised by the ChannelHub contract.
type ChannelHubEscrowDepositInitiatedIterator struct {
	Event *ChannelHubEscrowDepositInitiated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowDepositInitiatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowDepositInitiated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowDepositInitiated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowDepositInitiatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowDepositInitiatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowDepositInitiated represents a EscrowDepositInitiated event raised by the ChannelHub contract.
type ChannelHubEscrowDepositInitiated struct {
	EscrowId  [32]byte
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEscrowDepositInitiated is a free log retrieval operation binding the contract event 0xede7867afa7cdb9c443667efd8244d98bf9df1dce68e60dc94dca6605125ca76.
//
// Solidity: event EscrowDepositInitiated(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowDepositInitiated(opts *bind.FilterOpts, escrowId [][32]byte, channelId [][32]byte) (*ChannelHubEscrowDepositInitiatedIterator, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowDepositInitiated", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowDepositInitiatedIterator{contract: _ChannelHub.contract, event: "EscrowDepositInitiated", logs: logs, sub: sub}, nil
}

// WatchEscrowDepositInitiated is a free log subscription operation binding the contract event 0xede7867afa7cdb9c443667efd8244d98bf9df1dce68e60dc94dca6605125ca76.
//
// Solidity: event EscrowDepositInitiated(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowDepositInitiated(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowDepositInitiated, escrowId [][32]byte, channelId [][32]byte) (event.Subscription, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowDepositInitiated", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowDepositInitiated)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositInitiated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowDepositInitiated is a log parse operation binding the contract event 0xede7867afa7cdb9c443667efd8244d98bf9df1dce68e60dc94dca6605125ca76.
//
// Solidity: event EscrowDepositInitiated(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowDepositInitiated(log types.Log) (*ChannelHubEscrowDepositInitiated, error) {
	event := new(ChannelHubEscrowDepositInitiated)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositInitiated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowDepositInitiatedOnHomeIterator is returned from FilterEscrowDepositInitiatedOnHome and is used to iterate over the raw logs and unpacked data for EscrowDepositInitiatedOnHome events raised by the ChannelHub contract.
type ChannelHubEscrowDepositInitiatedOnHomeIterator struct {
	Event *ChannelHubEscrowDepositInitiatedOnHome // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowDepositInitiatedOnHomeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowDepositInitiatedOnHome)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowDepositInitiatedOnHome)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowDepositInitiatedOnHomeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowDepositInitiatedOnHomeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowDepositInitiatedOnHome represents a EscrowDepositInitiatedOnHome event raised by the ChannelHub contract.
type ChannelHubEscrowDepositInitiatedOnHome struct {
	EscrowId  [32]byte
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEscrowDepositInitiatedOnHome is a free log retrieval operation binding the contract event 0x471c4ebe4e57d25ef7117e141caac31c6b98f067b8098a7a7bbd38f637c2f980.
//
// Solidity: event EscrowDepositInitiatedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowDepositInitiatedOnHome(opts *bind.FilterOpts, escrowId [][32]byte, channelId [][32]byte) (*ChannelHubEscrowDepositInitiatedOnHomeIterator, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowDepositInitiatedOnHome", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowDepositInitiatedOnHomeIterator{contract: _ChannelHub.contract, event: "EscrowDepositInitiatedOnHome", logs: logs, sub: sub}, nil
}

// WatchEscrowDepositInitiatedOnHome is a free log subscription operation binding the contract event 0x471c4ebe4e57d25ef7117e141caac31c6b98f067b8098a7a7bbd38f637c2f980.
//
// Solidity: event EscrowDepositInitiatedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowDepositInitiatedOnHome(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowDepositInitiatedOnHome, escrowId [][32]byte, channelId [][32]byte) (event.Subscription, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowDepositInitiatedOnHome", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowDepositInitiatedOnHome)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositInitiatedOnHome", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowDepositInitiatedOnHome is a log parse operation binding the contract event 0x471c4ebe4e57d25ef7117e141caac31c6b98f067b8098a7a7bbd38f637c2f980.
//
// Solidity: event EscrowDepositInitiatedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowDepositInitiatedOnHome(log types.Log) (*ChannelHubEscrowDepositInitiatedOnHome, error) {
	event := new(ChannelHubEscrowDepositInitiatedOnHome)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositInitiatedOnHome", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowDepositsPurgedIterator is returned from FilterEscrowDepositsPurged and is used to iterate over the raw logs and unpacked data for EscrowDepositsPurged events raised by the ChannelHub contract.
type ChannelHubEscrowDepositsPurgedIterator struct {
	Event *ChannelHubEscrowDepositsPurged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowDepositsPurgedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowDepositsPurged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowDepositsPurged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowDepositsPurgedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowDepositsPurgedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowDepositsPurged represents a EscrowDepositsPurged event raised by the ChannelHub contract.
type ChannelHubEscrowDepositsPurged struct {
	PurgedCount *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterEscrowDepositsPurged is a free log retrieval operation binding the contract event 0x61815f4b11c6ea4e14a2e448a010bed8efdc3e53a15efbf183d16a31085cd145.
//
// Solidity: event EscrowDepositsPurged(uint256 purgedCount)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowDepositsPurged(opts *bind.FilterOpts) (*ChannelHubEscrowDepositsPurgedIterator, error) {

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowDepositsPurged")
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowDepositsPurgedIterator{contract: _ChannelHub.contract, event: "EscrowDepositsPurged", logs: logs, sub: sub}, nil
}

// WatchEscrowDepositsPurged is a free log subscription operation binding the contract event 0x61815f4b11c6ea4e14a2e448a010bed8efdc3e53a15efbf183d16a31085cd145.
//
// Solidity: event EscrowDepositsPurged(uint256 purgedCount)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowDepositsPurged(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowDepositsPurged) (event.Subscription, error) {

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowDepositsPurged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowDepositsPurged)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositsPurged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowDepositsPurged is a log parse operation binding the contract event 0x61815f4b11c6ea4e14a2e448a010bed8efdc3e53a15efbf183d16a31085cd145.
//
// Solidity: event EscrowDepositsPurged(uint256 purgedCount)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowDepositsPurged(log types.Log) (*ChannelHubEscrowDepositsPurged, error) {
	event := new(ChannelHubEscrowDepositsPurged)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositsPurged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowWithdrawalChallengedIterator is returned from FilterEscrowWithdrawalChallenged and is used to iterate over the raw logs and unpacked data for EscrowWithdrawalChallenged events raised by the ChannelHub contract.
type ChannelHubEscrowWithdrawalChallengedIterator struct {
	Event *ChannelHubEscrowWithdrawalChallenged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowWithdrawalChallengedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowWithdrawalChallenged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowWithdrawalChallenged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowWithdrawalChallengedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowWithdrawalChallengedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowWithdrawalChallenged represents a EscrowWithdrawalChallenged event raised by the ChannelHub contract.
type ChannelHubEscrowWithdrawalChallenged struct {
	EscrowId          [32]byte
	State             State
	ChallengeExpireAt uint64
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterEscrowWithdrawalChallenged is a free log retrieval operation binding the contract event 0xb8568a1f475f3c76759a620e08a653d28348c5c09e2e0bc91d533339801fefd8.
//
// Solidity: event EscrowWithdrawalChallenged(bytes32 indexed escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state, uint64 challengeExpireAt)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowWithdrawalChallenged(opts *bind.FilterOpts, escrowId [][32]byte) (*ChannelHubEscrowWithdrawalChallengedIterator, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowWithdrawalChallenged", escrowIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowWithdrawalChallengedIterator{contract: _ChannelHub.contract, event: "EscrowWithdrawalChallenged", logs: logs, sub: sub}, nil
}

// WatchEscrowWithdrawalChallenged is a free log subscription operation binding the contract event 0xb8568a1f475f3c76759a620e08a653d28348c5c09e2e0bc91d533339801fefd8.
//
// Solidity: event EscrowWithdrawalChallenged(bytes32 indexed escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state, uint64 challengeExpireAt)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowWithdrawalChallenged(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowWithdrawalChallenged, escrowId [][32]byte) (event.Subscription, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowWithdrawalChallenged", escrowIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowWithdrawalChallenged)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowWithdrawalChallenged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowWithdrawalChallenged is a log parse operation binding the contract event 0xb8568a1f475f3c76759a620e08a653d28348c5c09e2e0bc91d533339801fefd8.
//
// Solidity: event EscrowWithdrawalChallenged(bytes32 indexed escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state, uint64 challengeExpireAt)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowWithdrawalChallenged(log types.Log) (*ChannelHubEscrowWithdrawalChallenged, error) {
	event := new(ChannelHubEscrowWithdrawalChallenged)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowWithdrawalChallenged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowWithdrawalFinalizedIterator is returned from FilterEscrowWithdrawalFinalized and is used to iterate over the raw logs and unpacked data for EscrowWithdrawalFinalized events raised by the ChannelHub contract.
type ChannelHubEscrowWithdrawalFinalizedIterator struct {
	Event *ChannelHubEscrowWithdrawalFinalized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowWithdrawalFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowWithdrawalFinalized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowWithdrawalFinalized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowWithdrawalFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowWithdrawalFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowWithdrawalFinalized represents a EscrowWithdrawalFinalized event raised by the ChannelHub contract.
type ChannelHubEscrowWithdrawalFinalized struct {
	EscrowId  [32]byte
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEscrowWithdrawalFinalized is a free log retrieval operation binding the contract event 0x2fdac1380dbe23ae259b6871582b7f33e34461547f400bdd20d74991250317d1.
//
// Solidity: event EscrowWithdrawalFinalized(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowWithdrawalFinalized(opts *bind.FilterOpts, escrowId [][32]byte, channelId [][32]byte) (*ChannelHubEscrowWithdrawalFinalizedIterator, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowWithdrawalFinalized", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowWithdrawalFinalizedIterator{contract: _ChannelHub.contract, event: "EscrowWithdrawalFinalized", logs: logs, sub: sub}, nil
}

// WatchEscrowWithdrawalFinalized is a free log subscription operation binding the contract event 0x2fdac1380dbe23ae259b6871582b7f33e34461547f400bdd20d74991250317d1.
//
// Solidity: event EscrowWithdrawalFinalized(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowWithdrawalFinalized(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowWithdrawalFinalized, escrowId [][32]byte, channelId [][32]byte) (event.Subscription, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowWithdrawalFinalized", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowWithdrawalFinalized)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowWithdrawalFinalized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowWithdrawalFinalized is a log parse operation binding the contract event 0x2fdac1380dbe23ae259b6871582b7f33e34461547f400bdd20d74991250317d1.
//
// Solidity: event EscrowWithdrawalFinalized(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowWithdrawalFinalized(log types.Log) (*ChannelHubEscrowWithdrawalFinalized, error) {
	event := new(ChannelHubEscrowWithdrawalFinalized)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowWithdrawalFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowWithdrawalFinalizedOnHomeIterator is returned from FilterEscrowWithdrawalFinalizedOnHome and is used to iterate over the raw logs and unpacked data for EscrowWithdrawalFinalizedOnHome events raised by the ChannelHub contract.
type ChannelHubEscrowWithdrawalFinalizedOnHomeIterator struct {
	Event *ChannelHubEscrowWithdrawalFinalizedOnHome // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowWithdrawalFinalizedOnHomeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowWithdrawalFinalizedOnHome)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowWithdrawalFinalizedOnHome)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowWithdrawalFinalizedOnHomeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowWithdrawalFinalizedOnHomeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowWithdrawalFinalizedOnHome represents a EscrowWithdrawalFinalizedOnHome event raised by the ChannelHub contract.
type ChannelHubEscrowWithdrawalFinalizedOnHome struct {
	EscrowId  [32]byte
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEscrowWithdrawalFinalizedOnHome is a free log retrieval operation binding the contract event 0x6d0cf3d243d63f08f50db493a8af34b27d4e3bc9ec4098e82700abfeffe2d498.
//
// Solidity: event EscrowWithdrawalFinalizedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowWithdrawalFinalizedOnHome(opts *bind.FilterOpts, escrowId [][32]byte, channelId [][32]byte) (*ChannelHubEscrowWithdrawalFinalizedOnHomeIterator, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowWithdrawalFinalizedOnHome", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowWithdrawalFinalizedOnHomeIterator{contract: _ChannelHub.contract, event: "EscrowWithdrawalFinalizedOnHome", logs: logs, sub: sub}, nil
}

// WatchEscrowWithdrawalFinalizedOnHome is a free log subscription operation binding the contract event 0x6d0cf3d243d63f08f50db493a8af34b27d4e3bc9ec4098e82700abfeffe2d498.
//
// Solidity: event EscrowWithdrawalFinalizedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowWithdrawalFinalizedOnHome(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowWithdrawalFinalizedOnHome, escrowId [][32]byte, channelId [][32]byte) (event.Subscription, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowWithdrawalFinalizedOnHome", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowWithdrawalFinalizedOnHome)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowWithdrawalFinalizedOnHome", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowWithdrawalFinalizedOnHome is a log parse operation binding the contract event 0x6d0cf3d243d63f08f50db493a8af34b27d4e3bc9ec4098e82700abfeffe2d498.
//
// Solidity: event EscrowWithdrawalFinalizedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowWithdrawalFinalizedOnHome(log types.Log) (*ChannelHubEscrowWithdrawalFinalizedOnHome, error) {
	event := new(ChannelHubEscrowWithdrawalFinalizedOnHome)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowWithdrawalFinalizedOnHome", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowWithdrawalInitiatedIterator is returned from FilterEscrowWithdrawalInitiated and is used to iterate over the raw logs and unpacked data for EscrowWithdrawalInitiated events raised by the ChannelHub contract.
type ChannelHubEscrowWithdrawalInitiatedIterator struct {
	Event *ChannelHubEscrowWithdrawalInitiated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowWithdrawalInitiatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowWithdrawalInitiated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowWithdrawalInitiated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowWithdrawalInitiatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowWithdrawalInitiatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowWithdrawalInitiated represents a EscrowWithdrawalInitiated event raised by the ChannelHub contract.
type ChannelHubEscrowWithdrawalInitiated struct {
	EscrowId  [32]byte
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEscrowWithdrawalInitiated is a free log retrieval operation binding the contract event 0x17eb0a6bd5a0de45d1029ce3444941070e149df35b22176fc439f930f73c09f7.
//
// Solidity: event EscrowWithdrawalInitiated(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowWithdrawalInitiated(opts *bind.FilterOpts, escrowId [][32]byte, channelId [][32]byte) (*ChannelHubEscrowWithdrawalInitiatedIterator, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowWithdrawalInitiated", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowWithdrawalInitiatedIterator{contract: _ChannelHub.contract, event: "EscrowWithdrawalInitiated", logs: logs, sub: sub}, nil
}

// WatchEscrowWithdrawalInitiated is a free log subscription operation binding the contract event 0x17eb0a6bd5a0de45d1029ce3444941070e149df35b22176fc439f930f73c09f7.
//
// Solidity: event EscrowWithdrawalInitiated(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowWithdrawalInitiated(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowWithdrawalInitiated, escrowId [][32]byte, channelId [][32]byte) (event.Subscription, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowWithdrawalInitiated", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowWithdrawalInitiated)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowWithdrawalInitiated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowWithdrawalInitiated is a log parse operation binding the contract event 0x17eb0a6bd5a0de45d1029ce3444941070e149df35b22176fc439f930f73c09f7.
//
// Solidity: event EscrowWithdrawalInitiated(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowWithdrawalInitiated(log types.Log) (*ChannelHubEscrowWithdrawalInitiated, error) {
	event := new(ChannelHubEscrowWithdrawalInitiated)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowWithdrawalInitiated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowWithdrawalInitiatedOnHomeIterator is returned from FilterEscrowWithdrawalInitiatedOnHome and is used to iterate over the raw logs and unpacked data for EscrowWithdrawalInitiatedOnHome events raised by the ChannelHub contract.
type ChannelHubEscrowWithdrawalInitiatedOnHomeIterator struct {
	Event *ChannelHubEscrowWithdrawalInitiatedOnHome // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowWithdrawalInitiatedOnHomeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowWithdrawalInitiatedOnHome)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowWithdrawalInitiatedOnHome)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowWithdrawalInitiatedOnHomeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowWithdrawalInitiatedOnHomeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowWithdrawalInitiatedOnHome represents a EscrowWithdrawalInitiatedOnHome event raised by the ChannelHub contract.
type ChannelHubEscrowWithdrawalInitiatedOnHome struct {
	EscrowId  [32]byte
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEscrowWithdrawalInitiatedOnHome is a free log retrieval operation binding the contract event 0x587faad1bcd589ce902468251883e1976a645af8563c773eed7356d78433210c.
//
// Solidity: event EscrowWithdrawalInitiatedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowWithdrawalInitiatedOnHome(opts *bind.FilterOpts, escrowId [][32]byte, channelId [][32]byte) (*ChannelHubEscrowWithdrawalInitiatedOnHomeIterator, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowWithdrawalInitiatedOnHome", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowWithdrawalInitiatedOnHomeIterator{contract: _ChannelHub.contract, event: "EscrowWithdrawalInitiatedOnHome", logs: logs, sub: sub}, nil
}

// WatchEscrowWithdrawalInitiatedOnHome is a free log subscription operation binding the contract event 0x587faad1bcd589ce902468251883e1976a645af8563c773eed7356d78433210c.
//
// Solidity: event EscrowWithdrawalInitiatedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowWithdrawalInitiatedOnHome(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowWithdrawalInitiatedOnHome, escrowId [][32]byte, channelId [][32]byte) (event.Subscription, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowWithdrawalInitiatedOnHome", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowWithdrawalInitiatedOnHome)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowWithdrawalInitiatedOnHome", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowWithdrawalInitiatedOnHome is a log parse operation binding the contract event 0x587faad1bcd589ce902468251883e1976a645af8563c773eed7356d78433210c.
//
// Solidity: event EscrowWithdrawalInitiatedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowWithdrawalInitiatedOnHome(log types.Log) (*ChannelHubEscrowWithdrawalInitiatedOnHome, error) {
	event := new(ChannelHubEscrowWithdrawalInitiatedOnHome)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowWithdrawalInitiatedOnHome", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubFundsClaimedIterator is returned from FilterFundsClaimed and is used to iterate over the raw logs and unpacked data for FundsClaimed events raised by the ChannelHub contract.
type ChannelHubFundsClaimedIterator struct {
	Event *ChannelHubFundsClaimed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubFundsClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubFundsClaimed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubFundsClaimed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubFundsClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubFundsClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubFundsClaimed represents a FundsClaimed event raised by the ChannelHub contract.
type ChannelHubFundsClaimed struct {
	Account     common.Address
	Token       common.Address
	Destination common.Address
	Amount      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterFundsClaimed is a free log retrieval operation binding the contract event 0x7b8d70738154be94a9a068a6d2f5dd8cfc65c52855859dc8f47de1ff185f8b55.
//
// Solidity: event FundsClaimed(address indexed account, address indexed token, address indexed destination, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) FilterFundsClaimed(opts *bind.FilterOpts, account []common.Address, token []common.Address, destination []common.Address) (*ChannelHubFundsClaimedIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var destinationRule []interface{}
	for _, destinationItem := range destination {
		destinationRule = append(destinationRule, destinationItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "FundsClaimed", accountRule, tokenRule, destinationRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubFundsClaimedIterator{contract: _ChannelHub.contract, event: "FundsClaimed", logs: logs, sub: sub}, nil
}

// WatchFundsClaimed is a free log subscription operation binding the contract event 0x7b8d70738154be94a9a068a6d2f5dd8cfc65c52855859dc8f47de1ff185f8b55.
//
// Solidity: event FundsClaimed(address indexed account, address indexed token, address indexed destination, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) WatchFundsClaimed(opts *bind.WatchOpts, sink chan<- *ChannelHubFundsClaimed, account []common.Address, token []common.Address, destination []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var destinationRule []interface{}
	for _, destinationItem := range destination {
		destinationRule = append(destinationRule, destinationItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "FundsClaimed", accountRule, tokenRule, destinationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubFundsClaimed)
				if err := _ChannelHub.contract.UnpackLog(event, "FundsClaimed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseFundsClaimed is a log parse operation binding the contract event 0x7b8d70738154be94a9a068a6d2f5dd8cfc65c52855859dc8f47de1ff185f8b55.
//
// Solidity: event FundsClaimed(address indexed account, address indexed token, address indexed destination, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) ParseFundsClaimed(log types.Log) (*ChannelHubFundsClaimed, error) {
	event := new(ChannelHubFundsClaimed)
	if err := _ChannelHub.contract.UnpackLog(event, "FundsClaimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubMigrationInFinalizedIterator is returned from FilterMigrationInFinalized and is used to iterate over the raw logs and unpacked data for MigrationInFinalized events raised by the ChannelHub contract.
type ChannelHubMigrationInFinalizedIterator struct {
	Event *ChannelHubMigrationInFinalized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubMigrationInFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubMigrationInFinalized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubMigrationInFinalized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubMigrationInFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubMigrationInFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubMigrationInFinalized represents a MigrationInFinalized event raised by the ChannelHub contract.
type ChannelHubMigrationInFinalized struct {
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterMigrationInFinalized is a free log retrieval operation binding the contract event 0x7b20773c41402791c5f18914dbbeacad38b1ebcc4c55d8eb3bfe0a4cde26c826.
//
// Solidity: event MigrationInFinalized(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterMigrationInFinalized(opts *bind.FilterOpts, channelId [][32]byte) (*ChannelHubMigrationInFinalizedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "MigrationInFinalized", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubMigrationInFinalizedIterator{contract: _ChannelHub.contract, event: "MigrationInFinalized", logs: logs, sub: sub}, nil
}

// WatchMigrationInFinalized is a free log subscription operation binding the contract event 0x7b20773c41402791c5f18914dbbeacad38b1ebcc4c55d8eb3bfe0a4cde26c826.
//
// Solidity: event MigrationInFinalized(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchMigrationInFinalized(opts *bind.WatchOpts, sink chan<- *ChannelHubMigrationInFinalized, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "MigrationInFinalized", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubMigrationInFinalized)
				if err := _ChannelHub.contract.UnpackLog(event, "MigrationInFinalized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMigrationInFinalized is a log parse operation binding the contract event 0x7b20773c41402791c5f18914dbbeacad38b1ebcc4c55d8eb3bfe0a4cde26c826.
//
// Solidity: event MigrationInFinalized(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseMigrationInFinalized(log types.Log) (*ChannelHubMigrationInFinalized, error) {
	event := new(ChannelHubMigrationInFinalized)
	if err := _ChannelHub.contract.UnpackLog(event, "MigrationInFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubMigrationInInitiatedIterator is returned from FilterMigrationInInitiated and is used to iterate over the raw logs and unpacked data for MigrationInInitiated events raised by the ChannelHub contract.
type ChannelHubMigrationInInitiatedIterator struct {
	Event *ChannelHubMigrationInInitiated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubMigrationInInitiatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubMigrationInInitiated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubMigrationInInitiated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubMigrationInInitiatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubMigrationInInitiatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubMigrationInInitiated represents a MigrationInInitiated event raised by the ChannelHub contract.
type ChannelHubMigrationInInitiated struct {
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterMigrationInInitiated is a free log retrieval operation binding the contract event 0x26afbcb9eb52c21f42eb9cfe8f263718ffb65afbf84abe8ad8cce2acfb2242b8.
//
// Solidity: event MigrationInInitiated(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterMigrationInInitiated(opts *bind.FilterOpts, channelId [][32]byte) (*ChannelHubMigrationInInitiatedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "MigrationInInitiated", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubMigrationInInitiatedIterator{contract: _ChannelHub.contract, event: "MigrationInInitiated", logs: logs, sub: sub}, nil
}

// WatchMigrationInInitiated is a free log subscription operation binding the contract event 0x26afbcb9eb52c21f42eb9cfe8f263718ffb65afbf84abe8ad8cce2acfb2242b8.
//
// Solidity: event MigrationInInitiated(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchMigrationInInitiated(opts *bind.WatchOpts, sink chan<- *ChannelHubMigrationInInitiated, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "MigrationInInitiated", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubMigrationInInitiated)
				if err := _ChannelHub.contract.UnpackLog(event, "MigrationInInitiated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMigrationInInitiated is a log parse operation binding the contract event 0x26afbcb9eb52c21f42eb9cfe8f263718ffb65afbf84abe8ad8cce2acfb2242b8.
//
// Solidity: event MigrationInInitiated(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseMigrationInInitiated(log types.Log) (*ChannelHubMigrationInInitiated, error) {
	event := new(ChannelHubMigrationInInitiated)
	if err := _ChannelHub.contract.UnpackLog(event, "MigrationInInitiated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubMigrationOutFinalizedIterator is returned from FilterMigrationOutFinalized and is used to iterate over the raw logs and unpacked data for MigrationOutFinalized events raised by the ChannelHub contract.
type ChannelHubMigrationOutFinalizedIterator struct {
	Event *ChannelHubMigrationOutFinalized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubMigrationOutFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubMigrationOutFinalized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubMigrationOutFinalized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubMigrationOutFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubMigrationOutFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubMigrationOutFinalized represents a MigrationOutFinalized event raised by the ChannelHub contract.
type ChannelHubMigrationOutFinalized struct {
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterMigrationOutFinalized is a free log retrieval operation binding the contract event 0x9a6f675cc94b83b55f1ecc0876affd4332a30c92e6faa2aca0199b1b6df922c3.
//
// Solidity: event MigrationOutFinalized(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterMigrationOutFinalized(opts *bind.FilterOpts, channelId [][32]byte) (*ChannelHubMigrationOutFinalizedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "MigrationOutFinalized", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubMigrationOutFinalizedIterator{contract: _ChannelHub.contract, event: "MigrationOutFinalized", logs: logs, sub: sub}, nil
}

// WatchMigrationOutFinalized is a free log subscription operation binding the contract event 0x9a6f675cc94b83b55f1ecc0876affd4332a30c92e6faa2aca0199b1b6df922c3.
//
// Solidity: event MigrationOutFinalized(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchMigrationOutFinalized(opts *bind.WatchOpts, sink chan<- *ChannelHubMigrationOutFinalized, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "MigrationOutFinalized", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubMigrationOutFinalized)
				if err := _ChannelHub.contract.UnpackLog(event, "MigrationOutFinalized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMigrationOutFinalized is a log parse operation binding the contract event 0x9a6f675cc94b83b55f1ecc0876affd4332a30c92e6faa2aca0199b1b6df922c3.
//
// Solidity: event MigrationOutFinalized(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseMigrationOutFinalized(log types.Log) (*ChannelHubMigrationOutFinalized, error) {
	event := new(ChannelHubMigrationOutFinalized)
	if err := _ChannelHub.contract.UnpackLog(event, "MigrationOutFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubMigrationOutInitiatedIterator is returned from FilterMigrationOutInitiated and is used to iterate over the raw logs and unpacked data for MigrationOutInitiated events raised by the ChannelHub contract.
type ChannelHubMigrationOutInitiatedIterator struct {
	Event *ChannelHubMigrationOutInitiated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubMigrationOutInitiatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubMigrationOutInitiated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubMigrationOutInitiated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubMigrationOutInitiatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubMigrationOutInitiatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubMigrationOutInitiated represents a MigrationOutInitiated event raised by the ChannelHub contract.
type ChannelHubMigrationOutInitiated struct {
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterMigrationOutInitiated is a free log retrieval operation binding the contract event 0x3142fb397e715d80415dff7b527bf1c451def4675da6e1199ee1b4588e3f630a.
//
// Solidity: event MigrationOutInitiated(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterMigrationOutInitiated(opts *bind.FilterOpts, channelId [][32]byte) (*ChannelHubMigrationOutInitiatedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "MigrationOutInitiated", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubMigrationOutInitiatedIterator{contract: _ChannelHub.contract, event: "MigrationOutInitiated", logs: logs, sub: sub}, nil
}

// WatchMigrationOutInitiated is a free log subscription operation binding the contract event 0x3142fb397e715d80415dff7b527bf1c451def4675da6e1199ee1b4588e3f630a.
//
// Solidity: event MigrationOutInitiated(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchMigrationOutInitiated(opts *bind.WatchOpts, sink chan<- *ChannelHubMigrationOutInitiated, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "MigrationOutInitiated", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubMigrationOutInitiated)
				if err := _ChannelHub.contract.UnpackLog(event, "MigrationOutInitiated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMigrationOutInitiated is a log parse operation binding the contract event 0x3142fb397e715d80415dff7b527bf1c451def4675da6e1199ee1b4588e3f630a.
//
// Solidity: event MigrationOutInitiated(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseMigrationOutInitiated(log types.Log) (*ChannelHubMigrationOutInitiated, error) {
	event := new(ChannelHubMigrationOutInitiated)
	if err := _ChannelHub.contract.UnpackLog(event, "MigrationOutInitiated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubTransferFailedIterator is returned from FilterTransferFailed and is used to iterate over the raw logs and unpacked data for TransferFailed events raised by the ChannelHub contract.
type ChannelHubTransferFailedIterator struct {
	Event *ChannelHubTransferFailed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubTransferFailedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubTransferFailed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubTransferFailed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubTransferFailedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubTransferFailedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubTransferFailed represents a TransferFailed event raised by the ChannelHub contract.
type ChannelHubTransferFailed struct {
	Recipient common.Address
	Token     common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterTransferFailed is a free log retrieval operation binding the contract event 0xbf182be802245e8ed88e4b8d3e4344c0863dd2a70334f089fd07265389306fcf.
//
// Solidity: event TransferFailed(address indexed recipient, address indexed token, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) FilterTransferFailed(opts *bind.FilterOpts, recipient []common.Address, token []common.Address) (*ChannelHubTransferFailedIterator, error) {

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "TransferFailed", recipientRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubTransferFailedIterator{contract: _ChannelHub.contract, event: "TransferFailed", logs: logs, sub: sub}, nil
}

// WatchTransferFailed is a free log subscription operation binding the contract event 0xbf182be802245e8ed88e4b8d3e4344c0863dd2a70334f089fd07265389306fcf.
//
// Solidity: event TransferFailed(address indexed recipient, address indexed token, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) WatchTransferFailed(opts *bind.WatchOpts, sink chan<- *ChannelHubTransferFailed, recipient []common.Address, token []common.Address) (event.Subscription, error) {

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "TransferFailed", recipientRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubTransferFailed)
				if err := _ChannelHub.contract.UnpackLog(event, "TransferFailed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTransferFailed is a log parse operation binding the contract event 0xbf182be802245e8ed88e4b8d3e4344c0863dd2a70334f089fd07265389306fcf.
//
// Solidity: event TransferFailed(address indexed recipient, address indexed token, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) ParseTransferFailed(log types.Log) (*ChannelHubTransferFailed, error) {
	event := new(ChannelHubTransferFailed)
	if err := _ChannelHub.contract.UnpackLog(event, "TransferFailed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubValidatorRegisteredIterator is returned from FilterValidatorRegistered and is used to iterate over the raw logs and unpacked data for ValidatorRegistered events raised by the ChannelHub contract.
type ChannelHubValidatorRegisteredIterator struct {
	Event *ChannelHubValidatorRegistered // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubValidatorRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubValidatorRegistered)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubValidatorRegistered)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubValidatorRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubValidatorRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubValidatorRegistered represents a ValidatorRegistered event raised by the ChannelHub contract.
type ChannelHubValidatorRegistered struct {
	Node        common.Address
	ValidatorId uint8
	Validator   common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterValidatorRegistered is a free log retrieval operation binding the contract event 0x2366b94a706a0cfc2dca2fe8be9410b6fba2db75e3e9d3f03b3c2fb0b051efad.
//
// Solidity: event ValidatorRegistered(address indexed node, uint8 indexed validatorId, address indexed validator)
func (_ChannelHub *ChannelHubFilterer) FilterValidatorRegistered(opts *bind.FilterOpts, node []common.Address, validatorId []uint8, validator []common.Address) (*ChannelHubValidatorRegisteredIterator, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}
	var validatorIdRule []interface{}
	for _, validatorIdItem := range validatorId {
		validatorIdRule = append(validatorIdRule, validatorIdItem)
	}
	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "ValidatorRegistered", nodeRule, validatorIdRule, validatorRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubValidatorRegisteredIterator{contract: _ChannelHub.contract, event: "ValidatorRegistered", logs: logs, sub: sub}, nil
}

// WatchValidatorRegistered is a free log subscription operation binding the contract event 0x2366b94a706a0cfc2dca2fe8be9410b6fba2db75e3e9d3f03b3c2fb0b051efad.
//
// Solidity: event ValidatorRegistered(address indexed node, uint8 indexed validatorId, address indexed validator)
func (_ChannelHub *ChannelHubFilterer) WatchValidatorRegistered(opts *bind.WatchOpts, sink chan<- *ChannelHubValidatorRegistered, node []common.Address, validatorId []uint8, validator []common.Address) (event.Subscription, error) {

	var nodeRule []interface{}
	for _, nodeItem := range node {
		nodeRule = append(nodeRule, nodeItem)
	}
	var validatorIdRule []interface{}
	for _, validatorIdItem := range validatorId {
		validatorIdRule = append(validatorIdRule, validatorIdItem)
	}
	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "ValidatorRegistered", nodeRule, validatorIdRule, validatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubValidatorRegistered)
				if err := _ChannelHub.contract.UnpackLog(event, "ValidatorRegistered", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseValidatorRegistered is a log parse operation binding the contract event 0x2366b94a706a0cfc2dca2fe8be9410b6fba2db75e3e9d3f03b3c2fb0b051efad.
//
// Solidity: event ValidatorRegistered(address indexed node, uint8 indexed validatorId, address indexed validator)
func (_ChannelHub *ChannelHubFilterer) ParseValidatorRegistered(log types.Log) (*ChannelHubValidatorRegistered, error) {
	event := new(ChannelHubValidatorRegistered)
	if err := _ChannelHub.contract.UnpackLog(event, "ValidatorRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubWithdrawnIterator is returned from FilterWithdrawn and is used to iterate over the raw logs and unpacked data for Withdrawn events raised by the ChannelHub contract.
type ChannelHubWithdrawnIterator struct {
	Event *ChannelHubWithdrawn // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubWithdrawn)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubWithdrawn)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubWithdrawn represents a Withdrawn event raised by the ChannelHub contract.
type ChannelHubWithdrawn struct {
	Wallet common.Address
	Token  common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterWithdrawn is a free log retrieval operation binding the contract event 0xd1c19fbcd4551a5edfb66d43d2e337c04837afda3482b42bdf569a8fccdae5fb.
//
// Solidity: event Withdrawn(address indexed wallet, address indexed token, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) FilterWithdrawn(opts *bind.FilterOpts, wallet []common.Address, token []common.Address) (*ChannelHubWithdrawnIterator, error) {

	var walletRule []interface{}
	for _, walletItem := range wallet {
		walletRule = append(walletRule, walletItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "Withdrawn", walletRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubWithdrawnIterator{contract: _ChannelHub.contract, event: "Withdrawn", logs: logs, sub: sub}, nil
}

// WatchWithdrawn is a free log subscription operation binding the contract event 0xd1c19fbcd4551a5edfb66d43d2e337c04837afda3482b42bdf569a8fccdae5fb.
//
// Solidity: event Withdrawn(address indexed wallet, address indexed token, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) WatchWithdrawn(opts *bind.WatchOpts, sink chan<- *ChannelHubWithdrawn, wallet []common.Address, token []common.Address) (event.Subscription, error) {

	var walletRule []interface{}
	for _, walletItem := range wallet {
		walletRule = append(walletRule, walletItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "Withdrawn", walletRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubWithdrawn)
				if err := _ChannelHub.contract.UnpackLog(event, "Withdrawn", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWithdrawn is a log parse operation binding the contract event 0xd1c19fbcd4551a5edfb66d43d2e337c04837afda3482b42bdf569a8fccdae5fb.
//
// Solidity: event Withdrawn(address indexed wallet, address indexed token, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) ParseWithdrawn(log types.Log) (*ChannelHubWithdrawn, error) {
	event := new(ChannelHubWithdrawn)
	if err := _ChannelHub.contract.UnpackLog(event, "Withdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
