import { Address, Hex, stringToHex } from "viem";
import { NitroliteRPCMessage, NitroliteErrorCode, MessageSigner, MessageVerifier, RPCMessage } from "./types";
import { getCurrentTimestamp, generateRequestId } from "./utils";

/**
 * NitroliteRPC utility class for creating and signing RPC messages
 *
 * This class provides core utilities for creating RPC messages
 * which are used for communication between clients and servers in the Nitrolite protocol.
 * It handles request/response messages as well as error messages.
 */
export class NitroliteRPC {
    /** Constant representing the Host role (Client) */
    static readonly HOST = 0;

    /** Constant representing the Guest role (Server) */
    static readonly GUEST = 1;

    /**
     * Creates a NitroliteRPC request message
     *
     * @param method - The RPC method to call
     * @param params - The parameters for the method call
     * @param requestId - The unique ID for this request
     * @param timestamp - The timestamp for this request
     * @returns A formatted NitroliteRPCMessage object containing the request
     */
    static createRequest(
        method: string,
        params: any[] = [],
        requestId: number = generateRequestId(),
        timestamp: number = getCurrentTimestamp()
    ): NitroliteRPCMessage {
        return { req: [requestId, method, params, timestamp] };
    }

    /**
     * Creates a NitroliteRPC response message
     *
     * @param requestId - The ID of the request this response is for
     * @param method - The method that was called
     * @param result - The result data from the method call
     * @param timestamp - The timestamp for this response
     * @returns A formatted NitroliteRPCMessage object containing the response
     */
    static createResponse(requestId: number, method: string, result: any[] = [], timestamp: number = getCurrentTimestamp()): NitroliteRPCMessage {
        return { res: [requestId, method, result, timestamp] };
    }

    /**
     * Creates a NitroliteRPC error message
     *
     * @param requestId - The ID of the request that caused the error
     * @param code - The error code
     * @param message - The error message
     * @param timestamp - The timestamp for this error
     * @returns A formatted NitroliteRPCMessage object containing the error
     */
    static createError(requestId: number, code: number, message: string, timestamp: number = getCurrentTimestamp()): NitroliteRPCMessage {
        return { err: [requestId, code, message, timestamp] };
    }

    /**
     * Converts any value to a hex string representation
     *
     * @param value - The value to convert to hex
     * @returns The value as a Hex string
     * @private
     */
    private static toHex(value: any): Hex {
        return typeof value === "string" ? stringToHex(value) : stringToHex(JSON.stringify(value));
    }

    /**
     * Converts a NitroliteRPCMessage to the RPCMessage format for contract interaction
     *
     * @param message - The NitroliteRPCMessage to convert
     * @returns The formatted RPCMessage suitable for use with contracts
     * @throws Error if the message doesn't contain req, res, or err field
     */
    static toRPCMessage(message: NitroliteRPCMessage): RPCMessage {
        if (message.req) {
            const [requestID, method, params, timestamp] = message.req;
            return {
                requestID: BigInt(requestID),
                method,
                params: params.map(this.toHex),
                result: [],
                timestamp: BigInt(timestamp),
            };
        }

        if (message.res) {
            const [requestID, method, result, timestamp] = message.res;
            // Extract params from a request if available, otherwise empty
            const params = message.req ? message.req[2].map(this.toHex) : [];

            return {
                requestID: BigInt(requestID),
                method,
                params,
                result: result.map(this.toHex),
                timestamp: BigInt(timestamp),
            };
        }

        if (message.err) {
            const [requestID, code, errorMessage, timestamp] = message.err;
            return {
                requestID: BigInt(requestID),
                method: "error",
                params: [stringToHex(code.toString())],
                result: [stringToHex(errorMessage)],
                timestamp: BigInt(timestamp),
            };
        }

        throw new Error("Invalid message: must contain req, res, or err field");
    }

    /**
     * Extracts the payload from a message for signing
     *
     * @param message - The NitroliteRPCMessage to extract the payload from
     * @returns The stringified payload ready for signing
     * @throws Error if the message doesn't contain req, res, or err field
     * @private
     */
    static getMessagePayload(message: NitroliteRPCMessage): string {
        if (message.req) return JSON.stringify(message.req);
        if (message.res) return JSON.stringify(message.res);
        if (message.err) return JSON.stringify(message.err);
        throw new Error("Invalid message: must contain req, res, or err field");
    }

    /**
     * Signs a NitroliteRPC message using the provided signer function
     *
     * @param message - The message to sign
     * @param signer - The signing function that will produce the signature
     * @returns The original message with the signature attached
     */
    static async signMessage(message: NitroliteRPCMessage, signer: MessageSigner): Promise<NitroliteRPCMessage> {
        const payload = this.getMessagePayload(message);
        const signature = await signer(payload);

        return { ...message, sig: signature };
    }

    /**
     * Verifies a signature for a NitroliteRPC message
     *
     * @param message - The signed message to verify
     * @param expectedSigner - The address of the expected signer
     * @param verifier - The verification function to use
     * @returns True if the signature is valid, false otherwise
     */
    static async verifyMessage(message: NitroliteRPCMessage, expectedSigner: Address, verifier: MessageVerifier): Promise<boolean> {
        if (!message.sig) return false;

        try {
            const payload = this.getMessagePayload(message);
            return verifier(payload, message.sig, expectedSigner);
        } catch (error) {
            return false;
        }
    }
}
