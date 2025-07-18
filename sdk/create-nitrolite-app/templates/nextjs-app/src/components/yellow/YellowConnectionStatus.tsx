'use client';

import { useYellowWebSocketContext } from './YellowWebSocketProvider';

interface YellowConnectionStatusProps {
    className?: string;
    showDetails?: boolean;
}

export const YellowConnectionStatus = ({ className = '', showDetails = false }: YellowConnectionStatusProps) => {
    const { isConnected, isConnecting, error, sessionAddress } = useYellowWebSocketContext();

    const getStatusColor = () => {
        if (isConnected) return 'text-green-400';
        if (isConnecting) return 'text-yellow-400';
        if (error) return 'text-red-400';
        return 'text-gray-400';
    };

    const getStatusText = () => {
        if (isConnected) return 'Connected to Yellow';
        if (isConnecting) return 'Connecting to Yellow...';
        if (error) return 'Failed to connect to Yellow';
        return 'Disconnected from Yellow';
    };

    const getStatusIcon = () => {
        if (isConnected) return '🟢';
        if (isConnecting) return '🟡';
        if (error) return '🔴';
        return '⚫';
    };

    return (
        <div className={`flex items-center space-x-2 ${className}`}>
            <span className="text-sm">{getStatusIcon()}</span>
            <span className={`text-xs font-medium ${getStatusColor()}`}>{getStatusText()}</span>

            {showDetails && (
                <div className="text-xs text-gray-500">
                    {sessionAddress && (
                        <div className="mt-1">
                            <span className="text-gray-400">Session: </span>
                            <span className="font-mono">
                                {sessionAddress.slice(0, 6)}...{sessionAddress.slice(-4)}
                            </span>
                        </div>
                    )}
                    {error && <div className="mt-1 text-red-400">{error}</div>}
                </div>
            )}
        </div>
    );
};
