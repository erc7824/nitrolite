import { Address, Hex } from 'viem';
import {
    NitroliteRPCMessage,
    RequestData,
    NitroliteRPCErrorDetail,
    AccountID,
    MessageSigner,
    SingleMessageVerifier,
    MultiMessageVerifier,
    ParsedResponse,
    ResponsePayload,
    ApplicationRPCMessage,
    RPCResponse,
    GetConfigRPCResponse,
    GetLedgerBalancesRPCResponse,
    GetLedgerEntriesRPCResponse,
    CreateApplicationRPCResponse,
    SubmitStateRPCResponse,
    CloseApplicationRPCResponse,
    GetAppDefinitionRPCResponse,
    GetAppSessionsRPCResponse,
    ResizeChannelRPCResponse,
    CloseChannelRPCResponse,
    GetChannelsRPCResponse,
    GetRPCHistoryRPCResponse,
    GetAssetsRPCResponse,
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
    static createRequest(
        requestId: number = generateRequestId(),
        method: string,
        params: any[] = [],
        timestamp: number = getCurrentTimestamp(),
    ): NitroliteRPCMessage {
        const requestData: RequestData = [requestId, method, params, timestamp];
        const message: NitroliteRPCMessage = { req: requestData };

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
        requestId: number = generateRequestId(),
        method: string,
        params: any[] = [],
        timestamp: number = getCurrentTimestamp(),
        sid: Hex,
    ): ApplicationRPCMessage {
        const requestData: RequestData = [requestId, method, params, timestamp];
        const message: ApplicationRPCMessage = { req: requestData, sid };

        return message;
    }

    /**
     * Parses a raw message string or object received from the broker,
     * validating its structure as a Nitrolite RPC response. Handles both
     * messages with and without the top-level 'sid' field.
     * Does NOT verify the signature.
     *
     * @param rawMessage - The raw JSON string or pre-parsed object received.
     * @returns A ParsedResponse object containing the extracted data and validation status.
     */
    static parseResponse(rawMessage: string | object): ParsedResponse {
        let message: any;

        try {
            message = typeof rawMessage === 'string' ? JSON.parse(rawMessage) : rawMessage;
        } catch (e) {
            console.error('Failed to parse incoming message:', e);
            return {
                isValid: false,
                error: 'Message parsing failed',
            };
        }

        if (
            !message ||
            typeof message !== 'object' ||
            !message.res ||
            !Array.isArray(message.res) ||
            message.res.length !== 4
        ) {
            return {
                isValid: false,
                error: "Invalid message structure: Missing or invalid 'res' array.",
            };
        }

        const [requestId, method, dataPayload, timestamp] = message.res;
        const sid = typeof message.sid === 'string' ? message.sid : undefined;

        if (
            typeof requestId !== 'number' ||
            typeof method !== 'string' ||
            !Array.isArray(dataPayload) ||
            typeof timestamp !== 'number'
        ) {
            return {
                isValid: false,
                requestId,
                method,
                sid,
                timestamp,
                error: "Invalid 'res' payload structure or types.",
            };
        }

        let data: any[] | NitroliteRPCErrorDetail;
        let isError = false;
        if (method === 'error') {
            isError = true;
            if (
                dataPayload.length > 0 &&
                typeof dataPayload[0] === 'object' &&
                dataPayload[0] !== null &&
                'error' in dataPayload[0]
            ) {
                data = dataPayload[0] as NitroliteRPCErrorDetail;
            } else {
                return {
                    isValid: false,
                    requestId,
                    method,
                    sid,
                    timestamp,
                    error: 'Malformed error response payload.',
                };
            }
        } else {
            data = dataPayload;
        }

        return {
            isValid: true,
            isError: isError,
            requestId: requestId,
            method: method,
            data: data,
            sid: sid,
            timestamp: timestamp,
        };
    }

    /**
     * Type guard to check if a response is a specific RPC response type.
     * @param response - The response to check
     * @param method - The method name to check against
     * @returns True if the response is of the specified type
     */
    static isResponseType<T extends RPCResponse>(response: ParsedResponse, method: T['method']): response is ParsedResponse & { data: T['params'] } {
        return response.isValid && !response.isError && response.method === method;
    }

    /**
     * Type guard for GetConfigRPCResponse
     */
    static isGetConfigResponse(response: ParsedResponse): response is ParsedResponse & { data: GetConfigRPCResponse['params'] } {
        return this.isResponseType(response, 'get_config');
    }

    /**
     * Type guard for GetLedgerBalancesRPCResponse
     */
    static isGetLedgerBalancesResponse(response: ParsedResponse): response is ParsedResponse & { data: GetLedgerBalancesRPCResponse['params'] } {
        return this.isResponseType(response, 'get_ledger_balances');
    }

    /**
     * Type guard for GetLedgerEntriesRPCResponse
     */
    static isGetLedgerEntriesResponse(response: ParsedResponse): response is ParsedResponse & { data: GetLedgerEntriesRPCResponse['params'] } {
        return this.isResponseType(response, 'get_ledger_entries');
    }

    /**
     * Type guard for CreateApplicationRPCResponse
     */
    static isCreateApplicationResponse(response: ParsedResponse): response is ParsedResponse & { data: CreateApplicationRPCResponse['params'] } {
        return this.isResponseType(response, 'create_application');
    }

    /**
     * Type guard for SubmitStateRPCResponse
     */
    static isSubmitStateResponse(response: ParsedResponse): response is ParsedResponse & { data: SubmitStateRPCResponse['params'] } {
        return this.isResponseType(response, 'submit_state');
    }

    /**
     * Type guard for CloseApplicationRPCResponse
     */
    static isCloseApplicationResponse(response: ParsedResponse): response is ParsedResponse & { data: CloseApplicationRPCResponse['params'] } {
        return this.isResponseType(response, 'close_application');
    }

    /**
     * Type guard for GetAppDefinitionRPCResponse
     */
    static isGetAppDefinitionResponse(response: ParsedResponse): response is ParsedResponse & { data: GetAppDefinitionRPCResponse['params'] } {
        return this.isResponseType(response, 'get_app_definition');
    }

    /**
     * Type guard for GetAppSessionsRPCResponse
     */
    static isGetAppSessionsResponse(response: ParsedResponse): response is ParsedResponse & { data: GetAppSessionsRPCResponse['params'] } {
        return this.isResponseType(response, 'get_app_sessions');
    }

    /**
     * Type guard for ResizeChannelRPCResponse
     */
    static isResizeChannelResponse(response: ParsedResponse): response is ParsedResponse & { data: ResizeChannelRPCResponse['params'] } {
        return this.isResponseType(response, 'resize_channel');
    }

    /**
     * Type guard for CloseChannelRPCResponse
     */
    static isCloseChannelResponse(response: ParsedResponse): response is ParsedResponse & { data: CloseChannelRPCResponse['params'] } {
        return this.isResponseType(response, 'close_channel');
    }

    /**
     * Type guard for GetChannelsRPCResponse
     */
    static isGetChannelsResponse(response: ParsedResponse): response is ParsedResponse & { data: GetChannelsRPCResponse['params'] } {
        return this.isResponseType(response, 'get_channels');
    }

    /**
     * Type guard for GetRPCHistoryRPCResponse
     */
    static isGetRPCHistoryResponse(response: ParsedResponse): response is ParsedResponse & { data: GetRPCHistoryRPCResponse['params'] } {
        return this.isResponseType(response, 'get_rpc_history');
    }

    /**
     * Type guard for GetAssetsRPCResponse
     */
    static isGetAssetsResponse(response: ParsedResponse): response is ParsedResponse & { data: GetAssetsRPCResponse['params'] } {
        return this.isResponseType(response, 'get_assets');
    }

    /**
     * Extracts the payload (req or res array) from a message for signing or verification.
     *
     * @param message - The NitroliteRPCMessage to extract the payload from.
     * @returns The payload array (RequestData, ResponseData, or ErrorResponseData).
     * @throws Error if the message doesn't contain a 'req' or 'res' field.
     * @private
     */
    private static getMessagePayload(message: NitroliteRPCMessage): RequestData | ResponsePayload {
        if (message.req) return message.req;
        if (message.res) return message.res;
        throw new Error("Invalid message: must contain 'req' or 'res' field to define payload.");
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
            if (typeof signature !== 'string' || signature === '') {
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
