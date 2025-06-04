import { Address, Hex, WalletClient } from 'viem';
import {
    MessageSigner,
    AccountID,
    RequestID,
    Timestamp,
    ParsedResponse,
    CloseAppSessionRequest,
    CreateAppSessionRequest,
    ResizeChannel,
    AuthRequest,
    PartialEIP712AuthMessage,
    EIP712AuthTypes,
    EIP712AuthDomain,
} from './types'; // Added ParsedResponse
import { NitroliteRPC } from './nitrolite';
import { generateRequestId, getCurrentTimestamp, parseRPCResponse } from './utils';

/**
 * Creates the signed, stringified message body for an 'auth_request'.
 * This request is sent in the context of a specific direct channel with the broker.
 *
 * @param clientAddress - The Ethereum address of the client authenticating.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createAuthRequestMessage(
    params: AuthRequest,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const allowances = Object.values(params.allowances || {}).map((v) => [v.symbol, v.amount]);
    const paramsArray = [
        params.wallet,
        params.participant,
        params.app_name,
        allowances,
        params.expire ?? '',
        params.scope ?? '',
        params.application ?? '',
    ];
    const request = NitroliteRPC.createRequest(requestId, 'auth_request', paramsArray, timestamp);

    request.sig = [''];

    return JSON.stringify(request);
}

/**
 * Creates the signed, stringified message body for an 'auth_verify' request
 * using an explicitly provided challenge string.
 * Use this if you have already parsed the 'auth_challenge' response yourself.
 *
 * @param signer - The function to sign the 'auth_verify' request payload.
 * @param challenge - The challenge string received from the broker in the 'auth_challenge' response.
 * @param requestId - Optional request ID for the 'auth_verify' request. Defaults to a generated ID.
 * @param timestamp - Optional timestamp for the 'auth_verify' request. Defaults to the current time.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage for 'auth_verify'.
 */
export async function createAuthVerifyMessageFromChallenge(
    signer: MessageSigner,
    challenge: string,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const params = [{ challenge: challenge }];

    const request = NitroliteRPC.createRequest(requestId, 'auth_verify', [params], timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for an 'auth_verify' request
 * by parsing the challenge from the raw 'auth_challenge' response received from the broker.
 *
 * @param signer - The function to sign the 'auth_verify' request payload.
 * @param rawChallengeResponse - The raw JSON string or object received from the broker containing the 'auth_challenge'.
 * @param requestId - Optional request ID for the 'auth_verify' request. Defaults to a generated ID.
 * @param timestamp - Optional timestamp for the 'auth_verify' request. Defaults to the current time.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage for 'auth_verify'.
 * @throws Error if the rawChallengeResponse is invalid, not an 'auth_challenge', or missing required data.
 */
export async function createAuthVerifyMessage(
    signer: MessageSigner,
    rawChallengeResponse: string | object,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const parsedResponse: ParsedResponse = NitroliteRPC.parseResponse(rawChallengeResponse);

    if (!parsedResponse.isValid) {
        throw new Error(`Invalid auth_challenge response received: ${parsedResponse.error}`);
    }
    if (parsedResponse.method !== 'auth_challenge') {
        throw new Error(`Expected 'auth_challenge' method in response, but received '${parsedResponse.method}'`);
    }

    if (
        !parsedResponse.data ||
        !Array.isArray(parsedResponse.data) ||
        parsedResponse.data.length === 0 ||
        typeof parsedResponse.data[0]?.challenge_message !== 'string'
    ) {
        throw new Error(
            "Malformed data in auth_challenge response: Expected array with object containing 'challenge_message'.",
        );
    }

    const challenge: string = parsedResponse.data[0].challenge_message;
    const params = [{ challenge: challenge }];

    const request = NitroliteRPC.createRequest(requestId, 'auth_verify', params, timestamp);

    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for an 'auth_verify' request
 * by providing JWT token received from the broker.
 *
 * @param jwtToken - The JWT token to use for the 'auth_verify' request.
 * @param requestId - Optional request ID for the 'auth_verify' request. Defaults to a generated ID.
 * @param timestamp - Optional timestamp for the 'auth_verify' request. Defaults to the current time.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage for 'auth_verify'.
 */
export async function createAuthVerifyMessageWithJWT(
    jwtToken: string,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const params = [{ jwt: jwtToken }];

    const request = NitroliteRPC.createRequest(requestId, 'auth_verify', params, timestamp);

    return JSON.stringify(request);
}

/**
 * Creates the signed, stringified message body for a 'ping' request.
 *
 * @param signer - The function to sign the request payload.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createPingMessage(
    signer: MessageSigner,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest(requestId, 'ping', [], timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'get_config' request.
 *
 * @param signer - The function to sign the request payload.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createGetConfigMessage(
    signer: MessageSigner,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest(requestId, 'get_config', [], timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'get_ledger_balances' request.
 *
 * @param signer - The function to sign the request payload.
 * @param participant - The participant address.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createGetLedgerBalancesMessage(
    signer: MessageSigner,
    participant: Address,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const params = [{ participant: participant }];
    const request = NitroliteRPC.createRequest(requestId, 'get_ledger_balances', params, timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'get_app_definition' request.
 *
 * @param signer - The function to sign the request payload.
 * @param appSessionId - The Application Session ID to get the definition for.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createGetAppDefinitionMessage(
    signer: MessageSigner,
    appSessionId: AccountID,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const params = [{ app_session_id: appSessionId }];
    const request = NitroliteRPC.createRequest(requestId, 'get_app_definition', params, timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'create_app_session' request.
 *
 * @param signer - The function to sign the request payload.
 * @param params - The specific parameters required by 'create_app_session'. See {@link CreateAppSessionRequest} for details.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createAppSessionMessage(
    signer: MessageSigner,
    params: CreateAppSessionRequest[],
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest(requestId, 'create_app_session', params, timestamp);

    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'close_app_session' request.
 * Note: This function only adds the *caller's* signature. Multi-sig coordination happens externally.
 *
 * @param signer - The function to sign the request payload.
 * @param params - The specific parameters required by 'close_app_session' (e.g., final allocations).
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage (with single signature).
 */
export async function createCloseAppSessionMessage(
    signer: MessageSigner,
    params: CloseAppSessionRequest[],
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest(requestId, 'close_app_session', params, timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for sending a generic 'message' within an application.
 *
 * @param signer - The function to sign the request payload.
 * @param appSessionId - The Application Session ID the message is scoped to.
 * @param messageParams - The actual message content/parameters being sent.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createApplicationMessage(
    signer: MessageSigner,
    appSessionId: Hex,
    messageParams: any[],
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createAppRequest(requestId, 'message', messageParams, timestamp, appSessionId);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'close_channel' request.
 *
 * @param signer - The function to sign the request payload.
 * @param channelId - The Channel ID to close.
 * @param params - Any specific parameters required by 'close_channel'.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createCloseChannelMessage(
    signer: MessageSigner,
    channelId: AccountID,
    fundDestination: Address,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const params = [{ channel_id: channelId, funds_destination: fundDestination }];
    const request = NitroliteRPC.createRequest(requestId, 'close_channel', params, timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'resize_channel' request.
 *
 * @param signer - The function to sign the request payload.
 * @param params - Any specific parameters required by 'resize_channel'. See {@link ResizeChannel} for details.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createResizeChannelMessage(
    signer: MessageSigner,
    params: ResizeChannel[],
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest(requestId, 'resize_channel', params, timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'get_channels' request.
 *
 * @param signer - The function to sign the request payload.
 * @param participant - The participant address.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createGetChannelsMessage(
    signer: MessageSigner,
    participant: Address,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest(requestId, 'get_channels', [{ participant }], timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates EIP-712 signing function for challenge verification with proper challenge extraction
 *
 * @param walletClient - The WalletClient instance to use for signing.
 * @param partialMessage - The partial EIP-712 message structure to complete with the challenge.
 * @param authDomain - The domain name for the EIP-712 signing context.
 * @returns A MessageSigner function that takes the challenge data and returns the EIP-712 signature.
 * @throws Error if the wallet client is not available or if challenge extraction fails.
 */
export function createEIP712AuthMessageSigner(
    walletClient: WalletClient,
    partialMessage: PartialEIP712AuthMessage,
    domain: EIP712AuthDomain,
): MessageSigner {
    return async (data: any): Promise<`0x${string}`> => {
        const address = walletClient.account?.address;
        const response = parseRPCResponse(data)

        if (response.method !== 'auth_challenge') {
            throw new Error(`Expected 'auth_challenge' method, but received '${response.method}'`);
        }

        const challengeUUID = response.params.challengeMessage;
        const message: Record<string, unknown> = {
            ...partialMessage,
            challenge: challengeUUID,
            wallet: address as Address,
        };

        try {
            // Sign with EIP-712
            const signature = await walletClient.signTypedData({
                account: walletClient.account!,
                domain,
                types: EIP712AuthTypes,
                primaryType: 'Policy',
                message,
            });

            return signature;
        } catch (eip712Error) {
            console.error('EIP-712 signing failed:', eip712Error);
            throw new Error(`EIP-712  signing failed: ${eip712Error}`);
        }
    };
}
