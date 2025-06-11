import { keccak256, encodeAbiParameters, Address, Hex, recoverMessageAddress, numberToHex, parseSignature } from 'viem';
import { State, StateHash, Signature, ChannelId } from '../client/types'; // Updated import path

/**
 * Compute the hash of a channel state in a canonical way (ignoring the signature)
 * @param channelId The channelId
 * @param state The state struct
 * @returns The state hash as Hex
 */
export function getStateHash(channelId: ChannelId, state: State): StateHash {
    const encoded = encodeAbiParameters(
        [
            { name: 'channelId', type: 'bytes32' },
            // For channel creation, state.version must be 0 (corresponds to INITIAL status)
            // For active channels, state.version must be greater than 0
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

    return keccak256(encoded);
}

/**
 * Function type for signing messages, compatible with Viem's WalletClient or Account.
 * @dev Signing should not add an EIP-191 prefix to the message.
 * @param args An object containing the message to sign in the `{ message: { raw: Hex } }` format.
 * @returns A promise that resolves to the signature as a Hex string.
 * @throws If the signing fails.
 */
type SignMessageFn = (args: { message: { raw: Hex } }) => Promise<Hex>;

/**
 * Create a signature for a state hash using a Viem WalletClient or Account compatible signer.
 * Uses the locally defined parseSignature function.
 * @dev `signMessage` function should NOT add an EIP-191 prefix to the stateHash. See {@link SignMessageFn}.
 * @param stateHash The hash of the state to sign.
 * @param signer An object with a `signMessage` method compatible with Viem's interface (e.g., WalletClient, Account).
 * @returns The signature object { v, r, s }
 * @throws If the signer cannot sign messages or signing/parsing fails.
 */
export async function signState(
    stateHash: StateHash,
    signMessage: SignMessageFn,
): Promise<{
    r: Hex;
    s: Hex;
    v: number;
}> {
    try {
        const signatureHex = await signMessage({ message: { raw: stateHash } });
        const parsedSig = parseSignature(signatureHex);

        // Handle both legacy (27/28) and EIP-155 (0/1) v values
        let v: number;
        if (parsedSig.v !== undefined) {
            v = Number(parsedSig.v);
        } else if (parsedSig.yParity !== undefined) {
            v = parsedSig.yParity + 27;
        } else {
            throw new Error('Invalid signature format: missing v or yParity value');
        }

        return {
            r: parsedSig.r,
            s: parsedSig.s,
            v,
        };
    } catch (error) {
        console.error('Error signing state hash:', error);
        throw new Error(`Failed to sign state hash: ${error instanceof Error ? error.message : String(error)}`);
    }
}

export function removeQuotesFromRS(input: Signature): Signature {
    const output = { ...input };

    if (typeof output.r === 'string') {
        output.r = output.r.replace(/^"(.*)"$/, '$1') as Hex;
    }

    if (typeof output.s === 'string') {
        output.s = output.s.replace(/^"(.*)"$/, '$1') as Hex;
    }

    return output;
}

/**
 * Verifies that a state hash was signed by the expected signer.
 * @param stateHash The hash of the state.
 * @param signature The signature object { v, r, s }.
 * @param expectedSigner The address of the participant expected to have signed.
 * @returns True if the signature is valid and recovers to the expected signer, false otherwise.
 */
export async function verifySignature(
    stateHash: StateHash,
    signature: Signature,
    expectedSigner: Address,
): Promise<boolean> {
    try {
        // Reconstruct the flat hex signature needed by recoverMessageAddress
        // Ensure v is adjusted if necessary (e.g., some signers might return 0/1 instead of 27/28)
        const vNormalized = signature.v < 27 ? signature.v + 27 : signature.v;
        const signatureHex = `${signature.r}${signature.s.slice(2)}${vNormalized.toString(16).padStart(2, '0')}` as Hex;

        const recoveredAddress = await recoverMessageAddress({
            message: { raw: stateHash },
            signature: signatureHex,
        });

        return recoveredAddress.toLowerCase() === expectedSigner.toLowerCase();
    } catch (error) {
        console.error('Signature verification failed:', error);
        return false;
    }
}
