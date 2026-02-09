// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {IERC20} from "lib/openzeppelin-contracts/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "lib/openzeppelin-contracts/contracts/token/ERC20/utils/SafeERC20.sol";
import {SafeCast} from "lib/openzeppelin-contracts/contracts/utils/math/SafeCast.sol";
import {ReentrancyGuard} from "lib/openzeppelin-contracts/contracts/utils/ReentrancyGuard.sol";
import {EnumerableSet} from "lib/openzeppelin-contracts/contracts/utils/structs/EnumerableSet.sol";
import {ECDSA} from "lib/openzeppelin-contracts/contracts/utils/cryptography/ECDSA.sol";
import {MessageHashUtils} from "lib/openzeppelin-contracts/contracts/utils/cryptography/MessageHashUtils.sol";

import {IVault} from "./interfaces/IVault.sol";
import {ISignatureValidator, ValidationResult, VALIDATION_FAILURE} from "./interfaces/ISignatureValidator.sol";
import {ChannelDefinition, ChannelStatus, EscrowStatus, State, Ledger, StateIntent, SigValidatorType} from "./interfaces/Types.sol";

import {Utils} from "./Utils.sol";
import {ChannelEngine} from "./ChannelEngine.sol";
import {EscrowDepositEngine} from "./EscrowDepositEngine.sol";
import {EscrowWithdrawalEngine} from "./EscrowWithdrawalEngine.sol";

/**
 * @title ChannelHub
 * @notice Main contract implementing the Nitrolite state channel protocol (single-chain operations)
 * @dev Uses unified transition pattern with ChannelEngine library for validation
 */
contract ChannelHub is IVault, ReentrancyGuard {
    using EnumerableSet for EnumerableSet.Bytes32Set;
    using SafeERC20 for IERC20;
    using SafeCast for int256;
    using SafeCast for uint256;
    using ECDSA for bytes32;
    using MessageHashUtils for bytes;
    using {Utils.isEmpty} for Ledger;

    uint8 public constant VERSION = 1;

    ISignatureValidator public immutable DEFAULT_SIG_VALIDATOR;

    event EscrowDepositsPurged(uint256 purgedCount);

    event ChannelCreated(
        bytes32 indexed channelId,
        address indexed user,
        address indexed node,
        ChannelDefinition definition,
        State initialState
    );
    event ChannelDeposited(bytes32 indexed channelId, State candidate);
    event ChannelWithdrawn(bytes32 indexed channelId, State candidate);
    event ChannelCheckpointed(bytes32 indexed channelId, State candidate);
    event ChannelChallenged(bytes32 indexed channelId, State candidate, uint64 challengeExpireAt);
    event ChannelClosed(bytes32 indexed channelId, State finalState);

    event EscrowDepositInitiated(bytes32 indexed escrowId, bytes32 indexed channelId, State state);
    event EscrowDepositInitiatedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, State state);
    event EscrowDepositChallenged(bytes32 indexed escrowId, State state, uint64 challengeExpireAt);
    event EscrowDepositFinalized(bytes32 indexed escrowId, bytes32 indexed channelId, State state);
    event EscrowDepositFinalizedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, State state);

    event EscrowWithdrawalInitiated(bytes32 indexed escrowId, bytes32 indexed channelId, State state);
    event EscrowWithdrawalInitiatedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, State state);
    event EscrowWithdrawalChallenged(bytes32 indexed escrowId, State state, uint64 challengeExpireAt);
    event EscrowWithdrawalFinalized(bytes32 indexed escrowId, bytes32 indexed channelId, State state);
    event EscrowWithdrawalFinalizedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, State state);

    event MigrationOutInitiated(bytes32 indexed channelId, State state);
    event MigrationInInitiated(bytes32 indexed channelId, State state);
    event MigrationOutFinalized(bytes32 indexed channelId, State state);
    event MigrationInFinalized(bytes32 indexed channelId, State state);

    error InvalidAddress();
    error InvalidAmount();
    error InvalidValue();
    error AddressCollision(address collision);
    error IncorrectChallengeDuration();
    error ChannelDoesNotExist(bytes32 channelId);

    struct ChannelMeta {
        ChannelStatus status;
        ChannelDefinition definition;
        State lastState;
        uint256 lockedFunds;
        uint64 challengeExpireAt;
    }

    struct EscrowDepositMeta {
        bytes32 channelId;
        EscrowStatus status;
        address user;
        address node;
        address signatureValidator;
        uint64 unlockAt;
        uint64 challengeExpireAt;
        uint256 lockedAmount;
        State initState;
    }

    struct EscrowWithdrawalMeta {
        bytes32 channelId;
        EscrowStatus status;
        address user;
        address node;
        address signatureValidator;
        uint64 challengeExpireAt;
        uint256 lockedAmount;
        State initState;
    }

    // ======== Contract Storage ==========

    // TODO: estimate these values better
    uint32 public constant MIN_CHALLENGE_DURATION = 1 days;

    uint32 public constant ESCROW_DEPOSIT_UNLOCK_DELAY = 3 hours;

    // NOTE: this value should not be small, so that as much escrow deposits as possible can be purged in one tx
    // but also not too large, to avoid hitting block gas limit during purge and incurring Denial-Of-Service attacks
    uint32 public constant MAX_DEPOSIT_ESCROW_PURGE = 64;

    mapping(bytes32 channelId => ChannelMeta meta) internal _channels;
    mapping(address user => EnumerableSet.Bytes32Set channelIds) internal _userChannels;

    mapping(bytes32 escrowId => EscrowDepositMeta meta) internal _escrowDeposits;
    // sorted by `unlockAt` ascending
    bytes32[] internal _escrowDepositIds;
    // points to the first non-purged escrow deposit
    uint256 public escrowHead;

    mapping(bytes32 escrowId => EscrowWithdrawalMeta meta) internal _escrowWithdrawals;

    mapping(address node => mapping(address token => uint256 balance)) internal _nodeBalances;

    // ========== Constructor ==========

    constructor(ISignatureValidator _defaultSigValidator) {
        require(address(_defaultSigValidator) != address(0), InvalidAddress());
        DEFAULT_SIG_VALIDATOR = _defaultSigValidator;
    }

    // ========== Getters ==========

    function getAccountBalance(address node, address token) external view returns (uint256) {
        return _nodeBalances[node][token];
    }

    function getChannelIds(address user) external view returns (bytes32[] memory) {
        return _userChannels[user].values();
    }

    // filter only non-closed and non-migrated-out channels
    function getOpenChannels(address user) external view returns (bytes32[] memory) {
        bytes32[] memory allChannels = _userChannels[user].values();
        uint256 openChannelCount = 0;
        for (uint256 i = 0; i < allChannels.length; i++) {
            if (
                _channels[allChannels[i]].status != ChannelStatus.CLOSED
                    && _channels[allChannels[i]].status != ChannelStatus.MIGRATED_OUT
            ) {
                openChannelCount++;
            }
        }

        bytes32[] memory openChannels = new bytes32[](openChannelCount);
        uint256 openChannelIndex = 0;
        for (uint256 i = 0; i < allChannels.length; i++) {
            if (
                _channels[allChannels[i]].status != ChannelStatus.CLOSED
                    && _channels[allChannels[i]].status != ChannelStatus.MIGRATED_OUT
            ) {
                openChannels[openChannelIndex] = allChannels[i];
                openChannelIndex++;
            }
        }

        return openChannels;
    }

    function getChannelData(bytes32 channelId)
        external
        view
        returns (
            ChannelStatus status,
            ChannelDefinition memory definition,
            State memory lastState,
            uint256 challengeExpiry,
            uint256 lockedFunds
        )
    {
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
            bytes32 channelId,
            EscrowStatus status,
            uint64 unlockAt,
            uint64 challengeExpiry,
            uint256 lockedAmount,
            State memory initState
        )
    {
        EscrowDepositMeta memory meta = _escrowDeposits[escrowId];
        channelId = meta.channelId;
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
            bytes32 channelId,
            EscrowStatus status,
            uint64 challengeExpiry,
            uint256 lockedAmount,
            State memory initState
        )
    {
        EscrowWithdrawalMeta memory meta = _escrowWithdrawals[escrowId];
        channelId = meta.channelId;
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

            // Skip already-finalized escrows so they don't block the queue
            if (meta.status == EscrowStatus.FINALIZED) {
                escrowHeadTemp++;
                continue;
            }
            // only still "INITIALIZED" escrows can be purged: "CHALLENGED" escrows require manual finalization
            if (meta.unlockAt <= block.timestamp && meta.status == EscrowStatus.INITIALIZED) {
                _nodeBalances[meta.node][meta.initState.nonHomeState.token] += meta.lockedAmount;
                meta.status = EscrowStatus.FINALIZED;
                meta.lockedAmount = 0;
                purgedCount++;
                escrowHeadTemp++;
                continue;
            } else {
                break;
            }
        }

        escrowHead = escrowHeadTemp;

        if (purgedCount != 0) {
            emit EscrowDepositsPurged(purgedCount);
        }
    }

    // ========== Channel lifecycle ==========

    // Create channel with DEPOSIT, WITHDRAW, or OPERATE intent
    // This enables users who already have off-chain virtual states with non-zero version
    // to create a channel and perform initial operation simultaneously
    function createChannel(ChannelDefinition calldata def, State calldata initState) external payable {
        require(
            initState.intent == StateIntent.DEPOSIT || initState.intent == StateIntent.WITHDRAW
                || initState.intent == StateIntent.OPERATE,
            "invalid state intent for channel creation"
        );

        bytes32 channelId = Utils.getChannelId(def, VERSION);

        _requireValidDefinition(def);

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[def.node][initState.homeState.token]);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, initState);

        _validateSignatures(channelId, def, initState);

        _applyEffects(channelId, def, initState, effects);
        _userChannels[def.user].add(channelId);

        // Emit appropriate event based on intent
        if (initState.intent == StateIntent.DEPOSIT) {
            emit ChannelDeposited(channelId, initState);
        } else if (initState.intent == StateIntent.WITHDRAW) {
            emit ChannelWithdrawn(channelId, initState);
        } else {
            emit ChannelCheckpointed(channelId, initState);
        }

        emit ChannelCreated(channelId, def.user, def.node, def, initState);
    }

    function depositToChannel(bytes32 channelId, State calldata candidate) public payable {
        require(candidate.intent == StateIntent.DEPOSIT, "invalid state intent");

        ChannelMeta storage meta = _channels[channelId];

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[meta.definition.node][candidate.homeState.token]);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        _validateSignatures(channelId, meta.definition, candidate);

        _applyEffects(channelId, meta.definition, candidate, effects);

        emit ChannelDeposited(channelId, candidate);
    }

    function withdrawFromChannel(bytes32 channelId, State calldata candidate) public payable {
        require(candidate.intent == StateIntent.WITHDRAW, "invalid state intent");

        ChannelMeta storage meta = _channels[channelId];

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[meta.definition.node][candidate.homeState.token]);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        _validateSignatures(channelId, meta.definition, candidate);

        _applyEffects(channelId, meta.definition, candidate, effects);

        emit ChannelWithdrawn(channelId, candidate);
    }

    function checkpointChannel(bytes32 channelId, State calldata candidate) external payable {
        require(candidate.intent == StateIntent.OPERATE, "can only checkpoint operate states");

        ChannelMeta storage meta = _channels[channelId];

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[meta.definition.node][candidate.homeState.token]);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        _validateSignatures(channelId, meta.definition, candidate);

        _applyEffects(channelId, meta.definition, candidate, effects);

        emit ChannelCheckpointed(channelId, candidate);
    }

    function challengeChannel(
        bytes32 channelId,
        State calldata candidate,
        bytes calldata challengerSig
    ) external payable {
        ChannelMeta storage meta = _channels[channelId];

        require(meta.status == ChannelStatus.OPERATING, "invalid channel status");

        State memory prevState = meta.lastState;
        require(candidate.version >= prevState.version, "challenge candidate must have higher or equal version");

        address user = meta.definition.user;
        address node = meta.definition.node;

        // If version is higher, process the new state
        if (candidate.version > prevState.version) {
            require(candidate.intent == StateIntent.OPERATE, "invalid intent");

            ChannelEngine.TransitionContext memory ctx =
                _buildChannelContext(channelId, _nodeBalances[meta.definition.node][candidate.homeState.token]);
            ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

            _validateSignatures(channelId, meta.definition, candidate);

            _applyTransitionEffects(channelId, meta.definition, candidate, effects);
        }
        // else: challenging with same version, state already processed

        _validateChallengerSignature(channelId, meta.definition, candidate, challengerSig, user, node);

        meta.status = ChannelStatus.DISPUTED;
        uint64 challengeExpiry = uint64(block.timestamp) + meta.definition.challengeDuration;
        meta.challengeExpireAt = challengeExpiry;

        emit ChannelChallenged(channelId, candidate, challengeExpiry);
    }

    function closeChannel(bytes32 channelId, State calldata candidate) external payable {
        require(candidate.intent == StateIntent.CLOSE, "invalid state intent");

        ChannelMeta storage meta = _channels[channelId];
        ChannelStatus status = meta.status;

        require(status == ChannelStatus.OPERATING || status == ChannelStatus.DISPUTED, "invalid channel status");

        State memory prevState = meta.lastState;
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
        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[meta.definition.node][candidate.homeState.token]);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        _validateSignatures(channelId, meta.definition, candidate);

        _applyEffects(channelId, meta.definition, candidate, effects);
        _userChannels[user].remove(channelId);

        emit ChannelClosed(channelId, candidate);
    }

    // ========= Cross-Chain Functions ==========

    function initiateEscrowDeposit(ChannelDefinition calldata def, State calldata candidate) external payable {
        require(candidate.intent == StateIntent.INITIATE_ESCROW_DEPOSIT, "invalid intent");

        bytes32 channelId = Utils.getChannelId(def, VERSION);

        _validateSignatures(channelId, def, candidate);

        bytes32 escrowId = Utils.getEscrowId(channelId, candidate.version);

        if (_isHomeChain(channelId)) {
            // HOME CHAIN: Update channel via ChannelEngine
            ChannelMeta storage meta = _channels[channelId];

            ChannelEngine.TransitionContext memory ctx =
                _buildChannelContext(channelId, _nodeBalances[meta.definition.node][candidate.homeState.token]);
            ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

            _applyEffects(channelId, meta.definition, candidate, effects);

            emit EscrowDepositInitiatedOnHome(escrowId, channelId, candidate);
        } else {
            // NON-HOME CHAIN: Create escrow record - recover addresses from signatures
            EscrowDepositEngine.TransitionContext memory ctx = _buildEscrowDepositContext(escrowId, 0);
            EscrowDepositEngine.TransitionEffects memory effects =
                EscrowDepositEngine.validateTransition(ctx, candidate);

            _applyEscrowDepositEffects(escrowId, channelId, candidate, effects, def.user, def.node, def.signatureValidator);
            _escrowDepositIds.push(escrowId);

            emit EscrowDepositInitiated(escrowId, channelId, candidate);
        }
    }

    function challengeEscrowDeposit(bytes32 escrowId, bytes calldata challengerSig) external {
        EscrowDepositMeta storage meta = _escrowDeposits[escrowId];
        require(!_isHomeChain(meta.channelId), "only non-home escrows can be challenged");

        EscrowDepositEngine.TransitionContext memory ctx = _buildEscrowDepositContext(escrowId, 0);
        EscrowDepositEngine.TransitionEffects memory effects = EscrowDepositEngine.validateChallenge(ctx);

        bytes32 channelId = meta.channelId;
        ChannelDefinition memory def = _channels[channelId].definition;
        _validateChallengerSignature(channelId, def, meta.initState, challengerSig, meta.user, meta.node);

        _applyEscrowDepositEffects(escrowId, channelId, meta.initState, effects, meta.user, meta.node, meta.signatureValidator);

        emit EscrowDepositChallenged(escrowId, meta.initState, effects.newChallengeExpiry);
    }

    function finalizeEscrowDeposit(bytes32 escrowId, State calldata candidate) external {
        EscrowDepositMeta storage meta = _escrowDeposits[escrowId];
        address user = meta.user;
        address node = meta.node;

        bool isHomeChain = _isHomeChain(meta.channelId);
        EscrowStatus status = meta.status;

        if (!isHomeChain && status == EscrowStatus.DISPUTED && block.timestamp > meta.challengeExpireAt) {
            // NON-HOME CHAIN: Unilateral finalization after challenge timeout
            meta.status = EscrowStatus.FINALIZED;
            uint256 lockedAmount = meta.lockedAmount;
            meta.lockedAmount = 0;
            meta.challengeExpireAt = 0;

            _pushFunds(node, meta.initState.nonHomeState.token, lockedAmount);

            emit EscrowDepositFinalized(escrowId, meta.channelId, candidate);
            return;
        }

        require(candidate.intent == StateIntent.FINALIZE_ESCROW_DEPOSIT, "invalid intent");

        // Construct minimal ChannelDefinition for signature validation
        ChannelDefinition memory def = ChannelDefinition({
            challengeDuration: 0, // not used in validation
            user: user,
            node: node,
            nonce: 0, // not used in validation
            signatureValidator: meta.signatureValidator,
            metadata: bytes32(0) // not used in validation
        });
        _validateSignatures(meta.channelId, def, candidate);

        if (isHomeChain) {
            // HOME CHAIN: Update channel via ChannelEngine
            ChannelMeta storage channelMeta = _channels[meta.channelId];
            ChannelEngine.TransitionContext memory ctx = _buildChannelContext(
                meta.channelId, _nodeBalances[channelMeta.definition.node][candidate.homeState.token]
            );
            ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

            _applyEffects(meta.channelId, channelMeta.definition, candidate, effects);

            emit EscrowDepositFinalizedOnHome(escrowId, meta.channelId, candidate);
            return;
        } else {
            // NON-HOME CHAIN: Update via EscrowDepositEngine
            EscrowDepositEngine.TransitionContext memory ctx =
                _buildEscrowDepositContext(escrowId, _nodeBalances[node][candidate.nonHomeState.token]);
            EscrowDepositEngine.TransitionEffects memory effects =
                EscrowDepositEngine.validateTransition(ctx, candidate);

            _applyEscrowDepositEffects(escrowId, meta.channelId, candidate, effects, user, node, meta.signatureValidator);

            emit EscrowDepositFinalized(escrowId, meta.channelId, candidate);
        }
    }

    function initiateEscrowWithdrawal(ChannelDefinition calldata def, State calldata candidate) external {
        require(candidate.intent == StateIntent.INITIATE_ESCROW_WITHDRAWAL, "invalid intent");

        bytes32 channelId = Utils.getChannelId(def, VERSION);
        _validateSignatures(channelId, def, candidate);

        bytes32 escrowId = Utils.getEscrowId(channelId, candidate.version);

        if (_isHomeChain(channelId)) {
            // HOME CHAIN
            ChannelMeta storage meta = _channels[channelId];

            ChannelEngine.TransitionContext memory ctx =
                _buildChannelContext(channelId, _nodeBalances[meta.definition.node][candidate.homeState.token]);
            ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

            _applyEffects(channelId, meta.definition, candidate, effects);

            emit EscrowWithdrawalInitiatedOnHome(escrowId, channelId, candidate);
        } else {
            // NON-HOME CHAIN
            EscrowWithdrawalEngine.TransitionContext memory ctx = _buildEscrowWithdrawalContext(escrowId, def.node);
            EscrowWithdrawalEngine.TransitionEffects memory effects =
                EscrowWithdrawalEngine.validateTransition(ctx, candidate);

            _applyEscrowWithdrawalEffects(escrowId, channelId, candidate, effects, def.user, def.node, def.signatureValidator);

            emit EscrowWithdrawalInitiated(escrowId, channelId, candidate);
        }
    }

    function challengeEscrowWithdrawal(bytes32 escrowId, bytes calldata challengerSig) external {
        EscrowWithdrawalMeta storage meta = _escrowWithdrawals[escrowId];
        require(!_isHomeChain(meta.channelId), "only non-home escrows can be challenged");

        EscrowWithdrawalEngine.TransitionContext memory ctx = _buildEscrowWithdrawalContext(escrowId, meta.node);
        EscrowWithdrawalEngine.TransitionEffects memory effects = EscrowWithdrawalEngine.validateChallenge(ctx);

        // Validate challenger signature
        bytes32 channelId = meta.channelId;
        address user = meta.user;
        address node = meta.node;
        ChannelDefinition memory def = _channels[channelId].definition;
        _validateChallengerSignature(channelId, def, meta.initState, challengerSig, user, node);

        _applyEscrowWithdrawalEffects(escrowId, channelId, meta.initState, effects, user, node, meta.signatureValidator);

        emit EscrowWithdrawalChallenged(escrowId, meta.initState, effects.newChallengeExpiry);
    }

    function finalizeEscrowWithdrawal(bytes32 escrowId, State calldata candidate) external {
        EscrowWithdrawalMeta storage meta = _escrowWithdrawals[escrowId];
        bytes32 channelId = meta.channelId;
        address user = meta.user;
        address node = meta.node;

        bool isHomeChain = _isHomeChain(channelId);
        EscrowStatus status = meta.status;

        if (!isHomeChain && status == EscrowStatus.DISPUTED && block.timestamp > meta.challengeExpireAt) {
            // NON-HOME CHAIN: Unilateral finalization after challenge timeout
            meta.status = EscrowStatus.FINALIZED;
            uint256 lockedAmount = meta.lockedAmount;
            meta.lockedAmount = 0;
            meta.challengeExpireAt = 0;

            _pushFunds(node, meta.initState.nonHomeState.token, lockedAmount);

            emit EscrowWithdrawalFinalized(escrowId, meta.channelId, candidate);
            return;
        }

        require(candidate.intent == StateIntent.FINALIZE_ESCROW_WITHDRAWAL, "invalid intent");

        // Construct minimal ChannelDefinition for signature validation
        ChannelDefinition memory def = ChannelDefinition({
            challengeDuration: 0, // not used in validation
            user: user,
            node: node,
            nonce: 0, // not used in validation
            signatureValidator: meta.signatureValidator,
            metadata: bytes32(0) // not used in validation
        });
        _validateSignatures(channelId, def, candidate);

        if (isHomeChain) {
            // HOME CHAIN: Update channel via ChannelEngine
            ChannelMeta storage channelMeta = _channels[channelId];
            ChannelEngine.TransitionContext memory ctx =
                _buildChannelContext(channelId, _nodeBalances[channelMeta.definition.node][candidate.homeState.token]);
            ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

            _applyEffects(channelId, channelMeta.definition, candidate, effects);

            emit EscrowWithdrawalFinalizedOnHome(escrowId, channelId, candidate);
        } else {
            // Non-Home chain: Update via EscrowWithdrawalEngine
            EscrowWithdrawalEngine.TransitionContext memory ctx = _buildEscrowWithdrawalContext(escrowId, node);
            EscrowWithdrawalEngine.TransitionEffects memory effects =
                EscrowWithdrawalEngine.validateTransition(ctx, candidate);

            _applyEscrowWithdrawalEffects(escrowId, channelId, candidate, effects, user, node, meta.signatureValidator);

            emit EscrowWithdrawalFinalized(escrowId, channelId, candidate);
        }
    }

    function initiateMigration(ChannelDefinition calldata def, State calldata candidate) external {
        require(candidate.intent == StateIntent.INITIATE_MIGRATION, "invalid intent");

        bytes32 channelId = Utils.getChannelId(def, VERSION);
        _validateSignatures(channelId, def, candidate);

        State memory targetCandidate = candidate;
        bool isHomeChain = _isHomeChain(channelId);

        if (!isHomeChain) {
            // Initiate migration IN (on new home chain)
            _requireValidDefinition(def);

            // Swap states before processing it, so that homeState = current chain
            targetCandidate.homeState = candidate.nonHomeState;
            targetCandidate.nonHomeState = candidate.homeState;
            targetCandidate.userSig = ""; // Invalidate signatures after swap
            targetCandidate.nodeSig = "";

            _userChannels[def.user].add(channelId);
        }

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[def.node][targetCandidate.homeState.token]);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, targetCandidate);

        _applyEffects(channelId, def, targetCandidate, effects);

        // event with the correct candidate state
        if (isHomeChain) {
            emit MigrationOutInitiated(channelId, candidate);
        } else {
            emit MigrationInInitiated(channelId, candidate);
        }
    }

    function finalizeMigration(bytes32 channelId, State calldata candidate) external {
        require(candidate.intent == StateIntent.FINALIZE_MIGRATION, "invalid intent");

        ChannelMeta storage meta = _channels[channelId];
        address user = meta.definition.user;

        _validateSignatures(channelId, meta.definition, candidate);

        State memory targetCandidate = candidate;
        // `_isHomeChain` cannot be used here as channel exists on both chains
        bool isHomeChain = candidate.nonHomeState.chainId == block.chainid;

        if (isHomeChain) {
            // Finalize migration OUT (on old home chain)
            // Swap states before validation to maintain invariant, so that homeState = current chain
            targetCandidate.homeState = candidate.nonHomeState;
            targetCandidate.nonHomeState = candidate.homeState;
            targetCandidate.userSig = ""; // Invalidate signatures after swap
            targetCandidate.nodeSig = "";

            _userChannels[user].remove(channelId);
        }

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[meta.definition.node][targetCandidate.homeState.token]);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, targetCandidate);

        _applyEffects(channelId, meta.definition, targetCandidate, effects);

        // event with the correct candidate state
        if (isHomeChain) {
            emit MigrationOutFinalized(channelId, candidate);
        } else {
            emit MigrationInFinalized(channelId, candidate);
        }
    }

    // ========= Internal ==========

    /**
     * @notice Validates signatures using the pluggable signature validator system
     * @dev The first byte of each signature determines which validator to use:
     *      - 0x00 (DEFAULT) -> use defaultSigValidator
     *      - 0x01 (CHANNEL) -> use channel's signatureValidator from definition
     * @param channelId The channel ID for signature validation
     * @param def The channel definition containing the channel-specific validator
     * @param state The state to validate (must include userSig and nodeSig)
     */
    function _validateSignatures(bytes32 channelId, ChannelDefinition memory def, State calldata state) internal view {
        require(state.userSig.length > 0, "empty user signature");
        require(state.nodeSig.length > 0, "empty node signature");

        bytes memory signingData = Utils.toSigningData(state);

        // Validate user signature
        SigValidatorType userValidatorType = SigValidatorType(uint8(state.userSig[0]));
        bytes memory userSigData = state.userSig[1:];
        ISignatureValidator userValidator = userValidatorType == SigValidatorType.DEFAULT
            ? DEFAULT_SIG_VALIDATOR
            : ISignatureValidator(def.signatureValidator);

        ValidationResult userResult = userValidator.validateSignature(channelId, signingData, userSigData, def.user);
        require(ValidationResult.unwrap(userResult) != ValidationResult.unwrap(VALIDATION_FAILURE), "invalid user signature");

        // Validate node signature
        SigValidatorType nodeValidatorType = SigValidatorType(uint8(state.nodeSig[0]));
        bytes memory nodeSigData = state.nodeSig[1:];
        ISignatureValidator nodeValidator = nodeValidatorType == SigValidatorType.DEFAULT
            ? DEFAULT_SIG_VALIDATOR
            : ISignatureValidator(def.signatureValidator);

        ValidationResult nodeResult = nodeValidator.validateSignature(channelId, signingData, nodeSigData, def.node);
        require(ValidationResult.unwrap(nodeResult) != ValidationResult.unwrap(VALIDATION_FAILURE), "invalid node signature");
    }

    function _validateChallengerSignature(
        bytes32 channelId,
        ChannelDefinition memory def,
        State memory state,
        bytes calldata challengerSig,
        address user,
        address node
    ) internal view {
        require(challengerSig.length > 0, "empty challenger signature");

        bytes memory signingData = Utils.toSigningData(state);

        // Extract validator type and signature data
        SigValidatorType validatorType = SigValidatorType(uint8(challengerSig[0]));
        bytes memory sigData = challengerSig[1:];
        ISignatureValidator validator = validatorType == SigValidatorType.DEFAULT
            ? DEFAULT_SIG_VALIDATOR
            : ISignatureValidator(def.signatureValidator);

        ValidationResult result = validator.validateChallengerSignature(channelId, signingData, sigData, user, node);
        require(ValidationResult.unwrap(result) != ValidationResult.unwrap(VALIDATION_FAILURE), "invalid challenger signature");
    }

    function _buildChannelContext(bytes32 channelId, uint256 nodeBalance)
        internal
        view
        returns (ChannelEngine.TransitionContext memory ctx)
    {
        ChannelMeta storage meta = _channels[channelId];

        ctx.status = meta.status;
        ctx.prevState = meta.lastState;
        ctx.lockedFunds = meta.lockedFunds;
        ctx.nodeAvailableFunds = nodeBalance;
        ctx.challengeExpiry = meta.challengeExpireAt;

        return ctx;
    }

    function _buildEscrowDepositContext(bytes32 escrowId, uint256 nodeAvailableFunds)
        internal
        view
        returns (EscrowDepositEngine.TransitionContext memory ctx)
    {
        EscrowDepositMeta storage meta = _escrowDeposits[escrowId];

        ctx.status = meta.status;
        ctx.initState = meta.initState;
        ctx.lockedAmount = meta.lockedAmount;
        ctx.unlockAt = meta.unlockAt;
        ctx.challengeExpiry = meta.challengeExpireAt;
        ctx.nodeAvailableFunds = nodeAvailableFunds;

        return ctx;
    }

    function _buildEscrowWithdrawalContext(bytes32 escrowId, address node)
        internal
        view
        returns (EscrowWithdrawalEngine.TransitionContext memory ctx)
    {
        EscrowWithdrawalMeta storage meta = _escrowWithdrawals[escrowId];

        ctx.status = meta.status;
        ctx.initState = meta.initState;
        ctx.lockedAmount = meta.lockedAmount;
        ctx.challengeExpiry = meta.challengeExpireAt;
        ctx.nodeAddress = node;

        return ctx;
    }

    function _isHomeChain(bytes32 channelId) internal view returns (bool) {
        ChannelStatus status = _channels[channelId].status;
        if (status == ChannelStatus.VOID || status == ChannelStatus.MIGRATED_OUT) {
            return false;
        }

        return _channels[channelId].lastState.homeState.chainId == block.chainid;
    }

    function _applyEffects(
        bytes32 channelId,
        ChannelDefinition memory def,
        State memory candidate,
        ChannelEngine.TransitionEffects memory effects
    ) internal {
        ChannelMeta storage meta = _channels[channelId];

        if (meta.status == ChannelStatus.VOID) {
            meta.definition = def;
        }

        _applyTransitionEffects(channelId, def, candidate, effects);

        if (effects.newStatus != ChannelStatus.VOID && meta.status != effects.newStatus) {
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
        ChannelDefinition memory def,
        State memory candidate,
        ChannelEngine.TransitionEffects memory effects
    ) internal {
        ChannelMeta storage meta = _channels[channelId];

        if (effects.updateLastState) {
            meta.lastState = candidate;
        }

        address token = candidate.homeState.token;

        // Process POSITIVE deltas first (additions to lockedFunds) to prevent underflow
        if (effects.userFundsDelta > 0) {
            uint256 amount = effects.userFundsDelta.toUint256();
            _pullFunds(def.user, token, amount);
            meta.lockedFunds += amount;
        }

        if (effects.nodeFundsDelta > 0) {
            uint256 amount = effects.nodeFundsDelta.toUint256();
            _nodeBalances[def.node][token] -= amount;
            meta.lockedFunds += amount;
        }

        // Then process NEGATIVE deltas (subtractions from lockedFunds)
        if (effects.userFundsDelta < 0) {
            uint256 amount = (-effects.userFundsDelta).toUint256();
            _pushFunds(def.user, token, amount);
            meta.lockedFunds -= amount;
        }

        if (effects.nodeFundsDelta < 0) {
            uint256 amount = (-effects.nodeFundsDelta).toUint256();
            _nodeBalances[def.node][token] += amount;
            meta.lockedFunds -= amount;
        }

        // Special handling for CLOSE: push nodeAllocation directly to node address
        if (effects.closeChannel && candidate.homeState.nodeAllocation > 0) {
            _pushFunds(def.node, token, candidate.homeState.nodeAllocation);
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

    function _applyEscrowDepositEffects(
        bytes32 escrowId,
        bytes32 channelId,
        State memory candidate,
        EscrowDepositEngine.TransitionEffects memory effects,
        address user,
        address node,
        address signatureValidator
    ) internal {
        EscrowDepositMeta storage meta = _escrowDeposits[escrowId];

        if (effects.newStatus != EscrowStatus.VOID) {
            meta.status = effects.newStatus;
        }

        if (effects.updateInitState) {
            meta.initState = candidate;
            meta.channelId = channelId;
            meta.user = user;
            meta.node = node;
            meta.signatureValidator = signatureValidator;
        }

        if (effects.newUnlockAt > 0) {
            meta.unlockAt = effects.newUnlockAt;
        }

        if (effects.newChallengeExpiry > 0) {
            meta.challengeExpireAt = effects.newChallengeExpiry;
        }

        // Determine the correct token to use (from init state for finalization, from candidate for initiation)
        address token = effects.updateInitState ? candidate.nonHomeState.token : meta.initState.nonHomeState.token;

        // Handle user funds (positive = pull from user)
        if (effects.userFundsDelta > 0) {
            uint256 amount = effects.userFundsDelta.toUint256();
            _pullFunds(user, token, amount);
            meta.lockedAmount += amount;
        } else if (effects.userFundsDelta < 0) {
            uint256 amount = (-effects.userFundsDelta).toUint256();
            _pushFunds(user, token, amount);
            meta.lockedAmount -= amount;
        }

        // Handle node funds (positive = pull from node vault, negative = release to vault)
        if (effects.nodeFundsDelta > 0) {
            uint256 amount = effects.nodeFundsDelta.toUint256();
            _nodeBalances[node][token] -= amount;
            meta.lockedAmount += amount;
        } else if (effects.nodeFundsDelta < 0) {
            uint256 amount = (-effects.nodeFundsDelta).toUint256();
            _nodeBalances[node][token] += amount;
            meta.lockedAmount -= amount;
        }

        // NOTE: purge escrow deposits to unlock unutilized node liquidity
        _purgeEscrowDeposits();
    }

    function _applyEscrowWithdrawalEffects(
        bytes32 escrowId,
        bytes32 channelId,
        State memory candidate,
        EscrowWithdrawalEngine.TransitionEffects memory effects,
        address user,
        address node,
        address signatureValidator
    ) internal {
        EscrowWithdrawalMeta storage meta = _escrowWithdrawals[escrowId];

        if (effects.newStatus != EscrowStatus.VOID) {
            meta.status = effects.newStatus;
        }

        if (effects.updateInitState) {
            meta.initState = candidate;
            meta.channelId = channelId;
            meta.user = user;
            meta.node = node;
            meta.signatureValidator = signatureValidator;
        }

        if (effects.newChallengeExpiry > 0) {
            meta.challengeExpireAt = effects.newChallengeExpiry;
        }

        // Determine the correct token to use (from init state for finalization, from candidate for initiation)
        address token = effects.updateInitState ? candidate.nonHomeState.token : meta.initState.nonHomeState.token;

        // Handle user funds (negative = push to user)
        if (effects.userFundsDelta > 0) {
            uint256 amount = effects.userFundsDelta.toUint256();
            _pullFunds(user, token, amount);
            meta.lockedAmount += amount;
        } else if (effects.userFundsDelta < 0) {
            uint256 amount = (-effects.userFundsDelta).toUint256();
            _pushFunds(user, token, amount);
            meta.lockedAmount -= amount;
        }

        // Handle node funds (positive = pull from node vault, negative = release to vault)
        if (effects.nodeFundsDelta > 0) {
            uint256 amount = effects.nodeFundsDelta.toUint256();
            _nodeBalances[node][token] -= amount;
            meta.lockedAmount += amount;
        } else if (effects.nodeFundsDelta < 0) {
            uint256 amount = (-effects.nodeFundsDelta).toUint256();
            _nodeBalances[node][token] += amount;
            meta.lockedAmount -= amount;
        }

        // NOTE: purge escrow deposits to unlock unutilized node liquidity
        _purgeEscrowDeposits();
    }

    function _requireValidDefinition(ChannelDefinition calldata def) internal pure {
        require(def.user != address(0), InvalidAddress());
        require(def.node != address(0), InvalidAddress());
        require(def.user != def.node, AddressCollision(def.user));
        require(def.challengeDuration >= MIN_CHALLENGE_DURATION, IncorrectChallengeDuration());
        require(def.signatureValidator != address(0), InvalidAddress());
    }
}
