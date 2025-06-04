import { createAuthRequestMessage, createAuthVerifyMessage, createAuthVerifyMessageWithJWT } from '@erc7824/nitrolite';
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

const AUTH_TYPES = {
    Policy: [
        { name: 'challenge', type: 'string' },
        { name: 'scope', type: 'string' },
        { name: 'wallet', type: 'address' },
        { name: 'application', type: 'address' },
        { name: 'participant', type: 'address' },
        { name: 'expire', type: 'uint256' },
        { name: 'allowances', type: 'Allowance[]' },
    ],
    Allowance: [
        { name: 'asset', type: 'string' },
        { name: 'amount', type: 'uint256' },
    ],
};

const expire = String(Math.floor(Date.now() / 1000) + 24 * 60 * 60);

/**
 * Creates EIP-712 signing function for challenge verification with proper challenge extraction
 */
function createEIP712SigningFunction(walletClient: any, stateSigner: WalletSigner) {
    if (!walletClient) {
        throw new Error('No wallet client available for EIP-712 signing');
    }

    return async (data: any): Promise<Hex> => {
        console.log('Signing auth_verify challenge with EIP-712:', data);

        let challengeUUID = '';
        const address = walletClient.account?.address ? getAddress(walletClient.account.address) : null;
        
        if (!address) {
            throw new Error('No wallet address available for signing');
        }

        // For Snake game, we don't have complex channel state, so we'll use 0 for amount
        const totalAmount = 0;

        // The data coming in is the array from createAuthVerifyMessage
        // Format: [timestamp, "auth_verify", [{"address": "0x...", "challenge": "uuid"}], timestamp]
        if (Array.isArray(data)) {
            console.log('Data is array, extracting challenge from position [2][0].challenge');

            // Direct array access - data[2] should be the array with the challenge object
            if (data.length >= 3 && Array.isArray(data[2]) && data[2].length > 0) {
                const challengeObject = data[2][0];

                if (challengeObject && challengeObject.challenge) {
                    challengeUUID = challengeObject.challenge;
                    console.log('Extracted challenge UUID from array:', challengeUUID);
                }
            }
        } else if (typeof data === 'string') {
            try {
                const parsed = JSON.parse(data);

                console.log('Parsed challenge data:', parsed);

                // Handle different message structures
                if (parsed.res && Array.isArray(parsed.res)) {
                    // auth_challenge response: {"res": [id, "auth_challenge", {"challenge": "uuid"}, timestamp]}
                    if (parsed.res[1] === 'auth_challenge' && parsed.res[2]) {
                        challengeUUID = parsed.res[2].challenge_message || parsed.res[2].challenge;
                        console.log('Extracted challenge UUID from auth_challenge:', challengeUUID);
                    }
                    // auth_verify message: [timestamp, "auth_verify", [{"address": "0x...", "challenge": "uuid"}], timestamp]
                    else if (parsed.res[1] === 'auth_verify' && Array.isArray(parsed.res[2]) && parsed.res[2][0]) {
                        challengeUUID = parsed.res[2][0].challenge;
                        console.log('Extracted challenge UUID from auth_verify:', challengeUUID);
                    }
                }
                // Direct array format
                else if (Array.isArray(parsed) && parsed.length >= 3 && Array.isArray(parsed[2])) {
                    challengeUUID = parsed[2][0]?.challenge;
                    console.log('Extracted challenge UUID from direct array:', challengeUUID);
                }
            } catch (e) {
                console.error('Could not parse challenge data:', e);
                console.log('Using raw string as challenge');
                challengeUUID = data;
            }
        } else if (data && typeof data === 'object') {
            // If data is already an object, try to extract challenge
            challengeUUID = data.challenge || data.challenge_message;
            console.log('Extracted challenge from object:', challengeUUID);
        }

        if (!challengeUUID || challengeUUID.includes('[') || challengeUUID.includes('{')) {
            console.error('Challenge extraction failed or contains invalid characters:', challengeUUID);
            throw new Error('Could not extract valid challenge UUID for EIP-712 signing');
        }

        console.log('Final challenge UUID for EIP-712:', challengeUUID);
        console.log('Signing for address (original):', address);
        console.log('Signing for address (type):', typeof address);
        console.log('Auth domain:', getAuthDomain());

        // Create EIP-712 message with ONLY the challenge UUID
        const message = {
            challenge: challengeUUID,
            scope: 'snake-game',
            wallet: address as `0x${string}`,
            application: address as `0x${string}`,
            participant: stateSigner.address as `0x${string}`,
            expire: expire, // 24 hours in seconds
            allowances: [
                {
                    asset: 'usdc',
                    amount: totalAmount.toString(),
                },
            ],
        };

        try {
            // Sign with EIP-712
            const signature = await walletClient.signTypedData({
                account: walletClient.account!,
                domain: getAuthDomain(),
                types: AUTH_TYPES,
                primaryType: 'Policy',
                message: message,
            });

            console.log('EIP-712 signature generated for challenge:', signature);
            return signature;
        } catch (eip712Error) {
            console.error('EIP-712 signing failed:', eip712Error);
            console.log('Attempting fallback to regular message signing...');

            try {
                // Fallback to regular message signing if EIP-712 fails
                const fallbackMessage = `Authentication challenge for ${address}: ${challengeUUID}`;

                console.log('Fallback message:', fallbackMessage);

                const fallbackSignature = await walletClient.signMessage({
                    message: fallbackMessage,
                    account: walletClient.account!,
                });

                console.log('Fallback signature generated:', fallbackSignature);
                return fallbackSignature as `0x${string}`;
            } catch (fallbackError) {
                console.error('Fallback signing also failed:', fallbackError);
                throw new Error(`Both EIP-712 and fallback signing failed: ${(eip712Error as Error).message}`);
            }
        }
    };
}

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
                authRequest = await createAuthRequestMessage({
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
                });
                usingJWT = false;
            }
        } else {
            console.log('No JWT token found, proceeding with challenge-response authentication');
            authRequest = await createAuthRequestMessage({
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
            });
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
            }, timeout);
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
                        console.log('Creating EIP-712 signing function...');
                        const eip712SigningFunction = createEIP712SigningFunction(walletClient, signer);

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
                            const fallbackAuthRequest = await createAuthRequestMessage({
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
                            });
                            
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
