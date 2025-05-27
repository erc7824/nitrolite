# Nitrolite: State Channel Framework

**Nitrolite** is a lightweight state channel framework that enables off-chain interaction between participants, with an on-chain contract providing:

- **Custody** of tokens (ERC-20 and native) for each channel.
- **Mutual close** when participants agree on a final state.
- **Challenge/response** mechanism allowing a party to unilaterally finalize if needed.

State channel infrastructure has two main components:

- **IChannel** escrow which stores funds and can support and run adjudication on multiple channels
- **Adjudicators** are small contracts which validate state transitions to a candidate state against proofs

It is expected that Nitrolite implementation are 2 participants channels, where the first participants
is the channel creator and is called the client, the counterpart is clearnode a webservice which provide
chain abstraction services to connected clients. Allowing them to transact off-chain in a logical layer-3
with chain abstraction.

Nitrolite is the EVM implementation which connects to clearnode, other blockchains have other implementations
which must have equivalent escrow mechanisms.

## Interface Structure

### ChannelId

ChannelId hash is computed as:

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

### AppId

The `App` structure is the equivalent of the `Channel` type but for the off-chain realm. While channels operate on-chain with adjudicators validating state transitions, apps operate off-chain using a quorum-based consensus mechanism where participants have configurable weights toward meeting a required quorum threshold.

AppId hash is computed as:

```solidity
keccak256(
  abi.encode(
    app.protocol,
    app.participants,
    app.weights,
    app.quorum,
    app.challenge,
    app.nonce
  )
);
```

### StateHash

StateHash is used for signatures and stored in `state.sigs`:

```solidity
keccak256(
  abi.encode(
    channelId,
    state.data,
    state.version,
    state.allocations
  )
);
```

For signature verification, the stateHash is bare signed without EIP-191 since the protocol is intended to be chain-agnostic.

### `Types.sol`

Contains shared type definitions:

```solidity
struct Signature {
    uint8 v;
    bytes32 r;
    bytes32 s;
}

struct Amount {
    address token; // ERC-20 token address (address(0) for native tokens)
    uint256 amount; // Token amount
}

struct Allocation {
    address destination; // Where funds are sent on channel closure
    address token; // ERC-20 token contract address (address(0) for native tokens)
    uint256 amount; // Token amount allocated
}

struct Channel {
    address[] participants; // List of participants in the channel
    address adjudicator; // Address of the contract that validates state transitions
    uint64 challenge; // Duration in seconds for dispute resolution period
    uint64 nonce; // Unique per channel with same participants and adjudicator
}

struct App {
    string protocol; // String protocol/version "NitroRPC/0.2"
    address[] participants; // Array of participants in the app
    uint8[] weights; // Signers weights for this app [50, 50, 80, 20, 20]
    uint64 quorum; // Example value 100 would be the signature threshold
    uint64 challenge; // Duration in seconds for dispute resolution period
    uint64 nonce; // Unique per channel with same participants and adjudicator
}

struct State {
    StateIntent intent; // Intent of the state
    uint256 version; // State version incremental number to compare most recent
    bytes data; // Application data encoded, decoded by the adjudicator for business logic
    Allocation[] allocations; // Combined asset allocation and destination for each participant
    Signature[] sigs; // stateHash signatures from participants
}

enum Status {
    VOID,     // Channel was not created, State.version must be 0
    INITIAL,  // Channel is created and in funding process, State.version must be 0
    ACTIVE,   // Channel fully funded and operational, State.version is greater than 0
    DISPUTE,  // Challenge period is active
    FINAL     // Final state, channel can be closed
}

enum StateIntent {
    OPERATE,    // Operate is the app state 
    INITIALIZE, // Initial funding state
    RESIZE,     // Resize state
    FINALIZE    // Final closing state
}
```

### `IComparable.sol`

Interface for contracts that can determine ordering between states:

```solidity
interface IComparable {
    /**
     * @notice Compares two states to determine their relative ordering
     * @dev Implementations should return:
     *      -1 if candidate is less recent than previous
     *       0 if candidate is equally recent as previous
     *       1 if candidate is more recent than previous
     * @param candidate The state being evaluated
     * @param previous The reference state to compare against
     * @return result The comparison result:
     *         -1: candidate < previous (candidate is older)
     *          0: candidate == previous (same recency)
     *          1: candidate > previous (candidate is newer)
     */
    function compare(State calldata candidate, State calldata previous) external view returns (int8 result);
}
```

### `IAdjudicator.sol`

The adjudicator contract must implement:

```solidity
interface IAdjudicator {
    /**
     * @notice Validates a candidate state based on application-specific rules
     * @dev Used to determine if a state is valid during challenges or checkpoints
     * @param chan The channel configuration with participants, adjudicator, challenge period, and nonce
     * @param candidate The proposed state to be validated
     * @param proofs Array of previous states that provide context for validation
     * @return valid True if the candidate state is valid according to application rules
     */
    function adjudicate(Channel calldata chan, State calldata candidate, State[] calldata proofs)
        external
        view
        returns (bool valid);
}
```

### `IDeposit.sol`

Interface for contracts that allow users to deposit and withdraw token funds:

```solidity
interface IDeposit {
    /**
     * @notice Deposits tokens into the contract
     * @dev For native tokens, the value should be sent with the transaction
     * @param token Token address (use address(0) for native tokens)
     * @param amount Amount of tokens to deposit
     */
    function deposit(address token, uint256 amount) external payable;

    /**
     * @notice Withdraws tokens from the contract
     * @dev Can only withdraw available (not locked in channels) funds
     * @param token Token address (use address(0) for native tokens)
     * @param amount Amount of tokens to withdraw
     */
    function withdraw(address token, uint256 amount) external;
}
```

### `IChannel.sol` Interface

The main state channel interface implements:

```solidity
interface IChannel {
    event Created(bytes32 indexed channelId, Channel channel, State initial);
    event Joined(bytes32 indexed channelId, uint256 index);
    event Opened(bytes32 indexed channelId);
    event Challenged(bytes32 indexed channelId, uint256 expiration);
    event Checkpointed(bytes32 indexed channelId);
    event Resized(bytes32 indexed channelId, int256[] deltaAllocations);
    event Closed(bytes32 indexed channelId);

    /**
     * @notice Creates a new channel and initializes funding
     * @dev The creator must sign the funding state with StateIntent.INITIALIZE
     * @param ch Channel configuration with participants, adjudicator, challenge period, and nonce
     * @param initial Initial state with StateIntent.INITIALIZE and expected allocations
     * @return channelId Unique identifier for the created channel
     */
    function create(Channel calldata ch, State calldata initial) external returns (bytes32 channelId);

    /**
     * @notice Allows a participant to join a channel by signing the funding state
     * @dev Participant must provide signature on the same funding state with StateIntent.INITIALIZE
     * @param channelId Unique identifier for the channel
     * @param index Index of the participant in the channel's participants array
     * @param sig Signature of the participant on the funding state
     * @return channelId Unique identifier for the joined channel
     */
    function join(bytes32 channelId, uint256 index, Signature calldata sig) external returns (bytes32);

    /**
     * @notice Finalizes a channel with a mutually signed closing state
     * @dev Requires all participants' signatures on a state with StateIntent.FINALIZE,
     *      or can be called after challenge period expires with the last valid state
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state to be finalized
     * @param proofs Additional states required by the adjudicator to validate the candidate
     */
    function close(bytes32 channelId, State calldata candidate, State[] calldata proofs) external;

    /**
     * @notice All participants agree in setting a new allocation resulting in locking or unlocking funds
     * @dev Used for resizing channel allocations without withdrawing funds
     * @param channelId Unique identifier for the channel to resize
     * @param candidate The state that is to be true after resizing, containing the delta allocations
     * @param proofs An array of states supporting the claim that the candidate is true
     * NOTE: proof is needed to improve UX and allow resized state to follow any state (no need for consensus)
     */
    function resize(
        bytes32 channelId,
        State calldata candidate,
        State[] calldata proofs
    ) external;

    /**
     * @notice Initiates or updates a challenge with a signed state
     * @dev Starts a challenge period during which participants can respond with newer states
     * @param channelId Unique identifier for the channel
     * @param candidate The state being submitted as the latest valid state
     * @param proofs Additional states required by the adjudicator to validate the candidate
     */
    function challenge(bytes32 channelId, State calldata candidate, State[] calldata proofs) external;

    /**
     * @notice Records a valid state on-chain without initiating a challenge
     * @dev Used to establish on-chain proof of the latest state to prevent future disputes
     * @param channelId Unique identifier for the channel
     * @param candidate The state to checkpoint
     * @param proofs Additional states required by the adjudicator to validate the candidate
     */
    function checkpoint(bytes32 channelId, State calldata candidate, State[] calldata proofs) external;
}
```

## Funding Protocol

### Creation Phase

1. The Creator must:
   - Construct a channel configuration with participants, adjudicator, challenge period, and nonce
   - Prepare an initial state where `state.intent` is set to `StateIntent.INITIALIZE`
   - Define expected token deposits for all participants in the `state.allocations` array
   - Compute the Funding stateHash of this initial deposit state
   - Include creator's stateHash signature in the `state.sigs` array at position 0
   - Call the `create` function with the channel configuration and initial signed state

2. The system must:
   - Verify the Creator's signature on the funding stateHash
   - Verify creator has sufficient balance to fund required allocation
   - Lock the Creator's funds according to the allocation
   - Set the channel status to `INITIAL`
   - Emit a `Created` event with the channelId, channel configuration, and expected deposits

### Joining Phase

1. Each non-Creator participant must:
   - Verify the channelId and expected allocations
   - Sign the same funding stateHash (with `StateIntent.INITIALIZE`)
   - Call the `join` function with the channelId, their participant index, and signature

2. The system must:
   - Verify the participant's signature against the funding stateHash
   - Confirm the signer matches the expected participant at the given index
   - Lock the participant's funds according to the allocation
   - Track the actual deposit in the channel metadata
   - Emit a `Joined` event with the channelId and participant index

3. When all participants have joined, the system must:
   - Verify that all expected deposits are fulfilled
   - Set the channel status to `ACTIVE`
   - Emit an `Opened` event with the channelId

## State Versioning

Each state in a channel is uniquely identified by an incremental version number, stored in the `state.version` field. This ensures proper ordering of states during challenges and checkpoints:

1. When comparing two states during challenge/checkpoint operations, the system:
   - First attempts to use the `IComparable` interface if the channel's adjudicator implements it
   - If `IComparable` is not implemented, falls back to comparing `state.version` values directly

2. Version number rules:
   - For channel creation, `state.version` must be 0 (corresponds to `INITIAL` status)
   - For active channels, `state.version` must be greater than 0
   - Each new state should have a higher version number than the previous state
   - Higher version numbers indicate more recent states

This approach ensures that the most recent valid state always prevails during dispute resolution.

## Channel Closure

### Cooperative Close

1. To close cooperatively, any participant may:
   - Prepare a final state where `state.intent` is set to `StateIntent.FINALIZE`
   - Collect signatures from all participants on this final state
   - Call the `close` function with the channelId, final state, and any required proofs

2. The system must:
   - Verify all participant signatures on the closing stateHash
   - Verify the state contains `StateIntent.FINALIZE`
   - Distribute funds according to the final state's allocations
   - Set the channel status to `FINAL`
   - Delete the channel and emit a `Closed` event

### Challenge-Response Process

1. To initiate a challenge, a participant may:
   - Call the `challenge` function with their latest valid state and required proofs

2. The system must:
   - Verify the submitted state via the adjudicator
   - If valid, store the state and start the challenge period
   - Set a challenge expiration timestamp (current time + challenge duration)
   - Set the channel status to `DISPUTE`
   - Emit a `Challenged` event with the channelId and expiration time

3. During the challenge period, any participant may:
   - Submit a more recent valid state by calling `challenge` again
   - If the new state is valid and more recent (determined by comparing version numbers or using IComparable), the system must update the stored state and reset the challenge period

4. After the challenge period expires, any participant may call `close` to distribute funds according to the last valid challenged state

### Checkpointing

1. Any participant may:
   - Call the `checkpoint` function with a valid state and required proofs

2. The system must:
   - Verify the submitted state via the adjudicator
   - If valid and more recent (determined by comparing version numbers or using IComparable), store the state without starting a challenge period
   - Emit a `Checkpointed` event with the channelId

## Project Structure

```txt
src
├── Custody.sol
├── Utils.sol
├── adjudicators
│   ├── Consensus.sol
│   ├── Counter.sol
│   ├── MicroPayment.sol
└── interfaces
    ├── IAdjudicator.sol  # Interface for state validation
    ├── IChannel.sol      # Main interface for the state channel system
    ├── IComparable.sol   # Interface for determining state ordering
    ├── IDeposit.sol      # Interface for token deposit and withdrawal
    └── Types.sol         # Shared types used in the state channel system
```

### Custody Contract

The `Custody.sol` contract implements the `IChannel` and `IDeposit` interfaces, managing state channels and enforcing rules for creating, joining, closing, challenging, and checkpointing channels.
This implementation strictly supports only 2-participant channels with fixed roles: CREATOR (index 0) and BROKER (index 1).

```solidity
uint256 constant CREATOR = 0; // Participant index for the channel creator
uint256 constant BROKER = 1; // Participant index for the broker in clearnet context

struct Metadata {
    Channel chan;             // Channel configuration
    Status stage;             // Current channel status
    address creator;          // Creator address (caller of create function)
    Amount[2] expectedDeposits; // Fixed array for CREATOR (0) and BROKER (1) expected deposits
    Amount[2] actualDeposits;  // Fixed array for tracking actual deposits by CREATOR and BROKER
    uint256 challengeExpire;  // If non-zero channel will resolve to lastValidState when challenge Expires
    State lastValidState;     // Last valid state when adjudicator was called
    mapping(address token => uint256 balance) tokenBalances; // Token balances for the channel
}

struct Account {
    uint256 available;        // Available amount that can be withdrawn or allocated to channels
    uint256 locked;           // Amount currently allocated to channels
}

struct Ledger {
    mapping(address token => Account funds) tokens; // Token balances
    EnumerableSet.Bytes32Set channels; // Set of user ChannelId
}
```

### Remittance Adjudicator

The `Remittance.sol` contract implements the `IAdjudicator` and `IComparable` interfaces. Its purpose is to validate that allocation transfers between participants are legitimate based on cryptographically signed state hashes.

State data in the Remittance adjudicator contains a payment intent with the following structure:

```solidity
struct Intent {
    uint8 payer;      // Index of the paying participant (0 for CREATOR, 1 for BROKER)
    Amount transfer;  // Amount and token being transferred
}
```

When validating a state transition, the adjudicator requires:

- For the first state transition (version == 1): The funding state as proof (`proofs[0]`)
- For subsequent transitions (version > 1): Both the funding state (`proofs[0]`) and the previous valid state (`proofs[1]`)

A state transition is considered valid when:

1. The payer (participant whose allocation decreases) has signed the state hash
2. The allocation changes correctly reflect the intended transfer
3. The new allocations equal the previous allocations plus the intent transfer
4. The state version is incremented properly

This verification process ensures that all fund transfers between participants are authorized and accurately recorded.

## Roadmap

The following features are planned for future development:

1. **Enhanced multi-party channels support**
   - Further refinement of multi-party state validation
   - Improved handling of partially funded channels with multiple participants

2. **Nitrolite protocol as a unified virtual ledger (clearnet)**
   - Abstract from the underlying blockchain used
   - Support for cross-chain applications
