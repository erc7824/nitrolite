
---

# Yellow Network: State & Interaction Flows 

Description generated from the State spec by Gemini

Documentation

## 1. Overview

The Yellow Network is a high-performance overlay mesh designed to operate on top of multiple blockchains. It provides users with a unified account, represented by a single `CommonState`, which aggregates their asset holdings across all integrated chains. This architecture enables high-speed, low-cost off-chain transactions, deep liquidity, and a seamless cross-chain user experience.

All state transitions within the network are authorized by users and validated by a dynamic, weighted quorum of **Validators**. The on-chain **Adjudicator** contract serves as the ultimate arbiter for deposits, withdrawals, and dispute resolution.

## 2. Core Concepts

### Actors

* **User**: The owner of the funds and the `CommonState`. The user authorizes all actions by signing intents or states with their primary key or a delegated session key.
* **Validators**: A permissioned set of off-chain nodes responsible for validating user transactions, creating new states, and attesting to their validity with signatures. The validator set and their signature weights are managed by the `Adjudicator` contract.
* **Adjudicator Contract**: The on-chain smart contract that acts as the ultimate source of truth. It holds all user funds, manages the validator registry, processes deposits, and finalizes withdrawals.

### Core Data Structures

#### `CommonState`
This is the master object representing a user's account. It bundles the state data with the necessary signatures to prove its validity.

* `State`: The `UnsignedCommonState` containing the actual account data.
* `OwnerSig`: The user's signature, required to authorize on-chain actions like withdrawals.
* `ValidatorSigs`: A quorum of validator signatures attesting to the validity of the state.

#### `UnsignedCommonState`
This is the core data that is passed around and signed.

* `Nonce`: A strictly increasing `uint64` that sequences states and prevents all forms of replay attacks.
* `ChainStates`: An array detailing the user's token balances on each integrated blockchain.
* `SessionKeys`: A list of delegated, permissioned keys that can act on the user's behalf.
* `StateData`: A flexible field for arbitrary application-level metadata.

---

## 3. Interaction Flows

### 🏦 Flow 1: Onboarding (Deposit)

A deposit is an on-chain action that credits a user's account within the Yellow Network.

1.  **On-Chain Deposit**: The user interacts directly with the `Adjudicator` contract on a specific blockchain. They call the standard ERC20 `approve()` function, followed by the `Adjudicator.deposit(token, amount)` function.
2.  **Event Emission**: The `Adjudicator` contract securely receives the funds and emits a `Deposit` event containing the user's address, token, amount, and chain ID.
3.  **Off-Chain Acknowledgment**: Yellow Network validators, who are constantly monitoring the `Adjudicator`, see this `Deposit` event.
4.  **State Creation**: The validators create a new `CommonState` for the user. This new state has an incremented `nonce` and an updated balance in the `ChainStates` array reflecting the deposit.
5.  **Confirmation**: The validators sign this new `CommonState` and deliver it to the user. The user's account is now credited, and they can begin transacting off-chain. The user's on-chain transaction serves as the authorization; no separate signature is needed.

### 💸 Flow 2: Core Transaction (Transfer)

A transfer is an off-chain operation between two users within the network.

1.  **Intent Creation**: The sender creates and signs an `UnsignedTransferIntent` with their primary key or a valid session key. This intent specifies their current `StateNonce`, the destination address, and the assets to be transferred.
2.  **Submission**: The sender submits the signed `TransferIntent` to a Clearnode in the network.
3.  **Validation & State Transition**: The validator network receives the intent. They verify:
    * The signature is valid (from the owner or a permitted session key).
    * The `StateNonce` matches the sender's current state.
    * The sender has sufficient funds.
4.  **Atomic Update**: If valid, the validators atomically create **two** new `CommonState` objects:
    * **For the Sender**: A state with a decremented balance and `nonce + 1`.
    * **For the Receiver**: A state with an incremented balance and `nonce + 1`.
5.  **Confirmation**: The validators sign both new `CommonState` objects and deliver them to the sender and receiver, respectively. The transfer is now complete.

### 📤 Flow 3: Offboarding (Withdrawal)

The system provides two methods for withdrawing funds back to the main blockchain.

#### A. Cooperative (Immediate) Withdrawal

This is the standard, fast path for users when the network is operating normally.

1.  **Intent Creation**: The user creates and signs a `WithdrawIntent`.
2.  **Request for Co-Signature**: The user submits their signed intent to the network.
3.  **Validator Attestation**: The validators verify the request. They **lock** the user's account to prevent further off-chain activity and co-sign the `WithdrawIntent`, creating a `SignedWithdrawIntent`. This object serves as a **Proof-of-Finality**.
4.  **On-Chain Submission**: The user submits the `SignedWithdrawIntent` to the `Adjudicator` contract.
5.  **Immediate Payout**: Because the contract receives proof of finality from the validators, it waives the long challenge period. The funds are released after a minimal finalization delay (e.g., 1-2 blocks) for chain stability.

#### B. Uncooperative (Delayed) Withdrawal

This is the user's safety valve if the network is unresponsive.

1.  **Intent Creation**: The user creates and signs a `WithdrawIntent`, but is unable to get validator signatures.
2.  **On-Chain Submission**: The user submits the `WithdrawIntent` (signed only by them) to the `Adjudicator` contract.
3.  **Challenge Period Begins**: This action initiates a long, on-chain **challenge period** (e.g., hours or days).
4.  **Resolution**: One of two things will happen:
    * **If the period expires**: The withdrawal is considered valid, and the user can claim their funds.
    * **If challenged**: A validator can submit a `CommonState` with a higher nonce during the window. This proves the user's withdrawal request was based on a stale state, and the request is canceled.

### 🔑 Flow 4: Session Keys

Session keys allow users to delegate limited permissions without exposing their main key.

1.  **Management**: To add or revoke a session key, the owner signs a specific intent (e.g., `UpdateSessionKeyIntent`). This is processed by the validators, who issue a new `CommonState` with the updated `SessionKeys` list.
2.  **Usage**: A session key can sign a `TransferIntent`. When validators receive it, they check both the signature and that the transfer complies with the `Permissions` (e.g., spending limits, expiry) defined in the `CommonState`.
3.  **Security**: Session keys **cannot** sign critical intents like `WithdrawIntent` or intents that manage other session keys.

### 🛡️ Flow 5: Forced Settlement (Debt Resolution)

This protocol is used by the network to settle a debt from a user who has gone offline.

1.  **Scenario**: A user authorizes a transfer that incurs a fee but then disappears before receiving their new `CommonState`.
2.  **Proof Submission**: The network submits a proof to the `Adjudicator` contract containing:
    * The last `CommonState` `A` that was fully signed by the user.
    * An array of subsequent, user-signed `TransferIntents`.
    * The final `CommonState` `C` (signed only by validators) that is the result of applying the intents to state `A`.
3.  **On-Chain Verification**: The `Adjudicator` contract computationally verifies that the intents correctly transition state `A` to state `C`. This is a gas-intensive operation that serves as a strong guarantee.
4.  **Challenge & Finalization**: If the proof is valid, a challenge period begins for state `C`, allowing the user to dispute it. If undisputed, the state is finalized, and the network can claim the debt.