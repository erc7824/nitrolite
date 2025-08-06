import { keccak256, encodeAbiParameters, Address, Hex, recoverMessageAddress, encodePacked } from 'viem';
import { UnsignedState, State, StateHash, Signature, ChannelId } from '../client/types'; // Updated import path

/**
 * Packs a channel state into a canonical format for hashing and signing.
 * @param channelId The ID of the channel.
 * @param state The state to pack.
 * @returns The packed state as Hex.
 */
export function getPackedState(channelId: ChannelId, state: UnsignedState): Hex {
    return encodeAbiParameters(
        [
            { name: 'channelId', type: 'bytes32' },
            {
                name: 'intent',
                type: 'uint8',
            },
            {
                name: 'version',
                type: 'uint256',
            },
            { name: 'data', type: 'bytes' },
            {
                name: 'allocations',
                type: 'tuple[]',
                components: [
                    { name: 'destination', type: 'address' },
                    { name: 'token', type: 'address' },
                    { name: 'amount', type: 'uint256' },
                ],
            },
        ],
        [channelId, state.intent, state.version, state.data, state.allocations],
    );
}

/**
 * Compute the hash of a channel state in a canonical way (ignoring the signature)
 * @param channelId The channelId
 * @param state The state struct
 * @returns The state hash as Hex
 */
export function getStateHash(channelId: ChannelId, state: UnsignedState): StateHash {
    return keccak256(getPackedState(channelId, state)) as StateHash;
}

/**
 * Get a packed challenge state for a channel.
 * This function encodes the packed state and the challenge string.ß
 * @param channelId The ID of the channel.
 * @param state The state to calculate with.
 * @returns The encoded and packed challenge state as a Hex string.
 */
export function getPackedChallengeState(channelId: ChannelId, state: State): Hex {
    const packedState = getPackedState(channelId, state);
    const encoded = encodePacked(['bytes', 'string'], [packedState, 'challenge']);

    return encoded;
}

/**
 * Calculate a challenge state for a channel.
 * This function encodes the packed state and the challenge string and hashes it
 * @param channelId The ID of the channel.
 * @param state The state to calculate with.
 * @returns The challenge hash as a Hex string.
 */
export function getChallengeHash(channelId: ChannelId, state: State): Hex {
    return keccak256(getPackedChallengeState(channelId, state));
}

// TODO: extract into an interface and provide on NitroliteClient creation
/**
 * Verifies a raw ECDSA signature over a hash of a packed state.
 * @param stateHash The hash of the state.
 * @param signature The signature to verify.
 * @param expectedSigner The address of the participant expected to have signed.
 * @returns True if the signature is valid and recovers to the expected signer, false otherwise.
 */
export async function verifySignature(
    channelId: ChannelId,
    state: State,
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
