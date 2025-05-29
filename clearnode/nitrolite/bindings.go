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

// Signature is an auto generated low-level Go binding around an user-defined struct.
type Signature struct {
	V uint8
	R [32]byte
	S [32]byte
}

// State is an auto generated low-level Go binding around an user-defined struct.
type State struct {
	Intent      uint8
	Version     *big.Int
	Data        []byte
	Allocations []Allocation
	Sigs        []Signature
}

// CustodyMetaData contains all meta data concerning the Custody contract.
var CustodyMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"internalType\":\"structSignature[]\",\"name\":\"sigs\",\"type\":\"tuple[]\"}],\"internalType\":\"structState\",\"name\":\"candidate\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"internalType\":\"structSignature[]\",\"name\":\"sigs\",\"type\":\"tuple[]\"}],\"internalType\":\"structState[]\",\"name\":\"proofs\",\"type\":\"tuple[]\"}],\"name\":\"challenge\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"internalType\":\"structSignature[]\",\"name\":\"sigs\",\"type\":\"tuple[]\"}],\"internalType\":\"structState\",\"name\":\"candidate\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"internalType\":\"structSignature[]\",\"name\":\"sigs\",\"type\":\"tuple[]\"}],\"internalType\":\"structState[]\",\"name\":\"proofs\",\"type\":\"tuple[]\"}],\"name\":\"checkpoint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"internalType\":\"structSignature[]\",\"name\":\"sigs\",\"type\":\"tuple[]\"}],\"internalType\":\"structState\",\"name\":\"candidate\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"internalType\":\"structSignature[]\",\"name\":\"sigs\",\"type\":\"tuple[]\"}],\"internalType\":\"structState[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"name\":\"close\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address[]\",\"name\":\"participants\",\"type\":\"address[]\"},{\"internalType\":\"address\",\"name\":\"adjudicator\",\"type\":\"address\"},{\"internalType\":\"uint64\",\"name\":\"challenge\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"}],\"internalType\":\"structChannel\",\"name\":\"ch\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"internalType\":\"structSignature[]\",\"name\":\"sigs\",\"type\":\"tuple[]\"}],\"internalType\":\"structState\",\"name\":\"initial\",\"type\":\"tuple\"}],\"name\":\"create\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"getAccountChannels\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"getAccountInfo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"available\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"channelCount\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"internalType\":\"structSignature\",\"name\":\"sig\",\"type\":\"tuple\"}],\"name\":\"join\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"internalType\":\"structSignature[]\",\"name\":\"sigs\",\"type\":\"tuple[]\"}],\"internalType\":\"structState\",\"name\":\"candidate\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"internalType\":\"structSignature[]\",\"name\":\"sigs\",\"type\":\"tuple[]\"}],\"internalType\":\"structState[]\",\"name\":\"proofs\",\"type\":\"tuple[]\"}],\"name\":\"resize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"expiration\",\"type\":\"uint256\"}],\"name\":\"Challenged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"}],\"name\":\"Checkpointed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"internalType\":\"structSignature[]\",\"name\":\"sigs\",\"type\":\"tuple[]\"}],\"indexed\":false,\"internalType\":\"structState\",\"name\":\"finalState\",\"type\":\"tuple\"}],\"name\":\"Closed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"wallet\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"participants\",\"type\":\"address[]\"},{\"internalType\":\"address\",\"name\":\"adjudicator\",\"type\":\"address\"},{\"internalType\":\"uint64\",\"name\":\"challenge\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"}],\"indexed\":false,\"internalType\":\"structChannel\",\"name\":\"channel\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"enumStateIntent\",\"name\":\"intent\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"destination\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structAllocation[]\",\"name\":\"allocations\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"internalType\":\"structSignature[]\",\"name\":\"sigs\",\"type\":\"tuple[]\"}],\"indexed\":false,\"internalType\":\"structState\",\"name\":\"initial\",\"type\":\"tuple\"}],\"name\":\"Created\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"wallet\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deposited\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"Joined\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"}],\"name\":\"Opened\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"int256[]\",\"name\":\"deltaAllocations\",\"type\":\"int256[]\"}],\"name\":\"Resized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"wallet\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Withdrawn\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"ChallengeNotExpired\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ChannelNotFinal\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"channelId\",\"type\":\"bytes32\"}],\"name\":\"ChannelNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ECDSAInvalidSignature\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"length\",\"type\":\"uint256\"}],\"name\":\"ECDSAInvalidSignatureLength\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"ECDSAInvalidSignatureS\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"available\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"required\",\"type\":\"uint256\"}],\"name\":\"InsufficientBalance\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidAdjudicator\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidAllocations\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidAmount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidChallengePeriod\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidParticipant\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidState\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidStateSignatures\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidStatus\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidValue\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"SafeERC20FailedOperation\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"TransferFailed\",\"type\":\"error\"}]",
	Bin: "0x0x60808060405234601557613b17908161001a8239f35b5f80fdfe60806040526004361015610011575f80fd5b5f3560e01c8063259311c914611dd157806347e7ef2414611c725780636332fef614611c0b5780637de7ad62146117805780637e2d8d72146116c9578063a22b823d1461149c578063d0cce1e814611112578063d37ff7b514610827578063de22731f1461021a5763f3fef3a314610087575f80fd5b34610216576040600319360112610216576100a0612606565b60243590335f52600160205260405f20906001600160a01b0381165f528160205260405f2054918383106101e6576001600160a01b0392508282165f5260205260405f206100ef84825461290c565b9055169081610194575f80808084335af13d1561018f573d610110816126c3565b9061011e60405192836126a0565b81525f60203d92013e5b1561015c575b6040519081527fd1c19fbcd4551a5edfb66d43d2e337c04837afda3482b42bdf569a8fccdae5fb60203392a3005b907fbf182be8000000000000000000000000000000000000000000000000000000005f526004523360245260445260645ffd5b610128565b6101e16040517fa9059cbb000000000000000000000000000000000000000000000000000000006020820152336024820152826044820152604481526101db6064826126a0565b83613806565b61012e565b50507fcf479181000000000000000000000000000000000000000000000000000000005f5260045260245260445ffd5b5f80fd5b34610216576102283661258a565b5050815f525f60205260405f206003810160ff8154166005811015610813578015610800576002810361078e5750823560048110156102165761026a816129e9565b600381036105fd57602084013580156105fd576080850191600261028e8488612b1f565b905003610766576102b16102a186612b55565b6102ab36896127c6565b90613386565b1561076657600d8501906102c4816129e9565b60ff60ff198354169116179055600e840155600f83016102e76040860186612bb1565b9067ffffffffffffffff8211610752576103018354612be4565b601f8111610717575b505f90601f83116001146106b35761033992915f91836106a8575b50508160011b915f199060031b1c19161790565b90555b6010830161034d6060860186612b1f565b91906103598383612c32565b905f5260205f205f915b838310610642575050505061037c601184019185612b1f565b91906103888383612cb4565b905f5260205f205f915b8383106106255750505050600460ff198254161790555b6103c06103b96060840184612b1f565b36916126f7565b60028151036105fd575f5b600281106105e0575050335f5260016020526103ed83600160405f20016138db565b505f5b600281106105a7575050815f525f60205260405f2080545f82558061058d575b505f60018201555f60028201555f600382015561042f6004820161332e565b61043b6008820161332e565b5f600c8201555f600d8201555f600e820155600f810161045b8154612be4565b908161054a575b5050601081018054905f815581610511575b50506011018054905f8155816104c4575b837f3646844802330633cc652490829391a0e9ddb82143a86a7e39ca148dfb05c9106104bf85604051918291602083526020830190612da8565b0390a2005b816003029160038304036104fd575f5260205f20908101905b8181101561048557805f600392555f60018201555f6002820155016104dd565b634e487b7160e01b5f52601160045260245ffd5b816003029160038304036104fd575f5260205f20908101905b8181101561047457805f600392555f60018201555f60028201550161052a565b81601f5f93116001146105615750555b8380610462565b8183526020832061057d91601f0160051c810190600101612c1c565b808252816020812091555561055a565b6105a190825f5260205f2090810190612c1c565b83610410565b806001600160a01b036105bc60019385613189565b90549060031b1c165f52816020526105d9858360405f20016138db565b50016103f0565b806105f76105f060019385612936565b5187613753565b016103cb565b7fbaf3f0f7000000000000000000000000000000000000000000000000000000005f5260045ffd5b600360608261063660019486612d22565b01920192019190610392565b60036060826001600160a01b0361065a600195612ca0565b166001600160a01b031986541617855561067660208201612ca0565b6001600160a01b0385870191166001600160a01b03198254161790556040810135600286015501920192019190610363565b013590508980610325565b601f19831691845f5260205f20925f5b8181106106ff57509084600195949392106106e6575b505050811b01905561033c565b01355f19600384901b60f8161c191690558880806106d9565b919360206001819287870135815501950192016106c3565b61074290845f5260205f20601f850160051c81019160208610610748575b601f0160051c0190612c1c565b8861030a565b9091508190610735565b634e487b7160e01b5f52604160045260245ffd5b7f773a750f000000000000000000000000000000000000000000000000000000005f5260045ffd5b6003036107d857600c82015442106107b057600460ff198254161790556103a9565b7f151f07fe000000000000000000000000000000000000000000000000000000005f5260045ffd5b7ff525e320000000000000000000000000000000000000000000000000000000005f5260045ffd5b846379c1d89f60e11b5f5260045260245ffd5b634e487b7160e01b5f52602160045260245ffd5b346102165760406003193601126102165760043567ffffffffffffffff8111610216578060040190803603906080600319830112610216576024359167ffffffffffffffff831161021657826004019160a060031985360301126102165760026108918680613206565b9050148015906110ea575b80156110bc575b8015611073575b61104b57602481016001600160a01b036108c382612ca0565b1615611023576044820190610e1067ffffffffffffffff6108e38461323c565b1610610ffb578435926004841015610216576108fe846129e9565b600184036105fd57602487013596876105fd5761092361091e368b613266565b6136b6565b97885f525f60205260ff600360405f200154166005811015610813576107d857610960610950368c613266565b61095a368b6127c6565b906134dd565b60848301906001610971838c612b1f565b90500361076657610982828b612b1f565b15610dd9576109918d80613206565b92909215610dd9576109b06109a86109b694612ca0565b92369061278d565b906135c4565b1561076657606483019260026109cc858c612b1f565b905003610fd3578a5f525f60205260405f20926109e98d80613206565b9067ffffffffffffffff821161075257680100000000000000008211610752578554828755808310610fb7575b50855f5260205f205f5b838110610f9c5750505050600184016001600160a01b03610a4089612ca0565b166001600160a01b0319825416178155610a598961323c565b7fffffffff0000000000000000ffffffffffffffffffffffffffffffffffffffff7bffffffffffffffff000000000000000000000000000000000000000083549260a01b1691161790556002840198606487019967ffffffffffffffff610abf8c61323c565b82547fffffffffffffffffffffffffffffffffffffffffffffffff00000000000000001691161790556003850180547fffffffffffffffffffffff000000000000000000000000000000000000000000163360081b74ffffffffffffffffffffffffffffffffffffffff0016176001179055600d850190610b3f816129e9565b60ff60ff198354169116179055600e840155610b626044600f850192018b612bb1565b9067ffffffffffffffff821161075257610b7c8354612be4565b601f8111610f6c575b505f90601f8311600114610f0857610bb392915f9183610efd5750508160011b915f199060031b1c19161790565b90555b60108201610bc4848b612b1f565b9190610bd08383612c32565b905f5260205f205f915b838310610e975750505050610bf360118301918a612b1f565b9190610bff8383612cb4565b905f5260205f205f915b838310610e7a57505050505f916008600483019201925b60028110610ded57505090610c37610c7c9261319e565b90610c536001600160a01b038351166020840151908c336135fc565b9060206001916001600160a01b0380825116166001600160a01b03198554161784550151910155565b610c868880613206565b15610dd957610c9c6001600160a01b0391612ca0565b165f526001602052610cb487600160405f2001613873565b5060405194604086527fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffdd60c087019935910181121561021657016024600482013591019767ffffffffffffffff8211610216578160051b36038913610216578190608060408801525260e0850197905f5b818110610db15750505092610da660209767ffffffffffffffff610d927f7044488f9b947dc40d596a71992214b1050317a18ab1dced28e9d22320c398429682610d8787986001600160a01b03610d7c8f9d61261c565b1660608a0152613251565b166080870152613251565b1660a0840152828103898401523396612da8565b0390a3604051908152f35b9091986020806001926001600160a01b03610dcb8e61261c565b168152019a01929101610d25565b634e487b7160e01b5f52603260045260245ffd5b80610e528b6040610e2384610e1188610e1d6020610e1760019b610e11858b612b1f565b90612d63565b01612ca0565b95612b1f565b01356001600160a01b0360405192610e3a84612668565b1682526020820152610e4c8387612d50565b906131c5565b610e74604051610e6181612668565b5f81525f6020820152610e4c8388612d50565b01610c20565b6003606082610e8b60019486612d22565b01920192019190610c09565b60036060826001600160a01b03610eaf600195612ca0565b166001600160a01b0319865416178555610ecb60208201612ca0565b6001600160a01b0385870191166001600160a01b03198254161790556040810135600286015501920192019190610bda565b013590508f80610325565b601f19831691845f5260205f20925f5b818110610f545750908460019594939210610f3b575b505050811b019055610bb6565b01355f19600384901b60f8161c191690558e8080610f2e565b91936020600181928787013581550195019201610f18565b610f9690845f5260205f20601f850160051c8101916020861061074857601f0160051c0190612c1c565b8e610b85565b6001906020610faa85612ca0565b9401938184015501610a20565b610fcd90875f528360205f209182019101612c1c565b8f610a16565b7f52e4cb1c000000000000000000000000000000000000000000000000000000005f5260045ffd5b7fb4e12433000000000000000000000000000000000000000000000000000000005f5260045ffd5b7fea9e70ce000000000000000000000000000000000000000000000000000000005f5260045ffd5b7fa145c43e000000000000000000000000000000000000000000000000000000005f5260045ffd5b5061107e8580613206565b15610dd95761108c90612ca0565b6110968680613206565b60011015610dd9576001600160a01b036110b36020829301612ca0565b169116146108aa565b506110c78580613206565b60011015610dd9576110e360206001600160a01b039201612ca0565b16156108a3565b506110f58580613206565b15610dd95761110b6001600160a01b0391612ca0565b161561089c565b34610216576111203661258a565b91835f525f60205260405f2092600384019260ff84541691600583101561081357821561148957600483146107d857608084019261115e8486612b1f565b905015610766576001148061147c575b6105fd5760206111a4916001600160a01b036001890154169360405193849283926305b959ef60e01b8452898c60048601612f08565b0381855afa908115611471575f91611442575b50156105fd57602083013590600d8601600e870191825484146105fd576111e890866111e284612fa0565b91613415565b156105fd578435600481101561021657611201816129e9565b60ff60ff19835416911617905555600f84016112206040840184612bb1565b9067ffffffffffffffff82116107525761123a8354612be4565b601f8111611412575b505f90601f83116001146113ae5761127192915f91836106a85750508160011b915f199060031b1c19161790565b90555b601084016112856060840184612b1f565b91906112918383612c32565b905f5260205f205f915b83831061134857505050506112b4906011850192612b1f565b91906112c08383612cb4565b905f5260205f205f915b83831061132b575050505060ff815416600581101561081357600314611312575b827f1f681d6befe6e92b986338164917aaa3f065b8d2de29bb520aa373114e5ec0345f80a2005b5f91600c91600260ff19825416179055015581806112eb565b600360608261133c60019486612d22565b019201920191906112ca565b60036060826001600160a01b03611360600195612ca0565b166001600160a01b031986541617855561137c60208201612ca0565b6001600160a01b0385870191166001600160a01b0319825416179055604081013560028601550192019201919061129b565b601f19831691845f5260205f20925f5b8181106113fa57509084600195949392106113e1575b505050811b019055611274565b01355f19600384901b60f8161c191690558880806113d4565b919360206001819287870135815501950192016113be565b61143c90845f5260205f20601f850160051c8101916020861061074857601f0160051c0190612c1c565b88611243565b611464915060203d60201161146a575b61145c81836126a0565b81019061294a565b876111b7565b503d611452565b6040513d5f823e3d90fd5b506020840135151561116e565b866379c1d89f60e11b5f5260045260245ffd5b346102165760a06003193601126102165760043560607fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffbc36011261021657805f525f60205260405f20906003820160ff81541660058110156108135780156116b6575f19016107d85760016024350361104b57600b83015461104b5761153061152484612b55565b61095a600d8601612fa0565b926001600160a01b0361154282613171565b90549060031b1c1693604051916115588361264c565b60443560ff8116968782036102165761158691855260643594856020820152608435948560408301526135c4565b15610766576011810191825492680100000000000000008410156107525760018401808255841015610dd9576020976001600160a01b0395611634956002935f5260038b5f209102019160ff198354161782556001820155015561162f6115ef6006830161319e565b61160285825116898301519089336135fc565b600a83019060206001916001600160a01b0380825116166001600160a01b03198554161784550151910155565b613171565b90549060031b1c165f526001835261165282600160405f2001613873565b50600260ff19825416179055807fe8e915db7b3549b9e9e9b3e2ec2dc3edd1f76961504366998824836401f6846a8360405160018152a260405190807fd087f17acc177540af5f382bc30c65363705b90855144d285a822536ee11fdd15f80a28152f35b826379c1d89f60e11b5f5260045260245ffd5b34610216576020600319360112610216576001600160a01b036116ea612606565b165f526001602052600160405f2001604051806020835491828152019081935f5260205f20905f5b81811061176a57505050816117289103826126a0565b604051918291602083019060208452518091526040830191905f5b818110611751575050500390f35b8251845285945060209384019390920191600101611743565b8254845260209093019260019283019201611712565b346102165761178e3661258a565b9091835f525f60205260405f2092600384019260ff84541691600583101561081357821561148957600483146107d85760808401926117cd8486612b1f565b9050156107665760011480611bfe575b6105fd5760408401916117f08386612bb1565b9050611b46575b6118018386612bb1565b9050158015611b0a575b611a74575b5050600d8501833560048110156102165761182a816129e9565b60ff60ff1983541691161790556020830135600e86015561184f600f86019184612bb1565b9067ffffffffffffffff8211610752576118698354612be4565b601f8111611a44575b505f90601f83116001146119e0576118a092915f91836106a85750508160011b915f199060031b1c19161790565b90555b601084016118b46060840184612b1f565b91906118c08383612c32565b905f5260205f205f915b83831061197a57505050506118e3906011850192612b1f565b91906118ef8383612cb4565b905f5260205f205f915b83831061195d57867f08818bbbf6e59017d5461143d9f1c4e3fb74703f7fb792c207cbeed4b344cefc60208888600c61194367ffffffffffffffff600185015460a01c1642612b12565b9201918255600360ff1982541617905554604051908152a2005b600360608261196e60019486612d22565b019201920191906118f9565b60036060826001600160a01b03611992600195612ca0565b166001600160a01b03198654161785556119ae60208201612ca0565b6001600160a01b0385870191166001600160a01b031982541617905560408101356002860155019201920191906118ca565b601f19831691845f5260205f20925f5b818110611a2c5750908460019594939210611a13575b505050811b0190556118a3565b01355f19600384901b60f8161c19169055888080611a06565b919360206001819287870135815501950192016119f0565b611a6e90845f5260205f20601f850160051c8101916020861061074857601f0160051c0190612c1c565b88611872565b6020611aa6916001600160a01b0360018a0154169360405193849283926305b959ef60e01b84528a8d60048601612f08565b0381855afa908115611471575f91611aeb575b50156105fd57600e8601546020850135146105fd57611adf90846111e2600d8901612fa0565b156105fd578680611810565b611b04915060203d60201161146a5761145c81836126a0565b88611ab9565b508435600481101561021657611b1f816129e9565b60018114159081611b31575b5061180b565b60029150611b3e816129e9565b141589611b2b565b8435600481101561021657611b5a816129e9565b60018103611b69575b506117f7565b600290611b75816129e9565b14611b81575b88611b63565b611b946005880154600789015490612b12565b60608601611ba28188612b1f565b91909115610dd957611bb49088612b1f565b60019291921015610dd957604060a0611bd1930135910135612b12565b14611b7b577f52e4cb1c000000000000000000000000000000000000000000000000000000005f5260045ffd5b50602084013515156117dd565b3461021657604060031936011261021657611c24612606565b6024356001600160a01b0381168103610216576001600160a01b03604092165f5260016020526001600160a01b03825f2091165f52806020526001825f205491015482519182526020820152f35b604060031936011261021657611c86612606565b602435908115611da9576001600160a01b03169081158015611d7b57813403611d53575b335f52600160205260405f20835f5260205260405f20611ccb838254612b12565b905515611d01575b6040519081527f8752a472e571a816aea92eec8dae9baf628e840f4929fbcc2d155e6233ff68a760203392a3005b611d4e6040517f23b872dd000000000000000000000000000000000000000000000000000000006020820152336024820152306044820152826064820152606481526101db6084826126a0565b611cd3565b7faa7feadc000000000000000000000000000000000000000000000000000000005f5260045ffd5b3415611caa577faa7feadc000000000000000000000000000000000000000000000000000000005f5260045ffd5b7f2c5211c6000000000000000000000000000000000000000000000000000000005f5260045ffd5b3461021657611ddf3661258a565b919290815f525f60205260405f209160ff6003840154166005811015610813578015612577577ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe016107d85783156105fd57813592609e1983360301938481121561021657611e5190840136906127c6565b5f198601968688116104fd57601f19611e82611e6c8a6126df565b99611e7a6040519b8c6126a0565b808b526126df565b015f5b81811061254057505060015b87811015611ee1578060051b86013587811215610216575f198201919087018183116104fd57828b611ed3600195611ecd611eda9536906127c6565b92612936565b528b612936565b5001611e91565b50876001600160a01b0360018501541660405180926305b959ef60e01b825260606004830152611f29611f1760648401896129a2565b60031984820301602485015287612a18565b600319838203016044840152815180825260208201916020808360051b8301019401925f915b83831061250f57505050505091818060209403915afa908115611471575f916124f0575b50156105fd576020810135916020810151600181018091116104fd5783036105fd5760600192611fa38451613350565b60608201611fbc611fb76103b98386612b1f565b613350565b611fd2611fc883612b55565b6102ab36866127c6565b156107665760408301611fe58185612bb1565b810195906020818803126102165780359067ffffffffffffffff821161021657019686601f890112156102165787359661201e886126df565b9861202c6040519a8b6126a0565b888a5260208a01906020829a60051b82010192831161021657602001905b8282106124e057505050853590600482101561021657612069826129e9565b600282036105fd575161207f6103b98689612b1f565b9060028a51036105fd576120b18160406120a88161209f6120e996612919565b51015192612926565b51015190612b12565b916120e36120d38c6120cc6120c582612919565b5191612926565b51906133fa565b9160406120a88161209f84612919565b926133fa565b03610fd3576120fb6103b98588612b1f565b600486016001600160a01b03815416805f5b8d8d6002831061248257925050505f5b600281106123e557505050905f9060088801915b600281106123a25750505050600d85019061214b816129e9565b60ff60ff198354169116179055600e84015561216b600f84019185612bb1565b9067ffffffffffffffff8211610752576121858354612be4565b601f8111612372575b505f90601f831160011461230e576121bc92915f91836123035750508160011b915f199060031b1c19161790565b90555b6121cd601083019184612b1f565b91906121d98383612c32565b905f5260205f205f915b83831061229d575050505060116122009101916080810190612b1f565b919061220c8383612cb4565b905f5260205f205f915b838310612280575050505060405191602083019060208452518091526040830191905f5b81811061226a57857ff3b6c524f73df7344d9fcf2f960a57aba7fba7e292d8b79ed03d786f7b2b112f86860387a2005b825184526020938401939092019160010161223a565b600360608261229160019486612d22565b01920192019190612216565b60036060826001600160a01b036122b5600195612ca0565b166001600160a01b03198654161785556122d160208201612ca0565b6001600160a01b0385870191166001600160a01b031982541617905560408101356002860155019201920191906121e3565b013590508a80610325565b601f19831691845f5260205f20925f5b81811061235a5750908460019594939210612341575b505050811b0190556121bf565b01355f19600384901b60f8161c19169055898080612334565b9193602060018192878701358155019501920161231e565b61239c90845f5260205f20601f850160051c8101916020861061074857601f0160051c0190612c1c565b8961218e565b8060406123b160019385612936565b510151826123bf8388612d50565b50015560406123ce8285612936565b510151826123dc8387612d50565b50015501612131565b895f6123f18385612936565b5112612404575b508c915060010161211d565b816001600160a01b0361241a8261242894613189565b90549060031b1c1693612936565b517f800000000000000000000000000000000000000000000000000000000000000081146104fd576001928f61247992604051926124658461264c565b83528660208401525f036040830152613753565b8c9150896123f8565b60019383928d5f6124938686612936565b51136124a7575b505050505001819061210d565b846001600160a01b036124c06124d5976124ce94613189565b90549060031b1c1694612936565b51926135fc565b808d8f85908d61249a565b813581526020918201910161204a565b612509915060203d60201161146a5761145c81836126a0565b85611f73565b91939550919360208061252e83601f1986600196030187528951612a18565b97019301930190928795949293611f4f565b60209060405161254f81612630565b5f81525f83820152606060408201526060808201526060608082015282828d01015201611e85565b506379c1d89f60e11b5f5260045260245ffd5b906060600319830112610216576004359160243567ffffffffffffffff81116102165760a06003198284030112610216576004019160443567ffffffffffffffff811161021657826023820112156102165780600401359267ffffffffffffffff84116102165760248460051b83010111610216576024019190565b600435906001600160a01b038216820361021657565b35906001600160a01b038216820361021657565b60a0810190811067ffffffffffffffff82111761075257604052565b6060810190811067ffffffffffffffff82111761075257604052565b6040810190811067ffffffffffffffff82111761075257604052565b6080810190811067ffffffffffffffff82111761075257604052565b90601f601f19910116810190811067ffffffffffffffff82111761075257604052565b67ffffffffffffffff811161075257601f01601f191660200190565b67ffffffffffffffff81116107525760051b60200190565b929192612703826126df565b9361271160405195866126a0565b606060208685815201930282019181831161021657925b8284106127355750505050565b6060848303126102165760206060916040516127508161264c565b6127598761261c565b815261276683880161261c565b8382015260408701356040820152815201930192612728565b359060ff8216820361021657565b9190826060910312610216576040516127a58161264c565b60408082946127b38161277f565b8452602081013560208501520135910152565b919060a083820312610216576040516127de81612630565b80938035600481101561021657825260208101356020830152604081013567ffffffffffffffff811161021657810183601f82011215610216578035612823816126c3565b9161283160405193846126a0565b818352856020838301011161021657815f92602080930183860137830101526040830152606081013567ffffffffffffffff811161021657810183601f820112156102165783816020612886933591016126f7565b606083015260808101359067ffffffffffffffff8211610216570182601f820112156102165780356128b7816126df565b936128c560405195866126a0565b8185526020606081870193028401019281841161021657602001915b8383106128f2575050505060800152565b6020606091612901848661278d565b8152019201916128e1565b919082039182116104fd57565b805115610dd95760200190565b805160011015610dd95760400190565b8051821015610dd95760209160051b010190565b90816020910312610216575180151581036102165790565b90602082549182815201915f5260205f20905f5b8181106129835750505090565b82546001600160a01b0316845260209093019260019283019201612976565b9060808152606067ffffffffffffffff60026129c16080850186612962565b948260018201546001600160a01b038116602088015260a01c16604086015201541691015290565b6004111561081357565b90601f19601f602080948051918291828752018686015e5f8582860101520116010190565b8051612a23816129e9565b825260208101516020830152612a48604082015160a0604085015260a08401906129f3565b906060810151918381036060850152602080845192838152019301905f5b818110612aca5750505060800151916080818303910152602080835192838152019201905f5b818110612a995750505090565b909192602060606001926040875160ff81511683528481015185840152015160408201520194019101919091612a8c565b9091936020612b0860019287519060406060926001600160a01b0381511683526001600160a01b036020820151166020840152015160408201520190565b9501929101612a66565b919082018092116104fd57565b903590601e1981360301821215610216570180359067ffffffffffffffff82116102165760200191606082023603831361021657565b90604051612b6281612684565b606067ffffffffffffffff60028395604051612b8981612b828185612962565b03826126a0565b85528260018201546001600160a01b038116602088015260a01c166040860152015416910152565b903590601e1981360301821215610216570180359067ffffffffffffffff82116102165760200191813603831361021657565b90600182811c92168015612c12575b6020831014612bfe57565b634e487b7160e01b5f52602260045260245ffd5b91607f1691612bf3565b818110612c27575050565b5f8155600101612c1c565b9068010000000000000000811161075257815491818155828210612c5557505050565b826003029260038404036104fd57816003029160038304036104fd575f5260205f2091820191015b818110612c88575050565b805f600392555f60018201555f600282015501612c7d565b356001600160a01b03811681036102165790565b9068010000000000000000811161075257815491818155828210612cd757505050565b826003029260038404036104fd57816003029160038304036104fd575f5260205f2091820191015b818110612d0a575050565b805f600392555f60018201555f600282015501612cff565b90803560ff81168091036102165760029160409160ff19855416178455602081013560018501550135910155565b906002811015610dd95760011b01905f90565b9190811015610dd9576060020190565b9035601e198236030181121561021657016020813591019167ffffffffffffffff821161021657606082023603831361021657565b8035600481101561021657612dbc816129e9565b8252602081013560208301526040810135601e198236030181121561021657810160208135910167ffffffffffffffff82116102165781360381136102165781601f1992601f9260a060408801528160a088015260c08701375f60c08287010152011682019060c082019160e0612e366060840184612d73565b86840360c0016060880152948590529101925f5b818110612ebc57505050612e648160806020930190612d73565b92909360808183039101528281520191905f5b818110612e845750505090565b90919260608060019260ff612e988861277f565b16815260208701356020820152604087013560408201520194019101919091612e77565b9091936060806001926001600160a01b03612ed68961261c565b1681526001600160a01b03612eed60208a0161261c565b16602082015260408881013590820152019501929101612e4a565b91612f1e612f2c926060855260608501906129a2565b908382036020850152612da8565b906040818303910152828152602081019260208160051b83010193835f91609e1982360301945b848410612f64575050505050505090565b90919293949596601f19828203018352873587811215610216576020612f8f60019387839401612da8565b990193019401929195949390612f53565b90604051612fad81612630565b809260ff815416612fbd816129e9565b8252600181015460208301526002810160405190815f825492612fdf84612be4565b808452936001811690811561314f575060011461310b575b50613004925003826126a0565b6040830152600381018054613018816126df565b9161302660405193846126a0565b81835260208301905f5260205f205f915b8383106130c2575050505090600491606084015201908154613058816126df565b9261306660405194856126a0565b81845260208401905f5260205f205f915b838310613088575050505060800152565b6003602060019260405161309b8161264c565b60ff8654168152848601548382015260028601546040820152815201920192019190613077565b600360206001926040516130d58161264c565b6001600160a01b0386541681526001600160a01b0385870154168382015260028601546040820152815201920192019190613037565b90505f9291925260205f20905f915b818310613133575050906020613004928201015f612ff7565b602091935080600191548385880101520191019091839261311a565b6020935061300495925060ff1991501682840152151560051b8201015f612ff7565b805460011015610dd9575f52600160205f2001905f90565b8054821015610dd9575f5260205f2001905f90565b906040516131ab81612668565b6020600182946001600160a01b0381541684520154910152565b91906131f3576020816001600160a01b03806001945116166001600160a01b03198554161784550151910155565b634e487b7160e01b5f525f60045260245ffd5b903590601e1981360301821215610216570180359067ffffffffffffffff821161021657602001918160051b3603831361021657565b3567ffffffffffffffff811681036102165790565b359067ffffffffffffffff8216820361021657565b919091608081840312610216576040519061328082612684565b8193813567ffffffffffffffff81116102165782019080601f830112156102165781356132ac816126df565b926132ba60405194856126a0565b81845260208085019260051b82010192831161021657602001905b82821061331657505050606080926133119285526132f56020820161261c565b602086015261330660408201613251565b604086015201613251565b910152565b602080916133238461261c565b8152019101906132d5565b60048101905b81811061333f575050565b5f8082556001820155600201613334565b60028151036105fd576001600160a01b03602061337b828261337186612919565b5101511693612926565b51015116036105fd57565b90608061339382846134dd565b9101906002825151036133f3575f5b600281106133b35750505050600190565b6133de6133c1828551612936565b516001600160a01b036133d5848851612936565b511690846135c4565b156133eb576001016133a2565b505050505f90565b5050505f90565b9190915f83820193841291129080158216911516176104fd57565b919091602060405180927fcc2a842d00000000000000000000000000000000000000000000000000000000825260406004830152816001600160a01b0381613475613463604483018b612da8565b60031983820301602484015289612a18565b0392165afa5f91816134a0575b5061349557506020809101519101351190565b90505f8092500b1390565b9091506020813d6020116134d5575b816134bc602093836126a0565b810103126102165751805f0b810361021657905f613482565b3d91506134af565b6134e6906136b6565b908051906134f3826129e9565b60208101519161353860606040840151930151926040519485936020850197885261351d816129e9565b6040850152606084015260a0608084015260c08301906129f3565b91601f198284030160a0830152602080825194858152019101925f5b818110613576575050613570925003601f1981018352826126a0565b51902090565b9160019193506135b560209186519060406060926001600160a01b0381511683526001600160a01b036020820151166020840152015160408201520190565b94019101918492939193613554565b6135f5926135ec6001600160a01b039392849360ff8151166040602083015192015192613998565b90959195613a1a565b1691161490565b83156136b0576001600160a01b03165f52600160205260405f206001600160a01b0383165f528060205260405f2054848110613680578461363c9161290c565b906001600160a01b0384165f5260205260405f20555f525f6020526001600160a01b03601260405f200191165f5260205261367c60405f20918254612b12565b9055565b84907fcf479181000000000000000000000000000000000000000000000000000000005f5260045260245260445ffd5b50505050565b80516001600160a01b0360208301511667ffffffffffffffff60608160408601511694015116604051928391602083019560c084019460a088528351809652602060e086019401955f5b81811061373157505061357095506040850152606084015260808301524660a083015203601f1981018352826126a0565b87516001600160a01b0316865260209788019789975090950194600101613700565b906040810191825115613801575f525f602052601260405f20019160208201916001600160a01b0380845116165f528360205260405f20549384156137fa576001600160a01b0392518086115f146137ef576137b090809661290c565b908380865116165f5260205260405f205551165f5260016020526001600160a01b038060405f20925116165f5260205261367c60405f20918254612b12565b506137b0858061290c565b5050505050565b505050565b905f602091828151910182855af115611471575f513d61386a57506001600160a01b0381163b155b6138355750565b6001600160a01b03907f5274afe7000000000000000000000000000000000000000000000000000000005f521660045260245ffd5b6001141561382e565b6001810190825f528160205260405f2054155f146133f357805468010000000000000000811015610752576138c86138b2826001879401855584613189565b819391549060031b91821b915f19901b19161790565b905554915f5260205260405f2055600190565b906001820191815f528260205260405f20548015155f146133eb575f1981018181116104fd5782545f198101919082116104fd57818103613963575b5050508054801561394f575f1901906139308282613189565b8154905f199060031b1b19169055555f526020525f6040812055600190565b634e487b7160e01b5f52603160045260245ffd5b6139836139736138b29386613189565b90549060031b1c92839286613189565b90555f528360205260405f20555f8080613917565b91907f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a08411613a0f579160209360809260ff5f9560405194855216868401526040830152606082015282805260015afa15611471575f516001600160a01b03811615613a0557905f905f90565b505f906001905f90565b5050505f9160039190565b613a23816129e9565b80613a2c575050565b613a35816129e9565b60018103613a65577ff645eedf000000000000000000000000000000000000000000000000000000005f5260045ffd5b613a6e816129e9565b60028103613aa257507ffce698f7000000000000000000000000000000000000000000000000000000005f5260045260245ffd5b600390613aae816129e9565b14613ab65750565b7fd78bce0c000000000000000000000000000000000000000000000000000000005f5260045260245ffdfea26469706673582212200e489ce255e4fea587bb266fdef789d3a2a763c31d7b38b383eddaaff74e2dc764736f6c634300081c0033",
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

// GetAccountChannels is a free data retrieval call binding the contract method 0x7e2d8d72.
//
// Solidity: function getAccountChannels(address account) view returns(bytes32[])
func (_Custody *CustodyCaller) GetAccountChannels(opts *bind.CallOpts, account common.Address) ([][32]byte, error) {
	var out []interface{}
	err := _Custody.contract.Call(opts, &out, "getAccountChannels", account)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetAccountChannels is a free data retrieval call binding the contract method 0x7e2d8d72.
//
// Solidity: function getAccountChannels(address account) view returns(bytes32[])
func (_Custody *CustodySession) GetAccountChannels(account common.Address) ([][32]byte, error) {
	return _Custody.Contract.GetAccountChannels(&_Custody.CallOpts, account)
}

// GetAccountChannels is a free data retrieval call binding the contract method 0x7e2d8d72.
//
// Solidity: function getAccountChannels(address account) view returns(bytes32[])
func (_Custody *CustodyCallerSession) GetAccountChannels(account common.Address) ([][32]byte, error) {
	return _Custody.Contract.GetAccountChannels(&_Custody.CallOpts, account)
}

// GetAccountInfo is a free data retrieval call binding the contract method 0x6332fef6.
//
// Solidity: function getAccountInfo(address user, address token) view returns(uint256 available, uint256 channelCount)
func (_Custody *CustodyCaller) GetAccountInfo(opts *bind.CallOpts, user common.Address, token common.Address) (struct {
	Available    *big.Int
	ChannelCount *big.Int
}, error) {
	var out []interface{}
	err := _Custody.contract.Call(opts, &out, "getAccountInfo", user, token)

	outstruct := new(struct {
		Available    *big.Int
		ChannelCount *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Available = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.ChannelCount = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetAccountInfo is a free data retrieval call binding the contract method 0x6332fef6.
//
// Solidity: function getAccountInfo(address user, address token) view returns(uint256 available, uint256 channelCount)
func (_Custody *CustodySession) GetAccountInfo(user common.Address, token common.Address) (struct {
	Available    *big.Int
	ChannelCount *big.Int
}, error) {
	return _Custody.Contract.GetAccountInfo(&_Custody.CallOpts, user, token)
}

// GetAccountInfo is a free data retrieval call binding the contract method 0x6332fef6.
//
// Solidity: function getAccountInfo(address user, address token) view returns(uint256 available, uint256 channelCount)
func (_Custody *CustodyCallerSession) GetAccountInfo(user common.Address, token common.Address) (struct {
	Available    *big.Int
	ChannelCount *big.Int
}, error) {
	return _Custody.Contract.GetAccountInfo(&_Custody.CallOpts, user, token)
}

// Challenge is a paid mutator transaction binding the contract method 0x7de7ad62.
//
// Solidity: function challenge(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[])[] proofs) returns()
func (_Custody *CustodyTransactor) Challenge(opts *bind.TransactOpts, channelId [32]byte, candidate State, proofs []State) (*types.Transaction, error) {
	return _Custody.contract.Transact(opts, "challenge", channelId, candidate, proofs)
}

// Challenge is a paid mutator transaction binding the contract method 0x7de7ad62.
//
// Solidity: function challenge(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[])[] proofs) returns()
func (_Custody *CustodySession) Challenge(channelId [32]byte, candidate State, proofs []State) (*types.Transaction, error) {
	return _Custody.Contract.Challenge(&_Custody.TransactOpts, channelId, candidate, proofs)
}

// Challenge is a paid mutator transaction binding the contract method 0x7de7ad62.
//
// Solidity: function challenge(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[])[] proofs) returns()
func (_Custody *CustodyTransactorSession) Challenge(channelId [32]byte, candidate State, proofs []State) (*types.Transaction, error) {
	return _Custody.Contract.Challenge(&_Custody.TransactOpts, channelId, candidate, proofs)
}

// Checkpoint is a paid mutator transaction binding the contract method 0xd0cce1e8.
//
// Solidity: function checkpoint(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[])[] proofs) returns()
func (_Custody *CustodyTransactor) Checkpoint(opts *bind.TransactOpts, channelId [32]byte, candidate State, proofs []State) (*types.Transaction, error) {
	return _Custody.contract.Transact(opts, "checkpoint", channelId, candidate, proofs)
}

// Checkpoint is a paid mutator transaction binding the contract method 0xd0cce1e8.
//
// Solidity: function checkpoint(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[])[] proofs) returns()
func (_Custody *CustodySession) Checkpoint(channelId [32]byte, candidate State, proofs []State) (*types.Transaction, error) {
	return _Custody.Contract.Checkpoint(&_Custody.TransactOpts, channelId, candidate, proofs)
}

// Checkpoint is a paid mutator transaction binding the contract method 0xd0cce1e8.
//
// Solidity: function checkpoint(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[])[] proofs) returns()
func (_Custody *CustodyTransactorSession) Checkpoint(channelId [32]byte, candidate State, proofs []State) (*types.Transaction, error) {
	return _Custody.Contract.Checkpoint(&_Custody.TransactOpts, channelId, candidate, proofs)
}

// Close is a paid mutator transaction binding the contract method 0xde22731f.
//
// Solidity: function close(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[])[] ) returns()
func (_Custody *CustodyTransactor) Close(opts *bind.TransactOpts, channelId [32]byte, candidate State, arg2 []State) (*types.Transaction, error) {
	return _Custody.contract.Transact(opts, "close", channelId, candidate, arg2)
}

// Close is a paid mutator transaction binding the contract method 0xde22731f.
//
// Solidity: function close(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[])[] ) returns()
func (_Custody *CustodySession) Close(channelId [32]byte, candidate State, arg2 []State) (*types.Transaction, error) {
	return _Custody.Contract.Close(&_Custody.TransactOpts, channelId, candidate, arg2)
}

// Close is a paid mutator transaction binding the contract method 0xde22731f.
//
// Solidity: function close(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[])[] ) returns()
func (_Custody *CustodyTransactorSession) Close(channelId [32]byte, candidate State, arg2 []State) (*types.Transaction, error) {
	return _Custody.Contract.Close(&_Custody.TransactOpts, channelId, candidate, arg2)
}

// Create is a paid mutator transaction binding the contract method 0xd37ff7b5.
//
// Solidity: function create((address[],address,uint64,uint64) ch, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[]) initial) returns(bytes32 channelId)
func (_Custody *CustodyTransactor) Create(opts *bind.TransactOpts, ch Channel, initial State) (*types.Transaction, error) {
	return _Custody.contract.Transact(opts, "create", ch, initial)
}

// Create is a paid mutator transaction binding the contract method 0xd37ff7b5.
//
// Solidity: function create((address[],address,uint64,uint64) ch, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[]) initial) returns(bytes32 channelId)
func (_Custody *CustodySession) Create(ch Channel, initial State) (*types.Transaction, error) {
	return _Custody.Contract.Create(&_Custody.TransactOpts, ch, initial)
}

// Create is a paid mutator transaction binding the contract method 0xd37ff7b5.
//
// Solidity: function create((address[],address,uint64,uint64) ch, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[]) initial) returns(bytes32 channelId)
func (_Custody *CustodyTransactorSession) Create(ch Channel, initial State) (*types.Transaction, error) {
	return _Custody.Contract.Create(&_Custody.TransactOpts, ch, initial)
}

// Deposit is a paid mutator transaction binding the contract method 0x47e7ef24.
//
// Solidity: function deposit(address token, uint256 amount) payable returns()
func (_Custody *CustodyTransactor) Deposit(opts *bind.TransactOpts, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Custody.contract.Transact(opts, "deposit", token, amount)
}

// Deposit is a paid mutator transaction binding the contract method 0x47e7ef24.
//
// Solidity: function deposit(address token, uint256 amount) payable returns()
func (_Custody *CustodySession) Deposit(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Custody.Contract.Deposit(&_Custody.TransactOpts, token, amount)
}

// Deposit is a paid mutator transaction binding the contract method 0x47e7ef24.
//
// Solidity: function deposit(address token, uint256 amount) payable returns()
func (_Custody *CustodyTransactorSession) Deposit(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Custody.Contract.Deposit(&_Custody.TransactOpts, token, amount)
}

// Join is a paid mutator transaction binding the contract method 0xa22b823d.
//
// Solidity: function join(bytes32 channelId, uint256 index, (uint8,bytes32,bytes32) sig) returns(bytes32)
func (_Custody *CustodyTransactor) Join(opts *bind.TransactOpts, channelId [32]byte, index *big.Int, sig Signature) (*types.Transaction, error) {
	return _Custody.contract.Transact(opts, "join", channelId, index, sig)
}

// Join is a paid mutator transaction binding the contract method 0xa22b823d.
//
// Solidity: function join(bytes32 channelId, uint256 index, (uint8,bytes32,bytes32) sig) returns(bytes32)
func (_Custody *CustodySession) Join(channelId [32]byte, index *big.Int, sig Signature) (*types.Transaction, error) {
	return _Custody.Contract.Join(&_Custody.TransactOpts, channelId, index, sig)
}

// Join is a paid mutator transaction binding the contract method 0xa22b823d.
//
// Solidity: function join(bytes32 channelId, uint256 index, (uint8,bytes32,bytes32) sig) returns(bytes32)
func (_Custody *CustodyTransactorSession) Join(channelId [32]byte, index *big.Int, sig Signature) (*types.Transaction, error) {
	return _Custody.Contract.Join(&_Custody.TransactOpts, channelId, index, sig)
}

// Resize is a paid mutator transaction binding the contract method 0x259311c9.
//
// Solidity: function resize(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[])[] proofs) returns()
func (_Custody *CustodyTransactor) Resize(opts *bind.TransactOpts, channelId [32]byte, candidate State, proofs []State) (*types.Transaction, error) {
	return _Custody.contract.Transact(opts, "resize", channelId, candidate, proofs)
}

// Resize is a paid mutator transaction binding the contract method 0x259311c9.
//
// Solidity: function resize(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[])[] proofs) returns()
func (_Custody *CustodySession) Resize(channelId [32]byte, candidate State, proofs []State) (*types.Transaction, error) {
	return _Custody.Contract.Resize(&_Custody.TransactOpts, channelId, candidate, proofs)
}

// Resize is a paid mutator transaction binding the contract method 0x259311c9.
//
// Solidity: function resize(bytes32 channelId, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[]) candidate, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[])[] proofs) returns()
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
	Expiration *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterChallenged is a free log retrieval operation binding the contract event 0x08818bbbf6e59017d5461143d9f1c4e3fb74703f7fb792c207cbeed4b344cefc.
//
// Solidity: event Challenged(bytes32 indexed channelId, uint256 expiration)
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

// WatchChallenged is a free log subscription operation binding the contract event 0x08818bbbf6e59017d5461143d9f1c4e3fb74703f7fb792c207cbeed4b344cefc.
//
// Solidity: event Challenged(bytes32 indexed channelId, uint256 expiration)
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

// ParseChallenged is a log parse operation binding the contract event 0x08818bbbf6e59017d5461143d9f1c4e3fb74703f7fb792c207cbeed4b344cefc.
//
// Solidity: event Challenged(bytes32 indexed channelId, uint256 expiration)
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
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterCheckpointed is a free log retrieval operation binding the contract event 0x1f681d6befe6e92b986338164917aaa3f065b8d2de29bb520aa373114e5ec034.
//
// Solidity: event Checkpointed(bytes32 indexed channelId)
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

// WatchCheckpointed is a free log subscription operation binding the contract event 0x1f681d6befe6e92b986338164917aaa3f065b8d2de29bb520aa373114e5ec034.
//
// Solidity: event Checkpointed(bytes32 indexed channelId)
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

// ParseCheckpointed is a log parse operation binding the contract event 0x1f681d6befe6e92b986338164917aaa3f065b8d2de29bb520aa373114e5ec034.
//
// Solidity: event Checkpointed(bytes32 indexed channelId)
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

// FilterClosed is a free log retrieval operation binding the contract event 0x3646844802330633cc652490829391a0e9ddb82143a86a7e39ca148dfb05c910.
//
// Solidity: event Closed(bytes32 indexed channelId, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[]) finalState)
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

// WatchClosed is a free log subscription operation binding the contract event 0x3646844802330633cc652490829391a0e9ddb82143a86a7e39ca148dfb05c910.
//
// Solidity: event Closed(bytes32 indexed channelId, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[]) finalState)
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

// ParseClosed is a log parse operation binding the contract event 0x3646844802330633cc652490829391a0e9ddb82143a86a7e39ca148dfb05c910.
//
// Solidity: event Closed(bytes32 indexed channelId, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[]) finalState)
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

// FilterCreated is a free log retrieval operation binding the contract event 0x7044488f9b947dc40d596a71992214b1050317a18ab1dced28e9d22320c39842.
//
// Solidity: event Created(bytes32 indexed channelId, address indexed wallet, (address[],address,uint64,uint64) channel, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[]) initial)
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

// WatchCreated is a free log subscription operation binding the contract event 0x7044488f9b947dc40d596a71992214b1050317a18ab1dced28e9d22320c39842.
//
// Solidity: event Created(bytes32 indexed channelId, address indexed wallet, (address[],address,uint64,uint64) channel, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[]) initial)
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

// ParseCreated is a log parse operation binding the contract event 0x7044488f9b947dc40d596a71992214b1050317a18ab1dced28e9d22320c39842.
//
// Solidity: event Created(bytes32 indexed channelId, address indexed wallet, (address[],address,uint64,uint64) channel, (uint8,uint256,bytes,(address,address,uint256)[],(uint8,bytes32,bytes32)[]) initial)
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
