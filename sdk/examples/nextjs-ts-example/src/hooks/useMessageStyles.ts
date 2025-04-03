import { useMemo } from 'react';

// Define message type constant
export type MessageType = 'system' | 'error' | 'success' | 'info' | 'sent' | 'received' | 'warning' | 
                          'user-ping' | 'user-pong' | 'guest-ping' | 'guest-pong' | 'ping' | 'pong';

// Define styles once, outside the component for better performance
const MESSAGE_STYLES: Record<MessageType, string> = {
    system: 'bg-gray-100 text-gray-600 border border-gray-200',
    error: 'bg-red-50 text-red-600 border-l-2 border-red-500',
    success: 'bg-green-50 text-green-700 border-l-2 border-green-500',
    info: 'bg-gray-50 text-gray-700 border border-gray-200',
    sent: 'bg-[#3531ff]/10 text-[#3531ff] border-l-2 border-[#3531ff]',
    received: 'bg-indigo-50 text-indigo-700 border-l-2 border-indigo-500',
    warning: 'bg-yellow-50 text-yellow-700 border-l-2 border-yellow-500',
    // User ping/pong messages (blue colors)
    'user-ping': 'bg-blue-100 text-blue-700 border-l-2 border-blue-500 font-bold',
    'user-pong': 'bg-sky-100 text-sky-700 border-l-2 border-sky-500 font-bold',
    // Guest ping/pong messages (purple/pink colors)
    'guest-ping': 'bg-purple-100 text-purple-700 border-l-2 border-purple-500 font-bold',
    'guest-pong': 'bg-pink-100 text-pink-700 border-l-2 border-pink-500 font-bold',
    // Legacy styles for backward compatibility
    ping: 'bg-purple-100 text-purple-700 border-l-2 border-purple-500 font-bold',
    pong: 'bg-pink-100 text-pink-700 border-l-2 border-pink-500 font-bold',
};

export function useMessageStyles() {
    // Memoize to prevent unnecessary re-rendering
    return useMemo(() => MESSAGE_STYLES, []);
}
