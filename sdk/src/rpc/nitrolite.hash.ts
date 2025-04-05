import { Address, Hex, keccak256, encodeAbiParameters } from "viem";
import { NitroliteRPCMessage, MessageSigner, MessageVerifier, RPCMessage } from "./types";
import { NitroliteRPC } from "./nitrolite";

/**
 * NitroliteRPCHash utility class for state hash calculations and signatures
 * 
 * This class provides utilities for creating and verifying state hashes
 * which are used for off-chain state verification in the Nitrolite protocol.
 */
export class NitroliteRPCHash {
    /** ABI parameter definitions for request state hash calculations */
    private static readonly REQ_STATE_PARAMS = [
        { name: "channelId", type: "bytes32" },
        { name: "allocations", type: "tuple[]" },
        { name: "requestID", type: "uint64" },
        { name: "method", type: "string" },
        { name: "params", type: "bytes[]" },
        { name: "timestamp", type: "uint64" },
    ];
    
    /** ABI parameter definitions for response state hash calculations */
    private static readonly RES_STATE_PARAMS = [
        { name: "channelId", type: "bytes32" },
        { name: "allocations", type: "tuple[]" },
        { name: "requestID", type: "uint64" },
        { name: "method", type: "string" },
        { name: "params", type: "bytes[]" },
        { name: "result", type: "bytes[]" },
        { name: "timestamp", type: "uint64" },
    ];

    /**
     * Calculates the request state hash for verification
     * 
     * @param cid - The channel ID as a Hex string
     * @param out - The output allocations array
     * @param message - The RPCMessage to hash
     * @returns The state hash as a Hex string
     */
    static getReqStateHash(cid: Hex, out: any[], message: RPCMessage): Hex {
        return keccak256(
            encodeAbiParameters(
                this.REQ_STATE_PARAMS,
                [cid, out, message.requestID, message.method, message.params, message.timestamp]
            )
        );
    }

    /**
     * Calculates the response state hash for verification
     * 
     * @param cid - The channel ID as a Hex string
     * @param out - The output allocations array
     * @param message - The RPCMessage to hash
     * @returns The state hash as a Hex string
     */
    static getResStateHash(cid: Hex, out: any[], message: RPCMessage): Hex {
        return keccak256(
            encodeAbiParameters(
                this.RES_STATE_PARAMS,
                [cid, out, message.requestID, message.method, message.params, message.result, message.timestamp]
            )
        );
    }

    /**
     * Signs a NitroliteRPC message using the state hash method
     * 
     * @param cid - The channel ID as a Hex string
     * @param out - The output allocations array
     * @param message - The message to sign
     * @param signer - The signing function that will produce the signature
     * @param isHost - Whether the signer is in the host role (default: true)
     * @returns The signature as a Hex string
     */
    static async signStateHash(
        cid: Hex, 
        out: any[], 
        message: NitroliteRPCMessage, 
        signer: MessageSigner, 
        isHost: boolean = true
    ): Promise<Hex> {
        const rpcMessage = NitroliteRPC.toRPCMessage(message);
        const stateHash = isHost 
            ? this.getReqStateHash(cid, out, rpcMessage) 
            : this.getResStateHash(cid, out, rpcMessage);

        return await signer(stateHash);
    }

    /**
     * Verifies a state hash signature
     * 
     * @param cid - The channel ID as a Hex string
     * @param out - The output allocations array
     * @param message - The message that was signed
     * @param signature - The signature to verify
     * @param expectedSigner - The address of the expected signer
     * @param verifier - The verification function to use
     * @param isHost - Whether the signer is in the host role (default: true)
     * @returns True if the signature is valid, false otherwise
     */
    static async verifyStateHash(
        cid: Hex,
        out: any[],
        message: NitroliteRPCMessage,
        signature: Hex,
        expectedSigner: Address,
        verifier: MessageVerifier,
        isHost: boolean = true
    ): Promise<boolean> {
        const rpcMessage = NitroliteRPC.toRPCMessage(message);
        const stateHash = isHost 
            ? this.getReqStateHash(cid, out, rpcMessage) 
            : this.getResStateHash(cid, out, rpcMessage);

        return verifier(stateHash, signature, expectedSigner);
    }
}