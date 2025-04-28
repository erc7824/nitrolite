import { Address } from "viem";
import { MessageSigner, AccountID, Intent, RequestID, Timestamp, ParsedResponse, CloseApplicationRequest, CreateApplicationRequest } from "./types"; // Added ParsedResponse
import { NitroliteRPC } from "./nitrolite";
import { generateRequestId, getCurrentTimestamp } from "./utils";

/**
 * Creates the signed, stringified message body for an 'auth_request'.
 * This request is sent in the context of a specific direct channel with the broker.
 *
 * @param signer - The function to sign the request payload.
 * @param clientAddress - The Ethereum address of the client authenticating.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createAuthRequestMessage(
    signer: MessageSigner,
    clientAddress: Address,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp()
): Promise<string> {
    const params = [clientAddress];

    const request = NitroliteRPC.createRequest(requestId, "auth_request", params, timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for an 'auth_verify' request
 * using an explicitly provided challenge string.
 * Use this if you have already parsed the 'auth_challenge' response yourself.
 *
 * @param signer - The function to sign the 'auth_verify' request payload.
 * @param clientAddress - The Ethereum address of the client authenticating.
 * @param challenge - The challenge string extracted from the 'auth_challenge' response.
 * @param requestId - Optional request ID for the 'auth_verify' request. Defaults to a generated ID.
 * @param timestamp - Optional timestamp for the 'auth_verify' request. Defaults to the current time.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage for 'auth_verify'.
 */
export async function createAuthVerifyMessageFromChallenge(
    signer: MessageSigner,
    clientAddress: Address,
    challenge: string,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp()
): Promise<string> {
    const verificationData = { address: clientAddress, challenge: challenge };
    const params = [verificationData];

    const request = NitroliteRPC.createRequest(requestId, "auth_verify", params, timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for an 'auth_verify' request
 * by parsing the challenge from the raw 'auth_challenge' response received from the broker.
 *
 * @param signer - The function to sign the 'auth_verify' request payload.
 * @param rawChallengeResponse - The raw JSON string or object received from the broker containing the 'auth_challenge'.
 * @param clientAddress - The Ethereum address of the client authenticating (must match the address used in 'auth_request').
 * @param requestId - Optional request ID for the 'auth_verify' request. Defaults to a generated ID.
 * @param timestamp - Optional timestamp for the 'auth_verify' request. Defaults to the current time.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage for 'auth_verify'.
 * @throws Error if the rawChallengeResponse is invalid, not an 'auth_challenge', or missing required data.
 */
export async function createAuthVerifyMessage(
    signer: MessageSigner,
    rawChallengeResponse: string | object,
    clientAddress: Address,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp()
): Promise<string> {
    const parsedResponse: ParsedResponse = NitroliteRPC.parseResponse(rawChallengeResponse);

    if (!parsedResponse.isValid) {
        throw new Error(`Invalid auth_challenge response received: ${parsedResponse.error}`);
    }
    if (parsedResponse.method !== "auth_challenge") {
        throw new Error(`Expected 'auth_challenge' method in response, but received '${parsedResponse.method}'`);
    }

    if (
        !parsedResponse.data ||
        !Array.isArray(parsedResponse.data) ||
        parsedResponse.data.length === 0 ||
        typeof parsedResponse.data[0]?.challenge_message !== "string"
    ) {
        throw new Error("Malformed data in auth_challenge response: Expected array with object containing 'challenge_message'.");
    }

    const challenge: string = parsedResponse.data[0].challenge_message;

    const verificationData = { address: clientAddress, challenge: challenge };
    const params = [verificationData];

    const request = NitroliteRPC.createRequest(requestId, "auth_verify", params, timestamp);

    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
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
    timestamp: Timestamp = getCurrentTimestamp()
): Promise<string> {
    const request = NitroliteRPC.createRequest(requestId, "ping", [], timestamp);
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
    channelId: AccountID,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp()
): Promise<string> {
    const request = NitroliteRPC.createRequest(requestId, "get_config", [], timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'get_ledger_balances' request.
 *
 * @param signer - The function to sign the request payload.
 * @param channelId - The Channel ID to get ledger balances.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createGetLedgerBalancesMessage(
    signer: MessageSigner,
    channelId: AccountID,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp()
): Promise<string> {
    const params = [{ acc: channelId }];
    const request = NitroliteRPC.createRequest(requestId, "get_ledger_balances", [params], timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'get_app_definition' request.
 *
 * @param signer - The function to sign the request payload.
 * @param appId - The Application ID to get the definition for.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createGetAppDefinitionMessage(
    signer: MessageSigner,
    appId: AccountID,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp()
): Promise<string> {
    const params = [{ acc: appId }];
    const request = NitroliteRPC.createRequest(requestId, "get_app_definition", params, timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'create_application' request.
 *
 * @param signer - The function to sign the request payload.
 * @param params - The specific parameters required by 'create_application'. See {@link CreateApplicationRequest} for details.
 * @param intent - The initial allocation intent for the application.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createApplicationMessage(
    signer: MessageSigner,
    params: CreateApplicationRequest[],
    intent: Intent,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp()
): Promise<string> {
    const request = NitroliteRPC.createRequest(requestId, "create_application", params, timestamp, intent);

    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'close_application' request.
 * Note: This function only adds the *caller's* signature. Multi-sig coordination happens externally.
 *
 * @param signer - The function to sign the request payload.
 * @param params - The specific parameters required by 'close_application' (e.g., final allocations).
 * @param intent - The final allocation intent.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage (with single signature).
 */
export async function createCloseApplicationMessage(
    signer: MessageSigner,
    params: CloseApplicationRequest[],
    intent: Intent,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp()
): Promise<string> {
    const request = NitroliteRPC.createRequest(requestId, "close_application", params, timestamp, intent);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for sending a generic 'message' within an application.
 *
 * @param signer - The function to sign the request payload.
 * @param appId - The Application ID the message is scoped to.
 * @param messageParams - The actual message content/parameters being sent.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createApplicationRPCMessage(
    signer: MessageSigner,
    appId: AccountID,
    messageParams: any[],
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp()
): Promise<string> {
    const request = NitroliteRPC.createAppRequest(requestId, "message", messageParams, timestamp, appId);
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
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp()
): Promise<string> {
    const params = [{ channel_id: channelId }];
    const request = NitroliteRPC.createRequest(requestId, "close_channel", params, timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}
