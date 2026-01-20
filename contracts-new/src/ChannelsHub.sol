// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {IERC20} from "lib/openzeppelin-contracts/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "lib/openzeppelin-contracts/contracts/token/ERC20/utils/SafeERC20.sol";
import {SafeCast} from "lib/openzeppelin-contracts/contracts/utils/math/SafeCast.sol";
import {ReentrancyGuard} from "lib/openzeppelin-contracts/contracts/utils/ReentrancyGuard.sol";
import {EnumerableSet} from "lib/openzeppelin-contracts/contracts/utils/structs/EnumerableSet.sol";

import {IVault} from "./interfaces/IVault.sol";
import {Definition, ChannelStatus, EscrowStatus, CrossChainState, State, StateIntent} from "./interfaces/Types.sol";

import {Utils} from "./Utils.sol";
import {ChannelEngine} from "./ChannelEngine.sol";

/**
 * @title ChannelsHub
 * @notice Main contract implementing the Nitrolite state channel protocol (single-chain operations)
 * @dev Uses unified transition pattern with ChannelEngine library for validation
 */
contract ChannelsHub is IVault, ReentrancyGuard {
    using EnumerableSet for EnumerableSet.Bytes32Set;
    using SafeERC20 for IERC20;
    using SafeCast for int256;
    using SafeCast for uint256;
    using {Utils.validateSignatures, Utils.validateChallengerSignature} for CrossChainState;
    using {Utils.isEmpty} for State;

    error InvalidAddress();
    error InvalidAmount();
    error InvalidValue();
    error AddressCollision(address collision);
    error IncorrectChallengeDuration();

    error ChannelDoesNotExist(bytes32 channelId);

    struct ChannelMeta {
        ChannelStatus status;
        Definition definition;
        CrossChainState lastState;
        uint256 lockedFunds;
        uint64 challengeExpireAt;
    }

    struct EscrowDepositMeta {
        EscrowStatus status;
        address user;
        address node;
        uint64 unlockAt;
        uint64 challengeExpireAt;
        uint256 lockedAmount;
        CrossChainState initState;
    }

    struct EscrowWithdrawalMeta {
        EscrowStatus status;
        address user;
        address node;
        uint64 challengeExpireAt;
        uint256 lockedAmount;
        CrossChainState initState;
    }

    // ======== Contract Storage ==========

    // TODO: estimate these values better
    uint32 constant public MIN_CHALLENGE_DURATION = 1 days;

    uint32 constant public ESCROW_DEPOSIT_UNLOCK_DELAY = 12 hours;

    uint32 constant public MAX_DEPOSIT_ESCROW_PURGE = type(uint32).max;

    mapping(bytes32 channelId => ChannelMeta meta) internal _channels;
    mapping(address user => EnumerableSet.Bytes32Set channelIds) internal _userChannels;

    mapping(bytes32 escrowId => EscrowDepositMeta meta) internal _escrowDeposits;
    // sorted by `unlockAt` ascending
    bytes32[] internal _escrowDepositIds;
    // points to the first non-purged escrow deposit
    uint256 public escrowHead;

    mapping(bytes32 escrowId => EscrowWithdrawalMeta meta) internal _escrowWithdrawals;

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

    function getChannels(address user) external view returns (bytes32[] memory) {
        return _userChannels[user].values();
    }

    // filter only non-closed and non-migrated-out channels
    function getOpenChannels(address user) external view returns (bytes32[] memory) {
        bytes32[] memory allChannels = _userChannels[user].values();
        bytes32[] memory openChannelsTemp = new bytes32[](allChannels.length);
        uint256 count = 0;
        for (uint256 i = 0; i < allChannels.length; i++) {
            bytes32 channelId = allChannels[i];
            ChannelMeta memory meta = _channels[channelId];
            if (meta.status != ChannelStatus.CLOSED && meta.status != ChannelStatus.MIGRATED_OUT) {
                openChannelsTemp[count] = channelId;
                count++;
            }
        }
        bytes32[] memory openChannels = new bytes32[](count);
        for (uint256 i = 0; i < count; i++) {
            openChannels[i] = openChannelsTemp[i];
        }
        return openChannels;
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
        ChannelMeta memory meta = _channels[channelId];
        status = meta.status;
        definition = meta.definition;
        lastState = meta.lastState;
        challengeExpiry = meta.challengeExpireAt;
        lockedFunds = meta.lockedFunds;
    }

    function getEscrowDepositData(bytes32 escrowId)
        external
        view
        returns (
            EscrowStatus status,
            uint64 unlockAt,
            uint64 challengeExpiry,
            uint256 lockedAmount,
            CrossChainState memory initState
        ) {
        EscrowDepositMeta memory meta = _escrowDeposits[escrowId];
        status = meta.status;
        unlockAt = meta.unlockAt;
        challengeExpiry = meta.challengeExpireAt;
        lockedAmount = meta.lockedAmount;
        initState = meta.initState;
    }

    function getEscrowWithdrawalData(bytes32 escrowId)
        external
        view
        returns (
            EscrowStatus status,
            uint64 challengeExpiry,
            uint256 lockedAmount,
            CrossChainState memory initState
        ) {
        EscrowWithdrawalMeta memory meta = _escrowWithdrawals[escrowId];
        status = meta.status;
        challengeExpiry = meta.challengeExpireAt;
        lockedAmount = meta.lockedAmount;
        initState = meta.initState;
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

    // TODO: extract into a separate contract
    // ========= Escrow Deposit Purge ==========
    function getUnlockableEscrowDepositAmount() external view returns (uint256 totalUnlockable) {
        uint256 totalDeposits = _escrowDepositIds.length;
        uint256 escrowHeadTemp = escrowHead;

        while (escrowHeadTemp < totalDeposits) {
            bytes32 escrowId = _escrowDepositIds[escrowHeadTemp];
            EscrowDepositMeta storage meta = _escrowDeposits[escrowId];

            if (meta.unlockAt <= block.timestamp && meta.status == EscrowStatus.INITIALIZED) {
                totalUnlockable += meta.lockedAmount;
            } else {
                break;
            }

            escrowHeadTemp++;
        }
    }

    function getUnlockableEscrowDepositCount() external view returns (uint256 count) {
        uint256 totalDeposits = _escrowDepositIds.length;
        uint256 escrowHeadTemp = escrowHead;

        while (escrowHeadTemp < totalDeposits) {
            bytes32 escrowId = _escrowDepositIds[escrowHeadTemp];
            EscrowDepositMeta storage meta = _escrowDeposits[escrowId];

            if (meta.unlockAt <= block.timestamp && meta.status == EscrowStatus.INITIALIZED) {
                count++;
            } else {
                break;
            }

            escrowHeadTemp++;
        }
    }

    function getEscrowDepositIds(uint256 page, uint256 pageSize) external view returns (bytes32[] memory ids) {
        uint256 totalDeposits = _escrowDepositIds.length;
        uint256 start = page * pageSize;
        if (start >= totalDeposits) {
            return new bytes32[](0);
        }
        uint256 end = start + pageSize;
        if (end > totalDeposits) {
            end = totalDeposits;
        }
        ids = new bytes32[](end - start);
        for (uint256 i = start; i < end; i++) {
            ids[i - start] = _escrowDepositIds[i];
        }
    }

    function purgeEscrowDeposits(uint256 maxToPurge) external {
        _purgeEscrowDeposits(maxToPurge);
    }

    function _purgeEscrowDeposits() internal {
        _purgeEscrowDeposits(MAX_DEPOSIT_ESCROW_PURGE);
    }

    function _purgeEscrowDeposits(uint256 maxToPurge) internal {
        uint256 purgedCount = 0;
        uint256 totalDeposits = _escrowDepositIds.length;
        uint256 escrowHeadTemp = escrowHead;

        while (escrowHeadTemp < totalDeposits && purgedCount < maxToPurge) {
            bytes32 escrowId = _escrowDepositIds[escrowHeadTemp];
            EscrowDepositMeta storage meta = _escrowDeposits[escrowId];

            // only still "INITIALIZED" escrows can be purged: "CHALLENGED" escrows require manual finalization, while "FINALIZED" were already manually purged
            if (meta.unlockAt <= block.timestamp && meta.status == EscrowStatus.INITIALIZED) {
                _nodeBalances[meta.node][meta.initState.nonHomeState.token] += meta.lockedAmount;
                meta.status = EscrowStatus.FINALIZED;
                meta.lockedAmount = 0;
                purgedCount++;
            } else {
                break;
            }

            escrowHeadTemp++;
        }

        escrowHead = escrowHeadTemp;

        emit EscrowDepositsPurged(purgedCount);
    }

    event EscrowDepositsPurged(uint256 purgedCount);

    // ========== Channel lifecycle ==========

    function createChannel(Definition calldata def, CrossChainState calldata initCCS) external payable {
        bytes32 channelId = Utils.getChannelId(def);

        _requireValidDefinition(def);

        ChannelEngine.TransitionContext memory ctx = _buildContext(channelId, def.node, initCCS.homeState.token);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, initCCS);

        initCCS.validateSignatures(channelId, def.user, def.node);

        _applyEffects(channelId, def, initCCS, effects);
        _userChannels[def.user].add(channelId);

        emit ChannelCreated(channelId, def.user, def.node, def, initCCS);
    }

    event ChannelCreated(bytes32 indexed channelId, address indexed user, address indexed node, Definition definition, CrossChainState initialState);

    function depositToChannel(bytes32 channelId, CrossChainState calldata candidate) public payable {
        ChannelMeta storage meta = _channels[channelId];

        ChannelEngine.TransitionContext memory ctx = _buildContext(channelId, meta.definition.node, candidate.homeState.token);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        candidate.validateSignatures(channelId, meta.definition.user, meta.definition.node);

        _applyEffects(channelId, meta.definition, candidate, effects);

        emit ChannelDeposited(channelId, candidate);
    }

    event ChannelDeposited(bytes32 indexed channelId, CrossChainState candidate);

    function withdrawFromChannel(bytes32 channelId, CrossChainState calldata candidate) public payable {
        ChannelMeta storage meta = _channels[channelId];

        ChannelEngine.TransitionContext memory ctx = _buildContext(channelId, meta.definition.node, candidate.homeState.token);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        candidate.validateSignatures(channelId, meta.definition.user, meta.definition.node);

        _applyEffects(channelId, meta.definition, candidate, effects);

        emit ChannelWithdrawn(channelId, candidate);
    }

    event ChannelWithdrawn(bytes32 indexed channelId, CrossChainState candidate);

    function checkpointChannel(bytes32 channelId, CrossChainState calldata candidate, CrossChainState[] calldata proof) external payable {
        ChannelMeta storage meta = _channels[channelId];

        ChannelEngine.TransitionContext memory ctx = _buildContext(channelId, meta.definition.node, candidate.homeState.token);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        candidate.validateSignatures(channelId, meta.definition.user, meta.definition.node);

        _applyEffects(channelId, meta.definition, candidate, effects);

        emit ChannelCheckpointed(channelId, candidate);
    }

    event ChannelCheckpointed(bytes32 indexed channelId, CrossChainState candidate);

    function challengeChannel(bytes32 channelId, CrossChainState calldata candidate, CrossChainState[] calldata proof, bytes calldata challengerSig) external payable {
        ChannelMeta storage meta = _channels[channelId];

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
        meta.challengeExpireAt = challengeExpiry;

        emit ChannelChallenged(channelId, candidate, challengeExpiry);
    }

    event ChannelChallenged(bytes32 indexed channelId, CrossChainState candidate, uint256 challengeExpiry);

    function closeChannel(bytes32 channelId, CrossChainState calldata candidate, CrossChainState[] calldata proof) external payable {
        ChannelMeta storage meta = _channels[channelId];
        ChannelStatus status = meta.status;

        require(status == ChannelStatus.OPERATING || status == ChannelStatus.DISPUTED, "invalid channel status");

        CrossChainState memory prevState = meta.lastState;
        address node = meta.definition.node;
        address user = meta.definition.user;

        // Path 1: Unilateral closure after challenge timeout
        if (status == ChannelStatus.DISPUTED && block.timestamp > meta.challengeExpireAt) {
            meta.status = ChannelStatus.CLOSED;
            meta.lockedFunds = 0;
            meta.challengeExpireAt = 0;

            _pushFunds(user, prevState.homeState.token, prevState.homeState.userAllocation);
            _pushFunds(node, prevState.homeState.token, prevState.homeState.nodeAllocation);

            _userChannels[user].remove(channelId);

            emit ChannelClosed(channelId, prevState);
            return;
        }

        // Path 2: Cooperative closure with signed CLOSE state
        ChannelEngine.TransitionContext memory ctx = _buildContext(channelId, node, candidate.homeState.token);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        candidate.validateSignatures(channelId, user, node);

        _applyEffects(channelId, meta.definition, candidate, effects);
        _userChannels[user].remove(channelId);

        emit ChannelClosed(channelId, candidate);
    }

    event ChannelClosed(bytes32 indexed channelId, CrossChainState finalState);

    // ========= Cross-Chain Functions ==========

    function initiateEscrowDeposit(
        bytes32 channelId,
        CrossChainState calldata candidate
    ) external payable {
        // calculate escrowId
        // append escrowId to _escrowDepositIds
        revert("not implemented");
    }

    function challengeEscrowDeposit(
        bytes32 escrowId,
        CrossChainState calldata candidate
    ) external payable {
        revert("not implemented");
    }

    function finalizeEscrowDeposit(
        bytes32 escrowId,
        CrossChainState calldata candidate
    ) external payable {
        // mark escrow as finalized
        // unlock funds to node manually here
        revert("not implemented");
    }

    function initiateEscrowWithdrawal(
        bytes32 channelId,
        CrossChainState calldata candidate
    ) external payable {
        revert("not implemented");
    }

    function challengeEscrowWithdrawal(
        bytes32 escrowId,
        CrossChainState calldata candidate,
        bytes calldata challengerSig
    ) external payable {
        revert("not implemented");
    }

    function finalizeEscrowWithdrawal(
        bytes32 escrowId,
        CrossChainState calldata candidate
    ) external payable {
        revert("not implemented");
    }

    function initiateMigration(
        Definition memory def,
        CrossChainState calldata candidate
    ) external payable {
        revert("not implemented");
    }

    function finalizeMigration(
        bytes32 channelId,
        CrossChainState calldata candidate
    ) external payable {
        // _userChannels[user].remove(channelId);
        revert("not implemented");
    }

    // ========= Internal ==========

    function _buildContext(
        bytes32 channelId,
        address node,
        address token
    ) internal view returns (ChannelEngine.TransitionContext memory ctx) {
        ChannelMeta storage meta = _channels[channelId];

        ctx.status = meta.status;
        ctx.prevState = meta.lastState;
        ctx.lockedFunds = meta.lockedFunds;
        ctx.nodeAvailableFunds = _nodeBalances[node][token];
        ctx.challengeExpiry = meta.challengeExpireAt;

        return ctx;
    }

    function _applyEffects(
        bytes32 channelId,
        Definition memory def,
        CrossChainState calldata candidate,
        ChannelEngine.TransitionEffects memory effects
    ) internal {
        ChannelMeta storage meta = _channels[channelId];

        if (meta.status == ChannelStatus.VOID) {
            meta.definition = def;
        }

        _applyTransitionEffects(channelId, def, candidate, effects);

        if (effects.newStatus != ChannelStatus.VOID) {
            meta.status = effects.newStatus;
        }

        if (effects.clearDispute) {
            meta.status = ChannelStatus.OPERATING;
            meta.challengeExpireAt = 0;
        }

        if (effects.closeChannel) {
            meta.lockedFunds = 0;
            meta.challengeExpireAt = 0;
        }
    }

    function _applyTransitionEffects(
        bytes32 channelId,
        Definition memory def,
        CrossChainState calldata candidate,
        ChannelEngine.TransitionEffects memory effects
    ) internal {
        ChannelMeta storage meta = _channels[channelId];

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

        // NOTE: purge escrow deposits to unlock unutilized node liquidity
        _purgeEscrowDeposits();
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
