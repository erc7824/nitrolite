import { Address, Hex } from 'viem';
import { RPCMethod, RequestID, Timestamp, AppDefinition, ChannelStatus } from './types';

/**
 * Represents a generic RPC message structure that includes common fields.
 * This interface is extended by specific RPC request and response types.
 */
interface GenericRPCMessage {
  requestId: RequestID;
  timestamp?: Timestamp;
  signatures?: Hex[];
}

/**
 * Represents the parameters for the 'auth_challenge' RPC method.
 */
export interface AuthChallengeRPCParams {
  challengeMessage: string;
}

/**
 * Represents the response structure for the 'auth_challenge' RPC method.
 */
export interface AuthChallengeRPCResponse extends GenericRPCMessage {
  method: RPCMethod.AuthChallenge;
  params: AuthChallengeRPCParams;
}

/**
 * Represents the parameters for the 'auth_verify' RPC method.
 */
export interface AuthVerifyRPCParams {
  address: Address;
  jwtToken: string;
  sessionKey: Address;
  success: boolean;
}

/**
 * Represents the parameters for the 'error' RPC method.
 */
export interface ErrorRPCParams {
  error: string;
}

/**
 * Represents the parameters for the 'get_config' RPC method.
 */
export interface GetConfigRPCParams {
  broker_address: Address;
  networks: {
    name: string;
    chain_id: number;
    custody_address: Address;
    adjudicator_address: Address;
  }[];
}

/**
 * Represents the parameters for the 'get_ledger_balances' RPC method.
 */
export interface GetLedgerBalancesRPCParams {
  asset: string;
  amount: string;
}

/**
 * Represents the parameters for the 'get_ledger_entries' RPC method.
 */
export interface GetLedgerEntriesRPCParams {
  id: number;
  account_id: string;
  account_type: string;
  asset: string;
  participant: Address;
  credit: string;
  debit: string;
  created_at: string;
}

/**
 * Represents the parameters for the 'create_app_session' RPC method.
 */
export interface CreateAppSessionRPCParams {
  app_session_id: Hex;
  version: number;
  status: ChannelStatus;
}

/**
 * Represents the parameters for the 'submit_state' RPC method.
 */
export interface SubmitStateRPCParams {
  app_session_id: Hex;
  version: number;
  status: ChannelStatus;
}

/**
 * Represents the parameters for the 'close_app_session' RPC method.
 */
export interface CloseAppSessionRPCParams {
  app_session_id: Hex;
  version: number;
  status: ChannelStatus;
}

/**
 * Represents the parameters for the 'get_app_definition' RPC method.
 */
export interface GetAppDefinitionRPCParams extends AppDefinition { }

/**
 * Represents the parameters for the 'get_app_sessions' RPC method.
 */
export interface GetAppSessionsRPCParams {
  app_session_id: Hex;
  status: ChannelStatus;
  participants: Address[];
  protocol: string;
  challenge: number;
  weights: number[];
  quorum: number;
  version: number;
  nonce: number;
  created_at: string;
  updated_at: string;
}

/**
 * Represents the parameters for the 'resize_channel' RPC method.
 */
export interface ResizeChannelRPCParams {
  channel_id: Hex;
  state_data: string;
  intent: number;
  version: number;
  allocations: {
    destination: Address;
    token: Address;
    amount: string;
  }[];
  state_hash: string;
  server_signature: {
    v: string;
    r: string;
    s: string;
  };
}

/**
 * Represents the parameters for the 'close_channel' RPC method.
 */
export interface CloseChannelRPCParams {
  channel_id: Hex;
  intent: number;
  version: number;
  state_data: string;
  allocations: {
    destination: Address;
    token: Address;
    amount: string;
  }[];
  state_hash: string;
  server_signature: {
    v: string;
    r: string;
    s: string;
  };
}

/**
 * Represents the parameters for the 'get_channels' RPC method.
 */
export interface GetChannelsRPCParams {
  channel_id: Hex;
  participant: Address;
  status: ChannelStatus;
  token: Address;
  wallet: Address;
  amount: string;
  chain_id: number;
  adjudicator: Address;
  challenge: number;
  nonce: number;
  version: number;
  created_at: string;
  updated_at: string;
}

/**
 * Represents the parameters for the 'get_rpc_history' RPC method.
 */
export interface GetRPCHistoryRPCParams {
  id: number;
  sender: Address;
  req_id: number;
  method: string;
  params: string;
  timestamp: number;
  req_sig: Hex[];
  res_sig: Hex[];
  response: string;
}

/**
 * Represents the parameters for the 'get_assets' RPC method.
 */
export interface GetAssetsRPCParams {
  token: Address;
  chain_id: number;
  symbol: string;
  decimals: number;
}

/**
 * Represents the response structure for an error response.
 */
export interface ErrorRPCResponse extends GenericRPCMessage {
  method: RPCMethod.Error;
  params: ErrorRPCParams;
}

/**
 * Represents the response structure for the 'get_config' RPC method.
 */
export interface GetConfigRPCResponse extends GenericRPCMessage {
  method: RPCMethod.GetConfig;
  params: GetConfigRPCParams;
}

/**
 * Represents the response structure for the 'get_ledger_balances' RPC method.
 */
export interface GetLedgerBalancesRPCResponse extends GenericRPCMessage {
  method: RPCMethod.GetLedgerBalances;
  params: GetLedgerBalancesRPCParams[];
}

/**
 * Represents the response structure for the 'get_ledger_entries' RPC method.
 */
export interface GetLedgerEntriesRPCResponse extends GenericRPCMessage {
  method: RPCMethod.GetLedgerEntries;
  params: GetLedgerEntriesRPCParams[];
}

/**
 * Represents the response structure for the 'create_app_session' RPC method.
 */
export interface CreateAppSessionRPCResponse extends GenericRPCMessage {
  method: RPCMethod.CreateAppSession;
  params: CreateAppSessionRPCParams;
}

/**
 * Represents the response structure for the 'submit_state' RPC method.
 */
export interface SubmitStateRPCResponse extends GenericRPCMessage {
  method: RPCMethod.SubmitState;
  params: SubmitStateRPCParams;
}

/**
 * Represents the response structure for the 'close_app_session' RPC method.
 */
export interface CloseAppSessionRPCResponse extends GenericRPCMessage {
  method: RPCMethod.CloseAppSession;
  params: CloseAppSessionRPCParams;
}

/**
 * Represents the response structure for the 'get_app_definition' RPC method.
 */
export interface GetAppDefinitionRPCResponse extends GenericRPCMessage {
  method: RPCMethod.GetAppDefinition;
  params: GetAppDefinitionRPCParams;
}

/**
 * Represents the response structure for the 'get_app_sessions' RPC method.
 */
export interface GetAppSessionsRPCResponse extends GenericRPCMessage {
  method: RPCMethod.GetAppSessions;
  params: GetAppSessionsRPCParams[];
}

/**
 * Represents the response structure for the 'resize_channel' RPC method.
 */
export interface ResizeChannelRPCResponse extends GenericRPCMessage {
  method: RPCMethod.ResizeChannel;
  params: ResizeChannelRPCParams;
}

/**
 * Represents the response structure for the 'close_channel' RPC method.
 */
export interface CloseChannelRPCResponse extends GenericRPCMessage {
  method: RPCMethod.CloseChannel;
  params: CloseChannelRPCParams;
}

/**
 * Represents the response structure for the 'get_channels' RPC method.
 */
export interface GetChannelsRPCResponse extends GenericRPCMessage {
  method: RPCMethod.GetChannels;
  params: GetChannelsRPCParams[];
}

/**
 * Represents the response structure for the 'get_rpc_history' RPC method.
 */
export interface GetRPCHistoryRPCResponse extends GenericRPCMessage {
  method: RPCMethod.GetRPCHistory;
  params: GetRPCHistoryRPCParams[];
}

/**
 * Represents the response structure for the 'get_assets' RPC method.
 */
export interface GetAssetsRPCResponse extends GenericRPCMessage {
  method: RPCMethod.GetAssets;
  params: GetAssetsRPCParams[];
}

/**
 * Represents the response structure for the 'auth_verify' RPC method.
 */
export interface AuthVerifyRPCResponse extends GenericRPCMessage {
  method: RPCMethod.AuthVerify;
  params: AuthVerifyRPCParams;
}

/**
 * Union type for all possible RPC response types.
 * This allows for type-safe handling of different response structures.
 */
export type RPCResponse =
  | AuthChallengeRPCResponse
  | AuthVerifyRPCResponse
  | ErrorRPCResponse
  | GetConfigRPCResponse
  | GetLedgerBalancesRPCResponse
  | GetLedgerEntriesRPCResponse
  | CreateAppSessionRPCResponse
  | SubmitStateRPCResponse
  | CloseAppSessionRPCResponse
  | GetAppDefinitionRPCResponse
  | GetAppSessionsRPCResponse
  | ResizeChannelRPCResponse
  | CloseChannelRPCResponse
  | GetChannelsRPCResponse
  | GetRPCHistoryRPCResponse
  | GetAssetsRPCResponse;

/**
 * Maps RPC methods to their corresponding parameter types.
 */
export type RPCParamsByMethod = {
  [RPCMethod.AuthChallenge]: AuthChallengeRPCParams;
  [RPCMethod.AuthVerify]: AuthVerifyRPCParams;
  [RPCMethod.Error]: ErrorRPCParams;
  [RPCMethod.GetConfig]: GetConfigRPCParams;
  [RPCMethod.GetLedgerBalances]: GetLedgerBalancesRPCParams[];
  [RPCMethod.GetLedgerEntries]: GetLedgerEntriesRPCParams[];
  [RPCMethod.CreateAppSession]: CreateAppSessionRPCParams;
  [RPCMethod.SubmitState]: SubmitStateRPCParams;
  [RPCMethod.CloseAppSession]: CloseAppSessionRPCParams;
  [RPCMethod.GetAppDefinition]: GetAppDefinitionRPCParams;
  [RPCMethod.GetAppSessions]: GetAppSessionsRPCParams[];
  [RPCMethod.ResizeChannel]: ResizeChannelRPCParams;
  [RPCMethod.CloseChannel]: CloseChannelRPCParams;
  [RPCMethod.GetChannels]: GetChannelsRPCParams[];
  [RPCMethod.GetRPCHistory]: GetRPCHistoryRPCParams[];
  [RPCMethod.GetAssets]: GetAssetsRPCParams[];
};
