import { createAuthRequestMessage, createAuthVerifyMessage, createAuthVerifyMessageWithJWT, createEIP712AuthMessageSigner } from '@erc7824/nitrolite';
import type { WalletSigner } from '../crypto';
import type { Hex } from 'viem';
import { getAddress } from 'viem';

/**
 * EIP-712 domain and types for auth_verify challenge
 */
const getAuthDomain = () => {
    return {
        name: 'Snake Game',
    };
};


const expire = String(Math.floor(Date.now() / 1000) + 24 * 60 * 60);


/**
 * Authenticates with the WebSocket server using:
 * 1. auth_request: empty signature with wallet address
 * 2. auth_verify: EIP-712 signature for challenge verification (with proper challenge extraction)
 *
 * @param ws - The WebSocket connection
 * @param walletClient - The wallet client for signing
 * @param signer - The local keys signer (session key)
 * @param timeout - Timeout in milliseconds for the entire process
 * @returns A Promise that resolves when authenticated
 */
export async function authenticate(
    ws: WebSocket,
    walletClient: any,
    signer: WalletSigner,
    timeout: number = 15000
): Promise<void> {
    if (!ws) throw new Error('WebSocket not connected');

    if (!walletClient) {
        throw new Error('No wallet client available for authentication');
    }

    const rawWalletAddress = walletClient.account?.address;

    if (!rawWalletAddress) {
        throw new Error('No wallet address available for authentication');
    }

    // Ensure the address is properly checksummed for EIP-55 compliance
    const walletAddress = getAddress(rawWalletAddress);

    console.log('Starting authentication with:');
    console.log('- Wallet address:', walletAddress);
    console.log('- Session key:', signer.address);
    console.log('- Empty signature for auth_request');
    console.log('- EIP-712 signature for auth_verify challenge (UUID only)');

    const authMessage = {
        wallet: walletAddress as Hex,
        participant: signer.address as Hex,
        app_name: 'Snake Game',
        expire: expire, // 24 hours in seconds
        scope: 'snake-game',
        application: walletAddress as Hex,
        allowances: [
            {
                symbol: 'usdc',
                amount: '0',
            },
        ],
    };

    // Step 1: Send auth_request with empty signature and wallet address
    try {
        let authRequest: string;
        let usingJWT = false;
        const jwtToken = window.localStorage.getItem('jwt_token');

        if (jwtToken) {
            console.log('JWT token found, attempting JWT authentication:', jwtToken);
            try {
                authRequest = await createAuthVerifyMessageWithJWT(jwtToken);
                usingJWT = true;
            } catch (jwtError) {
                console.warn('JWT auth failed, falling back to signer authentication:', jwtError);
                // Remove invalid JWT token
                window.localStorage.removeItem('jwt_token');
                authRequest = await createAuthRequestMessage(authMessage);
                usingJWT = false;
            }
        } else {
            console.log('No JWT token found, proceeding with challenge-response authentication');
            authRequest = await createAuthRequestMessage(authMessage);
            usingJWT = false;
        }

        console.log(`Sending auth_request (${usingJWT ? 'JWT' : 'challenge-response'}):`, authRequest);
        ws.send(authRequest);
    } catch (requestError) {
        console.error('Error creating auth_request:', requestError);
        throw new Error(`Failed to create auth_request: ${(requestError as Error).message}`);
    }

    return new Promise((resolve, reject) => {
        if (!ws) return reject(new Error('WebSocket not connected'));

        let authTimeoutId: number | null = null;

        const cleanup = () => {
            if (authTimeoutId) {
                clearTimeout(authTimeoutId);
                authTimeoutId = null;
            }
            ws.removeEventListener('message', handleAuthResponse);
        };

        const resetTimeout = () => {
            if (authTimeoutId) {
                clearTimeout(authTimeoutId);
            }
            authTimeoutId = setTimeout(() => {
                cleanup();
                reject(new Error('Authentication timeout'));
            }, timeout) as unknown as number;
        };

        const handleAuthResponse = async (event: MessageEvent) => {
            let response;

            try {
                response = JSON.parse(event.data);
                console.log('Received auth message:', response);
            } catch (error) {
                console.error('Error parsing auth response:', error);
                console.log('Raw auth message:', event.data);
                return;
            }

            try {
                // Check for challenge response: {"res": [id, "auth_challenge", {"challenge": "uuid"}, timestamp]}
                if (response.res && response.res[1] === 'auth_challenge') {
                    console.log('Received auth_challenge, preparing EIP-712 auth_verify...');
                    resetTimeout(); // Reset timeout while we process and send verify

                    try {
                        // Step 2: Create EIP-712 signing function for challenge verification
                        if (!walletClient) {
                            throw new Error('No wallet client available for EIP-712 signing');
                        }

                        console.log('Creating EIP-712 signing function...');
                        const eip712SigningFunction = createEIP712AuthMessageSigner(walletClient, {
                            scope: authMessage.scope,
                            application: authMessage.application,
                            participant: authMessage.participant,
                            expire: authMessage.expire,
                            allowances: authMessage.allowances.map((allowance) => ({
                                asset: allowance.symbol,
                                amount: allowance.amount.toString(),
                            })),
                        }, getAuthDomain());

                        console.log('Calling createAuthVerifyMessage...');
                        // Create and send verification message with EIP-712 signature
                        const authVerify = await createAuthVerifyMessage(
                            eip712SigningFunction,
                            event.data, // Pass the raw challenge response string/object
                        );

                        console.log('Sending auth_verify with EIP-712 signature');
                        ws.send(authVerify);
                        console.log('auth_verify sent successfully');
                    } catch (eip712Error) {
                        console.error('Error creating EIP-712 auth_verify:', eip712Error);
                        console.error('Error stack:', (eip712Error as Error).stack);

                        cleanup();
                        reject(new Error(`EIP-712 auth_verify failed: ${(eip712Error as Error).message}`));
                        return;
                    }
                }
                // Check for success response
                else if (response.res && (response.res[1] === 'auth_verify' || response.res[1] === 'auth_success')) {
                    console.log('Authentication successful');

                    // If response contains a JWT token, store it
                    if (response.res[2]?.[0]?.['jwt_token']) {
                        console.log('JWT token received:', response.res[2][0]['jwt_token']);
                        window.localStorage.setItem('jwt_token', response.res[2][0]['jwt_token']);
                    }

                    cleanup();
                    resolve();
                }
                // Check for error response
                else if (response.err || (response.res && response.res[1] === 'error')) {
                    const errorMsg =
                        response.err?.[1] || response.error || response.res?.[2]?.[0]?.error || 'Authentication failed';

                    console.error('Authentication failed:', errorMsg);

                    // Check if this is a JWT authentication failure and fallback to signer auth
                    const errorString = String(errorMsg).toLowerCase();
                    if (errorString.includes('jwt') || errorString.includes('token') || errorString.includes('invalid') || errorString.includes('expired')) {
                        console.warn('JWT authentication failed on server, attempting fallback to signer authentication');
                        window.localStorage.removeItem('jwt_token');

                        try {
                            // Restart authentication with signer
                            const fallbackAuthRequest = await createAuthRequestMessage(authMessage);

                            console.log('Sending fallback auth_request with signer:', fallbackAuthRequest);
                            ws.send(fallbackAuthRequest);
                            resetTimeout(); // Reset timeout for the fallback attempt
                            return; // Continue listening for the fallback response
                        } catch (fallbackError) {
                            console.error('Fallback to signer authentication failed:', fallbackError);
                            cleanup();
                            reject(new Error(`Both JWT and signer authentication failed: ${fallbackError}`));
                            return;
                        }
                    }

                    window.localStorage.removeItem('jwt_token');
                    cleanup();
                    reject(new Error(String(errorMsg)));
                } else {
                    console.log('Received non-auth message during auth, continuing to listen:', response);
                    // Keep listening if it wasn't a final success/error
                }
            } catch (error) {
                console.error('Error handling auth response:', error);
                console.error('Error stack:', (error as Error).stack);
                cleanup();
                reject(new Error(`Authentication error: ${error instanceof Error ? error.message : String(error)}`));
            }
        };

        ws.addEventListener('message', handleAuthResponse);
        resetTimeout(); // Start the initial timeout
    });
}
