// Re-export websocket-related hooks
export * from './useWebSocket';
export * from './useConnectionStatus';

// Default export for convenience
import { useWebSocket } from './useWebSocket';
export default useWebSocket;