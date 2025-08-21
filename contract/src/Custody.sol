// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {EIP712} from "lib/openzeppelin-contracts/contracts/utils/cryptography/EIP712.sol";
import {IERC20} from "lib/openzeppelin-contracts/contracts/interfaces/IERC20.sol";
import {SafeERC20} from "lib/openzeppelin-contracts/contracts/token/ERC20/utils/SafeERC20.sol";
import {EnumerableSet} from "lib/openzeppelin-contracts/contracts/utils/structs/EnumerableSet.sol";

import {IAdjudicator} from "./interfaces/IAdjudicator.sol";
import {IChannelReader} from "./interfaces/IChannelReader.sol";
import {IComparable} from "./interfaces/IComparable.sol";
import {IChannel} from "./interfaces/IChannel.sol";
import {IDeposit} from "./interfaces/IDeposit.sol";
import {Channel, State, Allocation, ChannelStatus, StateIntent, Amount} from "./interfaces/Types.sol";
import {Utils} from "./Utils.sol";

/**
 * @title Custody
 * @notice A simple custody contract for state channels that delegates most state transition logic to an adjudicator
 * @dev This implementation currently only supports 2 participant channels (CLIENT and SERVER)
 */
contract Custody is IChannel, IDeposit, IChannelReader, EIP712 {
    using EnumerableSet for EnumerableSet.Bytes32Set;
    using SafeERC20 for IERC20;
    using Utils for State;

    // Errors
    // TODO: sort errors
    error ChannelNotFound(bytes32 channelId);
    error ChannelNotFinal();
    error InvalidParticipant();
    error InvalidStatus();
    error InvalidState();
    error InvalidAllocations();
    error DepositAlreadyFulfilled();
    error DepositsNotFulfilled(uint256 expectedFulfilled, uint256 actualFulfilled);
    error InvalidStateSignatures();
    error InvalidAdjudicator();
    error InvalidChallengerSignature();
    error InvalidChallengePeriod();
    error InvalidValue();
    error InvalidAmount();
    error TransferFailed(address token, address to, uint256 amount);
    error ChallengeNotExpired();
    error InsufficientBalance(uint256 available, uint256 required);

    // Custody contract restricts number of participants to 2
    uint256 constant PART_NUM = 2;
    uint256 constant CLIENT_IDX = 0; // Participant index for the channel creator
    uint256 constant SERVER_IDX = 1; // Participant index for the server in clearnet context

    uint256 public constant MIN_CHALLENGE_PERIOD = 1 hours;

    bytes32 public constant CHALLENGE_STATE_TYPEHASH = keccak256(
        "AllowChallengeStateHash(bytes32 channelId,uint8 intent,uint256 version,bytes data,Allocation[] allocations)Allocation(address destination,address token,uint256 amount)"
    );

    // Recommended structure to keep track of states
    struct Metadata {
        Channel chan; // Opener define channel configuration
        ChannelStatus stage;
        address[2] wallets; // depositing and resizing wallets for CLIENT and SERVER
        // Fixed arrays for exactly 2 participants (CLIENT and SERVER)
        // TODO: store `uint256` instead of `Amount`, as tokens are the same
        Amount[2] expectedDeposits; // CLIENT defines Token per participant
        Amount[2] actualDeposits; // Tracks deposits made by each participant
        uint256 challengeExpire; // If non-zero channel will resolve to lastValidState when challenge Expires
        State lastValidState; // Last valid state when adjudicator was called
        mapping(address token => uint256 balance) tokenBalances; // Token balances for the channel
    }

    struct Ledger {
        mapping(address token => uint256 available) tokens; // Available amount that can be withdrawn or allocated to channels
        EnumerableSet.Bytes32Set channels; // Set of user ChannelId
    }

    mapping(bytes32 channelId => Metadata chMeta) internal _channels;
    mapping(address account => Ledger ledger) internal _ledgers;

    // ========== Constructor ==========

    constructor() EIP712("Nitrolite:Custody", "0.3.0") {
        // No state initialization needed
    }

    // ========== Read methods ==========

    /**
     * @notice Gets the balances of multiple accounts for multiple tokens
     * @dev Returns a 2D array where each inner array corresponds to the balances of the tokens for each account
     * @param accounts Array of account addresses to check balances for
     * @param tokens Array of token addresses to check balances for (use address(0) for native tokens)
     * @return A 2D array of balances, where each inner array corresponds to the balances of the tokens for each account
     */
    function getAccountsBalances(address[] calldata accounts, address[] calldata tokens)
        external
        view
        returns (uint256[][] memory)
    {
        uint256[][] memory balances = new uint256[][](accounts.length);
        for (uint256 i = 0; i < accounts.length; i++) {
            balances[i] = new uint256[](tokens.length);
            for (uint256 j = 0; j < tokens.length; j++) {
                balances[i][j] = _ledgers[accounts[i]].tokens[tokens[j]];
            }
        }
        return balances;
    }

    /**
     * @notice Get the list of open channels for a list of accounts
     * @param accounts Array of account addresses to check for open channels
     * @return Array of arrays, where each inner array contains channel IDs for the corresponding account
     */
    function getOpenChannels(address[] memory accounts) external view returns (bytes32[][] memory) {
        bytes32[][] memory channels = new bytes32[][](accounts.length);
        for (uint256 i = 0; i < accounts.length; i++) {
            channels[i] = _ledgers[accounts[i]].channels.values();
        }
        return channels;
    }

    /**
     * @notice Get detailed information about a specific channel
     * @param channelId The unique identifier of the channel
     * @return channel The Channel configuration
     * @return status The current status of the channel
     * @return wallets The list of wallets that have funded the channel
     * @return challengeExpiry The challenge expiry timestamp
     * @return lastValidState The last valid state of the channel
     */
    function getChannelData(bytes32 channelId)
        external
        view
        returns (
            Channel memory channel,
            ChannelStatus status,
            address[] memory wallets,
            uint256 challengeExpiry,
            State memory lastValidState
        )
    {
        Metadata storage meta = _channels[channelId];
        channel = meta.chan;
        status = meta.stage;
        wallets = new address[](PART_NUM);
        for (uint256 i = 0; i < PART_NUM; i++) {
            wallets[i] = meta.wallets[i];
        }
        challengeExpiry = meta.challengeExpire;
        lastValidState = meta.lastValidState;
    }

    /**
     * @notice Get the balance of a channel for a list of tokens
     * @param channelId The unique identifier of the channel
     * @param tokens Array of token addresses to check balances for (use address(0) for native tokens)
     * @return balances Array of balances corresponding to the provided tokens
     */
    function getChannelBalances(bytes32 channelId, address[] memory tokens)
        external
        view
        returns (uint256[] memory balances)
    {
        balances = new uint256[](tokens.length);
        for (uint256 i = 0; i < tokens.length; i++) {
            balances[i] = _channels[channelId].tokenBalances[tokens[i]];
        }
    }

    // ========== Write public methods ==========

    /**
     * @notice Deposit tokens into the contract
     * @param account The account that funds will be deposited to
     * @param token The token address to deposit (use address(0) for native tokens)
     * @param amount The amount of tokens to deposit
     */
    function deposit(address account, address token, uint256 amount) public payable {
        if (amount == 0) revert InvalidAmount();

        if (token == address(0)) {
            if (msg.value != amount) revert InvalidValue();
        } else {
            if (msg.value != 0) revert InvalidValue();
        }

        _ledgers[account].tokens[token] += amount;

        if (token != address(0)) {
            IERC20(token).safeTransferFrom(msg.sender, address(this), amount);
        }

        emit Deposited(account, token, amount);
    }

    /**
     * @notice Withdraw tokens from the contract
     * @dev Can only withdraw available (not locked in channels) funds
     * @param token The token address to withdraw (use address(0) for native tokens)
     * @param amount The amount of tokens to withdraw
     */
    function withdraw(address token, uint256 amount) external {
        address account = msg.sender;
        Ledger storage ledger = _ledgers[account];
        uint256 available = ledger.tokens[token];
        if (available < amount) revert InsufficientBalance(available, amount);

        ledger.tokens[token] -= amount;

        _transfer(token, account, amount);

        emit Withdrawn(account, token, amount);
    }

    /**
     * @notice Create a channel by depositing assets
     * @param ch Channel configuration
     * @param initial is the initial State defined by the opener, it contains the expected allocation
     * @return channelId Unique identifier for the channel
     */
    function create(Channel calldata ch, State calldata initial) public returns (bytes32 channelId) {
        // TODO: add checks that there are only 2 allocations, they have the same token (here and throughout the code)
        // checks
        if (
            ch.participants.length != PART_NUM || ch.participants[CLIENT_IDX] == address(0)
                || ch.participants[SERVER_IDX] == address(0) || ch.participants[CLIENT_IDX] == ch.participants[SERVER_IDX]
        ) revert InvalidParticipant();
        if (ch.adjudicator == address(0)) revert InvalidAdjudicator();
        if (ch.challenge < MIN_CHALLENGE_PERIOD) revert InvalidChallengePeriod();

        // TODO: security hardening: check that `participants[0]` is authorized by the wallet

        // TODO: replace with `require(...)`
        if (initial.intent != StateIntent.INITIALIZE) revert InvalidState();
        if (initial.version != 0) revert InvalidState();

        channelId = Utils.getChannelId(ch);
        if (_channels[channelId].stage != ChannelStatus.VOID) revert InvalidStatus();

        if (initial.sigs.length == 0 || initial.sigs.length > PART_NUM) revert InvalidStateSignatures();

        // TODO: later we can lift the restriction that first sig must be from CLIENT
        if (
            !initial.verifyStateSignature(
                channelId, _domainSeparatorV4(), initial.sigs[CLIENT_IDX], ch.participants[CLIENT_IDX]
            )
        ) {
            revert InvalidStateSignatures();
        }

        // NOTE: even if there is not allocation planned, it should be present as `Allocation{address(0), 0}`
        if (initial.allocations.length != PART_NUM) revert InvalidAllocations();

        // effects
        address wallet = msg.sender;
        Metadata storage meta = _channels[channelId];
        meta.chan = ch;
        meta.stage = ChannelStatus.INITIAL;
        meta.wallets[CLIENT_IDX] = wallet;
        meta.lastValidState = initial;

        // NOTE: allocations MUST come in the same order as participants in deposit
        for (uint256 i = 0; i < PART_NUM; i++) {
            address token = initial.allocations[i].token;
            uint256 amount = initial.allocations[i].amount;

            // even if participant does not have an allocation, still track that
            meta.expectedDeposits[i] = Amount({token: token, amount: amount});
            meta.actualDeposits[i] = Amount({token: address(0), amount: 0}); // Initialize actual deposits to zero
        }

        // NOTE: it is allowed for depositor (and wallet) to be different from channel creator (participant)
        // This enables logic of "session keys" where a user can create a channel on behalf of another account, but will lock their own funds
        // if (ch.participants[CLIENT_IDX]; != wallet) revert InvalidParticipant();
        Amount memory creatorDeposit = meta.expectedDeposits[CLIENT_IDX];
        meta.actualDeposits[CLIENT_IDX] = creatorDeposit;
        _ledgers[ch.participants[CLIENT_IDX]].channels.add(channelId);

        // interactions
        _lockAccountFundsToChannel(wallet, channelId, creatorDeposit.token, creatorDeposit.amount);

        emit Created(channelId, wallet, ch, initial);

        if (initial.sigs.length == PART_NUM) {
            address serverAddress = ch.participants[SERVER_IDX];
            if (!initial.verifyStateSignature(channelId, _domainSeparatorV4(), initial.sigs[SERVER_IDX], serverAddress))
            {
                revert InvalidStateSignatures();
            }

            meta.stage = ChannelStatus.ACTIVE;
            Amount memory expectedDeposit = meta.expectedDeposits[SERVER_IDX];
            meta.actualDeposits[SERVER_IDX] = expectedDeposit;
            meta.wallets[SERVER_IDX] = serverAddress;
            _ledgers[serverAddress].channels.add(channelId);

            _lockAccountFundsToChannel(serverAddress, channelId, expectedDeposit.token, expectedDeposit.amount);

            emit Joined(channelId, SERVER_IDX);
            emit Opened(channelId);
        }

        return channelId;
    }

    /**
     * @notice Deposit funds and create a channel
     * @dev This function allows a user to deposit funds and create a channel in one transaction
     * @param token The token address to deposit (use address(0) for native tokens)
     * @param amount The amount of tokens to deposit
     * @param ch Channel configuration with participants, adjudicator, challenge period, and nonce
     * @param initial Initial state with StateIntent.INITIALIZE and expected allocations
     * @return channelId Unique identifier for the created channel
     */
    function depositAndCreate(address token, uint256 amount, Channel calldata ch, State calldata initial)
        external
        payable
        returns (bytes32)
    {
        deposit(msg.sender, token, amount);
        return create(ch, initial);
    }

    /**
     * @notice Allows a SERVER to join a channel by signing the funding state
     * @param channelId Unique identifier for the channel
     * @param index Index of the participant in the channel's participants array (must be 1 for SERVER)
     * @param sig Signature of SERVER on the funding state
     * @return The channelId of the joined channel
     */
    function join(bytes32 channelId, uint256 index, bytes calldata sig) external returns (bytes32) {
        Metadata storage meta = _channels[channelId];

        // checks
        if (meta.stage == ChannelStatus.VOID) revert ChannelNotFound(channelId);
        if (meta.stage != ChannelStatus.INITIAL) revert InvalidStatus();

        if (index != SERVER_IDX) revert InvalidParticipant();
        if (meta.actualDeposits[SERVER_IDX].amount != 0) revert DepositAlreadyFulfilled();

        if (
            !meta.lastValidState.verifyStateSignature(
                channelId, _domainSeparatorV4(), sig, meta.chan.participants[SERVER_IDX]
            )
        ) revert InvalidStateSignatures();

        State memory lastValidState = meta.lastValidState;
        bytes[] memory sigs = new bytes[](PART_NUM);
        sigs[CLIENT_IDX] = lastValidState.sigs[CLIENT_IDX];
        sigs[SERVER_IDX] = sig;
        lastValidState.sigs = sigs;

        if (!IAdjudicator(meta.chan.adjudicator).adjudicate(meta.chan, lastValidState, new State[](0))) {
            revert InvalidState();
        }

        // effects
        Amount memory expectedDeposit = meta.expectedDeposits[SERVER_IDX];
        address wallet = msg.sender;

        meta.actualDeposits[SERVER_IDX] = expectedDeposit;
        meta.wallets[SERVER_IDX] = wallet;
        meta.lastValidState = lastValidState;
        meta.stage = ChannelStatus.ACTIVE;

        _ledgers[meta.chan.participants[SERVER_IDX]].channels.add(channelId);

        // interactions
        _lockAccountFundsToChannel(wallet, channelId, expectedDeposit.token, expectedDeposit.amount);

        emit Joined(channelId, SERVER_IDX);
        emit Opened(channelId);

        return channelId;
    }

    /**
     * @notice Finalize the channel with a mutually signed state
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * NOTE: Custody implementation does NOT require the `proofs` parameter for the close function.
     */
    function close(bytes32 channelId, State calldata candidate, State[] calldata) public {
        Metadata storage meta = _channels[channelId];

        // checks
        if (meta.stage == ChannelStatus.VOID) revert ChannelNotFound(channelId);

        if (meta.stage == ChannelStatus.ACTIVE) {
            if (candidate.intent != StateIntent.FINALIZE) revert InvalidState();
            if (candidate.version == 0) revert InvalidState();

            if (candidate.sigs.length != PART_NUM) revert InvalidStateSignatures();
            if (!_verifyAllSignatures(meta.chan, candidate)) revert InvalidStateSignatures();

            meta.lastValidState = candidate;
        } else if (meta.stage == ChannelStatus.DISPUTE) {
            // Can overwrite any challenge state with a valid final state
            if (block.timestamp < meta.challengeExpire) {
                if (candidate.intent != StateIntent.FINALIZE) revert InvalidState();

                if (!_verifyAllSignatures(meta.chan, candidate)) revert InvalidStateSignatures();

                meta.challengeExpire = 0;
                meta.lastValidState = candidate;
            } else {
                // Already in DISPUTE with an expired challenge - can proceed to finalization
            }
        } else {
            revert InvalidStatus();
        }

        _closeEffectsAndInteractions(channelId, meta.lastValidState.allocations);

        emit Closed(channelId, candidate);
    }

    /**
     * @notice Unilaterally post a state when the other party is uncooperative
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     * @param challengerSig Challenger signature over `keccak256(abi.encode(stateHash, "challenge"))` to disallow 3rd party
     * to challenge with a stolen state and its signature
     */
    function challenge(
        bytes32 channelId,
        State calldata candidate,
        State[] calldata proofs,
        bytes calldata challengerSig
    ) external {
        Metadata storage meta = _channels[channelId];

        // checks
        if (meta.stage == ChannelStatus.VOID) revert ChannelNotFound(channelId);
        if (meta.stage == ChannelStatus.DISPUTE || meta.stage == ChannelStatus.FINAL) revert InvalidStatus();
        if (candidate.intent == StateIntent.FINALIZE) revert InvalidState();

        _requireChallengerIsParticipant(channelId, candidate, meta.chan.participants, challengerSig);

        StateIntent lastValidStateIntent = meta.lastValidState.intent;

        if (meta.stage == ChannelStatus.INITIAL) {
            // main goal: verify Candidate == LastValidState, close channel
            if (!Utils.statesAreEqual(candidate, meta.lastValidState)) {
                revert InvalidState();
            }

            _closeEffectsAndInteractions(channelId, candidate.allocations);

            emit Challenged(channelId, candidate, block.timestamp);
            emit Closed(channelId, candidate);
            return;
        }

        // meta.stage == ChannelStatus.ACTIVE
        // main goal: verify Candidate is valid and >= LastValidState (in RESIZE case states should be equal)
        if (lastValidStateIntent == StateIntent.INITIALIZE) {
            if (candidate.intent == StateIntent.INITIALIZE) {
                if (!Utils.statesAreEqual(candidate, meta.lastValidState)) revert InvalidState();
            } else {
                if (!_isMoreRecent(meta.chan.adjudicator, candidate, meta.lastValidState)) revert InvalidState();
                if (!IAdjudicator(meta.chan.adjudicator).adjudicate(meta.chan, candidate, proofs)) {
                    revert InvalidState();
                }
            }
        } else if (lastValidStateIntent == StateIntent.OPERATE) {
            if (candidate.intent != StateIntent.OPERATE) revert InvalidState();
            if (!Utils.statesAreEqual(candidate, meta.lastValidState)) {
                if (!_isMoreRecent(meta.chan.adjudicator, candidate, meta.lastValidState)) revert InvalidState();
                if (!IAdjudicator(meta.chan.adjudicator).adjudicate(meta.chan, candidate, proofs)) {
                    revert InvalidState();
                }
            }
        } else if (lastValidStateIntent == StateIntent.RESIZE) {
            if (candidate.intent == StateIntent.INITIALIZE) revert InvalidState();
            if (candidate.intent == StateIntent.OPERATE) {
                if (!_isMoreRecent(meta.chan.adjudicator, candidate, meta.lastValidState)) revert InvalidState();
                if (!IAdjudicator(meta.chan.adjudicator).adjudicate(meta.chan, candidate, proofs)) {
                    revert InvalidState();
                }
            } else if (candidate.intent == StateIntent.RESIZE) {
                if (!Utils.statesAreEqual(candidate, meta.lastValidState)) {
                    revert InvalidState();
                }
            } else {
                revert InvalidState(); // should not happen, but added for readability
            }
        } else {
            revert InvalidState(); // should not happen, but added for readability
        }

        // effects
        uint256 challengeExpiration = block.timestamp + meta.chan.challenge;
        meta.challengeExpire = challengeExpiration;
        meta.lastValidState = candidate;
        meta.stage = ChannelStatus.DISPUTE;

        emit Challenged(channelId, candidate, challengeExpiration);
    }

    /**
     * @notice Unilaterally post a state to store it on-chain to prevent future disputes
     * @param channelId Unique identifier for the channel
     * @param candidate The latest known valid state
     * @param proofs is an array of valid state required by the adjudicator
     */
    function checkpoint(bytes32 channelId, State calldata candidate, State[] calldata proofs) external {
        Metadata storage meta = _channels[channelId];

        // checks
        if (meta.stage == ChannelStatus.VOID) revert ChannelNotFound(channelId);
        if (meta.stage == ChannelStatus.FINAL) revert InvalidStatus();

        // if INITIALIZE, call `join(...)`. If RESIZE, call `resize(...)`. If FINALIZE, call `close(...)`.
        if (candidate.intent != StateIntent.OPERATE) {
            revert InvalidState();
        }

        StateIntent lastValidStateIntent = meta.lastValidState.intent;

        // main goal: verify Candidate is valid and > LastValidState

        if (meta.stage == ChannelStatus.INITIAL) {
            revert InvalidStatus(); // Cannot checkpoint in INITIAL stage, use `join(...)` instead
        } else if (meta.stage == ChannelStatus.ACTIVE) {
            if (!_isMoreRecent(meta.chan.adjudicator, candidate, meta.lastValidState)) revert InvalidState();
            if (!IAdjudicator(meta.chan.adjudicator).adjudicate(meta.chan, candidate, proofs)) revert InvalidState();
        } else {
            // meta.stage == ChannelStatus.DISPUTE
            if (!IAdjudicator(meta.chan.adjudicator).adjudicate(meta.chan, candidate, proofs)) revert InvalidState();

            if (lastValidStateIntent == StateIntent.OPERATE) {
                if (!_isMoreRecent(meta.chan.adjudicator, candidate, meta.lastValidState)) revert InvalidState();
            }

            meta.challengeExpire = 0;
        }

        // effects
        meta.stage = ChannelStatus.ACTIVE;
        meta.lastValidState = candidate;

        emit Checkpointed(channelId, candidate);
    }

    /**
     * @notice All participants agree in setting a new allocation resulting in locking or unlocking funds
     * @dev Used for resizing channel allocations without withdrawing funds
     * @param channelId Unique identifier for the channel to resize
     * @param candidate The state that is to be true after resizing, containing the delta allocations
     * @param proofs An array of states supporting the claim that the candidate is true
     * NOTE: proof is needed to improve UX and allow resized state to follow any state (no need for consensus)
     */
    function resize(bytes32 channelId, State calldata candidate, State[] calldata proofs) external {
        Metadata storage meta = _channels[channelId];

        // checks
        if (meta.stage == ChannelStatus.VOID) revert ChannelNotFound(channelId);
        if (meta.stage != ChannelStatus.ACTIVE) revert InvalidStatus();

        if (proofs.length == 0) revert InvalidState();
        if (candidate.intent != StateIntent.RESIZE) revert InvalidState();
        State memory precedingState = proofs[0];
        if (candidate.version != precedingState.version + 1) revert InvalidState();

        _requireCorrectAllocations(precedingState.allocations);
        _requireCorrectAllocations(candidate.allocations);

        // Verify all participants have signed the resize state
        if (!_verifyAllSignatures(meta.chan, candidate)) revert InvalidStateSignatures();

        // Decode the resize amounts
        // TODO: extract `int256[]` into an alias type
        int256[] memory resizeAmounts = abi.decode(candidate.data, (int256[]));

        _requireCorrectDelta(precedingState.allocations, candidate.allocations, resizeAmounts);

        // NOTE: this is required as `proofs[0:]` over arrays of dynamic types (State is dynamic) is not supported by Solidity compiler as of 0.8.29.
        State[] memory precedingProofs = new State[](proofs.length - 1);
        for (uint256 i = 1; i < proofs.length; i++) {
            precedingProofs[i - 1] = proofs[i];
        }

        if (!IAdjudicator(meta.chan.adjudicator).adjudicate(meta.chan, precedingState, precedingProofs)) {
            revert InvalidState();
        }

        // effects
        meta.lastValidState = candidate;

        // interactions
        _processResize(channelId, meta, resizeAmounts, candidate.allocations);

        emit Resized(channelId, resizeAmounts);
    }

    // ========== Write internal methods ==========

    function _transfer(address token, address to, uint256 amount) internal {
        if (token == address(0)) {
            (bool success,) = to.call{value: amount}("");
            if (!success) revert TransferFailed(token, to, amount);
        } else {
            IERC20(token).safeTransfer(to, amount);
        }
    }

    /**
     * @notice Lock funds from an account to a channel
     * @dev Used during channel creation and joining for 2-participant channels
     */
    function _lockAccountFundsToChannel(address account, bytes32 channelId, address token, uint256 amount) internal {
        if (amount == 0) return;

        Ledger storage ledger = _ledgers[account];
        uint256 available = ledger.tokens[token];
        if (available < amount) revert InsufficientBalance(available, amount);

        ledger.tokens[token] = available - amount; // avoiding "-=" saves gas on a storage lookup
        _channels[channelId].tokenBalances[token] += amount;
    }

    function _closeEffectsAndInteractions(bytes32 channelId, Allocation[] memory allocations) internal {
        Metadata storage meta = _channels[channelId];

        // effects
        // NOTE: FINAL state is ephemeral because `meta` is deleted at the end of this function
        meta.stage = ChannelStatus.FINAL;

        // interactions
        _unlockAllocations(channelId, allocations);

        // "delete" effects
        for (uint256 i = 0; i < PART_NUM; i++) {
            address participant = meta.chan.participants[i];
            _ledgers[participant].channels.remove(channelId);
        }

        delete _channels[channelId];
    }

    /**
     * @notice Internal function to close a channel and distribute funds
     * @param channelId The channel identifier
     * @param allocations The allocations to distribute
     */
    function _unlockAllocations(bytes32 channelId, Allocation[] memory allocations) internal {
        if (allocations.length != PART_NUM) revert InvalidState();

        for (uint256 i = 0; i < PART_NUM; i++) {
            _unlockAllocation(channelId, allocations[i]);
        }
    }

    // Does not perform checks to allow transferring partial balances in case of partial deposit
    function _unlockAllocation(bytes32 channelId, Allocation memory alloc) internal {
        if (alloc.amount == 0) return;

        Metadata storage meta = _channels[channelId];
        uint256 channelBalance = meta.tokenBalances[alloc.token];
        if (channelBalance == 0) return;

        uint256 correctedAmount = channelBalance > alloc.amount ? alloc.amount : channelBalance;
        meta.tokenBalances[alloc.token] = channelBalance - correctedAmount; // avoiding "-=" saves gas on a storage lookup
        _ledgers[alloc.destination].tokens[alloc.token] += correctedAmount;
    }

    /**
     * @notice Verifies that both signatures are valid for the given state in a 2-participant channel
     * @param chan The channel configuration
     * @param state The state to verify signatures for
     * @return valid True if both signatures are valid
     */
    function _verifyAllSignatures(Channel memory chan, State memory state) internal returns (bool valid) {
        if (state.sigs.length != PART_NUM) {
            return false;
        }

        bytes32 channelId = Utils.getChannelId(chan);

        for (uint256 i = 0; i < PART_NUM; i++) {
            if (!state.verifyStateSignature(channelId, _domainSeparatorV4(), state.sigs[i], chan.participants[i])) {
                return false;
            }
        }

        return true;
    }

    function _requireChallengerIsParticipant(
        bytes32 channelId,
        State memory state,
        address[] memory participants,
        bytes memory challengerSig
    ) internal view {
        // NOTE: ERC-6492 signature is NOT checked as at this point participants should already be deployed

        // NOTE: the "challenge" suffix substitution for raw ECDSA and EIP-191 signatures
        bytes memory packedChallengeState = abi.encodePacked(Utils.getPackedState(channelId, state), "challenge");
        address rawSigner = Utils.recoverRawECDSASigner(packedChallengeState, challengerSig);
        address eip191Signer = Utils.recoverEIP191Signer(packedChallengeState, challengerSig);
        address eip712Signer = Utils.recoverStateEIP712Signer(
            _domainSeparatorV4(), CHALLENGE_STATE_TYPEHASH, channelId, state, challengerSig
        );

        for (uint256 i = 0; i < participants.length; i++) {
            address participant = participants[i];
            if (participant.code.length != 0) {
                if (Utils.isValidERC1271Signature(keccak256(packedChallengeState), challengerSig, participant)) {
                    return;
                }
            } else {
                if (rawSigner == participant || eip191Signer == participant || eip712Signer == participant) {
                    return;
                }
            }
        }

        revert InvalidChallengerSignature();
    }

    /**
     * @notice Helper function to compare two states for recency
     * @param adjudicator The adjudicator contract address
     * @param candidate The candidate state
     * @param previous The previous state to compare against
     * @return True if the candidate state is strictly more recent than the previous state
     * @dev Returns false if states have equal version numbers or if candidate is older
     */
    function _isMoreRecent(address adjudicator, State memory candidate, State memory previous)
        internal
        view
        returns (bool)
    {
        // TODO: add support to ERC-165
        // Try to use IComparable if the adjudicator implements it
        // TODO: remove comparable altogether?
        try IComparable(adjudicator).compare(candidate, previous) returns (int8 result) {
            // Must return strictly positive result (>0), equal versions (==0) are not considered more recent
            return result > 0;
        } catch {
            // If IComparable is not implemented, fall back to comparing version numbers
            // Must be strictly greater, equal versions are not considered more recent
            return candidate.version > previous.version;
        }
    }

    function _requireCorrectAllocations(Allocation[] memory allocations) internal pure {
        if (allocations.length != PART_NUM) revert InvalidState();
        if (allocations[CLIENT_IDX].token != allocations[SERVER_IDX].token) revert InvalidState();
    }

    /// @notice Allows "implicit transfer" between CLIENT and SERVER, which is useful in situations where
    /// a participant wants to top-up a channel only to transfer funds to the other participant, so they can withdraw it
    /// @dev "implicit transfer" means that only the sum of "initial + resize == final" is checked, not the individual amounts_channels
    /// Explicit delta can be calculated as |final[i] - initial[i] - resize[i]|, where i can be CLIENT or SERVER
    function _requireCorrectDelta(
        Allocation[] memory initialAllocations,
        Allocation[] memory finalAllocations,
        int256[] memory delta
    ) internal pure {
        if (delta.length != PART_NUM) revert InvalidState();

        uint256 sumBefore = initialAllocations[CLIENT_IDX].amount + initialAllocations[SERVER_IDX].amount;
        int256 sumDelta = delta[CLIENT_IDX] + delta[SERVER_IDX];
        uint256 sumAfter = finalAllocations[CLIENT_IDX].amount + finalAllocations[SERVER_IDX].amount;

        if (int256(sumBefore) + sumDelta != int256(sumAfter)) {
            revert InvalidAllocations();
        }
    }

    /// @notice Supports "implicit transfer"
    /// @dev Positive deltas must be processed first as they add more funds to the channel that the negative delta may want to withdraw
    function _processResize(
        bytes32 channelId,
        Metadata storage chMeta,
        int256[] memory resizeAmounts,
        Allocation[] memory finalAllocations
    ) internal {
        // NOTE: all tokens are the same
        address token = chMeta.expectedDeposits[CLIENT_IDX].token;

        // First pass: Process all positive resizes
        for (uint256 i = 0; i < PART_NUM; i++) {
            if (resizeAmounts[i] > 0) {
                _lockAccountFundsToChannel(chMeta.wallets[i], channelId, token, uint256(resizeAmounts[i]));
            }
        }

        // Second pass: Process all negative resizes
        for (uint256 i = 0; i < PART_NUM; i++) {
            if (resizeAmounts[i] < 0) {
                _unlockAllocation(channelId, Allocation(chMeta.wallets[i], token, uint256(-resizeAmounts[i])));
            }
        }

        for (uint256 i = 0; i < PART_NUM; i++) {
            chMeta.expectedDeposits[i].amount = finalAllocations[i].amount;
            chMeta.actualDeposits[i].amount = finalAllocations[i].amount;
        }
    }
}
