import { Address, Hex } from 'viem';
import {
    NitroliteRPCMessage,
    MessageSigner,
    SingleMessageVerifier,
    MultiMessageVerifier,
    ApplicationRPCMessage,
    RPCRequest,
    RPCData,
} from './types';
import { getCurrentTimestamp, generateRequestId } from './utils';

/**
 * NitroliteRPC utility class for creating, signing, and parsing RPC messages
 * according to the Clearnet protocol specification (which uses NitroRPC principles).
 */
export class NitroliteRPC {
    /**
     * Creates a NitroliteRPC request message.
     *
     * @param requestId - Unique ID for the request. Defaults to a generated ID.
     * @param method - The RPC method name.
     * @param params - Parameters for the method call.
     * @param timestamp - Timestamp for the request. Defaults to the current time.
     * @returns A formatted NitroliteRPCMessage object for the request.
     */
    static createRequest({
        method,
        params = {},
        requestId = generateRequestId(),
        timestamp = getCurrentTimestamp(),
        signatures = [],
    }: RPCRequest): NitroliteRPCMessage {
        const requestData: RPCData = [requestId, method, params, timestamp];
        const message: NitroliteRPCMessage = { req: requestData, sig: signatures };
        return message;
    }

    /**
     * Creates a ApplicationRPCMessage request message.
     *
     * @param requestId - Unique ID for the request. Defaults to a generated ID.
     * @param method - The RPC method name.
     * @param params - Parameters for the method call.
     * @param timestamp - Timestamp for the request. Defaults to the current time.
     * @param sid - Application session ID.
     * @returns A formatted NitroliteRPCMessage object for the request.
     */
    static createAppRequest(
        { requestId = generateRequestId(), method, params = {}, timestamp = getCurrentTimestamp() }: RPCRequest,
        sid: Hex,
    ): ApplicationRPCMessage {
        const requestData: RPCData = [requestId, method, params, timestamp];
        const message: ApplicationRPCMessage = { req: requestData, sid };
        return message;
    }

    /**
     * Extracts the payload (req or res array) from a message for signing or verification.
     *
     * @param message - The NitroliteRPCMessage to extract the payload from.
     * @returns The payload array (RequestData, ResponseData, or ErrorResponseData).
     * @throws Error if the message doesn't contain a 'req' or 'res' field.
     * @private
     */
    private static getMessagePayload(message: NitroliteRPCMessage): RPCData {
        if (message.req) return message.req;
        if (message.res) return message.res;
        throw new Error("Message must contain either 'req' or 'res' field");
    }

    /**
     * Signs a NitroliteRPC request message using the provided signer function.
     * The signature is added to the 'sig' field as an array.
     * The original message object is mutated.
     *
     * @param message - The request message to sign (must contain 'req').
     * @param signer - The signing function that takes the payload array and returns a signature.
     *                 Can use either signMessage (for general RPC messages) or signStateData (for state channel operations).
     * @returns The original message object mutated with the signature attached.
     */
    static async signRequestMessage(message: NitroliteRPCMessage, signer: MessageSigner): Promise<NitroliteRPCMessage> {
        if (!message.req) {
            throw new Error("signRequestMessage can only sign request messages containing 'req'.");
        }
        const payload = this.getMessagePayload(message);
        const signature = await signer(payload);
        message.sig = [signature];
        return message;
    }

    /**
     * Verifies a single signature for a NitroliteRPC message.
     * Assumes the signature is the first (or only) element in the 'sig' array.
     * NOTE: This is NOT called by parseResponse.
     *
     * @param message - The signed message to verify.
     * @param expectedSigner - The Ethereum address of the expected signer.
     * @param verifier - The verification function for single signatures.
     * @returns A Promise resolving to true if the signature exists and is valid, false otherwise.
     */
    static async verifySingleSignature(
        message: NitroliteRPCMessage,
        expectedSigner: Address,
        verifier: SingleMessageVerifier,
    ): Promise<boolean> {
        if (!message.sig || !Array.isArray(message.sig) || message.sig.length === 0) {
            return false;
        }

        const signature = message.sig[0];

        if (message.sig.length > 1) {
            console.error(
                'verifySingleSignature called on message with multiple signatures. Verifying only the first one.',
            );
        }

        try {
            const payload = this.getMessagePayload(message);
            if (typeof signature !== 'string') {
                return false;
            }
            return await verifier(payload, signature as Hex, expectedSigner);
        } catch (error) {
            console.error('Error during single signature verification:', error);
            return false;
        }
    }

    /**
     * Verifies multiple signatures for a NitroliteRPC message (e.g., for closing applications).
     * NOTE: This is NOT called by parseResponse.
     *
     * @param message - The signed message to verify (expects 'sig' to be an array).
     * @param expectedSigners - An array of Ethereum addresses of the required signers.
     * @param verifier - The verification function for multiple signatures.
     * @returns A Promise resolving to true if the signature field contains a valid set of signatures from the expected signers, false otherwise.
     */
    static async verifyMultipleSignatures(
        message: NitroliteRPCMessage,
        expectedSigners: Address[],
        verifier: MultiMessageVerifier,
    ): Promise<boolean> {
        if (!message.sig || !Array.isArray(message.sig)) {
            return false;
        }

        try {
            const payload = this.getMessagePayload(message);
            return await verifier(payload, message.sig as Hex[], expectedSigners);
        } catch (error) {
            console.error('Error during multiple signature verification:', error);
            return false;
        }
    }
}
