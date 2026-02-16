// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {SafeCast} from "@openzeppelin/contracts/utils/math/SafeCast.sol";
import {EscrowStatus, State, StateIntent} from "./interfaces/Types.sol";
import {Utils} from "./Utils.sol";
import {WadMath} from "./WadMath.sol";

/**
 * @title EscrowDepositEngine
 * @notice Validation and calculation engine for escrow deposit operations on non-home chain
 */
library EscrowDepositEngine {
    using SafeCast for int256;
    using SafeCast for uint256;
    using WadMath for uint256;
    using WadMath for int256;

    // ========== Constants ==========

    uint64 constant UNLOCK_DELAY = 3 hours;
    uint64 constant CHALLENGE_DURATION = 1 days;

    // ========== Structs ==========

    struct TransitionContext {
        EscrowStatus status;
        State initState;
        uint256 lockedAmount;
        uint64 unlockAt;
        uint64 challengeExpiry;
        uint256 nodeAvailableFunds;
    }

    struct TransitionEffects {
        int256 userFundsDelta;
        int256 nodeFundsDelta;
        EscrowStatus newStatus;
        uint64 newUnlockAt;
        uint64 newChallengeExpiry;
        bool updateInitState;
    }

    // ========== Public Functions ==========

    /**
     * @notice Unified validation and calculation for escrow deposit state transitions
     * @dev Three phases: universal validation → intent-specific calculation → universal invariants
     * @param ctx Current escrow context from storage
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

    /**
     * @notice Validate challenge operation (no candidate state)
     * @param ctx Current escrow context
     * @return effects The calculated effects to apply
     */
    function validateChallenge(TransitionContext memory ctx) external view returns (TransitionEffects memory effects) {
        require(ctx.status == EscrowStatus.INITIALIZED, "invalid status for challenge");
        require(block.timestamp < ctx.unlockAt, "unlock period has passed");

        effects.newStatus = EscrowStatus.DISPUTED;
        effects.newChallengeExpiry = uint64(block.timestamp) + CHALLENGE_DURATION;
        effects.updateInitState = false;

        return effects;
    }

    // ========== Internal: Phase 1 - Universal Validation ==========

    function _validateUniversal(TransitionContext memory ctx, State memory candidate) internal view {
        require(ctx.status != EscrowStatus.FINALIZED, "escrow already finalized");
        uint64 blockchainId = uint64(block.chainid);
        require(candidate.homeLedger.chainId != blockchainId, "must not be on home chain");
        require(candidate.nonHomeLedger.chainId == blockchainId, "must be on non-home chain");
        require(candidate.version > 0, "invalid version");

        // Validate token decimals for nonHomeLedger (current chain)
        Utils.validateTokenDecimals(candidate.nonHomeLedger);

        // Validate allocations equal net flows
        uint256 allocsSum = candidate.nonHomeLedger.userAllocation + candidate.nonHomeLedger.nodeAllocation;
        int256 netFlowsSum = candidate.nonHomeLedger.userNetFlow + candidate.nonHomeLedger.nodeNetFlow;

        require(netFlowsSum >= 0, "negative net flow sum");
        require(allocsSum == netFlowsSum.toUint256(), "invalid allocation sum");
    }

    // ========== Internal: Phase 2 - Intent-Specific Calculation ==========

    function _calculateEffectsByIntent(TransitionContext memory ctx, State memory candidate)
        internal
        view
        returns (TransitionEffects memory effects)
    {
        StateIntent intent = candidate.intent;

        if (intent == StateIntent.INITIATE_ESCROW_DEPOSIT) {
            effects = _calculateInitiateEffects(ctx, candidate);
        } else if (intent == StateIntent.FINALIZE_ESCROW_DEPOSIT) {
            effects = _calculateFinalizeEffects(ctx, candidate);
        } else {
            require(false, "invalid intent for escrow deposit");
        }

        return effects;
    }

    function _calculateInitiateEffects(TransitionContext memory ctx, State memory candidate)
        internal
        view
        returns (TransitionEffects memory effects)
    {
        // INITIATE: User deposits on non-home, node locks on home
        require(ctx.status == EscrowStatus.VOID, "escrow already exists");
        uint256 depositAmount = candidate.nonHomeLedger.userAllocation;
        require(candidate.nonHomeLedger.userNetFlow == depositAmount.toInt256(), "invalid user net flow");
        require(candidate.nonHomeLedger.nodeAllocation == 0, "node allocation must be zero on non-home");
        require(candidate.nonHomeLedger.nodeNetFlow == 0, "node net flow must be zero on non-home");

        // Validate that home state shows node locking equal amount
        require(
            candidate.homeLedger.nodeAllocation.toWad(candidate.homeLedger.decimals)
                == depositAmount.toWad(candidate.nonHomeLedger.decimals),
            "home node alloc must match non-home user deposit"
        );

        // Calculate effects
        effects.userFundsDelta = depositAmount.toInt256(); // Pull from user
        effects.newStatus = EscrowStatus.INITIALIZED;
        effects.newUnlockAt = uint64(block.timestamp) + UNLOCK_DELAY;
        effects.updateInitState = true;

        return effects;
    }

    function _calculateFinalizeEffects(TransitionContext memory ctx, State memory candidate)
        internal
        pure
        returns (TransitionEffects memory effects)
    {
        // FINALIZE: Node claims with finalization proof
        require(
            ctx.status == EscrowStatus.INITIALIZED || ctx.status == EscrowStatus.DISPUTED, "invalid status for finalize"
        );

        // Must be immediate successor
        require(candidate.version == ctx.initState.version + 1, "candidate must be immediate successor");
        require(ctx.initState.intent == StateIntent.INITIATE_ESCROW_DEPOSIT, "initial intent must be initiate");

        uint256 depositAmount = ctx.initState.nonHomeLedger.userAllocation;
        require(candidate.nonHomeLedger.userNetFlow == depositAmount.toInt256(), "invalid user net flow");
        require(candidate.nonHomeLedger.nodeNetFlow == -(depositAmount).toInt256(), "invalid node net flow");
        require(candidate.nonHomeLedger.userAllocation == 0, "user allocation must be zero on non-home");
        require(candidate.nonHomeLedger.nodeAllocation == 0, "node allocation must be zero on non-home");

        // Check home - non-home state consistency
        uint256 userHomeAllocDelta = candidate.homeLedger.userAllocation - ctx.initState.homeLedger.userAllocation;
        require(
            userHomeAllocDelta.toWad(candidate.homeLedger.decimals)
                == depositAmount.toWad(ctx.initState.nonHomeLedger.decimals),
            "home user allocation must increase by deposit amount"
        );
        require(candidate.homeLedger.nodeAllocation == 0, "home node allocation must be zero");
        int256 userHomeNfDelta = candidate.homeLedger.userNetFlow - ctx.initState.homeLedger.userNetFlow;
        require(userHomeNfDelta == 0, "home user net flow must not change");

        // Calculate effects
        effects.nodeFundsDelta = -(ctx.lockedAmount).toInt256(); // Release to node vault
        effects.newStatus = EscrowStatus.FINALIZED;
        effects.updateInitState = false;

        return effects;
    }

    // ========== Internal: Phase 3 - Universal Invariants ==========

    function _validateInvariants(TransitionContext memory ctx, State memory candidate, TransitionEffects memory effects)
        internal
        pure
    {
        require(effects.userFundsDelta != 0 || effects.nodeFundsDelta != 0, "no fund movement");

        int256 totalDelta = effects.userFundsDelta + effects.nodeFundsDelta;

        if (candidate.intent == StateIntent.INITIATE_ESCROW_DEPOSIT) {
            // On initiate: funds locked (positive delta)
            require(totalDelta == effects.userFundsDelta, "fund conservation on initiate");
        } else if (candidate.intent == StateIntent.FINALIZE_ESCROW_DEPOSIT) {
            // On finalize: funds released (negative delta)
            require(totalDelta == effects.nodeFundsDelta, "fund conservation on finalize");
            require(
                (-effects.nodeFundsDelta).toUint256() == ctx.lockedAmount, "released amount must equal locked amount"
            );
        }
    }
}
