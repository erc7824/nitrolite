import { useMemo } from 'react';

// Define message type constant
export type MessageType = 'system' | 'error' | 'success' | 'info' | 'sent' | 'received' | 'warning';

// Define styles once, outside the component for better performance
const MESSAGE_STYLES: Record<MessageType, string> = {
    system: 'bg-gray-100 text-gray-600 border border-gray-200',
    error: 'bg-red-50 text-red-600 border-l-2 border-red-500',
    success: 'bg-green-50 text-green-700 border-l-2 border-green-500',
    info: 'bg-gray-50 text-gray-700 border border-gray-200',
    sent: 'bg-[#3531ff]/10 text-[#3531ff] border-l-2 border-[#3531ff]',
    received: 'bg-indigo-50 text-indigo-700 border-l-2 border-indigo-500',
    warning: 'bg-yellow-50 text-yellow-700 border-l-2 border-yellow-500',
};

export function useMessageStyles() {
    // Memoize to prevent unnecessary re-rendering
    return useMemo(() => MESSAGE_STYLES, []);
}