// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {SafeCast} from "lib/openzeppelin-contracts/contracts/utils/math/SafeCast.sol";
import {ChannelStatus, CrossChainState, StateIntent, State} from "./interfaces/Types.sol";

/**
 * @title ChannelEngine
 * @notice Unified validation and calculation engine for all channel state transitions
 */
library ChannelEngine {
    using SafeCast for int256;
    using SafeCast for uint256;

    // ========== Structs ==========

    struct TransitionContext {
        ChannelStatus status;
        CrossChainState prevState;
        uint256 lockedFunds;
        uint256 nodeAvailableFunds;
        uint64 challengeExpiry;
    }

    struct TransitionEffects {
        // Fund movements (positive = pull/lock, negative = push/release)
        int256 userFundsDelta;         // Funds to pull from user (>0) or push to user (<0)
        int256 nodeFundsDelta;         // Funds to lock from node vault (>0) or release (<0)

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
    function validateTransition(
        TransitionContext memory ctx,
        CrossChainState memory candidate
    ) external view returns (TransitionEffects memory effects) {
        // Phase 1: Universal validation
        _validateUniversal(ctx, candidate);

        // Phase 2: Intent-specific calculation
        effects = _calculateEffectsByIntent(ctx, candidate);

        // Phase 3: Universal invariants
        _validateInvariants(ctx, candidate, effects);

        return effects;
    }

    // ========== Internal: Phase 1 - Universal Validation ==========

    function _validateUniversal(
        TransitionContext memory ctx,
        CrossChainState memory candidate
    ) internal view {
        require(candidate.homeState.chainId == block.chainid, "invalid chain id");
        require(candidate.version > ctx.prevState.version || ctx.prevState.version == 0, "invalid version");

        uint256 allocsSum = candidate.homeState.userAllocation + candidate.homeState.nodeAllocation;
        int256 netFlowsSum = candidate.homeState.userNetFlow + candidate.homeState.nodeNetFlow;

        require(netFlowsSum >= 0, "negative net flow sum");
        require(allocsSum == uint256(netFlowsSum), "invalid allocation sum");
    }

    // ========== Internal: Phase 2 - Intent-Specific Calculation ==========

    function _calculateEffectsByIntent(
        TransitionContext memory ctx,
        CrossChainState memory candidate
    ) internal pure returns (TransitionEffects memory effects) {
        int256 userDelta = candidate.homeState.userNetFlow - ctx.prevState.homeState.userNetFlow;
        int256 nodeDelta = candidate.homeState.nodeNetFlow - ctx.prevState.homeState.nodeNetFlow;

        StateIntent intent = candidate.intent;

        if (intent == StateIntent.CREATE) {
            effects = _calculateCreateEffects(ctx, candidate, userDelta);
        } else if (intent == StateIntent.DEPOSIT) {
            effects = _calculateDepositEffects(ctx, candidate, userDelta, nodeDelta);
        } else if (intent == StateIntent.WITHDRAW) {
            effects = _calculateWithdrawEffects(ctx, candidate, userDelta, nodeDelta);
        } else if (intent == StateIntent.OPERATE) {
            effects = _calculateOperateEffects(ctx, candidate, userDelta, nodeDelta);
        } else if (intent == StateIntent.CLOSE) {
            effects = _calculateCloseEffects(ctx, candidate, userDelta, nodeDelta);
        } else {
            require(false, "invalid intent");
        }

        effects.updateLastState = true;
        return effects;
    }

    function _calculateCreateEffects(
        TransitionContext memory ctx,
        CrossChainState memory candidate,
        int256 userDelta
    ) internal pure returns (TransitionEffects memory effects) {
        // CREATE-specific validations
        require(candidate.version == 0, "invalid version");
        require(ctx.status == ChannelStatus.VOID, "invalid status");
        require(candidate.homeState.nodeAllocation == 0, "node allocation must be zero");
        require(_isStateEmpty(candidate.nonHomeState), "non-home state must be empty");

        // Calculate effects
        effects.userFundsDelta = userDelta;  // Pull user's initial deposit
        effects.newStatus = ChannelStatus.OPERATING;
        effects.clearDispute = false;

        return effects;
    }

    function _calculateDepositEffects(
        TransitionContext memory ctx,
        CrossChainState memory candidate,
        int256 userDelta,
        int256 nodeDelta
    ) internal pure returns (TransitionEffects memory effects) {
        // DEPOSIT-specific validations
        require(ctx.status == ChannelStatus.OPERATING || ctx.status == ChannelStatus.DISPUTED, "invalid status");
        require(userDelta > 0, "invalid user delta");

        // Calculate effects
        effects.userFundsDelta = userDelta;  // Pull deposit from user
        effects.nodeFundsDelta = nodeDelta;  // May lock more from node or release
        effects.clearDispute = (ctx.status == ChannelStatus.DISPUTED);

        return effects;
    }

    function _calculateWithdrawEffects(
        TransitionContext memory ctx,
        CrossChainState memory candidate,
        int256 userDelta,
        int256 nodeDelta
    ) internal pure returns (TransitionEffects memory effects) {
        // WITHDRAW-specific validations
        require(ctx.status == ChannelStatus.OPERATING || ctx.status == ChannelStatus.DISPUTED, "invalid status");
        require(userDelta < 0, "invalid user delta");
        require(candidate.homeState.userAllocation <= ctx.prevState.homeState.userAllocation, "withdrawal exceeds allocation");

        // Calculate effects
        effects.userFundsDelta = userDelta;  // Negative = push to user
        effects.nodeFundsDelta = nodeDelta;
        effects.clearDispute = (ctx.status == ChannelStatus.DISPUTED);

        return effects;
    }

    function _calculateOperateEffects(
        TransitionContext memory ctx,
        CrossChainState memory candidate,
        int256 userDelta,
        int256 nodeDelta
    ) internal pure returns (TransitionEffects memory effects) {
        // OPERATE-specific validations (checkpoint)
        require(ctx.status == ChannelStatus.OPERATING || ctx.status == ChannelStatus.DISPUTED, "invalid status");
        require(userDelta == 0, "invalid user delta");

        // Calculate effects
        effects.nodeFundsDelta = nodeDelta;  // Only node balance adjustments
        effects.clearDispute = (ctx.status == ChannelStatus.DISPUTED);

        return effects;
    }

    function _calculateCloseEffects(
        TransitionContext memory ctx,
        CrossChainState memory candidate,
        int256 userDelta,
        int256 nodeDelta
    ) internal pure returns (TransitionEffects memory effects) {
        // CLOSE-specific validations
        require(ctx.status == ChannelStatus.OPERATING || ctx.status == ChannelStatus.DISPUTED, "invalid status");

        uint256 allocsSum = candidate.homeState.userAllocation + candidate.homeState.nodeAllocation;
        require(allocsSum <= ctx.lockedFunds, "allocation exceeds locked funds");

        // Calculate effects
        // Push allocations to parties (negative = push out from channel)
        effects.userFundsDelta = userDelta;
        effects.nodeFundsDelta = nodeDelta;
        effects.newStatus = ChannelStatus.CLOSED;
        effects.closeChannel = true;

        return effects;
    }

    // ========== Internal: Phase 3 - Universal Invariants ==========

    function _validateInvariants(
        TransitionContext memory ctx,
        CrossChainState memory candidate,
        TransitionEffects memory effects
    ) internal pure {
        int256 expectedLocked = int256(ctx.lockedFunds) + effects.userFundsDelta + effects.nodeFundsDelta;
        require(expectedLocked >= 0, "negative locked funds");

        // Check that allocations equal expected locked funds (unless deleting)
        if (!effects.closeChannel) {
            uint256 allocsSum = candidate.homeState.userAllocation + candidate.homeState.nodeAllocation;
            require(allocsSum == uint256(expectedLocked), "locked funds consistency violation");
        }

        // Check node has sufficient funds for positive nodeDelta
        if (effects.nodeFundsDelta > 0) {
            require(ctx.nodeAvailableFunds >= uint256(effects.nodeFundsDelta), "insufficient node balance");
        }
    }

    // ========== Helpers ==========

    function _isStateEmpty(State memory state) internal pure returns (bool) {
        return state.chainId == 0;
    }
}
