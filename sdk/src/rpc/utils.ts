import { Address, Hex, stringToHex } from 'viem';
import { NitroliteRPCMessage, RPCMethod, RPCResponseParamsByMethod, RPCResponse } from './types';
import {
    AuthChallengeRPCResponseParams,
    AuthVerifyRPCResponseParams,
    ErrorRPCResponseParams,
    GetConfigRPCResponseParams,
    GetLedgerBalancesRPCResponseParams,
    GetLedgerEntriesRPCResponseParams,
    CreateAppSessionRPCResponseParams,
    SubmitStateRPCResponseParams,
    CloseAppSessionRPCResponseParams,
    GetAppDefinitionRPCResponseParams,
    GetAppSessionsRPCResponseParams,
    ResizeChannelRPCResponseParams,
    CloseChannelRPCResponseParams,
    GetChannelsRPCResponseParams,
    GetRPCHistoryRPCResponseParams,
    GetAssetsRPCResponseParams,
    AuthRequestRPCResponseParams,
    PingRPCResponseParams,
    PongRPCResponseParams,
    MessageRPCResponseParams,
    BalanceUpdateRPCResponseParams,
    ChannelUpdateRPCResponseParams,
} from './types/response';

/**
 * Get the current time in milliseconds
 *
 * @returns The current timestamp in milliseconds
 */
export function getCurrentTimestamp(): number {
    return Date.now();
}

/**
 * Generate a unique request ID
 *
 * @returns A unique request ID
 */
export function generateRequestId(): number {
    return Math.floor(Date.now() + Math.random() * 10000);
}

/**
 * Extract the request ID from a message
 *
 * @param message The message to extract from
 * @returns The request ID, or undefined if not found
 */
export function getRequestId(message: any): number | undefined {
    if (message.req) return message.req[0];
    if (message.res) return message.res[0];
    if (message.err) return message.err[0];
    return undefined;
}

/**
 * Extract the method name from a request or response
 *
 * @param message The message to extract from
 * @returns The method name, or undefined if not found
 */
export function getMethod(message: any): string | undefined {
    if (message.req) return message.req[1];
    if (message.res) return message.res[1];
    return undefined;
}

/**
 * Extract parameters from a request
 *
 * @param message The request message
 * @returns The parameters, or an empty array if not found
 */
export function getParams(message: any): any[] {
    if (message.req) return message.req[2] || [];
    return [];
}

/**
 * Extract result from a response
 *
 * @param message The response message
 * @returns The result, or an empty array if not found
 */
export function getResult(message: any): any[] {
    if (message.res) return message.res[2] || [];
    return [];
}

/**
 * Extract timestamp from a message
 *
 * @param message The message to extract from
 * @returns The timestamp, or undefined if not found
 */
export function getTimestamp(message: any): number | undefined {
    if (message.req) return message.req[3];
    if (message.res) return message.res[3];
    if (message.err) return message.err[3];
    return undefined;
}

/**
 * Extract error details from an error message
 *
 * @param message The error message
 * @returns The error details, or undefined if not found
 */
export function getError(message: any): { code: number; message: string } | undefined {
    if (message.err) {
        return {
            code: message.err[1],
            message: message.err[2],
        };
    }
    return undefined;
}

/**
 * Convert parameters or results to bytes format for smart contract interaction
 *
 * @param values Array of values to convert
 * @returns Array of hex strings
 */
export function toBytes(values: any[]): Hex[] {
    return values.map((v) => (typeof v === 'string' ? stringToHex(v) : stringToHex(JSON.stringify(v))));
}

/**
 * Validates that a response timestamp is greater than the request timestamp
 *
 * @param request The request message
 * @param response The response message
 * @returns True if the response timestamp is valid
 */
export function isValidResponseTimestamp(request: NitroliteRPCMessage, response: NitroliteRPCMessage): boolean {
    const requestTimestamp = getTimestamp(request);
    const responseTimestamp = getTimestamp(response);

    if (requestTimestamp === undefined || responseTimestamp === undefined) {
        return false;
    }

    return responseTimestamp > requestTimestamp;
}

/**
 * Validates that a response request ID matches the request
 *
 * @param request The request message
 * @param response The response message
 * @returns True if the response request ID is valid
 */
export function isValidResponseRequestId(request: NitroliteRPCMessage, response: NitroliteRPCMessage): boolean {
    const requestId = getRequestId(request);
    const responseId = getRequestId(response);

    if (requestId === undefined || responseId === undefined) {
        return false;
    }

    return responseId === requestId;
}

/**
 * Parses a raw RPC response string into a structured RPCResponse object
 * @param response The raw RPC response string to parse
 * @returns An RPCResponse object containing the parsed data
 */
export function parseRPCResponse(response: string): RPCResponse {
    // TODO: Add support for other rpc protocols besides websocket
    try {
        const parsed = JSON.parse(response);

        if (!Array.isArray(parsed.res) || parsed.res.length !== 4) {
            throw new Error('Invalid RPC response format');
        }

        const method = parsed.res[1] as keyof RPCResponseParamsByMethod;
        const responseObj = {
            method: method as RPCMethod,
            requestId: parsed.res[0],
            timestamp: parsed.res[3],
            signatures: parsed.sig || [],
            params: parseRPCParameters(method, parsed.res[2]),
        } as RPCResponse;

        return responseObj;
    } catch (e) {
        throw new Error(`Failed to parse RPC response: ${e}`);
    }
}

function parseRPCParameters<M extends keyof RPCResponseParamsByMethod>(
    method: M,
    params: Array<any>,
): RPCResponseParamsByMethod[M] {
    switch (method) {
        case RPCMethod.AuthChallenge:
            return {
                challengeMessage: extractRPCParameter<string>(params, 'challenge_message'),
            } as AuthChallengeRPCResponseParams as RPCResponseParamsByMethod[M];
        case RPCMethod.AuthVerify:
            return {
                address: extractRPCParameter<Address>(params, 'address'),
                jwtToken: extractRPCParameter<string>(params, 'jwt_token'),
                sessionKey: extractRPCParameter<string>(params, 'session_key') as Hex,
                success: extractRPCParameter<boolean>(params, 'success'),
            } as AuthVerifyRPCResponseParams as RPCResponseParamsByMethod[M];
        case RPCMethod.AuthRequest:
            return {
                challengeMessage: extractRPCParameter<string>(params, 'challenge_message'),
            } as AuthRequestRPCResponseParams as RPCResponseParamsByMethod[M];
        case RPCMethod.Error:
            return {
                error: extractRPCParameter<string>(params, 'error'),
            } as ErrorRPCResponseParams as RPCResponseParamsByMethod[M];
        case RPCMethod.GetConfig:
            return {
                broker_address: extractRPCParameter<Address>(params, 'broker_address'),
                networks: extractRPCParameter<GetConfigRPCResponseParams['networks']>(params, 'networks'),
            } as GetConfigRPCResponseParams as RPCResponseParamsByMethod[M];
        case RPCMethod.GetLedgerBalances:
            return params as RPCResponseParamsByMethod[M];
        case RPCMethod.GetLedgerEntries:
            return params as RPCResponseParamsByMethod[M];
        case RPCMethod.CreateAppSession:
            return {
                app_session_id: extractRPCParameter<Hex>(params, 'app_session_id'),
                version: extractRPCParameter<number>(params, 'version'),
                status: extractRPCParameter<string>(params, 'status'),
            } as CreateAppSessionRPCResponseParams as RPCResponseParamsByMethod[M];
        case RPCMethod.SubmitState:
            return {
                app_session_id: extractRPCParameter<Hex>(params, 'app_session_id'),
                version: extractRPCParameter<number>(params, 'version'),
                status: extractRPCParameter<string>(params, 'status'),
            } as SubmitStateRPCResponseParams as RPCResponseParamsByMethod[M];
        case RPCMethod.CloseAppSession:
            return {
                app_session_id: extractRPCParameter<Hex>(params, 'app_session_id'),
                version: extractRPCParameter<number>(params, 'version'),
                status: extractRPCParameter<string>(params, 'status'),
            } as CloseAppSessionRPCResponseParams as RPCResponseParamsByMethod[M];
        case RPCMethod.GetAppDefinition:
            return extractRPCParameter<GetAppDefinitionRPCResponseParams>(
                params,
                'definition',
            ) as RPCResponseParamsByMethod[M];
        case RPCMethod.GetAppSessions:
            return params as RPCResponseParamsByMethod[M];
        case RPCMethod.ResizeChannel:
            return {
                channel_id: extractRPCParameter<Hex>(params, 'channel_id'),
                intent: extractRPCParameter<number>(params, 'intent'),
                version: extractRPCParameter<number>(params, 'version'),
                state_data: extractRPCParameter<string>(params, 'state_data'),
                allocations: extractRPCParameter<ResizeChannelRPCResponseParams['allocations']>(params, 'allocations'),
                state_hash: extractRPCParameter<string>(params, 'state_hash'),
                server_signature: extractRPCParameter<ResizeChannelRPCResponseParams['server_signature']>(
                    params,
                    'server_signature',
                ),
            } as ResizeChannelRPCResponseParams as RPCResponseParamsByMethod[M];
        case RPCMethod.CloseChannel:
            return {
                channel_id: extractRPCParameter<Hex>(params, 'channel_id'),
                intent: extractRPCParameter<number>(params, 'intent'),
                version: extractRPCParameter<number>(params, 'version'),
                state_data: extractRPCParameter<string>(params, 'state_data'),
                allocations: extractRPCParameter<CloseChannelRPCResponseParams['allocations']>(params, 'allocations'),
                state_hash: extractRPCParameter<string>(params, 'state_hash'),
                server_signature: extractRPCParameter<CloseChannelRPCResponseParams['server_signature']>(
                    params,
                    'server_signature',
                ),
            } as CloseChannelRPCResponseParams as RPCResponseParamsByMethod[M];
        case RPCMethod.GetChannels:
            return params as RPCResponseParamsByMethod[M];
        case RPCMethod.GetRPCHistory:
            return params as RPCResponseParamsByMethod[M];
        case RPCMethod.GetAssets:
            return params as RPCResponseParamsByMethod[M];
        case RPCMethod.Ping:
            return {} as PingRPCResponseParams as RPCResponseParamsByMethod[M];
        case RPCMethod.Pong:
            return {} as PongRPCResponseParams as RPCResponseParamsByMethod[M];
        case RPCMethod.Message:
            return {} as MessageRPCResponseParams as RPCResponseParamsByMethod[M];
        case RPCMethod.BalanceUpdate:
            return params as RPCResponseParamsByMethod[M];
        case RPCMethod.ChannelsUpdate:
            return params as RPCResponseParamsByMethod[M];
        case RPCMethod.ChannelUpdate:
            return params as RPCResponseParamsByMethod[M];
        default:
            throw new Error(`Unsupported RPC method: ${method}`);
    }
}

function extractRPCParameter<T>(res: Array<any>, key: string): T {
    if (Array.isArray(res)) {
        return res[0]?.[key];
    }

    return res[key];
}
