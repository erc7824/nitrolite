import { Address, Hex } from 'viem';
import {
    RPCMessage,
    RPCMessageType,
    RPCRequest,
    RequestID,
    Timestamp,
    StateSigner,
    StateVerifier,
    MultiStateVerifier,
} from './types';
import { getCurrentTimestamp, generateRequestId } from './utils';

/**
 * NitroliteRPC utility class for creating and working with RPC messages
 * according to the Clearnode API v1 specification.
 */
export class NitroliteRPC {
    /**
     * Creates an RPC request message in wire format.
     * Wire format: [type, requestId, method, params, timestamp]
     *
     * @param request - The request object containing method, params, requestId, and timestamp
     * @returns A formatted RPCMessage array for the request
     */
    static createRequest(request: RPCRequest): RPCMessage {
        const {
            method,
            params = {},
            requestId = generateRequestId(),
            timestamp = getCurrentTimestamp(),
        } = request;

        return [RPCMessageType.Request, requestId, method, params, timestamp];
    }

    /**
     * Converts an RPCMessage wire format to a request object.
     *
     * @param message - The RPCMessage array to parse
     * @returns The parsed request object
     * @throws Error if the message is not a valid request
     */
    static parseMessage(message: RPCMessage): {
        type: RPCMessageType;
        requestId: RequestID;
        method: string;
        params: Record<string, unknown>;
        timestamp: Timestamp;
    } {
        if (!Array.isArray(message) || message.length !== 5) {
            throw new Error('Invalid RPC message format');
        }

        const [type, requestId, method, params, timestamp] = message;

        return {
            type,
            requestId,
            method,
            params,
            timestamp,
        };
    }

    /**
     * Signs state data using the provided signer function.
     * This is used for operations that require state signatures (e.g., submit_state).
     *
     * @param data - The state data to sign
     * @param signer - The signing function
     * @returns The signature as a Hex string
     */
    static async signStateData(data: unknown, signer: StateSigner): Promise<Hex> {
        return await signer(data);
    }

    /**
     * Verifies a single state signature.
     *
     * @param data - The state data that was signed
     * @param signature - The signature to verify
     * @param expectedSigner - The expected signer's address
     * @param verifier - The verification function
     * @returns True if the signature is valid, false otherwise
     */
    static async verifyStateSignature(
        data: unknown,
        signature: Hex,
        expectedSigner: Address,
        verifier: StateVerifier,
    ): Promise<boolean> {
        try {
            return await verifier(data, signature, expectedSigner);
        } catch (error) {
            console.error('Error during state signature verification:', error);
            return false;
        }
    }

    /**
     * Verifies multiple state signatures (e.g., for app session operations requiring quorum).
     *
     * @param data - The state data that was signed
     * @param signatures - Array of signatures to verify
     * @param expectedSigners - Array of expected signers' addresses
     * @param verifier - The verification function for multiple signatures
     * @returns True if all required signatures are valid, false otherwise
     */
    static async verifyMultipleStateSignatures(
        data: unknown,
        signatures: Hex[],
        expectedSigners: Address[],
        verifier: MultiStateVerifier,
    ): Promise<boolean> {
        try {
            return await verifier(data, signatures, expectedSigners);
        } catch (error) {
            console.error('Error during multiple state signature verification:', error);
            return false;
        }
    }
}
