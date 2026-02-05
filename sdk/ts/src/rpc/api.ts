/**
 * RPC API request and response definitions
 * This file implements the API request and response definitions
 * with versioned types organized by functional groups
 */

import { Address } from 'viem';
import {
  ChannelV1,
  ChannelDefinitionV1,
  StateV1,
  BalanceEntryV1,
  TransactionV1,
  PaginationParamsV1,
  PaginationMetadataV1,
  AssetV1,
  BlockchainInfoV1,
} from './types';
import {
  AppDefinitionV1,
  AppStateUpdateV1,
  AppSessionInfoV1,
  AppAllocationV1,
  AssetAllowanceV1,
  SessionKeyV1,
  SignedAppStateUpdateV1,
} from '../app/types';
import { TransactionType } from '../core/types';

// ============================================================================
// Channels Group - V1 API
// ============================================================================

export interface ChannelsV1GetHomeChannelRequest {
  /** User's wallet address */
  wallet: Address;
  /** Asset symbol */
  asset: string;
}

export interface ChannelsV1GetHomeChannelResponse {
  /** On-chain channel information */
  channel: ChannelV1;
}

export interface ChannelsV1GetEscrowChannelRequest {
  /** Escrow channel ID */
  escrowChannelId: string;
}

export interface ChannelsV1GetEscrowChannelResponse {
  /** On-chain channel information */
  channel: ChannelV1;
}

export interface ChannelsV1GetChannelsRequest {
  /** User's wallet address */
  wallet: Address;
  /** Status filter */
  status?: string;
  /** Asset filter */
  asset?: string;
  /** Pagination parameters */
  pagination?: PaginationParamsV1;
}

export interface ChannelsV1GetChannelsResponse {
  /** List of channels */
  channels: ChannelV1[];
  /** Pagination information */
  metadata: PaginationMetadataV1;
}

export interface ChannelsV1GetLatestStateRequest {
  /** User's wallet address */
  wallet: Address;
  /** Asset symbol */
  asset: string;
  /** Enable to get the latest signed state */
  onlySigned: boolean;
}

export interface ChannelsV1GetLatestStateResponse {
  /** Current state of the user */
  state: StateV1;
}

export interface ChannelsV1GetStatesRequest {
  /** User's wallet address */
  wallet: Address;
  /** Asset symbol */
  asset: string;
  /** User epoch index filter */
  epoch?: bigint; // uint64
  /** Home/Escrow Channel ID filter */
  channelId?: string;
  /** Return only signed states */
  onlySigned: boolean;
  /** Pagination parameters */
  pagination?: PaginationParamsV1;
}

export interface ChannelsV1GetStatesResponse {
  /** List of states */
  states: StateV1[];
  /** Pagination information */
  metadata: PaginationMetadataV1;
}

export interface ChannelsV1RequestCreationRequest {
  /** State to be submitted */
  state: StateV1;
  /** Definition of the channel to be created */
  channelDefinition: ChannelDefinitionV1;
}

export interface ChannelsV1RequestCreationResponse {
  /** Node's signature for the state */
  signature: string;
}

export interface ChannelsV1SubmitStateRequest {
  /** State to be submitted */
  state: StateV1;
}

export interface ChannelsV1SubmitStateResponse {
  /** Node's signature for the state */
  signature: string;
}

export interface ChannelsV1HomeChannelCreatedEvent {
  /** Created home channel information */
  channel: ChannelV1;
  /** Initial state of the home channel */
  initialState: StateV1;
}

// ============================================================================
// App Sessions Group - V1 API
// ============================================================================

export interface AppSessionsV1SubmitDepositStateRequest {
  /** Application session state update to be submitted */
  appStateUpdate: AppStateUpdateV1;
  /** List of participant signatures for the app state update */
  quorumSigs: string[];
  /** User state */
  userState: StateV1;
}

export interface AppSessionsV1SubmitDepositStateResponse {
  /** Node's signature for the deposit state */
  stateNodeSig: string;
}

export interface AppSessionsV1SubmitAppStateRequest {
  /** Application session state update to be submitted */
  appStateUpdate: AppStateUpdateV1;
  /** Signature quorum for the application session */
  quorumSigs: string[];
}

export interface AppSessionsV1SubmitAppStateResponse {}

export interface AppSessionsV1RebalanceAppSessionsRequest {
  /** List of signed application session state updates */
  signedUpdates: SignedAppStateUpdateV1[];
}

export interface AppSessionsV1RebalanceAppSessionsResponse {
  /** Unique identifier for this rebalancing operation */
  batchId: string;
}

export interface AppSessionsV1GetAppDefinitionRequest {
  /** Application session ID */
  appSessionId: string;
}

export interface AppSessionsV1GetAppDefinitionResponse {
  /** Application definition */
  definition: AppDefinitionV1;
}

export interface AppSessionsV1GetAppSessionsRequest {
  /** Application session ID filter */
  appSessionId?: string;
  /** Participant wallet address filter */
  participant?: Address;
  /** Status filter (open/closed) */
  status?: string;
  /** Pagination parameters */
  pagination?: PaginationParamsV1;
}

export interface AppSessionsV1GetAppSessionsResponse {
  /** List of application sessions */
  appSessions: AppSessionInfoV1[];
  /** Pagination information */
  metadata: PaginationMetadataV1;
}

export interface AppSessionsV1CreateAppSessionRequest {
  /** Application definition including participants and quorum */
  definition: AppDefinitionV1;
  /** Optional JSON stringified session data */
  sessionData: string;
  /** Participant signatures for the app session creation */
  quorumSigs?: string[];
}

export interface AppSessionsV1CreateAppSessionResponse {
  /** Created application session ID */
  appSessionId: string;
  /** Initial version of the session */
  version: string;
  /** Status of the session */
  status: string;
}

export interface AppSessionsV1CloseAppSessionRequest {
  /** Application session ID to close */
  appSessionId: string;
  /** Final asset allocations when closing the session */
  allocations: AppAllocationV1[];
  /** Optional final JSON stringified session data */
  sessionData?: string;
}

export interface AppSessionsV1CloseAppSessionResponse {
  /** Closed application session ID */
  appSessionId: string;
  /** Final version of the session */
  version: string;
  /** Status of the session (closed) */
  status: string;
}

// ============================================================================
// Session Keys Group - V1 API
// ============================================================================

export interface SessionKeysV1RegisterRequest {
  /** User wallet address */
  address: Address;
  /** Session key address for delegation */
  sessionKey?: string;
  /** Application name for analytics */
  application?: string;
  /** Asset allowances for the session */
  allowances?: AssetAllowanceV1[];
  /** Permission scope */
  scope?: string;
  /** Session expiration timestamp */
  expiresAt?: bigint; // uint64
}

export interface SessionKeysV1RegisterResponse {}

export interface SessionKeysV1RevokeSessionKeyRequest {
  /** Address of the session key to revoke */
  sessionKey: string;
}

export interface SessionKeysV1RevokeSessionKeyResponse {
  /** Address of the revoked session key */
  sessionKey: string;
}

export interface SessionKeysV1GetSessionKeysRequest {
  /** User's wallet address */
  wallet: Address;
}

export interface SessionKeysV1GetSessionKeysResponse {
  /** List of active session keys */
  sessionKeys: SessionKeyV1[];
}

// ============================================================================
// User Group - V1 API
// ============================================================================

export interface UserV1GetBalancesRequest {
  /** User's wallet address */
  wallet: Address;
}

export interface UserV1GetBalancesResponse {
  /** List of asset balances */
  balances: BalanceEntryV1[];
}

export interface UserV1GetTransactionsRequest {
  /** User's wallet address */
  wallet: Address;
  /** Asset symbol filter */
  asset?: string;
  /** Transaction type filter */
  txType?: TransactionType;
  /** Pagination parameters */
  pagination?: PaginationParamsV1;
  /** Start time filter (Unix timestamp) */
  fromTime?: bigint; // uint64
  /** End time filter (Unix timestamp) */
  toTime?: bigint; // uint64
}

export interface UserV1GetTransactionsResponse {
  /** List of transactions */
  transactions: TransactionV1[];
  /** Pagination information */
  metadata: PaginationMetadataV1;
}

// ============================================================================
// Node Group - V1 API
// ============================================================================

export interface NodeV1PingRequest {}

export interface NodeV1PingResponse {}

export interface NodeV1GetConfigRequest {}

export interface NodeV1GetConfigResponse {
  /** Node wallet address */
  node_address: Address;
  /** Node software version */
  node_version: string;
  /** List of supported networks */
  blockchains: BlockchainInfoV1[];
}

export interface NodeV1GetAssetsRequest {
  /** Blockchain network ID filter */
  blockchainId?: bigint; // uint64
}

export interface NodeV1GetAssetsResponse {
  /** List of supported assets */
  assets: AssetV1[];
}
