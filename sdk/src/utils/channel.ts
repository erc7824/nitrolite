import { keccak256, encodeAbiParameters, Address } from "viem";
import { Channel, ChannelId } from "../services/types"; // Updated import path

/**
 * Compute the unique identifier for a channel based on its configuration.
 * The parameters included and their order should match the smart contract's channel ID calculation.
 * @param channel The channel configuration object.
 * @returns The channel identifier as Hex.
 */
export function getChannelId(channel: Channel): ChannelId {
    const encoded = encodeAbiParameters(
        [
            { name: "participants", type: "address[]" },
            { name: "adjudicator", type: "address" },
            { name: "challenge", type: "uint64" },
            { name: "nonce", type: "uint64" },
        ],
        [channel.participants, channel.adjudicator, channel.challenge, channel.nonce]
    );

    return keccak256(encoded);
}

/**
 * Generate a robust nonce for channel creation.
 * This mitigates collision risks by combining timestamp, randomness, and optionally an address.
 * @param address Optional address to mix into the nonce for further uniqueness.
 * @returns A unique BigInt nonce suitable for Channel.channelNonce.
 */
export function generateChannelNonce(address?: Address): bigint {
    const timestamp = BigInt(Date.now());
    const randomComponent = BigInt(Math.floor(Math.random() * 0xffffffff));
    let nonce = (timestamp << 32n) | randomComponent;

    if (address) {
        const addressComponent = BigInt(`0x${address.slice(-16)}`);
        nonce = nonce ^ addressComponent;
    }

    return nonce;
}
