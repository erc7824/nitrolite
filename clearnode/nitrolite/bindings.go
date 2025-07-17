// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package nitrolite

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
	_ = abi.ConvertType
)

// Allocation is an auto generated low-level Go binding around an user-defined struct.
type Allocation struct {
	Destination common.Address
	Token       common.Address
	Amount      *big.Int
}

// Channel is an auto generated low-level Go binding around an user-defined struct.
type Channel struct {
	Participants []common.Address
	Adjudicator  common.Address
	Challenge    uint64
	Nonce        uint64
}

// State is an auto generated low-level Go binding around an user-defined struct.
type State struct {
	Intent      uint8
	Version     *big.Int
	Data        []byte
	Allocations []Allocation
	Sigs        [][]byte
}

// CustodyMetaData contains all meta data concerning the Custody contract.
var CustodyMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes[]\",\"name\":\"sigs\",\"type\":\"bytes[]\"}],\"internalType\":\"structState\",\"name\":\"candidate\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes[]\",\"name\":\"sigs\",\"type\":\"bytes[]\"}],\"internalType\":\"structState[]\",\"name\":\"proofs\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes\",\"name\":\"challengerSig\",\"type\":\"bytes\"}],\"name\":\"challenge\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes[]\",\"name\":\"sigs\",\"type\":\"bytes[]\"}],\"internalType\":\"structState\",\"name\":\"candidate\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes[]\",\"name\":\"sigs\",\"type\":\"bytes[]\"}],\"internalType\":\"structState[]\",\"name\":\"proofs\",\"type\":\"tuple[]\"}],\"name\":\"checkpoint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes[]\",\"name\":\"sigs\",\"type\":\"bytes[]\"}],\"internalType\":\"structState\",\"name\":\"candidate\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes[]\",\"name\":\"sigs\",\"type\":\"bytes[]\"}],\"internalType\":\"structState[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"name\":\"close\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address[]\",\"name\":\"participants\",\"type\":\"address[]\"},{\"internalType\":\"address\",\"name\":\"adjudicator\",\"type\":\"address\"},{\"internalType\":\"uint64\",\"name\":\"challenge\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"}],\"internalType\":\"structChannel\",\"name\":\"ch\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes[]\",\"name\":\"sigs\",\"type\":\"bytes[]\"}],\"internalType\":\"structState\",\"name\":\"initial\",\"type\":\"tuple\"}],\"name\":\"create\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"participants\",\"type\":\"address[]\"},{\"internalType\":\"address\",\"name\":\"adjudicator\",\"type\":\"address\"},{\"internalType\":\"uint64\",\"name\":\"challenge\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"}],\"internalType\":\"structChannel\",\"name\":\"ch\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes[]\",\"name\":\"sigs\",\"type\":\"bytes[]\"}],\"internalType\":\"structState\",\"name\":\"initial\",\"type\":\"tuple\"}],\"name\":\"depositAndCreate\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"eip712Domain\",\"outputs\":[{\"internalType\":\"bytes1\",\"name\":\"fields\",\"type\":\"bytes1\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"version\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"verifyingContract\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"salt\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"extensions\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"accounts\",\"type\":\"address[]\"},{\"internalType\":\"address[]\",\"name\":\"tokens\",\"type\":\"address[]\"}],\"name\":\"getAccountsBalances\",\"outputs\":[{\"internalType\":\"uint256[][]\",\"name\":\"\",\"type\":\"uint256[][]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"internalType\":\"address[]\",\"name\":\"tokens\",\"type\":\"address[]\"}],\"name\":\"getChannelBalances\",\"outputs\":[{\"internalType\":\"uint256[]\",\"name\":\"balances\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"}],\"name\":\"getChannelData\",\"outputs\":[{\"components\":[{\"internalType\":\"address[]\",\"name\":\"participants\",\"type\":\"address[]\"},{\"internalType\":\"address\",\"name\":\"adjudicator\",\"type\":\"address\"},{\"internalType\":\"uint64\",\"name\":\"challenge\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"}],\"internalType\":\"structChannel\",\"name\":\"channel\",\"type\":\"tuple\"},{\"internalType\":\"enumChannelStatus\",\"name\":\"status\",\"type\":\"uint8\"},{\"internalType\":\"address[]\",\"name\":\"wallets\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"challengeExpiry\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes[]\",\"name\":\"sigs\",\"type\":\"bytes[]\"}],\"internalType\":\"structState\",\"name\":\"lastValidState\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"accounts\",\"type\":\"address[]\"}],\"name\":\"getOpenChannels\",\"outputs\":[{\"internalType\":\"bytes32[][]\",\"name\":\"\",\"type\":\"bytes32[][]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"sig\",\"type\":\"bytes\"}],\"name\":\"join\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes[]\",\"name\":\"sigs\",\"type\":\"bytes[]\"}],\"internalType\":\"structState\",\"name\":\"candidate\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes[]\",\"name\":\"sigs\",\"type\":\"bytes[]\"}],\"internalType\":\"structState[]\",\"name\":\"proofs\",\"type\":\"tuple[]\"}],\"name\":\"resize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes[]\",\"name\":\"sigs\",\"type\":\"bytes[]\"}],\"indexed\":false,\"internalType\":\"structState\",\"name\":\"state\",\"type\":\"tuple\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"expiration\",\"type\":\"uint256\"}],\"name\":\"Challenged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes[]\",\"name\":\"sigs\",\"type\":\"bytes[]\"}],\"indexed\":false,\"internalType\":\"structState\",\"name\":\"state\",\"type\":\"tuple\"}],\"name\":\"Checkpointed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes[]\",\"name\":\"sigs\",\"type\":\"bytes[]\"}],\"indexed\":false,\"internalType\":\"structState\",\"name\":\"finalState\",\"type\":\"tuple\"}],\"name\":\"Closed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"wallet\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"participants\",\"type\":\"address[]\"},{\"internalType\":\"address\",\"name\":\"adjudicator\",\"type\":\"address\"},{\"internalType\":\"uint64\",\"name\":\"challenge\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"}],\"indexed\":false,\"internalType\":\"structChannel\",\"name\":\"channel\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes[]\",\"name\":\"sigs\",\"type\":\"bytes[]\"}],\"indexed\":false,\"internalType\":\"structState\",\"name\":\"initial\",\"type\":\"tuple\"}],\"name\":\"Created\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"wallet\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deposited\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"EIP712DomainChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"Joined\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"}],\"name\":\"Opened\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"int256[]\",\"name\":\"deltaAllocations\",\"type\":\"int256[]\"}],\"name\":\"Resized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"wallet\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Withdrawn\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"ChallengeNotExpired\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ChannelNotFinal\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"}],\"name\":\"ChannelNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DepositAlreadyFulfilled\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"expectedFulfilled\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"actualFulfilled\",\"type\":\"uint256\"}],\"name\":\"DepositsNotFulfilled\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ECDSAInvalidSignature\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"length\",\"type\":\"uint256\"}],\"name\":\"ECDSAInvalidSignatureLength\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"ECDSAInvalidSignatureS\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"available\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"required\",\"type\":\"uint256\"}],\"name\":\"InsufficientBalance\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidAdjudicator\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidAllocations\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidAmount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidChallengePeriod\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidChallengerSignature\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidParticipant\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidShortString\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidState\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidStateSignatures\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidStatus\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidValue\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"SafeERC20FailedOperation\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"str\",\"type\":\"string\"}],\"name\":\"StringTooLong\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"TransferFailed\",\"type\":\"error\"}]",
	Bin: "0x0x610160604052346101365760405161001860408261013a565b601181526020810190704e6974726f6c6974653a437573746f647960781b82526040519161004760408461013a565b600683526020830191650c0b8c8b8c8d60d21b835261006581610171565b6101205261007284610307565b61014052519020918260e05251902080610100524660a0526040519060208201927f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f8452604083015260608201524660808201523060a082015260a081526100db60c08261013a565b5190206080523060c0526040516153549081610440823960805181614a97015260a05181614b4e015260c05181614a68015260e05181614ae601526101005181614b0c01526101205181611068015261014051816110940152f35b5f80fd5b601f909101601f19168101906001600160401b0382119082101761015d57604052565b634e487b7160e01b5f52604160045260245ffd5b908151602081105f146101eb575090601f8151116101ab57602081519101516020821061019c571790565b5f198260200360031b1b161790565b604460209160405192839163305a27a960e01b83528160048401528051918291826024860152018484015e5f828201840152601f01601f19168101030190fd5b6001600160401b03811161015d575f54600181811c911680156102fd575b60208210146102e957601f81116102b7575b50602092601f821160011461025857928192935f9261024d575b50508160011b915f199060031b1c1916175f5560ff90565b015190505f80610235565b601f198216935f8052805f20915f5b86811061029f5750836001959610610287575b505050811b015f5560ff90565b01515f1960f88460031b161c191690555f808061027a565b91926020600181928685015181550194019201610267565b5f8052601f60205f20910160051c810190601f830160051c015b8181106102de575061021b565b5f81556001016102d1565b634e487b7160e01b5f52602260045260245ffd5b90607f1690610209565b908151602081105f14610332575090601f8151116101ab57602081519101516020821061019c571790565b6001600160401b03811161015d57600154600181811c91168015610435575b60208210146102e957601f8111610402575b50602092601f82116001146103a157928192935f92610396575b50508160011b915f199060031b1c19161760015560ff90565b015190505f8061037d565b601f1982169360015f52805f20915f5b8681106103ea57508360019596106103d2575b505050811b0160015560ff90565b01515f1960f88460031b161c191690555f80806103c4565b919260206001819286850151815501940192016103b1565b60015f52601f60205f20910160051c810190601f830160051c015b81811061042a5750610363565b5f815560010161041d565b90607f169061035156fe60806040526004361015610011575f80fd5b5f3560e01c8062e2bb2c146123555780631474e410146122d1578063183b499814611af05780632f33c4d6146119365780634a7e7798146118ca5780635a9eb80e146118285780637f9ebbd7146111705780638340f5491461113957806384b0196e14611050578063bab3290a14610a6e578063d710e92f146108e0578063e617208c1461078d578063ecf668fd146102455763f3fef3a3146100b2575f80fd5b34610241576040600319360112610241576100cb6123cb565b60243590335f52600360205260405f20906001600160a01b0381165f528160205260405f205491838310610211576001600160a01b0392508282165f5260205260405f2061011a8482546137c5565b90551690816101bf575f80808084335af13d156101ba573d61013b8161275c565b90610149604051928361251b565b81525f60203d92013e5b15610187575b6040519081527fd1c19fbcd4551a5edfb66d43d2e337c04837afda3482b42bdf569a8fccdae5fb60203392a3005b907fbf182be8000000000000000000000000000000000000000000000000000000005f526004523360245260445260645ffd5b610153565b61020c6040517fa9059cbb0000000000000000000000000000000000000000000000000000000060208201523360248201528260448201526044815261020660648261251b565b83614fc1565b610159565b50507fcf479181000000000000000000000000000000000000000000000000000000005f5260045260245260445ffd5b5f80fd5b346102415761025336612454565b835f52600260205260405f2091600383019160ff835416906005821015610779578115610766576004821461073e578535926004841015610241576102978461264f565b8361064257600f86019260ff84541690600181145f146102d9577ff525e320000000000000000000000000000000000000000000000000000000005f5260045ffd5b60020361068b57506001600160a01b036001870154169161030d6102fd368a612854565b61030686612af3565b9085614858565b15610642576103379260209260405180958194829363030232af60e21b84528d8d60048601612e43565b03915afa908115610680575f91610651575b501561064257610364925b600260ff19825416179055612ee8565b602082013560108201556011810161037f6040840184612f01565b9067ffffffffffffffff8211610559576103a38261039d85546129b8565b85612f4a565b5f90601f83116001146105de576103d192915f91836105d3575b50508160011b915f199060031b1c19161790565b90555b601281016103e56060840184612bb1565b91906103f18383612f8f565b905f5260205f205f915b83831061056d578686601387016104156080830183613011565b90916104218282613095565b5f9081526020812092805b83831061047257867f8cade4fe25d72146dc0dbe08ea2712bdcca7e2c996e2dce1e69f20e30ee1c5c361046d88604051918291602083526020830190612c6e565b0390a2005b61047c8183612f01565b9067ffffffffffffffff8211610559576104a08261049a89546129b8565b89612f4a565b5f90601f83116001146104ef57926104d5836001959460209487965f926104e45750508160011b915f199060031b1c19161790565b88555b0195019201919361042c565b013590508d806103bd565b601f19831691885f5260205f20925f5b8181106105415750936020936001969387969383889510610528575b505050811b0188556104d8565b01355f19600384901b60f8161c191690558c808061051b565b919360206001819287870135815501950192016104ff565b634e487b7160e01b5f52604160045260245ffd5b60036060826001600160a01b03610585600195612ffd565b166001600160a01b03198654161785556105a160208201612ffd565b6001600160a01b0385870191166001600160a01b031982541617905560408101356002860155019201920191906103fb565b0135905087806103bd565b601f19831691845f5260205f20925f5b81811061062a5750908460019594939210610611575b505050811b0190556103d4565b01355f19600384901b60f8161c19169055868080610604565b919360206001819287870135815501950192016105ee565b63baf3f0f760e01b5f5260045ffd5b610673915060203d602011610679575b61066b818361251b565b810190612de4565b87610349565b503d610661565b6040513d5f823e3d90fd5b909160206106bf916001600160a01b0360018a01541694604051938492839263030232af60e21b84528d8d60048601612e43565b0381865afa908115610680575f9161071f575b5015610642576106e18161264f565b156106f7575b50610364925f600e860155610354565b610714906107053688612854565b61070e84612af3565b91614858565b1561064257866106e7565b610738915060203d6020116106795761066b818361251b565b896106d2565b7ff525e320000000000000000000000000000000000000000000000000000000005f5260045ffd5b866379c1d89f60e11b5f5260045260245ffd5b634e487b7160e01b5f52602160045260245ffd5b34610241576020600319360112610241575f60606040516107ad816124c7565b81815282602082015282604082015201526107c66137d2565b506004355f52600260205260405f206107de81613770565b60ff60038301541691604051906107f660608361251b565b600282526040366020840137600481015f5b600281106108b2575050610823600f600e8301549201612af3565b6040519460a0865267ffffffffffffffff606061084d8751608060a08b01526101208a0190612613565b966001600160a01b0360208201511660c08a01528260408201511660e08a015201511661010087015260058110156107795785946108ae9461089b9260208801528682036040880152612613565b9160608501528382036080850152612659565b0390f35b806001600160a01b036108c76001938561394b565b90549060031b1c166108d9828761382d565b5201610808565b346102415760206003193601126102415760043567ffffffffffffffff811161024157610911903690600401612556565b8051906109366109208361253e565b9261092e604051948561251b565b80845261253e565b90610949601f19602085019301836138c4565b5f5b81518110156109dd576001600160a01b03610966828461382d565b51165f526003602052600160405f20016040519081602082549182815201915f5260205f20905f905b8082106109c557505050906109a98160019493038261251b565b6109b3828761382d565b526109be818661382d565b500161094b565b9091926001602081928654815201940192019061098f565b50509060405191829160208301906020845251809152604083019060408160051b85010192915f905b828210610a1557505050500390f35b9193909294603f19908203018252845190602080835192838152019201905f905b808210610a56575050506020806001929601920192018594939192610a06565b90919260208060019286518152019401920190610a36565b346102415760606003193601126102415760043560443567ffffffffffffffff811161024157610aa2903690600401612426565b825f93929352600260205260405f2090600382019060ff825416600581101561077957801561103d575f190161073e5760016024350361101557600d830154610fed57600f830194610b2b610af687612af3565b610afe614a5e565b906001600160a01b03610b10886143d0565b90549060031b1c169188610b25368888612778565b92614b74565b15610fc557610b3986612af3565b90610b8960405191610b4c60608461251b565b60028352610b5e6040602085016138c4565b6080840194610b6d86516137fc565b51610b77856137fc565b52610b81846137fc565b503691612778565b610b928261381d565b52610b9c8161381d565b50825260018401546001600160a01b0316604051906020610bbd818461251b565b5f8352601f19015f5b818110610fae575050610bf391602091604051808095819463030232af60e21b8352888c60048501613841565b03915afa908115610680575f91610f8f575b501561064257610c71610c1a600886016139be565b96610c4c88600c88019060206001916001600160a01b0380825116166001600160a01b03198554161784550151910155565b6005860180546001600160a01b03191633179055825190610c6c8261264f565b612ee8565b6020810151601085015560118401604082015180519067ffffffffffffffff821161055957610ca48261039d85546129b8565b602090601f8311600114610f2c57610cd292915f9183610f215750508160011b915f199060031b1c19161790565b90555b606060128501910151906020825192610cee8484612f8f565b01905f5260205f205f915b838310610ebe57505050506013830190516020815191610d198385613095565b01915f5260205f20915f905b828210610ddd57845460ff19166002178555602087610d858a6001600160a01b03610d4f8b6143d0565b90549060031b1c165f5260038452610d6d83600160405f2001614f59565b50836001600160a01b03825116910151908333614d06565b807fe8e915db7b3549b9e9e9b3e2ec2dc3edd1f76961504366998824836401f6846a8360405160018152a260405190807fd087f17acc177540af5f382bc30c65363705b90855144d285a822536ee11fdd15f80a28152f35b805180519067ffffffffffffffff821161055957610e0582610dff88546129b8565b88612f4a565b602090601f8311600114610e555792610e3b836001959460209487965f92610e4a5750508160011b915f199060031b1c19161790565b87555b01940191019092610d25565b015190508e806103bd565b90601f19831691875f52815f20925f5b818110610ea65750936020936001969387969383889510610e8e575b505050811b018755610e3e565b01515f1960f88460031b161c191690558d8080610e81565b92936020600181928786015181550195019301610e65565b60036020826040600194516001600160a01b0380825116166001600160a01b03198854161787556001600160a01b0384820151166001600160a01b0387890191166001600160a01b03198254161790550151600286015501920192019190610cf9565b015190508a806103bd565b90601f19831691845f52815f20925f5b818110610f775750908460019594939210610f5f575b505050811b019055610cd5565b01515f1960f88460031b161c19169055898080610f52565b92936020600181928786015181550195019301610f3c565b610fa8915060203d6020116106795761066b818361251b565b87610c05565b602090610fb96137d2565b82828701015201610bc6565b7f773a750f000000000000000000000000000000000000000000000000000000005f5260045ffd5b7f1b136079000000000000000000000000000000000000000000000000000000005f5260045ffd5b7fa145c43e000000000000000000000000000000000000000000000000000000005f5260045ffd5b846379c1d89f60e11b5f5260045260245ffd5b34610241575f6003193601126102415761110b61108c7f000000000000000000000000000000000000000000000000000000000000000061502e565b6108ae6110b87f000000000000000000000000000000000000000000000000000000000000000061509f565b611119604051916110ca60208461251b565b5f83525f3681376040519586957f0f00000000000000000000000000000000000000000000000000000000000000875260e0602088015260e08701906125ee565b9085820360408701526125ee565b904660608501523060808501525f60a085015283820360c08501526125bb565b60606003193601126102415761114d6123cb565b602435906001600160a01b03821682036102415761116e9160443591614271565b005b346102415761117e36612454565b5050815f52600260205260405f2060ff600382015416600581101561077957801561181557600281036114e5575081356004811015610241576111c08161264f565b60038103610642576020830135801561064257608084019160026111e48487613011565b905003610fc5576112076111f785613770565b6112013688612854565b90614953565b15610fc55761121990600f8501612ee8565b60108301556011820161122f6040850185612f01565b9067ffffffffffffffff82116105595761124d8261039d85546129b8565b5f90601f83116001146114815761127a92915f91836114765750508160011b915f199060031b1c19161790565b90555b6012820161128e6060850185612bb1565b919061129a8383612f8f565b905f5260205f205f915b83831061141057505050506112bd601383019184613011565b90916112c98282613095565b5f9081526020812092805b83831061132f5750505050507fd3fa0f35ad809781b5c95d9f324b2621475e3d03254a60808cf804b663a704969161131b611315601261046d945b01612a72565b856146b8565b604051918291602083526020830190612c6e565b6113398183612f01565b9067ffffffffffffffff8211610559576113578261049a89546129b8565b5f90601f83116001146113a6579261138c836001959460209487965f9261139b5750508160011b915f199060031b1c19161790565b88555b019501920191936112d4565b013590508e806103bd565b601f19831691885f5260205f20925f5b8181106113f857509360209360019693879693838895106113df575b505050811b01885561138f565b01355f19600384901b60f8161c191690558d80806113d2565b919360206001819287870135815501950192016113b6565b60036060826001600160a01b03611428600195612ffd565b166001600160a01b031986541617855561144460208201612ffd565b6001600160a01b0385870191166001600160a01b031982541617905560408101356002860155019201920191906112a4565b0135905088806103bd565b601f19831691845f5260205f20925f5b8181106114cd57509084600195949392106114b4575b505050811b01905561127d565b01355f19600384901b60f8161c191690558780806114a7565b91936020600181928787013581550195019201611491565b60030361073e57600e810180544210156117e157823560048110156102415761150d8161264f565b600381036106425761152b61152184613770565b6112013687612854565b15610fc5575f61153f9255600f8301612ee8565b602082013560108201556011810161155a6040840184612f01565b9067ffffffffffffffff8211610559576115788261039d85546129b8565b5f90601f831160011461177d576115a592915f91836105d35750508160011b915f199060031b1c19161790565b90555b601281016115b96060840184612bb1565b91906115c58383612f8f565b905f5260205f205f915b8383106117175750505050601381016115eb6080840184613011565b90916115f78282613095565b5f9081526020812092805b8383106116415750505050507fd3fa0f35ad809781b5c95d9f324b2621475e3d03254a60808cf804b663a704969161131b611315601261046d9461130f565b61164b8183612f01565b9067ffffffffffffffff8211610559576116698261049a89546129b8565b5f90601f83116001146116ad579261169e836001959460209487965f9261139b5750508160011b915f199060031b1c19161790565b88555b01950192019193611602565b601f19831691885f5260205f20925f5b8181106116ff57509360209360019693879693838895106116e6575b505050811b0188556116a1565b01355f19600384901b60f8161c191690558d80806116d9565b919360206001819287870135815501950192016116bd565b60036060826001600160a01b0361172f600195612ffd565b166001600160a01b031986541617855561174b60208201612ffd565b6001600160a01b0385870191166001600160a01b031982541617905560408101356002860155019201920191906115cf565b601f19831691845f5260205f20925f5b8181106117c957509084600195949392106117b0575b505050811b0190556115a8565b01355f19600384901b60f8161c191690558680806117a3565b9193602060018192878701358155019501920161178d565b507fd3fa0f35ad809781b5c95d9f324b2621475e3d03254a60808cf804b663a704969161131b611315601261046d9461130f565b836379c1d89f60e11b5f5260045260245ffd5b346102415760406003193601126102415760043560243567ffffffffffffffff81116102415761185c903690600401612556565b61186681516138e0565b5f5b82518110156118b457600190845f526002602052601460405f20016001600160a01b0380611896848861382d565b5116165f5260205260405f20546118ad828561382d565b5201611868565b604051602080825281906108ae908201856125bb565b346102415760406003193601126102415760043567ffffffffffffffff811161024157608060031982360301126102415760243567ffffffffffffffff81116102415760a060031982360301126102415760209161192e91600401906004016139e5565b604051908152f35b346102415760406003193601126102415760043567ffffffffffffffff8111610241576119679036906004016123f5565b60243567ffffffffffffffff8111610241576119879036906004016123f5565b9190926119938261253e565b936119a1604051958661251b565b8285526119ad8361253e565b936119c0601f19602088019601866138c4565b5f5b848110611a5957858760405191829160208301906020845251809152604083019060408160051b85010192915f905b828210611a0057505050500390f35b9193909294603f19908203018252845190602080835192838152019201905f905b808210611a415750505060208060019296019201920185949391926119f1565b90919260208060019286518152019401920190611a21565b611a62826138e0565b611a6c828961382d565b52611a77818861382d565b505f5b828110611a8a57506001016119c2565b6001906001600160a01b03611aa8611aa3858a8a613911565b612ffd565b165f52600360205260405f206001600160a01b03611aca611aa384888a613911565b165f5260205260405f2054611ae982611ae3868d61382d565b5161382d565b5201611a7a565b3461024157611afe36612454565b9190835f52600260205260405f209360ff60038601541660058110156107795780156122be577ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0161073e57831561064257823592600484101561024157611b658461264f565b6002840361064257823593609e1984360301948581121561024157611b8d9085013690612854565b966020830135602089015160018101809111611d81578103610642576060890193611bb8855161491d565b6060810191611bd9611bd4611bcd8585612bb1565b36916127cc565b61491d565b611bef611be585613770565b6112013685612854565b15610fc557611c016040830183612f01565b810195906020818803126102415780359067ffffffffffffffff821161024157019686601f8901121561024157873596611c3a8861253e565b98611c486040519a8b61251b565b888a5260208a01906020829a60051b82010192831161024157602001905b8282106122ae5750505051611c7e611bcd8686612bb1565b90600289510361064257611cb0816040611ca781611c9e611ce7966137fc565b5101519261381d565b51015190612edb565b91611ce1611cd1611cc08c6137fc565b51611cca8d61381d565b51906149d1565b916040611ca781611c9e846137fc565b926149d1565b03612286575f198b019b8b8d11611d8157611d018d61253e565b9c806040519e8f90611d13908261251b565b52611d1d9061253e565b601f19018d5f5b82811061226d5750505060015b8c811015611d95578060051b8b01358c811215610241575f19820191908c01818311611d81578f92600193611d6a611d7a933690612854565b611d74838361382d565b5261382d565b5001611d31565b634e487b7160e01b5f52601160045260245ffd5b5060208d611dc9926001600160a01b0360018a0154169060405180958194829363030232af60e21b84528d60048501613841565b03915afa908115610680575f9161224e575b501561064257611dee90600f8601612ee8565b601084015560118301611e046040830183612f01565b9067ffffffffffffffff821161055957611e228261039d85546129b8565b5f90601f83116001146121ea57611e4f92915f91836121df5750508160011b915f199060031b1c19161790565b90555b60128301611e608383612bb1565b9190611e6c8383612f8f565b905f5260205f205f915b838310612179575050505060138301611e926080830183613011565b9091611e9e8282613095565b5f9081526020812092805b838310612098575050505050611ec291611bcd91612bb1565b9060068101916001600160a01b038354165f5b6002811061203f57505f5b60028110611f9f575050600a5f9201915b60028110611f5c575050505060405191602083019060208452518091526040830191905f5b818110611f4657857ff3b6c524f73df7344d9fcf2f960a57aba7fba7e292d8b79ed03d786f7b2b112f86860387a2005b8251845260209384019390920191600101611f16565b806040611f6b6001938561382d565b51015182611f79838861396a565b5001556040611f88828561382d565b51015182611f96838761396a565b50015501611ef1565b5f611fad828997969761382d565b5112611fbf575b600101939293611ee0565b6001600160a01b03611fd4826004880161394b565b90549060031b1c1690611fe7818961382d565b517f80000000000000000000000000000000000000000000000000000000000000008114611d81576001926120389160405191612023836124ab565b82528560208301525f0360408201528a614ea5565b9050611fb4565b805f6120506001938a98979861382d565b5113612060575b01939293611ed5565b6120936001600160a01b036120788360048a0161394b565b90549060031b1c16848b61208c858d61382d565b5192614d06565b612057565b6120a28183612f01565b9067ffffffffffffffff8211610559576120c08261049a89546129b8565b5f90601f831160011461210f57926120f5836001959460209487965f926121045750508160011b915f199060031b1c19161790565b88555b01950192019193611ea9565b013590505f806103bd565b601f19831691885f5260205f20925f5b8181106121615750936020936001969387969383889510612148575b505050811b0188556120f8565b01355f19600384901b60f8161c191690555f808061213b565b9193602060018192878701358155019501920161211f565b60036060826001600160a01b03612191600195612ffd565b166001600160a01b03198654161785556121ad60208201612ffd565b6001600160a01b0385870191166001600160a01b03198254161790556040810135600286015501920192019190611e76565b013590508a806103bd565b601f19831691845f5260205f20925f5b818110612236575090846001959493921061221d575b505050811b019055611e52565b01355f19600384901b60f8161c19169055898080612210565b919360206001819287870135815501950192016121fa565b612267915060203d6020116106795761066b818361251b565b89611ddb565b602091828261227a6137d2565b92010152018e90611d24565b7f52e4cb1c000000000000000000000000000000000000000000000000000000005f5260045ffd5b8135815260209182019101611c66565b506379c1d89f60e11b5f5260045260245ffd5b346102415760806003193601126102415760243567ffffffffffffffff81116102415760a060031982360301126102415760443567ffffffffffffffff8111610241576123229036906004016123f5565b916064359267ffffffffffffffff84116102415761234761116e943690600401612426565b9390926004016004356130e1565b6080600319360112610241576123696123cb565b6044359067ffffffffffffffff82116102415760806003198336030112610241576064359167ffffffffffffffff83116102415760a06003198436030112610241576020926123bf61192e936024359033614271565b600401906004016139e5565b600435906001600160a01b038216820361024157565b35906001600160a01b038216820361024157565b9181601f840112156102415782359167ffffffffffffffff8311610241576020808501948460051b01011161024157565b9181601f840112156102415782359167ffffffffffffffff8311610241576020838186019501011161024157565b6060600319820112610241576004359160243567ffffffffffffffff81116102415760a0600319828503011261024157600401916044359067ffffffffffffffff8211610241576124a7916004016123f5565b9091565b6060810190811067ffffffffffffffff82111761055957604052565b6080810190811067ffffffffffffffff82111761055957604052565b60a0810190811067ffffffffffffffff82111761055957604052565b6040810190811067ffffffffffffffff82111761055957604052565b90601f601f19910116810190811067ffffffffffffffff82111761055957604052565b67ffffffffffffffff81116105595760051b60200190565b9080601f8301121561024157813561256d8161253e565b9261257b604051948561251b565b81845260208085019260051b82010192831161024157602001905b8282106125a35750505090565b602080916125b0846123e1565b815201910190612596565b90602080835192838152019201905f5b8181106125d85750505090565b82518452602093840193909201916001016125cb565b90601f19601f602080948051918291828752018686015e5f8582860101520116010190565b90602080835192838152019201905f5b8181106126305750505090565b82516001600160a01b0316845260209384019390920191600101612623565b6004111561077957565b80516126648161264f565b825260208101516020830152612689604082015160a0604085015260a08401906125ee565b906060810151918381036060850152602080845192838152019301905f5b8181106127145750505060800151916080818303910152815180825260208201916020808360051b8301019401925f915b8383106126e757505050505090565b909192939460208061270583601f19866001960301875289516125ee565b970193019301919392906126d8565b909193602061275260019287519060406060926001600160a01b0381511683526001600160a01b036020820151166020840152015160408201520190565b95019291016126a7565b67ffffffffffffffff811161055957601f01601f191660200190565b9291926127848261275c565b91612792604051938461251b565b829481845281830111610241578281602093845f960137010152565b9080601f83011215610241578160206127c993359101612778565b90565b9291926127d88261253e565b936127e6604051958661251b565b606060208685815201930282019181831161024157925b82841061280a5750505050565b606084830312610241576020606091604051612825816124ab565b61282e876123e1565b815261283b8388016123e1565b83820152604087013560408201528152019301926127fd565b919060a0838203126102415760405161286c816124e3565b80938035600481101561024157825260208101356020830152604081013567ffffffffffffffff811161024157836128a59183016127ae565b6040830152606081013567ffffffffffffffff811161024157810183601f8201121561024157838160206128db933591016127cc565b606083015260808101359067ffffffffffffffff821161024157019180601f8401121561024157823561290d8161253e565b9361291b604051958661251b565b81855260208086019260051b820101918383116102415760208201905b83821061294a57505050505060800152565b813567ffffffffffffffff81116102415760209161296d878480948801016127ae565b815201910190612938565b90602082549182815201915f5260205f20905f5b8181106129995750505090565b82546001600160a01b031684526020909301926001928301920161298c565b90600182811c921680156129e6575b60208310146129d257565b634e487b7160e01b5f52602260045260245ffd5b91607f16916129c7565b5f92918154916129ff836129b8565b8083529260018116908115612a545750600114612a1b57505050565b5f9081526020812093945091925b838310612a3a575060209250010190565b600181602092949394548385870101520191019190612a29565b9050602094955060ff1991509291921683830152151560051b010190565b908154612a7e8161253e565b92612a8c604051948561251b565b81845260208401905f5260205f205f915b838310612aaa5750505050565b60036020600192604051612abd816124ab565b6001600160a01b0386541681526001600160a01b0385870154168382015260028601546040820152815201920192019190612a9d565b90604051612b00816124e3565b6004819360ff815416612b128161264f565b835260018101546020840152604051612b3981612b3281600286016129f0565b038261251b565b6040840152612b4a60038201612a72565b606084015201908154612b5c8161253e565b92612b6a604051948561251b565b81845260208401905f5260205f205f915b838310612b8c575050505060800152565b600160208192604051612ba381612b3281896129f0565b815201920192019190612b7b565b903590601e1981360301821215610241570180359067ffffffffffffffff82116102415760200191606082023603831361024157565b9035601e198236030181121561024157016020813591019167ffffffffffffffff821161024157813603831361024157565b601f8260209493601f1993818652868601375f8582860101520116010190565b9035601e198236030181121561024157016020813591019167ffffffffffffffff8211610241578160051b3603831361024157565b8035600481101561024157612c828161264f565b825260208101356020830152612caf612c9e6040830183612be7565b60a0604086015260a0850191612c19565b906060810135601e19823603018112156102415781016020813591019267ffffffffffffffff82116102415760608202360384136102415784810360608601528181526020019392905f5b818110612d7a57505050806080612d12920190612c39565b90916080818503910152808352602083019260208260051b82010193835f925b848410612d425750505050505090565b909192939495602080612d6a83601f198660019603018852612d648b88612be7565b90612c19565b9801940194019294939190612d32565b909193946060806001926001600160a01b03612d95896123e1565b1681526001600160a01b03612dac60208a016123e1565b1660208201526040888101359082015201969501929101612cfa565b929190612ddf602091604086526040860190612c6e565b930152565b90816020910312610241575180151581036102415790565b9060808152606067ffffffffffffffff6002612e1b6080850186612978565b948260018201546001600160a01b038116602088015260a01c16604086015201541691015290565b91612e59612e6792606085526060850190612dfc565b908382036020850152612c6e565b906040818303910152828152602081019260208160051b83010193835f91609e1982360301945b848410612e9f575050505050505090565b90919293949596601f19828203018352873587811215610241576020612eca60019387839401612c6e565b990193019401929195949390612e8e565b91908201809211611d8157565b90612ef28161264f565b60ff60ff198354169116179055565b903590601e1981360301821215610241570180359067ffffffffffffffff82116102415760200191813603831361024157565b818110612f3f575050565b5f8155600101612f34565b9190601f8111612f5957505050565b612f83925f5260205f20906020601f840160051c83019310612f85575b601f0160051c0190612f34565b565b9091508190612f76565b9068010000000000000000811161055957815491818155828210612fb257505050565b82600302926003840403611d815781600302916003830403611d81575f5260205f2091820191015b818110612fe5575050565b805f600392555f60018201555f600282015501612fda565b356001600160a01b03811681036102415790565b903590601e1981360301821215610241570180359067ffffffffffffffff821161024157602001918160051b3603831361024157565b61305181546129b8565b908161305b575050565b81601f5f931160011461306c575055565b8183526020832061308891601f0160051c810190600101612f34565b8082528160208120915555565b90680100000000000000008111610559578154918181558282106130b857505050565b5f5260205f2091820191015b8181106130cf575050565b806130db600192613047565b016130c4565b93909195949295845f52600260205260405f2091600383019760ff89541692600584101561077957831561375d575f600385148015613751575b61073e5786359360048510159889610241576131368661264f565b600386146106425761316361317161317893613152368d612854565b9361316a8c60405195868092612978565b038561251b565b3691612778565b918c6143fd565b600f86019460ff86541691610779576001146136aa5787906131998161264f565b600181036135575750610241576131af8361264f565b600183036134c35750506131d56131c63686612854565b6131cf84612af3565b90614651565b15610642575b6131f667ffffffffffffffff600185015460a01c1642612edb565b9485600e8501556102415761320a91612ee8565b60208201356010820155601181016132256040840184612f01565b9067ffffffffffffffff8211610559576132438261039d85546129b8565b5f90601f831160011461345f5761327092915f91836121045750508160011b915f199060031b1c19161790565b90555b601281016132846060840184612bb1565b91906132908383612f8f565b905f5260205f205f915b8383106133f95750505050601301946132b66080830183613011565b90966132c28282613095565b5f9081526020812097805b83831061331c57505050507f44c1980976c3af1eb75b2a3b7d8c7e01f69168c0fe45dd229faf143233722e1793949550600360ff1982541617905561331760405192839283612dc8565b0390a2565b6133268183612f01565b9067ffffffffffffffff82116105595761334b828d61334581546129b8565b90612f4a565b5f90601f831160011461338f5792613380836001959460209487965f926121045750508160011b915f199060031b1c19161790565b8d555b019a01920191986132cd565b601f198316918d5f5260205f20925f5b8181106133e157509360209360019693879693838895106133c8575b505050811b018d55613383565b01355f19600384901b60f8161c191690555f80806133bb565b9193602060018192878701358155019501920161339f565b60036060826001600160a01b03613411600195612ffd565b166001600160a01b031986541617855561342d60208201612ffd565b6001600160a01b0385870191166001600160a01b0319825416179055604081013560028601550192019201919061329a565b601f19831691845f5260205f20925f5b8181106134ab5750908460019594939210613492575b505050811b019055613273565b01355f19600384901b60f8161c191690555f8080613485565b9193602060018192878701358155019501920161346f565b6001600160a01b03600186015416916134e86134df3689612854565b61030687612af3565b15610642576135129260209260405180958194829363030232af60e21b84528c8c60048601612e43565b03915afa908115610680575f91613538575b506131db5763baf3f0f760e01b5f5260045ffd5b613551915060203d6020116106795761066b818361251b565b5f613524565b90506135628161264f565b80613623575086610241576135768361264f565b82610642576135916135883688612854565b6131cf86612af3565b1561359f575b50505b6131db565b6001600160a01b03600186015416916135bb6134df3689612854565b15610642576135e59260209260405180958194829363030232af60e21b84528c8c60048601612e43565b03915afa908115610680575f91613604575b5015610642575f80613597565b61361d915060203d6020116106795761066b818361251b565b5f6135f7565b60029197506136318161264f565b036106425761363f8261264f565b60018214610642575f956136528361264f565b82613673576001600160a01b03600186015416916134e86134df3689612854565b505093505f936136828161264f565b60028103610642576136976131c63686612854565b61359a5763baf3f0f760e01b5f5260045ffd5b505050509193949550506136cc91506131cf6136c63685612854565b91612af3565b156106425761331781613709611315611bcd60607fd3fa0f35ad809781b5c95d9f324b2621475e3d03254a60808cf804b663a70496960184612bb1565b837f44c1980976c3af1eb75b2a3b7d8c7e01f69168c0fe45dd229faf143233722e176040518061373a428683612dc8565b0390a2604051918291602083526020830190612c6e565b50505f6004851461311b565b876379c1d89f60e11b5f5260045260245ffd5b9060405161377d816124c7565b606067ffffffffffffffff6002839560405161379d81612b328185612978565b85528260018201546001600160a01b038116602088015260a01c166040860152015416910152565b91908203918211611d8157565b604051906137df826124e3565b60606080835f81525f602082015282604082015282808201520152565b8051156138095760200190565b634e487b7160e01b5f52603260045260245ffd5b8051600110156138095760400190565b80518210156138095760209160051b010190565b9161385761386592606085526060850190612dfc565b908382036020850152612659565b906040818303910152815180825260208201916020808360051b8301019401925f915b83831061389757505050505090565b90919293946020806138b583601f1986600196030187528951612659565b97019301930191939290613888565b5f5b8281106138d257505050565b6060828201526020016138c6565b906138ea8261253e565b6138f7604051918261251b565b828152601f19613907829461253e565b0190602036910137565b91908110156138095760051b0190565b3567ffffffffffffffff811681036102415790565b359067ffffffffffffffff8216820361024157565b60028210156138095701905f90565b9190811015613809576060020190565b9060028110156138095760011b01905f90565b91906139ab576020816001600160a01b03806001945116166001600160a01b03198554161784550151910155565b634e487b7160e01b5f525f60045260245ffd5b906040516139cb816124ff565b6020600182946001600160a01b0381541684520154910152565b9060026139f28380613011565b905014801590614249575b801561421b575b80156141d2575b61101557602082016001600160a01b03613a2482612ffd565b16156141aa576040830192610e1067ffffffffffffffff613a4486613921565b161061418257823593600485101561024157613a5f8561264f565b600185036106425760208401359081610642576080833603126102415760405191613a89836124c7565b833567ffffffffffffffff811161024157613aa79036908601612556565b8352613ab2856123e1565b6020840152613ac082613936565b6040840152613ae06060850193613ad685613936565b60608201526149ec565b96875f52600260205260ff600360405f2001541660058110156107795761073e57608087016001613b11828a613011565b905003610fc557613b223689612854565b613b2a614a5e565b90613b35838b613011565b156138095780613b4491612f01565b919092613b518a80613011565b9490941561380957610b258e91613b6a613b7297612ffd565b953691612778565b15610fc55760608801926002613b88858b612bb1565b90500361228657895f52600260205260405f2092613ba68880613011565b9067ffffffffffffffff821161055957680100000000000000008211610559578554828755808310614166575b50855f5260205f205f5b83811061414b5750505050613ccf90600185016001600160a01b03613c018c612ffd565b166001600160a01b0319825416178155613c1a88613921565b7fffffffff0000000000000000ffffffffffffffffffffffffffffffffffffffff7bffffffffffffffff000000000000000000000000000000000000000083549260a01b1691161790556002850167ffffffffffffffff613c7a8a613921565b82547fffffffffffffffffffffffffffffffffffffffffffffffff000000000000000016911617905560038501805460ff191660011790556004850180546001600160a01b03191633179055600f8501612ee8565b601083015560118201613ce560408a018a612f01565b9067ffffffffffffffff821161055957613d038261039d85546129b8565b5f90601f83116001146140e757613d3092915f91836121045750508160011b915f199060031b1c19161790565b90555b60128201613d41848a612bb1565b9190613d4d8383612f8f565b905f5260205f205f915b8383106140815750505050613d70601383019189613011565b9091613d7c8282613095565b5f9081526020812092805b838310613fab5750505050505f91600a600683019201925b60028110613f1e575050613db5613de0916139be565b80929060206001916001600160a01b0380825116166001600160a01b03198554161784550151910155565b613dea8480613011565b9190911561380957613e3d906001600160a01b03613e0c8a9798969594612ffd565b165f526003602052613e2486600160405f2001614f59565b5060206001600160a01b03825116910151908633614d06565b6040519260408452613e5360c085019680612c39565b608060408701529687905260e08501965f5b818110613eea575050509467ffffffffffffffff613ecf859482613ec4613ee4966001600160a01b03613eb97f4dd0384c1acc40a5edb69575b4a1caa43c2c2852ef96f7ecfc4a6705ddb8ccc79c9d6123e1565b1660608a0152613936565b166080870152613936565b1660a084015282810360208401523396612c6e565b0390a390565b91965091929394966020806001926001600160a01b03613f098b6123e1565b16815201970191019189969795949392613e65565b80613f838a6040613f5484613f4288613f4e6020613f4860019b613f42858b612bb1565b9061395a565b01612ffd565b95612bb1565b01356001600160a01b0360405192613f6b846124ff565b1682526020820152613f7d838761396a565b9061397d565b613fa5604051613f92816124ff565b5f81525f6020820152613f7d838861396a565b01613d9f565b613fb58183612f01565b9067ffffffffffffffff821161055957613fd38261049a89546129b8565b5f90601f83116001146140175792614008836001959460209487965f926121045750508160011b915f199060031b1c19161790565b88555b01950192019193613d87565b601f19831691885f5260205f20925f5b8181106140695750936020936001969387969383889510614050575b505050811b01885561400b565b01355f19600384901b60f8161c191690555f8080614043565b91936020600181928787013581550195019201614027565b60036060826001600160a01b03614099600195612ffd565b166001600160a01b03198654161785556140b560208201612ffd565b6001600160a01b0385870191166001600160a01b03198254161790556040810135600286015501920192019190613d57565b601f19831691845f5260205f20925f5b818110614133575090846001959493921061411a575b505050811b019055613d33565b01355f19600384901b60f8161c191690555f808061410d565b919360206001819287870135815501950192016140f7565b600190602061415985612ffd565b9401938184015501613bdd565b61417c90875f528360205f209182019101612f34565b5f613bd3565b7fb4e12433000000000000000000000000000000000000000000000000000000005f5260045ffd5b7fea9e70ce000000000000000000000000000000000000000000000000000000005f5260045ffd5b506141dd8280613011565b15613809576141eb90612ffd565b6141f58380613011565b60011015613809576001600160a01b036142126020829301612ffd565b16911614613a0b565b506142268280613011565b600110156138095761424260206001600160a01b039201612ffd565b1615613a04565b506142548280613011565b156138095761426a6001600160a01b0391612ffd565b16156139fd565b9082156143a8576001600160a01b031691821591821561437157813403614349577f8752a472e571a816aea92eec8dae9baf628e840f4929fbcc2d155e6233ff68a7916001600160a01b036020925b1693845f526003835260405f20865f52835260405f206142e1838254612edb565b9055156142f2575b604051908152a3565b6143446040517f23b872dd00000000000000000000000000000000000000000000000000000000848201523360248201523060448201528260648201526064815261433e60848261251b565b86614fc1565b6142e9565b7faa7feadc000000000000000000000000000000000000000000000000000000005f5260045ffd5b34614349577f8752a472e571a816aea92eec8dae9baf628e840f4929fbcc2d155e6233ff68a7916001600160a01b036020926142c0565b7f2c5211c6000000000000000000000000000000000000000000000000000000005f5260045ffd5b805460011015613809575f52600160205f2001905f90565b8054821015613809575f5260205f2001905f90565b919092825f52600260205261441c61441760405f20613770565b6149ec565b9084516144288161264f565b6020860194855193604088019461446586519461445760608c0196875190604051958694602086019889614e26565b03601f19810183528261251b565b51902060405160208101918252604080820152600960608201527f6368616c6c656e676500000000000000000000000000000000000000000000006080820152608081526144b460a08261251b565b5190206144d66144d06144c7868461519b565b909291926151d5565b87614e66565b614647577f19457468657265756d205369676e6564204d6573736167653a0a3332000000005f52601c526145196145136144c785603c5f2061519b565b86614e66565b61463e576146129661460d9661460495604294614534614a5e565b9351926145408461264f565b519151602081519101209051604051614569816144576020820194602086526040830190614dc1565b5190209160405193602085019586527f74875af04779d70f933aef147d5751a32a32b3fa275f5022499f396ea394cf5360408601526145a78161264f565b6060850152608084015260a083015260c082015260c081526145ca60e08261251b565b519020604051917f19010000000000000000000000000000000000000000000000000000000000008352600283015260228201522061519b565b909391936151d5565b614e66565b612f83577f61a44f6e000000000000000000000000000000000000000000000000000000005f5260045ffd5b50505050505050565b5050505050505050565b60405161466e816144576020820194602086526040830190612659565b5190209060405161468f816144576020820194602086526040830190612659565b5190201490565b60048101905b8181106146a7575050565b5f808255600182015560020161469c565b90815f52600260205260405f209060038201600460ff198254161790556002815103610642575f5b6002811061483b5750505f5b600281106148015750505f52600260205260405f2080545f8255806147e7575b505f60018201555f60028201555f6003820155614738600682016147338160048501612f34565b614696565b614744600a8201614696565b5f600e8201555f600f8201555f601082015561476260118201613047565b601281018054905f8155816147ae575b50506013018054905f815581614786575050565b5f5260205f20908101905b81811061479c575050565b806147a8600192613047565b01614791565b81600302916003830403611d81575f5260205f20908101905b8181101561477257805f600392555f60018201555f6002820155016147c7565b6147fb90825f5260205f2090810190612f34565b5f61470c565b806001600160a01b03614816600193856143e8565b90549060031b1c165f526003602052614834848360405f20016150d6565b50016146ec565b8061485261484b6001938561382d565b5186614ea5565b016146e0565b602060405180927f7f7d6ab500000000000000000000000000000000000000000000000000000000825260406004830152816001600160a01b03816148b56148a3604483018a612659565b6003198382030160248401528a612659565b0392165afa5f91816148e0575b506148d557506020809101519101511090565b90505f8092500b1390565b9091506020813d602011614915575b816148fc6020938361251b565b810103126102415751805f0b810361024157905f6148c2565b3d91506148ef565b6002815103610642576001600160a01b036020614948828261493e866137fc565b510151169361381d565b510151160361064257565b60808201916002835151036149ca5761496b826149ec565b915f5b60028110614980575050505050600190565b6149b461498b614a5e565b61499683885161382d565b516001600160a01b036149aa85875161382d565b5116918787614b74565b156149c15760010161496e565b50505050505f90565b5050505f90565b9190915f8382019384129112908015821691151617611d8157565b805190614a586001600160a01b036020830151169167ffffffffffffffff6060816040840151169201511692604051938492614a34602085019760a0895260c0860190612613565b926040850152606084015260808301524660a083015203601f19810183528261251b565b51902090565b6001600160a01b037f000000000000000000000000000000000000000000000000000000000000000016301480614b4b575b15614ab9577f000000000000000000000000000000000000000000000000000000000000000090565b60405160208101907f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f82527f000000000000000000000000000000000000000000000000000000000000000060408201527f000000000000000000000000000000000000000000000000000000000000000060608201524660808201523060a082015260a08152614a5860c08261251b565b507f00000000000000000000000000000000000000000000000000000000000000004614614a90565b926001600160a01b03949193805193614b8c8561264f565b6020820195865194614bd58960408601988951614bbe606089019a6144578c5160405194859360208501978c89614e26565b51902092614bcc888561519b565b909491946151d5565b1699168914614cf8577f19457468657265756d205369676e6564204d6573736167653a0a3332000000005f52601c52876001600160a01b03614c1d6144c787603c5f2061519b565b1614614ceb577fdeeda0875e527d63774890b89d23bff91e0ec84761e45d2b578ee592780095bb8214614cdf576001600160a01b0396614cd0966144c7966042955192614c698461264f565b519151602081519101209051604051614c92816144576020820194602086526040830190614dc1565b519020916040519360208501957fb02e61f8dbbfba070321cdff64845b04358ee41db88ba372ef47e352446f4b9c875260408601526145a78161264f565b1614614cda575f90565b600190565b50505050505050505f90565b5050505050505050600190565b505050505050505050600190565b8315614dbb576001600160a01b03165f52600360205260405f206001600160a01b0383165f528060205260405f2054848110614d8b5784614d46916137c5565b906001600160a01b0384165f5260205260405f20555f5260026020526001600160a01b03601460405f200191165f52602052614d8760405f20918254612edb565b9055565b84907fcf479181000000000000000000000000000000000000000000000000000000005f5260045260245260445ffd5b50505050565b90602080835192838152019201905f5b818110614dde5750505090565b9091926020614e1c60019286519060406060926001600160a01b0381511683526001600160a01b036020820151166020840152015160408201520190565b9401929101614dd1565b939092614e58926127c996948652614e3d8161264f565b6020860152604085015260a0606085015260a08401906125ee565b916080818403910152614dc1565b905f5b82518110156149ca576001600160a01b03614e84828561382d565b51166001600160a01b03831614614e9d57600101614e69565b505050600190565b906040810191825115614f54575f526002602052601460405f20019160208201916001600160a01b0380845116165f528360205260405f2054938415614f4d576001600160a01b0392518086115f14614f4257614f039080966137c5565b908380865116165f5260205260405f205551165f5260036020526001600160a01b038060405f20925116165f52602052614d8760405f20918254612edb565b50614f0385806137c5565b5050505050565b505050565b6001810190825f528160205260405f2054155f146149ca5780546801000000000000000081101561055957614fae614f988260018794018555846143e8565b819391549060031b91821b915f19901b19161790565b905554915f5260205260405f2055600190565b905f602091828151910182855af115610680575f513d61502557506001600160a01b0381163b155b614ff05750565b6001600160a01b03907f5274afe7000000000000000000000000000000000000000000000000000000005f521660045260245ffd5b60011415614fe9565b60ff811461508d5760ff811690601f8211615065576040519161505260408461251b565b6020808452838101919036833783525290565b7fb3512b0c000000000000000000000000000000000000000000000000000000005f5260045ffd5b506040516127c981612b32815f6129f0565b60ff81146150c35760ff811690601f8211615065576040519161505260408461251b565b506040516127c981612b328160016129f0565b906001820191815f528260205260405f20548015155f14615193575f198101818111611d815782545f19810191908211611d815781810361515e575b5050508054801561514a575f19019061512b82826143e8565b8154905f199060031b1b19169055555f526020525f6040812055600190565b634e487b7160e01b5f52603160045260245ffd5b61517e61516e614f9893866143e8565b90549060031b1c928392866143e8565b90555f528360205260405f20555f8080615112565b505050505f90565b81519190604183036151cb576151c49250602082015190606060408401519301515f1a9061529c565b9192909190565b50505f9160029190565b6151de8161264f565b806151e7575050565b6151f08161264f565b60018103615220577ff645eedf000000000000000000000000000000000000000000000000000000005f5260045ffd5b6152298161264f565b6002810361525d57507ffce698f7000000000000000000000000000000000000000000000000000000005f5260045260245ffd5b6003906152698161264f565b146152715750565b7fd78bce0c000000000000000000000000000000000000000000000000000000005f5260045260245ffd5b91907f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a08411615313579160209360809260ff5f9560405194855216868401526040830152606082015282805260015afa15610680575f516001600160a01b0381161561530957905f905f90565b505f906001905f90565b5050505f916003919056fea2646970667358221220a42ae78844a10ed8f2b9c143b570a4d8ecbb65dda4c532c3bdd322fa72ea304064736f6c634300081b0033",
}

// CustodyABI is the input ABI used to generate the binding from.
// Deprecated: Use CustodyMetaData.ABI instead.
var CustodyABI = CustodyMetaData.ABI

// CustodyBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use CustodyMetaData.Bin instead.
var CustodyBin = CustodyMetaData.Bin

// DeployCustody deploys a new Ethereum contract, binding an instance of Custody to it.
func DeployCustody(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Custody, error) {
	parsed, err := CustodyMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(CustodyBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Custody{CustodyCaller: CustodyCaller{contract: contract}, CustodyTransactor: CustodyTransactor{contract: contract}, CustodyFilterer: CustodyFilterer{contract: contract}}, nil
}

// Custody is an auto generated Go binding around an Ethereum contract.
type Custody struct {
	CustodyCaller     // Read-only binding to the contract
	CustodyTransactor // Write-only binding to the contract
	CustodyFilterer   // Log filterer for contract events
}

// CustodyCaller is an auto generated read-only Go binding around an Ethereum contract.
type CustodyCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CustodyTransactor is an auto generated write-only Go binding around an Ethereum contract.
type CustodyTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CustodyFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type CustodyFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CustodySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type CustodySession struct {
	Contract     *Custody          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// CustodyCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type CustodyCallerSession struct {
	Contract *CustodyCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// CustodyTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type CustodyTransactorSession struct {
	Contract     *CustodyTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// CustodyRaw is an auto generated low-level Go binding around an Ethereum contract.
type CustodyRaw struct {
	Contract *Custody // Generic contract binding to access the raw methods on
}

// CustodyCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type CustodyCallerRaw struct {
	Contract *CustodyCaller // Generic read-only contract binding to access the raw methods on
}

// CustodyTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type CustodyTransactorRaw struct {
	Contract *CustodyTransactor // Generic write-only contract binding to access the raw methods on
}

// NewCustody creates a new instance of Custody, bound to a specific deployed contract.
func NewCustody(address common.Address, backend bind.ContractBackend) (*Custody, error) {
	contract, err := bindCustody(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Custody{CustodyCaller: CustodyCaller{contract: contract}, CustodyTransactor: CustodyTransactor{contract: contract}, CustodyFilterer: CustodyFilterer{contract: contract}}, nil
}

// NewCustodyCaller creates a new read-only instance of Custody, bound to a specific deployed contract.
func NewCustodyCaller(address common.Address, caller bind.ContractCaller) (*CustodyCaller, error) {
	contract, err := bindCustody(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &CustodyCaller{contract: contract}, nil
}

// NewCustodyTransactor creates a new write-only instance of Custody, bound to a specific deployed contract.
func NewCustodyTransactor(address common.Address, transactor bind.ContractTransactor) (*CustodyTransactor, error) {
	contract, err := bindCustody(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &CustodyTransactor{contract: contract}, nil
}

// NewCustodyFilterer creates a new log filterer instance of Custody, bound to a specific deployed contract.
func NewCustodyFilterer(address common.Address, filterer bind.ContractFilterer) (*CustodyFilterer, error) {
	contract, err := bindCustody(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &CustodyFilterer{contract: contract}, nil
}

// bindCustody binds a generic wrapper to an already deployed contract.
func bindCustody(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := CustodyMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Custody *CustodyRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Custody.Contract.CustodyCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Custody *CustodyRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Custody.Contract.CustodyTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Custody *CustodyRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Custody.Contract.CustodyTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Custody *CustodyCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Custody.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Custody *CustodyTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Custody.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Custody *CustodyTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Custody.Contract.contract.Transact(opts, method, params...)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_Custody *CustodyCaller) Eip712Domain(opts *bind.CallOpts) (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	var out []interface{}
	err := _Custody.contract.Call(opts, &out, "eip712Domain")

	outstruct := new(struct {
		Fields            [1]byte
		Name              string
		Version           string
		ChainId           *big.Int
		VerifyingContract common.Address
		Salt              [32]byte
		Extensions        []*big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Fields = *abi.ConvertType(out[0], new([1]byte)).(*[1]byte)
	outstruct.Name = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.Version = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.ChainId = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.VerifyingContract = *abi.ConvertType(out[4], new(common.Address)).(*common.Address)
	outstruct.Salt = *abi.ConvertType(out[5], new([32]byte)).(*[32]byte)
	outstruct.Extensions = *abi.ConvertType(out[6], new([]*big.Int)).(*[]*big.Int)

	return *outstruct, err

}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_Custody *CustodySession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _Custody.Contract.Eip712Domain(&_Custody.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_Custody *CustodyCallerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _Custody.Contract.Eip712Domain(&_Custody.CallOpts)
}

// GetAccountsBalances is a free data retrieval call binding the contract method 0x2f33c4d6.
//
// Solidity: function getAccountsBalances(address[] accounts, address[] tokens) view returns(uint256[][])
func (_Custody *CustodyCaller) GetAccountsBalances(opts *bind.CallOpts, accounts []common.Address, tokens []common.Address) ([][]*big.Int, error) {
	var out []interface{}
	err := _Custody.contract.Call(opts, &out, "getAccountsBalances", accounts, tokens)

	if err != nil {
		return *new([][]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([][]*big.Int)).(*[][]*big.Int)

	return out0, err

}

// GetAccountsBalances is a free data retrieval call binding the contract method 0x2f33c4d6.
//
// Solidity: function getAccountsBalances(address[] accounts, address[] tokens) view returns(uint256[][])
func (_Custody *CustodySession) GetAccountsBalances(accounts []common.Address, tokens []common.Address) ([][]*big.Int, error) {
	return _Custody.Contract.GetAccountsBalances(&_Custody.CallOpts, accounts, tokens)
}

// GetAccountsBalances is a free data retrieval call binding the contract method 0x2f33c4d6.
//
// Solidity: function getAccountsBalances(address[] accounts, address[] tokens) view returns(uint256[][])
func (_Custody *CustodyCallerSession) GetAccountsBalances(accounts []common.Address, tokens []common.Address) ([][]*big.Int, error) {
	return _Custody.Contract.GetAccountsBalances(&_Custody.CallOpts, accounts, tokens)
}

// GetChannelBalances is a free data retrieval call binding the contract method 0x5a9eb80e.
//
// Solidity: function getChannelBalances(bytes32 channelId, address[] tokens) view returns(uint256[] balances)
func (_Custody *CustodyCaller) GetChannelBalances(opts *bind.CallOpts, channelId [32]byte, tokens []common.Address) ([]*big.Int, error) {
	var out []interface{}
	err := _Custody.contract.Call(opts, &out, "getChannelBalances", channelId, tokens)

	if err != nil {
		return *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]*big.Int)).(*[]*big.Int)

	return out0, err

}

// GetChannelBalances is a free data retrieval call binding the contract method 0x5a9eb80e.
//
// Solidity: function getChannelBalances(bytes32 channelId, address[] tokens) view returns(uint256[] balances)
func (_Custody *CustodySession) GetChannelBalances(channelId [32]byte, tokens []common.Address) ([]*big.Int, error) {
	return _Custody.Contract.GetChannelBalances(&_Custody.CallOpts, channelId, tokens)
}

// GetChannelBalances is a free data retrieval call binding the contract method 0x5a9eb80e.
//
// Solidity: function getChannelBalances(bytes32 channelId, address[] tokens) view returns(uint256[] balances)
func (_Custody *CustodyCallerSession) GetChannelBalances(channelId [32]byte, tokens []common.Address) ([]*big.Int, error) {
	return _Custody.Contract.GetChannelBalances(&_Custody.CallOpts, channelId, tokens)
}

// GetChannelData is a free data retrieval call binding the contract method 0xe617208c.
//
// Solidity: function getChannelData(bytes32 channelId) view returns((address[],address,uint64,uint64) channel, uint8 status, address[] wallets, uint256 challengeExpiry, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) lastValidState)
func (_Custody *CustodyCaller) GetChannelData(opts *bind.CallOpts, channelId [32]byte) (struct {
	Channel         Channel
	Status          uint8
	Wallets         []common.Address
	ChallengeExpiry *big.Int
	LastValidState  State
}, error) {
	var out []interface{}
	err := _Custody.contract.Call(opts, &out, "getChannelData", channelId)

	outstruct := new(struct {
		Channel         Channel
		Status          uint8
		Wallets         []common.Address
		ChallengeExpiry *big.Int
		LastValidState  State
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Channel = *abi.ConvertType(out[0], new(Channel)).(*Channel)
	outstruct.Status = *abi.ConvertType(out[1], new(uint8)).(*uint8)
	outstruct.Wallets = *abi.ConvertType(out[2], new([]common.Address)).(*[]common.Address)
	outstruct.ChallengeExpiry = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.LastValidState = *abi.ConvertType(out[4], new(State)).(*State)

	return *outstruct, err

}

// GetChannelData is a free data retrieval call binding the contract method 0xe617208c.
//
// Solidity: function getChannelData(bytes32 channelId) view returns((address[],address,uint64,uint64) channel, uint8 status, address[] wallets, uint256 challengeExpiry, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) lastValidState)
func (_Custody *CustodySession) GetChannelData(channelId [32]byte) (struct {
	Channel         Channel
	Status          uint8
	Wallets         []common.Address
	ChallengeExpiry *big.Int
	LastValidState  State
}, error) {
	return _Custody.Contract.GetChannelData(&_Custody.CallOpts, channelId)
}

// GetChannelData is a free data retrieval call binding the contract method 0xe617208c.
//
// Solidity: function getChannelData(bytes32 channelId) view returns((address[],address,uint64,uint64) channel, uint8 status, address[] wallets, uint256 challengeExpiry, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) lastValidState)
func (_Custody *CustodyCallerSession) GetChannelData(channelId [32]byte) (struct {
	Channel         Channel
	Status          uint8
	Wallets         []common.Address
	ChallengeExpiry *big.Int
	LastValidState  State
}, error) {
	return _Custody.Contract.GetChannelData(&_Custody.CallOpts, channelId)
}

// GetOpenChannels is a free data retrieval call binding the contract method 0xd710e92f.
//
// Solidity: function getOpenChannels(address[] accounts) view returns(bytes32[][])
func (_Custody *CustodyCaller) GetOpenChannels(opts *bind.CallOpts, accounts []common.Address) ([][][32]byte, error) {
	var out []interface{}
	err := _Custody.contract.Call(opts, &out, "getOpenChannels", accounts)

	if err != nil {
		return *new([][][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][][32]byte)).(*[][][32]byte)

	return out0, err

}

// GetOpenChannels is a free data retrieval call binding the contract method 0xd710e92f.
//
// Solidity: function getOpenChannels(address[] accounts) view returns(bytes32[][])
func (_Custody *CustodySession) GetOpenChannels(accounts []common.Address) ([][][32]byte, error) {
	return _Custody.Contract.GetOpenChannels(&_Custody.CallOpts, accounts)
}

// GetOpenChannels is a free data retrieval call binding the contract method 0xd710e92f.
//
// Solidity: function getOpenChannels(address[] accounts) view returns(bytes32[][])
func (_Custody *CustodyCallerSession) GetOpenChannels(accounts []common.Address) ([][][32]byte, error) {
	return _Custody.Contract.GetOpenChannels(&_Custody.CallOpts, accounts)
}

// Challenge is a paid mutator transaction binding the contract method 0x1474e410.
//
// Solidity: function challenge(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],bytes[])[] proofs, bytes challengerSig) returns()
func (_Custody *CustodyTransactor) Challenge(opts *bind.TransactOpts, channelId [32]byte, candidate State, proofs []State, challengerSig []byte) (*types.Transaction, error) {
	return _Custody.contract.Transact(opts, "challenge", channelId, candidate, proofs, challengerSig)
}

// Challenge is a paid mutator transaction binding the contract method 0x1474e410.
//
// Solidity: function challenge(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],bytes[])[] proofs, bytes challengerSig) returns()
func (_Custody *CustodySession) Challenge(channelId [32]byte, candidate State, proofs []State, challengerSig []byte) (*types.Transaction, error) {
	return _Custody.Contract.Challenge(&_Custody.TransactOpts, channelId, candidate, proofs, challengerSig)
}

// Challenge is a paid mutator transaction binding the contract method 0x1474e410.
//
// Solidity: function challenge(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],bytes[])[] proofs, bytes challengerSig) returns()
func (_Custody *CustodyTransactorSession) Challenge(channelId [32]byte, candidate State, proofs []State, challengerSig []byte) (*types.Transaction, error) {
	return _Custody.Contract.Challenge(&_Custody.TransactOpts, channelId, candidate, proofs, challengerSig)
}

// Checkpoint is a paid mutator transaction binding the contract method 0xecf668fd.
//
// Solidity: function checkpoint(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],bytes[])[] proofs) returns()
func (_Custody *CustodyTransactor) Checkpoint(opts *bind.TransactOpts, channelId [32]byte, candidate State, proofs []State) (*types.Transaction, error) {
	return _Custody.contract.Transact(opts, "checkpoint", channelId, candidate, proofs)
}

// Checkpoint is a paid mutator transaction binding the contract method 0xecf668fd.
//
// Solidity: function checkpoint(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],bytes[])[] proofs) returns()
func (_Custody *CustodySession) Checkpoint(channelId [32]byte, candidate State, proofs []State) (*types.Transaction, error) {
	return _Custody.Contract.Checkpoint(&_Custody.TransactOpts, channelId, candidate, proofs)
}

// Checkpoint is a paid mutator transaction binding the contract method 0xecf668fd.
//
// Solidity: function checkpoint(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],bytes[])[] proofs) returns()
func (_Custody *CustodyTransactorSession) Checkpoint(channelId [32]byte, candidate State, proofs []State) (*types.Transaction, error) {
	return _Custody.Contract.Checkpoint(&_Custody.TransactOpts, channelId, candidate, proofs)
}

// Close is a paid mutator transaction binding the contract method 0x7f9ebbd7.
//
// Solidity: function close(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],bytes[])[] ) returns()
func (_Custody *CustodyTransactor) Close(opts *bind.TransactOpts, channelId [32]byte, candidate State, arg2 []State) (*types.Transaction, error) {
	return _Custody.contract.Transact(opts, "close", channelId, candidate, arg2)
}

// Close is a paid mutator transaction binding the contract method 0x7f9ebbd7.
//
// Solidity: function close(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],bytes[])[] ) returns()
func (_Custody *CustodySession) Close(channelId [32]byte, candidate State, arg2 []State) (*types.Transaction, error) {
	return _Custody.Contract.Close(&_Custody.TransactOpts, channelId, candidate, arg2)
}

// Close is a paid mutator transaction binding the contract method 0x7f9ebbd7.
//
// Solidity: function close(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],bytes[])[] ) returns()
func (_Custody *CustodyTransactorSession) Close(channelId [32]byte, candidate State, arg2 []State) (*types.Transaction, error) {
	return _Custody.Contract.Close(&_Custody.TransactOpts, channelId, candidate, arg2)
}

// Create is a paid mutator transaction binding the contract method 0x4a7e7798.
//
// Solidity: function create((address[],address,uint64,uint64) ch, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) initial) returns(bytes32 channelId)
func (_Custody *CustodyTransactor) Create(opts *bind.TransactOpts, ch Channel, initial State) (*types.Transaction, error) {
	return _Custody.contract.Transact(opts, "create", ch, initial)
}

// Create is a paid mutator transaction binding the contract method 0x4a7e7798.
//
// Solidity: function create((address[],address,uint64,uint64) ch, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) initial) returns(bytes32 channelId)
func (_Custody *CustodySession) Create(ch Channel, initial State) (*types.Transaction, error) {
	return _Custody.Contract.Create(&_Custody.TransactOpts, ch, initial)
}

// Create is a paid mutator transaction binding the contract method 0x4a7e7798.
//
// Solidity: function create((address[],address,uint64,uint64) ch, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) initial) returns(bytes32 channelId)
func (_Custody *CustodyTransactorSession) Create(ch Channel, initial State) (*types.Transaction, error) {
	return _Custody.Contract.Create(&_Custody.TransactOpts, ch, initial)
}

// Deposit is a paid mutator transaction binding the contract method 0x8340f549.
//
// Solidity: function deposit(address account, address token, uint256 amount) payable returns()
func (_Custody *CustodyTransactor) Deposit(opts *bind.TransactOpts, account common.Address, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Custody.contract.Transact(opts, "deposit", account, token, amount)
}

// Deposit is a paid mutator transaction binding the contract method 0x8340f549.
//
// Solidity: function deposit(address account, address token, uint256 amount) payable returns()
func (_Custody *CustodySession) Deposit(account common.Address, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Custody.Contract.Deposit(&_Custody.TransactOpts, account, token, amount)
}

// Deposit is a paid mutator transaction binding the contract method 0x8340f549.
//
// Solidity: function deposit(address account, address token, uint256 amount) payable returns()
func (_Custody *CustodyTransactorSession) Deposit(account common.Address, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Custody.Contract.Deposit(&_Custody.TransactOpts, account, token, amount)
}

// DepositAndCreate is a paid mutator transaction binding the contract method 0x00e2bb2c.
//
// Solidity: function depositAndCreate(address token, uint256 amount, (address[],address,uint64,uint64) ch, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) initial) payable returns(bytes32)
func (_Custody *CustodyTransactor) DepositAndCreate(opts *bind.TransactOpts, token common.Address, amount *big.Int, ch Channel, initial State) (*types.Transaction, error) {
	return _Custody.contract.Transact(opts, "depositAndCreate", token, amount, ch, initial)
}

// DepositAndCreate is a paid mutator transaction binding the contract method 0x00e2bb2c.
//
// Solidity: function depositAndCreate(address token, uint256 amount, (address[],address,uint64,uint64) ch, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) initial) payable returns(bytes32)
func (_Custody *CustodySession) DepositAndCreate(token common.Address, amount *big.Int, ch Channel, initial State) (*types.Transaction, error) {
	return _Custody.Contract.DepositAndCreate(&_Custody.TransactOpts, token, amount, ch, initial)
}

// DepositAndCreate is a paid mutator transaction binding the contract method 0x00e2bb2c.
//
// Solidity: function depositAndCreate(address token, uint256 amount, (address[],address,uint64,uint64) ch, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) initial) payable returns(bytes32)
func (_Custody *CustodyTransactorSession) DepositAndCreate(token common.Address, amount *big.Int, ch Channel, initial State) (*types.Transaction, error) {
	return _Custody.Contract.DepositAndCreate(&_Custody.TransactOpts, token, amount, ch, initial)
}

// Join is a paid mutator transaction binding the contract method 0xbab3290a.
//
// Solidity: function join(bytes32 channelId, uint256 index, bytes sig) returns(bytes32)
func (_Custody *CustodyTransactor) Join(opts *bind.TransactOpts, channelId [32]byte, index *big.Int, sig []byte) (*types.Transaction, error) {
	return _Custody.contract.Transact(opts, "join", channelId, index, sig)
}

// Join is a paid mutator transaction binding the contract method 0xbab3290a.
//
// Solidity: function join(bytes32 channelId, uint256 index, bytes sig) returns(bytes32)
func (_Custody *CustodySession) Join(channelId [32]byte, index *big.Int, sig []byte) (*types.Transaction, error) {
	return _Custody.Contract.Join(&_Custody.TransactOpts, channelId, index, sig)
}

// Join is a paid mutator transaction binding the contract method 0xbab3290a.
//
// Solidity: function join(bytes32 channelId, uint256 index, bytes sig) returns(bytes32)
func (_Custody *CustodyTransactorSession) Join(channelId [32]byte, index *big.Int, sig []byte) (*types.Transaction, error) {
	return _Custody.Contract.Join(&_Custody.TransactOpts, channelId, index, sig)
}

// Resize is a paid mutator transaction binding the contract method 0x183b4998.
//
// Solidity: function resize(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],bytes[])[] proofs) returns()
func (_Custody *CustodyTransactor) Resize(opts *bind.TransactOpts, channelId [32]byte, candidate State, proofs []State) (*types.Transaction, error) {
	return _Custody.contract.Transact(opts, "resize", channelId, candidate, proofs)
}

// Resize is a paid mutator transaction binding the contract method 0x183b4998.
//
// Solidity: function resize(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],bytes[])[] proofs) returns()
func (_Custody *CustodySession) Resize(channelId [32]byte, candidate State, proofs []State) (*types.Transaction, error) {
	return _Custody.Contract.Resize(&_Custody.TransactOpts, channelId, candidate, proofs)
}

// Resize is a paid mutator transaction binding the contract method 0x183b4998.
//
// Solidity: function resize(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],bytes[])[] proofs) returns()
func (_Custody *CustodyTransactorSession) Resize(channelId [32]byte, candidate State, proofs []State) (*types.Transaction, error) {
	return _Custody.Contract.Resize(&_Custody.TransactOpts, channelId, candidate, proofs)
}

// Withdraw is a paid mutator transaction binding the contract method 0xf3fef3a3.
//
// Solidity: function withdraw(address token, uint256 amount) returns()
func (_Custody *CustodyTransactor) Withdraw(opts *bind.TransactOpts, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Custody.contract.Transact(opts, "withdraw", token, amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0xf3fef3a3.
//
// Solidity: function withdraw(address token, uint256 amount) returns()
func (_Custody *CustodySession) Withdraw(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Custody.Contract.Withdraw(&_Custody.TransactOpts, token, amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0xf3fef3a3.
//
// Solidity: function withdraw(address token, uint256 amount) returns()
func (_Custody *CustodyTransactorSession) Withdraw(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Custody.Contract.Withdraw(&_Custody.TransactOpts, token, amount)
}

// CustodyChallengedIterator is returned from FilterChallenged and is used to iterate over the raw logs and unpacked data for Challenged events raised by the Custody contract.
type CustodyChallengedIterator struct {
	Event *CustodyChallenged // Event containing the contract specifics and raw log

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
func (it *CustodyChallengedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CustodyChallenged)
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
		it.Event = new(CustodyChallenged)
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
func (it *CustodyChallengedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CustodyChallengedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CustodyChallenged represents a Challenged event raised by the Custody contract.
type CustodyChallenged struct {
	ChannelId  [32]byte
	State      State
	Expiration *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterChallenged is a free log retrieval operation binding the contract event 0x44c1980976c3af1eb75b2a3b7d8c7e01f69168c0fe45dd229faf143233722e17.
//
// Solidity: event Challenged(bytes32 indexed channelId, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) state, uint256 expiration)
func (_Custody *CustodyFilterer) FilterChallenged(opts *bind.FilterOpts, channelId [][32]byte) (*CustodyChallengedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _Custody.contract.FilterLogs(opts, "Challenged", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &CustodyChallengedIterator{contract: _Custody.contract, event: "Challenged", logs: logs, sub: sub}, nil
}

// WatchChallenged is a free log subscription operation binding the contract event 0x44c1980976c3af1eb75b2a3b7d8c7e01f69168c0fe45dd229faf143233722e17.
//
// Solidity: event Challenged(bytes32 indexed channelId, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) state, uint256 expiration)
func (_Custody *CustodyFilterer) WatchChallenged(opts *bind.WatchOpts, sink chan<- *CustodyChallenged, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _Custody.contract.WatchLogs(opts, "Challenged", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CustodyChallenged)
				if err := _Custody.contract.UnpackLog(event, "Challenged", log); err != nil {
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

// ParseChallenged is a log parse operation binding the contract event 0x44c1980976c3af1eb75b2a3b7d8c7e01f69168c0fe45dd229faf143233722e17.
//
// Solidity: event Challenged(bytes32 indexed channelId, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) state, uint256 expiration)
func (_Custody *CustodyFilterer) ParseChallenged(log types.Log) (*CustodyChallenged, error) {
	event := new(CustodyChallenged)
	if err := _Custody.contract.UnpackLog(event, "Challenged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CustodyCheckpointedIterator is returned from FilterCheckpointed and is used to iterate over the raw logs and unpacked data for Checkpointed events raised by the Custody contract.
type CustodyCheckpointedIterator struct {
	Event *CustodyCheckpointed // Event containing the contract specifics and raw log

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
func (it *CustodyCheckpointedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CustodyCheckpointed)
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
		it.Event = new(CustodyCheckpointed)
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
func (it *CustodyCheckpointedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CustodyCheckpointedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CustodyCheckpointed represents a Checkpointed event raised by the Custody contract.
type CustodyCheckpointed struct {
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterCheckpointed is a free log retrieval operation binding the contract event 0x8cade4fe25d72146dc0dbe08ea2712bdcca7e2c996e2dce1e69f20e30ee1c5c3.
//
// Solidity: event Checkpointed(bytes32 indexed channelId, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) state)
func (_Custody *CustodyFilterer) FilterCheckpointed(opts *bind.FilterOpts, channelId [][32]byte) (*CustodyCheckpointedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _Custody.contract.FilterLogs(opts, "Checkpointed", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &CustodyCheckpointedIterator{contract: _Custody.contract, event: "Checkpointed", logs: logs, sub: sub}, nil
}

// WatchCheckpointed is a free log subscription operation binding the contract event 0x8cade4fe25d72146dc0dbe08ea2712bdcca7e2c996e2dce1e69f20e30ee1c5c3.
//
// Solidity: event Checkpointed(bytes32 indexed channelId, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) state)
func (_Custody *CustodyFilterer) WatchCheckpointed(opts *bind.WatchOpts, sink chan<- *CustodyCheckpointed, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _Custody.contract.WatchLogs(opts, "Checkpointed", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CustodyCheckpointed)
				if err := _Custody.contract.UnpackLog(event, "Checkpointed", log); err != nil {
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

// ParseCheckpointed is a log parse operation binding the contract event 0x8cade4fe25d72146dc0dbe08ea2712bdcca7e2c996e2dce1e69f20e30ee1c5c3.
//
// Solidity: event Checkpointed(bytes32 indexed channelId, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) state)
func (_Custody *CustodyFilterer) ParseCheckpointed(log types.Log) (*CustodyCheckpointed, error) {
	event := new(CustodyCheckpointed)
	if err := _Custody.contract.UnpackLog(event, "Checkpointed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CustodyClosedIterator is returned from FilterClosed and is used to iterate over the raw logs and unpacked data for Closed events raised by the Custody contract.
type CustodyClosedIterator struct {
	Event *CustodyClosed // Event containing the contract specifics and raw log

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
func (it *CustodyClosedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CustodyClosed)
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
		it.Event = new(CustodyClosed)
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
func (it *CustodyClosedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CustodyClosedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CustodyClosed represents a Closed event raised by the Custody contract.
type CustodyClosed struct {
	ChannelId  [32]byte
	FinalState State
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterClosed is a free log retrieval operation binding the contract event 0xd3fa0f35ad809781b5c95d9f324b2621475e3d03254a60808cf804b663a70496.
//
// Solidity: event Closed(bytes32 indexed channelId, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) finalState)
func (_Custody *CustodyFilterer) FilterClosed(opts *bind.FilterOpts, channelId [][32]byte) (*CustodyClosedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _Custody.contract.FilterLogs(opts, "Closed", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &CustodyClosedIterator{contract: _Custody.contract, event: "Closed", logs: logs, sub: sub}, nil
}

// WatchClosed is a free log subscription operation binding the contract event 0xd3fa0f35ad809781b5c95d9f324b2621475e3d03254a60808cf804b663a70496.
//
// Solidity: event Closed(bytes32 indexed channelId, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) finalState)
func (_Custody *CustodyFilterer) WatchClosed(opts *bind.WatchOpts, sink chan<- *CustodyClosed, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _Custody.contract.WatchLogs(opts, "Closed", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CustodyClosed)
				if err := _Custody.contract.UnpackLog(event, "Closed", log); err != nil {
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

// ParseClosed is a log parse operation binding the contract event 0xd3fa0f35ad809781b5c95d9f324b2621475e3d03254a60808cf804b663a70496.
//
// Solidity: event Closed(bytes32 indexed channelId, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) finalState)
func (_Custody *CustodyFilterer) ParseClosed(log types.Log) (*CustodyClosed, error) {
	event := new(CustodyClosed)
	if err := _Custody.contract.UnpackLog(event, "Closed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CustodyCreatedIterator is returned from FilterCreated and is used to iterate over the raw logs and unpacked data for Created events raised by the Custody contract.
type CustodyCreatedIterator struct {
	Event *CustodyCreated // Event containing the contract specifics and raw log

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
func (it *CustodyCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CustodyCreated)
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
		it.Event = new(CustodyCreated)
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
func (it *CustodyCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CustodyCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CustodyCreated represents a Created event raised by the Custody contract.
type CustodyCreated struct {
	ChannelId [32]byte
	Wallet    common.Address
	Channel   Channel
	Initial   State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterCreated is a free log retrieval operation binding the contract event 0x4dd0384c1acc40a5edb69575b4a1caa43c2c2852ef96f7ecfc4a6705ddb8ccc7.
//
// Solidity: event Created(bytes32 indexed channelId, address indexed wallet, (address[],address,uint64,uint64) channel, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) initial)
func (_Custody *CustodyFilterer) FilterCreated(opts *bind.FilterOpts, channelId [][32]byte, wallet []common.Address) (*CustodyCreatedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}
	var walletRule []interface{}
	for _, walletItem := range wallet {
		walletRule = append(walletRule, walletItem)
	}

	logs, sub, err := _Custody.contract.FilterLogs(opts, "Created", channelIdRule, walletRule)
	if err != nil {
		return nil, err
	}
	return &CustodyCreatedIterator{contract: _Custody.contract, event: "Created", logs: logs, sub: sub}, nil
}

// WatchCreated is a free log subscription operation binding the contract event 0x4dd0384c1acc40a5edb69575b4a1caa43c2c2852ef96f7ecfc4a6705ddb8ccc7.
//
// Solidity: event Created(bytes32 indexed channelId, address indexed wallet, (address[],address,uint64,uint64) channel, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) initial)
func (_Custody *CustodyFilterer) WatchCreated(opts *bind.WatchOpts, sink chan<- *CustodyCreated, channelId [][32]byte, wallet []common.Address) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}
	var walletRule []interface{}
	for _, walletItem := range wallet {
		walletRule = append(walletRule, walletItem)
	}

	logs, sub, err := _Custody.contract.WatchLogs(opts, "Created", channelIdRule, walletRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CustodyCreated)
				if err := _Custody.contract.UnpackLog(event, "Created", log); err != nil {
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

// ParseCreated is a log parse operation binding the contract event 0x4dd0384c1acc40a5edb69575b4a1caa43c2c2852ef96f7ecfc4a6705ddb8ccc7.
//
// Solidity: event Created(bytes32 indexed channelId, address indexed wallet, (address[],address,uint64,uint64) channel, (uint8,uint256,bytes,(address,address,uint256)[],bytes[]) initial)
func (_Custody *CustodyFilterer) ParseCreated(log types.Log) (*CustodyCreated, error) {
	event := new(CustodyCreated)
	if err := _Custody.contract.UnpackLog(event, "Created", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CustodyDepositedIterator is returned from FilterDeposited and is used to iterate over the raw logs and unpacked data for Deposited events raised by the Custody contract.
type CustodyDepositedIterator struct {
	Event *CustodyDeposited // Event containing the contract specifics and raw log

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
func (it *CustodyDepositedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CustodyDeposited)
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
		it.Event = new(CustodyDeposited)
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
func (it *CustodyDepositedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CustodyDepositedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CustodyDeposited represents a Deposited event raised by the Custody contract.
type CustodyDeposited struct {
	Wallet common.Address
	Token  common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDeposited is a free log retrieval operation binding the contract event 0x8752a472e571a816aea92eec8dae9baf628e840f4929fbcc2d155e6233ff68a7.
//
// Solidity: event Deposited(address indexed wallet, address indexed token, uint256 amount)
func (_Custody *CustodyFilterer) FilterDeposited(opts *bind.FilterOpts, wallet []common.Address, token []common.Address) (*CustodyDepositedIterator, error) {

	var walletRule []interface{}
	for _, walletItem := range wallet {
		walletRule = append(walletRule, walletItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Custody.contract.FilterLogs(opts, "Deposited", walletRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return &CustodyDepositedIterator{contract: _Custody.contract, event: "Deposited", logs: logs, sub: sub}, nil
}

// WatchDeposited is a free log subscription operation binding the contract event 0x8752a472e571a816aea92eec8dae9baf628e840f4929fbcc2d155e6233ff68a7.
//
// Solidity: event Deposited(address indexed wallet, address indexed token, uint256 amount)
func (_Custody *CustodyFilterer) WatchDeposited(opts *bind.WatchOpts, sink chan<- *CustodyDeposited, wallet []common.Address, token []common.Address) (event.Subscription, error) {

	var walletRule []interface{}
	for _, walletItem := range wallet {
		walletRule = append(walletRule, walletItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Custody.contract.WatchLogs(opts, "Deposited", walletRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CustodyDeposited)
				if err := _Custody.contract.UnpackLog(event, "Deposited", log); err != nil {
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
func (_Custody *CustodyFilterer) ParseDeposited(log types.Log) (*CustodyDeposited, error) {
	event := new(CustodyDeposited)
	if err := _Custody.contract.UnpackLog(event, "Deposited", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CustodyEIP712DomainChangedIterator is returned from FilterEIP712DomainChanged and is used to iterate over the raw logs and unpacked data for EIP712DomainChanged events raised by the Custody contract.
type CustodyEIP712DomainChangedIterator struct {
	Event *CustodyEIP712DomainChanged // Event containing the contract specifics and raw log

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
func (it *CustodyEIP712DomainChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CustodyEIP712DomainChanged)
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
		it.Event = new(CustodyEIP712DomainChanged)
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
func (it *CustodyEIP712DomainChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CustodyEIP712DomainChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CustodyEIP712DomainChanged represents a EIP712DomainChanged event raised by the Custody contract.
type CustodyEIP712DomainChanged struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEIP712DomainChanged is a free log retrieval operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_Custody *CustodyFilterer) FilterEIP712DomainChanged(opts *bind.FilterOpts) (*CustodyEIP712DomainChangedIterator, error) {

	logs, sub, err := _Custody.contract.FilterLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return &CustodyEIP712DomainChangedIterator{contract: _Custody.contract, event: "EIP712DomainChanged", logs: logs, sub: sub}, nil
}

// WatchEIP712DomainChanged is a free log subscription operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_Custody *CustodyFilterer) WatchEIP712DomainChanged(opts *bind.WatchOpts, sink chan<- *CustodyEIP712DomainChanged) (event.Subscription, error) {

	logs, sub, err := _Custody.contract.WatchLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CustodyEIP712DomainChanged)
				if err := _Custody.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
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

// ParseEIP712DomainChanged is a log parse operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_Custody *CustodyFilterer) ParseEIP712DomainChanged(log types.Log) (*CustodyEIP712DomainChanged, error) {
	event := new(CustodyEIP712DomainChanged)
	if err := _Custody.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CustodyJoinedIterator is returned from FilterJoined and is used to iterate over the raw logs and unpacked data for Joined events raised by the Custody contract.
type CustodyJoinedIterator struct {
	Event *CustodyJoined // Event containing the contract specifics and raw log

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
func (it *CustodyJoinedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CustodyJoined)
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
		it.Event = new(CustodyJoined)
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
func (it *CustodyJoinedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CustodyJoinedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CustodyJoined represents a Joined event raised by the Custody contract.
type CustodyJoined struct {
	ChannelId [32]byte
	Index     *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterJoined is a free log retrieval operation binding the contract event 0xe8e915db7b3549b9e9e9b3e2ec2dc3edd1f76961504366998824836401f6846a.
//
// Solidity: event Joined(bytes32 indexed channelId, uint256 index)
func (_Custody *CustodyFilterer) FilterJoined(opts *bind.FilterOpts, channelId [][32]byte) (*CustodyJoinedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _Custody.contract.FilterLogs(opts, "Joined", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &CustodyJoinedIterator{contract: _Custody.contract, event: "Joined", logs: logs, sub: sub}, nil
}

// WatchJoined is a free log subscription operation binding the contract event 0xe8e915db7b3549b9e9e9b3e2ec2dc3edd1f76961504366998824836401f6846a.
//
// Solidity: event Joined(bytes32 indexed channelId, uint256 index)
func (_Custody *CustodyFilterer) WatchJoined(opts *bind.WatchOpts, sink chan<- *CustodyJoined, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _Custody.contract.WatchLogs(opts, "Joined", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CustodyJoined)
				if err := _Custody.contract.UnpackLog(event, "Joined", log); err != nil {
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

// ParseJoined is a log parse operation binding the contract event 0xe8e915db7b3549b9e9e9b3e2ec2dc3edd1f76961504366998824836401f6846a.
//
// Solidity: event Joined(bytes32 indexed channelId, uint256 index)
func (_Custody *CustodyFilterer) ParseJoined(log types.Log) (*CustodyJoined, error) {
	event := new(CustodyJoined)
	if err := _Custody.contract.UnpackLog(event, "Joined", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CustodyOpenedIterator is returned from FilterOpened and is used to iterate over the raw logs and unpacked data for Opened events raised by the Custody contract.
type CustodyOpenedIterator struct {
	Event *CustodyOpened // Event containing the contract specifics and raw log

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
func (it *CustodyOpenedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CustodyOpened)
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
		it.Event = new(CustodyOpened)
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
func (it *CustodyOpenedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CustodyOpenedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CustodyOpened represents a Opened event raised by the Custody contract.
type CustodyOpened struct {
	ChannelId [32]byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterOpened is a free log retrieval operation binding the contract event 0xd087f17acc177540af5f382bc30c65363705b90855144d285a822536ee11fdd1.
//
// Solidity: event Opened(bytes32 indexed channelId)
func (_Custody *CustodyFilterer) FilterOpened(opts *bind.FilterOpts, channelId [][32]byte) (*CustodyOpenedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _Custody.contract.FilterLogs(opts, "Opened", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &CustodyOpenedIterator{contract: _Custody.contract, event: "Opened", logs: logs, sub: sub}, nil
}

// WatchOpened is a free log subscription operation binding the contract event 0xd087f17acc177540af5f382bc30c65363705b90855144d285a822536ee11fdd1.
//
// Solidity: event Opened(bytes32 indexed channelId)
func (_Custody *CustodyFilterer) WatchOpened(opts *bind.WatchOpts, sink chan<- *CustodyOpened, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _Custody.contract.WatchLogs(opts, "Opened", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CustodyOpened)
				if err := _Custody.contract.UnpackLog(event, "Opened", log); err != nil {
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

// ParseOpened is a log parse operation binding the contract event 0xd087f17acc177540af5f382bc30c65363705b90855144d285a822536ee11fdd1.
//
// Solidity: event Opened(bytes32 indexed channelId)
func (_Custody *CustodyFilterer) ParseOpened(log types.Log) (*CustodyOpened, error) {
	event := new(CustodyOpened)
	if err := _Custody.contract.UnpackLog(event, "Opened", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CustodyResizedIterator is returned from FilterResized and is used to iterate over the raw logs and unpacked data for Resized events raised by the Custody contract.
type CustodyResizedIterator struct {
	Event *CustodyResized // Event containing the contract specifics and raw log

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
func (it *CustodyResizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CustodyResized)
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
		it.Event = new(CustodyResized)
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
func (it *CustodyResizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CustodyResizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CustodyResized represents a Resized event raised by the Custody contract.
type CustodyResized struct {
	ChannelId        [32]byte
	DeltaAllocations []*big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterResized is a free log retrieval operation binding the contract event 0xf3b6c524f73df7344d9fcf2f960a57aba7fba7e292d8b79ed03d786f7b2b112f.
//
// Solidity: event Resized(bytes32 indexed channelId, int256[] deltaAllocations)
func (_Custody *CustodyFilterer) FilterResized(opts *bind.FilterOpts, channelId [][32]byte) (*CustodyResizedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _Custody.contract.FilterLogs(opts, "Resized", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &CustodyResizedIterator{contract: _Custody.contract, event: "Resized", logs: logs, sub: sub}, nil
}

// WatchResized is a free log subscription operation binding the contract event 0xf3b6c524f73df7344d9fcf2f960a57aba7fba7e292d8b79ed03d786f7b2b112f.
//
// Solidity: event Resized(bytes32 indexed channelId, int256[] deltaAllocations)
func (_Custody *CustodyFilterer) WatchResized(opts *bind.WatchOpts, sink chan<- *CustodyResized, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _Custody.contract.WatchLogs(opts, "Resized", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CustodyResized)
				if err := _Custody.contract.UnpackLog(event, "Resized", log); err != nil {
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

// ParseResized is a log parse operation binding the contract event 0xf3b6c524f73df7344d9fcf2f960a57aba7fba7e292d8b79ed03d786f7b2b112f.
//
// Solidity: event Resized(bytes32 indexed channelId, int256[] deltaAllocations)
func (_Custody *CustodyFilterer) ParseResized(log types.Log) (*CustodyResized, error) {
	event := new(CustodyResized)
	if err := _Custody.contract.UnpackLog(event, "Resized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CustodyWithdrawnIterator is returned from FilterWithdrawn and is used to iterate over the raw logs and unpacked data for Withdrawn events raised by the Custody contract.
type CustodyWithdrawnIterator struct {
	Event *CustodyWithdrawn // Event containing the contract specifics and raw log

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
func (it *CustodyWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CustodyWithdrawn)
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
		it.Event = new(CustodyWithdrawn)
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
func (it *CustodyWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CustodyWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CustodyWithdrawn represents a Withdrawn event raised by the Custody contract.
type CustodyWithdrawn struct {
	Wallet common.Address
	Token  common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterWithdrawn is a free log retrieval operation binding the contract event 0xd1c19fbcd4551a5edfb66d43d2e337c04837afda3482b42bdf569a8fccdae5fb.
//
// Solidity: event Withdrawn(address indexed wallet, address indexed token, uint256 amount)
func (_Custody *CustodyFilterer) FilterWithdrawn(opts *bind.FilterOpts, wallet []common.Address, token []common.Address) (*CustodyWithdrawnIterator, error) {

	var walletRule []interface{}
	for _, walletItem := range wallet {
		walletRule = append(walletRule, walletItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Custody.contract.FilterLogs(opts, "Withdrawn", walletRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return &CustodyWithdrawnIterator{contract: _Custody.contract, event: "Withdrawn", logs: logs, sub: sub}, nil
}

// WatchWithdrawn is a free log subscription operation binding the contract event 0xd1c19fbcd4551a5edfb66d43d2e337c04837afda3482b42bdf569a8fccdae5fb.
//
// Solidity: event Withdrawn(address indexed wallet, address indexed token, uint256 amount)
func (_Custody *CustodyFilterer) WatchWithdrawn(opts *bind.WatchOpts, sink chan<- *CustodyWithdrawn, wallet []common.Address, token []common.Address) (event.Subscription, error) {

	var walletRule []interface{}
	for _, walletItem := range wallet {
		walletRule = append(walletRule, walletItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Custody.contract.WatchLogs(opts, "Withdrawn", walletRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CustodyWithdrawn)
				if err := _Custody.contract.UnpackLog(event, "Withdrawn", log); err != nil {
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
func (_Custody *CustodyFilterer) ParseWithdrawn(log types.Log) (*CustodyWithdrawn, error) {
	event := new(CustodyWithdrawn)
	if err := _Custody.contract.UnpackLog(event, "Withdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
