# Nitrolite: State Channel Framework

**Nitrolite** refers to a type of powdered, high-explosive material with an ammonium nitrate base, used in mining, construction, and military applications.

This document describes a minimal **2-party state channel** that enables off-chain interaction between participants, with an on-chain contract providing:

- **Custody** of ERC-20 tokens for each channel.
- **Mutual close** when participants agree a final state.
- **Challenge/response** mechanism allowing a party to unilaterally finalize if needed.

> **Note:** The current implementation has been simplified to support only 2 participants per channel. Once the protocol is battle-tested, we plan to extend support for multiple participants as outlined in the Roadmap.

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
signature use ec25519 without eip-191 prefix as the protocol is chain-agnostic.

```solidity
keccak256(
  abi.encode(
    channelId,
    state.data,
    state.allocations
  )
);
```

### `Types.sol`

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
    Signature[] sigs; // stateHash signatures
}

enum Status {
    VOID,     // Channel was never active (zero-initialized)
    PARTIAL,  // Partial funding waiting for other participants
    ACTIVE,   // Channel fully funded and valid state
    FINAL,    // This is the FINAL state, channel can be closed
    INVALID   // Channel state is invalid
}

// This struct has been moved to Custody.sol with additional fields
// Kept here for backward compatibility, but should be migrated to use the Custody.sol version
struct Metadata {
    Channel chan; // Opener define channel configuration
    Status status; // Current channel status
    uint256 challengeExpire; // If non-zero channel will resolve to lastValidState when challenge Expires
    State lastValidState; // Last valid state when adjudicator was called
}
```

### `IAdjudicator.sol`

The adjudicator contract must implement:

```solidity
interface IAdjudicator {
    /**
     * @notice Validates the application state and determines the outcome of a channel
     * @dev This function evaluates the validity of a candidate state against provided proofs
     * @param chan The channel information containing participants, adjudicator, nonce, and challenge period
     * @param candidate The proposed state to be validated
     * @param proofs Array of previous states that may be used to validate the candidate state
     * @return valid is true if the candidate is approved
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        view
        returns (bool valid);
}
```

- **Parameters**:
  - `chan`: Channel configuration
  - `candidate`: The proposed state to be validated
  - `proofs`: Array of previous states that may be used to validate the candidate state
- **Returns**:
  - `valid`: Boolean indicating if the candidate state is approved

### `IDeposit.sol`

Interface for contracts that allow users to deposit and withdraw token funds. This interface is about pre-funding the contract to make calls to reset easier when we want to resize a channel allocation. Participants usually make their respective Deposit before calling open or reset, and the initial deposit balance must be higher or equal than the allocation to the channel.

```solidity
interface IDeposit {
    /**
     * @notice Deposits tokens into the contract
     * @dev Any user can deposit tokens
     * @param token Address of the ERC20 token to deposit
     * @param amount Amount of tokens to deposit
     */
    function deposit(address token, uint256 amount) external payable;

    /**
     * @notice Withdraws tokens from the contract
     * @dev Any user can withdraw their previously deposited tokens
     * @param token Address of the ERC20 token to withdraw
     * @param amount Amount of tokens to withdraw
     */
    function withdraw(address token, uint256 amount) external;
}
```

## `IChannel.sol` Interface

The main state channel interface implements:

```solidity
interface IChannel {
    event ChannelPartiallyFunded(bytes32 indexed channelId, Channel channel);
    event ChannelOpened(bytes32 indexed channelId, Channel channel);
    event ChannelChallenged(bytes32 indexed channelId, uint256 expiration);
    event ChannelCheckpointed(bytes32 indexed channelId);
    event ChannelClosed(bytes32 indexed channelId);

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
     * @notice Reset will close and open channel for resizing allocations
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     * @param ch Channel configuration
     * @param deposit is the initial State defined by the opener, it contains the expected allocation
     */
    function reset(
        bytes32 channelId,
        State calldata candidate,
        State[] calldata proofs,
        Channel calldata ch,
        State calldata deposit
    ) external;

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
   - **Process**:
     - When first participant calls this method, they provide the initial allocation and expected number of participants for the adjudicator
     - If their signature on the state is valid and their Allocation Transfer was successful (taking from initial `deposit()`), we check if state is valid with the adjudicator
     - The channel status is moved from VOID to PARTIAL and an event is emitted
     - Participants listen to the contract events, and if they approve the Allocation and State suggested by initiator, they can append their signature to the same State and call Open()
     - Once the last participant has successfully locked allocation to the channel, channel becomes Status.ACTIVE, and they can transact off-chain
   - **Effects**:  
     - Transfers token amounts from the caller to the contract
     - Call adjudicate to activate the channel
     - Returns unique channelId

2. **Close Channel (Cooperative Close)**  
   `close(bytes32 channelId, State candidate, State[] proofs)`  
   - **Purpose**: Finalize the channel immediately with a valid state.
   - **Logic**:
     - Calls `adjudicate` on the channel's adjudicator with the candidate state and proofs
     - Verifies all participant signatures are present
     - If state is valid, sets Status.FINAL and unallocates funds back to their user balance
     - Closes the channel

3. **Reset Channel**
`reset(bytes32 channelId, State candidate, State[] proofs, Channel ch, State deposit)`
   - **Purpose**: Close and reopen a channel to resize allocations.
   - **Logic**:
     - Closes the existing channel with the valid candidate state
     - Opens a new channel with the provided configuration and deposit
     - Used when allocation adjustments (deposits/withdrawals) are needed

4. **Challenge Channel**
`challenge(bytes32 channelId, State candidate, State[] proofs)`  
   - **Purpose**: Unilaterally post a state when the other party is uncooperative.
   - **Logic**:
     - Verifies the submitted state is valid via `adjudicate`
     - Either party can Challenge with a valid State
     - If valid, records the proposed state and starts the challenge period

5. **Checkpoint**
`checkpoint(bytes32 channelId, State candidate, State[] proofs)`  
   - **Purpose**: Store a valid state on-chain to prevent future disputes.
   - **Logic**:
     - Verifies the submitted state is valid via `adjudicate`
     - Either party can Checkpoint with a valid State
     - Records the state without initiating channel closure

6. **Reclaim**  
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
├── Utils.sol
├── adjudicators
│   ├── Consensus.sol
│   ├── Counter.sol
│   ├── MicroPayment.sol
└── interfaces
    ├── IAdjudicator.sol  # Interface for state validation and outcome determination
    ├── IChannel.sol      # Main interface for the state channel system
    ├── IDeposit.sol      # Interface for token deposit and withdrawal
    └── Types.sol         # Shared types used in the state channel system
```

### Custody.sol implementation

The `Custody.sol` contract implements the `IChannel` interface, managing the state channels and enforcing the rules for opening, closing, challenging, and reclaiming funds. It also contains the Status enum that defines the possible channel states.

```solidity
enum Status {
    VOID,     // Channel was never active (zero-initialized)
    PARTIAL,  // Partial funding waiting for other participants
    ACTIVE,   // Channel fully funded and valid state
    FINAL,    // This is the FINAL state, channel can be closed
    INVALID   // Channel state is invalid
}
```

#### Requirements

- Only state which adjudicator returns valid can replace previously lastValidState
- `open` is called first by the Host creating the initial funding State `deposit` which contains expected deposits
  - When Guest join the channel a call to the adjudicator will be made to validate state transitions from PARTIAL to ACTIVE
- `close` will be closing the channel if channel is ACTIVE, and adjudicator maybe return FINAL allowing token distribution
- `challenge` if the adjudicator returns valid, State is saved and challenge can be start by setting challengeExpire = now + ch.challenge
- `checkpoint` if the adjudicator returns valid, State is saved on-chain
- `reclaim` is called after challengeExpire time to distribute the tokens

```solidity
// This is the recommended internal structure for tracking channel state
struct Metadata {
    Channel chan;             // Opener define channel configuration
    Status status;            // Current channel status
    uint256 challengeExpire;  // If non-zero channel will resolve to lastValidState when challenge Expires
    State lastValidState;     // Last valid state when adjudicator was called
}

// ChannelId to Data
mapping(bytes32 => Metadata) private channels;
```

### Trivial Adjudicator

The Trivial adjudicator provides a basic implementation for validating state transitions. It always returns ACTIVE status, allowing testing the framework with simple state validation rules.

## Roadmap

The following features are planned for future development:

1. **Support for multiparty channels**
   - Refactor the `Channel.participants` structure to support variable-length arrays of participants
   - Update allocation handling to match the number of participants
   - Enhance the signature collection and verification process for multiple parties
   - Modify adjudicators to support multi-party state validation
   - Update state transition logic for partially funded channels with multiple participants
