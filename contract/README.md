# Nitrolite: State Channel Framework

**Nitrolite** refers to a type of powdered, high-explosive material with an ammonium nitrate base, used in mining, construction, and military applications.

This document describes a minimal **2-party state channel** that enables off-chain interaction between participants, with an on-chain contract providing:

- **Custody** of ERC-20 tokens for each channel.
- **Mutual close** when participants agree a final state.
- **Challenge/response** mechanism allowing a party to unilaterally finalize if needed.

State channel infrastructure has two main components:

- **IChannel** escrow which stores funds and can support and run adjudication on multiple channels
- **Adjudicator** are small contracts which can validate state transitions to a candidate state against proofs

## Interface Structure

### ChannelId

ChannelId hash are computed the following way:

```solidity
keccak256(
  abi.encode(
    ch.participants,
    ch.adjudicator,
    ch.challenge,
    ch.nonce
  )
);
```

### StateHash

StateHash are used in signature and often stored in `state.sigs`

```solidity
keccak256(
  abi.encode(
    channelId,
    state.data,
    state.allocations
  )
);
```

### `Types`

Contains shared type definitions:

```solidity
struct Signature {
    uint8 v;
    bytes32 r;
    bytes32 s;
}

struct Allocation {
    address destination; // Where funds are sent on channel closure
    address token; // ERC-20 token contract address
    uint256 amount; // Token amount allocated
}

struct Channel {
    address[2] participants; // List of participants in the channel [Host, Guest]
    address adjudicator; // Address of the contract that validates final states
    uint64 challenge; // Duration in second, Participants can dispute by submitting newer valid state during challenge
    uint64 nonce; // Unique per channel with same participants and adjudicator
}

struct State {
    bytes data; // Application data encoded, decoded by the adjudicator for business logic
    Allocation[2] allocations; // Combined asset allocation and destination for each participant
    Signature[2] sigs; // stateHash signatures
}

// Recommended structure to keep track of states
struct Metadata {
    Channel chan; // Opener define channel configuration
    uint256 challengeExpire; // If non-zero channel will resolve to lastValidState when challenge Expires
    State lastValidState; // Last valid state when adjudicator was called
}
```

### `IAdjudicator`

The adjudicator contract must implement:

```solidity
interface IAdjudicator {
    enum Status {
        VOID,     // Channel was never active or have an anomaly
        PARTIAL,  // Partial funding waiting for other participants
        ACTIVE,   // Channel fully funded using open or state are valid
        INVALID,  // Channel state is invalid
        FINAL     // This is the FINAL State channel can be closed
    }

    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        view
        returns (Status decision);
}
```

- **Parameters**:
  - `chan`: Channel configuration
  - `candidate`: The State being validated, containing application-specific data
  - `proofs`: Previous valid states for reference in validation
- **Returns**:
  - `decision`: Status of the channel after adjudication

## IChannel Interface

The main state channel interface implements:

```solidity
interface IChannel {
    /**
     * @notice Open or join a channel by depositing assets
     * @param ch Channel configuration
     * @param deposit is the initial State defined by the opener, it contains the expected allocation
     * @return channelId Unique identifier for the channel
     */
    function open(Channel calldata ch, State calldata deposit) external returns (bytes32 channelId);

    /**
     * @notice Finalize the channel with a mutually signed state
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     */
    function close(bytes32 channelId, State calldata candidate, State[] calldata proofs) external;

    /**
     * @notice Unilaterally post a state when the other party is uncooperative
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     */
    function challenge(bytes32 channelId, State calldata candidate, State[] calldata proofs) external;

    /**
     * @notice Unilaterally post a state to store it on-chain to prevent future disputes
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     */
    function checkpoint(bytes32 channelId, State calldata candidate, State[] calldata proofs) external;

    /**
     * @notice Conclude the channel after challenge period expires
     * @param channelId Unique identifier for the channel
     */
    function reclaim(bytes32 channelId) external;
}
```

### Protocol Details

1. **Open Channel**  
   `open(Channel ch, State deposit) returns (bytes32 channelId)`
   - **Purpose**: Open or join a channel by depositing participants assets into the contract.
   First depositor is the Host, second depositor is the Guest.
   - **Notice**: Participants are only used to sign state and might not be the caller of the smart-contract,
   Moreover participant address are not payout destination addresses.
   - **Effects**:  
     - Transfers token amounts from the caller to the contract
     - Call adjudicate to activate the channel
     - Returns unique channelId

2. **Close Channel (Cooperative Close)**  
   `close(bytes32 channelId, State candidate, State[] proofs)`  
   - **Purpose**: Finalize the channel immediately with a valid state.
   - **Logic**:
     - Calls `adjudicate` on the channel's adjudicator with the candidate state and proofs
     - If valid, distributes tokens according to the state's allocations
     - Closes the channel

3. **Challenge Channel**  
   `challenge(bytes32 channelId, State candidate, State[] proofs)`  
   - **Purpose**: Unilaterally post a state when the other party is uncooperative.
   - **Logic**:
     - Verifies the submitted state is valid via `adjudicate`
     - If valid, records the proposed state and starts the challenge period

4. **Checkpoint**  
   `checkpoint(bytes32 channelId, State candidate, State[] proofs)`  
   - **Purpose**: Store a valid state on-chain to prevent future disputes.
   - **Logic**:
     - Verifies the submitted state is valid via `adjudicate`
     - Records the state without initiating channel closure

5. **Reclaim**  
   `reclaim(bytes32 channelId)`  
   - **Purpose**: Conclude the channel after challenge period expires.
   - **Logic**:  
     - Distributes tokens according to the last valid state's allocations
     - Closes the channel

## High-Level Flow

1. **Channel Creation**:  
   - Two participants deposit ERC20 tokens into the contract using `open` with an initial state.
2. **Off-Chain Updates**:  
   - The parties exchange and co-sign states off-chain, with application-specific data encoded in the `data` field.
3. **Happy Path (Cooperative Close)**:  
   - A final state is validated by the adjudicator.
   - Either party calls `close` with the candidate state and any required proofs.
   - The adjudicator verifies the state's validity, and the contract uses the state's allocations for distribution.
4. **Intermediate State Record (Checkpoint)**:
   - At any point, either party can call `checkpoint` to record a valid state on-chain.
   - This doesn't close the channel but provides protection against future disputes.
5. **Unhappy Path (Challenge)**:  
   - One party calls `challenge` with their most recent valid state and any required proofs.
   - The counterparty may respond with a more recent valid state using another `challenge`.
   - After the challenge period expires, `reclaim` settles funds according to the allocations in the last adjudicated valid state.

## Project Structure

```
src
├── Custody.sol
├── adjudicators
│   └── Trivial.sol
└── interfaces
    ├── IAdjudicator.sol  # Interface for state validation and outcome determination
    ├── IChannel.sol      # Main interface for the state channel system
    └── Types.sol         # Shared types used in the state channel system
```

### Custody.sol implementation

The `Custody.sol` contract implements the `IChannel` interface, managing the state channels and enforcing the rules for opening, closing, challenging, and reclaiming funds.

#### Requirements

- Only state which adjudicator return valid can replace previously lastValidState
- `open` is called first by the Host creating the initial funding State `deposit` which contains expected deposits
  - When Guest join the channel a call to the adjudicator will be made to validate state transitions from PARTIAL to ACTIVE
- `close` will be closing the channel if channel is ACTIVE, and adjudicator maybe return FINAL allowing token distribution
- `challenge` if the adjudicator return ACTIVE, State is saved and challenge can be start by setting challengeExpire = now + ch.challenge
- `checkpoint` if the adjudicator return ACTIVE, State is saved on-chain
- `reclaim` is called after challengeExpire time to distribute the tokens

```solidity
// This is the recommended internal structure for tracking channel state
struct Metadata {
    Channel chan;             // Opener define channel configuration
    uint256 challengeExpire;  // If non-zero channel will resolve to lastValidState when challenge Expires
    State lastValidState;     // Last valid state when adjudicator was called
}

// ChannelId to Data
mapping(bytes32 => Metadata) private channels;
```

### Trivial Adjudicator

The Trivial adjudicator provides a basic implementation for validating state transitions. It always returns ACTIVE status, allowing testing the framework with simple state validation rules.
