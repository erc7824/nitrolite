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
	ChallengeDuration uint32
	User              common.Address
	Node              common.Address
	Nonce             uint64
	Metadata          [32]byte
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
	Version      uint64
	Intent       uint8
	Metadata     [32]byte
	HomeState    Ledger
	NonHomeState Ledger
	UserSig      []byte
	NodeSig      []byte
}

// ChannelHubMetaData contains all meta data concerning the ChannelHub contract.
var ChannelHubMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"ESCROW_DEPOSIT_UNLOCK_DELAY\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MAX_DEPOSIT_ESCROW_PURGE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MIN_CHALLENGE_DURATION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"challengeChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"proof\",\"type\":\"tuple[]\",\"internalType\":\"structState[]\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengerSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"challengeEscrowDeposit\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"challengerSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"challengeEscrowWithdrawal\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"challengerSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"checkpointChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"proof\",\"type\":\"tuple[]\",\"internalType\":\"structState[]\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"closeChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"proof\",\"type\":\"tuple[]\",\"internalType\":\"structState[]\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"createChannel\",\"inputs\":[{\"name\":\"def\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"initCCS\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"depositToChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"depositToVault\",\"inputs\":[{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"escrowHead\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"finalizeEscrowDeposit\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"finalizeEscrowWithdrawal\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"finalizeMigration\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getAccountBalance\",\"inputs\":[{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getChannelData\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumChannelStatus\"},{\"name\":\"definition\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"lastState\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengeExpiry\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"lockedFunds\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getChannelIds\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEscrowDepositData\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumEscrowStatus\"},{\"name\":\"unlockAt\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"challengeExpiry\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"lockedAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"initState\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEscrowDepositIds\",\"inputs\":[{\"name\":\"page\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"pageSize\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"ids\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEscrowWithdrawalData\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumEscrowStatus\"},{\"name\":\"challengeExpiry\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"lockedAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"initState\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOpenChannels\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getUnlockableEscrowDepositAmount\",\"inputs\":[],\"outputs\":[{\"name\":\"totalUnlockable\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getUnlockableEscrowDepositCount\",\"inputs\":[],\"outputs\":[{\"name\":\"count\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initiateEscrowDeposit\",\"inputs\":[{\"name\":\"def\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"initiateEscrowWithdrawal\",\"inputs\":[{\"name\":\"def\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"initiateMigration\",\"inputs\":[{\"name\":\"def\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"purgeEscrowDeposits\",\"inputs\":[{\"name\":\"maxToPurge\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"withdrawFromChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"withdrawFromVault\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"ChannelChallenged\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengeExpireAt\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelCheckpointed\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelClosed\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"finalState\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelCreated\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"user\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"definition\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"initialState\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelDeposited\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelWithdrawn\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Deposited\",\"inputs\":[{\"name\":\"wallet\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositChallenged\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengeExpireAt\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositFinalized\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositFinalizedOnHome\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositInitiated\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositInitiatedOnHome\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositsPurged\",\"inputs\":[{\"name\":\"purgedCount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalChallenged\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengeExpireAt\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalFinalized\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalFinalizedOnHome\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalInitiated\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalInitiatedOnHome\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MigrationInFinalized\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MigrationInInitiated\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MigrationOutFinalized\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MigrationOutInitiated\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Withdrawn\",\"inputs\":[{\"name\":\"wallet\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AddressCollision\",\"inputs\":[{\"name\":\"collision\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ChannelDoesNotExist\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureLength\",\"inputs\":[{\"name\":\"length\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureS\",\"inputs\":[{\"name\":\"s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"IncorrectChallengeDuration\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAddress\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAmount\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidValue\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"SafeCastOverflowedIntToUint\",\"inputs\":[{\"name\":\"value\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"type\":\"error\",\"name\":\"SafeERC20FailedOperation\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}]}]",
	Bin: "0x6080806040523460395760017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f005561639e908161003e8239f35b5f80fdfe60806040526004361015610011575f80fd5b5f3560e01c80630f00bcbb146120dd57806312d5c0dd1461203f57806313c380ed1461202857806317536c0614611f75578063187576d814611f035780632835312914611d095780633115f63014611b805780634adf728d14611a3e57806351c7a75f146117f857806353269198146115af578063587675e8146115505780635a0745b4146115345780635b9acbf9146115065780636898234b146113585780636af820bd1461133d5780637e7985f91461132657806382d3e15d1461130957806394191051146112ec578063a5c6225114611188578063b88c12e614610eeb578063c30159d514610b5e578063c74a2d1014610a80578063cd68b37a14610a6a578063d888ccae1461091d578063dd73d4941461082b578063e045e8d1146106a9578063e617208c1461051f578063e8265af714610477578063ecf3d7e8146103055763f4ac51f514610163575f80fd5b61016c36612338565b6020810135600b81101561030157600361018691146131d6565b815f525f60205260405f206002600182019101916001600160a01b038354165f52600660205260405f206001600160a01b036101c460808401612952565b165f5260205261021260e0826101de60405f20548861407d565b60405193849283927fa8b4483c00000000000000000000000000000000000000000000000000000000845260048401612d78565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49384156102f6577f6085f5128b19e0d3cc37524413de47259383f0f75265d5d66f41778696206696946102ba946102a6935f926102bf575b506102906102959293610276368861289b565b908a6001600160a01b0380865460201c1692541692613e26565b612d8f565b61029f368561289b565b90876140d7565b604051918291602083526020830190612b59565b0390a2005b61029592506102e86102909160e03d60e0116102ef575b6102e08183612706565b810190612988565b9250610263565b503d6102d6565b6040513d5f823e3d90fd5b5f80fd5b34610301576103133661239b565b91906001600160a01b0382161561044f57821561042757335f52600660205260405f206001600160a01b0382165f5260205260405f20548381106103e35783826001600160a01b03946103698361038f9561327f565b335f52600660205260405f208784165f5260205260405f205561038a61605c565b6142b9565b60017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f005560405192835216907fd1c19fbcd4551a5edfb66d43d2e337c04837afda3482b42bdf569a8fccdae5fb60203392a3005b606460405162461bcd60e51b815260206004820152601460248201527f696e73756666696369656e742062616c616e63650000000000000000000000006044820152fd5b7f2c5211c6000000000000000000000000000000000000000000000000000000005f5260045ffd5b7fe6c4247b000000000000000000000000000000000000000000000000000000005f5260045ffd5b34610301575f600319360112610301575f60035490600454915b8083106104a4575b602082604051908152f35b906104ae83612bfd565b90549060031b1c5f52600260205260405f204267ffffffffffffffff600283015460a01c16111580610506575b156104ff576104f89160046104f2920154906131c9565b92612c57565b9190610491565b5090610499565b50600160ff8183015416610519816124f5565b146104db565b34610301576020600319360112610301575f608060405161053f816126ce565b828152826020820152826040820152826060820152015261055e613d66565b506004355f525f60205260405f2060405190610579826126ce565b61058760ff82541683613daa565b61059360018201612d8f565b90602083019182526105a760048201613826565b6040840190815267ffffffffffffffff60136012840154936060870194855201541693608081019485525192600684101561067c576106059461065967ffffffffffffffff61066c935194519251169451936040519788809861265a565b60208701906080809163ffffffff81511684526001600160a01b0360208201511660208501526001600160a01b03604082015116604085015267ffffffffffffffff60608201511660608501520151910152565b61012060c0860152610120850190612531565b9160e08401526101008301520390f35b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602160045260245ffd5b34610301576107326106ba366124c5565b825f9492939452600560205260405f206106dd6106d78254613f1c565b1561367e565b6002810160a06106f76001600160a01b0383541688614e67565b604051809681927f24063eba0000000000000000000000000000000000000000000000000000000083526020600484015260248301906133a4565b038173__$b69fb814c294bfc16f92e50d7aeced4bde$__5af49384156102f6575f946107fa575b508154928260018594015460081c6001600160a01b03168093546001600160a01b03169460048693019861078c8a613826565b9436906107989261284a565b906107a29461557e565b836107ac86613826565b6107b69488614ed2565b6060015167ffffffffffffffff166040519182916107d491836138c9565b037fb8568a1f475f3c76759a620e08a653d28348c5c09e2e0bc91d533339801fefd891a2005b61081d91945060a03d60a011610824575b6108158183612706565b810190613345565b9286610759565b503d61080b565b3461030157602060031936011261030157610844613d66565b506004355f52600560205260405f2060405161085f816126b2565b815481526109196001830154916001600160a01b0360ff8416936020830194610887816124f5565b855260081c16604082015267ffffffffffffffff6002850154936001600160a01b038516606084015281608084019560a01c16855260c06108d6600460038901549860a08701998a5201613826565b930192835251936108e6856124f5565b5116935190519060405194846108fc87966124f5565b855260208501526040840152608060608401526080830190612531565b0390f35b3461030157602060031936011261030157610936613d66565b506004355f52600260205260405f20604051610100810181811067ffffffffffffffff821117610a3d57604052815481526109196001830154926001600160a01b0360ff851694602085019561098b816124f5565b865260081c1660408401526002810154926001600160a01b038416606082015267ffffffffffffffff608082019460a01c16845267ffffffffffffffff80806003850154169660a0840197885260e06109f2600560048801549760c0880198895201613826565b94019384525195610a02876124f5565b511695511691519051916040519585610a1b88976124f5565b865260208601526040850152606084015260a0608084015260a0830190612531565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b610a7e610a763661243f565b505090613a68565b005b610a8936612338565b6020810135600b811015610301576004610aa391146131d6565b815f525f60205260405f206002600182019101916001600160a01b038354165f52600660205260405f206001600160a01b03610ae160808401612952565b165f52602052610afb60e0826101de60405f20548861407d565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49384156102f6577f188e0ade7d115cc397426774adb960ae3e8c83e72f0a6cad4b7085e1d60bf986946102ba946102a6935f926102bf57506102906102959293610276368861289b565b60806003193601126103015760043560243567ffffffffffffffff811161030157806004019061026060031982360301126103015760443567ffffffffffffffff811161030157610bb390369060040161240e565b505060643567ffffffffffffffff811161030157610bd5903690600401612497565b9190845f525f60205260405f209260ff845416600681101561067c576001610bfd9114613a1d565b610c0960048501613826565b90610c1386613221565b67ffffffffffffffff80845116911610610e81578660018601936001600160a01b03855460201c16956001600160a01b036002890154169467ffffffffffffffff80610c5e8c613221565b925116911611610d55575b509163ffffffff9591610c8c610c929594610c84368c61289b565b93369161284a565b9161557e565b600260ff19845416178355541667ffffffffffffffff4216019167ffffffffffffffff8311610d28577f07b9206d5a6026d3bd2a8f9a9b79f6fa4bfbd6a016975829fbaf07488019f28a926013610d1c930167ffffffffffffffff821667ffffffffffffffff1982541617905567ffffffffffffffff604051938493604085526040850190612b59565b911660208301520390a2005b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b9591509291602486013595600b87101561030157610d76610de19715612667565b835f5260066020526001600160a01b03610d96608460405f209301612952565b165f5260205260e088610dad60405f20548c61407d565b60405198899283927fa8b4483c00000000000000000000000000000000000000000000000000000000845260048401612d78565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af480156102f657610c8c610c9295610e4e8b8d9488888763ffffffff9e5f94610e5a575b50610e2d610e329495369061289b565b613e26565b8c610e47610e3f8c612d8f565b91369061289b565b9086615683565b92949550509195610c69565b610e329450610e7a610e2d9160e03d60e0116102ef576102e08183612706565b9450610e1d565b608460405162461bcd60e51b815260206004820152603560248201527f6368616c6c656e67652063616e646964617465206d757374206861766520686960448201527f67686572206f7220657175616c2076657273696f6e00000000000000000000006064820152fd5b3461030157610ef9366122f8565b6020810135600b811015610301576007610f139114612667565b610f25610f20368461274f565b613db6565b91610f30368361289b565b91610f5760208301936040610f4486612952565b94019386610f5186612952565b92613e26565b610f69610f6382613221565b85614e3d565b92610f7385613f1c565b1561109e57506110029150835f525f60205260405f20906001600160a01b036002830154165f52600660205260405f206001600160a01b03610fb760808401612952565b165f5260205260e081610fce60405f20548861407d565b60405195869283927fa8b4483c00000000000000000000000000000000000000000000000000000000845260048401612d78565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49283156102f6577f587faad1bcd589ce902468251883e1976a645af8563c773eed7356d78433210c93611071936102a6925f92611076575b5060016110609101612d8f565b61106a368561289b565b90886140d7565b0390a3005b61106091925061109660019160e03d60e0116102ef576102e08183612706565b929150611053565b906110ea60a0826110b76110b187612952565b88614e67565b60405193849283927eea54e7000000000000000000000000000000000000000000000000000000008452600484016133fe565b038173__$b69fb814c294bfc16f92e50d7aeced4bde$__5af480156102f6577f17eb0a6bd5a0de45d1029ce3444941070e149df35b22176fc439f930f73c09f794611071946102a6935f9361115f575b5061114761114d91612952565b91612952565b91611158368661289b565b8989614ed2565b61114d9193506111806111479160a03d60a011610824576108158183612706565b93915061113a565b3461030157611196366124c5565b90825f52600260205260405f20906111b16106d78354613f1c565b6111fa60c06111bf86614392565b604051809381927f6666e4c0000000000000000000000000000000000000000000000000000000008352602060048401526024830190612cee565b038173__$682d6198b4eca5bc7e038b912a26498e7e$__5af49384156102f6577fba075bd445233f7cad862c72f0343b3503aad9c8e704a2295f122b82abf8e8019467ffffffffffffffff936080935f926112b6575b506002826112a3939461129389549160058b019a61126d8c613826565b84610c8c6001600160a01b0380600186015460081c16998a95015416998a95369161284a565b61129c89613826565b908b614450565b015116906102ba604051928392836138c9565b6112a392506112de60029160c03d60c0116112e5575b6112d68183612706565b810190612c7e565b9250611250565b503d6112cc565b34610301575f600319360112610301576020604051620151808152f35b34610301575f600319360112610301576020600454604051908152f35b3461030157610a7e61133736612338565b90613415565b34610301575f60031936011261030157602060405160408152f35b34610301576020600319360112610301576001600160a01b03611379612371565b165f52600160205260405f20604051808260208294549384815201905f5260205f20925f5b8181106114ed5750506113b392500382612706565b5f5f5b8251811015611438576113c9818461328c565b515f525f60205260ff60405f205416600681101561067c57600314158061140c575b6113f8575b6001016113b6565b90611404600191612c57565b9190506113f0565b50611417818461328c565b515f525f60205260ff60405f205416600681101561067c57600514156113eb565b506114429061324e565b905f915f5b82518110156114df5761145a818461328c565b515f525f60205260ff60405f205416600681101561067c5760031415806114b3575b611489575b600101611447565b926114ab60019161149a868661328c565b516114a5828661328c565b52612c57565b939050611481565b506114be818461328c565b515f525f60205260ff60405f205416600681101561067c576005141561147c565b6040518061091984826123d5565b845483526001948501948694506020909301920161139e565b34610301576040600319360112610301576109196115286024356004356132a0565b604051918291826123d5565b34610301575f600319360112610301576020604051612a308152f35b3461030157604060031936011261030157611569612371565b602435906001600160a01b0382168203610301576001600160a01b03165f5260066020526001600160a01b0360405f2091165f52602052602060405f2054604051908152f35b34610301576115bd36612338565b6020810135600b81101561030157600a6115d79114612667565b815f525f60205260405f206116b76001820160026001600160a01b03825460201c1693016116156001600160a01b038254168588610e2d368a61289b565b6001600160a01b03611627368761289b565b9161014087019561163787613221565b67ffffffffffffffff164614968761178e575b505054165f52600660205260405f206001600160a01b038060206060850151015116165f5260205260e08161168360405f20548961407d565b60405195869283927fa8b4483c00000000000000000000000000000000000000000000000000000000845260048401612a58565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49182156102f6576116f2935f93611769575b506116ec90612d8f565b866140d7565b15611730576102ba7f9a6f675cc94b83b55f1ecc0876affd4332a30c92e6faa2aca0199b1b6df922c391604051918291602083526020830190612b59565b6102ba7f7b20773c41402791c5f18914dbbeacad38b1ebcc4c55d8eb3bfe0a4cde26c82691604051918291602083526020830190612b59565b6116ec9193506117879060e03d60e0116102ef576102e08183612706565b92906116e2565b6117999036906127bf565b60608501526117ab3660608a016127bf565b60808501526040516117be602082612706565b5f815260a08501526040516117d4602082612706565b5f815260c08501525f5260016020526117f08860405f20616108565b50888061164a565b611801366122f8565b906020820135600b81101561030157600561181c9114612667565b611829610f20368361274f565b91611834368261289b565b6118546020840191604061184784612952565b95019486610f5187612952565b611860610f6383613221565b9261186a85613f1c565b1561190b5750506118ae90835f525f60205260405f20906001600160a01b036002830154165f52600660205260405f206001600160a01b03610fb760808401612952565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49283156102f6577f471c4ebe4e57d25ef7117e141caac31c6b98f067b8098a7a7bbd38f637c2f98093611071936102a6925f92611076575060016110609101612d8f565b906119509160c08461191c87614392565b60405195869283927fbbc42f3400000000000000000000000000000000000000000000000000000000845260048401612d53565b038173__$682d6198b4eca5bc7e038b912a26498e7e$__5af49182156102f65761199a935f93611a15575b5061114761198891612952565b91611993368661289b565b8787614450565b60035468010000000000000000811015610a3d577fede7867afa7cdb9c443667efd8244d98bf9df1dce68e60dc94dca6605125ca7691836119ff6119e984600161107196016003556003612c42565b81939154905f199060031b92831b921b19161790565b9055604051918291602083526020830190612b59565b611988919350611a366111479160c03d60c0116112e5576112d68183612706565b93915061197b565b611a473661243f565b50506020810135600b81101561030157611b1657815f525f60205260405f206002600182019101916001600160a01b038354165f52600660205260405f206001600160a01b03611a9960808401612952565b165f52602052611ab360e0826101de60405f20548861407d565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49384156102f6577f567044ba1cdd4671ac3979c114241e1e3b56c9e9051f63f2f234f7a2795019cc946102ba946102a6935f926102bf57506102906102959293610276368861289b565b608460405162461bcd60e51b815260206004820152602260248201527f63616e206f6e6c7920636865636b706f696e74206f706572617465207374617460448201527f65730000000000000000000000000000000000000000000000000000000000006064820152fd5b3461030157602060031936011261030157600354600480549190355f5b82841080611d00575b15611cd357611bb484612bfd565b90549060031b1c5f52600260205260405f206001810160ff815416611bd8816124f5565b60038114611cc0576002830154904267ffffffffffffffff8360a01c1611159081611cac575b5015611c745782611c6b94925f926004611c659601926001600160a01b0384549216855260066020526001600160a01b0380600c6040882093015460401c16168552602052611c52604085209182546131c9565b9055600360ff1982541617905555612c57565b93612c57565b915b9192611b9d565b5050509050602091507f61815f4b11c6ea4e14a2e448a010bed8efdc3e53a15efbf183d16a31085cd145925b600455604051908152a1005b60019150611cb9816124f5565b1488611bfe565b50505092611ccd90612c57565b91611c6d565b9050602091507f61815f4b11c6ea4e14a2e448a010bed8efdc3e53a15efbf183d16a31085cd14592611ca0565b50818110611ba6565b611d12366122f8565b6020810135600b811015610301576001611d2c91146131d6565b611d39610f20368461274f565b611d4283613f70565b604083016001600160a01b03611d5782612952565b165f52600660205260405f206001600160a01b03611d7760808601612952565b165f52602052611d9160e0846101de60405f20548661407d565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af480156102f6577fae3d48960dc29080438681a58800ea8520315e5fb998f450c039b9269201864f92611e86925f92611ed3575b506001600160a01b0380611ece92611e2f611df8368b61289b565b95611e1460208d0197611e0a89612952565b8c610f5187612952565b611e1e368d61274f565b611e28368d61289b565b908b6140d7565b81611e3986612952565b165f526001602052611e4e8860405f20616003565b506080611e5a86612952565b9a83611e91611e6885612952565b94826040519b8c9b63ffffffff611e7e88612729565b168d52612387565b1660208b0152612387565b16604088015267ffffffffffffffff611eac6060830161273a565b1660608801520135608086015260c060a08601521697169560c0830190612b59565b0390a4005b611ece9192506001600160a01b03611efa819260e03d60e0116102ef576102e08183612706565b93925050611ddd565b34610301576020600319360112610301576001600160a01b03611f24612371565b165f52600160205260405f206040519081602082549182815201915f5260205f20905f5b818110611f5f576109198561152881870382612706565b8254845260209093019260019283019201611f48565b6001600160a01b03611f863661239b565b9290911690811561044f5782156104275760206001600160a01b037f8752a472e571a816aea92eec8dae9baf628e840f4929fbcc2d155e6233ff68a792845f526006835260405f208282165f52835260405f20611fe48782546131c9565b9055611fee61605c565b611ff9868233614be5565b60017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f00556040519586521693a3005b3461030157610a7e61203936612338565b90612deb565b34610301575f600319360112610301575f60035490600454915b80831061206b57602082604051908152f35b9061207583612bfd565b90549060031b1c5f52600260205260405f204267ffffffffffffffff600283015460a01c16111590816120c2575b50156120bc576104f26120b591612c57565b9190612059565b90610499565b600180925060ff910154166120d6816124f5565b14846120a3565b34610301576120eb366122f8565b6020810135600b8110156103015760096121059114612667565b612112610f20368461274f565b916121a5612120368461289b565b91602081019261214261213285612952565b91604084019288610f5185612952565b6001600160a01b0361216f612157368861289b565b9261216189613f1c565b968715612278575b50612952565b165f52600660205260405f206001600160a01b038060206060850151015116165f5260205260e08161168360405f20548961407d565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49182156102f6576121dc935f93612253575b506116ec90369061274f565b1561221a576102ba7f3142fb397e715d80415dff7b527bf1c451def4675da6e1199ee1b4588e3f630a91604051918291602083526020830190612b59565b6102ba7f26afbcb9eb52c21f42eb9cfe8f263718ffb65afbf84abe8ad8cce2acfb2242b891604051918291602083526020830190612b59565b6116ec9193506122719060e03d60e0116102ef576102e08183612706565b92906121d0565b6122dc849161228688613f70565b612294366101408d016127bf565b60608801526122a63660608d016127bf565b60808801526040516122b9602082612706565b5f815260a08801526040516122cf602082612706565b5f815260c0880152612952565b165f5260016020526122f18960405f20616003565b5089612169565b90600319820160c081126103015760a0136103015760049160a4359067ffffffffffffffff82116103015760031982610260920301126103015760040190565b90604060031983011261030157600435916024359067ffffffffffffffff82116103015760031982610260920301126103015760040190565b600435906001600160a01b038216820361030157565b35906001600160a01b038216820361030157565b6003196060910112610301576004356001600160a01b038116810361030157906024356001600160a01b0381168103610301579060443590565b60206040818301928281528451809452019201905f5b8181106123f85750505090565b82518452602093840193909201916001016123eb565b9181601f840112156103015782359167ffffffffffffffff8311610301576020808501948460051b01011161030157565b6060600319820112610301576004359160243567ffffffffffffffff811161030157610260600319828503011261030157600401916044359067ffffffffffffffff8211610301576124939160040161240e565b9091565b9181601f840112156103015782359167ffffffffffffffff8311610301576020838186019501011161030157565b90604060031983011261030157600435916024359067ffffffffffffffff82116103015761249391600401612497565b6004111561067c57565b90600b82101561067c5752565b90601f19601f602080948051918291828752018686015e5f8582860101520116010190565b6126579167ffffffffffffffff8251168152612555602083015160208301906124ff565b604082015160408201526125c36060830151606083019060c0809167ffffffffffffffff81511684526001600160a01b03602082015116602085015260ff6040820151166040850152606081015160608501526080810151608085015260a081015160a08501520151910152565b608082810151805167ffffffffffffffff1661014084015260208101516001600160a01b0316610160840152604081015160ff1661018084015260608101516101a0840152908101516101c083015260a08101516101e083015260c0015161020082015260c061264560a084015161026061022085015261026084019061250c565b9201519061024081840391015261250c565b90565b90600682101561067c5752565b1561266e57565b606460405162461bcd60e51b815260206004820152600e60248201527f696e76616c696420696e74656e740000000000000000000000000000000000006044820152fd5b60e0810190811067ffffffffffffffff821117610a3d57604052565b60a0810190811067ffffffffffffffff821117610a3d57604052565b60c0810190811067ffffffffffffffff821117610a3d57604052565b90601f601f19910116810190811067ffffffffffffffff821117610a3d57604052565b359063ffffffff8216820361030157565b359067ffffffffffffffff8216820361030157565b91908260a091031261030157604051612767816126ce565b608080829461277581612729565b845261278360208201612387565b602085015261279460408201612387565b60408501526127a56060820161273a565b60608501520135910152565b359060ff8216820361030157565b91908260e0910312610301576040516127d7816126b2565b60c08082946127e58161273a565b84526127f360208201612387565b6020850152612804604082016127b1565b6040850152606081013560608501526080810135608085015260a081013560a08501520135910152565b67ffffffffffffffff8111610a3d57601f01601f191660200190565b9291926128568261282e565b916128646040519384612706565b829481845281830111610301578281602093845f960137010152565b9080601f83011215610301578160206126579335910161284a565b91906102608382031261030157604051906128b5826126b2565b81936128c08161273a565b83526020810135600b811015610301576020840152604081013560408401526128ec82606083016127bf565b60608401526128ff8261014083016127bf565b608084015261022081013567ffffffffffffffff81116103015782612925918301612880565b60a08401526102408101359167ffffffffffffffff83116103015760c09261294d9201612880565b910152565b356001600160a01b03811681036103015790565b519067ffffffffffffffff8216820361030157565b5190811515820361030157565b908160e091031261030157604051906129a0826126b2565b80518252602081015160208301526040810151600681101561030157612a019160c09160408501526129d460608201612966565b60608501526129e56080820161297b565b60808501526129f660a0820161297b565b60a08501520161297b565b60c082015290565b90612a1581835161265a565b608067ffffffffffffffff81612a3a602086015160a0602087015260a0860190612531565b94604081015160408601526060810151606086015201511691015290565b9091612a6f61265793604084526040840190612a09565b916020818403910152612531565b60c0809167ffffffffffffffff612a938261273a565b1684526001600160a01b03612aaa60208301612387565b16602085015260ff612abe604083016127b1565b166040850152606081013560608501526080810135608085015260a081013560a08501520135910152565b90357fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe18236030181121561030157016020813591019167ffffffffffffffff821161030157813603831361030157565b601f8260209493601f1993818652868601375f8582860101520116010190565b67ffffffffffffffff612b6b8261273a565b168252602081013591600b83101561030157612b8e6126579360208301906124ff565b60408201356040820152612ba86060820160608401612a7d565b612bba61014082016101408401612a7d565b612bee612be2612bce610220850185612ae9565b610260610220860152610260850191612b39565b92610240810190612ae9565b91610240818503910152612b39565b600354811015612c155760035f5260205f2001905f90565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603260045260245ffd5b8054821015612c15575f5260205f2001905f90565b5f198114610d285760010190565b90612c6f816124f5565b60ff60ff198354169116179055565b908160c09103126103015760405190612c96826126ea565b80518252602081015160208301526040810151600481101561030157612ce69160a0916040850152612cca60608201612966565b6060850152612cdb60808201612966565b60808501520161297b565b60a082015290565b908151612cfa816124f5565b815260a080612d18602085015160c0602086015260c0850190612531565b936040810151604085015267ffffffffffffffff606082015116606085015267ffffffffffffffff6080820151166080850152015191015290565b9091612d6a61265793604084526040840190612cee565b916020818403910152612b59565b9091612d6a61265793604084526040840190612a09565b90604051612d9c816126ce565b6080600282946001600160a01b03815463ffffffff8116865260201c16602085015267ffffffffffffffff60018201546001600160a01b038116604087015260a01c1660608501520154910152565b805f52600260205260405f2060018101908154916001600160a01b038360081c16926001600160a01b0360028401541691612e268454613f1c565b91821560ff8216816131b5575b508061319e575b6130f55750506020860135600b811015610301576006612e5a9114612667565b612e718285612e69368a61289b565b865490613e26565b15612f7b5750612ecf9150805490815f525f60205260e085610fce60405f20946001600160a01b036002870154165f52600660205260405f206001600160a01b03612ebe60808601612952565b165f5260205260405f20549061407d565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49283156102f6577f32e24720f56fd5a7f4cb219d7ff3278ae95196e79c85b5801395894a6f53466c93612f5593612f3f925f92612f5a575b50612f2f600185549201612d8f565b612f39368a61289b565b916140d7565b5493604051918291602083526020830190612b59565b0390a3565b612f7491925060e03d60e0116102ef576102e08183612706565b905f612f20565b90815f52600660205260405f206001600160a01b03612f9d6101608801612952565b165f5260205261307160c08660405f205460405190612fbb826126ea565b5f82528860208301612fcb613d66565b815267ffffffffffffffff600360408601925f8452606087015f815260808801945f865260a08901965f88525f52600260205260405f209260ff600185015416613014816124f5565b8a5261302260058501613826565b90526004830154905283600283015460a01c16905201541690525260405193849283927fbbc42f3400000000000000000000000000000000000000000000000000000000845260048401612d53565b038173__$682d6198b4eca5bc7e038b912a26498e7e$__5af49283156102f6577f1b92e8ef67d8a7c0d29c99efcd180a5e0d98d60ac41d52abbbb5950882c78e4e94612f5594612f3f935f916130d6575b5084546130cf368b61289b565b9089614450565b6130ef915060c03d60c0116112e5576112d68183612706565b5f6130c2565b7f1b92e8ef67d8a7c0d29c99efcd180a5e0d98d60ac41d52abbbb5950882c78e4e9550613164925090600360ff19612f559695931617905560048301905f825492556003840167ffffffffffffffff1981541690556001600160a01b03600c85015460401c169061038a61605c565b60017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f00555493604051918291602083526020830190612b59565b5067ffffffffffffffff6003860154164211612e3a565b600291506131c2816124f5565b145f612e33565b91908201809211610d2857565b156131dd57565b606460405162461bcd60e51b815260206004820152601460248201527f696e76616c696420737461746520696e74656e740000000000000000000000006044820152fd5b3567ffffffffffffffff811681036103015790565b67ffffffffffffffff8111610a3d5760051b60200190565b9061325882613236565b6132656040519182612706565b828152601f196132758294613236565b0190602036910137565b91908203918211610d2857565b8051821015612c155760209160051b010190565b91906003549080840293808504821490151715610d285781841015613329576132c990846131c9565b90808211613321575b506132e56132e0848361327f565b61324e565b92805b8281106132f457505050565b80613300600192612bfd565b90549060031b1c61331a613314858461327f565b8861328c565b52016132e8565b90505f6132d2565b5050905060405161333b602082612706565b5f81525f36813790565b908160a0910312610301576040519061335d826126ce565b8051825260208101516020830152604081015160048110156103015761339c91608091604085015261339160608201612966565b60608501520161297b565b608082015290565b9081516133b0816124f5565b815260806001600160a01b03816133d6602086015160a0602087015260a0860190612531565b946040810151604086015267ffffffffffffffff606082015116606086015201511691015290565b9091612d6a612657936040845260408401906133a4565b805f52600560205260405f20805492600182018054906001600160a01b038260081c1693600281018054926001600160a01b038416946134548a613f1c565b9485159060ff831682613668575b5081613651575b506135c957505050506020830135600b81101561030157600861348c9114612667565b61349c828588610e2d368861289b565b1561353d57506134e09150835f525f60205260405f20906001600160a01b036002830154165f52600660205260405f206001600160a01b03610fb760808401612952565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49283156102f6577f6d0cf3d243d63f08f50db493a8af34b27d4e3bc9ec4098e82700abfeffe2d49893612f55936102a6925f92611076575060016110609101612d8f565b9061354e60a0826110b78588614e67565b038173__$b69fb814c294bfc16f92e50d7aeced4bde$__5af49283156102f6577f2fdac1380dbe23ae259b6871582b7f33e34461547f400bdd20d74991250317d194612f55946102a6935f916135aa575b50611158368661289b565b6135c3915060a03d60a011610824576108158183612706565b5f61359f565b612f559699507f2fdac1380dbe23ae259b6871582b7f33e34461547f400bdd20d74991250317d19750613164945060ff196003915f97949597501617905560038401915f835493557fffffffff0000000000000000ffffffffffffffffffffffffffffffffffffffff81541690556001600160a01b03600b85015460401c169061038a61605c565b67ffffffffffffffff915060a01c1642115f613469565b6002919250613676816124f5565b14905f613462565b1561368557565b608460405162461bcd60e51b815260206004820152602760248201527f6f6e6c79206e6f6e2d686f6d6520657363726f77732063616e2062652063686160448201527f6c6c656e676564000000000000000000000000000000000000000000000000006064820152fd5b906040516136fc816126b2565b60c06004829460ff815467ffffffffffffffff811686526001600160a01b038160401c16602087015260e01c1660408501526001810154606085015260028101546080850152600381015460a08501520154910152565b90600182811c9216801561379a575b602083101461376d57565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b91607f1691613762565b5f92918154916137b383613753565b808352926001811690811561380857506001146137cf57505050565b5f9081526020812093945091925b8383106137ee575060209250010190565b6001816020929493945483858701015201910191906137dd565b9050602094955060ff1991509291921683830152151560051b010190565b90604051613833816126b2565b809260ff815467ffffffffffffffff8116845260401c1691600b83101561067c576138c560c092600d94602084015260018101546040840152613878600282016136ef565b6060840152613889600782016136ef565b60808401526040516138a9816138a281600c86016137a4565b0382612706565b60a08401526138be60405180968193016137a4565b0384612706565b0152565b9067ffffffffffffffff613a16602092959495604085526138fd8154848116604088015260ff606088019160401c166124ff565b6001810154608086015261396860a0860160028301600460c09160ff815467ffffffffffffffff811686526001600160a01b038160401c16602087015260e01c1660408501526001810154606085015260028101546080850152600381015460a08501520154910152565b600781015467ffffffffffffffff8116610180870152604081901c6001600160a01b03166101a087015260e01c60ff166101c086015260088101546101e08601526009810154610200860152600a810154610220860152600b81015461024086015261026080860152600d6139e46102a08701600c84016137a4565b917fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc087840301610280880152016137a4565b9416910152565b15613a2457565b606460405162461bcd60e51b815260206004820152601660248201527f696e76616c6964206368616e6e656c20737461747573000000000000000000006044820152fd5b906020810135600b811015610301576002613a8391146131d6565b815f525f60205260405f2060ff8154169160068310158061067c57600184148015613d23575b613ab290613a1d565b613abe60048401613826565b92600181019360028201916001600160a01b03835416966001600160a01b03875460201c169461067c5760021480613d0c575b613c005750506001600160a01b039054165f52600660205260405f206001600160a01b03613b2160808501612952565b165f52602052613b3b60e0836101de60405f20548961407d565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49485156102f6577f04cd8c68bf83e7bc531ca5a5d75c34e36513c2acf81e07e6470ba79e29da13a895613bd495613bab935f92613bd9575b50610290613ba19293868b610e2d368b61289b565b61106a368661289b565b5f526001602052613bbf8460405f20616108565b50604051918291602083526020830190612b59565b0390a2565b613ba19250613bf96102909160e03d60e0116102ef576102e08183612706565b9250613b8c565b613bd495507f04cd8c68bf83e7bc531ca5a5d75c34e36513c2acf81e07e6470ba79e29da13a896919450613cbf9250601390600360ff198254161781555f60128201550167ffffffffffffffff19815416905560608401613c7e815160606001600160a01b0360208301511691015190613c7861605c565b866142b9565b60017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f00555160a06001600160a01b036020830151169101519161038a61605c565b60017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f00555f526001602052613cf78460405f20616108565b50604051918291602083526020830190612531565b5067ffffffffffffffff6013820154164211613af1565b505f905060028414613aa9565b60405190613d3d826126b2565b5f60c0838281528260208201528260408201528260608201528260808201528260a08201520152565b60405190613d73826126b2565b606060c0835f81525f60208201525f6040820152613d8f613d30565b83820152613d9b613d30565b60808201528260a08201520152565b600682101561067c5752565b604051613e116020820180936080809163ffffffff81511684526001600160a01b0360208201511660208501526001600160a01b03604082015116604085015267ffffffffffffffff60608201511660608501520151910152565b60a08152613e2060c082612706565b51902090565b6001600160a01b03613e789493613e6f829360c0613e4f613e4a613e679884615cfd565b615e1e565b91613e5e60a0820151846161e5565b9099919961621f565b0151906161e5565b9097919761621f565b16911603613ed8576001600160a01b03809116911603613e9457565b606460405162461bcd60e51b815260206004820152601660248201527f696e76616c6964206e6f6465207369676e6174757265000000000000000000006044820152fd5b606460405162461bcd60e51b815260206004820152601660248201527f696e76616c69642075736572207369676e6174757265000000000000000000006044820152fd5b805f525f60205260ff60405f205416600681101561067c578015908115613f65575b50613f60575f525f60205267ffffffffffffffff600660405f20015416461490565b505f90565b60059150145f613f3e565b602081016001600160a01b03613f8582612952565b161561044f57604082016001600160a01b03613fa082612952565b161561044f57613fcf906001600160a01b0380613fc5613fbf86612952565b93612952565b1691161491612952565b9061401757503563ffffffff8116809103610301576201518011613fef57565b7f0596b15b000000000000000000000000000000000000000000000000000000005f5260045ffd5b6001600160a01b03907fabfa558d000000000000000000000000000000000000000000000000000000005f521660045260245ffd5b60405190614059826126ce565b5f608083828152614068613d66565b60208201528260408201528260608201520152565b90601367ffffffffffffffff9161409261404c565b935f525f60205260405f20906140ac60ff83541686613daa565b6140b860048301613826565b6020860152601282015460408601526060850152015416608082015290565b90929192815f525f60205260405f209360ff85541691600683101561067c57614106938593156141b957615683565b604081018051600681101561067c57151580614197575b614176575b508060a060c0920151614153575b01516141395750565b5f6012820155601301805467ffffffffffffffff19169055565b600160ff198454161783556013830167ffffffffffffffff198154169055614130565b5190600682101561067c5760c09160ff60ff19855416911617835590614122565b5060ff835416815190600682101561067c57600681101561067c57141561411d565b6001870163ffffffff8351168154907fffffffffffffffff00000000000000000000000000000000000000000000000077ffffffffffffffffffffffffffffffffffffffff00000000602087015160201b169216171790556142aa600288016001600160a01b0380604086015116167fffffffffffffffffffffffff000000000000000000000000000000000000000082541617815567ffffffffffffffff6060850151167fffffffff0000000000000000ffffffffffffffffffffffffffffffffffffffff7bffffffffffffffff000000000000000000000000000000000000000083549260a01b169116179055565b60808201516003880155615683565b90821561438d576001600160a01b0316806142e757505f8080936001600160a01b0382941682f1156102f657565b916001600160a01b03604051927fa9059cbb000000000000000000000000000000000000000000000000000000005f521660045260245260205f60448180865af19060015f511482161561436c575b604052156143415750565b7f5274afe7000000000000000000000000000000000000000000000000000000005f5260045260245ffd5b90600181151661438457823b15153d15161690614336565b503d5f823e3d90fd5b505050565b5f604051916143a0836126ea565b818352602083016143af613d66565b815267ffffffffffffffff6003604086019285845260608701868152608088019487865260a089019688885288526002602052604088209260ff6001850154166143f8816124f5565b8a5261440660058501613826565b90526004830154905283600283015460a01c16905201541690525290565b7f80000000000000000000000000000000000000000000000000000000000000008114610d28575f0390565b9594939290955f52600260205260405f2091604082018051614471816124f5565b61447a816124f5565b614bc8575b5060a082018051614715575b602093949596975067ffffffffffffffff606084015116806146ba575b5067ffffffffffffffff60808401511680614695575b50511561467c578260806001600160a01b0392015101511680945b8251905f8213156146155761450391506144f384516160d3565b9283916144fe61605c565b614be5565b60017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0055614536600485019182546131c9565b90555b0180515f8113156145a1575090614594926001600160a01b0361455e600494516160d3565b95165f5260066020526001600160a01b0360405f2091165f5260205260405f2061458985825461327f565b9055019182546131c9565b90555b61459f614cdf565b565b9190505f82126145b5575b50505050614597565b61460a926001600160a01b036145d46145cf600495614424565b6160d3565b95165f5260066020526001600160a01b0360405f2091165f5260205260405f206145ff8582546131c9565b90550191825461327f565b90555f8080806145ac565b5f8212614625575b505050614539565b6146346145cf61463f93614424565b92839161038a61605c565b60017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f00556146726004850191825461327f565b9055835f8061461d565b506001600160a01b03600c84015460401c1680946144d9565b67ffffffffffffffff60038701911667ffffffffffffffff198254161790555f6144be565b61470f9060028701907fffffffff0000000000000000ffffffffffffffffffffffffffffffffffffffff7bffffffffffffffff000000000000000000000000000000000000000083549260a01b169116179055565b5f6144a8565b6005840167ffffffffffffffff808451161667ffffffffffffffff198254161781556020830151600b81101561067c577fffffffffffffffffffffffffffffffffffffffffffffff00ffffffffffffffff68ff000000000000000083549260401b1691161790556040820151600685015560c06007850160608401519067ffffffffffffffff808351161667ffffffffffffffff19825416178155602082015181547fffffff000000000000000000000000000000000000000000ffffffffffffffff7bffffffffffffffffffffffffffffffffffffffff00000000000000007cff00000000000000000000000000000000000000000000000000000000604087015160e01b169360401b16911617179055606081015160088701556080810151600987015560a0810151600a8701550151600b85015560c0600c850160808401519067ffffffffffffffff808351161667ffffffffffffffff19825416178155602082015181547fffffff000000000000000000000000000000000000000000ffffffffffffffff7bffffffffffffffffffffffffffffffffffffffff00000000000000007cff00000000000000000000000000000000000000000000000000000000604087015160e01b169360401b169116171790556060810151600d8701556080810151600e87015560a0810151600f870155015160108501556011840160a083015180519067ffffffffffffffff8211610a3d5781906149398454613753565b601f8111614b78575b50602090601f8311600114614b17575f92614b0c575b50505f198260011b9260031b1c19161790555b601284019760c083015198895167ffffffffffffffff8111610a3d576149918254613753565b601f8111614ac7575b506020601f8211600114614a5f578190602098999a9b9c5f92614a54575b50505f198260011b9260031b1c19161790555b85556001850180547fffffffffffffffffffffff0000000000000000000000000000000000000000ff16600888901b74ffffffffffffffffffffffffffffffffffffffff0016179055600285016001600160a01b0388167fffffffffffffffffffffffff000000000000000000000000000000000000000082541617905587969594935061448b565b015190505f806149b8565b601f1982169b835f52815f209c5f5b818110614aaf5750916020999a9b9c9d91846001959410614a97575b505050811b0190556149cb565b01515f1960f88460031b161c191690555f8080614a8a565b838301518f556001909e019d60209384019301614a6e565b825f5260205f20601f830160051c81019160208410614b02575b601f0160051c01905b818110614af7575061499a565b5f8155600101614aea565b9091508190614ae1565b015190505f80614958565b5f85815282812093601f1916905b818110614b605750908460019594939210614b48575b505050811b01905561496b565b01515f1960f88460031b161c191690555f8080614b3b565b92936020600181928786015181550195019301614b25565b909150835f5260205f20601f840160051c81019160208510614bbe575b90601f859493920160051c01905b818110614bb05750614942565b5f8155849350600101614ba3565b9091508190614b95565b614bdf9051614bd6816124f5565b60018501612c65565b5f61447f565b90821561438d576001600160a01b03169182158015614cb157813403614c89575b15614c1057505050565b6001600160a01b03604051927f23b872dd000000000000000000000000000000000000000000000000000000005f52166004523060245260445260205f60648180865af19060015f5114821615614c71575b6040525f606052156143415750565b90600181151661438457823b15153d15161690614c62565b7faa7feadc000000000000000000000000000000000000000000000000000000005f5260045ffd5b3415614c06577faa7feadc000000000000000000000000000000000000000000000000000000005f5260045ffd5b6003546004545f5b82821080614e33575b15614e0857614cfe82612bfd565b90549060031b1c5f52600260205260405f206001810160ff815416614d22816124f5565b60038114614df5576002830154904267ffffffffffffffff8360a01c1611159081614de1575b5015614dab5782614da294925f926004614d9c9601926001600160a01b0384549216855260066020526001600160a01b0380600c6040882093015460401c16168552602052611c52604085209182546131c9565b91612c57565b915b9190614ce7565b5050507f61815f4b11c6ea4e14a2e448a010bed8efdc3e53a15efbf183d16a31085cd14592506020915b600455604051908152a1565b60019150614dee816124f5565b145f614d48565b50505090614e0290612c57565b91614da4565b7f61815f4b11c6ea4e14a2e448a010bed8efdc3e53a15efbf183d16a31085cd1459250602091614dd5565b5060408110614cf0565b9067ffffffffffffffff604051916020830193845216604082015260408152613e20606082612706565b906001600160a01b0390614e7961404c565b925f52600560205267ffffffffffffffff600260405f2060ff600182015416614ea1816124f5565b8652614eaf60048201613826565b602087015260038101546040870152015460a01c16606084015216608082015290565b9594939290955f52600560205260405f2091604082018051614ef3816124f5565b614efc816124f5565b61556a575b506080820180516150b7575b602093949596975067ffffffffffffffff6060840151168061505c575b505115615043578260806001600160a01b0392015101511680945b8251905f821315614fe757614f5f91506144f384516160d3565b60017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0055614f92600385019182546131c9565b90555b0180515f811315614fba575090614594926001600160a01b0361455e600394516160d3565b9190505f8212614fcd5750505050614597565b61460a926001600160a01b036145d46145cf600395614424565b5f8212614ff7575b505050614f95565b6146346145cf61500693614424565b60017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f00556150396003850191825461327f565b9055835f80614fef565b506001600160a01b03600b84015460401c168094614f45565b6150b19060028701907fffffffff0000000000000000ffffffffffffffffffffffffffffffffffffffff7bffffffffffffffff000000000000000000000000000000000000000083549260a01b169116179055565b5f614f2a565b6004840167ffffffffffffffff808451161667ffffffffffffffff198254161781556020830151600b81101561067c577fffffffffffffffffffffffffffffffffffffffffffffff00ffffffffffffffff68ff000000000000000083549260401b1691161790556040820151600585015560c06006850160608401519067ffffffffffffffff808351161667ffffffffffffffff19825416178155602082015181547fffffff000000000000000000000000000000000000000000ffffffffffffffff7bffffffffffffffffffffffffffffffffffffffff00000000000000007cff00000000000000000000000000000000000000000000000000000000604087015160e01b169360401b16911617179055606081015160078701556080810151600887015560a081015160098701550151600a85015560c0600b850160808401519067ffffffffffffffff808351161667ffffffffffffffff19825416178155602082015181547fffffff000000000000000000000000000000000000000000ffffffffffffffff7bffffffffffffffffffffffffffffffffffffffff00000000000000007cff00000000000000000000000000000000000000000000000000000000604087015160e01b169360401b169116171790556060810151600c8701556080810151600d87015560a0810151600e8701550151600f8501556010840160a083015180519067ffffffffffffffff8211610a3d5781906152db8454613753565b601f811161551a575b50602090601f83116001146154b9575f926154ae575b50505f198260011b9260031b1c19161790555b601184019760c083015198895167ffffffffffffffff8111610a3d576153338254613753565b601f8111615469575b506020601f8211600114615401578190602098999a9b9c5f926153f6575b50505f198260011b9260031b1c19161790555b85556001850180547fffffffffffffffffffffff0000000000000000000000000000000000000000ff16600888901b74ffffffffffffffffffffffffffffffffffffffff0016179055600285016001600160a01b0388167fffffffffffffffffffffffff0000000000000000000000000000000000000000825416179055879695949350614f0d565b015190505f8061535a565b601f1982169b835f52815f209c5f5b8181106154515750916020999a9b9c9d91846001959410615439575b505050811b01905561536d565b01515f1960f88460031b161c191690555f808061542c565b838301518f556001909e019d60209384019301615410565b825f5260205f20601f830160051c810191602084106154a4575b601f0160051c01905b818110615499575061533c565b5f815560010161548c565b9091508190615483565b015190505f806152fa565b5f85815282812093601f1916905b81811061550257509084600195949392106154ea575b505050811b01905561530d565b01515f1960f88460031b161c191690555f80806154dd565b929360206001819287860151815501950193016154c7565b909150835f5260205f20601f840160051c81019160208510615560575b90601f859493920160051c01905b81811061555257506152e4565b5f8155849350600101615545565b9091508190615537565b6155789051614bd6816124f5565b5f614f01565b61560d6001600160a01b03936156086020613e4a6009826155a38a9961561699615cfd565b6040519481869251918291018484015e81017f6368616c6c656e676500000000000000000000000000000000000000000000008382015203017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe9810184520182612706565b6161e5565b9092919261621f565b1691168114918215615670575b50501561562c57565b606460405162461bcd60e51b815260206004820152601f60248201527f6368616c6c656e676572206d757374206265206e6f6465206f722075736572006044820152fd5b6001600160a01b03161490505f80615623565b9291925f525f60205260405f209260808301516158cf575b6060019160c06001600160a01b03602085510151169180515f81135f14615862575080516156da81856001600160a01b036020890151166144fe61605c565b60017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f005561570d601288019182546131c9565b90555b6020810180515f8113156157f95750516001600160a01b036040860151165f52600660205260405f206001600160a01b0385165f5260205260405f2061575782825461327f565b9055615768601288019182546131c9565b90555b01511515806157eb575b615786575b5050505061459f614cdf565b6157e0926157af60a0926001600160a01b0360406012960151169084845101519161038a61605c565b60017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0055510151920191825461327f565b90555f80808061577a565b5060a0835101511515615775565b90505f8112615809575b5061576b565b61581290614424565b6001600160a01b036040860151165f52600660205260405f206001600160a01b0385165f5260205260405f206158498282546131c9565b905561585a6012880191825461327f565b90555f615803565b5f8112615870575b50615710565b61587990614424565b61589481856001600160a01b0360208901511661038a61605c565b60017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f00556158c76012880191825461327f565b90555f61586a565b6004840167ffffffffffffffff808351161667ffffffffffffffff198254161781556020820151600b81101561067c577fffffffffffffffffffffffffffffffffffffffffffffff00ffffffffffffffff68ff000000000000000083549260401b1691161790556040810151600585015560c06006850160608301519067ffffffffffffffff808351161667ffffffffffffffff19825416178155602082015181547fffffff000000000000000000000000000000000000000000ffffffffffffffff7bffffffffffffffffffffffffffffffffffffffff00000000000000007cff00000000000000000000000000000000000000000000000000000000604087015160e01b169360401b16911617179055606081015160078701556080810151600887015560a081015160098701550151600a85015560c0600b850160808301519067ffffffffffffffff808351161667ffffffffffffffff19825416178155602082015181547fffffff000000000000000000000000000000000000000000ffffffffffffffff7bffffffffffffffffffffffffffffffffffffffff00000000000000007cff00000000000000000000000000000000000000000000000000000000604087015160e01b169360401b169116171790556060810151600c8701556080810151600d87015560a0810151600e8701550151600f8501556010840160a082015180519067ffffffffffffffff8211610a3d578190615af38454613753565b601f8111615cad575b50602090601f8311600114615c4c575f92615c41575b50505f198260011b9260031b1c19161790555b6011840160c082015180519067ffffffffffffffff8211610a3d57615b4a8354613753565b601f8111615bfc575b50602090601f8311600114615b95576060949392915f9183615b8a575b50505f198260011b9260031b1c19161790555b905061569b565b015190505f80615b70565b90601f19831691845f52815f20925f5b818110615be4575091600193918560609897969410615bcc575b505050811b019055615b83565b01515f1960f88460031b161c191690555f8080615bbf565b92936020600181928786015181550195019301615ba5565b835f5260205f20601f840160051c81019160208510615c37575b601f0160051c01905b818110615c2c5750615b53565b5f8155600101615c1f565b9091508190615c16565b015190505f80615b12565b5f85815282812093601f1916905b818110615c955750908460019594939210615c7d575b505050811b019055615b25565b01515f1960f88460031b161c191690555f8080615c70565b92936020600181928786015181550195019301615c5a565b909150835f5260205f20601f840160051c81019160208510615cf3575b90601f859493920160051c01905b818110615ce55750615afc565b5f8155849350600101615cd8565b9091508190615cca565b67ffffffffffffffff8151166020820151600b81101561067c5782615dac91615d4b6040615e0d9601519160806060850151940151956040519860208a0152604089015260608801906124ff565b608086015260a085019060c0809167ffffffffffffffff81511684526001600160a01b03602082015116602085015260ff6040820151166040850152606081015160608501526080810151608085015260a081015160a08501520151910152565b805167ffffffffffffffff1661018084015260208101516001600160a01b03166101a0840152604081015160ff166101c084015260608101516101e0840152608081015161020084015260a081015161022084015260c00151610240830152565b610240815261265761026082612706565b8051905f827a184f03e93ff9f4daa797ed6e38ed64bf6a1f010000000000000000811015615fdb575b806d04ee2d6d415b85acef8100000000600a921015615fc0575b662386f26fc10000811015615fac575b6305f5e100811015615f9b575b612710811015615f8c575b6064811015615f7e575b1015615f76575b6001810192600a5f196021615ec7615eb18861282e565b97615ebf604051998a612706565b80895261282e565b94601f196020890196013687378701015b01917f30313233343536373839616263646566000000000000000000000000000000008282061a835304908115615f1457600a905f1990615ed8565b5050613e2090603a6020604051948593828501977f19457468657265756d205369676e6564204d6573736167653a0a00000000000089525180918587015e8401908382015f8152815193849201905e01015f815203601f198101835282612706565b600101615e9a565b606460029104920191615e93565b61271060049104920191615e89565b6305f5e10060089104920191615e7e565b662386f26fc1000060109104920191615e71565b6d04ee2d6d415b85acef810000000060209104920191615e61565b50604090507a184f03e93ff9f4daa797ed6e38ed64bf6a1f0100000000000000008304615e47565b6001810190825f528160205260405f2054155f1461605557805468010000000000000000811015610a3d576160426119e9826001879401855584612c42565b905554915f5260205260405f2055600190565b5050505f90565b60027f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0054146160ab5760027f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0055565b7f3ee5aeb5000000000000000000000000000000000000000000000000000000005f5260045ffd5b5f81126160dd5790565b7fa8ce4432000000000000000000000000000000000000000000000000000000005f5260045260245ffd5b906001820191815f528260205260405f20548015155f146161dd575f198101818111610d28578254905f198201918211610d28578181036161a8575b5050508054801561617b575f19019061615d8282612c42565b5f1982549160031b1b19169055555f526020525f6040812055600190565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603160045260245ffd5b6161c86161b86119e99386612c42565b90549060031b1c92839286612c42565b90555f528360205260405f20555f8080616144565b505050505f90565b81519190604183036162155761620e9250602082015190606060408401519301515f1a906162e6565b9192909190565b50505f9160029190565b616228816124f5565b80616231575050565b61623a816124f5565b6001810361626a577ff645eedf000000000000000000000000000000000000000000000000000000005f5260045ffd5b616273816124f5565b600281036162a757507ffce698f7000000000000000000000000000000000000000000000000000000005f5260045260245ffd5b6003906162b3816124f5565b146162bb5750565b7fd78bce0c000000000000000000000000000000000000000000000000000000005f5260045260245ffd5b91907f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0841161635d579160209360809260ff5f9560405194855216868401526040830152606082015282805260015afa156102f6575f516001600160a01b0381161561635357905f905f90565b505f906001905f90565b5050505f916003919056fea26469706673582212209006b4d3d1640372d8aa906105d0ed57e3886635e70f2d2d8b5c29fb5083e26764736f6c634300081e0033",
}

// ChannelHubABI is the input ABI used to generate the binding from.
// Deprecated: Use ChannelHubMetaData.ABI instead.
var ChannelHubABI = ChannelHubMetaData.ABI

// ChannelHubBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ChannelHubMetaData.Bin instead.
var ChannelHubBin = ChannelHubMetaData.Bin

// DeployChannelHub deploys a new Ethereum contract, binding an instance of ChannelHub to it.
func DeployChannelHub(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ChannelHub, error) {
	parsed, err := ChannelHubMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ChannelHubBin), backend)
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
// Solidity: function getChannelData(bytes32 channelId) view returns(uint8 status, (uint32,address,address,uint64,bytes32) definition, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) lastState, uint256 challengeExpiry, uint256 lockedFunds)
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
// Solidity: function getChannelData(bytes32 channelId) view returns(uint8 status, (uint32,address,address,uint64,bytes32) definition, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) lastState, uint256 challengeExpiry, uint256 lockedFunds)
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
// Solidity: function getChannelData(bytes32 channelId) view returns(uint8 status, (uint32,address,address,uint64,bytes32) definition, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) lastState, uint256 challengeExpiry, uint256 lockedFunds)
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
// Solidity: function getEscrowDepositData(bytes32 escrowId) view returns(uint8 status, uint64 unlockAt, uint64 challengeExpiry, uint256 lockedAmount, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState)
func (_ChannelHub *ChannelHubCaller) GetEscrowDepositData(opts *bind.CallOpts, escrowId [32]byte) (struct {
	Status          uint8
	UnlockAt        uint64
	ChallengeExpiry uint64
	LockedAmount    *big.Int
	InitState       State
}, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getEscrowDepositData", escrowId)

	outstruct := new(struct {
		Status          uint8
		UnlockAt        uint64
		ChallengeExpiry uint64
		LockedAmount    *big.Int
		InitState       State
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Status = *abi.ConvertType(out[0], new(uint8)).(*uint8)
	outstruct.UnlockAt = *abi.ConvertType(out[1], new(uint64)).(*uint64)
	outstruct.ChallengeExpiry = *abi.ConvertType(out[2], new(uint64)).(*uint64)
	outstruct.LockedAmount = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.InitState = *abi.ConvertType(out[4], new(State)).(*State)

	return *outstruct, err

}

// GetEscrowDepositData is a free data retrieval call binding the contract method 0xd888ccae.
//
// Solidity: function getEscrowDepositData(bytes32 escrowId) view returns(uint8 status, uint64 unlockAt, uint64 challengeExpiry, uint256 lockedAmount, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState)
func (_ChannelHub *ChannelHubSession) GetEscrowDepositData(escrowId [32]byte) (struct {
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
// Solidity: function getEscrowDepositData(bytes32 escrowId) view returns(uint8 status, uint64 unlockAt, uint64 challengeExpiry, uint256 lockedAmount, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState)
func (_ChannelHub *ChannelHubCallerSession) GetEscrowDepositData(escrowId [32]byte) (struct {
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
// Solidity: function getEscrowWithdrawalData(bytes32 escrowId) view returns(uint8 status, uint64 challengeExpiry, uint256 lockedAmount, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState)
func (_ChannelHub *ChannelHubCaller) GetEscrowWithdrawalData(opts *bind.CallOpts, escrowId [32]byte) (struct {
	Status          uint8
	ChallengeExpiry uint64
	LockedAmount    *big.Int
	InitState       State
}, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getEscrowWithdrawalData", escrowId)

	outstruct := new(struct {
		Status          uint8
		ChallengeExpiry uint64
		LockedAmount    *big.Int
		InitState       State
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Status = *abi.ConvertType(out[0], new(uint8)).(*uint8)
	outstruct.ChallengeExpiry = *abi.ConvertType(out[1], new(uint64)).(*uint64)
	outstruct.LockedAmount = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.InitState = *abi.ConvertType(out[3], new(State)).(*State)

	return *outstruct, err

}

// GetEscrowWithdrawalData is a free data retrieval call binding the contract method 0xdd73d494.
//
// Solidity: function getEscrowWithdrawalData(bytes32 escrowId) view returns(uint8 status, uint64 challengeExpiry, uint256 lockedAmount, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState)
func (_ChannelHub *ChannelHubSession) GetEscrowWithdrawalData(escrowId [32]byte) (struct {
	Status          uint8
	ChallengeExpiry uint64
	LockedAmount    *big.Int
	InitState       State
}, error) {
	return _ChannelHub.Contract.GetEscrowWithdrawalData(&_ChannelHub.CallOpts, escrowId)
}

// GetEscrowWithdrawalData is a free data retrieval call binding the contract method 0xdd73d494.
//
// Solidity: function getEscrowWithdrawalData(bytes32 escrowId) view returns(uint8 status, uint64 challengeExpiry, uint256 lockedAmount, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState)
func (_ChannelHub *ChannelHubCallerSession) GetEscrowWithdrawalData(escrowId [32]byte) (struct {
	Status          uint8
	ChallengeExpiry uint64
	LockedAmount    *big.Int
	InitState       State
}, error) {
	return _ChannelHub.Contract.GetEscrowWithdrawalData(&_ChannelHub.CallOpts, escrowId)
}

// GetOpenChannels is a free data retrieval call binding the contract method 0x6898234b.
//
// Solidity: function getOpenChannels(address user) view returns(bytes32[])
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
// Solidity: function getOpenChannels(address user) view returns(bytes32[])
func (_ChannelHub *ChannelHubSession) GetOpenChannels(user common.Address) ([][32]byte, error) {
	return _ChannelHub.Contract.GetOpenChannels(&_ChannelHub.CallOpts, user)
}

// GetOpenChannels is a free data retrieval call binding the contract method 0x6898234b.
//
// Solidity: function getOpenChannels(address user) view returns(bytes32[])
func (_ChannelHub *ChannelHubCallerSession) GetOpenChannels(user common.Address) ([][32]byte, error) {
	return _ChannelHub.Contract.GetOpenChannels(&_ChannelHub.CallOpts, user)
}

// GetUnlockableEscrowDepositAmount is a free data retrieval call binding the contract method 0xe8265af7.
//
// Solidity: function getUnlockableEscrowDepositAmount() view returns(uint256 totalUnlockable)
func (_ChannelHub *ChannelHubCaller) GetUnlockableEscrowDepositAmount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getUnlockableEscrowDepositAmount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetUnlockableEscrowDepositAmount is a free data retrieval call binding the contract method 0xe8265af7.
//
// Solidity: function getUnlockableEscrowDepositAmount() view returns(uint256 totalUnlockable)
func (_ChannelHub *ChannelHubSession) GetUnlockableEscrowDepositAmount() (*big.Int, error) {
	return _ChannelHub.Contract.GetUnlockableEscrowDepositAmount(&_ChannelHub.CallOpts)
}

// GetUnlockableEscrowDepositAmount is a free data retrieval call binding the contract method 0xe8265af7.
//
// Solidity: function getUnlockableEscrowDepositAmount() view returns(uint256 totalUnlockable)
func (_ChannelHub *ChannelHubCallerSession) GetUnlockableEscrowDepositAmount() (*big.Int, error) {
	return _ChannelHub.Contract.GetUnlockableEscrowDepositAmount(&_ChannelHub.CallOpts)
}

// GetUnlockableEscrowDepositCount is a free data retrieval call binding the contract method 0x12d5c0dd.
//
// Solidity: function getUnlockableEscrowDepositCount() view returns(uint256 count)
func (_ChannelHub *ChannelHubCaller) GetUnlockableEscrowDepositCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getUnlockableEscrowDepositCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetUnlockableEscrowDepositCount is a free data retrieval call binding the contract method 0x12d5c0dd.
//
// Solidity: function getUnlockableEscrowDepositCount() view returns(uint256 count)
func (_ChannelHub *ChannelHubSession) GetUnlockableEscrowDepositCount() (*big.Int, error) {
	return _ChannelHub.Contract.GetUnlockableEscrowDepositCount(&_ChannelHub.CallOpts)
}

// GetUnlockableEscrowDepositCount is a free data retrieval call binding the contract method 0x12d5c0dd.
//
// Solidity: function getUnlockableEscrowDepositCount() view returns(uint256 count)
func (_ChannelHub *ChannelHubCallerSession) GetUnlockableEscrowDepositCount() (*big.Int, error) {
	return _ChannelHub.Contract.GetUnlockableEscrowDepositCount(&_ChannelHub.CallOpts)
}

// ChallengeChannel is a paid mutator transaction binding the contract method 0xc30159d5.
//
// Solidity: function challengeChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes)[] proof, bytes challengerSig) payable returns()
func (_ChannelHub *ChannelHubTransactor) ChallengeChannel(opts *bind.TransactOpts, channelId [32]byte, candidate State, proof []State, challengerSig []byte) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "challengeChannel", channelId, candidate, proof, challengerSig)
}

// ChallengeChannel is a paid mutator transaction binding the contract method 0xc30159d5.
//
// Solidity: function challengeChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes)[] proof, bytes challengerSig) payable returns()
func (_ChannelHub *ChannelHubSession) ChallengeChannel(channelId [32]byte, candidate State, proof []State, challengerSig []byte) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChallengeChannel(&_ChannelHub.TransactOpts, channelId, candidate, proof, challengerSig)
}

// ChallengeChannel is a paid mutator transaction binding the contract method 0xc30159d5.
//
// Solidity: function challengeChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes)[] proof, bytes challengerSig) payable returns()
func (_ChannelHub *ChannelHubTransactorSession) ChallengeChannel(channelId [32]byte, candidate State, proof []State, challengerSig []byte) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChallengeChannel(&_ChannelHub.TransactOpts, channelId, candidate, proof, challengerSig)
}

// ChallengeEscrowDeposit is a paid mutator transaction binding the contract method 0xa5c62251.
//
// Solidity: function challengeEscrowDeposit(bytes32 escrowId, bytes challengerSig) returns()
func (_ChannelHub *ChannelHubTransactor) ChallengeEscrowDeposit(opts *bind.TransactOpts, escrowId [32]byte, challengerSig []byte) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "challengeEscrowDeposit", escrowId, challengerSig)
}

// ChallengeEscrowDeposit is a paid mutator transaction binding the contract method 0xa5c62251.
//
// Solidity: function challengeEscrowDeposit(bytes32 escrowId, bytes challengerSig) returns()
func (_ChannelHub *ChannelHubSession) ChallengeEscrowDeposit(escrowId [32]byte, challengerSig []byte) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChallengeEscrowDeposit(&_ChannelHub.TransactOpts, escrowId, challengerSig)
}

// ChallengeEscrowDeposit is a paid mutator transaction binding the contract method 0xa5c62251.
//
// Solidity: function challengeEscrowDeposit(bytes32 escrowId, bytes challengerSig) returns()
func (_ChannelHub *ChannelHubTransactorSession) ChallengeEscrowDeposit(escrowId [32]byte, challengerSig []byte) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChallengeEscrowDeposit(&_ChannelHub.TransactOpts, escrowId, challengerSig)
}

// ChallengeEscrowWithdrawal is a paid mutator transaction binding the contract method 0xe045e8d1.
//
// Solidity: function challengeEscrowWithdrawal(bytes32 escrowId, bytes challengerSig) returns()
func (_ChannelHub *ChannelHubTransactor) ChallengeEscrowWithdrawal(opts *bind.TransactOpts, escrowId [32]byte, challengerSig []byte) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "challengeEscrowWithdrawal", escrowId, challengerSig)
}

// ChallengeEscrowWithdrawal is a paid mutator transaction binding the contract method 0xe045e8d1.
//
// Solidity: function challengeEscrowWithdrawal(bytes32 escrowId, bytes challengerSig) returns()
func (_ChannelHub *ChannelHubSession) ChallengeEscrowWithdrawal(escrowId [32]byte, challengerSig []byte) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChallengeEscrowWithdrawal(&_ChannelHub.TransactOpts, escrowId, challengerSig)
}

// ChallengeEscrowWithdrawal is a paid mutator transaction binding the contract method 0xe045e8d1.
//
// Solidity: function challengeEscrowWithdrawal(bytes32 escrowId, bytes challengerSig) returns()
func (_ChannelHub *ChannelHubTransactorSession) ChallengeEscrowWithdrawal(escrowId [32]byte, challengerSig []byte) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChallengeEscrowWithdrawal(&_ChannelHub.TransactOpts, escrowId, challengerSig)
}

// CheckpointChannel is a paid mutator transaction binding the contract method 0x4adf728d.
//
// Solidity: function checkpointChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes)[] proof) payable returns()
func (_ChannelHub *ChannelHubTransactor) CheckpointChannel(opts *bind.TransactOpts, channelId [32]byte, candidate State, proof []State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "checkpointChannel", channelId, candidate, proof)
}

// CheckpointChannel is a paid mutator transaction binding the contract method 0x4adf728d.
//
// Solidity: function checkpointChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes)[] proof) payable returns()
func (_ChannelHub *ChannelHubSession) CheckpointChannel(channelId [32]byte, candidate State, proof []State) (*types.Transaction, error) {
	return _ChannelHub.Contract.CheckpointChannel(&_ChannelHub.TransactOpts, channelId, candidate, proof)
}

// CheckpointChannel is a paid mutator transaction binding the contract method 0x4adf728d.
//
// Solidity: function checkpointChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes)[] proof) payable returns()
func (_ChannelHub *ChannelHubTransactorSession) CheckpointChannel(channelId [32]byte, candidate State, proof []State) (*types.Transaction, error) {
	return _ChannelHub.Contract.CheckpointChannel(&_ChannelHub.TransactOpts, channelId, candidate, proof)
}

// CloseChannel is a paid mutator transaction binding the contract method 0xcd68b37a.
//
// Solidity: function closeChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes)[] proof) payable returns()
func (_ChannelHub *ChannelHubTransactor) CloseChannel(opts *bind.TransactOpts, channelId [32]byte, candidate State, proof []State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "closeChannel", channelId, candidate, proof)
}

// CloseChannel is a paid mutator transaction binding the contract method 0xcd68b37a.
//
// Solidity: function closeChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes)[] proof) payable returns()
func (_ChannelHub *ChannelHubSession) CloseChannel(channelId [32]byte, candidate State, proof []State) (*types.Transaction, error) {
	return _ChannelHub.Contract.CloseChannel(&_ChannelHub.TransactOpts, channelId, candidate, proof)
}

// CloseChannel is a paid mutator transaction binding the contract method 0xcd68b37a.
//
// Solidity: function closeChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes)[] proof) payable returns()
func (_ChannelHub *ChannelHubTransactorSession) CloseChannel(channelId [32]byte, candidate State, proof []State) (*types.Transaction, error) {
	return _ChannelHub.Contract.CloseChannel(&_ChannelHub.TransactOpts, channelId, candidate, proof)
}

// CreateChannel is a paid mutator transaction binding the contract method 0x28353129.
//
// Solidity: function createChannel((uint32,address,address,uint64,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initCCS) payable returns()
func (_ChannelHub *ChannelHubTransactor) CreateChannel(opts *bind.TransactOpts, def ChannelDefinition, initCCS State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "createChannel", def, initCCS)
}

// CreateChannel is a paid mutator transaction binding the contract method 0x28353129.
//
// Solidity: function createChannel((uint32,address,address,uint64,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initCCS) payable returns()
func (_ChannelHub *ChannelHubSession) CreateChannel(def ChannelDefinition, initCCS State) (*types.Transaction, error) {
	return _ChannelHub.Contract.CreateChannel(&_ChannelHub.TransactOpts, def, initCCS)
}

// CreateChannel is a paid mutator transaction binding the contract method 0x28353129.
//
// Solidity: function createChannel((uint32,address,address,uint64,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initCCS) payable returns()
func (_ChannelHub *ChannelHubTransactorSession) CreateChannel(def ChannelDefinition, initCCS State) (*types.Transaction, error) {
	return _ChannelHub.Contract.CreateChannel(&_ChannelHub.TransactOpts, def, initCCS)
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

// InitiateEscrowDeposit is a paid mutator transaction binding the contract method 0x51c7a75f.
//
// Solidity: function initiateEscrowDeposit((uint32,address,address,uint64,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubTransactor) InitiateEscrowDeposit(opts *bind.TransactOpts, def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "initiateEscrowDeposit", def, candidate)
}

// InitiateEscrowDeposit is a paid mutator transaction binding the contract method 0x51c7a75f.
//
// Solidity: function initiateEscrowDeposit((uint32,address,address,uint64,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubSession) InitiateEscrowDeposit(def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.InitiateEscrowDeposit(&_ChannelHub.TransactOpts, def, candidate)
}

// InitiateEscrowDeposit is a paid mutator transaction binding the contract method 0x51c7a75f.
//
// Solidity: function initiateEscrowDeposit((uint32,address,address,uint64,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubTransactorSession) InitiateEscrowDeposit(def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.InitiateEscrowDeposit(&_ChannelHub.TransactOpts, def, candidate)
}

// InitiateEscrowWithdrawal is a paid mutator transaction binding the contract method 0xb88c12e6.
//
// Solidity: function initiateEscrowWithdrawal((uint32,address,address,uint64,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactor) InitiateEscrowWithdrawal(opts *bind.TransactOpts, def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "initiateEscrowWithdrawal", def, candidate)
}

// InitiateEscrowWithdrawal is a paid mutator transaction binding the contract method 0xb88c12e6.
//
// Solidity: function initiateEscrowWithdrawal((uint32,address,address,uint64,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubSession) InitiateEscrowWithdrawal(def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.InitiateEscrowWithdrawal(&_ChannelHub.TransactOpts, def, candidate)
}

// InitiateEscrowWithdrawal is a paid mutator transaction binding the contract method 0xb88c12e6.
//
// Solidity: function initiateEscrowWithdrawal((uint32,address,address,uint64,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactorSession) InitiateEscrowWithdrawal(def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.InitiateEscrowWithdrawal(&_ChannelHub.TransactOpts, def, candidate)
}

// InitiateMigration is a paid mutator transaction binding the contract method 0x0f00bcbb.
//
// Solidity: function initiateMigration((uint32,address,address,uint64,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactor) InitiateMigration(opts *bind.TransactOpts, def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "initiateMigration", def, candidate)
}

// InitiateMigration is a paid mutator transaction binding the contract method 0x0f00bcbb.
//
// Solidity: function initiateMigration((uint32,address,address,uint64,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubSession) InitiateMigration(def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.InitiateMigration(&_ChannelHub.TransactOpts, def, candidate)
}

// InitiateMigration is a paid mutator transaction binding the contract method 0x0f00bcbb.
//
// Solidity: function initiateMigration((uint32,address,address,uint64,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
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

// FilterChannelCreated is a free log retrieval operation binding the contract event 0xae3d48960dc29080438681a58800ea8520315e5fb998f450c039b9269201864f.
//
// Solidity: event ChannelCreated(bytes32 indexed channelId, address indexed user, address indexed node, (uint32,address,address,uint64,bytes32) definition, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initialState)
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

// WatchChannelCreated is a free log subscription operation binding the contract event 0xae3d48960dc29080438681a58800ea8520315e5fb998f450c039b9269201864f.
//
// Solidity: event ChannelCreated(bytes32 indexed channelId, address indexed user, address indexed node, (uint32,address,address,uint64,bytes32) definition, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initialState)
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

// ParseChannelCreated is a log parse operation binding the contract event 0xae3d48960dc29080438681a58800ea8520315e5fb998f450c039b9269201864f.
//
// Solidity: event ChannelCreated(bytes32 indexed channelId, address indexed user, address indexed node, (uint32,address,address,uint64,bytes32) definition, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initialState)
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
