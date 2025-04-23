import { Hex } from "viem";
import { NitroliteClient } from "./index"; // Import client type for context
import { CreateChannelParams, State, Channel, ChannelId, CloseChannelParams } from "./types";
import { generateChannelNonce, getChannelId, getStateHash, signState, encoders, removeQuotesFromRS } from "../utils";
import { MAGIC_NUMBERS } from "../config";
import * as Errors from "../errors";

/**
 * Shared logic for preparing the channel object, initial state, and signing it.
 * Used by both direct execution (createChannel) and preparation (prepareCreateChannelTransaction).
 * @param client - The NitroliteClient instance for context (account, addresses, walletClient).
 * @param params - Parameters for channel creation.
 * @returns An object containing the channel object, the signed initial state, and the channel ID.
 * @throws {Errors.MissingParameterError} If the default adjudicator address is missing.
 * @throws {Errors.InvalidParameterError} If participants are invalid.
 */
export async function _prepareAndSignInitialState(
    client: NitroliteClient,
    params: CreateChannelParams
): Promise<{ channel: Channel; initialState: State; channelId: ChannelId }> {
    const { initialAllocationAmounts, stateData } = params;
    const channelNonce = generateChannelNonce();

    const participants: [Hex, Hex] = [client.account.address, client.addresses.guestAddress];
    const tokenAddress = client.addresses.tokenAddress;
    const adjudicatorAddress = client.addresses.adjudicators?.["default"];
    if (!adjudicatorAddress) {
        throw new Errors.MissingParameterError("Default adjudicator address is not configured in addresses.adjudicators");
    }

    const challengeDuration = client.challengeDuration;

    if (!participants || participants.length !== 2) {
        throw new Errors.InvalidParameterError("Channel must have two participants.");
    }

    if (!initialAllocationAmounts || initialAllocationAmounts.length !== 2) {
        throw new Errors.InvalidParameterError("Initial allocation amounts must be provided for both participants.");
    }

    const channel: Channel = { participants, adjudicator: adjudicatorAddress, challenge: challengeDuration, nonce: channelNonce };

    const initialAppData = stateData ?? encoders["numeric"](MAGIC_NUMBERS.OPEN);
    const channelId = getChannelId(channel);

    const stateToSign: State = {
        data: initialAppData,
        allocations: [
            { destination: participants[0], token: tokenAddress, amount: initialAllocationAmounts[0] },
            { destination: participants[1], token: tokenAddress, amount: initialAllocationAmounts[1] },
        ],
        sigs: [],
    };

    const stateHash = getStateHash(channelId, stateToSign);

    const accountSignature = await signState(stateHash, client.walletClient.signMessage);

    const initialState: State = {
        ...stateToSign,
        sigs: [accountSignature],
    };

    return { channel, initialState, channelId };
}

/**
 * Shared logic for preparing the final state for closing a channel and signing it.
 * Used by both direct execution (closeChannel) and preparation (prepareCloseChannelTransaction).
 * @param client - The NitroliteClient instance for context (walletClient).
 * @param params - Parameters for closing the channel, containing the server-signed final state.
 * @returns An object containing the fully signed final state and the channel ID.
 */
export async function _prepareAndSignFinalState(
    client: NitroliteClient,
    params: CloseChannelParams
): Promise<{ finalStateWithSigs: State; channelId: ChannelId }> {
    const { finalState } = params;

    const channelId = finalState.channel_id;
    const finalSignatures = removeQuotesFromRS(finalState.server_signature)["server_signature"];

    const appData = encoders["numeric"](MAGIC_NUMBERS.CLOSE);

    const stateToSign: State = {
        data: appData,
        allocations: finalState.allocations,
        sigs: [],
    };

    const stateHash = getStateHash(channelId, stateToSign);

    const accountSignature = await signState(stateHash, client.walletClient.signMessage);

    const finalStateWithSigs: State = {
        ...stateToSign,
        sigs: [accountSignature, ...finalSignatures],
    };

    return { finalStateWithSigs, channelId };
}
