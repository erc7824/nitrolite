import { Address, Hex, stringToHex } from 'viem';
import {
    AuthChallengeRPCParams,
    AuthVerifyRPCParams,
    ErrorRPCParams,
    NitroliteRPCMessage,
    RPCMethod,
    RPCParamsByMethod,
    RPCResponse,
    AppDefinition,
    GetConfigRPCParams,
    GetLedgerBalancesRPCParams,
    GetLedgerEntriesRPCParams,
    CreateApplicationRPCParams,
    SubmitStateRPCParams,
    CloseApplicationRPCParams,
    GetAppDefinitionRPCParams,
    GetAppSessionsRPCParams,
    ResizeChannelRPCParams,
    CloseChannelRPCParams,
    GetChannelsRPCParams,
    GetRPCHistoryRPCParams,
    GetAssetsRPCParams,
} from './types';

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

        return {
            method: parsed.res[1] as RPCMethod,
            requestId: parsed.res[0],
            timestamp: parsed.res[3],
            signatures: parsed.sig || [],
            params: parseRPCParameters(parsed.res[1], parsed.res[2]),
        };
    } catch (e) {
        throw new Error(`Failed to parse RPC response: ${e}`);
    }
}

function parseRPCParameters<M extends RPCMethod>(method: M, params: Array<any>): RPCParamsByMethod[M] {
    switch (method) {
        case RPCMethod.AuthChallenge:
            return {
                challengeMessage: extractRPCParameter<string>(params, 'challenge_message'),
            } as AuthChallengeRPCParams as RPCParamsByMethod[M];
        case RPCMethod.AuthVerify:
            return {
                address: extractRPCParameter<Address>(params, 'address'),
                jwtToken: extractRPCParameter<string>(params, 'jwt_token'),
                sessionKey: extractRPCParameter<string>(params, 'session_key') as Hex,
                success: extractRPCParameter<boolean>(params, 'success'),
            } as AuthVerifyRPCParams as RPCParamsByMethod[M];
        case RPCMethod.Error:
            return {
                error: extractRPCParameter<string>(params, 'error'),
            } as ErrorRPCParams as RPCParamsByMethod[M];
        case RPCMethod.GetConfig:
            return {
                broker_address: extractRPCParameter<Address>(params, 'broker_address'),
                networks: extractRPCParameter<GetConfigRPCParams['networks']>(params, 'networks'),
            } as GetConfigRPCParams as RPCParamsByMethod[M];
        case RPCMethod.GetLedgerBalances:
            return extractRPCParameter<GetLedgerBalancesRPCParams[]>(params, 'balances') as RPCParamsByMethod[M];
        case RPCMethod.GetLedgerEntries:
            return extractRPCParameter<GetLedgerEntriesRPCParams[]>(params, 'entries') as RPCParamsByMethod[M];
        case RPCMethod.CreateApplication:
            return {
                app_session_id: extractRPCParameter<Hex>(params, 'app_session_id'),
                version: extractRPCParameter<number>(params, 'version'),
                status: extractRPCParameter<string>(params, 'status'),
            } as CreateApplicationRPCParams as RPCParamsByMethod[M];
        case RPCMethod.SubmitState:
            return {
                app_session_id: extractRPCParameter<Hex>(params, 'app_session_id'),
                version: extractRPCParameter<number>(params, 'version'),
                status: extractRPCParameter<string>(params, 'status'),
            } as SubmitStateRPCParams as RPCParamsByMethod[M];
        case RPCMethod.CloseApplication:
            return {
                app_session_id: extractRPCParameter<Hex>(params, 'app_session_id'),
                version: extractRPCParameter<number>(params, 'version'),
                status: extractRPCParameter<string>(params, 'status'),
            } as CloseApplicationRPCParams as RPCParamsByMethod[M];
        case RPCMethod.GetAppDefinition:
            return extractRPCParameter<GetAppDefinitionRPCParams>(params, 'definition') as RPCParamsByMethod[M];
        case RPCMethod.GetAppSessions:
            return extractRPCParameter<GetAppSessionsRPCParams[]>(params, 'sessions') as RPCParamsByMethod[M];
        case RPCMethod.ResizeChannel:
            return {
                channel_id: extractRPCParameter<Hex>(params, 'channel_id'),
                intent: extractRPCParameter<number>(params, 'intent'),
                version: extractRPCParameter<number>(params, 'version'),
                state_data: extractRPCParameter<string>(params, 'state_data'),
                allocations: extractRPCParameter<ResizeChannelRPCParams['allocations']>(params, 'allocations'),
                state_hash: extractRPCParameter<string>(params, 'state_hash'),
                server_signature: extractRPCParameter<ResizeChannelRPCParams['server_signature']>(params, 'server_signature'),
            } as ResizeChannelRPCParams as RPCParamsByMethod[M];
        case RPCMethod.CloseChannel:
            return {
                channel_id: extractRPCParameter<Hex>(params, 'channel_id'),
                intent: extractRPCParameter<number>(params, 'intent'),
                version: extractRPCParameter<number>(params, 'version'),
                state_data: extractRPCParameter<string>(params, 'state_data'),
                allocations: extractRPCParameter<CloseChannelRPCParams['allocations']>(params, 'allocations'),
                state_hash: extractRPCParameter<string>(params, 'state_hash'),
                server_signature: extractRPCParameter<CloseChannelRPCParams['server_signature']>(params, 'server_signature'),
            } as CloseChannelRPCParams as RPCParamsByMethod[M];
        case RPCMethod.GetChannels:
            return extractRPCParameter<GetChannelsRPCParams[]>(params, 'channels') as RPCParamsByMethod[M];
        case RPCMethod.GetRPCHistory:
            return extractRPCParameter<GetRPCHistoryRPCParams[]>(params, 'history') as RPCParamsByMethod[M];
        case RPCMethod.GetAssets:
            return extractRPCParameter<GetAssetsRPCParams[]>(params, 'assets') as RPCParamsByMethod[M];
        default:
            throw new Error(`Unknown method for parameter extraction: ${method}`);
    }
}

function extractRPCParameter<T>(res: Array<any>, key: string): T {
    if (Array.isArray(res)) {
        return res[0]?.[key];
    }

    return res[key];
}
