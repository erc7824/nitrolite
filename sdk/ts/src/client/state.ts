import { Address, zeroAddress } from 'viem';
import * as Errors from '../errors';
import {
    getChannelId,
    getPackedChallengeState,
} from '../utils';
import { PreparerDependencies } from './prepare';
import { StateSigner, WalletStateSigner } from './signer';
import {
    ChallengeChannelParams,
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
 * @param deps - The dependencies needed (account, addresses, walletClient, challengeDuration). See {@link PreparerDependencies}.
 * @param params - Parameters for channel creation. See {@link CreateChannelParams}.
 * @returns An object containing the signed initial state, and the channel ID.
 * @throws {Errors.MissingParameterError} If the default adjudicator address is missing.
 * @throws {Errors.InvalidParameterError} If participants are invalid.
 */
export async function _prepareAndSignInitialState(
    deps: PreparerDependencies,
    params: CreateChannelParams,
): Promise<{ initialState: State; channelId: ChannelId }> {
    const { channel, unsignedInitialState, serverSignature } = params;

    if (!unsignedInitialState) {
        throw new Errors.MissingParameterError('Initial state is required for creating the channel');
    }

    if (!unsignedInitialState.data) {
        throw new Errors.MissingParameterError('State data is required for creating the channel');
    }

    if (!unsignedInitialState.allocations || unsignedInitialState.allocations.length !== 2) {
        throw new Errors.InvalidParameterError('Initial allocation amounts must be provided for both participants.');
    }

    if (unsignedInitialState.allocations[0].amount !== 0n) {
        throw new Errors.InvalidParameterError('Initial allocation amount for the first participant must be zero.');
    }

    if (!channel) {
        throw new Errors.MissingParameterError("Channel's fixed part is required for creating the channel");
    }

    if (!channel?.adjudicator) {
        throw new Errors.MissingParameterError('Adjudicator address is required for creating the channel');
    }

    if (!channel.participants || channel.participants.length !== 2) {
        throw new Errors.InvalidParameterError('Channel must have exactly two participants.');
    }

    const signer = _checkParticipantAndGetSigner(deps, channel.participants[0]);
    const channelId = getChannelId(channel, deps.chainId);
    const accountSignature = await signer.signState(channelId, unsignedInitialState);
    const signedInitialState: State = {
        ...unsignedInitialState,
        // TODO: remove assumption, that current signer will always be the first participant
        sigs: [accountSignature, serverSignature],
    };

    return { initialState: signedInitialState, channelId };
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
    const challengeMsg = getPackedChallengeState(channelId, candidateState);
    const signer = await _fetchParticipantAndGetSigner(deps, channelId);
    const challengerSig = await signer.signRawMessage(challengeMsg);

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

    const signer = await _fetchParticipantAndGetSigner(deps, channelId);
    const accountSignature = await signer.signState(channelId, stateToSign);

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

    const signer = await _fetchParticipantAndGetSigner(deps, channelId);
    const accountSignature = await signer.signState(channelId, stateToSign);

    // Create a new state with signatures in the requested style
    const finalStateWithSigs: State = {
        ...stateToSign,
        sigs: [accountSignature, finalState.serverSignature],
    };

    return { finalStateWithSigs, channelId };
}

/**
 * Helper function supporting user participant to be either wallet or session key, returning the respective StateSigner.
 * This provides backward compatibility with channels created prior to v0.5.0.
 * Fetches the channel data from the blockchain to get the participant address.
 * @param deps - The dependencies needed (account, addresses, walletClient, challengeDuration). See {@link PreparerDependencies}.
 * @param channelId - The id of a channel to get the signer for.
 * @returns A StateSigner object depending on the user participant address.
 */
async function _fetchParticipantAndGetSigner(deps: PreparerDependencies, channelId: ChannelId): Promise<StateSigner> {
    const {channel: {participants}} = await deps.nitroliteService.getChannelData(channelId);
    let participant = participants.length == 2 ? participants[0] : zeroAddress;

    return _checkParticipantAndGetSigner(deps, participant);
}

/**
 * Helper function supporting user participant to be either wallet or session key, returning the respective StateSigner.
 * This provides backward compatibility with channels created prior to v0.5.0.
 * @param deps - The dependencies needed (account, addresses, walletClient, challengeDuration). See {@link PreparerDependencies}.
 * @param participant - The address of the user participant that a resulting StateSigner depends on.
 * @returns A StateSigner object depending on the user participant address.
 */
function _checkParticipantAndGetSigner(deps: PreparerDependencies, participant: Address): StateSigner {
    let signer = deps.stateSigner;
    if (participant == deps.walletClient.account.address) {
        signer = new WalletStateSigner(deps.walletClient);
    }

    return signer;
}
