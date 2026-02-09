import { Address, Hex, zeroHash } from 'viem';
import * as Errors from '../errors';
import { getPackedChallengeState } from '../utils';
import { PreparerDependencies } from './prepare';
import { StateSigner, WalletStateSigner } from './signer';
import {
    ChallengeChannelParams,
    ChannelId,
    CloseChannelParams,
    CreateChannelParams,
    Signature,
    UnsignedStateV1,
} from './types';

/**
 * Shared logic for preparing and signing channel creation.
 * Used by both direct execution (createChannel) and preparation (prepareCreateChannelTransaction).
 * @param deps - The dependencies needed (account, addresses, walletClient, chainId). See {@link PreparerDependencies}.
 * @param params - Parameters for channel creation. See {@link CreateChannelParams}.
 * @returns An object containing the channel ID and signed initial state.
 * @throws {Errors.MissingParameterError} If required parameters are missing.
 * @throws {Errors.InvalidParameterError} If parameters are invalid.
 */
export async function _prepareAndSignInitialState(
    deps: PreparerDependencies,
    params: CreateChannelParams,
): Promise<{ channelId: ChannelId; initialState: CreateChannelParams['initialState'] }> {
    const { definition, initialState } = params;

    if (!definition) {
        throw new Errors.MissingParameterError('Channel definition is required for creating the channel');
    }

    if (!initialState) {
        throw new Errors.MissingParameterError('Initial state is required for creating the channel');
    }

    // For V1, the state already includes signatures from both parties
    // The RPC layer should handle getting the node signature
    return { channelId: zeroHash as ChannelId, initialState };
}

/**
 * Shared logic for preparing the challenger signature for a challenge state.
 * Used by both direct execution (challengeChannel) and preparation (prepareChallengeChannelTransaction).
 * @param deps - The dependencies needed (account, addresses, walletClient, chainId). See {@link PreparerDependencies}.
 * @param params - Parameters for challenging the channel. See {@link ChallengeChannelParams}.
 * @returns An object containing the challenger signature and other challenge parameters.
 */
export async function _prepareAndSignChallengeState(
    deps: PreparerDependencies,
    params: ChallengeChannelParams,
): Promise<{
    channelId: ChannelId;
    candidateState: ChallengeChannelParams['candidateState'];
    proofs: ChallengeChannelParams['proofs'];
    challengerSig: Signature;
}> {
    const { channelId, candidateState, proofs = [], challengerSig } = params;

    // If challengerSig is already provided, return it
    if (challengerSig) {
        return { channelId, candidateState, proofs, challengerSig };
    }

    // Otherwise, generate the challenge signature
    const unsignedState: UnsignedStateV1 = {
        version: candidateState.version,
        intent: candidateState.intent,
        metadata: candidateState.metadata,
        homeState: candidateState.homeState,
        nonHomeState: candidateState.nonHomeState,
    };

    const challengeMsg = getPackedChallengeState(channelId, unsignedState);
    const signer = await _getChannelSigner(deps, channelId);
    const newChallengerSig = await signer.signRawMessage(challengeMsg);

    return { channelId, candidateState, proofs, challengerSig: newChallengerSig };
}

/**
 * Shared logic for preparing the final state for closing a channel.
 * Used by both direct execution (closeChannel) and preparation (prepareCloseChannelTransaction).
 * @param deps - The dependencies needed (account, addresses, walletClient, chainId). See {@link PreparerDependencies}.
 * @param params - Parameters for closing the channel. See {@link CloseChannelParams}.
 * @returns An object containing the channel ID, final state, and proofs.
 */
export async function _prepareAndSignFinalState(
    deps: PreparerDependencies,
    params: CloseChannelParams,
): Promise<{ channelId: ChannelId; finalState: CloseChannelParams['finalState']; proofs: CloseChannelParams['proofs'] }> {
    const { channelId, finalState, proofs = [] } = params;

    if (!finalState) {
        throw new Errors.MissingParameterError('Final state is required for closing the channel');
    }

    // For V1, the state already includes both signatures
    return { channelId, finalState, proofs };
}

/**
 * Helper function to get the appropriate signer for a channel.
 * Fetches the channel data from the blockchain to determine the correct signer.
 * @param deps - The dependencies needed (account, addresses, walletClient, chainId). See {@link PreparerDependencies}.
 * @param channelId - The id of a channel to get the signer for.
 * @returns A StateSigner object.
 */
async function _getChannelSigner(deps: PreparerDependencies, channelId: ChannelId): Promise<StateSigner> {
    const channelData = await deps.nitroliteService.getChannelData(channelId);
    const userAddress = channelData.definition.user;

    return _checkParticipantAndGetSigner(deps, userAddress);
}

/**
 * Helper function to determine which signer to use based on the participant address.
 * @param deps - The dependencies needed (account, addresses, walletClient, chainId). See {@link PreparerDependencies}.
 * @param participant - The address of the user participant.
 * @returns A StateSigner object.
 */
function _checkParticipantAndGetSigner(deps: PreparerDependencies, participant: Address): StateSigner {
    let signer = deps.stateSigner;
    if (participant === deps.walletClient.account.address) {
        signer = new WalletStateSigner(deps.walletClient);
    }

    return signer;
}
