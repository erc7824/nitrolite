import { keccak256, encodeAbiParameters, Address, Hex } from 'viem';
import { Channel, ChannelId, State } from '../client/types'; // Updated import path
import { RPCChannel, RPCChannelOperationState } from '../rpc';

/**
 * Compute the unique identifier for a channel based on its configuration.
 * The parameters included and their order should match the smart contract's channel ID calculation.
 * @param channel The channel configuration object.
 * @returns The channel identifier as Hex.
 */
export function getChannelId(channel: Channel, chainId: number): ChannelId {
    const encoded = encodeAbiParameters(
        [
            { name: 'participants', type: 'address[]' },
            { name: 'adjudicator', type: 'address' },
            { name: 'challenge', type: 'uint64' },
            { name: 'nonce', type: 'uint64' },
            { name: 'chainId', type: 'uint256' },
        ],
        // @ts-ignore
        [channel.participants, channel.adjudicator, channel.challenge, channel.nonce, chainId],
    );

    return keccak256(encoded);
}

/**
 * Generate a nonce for channel creation, ensuring it fits within int64 for database compatibility.
 * This mitigates collision risks by combining timestamp, randomness, and optionally an address.
 * NOTE: This reduces the potential range compared to a full uint64.
 * @param address Optional address to mix into the nonce for further uniqueness.
 * @returns A unique BigInt nonce suitable for int64 storage.
 */
export function generateChannelNonce(address?: Address): bigint {
    const timestamp = BigInt(Math.floor(Date.now() / 1000));
    const randomComponent = BigInt(Math.floor(Math.random() * 0xffffffff));

    let combinedNonce = (timestamp << 32n) | randomComponent;

    if (address) {
        // Remove any existing 0x prefix to avoid double prefix
        const cleanAddress = address.startsWith('0x') ? address.slice(2) : address;

        if (!/^[0-9a-fA-F]+$/.test(cleanAddress)) {
            throw new Error(`Invalid address format: ${address}. Address must be a valid hex string.`);
        }

        const addressComponent = BigInt(`0x${cleanAddress.slice(-16)}`);
        combinedNonce = combinedNonce ^ addressComponent;
    }

    // Mask to ensure the value fits within int64 (max value 0x7fffffffffffffff)
    // This clears the most significant bit (sign bit for int64).
    const maxInt64 = 0x7fffffffffffffffn;
    const nonce = combinedNonce & maxInt64;

    return nonce;
}

export function convertRPCToClientChannel(ch: RPCChannel): Channel {
    return {
        participants: ch.participants,
        adjudicator: ch.adjudicator,
        challenge: BigInt(ch.challenge),
        nonce: BigInt(ch.nonce),
    };
}

export function convertRPCToClientState(s: RPCChannelOperationState, sig: Hex): State {
    return {
        intent: s.intent,
        version: BigInt(s.version),
        data: s.stateData,
        allocations: s.allocations.map((a) => ({
            token: a.token,
            destination: a.destination,
            amount: a.amount,
        })),
        sigs: [sig],
    };
}
