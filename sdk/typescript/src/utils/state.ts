import { keccak256, encodeAbiParameters, Address, Hex, recoverMessageAddress, encodePacked } from 'viem';
import { UnsignedStateV1, State, StateHash, Signature, ChannelId, Ledger } from '../client/types';

/**
 * Packs a V1 channel state into a canonical format for hashing and signing.
 * @param channelId The ID of the channel.
 * @param state The unsigned state to pack (without signatures).
 * @returns The packed state as Hex.
 */
export function getPackedState(channelId: ChannelId, state: UnsignedStateV1): Hex {
    const ledgerComponents = [
        { name: 'chainId', type: 'uint64' },
        { name: 'token', type: 'address' },
        { name: 'decimals', type: 'uint8' },
        { name: 'userAllocation', type: 'uint256' },
        { name: 'userNetFlow', type: 'int256' },
        { name: 'nodeAllocation', type: 'uint256' },
        { name: 'nodeNetFlow', type: 'int256' },
    ];

    return encodeAbiParameters(
        [
            { name: 'channelId', type: 'bytes32' },
            { name: 'version', type: 'uint64' },
            { name: 'intent', type: 'uint8' },
            { name: 'metadata', type: 'bytes32' },
            {
                name: 'homeState',
                type: 'tuple',
                components: ledgerComponents,
            },
            {
                name: 'nonHomeState',
                type: 'tuple',
                components: ledgerComponents,
            },
        ],
        [
            channelId,
            state.version,
            state.intent,
            state.metadata,
            {
                chainId: state.homeState.chainId,
                token: state.homeState.token,
                decimals: state.homeState.decimals,
                userAllocation: state.homeState.userAllocation,
                userNetFlow: state.homeState.userNetFlow,
                nodeAllocation: state.homeState.nodeAllocation,
                nodeNetFlow: state.homeState.nodeNetFlow,
            },
            {
                chainId: state.nonHomeState.chainId,
                token: state.nonHomeState.token,
                decimals: state.nonHomeState.decimals,
                userAllocation: state.nonHomeState.userAllocation,
                userNetFlow: state.nonHomeState.userNetFlow,
                nodeAllocation: state.nonHomeState.nodeAllocation,
                nodeNetFlow: state.nonHomeState.nodeNetFlow,
            },
        ],
    );
}

/**
 * Compute the hash of a V1 channel state in a canonical way (ignoring the signatures)
 * @param channelId The channelId
 * @param state The unsigned state struct
 * @returns The state hash as Hex
 */
export function getStateHash(channelId: ChannelId, state: UnsignedStateV1): StateHash {
    return keccak256(getPackedState(channelId, state)) as StateHash;
}

/**
 * Get a packed challenge state for a channel.
 * This function encodes the packed state and the challenge string.
 * @param channelId The ID of the channel.
 * @param state The state to calculate with.
 * @returns The encoded and packed challenge state as a Hex string.
 */
export function getPackedChallengeState(channelId: ChannelId, state: UnsignedStateV1): Hex {
    const packedState = getPackedState(channelId, state);
    const encoded = encodePacked(['bytes', 'string'], [packedState, 'challenge']);

    return encoded;
}

/**
 * Calculate a challenge hash for a channel.
 * This function encodes the packed state and the challenge string and hashes it
 * @param channelId The ID of the channel.
 * @param state The unsigned state to calculate with.
 * @returns The challenge hash as a Hex string.
 */
export function getChallengeHash(channelId: ChannelId, state: UnsignedStateV1): Hex {
    return keccak256(getPackedChallengeState(channelId, state));
}

/**
 * Verifies a raw ECDSA signature over a hash of a packed state.
 * @param channelId The channel ID.
 * @param state The unsigned state.
 * @param signature The signature to verify.
 * @param expectedSigner The address of the participant expected to have signed.
 * @returns True if the signature is valid and recovers to the expected signer, false otherwise.
 */
export async function verifySignature(
    channelId: ChannelId,
    state: UnsignedStateV1,
    signature: Signature,
    expectedSigner: Address,
): Promise<boolean> {
    try {
        const stateHash = getStateHash(channelId, state);
        const recoveredAddress = await recoverMessageAddress({
            message: { raw: stateHash },
            signature: signature,
        });

        return recoveredAddress.toLowerCase() === expectedSigner.toLowerCase();
    } catch (error) {
        console.error('Signature verification failed:', error);
        return false;
    }
}
