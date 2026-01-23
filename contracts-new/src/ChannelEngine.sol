// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {SafeCast} from "lib/openzeppelin-contracts/contracts/utils/math/SafeCast.sol";
import {ChannelStatus, State, StateIntent, Ledger} from "./interfaces/Types.sol";
import {Utils} from "./Utils.sol";

/**
 * @title ChannelEngine
 * @notice Unified validation and calculation engine for all channel state transitions
 * @dev REQUIRES `state.homeState` to ALWAYS point to this (execution) chain. Otherwise, delta calculations will be incorrect.
 */
library ChannelEngine {
    using SafeCast for int256;
    using SafeCast for uint256;
    using {Utils.isEmpty} for Ledger;

    // ========== Structs ==========

    struct TransitionContext {
        ChannelStatus status;
        State prevState;
        uint256 lockedFunds;
        uint256 nodeAvailableFunds;
        uint64 challengeExpiry;
    }

    struct TransitionEffects {
        // Fund movements (positive = pull/lock, negative = push/release)
        int256 userFundsDelta; // Funds to pull from user (>0) or push to user (<0)
        int256 nodeFundsDelta; // Funds to lock from node vault (>0) or release (<0)

        // State updates
        ChannelStatus newStatus;
        uint64 newChallengeExpiry;
        bool updateLastState;
        bool clearDispute;
        bool closeChannel;
    }

    // ========== Public Functions ==========

    /**
     * @notice Unified validation and calculation for all channel state transitions
     * @dev Three phases: universal validation → intent-specific calculation → universal invariants
     * @param ctx Current channel context from storage
     * @param candidate New state to transition to
     * @return effects The calculated effects to apply
     */
    function validateTransition(TransitionContext memory ctx, State memory candidate)
        external
        view
        returns (TransitionEffects memory effects)
    {
        // Phase 1: Universal validation
        _validateUniversal(ctx, candidate);

        // Phase 2: Intent-specific calculation
        effects = _calculateEffectsByIntent(ctx, candidate);

        // Phase 3: Universal invariants
        _validateInvariants(ctx, candidate, effects);

        return effects;
    }

    // ========== Internal: Phase 1 - Universal Validation ==========

    function _validateUniversal(TransitionContext memory ctx, State memory candidate) internal view {
        // homeState always represents current chain
        require(candidate.homeState.chainId == block.chainid, "invalid chain id");
        require(candidate.version > ctx.prevState.version || ctx.prevState.version == 0, "invalid version");

        // Cross-chain escrow and migration operations require nonHomeState
        if (
            candidate.intent == StateIntent.INITIATE_ESCROW_DEPOSIT
                || candidate.intent == StateIntent.FINALIZE_ESCROW_DEPOSIT
                || candidate.intent == StateIntent.INITIATE_ESCROW_WITHDRAWAL
                || candidate.intent == StateIntent.FINALIZE_ESCROW_WITHDRAWAL
                || candidate.intent == StateIntent.INITIATE_MIGRATION
                || candidate.intent == StateIntent.FINALIZE_MIGRATION
        ) {
            require(!candidate.nonHomeState.isEmpty(), "non-home state required for cross-chain operations");
            require(candidate.nonHomeState.chainId != block.chainid, "invalid non-home chain id");
        }

        uint256 allocsSum = candidate.homeState.userAllocation + candidate.homeState.nodeAllocation;
        int256 netFlowsSum = candidate.homeState.userNetFlow + candidate.homeState.nodeNetFlow;

        require(netFlowsSum >= 0, "negative net flow sum");
        require(allocsSum == uint256(netFlowsSum), "invalid allocation sum");
    }

    // ========== Internal: Phase 2 - Intent-Specific Calculation ==========

    function _calculateEffectsByIntent(TransitionContext memory ctx, State memory candidate)
        internal
        view
        returns (TransitionEffects memory effects)
    {
        int256 userNfDelta = candidate.homeState.userNetFlow - ctx.prevState.homeState.userNetFlow;
        int256 nodeNfDelta = candidate.homeState.nodeNetFlow - ctx.prevState.homeState.nodeNetFlow;

        StateIntent intent = candidate.intent;

        if (intent == StateIntent.CREATE) {
            effects = _calculateCreateEffects(ctx, candidate, userNfDelta);
        } else if (intent == StateIntent.DEPOSIT) {
            effects = _calculateDepositEffects(ctx, candidate, userNfDelta, nodeNfDelta);
        } else if (intent == StateIntent.WITHDRAW) {
            effects = _calculateWithdrawEffects(ctx, candidate, userNfDelta, nodeNfDelta);
        } else if (intent == StateIntent.OPERATE) {
            effects = _calculateOperateEffects(ctx, candidate, userNfDelta, nodeNfDelta);
        } else if (intent == StateIntent.CLOSE) {
            effects = _calculateCloseEffects(ctx, candidate, userNfDelta, nodeNfDelta);
        } else if (intent == StateIntent.INITIATE_ESCROW_DEPOSIT) {
            effects = _calculateInitiateEscrowDepositEffects(ctx, candidate, userNfDelta, nodeNfDelta);
        } else if (intent == StateIntent.FINALIZE_ESCROW_DEPOSIT) {
            effects = _calculateFinalizeEscrowDepositEffects(ctx, candidate, userNfDelta, nodeNfDelta);
        } else if (intent == StateIntent.INITIATE_ESCROW_WITHDRAWAL) {
            effects = _calculateInitiateEscrowWithdrawalEffects(ctx, candidate, userNfDelta, nodeNfDelta);
        } else if (intent == StateIntent.FINALIZE_ESCROW_WITHDRAWAL) {
            effects = _calculateFinalizeEscrowWithdrawalEffects(ctx, candidate, userNfDelta, nodeNfDelta);
        } else if (intent == StateIntent.INITIATE_MIGRATION) {
            effects = _calculateInitiateMigrationEffects(ctx, candidate, userNfDelta, nodeNfDelta);
        } else if (intent == StateIntent.FINALIZE_MIGRATION) {
            effects = _calculateFinalizeMigrationEffects(ctx, candidate, userNfDelta, nodeNfDelta);
        } else {
            require(false, "invalid intent");
        }

        effects.updateLastState = true;
        return effects;
    }

    function _calculateCreateEffects(TransitionContext memory ctx, State memory candidate, int256 userNfDelta)
        internal
        pure
        returns (TransitionEffects memory effects)
    {
        // CREATE-specific validations
        require(candidate.version == 0, "invalid version");
        require(ctx.status == ChannelStatus.VOID, "invalid status");
        require(candidate.homeState.nodeAllocation == 0, "node allocation must be zero");
        require(candidate.nonHomeState.isEmpty(), "non-home state must be empty");

        // Calculate effects
        effects.userFundsDelta = userNfDelta; // Pull user's initial deposit
        effects.newStatus = ChannelStatus.OPERATING;
        effects.clearDispute = false;

        return effects;
    }

    function _calculateDepositEffects(
        TransitionContext memory ctx,
        State memory candidate,
        int256 userNfDelta,
        int256 nodeNfDelta
    ) internal pure returns (TransitionEffects memory effects) {
        // DEPOSIT-specific validations
        require(
            ctx.status == ChannelStatus.OPERATING || ctx.status == ChannelStatus.DISPUTED
                || ctx.status == ChannelStatus.MIGRATING_IN,
            "invalid status"
        );
        require(userNfDelta > 0, "invalid user delta");

        // Calculate effects
        effects.userFundsDelta = userNfDelta; // Pull deposit from user
        effects.nodeFundsDelta = nodeNfDelta; // May lock more from node or release
        effects.newStatus = ChannelStatus.OPERATING;
        effects.clearDispute = (ctx.status == ChannelStatus.DISPUTED);

        return effects;
    }

    function _calculateWithdrawEffects(
        TransitionContext memory ctx,
        State memory candidate,
        int256 userNfDelta,
        int256 nodeNfDelta
    ) internal pure returns (TransitionEffects memory effects) {
        // WITHDRAW-specific validations
        require(
            ctx.status == ChannelStatus.OPERATING || ctx.status == ChannelStatus.DISPUTED
                || ctx.status == ChannelStatus.MIGRATING_IN,
            "invalid status"
        );
        require(userNfDelta < 0, "invalid user delta");

        // Calculate effects
        effects.userFundsDelta = userNfDelta; // Negative = push to user
        effects.nodeFundsDelta = nodeNfDelta;
        effects.newStatus = ChannelStatus.OPERATING;
        effects.clearDispute = (ctx.status == ChannelStatus.DISPUTED);

        return effects;
    }

    function _calculateOperateEffects(
        TransitionContext memory ctx,
        State memory candidate,
        int256 userNfDelta,
        int256 nodeNfDelta
    ) internal pure returns (TransitionEffects memory effects) {
        // OPERATE-specific validations (checkpoint)
        require(
            ctx.status == ChannelStatus.OPERATING || ctx.status == ChannelStatus.DISPUTED
                || ctx.status == ChannelStatus.MIGRATING_IN,
            "invalid status"
        );
        require(userNfDelta == 0, "invalid user delta");
        require(candidate.homeState.nodeAllocation == 0, "node allocation must be zero");

        // Calculate effects
        effects.nodeFundsDelta = nodeNfDelta; // Only node balance adjustments
        effects.newStatus = ChannelStatus.OPERATING;
        effects.clearDispute = (ctx.status == ChannelStatus.DISPUTED);

        return effects;
    }

    function _calculateCloseEffects(
        TransitionContext memory ctx,
        State memory candidate,
        int256 userNfDelta,
        int256 nodeNfDelta
    ) internal pure returns (TransitionEffects memory effects) {
        // CLOSE-specific validations
        require(
            ctx.status == ChannelStatus.OPERATING || ctx.status == ChannelStatus.DISPUTED
                || ctx.status == ChannelStatus.MIGRATING_IN,
            "invalid status"
        );

        uint256 allocsSum = candidate.homeState.userAllocation + candidate.homeState.nodeAllocation;
        require(allocsSum <= ctx.lockedFunds, "allocation exceeds locked funds");

        // Calculate effects
        // Push allocations to parties (negative = push out from channel)
        effects.userFundsDelta = userNfDelta;
        effects.nodeFundsDelta = nodeNfDelta;
        effects.newStatus = ChannelStatus.CLOSED;
        effects.closeChannel = true;

        return effects;
    }

    function _calculateInitiateEscrowDepositEffects(
        TransitionContext memory ctx,
        State memory candidate,
        int256 userNfDelta,
        int256 nodeNfDelta
    ) internal pure returns (TransitionEffects memory effects) {
        // INITIATE_ESCROW_DEPOSIT-specific validations (Home Chain)
        // Node locks liquidity in channel for cross-chain deposit
        require(
            ctx.status == ChannelStatus.OPERATING || ctx.status == ChannelStatus.DISPUTED
                || ctx.status == ChannelStatus.MIGRATING_IN,
            "invalid status"
        );
        require(userNfDelta == 0, "user delta must be zero"); // no user funds movement
        // node fund movement may accommodate transfers, thus it can be both positive or negative
        // node allocation may not have changed if previous operation is also initiate escrow deposit

        // Check home - non-home state consistency
        uint256 depositAmount = candidate.nonHomeState.userAllocation;
        require(depositAmount > 0, "deposit amount must be positive");
        require(candidate.homeState.nodeAllocation == depositAmount, "invalid home node allocation");
        require(candidate.nonHomeState.userNetFlow == depositAmount.toInt256(), "invalid non-home user net flow");

        // Calculate effects
        effects.nodeFundsDelta = nodeNfDelta; // Only node balance adjustments
        effects.newStatus = ChannelStatus.OPERATING;
        effects.clearDispute = (ctx.status == ChannelStatus.DISPUTED);

        return effects;
    }

    function _calculateFinalizeEscrowDepositEffects(
        TransitionContext memory ctx,
        State memory candidate,
        int256 userNfDelta,
        int256 nodeNfDelta
    ) internal pure returns (TransitionEffects memory effects) {
        // FINALIZE_ESCROW_DEPOSIT-specific validations (Home Chain)
        // Previous on-chain state MUST be INITIATE_ESCROW_DEPOSIT
        // Funds stay in channel, just move from node allocation to user allocation
        require(
            ctx.status == ChannelStatus.OPERATING || ctx.status == ChannelStatus.DISPUTED
                || ctx.status == ChannelStatus.MIGRATING_IN,
            "invalid status"
        );
        require(candidate.homeState.nodeAllocation == 0, "node allocation must be zero");
        // nothing changes from initiate escrow deposit state
        require(userNfDelta == 0, "user delta must be 0");
        require(nodeNfDelta == 0, "node delta must be 0");

        // Check home - non-home state consistency
        require(ctx.prevState.intent == StateIntent.INITIATE_ESCROW_DEPOSIT, "invalid intent");
        require(candidate.version == ctx.prevState.version + 1, "invalid version");
        require(candidate.nonHomeState.userAllocation == 0, "invalid non-home user allocation");
        require(candidate.nonHomeState.nodeAllocation == 0, "invalid non-home node allocation");

        uint256 depositAmount = ctx.prevState.nonHomeState.userAllocation;
        require(candidate.nonHomeState.userNetFlow == depositAmount.toInt256(), "invalid non-home user net flow");
        require(candidate.nonHomeState.nodeNetFlow == -depositAmount.toInt256(), "invalid non-home node net flow");

        uint256 userAllocDelta = candidate.homeState.userAllocation - ctx.prevState.homeState.userAllocation;
        require(userAllocDelta == depositAmount, "user allocation delta mismatch");

        // Calculate effects - funds stay in channel, no external movement
        effects.userFundsDelta = 0;
        effects.nodeFundsDelta = 0;
        effects.newStatus = ChannelStatus.OPERATING;
        effects.clearDispute = (ctx.status == ChannelStatus.DISPUTED);

        return effects;
    }

    function _calculateInitiateEscrowWithdrawalEffects(
        TransitionContext memory ctx,
        State memory candidate,
        int256 userNfDelta,
        int256 nodeNfDelta
    ) internal pure returns (TransitionEffects memory effects) {
        // INITIATE_ESCROW_WITHDRAWAL-specific validations (Home Chain)
        // Previous on-chain state can be anything, so validate like an OPERATE state + non-home State
        // Prepare for cross-chain withdrawal (state validation only)
        require(
            ctx.status == ChannelStatus.OPERATING || ctx.status == ChannelStatus.DISPUTED
                || ctx.status == ChannelStatus.MIGRATING_IN,
            "invalid status"
        );
        require(userNfDelta == 0, "user delta must be zero"); // no user funds movement
        require(candidate.homeState.nodeAllocation == 0, "node allocation must be zero");

        // Check home - non-home state consistency
        require(candidate.nonHomeState.userNetFlow == 0, "withdrawal user net flow must be zero");
        require(candidate.nonHomeState.userAllocation == 0, "withdrawal user allocation must be zero");
        require(
            candidate.nonHomeState.nodeAllocation.toInt256() == candidate.nonHomeState.nodeNetFlow,
            "invalid non-home node net flow"
        );

        // Calculate effects - no immediate fund movement
        effects.nodeFundsDelta = nodeNfDelta; // Only node balance adjustments
        effects.newStatus = ChannelStatus.OPERATING;
        effects.clearDispute = (ctx.status == ChannelStatus.DISPUTED);

        return effects;
    }

    function _calculateFinalizeEscrowWithdrawalEffects(
        TransitionContext memory ctx,
        State memory candidate,
        int256 userNfDelta,
        int256 nodeNfDelta
    ) internal pure returns (TransitionEffects memory effects) {
        // FINALIZE_ESCROW_WITHDRAWAL-specific validations (Home Chain)
        // Previous on-chain state can be anything, so validate like an OPERATE state + non-home State
        // Decrease user allocation after cross-chain withdrawal completes
        require(
            ctx.status == ChannelStatus.OPERATING || ctx.status == ChannelStatus.DISPUTED
                || ctx.status == ChannelStatus.MIGRATING_IN,
            "invalid status"
        );
        require(userNfDelta == 0, "user delta must be zero"); // no user funds movement
        require(candidate.homeState.nodeAllocation == 0, "node allocation must be zero");

        // Check home - non-home state consistency
        require(candidate.nonHomeState.userAllocation == 0, "withdrawal user allocation must be zero");
        require(candidate.nonHomeState.nodeAllocation == 0, "withdrawal node allocation must be zero");
        require(
            candidate.nonHomeState.userNetFlow == -candidate.nonHomeState.nodeNetFlow,
            "invalid non-home user net flow"
        );

        // TODO: provide V-1 state (INITIATE_ESCROW_WITHDRAWAL) to validate against?

        // Calculate effects
        effects.nodeFundsDelta = nodeNfDelta; // Only node balance adjustments
        effects.newStatus = ChannelStatus.OPERATING;
        effects.clearDispute = (ctx.status == ChannelStatus.DISPUTED);

        return effects;
    }

    function _calculateInitiateMigrationEffects(
        TransitionContext memory ctx,
        State memory candidate,
        int256 userNfDelta,
        int256 nodeNfDelta
    ) internal view returns (TransitionEffects memory effects) {
        // INITIATE_MIGRATION: Can be called on both home and non-home chain

        if (ctx.status == ChannelStatus.VOID || ctx.status == ChannelStatus.MIGRATED_OUT) {
            // NON-HOME CHAIN (IN): Create MIGRATING_IN channel
            // HomeState represents new home (current chain)

            uint256 userNonHomeAlloc = candidate.nonHomeState.userAllocation;
            require(userNonHomeAlloc > 0, "old home must have user allocation");
            require(candidate.nonHomeState.nodeAllocation == 0, "old home node allocation must be zero");

            require(candidate.homeState.userAllocation == 0, "new home user allocation must be zero");
            require(candidate.homeState.nodeAllocation == userNonHomeAlloc, "node must deposit user allocation amount");
            require(candidate.homeState.nodeNetFlow == userNonHomeAlloc.toInt256(), "invalid new home node net flow");
            require(candidate.homeState.userNetFlow == 0, "new home user net flow must be zero");

            // Calculate effects - lock node funds
            // No delta calculation needed - creating fresh channel
            effects.nodeFundsDelta = candidate.homeState.nodeAllocation.toInt256();
            effects.newStatus = ChannelStatus.MIGRATING_IN;
        } else if (ctx.status == ChannelStatus.OPERATING || ctx.status == ChannelStatus.DISPUTED) {
            // HOME CHAIN (OUT): Update state
            require(candidate.homeState.chainId == block.chainid, "invalid chain id");

            require(userNfDelta == 0, "user delta must be zero"); // no user funds movement

            // Validate homeState (current chain)
            uint256 userHomeAlloc = candidate.homeState.userAllocation;
            require(userHomeAlloc > 0, "old home must have user allocation");
            require(candidate.homeState.nodeAllocation == 0, "node allocation must be zero");

            // Validate nonHomeState (target chain)
            require(candidate.nonHomeState.userAllocation == 0, "new home user allocation must be zero");
            require(candidate.nonHomeState.nodeAllocation == userHomeAlloc, "node must deposit user allocation amount");
            require(candidate.nonHomeState.nodeNetFlow == userHomeAlloc.toInt256(), "invalid new home node net flow");
            require(candidate.nonHomeState.userNetFlow == 0, "new home user net flow must be zero");

            // Calculate effects - may adjust node vault based on net flow delta
            effects.nodeFundsDelta = nodeNfDelta;
            effects.clearDispute = (ctx.status == ChannelStatus.DISPUTED);
        } else {
            revert("invalid status for initiate migration");
        }

        return effects;
    }

    function _calculateFinalizeMigrationEffects(
        TransitionContext memory ctx,
        State memory candidate,
        int256 userNfDelta,
        int256 nodeNfDelta
    ) internal view returns (TransitionEffects memory effects) {
        // FINALIZE_MIGRATION: Can be called on both new home and old home chain

        if (ctx.status == ChannelStatus.MIGRATING_IN) {
            // NEW HOME CHAIN (IN): Move MIGRATING_IN → OPERATING
            // The home state represents the new home (current chain)
            require(candidate.homeState.chainId == block.chainid, "invalid chain id");
            require(ctx.prevState.intent == StateIntent.INITIATE_MIGRATION, "invalid previous intent");
            require(candidate.version == ctx.prevState.version + 1, "invalid version");

            uint256 userMigratedAlloc = ctx.prevState.homeState.userAllocation;

            // Validate that this completes the migration
            require(
                candidate.homeState.userAllocation == userMigratedAlloc, "user allocation must match migrated amount"
            );
            require(candidate.homeState.nodeAllocation == 0, "node allocation must be zero");
            require(candidate.nonHomeState.userAllocation == 0, "old home user allocation must be zero");
            require(candidate.nonHomeState.nodeAllocation == 0, "old home node allocation must be zero");

            // Special delta calculation: previous state was swapped during INITIATE_MIGRATION
            // So prevState.homeState represents new home (current chain)
            // Calculate deltas normally - no special handling needed since state was swapped on storage
            require(userNfDelta == 0, "user net flow delta must be zero");
            require(nodeNfDelta == 0, "node net flow delta must be zero");

            // Calculate effects - just status change
            effects.newStatus = ChannelStatus.OPERATING;
        } else if (ctx.status == ChannelStatus.OPERATING || ctx.status == ChannelStatus.DISPUTED) {
            // OLD HOME CHAIN (OUT): Release funds and move to MIGRATED_OUT
            // HomeState represents old home (current chain)

            // Validate homeState
            require(candidate.homeState.userAllocation == 0, "old home user allocation must be zero");
            require(candidate.homeState.nodeAllocation == 0, "old home node allocation must be zero");

            // Validate nonHomeState (new home)
            require(candidate.nonHomeState.userAllocation > 0, "new home user allocation must be positive");
            require(candidate.nonHomeState.nodeAllocation == 0, "new home node allocation must be zero");

            // Calculate effects - release all currently locked funds to node vault
            effects.nodeFundsDelta = nodeNfDelta;
            effects.newStatus = ChannelStatus.MIGRATED_OUT;
            effects.clearDispute = (ctx.status == ChannelStatus.DISPUTED);
            effects.closeChannel = true;
        } else {
            revert("invalid status for finalize migration");
        }

        return effects;
    }

    // ========== Internal: Phase 3 - Universal Invariants ==========

    function _validateInvariants(
        TransitionContext memory ctx,
        State memory candidate,
        TransitionEffects memory effects
    ) internal pure {
        int256 expectedLocked =
            int256(ctx.lockedFunds) + effects.userFundsDelta + effects.nodeFundsDelta;
        require(expectedLocked >= 0, "negative locked funds");

        // Check that allocations equal expected locked funds (unless deleting)
        if (!effects.closeChannel) {
            uint256 allocsSum = candidate.homeState.userAllocation + candidate.homeState.nodeAllocation;
            require(allocsSum == uint256(expectedLocked), "locked funds consistency violation");
        }

        // Check node has sufficient funds for positive nodeNfDelta
        if (effects.nodeFundsDelta > 0) {
            require(ctx.nodeAvailableFunds >= uint256(effects.nodeFundsDelta), "insufficient node balance");
        }
    }
}
