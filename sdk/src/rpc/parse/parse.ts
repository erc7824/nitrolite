import { RPCResponse, RPCMethod } from '../types';
import { paramsParsers } from './index';
import { ParamsParser } from './common';

// Helper type to extract a specific response type from the main RPCResponse union.
type SpecificRPCResponse<T extends RPCMethod> = Extract<RPCResponse, { method: T }>;

/**
 * The core parsing engine. Parses any raw JSON RPC response.
 * This is the foundation for the specific parsers.
 */
export const parseAnyRPCResponse = (response: string): RPCResponse => {
    try {
        const parsed = JSON.parse(response);

        if (!Array.isArray(parsed.res) || parsed.res.length !== 4) {
            throw new Error('Invalid RPC response format');
        }

        const method = parsed.res[1] as RPCMethod;
        const parse = paramsParsers[method] as ParamsParser<unknown>;

        if (!parse) {
            throw new Error(`No parser found for method ${method}`);
        }

        const params = parse(parsed.res[2]);
        const responseObj = {
            method,
            requestId: parsed.res[0],
            timestamp: parsed.res[3],
            signatures: parsed.sig || [],
            params,
        } as RPCResponse;

        return responseObj;
    } catch (e) {
        throw new Error(`Failed to parse RPC response: ${e instanceof Error ? e.message : e}`);
    }
};

/**
 * INTERNAL: A generic parser that validates against an expected method.
 * This function acts as a type guard, ensuring the response matches what's expected.
 */
const _parseSpecificRPCResponse = <T extends RPCMethod>(
    response: string,
    expectedMethod: T,
): SpecificRPCResponse<T> => {
    const result = parseAnyRPCResponse(response);

    if (result.method !== expectedMethod) {
        throw new Error(`Expected RPC method to be '${expectedMethod}', but received '${result.method}'`);
    }

    return result as SpecificRPCResponse<T>;
};

/**
 * The main RPC response parsing utility.
 * This object provides a collection of type-safe parsers for each specific RPC method.
 * It offers the best developer experience, returning a fully typed object
 * without the need for manual type guards.
 *
 * @example
 * const result = rpcResponseParser.authChallenge(rawResponse);
 * // `result` is now fully typed as AuthChallengeResponse
 * console.log(result.params.challengeMessage);
 */
export const rpcResponseParser = {
    authChallenge: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.AuthChallenge),
    authVerify: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.AuthVerify),
    authRequest: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.AuthRequest),
    error: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.Error),
    getConfig: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetConfig),
    getLedgerBalances: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetLedgerBalances),
    getLedgerEntries: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetLedgerEntries),
    getLedgerTransactions: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetLedgerTransactions),
    createAppSession: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.CreateAppSession),
    submitAppState: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.SubmitAppState),
    closeAppSession: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.CloseAppSession),
    getAppDefinition: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetAppDefinition),
    getAppSessions: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetAppSessions),
    resizeChannel: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.ResizeChannel),
    closeChannel: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.CloseChannel),
    getChannels: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetChannels),
    getRPCHistory: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetRPCHistory),
    getAssets: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetAssets),
    assets: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.Assets),
    message: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.Message),
    balanceUpdate: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.BalanceUpdate),
    channelsUpdate: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.ChannelsUpdate),
    channelUpdate: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.ChannelUpdate),
    ping: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.Ping),
    pong: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.Pong),
    transfer: (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.Transfer),
};
