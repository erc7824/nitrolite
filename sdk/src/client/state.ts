import { Address, Hex } from 'viem';
import * as Errors from '../errors';
import { generateChannelNonce, getChallengeHash, getChannelId } from '../utils';
import { PreparerDependencies } from './prepare';
import {
    ChallengeChannelParams,
    Channel,
    ChannelId,
    CloseChannelParams,
    CreateChannelParams,
    ResizeChannelParams,
    State,
    StateIntent,
    Signature,
} from './types';

/**
 * Shared logic for preparing the channel object, initial state, and signing it.
 * Used by both direct execution (createChannel) and preparation (prepareCreateChannelTransaction).
 * @param tokenAddress The address of the token for the channel.
 * @param deps - The dependencies needed (account, addresses, walletClient, challengeDuration). See {@link PreparerDependencies}.
 * @param params - Parameters for channel creation. See {@link CreateChannelParams}.
 * @returns An object containing the channel object, the signed initial state, and the channel ID.
 * @throws {Errors.MissingParameterError} If the default adjudicator address is missing.
 * @throws {Errors.InvalidParameterError} If participants are invalid.
 */
export async function _prepareAndSignInitialState(
    tokenAddress: Address,
    deps: PreparerDependencies,
    params: CreateChannelParams,
): Promise<{ channel: Channel; initialState: State; channelId: ChannelId }> {
    const { initialAllocationAmounts, stateData } = params;

    if (!stateData) {
        throw new Errors.MissingParameterError('State data is required for creating the channel');
    }

    const channelNonce = generateChannelNonce(deps.account.address);

    const participants: [Hex, Hex] = [deps.account.address, deps.addresses.guestAddress];
    const channelParticipants: [Hex, Hex] = [deps.stateSigner.getAddress(), deps.addresses.guestAddress];
    const adjudicatorAddress = deps.addresses.adjudicator;
    if (!adjudicatorAddress) {
        throw new Errors.MissingParameterError(
            'Default adjudicator address is not configured in addresses.adjudicator',
        );
    }

    const challengeDuration = deps.challengeDuration;

    if (!participants || participants.length !== 2) {
        throw new Errors.InvalidParameterError('Channel must have two participants.');
    }

    if (!initialAllocationAmounts || initialAllocationAmounts.length !== 2) {
        throw new Errors.InvalidParameterError('Initial allocation amounts must be provided for both participants.');
    }

    const channel: Channel = {
        participants: channelParticipants,
        adjudicator: adjudicatorAddress,
        challenge: challengeDuration,
        nonce: channelNonce,
    };

    const channelId = getChannelId(channel, deps.chainId);

    const stateToSign: State = {
        data: stateData,
        intent: StateIntent.INITIALIZE,
        allocations: [
            { destination: participants[0], token: tokenAddress, amount: initialAllocationAmounts[0] },
            { destination: participants[1], token: tokenAddress, amount: initialAllocationAmounts[1] },
        ],
        // The state version is set to 0 for the initial state.
        version: 0n,
        sigs: [],
    };

    const accountSignature = await deps.stateSigner.signState(channelId, stateToSign);
    const initialState: State = {
        ...stateToSign,
        sigs: [accountSignature],
    };

    return { channel, initialState, channelId };
}

/**
 * Shared logic for preparing the challenger signature for a challenge state.
 * Used by both direct execution (challengeChannel) and preparation (prepareChallengeChannelTransaction).
 * @param deps - The dependencies needed (account, addresses, walletClient, challengeDuration). See {@link PreparerDependencies}.
 * @param params - Parameters for challenging the channel. See {@link ChallengeChannelParams}.
 * @returns An object containing the challenger signature.
 */
export async function _prepareAndSignChallengeState(
    deps: PreparerDependencies,
    params: ChallengeChannelParams,
): Promise<{
    channelId: ChannelId;
    candidateState: State;
    proofs: State[];
    challengerSig: Signature;
}> {
    const { channelId, candidateState, proofStates = [] } = params;
    const challengeHash = getChallengeHash(channelId, candidateState);
    const challengerSig = await deps.stateSigner.signRawMessage(challengeHash);

    return { channelId, candidateState, proofs: proofStates, challengerSig };
}

/**
 * Shared logic for preparing the resize state for a channel and signing it.
 * Used by both direct execution (resizeChannel) and preparation (prepareResizeChannelTransaction).
 * @param deps - The dependencies needed (account, addresses, walletClient, challengeDuration). See {@link PreparerDependencies}.
 * @param params - Parameters for resizing the channel, containing the server-signed final state. See {@link ResizeChannelParams}.
 * @returns An object containing the fully signed resize state and the channel ID.
 */
export async function _prepareAndSignResizeState(
    deps: PreparerDependencies,
    params: ResizeChannelParams,
): Promise<{ resizeStateWithSigs: State; proofs: State[]; channelId: ChannelId }> {
    const { resizeState, proofStates } = params;

    if (!resizeState.data) {
        throw new Errors.MissingParameterError('State data is required for closing the channel.');
    }

    const channelId = resizeState.channelId;

    const stateToSign: State = {
        data: resizeState.data,
        intent: resizeState.intent,
        allocations: resizeState.allocations,
        version: resizeState.version,
        sigs: [],
    };

    const accountSignature = await deps.stateSigner.signState(channelId, stateToSign);

    // Create a new state with signatures in the requested style
    const resizeStateWithSigs: State = {
        ...stateToSign,
        sigs: [accountSignature, resizeState.serverSignature],
    };

    let proofs: State[] = [...proofStates];

    return { resizeStateWithSigs, proofs, channelId };
}

/**
 * Shared logic for preparing the final state for closing a channel and signing it.
 * Used by both direct execution (closeChannel) and preparation (prepareCloseChannelTransaction).
 * @param deps - The dependencies needed (account, addresses, walletClient, challengeDuration). See {@link PreparerDependencies}.
 * @param params - Parameters for closing the channel, containing the server-signed final state. See {@link CloseChannelParams}.
 * @returns An object containing the fully signed final state and the channel ID.
 */
export async function _prepareAndSignFinalState(
    deps: PreparerDependencies,
    params: CloseChannelParams,
): Promise<{ finalStateWithSigs: State; channelId: ChannelId }> {
    const { stateData, finalState } = params;

    if (!stateData) {
        throw new Errors.MissingParameterError('State data is required for closing the channel.');
    }

    const channelId = finalState.channelId;

    const stateToSign: State = {
        data: stateData,
        intent: StateIntent.FINALIZE,
        allocations: finalState.allocations,
        version: finalState.version,
        sigs: [],
    };

    const accountSignature = await deps.stateSigner.signState(channelId, stateToSign);

    // Create a new state with signatures in the requested style
    const finalStateWithSigs: State = {
        ...stateToSign,
        sigs: [accountSignature, finalState.serverSignature],
    };

    return { finalStateWithSigs, channelId };
}
