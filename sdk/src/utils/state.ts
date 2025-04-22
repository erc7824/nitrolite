import { keccak256, encodeAbiParameters, Address, Hex, recoverMessageAddress, numberToHex } from "viem";
import { State, StateHash, Signature, Channel, ChannelId } from "../client/types"; // Updated import path
import { getChannelId } from "./channel";
import { secp256k1 } from "@noble/curves/secp256k1";

/**
 * Compute the hash of a channel state in a canonical way (ignoring the signature)
 * @param channelId The channelId
 * @param state The state struct
 * @returns The state hash as Hex
 */
export function getStateHash(channelId: ChannelId, state: State): StateHash {
    const encoded = encodeAbiParameters(
        [
            { name: "channelId", type: "bytes32" },
            { name: "data", type: "bytes" },
            { name: "allocations", type: "tuple(address destination, address token, uint256 amount)[]" },
        ],
        [channelId, state.data, state.allocations]
    );

    return keccak256(encoded);
}

type ParsedSignature = { r: Hex; s: Hex; v?: bigint | undefined; yParity?: number | undefined };

/**
 * @description Parses a hex formatted signature into a structured signature.
 * (Copied from viem source for local use)
 * @param signatureHex Signature in hex format.
 * @returns The structured signature {r, s, v?, yParity?}.
 */
export function parseSignature(signatureHex: Hex): ParsedSignature {
    // Ensure the input is a valid hex string
    if (!/^0x[0-9a-fA-F]*$/.test(signatureHex) || signatureHex.length !== 132) {
        throw new Error("Invalid signature hex format");
    }
    try {
        const signatureBytes = Buffer.from(signatureHex.slice(2), "hex");
        // Use fromCompact directly on the relevant bytes (first 64 bytes for r,s)
        const sig = secp256k1.Signature.fromCompact(signatureBytes.slice(0, 64));
        const r = sig.r;
        const s = sig.s;

        // The last byte is yParityOrV
        const yParityOrV = signatureBytes[64];

        const [v, yParity] = (() => {
            if (yParityOrV === 0 || yParityOrV === 1) return [undefined, yParityOrV]; // Only yParity
            if (yParityOrV === 27) return [BigInt(yParityOrV), 0]; // v = 27, yParity = 0
            if (yParityOrV === 28) return [BigInt(yParityOrV), 1]; // v = 28, yParity = 1
            // Handle EIP-155 replay protected signatures (v = chainId * 2 + 35 + yParity)
            // This basic parser might not fully handle EIP-155 decoding back to 27/28 + chainId
            // For simplicity here, we'll assume 27/28 or 0/1 based on the copied logic
            if (yParityOrV >= 35) {
                const yParityEIP155 = (yParityOrV - 35) % 2;
                return [BigInt(yParityOrV), yParityEIP155]; // Return the raw EIP-155 'v' and derived yParity
            }
            throw new Error(`Invalid yParityOrV value: ${yParityOrV}`);
        })();

        const result: ParsedSignature = {
            r: numberToHex(r, { size: 32 }),
            s: numberToHex(s, { size: 32 }),
        };
        if (v !== undefined) result.v = v;
        if (yParity !== undefined) result.yParity = yParity;

        return result;
    } catch (error) {
        throw new Error(`Failed to parse signature: ${error instanceof Error ? error.message : String(error)}`);
    }
}

type SignMessageFn = (args: { message: { raw: Hex } | string }) => Promise<Hex>;

/**
 * Create a signature for a state hash using a Viem WalletClient or Account compatible signer.
 * Uses the locally defined parseSignature function.
 * @param stateHash The hash of the state to sign.
 * @param signer An object with a `signMessage` method compatible with Viem's interface (e.g., WalletClient, Account).
 * @returns The signature object { v, r, s }
 * @throws If the signer cannot sign messages or signing/parsing fails.
 */
export async function signState(
    stateHash: StateHash,
    signMessage: SignMessageFn // Pass the function directly
): Promise<{
    r: Hex;
    s: Hex;
    v: number;
}> {
    try {
        const signatureHex = await signMessage({ message: { raw: stateHash } });
        const parsedSig = parseSignature(signatureHex);

        if (typeof parsedSig.v === "undefined") {
            throw new Error("Signature parsing did not return a 'v' value. Unexpected signature format.");
        }

        return {
            r: parsedSig.r,
            s: parsedSig.s,
            v: Number(parsedSig.v),
        };
    } catch (error) {
        console.error("Error signing state hash:", error);
        throw new Error(`Failed to sign state hash: ${error instanceof Error ? error.message : String(error)}`);
    }
}

/**
 * Verifies that a state hash was signed by the expected signer.
 * @param stateHash The hash of the state.
 * @param signature The signature object { v, r, s }.
 * @param expectedSigner The address of the participant expected to have signed.
 * @returns True if the signature is valid and recovers to the expected signer, false otherwise.
 */
export async function verifySignature(stateHash: StateHash, signature: Signature, expectedSigner: Address): Promise<boolean> {
    try {
        // Reconstruct the flat hex signature needed by recoverMessageAddress
        // Ensure v is adjusted if necessary (e.g., some signers might return 0/1 instead of 27/28)
        const vNormalized = signature.v < 27 ? signature.v + 27 : signature.v;
        const signatureHex = `${signature.r}${signature.s.slice(2)}${vNormalized.toString(16).padStart(2, "0")}` as Hex;

        const recoveredAddress = await recoverMessageAddress({
            message: { raw: stateHash },
            signature: signatureHex,
        });

        return recoveredAddress.toLowerCase() === expectedSigner.toLowerCase();
    } catch (error) {
        console.error("Signature verification failed:", error);
        return false;
    }
}
