import { keccak256, encodeAbiParameters, Address, Hex } from 'viem';
import { ChannelDefinition, ChannelId } from '../client/types';

/**
 * Compute the unique identifier for a V1 channel based on its definition.
 * The parameters included and their order should match the smart contract's channel ID calculation.
 * @param definition The channel definition object.
 * @param chainId The chain ID where the channel exists.
 * @returns The channel identifier as Hex.
 */
export function getChannelId(definition: ChannelDefinition, chainId: number): ChannelId {
    const encoded = encodeAbiParameters(
        [
            { name: 'challengeDuration', type: 'uint32' },
            { name: 'user', type: 'address' },
            { name: 'node', type: 'address' },
            { name: 'nonce', type: 'uint64' },
            { name: 'metadata', type: 'bytes32' },
            { name: 'chainId', type: 'uint256' },
        ],
        [definition.challengeDuration, definition.user, definition.node, definition.nonce, definition.metadata, BigInt(chainId)],
    );

    return keccak256(encoded);
}

/**
 * Calculate channel ID from individual parameters without constructing a ChannelDefinition object.
 * Convenience wrapper around getChannelId() for simpler usage.
 * @param user Address of the user.
 * @param node Address of the node.
 * @param nonce Channel nonce.
 * @param challengeDuration Challenge duration in seconds.
 * @param metadata Channel metadata as bytes32.
 * @param chainId The chain ID where the channel exists.
 * @returns The channel identifier as Hex.
 * @example
 * const channelId = calculateChannelId(
 *   "0x123...",
 *   "0x456...",
 *   1n,
 *   3600,
 *   "0x0000000000000000000000000000000000000000000000000000000000000000",
 *   1
 * );
 */
export function calculateChannelId(
    user: Address,
    node: Address,
    nonce: bigint,
    challengeDuration: number,
    metadata: Hex,
    chainId: number,
): ChannelId {
    const definition: ChannelDefinition = {
        challengeDuration,
        user,
        node,
        nonce,
        metadata,
    };

    return getChannelId(definition, chainId);
}

/**
 * Derive an escrow channel ID from a home channel ID.
 * In V1, escrow channels are deterministically derived from home channels.
 * @param homeChannelId The home channel ID.
 * @returns The escrow channel identifier as Hex.
 */
export function deriveEscrowChannelId(homeChannelId: ChannelId): ChannelId {
    // Escrow channel ID is the keccak256 hash of the home channel ID
    return keccak256(homeChannelId);
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
