import { createPingMessage, createECDSAMessageSigner, parseAnyRPCResponse, RPCMethod } from '@erc7824/nitrolite';
import { Wallet } from 'ethers';
import { WebSocket } from 'ws';
import {
    authenticateWithNitrolite,
    sendAuthRequest,
    isTokenExpiredError,
    processAuthResponse,
    clearJWTToken,
    parseNitroliteError,
} from './auth.js';
import type {
    NitroliteAuthContext,
    WSStatus,
    SessionKey,
    NitroliteConfig,
    NitroliteConnectionCallbacks,
} from './types.js';
import { UserRejectedError } from '../../types/index.js';
import { config, isDevelopment } from '../../config/index.js';
import { logger } from '../../utils/logger.js';

export const DEFAULT_CONFIG: NitroliteConfig = {
    wsUrl: process.env.YELLOW_WS_URL || 'wss://clearnet.yellow.com/ws',
    pingInterval: 30000,
    reconnectDelay: 2000,
    maxRetries: 5,
    requestTimeout: 30000,
};

// In-memory session storage for server
let sessionKeyStore: SessionKey | null = null;

export class NitroliteWebSocketClient {
    private ws: WebSocket | null = null;
    private status: WSStatus = 'disconnected';
    private sessionKey: SessionKey | null = null;
    private walletAddress: string | null = null;
    private privateKey: string | null = null;
    private isAuthenticated = false;
    private pingInterval: NodeJS.Timeout | null = null;
    private reconnectTimeout: NodeJS.Timeout | null = null;
    private retryCount = 0;
    private isDestroyed = false;
    private userRejectedAuth = false;
    private pendingChallenge: any = null;
    private rawChallengeMessage: string | null = null;
    private challengeTimeout: NodeJS.Timeout | null = null;
    private challengeKeepAliveInterval: NodeJS.Timeout | null = null;
    private authMessageHandler: ((event: MessageEvent) => void) | null = null;
    private authInProgress = false;

    private statusListeners = new Set<(status: WSStatus) => void>();
    private messageListeners = new Set<(message: any) => void>();
    private errorListeners = new Set<(error: Error) => void>();

    private config: NitroliteConfig;
    private callbacks: NitroliteConnectionCallbacks;

    constructor(config: Partial<NitroliteConfig> = {}, callbacks: NitroliteConnectionCallbacks = {}) {
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

    async connect(walletAddress: string, privateKey: string): Promise<void> {
        if (this.isDestroyed) {
            throw new Error('Client has been destroyed');
        }

        if (this.status === 'connecting' || this.isConnected) {
            logger.debug('connecting or connected');
            return;
        }

        if (this.userRejectedAuth) {
            throw new UserRejectedError('User previously rejected authentication');
        }

        if (!privateKey) {
            throw new Error('Private key is required for server-side authentication');
        }

        this.walletAddress = walletAddress;
        this.privateKey = privateKey;
        this.setStatus('connecting');

        try {
            await this.initializeSessionKey();
            await this.createWebSocketConnection();
            await this.authenticate();

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
            throw new Error('Not connected to Nitrolite WebSocket');
        }

        const message = typeof data === 'string' ? data : JSON.stringify(data);
        this.ws!.send(message);
    }

    async approveChallenge(): Promise<void> {
        if (!this.pendingChallenge) {
            throw new Error('No pending challenge to approve');
        }

        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            throw new Error('WebSocket not connected');
        }

        if (this.authInProgress) {
            logger.debug('Challenge approval already in progress, skipping duplicate');
            return;
        }

        try {
            this.authInProgress = true; // Prevent duplicate approvals

            const authContext: NitroliteAuthContext = {
                walletAddress: this.walletAddress!,
                sessionKey: this.sessionKey!,
                privateKey: this.privateKey!,
            };

            await authenticateWithNitrolite(
                this.ws,
                authContext,
                this.config.requestTimeout,
                this.pendingChallenge,
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
        } finally {
            this.authInProgress = false; // Always reset the flag
        }
    }

    rejectChallenge(): void {
        this.clearChallenge();
        this.userRejectedAuth = true;
        this.setStatus('disconnected');
        this.callbacks.onAuthFailed?.('Authentication rejected');
    }

    private clearChallenge(): void {
        this.pendingChallenge = null;
        this.rawChallengeMessage = null;
        if (this.challengeTimeout) {
            clearTimeout(this.challengeTimeout);
            this.challengeTimeout = null;
        }
        this.clearChallengeKeepAlive();
    }

    private async initializeSessionKey(): Promise<void> {
        // For server-side implementation, use the wallet private key as session key
        if (!this.privateKey) {
            throw new Error('Private key required for server session');
        }

        // Create wallet from private key to get the address
        const { Wallet } = await import('ethers');
        const wallet = new Wallet(this.privateKey);

        this.sessionKey = {
            privateKey: this.privateKey,
            address: wallet.address,
        };
        sessionKeyStore = this.sessionKey;
    }

    private async createWebSocketConnection(): Promise<void> {
        return new Promise((resolve, reject) => {
            try {
                logger.debug('wsurl', this.config.wsUrl);
                this.ws = new WebSocket(this.config.wsUrl);

                this.ws.onopen = () => {
                    logger.info('WebSocket connection opened');
                    resolve();
                };
                this.ws.onmessage = (event) => {
                    this.handleMessage(event);
                    if (this.authMessageHandler) {
                        this.authMessageHandler(event);
                    }
                };
                this.ws.onclose = () => {
                    logger.info('WebSocket connection closed');
                    this.handleDisconnection();
                };
                this.ws.onerror = (error) => {
                    logger.error('WebSocket error:', error);
                    reject(new Error('WebSocket connection failed'));
                };
            } catch (error) {
                reject(error);
            }
        });
    }

    private async authenticate(): Promise<void> {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            throw new Error('WebSocket connection is not established');
        }
        if (!this.sessionKey || !this.walletAddress || !this.privateKey) {
            throw new Error('Authentication context not available - missing private key');
        }

        const authContext: NitroliteAuthContext = {
            walletAddress: this.walletAddress,
            sessionKey: this.sessionKey,
            privateKey: this.privateKey,
        };

        await this.initializeManualAuth(authContext);
    }

    private async initializeManualAuth(authContext: NitroliteAuthContext): Promise<void> {
        if (this.authInProgress) {
            logger.debug('Authentication already in progress, skipping duplicate request');
            return;
        }

        this.authInProgress = true;
        logger.info('Starting authentication flow...');

        return new Promise((resolve, reject) => {
            const cleanup = () => {
                this.ws?.removeEventListener('message', handleMessage);
                this.authMessageHandler = null;
                this.authInProgress = false;
            };

            const handleMessage = async (event: MessageEvent) => {
                try {
                    let rawJsonMessage;
                    try {
                        rawJsonMessage = JSON.parse(event.data);
                    } catch {
                        return;
                    }

                    const nitroliteError = parseNitroliteError(rawJsonMessage);
                    if (nitroliteError.isTokenExpired) {
                        logger.info('Token error detected, clearing and restarting authentication...');

                        clearJWTToken();
                        this.isAuthenticated = false;
                        this.pendingChallenge = null;
                        this.clearChallengeKeepAlive();

                        cleanup();

                        if (!this.authInProgress) {
                            this.initializeManualAuth(authContext)
                                .then(() => {
                                    logger.info('Fresh authentication flow completed after token error');
                                })
                                .catch((error) => {
                                    logger.error('Fresh auth flow failed:', error);
                                    this.callbacks.onVerifyFailed?.(error.message || 'Re-authentication failed');
                                });
                        }

                        return;
                    }

                    let rawMessage;
                    try {
                        rawMessage = parseAnyRPCResponse(event.data);
                    } catch {
                        return;
                    }

                    if (rawMessage.method === RPCMethod.AuthChallenge) {
                        this.pendingChallenge = rawMessage;
                        this.setStatus('pending_auth');
                        this.callbacks.onChallengeReceived?.(rawMessage);

                        this.startChallengeKeepAlive();
                        resolve();
                    } else if (rawMessage.method === RPCMethod.AuthVerify) {
                        if (rawMessage.params?.success) {
                            cleanup();
                            this.isAuthenticated = true;
                            resolve();
                        } else {
                            this.callbacks.onVerifyFailed?.('Authentication verification failed');
                        }
                    } else if (rawMessage.method === RPCMethod.Error) {
                        const authResult = processAuthResponse(rawMessage);

                        if (authResult.tokenExpired) {
                            logger.info(
                                'Token error detected via processAuthResponse, clearing and restarting authentication...',
                            );

                            clearJWTToken();
                            this.isAuthenticated = false;
                            this.pendingChallenge = null;
                            this.clearChallengeKeepAlive();

                            cleanup();

                            if (!this.authInProgress) {
                                this.initializeManualAuth(authContext)
                                    .then(() => {
                                        logger.info('Fresh authentication flow completed after token error');
                                    })
                                    .catch((error) => {
                                        logger.error('Fresh auth flow failed:', error);
                                        this.callbacks.onVerifyFailed?.(error.message || 'Re-authentication failed');
                                    });
                            }

                            return;
                        } else {
                            this.callbacks.onVerifyFailed?.(rawMessage.params?.error || 'Authentication error');
                        }
                    }
                } catch (error) {
                    // Skip parsing errors
                }
            };

            this.ws!.addEventListener('message', handleMessage);
            this.authMessageHandler = handleMessage;

            sendAuthRequest(this.ws!, authContext).catch((error) => {
                cleanup();
                reject(error);
            });
        });
    }

    private startChallengeKeepAlive(): void {
        if (this.challengeKeepAliveInterval) {
            clearInterval(this.challengeKeepAliveInterval);
        }

        this.challengeKeepAliveInterval = setInterval(() => {
            if (this.sessionKey) {
                this.ping().catch(() => {
                    this.clearChallengeKeepAlive();
                });
            } else {
                this.clearChallengeKeepAlive();
            }
        }, 30000);
    }

    private clearChallengeKeepAlive(): void {
        if (this.challengeKeepAliveInterval) {
            clearInterval(this.challengeKeepAliveInterval);
            this.challengeKeepAliveInterval = null;
        }
    }

    private handleTokenExpiration(): void {
        logger.info('Handling token expiration - clearing auth state and triggering re-auth...');

        this.isAuthenticated = false;
        this.pendingChallenge = null;
        this.clearChallengeKeepAlive();
        this.setStatus('connecting');

        if (this.authInProgress) {
            logger.debug('Auth already in progress, skipping token expiration handling');
            return;
        }

        if (
            this.ws &&
            this.ws.readyState === WebSocket.OPEN &&
            this.sessionKey &&
            this.walletAddress &&
            this.privateKey
        ) {
            const authContext: NitroliteAuthContext = {
                walletAddress: this.walletAddress,
                sessionKey: this.sessionKey,
                privateKey: this.privateKey,
            };

            this.initializeManualAuth(authContext)
                .then(() => {
                    logger.info('Fresh authentication flow completed after token expiration');
                })
                .catch((error) => {
                    logger.error('Token expiration re-auth failed:', error);
                    this.emitError(new Error(`Re-authentication failed: ${error.message}`));
                });
        } else {
            logger.error('Cannot re-authenticate: missing context or connection');
            this.emitError(new Error('Cannot re-authenticate: connection or context unavailable'));
        }
    }

    private handleMessage(event: MessageEvent): void {
        try {
            // First check for raw JSON error format (for parseNitroliteError)
            let rawJsonMessage;
            try {
                rawJsonMessage = JSON.parse(event.data);
            } catch {
                return;
            }

            const nitroliteError = parseNitroliteError(rawJsonMessage);
            if (nitroliteError.isTokenExpired) {
                logger.info('Token error detected, clearing and triggering re-authentication...');
                clearJWTToken();
                this.handleTokenExpiration();
                return;
            }

            try {
                const response = parseAnyRPCResponse(event.data);

                if (response.method === RPCMethod.AuthChallenge) {
                    if (!this.isAuthenticated && !this.authInProgress) {
                        this.pendingChallenge = response;
                        this.rawChallengeMessage = event.data;
                        this.setStatus('pending_auth');
                        this.callbacks.onChallengeReceived?.(response);
                        this.startChallengeKeepAlive();
                    } else {
                        logger.debug('Ignoring auth_challenge - already authenticated or auth in progress');
                    }
                    return;
                }

                if (response.method === RPCMethod.AuthVerify) {
                    // Only process auth_verify if we're not already authenticated (prevent duplicates)
                    if (!this.isAuthenticated) {
                        logger.info('📥 Received auth_verify response:', JSON.stringify(response, null, 2));
                        const authResult = processAuthResponse(response);
                        if (authResult.success) {
                            logger.info('🎉 Authentication verification successful!');
                            this.isAuthenticated = true;
                            this.startPingInterval();
                            this.setStatus('connected');
                            this.retryCount = 0;
                            this.callbacks.onConnect?.();
                            this.callbacks.onAuthSuccess?.();
                        } else {
                            logger.warn('❌ Authentication verification failed:', authResult.error);
                            this.callbacks.onVerifyFailed?.(authResult.error || 'Authentication failed');
                        }
                    } else {
                        logger.debug('Ignoring auth_verify - already authenticated');
                    }
                } else if (response.method === RPCMethod.Pong) {
                    // Pong received - connection healthy
                } else if (response.method === RPCMethod.Error) {
                    if (isTokenExpiredError(response.params?.error)) {
                        logger.info('Token error in RPC response, clearing and triggering re-authentication...');
                        clearJWTToken();
                        this.handleTokenExpiration();
                    } else {
                        this.emitError(new Error('Nitrolite service error'));
                    }
                } else {
                    this.emitMessage(response);
                }
                return;
            } catch (rpcError) {
                if (isDevelopment) {
                    logger.debug('Failed to parse as RPC, handling as raw message:', rpcError);
                }
            }

            if (rawJsonMessage.method === RPCMethod.Assets) {
                this.emitMessage(rawJsonMessage);
                return;
            }

            this.emitMessage(rawJsonMessage);
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
        this.reconnectTimeout = setTimeout(() => {
            if (this.walletAddress && this.privateKey) {
                this.connect(this.walletAddress, this.privateKey).catch(() => {
                    // Reconnection failed - handled by handleConnectionError
                });
            }
        }, delay);
    }

    private startPingInterval(): void {
        this.pingInterval = setInterval(() => {
            if (this.isConnected) {
                this.ping().catch(() => {
                    // Ping failed - connection may be broken
                });
            }
        }, this.config.pingInterval);
    }

    private cleanup(): void {
        if (this.pingInterval) {
            clearInterval(this.pingInterval);
            this.pingInterval = null;
        }

        if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
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

    resetRejectionState(): void {
        this.userRejectedAuth = false;
    }

    handleChallengeMessage(challengeData: any): void {
        if (!this.isAuthenticated && challengeData.method === RPCMethod.AuthChallenge) {
            this.pendingChallenge = challengeData;
            this.setStatus('pending_auth');
            this.callbacks.onChallengeReceived?.(challengeData);
            this.startChallengeKeepAlive();
        }
    }
}

export function createNitroliteWebSocketClient(
    config?: Partial<NitroliteConfig>,
    callbacks?: NitroliteConnectionCallbacks,
): NitroliteWebSocketClient {
    return new NitroliteWebSocketClient(config, callbacks);
}

// Global client instance for server use
let globalClient: NitroliteWebSocketClient | null = null;

export async function initializeNitroliteClient(): Promise<NitroliteWebSocketClient> {
    logger.info('🚀 Initializing Nitrolite WebSocket client...');

    if (globalClient) {
        logger.info('📋 Using existing global client');
        return globalClient;
    }

    if (!config.walletPrivateKey) {
        throw new Error('WALLET_PRIVATE_KEY is required for Nitrolite client initialization');
    }

    const wallet = new Wallet(config.walletPrivateKey);
    const walletAddress = wallet.address;

    logger.info('💼 Wallet address:', walletAddress);
    logger.info('🌐 Yellow WebSocket URL:', config.yellowWsUrl);
    logger.info('🏷️  App name:', config.vApp.name);
    logger.info('🔍 App scope:', config.vApp.scope);

    globalClient = createNitroliteWebSocketClient(
        {
            wsUrl: config.yellowWsUrl,
        },
        {
            onConnect: () => {
                logger.info('✅ Nitrolite WebSocket connected successfully');
            },
            onDisconnect: () => {
                logger.warn('❌ Nitrolite WebSocket disconnected');
            },
            onError: (error) => {
                logger.error('💥 Nitrolite WebSocket error:', error.message);
            },
            onAuthSuccess: () => {
                logger.info('🔐 Nitrolite authentication successful');
            },
            onAuthFailed: (error) => {
                logger.error('🔒 Nitrolite authentication failed:', error);
            },
            onChallengeReceived: (challenge) => {
                logger.info('🤝 Nitrolite auth challenge received, will auto-approve...');
                logger.debug('Challenge details:', JSON.stringify(challenge, null, 2));
                // Auto-approve challenge for server (like frontend manual approval)
                if (globalClient && globalClient.hasPendingChallenge) {
                    globalClient.approveChallenge().catch((error) => {
                        logger.error('❌ Failed to auto-approve challenge:', error);
                    });
                }
            },
            onMessage: (message) => {
                logger.debug('📨 Nitrolite message received:', message);
            },
        },
    );

    // Connect the client
    logger.info('🔗 Connecting to Nitrolite WebSocket...');
    await globalClient.connect(walletAddress, config.walletPrivateKey);

    logger.info('✅ Nitrolite client initialization completed');
    return globalClient;
}

export function getNitroliteClient(): NitroliteWebSocketClient | null {
    return globalClient;
}
