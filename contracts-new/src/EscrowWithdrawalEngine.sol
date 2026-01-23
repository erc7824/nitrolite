// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {SafeCast} from "lib/openzeppelin-contracts/contracts/utils/math/SafeCast.sol";
import {EscrowStatus, State, StateIntent} from "./interfaces/Types.sol";

/**
 * @title EscrowWithdrawalEngine
 * @notice Validation and calculation engine for escrow withdrawal operations on non-home chain
 */
library EscrowWithdrawalEngine {
    using SafeCast for int256;
    using SafeCast for uint256;

    // ========== Constants ==========

    uint64 constant CHALLENGE_DURATION = 1 days;

    // ========== Structs ==========

    struct TransitionContext {
        EscrowStatus status;
        State initState;
        uint256 lockedAmount;
        uint64 challengeExpiry;
        address nodeAddress;
    }

    struct TransitionEffects {
        int256 userFundsDelta;
        int256 nodeFundsDelta;
        EscrowStatus newStatus;
        uint64 newChallengeExpiry;
        bool updateInitState;
    }

    // ========== Public Functions ==========

    /**
     * @notice Unified validation and calculation for escrow withdrawal state transitions
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

        effects.newStatus = EscrowStatus.DISPUTED;
        effects.newChallengeExpiry = uint64(block.timestamp) + CHALLENGE_DURATION;
        effects.updateInitState = false;

        return effects;
    }

    // ========== Internal: Phase 1 - Universal Validation ==========

    function _validateUniversal(TransitionContext memory ctx, State memory candidate) internal view {
        require(ctx.status != EscrowStatus.FINALIZED, "escrow already finalized");
        uint64 blockchainId = uint64(block.chainid);
        require(candidate.homeState.chainId != blockchainId, "must not be on home chain");
        require(candidate.nonHomeState.chainId == blockchainId, "must be on non-home chain");
        require(candidate.version > 0, "invalid version");

        // Validate allocations equal net flows
        uint256 allocsSum = candidate.nonHomeState.userAllocation + candidate.nonHomeState.nodeAllocation;
        int256 netFlowsSum = candidate.nonHomeState.userNetFlow + candidate.nonHomeState.nodeNetFlow;

        require(netFlowsSum >= 0, "negative net flow sum");
        require(allocsSum == uint256(netFlowsSum), "invalid allocation sum");
    }

    // ========== Internal: Phase 2 - Intent-Specific Calculation ==========

    function _calculateEffectsByIntent(TransitionContext memory ctx, State memory candidate)
        internal
        pure
        returns (TransitionEffects memory effects)
    {
        StateIntent intent = candidate.intent;

        if (intent == StateIntent.INITIATE_ESCROW_WITHDRAWAL) {
            effects = _calculateInitiateEffects(ctx, candidate);
        } else if (intent == StateIntent.FINALIZE_ESCROW_WITHDRAWAL) {
            effects = _calculateFinalizeEffects(ctx, candidate);
        } else {
            require(false, "invalid intent for escrow withdrawal");
        }

        return effects;
    }

    function _calculateInitiateEffects(TransitionContext memory ctx, State memory candidate)
        internal
        pure
        returns (TransitionEffects memory effects)
    {
        // INITIATE: Node locks funds for user withdrawal
        require(ctx.status == EscrowStatus.VOID, "escrow already exists");
        require(candidate.nonHomeState.userAllocation == 0, "user allocation must be zero on non-home");
        require(candidate.nonHomeState.userNetFlow == 0, "user net flow must be zero on non-home");
        uint256 withdrawalAmount = candidate.nonHomeState.nodeAllocation;
        require(
            candidate.nonHomeState.nodeAllocation == withdrawalAmount, "node allocation must equal withdrawal amount"
        );
        require(
            candidate.nonHomeState.nodeNetFlow == withdrawalAmount.toInt256(),
            "node net flow must equal withdrawal amount"
        );

        // Validate that home state shows user has allocation to withdraw
        require(candidate.homeState.userAllocation >= withdrawalAmount, "home user allocation must be sufficient");
        require(candidate.homeState.nodeAllocation == 0, "home node allocation must be zero");

        // Calculate effects
        effects.nodeFundsDelta = withdrawalAmount.toInt256(); // Pull from node vault
        effects.newStatus = EscrowStatus.INITIALIZED;
        effects.updateInitState = true;

        return effects;
    }

    function _calculateFinalizeEffects(TransitionContext memory ctx, State memory candidate)
        internal
        pure
        returns (TransitionEffects memory effects)
    {
        // FINALIZE: Release to user with finalization proof
        require(
            ctx.status == EscrowStatus.INITIALIZED || ctx.status == EscrowStatus.DISPUTED, "invalid status for finalize"
        );

        // Must be immediate successor
        require(candidate.version == ctx.initState.version + 1, "candidate must be immediate successor");
        require(ctx.initState.intent == StateIntent.INITIATE_ESCROW_WITHDRAWAL, "intent must be initiate");

        uint256 withdrawalAmount = ctx.initState.nonHomeState.nodeAllocation;
        require(candidate.nonHomeState.userAllocation == 0, "user allocation must be zero");
        require(candidate.nonHomeState.userNetFlow == -withdrawalAmount.toInt256(), "invalid user net flow");
        require(candidate.nonHomeState.nodeAllocation == 0, "node allocation must be zero");
        require(candidate.nonHomeState.nodeNetFlow == withdrawalAmount.toInt256(), "invalid node net flow");

        // Validate homeState shows user allocation decreased
        require(
            candidate.homeState.userAllocation < ctx.initState.homeState.userAllocation,
            "home user allocation must decrease"
        );
        uint256 homeUserAllocDelta = ctx.initState.homeState.userAllocation - candidate.homeState.userAllocation;
        require(homeUserAllocDelta == withdrawalAmount, "home user alloc delta must equal withdrawal amount");

        // Node net flow decreases (becomes more negative) by withdrawal amount
        int256 homeNodeNfDelta = candidate.homeState.nodeNetFlow - ctx.initState.homeState.nodeNetFlow;
        require(homeNodeNfDelta < 0, "home node net flow must decrease");
        require(
            (-homeNodeNfDelta).toUint256() == withdrawalAmount, "home node net flow delta must equal withdrawal amount"
        );

        // Calculate effects
        effects.userFundsDelta = -ctx.lockedAmount.toInt256(); // Push to user
        effects.newStatus = EscrowStatus.FINALIZED;
        effects.updateInitState = false;

        return effects;
    }

    // ========== Internal: Phase 3 - Universal Invariants ==========

    function _validateInvariants(
        TransitionContext memory ctx,
        State memory candidate,
        TransitionEffects memory effects
    ) internal pure {
        require(effects.userFundsDelta != 0 || effects.nodeFundsDelta != 0, "no fund movement");

        int256 totalDelta = effects.userFundsDelta + effects.nodeFundsDelta;

        if (candidate.intent == StateIntent.INITIATE_ESCROW_WITHDRAWAL) {
            // On initiate: node funds locked (positive delta)
            require(totalDelta == effects.nodeFundsDelta, "fund conservation on initiate");
        } else if (candidate.intent == StateIntent.FINALIZE_ESCROW_WITHDRAWAL) {
            // On finalize: user funds released (negative delta)
            require(totalDelta == effects.userFundsDelta, "fund conservation on finalize");
            require(
                (-effects.userFundsDelta).toUint256() == ctx.lockedAmount, "released amount must equal locked amount"
            );
        }
    }
}
