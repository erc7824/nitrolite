import { NitroliteRPC } from '@erc7824/nitrolite';
import { WalletSigner } from '../crypto';
import MessageService from '../services/MessageService';

/**
 * Authenticates with the WebSocket server and performs ping verification
 *
 * @param ws - The WebSocket connection
 * @param signer - The signer to use for authentication
 * @param timeout - Timeout in milliseconds
 * @returns A Promise that resolves when authenticated and ping verification is complete
 */
export async function authenticate(ws: WebSocket, signer: WalletSigner, timeout: number): Promise<void> {
    if (!ws) throw new Error('WebSocket not connected');

    // Step 1: Authenticate with the server
    await performAuthentication(ws, signer, timeout);
    
    // Step 2: Skip ping verification, it will be handled by the UI
}

/**
 * Performs the initial authentication with the WebSocket server
 */
export async function performAuthentication(ws: WebSocket, signer: WalletSigner, timeout: number): Promise<void> {
    const authRequest = NitroliteRPC.createRequest('auth', [signer.address]);
    const signedAuthRequest = await NitroliteRPC.signMessage(authRequest, signer.sign);

    return new Promise((resolve, reject) => {
        if (!ws) return reject(new Error('WebSocket not connected'));

        const authTimeout = setTimeout(() => {
            ws.removeEventListener('message', handleAuthResponse);
            reject(new Error('Authentication timeout'));
        }, timeout);

        const handleAuthResponse = (event: MessageEvent) => {
            let response;

            try {
                response = JSON.parse(event.data);
            } catch (error) {
                MessageService.error(`Error parsing auth response: ${error}`);
                return; // Continue waiting for valid responses
            }

            try {
                if ((response.res && response.res[1] === 'auth') || response.type === 'auth_success') {
                    clearTimeout(authTimeout);
                    ws.removeEventListener('message', handleAuthResponse);
                    MessageService.system('Authentication successful');
                    resolve();
                } else if ((response.err && response.err[1]) || response.type === 'auth_error') {
                    clearTimeout(authTimeout);
                    ws.removeEventListener('message', handleAuthResponse);
                    const errorMsg = response.err ? response.err[2] : response.error || 'Authentication failed';

                    MessageService.error(`Authentication error: ${errorMsg}`);
                    reject(new Error(errorMsg));
                }
            } catch (error) {
                MessageService.error(`Error handling auth response: ${error}`);
                clearTimeout(authTimeout);
                ws.removeEventListener('message', handleAuthResponse);
                reject(new Error(`Authentication error: ${error instanceof Error ? error.message : String(error)}`));
            }
        };

        ws.addEventListener('message', handleAuthResponse);
        ws.send(JSON.stringify(signedAuthRequest));
    });
}

/**
 * Performs a simplified single ping-pong verification to ensure connection stability
 */
export async function performSinglePingVerification(
    ws: WebSocket,
    signer: WalletSigner,
    timeout: number,
    options?: {
        channel?: string;
    },
): Promise<void> {
    const CHANNEL = options?.channel ?? 'public'; // Default to public channel

    MessageService.system(`Starting ping-pong verification on channel ${CHANNEL}`);

    // Create a promise to track the ping-pong completion
    return new Promise<void>((resolve, reject) => {
        let pingReceived = false;
        let pongReceived = false;

        // Function to send a message on the channel
        const sendChannelMessage = async (message: string): Promise<void> => {
            try {
                const shortenedKey = signer.publicKey.substring(0, 10) + '...';
                const request = NitroliteRPC.createRequest('publish', [CHANNEL, message, shortenedKey]);
                const signedRequest = await NitroliteRPC.signMessage(request, signer.sign);

                ws.send(JSON.stringify(signedRequest));
            } catch (error) {
                MessageService.error(
                    `Error sending message: ${error instanceof Error ? error.message : String(error)}`,
                );
            }
        };

        // Function to handle messages on the channel
        const handleMessage = (event: MessageEvent) => {
            let message;

            try {
                message = JSON.parse(event.data);
            } catch (error) {
                // Ignore parsing errors
                return;
            }

            try {
                // Check for channel messages - matching for ping/pong
                if (
                    message.data &&
                    message.type === 'channel_message' &&
                    message.data.channel === CHANNEL &&
                    message.data.content &&
                    message.data.sender
                ) {
                    const content = message.data.content.toString().trim().toLowerCase();
                    const sender = message.data.sender;
                    const ourKey = signer.publicKey.substring(0, 10) + '...';

                    // Only process messages from other users
                    if (sender !== ourKey) {
                        if (content === 'ping') {
                            pingReceived = true;
                            // Received a ping, respond with pong
                            MessageService.add({ 
                                text: `>guest: PING`, 
                                type: 'guest-ping', 
                                sender: 'guest' 
                            });
                            
                            // Send pong response
                            sendChannelMessage('pong');
                            MessageService.add({ 
                                text: `>user: PONG`, 
                                type: 'user-pong', 
                                sender: 'user' 
                            });
                        } 
                        else if (content === 'pong') {
                            pongReceived = true;
                            // Received a pong
                            MessageService.add({ 
                                text: `>guest: PONG`, 
                                type: 'guest-pong', 
                                sender: 'guest' 
                            });
                            
                            // Ping-pong completed successfully
                            if (pingReceived && pongReceived) {
                                cleanup();
                                resolve();
                            }
                        }
                    }
                }
            } catch (error) {
                // Ignore errors in message handling
            }
        };

        // Send initial ping
        const sendInitialPing = async (): Promise<void> => {
            try {
                await sendChannelMessage('ping');
                MessageService.add({ 
                    text: `>user: PING`, 
                    type: 'user-ping',
                    sender: 'user'
                });
            } catch (error) {
                MessageService.error(`Error sending ping: ${error instanceof Error ? error.message : String(error)}`);
                cleanup();
                reject(error);
            }
        };

        // Add message listener
        ws.addEventListener('message', handleMessage);

        // Start with a ping
        sendInitialPing();

        // Set timeout for the ping-pong process
        const pingPongTimeout = setTimeout(() => {
            MessageService.error('Ping-pong verification timed out');
            cleanup();
            reject(new Error('Ping-pong verification timed out'));
        }, timeout);

        // Clean up resources
        const cleanup = () => {
            clearTimeout(pingPongTimeout);
            ws.removeEventListener('message', handleMessage);
        };
    });
}

/**
 * Helper function to subscribe to a channel
 */
async function subscribeToChannel(
    ws: WebSocket,
    signer: WalletSigner,
    channel: string,
    timeout: number,
): Promise<void> {
    return new Promise<void>((resolve, reject) => {
        MessageService.system(`Subscribing to channel ${channel} for ping-pong verification`);

        const subscribeRequest = NitroliteRPC.createRequest('subscribe', [channel]);

        const handleSubscribeResponse = async (event: MessageEvent) => {
            let response;

            try {
                response = JSON.parse(event.data);
            } catch (error) {
                // Ignore parsing errors
                return;
            }

            try {
                if (
                    (response.res && response.res[1] === 'subscribe') ||
                    (response.type === 'subscribe_success' && response.data && response.data.channel === channel)
                ) {
                    ws.removeEventListener('message', handleSubscribeResponse);
                    clearTimeout(subscribeTimeout);
                    MessageService.system(`Successfully subscribed to channel ${channel}`);
                    resolve();
                }
            } catch (error) {
                // Ignore errors in handling responses
            }
        };

        // Set timeout for subscription
        const subscribeTimeout = setTimeout(() => {
            ws.removeEventListener('message', handleSubscribeResponse);
            reject(new Error(`Subscription to channel ${channel} timed out`));
        }, timeout);

        // Add listener for subscription response
        ws.addEventListener('message', handleSubscribeResponse);

        // Send subscription request
        NitroliteRPC.signMessage(subscribeRequest, signer.sign)
            .then((signedRequest) => {
                ws.send(JSON.stringify(signedRequest));
            })
            .catch((error) => {
                ws.removeEventListener('message', handleSubscribeResponse);
                clearTimeout(subscribeTimeout);
                reject(error);
            });
    });
}
