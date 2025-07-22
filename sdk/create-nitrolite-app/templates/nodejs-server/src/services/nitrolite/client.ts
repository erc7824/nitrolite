import { createPingMessage, createECDSAMessageSigner, RPCMethod } from '@erc7824/nitrolite';
import { Wallet } from 'ethers';
import type { MessageEvent as WSMessageEvent } from 'ws';
import type { WSStatus, NitroliteConfig, NitroliteConnectionCallbacks } from '../../types/index.js';
import { UserRejectedError } from '../../types/index.js';
import { config } from '../../config/index.js';
import { logger } from '../../utils/logger.js';
import { StatusManager } from './status-manager.js';
import { WebSocketManager } from './websocket-manager.js';
import { AuthenticationManager } from './auth-manager.js';
import { ChallengeManager } from './challenge-manager.js';
import { MessageRouter } from './message-router.js';
import { ConnectionManager } from './connection-manager.js';
import { NitroliteEventEmitter } from './event-emitter.js';

export const DEFAULT_CONFIG: NitroliteConfig = {
    wsUrl: process.env.YELLOW_WS_URL || 'wss://clearnet.yellow.com/ws',
    pingInterval: 30000,
    reconnectDelay: 2000,
    maxRetries: 5,
    requestTimeout: 30000,
};

export class NitroliteWebSocketClient {
    private readonly statusManager: StatusManager;
    private readonly wsManager: WebSocketManager;
    private readonly authManager: AuthenticationManager;
    private readonly challengeManager: ChallengeManager;
    private readonly messageRouter: MessageRouter;
    private readonly connectionManager: ConnectionManager;
    private readonly eventEmitter: NitroliteEventEmitter;

    private config: NitroliteConfig;
    private callbacks: NitroliteConnectionCallbacks;

    constructor(config: Partial<NitroliteConfig> = {}, callbacks: NitroliteConnectionCallbacks = {}) {
        this.config = { ...DEFAULT_CONFIG, ...config };
        this.callbacks = callbacks;

        this.statusManager = new StatusManager();
        this.wsManager = new WebSocketManager(this.config);
        this.authManager = new AuthenticationManager();
        this.challengeManager = new ChallengeManager();
        this.messageRouter = new MessageRouter();
        this.connectionManager = new ConnectionManager(this.config);
        this.eventEmitter = new NitroliteEventEmitter();

        this.setupEventHandlers();
    }

    private setupEventHandlers(): void {
        // WebSocket events
        this.wsManager.onMessage((event) => this.handleMessage(event));
        this.wsManager.onClose(() => this.handleDisconnection());
        this.wsManager.onError((error) => this.handleWSError(error));

        // Status change events
        this.statusManager.onStatusChange((status) => {
            this.eventEmitter.status.emit(status);
        });

        // Authentication events
        this.authManager.onTokenExpired(() => this.handleTokenExpiration());

        // Challenge events
        this.challengeManager.onChallengeReceived((challenge) => {
            this.statusManager.setStatus('pending_auth');
            this.callbacks.onChallengeReceived?.(challenge);
        });

        // Connection events
        this.connectionManager.onMaxRetriesReached(() => {
            this.statusManager.setStatus('failed');
        });

        // Message routing
        this.messageRouter.onAuthChallenge((data) => this.handleAuthChallenge(data));
        this.messageRouter.onAuthVerify((data) => this.handleAuthVerify(data));
        this.messageRouter.onError((data) => this.handleRPCError(data));
        this.messageRouter.onGeneralMessage((data) => this.emitMessage(data));
        this.messageRouter.onAssets((data) => this.emitMessage(data));
    }

    get isConnected(): boolean {
        return this.statusManager.isConnected && this.authManager.authenticated;
    }

    get currentStatus(): WSStatus {
        return this.statusManager.currentStatus;
    }

    get currentSessionAddress(): string | null {
        return this.authManager.currentSessionAddress;
    }

    get hasPendingChallenge(): boolean {
        return this.challengeManager.hasPendingChallenge;
    }

    get sessionSigner() {
        const context = this.authManager.authContext;
        if (!context?.sessionKey) {
            return null;
        }
        return createECDSAMessageSigner(context.sessionKey.privateKey as `0x${string}`);
    }

    onStatusChange(listener: (status: WSStatus) => void): () => void {
        return this.eventEmitter.status.add(listener);
    }

    onMessage(listener: (message: any) => void): () => void {
        return this.eventEmitter.message.add(listener);
    }

    onError(listener: (error: Error) => void): () => void {
        return this.eventEmitter.error.add(listener);
    }

    async connect(walletAddress: string, privateKey: string): Promise<void> {
        if (this.connectionManager.destroyed) {
            throw new Error('Client has been destroyed');
        }

        if (this.statusManager.isConnecting || this.isConnected) {
            return;
        }

        if (this.connectionManager.hasUserRejectedAuth) {
            throw new UserRejectedError('User previously rejected authentication');
        }

        if (!privateKey) {
            throw new Error('Private key is required for server-side authentication');
        }

        this.statusManager.setStatus('connecting');

        try {
            await this.authManager.initializeContext(walletAddress, privateKey);
            await this.wsManager.connect();
            await this.authenticate();

            if (!this.statusManager.isPendingAuth) {
                this.startPingInterval();
                this.statusManager.setStatus('connected');
                this.connectionManager.resetRetryCount();
                this.callbacks.onConnect?.();
                this.callbacks.onAuthSuccess?.();
            }
        } catch (error) {
            this.connectionManager.handleConnectionError(error as Error);
            this.handleConnectionError(error as Error);
            throw error;
        }
    }

    disconnect(): void {
        this.cleanup();
        this.wsManager.close();
        this.statusManager.setStatus('disconnected');
        this.authManager.reset();
        this.callbacks.onDisconnect?.();
    }

    destroy(): void {
        this.disconnect();
        this.statusManager.destroy();
        this.wsManager.destroy();
        this.authManager.destroy();
        this.challengeManager.destroy();
        this.messageRouter.destroy();
        this.connectionManager.destroy();
        this.eventEmitter.clear();
    }

    async ping(): Promise<void> {
        if (!this.wsManager.isOpen) {
            throw new Error('WebSocket not connected');
        }

        const context = this.authManager.authContext;
        if (!context?.sessionKey) {
            throw new Error('Session key not available');
        }

        const sessionSigner = createECDSAMessageSigner(context.sessionKey.privateKey as `0x${string}`);
        const pingMessage = await createPingMessage(sessionSigner);
        this.wsManager.send(pingMessage);
    }

    send(data: any): void {
        if (!this.isConnected) {
            throw new Error('Not connected to Nitrolite WebSocket');
        }

        const message = typeof data === 'string' ? data : JSON.stringify(data);
        this.wsManager.send(message);
    }

    async approveChallenge(): Promise<void> {
        if (!this.challengeManager.hasPendingChallenge) {
            throw new Error('No pending challenge to approve');
        }

        if (!this.wsManager.isOpen) {
            throw new Error('WebSocket not connected');
        }

        // Remove the inProgress check here since challenge handling is expected during auth flow

        logger.debug(`Challenge: ${JSON.stringify(this.challengeManager.challenge, null, 2)}`);

        try {
            await this.authManager.authenticate(
                this.wsManager,
                this.config.requestTimeout,
                this.challengeManager.challenge,
                this.challengeManager.rawMessage || undefined,
            );

            this.challengeManager.clearChallenge();
            this.startPingInterval();
            this.statusManager.setStatus('connected');
            this.connectionManager.resetRetryCount();
            this.callbacks.onConnect?.();
            this.callbacks.onAuthSuccess?.();
        } catch (error) {
            logger.error('❌ Challenge approval failed:', error);
            this.callbacks.onVerifyFailed?.(error instanceof Error ? error.message : 'Challenge approval failed');
        }
    }

    rejectChallenge(): void {
        this.challengeManager.clearChallenge();
        this.connectionManager.resetUserRejection();
        this.statusManager.setStatus('disconnected');
        this.callbacks.onAuthFailed?.('Authentication rejected');
    }

    private async authenticate(): Promise<void> {
        if (!this.wsManager.isOpen) {
            throw new Error('WebSocket connection is not established');
        }

        await this.authManager.sendAuthRequest(this.wsManager);
    }

    private handleMessage(event: WSMessageEvent): void {
        const dataStr = event.data.toString();

        // Log all messages received from clearnode WebSocket connection
        logger.debug('📨 Received message from clearnode WebSocket');
        logger.debug(`Raw message: ${dataStr}`);

        // Check for token expiration first
        try {
            const rawJsonMessage = JSON.parse(dataStr);
            if (this.authManager.checkForTokenExpiration(rawJsonMessage)) {
                return;
            }
        } catch {
            // Continue to message routing
        }

        // Route the message to appropriate handlers
        this.messageRouter.routeMessage(event);
    }

    private handleAuthChallenge(data: any): void {
        logger.debug(`Challenge data: ${JSON.stringify(data, null, 2)}`);

        if (!this.authManager.authenticated) {
            this.challengeManager.setChallenge(data, JSON.stringify(data));
            // Note: challenge is handled by global client callback, not here
        } else {
            logger.warn(`Ignoring auth_challenge - already authenticated: ${this.authManager.authenticated}`);
        }
    }

    private handleAuthVerify(data: any): void {
        logger.debug(`Auth verify data: ${JSON.stringify(data, null, 2)}`);

        if (!this.authManager.authenticated && !this.authManager.inProgress) {
            const result = this.authManager.handleAuthResponse(data);

            if (result.success) {
                this.startPingInterval();
                this.statusManager.setStatus('connected');
                this.connectionManager.resetRetryCount();
                this.callbacks.onConnect?.();
                this.callbacks.onAuthSuccess?.();
            } else {
                logger.error(`❌ Authentication verification failed: ${result.error}`);
                this.callbacks.onVerifyFailed?.(result.error || 'Authentication failed');
            }
        } else {
            if (this.authManager.authenticated) {
                logger.warn('⚠️  Ignoring auth_verify - already authenticated');
            } else if (this.authManager.inProgress) {
                logger.debug('Skipping message router processing - auth in progress');
            }
        }
    }

    private handleRPCError(data: any): void {
        if (this.authManager.checkForTokenExpiration(data.params?.error)) {
            return;
        }
        this.emitError(new Error('Nitrolite service error'));
    }

    private handleTokenExpiration(): void {
        this.statusManager.setStatus('connecting');

        const context = this.authManager.authContext;
        if (this.wsManager.isOpen && context) {
            this.authManager.sendAuthRequest(this.wsManager).catch((error) => {
                logger.error('❌ Token expiration re-auth failed:', error);
                this.emitError(new Error(`Re-authentication failed: ${error.message}`));
            });
        } else {
            logger.error('❌ Cannot re-authenticate: missing connection or context');
            this.emitError(new Error('Cannot re-authenticate: connection or context unavailable'));
        }
    }

    private handleWSError(error: any): void {
        logger.error('WebSocket error:', error);
        this.emitError(new Error('WebSocket connection error'));
    }

    private handleDisconnection(): void {
        this.cleanup();
        this.authManager.reset();

        if (!this.connectionManager.destroyed && this.connectionManager.shouldReconnect()) {
            this.scheduleReconnect();
        } else {
            this.statusManager.setStatus('disconnected');
            this.callbacks.onDisconnect?.();
        }
    }

    private handleConnectionError(error: Error): void {
        this.emitError(error);
        this.callbacks.onError?.(error);

        if (this.connectionManager.shouldReconnect()) {
            this.scheduleReconnect();
        } else {
            this.statusManager.setStatus('failed');
        }
    }

    private scheduleReconnect(): void {
        this.statusManager.setStatus('reconnecting');

        this.connectionManager.scheduleReconnect(async () => {
            const context = this.authManager.authContext;
            if (context) {
                await this.connect(context.walletAddress, context.privateKey);
            }
        });
    }

    private startPingInterval(): void {
        const context = this.authManager.authContext;
        this.connectionManager.startPingInterval(context?.sessionKey || null, (data: string) =>
            this.wsManager.send(data),
        );
    }

    private cleanup(): void {
        this.connectionManager.cleanup();
        this.challengeManager.clearChallenge();
    }

    private emitMessage(message: any): void {
        this.eventEmitter.message.emit(message);
        this.callbacks.onMessage?.(message);
    }

    private emitError(error: Error): void {
        this.eventEmitter.error.emit(error);
    }

    resetRejectionState(): void {
        this.connectionManager.resetUserRejection();
    }

    handleChallengeMessage(challengeData: any): void {
        if (!this.authManager.authenticated && challengeData.method === RPCMethod.AuthChallenge) {
            this.challengeManager.setChallenge(challengeData);
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
    if (globalClient) {
        return globalClient;
    }

    if (!config.walletPrivateKey) {
        throw new Error('WALLET_PRIVATE_KEY is required for Nitrolite client initialization');
    }

    const wallet = new Wallet(config.walletPrivateKey);
    const walletAddress = wallet.address;

    logger.info(`💼 Wallet address: ${walletAddress.slice(0, 6)}...${walletAddress.slice(-4)}`);
    logger.info('🌐 Yellow WebSocket URL:', config.yellowWsUrl);
    logger.info('🏷️  App name:', config.vApp.name);
    logger.info('🔍 App scope:', config.vApp.scope);

    globalClient = createNitroliteWebSocketClient(
        {
            wsUrl: config.yellowWsUrl,
        },
        {
            onConnect: () => {},
            onDisconnect: () => {
                logger.warn('❌ Nitrolite WebSocket disconnected');
            },
            onError: (error) => {
                logger.error('💥 Nitrolite WebSocket error:', error.message);
            },
            onAuthSuccess: () => {},
            onAuthFailed: (error) => {
                logger.error('🔒 Nitrolite authentication failed:', error);
            },
            onChallengeReceived: (challenge) => {
                logger.debug('Auth challenge received for processing');
                // Auto-approve challenge for server (like frontend manual approval)
                if (globalClient && globalClient.hasPendingChallenge) {
                    globalClient.approveChallenge().catch((error) => {
                        logger.error('❌ Failed to auto-approve challenge:', error);
                    });
                }
            },
            onMessage: (message) => {
                logger.debug('📨 Nitrolite message received');
            },
        },
    );

    // Connect the client
    await globalClient.connect(walletAddress, config.walletPrivateKey);

    return globalClient;
}

export function getNitroliteClient(): NitroliteWebSocketClient | null {
    return globalClient;
}
