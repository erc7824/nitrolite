import { Channel } from "@/types";
import MessageService from "@/websocket/services/MessageService";

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
  pendingRequests: Map<number, { resolve: Function; reject: Function }>,
  setChannel: (channel: Channel) => void,
  onMessageCallback?: (message: any) => void,
  onErrorCallback?: (error: Error) => void
): void {
  let response;

  // Parse incoming message
  try {
    response = JSON.parse(event.data);
  } catch (error) {
    const errorMessage = "Failed to parse server message";
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

    // Handle standard NitroRPC responses
    if (response.res) {
      const requestId = response.res[0];
      if (pendingRequests.has(requestId)) {
        pendingRequests.get(requestId)!.resolve(response.res[2]);
        pendingRequests.delete(requestId);
      }
    }
    // Handle error responses
    else if (response.err) {
      const requestId = response.err[0];
      const errorMessage = `Error ${response.err[1]}: ${response.err[2]}`;
      
      MessageService.error(errorMessage);
      
      if (pendingRequests.has(requestId)) {
        pendingRequests.get(requestId)!.reject(new Error(errorMessage));
        pendingRequests.delete(requestId);
      }
    }
    // Handle legacy/custom responses
    else if (response.type) {
      if (response.type === "subscribe_success" && response.data?.channel) {
        setChannel(response.data.channel as Channel);
      }

      // Resolve any pending requests with a requestId
      const requestId = response.requestId;
      if (requestId && pendingRequests.has(requestId)) {
        pendingRequests.get(requestId)!.resolve(response.data || response);
        pendingRequests.delete(requestId);
      }
    }
  } catch (error) {
    const errorMessage = `Error processing message: ${error instanceof Error ? error.message : String(error)}`;
    MessageService.error(errorMessage);
    onErrorCallback?.(new Error(errorMessage));
  }
}