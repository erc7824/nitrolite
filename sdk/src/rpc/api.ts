import { Address, Hex, keccak256, stringToBytes, WalletClient } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';
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
    RequestData,
    RPCMethod,
    RPCChannelStatus,
    ResponsePayload,
} from './types';
import { NitroliteRPC } from './nitrolite';
import { generateRequestId, getCurrentTimestamp } from './utils';
import {
    CloseAppSessionRequestParams,
    CreateAppSessionRequestParams,
    SubmitAppStateRequestParams,
    ResizeChannelRequestParams,
    GetLedgerTransactionsFilters,
    GetLedgerTransactionsRequestParams,
    TransferRequestParams,
} from './types/request';

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
    const allowances = Object.values(params.allowances || {}).map((v) => [v.asset, v.amount]);
    const paramsArray = [
        params.wallet,
        params.participant,
        params.app_name,
        allowances,
        params.expire ?? '',
        params.scope ?? '',
        params.application ?? '',
    ];
    const request = NitroliteRPC.createRequest(requestId, RPCMethod.AuthRequest, paramsArray, timestamp);
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

    const request = NitroliteRPC.createRequest(requestId, RPCMethod.AuthVerify, [params], timestamp);
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
    const params = [{ challenge: challenge.params.challengeMessage }];
    const request = NitroliteRPC.createRequest(requestId, RPCMethod.AuthVerify, params, timestamp);
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
    const request = NitroliteRPC.createRequest(requestId, RPCMethod.AuthVerify, params, timestamp);
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
    const request = NitroliteRPC.createRequest(requestId, RPCMethod.Ping, [], timestamp);
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
    const request = NitroliteRPC.createRequest(requestId, RPCMethod.GetConfig, [], timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
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
    const request = NitroliteRPC.createRequest(requestId, RPCMethod.GetUserTag, [], timestamp);
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
    const request = NitroliteRPC.createRequest(requestId, RPCMethod.GetLedgerBalances, params, timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'get_ledger_entries' request.
 *
 * @param signer - The function to sign the request payload.
 * @param accountId - The account ID to get entries for.
 * @param asset - Optional asset symbol to filter entries.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createGetLedgerEntriesMessage(
    signer: MessageSigner,
    accountId: string,
    asset?: string,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const params = [
        {
            account_id: accountId,
            ...(asset ? { asset } : {}),
        },
    ];
    const request = NitroliteRPC.createRequest(requestId, RPCMethod.GetLedgerEntries, params, timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'get_ledger_transactions' request.
 *
 * @param signer - The function to sign the request payload.
 * @param accountId - The account ID to get transactions for.
 * @param filters - Optional filters to apply to the transactions.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createGetLedgerTransactionsMessage(
    signer: MessageSigner,
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

    const paramsObj: GetLedgerTransactionsRequestParams = {
        account_id: accountId,
        ...filteredParams,
    };

    const params = [paramsObj];
    const request = NitroliteRPC.createRequest(requestId, RPCMethod.GetLedgerTransactions, params, timestamp);
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
    const request = NitroliteRPC.createRequest(requestId, RPCMethod.GetAppDefinition, params, timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'get_app_sessions' request.
 *
 * @param signer - The function to sign the request payload.
 * @param participant - Participant address to filter sessions.
 * @param status - Optional status to filter sessions.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createGetAppSessionsMessage(
    signer: MessageSigner,
    participant: Address,
    status?: string,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const params = [
        {
            participant,
            ...(status ? { status } : {}),
        },
    ];
    const request = NitroliteRPC.createRequest(requestId, RPCMethod.GetAppSessions, params, timestamp);
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
    params: CreateAppSessionRequestParams[],
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest(requestId, RPCMethod.CreateAppSession, params, timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'submit_state' request.
 *
 * @param signer - The function to sign the request payload.
 * @param params - The specific parameters required by 'submit_state'. See {@link SubmitAppStateRequestParams} for details.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createSubmitAppStateMessage(
    signer: MessageSigner,
    params: SubmitAppStateRequestParams[],
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest(requestId, RPCMethod.SubmitAppState, params, timestamp);
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
    params: CloseAppSessionRequestParams[],
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest(requestId, RPCMethod.CloseAppSession, params, timestamp);
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
    const request = NitroliteRPC.createAppRequest(requestId, RPCMethod.Message, messageParams, timestamp, appSessionId);
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
    const request = NitroliteRPC.createRequest(requestId, RPCMethod.CloseChannel, params, timestamp);
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
    params: ResizeChannelRequestParams[],
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const request = NitroliteRPC.createRequest(requestId, RPCMethod.ResizeChannel, params, timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest, (_, value) => (typeof value === 'bigint' ? value.toString() : value));
}

/**
 * Creates the signed, stringified message body for a 'get_channels' request.
 *
 * @param signer - The function to sign the request payload.
 * @param participant - Optional participant address to filter channels.
 * @param status - Optional status to filter channels.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createGetChannelsMessage(
    signer: MessageSigner,
    participant?: Address,
    status?: RPCChannelStatus,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const params = [
        {
            ...(participant ? { participant } : {}),
            ...(status ? { status } : {}),
        },
    ];
    const request = NitroliteRPC.createRequest(requestId, RPCMethod.GetChannels, params, timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);
    return JSON.stringify(signedRequest);
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
    const request = NitroliteRPC.createRequest(requestId, RPCMethod.GetRPCHistory, [], timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
}

/**
 * Creates the signed, stringified message body for a 'get_assets' request.
 *
 * @param signer - The function to sign the request payload.
 * @param chainId - Optional chain ID to filter assets.
 * @param requestId - Optional request ID.
 * @param timestamp - Optional timestamp.
 * @returns A Promise resolving to the JSON string of the signed NitroliteRPCMessage.
 */
export async function createGetAssetsMessage(
    signer: MessageSigner,
    chainId?: number,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    const params = [
        {
            ...(chainId ? { chain_id: chainId } : {}),
        },
    ];
    const request = NitroliteRPC.createRequest(requestId, RPCMethod.GetAssets, params, timestamp);
    const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);

    return JSON.stringify(signedRequest);
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
    transferParams: TransferRequestParams,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): Promise<string> {
    // Validate that exactly one destination type is provided (XOR logic)
    const hasDestination = !!transferParams.destination;
    const hasDestinationTag = !!transferParams.destination_user_tag;

    if (hasDestination === hasDestinationTag) {
        throw new Error(
            hasDestination
                ? 'Cannot provide both destination and destination_user_tag'
                : 'Either destination or destination_user_tag must be provided',
        );
    }

    const params = [transferParams];
    const request = NitroliteRPC.createRequest(requestId, RPCMethod.Transfer, params, timestamp);
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
    return async (payload: RequestData | ResponsePayload): Promise<Hex> => {
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
        const firstParam = Array.isArray(params) ? params[0] : undefined;

        if (
            typeof firstParam !== 'object' ||
            firstParam === null ||
            !('challenge' in firstParam) ||
            typeof firstParam.challenge !== 'string'
        ) {
            throw new Error('Invalid payload for AuthVerify: The challenge string is missing or malformed.');
        }

        // After the check, TypeScript knows `firstParam` is an object with a `challenge` property of type string.
        const challengeUUID: string = firstParam.challenge;

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
    return async (payload: RequestData | ResponsePayload): Promise<Hex> => {
        try {
            const messageBytes = keccak256(stringToBytes(JSON.stringify(payload)));
            const flatSignature = await privateKeyToAccount(privateKey).sign({ hash: messageBytes });

            return flatSignature as Hex;
        } catch (error) {
            console.error('ECDSA signing failed:', error);
            throw new Error(`EIP-712  signing failed: ${error}`);
        }
    };
}
