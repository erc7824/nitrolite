import { Hex } from "viem";
import { CreateChannelParams, State, Channel, ChannelId, CloseChannelParams, StateIntent } from "./types";
import { generateChannelNonce, getChannelId, getStateHash, signState, removeQuotesFromRS } from "../utils";
import * as Errors from "../errors";
import { PreparerDependencies } from "./prepare";

/**
 * Shared logic for preparing the channel object, initial state, and signing it.
 * Used by both direct execution (createChannel) and preparation (prepareCreateChannelTransaction).
 * @param deps - The dependencies needed (account, addresses, walletClient, challengeDuration). See {@link PreparerDependencies}.
 * @param params - Parameters for channel creation. See {@link CreateChannelParams}.
 * @returns An object containing the channel object, the signed initial state, and the channel ID.
 * @throws {Errors.MissingParameterError} If the default adjudicator address is missing.
 * @throws {Errors.InvalidParameterError} If participants are invalid.
 */
export async function _prepareAndSignInitialState(
    deps: PreparerDependencies,
    params: CreateChannelParams
): Promise<{ channel: Channel; initialState: State; channelId: ChannelId }> {
    const { initialAllocationAmounts, stateData } = params;

    if (!stateData) {
        throw new Errors.MissingParameterError("State data is required for creating the channel.");
    }

    const channelNonce = generateChannelNonce(deps.account.address);

    const participants: [Hex, Hex] = [deps.account.address, deps.addresses.guestAddress];
    const channelParticipants: [Hex, Hex] = [deps.stateWalletClient.account.address, deps.addresses.guestAddress];
    const tokenAddress = deps.addresses.tokenAddress;
    const adjudicatorAddress = deps.addresses.adjudicator;
    if (!adjudicatorAddress) {
        throw new Errors.MissingParameterError("Default adjudicator address is not configured in addresses.adjudicator");
    }

    const challengeDuration = deps.challengeDuration;

    if (!participants || participants.length !== 2) {
        throw new Errors.InvalidParameterError("Channel must have two participants.");
    }

    if (!initialAllocationAmounts || initialAllocationAmounts.length !== 2) {
        throw new Errors.InvalidParameterError("Initial allocation amounts must be provided for both participants.");
    }

    const channel: Channel = {
        participants: channelParticipants,
        adjudicator: adjudicatorAddress,
        challenge: challengeDuration,
        nonce: channelNonce,
    };

    const channelId = getChannelId(channel);

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

    const stateHash = getStateHash(channelId, stateToSign);

    const accountSignature = await signState(stateHash, deps.stateWalletClient.signMessage);
    const initialState: State = {
        ...stateToSign,
        sigs: [accountSignature],
    };

    return { channel, initialState, channelId };
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
    params: CloseChannelParams
): Promise<{ finalStateWithSigs: State; channelId: ChannelId }> {
    const { stateData, finalState } = params;

    if (!stateData) {
        throw new Errors.MissingParameterError("State data is required for closing the channel.");
    }

    const channelId = finalState.channelId;
    const serverSignature = removeQuotesFromRS(finalState.serverSignature);

    const stateToSign: State = {
        data: stateData,
        intent: StateIntent.FINALIZE,
        allocations: finalState.allocations,
        version: finalState.version,
        sigs: [],
    };

    const stateHash = getStateHash(channelId, stateToSign);

    const accountSignature = await signState(stateHash, deps.stateWalletClient.signMessage);

    // Create a new state with signatures in the requested style
    const finalStateWithSigs: State = {
        ...stateToSign,
        sigs: [accountSignature, serverSignature],
    };

    return { finalStateWithSigs, channelId };
}
