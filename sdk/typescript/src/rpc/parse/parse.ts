import { RPCResponse, RPCMethod, RPCMessageType } from '../types';
import { paramsParsers } from './index';
import { ParamsParser } from './common';

type SpecificRPCResponse<T extends RPCMethod> = Extract<RPCResponse, { method: T }>;

/**
 * Parses any raw JSON RPC response from wire format to structured object.
 * Wire format: [type, requestId, method, params, timestamp]
 */
export const parseAnyRPCResponse = (response: string): RPCResponse => {
    try {
        const parsed = JSON.parse(response);

        if (!Array.isArray(parsed) || parsed.length !== 5) {
            throw new Error('Invalid RPC response format: expected 5-element array');
        }

        const [type, requestId, method, params, timestamp] = parsed;

        if (type !== RPCMessageType.Response && type !== RPCMessageType.Event) {
            throw new Error(`Invalid message type: expected Response (2) or Event (3), got ${type}`);
        }

        const parse = paramsParsers[method as RPCMethod] as ParamsParser<unknown>;

        if (!parse) {
            throw new Error(`No parser found for method ${method}`);
        }

        const parsedParams = parse(params);
        const responseObj = {
            method: method as RPCMethod,
            requestId,
            timestamp,
            params: parsedParams,
        } as RPCResponse;

        return responseObj;
    } catch (e) {
        throw new Error(`Failed to parse RPC response: ${e instanceof Error ? e.message : e}`);
    }
};

/**
 * Generic parser that validates against an expected method.
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

export const parseErrorResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.Error);

export const parseGetConfigResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetConfig);

export const parseGetAssetsResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetAssets);

export const parsePingResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.Ping);

export const parseRegisterResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.Register);

export const parseGetSessionKeysResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetSessionKeys);

export const parseRevokeSessionKeyResponse = (raw: string) =>
    _parseSpecificRPCResponse(raw, RPCMethod.RevokeSessionKey);

export const parseGetBalancesResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetBalances);

export const parseGetTransactionsResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetTransactions);

export const parseGetHomeChannelResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetHomeChannel);

export const parseGetEscrowChannelResponse = (raw: string) =>
    _parseSpecificRPCResponse(raw, RPCMethod.GetEscrowChannel);

export const parseGetChannelsResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetChannels);

export const parseGetLatestStateResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetLatestState);

export const parseGetStatesResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetStates);

export const parseCreateChannelResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.CreateChannel);

export const parseSubmitStateResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.SubmitState);

export const parseGetAppDefinitionResponse = (raw: string) =>
    _parseSpecificRPCResponse(raw, RPCMethod.GetAppDefinition);

export const parseGetAppSessionsResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.GetAppSessions);

export const parseCreateAppSessionResponse = (raw: string) =>
    _parseSpecificRPCResponse(raw, RPCMethod.CreateAppSession);

export const parseSubmitAppStateResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.SubmitAppState);

export const parseSubmitDepositStateResponse = (raw: string) =>
    _parseSpecificRPCResponse(raw, RPCMethod.SubmitDepositState);

export const parseRebalanceAppSessionsResponse = (raw: string) =>
    _parseSpecificRPCResponse(raw, RPCMethod.RebalanceAppSessions);

export const parseMessageResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.Message);

export const parseAssetsResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.Assets);

export const parseBalanceUpdateResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.BalanceUpdate);

export const parseTransferNotificationResponse = (raw: string) =>
    _parseSpecificRPCResponse(raw, RPCMethod.TransferNotification);

export const parseChannelUpdateResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.ChannelUpdate);

export const parseChannelsUpdateResponse = (raw: string) => _parseSpecificRPCResponse(raw, RPCMethod.ChannelsUpdate);

export const parseAppSessionUpdateResponse = (raw: string) =>
    _parseSpecificRPCResponse(raw, RPCMethod.AppSessionUpdate);
