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

// AppRegistryMetaData contains all meta data concerning the AppRegistry contract.
var AppRegistryMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"asset_\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"unlockPeriod_\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"admin_\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"ADJUDICATOR_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"ASSET\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"DEFAULT_ADMIN_ROLE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"UNLOCK_PERIOD\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"asset\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"balanceOf\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRoleAdmin\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"grantRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"hasRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"lastSlashTimestamp\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"lock\",\"inputs\":[{\"name\":\"target\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"lockStateOf\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"enumILock.LockState\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"relock\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"callerConfirmation\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"revokeRole\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setSlashCooldown\",\"inputs\":[{\"name\":\"newCooldown\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"slash\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"recipient\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decision\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"slashCooldown\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"supportsInterface\",\"inputs\":[{\"name\":\"interfaceId\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"unlock\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"unlockTimestampOf\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"withdraw\",\"inputs\":[{\"name\":\"destination\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"Locked\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"deposited\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newBalance\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Relocked\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"balance\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleAdminChanged\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"previousAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"newAdminRole\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleGranted\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RoleRevoked\",\"inputs\":[{\"name\":\"role\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SlashCooldownUpdated\",\"inputs\":[{\"name\":\"oldCooldown\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newCooldown\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Slashed\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"recipient\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"decision\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"UnlockInitiated\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"balance\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"availableAt\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Withdrawn\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"balance\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AccessControlBadConfirmation\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AccessControlUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"neededRole\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"AlreadyUnlocking\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InsufficientBalance\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAddress\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAmount\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidPeriod\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotLocked\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotUnlocking\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"RecipientIsAdjudicator\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"SafeERC20FailedOperation\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"SlashCooldownActive\",\"inputs\":[{\"name\":\"availableAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"UnlockPeriodNotElapsed\",\"inputs\":[{\"name\":\"availableAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}]",
	Bin: "0x60c03461010d57601f61176038819003918201601f19168301916001600160401b038311848410176101115780849260609460405283398101031261010d5761004781610125565b90610059604060208301519201610125565b60017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0055916001600160a01b038116156100ef5781156100fe5760805260a0526001600160a01b038116156100ef576100b190610139565b5060405161157d90816101c3823960805181818161060d0152818161084c01528181610edd01526110e1015260a05181818161030b0152610aca0152f35b63e6c4247b60e01b5f5260045ffd5b6302e8f35960e31b5f5260045ffd5b5f80fd5b634e487b7160e01b5f52604160045260245ffd5b51906001600160a01b038216820361010d57565b6001600160a01b0381165f9081525f5160206117405f395f51905f52602052604090205460ff166101bd576001600160a01b03165f8181525f5160206117405f395f51905f5260205260408120805460ff191660011790553391907f2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d8180a4600190565b505f9056fe60806040526004361015610011575f80fd5b5f3560e01c8063010379b214610d3257806301ffc9a714610c7357806302f70fe614610c10578063248a9ca314610bbf578063256f8f0914610aed578063259a28cf14610a95578063282d3fdf146107ad5780632f2ff15d1461075157806336568abe146106c957806338d52e0f146106c45780634800d97f146106c457806351cff8d91461058a5780635a6ca4ae1461054f57806370a08231146104ed57806391d1485414610478578063a217fddf14610440578063a43e9cf4146103c1578063a69df4b5146102b5578063c53b573d14610205578063c77f5062146101ca578063d547741f146101675763f538f7dd1461010b575f80fd5b34610163575f7ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc3601126101635760206040517f5ad6c1ce52091d5f5a49dff6df7d4bf577735e77d8c13251e1ea9c82ce8c380d8152f35b5f80fd5b346101635760407ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc360112610163576101c86004356101a4611074565b906101c36101be825f526002602052600160405f20015490565b611201565b61147d565b005b34610163575f7ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc360112610163576020600454604051908152f35b34610163575f7ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc36011261016357335f52600160205260405f20541561028d57335f525f60205260405f2054335f5260016020525f60408120556040519081527fe2d155d9d74decc198a9a9b4f5bddb24ee0842ee745c9ce57e3573b971e50d9d60203392a2005b7ff61050dc000000000000000000000000000000000000000000000000000000005f5260045ffd5b34610163575f7ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc36011261016357335f525f60205260405f2054801561039957335f52600160205260405f2054610371576103307f000000000000000000000000000000000000000000000000000000000000000042611105565b335f5260016020528060405f205560405191825260208201527fd73fc4aafa5101b91e88b8c2fa75d8d6db73466773addb11c079d063c1e63c0c60403392a2005b7fe47b668b000000000000000000000000000000000000000000000000000000005f5260045ffd5b7f1834e265000000000000000000000000000000000000000000000000000000005f5260045ffd5b346101635760207ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc360112610163576104006103fb611051565b6111ba565b6040516003821015610413576020918152f35b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602160045260245ffd5b34610163575f7ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc3601126101635760206040515f8152f35b346101635760407ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc360112610163576104af611074565b6004355f52600260205273ffffffffffffffffffffffffffffffffffffffff60405f2091165f52602052602060ff60405f2054166040519015158152f35b346101635760207ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc3601126101635773ffffffffffffffffffffffffffffffffffffffff610539611051565b165f525f602052602060405f2054604051908152f35b34610163575f7ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc360112610163576020600354604051908152f35b346101635760207ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc360112610163576105c1611051565b6105c9611268565b335f52600160205260405f2054801561028d578042106106995750335f908152602081815260408083208054908490556001909252822091909155906106479082907f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff166112df565b6040519081527f7084f5476618d8e60b11ef0d7d3f06914655adb8793e28ff7f018d4c76d505d560203392a260017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0055005b7fc33c45d1000000000000000000000000000000000000000000000000000000005f5260045260245ffd5b611097565b346101635760407ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc36011261016357610700611074565b3373ffffffffffffffffffffffffffffffffffffffff821603610729576101c89060043561147d565b7f6697b232000000000000000000000000000000000000000000000000000000005f5260045ffd5b346101635760407ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc360112610163576101c860043561078e611074565b906107a86101be825f526002602052600160405f20015490565b6113a9565b346101635760407ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc360112610163576107e4611051565b602435906107f0611268565b8115610a6d5773ffffffffffffffffffffffffffffffffffffffff1690815f52600160205260405f2054610371576040517f70a082310000000000000000000000000000000000000000000000000000000081523060048201527f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16602082602481845afa9182156109e4575f92610a39575b50604051927f23b872dd000000000000000000000000000000000000000000000000000000005f52336004523060245260445260205f60648180855af160015f5114811615610a1a575b836040525f606052156109ef57826024816020937f70a082310000000000000000000000000000000000000000000000000000000082523060048301525afa9182156109e4575f926109ae575b5061095d6040917fd4665e3049283582ba6f9eba07a5b3e12dab49e02da99e8927a47af5d134bea59361113f565b835f525f60205261097181835f2054611105565b845f525f60205280835f205582519182526020820152a260017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0055005b91506020823d6020116109dc575b816109c96020938361114c565b810103126101635790519061095d61092f565b3d91506109bc565b6040513d5f823e3d90fd5b7f5274afe7000000000000000000000000000000000000000000000000000000005f5260045260245ffd5b6001811516610a3057813b15153d1516166108e2565b833d5f823e3d90fd5b9091506020813d602011610a65575b81610a556020938361114c565b8101031261016357519084610898565b3d9150610a48565b7f2c5211c6000000000000000000000000000000000000000000000000000000005f5260045ffd5b34610163575f7ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc3601126101635760206040517f00000000000000000000000000000000000000000000000000000000000000008152f35b346101635760207ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc36011261016357335f9081527fac33ff75c19e70fe83507db0d683fd3465c996598dc972688b7ace676c89077b60205260409020546004359060ff1615610b8f5760407f21dd401bad58f48f88b208aad5157305ac7e8ec5db030042aaec08ebd7f50e4c91600354908060035582519182526020820152a1005b7fe2517d3f000000000000000000000000000000000000000000000000000000005f52336004525f60245260445ffd5b346101635760207ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc360112610163576020610c086004355f526002602052600160405f20015490565b604051908152f35b346101635760207ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc3601126101635773ffffffffffffffffffffffffffffffffffffffff610c5c611051565b165f526001602052602060405f2054604051908152f35b346101635760207ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc360112610163576004357fffffffff00000000000000000000000000000000000000000000000000000000811680910361016357807f7965db0b0000000000000000000000000000000000000000000000000000000060209214908115610d08575b506040519015158152f35b7f01ffc9a70000000000000000000000000000000000000000000000000000000091501482610cfd565b346101635760807ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc36011261016357610d69611051565b60243560443573ffffffffffffffffffffffffffffffffffffffff811692838203610163576064359067ffffffffffffffff821161016357366023830112156101635781600401359067ffffffffffffffff821161016357366024838501011161016357335f9081527f7f9272e0d81bdb9b1e0d5b8b11f4dfdbe7d5ebf3410318d28f3e6024ce33d903602052604090205460ff161561100157610e0b611268565b60045460035480151580610ff8575b610fb7575b5050338614610f8f5773ffffffffffffffffffffffffffffffffffffffff1693845f525f60205260405f2054938415610f6757848211610f6757601f606093610f02847fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe094610eaf827ff204f6a8b6d1439051d5fbd742389d0f778d18d21016e81c8ad3d4558266454c9b61113f565b8b5f525f6020528060405f205515610f54575b4260045573ffffffffffffffffffffffffffffffffffffffff7f0000000000000000000000000000000000000000000000000000000000000000166112df565b80602460405197889687526040602088015282604088015201868601375f85828601015201168101030190a360017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0055005b8a5f5260016020525f6040812055610ec2565b7ff4d678b8000000000000000000000000000000000000000000000000000000005f5260045ffd5b7f8d3fe42e000000000000000000000000000000000000000000000000000000005f5260045ffd5b610fc091611105565b804210610fcd5780610e1f565b7f773faedc000000000000000000000000000000000000000000000000000000005f5260045260245ffd5b50811515610e1a565b7fe2517d3f000000000000000000000000000000000000000000000000000000005f52336004527f5ad6c1ce52091d5f5a49dff6df7d4bf577735e77d8c13251e1ea9c82ce8c380d60245260445ffd5b6004359073ffffffffffffffffffffffffffffffffffffffff8216820361016357565b6024359073ffffffffffffffffffffffffffffffffffffffff8216820361016357565b34610163575f7ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc36011261016357602060405173ffffffffffffffffffffffffffffffffffffffff7f0000000000000000000000000000000000000000000000000000000000000000168152f35b9190820180921161111257565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b9190820391821161111257565b90601f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0910116810190811067ffffffffffffffff82111761118d57604052565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b73ffffffffffffffffffffffffffffffffffffffff16805f525f60205260405f2054156111fc575f52600160205260405f2054156111f757600290565b600190565b505f90565b805f52600260205260405f2073ffffffffffffffffffffffffffffffffffffffff33165f5260205260ff60405f205416156112395750565b7fe2517d3f000000000000000000000000000000000000000000000000000000005f523360045260245260445ffd5b60027f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0054146112b75760027f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0055565b7f3ee5aeb5000000000000000000000000000000000000000000000000000000005f5260045ffd5b9173ffffffffffffffffffffffffffffffffffffffff604051927fa9059cbb000000000000000000000000000000000000000000000000000000005f521660045260245260205f60448180865af19060015f5114821615611388575b604052156113465750565b73ffffffffffffffffffffffffffffffffffffffff907f5274afe7000000000000000000000000000000000000000000000000000000005f521660045260245ffd5b9060018115166113a057823b15153d1516169061133b565b503d5f823e3d90fd5b805f52600260205260405f2073ffffffffffffffffffffffffffffffffffffffff83165f5260205260ff60405f205416155f1461147757805f52600260205260405f2073ffffffffffffffffffffffffffffffffffffffff83165f5260205260405f2060017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0082541617905573ffffffffffffffffffffffffffffffffffffffff339216907f2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d5f80a4600190565b50505f90565b805f52600260205260405f2073ffffffffffffffffffffffffffffffffffffffff83165f5260205260ff60405f2054165f1461147757805f52600260205260405f2073ffffffffffffffffffffffffffffffffffffffff83165f5260205260405f207fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00815416905573ffffffffffffffffffffffffffffffffffffffff339216907ff6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b5f80a460019056fea2646970667358221220d76073a01636f84db6fa5e41c0f05a2695c4df18ed5b02a8e189ed7ed6d4da3a64736f6c63430008220033ac33ff75c19e70fe83507db0d683fd3465c996598dc972688b7ace676c89077b",
}

// AppRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use AppRegistryMetaData.ABI instead.
var AppRegistryABI = AppRegistryMetaData.ABI

// AppRegistryBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use AppRegistryMetaData.Bin instead.
var AppRegistryBin = AppRegistryMetaData.Bin

// DeployAppRegistry deploys a new Ethereum contract, binding an instance of AppRegistry to it.
func DeployAppRegistry(auth *bind.TransactOpts, backend bind.ContractBackend, asset_ common.Address, unlockPeriod_ *big.Int, admin_ common.Address) (common.Address, *types.Transaction, *AppRegistry, error) {
	parsed, err := AppRegistryMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(AppRegistryBin), backend, asset_, unlockPeriod_, admin_)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &AppRegistry{AppRegistryCaller: AppRegistryCaller{contract: contract}, AppRegistryTransactor: AppRegistryTransactor{contract: contract}, AppRegistryFilterer: AppRegistryFilterer{contract: contract}}, nil
}

// AppRegistry is an auto generated Go binding around an Ethereum contract.
type AppRegistry struct {
	AppRegistryCaller     // Read-only binding to the contract
	AppRegistryTransactor // Write-only binding to the contract
	AppRegistryFilterer   // Log filterer for contract events
}

// AppRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type AppRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AppRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AppRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AppRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AppRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AppRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AppRegistrySession struct {
	Contract     *AppRegistry      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AppRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AppRegistryCallerSession struct {
	Contract *AppRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// AppRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AppRegistryTransactorSession struct {
	Contract     *AppRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// AppRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type AppRegistryRaw struct {
	Contract *AppRegistry // Generic contract binding to access the raw methods on
}

// AppRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AppRegistryCallerRaw struct {
	Contract *AppRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// AppRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AppRegistryTransactorRaw struct {
	Contract *AppRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAppRegistry creates a new instance of AppRegistry, bound to a specific deployed contract.
func NewAppRegistry(address common.Address, backend bind.ContractBackend) (*AppRegistry, error) {
	contract, err := bindAppRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AppRegistry{AppRegistryCaller: AppRegistryCaller{contract: contract}, AppRegistryTransactor: AppRegistryTransactor{contract: contract}, AppRegistryFilterer: AppRegistryFilterer{contract: contract}}, nil
}

// NewAppRegistryCaller creates a new read-only instance of AppRegistry, bound to a specific deployed contract.
func NewAppRegistryCaller(address common.Address, caller bind.ContractCaller) (*AppRegistryCaller, error) {
	contract, err := bindAppRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AppRegistryCaller{contract: contract}, nil
}

// NewAppRegistryTransactor creates a new write-only instance of AppRegistry, bound to a specific deployed contract.
func NewAppRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*AppRegistryTransactor, error) {
	contract, err := bindAppRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AppRegistryTransactor{contract: contract}, nil
}

// NewAppRegistryFilterer creates a new log filterer instance of AppRegistry, bound to a specific deployed contract.
func NewAppRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*AppRegistryFilterer, error) {
	contract, err := bindAppRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AppRegistryFilterer{contract: contract}, nil
}

// bindAppRegistry binds a generic wrapper to an already deployed contract.
func bindAppRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AppRegistryABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AppRegistry *AppRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AppRegistry.Contract.AppRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AppRegistry *AppRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AppRegistry.Contract.AppRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AppRegistry *AppRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AppRegistry.Contract.AppRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AppRegistry *AppRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AppRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AppRegistry *AppRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AppRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AppRegistry *AppRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AppRegistry.Contract.contract.Transact(opts, method, params...)
}

// ADJUDICATORROLE is a free data retrieval call binding the contract method 0xf538f7dd.
//
// Solidity: function ADJUDICATOR_ROLE() view returns(bytes32)
func (_AppRegistry *AppRegistryCaller) ADJUDICATORROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _AppRegistry.contract.Call(opts, &out, "ADJUDICATOR_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ADJUDICATORROLE is a free data retrieval call binding the contract method 0xf538f7dd.
//
// Solidity: function ADJUDICATOR_ROLE() view returns(bytes32)
func (_AppRegistry *AppRegistrySession) ADJUDICATORROLE() ([32]byte, error) {
	return _AppRegistry.Contract.ADJUDICATORROLE(&_AppRegistry.CallOpts)
}

// ADJUDICATORROLE is a free data retrieval call binding the contract method 0xf538f7dd.
//
// Solidity: function ADJUDICATOR_ROLE() view returns(bytes32)
func (_AppRegistry *AppRegistryCallerSession) ADJUDICATORROLE() ([32]byte, error) {
	return _AppRegistry.Contract.ADJUDICATORROLE(&_AppRegistry.CallOpts)
}

// ASSET is a free data retrieval call binding the contract method 0x4800d97f.
//
// Solidity: function ASSET() view returns(address)
func (_AppRegistry *AppRegistryCaller) ASSET(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AppRegistry.contract.Call(opts, &out, "ASSET")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ASSET is a free data retrieval call binding the contract method 0x4800d97f.
//
// Solidity: function ASSET() view returns(address)
func (_AppRegistry *AppRegistrySession) ASSET() (common.Address, error) {
	return _AppRegistry.Contract.ASSET(&_AppRegistry.CallOpts)
}

// ASSET is a free data retrieval call binding the contract method 0x4800d97f.
//
// Solidity: function ASSET() view returns(address)
func (_AppRegistry *AppRegistryCallerSession) ASSET() (common.Address, error) {
	return _AppRegistry.Contract.ASSET(&_AppRegistry.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_AppRegistry *AppRegistryCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _AppRegistry.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_AppRegistry *AppRegistrySession) DEFAULTADMINROLE() ([32]byte, error) {
	return _AppRegistry.Contract.DEFAULTADMINROLE(&_AppRegistry.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_AppRegistry *AppRegistryCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _AppRegistry.Contract.DEFAULTADMINROLE(&_AppRegistry.CallOpts)
}

// UNLOCKPERIOD is a free data retrieval call binding the contract method 0x259a28cf.
//
// Solidity: function UNLOCK_PERIOD() view returns(uint256)
func (_AppRegistry *AppRegistryCaller) UNLOCKPERIOD(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AppRegistry.contract.Call(opts, &out, "UNLOCK_PERIOD")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// UNLOCKPERIOD is a free data retrieval call binding the contract method 0x259a28cf.
//
// Solidity: function UNLOCK_PERIOD() view returns(uint256)
func (_AppRegistry *AppRegistrySession) UNLOCKPERIOD() (*big.Int, error) {
	return _AppRegistry.Contract.UNLOCKPERIOD(&_AppRegistry.CallOpts)
}

// UNLOCKPERIOD is a free data retrieval call binding the contract method 0x259a28cf.
//
// Solidity: function UNLOCK_PERIOD() view returns(uint256)
func (_AppRegistry *AppRegistryCallerSession) UNLOCKPERIOD() (*big.Int, error) {
	return _AppRegistry.Contract.UNLOCKPERIOD(&_AppRegistry.CallOpts)
}

// Asset is a free data retrieval call binding the contract method 0x38d52e0f.
//
// Solidity: function asset() view returns(address)
func (_AppRegistry *AppRegistryCaller) Asset(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AppRegistry.contract.Call(opts, &out, "asset")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Asset is a free data retrieval call binding the contract method 0x38d52e0f.
//
// Solidity: function asset() view returns(address)
func (_AppRegistry *AppRegistrySession) Asset() (common.Address, error) {
	return _AppRegistry.Contract.Asset(&_AppRegistry.CallOpts)
}

// Asset is a free data retrieval call binding the contract method 0x38d52e0f.
//
// Solidity: function asset() view returns(address)
func (_AppRegistry *AppRegistryCallerSession) Asset() (common.Address, error) {
	return _AppRegistry.Contract.Asset(&_AppRegistry.CallOpts)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address user) view returns(uint256)
func (_AppRegistry *AppRegistryCaller) BalanceOf(opts *bind.CallOpts, user common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AppRegistry.contract.Call(opts, &out, "balanceOf", user)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address user) view returns(uint256)
func (_AppRegistry *AppRegistrySession) BalanceOf(user common.Address) (*big.Int, error) {
	return _AppRegistry.Contract.BalanceOf(&_AppRegistry.CallOpts, user)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address user) view returns(uint256)
func (_AppRegistry *AppRegistryCallerSession) BalanceOf(user common.Address) (*big.Int, error) {
	return _AppRegistry.Contract.BalanceOf(&_AppRegistry.CallOpts, user)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_AppRegistry *AppRegistryCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _AppRegistry.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_AppRegistry *AppRegistrySession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _AppRegistry.Contract.GetRoleAdmin(&_AppRegistry.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_AppRegistry *AppRegistryCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _AppRegistry.Contract.GetRoleAdmin(&_AppRegistry.CallOpts, role)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_AppRegistry *AppRegistryCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _AppRegistry.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_AppRegistry *AppRegistrySession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _AppRegistry.Contract.HasRole(&_AppRegistry.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_AppRegistry *AppRegistryCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _AppRegistry.Contract.HasRole(&_AppRegistry.CallOpts, role, account)
}

// LastSlashTimestamp is a free data retrieval call binding the contract method 0xc77f5062.
//
// Solidity: function lastSlashTimestamp() view returns(uint256)
func (_AppRegistry *AppRegistryCaller) LastSlashTimestamp(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AppRegistry.contract.Call(opts, &out, "lastSlashTimestamp")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastSlashTimestamp is a free data retrieval call binding the contract method 0xc77f5062.
//
// Solidity: function lastSlashTimestamp() view returns(uint256)
func (_AppRegistry *AppRegistrySession) LastSlashTimestamp() (*big.Int, error) {
	return _AppRegistry.Contract.LastSlashTimestamp(&_AppRegistry.CallOpts)
}

// LastSlashTimestamp is a free data retrieval call binding the contract method 0xc77f5062.
//
// Solidity: function lastSlashTimestamp() view returns(uint256)
func (_AppRegistry *AppRegistryCallerSession) LastSlashTimestamp() (*big.Int, error) {
	return _AppRegistry.Contract.LastSlashTimestamp(&_AppRegistry.CallOpts)
}

// LockStateOf is a free data retrieval call binding the contract method 0xa43e9cf4.
//
// Solidity: function lockStateOf(address user) view returns(uint8)
func (_AppRegistry *AppRegistryCaller) LockStateOf(opts *bind.CallOpts, user common.Address) (uint8, error) {
	var out []interface{}
	err := _AppRegistry.contract.Call(opts, &out, "lockStateOf", user)

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// LockStateOf is a free data retrieval call binding the contract method 0xa43e9cf4.
//
// Solidity: function lockStateOf(address user) view returns(uint8)
func (_AppRegistry *AppRegistrySession) LockStateOf(user common.Address) (uint8, error) {
	return _AppRegistry.Contract.LockStateOf(&_AppRegistry.CallOpts, user)
}

// LockStateOf is a free data retrieval call binding the contract method 0xa43e9cf4.
//
// Solidity: function lockStateOf(address user) view returns(uint8)
func (_AppRegistry *AppRegistryCallerSession) LockStateOf(user common.Address) (uint8, error) {
	return _AppRegistry.Contract.LockStateOf(&_AppRegistry.CallOpts, user)
}

// SlashCooldown is a free data retrieval call binding the contract method 0x5a6ca4ae.
//
// Solidity: function slashCooldown() view returns(uint256)
func (_AppRegistry *AppRegistryCaller) SlashCooldown(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AppRegistry.contract.Call(opts, &out, "slashCooldown")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SlashCooldown is a free data retrieval call binding the contract method 0x5a6ca4ae.
//
// Solidity: function slashCooldown() view returns(uint256)
func (_AppRegistry *AppRegistrySession) SlashCooldown() (*big.Int, error) {
	return _AppRegistry.Contract.SlashCooldown(&_AppRegistry.CallOpts)
}

// SlashCooldown is a free data retrieval call binding the contract method 0x5a6ca4ae.
//
// Solidity: function slashCooldown() view returns(uint256)
func (_AppRegistry *AppRegistryCallerSession) SlashCooldown() (*big.Int, error) {
	return _AppRegistry.Contract.SlashCooldown(&_AppRegistry.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AppRegistry *AppRegistryCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _AppRegistry.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AppRegistry *AppRegistrySession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _AppRegistry.Contract.SupportsInterface(&_AppRegistry.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_AppRegistry *AppRegistryCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _AppRegistry.Contract.SupportsInterface(&_AppRegistry.CallOpts, interfaceId)
}

// UnlockTimestampOf is a free data retrieval call binding the contract method 0x02f70fe6.
//
// Solidity: function unlockTimestampOf(address user) view returns(uint256)
func (_AppRegistry *AppRegistryCaller) UnlockTimestampOf(opts *bind.CallOpts, user common.Address) (*big.Int, error) {
	var out []interface{}
	err := _AppRegistry.contract.Call(opts, &out, "unlockTimestampOf", user)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// UnlockTimestampOf is a free data retrieval call binding the contract method 0x02f70fe6.
//
// Solidity: function unlockTimestampOf(address user) view returns(uint256)
func (_AppRegistry *AppRegistrySession) UnlockTimestampOf(user common.Address) (*big.Int, error) {
	return _AppRegistry.Contract.UnlockTimestampOf(&_AppRegistry.CallOpts, user)
}

// UnlockTimestampOf is a free data retrieval call binding the contract method 0x02f70fe6.
//
// Solidity: function unlockTimestampOf(address user) view returns(uint256)
func (_AppRegistry *AppRegistryCallerSession) UnlockTimestampOf(user common.Address) (*big.Int, error) {
	return _AppRegistry.Contract.UnlockTimestampOf(&_AppRegistry.CallOpts, user)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_AppRegistry *AppRegistryTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AppRegistry.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_AppRegistry *AppRegistrySession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AppRegistry.Contract.GrantRole(&_AppRegistry.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_AppRegistry *AppRegistryTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AppRegistry.Contract.GrantRole(&_AppRegistry.TransactOpts, role, account)
}

// Lock is a paid mutator transaction binding the contract method 0x282d3fdf.
//
// Solidity: function lock(address target, uint256 amount) returns()
func (_AppRegistry *AppRegistryTransactor) Lock(opts *bind.TransactOpts, target common.Address, amount *big.Int) (*types.Transaction, error) {
	return _AppRegistry.contract.Transact(opts, "lock", target, amount)
}

// Lock is a paid mutator transaction binding the contract method 0x282d3fdf.
//
// Solidity: function lock(address target, uint256 amount) returns()
func (_AppRegistry *AppRegistrySession) Lock(target common.Address, amount *big.Int) (*types.Transaction, error) {
	return _AppRegistry.Contract.Lock(&_AppRegistry.TransactOpts, target, amount)
}

// Lock is a paid mutator transaction binding the contract method 0x282d3fdf.
//
// Solidity: function lock(address target, uint256 amount) returns()
func (_AppRegistry *AppRegistryTransactorSession) Lock(target common.Address, amount *big.Int) (*types.Transaction, error) {
	return _AppRegistry.Contract.Lock(&_AppRegistry.TransactOpts, target, amount)
}

// Relock is a paid mutator transaction binding the contract method 0xc53b573d.
//
// Solidity: function relock() returns()
func (_AppRegistry *AppRegistryTransactor) Relock(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AppRegistry.contract.Transact(opts, "relock")
}

// Relock is a paid mutator transaction binding the contract method 0xc53b573d.
//
// Solidity: function relock() returns()
func (_AppRegistry *AppRegistrySession) Relock() (*types.Transaction, error) {
	return _AppRegistry.Contract.Relock(&_AppRegistry.TransactOpts)
}

// Relock is a paid mutator transaction binding the contract method 0xc53b573d.
//
// Solidity: function relock() returns()
func (_AppRegistry *AppRegistryTransactorSession) Relock() (*types.Transaction, error) {
	return _AppRegistry.Contract.Relock(&_AppRegistry.TransactOpts)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_AppRegistry *AppRegistryTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _AppRegistry.contract.Transact(opts, "renounceRole", role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_AppRegistry *AppRegistrySession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _AppRegistry.Contract.RenounceRole(&_AppRegistry.TransactOpts, role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_AppRegistry *AppRegistryTransactorSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _AppRegistry.Contract.RenounceRole(&_AppRegistry.TransactOpts, role, callerConfirmation)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_AppRegistry *AppRegistryTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AppRegistry.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_AppRegistry *AppRegistrySession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AppRegistry.Contract.RevokeRole(&_AppRegistry.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_AppRegistry *AppRegistryTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _AppRegistry.Contract.RevokeRole(&_AppRegistry.TransactOpts, role, account)
}

// SetSlashCooldown is a paid mutator transaction binding the contract method 0x256f8f09.
//
// Solidity: function setSlashCooldown(uint256 newCooldown) returns()
func (_AppRegistry *AppRegistryTransactor) SetSlashCooldown(opts *bind.TransactOpts, newCooldown *big.Int) (*types.Transaction, error) {
	return _AppRegistry.contract.Transact(opts, "setSlashCooldown", newCooldown)
}

// SetSlashCooldown is a paid mutator transaction binding the contract method 0x256f8f09.
//
// Solidity: function setSlashCooldown(uint256 newCooldown) returns()
func (_AppRegistry *AppRegistrySession) SetSlashCooldown(newCooldown *big.Int) (*types.Transaction, error) {
	return _AppRegistry.Contract.SetSlashCooldown(&_AppRegistry.TransactOpts, newCooldown)
}

// SetSlashCooldown is a paid mutator transaction binding the contract method 0x256f8f09.
//
// Solidity: function setSlashCooldown(uint256 newCooldown) returns()
func (_AppRegistry *AppRegistryTransactorSession) SetSlashCooldown(newCooldown *big.Int) (*types.Transaction, error) {
	return _AppRegistry.Contract.SetSlashCooldown(&_AppRegistry.TransactOpts, newCooldown)
}

// Slash is a paid mutator transaction binding the contract method 0x010379b2.
//
// Solidity: function slash(address user, uint256 amount, address recipient, bytes decision) returns()
func (_AppRegistry *AppRegistryTransactor) Slash(opts *bind.TransactOpts, user common.Address, amount *big.Int, recipient common.Address, decision []byte) (*types.Transaction, error) {
	return _AppRegistry.contract.Transact(opts, "slash", user, amount, recipient, decision)
}

// Slash is a paid mutator transaction binding the contract method 0x010379b2.
//
// Solidity: function slash(address user, uint256 amount, address recipient, bytes decision) returns()
func (_AppRegistry *AppRegistrySession) Slash(user common.Address, amount *big.Int, recipient common.Address, decision []byte) (*types.Transaction, error) {
	return _AppRegistry.Contract.Slash(&_AppRegistry.TransactOpts, user, amount, recipient, decision)
}

// Slash is a paid mutator transaction binding the contract method 0x010379b2.
//
// Solidity: function slash(address user, uint256 amount, address recipient, bytes decision) returns()
func (_AppRegistry *AppRegistryTransactorSession) Slash(user common.Address, amount *big.Int, recipient common.Address, decision []byte) (*types.Transaction, error) {
	return _AppRegistry.Contract.Slash(&_AppRegistry.TransactOpts, user, amount, recipient, decision)
}

// Unlock is a paid mutator transaction binding the contract method 0xa69df4b5.
//
// Solidity: function unlock() returns()
func (_AppRegistry *AppRegistryTransactor) Unlock(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AppRegistry.contract.Transact(opts, "unlock")
}

// Unlock is a paid mutator transaction binding the contract method 0xa69df4b5.
//
// Solidity: function unlock() returns()
func (_AppRegistry *AppRegistrySession) Unlock() (*types.Transaction, error) {
	return _AppRegistry.Contract.Unlock(&_AppRegistry.TransactOpts)
}

// Unlock is a paid mutator transaction binding the contract method 0xa69df4b5.
//
// Solidity: function unlock() returns()
func (_AppRegistry *AppRegistryTransactorSession) Unlock() (*types.Transaction, error) {
	return _AppRegistry.Contract.Unlock(&_AppRegistry.TransactOpts)
}

// Withdraw is a paid mutator transaction binding the contract method 0x51cff8d9.
//
// Solidity: function withdraw(address destination) returns()
func (_AppRegistry *AppRegistryTransactor) Withdraw(opts *bind.TransactOpts, destination common.Address) (*types.Transaction, error) {
	return _AppRegistry.contract.Transact(opts, "withdraw", destination)
}

// Withdraw is a paid mutator transaction binding the contract method 0x51cff8d9.
//
// Solidity: function withdraw(address destination) returns()
func (_AppRegistry *AppRegistrySession) Withdraw(destination common.Address) (*types.Transaction, error) {
	return _AppRegistry.Contract.Withdraw(&_AppRegistry.TransactOpts, destination)
}

// Withdraw is a paid mutator transaction binding the contract method 0x51cff8d9.
//
// Solidity: function withdraw(address destination) returns()
func (_AppRegistry *AppRegistryTransactorSession) Withdraw(destination common.Address) (*types.Transaction, error) {
	return _AppRegistry.Contract.Withdraw(&_AppRegistry.TransactOpts, destination)
}

// AppRegistryLockedIterator is returned from FilterLocked and is used to iterate over the raw logs and unpacked data for Locked events raised by the AppRegistry contract.
type AppRegistryLockedIterator struct {
	Event *AppRegistryLocked // Event containing the contract specifics and raw log

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
func (it *AppRegistryLockedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AppRegistryLocked)
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
		it.Event = new(AppRegistryLocked)
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
func (it *AppRegistryLockedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AppRegistryLockedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AppRegistryLocked represents a Locked event raised by the AppRegistry contract.
type AppRegistryLocked struct {
	User       common.Address
	Deposited  *big.Int
	NewBalance *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterLocked is a free log retrieval operation binding the contract event 0xd4665e3049283582ba6f9eba07a5b3e12dab49e02da99e8927a47af5d134bea5.
//
// Solidity: event Locked(address indexed user, uint256 deposited, uint256 newBalance)
func (_AppRegistry *AppRegistryFilterer) FilterLocked(opts *bind.FilterOpts, user []common.Address) (*AppRegistryLockedIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _AppRegistry.contract.FilterLogs(opts, "Locked", userRule)
	if err != nil {
		return nil, err
	}
	return &AppRegistryLockedIterator{contract: _AppRegistry.contract, event: "Locked", logs: logs, sub: sub}, nil
}

// WatchLocked is a free log subscription operation binding the contract event 0xd4665e3049283582ba6f9eba07a5b3e12dab49e02da99e8927a47af5d134bea5.
//
// Solidity: event Locked(address indexed user, uint256 deposited, uint256 newBalance)
func (_AppRegistry *AppRegistryFilterer) WatchLocked(opts *bind.WatchOpts, sink chan<- *AppRegistryLocked, user []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _AppRegistry.contract.WatchLogs(opts, "Locked", userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AppRegistryLocked)
				if err := _AppRegistry.contract.UnpackLog(event, "Locked", log); err != nil {
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

// ParseLocked is a log parse operation binding the contract event 0xd4665e3049283582ba6f9eba07a5b3e12dab49e02da99e8927a47af5d134bea5.
//
// Solidity: event Locked(address indexed user, uint256 deposited, uint256 newBalance)
func (_AppRegistry *AppRegistryFilterer) ParseLocked(log types.Log) (*AppRegistryLocked, error) {
	event := new(AppRegistryLocked)
	if err := _AppRegistry.contract.UnpackLog(event, "Locked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AppRegistryRelockedIterator is returned from FilterRelocked and is used to iterate over the raw logs and unpacked data for Relocked events raised by the AppRegistry contract.
type AppRegistryRelockedIterator struct {
	Event *AppRegistryRelocked // Event containing the contract specifics and raw log

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
func (it *AppRegistryRelockedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AppRegistryRelocked)
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
		it.Event = new(AppRegistryRelocked)
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
func (it *AppRegistryRelockedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AppRegistryRelockedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AppRegistryRelocked represents a Relocked event raised by the AppRegistry contract.
type AppRegistryRelocked struct {
	User    common.Address
	Balance *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRelocked is a free log retrieval operation binding the contract event 0xe2d155d9d74decc198a9a9b4f5bddb24ee0842ee745c9ce57e3573b971e50d9d.
//
// Solidity: event Relocked(address indexed user, uint256 balance)
func (_AppRegistry *AppRegistryFilterer) FilterRelocked(opts *bind.FilterOpts, user []common.Address) (*AppRegistryRelockedIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _AppRegistry.contract.FilterLogs(opts, "Relocked", userRule)
	if err != nil {
		return nil, err
	}
	return &AppRegistryRelockedIterator{contract: _AppRegistry.contract, event: "Relocked", logs: logs, sub: sub}, nil
}

// WatchRelocked is a free log subscription operation binding the contract event 0xe2d155d9d74decc198a9a9b4f5bddb24ee0842ee745c9ce57e3573b971e50d9d.
//
// Solidity: event Relocked(address indexed user, uint256 balance)
func (_AppRegistry *AppRegistryFilterer) WatchRelocked(opts *bind.WatchOpts, sink chan<- *AppRegistryRelocked, user []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _AppRegistry.contract.WatchLogs(opts, "Relocked", userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AppRegistryRelocked)
				if err := _AppRegistry.contract.UnpackLog(event, "Relocked", log); err != nil {
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

// ParseRelocked is a log parse operation binding the contract event 0xe2d155d9d74decc198a9a9b4f5bddb24ee0842ee745c9ce57e3573b971e50d9d.
//
// Solidity: event Relocked(address indexed user, uint256 balance)
func (_AppRegistry *AppRegistryFilterer) ParseRelocked(log types.Log) (*AppRegistryRelocked, error) {
	event := new(AppRegistryRelocked)
	if err := _AppRegistry.contract.UnpackLog(event, "Relocked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AppRegistryRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the AppRegistry contract.
type AppRegistryRoleAdminChangedIterator struct {
	Event *AppRegistryRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *AppRegistryRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AppRegistryRoleAdminChanged)
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
		it.Event = new(AppRegistryRoleAdminChanged)
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
func (it *AppRegistryRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AppRegistryRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AppRegistryRoleAdminChanged represents a RoleAdminChanged event raised by the AppRegistry contract.
type AppRegistryRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_AppRegistry *AppRegistryFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*AppRegistryRoleAdminChangedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _AppRegistry.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &AppRegistryRoleAdminChangedIterator{contract: _AppRegistry.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_AppRegistry *AppRegistryFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *AppRegistryRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _AppRegistry.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AppRegistryRoleAdminChanged)
				if err := _AppRegistry.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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

// ParseRoleAdminChanged is a log parse operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_AppRegistry *AppRegistryFilterer) ParseRoleAdminChanged(log types.Log) (*AppRegistryRoleAdminChanged, error) {
	event := new(AppRegistryRoleAdminChanged)
	if err := _AppRegistry.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AppRegistryRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the AppRegistry contract.
type AppRegistryRoleGrantedIterator struct {
	Event *AppRegistryRoleGranted // Event containing the contract specifics and raw log

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
func (it *AppRegistryRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AppRegistryRoleGranted)
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
		it.Event = new(AppRegistryRoleGranted)
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
func (it *AppRegistryRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AppRegistryRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AppRegistryRoleGranted represents a RoleGranted event raised by the AppRegistry contract.
type AppRegistryRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_AppRegistry *AppRegistryFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*AppRegistryRoleGrantedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _AppRegistry.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &AppRegistryRoleGrantedIterator{contract: _AppRegistry.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_AppRegistry *AppRegistryFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *AppRegistryRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _AppRegistry.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AppRegistryRoleGranted)
				if err := _AppRegistry.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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

// ParseRoleGranted is a log parse operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_AppRegistry *AppRegistryFilterer) ParseRoleGranted(log types.Log) (*AppRegistryRoleGranted, error) {
	event := new(AppRegistryRoleGranted)
	if err := _AppRegistry.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AppRegistryRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the AppRegistry contract.
type AppRegistryRoleRevokedIterator struct {
	Event *AppRegistryRoleRevoked // Event containing the contract specifics and raw log

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
func (it *AppRegistryRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AppRegistryRoleRevoked)
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
		it.Event = new(AppRegistryRoleRevoked)
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
func (it *AppRegistryRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AppRegistryRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AppRegistryRoleRevoked represents a RoleRevoked event raised by the AppRegistry contract.
type AppRegistryRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_AppRegistry *AppRegistryFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*AppRegistryRoleRevokedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _AppRegistry.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &AppRegistryRoleRevokedIterator{contract: _AppRegistry.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_AppRegistry *AppRegistryFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *AppRegistryRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _AppRegistry.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AppRegistryRoleRevoked)
				if err := _AppRegistry.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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

// ParseRoleRevoked is a log parse operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_AppRegistry *AppRegistryFilterer) ParseRoleRevoked(log types.Log) (*AppRegistryRoleRevoked, error) {
	event := new(AppRegistryRoleRevoked)
	if err := _AppRegistry.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AppRegistrySlashCooldownUpdatedIterator is returned from FilterSlashCooldownUpdated and is used to iterate over the raw logs and unpacked data for SlashCooldownUpdated events raised by the AppRegistry contract.
type AppRegistrySlashCooldownUpdatedIterator struct {
	Event *AppRegistrySlashCooldownUpdated // Event containing the contract specifics and raw log

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
func (it *AppRegistrySlashCooldownUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AppRegistrySlashCooldownUpdated)
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
		it.Event = new(AppRegistrySlashCooldownUpdated)
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
func (it *AppRegistrySlashCooldownUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AppRegistrySlashCooldownUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AppRegistrySlashCooldownUpdated represents a SlashCooldownUpdated event raised by the AppRegistry contract.
type AppRegistrySlashCooldownUpdated struct {
	OldCooldown *big.Int
	NewCooldown *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterSlashCooldownUpdated is a free log retrieval operation binding the contract event 0x21dd401bad58f48f88b208aad5157305ac7e8ec5db030042aaec08ebd7f50e4c.
//
// Solidity: event SlashCooldownUpdated(uint256 oldCooldown, uint256 newCooldown)
func (_AppRegistry *AppRegistryFilterer) FilterSlashCooldownUpdated(opts *bind.FilterOpts) (*AppRegistrySlashCooldownUpdatedIterator, error) {

	logs, sub, err := _AppRegistry.contract.FilterLogs(opts, "SlashCooldownUpdated")
	if err != nil {
		return nil, err
	}
	return &AppRegistrySlashCooldownUpdatedIterator{contract: _AppRegistry.contract, event: "SlashCooldownUpdated", logs: logs, sub: sub}, nil
}

// WatchSlashCooldownUpdated is a free log subscription operation binding the contract event 0x21dd401bad58f48f88b208aad5157305ac7e8ec5db030042aaec08ebd7f50e4c.
//
// Solidity: event SlashCooldownUpdated(uint256 oldCooldown, uint256 newCooldown)
func (_AppRegistry *AppRegistryFilterer) WatchSlashCooldownUpdated(opts *bind.WatchOpts, sink chan<- *AppRegistrySlashCooldownUpdated) (event.Subscription, error) {

	logs, sub, err := _AppRegistry.contract.WatchLogs(opts, "SlashCooldownUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AppRegistrySlashCooldownUpdated)
				if err := _AppRegistry.contract.UnpackLog(event, "SlashCooldownUpdated", log); err != nil {
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

// ParseSlashCooldownUpdated is a log parse operation binding the contract event 0x21dd401bad58f48f88b208aad5157305ac7e8ec5db030042aaec08ebd7f50e4c.
//
// Solidity: event SlashCooldownUpdated(uint256 oldCooldown, uint256 newCooldown)
func (_AppRegistry *AppRegistryFilterer) ParseSlashCooldownUpdated(log types.Log) (*AppRegistrySlashCooldownUpdated, error) {
	event := new(AppRegistrySlashCooldownUpdated)
	if err := _AppRegistry.contract.UnpackLog(event, "SlashCooldownUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AppRegistrySlashedIterator is returned from FilterSlashed and is used to iterate over the raw logs and unpacked data for Slashed events raised by the AppRegistry contract.
type AppRegistrySlashedIterator struct {
	Event *AppRegistrySlashed // Event containing the contract specifics and raw log

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
func (it *AppRegistrySlashedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AppRegistrySlashed)
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
		it.Event = new(AppRegistrySlashed)
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
func (it *AppRegistrySlashedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AppRegistrySlashedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AppRegistrySlashed represents a Slashed event raised by the AppRegistry contract.
type AppRegistrySlashed struct {
	User      common.Address
	Amount    *big.Int
	Recipient common.Address
	Decision  []byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterSlashed is a free log retrieval operation binding the contract event 0xf204f6a8b6d1439051d5fbd742389d0f778d18d21016e81c8ad3d4558266454c.
//
// Solidity: event Slashed(address indexed user, uint256 amount, address indexed recipient, bytes decision)
func (_AppRegistry *AppRegistryFilterer) FilterSlashed(opts *bind.FilterOpts, user []common.Address, recipient []common.Address) (*AppRegistrySlashedIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _AppRegistry.contract.FilterLogs(opts, "Slashed", userRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return &AppRegistrySlashedIterator{contract: _AppRegistry.contract, event: "Slashed", logs: logs, sub: sub}, nil
}

// WatchSlashed is a free log subscription operation binding the contract event 0xf204f6a8b6d1439051d5fbd742389d0f778d18d21016e81c8ad3d4558266454c.
//
// Solidity: event Slashed(address indexed user, uint256 amount, address indexed recipient, bytes decision)
func (_AppRegistry *AppRegistryFilterer) WatchSlashed(opts *bind.WatchOpts, sink chan<- *AppRegistrySlashed, user []common.Address, recipient []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _AppRegistry.contract.WatchLogs(opts, "Slashed", userRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AppRegistrySlashed)
				if err := _AppRegistry.contract.UnpackLog(event, "Slashed", log); err != nil {
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

// ParseSlashed is a log parse operation binding the contract event 0xf204f6a8b6d1439051d5fbd742389d0f778d18d21016e81c8ad3d4558266454c.
//
// Solidity: event Slashed(address indexed user, uint256 amount, address indexed recipient, bytes decision)
func (_AppRegistry *AppRegistryFilterer) ParseSlashed(log types.Log) (*AppRegistrySlashed, error) {
	event := new(AppRegistrySlashed)
	if err := _AppRegistry.contract.UnpackLog(event, "Slashed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AppRegistryUnlockInitiatedIterator is returned from FilterUnlockInitiated and is used to iterate over the raw logs and unpacked data for UnlockInitiated events raised by the AppRegistry contract.
type AppRegistryUnlockInitiatedIterator struct {
	Event *AppRegistryUnlockInitiated // Event containing the contract specifics and raw log

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
func (it *AppRegistryUnlockInitiatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AppRegistryUnlockInitiated)
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
		it.Event = new(AppRegistryUnlockInitiated)
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
func (it *AppRegistryUnlockInitiatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AppRegistryUnlockInitiatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AppRegistryUnlockInitiated represents a UnlockInitiated event raised by the AppRegistry contract.
type AppRegistryUnlockInitiated struct {
	User        common.Address
	Balance     *big.Int
	AvailableAt *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterUnlockInitiated is a free log retrieval operation binding the contract event 0xd73fc4aafa5101b91e88b8c2fa75d8d6db73466773addb11c079d063c1e63c0c.
//
// Solidity: event UnlockInitiated(address indexed user, uint256 balance, uint256 availableAt)
func (_AppRegistry *AppRegistryFilterer) FilterUnlockInitiated(opts *bind.FilterOpts, user []common.Address) (*AppRegistryUnlockInitiatedIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _AppRegistry.contract.FilterLogs(opts, "UnlockInitiated", userRule)
	if err != nil {
		return nil, err
	}
	return &AppRegistryUnlockInitiatedIterator{contract: _AppRegistry.contract, event: "UnlockInitiated", logs: logs, sub: sub}, nil
}

// WatchUnlockInitiated is a free log subscription operation binding the contract event 0xd73fc4aafa5101b91e88b8c2fa75d8d6db73466773addb11c079d063c1e63c0c.
//
// Solidity: event UnlockInitiated(address indexed user, uint256 balance, uint256 availableAt)
func (_AppRegistry *AppRegistryFilterer) WatchUnlockInitiated(opts *bind.WatchOpts, sink chan<- *AppRegistryUnlockInitiated, user []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _AppRegistry.contract.WatchLogs(opts, "UnlockInitiated", userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AppRegistryUnlockInitiated)
				if err := _AppRegistry.contract.UnpackLog(event, "UnlockInitiated", log); err != nil {
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

// ParseUnlockInitiated is a log parse operation binding the contract event 0xd73fc4aafa5101b91e88b8c2fa75d8d6db73466773addb11c079d063c1e63c0c.
//
// Solidity: event UnlockInitiated(address indexed user, uint256 balance, uint256 availableAt)
func (_AppRegistry *AppRegistryFilterer) ParseUnlockInitiated(log types.Log) (*AppRegistryUnlockInitiated, error) {
	event := new(AppRegistryUnlockInitiated)
	if err := _AppRegistry.contract.UnpackLog(event, "UnlockInitiated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AppRegistryWithdrawnIterator is returned from FilterWithdrawn and is used to iterate over the raw logs and unpacked data for Withdrawn events raised by the AppRegistry contract.
type AppRegistryWithdrawnIterator struct {
	Event *AppRegistryWithdrawn // Event containing the contract specifics and raw log

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
func (it *AppRegistryWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AppRegistryWithdrawn)
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
		it.Event = new(AppRegistryWithdrawn)
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
func (it *AppRegistryWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AppRegistryWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AppRegistryWithdrawn represents a Withdrawn event raised by the AppRegistry contract.
type AppRegistryWithdrawn struct {
	User    common.Address
	Balance *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterWithdrawn is a free log retrieval operation binding the contract event 0x7084f5476618d8e60b11ef0d7d3f06914655adb8793e28ff7f018d4c76d505d5.
//
// Solidity: event Withdrawn(address indexed user, uint256 balance)
func (_AppRegistry *AppRegistryFilterer) FilterWithdrawn(opts *bind.FilterOpts, user []common.Address) (*AppRegistryWithdrawnIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _AppRegistry.contract.FilterLogs(opts, "Withdrawn", userRule)
	if err != nil {
		return nil, err
	}
	return &AppRegistryWithdrawnIterator{contract: _AppRegistry.contract, event: "Withdrawn", logs: logs, sub: sub}, nil
}

// WatchWithdrawn is a free log subscription operation binding the contract event 0x7084f5476618d8e60b11ef0d7d3f06914655adb8793e28ff7f018d4c76d505d5.
//
// Solidity: event Withdrawn(address indexed user, uint256 balance)
func (_AppRegistry *AppRegistryFilterer) WatchWithdrawn(opts *bind.WatchOpts, sink chan<- *AppRegistryWithdrawn, user []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _AppRegistry.contract.WatchLogs(opts, "Withdrawn", userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AppRegistryWithdrawn)
				if err := _AppRegistry.contract.UnpackLog(event, "Withdrawn", log); err != nil {
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

// ParseWithdrawn is a log parse operation binding the contract event 0x7084f5476618d8e60b11ef0d7d3f06914655adb8793e28ff7f018d4c76d505d5.
//
// Solidity: event Withdrawn(address indexed user, uint256 balance)
func (_AppRegistry *AppRegistryFilterer) ParseWithdrawn(log types.Log) (*AppRegistryWithdrawn, error) {
	event := new(AppRegistryWithdrawn)
	if err := _AppRegistry.contract.UnpackLog(event, "Withdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
