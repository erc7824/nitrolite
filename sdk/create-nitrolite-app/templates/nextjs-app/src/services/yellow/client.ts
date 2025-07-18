import { createPingMessage, createECDSAMessageSigner, parseAnyRPCResponse, RPCMethod } from '@erc7824/nitrolite';
import { generateKeyPair } from '../../utils/crypto';
import RequestStore, { type RequestInfo } from '../../store/RequestStore';
import {
    authenticateWithYellow,
    sendAuthRequest,
    isTokenExpiredError,
    processAuthResponse,
    clearJWTToken,
    parseYellowError,
} from './auth';
import type {
    YellowAuthContext,
    WSStatus,
    SessionKey,
    YellowConfig,
    YellowConnectionCallbacks,
    CompatibleWalletClient,
} from './types';
import { UserRejectedError } from './types';
import { config } from '@/utils/env';

const DEFAULT_CONFIG: YellowConfig = {
    wsUrl: config.yellowWsUrl,
    pingInterval: 30000,
    reconnectDelay: 2000,
    maxRetries: 5,
    requestTimeout: 30000,
};

const SESSION_KEY = 'myapp_session_key';

export class YellowWebSocketClient {
    private ws: WebSocket | null = null;
    private status: WSStatus = 'disconnected';
    private sessionKey: SessionKey | null = null;
    private walletAddress: string | null = null;
    private walletClient: CompatibleWalletClient | null = null;
    private signTypedData: YellowAuthContext['signTypedData'] | null = null;
    private isAuthenticated = false;
    private pingInterval: number | null = null;
    private reconnectTimeout: number | null = null;
    private retryCount = 0;
    private isDestroyed = false;
    private userRejectedAuth = false;
    private pendingChallenge: any = null;
    private rawChallengeMessage: string | null = null;
    private challengeTimeout: number | null = null;
    private challengeKeepAliveInterval: number | null = null;
    private authMessageHandler: ((event: MessageEvent) => void) | null = null;
    private authInProgress = false;

    private statusListeners = new Set<(status: WSStatus) => void>();
    private messageListeners = new Set<(message: any) => void>();
    private errorListeners = new Set<(error: Error) => void>();

    private config: YellowConfig;
    private callbacks: YellowConnectionCallbacks;

    constructor(config: Partial<YellowConfig> = {}, callbacks: YellowConnectionCallbacks = {}) {
        this.config = { ...DEFAULT_CONFIG, ...config };
        this.callbacks = callbacks;
    }

    get isConnected(): boolean {
        return this.status === 'connected' && this.isAuthenticated;
    }

    get currentStatus(): WSStatus {
        return this.status;
    }

    get currentSessionAddress(): string | null {
        return this.sessionKey?.address || null;
    }

    get hasPendingChallenge(): boolean {
        return this.pendingChallenge !== null;
    }

    get sessionSigner() {
        if (!this.sessionKey) {
            return null;
        }
        return createECDSAMessageSigner(this.sessionKey.privateKey as `0x${string}`);
    }

    onStatusChange(listener: (status: WSStatus) => void): () => void {
        this.statusListeners.add(listener);
        return () => this.statusListeners.delete(listener);
    }

    onMessage(listener: (message: any) => void): () => void {
        this.messageListeners.add(listener);
        return () => this.messageListeners.delete(listener);
    }

    onError(listener: (error: Error) => void): () => void {
        this.errorListeners.add(listener);
        return () => this.errorListeners.delete(listener);
    }

    async connect(
        walletAddress: string,
        signTypedData: YellowAuthContext['signTypedData'],
        walletClient?: CompatibleWalletClient | null,
    ): Promise<void> {
        if (this.isDestroyed) {
            throw new Error('Client has been destroyed');
        }

        if (this.status === 'connecting' || this.isConnected) {
            return;
        }

        if (this.userRejectedAuth) {
            throw new UserRejectedError('User previously rejected authentication');
        }

        // Ensure wallet client is available before proceeding
        if (!walletClient) {
            throw new Error('Wallet client is required for authentication');
        }

        this.walletAddress = walletAddress;
        this.signTypedData = signTypedData;
        this.walletClient = walletClient;
        this.setStatus('connecting');

        try {
            await this.initializeSessionKey();
            await this.createWebSocketConnection();
            await this.authenticate();

            // Only set connected status if not waiting for challenge approval
            if (this.status !== 'pending_auth') {
                this.startPingInterval();
                this.setStatus('connected');
                this.retryCount = 0;
                this.callbacks.onConnect?.();
                this.callbacks.onAuthSuccess?.();
            }
        } catch (error) {
            this.handleConnectionError(error as Error);
            throw error;
        }
    }

    disconnect(): void {
        this.cleanup();
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
        this.setStatus('disconnected');
        this.isAuthenticated = false;
        this.callbacks.onDisconnect?.();
    }

    destroy(): void {
        this.isDestroyed = true;
        this.disconnect();
        this.statusListeners.clear();
        this.messageListeners.clear();
        this.errorListeners.clear();
    }

    async ping(): Promise<void> {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN || !this.sessionKey) {
            throw new Error('Not connected or session key not available');
        }

        const sessionSigner = createECDSAMessageSigner(this.sessionKey.privateKey as `0x${string}`);
        const pingMessage = await createPingMessage(sessionSigner);
        this.ws!.send(pingMessage);
    }

    send(data: any): void {
        if (!this.isConnected) {
            throw new Error('Not connected to Yellow WebSocket');
        }

        const message = typeof data === 'string' ? data : JSON.stringify(data);
        this.ws!.send(message);
    }

    async sendWithResponse(
        data: any,
        options: { timeout?: number } = {},
    ): Promise<{ requestInfo: RequestInfo; response: any }> {
        if (!this.isConnected) {
            throw new Error('Not connected to Yellow WebSocket');
        }

        // Parse the message to extract request information
        let parsedMessage;
        try {
            const messageStr = typeof data === 'string' ? data : JSON.stringify(data);
            parsedMessage = JSON.parse(messageStr);
        } catch (error) {
            throw new Error('Invalid message format - could not parse JSON');
        }

        // Extract request information from the message
        let requestInfo: RequestInfo;
        if (parsedMessage.req) {
            const [requestId, method, params, timestamp] = parsedMessage.req;
            requestInfo = {
                requestId,
                method,
                params,
                timestamp: timestamp || Date.now(),
            };
        } else {
            throw new Error('Message does not contain valid RPC request format');
        }

        // Send the message
        const responsePromise = RequestStore.registerRequest(requestInfo, options);
        const message = typeof data === 'string' ? data : JSON.stringify(data);
        this.ws!.send(message);
        const response = await responsePromise;

        return {
            requestInfo,
            response,
        };
    }

    async approveChallenge(): Promise<void> {
        if (!this.pendingChallenge) {
            throw new Error('No pending challenge to approve');
        }

        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            throw new Error('WebSocket not connected');
        }

        try {
            const authContext: YellowAuthContext = {
                walletAddress: this.walletAddress!,
                sessionKey: this.sessionKey!,
                walletClient: this.walletClient!,
                signTypedData: this.signTypedData!,
            };

            // Pass the raw message along with the parsed challenge
            await authenticateWithYellow(
                this.ws,
                authContext,
                this.config.requestTimeout,
                this.pendingChallenge,
                // @ts-ignore
                this.rawChallengeMessage,
            );

            this.clearChallenge();
            this.isAuthenticated = true;
            this.startPingInterval();
            this.setStatus('connected');
            this.retryCount = 0;
            this.callbacks.onConnect?.();
            this.callbacks.onAuthSuccess?.();
        } catch (error) {
            this.callbacks.onVerifyFailed?.(error instanceof Error ? error.message : 'Challenge approval failed');
            // Don't disconnect on verify failure - keep WS alive
        }
    }

    rejectChallenge(): void {
        this.clearChallenge();
        this.userRejectedAuth = true;
        this.setStatus('disconnected');
        this.callbacks.onAuthFailed?.('User rejected authentication');
    }

    private clearChallenge(): void {
        this.pendingChallenge = null;
        this.rawChallengeMessage = null;
        if (this.challengeTimeout) {
            window.clearTimeout(this.challengeTimeout);
            this.challengeTimeout = null;
        }
        this.clearChallengeKeepAlive();
    }

    private async initializeSessionKey(): Promise<void> {
        try {
            const stored = localStorage.getItem(SESSION_KEY);
            if (stored) {
                const parsed = JSON.parse(stored);
                if (parsed.privateKey && parsed.address) {
                    this.sessionKey = parsed;
                    return;
                }
            }
        } catch {
            // Failed to load stored session key, will generate new one
        }

        this.sessionKey = await generateKeyPair();
        try {
            localStorage.setItem(SESSION_KEY, JSON.stringify(this.sessionKey));
        } catch {
            // Storage failed - continue without persistence
        }
    }

    private createWebSocketConnection(): Promise<void> {
        return new Promise((resolve, reject) => {
            try {
                this.ws = new WebSocket(this.config.wsUrl);

                this.ws.onopen = () => resolve();
                this.ws.onmessage = (event) => {
                    this.handleMessage(event);
                    // Also forward to auth handler if it exists
                    if (this.authMessageHandler) {
                        this.authMessageHandler(event);
                    }
                };
                this.ws.onclose = () => this.handleDisconnection();
                this.ws.onerror = () => reject(new Error('WebSocket connection failed'));
            } catch (error) {
                reject(error);
            }
        });
    }

    private async authenticate(): Promise<void> {
        if (
            !this.ws ||
            this.ws.readyState !== WebSocket.OPEN ||
            !this.sessionKey ||
            !this.walletAddress ||
            !this.signTypedData ||
            !this.walletClient
        ) {
            throw new Error('Authentication context not available - missing wallet client');
        }

        const authContext: YellowAuthContext = {
            walletAddress: this.walletAddress,
            sessionKey: this.sessionKey,
            walletClient: this.walletClient,
            signTypedData: this.signTypedData,
        };

        await this.initializeManualAuth(authContext);
    }

    private async initializeManualAuth(authContext: YellowAuthContext): Promise<void> {
        // Prevent multiple concurrent auth flows
        if (this.authInProgress) {
            console.log('Authentication already in progress, skipping duplicate request');
            return;
        }

        this.authInProgress = true;
        console.log('Starting authentication flow...');

        return new Promise((resolve, reject) => {
            const cleanup = () => {
                this.ws?.removeEventListener('message', handleMessage);
                this.authMessageHandler = null;
                this.authInProgress = false;
            };

            const handleMessage = async (event: MessageEvent) => {
                console.log('testing 123123auth_verify');

                try {
                    // First, try to parse as raw JSON to check for Yellow error format
                    let rawJsonMessage;
                    try {
                        rawJsonMessage = JSON.parse(event.data);
                    } catch {
                        return; // Skip non-JSON messages
                    }

                    // Check for Yellow error format first (handles "invalid JWT token" errors)
                    const yellowError = parseYellowError(rawJsonMessage);
                    if (yellowError.isTokenExpired) {
                        console.log(
                            'Token error detected in auth handler via parseYellowError, clearing and restarting authentication...',
                        );

                        // Force clear any malformed JWT token
                        clearJWTToken();

                        // Clear the expired state and reset auth context
                        this.isAuthenticated = false;
                        this.pendingChallenge = null;
                        this.clearChallengeKeepAlive();

                        // Restart the entire authentication flow
                        cleanup(); // Clean up current auth handler

                        // Only start fresh auth if not already in progress
                        if (!this.authInProgress) {
                            this.initializeManualAuth(authContext)
                                .then(() => {
                                    console.log(
                                        'Fresh authentication flow completed after token error in auth handler',
                                    );
                                })
                                .catch((error) => {
                                    console.error('Fresh auth flow failed in auth handler:', error);
                                    this.callbacks.onVerifyFailed?.(error.message || 'Re-authentication failed');
                                });
                        } else {
                            console.log('Auth already in progress, skipping duplicate restart in auth handler');
                        }

                        return; // Exit current handler
                    }

                    // Now try to parse as RPC message using nitrolite SDK
                    let rawMessage;
                    try {
                        rawMessage = parseAnyRPCResponse(event.data);
                    } catch {
                        return; // Skip non-RPC messages
                    }

                    if (rawMessage.method === RPCMethod.AuthChallenge) {
                        // Store challenge and notify UI
                        this.pendingChallenge = rawMessage;
                        this.setStatus('pending_auth');
                        this.callbacks.onChallengeReceived?.(rawMessage);

                        // Keep pinging to maintain connection while waiting
                        this.startChallengeKeepAlive();
                        resolve(); // Resolve to continue connection flow, but stay in pending_auth status
                    } else if (rawMessage.method === RPCMethod.AuthVerify) {
                        console.log('testing auth_verify', rawMessage);
                        if (rawMessage.params?.success) {
                            cleanup();
                            this.isAuthenticated = true;
                            resolve();
                        } else {
                            this.callbacks.onVerifyFailed?.('Authentication verification failed');
                            // Don't reject - keep connection alive
                        }
                    } else if (rawMessage.method === RPCMethod.Error) {
                        const authResult = processAuthResponse(rawMessage);

                        // Check if this is a token expiration error
                        if (authResult.tokenExpired) {
                            console.log(
                                'Token error detected via processAuthResponse, clearing and restarting authentication...',
                            );

                            // Force clear any malformed JWT token
                            clearJWTToken();

                            // Clear the expired state and reset auth context
                            this.isAuthenticated = false;
                            this.pendingChallenge = null;
                            this.clearChallengeKeepAlive();

                            // Restart the entire authentication flow
                            cleanup(); // Clean up current auth handler

                            // Only start fresh auth if not already in progress
                            if (!this.authInProgress) {
                                this.initializeManualAuth(authContext)
                                    .then(() => {
                                        console.log('Fresh authentication flow completed after token error');
                                    })
                                    .catch((error) => {
                                        console.error('Fresh auth flow failed:', error);
                                        this.callbacks.onVerifyFailed?.(error.message || 'Re-authentication failed');
                                    });
                            } else {
                                console.log('Auth already in progress, skipping duplicate restart');
                            }

                            return; // Exit current handler
                        } else {
                            this.callbacks.onVerifyFailed?.(rawMessage.params?.error || 'Authentication error');
                            // Don't reject - keep connection alive
                        }
                    }
                } catch (error) {
                    // Skip parsing errors
                }
            };

            this.ws!.addEventListener('message', handleMessage);
            this.authMessageHandler = handleMessage;

            // Send initial auth request
            sendAuthRequest(this.ws!, authContext).catch((error) => {
                cleanup();
                reject(error);
            });
        });
    }

    private startChallengeKeepAlive(): void {
        // Clear any existing keep-alive interval
        if (this.challengeKeepAliveInterval) {
            window.clearInterval(this.challengeKeepAliveInterval);
        }

        // Send ping every 30 seconds to keep connection alive
        this.challengeKeepAliveInterval = window.setInterval(() => {
            if (this.sessionKey) {
                this.ping().catch(() => {
                    // Ping failed - connection may be broken
                    this.clearChallengeKeepAlive();
                });
            } else {
                // Session key not available
                this.clearChallengeKeepAlive();
            }
        }, 30000);
    }

    private clearChallengeKeepAlive(): void {
        if (this.challengeKeepAliveInterval) {
            window.clearInterval(this.challengeKeepAliveInterval);
            this.challengeKeepAliveInterval = null;
        }
    }

    private handleTokenExpiration(): void {
        console.log('Handling token expiration - clearing auth state and triggering re-auth...');

        // Clear authentication state and reset challenge context
        this.isAuthenticated = false;
        this.pendingChallenge = null;
        this.clearChallengeKeepAlive();
        this.setStatus('connecting'); // Reset to connecting state to trigger fresh auth

        // Only proceed if auth is not already in progress
        if (this.authInProgress) {
            console.log('Auth already in progress, skipping token expiration handling');
            return;
        }

        // Ensure we have the necessary context for re-authentication
        if (
            this.ws &&
            this.ws.readyState === WebSocket.OPEN &&
            this.sessionKey &&
            this.walletAddress &&
            this.signTypedData &&
            this.walletClient
        ) {
            const authContext: YellowAuthContext = {
                walletAddress: this.walletAddress,
                sessionKey: this.sessionKey,
                walletClient: this.walletClient,
                signTypedData: this.signTypedData,
            };

            // Start fresh authentication flow
            this.initializeManualAuth(authContext)
                .then(() => {
                    console.log('Fresh authentication flow completed after token expiration in general flow');
                })
                .catch((error) => {
                    console.error('Token expiration re-auth failed:', error);
                    this.emitError(new Error(`Re-authentication failed: ${error.message}`));
                });
        } else {
            console.error('Cannot re-authenticate: missing context or connection');
            this.emitError(new Error('Cannot re-authenticate: connection or context unavailable'));
        }
    }

    private handleMessage(event: MessageEvent): void {
        try {
            let rawMessage;
            try {
                rawMessage = JSON.parse(event.data);
            } catch {
                return; // Skip non-JSON messages
            }

            // Check for Yellow error format first (handles both RPC and direct errors)
            const yellowError = parseYellowError(rawMessage);
            if (yellowError.isTokenExpired) {
                console.log('Token error detected via parseYellowError, clearing and triggering re-authentication...');
                clearJWTToken(); // Force clear malformed JWT
                this.handleTokenExpiration();
                return;
            }

            // Parse as RPC message using nitrolite SDK
            try {
                const response = parseAnyRPCResponse(event.data);

                // Check if this is a response to a pending request
                if (response.requestId) {
                    RequestStore.handleResponse(response.requestId, response);
                }

                // Handle auth challenge messages
                if (response.method === RPCMethod.AuthChallenge) {
                    if (!this.isAuthenticated) {
                        this.pendingChallenge = response;
                        this.rawChallengeMessage = event.data; // Store raw message for nitrolite
                        this.setStatus('pending_auth');
                        this.callbacks.onChallengeReceived?.(response);
                        this.startChallengeKeepAlive();
                    } else {
                        console.log('Ignoring auth_challenge because already authenticated');
                    }
                    return;
                }

                // Handle other RPC responses
                if (response.method === RPCMethod.Pong) {
                    // Pong received - connection healthy
                } else if (response.method === RPCMethod.Error) {
                    // Check if this is a token expiration error in general responses
                    if (isTokenExpiredError(response.params?.error)) {
                        console.log('Token error in RPC response, clearing and triggering re-authentication...');
                        clearJWTToken(); // Force clear malformed JWT
                        this.handleTokenExpiration();
                    } else {
                        this.emitError(new Error('Yellow service error'));
                    }
                } else {
                    this.emitMessage(response);
                }
                return;
            } catch (rpcError) {
                // If not parsable as RPC, continue to handle as raw message
                if (import.meta.env.DEV) {
                    console.log('Failed to parse as RPC, handling as raw message:', rpcError);
                }
            }

            // Handle special messages that don't follow RPC format
            if (rawMessage.method === RPCMethod.Assets) {
                this.emitMessage(rawMessage);
                return;
            }

            // If we get here, it's an unhandled raw message
            this.emitMessage(rawMessage);
        } catch {
            // Message parsing failed - skip
        }
    }

    private handleDisconnection(): void {
        this.cleanup();
        this.isAuthenticated = false;

        if (!this.isDestroyed && this.shouldReconnect()) {
            this.scheduleReconnect();
        } else {
            this.setStatus('disconnected');
            this.callbacks.onDisconnect?.();
        }
    }

    private handleConnectionError(error: Error): void {
        this.emitError(error);
        this.callbacks.onError?.(error);

        if (error instanceof UserRejectedError || UserRejectedError.isUserRejection(error)) {
            this.userRejectedAuth = true;
            this.setStatus('disconnected');
            this.callbacks.onAuthFailed?.(error.message);
        } else if (this.shouldReconnect()) {
            this.scheduleReconnect();
        } else {
            this.setStatus('failed');
        }
    }

    private shouldReconnect(): boolean {
        return !this.userRejectedAuth && this.retryCount < this.config.maxRetries;
    }

    private scheduleReconnect(): void {
        if (this.isDestroyed) return;

        this.retryCount++;
        this.setStatus('reconnecting');

        const delay = this.config.reconnectDelay * Math.pow(2, this.retryCount - 1);
        this.reconnectTimeout = window.setTimeout(() => {
            if (this.walletAddress && this.signTypedData) {
                this.connect(this.walletAddress, this.signTypedData, this.walletClient).catch(() => {
                    // Reconnection failed - handled by handleConnectionError
                });
            }
        }, delay);
    }

    private startPingInterval(): void {
        this.pingInterval = window.setInterval(() => {
            if (this.isConnected) {
                this.ping().catch(() => {
                    // Ping failed - connection may be broken
                });
            }
        }, this.config.pingInterval);
    }

    private cleanup(): void {
        if (this.pingInterval) {
            window.clearInterval(this.pingInterval);
            this.pingInterval = null;
        }

        if (this.reconnectTimeout) {
            window.clearTimeout(this.reconnectTimeout);
            this.reconnectTimeout = null;
        }

        this.clearChallenge();
    }

    private setStatus(status: WSStatus): void {
        if (this.status !== status) {
            this.status = status;
            this.statusListeners.forEach((listener) => listener(status));
        }
    }

    private emitMessage(message: any): void {
        this.messageListeners.forEach((listener) => listener(message));
        this.callbacks.onMessage?.(message);
    }

    private emitError(error: Error): void {
        this.errorListeners.forEach((listener) => listener(error));
    }

    // Reset user rejection flag (for new wallet connections)
    resetRejectionState(): void {
        this.userRejectedAuth = false;
    }

    // Method to handle challenge messages forwarded from providers
    handleChallengeMessage(challengeData: any): void {
        if (!this.isAuthenticated && challengeData.method === RPCMethod.AuthChallenge) {
            this.pendingChallenge = challengeData;
            this.setStatus('pending_auth');
            this.callbacks.onChallengeReceived?.(challengeData);
            this.startChallengeKeepAlive();
        }
    }
}

export function createYellowWebSocketClient(
    config?: Partial<YellowConfig>,
    callbacks?: YellowConnectionCallbacks,
): YellowWebSocketClient {
    return new YellowWebSocketClient(config, callbacks);
}
