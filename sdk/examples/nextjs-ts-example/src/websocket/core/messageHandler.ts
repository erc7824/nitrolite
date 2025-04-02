import { Channel } from "@/types";

/**
 * Handles incoming WebSocket messages
 * 
 * @param event - The message event
 * @param pendingRequests - Map of pending requests waiting for responses
 * @param setChannel - Function to set the current channel
 * @param onMessageCallback - Callback for all messages
 * @param onErrorCallback - Callback for errors
 */
export function handleMessage(
  event: MessageEvent,
  pendingRequests: Map<number, { resolve: Function; reject: Function }>,
  setChannel: (channel: Channel) => void,
  onMessageCallback?: (message: any) => void,
  onErrorCallback?: (error: Error) => void
): void {
  let response;

  try {
    response = JSON.parse(event.data);
  } catch (error) {
    console.error("Error parsing message:", error);
    console.log("Raw message:", event.data);
    // Notify about message parsing error but don't break the connection
    onErrorCallback?.(new Error("Failed to parse server message"));
    return;
  }

  try {
    // Notify callback about received message
    onMessageCallback?.(response);

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
      if (pendingRequests.has(requestId)) {
        pendingRequests.get(requestId)!.reject(
          new Error(`Error ${response.err[1]}: ${response.err[2]}`)
        );
        pendingRequests.delete(requestId);
      }
    }
    // Handle legacy/custom responses
    else if (response.type) {
      if (response.type === "auth_success") {
        // Authentication handled separately
      } else if (response.type === "subscribe_success" && response.data?.channel) {
        setChannel(response.data.channel as Channel);
      }

      // For all other responses with a requestId, resolve any pending requests
      const requestId = response.requestId;
      if (requestId && pendingRequests.has(requestId)) {
        pendingRequests.get(requestId)!.resolve(response.data || response);
        pendingRequests.delete(requestId);
      }
    }
  } catch (error) {
    console.error("Error handling message:", error);
    onErrorCallback?.(
      new Error(`Error processing message: ${error instanceof Error ? error.message : String(error)}`)
    );
  }
}