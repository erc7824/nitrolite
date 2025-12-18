import { Address, Hex, toHex, WalletClient } from 'viem';
import {
    MessageSigner,
    AccountID,
    RequestID,
    Timestamp,
    CreateAppSessionRequest,
    AuthRequestParams,
    PartialEIP712AuthMessage,
    EIP712AuthTypes,
    EIP712AuthDomain,
    EIP712AuthMessage,
    AuthChallengeResponse,
    RPCMethod,
    RPCData,
    GetLedgerTransactionsFilters,
    RPCChannelStatus,
    RPCProtocolVersion,
} from './types';
import { NitroliteRPC } from './nitrolite';
import { generateRequestId, getCurrentTimestamp } from './utils';
import {
    CloseAppSessionRequestParams,
    CreateAppSessionRequestParams,
    SubmitAppStateParamsPerProtocol,
    ResizeChannelRequestParams,
    GetLedgerTransactionsRequestParams,
    TransferRequestParams,
    CreateChannelRequestParams,
} from './types/request';
import { signRawECDSAMessage } from '../utils/sign';

/**
 * NOTE:
 * Some RPC message builders in this file have legacy variants that accept a `MessageSigner`.
 * These exist only for backward compatibility.
 *
 * Public RPC methods do NOT require signing.
 * Prefer `*V2` variants when available.
 */

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
    params: AuthRequestParams,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.AuthRequest,
        params,
        requestId,
        timestamp,
    });
    return JSON.stringify(request, (_, value) => (typeof value === 'bigint' ? Number(value) : value));
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
    const params = { challenge: challenge };

    const request = NitroliteRPC.createRequest({
        method: RPCMethod.AuthVerify,
        params,
        requestId,
        timestamp,
    });
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
    challenge: AuthChallengeResponse,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const params = { challenge: challenge.params.challengeMessage };
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.AuthVerify,
        params,
        requestId,
        timestamp,
    });
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
    const params = { jwt: jwtToken };
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.AuthVerify,
        params,
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}

/**
 * ⚠️ LEGACY — DO NOT USE IN NEW CODE
 *
 * Creates the stringified message body for a 'ping' request.
 *
 * Use {@link createPingMessageV2} instead.
 *
 * @deprecated Use createPingMessageV2(). This function will be removed in a future release.
 *
 * @param signer - Ignored. Previously required for compatibility.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to JSON string of the NitroliteRPCMessage.
 */
export async function createPingMessage(
    _signer: MessageSigner,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.Ping,
        params: {},
        requestId,
        timestamp,
    });

    return JSON.stringify(request);
}

/**
 * Creates the stringified message body for a 'ping' request.
 *
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns JSON string of the NitroliteRPCMessage.
 */
export function createPingMessageV2(
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.Ping,
        params: {},
        requestId,
        timestamp,
    });

    return JSON.stringify(request);
}

/**
 * ⚠️ LEGACY — DO NOT USE IN NEW CODE
 *
 * Creates the stringified message body for a 'get_config' request.
 *
 * Use {@link createGetConfigMessageV2} instead.
 *
 * @deprecated Use createGetConfigMessageV2(). This function will be removed in a future release.
 *
 * @param signer - Ignored. Previously required for compatibility.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to JSON string of the NitroliteRPCMessage.
 */
export async function createGetConfigMessage(
    _signer: MessageSigner,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetConfig,
        params: {},
        requestId,
        timestamp,
    });

    return JSON.stringify(request);
}

/**
 * Creates the stringified message body for a 'get_config' request.
 *
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns JSON string of the NitroliteRPCMessage.
 */
export function createGetConfigMessageV2(
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetConfig,
        params: {},
        requestId,
        timestamp,
    });

    return JSON.stringify(request);
}

/**
 * Creates the signed, stringified message body for a 'get_user_tag' request.
 *
 * @param signer - The function to sign the request payload.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createGetUserTagMessage(
    signer: MessageSigner,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetUserTag,
        params: {},
        requestId,
        timestamp,
    });
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'get_session_keys' request.
 * Retrieves all active (non-expired) session keys for the authenticated user.
 *
 * @param signer - The function to sign the request payload.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createGetSessionKeysMessage(
    signer: MessageSigner,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetSessionKeys,
        params: {},
        requestId,
        timestamp,
    });
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'get_ledger_balances' request.
 *
 * @param signer - The function to sign the request payload.
 * @param accountId - Optional account ID to filter balances.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createGetLedgerBalancesMessage(
    signer: MessageSigner,
    accountId?: string,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const params = accountId ? { account_id: accountId } : {};
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetLedgerBalances,
        params,
        requestId,
        timestamp,
    });
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * ⚠️ LEGACY — DO NOT USE IN NEW CODE
 *
 * Creates the stringified message body for a 'get_ledger_entries' request.
 *
 * Use {@link createGetLedgerEntriesMessageV2} instead.
 *
 * @deprecated Use createGetLedgerEntriesMessageV2(). This function will be removed in a future release.
 *
 * @param signer - Ignored. Previously required for compatibility.
 * @param accountId - The account ID to get entries for.
 * @param asset - Optional asset symbol to filter entries.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to JSON string of the NitroliteRPCMessage.
 */
export async function createGetLedgerEntriesMessage(
    _signer: MessageSigner,
    accountId: string,
    asset?: string,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const params = {
        account_id: accountId,
        ...(asset ? { asset } : {}),
    };
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetLedgerEntries,
        params,
        requestId,
        timestamp,
    });

    return JSON.stringify(request);
}

/**
 * Creates the stringified message body for a 'get_ledger_entries' request.
 *
 * @param accountId - The account ID to get entries for.
 * @param asset - Optional asset symbol to filter entries.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns JSON string of the NitroliteRPCMessage.
 */
export function createGetLedgerEntriesMessageV2(
    accountId: string,
    asset?: string,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const params = {
        account_id: accountId,
        ...(asset ? { asset } : {}),
    };
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetLedgerEntries,
        params,
        requestId,
        timestamp,
    });

    return JSON.stringify(request);
}

/**
 * ⚠️ LEGACY — DO NOT USE IN NEW CODE
 *
 * Creates the stringified message body for a 'get_ledger_transactions' request.
 *
 * Use {@link createGetLedgerTransactionsMessageV2} instead.
 *
 * @deprecated Use createGetLedgerTransactionsMessageV2(). This function will be removed in a future release.
 *
 * @param signer - Ignored. Previously required for compatibility.
 * @param accountId - The account ID to get transactions for.
 * @param filters - Optional filters to apply to the transactions.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to JSON string of the NitroliteRPCMessage.
 */
export async function createGetLedgerTransactionsMessage(
    _signer: MessageSigner,
    accountId: string,
    filters?: GetLedgerTransactionsFilters,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    // Build filtered parameters object
    const filteredParams: Partial<GetLedgerTransactionsFilters> = {};
    if (filters) {
        Object.entries(filters).forEach(([key, value]) => {
            if (value !== undefined && value !== null && value !== '') {
                (filteredParams as any)[key] = value;
            }
        });
    }

    const params: GetLedgerTransactionsRequestParams = {
        account_id: accountId,
        ...filteredParams,
    };

    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetLedgerTransactions,
        params,
        requestId,
        timestamp,
    });

    return JSON.stringify(request);
}

/**
 * Creates the stringified message body for a 'get_ledger_transactions' request.
 *
 * @param accountId - The account ID to get transactions for.
 * @param filters - Optional filters to apply to the transactions.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns JSON string of the NitroliteRPCMessage.
 */
export function createGetLedgerTransactionsMessageV2(
    accountId: string,
    filters?: GetLedgerTransactionsFilters,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    // Build filtered parameters object
    const filteredParams: Partial<GetLedgerTransactionsFilters> = {};
    if (filters) {
        Object.entries(filters).forEach(([key, value]) => {
            if (value !== undefined && value !== null && value !== '') {
                (filteredParams as any)[key] = value;
            }
        });
    }

    const params: GetLedgerTransactionsRequestParams = {
        account_id: accountId,
        ...filteredParams,
    };

    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetLedgerTransactions,
        params,
        requestId,
        timestamp,
    });

    return JSON.stringify(request);
}

/**
 * ⚠️ LEGACY — DO NOT USE IN NEW CODE
 *
 * Creates the stringified message body for a 'get_app_definition' request.
 *
 * Use {@link createGetAppDefinitionMessageV2} instead.
 *
 * @deprecated Use createGetAppDefinitionMessageV2(). This function will be removed in a future release.
 *
 * @param signer - Ignored. Previously required for compatibility.
 * @param appSessionId - The Application Session ID to get the definition for.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to JSON string of the NitroliteRPCMessage.
 */
export async function createGetAppDefinitionMessage(
    _signer: MessageSigner,
    appSessionId: AccountID,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const params = { app_session_id: appSessionId };
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetAppDefinition,
        params,
        requestId,
        timestamp,
    });

    return JSON.stringify(request);
}

/**
 * Creates the stringified message body for a 'get_app_definition' request.
 *
 * @param appSessionId - The Application Session ID to get the definition for.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns JSON string of the NitroliteRPCMessage.
 */
export function createGetAppDefinitionMessageV2(
    appSessionId: AccountID,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const params = { app_session_id: appSessionId };
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetAppDefinition,
        params,
        requestId,
        timestamp,
    });

    return JSON.stringify(request);
}

/**
 * ⚠️ LEGACY — DO NOT USE IN NEW CODE
 *
 * Creates the stringified message body for a 'get_app_sessions' request.
 *
 * Use {@link createGetAppSessionsMessageV2} instead.
 *
 * @deprecated Use createGetAppSessionsMessageV2(). This function will be removed in a future release.
 *
 * @param signer - Ignored. Previously required for compatibility.
 * @param participant - Participant address to filter sessions.
 * @param status - Optional status to filter sessions.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to JSON string of the NitroliteRPCMessage.
 */
export async function createGetAppSessionsMessage(
    _signer: MessageSigner,
    participant: Address,
    status?: RPCChannelStatus,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const params = {
        participant,
        ...(status ? { status } : {}),
    };
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetAppSessions,
        params,
        requestId,
        timestamp,
    });

    return JSON.stringify(request);
}

/**
 * Creates the stringified message body for a 'get_app_sessions' request.
 *
 * @param participant - Participant address to filter sessions.
 * @param status - Optional status to filter sessions.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns JSON string of the NitroliteRPCMessage.
 */
export function createGetAppSessionsMessageV2(
    participant: Address,
    status?: RPCChannelStatus,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const params = {
        participant,
        ...(status ? { status } : {}),
    };
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetAppSessions,
        params,
        requestId,
        timestamp,
    });

    return JSON.stringify(request);
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
    params: CreateAppSessionRequestParams,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.CreateAppSession,
        params,
        requestId,
        timestamp,
    });
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'submit_state' request.
 * Use the generic parameter to specify the protocol version and get type-safe parameter validation.
 *
 * @template P - The protocol version (use RPCProtocolVersion enum) to determine the required parameters structure.
 * @param signer - The function to sign the request payload.
 * @param params - The specific parameters required by 'submit_state' for the given protocol version.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 *
 * @example
 * // For NitroRPC/0.2
 * await createSubmitAppStateMessage<RPCProtocolVersion.NitroRPC_0_2>(signer, {
 *   app_session_id: '0x...',
 *   allocations: [...]
 * });
 *
 * @example
 * // For NitroRPC/0.4
 * await createSubmitAppStateMessage<RPCProtocolVersion.NitroRPC_0_4>(signer, {
 *   app_session_id: '0x...',
 *   intent: RPCAppStateIntent.Operate,
 *   version: 1,
 *   allocations: [...]
 * });
 */
export async function createSubmitAppStateMessage<P extends RPCProtocolVersion>(
    signer: MessageSigner,
    params: SubmitAppStateParamsPerProtocol[P],
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.SubmitAppState,
        params,
        requestId,
        timestamp,
    });
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
    params: CloseAppSessionRequestParams,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.CloseAppSession,
        params,
        requestId,
        timestamp,
    });
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
    messageParams: any,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createAppRequest(
        {
            method: RPCMethod.Message,
            params: messageParams,
            requestId,
            timestamp,
        },
        appSessionId,
    );
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'create_channel' request.
 *
 * @param signer - The function to sign the request payload.
 * @param params - Any specific parameters required by 'create_channel'. See {@link CreateChannelRequestParams} for details.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createCreateChannelMessage(
    signer: MessageSigner,
    params: CreateChannelRequestParams,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.CreateChannel,
        params,
        requestId,
        timestamp,
    });
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);
    return JSON.stringify(signedRequest, (_, value) => (typeof value === 'bigint' ? value.toString() : value));
}

/**
 * Creates the signed, stringified message body for a 'close_channel' request.
 *
 * @param signer - The function to sign the request payload.
 * @param channelId - The Channel ID to close.
 * @param fundDestination - The address where remaining funds should be sent upon channel closure.
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
    const params = { channel_id: channelId, funds_destination: fundDestination };
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.CloseChannel,
        params,
        requestId,
        timestamp,
    });
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'resize_channel' request.
 *
 * @param signer - The function to sign the request payload.
 * @param params - Any specific parameters required by 'resize_channel'. See {@link ResizeChannelRequestParams} for details.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createResizeChannelMessage(
    signer: MessageSigner,
    params: ResizeChannelRequestParams,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.ResizeChannel,
        params,
        requestId,
        timestamp,
    });
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest, (_, value) => (typeof value === 'bigint' ? value.toString() : value));
}

/**
 * ⚠️ LEGACY — DO NOT USE IN NEW CODE
 *
 * Creates the stringified message body for a 'get_channels' request.
 *
 * Use {@link createGetChannelsMessageV2} instead.
 *
 * @deprecated Use createGetChannelsMessageV2(). This function will be removed in a future release.
 *
 * @param signer - Ignored. Previously required for compatibility.
 * @param participant - Optional participant address to filter channels.
 * @param status - Optional status to filter channels.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to JSON string of the NitroliteRPCMessage.
 */
export async function createGetChannelsMessage(
    _signer: MessageSigner,
    participant?: Address,
    status?: RPCChannelStatus,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const params = {
        ...(participant ? { participant } : {}),
        ...(status ? { status } : {}),
    };
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetChannels,
        params,
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}

/**
 * Creates the stringified message body for a 'get_channels' request.
 *
 * @param participant - Optional participant address to filter channels.
 * @param status - Optional status to filter channels.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns JSON string of the NitroliteRPCMessage.
 */
export function createGetChannelsMessageV2(
    participant?: Address,
    status?: RPCChannelStatus,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const params = {
        ...(participant ? { participant } : {}),
        ...(status ? { status } : {}),
    };
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetChannels,
        params,
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}

/**
 * Creates the signed, stringified message body for a 'get_rpc_history' request.
 *
 * @param signer - The function to sign the request payload.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createGetRPCHistoryMessage(
    signer: MessageSigner,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetRPCHistory,
        params: {},
        requestId,
        timestamp,
    });
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * ⚠️ LEGACY — DO NOT USE IN NEW CODE
 *
 * Creates the stringified message body for a 'get_assets' request.
 *
 * Use {@link createGetAssetsMessageV2} instead.
 *
 * @deprecated Use createGetAssetsMessageV2(). This function will be removed in a future release.
 *
 * @param signer - Ignored. Previously required for compatibility.
 * @param chainId - Optional chain ID to filter assets.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to JSON string of the NitroliteRPCMessage.
 */
export async function createGetAssetsMessage(
    _signer: MessageSigner,
    chainId?: number,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const params = {
        ...(chainId ? { chain_id: chainId } : {}),
    };
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetAssets,
        params,
        requestId,
        timestamp,
    });

    return JSON.stringify(request);
}

/**
 * Creates the stringified message body for a 'get_assets' request.
 *
 * @param chainId - Optional chain ID to filter assets.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns JSON string of the NitroliteRPCMessage.
 */
export function createGetAssetsMessageV2(
    chainId?: number,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const params = {
        ...(chainId ? { chain_id: chainId } : {}),
    };
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetAssets,
        params,
        requestId,
        timestamp,
    });

    return JSON.stringify(request);
}

/**
 * Creates the signed, stringified message body for a 'transfer' request.
 *
 * @param signer - The function to sign the request payload.
 * @param transferParams - The transfer parameters including destination/destination_user_tag and allocations.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createTransferMessage(
    signer: MessageSigner,
    params: TransferRequestParams,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    // Validate that exactly one destination type is provided (XOR logic)
    const hasDestination = !!params.destination;
    const hasDestinationTag = !!params.destination_user_tag;

    if (hasDestination === hasDestinationTag) {
        throw new Error(
            hasDestination
                ? 'Cannot provide both destination and destination_user_tag'
                : 'Either destination or destination_user_tag must be provided',
        );
    }

    const request = NitroliteRPC.createRequest({
        method: RPCMethod.Transfer,
        params,
        requestId,
        timestamp,
    });
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates a revoke session key message
 *
 * @param signer - The message signer to sign the request
 * @param sessionKey - The session key address to revoke
 * @param requestId - Optional request ID (auto-generated if not provided)
 * @param timestamp - Optional timestamp (auto-generated if not provided)
 * @returns JSON string of the signed RPC message
 */
export async function createRevokeSessionKeyMessage(
    signer: MessageSigner,
    sessionKey: Address,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.RevokeSessionKey,
        params: { session_key: sessionKey },
        requestId,
        timestamp,
    });
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates a cleanup session key cache message
 *
 * @param signer - The message signer to sign the request
 * @param requestId - Optional request ID (auto-generated if not provided)
 * @param timestamp - Optional timestamp (auto-generated if not provided)
 * @returns JSON string of the signed RPC message
 */
export async function createCleanupSessionKeyCacheMessage(
    signer: MessageSigner,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.CleanupSessionKeyCache,
        params: {},
        requestId,
        timestamp,
    });
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
 */
export function createEIP712AuthMessageSigner(
    walletClient: WalletClient,
    partialMessage: PartialEIP712AuthMessage,
    domain: EIP712AuthDomain,
): MessageSigner {
    return async (payload: RPCData): Promise<Hex> => {
        const address = walletClient.account?.address;
        if (!address) {
            throw new Error('Wallet client is not connected or does not have an account.');
        }

        const method = payload[1];
        if (method !== RPCMethod.AuthVerify) {
            throw new Error(
                `This EIP-712 signer is designed only for the '${RPCMethod.AuthVerify}' method, but received '${method}'.`,
            );
        }

        // Safely extract the challenge from the payload for an AuthVerify request.
        // The expected structure is `[id, 'auth_verify', [{ challenge: '...' }], ts]`
        const params = payload[2];
        if (!('challenge' in params) || typeof params.challenge !== 'string') {
            throw new Error('Invalid payload for AuthVerify: The challenge string is missing or malformed.');
        }

        // After the check, TypeScript knows `params` is an object with a `challenge` property of type string.
        const challengeUUID: string = params.challenge;

        const message: EIP712AuthMessage = {
            ...partialMessage,
            challenge: challengeUUID,
            wallet: address,
        };

        try {
            // The message for signTypedData must be a plain object.
            const untypedMessage: Record<string, unknown> = { ...message };

            // Sign with EIP-712
            const signature = await walletClient.signTypedData({
                account: walletClient.account!,
                domain,
                types: EIP712AuthTypes,
                primaryType: 'Policy',
                message: untypedMessage,
            });

            return signature;
        } catch (eip712Error) {
            const errorMessage = eip712Error instanceof Error ? eip712Error.message : String(eip712Error);
            console.error('EIP-712 signing failed:', errorMessage);
            throw new Error(`EIP-712 signing failed: ${errorMessage}`);
        }
    };
}

/**
 * Creates a message signer function that uses ECDSA signing with a provided private key.
 *
 * Note: for session key signing only, do not use this method with EOA keys.
 * @param privateKey - The private key to use for ECDSA signing.
 * @returns A MessageSigner function that signs the payload using ECDSA.
 */
export function createECDSAMessageSigner(privateKey: Hex): MessageSigner {
    return async (payload: RPCData): Promise<Hex> => {
        try {
            const message = toHex(JSON.stringify(payload, (_, v) => (typeof v === 'bigint' ? v.toString() : v)));

            return signRawECDSAMessage(message, privateKey);
        } catch (error) {
            console.error('ECDSA signing failed:', error);
            throw new Error(`ECDSA signing failed: ${error}`);
        }
    };
}
