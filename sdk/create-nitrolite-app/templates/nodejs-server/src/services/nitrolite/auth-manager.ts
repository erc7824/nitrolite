import { Wallet } from 'ethers';
import type { NitroliteAuthContext, SessionKey } from '../../types/index.js';
import { 
    sendAuthRequest,
    authenticateWithNitrolite,
    isTokenExpiredError,
    processAuthResponse,
    clearJWTToken,
    parseNitroliteError,
    getStoredJWTToken,
    isJWTTokenValid
} from './auth.js';
import { logger } from '../../utils/logger.js';
import { EventEmitter } from './event-emitter.js';

export interface AuthEvents {
    success: void;
    failed: string;
    tokenExpired: void;
}

export class AuthenticationManager {
    private isAuthenticated = false;
    private sessionKey: SessionKey | null = null;
    private walletAddress: string | null = null;
    private privateKey: string | null = null;
    private authInProgress = false;
    
    private readonly successEmitter = new EventEmitter<void>();
    private readonly failedEmitter = new EventEmitter<string>();
    private readonly tokenExpiredEmitter = new EventEmitter<void>();

    get authenticated(): boolean {
        return this.isAuthenticated;
    }

    get currentSessionAddress(): string | null {
        return this.sessionKey?.address || null;
    }

    get inProgress(): boolean {
        return this.authInProgress;
    }

    get authContext(): NitroliteAuthContext | null {
        if (!this.walletAddress || !this.sessionKey || !this.privateKey) {
            return null;
        }
        
        return {
            walletAddress: this.walletAddress,
            sessionKey: this.sessionKey,
            privateKey: this.privateKey,
        };
    }

    async initializeContext(walletAddress: string, privateKey: string): Promise<void> {
        if (!privateKey) {
            throw new Error('Private key is required for server-side authentication');
        }

        this.walletAddress = walletAddress;
        this.privateKey = privateKey;
        
        const wallet = new Wallet(privateKey);
        this.sessionKey = {
            privateKey: privateKey,
            address: wallet.address,
        };
    }

    async sendAuthRequest(wsManager: any): Promise<void> {
        const context = this.authContext;
        if (!context) {
            throw new Error('Authentication context not initialized');
        }
        
        if (this.authInProgress) {
            logger.debug('Authentication already in progress, skipping duplicate request');
            return;
        }

        const ws = wsManager.rawWebSocket;
        if (!ws) {
            throw new Error('WebSocket not available');
        }

        // Check if we have a valid JWT token to use
        const storedJWT = getStoredJWTToken();
        if (storedJWT && isJWTTokenValid(storedJWT)) {
            logger.info('🔑 Using stored JWT token for authentication');
            logger.info(`JWT Token: ${storedJWT}`);
        } else {
            logger.info('🆕 No valid JWT token found, proceeding with standard auth request');
            if (storedJWT) {
                logger.info('🗑️  Clearing expired JWT token');
                clearJWTToken();
            }
        }

        this.authInProgress = true;
        try {
            await sendAuthRequest(ws, context);
        } catch (error) {
            this.authInProgress = false;
            throw error;
        }
    }

    async authenticate(wsManager: any, timeout: number, pendingChallenge?: any, rawMessage?: string): Promise<void> {
        logger.info('🔐🔐 AUTH MANAGER AUTHENTICATE CALLED');
        logger.info(`Pending challenge: ${JSON.stringify(pendingChallenge, null, 2)}`);
        logger.info(`Raw message: ${rawMessage}`);
        
        const context = this.authContext;
        if (!context) {
            throw new Error('Authentication context not initialized');
        }

        // Allow challenge processing during auth flow
        if (this.authInProgress && !pendingChallenge) {
            logger.warn('⚠️  Authentication already in progress, skipping duplicate');
            return;
        }

        const ws = wsManager.rawWebSocket;
        if (!ws) {
            throw new Error('WebSocket not available');
        }

        logger.info('🚀 Starting authenticateWithNitrolite...');
        this.authInProgress = true;
        try {
            const authResult = await authenticateWithNitrolite(ws, context, timeout, pendingChallenge, rawMessage);
            logger.info('✅ authenticateWithNitrolite completed successfully');
            
            // Print JWT token if received
            if (authResult.jwtToken) {
                logger.info('🎉 SUCCESS! JWT Token received and stored:');
                logger.info('🔑 ═══════════════════════════════════════════════════════════');
                logger.info(`🔑 JWT TOKEN: ${authResult.jwtToken}`);
                logger.info('🔑 ═══════════════════════════════════════════════════════════');
                logger.info('✅ JWT Token will be reused for future authentication requests');
            } else {
                logger.info('✅ Authentication successful but no JWT token received');
            }
            
            this.isAuthenticated = true;
            this.successEmitter.emit();
        } catch (error) {
            logger.error('❌ authenticateWithNitrolite failed:', error);
            this.failedEmitter.emit(error instanceof Error ? error.message : 'Authentication failed');
            throw error;
        } finally {
            this.authInProgress = false;
            logger.info('🔄 Auth progress reset to false');
        }
    }

    handleAuthResponse(response: any): { success: boolean; error?: string; tokenExpired?: boolean } {
        logger.info('🔍 Processing authentication response:');
        logger.info(`Auth response: ${JSON.stringify(response, null, 2)}`);
        
        const result = processAuthResponse(response);
        
        if (result.success) {
            this.isAuthenticated = true;
            
            // Always print JWT token when received - this is critical for verification
            if (result.jwtToken) {
                logger.info('🎉 SUCCESS! JWT Token received and stored:');
                logger.info('🔑 ═══════════════════════════════════════════════════════════');
                logger.info(`🔑 JWT TOKEN: ${result.jwtToken}`);
                logger.info('🔑 ═══════════════════════════════════════════════════════════');
                logger.info('✅ JWT Token will be reused for future authentication requests');
            } else {
                logger.info('✅ Authentication successful but no JWT token received');
            }
            
            this.successEmitter.emit();
        } else if (result.tokenExpired) {
            logger.info('⚠️  JWT Token expired, clearing and triggering re-authentication');
            this.handleTokenExpiration();
        } else if (result.error) {
            logger.error(`❌ Authentication failed: ${result.error}`);
            this.failedEmitter.emit(result.error);
        }
        
        return result;
    }

    checkForTokenExpiration(response: any): boolean {
        const nitroliteError = parseNitroliteError(response);
        if (nitroliteError.isTokenExpired || isTokenExpiredError(response.params?.error)) {
            this.handleTokenExpiration();
            return true;
        }
        return false;
    }

    private handleTokenExpiration(): void {
        logger.info('Token expired - clearing auth state');
        clearJWTToken();
        this.isAuthenticated = false;
        this.authInProgress = false;
        this.tokenExpiredEmitter.emit();
    }

    setAuthenticated(authenticated: boolean): void {
        this.isAuthenticated = authenticated;
        if (authenticated) {
            this.successEmitter.emit();
        }
    }

    reset(): void {
        this.isAuthenticated = false;
        this.authInProgress = false;
    }

    onAuthSuccess(listener: () => void): () => void {
        return this.successEmitter.add(listener);
    }

    onAuthFailed(listener: (error: string) => void): () => void {
        return this.failedEmitter.add(listener);
    }

    onTokenExpired(listener: () => void): () => void {
        return this.tokenExpiredEmitter.add(listener);
    }

    destroy(): void {
        this.reset();
        this.successEmitter.clear();
        this.failedEmitter.clear();
        this.tokenExpiredEmitter.clear();
    }
}