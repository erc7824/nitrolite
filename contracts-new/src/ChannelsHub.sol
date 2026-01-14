// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {IERC20} from "lib/openzeppelin-contracts/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "lib/openzeppelin-contracts/contracts/token/ERC20/utils/SafeERC20.sol";

import {IVault} from "./interfaces/IVault.sol";
import {Definition, ChannelStatus, CrossChainState, State, StateIntent} from "./interfaces/Types.sol";

import {Utils} from "./Utils.sol";

contract ChannelsHub is IVault {
    using {Utils.validateSignatures, Utils.validateChallengerSignature} for CrossChainState;
    using {Utils.isEmpty} for State;
    using SafeERC20 for IERC20;

    error InvalidAddress();
    error InvalidAmount();
    error InvalidValue();
    error AddressCollision(address collision);
    error IncorrectChallengeDuration();

    error ChannelDoesNotExist(bytes32 channelId);

    struct Metadata {
        ChannelStatus status;
        Definition definition;
        CrossChainState lastState;
        uint256 lockedFunds;
        uint64 challengeExpiry; // timestamp when challenge period ends
    }

    // ======== Contract Storage ==========

    uint32 constant MIN_CHALLENGE_DURATION = 1 days;

    mapping(bytes32 channelId => Metadata meta) internal _channels;

    mapping(address node => mapping(address token => uint256 balance)) internal _nodeBalances;

    // ========== Getters ==========

    // *** IVault ***

    function getAccountsBalances(address[] calldata accounts, address[] calldata tokens)
        external
        view
        returns (uint256[][] memory) {}

    // ******

    function getVaultBalance(address node, address token) external view returns (uint256) {
        // TODO: add escrowDepositUnlocked funds
        return _nodeBalances[node][token];
    }

    function getOpenChannels(address participant) external view returns (bytes32[] memory) {
        // return list of open channelIds between participant and node
    }

    function getChannelData(bytes32 channelId)
        external
        view
        returns (
            ChannelStatus status,
            Definition memory definition,
            CrossChainState memory lastState,
            uint256 challengeExpiry,
            uint256 lockedFunds
        ) {
        Metadata memory meta = _channels[channelId];
        status = meta.status;
        definition = meta.definition;
        lastState = meta.lastState;
        challengeExpiry = meta.challengeExpiry;
        lockedFunds = meta.lockedFunds;
    }

    function getEscrowDepositData(bytes32 escrowId)
        external
        view
        returns (
            Definition memory definition,
            CrossChainState memory lastState,
            uint256 participantLockedFunds,
            uint64 unlockExpiry,
            uint64 challengeExpiry
        ) {
        // return escrow deposit Metadata
    }

    function getEscrowWithdrawalData(bytes32 escrowId)
        external
        view
        returns (
            Definition memory definition,
            CrossChainState memory lastState,
            uint256 participantLockedFunds
        ) {
        // return escrow withdrawal Metadata
    }

    // *** IVault ***
    // TODO: replace with user-locked liquidity provision

    // user to deposit funds into node's balance
    function depositToVault(address node, address token, uint256 amount) external payable {
        require(node != address(0), InvalidAddress());
        require(amount > 0, InvalidAmount());

        _nodeBalances[node][token] += amount;

        _pullFunds(msg.sender, token, amount);

        emit Deposited(node, token, amount);
    }

    function withdrawFromVault(address to, address token, uint256 amount) external {
        require(to != address(0), InvalidAddress());
        require(amount > 0, InvalidAmount());

        uint256 currentBalance = _nodeBalances[msg.sender][token];
        require(currentBalance >= amount, "insufficient balance");

        _nodeBalances[msg.sender][token] = currentBalance - amount;

        _pushFunds(to, token, amount);

        emit Withdrawn(msg.sender, token, amount);
    }

    // ******

    // ========== Channel lifecycle ==========

    // usage:
    // - open a new channel with User deposit funds ("entry to Clearnet")
    // - open a new channel with Node transfer funds to User (User joins Clearnet after already receiving funds)
    // - open and close channel in one go for atomic for User receiving and withdrawing funds
    function createChannel(Definition calldata def, CrossChainState calldata initCCS) external payable {
        // -- checks --
        _requireValidDefinition(def);
        _requireValidState(initCCS);

        require(initCCS.version == 0, "invalid initial version");
        require(initCCS.intent == StateIntent.CREATE, "invalid initial intent");
        require(initCCS.nonHomeState.isEmpty(), "non-home state must be empty");
        require(initCCS.homeState.nodeAllocation == 0, "node balance must be zero in initial state");

        bytes32 channelId = Utils.getChannelId(def);

        initCCS.validateSignatures(channelId, def.participant, def.node);

        // -- effects --
        _channels[channelId] = Metadata({
            status: ChannelStatus.OPERATING,
            definition: def,
            lastState: initCCS,
            lockedFunds: initCCS.homeState.userAllocation,
            challengeExpiry: 0
        });

        // -- interactions --
        _pullFunds(
            def.participant,
            initCCS.homeState.token,
            initCCS.homeState.userAllocation
        );

        emit ChannelCreated(channelId, def.participant, def.node, def, initCCS);
    }

    event ChannelCreated(bytes32 indexed channelId, address indexed participant, address indexed node, Definition definition, CrossChainState initialState);

    function depositToChannel(bytes32 channelId, CrossChainState calldata candidate) public payable {
        // -- checks --
        ChannelStatus status = _validateChannelStatusAndChallenge(channelId);
        Metadata storage channelMeta = _channels[channelId];
        CrossChainState memory prevState = channelMeta.lastState;

        require(candidate.intent == StateIntent.DEPOSIT, "invalid intent");
        _requireValidState(candidate);

        address participant = channelMeta.definition.participant;
        address node = channelMeta.definition.node;

        _requireValidTransition(candidate, prevState, channelMeta.lockedFunds, _nodeBalances[node][candidate.homeState.token], false);

        int256 userDepositAmount = candidate.homeState.userNetFlow - prevState.homeState.userNetFlow;
        require(userDepositAmount > 0, "deposit amount must be positive");

        candidate.validateSignatures(channelId, participant, node);

        // -- effects --
        _clearDisputedStatus(channelId, status);
        _adjustNodeBalance(channelId, node, candidate.homeState.token, prevState, candidate);

        channelMeta.lastState = candidate;
        channelMeta.lockedFunds += uint256(userDepositAmount);

        // -- interactions --
        _pullFunds(
            participant,
            candidate.homeState.token,
            uint256(userDepositAmount)
        );

        emit ChannelDeposited(channelId, candidate);
    }

    event ChannelDeposited(bytes32 indexed channelId, CrossChainState candidate);

    function withdrawFromChannel(bytes32 channelId, CrossChainState calldata candidate) public payable {
        // -- checks --
        ChannelStatus status = _validateChannelStatusAndChallenge(channelId);
        Metadata storage channelMeta = _channels[channelId];
        CrossChainState memory prevState = channelMeta.lastState;

        require(candidate.intent == StateIntent.WITHDRAW, "invalid intent");
        _requireValidState(candidate);

        address participant = channelMeta.definition.participant;
        address node = channelMeta.definition.node;

        _requireValidTransition(candidate, prevState, channelMeta.lockedFunds, _nodeBalances[node][candidate.homeState.token], false);

        int256 userWithdrawalAmount = candidate.homeState.userNetFlow - prevState.homeState.userNetFlow;
        require(userWithdrawalAmount < 0, "invalid withdrawal amount");

        // Verify withdrawal doesn't exceed user allocation
        require(candidate.homeState.userAllocation <= prevState.homeState.userAllocation, "withdrawal exceeds allocation");

        candidate.validateSignatures(channelId, participant, node);

        // -- effects --
        _clearDisputedStatus(channelId, status);
        _adjustNodeBalance(channelId, node, candidate.homeState.token, prevState, candidate);

        channelMeta.lastState = candidate;
        channelMeta.lockedFunds -= uint256(-userWithdrawalAmount);

        // -- interactions --
        _pushFunds(
            participant,
            candidate.homeState.token,
            uint256(-userWithdrawalAmount)
        );

        emit ChannelWithdrawn(channelId, candidate);
    }

    event ChannelWithdrawn(bytes32 indexed channelId, CrossChainState candidate);

    // usage:
    // - "open" channel on new chain in case of chain migration
    function migrateChannelHere(Definition calldata def, CrossChainState calldata candidate, CrossChainState[] calldata proof) external payable {
        // -- checks --
        // require(candidate.homeChainId == block.chainid)
        // validate signatures over (channelId, CrossChainState)
        // validate proof:

        // -- effects --
        // store channel definition
        // store initial state
        // move `nodeLockedFunds` to `participantLockedFunds` in Metadata

        // -- interactions --
        // pull funds from participant
    }

    event ChannelMigrated(bytes32 indexed channelId, Definition definition, CrossChainState initialState);

    // usage:
    // - acknowledge latest funds change on Home chain
    // - resolve a challenge

    // - release Node's funds from a channel to the Node's internal vault balance (here)
    // - Node deposit funds into a home chain channel during bridging-deposit (here)
    function checkpointChannel(bytes32 channelId, CrossChainState calldata candidate, CrossChainState[] calldata proof) external payable {
        // -- checks --
        ChannelStatus status = _validateChannelStatusAndChallenge(channelId);
        Metadata storage channelMeta = _channels[channelId];
        CrossChainState memory prevState = channelMeta.lastState;

        require(candidate.intent == StateIntent.OPERATE, "invalid intent");
        _requireValidState(candidate);

        address participant = channelMeta.definition.participant;
        address node = channelMeta.definition.node;

        _requireValidTransition(candidate, prevState, channelMeta.lockedFunds, _nodeBalances[node][candidate.homeState.token], true);

        candidate.validateSignatures(channelId, participant, node);

        // -- effects --
        _clearDisputedStatus(channelId, status);
        _adjustNodeBalance(channelId, node, candidate.homeState.token, prevState, candidate);

        channelMeta.lastState = candidate;

        emit ChannelCheckpointed(channelId, candidate);
    }

    event ChannelCheckpointed(bytes32 indexed channelId, CrossChainState candidate);

    // usage:
    // - close a channel in case either party is unresponsive
    function challengeChannel(bytes32 channelId, CrossChainState calldata candidate, CrossChainState[] calldata proof, bytes calldata challengerSig) external payable {
        Metadata storage channelMeta = _channels[channelId];
        ChannelStatus status = channelMeta.status;

        if (status != ChannelStatus.OPERATING) {
            revert("invalid channel status");
        }

        CrossChainState memory prevState = channelMeta.lastState;
        require(candidate.version >= prevState.version, "challenge candidate must have higher version than previous state");

        address participant = channelMeta.definition.participant;
        address node = channelMeta.definition.node;

        if (candidate.version > prevState.version) {
            // check candidate
            require(candidate.intent == StateIntent.OPERATE, "invalid intent");
            _requireValidState(candidate);
            _requireValidTransition(candidate, prevState, channelMeta.lockedFunds, _nodeBalances[node][candidate.homeState.token], true);

            candidate.validateSignatures(channelId, participant, node);

            // -- effects --
            _adjustNodeBalance(channelId, node, candidate.homeState.token, prevState, candidate);
            channelMeta.lastState = candidate;
        } /* else {
            challenging with previous state, which was already processed
            only `candidate.version == prev.version` is checked
            `candidate == prevState` check is not required (but implied) as the protocol forbids a case where
            2 different states have the same version
            even if a different state with the same version is supplied, it does not affect
            the protocol, as the new state is not processed - only challenger signature is verified
        } */

        candidate.validateChallengerSignature(channelId, challengerSig, participant, node);

        channelMeta.status = ChannelStatus.DISPUTED;
        uint64 challengeExpiry = uint64(block.timestamp) + channelMeta.definition.challengeDuration;
        channelMeta.challengeExpiry = challengeExpiry;

        emit ChannelChallenged(channelId, candidate, challengeExpiry);
    }

    event ChannelChallenged(bytes32 indexed channelId, CrossChainState candidate, uint256 challengeExpiry);

    // usage:
    // - unilaterally close the channel withdrawing all funds
    // - "close" channel on old chain in case of chain migration
    function closeChannel(bytes32 channelId, CrossChainState calldata candidate, CrossChainState[] calldata proof) external payable {
        // -- checks --
        Metadata storage channelMeta = _channels[channelId];
        ChannelStatus status = channelMeta.status;

        if (status != ChannelStatus.OPERATING && status != ChannelStatus.DISPUTED) {
            revert("invalid channel status");
        }

        CrossChainState memory prevState = channelMeta.lastState;
        address node = channelMeta.definition.node;
        address participant = channelMeta.definition.participant;

        if (status == ChannelStatus.DISPUTED && block.timestamp > channelMeta.challengeExpiry) {
            // withdraw user funds according to lastState
            _pushFunds(
                participant,
                prevState.homeState.token,
                prevState.homeState.userAllocation
            );

            _pushFunds(
                node,
                prevState.homeState.token,
                prevState.homeState.nodeAllocation
            );

            emit ChannelClosed(channelId, prevState);
            return;
        }

        // status == ChannelStatus.OPERATING || status == ChannelStatus.DISPUTED && not expired

        // validate candidate
        require(candidate.intent == StateIntent.CLOSE, "invalid intent");
        _requireValidState(candidate);
        require(candidate.homeState.nodeAllocation == 0, "node allocation must be 0");
        // Additional closure validation
        require(candidate.homeState.userAllocation <= channelMeta.lockedFunds, "user allocation exceeds locked funds");
        _requireValidTransition(candidate, prevState, channelMeta.lockedFunds, _nodeBalances[node][candidate.homeState.token], true);


        candidate.validateSignatures(channelId, participant, node);

        // -- effects --
        _adjustNodeBalance(channelId, node, candidate.homeState.token, prevState, candidate);

        _pushFunds(
            participant,
            candidate.homeState.token,
            candidate.homeState.userAllocation
        );

        delete _channels[channelId];

        emit ChannelClosed(channelId, candidate);
    }

    event ChannelClosed(bytes32 indexed channelId, CrossChainState finalState);

    // ========= Escrow ==========

    struct EscrowDepositMetadata {
        Definition definition;
        CrossChainState lastState;
        uint256 participantLockedFunds;
        uint64 unlockExpiry;
        uint64 challengeExpiry;
    }

    // include Node address for simplicity?
    // EnumerableMap(bytes32 escrowId => EscrowDepositMetadata) escrowDeposits;

    // usage:
    // - lock user funds during bridging-deposit
    function initiateEscrowDeposit(Definition calldata def, CrossChainState calldata initCCS) external payable {
        // -- checks --
        // require(ccs.homeChainId != block.chainid)
        // validate signatures over (escrowId, Definition, CrossChainState)

        // -- effects --
        // store escrow data
        // add user funding amount to "escrow locked" amount
        // start `unlockTimer` (a constant on the Contract level)

        // -- interactions --
        // pull funds from user
    }

    event EscrowDepositInitiated(bytes32 indexed escrowId, address indexed participant, address indexed node, Definition definition, CrossChainState initialState);

    // usage:
    // - challenge bridging-deposit process
    function challengeEscrowDeposit(bytes32 escrowId, CrossChainState calldata candidate, CrossChainState[] calldata proof) external payable {}

    event EscrowDepositChallenged(bytes32 indexed escrowId, CrossChainState candidate, uint256 challengeExpiry);

    // usage: (optional)
    // - unlock Node's funds after bridging-deposit for better funds efficiency
    // - resolve a challenge during a bridging-deposit process
    function finalizeEscrowDeposit(bytes32 escrowId, CrossChainState calldata candidate, CrossChainState[2] calldata proof) external payable {}

    event EscrowDepositFinalized(bytes32 indexed escrowId, CrossChainState finalState);

    struct EscrowWithdrawalMetadata {
        Definition definition;
        CrossChainState lastState;
        uint256 participantLockedFunds;
    }

    // usage:
    // - lock Node's funds during bridging-withdrawal
    function initiateEscrowWithdrawal(Definition calldata def, CrossChainState calldata initCCS) external payable {}

    event EscrowWithdrawalInitiated(bytes32 indexed escrowId, address indexed participant, address indexed node, Definition definition, CrossChainState initialState);

    // usage:
    // - challenge bridging-withdrawal process
    function challengeEscrowWithdrawal(bytes32 escrowId, CrossChainState calldata candidate, CrossChainState[] calldata proof) external payable {}

    event EscrowWithdrawalChallenged(bytes32 indexed escrowId, CrossChainState candidate, uint256 challengeExpiry);

    // usage:
    // - unlock user funds during bridging-withdrawal
    function finalizeEscrowWithdrawal(bytes32 escrowId, CrossChainState calldata candidate) external payable {}

    event EscrowWithdrawalFinalized(bytes32 indexed escrowId, CrossChainState finalState);

    // ========= Internal ==========

    function _pullFunds(address from, address token, uint256 amount) internal {
        if (amount == 0) return;

        if (token == address(0)) {
            require(msg.value == amount, InvalidValue());
        } else {
            require(msg.value == 0, InvalidValue());
        }

        if (token != address(0)) {
            IERC20(token).safeTransferFrom(from, address(this), amount);
        }
    }

    function _pushFunds(address to, address token, uint256 amount) internal {
        if (amount == 0) return;

        if (token == address(0)) {
            payable(to).transfer(amount);
        } else {
            IERC20(token).safeTransfer(to, amount);
        }
    }


    function _requireValidDefinition(Definition calldata def) internal pure {
        require(def.participant != address(0), InvalidAddress());
        require(def.node != address(0), InvalidAddress());
        require(def.participant != def.node, AddressCollision(def.participant));
        require(def.challengeDuration >= MIN_CHALLENGE_DURATION, IncorrectChallengeDuration());
    }

    /**
     * @dev Validates that channel status is OPERATING or DISPUTED, and checks challenge expiry if DISPUTED
     * @param channelId The channel identifier
     * @return status The current channel status
     */
    function _validateChannelStatusAndChallenge(bytes32 channelId) internal view returns (ChannelStatus) {
        ChannelStatus status = _channels[channelId].status;
        if (status != ChannelStatus.OPERATING && status != ChannelStatus.DISPUTED) {
            revert("invalid channel status");
        }

        if (status == ChannelStatus.DISPUTED) {
            require(block.timestamp <= _channels[channelId].challengeExpiry, "challenge period has expired");
        }

        return status;
    }

    function _requireValidState(
        CrossChainState memory state
    ) internal view {
        require(state.homeState.chainId == block.chainid, "invalid home state chain id");

        uint256 allocsSum = state.homeState.userAllocation + state.homeState.nodeAllocation;
        int256 netFlowsSum = state.homeState.userNetFlow + state.homeState.nodeNetFlow;
        require(netFlowsSum >= 0, "negative net flow sum");
        require(allocsSum == uint256(netFlowsSum), "allocation/net flow sum mismatch");
    }

    /**
     * @dev Validates common candidate state properties against previous state, including
     * that allocations will match locked funds after all adjustments
     * Accounts for user net flow delta and node net flow delta
     * @param candidate The new state to validate
     * @param prevState The previous state to compare against
     * @param requireZeroUserDelta If true, requires user net flow delta to be zero
     */
    function _requireValidTransition(
        CrossChainState memory candidate,
        CrossChainState memory prevState,
        uint256 currentLockedFunds,
        uint256 nodeAvailableFunds,
        bool requireZeroUserDelta
    ) internal pure {
        require(candidate.version > prevState.version, "invalid version");
        require(candidate.homeState.token == prevState.homeState.token, "home state token mismatch");

        if (requireZeroUserDelta) {
            int256 userDeltaAmount = candidate.homeState.userNetFlow - prevState.homeState.userNetFlow;
            require(userDeltaAmount == 0, "user delta must be 0");
        }

        int256 userDelta = candidate.homeState.userNetFlow - prevState.homeState.userNetFlow;

        int256 nodeDelta = candidate.homeState.nodeNetFlow - prevState.homeState.nodeNetFlow;
        if (nodeDelta > 0) {
            require(nodeAvailableFunds >= uint256(nodeDelta), "insufficient node balance");
        }

        int256 expectedLockedFunds = int256(currentLockedFunds) + userDelta + nodeDelta;

        require(expectedLockedFunds >= 0, "negative locked funds");

        uint256 allocsSum = candidate.homeState.userAllocation + candidate.homeState.nodeAllocation;
        require(allocsSum == uint256(expectedLockedFunds), "locked funds consistency mismatch");

    }

    /**
     * @dev Adjusts node balance based on the net flow delta between states
     * @param channelId The channel identifier
     * @param node The node address
     * @param token The token address
     * @param prevState The previous state
     * @param newState The new state
     */
    function _adjustNodeBalance(
        bytes32 channelId,
        address node,
        address token,
        CrossChainState memory prevState,
        CrossChainState memory newState
    ) internal {
        int256 nodeDelta = newState.homeState.nodeNetFlow - prevState.homeState.nodeNetFlow;
        if (nodeDelta < 0) {
            // release Node's funds from the channel to the Node's internal vault balance
            _nodeBalances[node][token] += uint256(-nodeDelta);
            _channels[channelId].lockedFunds -= uint256(-nodeDelta);
        } else if (nodeDelta > 0) {
            // lock Node's funds into the channel from the Node's internal vault balance
            _nodeBalances[node][token] -= uint256(nodeDelta);
            _channels[channelId].lockedFunds += uint256(nodeDelta);
        }
    }

    /**
     * @dev Clears disputed status and resets challenge expiry if channel is in disputed state
     * @param channelId The channel identifier
     * @param status The current channel status
     */
    function _clearDisputedStatus(bytes32 channelId, ChannelStatus status) internal {
        if (status == ChannelStatus.DISPUTED) {
            _channels[channelId].status = ChannelStatus.OPERATING;
            _channels[channelId].challengeExpiry = 0;
        }
    }
}
