import { Address, Hex } from "viem";
import { NitroliteRPCMessage, NitroliteErrorCode, MessageSigner, MessageVerifier } from "./types";
import { getCurrentTimestamp, generateRequestId } from "./utils";

/**
 * NitroliteRPC utility class for creating and signing RPC messages
 *
 * This class provides utilities for working with the NitroliteRPC protocol to
 * communicate with a Golang broker via WebSockets or other transports.
 */
export class NitroliteRPC {
    /**
     * Create a NitroliteRPC request message
     *
     * @param method Method name to call
     * @param params Parameters for the method
     * @param requestId Unique identifier for the request (optional, will generate if not provided)
     * @param timestamp Current timestamp (optional, will use current time if not provided)
     * @returns A formatted request message (unsigned)
     *
     * @example
     * const request = NitroliteRPC.createRequest('subtract', [42, 23]);
     */
    static createRequest(
        method: string,
        params: any[] = [],
        requestId: number = generateRequestId(),
        timestamp: number = getCurrentTimestamp()
    ): NitroliteRPCMessage {
        return {
            req: [requestId, method, params, timestamp],
        };
    }

    /**
     * Create a NitroliteRPC response message
     *
     * @param requestId Request ID from the original request
     * @param method Method name from the original request
     * @param result Result of the method call
     * @param timestamp Current timestamp (optional, will use current time if not provided)
     * @returns A formatted response message (unsigned)
     *
     * @example
     * const response = NitroliteRPC.createResponse(1001, 'subtract', [19]);
     */
    static createResponse(requestId: number, method: string, result: any[] = [], timestamp: number = getCurrentTimestamp()): NitroliteRPCMessage {
        return {
            res: [requestId, method, result, timestamp],
        };
    }

    /**
     * Create a NitroliteRPC error message
     *
     * @param requestId Request ID from the original request
     * @param code Error code
     * @param message Error message
     * @param timestamp Current timestamp (optional, will use current time if not provided)
     * @returns A formatted error message (unsigned)
     *
     * @example
     * const error = NitroliteRPC.createError(1001, NitroliteErrorCode.METHOD_NOT_FOUND, 'Method not found');
     */
    static createError(requestId: number, code: number, message: string, timestamp: number = getCurrentTimestamp()): NitroliteRPCMessage {
        return {
            err: [requestId, code, message, timestamp],
        };
    }

    /**
     * Sign a NitroliteRPC message
     *
     * @param message The message to sign
     * @param signer Function that signs a hex string
     * @returns Promise that resolves to the signed message
     *
     * @example
     * const signedRequest = await NitroliteRPC.signMessage(
     *   request,
     *   (payload) => account.signMessage({ message: payload })
     * );
     */
    static async signMessage(message: NitroliteRPCMessage, signer: MessageSigner): Promise<NitroliteRPCMessage> {
        // Determine which field to sign based on message type
        let payload: string;

        if (message.req) {
            payload = JSON.stringify(message.req);
        } else if (message.res) {
            payload = JSON.stringify(message.res);
        } else if (message.err) {
            payload = JSON.stringify(message.err);
        } else {
            throw new Error("Invalid message: must contain req, res, or err field");
        }

        // Sign the payload
        const signature = await signer(payload);

        // Return a new message with the signature
        return {
            ...message,
            sig: signature,
        };
    }

    /**
     * Verify a signature for a NitroliteRPC message
     *
     * @param message The message to verify
     * @param expectedSigner The expected signer address
     * @param verifier Function that verifies a signature
     * @returns Promise that resolves to true if the signature is valid
     *
     * @example
     * const isValid = await NitroliteRPC.verifyMessage(
     *   signedMessage,
     *   '0x1234...',
     *   (payload, signature, address) => isValidSignatureForAddress(payload, signature, address)
     * );
     */
    static async verifyMessage(message: NitroliteRPCMessage, expectedSigner: Address, verifier: MessageVerifier): Promise<boolean> {
        if (!message.sig) {
            return false;
        }

        let payload: string;

        if (message.req) {
            payload = JSON.stringify(message.req);
        } else if (message.res) {
            payload = JSON.stringify(message.res);
        } else if (message.err) {
            payload = JSON.stringify(message.err);
        } else {
            return false;
        }

        return verifier(payload, message.sig, expectedSigner);
    }
}
