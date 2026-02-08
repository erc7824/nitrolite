import { Address, Hex } from 'viem';
import {
    RPCMethod,
    RequestID,
    Timestamp,
    StateSigner,
    RPCState,
    RPCAppDefinition,
    RPCAppSessionAllocation,
    RPCAppStateIntent,
    RPCAppStateUpdate,
    RPCSignedAppStateUpdate,
    GetTransactionsOptions,
    GetChannelsOptions,
    GetStatesOptions,
    GetAppSessionsOptions,
    RegisterOptions,
} from './types';
import { NitroliteRPC } from './nitrolite';
import { generateRequestId, getCurrentTimestamp } from './utils';

/**
 * API v1 Message Builders
 *
 * This file contains functions to create RPC request messages for the Clearnode API v1.
 * All messages are created in the wire format: [type, requestId, method, params, timestamp]
 *
 * Methods are organized by API group:
 * - Node: Configuration and system information
 * - User: User account information
 * - Channels: Channel and state management
 * - App Sessions: Application session management
 * - Session Keys: Session key management
 */

/**
 * Creates a 'node.v1.ping' request message.
 * Health check to verify connection is alive.
 *
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createPingMessage(
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
 * Creates a 'node.v1.get_config' request message.
 * Get node configuration and supported blockchains.
 *
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createGetConfigMessage(
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
 * Creates a 'node.v1.get_assets' request message.
 * Get list of supported assets, optionally filtered by blockchain.
 *
 * @param chainId - Optional blockchain ID to filter assets
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createGetAssetsMessage(
    chainId?: number,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const params = chainId !== undefined ? { chain_id: chainId } : {};
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetAssets,
        params,
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}

/**
 * Creates a 'user.v1.get_balances' request message.
 * Get user's asset balances.
 *
 * @param wallet - User's wallet address
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createGetBalancesMessage(
    wallet: Address,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetBalances,
        params: { wallet },
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}

/**
 * Creates a 'user.v1.get_transactions' request message.
 * Get user's transaction history with optional filtering and pagination.
 *
 * @param wallet - User's wallet address
 * @param options - Optional filters and pagination
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createGetTransactionsMessage(
    wallet: Address,
    options?: GetTransactionsOptions,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const params = {
        wallet,
        ...(options?.asset && { asset: options.asset }),
        ...(options?.tx_type && { tx_type: options.tx_type }),
        ...(options?.from_time && { from_time: options.from_time }),
        ...(options?.to_time && { to_time: options.to_time }),
        ...(options?.pagination && { pagination: options.pagination }),
    };

    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetTransactions,
        params,
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}

/**
 * Creates a 'channels.v1.get_home_channel' request message.
 * Retrieve current on-chain home channel information.
 *
 * @param wallet - User's wallet address
 * @param asset - Asset symbol
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createGetHomeChannelMessage(
    wallet: Address,
    asset: string,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetHomeChannel,
        params: { wallet, asset },
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}

/**
 * Creates a 'channels.v1.get_escrow_channel' request message.
 * Retrieve current on-chain escrow channel information.
 *
 * @param escrowChannelId - Escrow channel ID
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createGetEscrowChannelMessage(
    escrowChannelId: string,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetEscrowChannel,
        params: { escrow_channel_id: escrowChannelId },
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}

/**
 * Creates a 'channels.v1.get_channels' request message.
 * Retrieve all channels for a user with optional filtering.
 *
 * @param wallet - User's wallet address
 * @param options - Optional filters and pagination
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createGetChannelsMessage(
    wallet: Address,
    options?: GetChannelsOptions,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const params = {
        wallet,
        ...(options?.asset && { asset: options.asset }),
        ...(options?.status && { status: options.status }),
        ...(options?.pagination && { pagination: options.pagination }),
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
 * Creates a 'channels.v1.get_latest_state' request message.
 * Retrieve the current state of the user stored on the Node.
 *
 * @param wallet - User's wallet address
 * @param asset - Asset symbol
 * @param onlySigned - Get only signed states
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createGetLatestStateMessage(
    wallet: Address,
    asset: string,
    onlySigned: boolean,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetLatestState,
        params: { wallet, asset, only_signed: onlySigned },
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}

/**
 * Creates a 'channels.v1.get_states' request message.
 * Retrieve state history for a user with optional filtering.
 *
 * @param wallet - User's wallet address
 * @param asset - Asset symbol
 * @param onlySigned - Return only signed states
 * @param options - Optional filters and pagination
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createGetStatesMessage(
    wallet: Address,
    asset: string,
    onlySigned: boolean,
    options?: GetStatesOptions,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const params = {
        wallet,
        asset,
        only_signed: onlySigned,
        ...(options?.epoch !== undefined && { epoch: options.epoch }),
        ...(options?.channel_id && { channel_id: options.channel_id }),
        ...(options?.pagination && { pagination: options.pagination }),
    };

    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetStates,
        params,
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}

/**
 * Creates a 'channels.v1.request_creation' request message.
 * Request the creation of a channel from Node.
 *
 * @param state - The state to be submitted
 * @param channelDefinition - Definition of the channel to be created
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createCreateChannelMessage(
    state: RPCState,
    channelDefinition: { nonce: number; challenge: number },
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.CreateChannel,
        params: {
            state,
            channel_definition: channelDefinition,
        },
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}

/**
 * Creates a 'channels.v1.submit_state' request message.
 * Submit a cross-chain state transition.
 *
 * @param state - The state to be submitted
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createSubmitStateMessage(
    state: RPCState,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.SubmitState,
        params: { state },
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}

// App Session Methods (app_sessions.v1.*)

/**
 * Creates a 'app_sessions.v1.get_app_definition' request message.
 * Retrieve the application definition for a specific app session.
 *
 * @param appSessionId - The unique identifier of the application session
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createGetAppDefinitionMessage(
    appSessionId: Hex,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetAppDefinition,
        params: { app_session_id: appSessionId },
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}

/**
 * Creates a 'app_sessions.v1.get_app_sessions' request message.
 * Get list of application sessions with optional filtering and pagination.
 *
 * @param options - Optional filters and pagination
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createGetAppSessionsMessage(
    options?: GetAppSessionsOptions,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const params = {
        ...(options?.app_session_id && { app_session_id: options.app_session_id }),
        ...(options?.participant && { participant: options.participant }),
        ...(options?.status && { status: options.status }),
        ...(options?.pagination && { pagination: options.pagination }),
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
 * Creates a 'app_sessions.v1.create_app_session' request message.
 * Create a new application session.
 *
 * @param definition - Application definition including participants and quorum
 * @param quorumSigs - App Session creation signatures
 * @param sessionData - Optional JSON stringified session data
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createCreateAppSessionMessage(
    definition: RPCAppDefinition,
    quorumSigs: Hex[],
    sessionData?: string,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const params = {
        definition,
        quorum_sigs: quorumSigs,
        ...(sessionData !== undefined && { session_data: sessionData }),
    };

    const request = NitroliteRPC.createRequest({
        method: RPCMethod.CreateAppSession,
        params,
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}

/**
 * Creates a 'app_sessions.v1.submit_app_state' request message.
 * Submit an application session state update.
 *
 * @param appStateUpdate - The application session state update
 * @param quorumSigs - Quorum signatures from participants
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createSubmitAppStateMessage(
    appStateUpdate: {
        app_session_id: Hex;
        intent: RPCAppStateIntent;
        version: number;
        allocations: RPCAppSessionAllocation[];
        session_data?: string;
    },
    quorumSigs: Hex[],
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.SubmitAppState,
        params: {
            app_state_update: appStateUpdate,
            quorum_sigs: quorumSigs,
        },
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}

/**
 * Creates a 'app_sessions.v1.submit_deposit_state' request message.
 * Submit an application session deposit state update.
 *
 * @param appStateUpdate - The application session state update
 * @param quorumSigs - Quorum signatures
 * @param userState - The user state associated with the deposit
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createSubmitDepositStateMessage(
    appStateUpdate: RPCAppStateUpdate,
    quorumSigs: Hex[],
    userState: RPCState,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.SubmitDepositState,
        params: {
            app_state_update: appStateUpdate,
            quorum_sigs: quorumSigs,
            user_state: userState,
        },
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}

/**
 * Creates a 'app_sessions.v1.rebalance_app_sessions' request message.
 * Atomically rebalance multiple application sessions.
 *
 * @param signedUpdates - List of signed updates with intent 'rebalance'
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createRebalanceAppSessionsMessage(
    signedUpdates: RPCSignedAppStateUpdate[],
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.RebalanceAppSessions,
        params: { signed_updates: signedUpdates },
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}

// Session Key Methods (session_keys.v1.*)

/**
 * Creates a 'session_keys.v1.register' request message.
 * Register a new session key with allowances and expiration.
 *
 * @param address - User wallet address
 * @param options - Optional session key configuration
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createRegisterMessage(
    address: Address,
    options?: RegisterOptions,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const params = {
        address,
        ...(options?.session_key && { session_key: options.session_key }),
        ...(options?.application && { application: options.application }),
        ...(options?.allowances && { allowances: options.allowances }),
        ...(options?.scope && { scope: options.scope }),
        ...(options?.expires_at && { expires_at: options.expires_at }),
    };

    const request = NitroliteRPC.createRequest({
        method: RPCMethod.Register,
        params,
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}

/**
 * Creates a 'session_keys.v1.get_session_keys' request message.
 * Get all active session keys for a user.
 *
 * @param wallet - User's wallet address
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createGetSessionKeysMessage(
    wallet: Address,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.GetSessionKeys,
        params: { wallet },
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}

/**
 * Creates a 'session_keys.v1.revoke_session_key' request message.
 * Revoke an existing session key.
 *
 * @param sessionKey - The session key address to revoke
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createRevokeSessionKeyMessage(
    sessionKey: Address,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.RevokeSessionKey,
        params: { session_key: sessionKey },
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}

// Server Push / Application Messages

/**
 * Creates a 'message' request for application-specific communication.
 * The message parameters are handled by the virtual application.
 *
 * @param params - Application-specific parameters
 * @param requestId - Optional request ID
 * @param timestamp - Optional timestamp
 * @returns JSON string of the RPC message
 */
export function createApplicationMessage(
    params: Record<string, unknown>,
    requestId: RequestID = generateRequestId(),
    timestamp: Timestamp = getCurrentTimestamp(),
): string {
    const request = NitroliteRPC.createRequest({
        method: RPCMethod.Message,
        params,
        requestId,
        timestamp,
    });
    return JSON.stringify(request);
}
