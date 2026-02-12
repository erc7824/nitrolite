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
import {
    ChannelDefinition,
    ChannelStatus,
    DEFAULT_SIG_VALIDATOR_ID,
    EscrowStatus,
    State,
    Ledger,
    StateIntent,
    ParticipantIndex
} from "./interfaces/Types.sol";

import {Utils} from "./Utils.sol";
import {ChannelEngine} from "./ChannelEngine.sol";
import {EscrowDepositEngine} from "./EscrowDepositEngine.sol";
import {EscrowWithdrawalEngine} from "./EscrowWithdrawalEngine.sol";
import {EcdsaSignatureUtils} from "./sigValidators/EcdsaSignatureUtils.sol";

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

    event ValidatorRegistered(address indexed node, uint8 indexed validatorId, ISignatureValidator indexed validator);

    error InvalidAddress();
    error InvalidAmount();
    error InvalidValue();
    error AddressCollision(address collision);
    error IncorrectChallengeDuration();
    error ChannelDoesNotExist(bytes32 channelId);
    error InvalidValidatorId();
    error ValidatorAlreadyRegistered(address node, uint8 validatorId);
    error ValidatorNotRegistered(address node, uint8 validatorId);
    error InvalidSignature();

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
        uint256 approvedSignatureValidators; // Bitmask of approved validator IDs for user signatures
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
        uint256 approvedSignatureValidators; // Bitmask of approved validator IDs for user signatures
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

    // Validator ID 0x00 is reserved for DEFAULT_SIG_VALIDATOR
    // Validator IDs 0x01-0xFF are available for node-registered validators
    mapping(address node => mapping(uint8 validatorId => ISignatureValidator validator)) internal
        _nodeValidatorRegistry;

    // ========== Constructor ==========

    constructor(ISignatureValidator _defaultSigValidator) {
        require(address(_defaultSigValidator) != address(0), InvalidAddress());
        DEFAULT_SIG_VALIDATOR = _defaultSigValidator;
    }

    // ========== Getters ==========

    function getAccountBalance(address node, address token) external view returns (uint256) {
        return _nodeBalances[node][token];
    }

    function getNodeValidator(address node, uint8 validatorId) external view returns (ISignatureValidator) {
        return _nodeValidatorRegistry[node][validatorId];
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

    // ========= Validator Registry ==========

    /**
     * @notice Register a signature validator for a node using signature-based authorization
     * @dev Anyone can submit this transaction with a valid node signature, enabling relayer-friendly registration.
     *      The node's private key only signs the registration data, never sends transactions directly.
     *      This allows nodes to use cold storage or HSMs without exposing keys to transaction submission.
     *      The signature includes block.chainid to prevent cross-chain replay attacks.
     * @param node The node address that signed the registration
     * @param validatorId The validator ID (0x01-0xFF, 0x00 reserved for DEFAULT)
     * @param validator The validator contract address
     * @param signature Node's signature over abi.encode(validatorId, validator, block.chainid)
     */
    function registerNodeValidator(
        address node,
        uint8 validatorId,
        ISignatureValidator validator,
        bytes calldata signature
    ) external {
        require(validatorId != DEFAULT_SIG_VALIDATOR_ID, InvalidValidatorId());
        require(address(validator) != address(0), InvalidAddress());
        require(
            address(_nodeValidatorRegistry[node][validatorId]) == address(0),
            ValidatorAlreadyRegistered(node, validatorId)
        );

        // Verify node signed the registration data (includes chainid to prevent cross-chain replay)
        bytes memory message = abi.encode(validatorId, validator, block.chainid);
        require(EcdsaSignatureUtils.validateEcdsaSigner(message, signature, node), InvalidSignature());

        _nodeValidatorRegistry[node][validatorId] = validator;

        emit ValidatorRegistered(node, validatorId, validator);
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
                _nodeBalances[meta.node][meta.initState.nonHomeLedger.token] += meta.lockedAmount;
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
        _validateSignatures(channelId, initState, def.user, def.node, def.approvedSignatureValidators);

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[def.node][initState.homeLedger.token]);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, initState);

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
        _validateSignatures(
            channelId,
            candidate,
            meta.definition.user,
            meta.definition.node,
            meta.definition.approvedSignatureValidators
        );

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[meta.definition.node][candidate.homeLedger.token]);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        _applyEffects(channelId, meta.definition, candidate, effects);

        emit ChannelDeposited(channelId, candidate);
    }

    function withdrawFromChannel(bytes32 channelId, State calldata candidate) public payable {
        require(candidate.intent == StateIntent.WITHDRAW, "invalid state intent");

        ChannelMeta storage meta = _channels[channelId];
        _validateSignatures(
            channelId,
            candidate,
            meta.definition.user,
            meta.definition.node,
            meta.definition.approvedSignatureValidators
        );

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[meta.definition.node][candidate.homeLedger.token]);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        _applyEffects(channelId, meta.definition, candidate, effects);

        emit ChannelWithdrawn(channelId, candidate);
    }

    function checkpointChannel(bytes32 channelId, State calldata candidate) external payable {
        require(candidate.intent == StateIntent.OPERATE, "can only checkpoint operate states");

        ChannelMeta storage meta = _channels[channelId];
        _validateSignatures(
            channelId,
            candidate,
            meta.definition.user,
            meta.definition.node,
            meta.definition.approvedSignatureValidators
        );

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[meta.definition.node][candidate.homeLedger.token]);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        _applyEffects(channelId, meta.definition, candidate, effects);

        emit ChannelCheckpointed(channelId, candidate);
    }

    function challengeChannel(
        bytes32 channelId,
        State calldata candidate,
        bytes calldata challengerSig,
        ParticipantIndex challengerIdx
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
            _validateSignatures(channelId, candidate, user, node, meta.definition.approvedSignatureValidators);

            ChannelEngine.TransitionContext memory ctx =
                _buildChannelContext(channelId, _nodeBalances[meta.definition.node][candidate.homeLedger.token]);
            ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

            _applyTransitionEffects(channelId, meta.definition, candidate, effects);
        }
        // else: challenging with same version, state already processed

        (ISignatureValidator validator, bytes calldata sigData) =
            _extractValidator(challengerSig, node, meta.definition.approvedSignatureValidators);
        _validateChallengerSignature(channelId, candidate, sigData, validator, user, node, challengerIdx);

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

            _pushFunds(user, prevState.homeLedger.token, prevState.homeLedger.userAllocation);
            _pushFunds(node, prevState.homeLedger.token, prevState.homeLedger.nodeAllocation);

            _userChannels[user].remove(channelId);

            emit ChannelClosed(channelId, prevState);
            return;
        }

        // Path 2: Cooperative closure with signed CLOSE state
        _validateSignatures(channelId, candidate, user, node, meta.definition.approvedSignatureValidators);

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[meta.definition.node][candidate.homeLedger.token]);
        ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

        _applyEffects(channelId, meta.definition, candidate, effects);
        _userChannels[user].remove(channelId);

        emit ChannelClosed(channelId, candidate);
    }

    // ========= Cross-Chain Functions ==========

    function initiateEscrowDeposit(ChannelDefinition calldata def, State calldata candidate) external payable {
        require(candidate.intent == StateIntent.INITIATE_ESCROW_DEPOSIT, "invalid intent");

        bytes32 channelId = Utils.getChannelId(def, VERSION);
        _validateSignatures(channelId, candidate, def.user, def.node, def.approvedSignatureValidators);

        bytes32 escrowId = Utils.getEscrowId(channelId, candidate.version);

        if (_isHomeChain(channelId)) {
            // HOME CHAIN: Update channel via ChannelEngine
            ChannelMeta storage meta = _channels[channelId];

            ChannelEngine.TransitionContext memory ctx =
                _buildChannelContext(channelId, _nodeBalances[meta.definition.node][candidate.homeLedger.token]);
            ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

            _applyEffects(channelId, meta.definition, candidate, effects);

            emit EscrowDepositInitiatedOnHome(escrowId, channelId, candidate);
        } else {
            // NON-HOME CHAIN: Create escrow record - recover addresses from signatures
            EscrowDepositEngine.TransitionContext memory ctx = _buildEscrowDepositContext(escrowId, 0);
            EscrowDepositEngine.TransitionEffects memory effects =
                EscrowDepositEngine.validateTransition(ctx, candidate);

            _applyEscrowDepositEffects(
                escrowId, channelId, candidate, effects, def.user, def.node, def.approvedSignatureValidators
            );
            _escrowDepositIds.push(escrowId);

            emit EscrowDepositInitiated(escrowId, channelId, candidate);
        }
    }

    function challengeEscrowDeposit(bytes32 escrowId, bytes calldata challengerSig, ParticipantIndex challengerIdx)
        external
    {
        EscrowDepositMeta storage meta = _escrowDeposits[escrowId];
        require(!_isHomeChain(meta.channelId), "only non-home escrows can be challenged");

        bytes32 channelId = meta.channelId;
        (ISignatureValidator validator, bytes calldata sigData) =
            _extractValidator(challengerSig, meta.node, meta.approvedSignatureValidators);
        _validateChallengerSignature(channelId, meta.initState, sigData, validator, meta.user, meta.node, challengerIdx);

        EscrowDepositEngine.TransitionContext memory ctx = _buildEscrowDepositContext(escrowId, 0);
        EscrowDepositEngine.TransitionEffects memory effects = EscrowDepositEngine.validateChallenge(ctx);

        _applyEscrowDepositEffects(
            escrowId, channelId, meta.initState, effects, meta.user, meta.node, meta.approvedSignatureValidators
        );

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

            _pushFunds(node, meta.initState.nonHomeLedger.token, lockedAmount);

            emit EscrowDepositFinalized(escrowId, meta.channelId, candidate);
            return;
        }

        require(candidate.intent == StateIntent.FINALIZE_ESCROW_DEPOSIT, "invalid intent");

        if (isHomeChain) {
            // HOME CHAIN: Update channel via ChannelEngine
            ChannelMeta storage channelMeta = _channels[meta.channelId];
            _validateSignatures(
                meta.channelId, candidate, user, node, channelMeta.definition.approvedSignatureValidators
            );
            ChannelEngine.TransitionContext memory ctx = _buildChannelContext(
                meta.channelId, _nodeBalances[channelMeta.definition.node][candidate.homeLedger.token]
            );
            ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

            _applyEffects(meta.channelId, channelMeta.definition, candidate, effects);

            emit EscrowDepositFinalizedOnHome(escrowId, meta.channelId, candidate);
            return;
        } else {
            // NON-HOME CHAIN: Update via EscrowDepositEngine
            _validateSignatures(meta.channelId, candidate, user, node, meta.approvedSignatureValidators);

            EscrowDepositEngine.TransitionContext memory ctx =
                _buildEscrowDepositContext(escrowId, _nodeBalances[node][candidate.nonHomeLedger.token]);
            EscrowDepositEngine.TransitionEffects memory effects =
                EscrowDepositEngine.validateTransition(ctx, candidate);

            _applyEscrowDepositEffects(
                escrowId, meta.channelId, candidate, effects, user, node, meta.approvedSignatureValidators
            );

            emit EscrowDepositFinalized(escrowId, meta.channelId, candidate);
        }
    }

    function initiateEscrowWithdrawal(ChannelDefinition calldata def, State calldata candidate) external {
        require(candidate.intent == StateIntent.INITIATE_ESCROW_WITHDRAWAL, "invalid intent");

        bytes32 channelId = Utils.getChannelId(def, VERSION);
        _validateSignatures(channelId, candidate, def.user, def.node, def.approvedSignatureValidators);

        bytes32 escrowId = Utils.getEscrowId(channelId, candidate.version);

        if (_isHomeChain(channelId)) {
            // HOME CHAIN
            ChannelMeta storage meta = _channels[channelId];

            ChannelEngine.TransitionContext memory ctx =
                _buildChannelContext(channelId, _nodeBalances[meta.definition.node][candidate.homeLedger.token]);
            ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

            _applyEffects(channelId, meta.definition, candidate, effects);

            emit EscrowWithdrawalInitiatedOnHome(escrowId, channelId, candidate);
        } else {
            // NON-HOME CHAIN
            EscrowWithdrawalEngine.TransitionContext memory ctx = _buildEscrowWithdrawalContext(escrowId, def.node);
            EscrowWithdrawalEngine.TransitionEffects memory effects =
                EscrowWithdrawalEngine.validateTransition(ctx, candidate);

            _applyEscrowWithdrawalEffects(
                escrowId, channelId, candidate, effects, def.user, def.node, def.approvedSignatureValidators
            );

            emit EscrowWithdrawalInitiated(escrowId, channelId, candidate);
        }
    }

    function challengeEscrowWithdrawal(bytes32 escrowId, bytes calldata challengerSig, ParticipantIndex challengerIdx)
        external
    {
        EscrowWithdrawalMeta storage meta = _escrowWithdrawals[escrowId];
        require(!_isHomeChain(meta.channelId), "only non-home escrows can be challenged");

        EscrowWithdrawalEngine.TransitionContext memory ctx = _buildEscrowWithdrawalContext(escrowId, meta.node);
        EscrowWithdrawalEngine.TransitionEffects memory effects = EscrowWithdrawalEngine.validateChallenge(ctx);

        // Validate challenger signature
        bytes32 channelId = meta.channelId;
        address user = meta.user;
        address node = meta.node;
        (ISignatureValidator validator, bytes calldata sigData) =
            _extractValidator(challengerSig, node, meta.approvedSignatureValidators);
        _validateChallengerSignature(channelId, meta.initState, sigData, validator, user, node, challengerIdx);

        _applyEscrowWithdrawalEffects(
            escrowId, channelId, meta.initState, effects, user, node, meta.approvedSignatureValidators
        );

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

            _pushFunds(node, meta.initState.nonHomeLedger.token, lockedAmount);

            emit EscrowWithdrawalFinalized(escrowId, meta.channelId, candidate);
            return;
        }

        require(candidate.intent == StateIntent.FINALIZE_ESCROW_WITHDRAWAL, "invalid intent");

        if (isHomeChain) {
            // HOME CHAIN: Update channel via ChannelEngine
            ChannelMeta storage channelMeta = _channels[channelId];
            _validateSignatures(channelId, candidate, user, node, channelMeta.definition.approvedSignatureValidators);
            ChannelEngine.TransitionContext memory ctx =
                _buildChannelContext(channelId, _nodeBalances[channelMeta.definition.node][candidate.homeLedger.token]);
            ChannelEngine.TransitionEffects memory effects = ChannelEngine.validateTransition(ctx, candidate);

            _applyEffects(channelId, channelMeta.definition, candidate, effects);

            emit EscrowWithdrawalFinalizedOnHome(escrowId, channelId, candidate);
        } else {
            // Non-Home chain: Update via EscrowWithdrawalEngine
            _validateSignatures(channelId, candidate, user, node, meta.approvedSignatureValidators);

            EscrowWithdrawalEngine.TransitionContext memory ctx = _buildEscrowWithdrawalContext(escrowId, node);
            EscrowWithdrawalEngine.TransitionEffects memory effects =
                EscrowWithdrawalEngine.validateTransition(ctx, candidate);

            _applyEscrowWithdrawalEffects(
                escrowId, channelId, candidate, effects, user, node, meta.approvedSignatureValidators
            );

            emit EscrowWithdrawalFinalized(escrowId, channelId, candidate);
        }
    }

    function initiateMigration(ChannelDefinition calldata def, State calldata candidate) external {
        require(candidate.intent == StateIntent.INITIATE_MIGRATION, "invalid intent");

        bytes32 channelId = Utils.getChannelId(def, VERSION);
        _validateSignatures(channelId, candidate, def.user, def.node, def.approvedSignatureValidators);

        State memory targetCandidate = candidate;
        bool isHomeChain = _isHomeChain(channelId);

        if (!isHomeChain) {
            // Initiate migration IN (on new home chain)
            _requireValidDefinition(def);

            // Swap states before processing it, so that homeLedger = current chain
            targetCandidate.homeLedger = candidate.nonHomeLedger;
            targetCandidate.nonHomeLedger = candidate.homeLedger;
            targetCandidate.userSig = ""; // Invalidate signatures after swap
            targetCandidate.nodeSig = "";

            _userChannels[def.user].add(channelId);
        }

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[def.node][targetCandidate.homeLedger.token]);
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

        _validateSignatures(
            channelId, candidate, user, meta.definition.node, meta.definition.approvedSignatureValidators
        );

        State memory targetCandidate = candidate;
        // `_isHomeChain` cannot be used here as channel exists on both chains
        bool isHomeChain = candidate.nonHomeLedger.chainId == block.chainid;

        if (isHomeChain) {
            // Finalize migration OUT (on old home chain)
            // Swap states before validation to maintain invariant, so that homeLedger = current chain
            targetCandidate.homeLedger = candidate.nonHomeLedger;
            targetCandidate.nonHomeLedger = candidate.homeLedger;
            targetCandidate.userSig = ""; // Invalidate signatures after swap
            targetCandidate.nodeSig = "";

            _userChannels[user].remove(channelId);
        }

        ChannelEngine.TransitionContext memory ctx =
            _buildChannelContext(channelId, _nodeBalances[meta.definition.node][targetCandidate.homeLedger.token]);
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

    function _validateSignatures(
        bytes32 channelId,
        State calldata state,
        address user,
        address node,
        uint256 approvedSignatureValidators
    ) internal view {
        (ISignatureValidator userValidator, bytes calldata userSigData) =
            _extractValidator(state.userSig, node, approvedSignatureValidators);
        _validateSignature(channelId, state, userSigData, user, userValidator);
        (ISignatureValidator nodeValidator, bytes calldata nodeSigData) =
            _extractValidator(state.nodeSig, node, approvedSignatureValidators);
        _validateSignature(channelId, state, nodeSigData, node, nodeValidator);
    }

    /**
     * @notice Validates a single signature (stateless)
     * @param channelId The channel ID for signature validation
     * @param state The state to validate
     * @param sigData The signature data (without validator ID byte)
     * @param participant The expected signer's address
     * @param validator The validator to use for verification
     */
    function _validateSignature(
        bytes32 channelId,
        State calldata state,
        bytes calldata sigData,
        address participant,
        ISignatureValidator validator
    ) internal view {
        bytes memory signingData = Utils.toSigningData(state);
        ValidationResult result = validator.validateSignature(channelId, signingData, sigData, participant);
        require(ValidationResult.unwrap(result) != ValidationResult.unwrap(VALIDATION_FAILURE), "invalid signature");
    }

    function _extractValidator(bytes calldata signature, address node, uint256 approvedSignatureValidators)
        internal
        view
        returns (ISignatureValidator validator, bytes calldata sigData)
    {
        require(signature.length > 0, "empty signature");

        uint8 validatorId = uint8(signature[0]);

        if (validatorId == DEFAULT_SIG_VALIDATOR_ID) {
            // Default validator (0x00) is always available for both users and nodes
            validator = DEFAULT_SIG_VALIDATOR;
        } else {
            // Look up validator in node's registry
            require((approvedSignatureValidators >> validatorId) & 1 == 1, "Validator not approved");
            validator = _nodeValidatorRegistry[node][validatorId];
            require(address(validator) != address(0), ValidatorNotRegistered(node, validatorId));
        }

        sigData = _sliceCalldata(signature, 1);
    }

    function _sliceCalldata(bytes calldata data, uint256 start) internal pure returns (bytes calldata result) {
        assembly ("memory-safe") {
            result.offset := add(data.offset, start)
            result.length := sub(data.length, start)
        }
    }

    /**
     * @notice Validates a challenger's signature (stateless)
     * @param channelId The channel ID for signature validation
     * @param state The state being challenged
     * @param sigData The challenger's signature data (without validator ID byte)
     * @param validator The validator to use for verification
     * @param user The user's address
     * @param node The node's address
     */
    function _validateChallengerSignature(
        bytes32 channelId,
        State memory state,
        bytes calldata sigData,
        ISignatureValidator validator,
        address user,
        address node,
        ParticipantIndex challengerIdx
    ) internal view {
        bytes memory signingData = Utils.toSigningData(state);
        bytes memory challengerSigningData = abi.encodePacked(signingData, "challenge");
        address challenger = challengerIdx == ParticipantIndex.USER ? user : node;
        ValidationResult result = validator.validateSignature(channelId, challengerSigningData, sigData, challenger);
        require(
            ValidationResult.unwrap(result) != ValidationResult.unwrap(VALIDATION_FAILURE),
            "invalid challenger signature"
        );
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

        return _channels[channelId].lastState.homeLedger.chainId == block.chainid;
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

        address token = candidate.homeLedger.token;

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
        if (effects.closeChannel && candidate.homeLedger.nodeAllocation > 0) {
            _pushFunds(def.node, token, candidate.homeLedger.nodeAllocation);
            meta.lockedFunds -= candidate.homeLedger.nodeAllocation;
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
        uint256 approvedSignatureValidators
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
            meta.approvedSignatureValidators = approvedSignatureValidators;
        }

        if (effects.newUnlockAt > 0) {
            meta.unlockAt = effects.newUnlockAt;
        }

        if (effects.newChallengeExpiry > 0) {
            meta.challengeExpireAt = effects.newChallengeExpiry;
        }

        // Determine the correct token to use (from init state for finalization, from candidate for initiation)
        address token = effects.updateInitState ? candidate.nonHomeLedger.token : meta.initState.nonHomeLedger.token;

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
        uint256 approvedSignatureValidators
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
            meta.approvedSignatureValidators = approvedSignatureValidators;
        }

        if (effects.newChallengeExpiry > 0) {
            meta.challengeExpireAt = effects.newChallengeExpiry;
        }

        // Determine the correct token to use (from init state for finalization, from candidate for initiation)
        address token = effects.updateInitState ? candidate.nonHomeLedger.token : meta.initState.nonHomeLedger.token;

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
    }
}
