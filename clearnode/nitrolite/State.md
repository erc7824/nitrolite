
-----

# Yellow Network Definition

Description generated from the State spec by Gemini

Documentation

## 1\. Overview

The Yellow Network is a high-performance overlay mesh network designed to operate across multiple blockchains. It provides users with a unified account, identified by a unique `UserID`, which aggregates their asset holdings across all integrated chains into a single `CommonState`. This architecture enables high-speed, low-cost off-chain transactions, deep liquidity, and a seamless cross-chain user experience.

All state transitions within the network are authorized by users and validated by a dynamic, weighted quorum of **Ledger Nodes**. A master smart contract on Ethereum maintains the network configuration, while on-chain **Custody** and **Adjudicator** contracts on each supported chain handle deposits, withdrawals, and dispute resolution.

## 2\. Core Concepts

### Actors

  * **User**: The owner of the funds and the `CommonState`, identified by a unique Yellow Network `UserID`. The user authorizes all actions by signing intents or states with their primary key or a delegated session key.
  * **Ledger Nodes (Clearnode)**: A permissioned set of off-chain nodes responsible for validating user intents, issuing new states, reaching consensus, and attesting to state validity with their weighted signatures. Their registry and weights are managed by the Adjudicator contract.
  * **Master Registry Contract (on Ethereum)**: The primary on-chain contract defining the network's configuration, including approved ledger node keys, supported blockchains and the addresses of their respective Custody and Adjudicator contracts.
  * **Custody Contract (on each chain)**: Holds the pooled funds for all users on a specific blockchain, processes deposits, and executes withdrawals upon receiving valid proofs.
  * **Adjudicator Contract (on each chain)**: A companion to the Custody contract that validates state proofs. It verifies signatures, and resolves disputes by computationally verifying forced state transitions.

### Core Data Structures

#### `CommonState`

This is the master object representing a user's account, bundling the state data with the necessary signatures to prove its validity.

```go
type CommonState struct {
    State       UnsignedCommonState `json:"state"`
    OwnerSig    Signature           `json:"owner_sig"`
    NetworkSigs []Signature         `json:"network_sigs"`
}
```

#### `UnsignedCommonState`

This is the core data that is passed around and signed.

```go
type UnsignedCommonState struct {
    Nonce             uint64       `json:"nonce"`               // Strictly increasing nonce to prevent replay attacks.
    StateData         []byte       `json:"state_data"`          // Flexible field for arbitrary application-level metadata.
    Balances          []TokenAmount `json:"balances"`            // User ledger balance on each chain
    // ChainStates       []ChainState `json:"chain_states"`        // Details the user's token balances on each chain.
    ActiveSessionKeys []SessionKey `json:"active_session_keys"` // List of delegated, permissioned keys.
}
```

#### Supporting Structures

```go
// // ChainState details a user's funds on a single blockchain.
// type ChainState struct {
//     ChainID     uint32        `json:"chain_id"`
//     Allocations []TokenAmount `json:"allocations"`
// }

// TokenAmount represents a quantity of a specific asset.
type TokenAmount struct {
    Asset  string          `json:"asset"` // Yellow Network asset identifier.
    Amount decimal.Decimal `json:"amount"`
}

// SessionKey defines a delegated key and its permissions.
type SessionKey struct {
    KeyAddress  common.Address        `json:"key_address"`
    Permissions SessionKeyPermissions `json:"permissions"`
}

// SessionKeyPermissions specifies the limits and rules for a session key.
type SessionKeyPermissions struct {
    SpendingLimits []TokenAmount `json:"spending_limits,omitempty"`
    Expiry         uint64        `json:"expiry,omitempty"`
    Nonce          uint64        `json:"nonce,omitempty"`
}
```

-----

## 3\. Interaction Flows

### 🏦 Flow 1: Onboarding (Deposit)

A deposit is an on-chain action that credits a user's `CommonState` within the Yellow Network.

1.  **On-Chain Deposit**: The user interacts directly with the **Custody Contract** on a specific blockchain. They call the standard ERC20 `approve()` function, followed by the contract's `deposit(token, amount)` function.
2.  **Event Emission**: The Custody Contract securely receives the funds into the shared liquidity pool and emits a `Deposit` event containing the user's address, token, amount, and chain ID.
3.  **Off-Chain Acknowledgment**: Yellow Network Ledger Nodes, who constantly monitor the Custody Contracts, observe this `Deposit` event.
4.  **State Creation**: The Ledger Nodes create a new `CommonState` for the user. This new state has an incremented `nonce` and an updated balance in the corresponding `ChainState` within the `ChainStates` array.
5.  **Confirmation**: The Ledger Nodes sign this new `CommonState` and deliver it to the user. The on-chain `Deposit` event serves as authorization; no separate user signature is needed for this initial state creation.

### 💸 Flow 2: Core Transaction (Batch Transfer)

A transfer is an off-chain operation between two users within the network, allowing multiple assets to be sent simultaneously.

1.  **Intent Creation**: The sender creates and signs an `UnsignedBatchTransferIntent` with their primary key or a valid session key. This intent specifies their current `StateNonce`, the destination `UserID`, and an array of `Allocations` (assets and amounts) to be transferred.
2.  **Submission**: The sender submits the signed `BatchTransferIntent` to a Ledger Node.
3.  **Validation & State Transition**: The Ledger Node network receives the intent and verifies:
      * The signature is valid (from the owner or a permitted session key).
      * The `StateNonce` matches the sender's current state nonce.
      * The sender has sufficient funds for all specified allocations.
4.  **Atomic Update**: If valid, the Ledger Nodes atomically create **two** new `CommonState` objects:
      * **For the Sender**: A state with decremented balances and `nonce + 1`.
      * **For the Receiver**: A state with incremented balances and `nonce + 1`.
5.  **Confirmation**: The Ledger Nodes sign both new `CommonState` objects and deliver them to the sender and receiver, respectively. The transfer is now complete.

### 📤 Flow 3: Offboarding (Batch Withdrawal)

The system provides two methods for withdrawing funds back to a main blockchain.

#### A. Cooperative (Immediate) Withdrawal

This is the standard, fast path for users when the network is operating normally.

1.  **Intent Creation**: The user creates a `BatchWithdrawalIntent`, specifying the `StateNonce`, target `ChainID`, destination address, and a list of assets and amounts to withdraw. They sign it to create a user-signed intent.
2.  **Request for Co-Signature**: The user submits this signed intent to the Ledger Node network.
3.  **Validator Attestation**: The Ledger Nodes verify the request, lock the user's account to prevent further off-chain activity, and co-sign the intent, creating a `SignedBatchWithdrawalIntent`. This object includes the original intent, the user's signature, and a quorum of network signatures, serving as a **Proof-of-Finality**.
4.  **On-Chain Submission**: The user submits the `SignedBatchWithdrawalIntent` to the **Custody Contract** on the target chain.
5.  **Immediate Payout**: The contract's `Adjudicator` module verifies the network quorum signature. Because this proof of finality is present, the contract waives the long challenge period, and funds are released after a minimal finalization delay.

#### B. Uncooperative (Delayed) Withdrawal

This is the user's safety valve if the network is unresponsive or refuses to co-sign.

1.  **Intent Creation**: The user creates and signs a `BatchWithdrawalIntent` but is unable to get the required network signatures.
2.  **On-Chain Submission**: The user submits this intent (signed only by them) to the **Custody Contract**.
3.  **Challenge Period Begins**: This action initiates a long, on-chain **challenge period** (e.g., several hours or days).
4.  **Resolution**: One of two things will happen:
      * **If the period expires**: The withdrawal is considered valid, and the user can claim their funds from the Custody Contract.
      * **If challenged**: A Ledger Node can submit a `CommonState` with a higher nonce during the window. The `Adjudicator` verifies this newer state. If valid, it proves the user's withdrawal request was based on a stale state, and the request is canceled.

### 🔑 Flow 4: Session Keys

Session keys allow users to delegate limited, revocable permissions without exposing their primary key.

1.  **Management**: To add or revoke a session key, the owner signs a specific intent (e.g., `UpdateSessionKeysIntent`). This is processed by the Ledger Nodes, who issue a new `CommonState` with the updated `ActiveSessionKeys` list.
2.  **Usage**: A session key can sign a `BatchTransferIntent`. When Ledger Nodes receive it, they check the signature and verify that the transfer complies with the `Permissions` defined for that key in the user's `CommonState` (e.g., `SpendingLimits`, `Expiry`).
3.  **Security**: Session keys **cannot** sign critical intents like `BatchWithdrawalIntent` or intents that manage other session keys.

### 🛡️ Flow 5: Forced Settlement (Debt Resolution)

This protocol is used by the network to finalize a user's state on-chain, typically to settle a debt (e.g., fees) from a user who has gone offline.

1.  **Scenario**: A user authorizes one or more transfers but then disappears before acknowledging receipt of their new `CommonState`. The network needs to finalize this state on-chain.
2.  **Proof Submission**: The network submits a proof to the **Adjudicator Contract** containing:
      * The last `CommonState` `A` that was fully signed by the user and the network.
      * An array of subsequent, user-signed `BatchTransferIntents`.
      * The final `CommonState` `C` (signed only by Ledger Nodes) that is the result of applying the intents to state `A`.
3.  **On-Chain Verification**: The `Adjudicator` contract computationally verifies that the provided sequence of signed intents correctly transitions state `A` to state `C`. This is a potentially gas-intensive operation that provides a cryptographic guarantee of the state transition's validity.
4.  **Challenge & Finalization**: If the proof is valid, a challenge period begins for state `C`, allowing the user to dispute it with a newer, user-signed state if one exists. If undisputed, state `C` is finalized on-chain, settling the user's balance and allowing the network to resolve any outstanding obligations.