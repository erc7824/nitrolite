package nitrolite

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
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

type CommonState struct {
	State       UnsignedCommonState `json:"state"`
	OwnerSig    Signature           `json:"owner_sig"`
	NetworkSigs []Signature         `json:"network_sigs"`
}

// Adjudicator manages a registry of approved ledger nodes. Each node has a signature weight.
// Adjudicator verifies that networkSigs achieve a signature quorum threshold.
// Number of approved ledger nodes and quorum threshold can be adjusted in realtime by adjudicator contract.

type UnsignedCommonState struct {
	Nonce             uint64       `json:"nonce"`               // Common state nonce
	StateData         []byte       `json:"state_data"`          // Common state data
	ChainStates       []ChainState `json:"chain_states"`        // User allocation on each chain
	ActiveSessionKeys []SessionKey `json:"active_session_keys"` // List of active session keys.
}

// ActiveSessionKeys defines which keys can be used to sign new states or intents.

type ChainState struct {
	ChainID     uint32        `json:"chain_id"`
	Allocations []TokenAmount `json:"allocations"`
}

type TokenAmount struct {
	TokenAddress common.Address `json:"token"`
	RawAmount    *big.Int       `json:"amount"`
}

// SessionKey holds the public key and permissions for a delegated key.
type SessionKey struct {
	KeyAddress  common.Address        `json:"key_address"` // The public address of the session key.
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

// To perform a transfer on Yellow Network, User Creates and signs an Intent:

type TransferIntent struct {
	TransferIntent UnsignedTransferIntent `json:"transfer_intent"`
	Signature      Signature              `json:"signature"`
}

type UnsignedTransferIntent struct {
	StateNonce  uint64         `json:"state_nonce"` // Must match the sender's current CommonState nonce.
	Destination common.Address `json:"destination"`
	Allocations []Allocation   `json:"allocations"` // The assets and amounts to be transferred.
}

// Validators sign the new CommonState both for sender and receiver, store them and return them to the users.
// Users can then anytime submit the signed CommonState to the Custody contract to settle. User also needs to provide his signature for the CommonState.

// As when Network create new CommonStates, they have only signatures of validators, they can not submit the CommonState to the Custody contract straight away.
// However, if network needs to source money the money user owes, it calls the Adjudicator contract.
// Adjudicator contract accepts:
// - last CommonState A signed by user and validators.
// - array of transfer intents signed by user. (Proofs)
// - final CommonState C signed by validators.

// Adjudicator contract verifies that provided signed transfer intents lead from state A to state C, so it can accept state C.

// SignedBatchWithdrawalIntent is the complete object submitted to the Custody contract and verified by Adjudicator.
type SignedBatchWithdrawalIntent struct {
	Intent      BatchWithdrawalIntent `json:"intent"`
	Signature   Signature             `json:"signature"`    // User's signature on the Intent.
	NetworkSigs []Signature           `json:"network_sigs"` // Quorum of valid network signatures on the Intent.
}

// BatchWithdrawalIntent is a user's declaration of their intent to withdraw funds.
type BatchWithdrawalIntent struct {
	StateNonce  uint64         `json:"state_nonce"` // Must match the sender's current CommonState nonce.
	ChainID     uint32         `json:"chain_id"`    // Chain to withdraw.
	Destination common.Address `json:"destination"` // Destination for the withdrawn funds, typically the owner's address.
	Withdrawals []TokenAmount  `json:"withdrawals"` // A list of tokens and amounts to withdraw. Can be full or a partial amount.
}

// User has 2 options to withdraw funds:

// 1. Immediate cooperative withdraw with ledger node signatures.
// - Submit a Withdraw Intent to the Yellow Network, get ledger node signatures, and then submit the signed withdraw state to the Custody contract.

// 2. Withdraw with cooldown period with no ledger node signatures of Withdraw Intent.
// - Submit a Withdraw Intent to the Yellow Network without ledger node signatures (in case network is down),
// which will initialize a delayed withdraw. (like challenge or unlock period in yellow vault).
// Validator can use this time window to submit a newer valid state.

// Deposits

// The Network validators monitor Deposit events.

// Upon seeing a new Deposit event, the validators create a new UnsignedCommonState for the user.
// This new state will have an incremented nonce and an updated ChainStates with the deposited funds.

// The validators sign this new CommonState and credit the user's account within the network.
// The user doesn't need to sign a separate intent for deposits. // The Deposit event emitted by the contract is the authorization for the validators to update the user's state.
