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
	ABI: "[{\"type\":\"function\",\"name\":\"ESCROW_DEPOSIT_UNLOCK_DELAY\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MAX_DEPOSIT_ESCROW_PURGE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MIN_CHALLENGE_DURATION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"challengeChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"proof\",\"type\":\"tuple[]\",\"internalType\":\"structState[]\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengerSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"challengeEscrowDeposit\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"challengerSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"challengeEscrowWithdrawal\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"challengerSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"checkpointChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"proof\",\"type\":\"tuple[]\",\"internalType\":\"structState[]\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"closeChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"proof\",\"type\":\"tuple[]\",\"internalType\":\"structState[]\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"createChannel\",\"inputs\":[{\"name\":\"def\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"initState\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"depositToChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"depositToVault\",\"inputs\":[{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"escrowHead\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"finalizeEscrowDeposit\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"finalizeEscrowWithdrawal\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"finalizeMigration\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getAccountBalance\",\"inputs\":[{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getChannelData\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumChannelStatus\"},{\"name\":\"definition\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"lastState\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengeExpiry\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"lockedFunds\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getChannelIds\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEscrowDepositData\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumEscrowStatus\"},{\"name\":\"unlockAt\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"challengeExpiry\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"lockedAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"initState\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEscrowDepositIds\",\"inputs\":[{\"name\":\"page\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"pageSize\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"ids\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEscrowWithdrawalData\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumEscrowStatus\"},{\"name\":\"challengeExpiry\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"lockedAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"initState\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOpenChannels\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getUnlockableEscrowDepositAmount\",\"inputs\":[],\"outputs\":[{\"name\":\"totalUnlockable\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getUnlockableEscrowDepositCount\",\"inputs\":[],\"outputs\":[{\"name\":\"count\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initiateEscrowDeposit\",\"inputs\":[{\"name\":\"def\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"initiateEscrowWithdrawal\",\"inputs\":[{\"name\":\"def\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"initiateMigration\",\"inputs\":[{\"name\":\"def\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"purgeEscrowDeposits\",\"inputs\":[{\"name\":\"maxToPurge\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"withdrawFromChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"withdrawFromVault\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"ChannelChallenged\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengeExpireAt\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelCheckpointed\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelClosed\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"finalState\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelCreated\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"user\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"definition\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"initialState\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelDeposited\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelWithdrawn\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Deposited\",\"inputs\":[{\"name\":\"wallet\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositChallenged\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengeExpireAt\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositFinalized\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositFinalizedOnHome\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositInitiated\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositInitiatedOnHome\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositsPurged\",\"inputs\":[{\"name\":\"purgedCount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalChallenged\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengeExpireAt\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalFinalized\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalFinalizedOnHome\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalInitiated\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalInitiatedOnHome\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MigrationInFinalized\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MigrationInInitiated\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MigrationOutFinalized\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MigrationOutInitiated\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeState\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Withdrawn\",\"inputs\":[{\"name\":\"wallet\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AddressCollision\",\"inputs\":[{\"name\":\"collision\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ChannelDoesNotExist\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureLength\",\"inputs\":[{\"name\":\"length\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureS\",\"inputs\":[{\"name\":\"s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"IncorrectChallengeDuration\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAddress\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAmount\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidValue\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"SafeCastOverflowedIntToUint\",\"inputs\":[{\"name\":\"value\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"type\":\"error\",\"name\":\"SafeERC20FailedOperation\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}]}]",
	Bin: "0x6080806040523460395760017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0055615fe8908161003e8239f35b5f80fdfe60806040526004361015610011575f80fd5b5f3560e01c80630f00bcbb1461214c57806312d5c0dd146120af57806313c380ed1461209857806317536c0614611ff8578063187576d814611f865780632835312914611c385780633115f63014611ab05780634adf728d1461196e57806351c7a75f146117415780635326919814611512578063587675e8146114b35780635a0745b4146114975780635b9acbf9146114695780636898234b146112bb5780636af820bd146112a05780637e7985f91461128957806382d3e15d1461126c578063941910511461124f578063a5c62251146110ec578063b88c12e614610e68578063c30159d514610b17578063c74a2d1014610a39578063cd68b37a14610a23578063d888ccae146108e2578063dd73d494146107e1578063e045e8d114610660578063e617208c146104f2578063e8265af71461044b578063ecf3d7e8146102ec5763f4ac51f514610163575f80fd5b61016c366123a7565b6020810135600a8110156102e857600261018691146139b7565b815f525f60205260405f206002600182019101916001600160a01b038354165f52600660205260405f206001600160a01b036101c4608084016129af565b165f526020526101f960e0826101de60405f205488613fed565b6040519384928392632a2d120f60e21b845260048401612d96565b0381730827b6aaa03475a8bf59ee1a2bd76903ddfbadb65af49384156102dd577f6085f5128b19e0d3cc37524413de47259383f0f75265d5d66f41778696206696946102a19461028d935f926102a6575b5061027761027c929361025d36886128fa565b908a6001600160a01b0380865460201c1692541692613d97565b612dad565b61028636856128fa565b9087614046565b604051918291602083526020830190612b94565b0390a2005b61027c92506102cf6102779160e03d60e0116102d6575b6102c78183612769565b8101906129e4565b925061024a565b503d6102bd565b6040513d5f823e3d90fd5b5f80fd5b346102e8576102fa3661240a565b91906001600160a01b038216156104235782156103fb57335f52600660205260405f206001600160a01b0382165f5260205260405f20548381106103b75783826001600160a01b03946103508361037695613221565b335f52600660205260405f208784165f5260205260405f2055610371615cc4565b6141f6565b60015f516020615f935f395f51905f525560405192835216907fd1c19fbcd4551a5edfb66d43d2e337c04837afda3482b42bdf569a8fccdae5fb60203392a3005b606460405162461bcd60e51b815260206004820152601460248201527f696e73756666696369656e742062616c616e63650000000000000000000000006044820152fd5b7f2c5211c6000000000000000000000000000000000000000000000000000000005f5260045ffd5b7fe6c4247b000000000000000000000000000000000000000000000000000000005f5260045ffd5b346102e8575f3660031901126102e8575f60035490600454915b808310610478575b602082604051908152f35b9061048283612c37565b90549060031b1c5f52600260205260405f20426001600160401b03600283015460a01c161115806104d9575b156104d2576104cb9160046104c5920154906131b7565b92612c78565b9190610465565b509061046d565b50600160ff81830154166104ec8161255f565b146104ae565b346102e85760203660031901126102e8575f608060405161051281612733565b8281528260208201528260408201528260608201520152610531613cd8565b506004355f525f60205260405f206040519061054c82612733565b61055a60ff82541683613d1c565b61056660018201612dad565b906020830191825261057a60048201613797565b604084019081526001600160401b0360136012840154936060870194855201541693608081019485525192600684101561064c576105d6946106296001600160401b0361063c93519451925116945193604051978880986126c0565b60208701906080809163ffffffff81511684526001600160a01b0360208201511660208501526001600160a01b0360408201511660408501526001600160401b0360608201511660608501520151910152565b61012060c086015261012085019061259a565b9160e08401526101008301520390f35b634e487b7160e01b5f52602160045260245ffd5b346102e8576106e961067136612530565b825f9492939452600560205260405f2061069461068e8254613e8d565b1561360a565b6002810160a06106ae6001600160a01b0383541688614c85565b604051809681927f24063eba000000000000000000000000000000000000000000000000000000008352602060048401526024830190613346565b038173e749ecf94531a3ddbc5f293b91c9dabe1df44a065af49384156102dd575f946107b0575b508154928260018594015460081c6001600160a01b03168093546001600160a01b0316946004869301986107438a613797565b94369061074f926128a9565b90610759946152b4565b8361076386613797565b61076d9488614cef565b606001516001600160401b031660405191829161078a9183613839565b037fb8568a1f475f3c76759a620e08a653d28348c5c09e2e0bc91d533339801fefd891a2005b6107d391945060a03d60a0116107da575b6107cb8183612769565b8101906132e7565b9286610710565b503d6107c1565b346102e85760203660031901126102e8576107fa613cd8565b506004355f52600560205260405f206040519061081682612718565b805482526108de6001820154916001600160a01b0360ff841693602086019461083e8161255f565b855260081c1660408501526002810154936001600160a01b03851660608201526001600160401b03608082019560a01c1685526001600160401b03610891600460038501549460a0850195865201613797565b9160c08101928352519451956108a68761255f565b5116915190519160405195869586526108be8161255f565b60208601526040850152606084015260a0608084015260a083019061259a565b0390f35b346102e85760203660031901126102e8576108fb613cd8565b506004355f52600260205260405f206040519061010082018281106001600160401b03821117610a0f57604052805482526108de6001820154916001600160a01b0360ff84169360208601946109508161255f565b855260081c1660408501526002810154936001600160a01b03851660608201526001600160401b03608082019560a01c1685526001600160401b036003830154169160a082019283526001600160401b03806109ba600560048501549460c0870195865201613797565b9360e08101948552519651976109cf8961255f565b5116935116905191519260405196879687526109ea8161255f565b602087015260408601526060850152608084015260c060a084015260c083019061259a565b634e487b7160e01b5f52604160045260245ffd5b610a37610a2f366124ad565b505090613a02565b005b610a42366123a7565b6020810135600a8110156102e8576003610a5c91146139b7565b815f525f60205260405f206002600182019101916001600160a01b038354165f52600660205260405f206001600160a01b03610a9a608084016129af565b165f52602052610ab460e0826101de60405f205488613fed565b0381730827b6aaa03475a8bf59ee1a2bd76903ddfbadb65af49384156102dd577f188e0ade7d115cc397426774adb960ae3e8c83e72f0a6cad4b7085e1d60bf986946102a19461028d935f926102a6575061027761027c929361025d36886128fa565b60803660031901126102e8576004356024356001600160401b0381116102e857806004019061026060031982360301126102e8576044356001600160401b0381116102e857610b6a90369060040161247d565b50506064356001600160401b0381116102e857610b8b903690600401612503565b9190845f525f60205260405f209260ff845416600681101561064c576001610bb3911461396c565b610bbf60048501613797565b90610bc9866131c4565b6001600160401b0380845116911610610dfe578660018601936001600160a01b03855460201c16956001600160a01b03600289015416946001600160401b0380610c128c6131c4565b925116911611610ceb575b509163ffffffff9591610c40610c469594610c38368c6128fa565b9336916128a9565b916152b4565b600260ff1984541617835554166001600160401b03421601916001600160401b038311610cd7577f07b9206d5a6026d3bd2a8f9a9b79f6fa4bfbd6a016975829fbaf07488019f28a926013610ccb93016001600160401b0382166001600160401b03198254161790556001600160401b03604051938493604085526040850190612b94565b911660208301520390a2005b634e487b7160e01b5f52601160045260245ffd5b9591509291602486013595600a8710156102e857610d0c610d5e97156126cd565b835f5260066020526001600160a01b03610d2c608460405f2093016129af565b165f5260205260e088610d4360405f20548c613fed565b6040519889928392632a2d120f60e21b845260048401612d96565b0381730827b6aaa03475a8bf59ee1a2bd76903ddfbadb65af480156102dd57610c40610c4695610dcb8b8d9488888763ffffffff9e5f94610dd7575b50610daa610daf949536906128fa565b613d97565b8c610dc4610dbc8c612dad565b9136906128fa565b908661539b565b92949550509195610c1d565b610daf9450610df7610daa9160e03d60e0116102d6576102c78183612769565b9450610d9a565b608460405162461bcd60e51b815260206004820152603560248201527f6368616c6c656e67652063616e646964617465206d757374206861766520686960448201527f67686572206f7220657175616c2076657273696f6e00000000000000000000006064820152fd5b346102e857610e7636612367565b6020810135600a8110156102e8576006610e9091146126cd565b610ea2610e9d36846127af565b613d28565b91610ead36836128fa565b91610ed460208301936040610ec1866129af565b94019386610ece866129af565b92613d97565b610ee6610ee0826131c4565b85614c5c565b92610ef085613e8d565b156110025750610f669150835f525f60205260405f20906001600160a01b036002830154165f52600660205260405f206001600160a01b03610f34608084016129af565b165f5260205260e081610f4b60405f205488613fed565b6040519586928392632a2d120f60e21b845260048401612d96565b0381730827b6aaa03475a8bf59ee1a2bd76903ddfbadb65af49283156102dd577f587faad1bcd589ce902468251883e1976a645af8563c773eed7356d78433210c93610fd59361028d925f92610fda575b506001610fc49101612dad565b610fce36856128fa565b9088614046565b0390a3005b610fc4919250610ffa60019160e03d60e0116102d6576102c78183612769565b929150610fb7565b9061104e60a08261101b611015876129af565b88614c85565b60405193849283927eea54e70000000000000000000000000000000000000000000000000000000084526004840161339f565b038173e749ecf94531a3ddbc5f293b91c9dabe1df44a065af480156102dd577f17eb0a6bd5a0de45d1029ce3444941070e149df35b22176fc439f930f73c09f794610fd59461028d935f936110c3575b506110ab6110b1916129af565b916129af565b916110bc36866128fa565b8989614cef565b6110b19193506110e46110ab9160a03d60a0116107da576107cb8183612769565b93915061109e565b346102e8576110fa36612530565b90825f52600260205260405f209061111561068e8354613e8d565b61115e60c0611123866142cf565b604051809381927f6666e4c0000000000000000000000000000000000000000000000000000000008352602060048401526024830190612d0e565b0381739d8193e5655a36ffb9cd7d88d31c91d2650896d05af49384156102dd577fba075bd445233f7cad862c72f0343b3503aad9c8e704a2295f122b82abf8e801946001600160401b03936080935f92611219575b5060028261120693946111f689549160058b019a6111d08c613797565b84610c406001600160a01b0380600186015460081c16998a95015416998a9536916128a9565b6111ff89613797565b908b61438c565b015116906102a160405192839283613839565b611206925061124160029160c03d60c011611248575b6112398183612769565b810190612c9e565b92506111b3565b503d61122f565b346102e8575f3660031901126102e8576020604051620151808152f35b346102e8575f3660031901126102e8576020600454604051908152f35b346102e857610a3761129a366123a7565b906133b6565b346102e8575f3660031901126102e857602060405160408152f35b346102e85760203660031901126102e8576001600160a01b036112dc6123e0565b165f52600160205260405f20604051808260208294549384815201905f5260205f20925f5b81811061145057505061131692500382612769565b5f5f5b825181101561139b5761132c818461322e565b515f525f60205260ff60405f205416600681101561064c57600314158061136f575b61135b575b600101611319565b90611367600191612c78565b919050611353565b5061137a818461322e565b515f525f60205260ff60405f205416600681101561064c576005141561134e565b506113a5906131ef565b905f915f5b8251811015611442576113bd818461322e565b515f525f60205260ff60405f205416600681101561064c576003141580611416575b6113ec575b6001016113aa565b9261140e6001916113fd868661322e565b51611408828661322e565b52612c78565b9390506113e4565b50611421818461322e565b515f525f60205260ff60405f205416600681101561064c57600514156113df565b604051806108de8482612444565b8454835260019485019486945060209093019201611301565b346102e85760403660031901126102e8576108de61148b602435600435613242565b60405191829182612444565b346102e8575f3660031901126102e8576020604051612a308152f35b346102e85760403660031901126102e8576114cc6123e0565b602435906001600160a01b03821682036102e8576001600160a01b03165f5260066020526001600160a01b0360405f2091165f52602052602060405f2054604051908152f35b346102e857611520366123a7565b6020810135600a8110156102e857600961153a91146126cd565b815f525f60205260405f206116006001820160026001600160a01b03825460201c1693016115786001600160a01b038254168588610daa368a6128fa565b6001600160a01b0361158a36876128fa565b9161014087019561159a876131c4565b6001600160401b0316461496876116d7575b505054165f52600660205260405f206001600160a01b038060206060850151015116165f5260205260e0816115e560405f205489613fed565b6040519586928392632a2d120f60e21b845260048401612ab3565b0381730827b6aaa03475a8bf59ee1a2bd76903ddfbadb65af49182156102dd5761163b935f936116b2575b5061163590612dad565b86614046565b15611679576102a17f9a6f675cc94b83b55f1ecc0876affd4332a30c92e6faa2aca0199b1b6df922c391604051918291602083526020830190612b94565b6102a17f7b20773c41402791c5f18914dbbeacad38b1ebcc4c55d8eb3bfe0a4cde26c82691604051918291602083526020830190612b94565b6116359193506116d09060e03d60e0116102d6576102c78183612769565b929061162b565b6116e290369061281f565b60608501526116f43660608a0161281f565b6080850152604051611707602082612769565b5f815260a085015260405161171d602082612769565b5f815260c08501525f5260016020526117398860405f20615d4a565b5088806115ac565b61174a36612367565b906020820135600a8110156102e857600461176591146126cd565b611772610e9d36836127af565b9161177d36826128fa565b61179d60208401916040611790846129af565b95019486610ece876129af565b6117a9610ee0836131c4565b926117b385613e8d565b156118545750506117f790835f525f60205260405f20906001600160a01b036002830154165f52600660205260405f206001600160a01b03610f34608084016129af565b0381730827b6aaa03475a8bf59ee1a2bd76903ddfbadb65af49283156102dd577f471c4ebe4e57d25ef7117e141caac31c6b98f067b8098a7a7bbd38f637c2f98093610fd59361028d925f92610fda57506001610fc49101612dad565b906118809160c084611865876142cf565b6040519586928392632ef10bcd60e21b845260048401612d71565b0381739d8193e5655a36ffb9cd7d88d31c91d2650896d05af49182156102dd576118ca935f93611945575b506110ab6118b8916129af565b916118c336866128fa565b878761438c565b60035468010000000000000000811015610a0f577fede7867afa7cdb9c443667efd8244d98bf9df1dce68e60dc94dca6605125ca76918361192f611919846001610fd596016003556003612c63565b819391549060031b91821b915f19901b19161790565b9055604051918291602083526020830190612b94565b6118b89193506119666110ab9160c03d60c011611248576112398183612769565b9391506118ab565b611977366124ad565b50506020810135600a8110156102e857611a4657815f525f60205260405f206002600182019101916001600160a01b038354165f52600660205260405f206001600160a01b036119c9608084016129af565b165f526020526119e360e0826101de60405f205488613fed565b0381730827b6aaa03475a8bf59ee1a2bd76903ddfbadb65af49384156102dd577f567044ba1cdd4671ac3979c114241e1e3b56c9e9051f63f2f234f7a2795019cc946102a19461028d935f926102a6575061027761027c929361025d36886128fa565b608460405162461bcd60e51b815260206004820152602260248201527f63616e206f6e6c7920636865636b706f696e74206f706572617465207374617460448201527f65730000000000000000000000000000000000000000000000000000000000006064820152fd5b346102e85760203660031901126102e857600354600480549190355f5b82841080611c2f575b15611c0257611ae484612c37565b90549060031b1c5f52600260205260405f206001810160ff815416611b088161255f565b60038114611bef57600283015490426001600160401b038360a01c1611159081611bdb575b5015611ba35782611b9a94925f926004611b949601926001600160a01b0384549216855260066020526001600160a01b0380600c6040882093015460401c16168552602052611b81604085209182546131b7565b9055805460ff1916600317905555612c78565b93612c78565b915b9192611acd565b5050509050602091507f61815f4b11c6ea4e14a2e448a010bed8efdc3e53a15efbf183d16a31085cd145925b600455604051908152a1005b60019150611be88161255f565b1488611b2d565b50505092611bfc90612c78565b91611b9c565b9050602091507f61815f4b11c6ea4e14a2e448a010bed8efdc3e53a15efbf183d16a31085cd14592611bcf565b50818110611ad6565b611c4136612367565b6020810135600a8110156102e8575f9060028114808015611f79575b8015611f6a575b15611f0057611cd0611c79610e9d36886127af565b93611c8387613ee0565b60408701906001600160a01b03611c99836129af565b165f52600660205260405f206001600160a01b03611cb960808a016129af565b165f5260205260e087610f4b60405f205489613fed565b0381730827b6aaa03475a8bf59ee1a2bd76903ddfbadb65af49283156102dd575f93611edf575b50611d3c611d0536896128fa565b93611d2160208b0195611d17876129af565b8a610ece886129af565b611d2b368b6127af565b611d35368b6128fa565b9089614046565b6001600160a01b03611d4d846129af565b165f526001602052611d628660405f20615c6b565b506102e8576001600160a01b0380611e5b927fae3d48960dc29080438681a58800ea8520315e5fb998f450c039b9269201864f96611e14965f14611e605750877f6085f5128b19e0d3cc37524413de47259383f0f75265d5d66f417786962066966040516020815280611dd98d6020830190612b94565b0390a25b6080611de8866129af565b9a83611e1f611df6856129af565b94826040519b8c9b63ffffffff611e0c8861278a565b168d526123f6565b1660208b01526123f6565b1660408801526001600160401b03611e396060830161279b565b1660608801520135608086015260c060a08601521697169560c0830190612b94565b0390a4005b600303611ea757877f188e0ade7d115cc397426774adb960ae3e8c83e72f0a6cad4b7085e1d60bf9866040516020815280611e9f8d6020830190612b94565b0390a2611ddd565b877f567044ba1cdd4671ac3979c114241e1e3b56c9e9051f63f2f234f7a2795019cc6040516020815280611e9f8d6020830190612b94565b611ef991935060e03d60e0116102d6576102c78183612769565b9188611cf7565b608460405162461bcd60e51b815260206004820152602960248201527f696e76616c696420737461746520696e74656e7420666f72206368616e6e656c60448201527f206372656174696f6e00000000000000000000000000000000000000000000006064820152fd5b50916102e8575f918115611c64565b505f925060038214611c5d565b346102e85760203660031901126102e8576001600160a01b03611fa76123e0565b165f52600160205260405f206040519081602082549182815201915f5260205f20905f5b818110611fe2576108de8561148b81870382612769565b8254845260209093019260019283019201611fcb565b6001600160a01b036120093661240a565b929091169081156104235782156103fb5760206001600160a01b037f8752a472e571a816aea92eec8dae9baf628e840f4929fbcc2d155e6233ff68a792845f526006835260405f208282165f52835260405f206120678782546131b7565b9055612071615cc4565b61207c868233614a37565b60015f516020615f935f395f51905f52556040519586521693a3005b346102e857610a376120a9366123a7565b90612e08565b346102e8575f3660031901126102e8575f60035490600454915b8083106120db57602082604051908152f35b906120e583612c37565b90549060031b1c5f52600260205260405f20426001600160401b03600283015460a01c1611159081612131575b501561212b576104c561212491612c78565b91906120c9565b9061046d565b600180925060ff910154166121458161255f565b1484612112565b346102e85761215a36612367565b6020810135600a8110156102e857600861217491146126cd565b612181610e9d36846127af565b9161221461218f36846128fa565b9160208101926121b16121a1856129af565b91604084019288610ece856129af565b6001600160a01b036121de6121c636886128fa565b926121d089613e8d565b9687156122e7575b506129af565b165f52600660205260405f206001600160a01b038060206060850151015116165f5260205260e0816115e560405f205489613fed565b0381730827b6aaa03475a8bf59ee1a2bd76903ddfbadb65af49182156102dd5761224b935f936122c2575b506116359036906127af565b15612289576102a17f3142fb397e715d80415dff7b527bf1c451def4675da6e1199ee1b4588e3f630a91604051918291602083526020830190612b94565b6102a17f26afbcb9eb52c21f42eb9cfe8f263718ffb65afbf84abe8ad8cce2acfb2242b891604051918291602083526020830190612b94565b6116359193506122e09060e03d60e0116102d6576102c78183612769565b929061223f565b61234b84916122f588613ee0565b612303366101408d0161281f565b60608801526123153660608d0161281f565b6080880152604051612328602082612769565b5f815260a088015260405161233e602082612769565b5f815260c08801526129af565b165f5260016020526123608960405f20615c6b565b50896121d8565b90600319820160c081126102e85760a0136102e85760049160a435906001600160401b0382116102e8576102609082900360031901126102e85760040190565b9060406003198301126102e85760043591602435906001600160401b0382116102e8576102609082900360031901126102e85760040190565b600435906001600160a01b03821682036102e857565b35906001600160a01b03821682036102e857565b60609060031901126102e8576004356001600160a01b03811681036102e857906024356001600160a01b03811681036102e8579060443590565b60206040818301928281528451809452019201905f5b8181106124675750505090565b825184526020938401939092019160010161245a565b9181601f840112156102e8578235916001600160401b0383116102e8576020808501948460051b0101116102e857565b60606003198201126102e857600435916024356001600160401b0381116102e85761026081840360031901126102e85760040191604435906001600160401b0382116102e8576124ff9160040161247d565b9091565b9181601f840112156102e8578235916001600160401b0383116102e857602083818601950101116102e857565b9060406003198301126102e85760043591602435906001600160401b0382116102e8576124ff91600401612503565b6004111561064c57565b90600a82101561064c5752565b805180835260209291819084018484015e5f828201840152601f01601f1916010190565b6126bd916001600160401b0382511681526125bd60208301516020830190612569565b6040820151604082015261262a6060830151606083019060c080916001600160401b0381511684526001600160a01b03602082015116602085015260ff6040820151166040850152606081015160608501526080810151608085015260a081015160a08501520151910152565b60808281015180516001600160401b031661014084015260208101516001600160a01b0316610160840152604081015160ff1661018084015260608101516101a0840152908101516101c083015260a08101516101e083015260c0015161020082015260c06126ab60a0840151610260610220850152610260840190612576565b92015190610240818403910152612576565b90565b90600682101561064c5752565b156126d457565b606460405162461bcd60e51b815260206004820152600e60248201527f696e76616c696420696e74656e740000000000000000000000000000000000006044820152fd5b60e081019081106001600160401b03821117610a0f57604052565b60a081019081106001600160401b03821117610a0f57604052565b60c081019081106001600160401b03821117610a0f57604052565b90601f801991011681019081106001600160401b03821117610a0f57604052565b359063ffffffff821682036102e857565b35906001600160401b03821682036102e857565b91908260a09103126102e8576040516127c781612733565b60808082946127d58161278a565b84526127e3602082016123f6565b60208501526127f4604082016123f6565b60408501526128056060820161279b565b60608501520135910152565b359060ff821682036102e857565b91908260e09103126102e85760405161283781612718565b60c08082946128458161279b565b8452612853602082016123f6565b602085015261286460408201612811565b6040850152606081013560608501526080810135608085015260a081013560a08501520135910152565b6001600160401b038111610a0f57601f01601f191660200190565b9291926128b58261288e565b916128c36040519384612769565b8294818452818301116102e8578281602093845f960137010152565b9080601f830112156102e8578160206126bd933591016128a9565b9190610260838203126102e8576040519061291482612718565b819361291f8161279b565b83526020810135600a8110156102e85760208401526040810135604084015261294b826060830161281f565b606084015261295e82610140830161281f565b60808401526102208101356001600160401b0381116102e857826129839183016128df565b60a0840152610240810135916001600160401b0383116102e85760c0926129aa92016128df565b910152565b356001600160a01b03811681036102e85790565b51906001600160401b03821682036102e857565b519081151582036102e857565b908160e09103126102e857604051906129fc82612718565b8051825260208101516020830152604081015160068110156102e857612a5d9160c0916040850152612a30606082016129c3565b6060850152612a41608082016129d7565b6080850152612a5260a082016129d7565b60a0850152016129d7565b60c082015290565b90612a718183516126c0565b60806001600160401b0381612a95602086015160a0602087015260a086019061259a565b94604081015160408601526060810151606086015201511691015290565b9091612aca6126bd93604084526040840190612a65565b91602081840391015261259a565b60c080916001600160401b03612aed8261279b565b1684526001600160a01b03612b04602083016123f6565b16602085015260ff612b1860408301612811565b166040850152606081013560608501526080810135608085015260a081013560a08501520135910152565b9035601e19823603018112156102e85701602081359101916001600160401b0382116102e85781360383136102e857565b908060209392818452848401375f828201840152601f01601f1916010190565b6001600160401b03612ba58261279b565b168252602081013591600a8310156102e857612bc86126bd936020830190612569565b60408201356040820152612be26060820160608401612ad8565b612bf461014082016101408401612ad8565b612c28612c1c612c08610220850185612b43565b610260610220860152610260850191612b74565b92610240810190612b43565b91610240818503910152612b74565b600354811015612c4f5760035f5260205f2001905f90565b634e487b7160e01b5f52603260045260245ffd5b8054821015612c4f575f5260205f2001905f90565b5f198114610cd75760010190565b90612c908161255f565b60ff80198354169116179055565b908160c09103126102e85760405190612cb68261274e565b8051825260208101516020830152604081015160048110156102e857612d069160a0916040850152612cea606082016129c3565b6060850152612cfb608082016129c3565b6080850152016129d7565b60a082015290565b908151612d1a8161255f565b815260a080612d38602085015160c0602086015260c085019061259a565b93604081015160408501526001600160401b0360608201511660608501526001600160401b036080820151166080850152015191015290565b9091612d886126bd93604084526040840190612d0e565b916020818403910152612b94565b9091612d886126bd93604084526040840190612a65565b90604051612dba81612733565b6080600282946001600160a01b03815463ffffffff8116865260201c1660208501526001600160401b0360018201546001600160a01b038116604087015260a01c1660608501520154910152565b805f52600260205260405f2060018101908154916001600160a01b038360081c16926001600160a01b0360028401541691612e438454613e8d565b91821560ff8216816131a3575b508061318d575b6130f85750506020860135600a8110156102e8576005612e7791146126cd565b612e8e8285612e86368a6128fa565b865490613d97565b15612f985750612eec9150805490815f525f60205260e085610f4b60405f20946001600160a01b036002870154165f52600660205260405f206001600160a01b03612edb608086016129af565b165f5260205260405f205490613fed565b0381730827b6aaa03475a8bf59ee1a2bd76903ddfbadb65af49283156102dd577f32e24720f56fd5a7f4cb219d7ff3278ae95196e79c85b5801395894a6f53466c93612f7293612f5c925f92612f77575b50612f4c600185549201612dad565b612f56368a6128fa565b91614046565b5493604051918291602083526020830190612b94565b0390a3565b612f9191925060e03d60e0116102d6576102c78183612769565b905f612f3d565b90815f52600660205260405f206001600160a01b03612fba61016088016129af565b165f5260205261307460c08660405f205460405190612fd88261274e565b5f82528860208301612fe8613cd8565b81526001600160401b03600360408601925f8452606087015f815260808801945f865260a08901965f88525f52600260205260405f209260ff6001850154166130308161255f565b8a5261303e60058501613797565b90526004830154905283600283015460a01c1690520154169052526040519384928392632ef10bcd60e21b845260048401612d71565b0381739d8193e5655a36ffb9cd7d88d31c91d2650896d05af49283156102dd577f1b92e8ef67d8a7c0d29c99efcd180a5e0d98d60ac41d52abbbb5950882c78e4e94612f7294612f5c935f916130d9575b5084546130d2368b6128fa565b908961438c565b6130f2915060c03d60c011611248576112398183612769565b5f6130c5565b7f1b92e8ef67d8a7c0d29c99efcd180a5e0d98d60ac41d52abbbb5950882c78e4e95506131669250906003612f7295949260ff191617905560048301905f82549255600384016001600160401b031981541690556001600160a01b03600c85015460401c1690610371615cc4565b60015f516020615f935f395f51905f52555493604051918291602083526020830190612b94565b506001600160401b036003860154164211612e57565b600291506131b08161255f565b145f612e50565b91908201809211610cd757565b356001600160401b03811681036102e85790565b6001600160401b038111610a0f5760051b60200190565b906131f9826131d8565b6132066040519182612769565b8281528092613217601f19916131d8565b0190602036910137565b91908203918211610cd757565b8051821015612c4f5760209160051b010190565b91906003549080840293808504821490151715610cd757818410156132cb5761326b90846131b7565b908082116132c3575b506132876132828483613221565b6131ef565b92805b82811061329657505050565b806132a2600192612c37565b90549060031b1c6132bc6132b68584613221565b8861322e565b520161328a565b90505f613274565b505090506040516132dd602082612769565b5f81525f36813790565b908160a09103126102e857604051906132ff82612733565b8051825260208101516020830152604081015160048110156102e85761333e916080916040850152613333606082016129c3565b6060850152016129d7565b608082015290565b9081516133528161255f565b815260806001600160a01b0381613378602086015160a0602087015260a086019061259a565b94604081015160408601526001600160401b03606082015116606086015201511691015290565b9091612d886126bd93604084526040840190613346565b805f52600560205260405f20805492600182018054906001600160a01b038260081c1693600281018054926001600160a01b038416946133f58a613e8d565b9485159060ff8316826135f4575b50816135de575b5061356a57505050506020830135600a8110156102e857600761342d91146126cd565b61343d828588610daa36886128fa565b156134de57506134819150835f525f60205260405f20906001600160a01b036002830154165f52600660205260405f206001600160a01b03610f34608084016129af565b0381730827b6aaa03475a8bf59ee1a2bd76903ddfbadb65af49283156102dd577f6d0cf3d243d63f08f50db493a8af34b27d4e3bc9ec4098e82700abfeffe2d49893612f729361028d925f92610fda57506001610fc49101612dad565b906134ef60a08261101b8588614c85565b038173e749ecf94531a3ddbc5f293b91c9dabe1df44a065af49283156102dd577f2fdac1380dbe23ae259b6871582b7f33e34461547f400bdd20d74991250317d194612f729461028d935f9161354b575b506110bc36866128fa565b613564915060a03d60a0116107da576107cb8183612769565b5f613540565b612f729699507f2fdac1380dbe23ae259b6871582b7f33e34461547f400bdd20d74991250317d1975061316694506003905f969394965060ff191617905560038401915f8354935567ffffffffffffffff60a01b1981541690556001600160a01b03600b85015460401c1690610371615cc4565b6001600160401b03915060a01c1642115f61340a565b60029192506136028161255f565b14905f613403565b1561361157565b608460405162461bcd60e51b815260206004820152602760248201527f6f6e6c79206e6f6e2d686f6d6520657363726f77732063616e2062652063686160448201527f6c6c656e676564000000000000000000000000000000000000000000000000006064820152fd5b9060405161368881612718565b60c06004829460ff81546001600160401b03811686526001600160a01b038160401c16602087015260e01c1660408501526001810154606085015260028101546080850152600381015460a08501520154910152565b90600182811c9216801561370c575b60208310146136f857565b634e487b7160e01b5f52602260045260245ffd5b91607f16916136ed565b5f9291815491613725836136de565b808352926001811690811561377a575060011461374157505050565b5f9081526020812093945091925b838310613760575060209250010190565b60018160209294939454838587010152019101919061374f565b915050602093945060ff929192191683830152151560051b010190565b906040516137a481612718565b809260ff81546001600160401b038116845260401c1691600a83101561064c5761383560c092600d946020840152600181015460408401526137e86002820161367b565b60608401526137f96007820161367b565b60808401526040516138198161381281600c8601613716565b0382612769565b60a084015261382e6040518096819301613716565b0384612769565b0152565b906001600160401b036139656020929594956040855261386c8154848116604088015260ff606088019160401c16612569565b600181015460808601526138d660a0860160028301600460c09160ff81546001600160401b03811686526001600160a01b038160401c16602087015260e01c1660408501526001810154606085015260028101546080850152600381015460a08501520154910152565b60078101546001600160401b038116610180870152604081901c6001600160a01b03166101a087015260e01c60ff166101c086015260088101546101e08601526009810154610200860152600a810154610220860152600b81015461024086015261026080860152600d6139516102a08701600c8401613716565b868103603f19016102808801529101613716565b9416910152565b1561397357565b606460405162461bcd60e51b815260206004820152601660248201527f696e76616c6964206368616e6e656c20737461747573000000000000000000006044820152fd5b156139be57565b606460405162461bcd60e51b815260206004820152601460248201527f696e76616c696420737461746520696e74656e740000000000000000000000006044820152fd5b906020810135600a8110156102e8576001613a1d91146139b7565b815f525f60205260405f2060ff8154169160068310158061064c57600184148015613c95575b613a4c9061396c565b613a5860048401613797565b92600181019360028201916001600160a01b03835416966001600160a01b03875460201c169461064c5760021480613c7f575b613b9a5750506001600160a01b039054165f52600660205260405f206001600160a01b03613abb608085016129af565b165f52602052613ad560e0836101de60405f205489613fed565b0381730827b6aaa03475a8bf59ee1a2bd76903ddfbadb65af49485156102dd577f04cd8c68bf83e7bc531ca5a5d75c34e36513c2acf81e07e6470ba79e29da13a895613b6e95613b45935f92613b73575b50610277613b3b9293868b610daa368b6128fa565b610fce36866128fa565b5f526001602052613b598460405f20615d4a565b50604051918291602083526020830190612b94565b0390a2565b613b3b9250613b936102779160e03d60e0116102d6576102c78183612769565b9250613b26565b613b6e95507f04cd8c68bf83e7bc531ca5a5d75c34e36513c2acf81e07e6470ba79e29da13a896919450613c459250601390600360ff198254161781555f6012820155016001600160401b0319815416905560608401613c17815160606001600160a01b0360208301511691015190613c11615cc4565b866141f6565b60015f516020615f935f395f51905f52555160a06001600160a01b0360208301511691015191610371615cc4565b60015f516020615f935f395f51905f52555f526001602052613c6a8460405f20615d4a565b5060405191829160208352602083019061259a565b506001600160401b036013820154164211613a8b565b505f905060028414613a43565b60405190613caf82612718565b5f60c0838281528260208201528260408201528260608201528260808201528260a08201520152565b60405190613ce582612718565b606060c0835f81525f60208201525f6040820152613d01613ca2565b83820152613d0d613ca2565b60808201528260a08201520152565b600682101561064c5752565b604051613d826020820180936080809163ffffffff81511684526001600160a01b0360208201511660208501526001600160a01b0360408201511660408501526001600160401b0360608201511660608501520151910152565b60a08152613d9160c082612769565b51902090565b6001600160a01b03613de99493613de0829360c0613dc0613dbb613dd8988461596b565b615a89565b91613dcf60a082015184615e0f565b90999199615e49565b015190615e0f565b90979197615e49565b16911603613e49576001600160a01b03809116911603613e0557565b606460405162461bcd60e51b815260206004820152601660248201527f696e76616c6964206e6f6465207369676e6174757265000000000000000000006044820152fd5b606460405162461bcd60e51b815260206004820152601660248201527f696e76616c69642075736572207369676e6174757265000000000000000000006044820152fd5b805f525f60205260ff60405f205416600681101561064c578015908115613ed5575b50613ed0575f525f6020526001600160401b03600660405f20015416461490565b505f90565b60059150145f613eaf565b602081016001600160a01b03613ef5826129af565b161561042357604082016001600160a01b03613f10826129af565b161561042357613f3f906001600160a01b0380613f35613f2f866129af565b936129af565b16911614916129af565b90613f8757503563ffffffff81168091036102e8576201518011613f5f57565b7f0596b15b000000000000000000000000000000000000000000000000000000005f5260045ffd5b6001600160a01b03907fabfa558d000000000000000000000000000000000000000000000000000000005f521660045260245ffd5b60405190613fc982612733565b5f608083828152613fd8613cd8565b60208201528260408201528260608201520152565b9060136001600160401b0391614001613fbc565b935f525f60205260405f209061401b60ff83541686613d1c565b61402760048301613797565b6020860152601282015460408601526060850152015416608082015290565b90929192815f525f60205260405f209360ff85541691600683101561064c57614075938593156141275761539b565b604081018051600681101561064c57151580614105575b6140e5575b508060a060c09201516140c2575b01516140a85750565b5f6012820155601301805467ffffffffffffffff19169055565b825460ff1916600117835560138301805467ffffffffffffffff1916905561409f565b5190600682101561064c5760c09160ff8019855416911617835590614091565b5060ff835416815190600682101561064c57600681101561064c57141561408c565b6001870163ffffffff8351168154907fffffffffffffffff00000000000000000000000000000000000000000000000077ffffffffffffffffffffffffffffffffffffffff00000000602087015160201b169216171790556141e7600288016001600160a01b03806040860151161673ffffffffffffffffffffffffffffffffffffffff198254161781556001600160401b0360608501511667ffffffffffffffff60a01b1967ffffffffffffffff60a01b83549260a01b169116179055565b6080820151600388015561539b565b9082156142ca576001600160a01b03168061422457505f8080936001600160a01b0382941682f1156102dd57565b916001600160a01b03604051927fa9059cbb000000000000000000000000000000000000000000000000000000005f521660045260245260205f60448180865af19060015f51148216156142a9575b6040521561427e5750565b7f5274afe7000000000000000000000000000000000000000000000000000000005f5260045260245ffd5b9060018115166142c157823b15153d15161690614273565b503d5f823e3d90fd5b505050565b5f604051916142dd8361274e565b818352602083016142ec613cd8565b81526001600160401b036003604086019285845260608701868152608088019487865260a089019688885288526002602052604088209260ff6001850154166143348161255f565b8a5261434260058501613797565b90526004830154905283600283015460a01c16905201541690525290565b7f80000000000000000000000000000000000000000000000000000000000000008114610cd7575f0390565b9594939290955f52600260205260405f20916040820180516143ad8161255f565b6143b68161255f565b614a1a575b5060a082018051614602575b60209394959697506001600160401b03606084015116806145cc575b506001600160401b03608084015116806145a9575b505115614590578260806001600160a01b0392015101511680945b8251905f82131561453c5761443d915061442d8451615d15565b928391614438615cc4565b614a37565b60015f516020615f935f395f51905f525561445d600485019182546131b7565b90555b0180515f8113156144c85750906144bb926001600160a01b0361448560049451615d15565b95165f5260066020526001600160a01b0360405f2091165f5260205260405f206144b0858254613221565b9055019182546131b7565b90555b6144c6614aff565b565b9190505f82126144dc575b505050506144be565b614531926001600160a01b036144fb6144f6600495614360565b615d15565b95165f5260066020526001600160a01b0360405f2091165f5260205260405f206145268582546131b7565b905501918254613221565b90555f8080806144d3565b5f821261454c575b505050614460565b61455b6144f661456693614360565b928391610371615cc4565b60015f516020615f935f395f51905f525561458660048501918254613221565b9055835f80614544565b506001600160a01b03600c84015460401c168094614413565b6001600160401b036003870191166001600160401b03198254161790555f6143f8565b6145fc90600287019067ffffffffffffffff60a01b1967ffffffffffffffff60a01b83549260a01b169116179055565b5f6143e3565b600584016001600160401b0380845116166001600160401b03198254161781556020830151600a81101561064c5768ff000000000000000082549160401b169068ff000000000000000019161790556040820151600685015560c0600785016060840151906001600160401b0380835116166001600160401b03198254161781556020820151815468010000000000000000600160e81b031968010000000000000000600160e01b0360ff60e01b604087015160e01b169360401b16911617179055606081015160088701556080810151600987015560a0810151600a8701550151600b85015560c0600c85016080840151906001600160401b0380835116166001600160401b03198254161781556020820151815468010000000000000000600160e81b031968010000000000000000600160e01b0360ff60e01b604087015160e01b169360401b169116171790556060810151600d8701556080810151600e87015560a0810151600f870155015160108501556011840160a08301518051906001600160401b038211610a0f57819061479d84546136de565b601f81116149ca575b50602090601f8311600114614967575f9261495c575b50508160011b915f199060031b1c19161790555b601284019760c08301519889516001600160401b038111610a0f576147f582546136de565b601f8111614917575b506020601f82116001146148af578190602098999a9b9c5f926148a4575b50508160011b915f199060031b1c19161790555b855560018501805474ffffffffffffffffffffffffffffffffffffffff001916600888901b74ffffffffffffffffffffffffffffffffffffffff0016179055600285016001600160a01b03881673ffffffffffffffffffffffffffffffffffffffff198254161790558796959493506143c7565b015190505f8061481c565b601f1982169b835f52815f209c5f5b8181106148ff5750916020999a9b9c9d918460019594106148e7575b505050811b019055614830565b01515f1960f88460031b161c191690555f80806148da565b838301518f556001909e019d602093840193016148be565b825f5260205f20601f830160051c81019160208410614952575b601f0160051c01905b81811061494757506147fe565b5f815560010161493a565b9091508190614931565b015190505f806147bc565b5f8581528281209350601f198516905b8181106149b2575090846001959493921061499a575b505050811b0190556147d0565b01515f1960f88460031b161c191690555f808061498d565b92936020600181928786015181550195019301614977565b909150835f5260205f20601f840160051c81019160208510614a10575b90601f859493920160051c01905b818110614a0257506147a6565b5f81558493506001016149f5565b90915081906149e7565b614a319051614a288161255f565b60018501612c86565b5f6143bb565b9082156142ca576001600160a01b03169182158015614aea57813403614adb575b15614a6257505050565b6001600160a01b03604051927f23b872dd000000000000000000000000000000000000000000000000000000005f52166004523060245260445260205f60648180865af19060015f5114821615614ac3575b6040525f6060521561427e5750565b9060018115166142c157823b15153d15161690614ab4565b632a9ffab760e21b5f5260045ffd5b3415614a5857632a9ffab760e21b5f5260045ffd5b6003546004545f5b82821080614c52575b15614c2757614b1e82612c37565b90549060031b1c5f52600260205260405f206001810160ff815416614b428161255f565b60038114614c1457600283015490426001600160401b038360a01c1611159081614c00575b5015614bca5782614bc194925f926004614bbb9601926001600160a01b0384549216855260066020526001600160a01b0380600c6040882093015460401c16168552602052611b81604085209182546131b7565b91612c78565b915b9190614b07565b5050507f61815f4b11c6ea4e14a2e448a010bed8efdc3e53a15efbf183d16a31085cd14592506020915b600455604051908152a1565b60019150614c0d8161255f565b145f614b67565b50505090614c2190612c78565b91614bc3565b7f61815f4b11c6ea4e14a2e448a010bed8efdc3e53a15efbf183d16a31085cd1459250602091614bf4565b5060408110614b10565b906001600160401b03604051916020830193845216604082015260408152613d91606082612769565b906001600160a01b0390614c97613fbc565b925f5260056020526001600160401b03600260405f2060ff600182015416614cbe8161255f565b8652614ccc60048201613797565b602087015260038101546040870152015460a01c16606084015216608082015290565b9594939290955f52600560205260405f2091604082018051614d108161255f565b614d198161255f565b6152a0575b50608082018051614e88575b60209394959697506001600160401b0360608401511680614e52575b505115614e39578260806001600160a01b0392015101511680945b8251905f821315614df057614d7b915061442d8451615d15565b60015f516020615f935f395f51905f5255614d9b600385019182546131b7565b90555b0180515f811315614dc35750906144bb926001600160a01b0361448560039451615d15565b9190505f8212614dd657505050506144be565b614531926001600160a01b036144fb6144f6600395614360565b5f8212614e00575b505050614d9e565b61455b6144f6614e0f93614360565b60015f516020615f935f395f51905f5255614e2f60038501918254613221565b9055835f80614df8565b506001600160a01b03600b84015460401c168094614d61565b614e8290600287019067ffffffffffffffff60a01b1967ffffffffffffffff60a01b83549260a01b169116179055565b5f614d46565b600484016001600160401b0380845116166001600160401b03198254161781556020830151600a81101561064c5768ff000000000000000082549160401b169068ff000000000000000019161790556040820151600585015560c0600685016060840151906001600160401b0380835116166001600160401b03198254161781556020820151815468010000000000000000600160e81b031968010000000000000000600160e01b0360ff60e01b604087015160e01b169360401b16911617179055606081015160078701556080810151600887015560a081015160098701550151600a85015560c0600b85016080840151906001600160401b0380835116166001600160401b03198254161781556020820151815468010000000000000000600160e81b031968010000000000000000600160e01b0360ff60e01b604087015160e01b169360401b169116171790556060810151600c8701556080810151600d87015560a0810151600e8701550151600f8501556010840160a08301518051906001600160401b038211610a0f57819061502384546136de565b601f8111615250575b50602090601f83116001146151ed575f926151e2575b50508160011b915f199060031b1c19161790555b601184019760c08301519889516001600160401b038111610a0f5761507b82546136de565b601f811161519d575b506020601f8211600114615135578190602098999a9b9c5f9261512a575b50508160011b915f199060031b1c19161790555b855560018501805474ffffffffffffffffffffffffffffffffffffffff001916600888901b74ffffffffffffffffffffffffffffffffffffffff0016179055600285016001600160a01b03881673ffffffffffffffffffffffffffffffffffffffff19825416179055879695949350614d2a565b015190505f806150a2565b601f1982169b835f52815f209c5f5b8181106151855750916020999a9b9c9d9184600195941061516d575b505050811b0190556150b6565b01515f1960f88460031b161c191690555f8080615160565b838301518f556001909e019d60209384019301615144565b825f5260205f20601f830160051c810191602084106151d8575b601f0160051c01905b8181106151cd5750615084565b5f81556001016151c0565b90915081906151b7565b015190505f80615042565b5f8581528281209350601f198516905b8181106152385750908460019594939210615220575b505050811b019055615056565b01515f1960f88460031b161c191690555f8080615213565b929360206001819287860151815501950193016151fd565b909150835f5260205f20601f840160051c81019160208510615296575b90601f859493920160051c01905b818110615288575061502c565b5f815584935060010161527b565b909150819061526d565b6152ae9051614a288161255f565b5f614d1e565b6153256001600160a01b03936153206020613dbb6009826152d98a9961532e9961596b565b6040519481869251918291018484015e81017f6368616c6c656e67650000000000000000000000000000000000000000000000838201520301601619810184520182612769565b615e0f565b90929192615e49565b1691168114918215615388575b50501561534457565b606460405162461bcd60e51b815260206004820152601f60248201527f6368616c6c656e676572206d757374206265206e6f6465206f722075736572006044820152fd5b6001600160a01b03161490505f8061533b565b9291925f525f60205260405f209260808301516155c3575b6060019160c06001600160a01b03602085510151169180515f8113615577575b506020810180515f811361551e575b5081515f81126154cf575b50515f8112615473575b500151151580615465575b615413575b505050506144c6614aff565b61545a9261543c60a0926001600160a01b03604060129601511690848451015191610371615cc4565b60015f516020615f935f395f51905f52555101519201918254613221565b90555f808080615407565b5060a0835101511515615402565b6144f661547f91614360565b6001600160a01b036040860151165f52600660205260405f206001600160a01b0385165f5260205260405f206154b68282546131b7565b90556154c760128801918254613221565b90555f6153f7565b6144f66154db91614360565b6154f681866001600160a01b0360208a015116610371615cc4565b60015f516020615f935f395f51905f525561551660128901918254613221565b90555f6153ed565b61552790615d15565b6001600160a01b036040870151165f52600660205260405f206001600160a01b0386165f5260205260405f2061555e828254613221565b905561556f601289019182546131b7565b90555f6153e2565b61558090615d15565b61559b81856001600160a01b03602089015116614438615cc4565b60015f516020615f935f395f51905f52556155bb601288019182546131b7565b90555f6153d3565b600484016001600160401b0380835116166001600160401b03198254161781556020820151600a81101561064c5768ff000000000000000082549160401b169068ff000000000000000019161790556040810151600585015560c0600685016060830151906001600160401b0380835116166001600160401b03198254161781556020820151815468010000000000000000600160e81b031968010000000000000000600160e01b0360ff60e01b604087015160e01b169360401b16911617179055606081015160078701556080810151600887015560a081015160098701550151600a85015560c0600b85016080830151906001600160401b0380835116166001600160401b03198254161781556020820151815468010000000000000000600160e81b031968010000000000000000600160e01b0360ff60e01b604087015160e01b169360401b169116171790556060810151600c8701556080810151600d87015560a0810151600e8701550151600f8501556010840160a08201518051906001600160401b038211610a0f57819061575e84546136de565b601f811161591b575b50602090601f83116001146158b8575f926158ad575b50508160011b915f199060031b1c19161790555b6011840160c08201518051906001600160401b038211610a0f576157b583546136de565b601f8111615868575b50602090601f8311600114615801576060949392915f91836157f6575b50508160011b915f199060031b1c19161790555b90506153b3565b015190505f806157db565b90601f19831691845f52815f20925f5b818110615850575091600193918560609897969410615838575b505050811b0190556157ef565b01515f1960f88460031b161c191690555f808061582b565b92936020600181928786015181550195019301615811565b835f5260205f20601f840160051c810191602085106158a3575b601f0160051c01905b81811061589857506157be565b5f815560010161588b565b9091508190615882565b015190505f8061577d565b5f8581528281209350601f198516905b81811061590357509084600195949392106158eb575b505050811b019055615791565b01515f1960f88460031b161c191690555f80806158de565b929360206001819287860151815501950193016158c8565b909150835f5260205f20601f840160051c81019160208510615961575b90601f859493920160051c01905b8181106159535750615767565b5f8155849350600101615946565b9091508190615938565b6001600160401b038151166020820151600a81101561064c5782615a18916159b86040615a789601519160806060850151940151956040519860208a015260408901526060880190612569565b608086015260a085019060c080916001600160401b0381511684526001600160a01b03602082015116602085015260ff6040820151166040850152606081015160608501526080810151608085015260a081015160a08501520151910152565b80516001600160401b031661018084015260208101516001600160a01b03166101a0840152604081015160ff166101c084015260608101516101e0840152608081015161020084015260a081015161022084015260c00151610240830152565b61024081526126bd61026082612769565b8051905f827a184f03e93ff9f4daa797ed6e38ed64bf6a1f010000000000000000811015615c43575b806d04ee2d6d415b85acef8100000000600a921015615c28575b662386f26fc10000811015615c14575b6305f5e100811015615c03575b612710811015615bf4575b6064811015615be6575b1015615bde575b6001810192600a6021615b30615b1a8761288e565b96615b286040519889612769565b80885261288e565b602087019490601f19013686378601015b5f1901917f30313233343536373839616263646566000000000000000000000000000000008282061a835304908115615b7c57600a90615b41565b5050613d9190603a6020604051948593828501977f19457468657265756d205369676e6564204d6573736167653a0a00000000000089525180918587015e8401908382015f8152815193849201905e01015f815203601f198101835282612769565b600101615b05565b606460029104920191615afe565b61271060049104920191615af4565b6305f5e10060089104920191615ae9565b662386f26fc1000060109104920191615adc565b6d04ee2d6d415b85acef810000000060209104920191615acc565b50604090507a184f03e93ff9f4daa797ed6e38ed64bf6a1f0100000000000000008304615ab2565b6001810190825f528160205260405f2054155f14615cbd57805468010000000000000000811015610a0f57615caa611919826001879401855584612c63565b905554915f5260205260405f2055600190565b5050505f90565b60025f516020615f935f395f51905f525414615ced5760025f516020615f935f395f51905f5255565b7f3ee5aeb5000000000000000000000000000000000000000000000000000000005f5260045ffd5b5f8112615d1f5790565b7fa8ce4432000000000000000000000000000000000000000000000000000000005f5260045260245ffd5b906001820191815f528260205260405f20548015155f14615e07575f198101818111610cd75782545f19810191908211610cd757818103615dd2575b50505080548015615dbe575f190190615d9f8282612c63565b8154905f199060031b1b19169055555f526020525f6040812055600190565b634e487b7160e01b5f52603160045260245ffd5b615df2615de26119199386612c63565b90549060031b1c92839286612c63565b90555f528360205260405f20555f8080615d86565b505050505f90565b8151919060418303615e3f57615e389250602082015190606060408401519301515f1a90615f10565b9192909190565b50505f9160029190565b615e528161255f565b80615e5b575050565b615e648161255f565b60018103615e94577ff645eedf000000000000000000000000000000000000000000000000000000005f5260045ffd5b615e9d8161255f565b60028103615ed157507ffce698f7000000000000000000000000000000000000000000000000000000005f5260045260245ffd5b600390615edd8161255f565b14615ee55750565b7fd78bce0c000000000000000000000000000000000000000000000000000000005f5260045260245ffd5b91907f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a08411615f87579160209360809260ff5f9560405194855216868401526040830152606082015282805260015afa156102dd575f516001600160a01b03811615615f7d57905f905f90565b505f906001905f90565b5050505f916003919056fe9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f00a26469706673582212200fd9539b3eb79e5b8d367c0198d5a9c4948d41768cf426c211b035f07250dbeb64736f6c634300081e0033",
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
// Solidity: function createChannel((uint32,address,address,uint64,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState) payable returns()
func (_ChannelHub *ChannelHubTransactor) CreateChannel(opts *bind.TransactOpts, def ChannelDefinition, initState State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "createChannel", def, initState)
}

// CreateChannel is a paid mutator transaction binding the contract method 0x28353129.
//
// Solidity: function createChannel((uint32,address,address,uint64,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState) payable returns()
func (_ChannelHub *ChannelHubSession) CreateChannel(def ChannelDefinition, initState State) (*types.Transaction, error) {
	return _ChannelHub.Contract.CreateChannel(&_ChannelHub.TransactOpts, def, initState)
}

// CreateChannel is a paid mutator transaction binding the contract method 0x28353129.
//
// Solidity: function createChannel((uint32,address,address,uint64,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState) payable returns()
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
