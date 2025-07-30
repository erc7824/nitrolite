import { keccak256, encodeAbiParameters, Address, Hex, recoverMessageAddress, numberToHex, parseSignature } from 'viem';
import { State, StateHash, Signature, ChannelId } from '../client/types'; // Updated import path
import { get } from 'http';

/**
 * Packs a channel state into a canonical format for hashing and signing.
 * @param channelId The ID of the channel.
 * @param state The state to pack.
 * @returns The packed state as Hex.
 */
export function getPackedState(channelId: ChannelId, state: State): Hex {
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
export function getStateHash(channelId: ChannelId, state: State): StateHash {
    return keccak256(getPackedState(channelId, state)) as StateHash;
}

/**
 * Function type for signing messages, compatible with Viem's WalletClient or Account.
 * @dev Signing should NOT add an EIP-191 prefix to the message.
 * @param args An object containing the message to sign in the `{ message: { raw: Hex } }` format.
 * @returns A promise that resolves to the signature as a Hex string.
 * @throws If the signing fails.
 */
type SignMessageFn = (args: { message: { raw: Hex } }) => Promise<Hex>;

/**
 * Create a raw ECDSA signature for a hash over a packed state using a Viem WalletClient or Account compatible signer.
 * Uses the locally defined parseSignature function.
 * @dev `signMessage` function should NOT add an EIP-191 prefix to the stateHash. See {@link SignMessageFn}.
 * @param stateHash The hash of the state to sign.
 * @param signer An object with a `signMessage` method compatible with Viem's interface (e.g., WalletClient, Account).
 * @returns The signature over the state hash.
 * @throws If the signer cannot sign messages or signing/parsing fails.
 */
export async function signState(
    stateHash: StateHash,
    signMessage: SignMessageFn,
): Promise<Signature> {
    try {
        return await signMessage({ message: { raw: stateHash } });
    } catch (error) {
        console.error('Error signing state hash:', error);
        throw new Error(`Failed to sign state hash: ${error instanceof Error ? error.message : String(error)}`);
    }
}

/**
 * Verifies a raw ECDSA signature over a hash of a packed state.
 * @param stateHash The hash of the state.
 * @param signature The signature to verify.
 * @param expectedSigner The address of the participant expected to have signed.
 * @returns True if the signature is valid and recovers to the expected signer, false otherwise.
 */
export async function verifySignature(
    stateHash: StateHash,
    signature: Signature,
    expectedSigner: Address,
): Promise<boolean> {
    try {
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
