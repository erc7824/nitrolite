// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {IERC20} from "lib/openzeppelin-contracts/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "lib/openzeppelin-contracts/contracts/token/ERC20/utils/SafeERC20.sol";
import {SafeCast} from "lib/openzeppelin-contracts/contracts/utils/math/SafeCast.sol";
import {ReentrancyGuard} from "lib/openzeppelin-contracts/contracts/utils/ReentrancyGuard.sol";

import {IVault} from "./interfaces/IVault.sol";
import {Definition, ChannelStatus, CrossChainState, State, StateIntent} from "./interfaces/Types.sol";

import {Utils} from "./Utils.sol";
import {ChannelEngine} from "./ChannelEngine.sol";

/**
 * @title ChannelsHub
 * @notice Main contract implementing the Nitrolite state channel protocol (single-chain operations)
 * @dev Uses unified transition pattern with ChannelEngine library for validation
 */
contract ChannelsHub is IVault, ReentrancyGuard {
    using {Utils.validateSignatures, Utils.validateChallengerSignature} for CrossChainState;
    using {Utils.isEmpty} for State;
    using SafeERC20 for IERC20;
    using SafeCast for int256;
    using SafeCast for uint256;

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
        return _nodeBalances[node][token];
    }

    function getOpenChannels(address user) external view returns (bytes32[] memory) {
        // return list of open channelIds between user and node
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

    // *** IVault ***

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

    function createChannel(Definition calldata def, CrossChainState calldata initCCS) external payable {
        bytes32 channelId = Utils.getChannelId(def);

        _requireValidDefinition(def);

        ChannelEngine.TransitionContext memory ctx = _buildContext(channelId, def.node, initCCS.homeState.token);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, initCCS);

        initCCS.validateSignatures(channelId, def.user, def.node);

        _applyEffects(channelId, def, initCCS, effects);

        emit ChannelCreated(channelId, def.user, def.node, def, initCCS);
    }

    event ChannelCreated(bytes32 indexed channelId, address indexed user, address indexed node, Definition definition, CrossChainState initialState);

    function depositToChannel(bytes32 channelId, CrossChainState calldata candidate) public payable {
        Metadata storage meta = _channels[channelId];

        ChannelEngine.TransitionContext memory ctx = _buildContext(channelId, meta.definition.node, candidate.homeState.token);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        candidate.validateSignatures(channelId, meta.definition.user, meta.definition.node);

        _applyEffects(channelId, meta.definition, candidate, effects);

        emit ChannelDeposited(channelId, candidate);
    }

    event ChannelDeposited(bytes32 indexed channelId, CrossChainState candidate);

    function withdrawFromChannel(bytes32 channelId, CrossChainState calldata candidate) public payable {
        Metadata storage meta = _channels[channelId];

        ChannelEngine.TransitionContext memory ctx = _buildContext(channelId, meta.definition.node, candidate.homeState.token);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        candidate.validateSignatures(channelId, meta.definition.user, meta.definition.node);

        _applyEffects(channelId, meta.definition, candidate, effects);

        emit ChannelWithdrawn(channelId, candidate);
    }

    event ChannelWithdrawn(bytes32 indexed channelId, CrossChainState candidate);

    function checkpointChannel(bytes32 channelId, CrossChainState calldata candidate, CrossChainState[] calldata proof) external payable {
        Metadata storage meta = _channels[channelId];

        ChannelEngine.TransitionContext memory ctx = _buildContext(channelId, meta.definition.node, candidate.homeState.token);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        candidate.validateSignatures(channelId, meta.definition.user, meta.definition.node);

        _applyEffects(channelId, meta.definition, candidate, effects);

        emit ChannelCheckpointed(channelId, candidate);
    }

    event ChannelCheckpointed(bytes32 indexed channelId, CrossChainState candidate);

    function challengeChannel(bytes32 channelId, CrossChainState calldata candidate, CrossChainState[] calldata proof, bytes calldata challengerSig) external payable {
        Metadata storage meta = _channels[channelId];

        require(meta.status == ChannelStatus.OPERATING, "invalid channel status");

        CrossChainState memory prevState = meta.lastState;
        require(candidate.version >= prevState.version, "challenge candidate must have higher or equal version");

        address user = meta.definition.user;
        address node = meta.definition.node;

        // If version is higher, process the new state
        if (candidate.version > prevState.version) {
            require(candidate.intent == StateIntent.OPERATE, "invalid intent");

            ChannelEngine.TransitionContext memory ctx = _buildContext(channelId, node, candidate.homeState.token);
            ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

            candidate.validateSignatures(channelId, user, node);

            _applyTransitionEffects(channelId, meta.definition, candidate, effects);
        }
        // else: challenging with same version, state already processed

        candidate.validateChallengerSignature(channelId, challengerSig, user, node);

        meta.status = ChannelStatus.DISPUTED;
        uint64 challengeExpiry = uint64(block.timestamp) + meta.definition.challengeDuration;
        meta.challengeExpiry = challengeExpiry;

        emit ChannelChallenged(channelId, candidate, challengeExpiry);
    }

    event ChannelChallenged(bytes32 indexed channelId, CrossChainState candidate, uint256 challengeExpiry);

    function closeChannel(bytes32 channelId, CrossChainState calldata candidate, CrossChainState[] calldata proof) external payable {
        Metadata storage meta = _channels[channelId];
        ChannelStatus status = meta.status;

        require(status == ChannelStatus.OPERATING || status == ChannelStatus.DISPUTED, "invalid channel status");

        CrossChainState memory prevState = meta.lastState;
        address node = meta.definition.node;
        address user = meta.definition.user;

        // Path 1: Unilateral closure after challenge timeout
        if (status == ChannelStatus.DISPUTED && block.timestamp > meta.challengeExpiry) {
            meta.status = ChannelStatus.CLOSED;
            meta.lockedFunds = 0;
            meta.challengeExpiry = 0;

            _pushFunds(user, prevState.homeState.token, prevState.homeState.userAllocation);
            _pushFunds(node, prevState.homeState.token, prevState.homeState.nodeAllocation);

            emit ChannelClosed(channelId, prevState);
            return;
        }

        // Path 2: Cooperative closure with signed CLOSE state
        ChannelEngine.TransitionContext memory ctx = _buildContext(channelId, node, candidate.homeState.token);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        candidate.validateSignatures(channelId, user, node);

        _applyEffects(channelId, meta.definition, candidate, effects);

        emit ChannelClosed(channelId, candidate);
    }

    event ChannelClosed(bytes32 indexed channelId, CrossChainState finalState);

    // ========= Internal ==========

    function _buildContext(
        bytes32 channelId,
        address node,
        address token
    ) internal view returns (ChannelEngine.TransitionContext memory ctx) {
        Metadata storage meta = _channels[channelId];

        ctx.status = meta.status;
        ctx.prevState = meta.lastState;
        ctx.lockedFunds = meta.lockedFunds;
        ctx.nodeAvailableFunds = _nodeBalances[node][token];
        ctx.challengeExpiry = meta.challengeExpiry;

        return ctx;
    }

    function _applyEffects(
        bytes32 channelId,
        Definition memory def,
        CrossChainState calldata candidate,
        ChannelEngine.TransitionEffects memory effects
    ) internal {
        Metadata storage meta = _channels[channelId];

        if (meta.status == ChannelStatus.VOID) {
            meta.definition = def;
        }

        _applyTransitionEffects(channelId, def, candidate, effects);

        if (effects.newStatus != ChannelStatus.VOID) {
            meta.status = effects.newStatus;
        }

        if (effects.clearDispute) {
            meta.status = ChannelStatus.OPERATING;
            meta.challengeExpiry = 0;
        }

        if (effects.closeChannel) {
            meta.lockedFunds = 0;
            meta.challengeExpiry = 0;
        }
    }

    function _applyTransitionEffects(
        bytes32 channelId,
        Definition memory def,
        CrossChainState calldata candidate,
        ChannelEngine.TransitionEffects memory effects
    ) internal {
        Metadata storage meta = _channels[channelId];

        if (effects.updateLastState) {
            meta.lastState = candidate;
        }

        if (effects.userFundsDelta > 0) {
            uint256 amount = uint256(effects.userFundsDelta);
            _pullFunds(def.user, candidate.homeState.token, amount);
            meta.lockedFunds += amount;
        } else if (effects.userFundsDelta < 0) {
            uint256 amount = uint256(-effects.userFundsDelta);
            _pushFunds(def.user, candidate.homeState.token, amount);
            meta.lockedFunds -= amount;
        }

        if (effects.nodeFundsDelta > 0) {
            uint256 amount = uint256(effects.nodeFundsDelta);
            _nodeBalances[def.node][candidate.homeState.token] -= amount;
            meta.lockedFunds += amount;
        } else if (effects.nodeFundsDelta < 0) {
            uint256 amount = uint256(-effects.nodeFundsDelta);
            _nodeBalances[def.node][candidate.homeState.token] += amount;
            meta.lockedFunds -= amount;
        }

        // Special handling for CLOSE: push nodeAllocation directly to node address
        if (effects.closeChannel && candidate.homeState.nodeAllocation > 0) {
            _pushFunds(def.node, candidate.homeState.token, candidate.homeState.nodeAllocation);
            meta.lockedFunds -= candidate.homeState.nodeAllocation;
        }
    }

    function _pullFunds(address from, address token, uint256 amount) internal nonReentrant {
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

    function _pushFunds(address to, address token, uint256 amount) internal nonReentrant {
        if (amount == 0) return;

        if (token == address(0)) {
            payable(to).transfer(amount);
        } else {
            IERC20(token).safeTransfer(to, amount);
        }
    }

    function _requireValidDefinition(Definition calldata def) internal pure {
        require(def.user != address(0), InvalidAddress());
        require(def.node != address(0), InvalidAddress());
        require(def.user != def.node, AddressCollision(def.user));
        require(def.challengeDuration >= MIN_CHALLENGE_DURATION, IncorrectChallengeDuration());
    }
}
