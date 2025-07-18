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
import type { YellowAuthContext } from './types';
import { UserRejectedError } from './types';
import { config } from '../../utils/env';
import type { Address } from 'viem';

export const JWT_STORAGE_KEY = 'jwt_token';
const AUTH_TIMEOUT = 30000; // 30 seconds
// should be global as its should be same value for all auth messages
const expire = String(Math.floor(Date.now() / 1000) + 24 * 60 * 60);

/**
 * EIP-712 domain for Yellow authentication
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
        wallet: walletAddress as Address,
        participant: sessionAddress as Address,
        app_name: config.vApp.name,
        expire: expire,
        scope: config.vApp.scope,
        application: walletAddress as Address,
        allowances: [
            {
                asset: config.asset,
                amount: (isNaN(Number(totalAmount)) ? '0' : totalAmount)?.toString(),
            },
        ],
    };

    // Include challenge if provided
    if (challenge) {
        (message as any).challenge = challenge;
    }

    return message;
};

/**
 * JWT token management with error handling
 */
export const getStoredJWTToken = (): string | null => {
    try {
        return localStorage.getItem(JWT_STORAGE_KEY);
    } catch {
        return null;
    }
};

export const storeJWTToken = (token: string): void => {
    try {
        localStorage.setItem(JWT_STORAGE_KEY, token);
    } catch {
        // Storage failed - continue without JWT caching
    }
};

export const removeJWTToken = (): void => {
    try {
        console.log('🗑️ Removing JWT token from localStorage due to expiration/error');
        localStorage.removeItem(JWT_STORAGE_KEY);
    } catch {
        // Removal failed - not critical
        console.warn('Failed to remove JWT token from localStorage');
    }
};

/**
 * Force clear any existing JWT token - useful for malformed tokens
 */
export const clearJWTToken = (): void => {
    try {
        console.log('🧹 Force clearing JWT token from localStorage');
        localStorage.removeItem(JWT_STORAGE_KEY);
    } catch {
        console.warn('Failed to force clear JWT token from localStorage');
    }
};

export const isJWTTokenValid = (token: string): boolean => {
    // Basic validation: check if token is a non-empty string
    if (typeof token !== 'string' || token.trim() === '') {
        return false;
    }

    // Check for expiration
    try {
        const payload = JSON.parse(atob(token.split('.')[1]));

        const currentTime = Math.floor(Date.now() / 1000);
        return payload.exp > currentTime;
    } catch (error) {
        console.warn('JWT token validation failed:', error);
        return false;
    }
};

/**
 * Sends initial authentication request
 */
export const sendAuthRequest = async (ws: WebSocket, authContext: YellowAuthContext): Promise<void> => {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
        throw new Error('WebSocket not connected');
    }

    const jwtToken = getStoredJWTToken();
    let authRequest: string;

    try {
        if (jwtToken && isJWTTokenValid(jwtToken)) {
            console.log('Sending auth_verify with existing JWT token');
            authRequest = await createAuthVerifyMessageWithJWT(jwtToken);
        } else {
            console.log('No JWT token found, sending fresh auth_request');
            const authMessage = createAuthMessage(authContext.walletAddress, authContext.sessionKey.address);
            authRequest = await createAuthRequestMessage(authMessage as any);
        }

        console.log('Sending auth message via WebSocket:', authRequest.substring(0, 100) + '...');
        ws.send(authRequest);
    } catch (error) {
        throw new Error(`Auth request failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
};

/**
 * Handles auth challenge and sends verify message using WalletClient EIP-712 signing
 */
export const handleAuthChallenge = async (
    ws: WebSocket,
    challengeResponse: AuthChallengeResponse,
    authContext: YellowAuthContext,
    rawMessage?: string,
): Promise<void> => {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
        throw new Error('WebSocket not connected for auth verify');
    }

    const challenge = challengeResponse.params?.challengeMessage;
    const authMessage = createAuthMessage(authContext.walletAddress, authContext.sessionKey.address, '0', challenge);

    try {
        // Verify we have wallet client for EIP-712 signing
        if (!authContext.walletClient) {
            throw new Error('No wallet client available for EIP-712 signing');
        }

        const eip712SigningFunction = createEIP712AuthMessageSigner(
            // @ts-ignore
            authContext.walletClient,
            {
                scope: authMessage.scope,
                application: authMessage.application as `0x${string}`,
                participant: authMessage.participant as `0x${string}`,
                expire: expire,
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

        ws.send(authVerifyMessage);
    } catch (error) {
        console.error('handleAuthChallenge error details:', {
            errorType: error?.constructor?.name,
            errorMessage: error instanceof Error ? error.message : String(error),
            errorStack: error instanceof Error ? error.stack : undefined,
        });

        // Check if user rejected the signing request
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
            console.log('🔍 Token error detected in message:', msg);
        }
        return isExpired;
    };

    if (typeof error === 'string') {
        return checkMessage(error);
    }

    if (error && typeof error === 'object') {
        // Check for nested error structures
        if (error.error && typeof error.error === 'string') {
            return checkMessage(error.error);
        }

        // Check for array of errors
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
    if (response.method === RPCMethod.AuthVerify && response.params?.success) {
        const result: { success: boolean; jwtToken?: string } = { success: true };

        if (response.params.jwtToken) {
            result.jwtToken = response.params.jwtToken;
            storeJWTToken(response.params.jwtToken);
        }

        return result;
    }

    if (response.method === RPCMethod.Error) {
        const errorMsg = response.params?.error || 'Authentication failed';
        const tokenExpired = isTokenExpiredError(response.params?.error);

        // Always remove JWT on error, but especially on token expiration
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

/**
 * Main authentication function with proper error handling and timeout
 */
export const authenticateWithYellow = async (
    ws: WebSocket,
    authContext: YellowAuthContext,
    timeout: number = AUTH_TIMEOUT,
    pendingChallenge?: any,
    rawMessage?: string,
): Promise<void> => {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
        throw new Error('WebSocket not connected');
    }

    return new Promise((resolve, reject) => {
        let authTimeoutId: number | null = null;
        let isResolved = false;

        const cleanup = () => {
            if (authTimeoutId) {
                window.clearTimeout(authTimeoutId);
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
                window.clearTimeout(authTimeoutId);
            }
            authTimeoutId = window.setTimeout(() => {
                rejectAuth(new Error('Authentication timeout'));
            }, timeout);
        };

        const handleAuthMessage = async (event: MessageEvent) => {
            try {
                // Parse and filter auth-related messages
                let rawMessage;
                try {
                    rawMessage = JSON.parse(event.data);
                } catch {
                    return; // Skip non-JSON messages
                }

                // Skip non-auth messages during authentication
                if (
                    rawMessage.method &&
                    ![RPCMethod.AuthChallenge, RPCMethod.AuthVerify, 'error'].includes(rawMessage.method)
                ) {
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

        // Set up message listener and timeout
        ws.addEventListener('message', handleAuthMessage);
        resetTimeout();

        // If we have a pending challenge, handle it directly
        if (pendingChallenge) {
            handleAuthChallenge(ws, pendingChallenge, authContext, rawMessage)
                .then(() => {
                    console.log('handleAuthChallenge completed successfully');
                })
                .catch((error) => {
                    console.error('handleAuthChallenge failed with error:', {
                        errorType: error?.constructor?.name,
                        errorMessage: error instanceof Error ? error.message : String(error),
                        errorStack: error instanceof Error ? error.stack : undefined,
                    });
                    rejectAuth(error instanceof Error ? error : new Error('Challenge handling failed'));
                });
        } else {
            // Send initial auth request
            sendAuthRequest(ws, authContext)
                .then(() => {
                    console.log('Auth request sent successfully');
                })
                .catch((error) => {
                    console.error('sendAuthRequest failed:', error);
                    rejectAuth(
                        new Error(`Auth request failed: ${error instanceof Error ? error.message : 'Unknown error'}`),
                    );
                });
        }
    });
};

/**
 * Parses the specific error format from Yellow WebSocket responses
 * Example: {"res":[1751872302081,"error",["invalid JWT token"],1751872299463],"sig":["0x..."]}
 * Example: {"res":[1750671978677,"error",[{"error":"token has invalid claims: token is expired"}],1750671978677],"sig":["0x..."]}
 */
export const parseYellowError = (response: any): { isTokenExpired: boolean; errorMessage: string } => {
    try {
        // Handle string responses (might be truncated JSON)
        if (typeof response === 'string') {
            try {
                const parsed = JSON.parse(response);
                return parseYellowError(parsed);
            } catch {
                // If it's not valid JSON, check if it contains error indicators
                const isTokenError = isTokenExpiredError(response);
                return {
                    isTokenExpired: isTokenError,
                    errorMessage: response,
                };
            }
        }

        // Check if response has the expected structure: {"res":[timestamp,"error",[errorData],timestamp],"sig":[...]}
        if (response && response.res && Array.isArray(response.res)) {
            const [, method, params] = response.res;

            if (method === 'error' && Array.isArray(params) && params.length > 0) {
                // Handle both string errors and object errors
                const errorData = params[0];

                if (typeof errorData === 'string') {
                    // Direct string error: ["invalid JWT token"]
                    return {
                        isTokenExpired: isTokenExpiredError(errorData),
                        errorMessage: errorData,
                    };
                } else if (errorData && typeof errorData === 'object' && errorData.error) {
                    // Object error: [{"error":"token has invalid claims: token is expired"}]
                    return {
                        isTokenExpired: isTokenExpiredError(errorData.error),
                        errorMessage: errorData.error,
                    };
                }
            }
        }

        // Check if it's a raw error object
        if (response && response.error) {
            return {
                isTokenExpired: isTokenExpiredError(response.error),
                errorMessage: response.error,
            };
        }

        // Check for error arrays (handle truncated or partial messages)
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

        // Fallback to direct error check
        return {
            isTokenExpired: isTokenExpiredError(response),
            errorMessage: typeof response === 'string' ? response : JSON.stringify(response) || 'Unknown error',
        };
    } catch (error) {
        console.warn('Error parsing Yellow WebSocket response:', error, 'Original response:', response);
        return {
            isTokenExpired: false,
            errorMessage: 'Error parsing response',
        };
    }
};
