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

/** Parses `auth_challenge` response */
export const parseAuthChallengeResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.AuthChallenge);

/** Parses `auth_verify` response */
export const parseAuthVerifyResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.AuthVerify);

/** Parses `auth_request` response */
export const parseAuthRequestResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.AuthRequest);

/** Parses `error` response */
export const parseErrorResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.Error);

/** Parses `get_config` response */
export const parseGetConfigResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetConfig);

/** Parses `get_ledger_balances` response */
export const parseGetLedgerBalancesResponse = (raw: string) =>
    _parseSpecificRPCResponse(raw, RPCMethod.GetLedgerBalances);

/** Parses `get_ledger_entries` response */
export const parseGetLedgerEntriesResponse = (raw: string) =>
    _parseSpecificRPCResponse(raw, RPCMethod.GetLedgerEntries);

/** Parses `get_ledger_transactions` response */
export const parseGetLedgerTransactionsResponse = (raw: string) =>
    _parseSpecificRPCResponse(raw, RPCMethod.GetLedgerTransactions);

/** Parses `get_user_tag` response */
export const parseGetUserTagResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetUserTag);

/** Parses `get_session_keys` response */
export const parseGetSessionKeysResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetSessionKeys)

/** Parses `create_app_session` response */
export const parseCreateAppSessionResponse = (raw: string) =>
    _parseSpecificRPCResponse(raw, RPCMethod.CreateAppSession);

/** Parses `submit_app_state` response */
export const parseSubmitAppStateResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.SubmitAppState);

/** Parses `close_app_session` response */
export const parseCloseAppSessionResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.CloseAppSession);

/** Parses `get_app_definition` response */
export const parseGetAppDefinitionResponse = (raw: string) =>
    _parseSpecificRPCResponse(raw, RPCMethod.GetAppDefinition);

/** Parses `get_app_sessions` response */
export const parseGetAppSessionsResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetAppSessions);

/** Parses `create_channel` response */
export const parseCreateChannelResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.CreateChannel);

/** Parses `resize_channel` response */
export const parseResizeChannelResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.ResizeChannel);

/** Parses `close_channel` response */
export const parseCloseChannelResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.CloseChannel);

/** Parses `get_channels` response */
export const parseGetChannelsResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetChannels);

/** Parses `get_rpc_history` response */
export const parseGetRPCHistoryResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetRPCHistory);

/** Parses `get_assets` response */
export const parseGetAssetsResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetAssets);

/** Parses `assets` response */
export const parseAssetsResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.Assets);

/** Parses `message` response */
export const parseMessageResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.Message);

/** Parses `bu` response */
export const parseBalanceUpdateResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.BalanceUpdate);

/** Parses `channels` response */
export const parseChannelsUpdateResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.ChannelsUpdate);

/** Parses `cu` response */
export const parseChannelUpdateResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.ChannelUpdate);

/** Parses `ping` response */
export const parsePingResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.Ping);

/** Parses `pong` response */
export const parsePongResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.Pong);

/** Parses `transfer` response */
export const parseTransferResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.Transfer);

/** Parses `cleanup_session_key_cache` response */
export const parseCleanupSessionKeyCacheResponse = (raw: string) =>
    _parseSpecificRPCResponse(raw, RPCMethod.CleanupSessionKeyCache);

/** Parses `tr` response */
export const parseTransferNotificationResponse = (raw: string) =>
    _parseSpecificRPCResponse(raw, RPCMethod.TransferNotification);

/** Parses `asu` response */
export const parseAppSessionUpdateResponse = (raw: string) =>
    _parseSpecificRPCResponse(raw, RPCMethod.AppSessionUpdate);
