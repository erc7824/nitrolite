import {
    createAuthRequestMessage,
    createAuthVerifyMessageWithJWT,
    createAuthVerifyMessage,
    createEIP712AuthMessageSigner,
    parseAnyRPCResponse,
    type RPCResponse,
    type AuthChallengeResponse,
    RPCMethod,
    type AuthRequestParams,
} from '@erc7824/nitrolite';
import { Wallet } from 'ethers';
import { WebSocket } from 'ws';
import { UserRejectedError, NitroliteAuthContext } from '../../types/index.js';
import { config } from '../../config/index.js';
import { logger } from '../../utils/logger.js';

export const JWT_STORAGE_KEY = 'jwt_token';
const AUTH_TIMEOUT = 30000; // 30 seconds
const expire = String(Math.floor(Date.now() / 1000) + 24 * 60 * 60);

// In-memory JWT storage for server
let jwtTokenStore: string | null = null;

/**
 * EIP-712 domain for authentication
 */
const getAuthDomain = () => ({
    name: config.vApp.name,
});

/**
 * Creates the authentication message structure
 */
export const createAuthMessage = (
    walletAddress: string,
    sessionAddress: string,
    totalAmount: string = '0',
    challenge?: string,
): AuthRequestParams => {
    const message: AuthRequestParams = {
        wallet: walletAddress as `0x${string}`,
        participant: sessionAddress as `0x${string}`,
        app_name: config.vApp.name,
        expire: expire,
        scope: config.vApp.scope,
        application: walletAddress as `0x${string}`,
        allowances: [
            {
                asset: config.asset,
                amount: (isNaN(Number(totalAmount)) ? '0' : totalAmount)?.toString(),
            },
        ],
    };

    if (challenge) {
        (message as any).challenge = challenge;
    }

    return message;
};

/**
 * JWT token management for server
 */
export const getStoredJWTToken = (): string | null => {
    return jwtTokenStore;
};

export const storeJWTToken = (token: string): void => {
    jwtTokenStore = token;
    logger.info('✅ JWT token stored in memory');
    logger.info('🔑 JWT Token:', token.substring(0, 50) + '...' + token.substring(token.length - 10));
};

export const removeJWTToken = (): void => {
    logger.info('Removing JWT token due to expiration/error');
    jwtTokenStore = null;
};

export const clearJWTToken = (): void => {
    logger.info('Force clearing JWT token');
    jwtTokenStore = null;
};

export const isJWTTokenValid = (token: string): boolean => {
    if (typeof token !== 'string' || token.trim() === '') {
        return false;
    }

    try {
        const payload = JSON.parse(Buffer.from(token.split('.')[1], 'base64').toString());
        const currentTime = Math.floor(Date.now() / 1000);
        return payload.exp > currentTime;
    } catch (error) {
        logger.warn('JWT token validation failed:', error);
        return false;
    }
};

/**
 * Sends initial authentication request
 */
export const sendAuthRequest = async (ws: WebSocket, authContext: NitroliteAuthContext): Promise<void> => {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
        throw new Error('WebSocket not connected');
    }

    const jwtToken = getStoredJWTToken();
    let authRequest: string;

    try {
        if (jwtToken && isJWTTokenValid(jwtToken)) {
            logger.info('🔐 Sending auth_verify with existing JWT token');
            logger.debug('JWT Token preview:', jwtToken.substring(0, 30) + '...');
            authRequest = await createAuthVerifyMessageWithJWT(jwtToken);
        } else {
            logger.info('🆕 No valid JWT token found, sending fresh auth_request');
            const authMessage = createAuthMessage(authContext.walletAddress, authContext.sessionKey.address);
            authRequest = await createAuthRequestMessage(authMessage as any);
        }

        logger.debug('Sending auth message via WebSocket:', authRequest.substring(0, 100) + '...');
        ws.send(authRequest);
    } catch (error) {
        throw new Error(`Auth request failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
};

/**
 * Handles auth challenge and sends verify message using server-side ECDSA signing
 */
export const handleAuthChallenge = async (
    ws: WebSocket,
    challengeResponse: AuthChallengeResponse,
    authContext: NitroliteAuthContext,
    rawMessage?: string,
): Promise<void> => {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
        throw new Error('WebSocket not connected for auth verify');
    }

    const challenge = challengeResponse.params?.challengeMessage;
    logger.info('🤝 Handling auth challenge with challenge message:', challenge);
    const authMessage = createAuthMessage(authContext.walletAddress, authContext.sessionKey.address, '0', challenge);
    logger.debug('📝 Created auth message:', JSON.stringify(authMessage));

    try {
        // Create wallet from private key for signing
        const wallet = new Wallet(authContext.privateKey);

        // Create a server-side wallet client adapter for EIP-712 signing
        const serverWalletClient = {
            account: {
                address: wallet.address as `0x${string}`,
            },
            signTypedData: async ({ domain, types, message }: any) => {
                // Use ethers wallet to sign EIP-712 message
                return (await wallet.signTypedData(domain, types, message)) as `0x${string}`;
            },
        };

        // Create EIP-712 message signer (like the frontend does)
        const eip712SigningFunction = createEIP712AuthMessageSigner(
            serverWalletClient as any,
            {
                scope: authMessage.scope,
                application: authMessage.application as `0x${string}`,
                participant: authMessage.participant as `0x${string}`,
                expire: authMessage.expire,
                allowances: authMessage.allowances.map((allowance) => ({
                    asset: allowance.asset,
                    amount: allowance.amount,
                })),
            },
            getAuthDomain(),
        );

        // Use raw message if provided, otherwise fall back to challengeResponse
        const messageForVerify = rawMessage
            ? (parseAnyRPCResponse(rawMessage) as AuthChallengeResponse)
            : challengeResponse;

        const authVerifyMessage = await createAuthVerifyMessage(eip712SigningFunction, messageForVerify);
        logger.info('📤 Sending auth_verify message to Yellow network (using EIP-712 signature)');

        ws.send(authVerifyMessage);
    } catch (error) {
        logger.error('handleAuthChallenge error details:', {
            errorType: error?.constructor?.name,
            errorMessage: error instanceof Error ? error.message : String(error),
            errorStack: error instanceof Error ? error.stack : undefined,
        });

        if (error instanceof Error && UserRejectedError.isUserRejection(error)) {
            throw new UserRejectedError(error.message);
        }

        throw new Error(
            `Auth verify message creation failed: ${error instanceof Error ? error.message : 'Unknown error'}`,
        );
    }
};

/**
 * Checks if the error indicates token expiration
 */
export const isTokenExpiredError = (error: any): boolean => {
    const checkMessage = (msg: string) => {
        const isExpired =
            msg.includes('token is expired') ||
            msg.includes('token has invalid claims') ||
            msg.includes('token is malformed') ||
            msg.includes('invalid JWT token') ||
            msg.includes('invalid JWT');
        if (isExpired) {
            logger.debug('Token error detected in message:', msg);
        }
        return isExpired;
    };

    if (typeof error === 'string') {
        return checkMessage(error);
    }

    if (error && typeof error === 'object') {
        if (error.error && typeof error.error === 'string') {
            return checkMessage(error.error);
        }

        if (Array.isArray(error) && error.length > 0) {
            return error.some((e) => {
                if (typeof e === 'string') {
                    return checkMessage(e);
                }
                return e && typeof e === 'object' && e.error && checkMessage(e.error);
            });
        }
    }

    return false;
};

/**
 * Processes authentication response
 */
export const processAuthResponse = (
    response: RPCResponse,
): {
    success: boolean;
    jwtToken?: string;
    error?: string;
    tokenExpired?: boolean;
} => {
    logger.info('🔍 Processing auth response:', JSON.stringify(response, null, 2));

    if (response.method === RPCMethod.AuthVerify && response.params?.success) {
        const result: { success: boolean; jwtToken?: string } = { success: true };

        if (response.params.jwtToken) {
            result.jwtToken = response.params.jwtToken;
            logger.info('🎉 Authentication successful! Received JWT token');
            storeJWTToken(response.params.jwtToken);
        } else {
            logger.warn('⚠️ Authentication successful but no JWT token received');
        }

        return result;
    }

    if (response.method === RPCMethod.Error) {
        const errorMsg = response.params?.error || 'Authentication failed';
        const tokenExpired = isTokenExpiredError(response.params?.error);

        removeJWTToken();

        return {
            success: false,
            error: String(errorMsg),
            tokenExpired,
        };
    }

    if (response.method === RPCMethod.AuthVerify && !response.params?.success) {
        return { success: false, error: 'Authentication verification failed' };
    }

    return { success: false, error: 'Unknown authentication response' };
};

export const authenticateWithNitrolite = async (
    ws: WebSocket,
    authContext: NitroliteAuthContext,
    timeout: number = AUTH_TIMEOUT,
    pendingChallenge?: any,
    rawMessage?: string,
): Promise<void> => {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
        throw new Error('WebSocket not connected');
    }

    return new Promise((resolve, reject) => {
        let authTimeoutId: NodeJS.Timeout | null = null;
        let isResolved = false;

        const cleanup = () => {
            if (authTimeoutId) {
                clearTimeout(authTimeoutId);
                authTimeoutId = null;
            }
            ws.removeEventListener('message', handleAuthMessage);
        };

        const resolveAuth = (result?: any) => {
            if (isResolved) return;
            isResolved = true;
            cleanup();
            resolve(result);
        };

        const rejectAuth = (error: Error) => {
            if (isResolved) return;
            isResolved = true;
            cleanup();
            reject(error);
        };

        const resetTimeout = () => {
            if (authTimeoutId) {
                clearTimeout(authTimeoutId);
            }
            authTimeoutId = setTimeout(() => {
                rejectAuth(new Error('Authentication timeout'));
            }, timeout);
        };

        const handleAuthMessage = async (event: MessageEvent) => {
            try {
                // Parse and filter auth-related messages
                let rawMessage;
                try {
                    rawMessage = parseAnyRPCResponse(event.data);
                } catch {
                    logger.error('failed to parse incoming event', event.data);
                    return;
                }

                // Skip non-auth messages during authentication
                if (![RPCMethod.AuthChallenge, RPCMethod.AuthVerify, 'error'].includes(rawMessage.method)) {
                    return;
                }

                // Only parse with nitrolite SDK if it's an auth-related message
                const response = parseAnyRPCResponse(event.data);

                if (response.method === RPCMethod.AuthChallenge) {
                    resetTimeout();
                    try {
                        await handleAuthChallenge(ws, response, authContext);
                    } catch (error) {
                        rejectAuth(
                            new Error(
                                `Challenge handling failed: ${error instanceof Error ? error.message : 'Unknown error'}`,
                            ),
                        );
                    }
                } else if (response.method === RPCMethod.AuthVerify || response.method === RPCMethod.Error) {
                    const result = processAuthResponse(response);
                    if (result.success) {
                        resolveAuth();
                    } else {
                        rejectAuth(new Error(result.error || 'Authentication failed'));
                    }
                }
            } catch (error) {
                // Skip unsupported RPC methods
                if (error instanceof Error && error.message.includes('Unsupported RPC method')) {
                    return;
                }

                rejectAuth(
                    new Error(
                        `Auth message processing failed: ${error instanceof Error ? error.message : 'Unknown error'}`,
                    ),
                );
            }
        };

        ws.addEventListener('message', handleAuthMessage);
        resetTimeout();

        // If we have a pending challenge, handle it directly
        if (pendingChallenge) {
            handleAuthChallenge(ws, pendingChallenge, authContext, rawMessage)
                .then(() => {
                    logger.debug('handleAuthChallenge completed successfully');
                })
                .catch((error) => {
                    logger.error('handleAuthChallenge failed with error:', {
                        errorType: error?.constructor?.name,
                        errorMessage: error instanceof Error ? error.message : String(error),
                        errorStack: error instanceof Error ? error.stack : undefined,
                    });
                    rejectAuth(error instanceof Error ? error : new Error('Challenge handling failed'));
                });
        } else {
            sendAuthRequest(ws, authContext)
                .then(() => {
                    logger.debug('Auth request sent successfully');
                })
                .catch((error) => {
                    logger.error('sendAuthRequest failed:', error);
                    rejectAuth(
                        new Error(`Auth request failed: ${error instanceof Error ? error.message : 'Unknown error'}`),
                    );
                });
        }
    });
};

/**
 * Parses the specific error format from Nitrolite WebSocket responses
 */
export const parseNitroliteError = (response: any): { isTokenExpired: boolean; errorMessage: string } => {
    try {
        if (typeof response === 'string') {
            try {
                const parsed = JSON.parse(response);
                return parseNitroliteError(parsed);
            } catch {
                const isTokenError = isTokenExpiredError(response);
                return {
                    isTokenExpired: isTokenError,
                    errorMessage: response,
                };
            }
        }

        if (response && response.res && Array.isArray(response.res)) {
            const [, method, params] = response.res;

            if (method === 'error' && Array.isArray(params) && params.length > 0) {
                const errorData = params[0];

                if (typeof errorData === 'string') {
                    return {
                        isTokenExpired: isTokenExpiredError(errorData),
                        errorMessage: errorData,
                    };
                } else if (errorData && typeof errorData === 'object' && errorData.error) {
                    return {
                        isTokenExpired: isTokenExpiredError(errorData.error),
                        errorMessage: errorData.error,
                    };
                }
            }
        }

        if (response && response.error) {
            return {
                isTokenExpired: isTokenExpiredError(response.error),
                errorMessage: response.error,
            };
        }

        if (Array.isArray(response)) {
            for (const item of response) {
                if (typeof item === 'string') {
                    const isTokenError = isTokenExpiredError(item);
                    if (isTokenError) {
                        return {
                            isTokenExpired: isTokenError,
                            errorMessage: item,
                        };
                    }
                } else if (item && typeof item === 'object' && item.error) {
                    return {
                        isTokenExpired: isTokenExpiredError(item.error),
                        errorMessage: item.error,
                    };
                }
            }
        }

        return {
            isTokenExpired: isTokenExpiredError(response),
            errorMessage: typeof response === 'string' ? response : JSON.stringify(response) || 'Unknown error',
        };
    } catch (error) {
        logger.warn('Error parsing Nitrolite WebSocket response:', error);
        return {
            isTokenExpired: false,
            errorMessage: 'Error parsing response',
        };
    }
};
