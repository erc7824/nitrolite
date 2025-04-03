import { Channel } from '@/types';
import MessageService from '@/websocket/services/MessageService';

/**
 * Handles incoming WebSocket messages
 * @param event - The message event
 * @param pendingRequests - Map of pending requests waiting for responses
 * @param setChannel - Function to set the current channel
 * @param onMessageCallback - Optional callback for all messages
 * @param onErrorCallback - Optional callback for errors
 */
export function handleMessage(
    event: MessageEvent,
    pendingRequests: Map<number, { resolve: (value: unknown) => void; reject: (reason: Error) => void }>,
    setChannel: (channel: Channel) => void,
    onMessageCallback?: (message: unknown) => void,
    onErrorCallback?: (error: Error) => void,
): void {
    let response;

    // Parse incoming message
    try {
        response = JSON.parse(event.data);
    } catch (error) {
        const errorMessage = `Failed to parse server message ${error}`;

        // Log the raw data for debugging without using console
        MessageService.error(`${errorMessage}: ${event.data}`);
        onErrorCallback?.(new Error(errorMessage));
        return;
    }

    try {
        // Notify callback about received message
        onMessageCallback?.(response);

        // Process message with MessageService
        MessageService.handleWebSocketMessage(response);

        // Type guard to check object shape
        const hasProperty = <T extends object, K extends string>(obj: T, prop: K): obj is T & Record<K, unknown> => {
            return prop in obj;
        };

        if (typeof response === 'object' && response !== null) {
            // Handle standard NitroRPC responses
            if (hasProperty(response, 'res') && Array.isArray(response.res) && response.res.length >= 3) {
                const requestId = typeof response.res[0] === 'number' ? response.res[0] : -1;

                if (pendingRequests.has(requestId)) {
                    pendingRequests.get(requestId)!.resolve(response.res[2]);
                    pendingRequests.delete(requestId);
                }
            }
            // Handle error responses
            else if (hasProperty(response, 'err') && Array.isArray(response.err) && response.err.length >= 3) {
                const requestId = typeof response.err[0] === 'number' ? response.err[0] : -1;
                const errorCode = String(response.err[1]);
                const errorDesc = String(response.err[2]);
                const errorMessage = `Error ${errorCode}: ${errorDesc}`;

                MessageService.error(errorMessage);

                if (pendingRequests.has(requestId)) {
                    pendingRequests.get(requestId)!.reject(new Error(errorMessage));
                    pendingRequests.delete(requestId);
                }
            }
            // Handle legacy/custom responses
            else if (hasProperty(response, 'type') && typeof response.type === 'string') {
                if (
                    response.type === 'subscribe_success' &&
                    hasProperty(response, 'data') &&
                    typeof response.data === 'object' &&
                    response.data &&
                    hasProperty(response.data, 'channel') &&
                    (response.data.channel === 'public' ||
                        response.data.channel === 'game' ||
                        response.data.channel === 'trade' ||
                        response.data.channel === 'private')
                ) {
                    setChannel(response.data.channel);
                }

                // Resolve any pending requests with a requestId
                if (hasProperty(response, 'requestId') && typeof response.requestId === 'number') {
                    const requestId = response.requestId;

                    if (pendingRequests.has(requestId)) {
                        const result = hasProperty(response, 'data') ? response.data : response;

                        pendingRequests.get(requestId)!.resolve(result);
                        pendingRequests.delete(requestId);
                    }
                }
            }
        }
    } catch (error) {
        const errorMessage = `Error processing message: ${error instanceof Error ? error.message : String(error)}`;

        MessageService.error(errorMessage);
        onErrorCallback?.(new Error(errorMessage));
    }
}
