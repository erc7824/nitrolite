import { useMemo } from 'react';

// Define message type constant
export type MessageType = 'system' | 'error' | 'success' | 'info' | 'sent' | 'received' | 'warning';

// Define styles once, outside the component for better performance
const MESSAGE_STYLES: Record<MessageType, string> = {
  "system": "bg-gray-700 text-blue-300",
  "error": "bg-red-900 bg-opacity-50 text-red-300 border-l-2 border-red-500",
  "success": "bg-green-900 bg-opacity-50 text-green-300 border-l-2 border-green-500",
  "info": "bg-gray-700 text-gray-300",
  "sent": "bg-secondary-900 bg-opacity-50 text-secondary-300 border-l-2 border-secondary-500",
  "received": "bg-primary-900 bg-opacity-50 text-primary-300 border-l-2 border-primary-500",
  "warning": "bg-yellow-900 bg-opacity-50 text-yellow-300 border-l-2 border-yellow-500"
};

export function useMessageStyles() {
  // Memoize to prevent unnecessary re-rendering
  return useMemo(() => MESSAGE_STYLES, []);
}