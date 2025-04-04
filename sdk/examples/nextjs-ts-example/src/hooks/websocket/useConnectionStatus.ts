import { useMemo } from 'react';
import { WSStatus } from '@/types';

// Define status mappings once, outside the component
const STATUS_COLORS = {
    connected: 'bg-green-500',
    connecting: 'bg-yellow-500 animate-pulse',
    reconnecting: 'bg-yellow-500 animate-pulse',
    authenticating: 'bg-yellow-500 animate-pulse',
    disconnected: 'bg-red-500',
    reconnect_failed: 'bg-red-500',
    auth_failed: 'bg-red-500',
    waiting: 'bg-gray-500',
};

const STATUS_TEXTS = {
    connected: 'Connected',
    connecting: 'Connecting...',
    disconnected: 'Disconnected',
    reconnecting: 'Reconnecting...',
    reconnect_failed: 'Reconnection failed',
    auth_failed: 'Authentication failed',
    authenticating: 'Authenticating...',
    waiting: 'Waiting for connection...',
};

export function useConnectionStatus(status: WSStatus) {
    const statusColor = useMemo(() => STATUS_COLORS[status] || 'bg-gray-500', [status]);

    const statusText = useMemo(
        () => STATUS_TEXTS[status] || status.charAt(0).toUpperCase() + status.slice(1),
        [status],
    );

    return {
        statusColor,
        statusText,
    };
}