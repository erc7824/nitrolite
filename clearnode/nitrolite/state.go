package nitrolite

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

type Intent uint8

const (
	IntentOPERATE    Intent = 0
	IntentINITIALIZE Intent = 1
	IntentRESIZE     Intent = 2
	IntentFINALIZE   Intent = 3
)

// EncodeState encodes channel state into a byte array using channelID, intent, version, state data, and allocations.
func EncodeState(channelID common.Hash, intent Intent, version *big.Int, stateData []byte, allocations []Allocation) ([]byte, error) {
	allocationType, err := abi.NewType("tuple[]", "", []abi.ArgumentMarshaling{
		{Name: "destination", Type: "address"},
		{Name: "token", Type: "address"},
		{Name: "amount", Type: "uint256"},
	})
	if err != nil {
		return nil, err
	}

	intentType, err := abi.NewType("uint8", "", nil)
	if err != nil {
		return nil, err
	}
	versionType, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, err
	}

	args := abi.Arguments{
		{Type: abi.Type{T: abi.FixedBytesTy, Size: 32}}, // channelID
		{Type: intentType},               // intent
		{Type: versionType},              // version
		{Type: abi.Type{T: abi.BytesTy}}, // stateData
		{Type: allocationType},           // allocations (tuple[])
	}

	packed, err := args.Pack(channelID, uint8(intent), version, stateData, allocations)
	if err != nil {
		return nil, err
	}
	return packed, nil
}

// COMMON STATE DEFINITION

// Assumption: let's abstract from state channel framework, and try do define a common state without being limited by state-channels.

// NETWORK CONFIGURATION

// - A master smart contract, deployed on Ethereum, acts as a registry for main network configuration.
// This config defines registry of supported blockchains, networks on each blockchain, and smart contracts for each network.
// The smart contract also manages registry of approved ledger nodes, each with a signature weight.

// Per-Chain Configuration
// - Adjudicator on each chain will need to know mapping of YN asset id to token on it's network.
// So it keeps a registry of tokens it supports and maps them to YN tokens.
// Adjudicator should also mirror the registry of approved ledger nodes with their signature weights.

// YN ACCOUNT DEFINITION

// As we plan to support non-EVM chains, we will issue users with a unique Yellow Network account identifier (UserID).

/*
SQL Schema reference:
CREATE TABLE user_keys (
	user_id text PRIMARY KEY,
	key_type text NOT NULL,
	address text NOT NULL UNIQUE,
	permissions text NOT NULL // Includes target chains and apps and actions that can be performed with this key.
)
*/

type State struct {
	UnsignedState UnsignedState `json:"state"`
	OwnerSig      Signature     `json:"owner_sig"`
	NetworkSigs   []Signature   `json:"network_sigs"`
}

type UnsignedState struct {
	Nonce             uint64        `json:"nonce"`               // Common state nonce
	Data              []byte        `json:"data"`                // Common state data
	ChainStates       []ChainState  `json:"chain_states"`        // User allocation on each chain
	Balances          []TokenAmount `json:"balances"`            // User ledger balance on each chain
	ActiveSessionKeys []SessionKey  `json:"active_session_keys"` // List of active session keys.
}

// ActiveSessionKeys defines which keys can be used to sign new states or intents.

type ChainState struct {
	ChainID     uint32        `json:"chain_id"`
	Allocations []TokenAmount `json:"allocations"`
}

type TokenAmount struct {
	Asset  string          `json:"asset"` // Asset identifier on YN (todo: define strict formatting rules)
	Amount decimal.Decimal `json:"amount"`
}

// SessionKey holds the public key and permissions for a delegated key.
type SessionKey struct {
	Address     common.Address        `json:"address"` // The public address of the session key.
	Permissions SessionKeyPermissions `json:"permissions"`
}

// SessionKeyPermissions defines what a session key is allowed to do.
type SessionKeyPermissions struct {
	SpendingLimits []TokenAmount `json:"spending_limits,omitempty"`
	Expiry         uint64        `json:"expiry,omitempty"` // The timestamp when this key expires (seconds)
	Nonce          uint64        `json:"nonce,omitempty"`  // A nonce to prevent replay of the session key authorization.
}

// User deposits money on smart contract. Smart contract account is a big state channel with Yellow Network.
// Yellow network users are participants of this big state channel. State changes are validated by set of validators.
// Funds are in the same pool, delivering high liquidity and funds efficiency.

// CLEARNODE COMMON STATE INTEGRATION

// To perform a transfer on Yellow Network, a user Creates and signs an Intent:

type BatchTransferIntent struct {
	TransferIntent UnsignedBatchTransferIntent `json:"transfer_intent"`
	Signature      Signature                   `json:"signature"` // Recovered address from signature must match the recovered address from the referenced StateNonce.
}

type UnsignedBatchTransferIntent struct {
	StateNonce  uint64         `json:"state_nonce"` // Must match the sender's current State nonce.
	Destination common.Address `json:"destination"` // YN account identifier.
	Allocations []TokenAmount  `json:"allocations"` // The assets and amounts to be transferred.
}

// Validators sign the new State both for sender and receiver, store them and return them to the users.
// Users can then anytime submit the signed State to the Custody contract to settle. User also needs to provide his signature for the State.

// As when Network create new CommonStates, they have only signatures of validators, they can not submit the State to the Custody contract straight away.
// However, if network needs to source money the money user owes, it calls the Adjudicator contract.
// Adjudicator contract accepts:
// - last State A signed by user and validators.
// - array of transfer intents signed by user. (Proofs)
// - final State C signed by validators.

// Adjudicator verifies that networkSigs achieve a signature quorum threshold.
// Adjudicator contract verifies that provided signed transfer intents lead from state A to state C, so it can accept state C.

// SignedBatchWithdrawalIntent is the complete object submitted to the Custody contract and verified by Adjudicator.
type SignedBatchWithdrawalIntent struct {
	Intent      BatchWithdrawalIntent `json:"intent"`
	Signature   Signature             `json:"signature"`    // Recovered address from signature must match the recovered address from the referenced StateNonce.
	NetworkSigs []Signature           `json:"network_sigs"` // Quorum of valid network signatures on the Intent.
}

// BatchWithdrawalIntent is a user's declaration of their intent to withdraw funds.
type BatchWithdrawalIntent struct {
	StateNonce  uint64         `json:"state_nonce"` // Must match the sender's current State nonce.
	ChainID     uint32         `json:"chain_id"`    // Chain to withdraw.
	Destination common.Address `json:"destination"` // Destination for the withdrawn funds on the target chain, typically the owner's address.
	Withdrawals []TokenAmount  `json:"withdrawals"` // A list of tokens and amounts to withdraw. Can be full or a partial amount.
}

// User requests a withdrawal from our node by providing it with BatchWithdrawalIntent;
// Node keeps this Intent inside "mempool", where other nodes or "validators" can add their proofs;

// User has 2 options to withdraw funds:

// 1. Immediate cooperative withdraw with ledger node signatures.
// - Submit a Withdraw Intent to the Yellow Network, get ledger node signatures, and then submit the signed withdraw state to the Custody contract.

// 2. Withdraw with cooldown period with no ledger node signatures of Withdraw Intent.
// - Submit a Withdraw Intent to the Yellow Network without ledger node signatures (in case network is down),
// which will initialize a delayed withdraw. (like challenge or unlock period in yellow vault).
// Validator can use this time window to submit a newer valid state.

// Deposits

// The Network validators monitor Deposit events.

// Upon seeing a new Deposit event, the validators create a new UnsignedState for the user.
// This new state will have an incremented nonce and an updated ChainStates with the deposited funds.

// The validators sign this new State and credit the user's account within the network.
// The user doesn't need to sign a separate intent for deposits. // The Deposit event emitted by the contract is the authorization for the validators to update the user's state.
