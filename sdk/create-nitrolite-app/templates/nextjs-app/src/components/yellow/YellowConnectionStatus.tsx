'use client';

import { useEffect, useRef } from 'react';
import { useYellowWebSocketContext } from './YellowWebSocketProvider';
import { useLoginState } from '../../hooks/useLoginState';

interface YellowConnectionStatusProps {
    className?: string;
    showDetails?: boolean;
    autoReconnect?: boolean;
    reconnectDelay?: number;
}

export const YellowConnectionStatus = ({
    className = '',
    showDetails = false,
    autoReconnect = true,
    reconnectDelay = 5000,
}: YellowConnectionStatusProps) => {
    const { isConnected, isConnecting, error, sessionAddress, connect } = useYellowWebSocketContext();
    const { isLoggedIn, walletAddress, login } = useLoginState();
    const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
    const lastErrorRef = useRef<string | null>(null);

    // Auto-reconnect logic
    useEffect(() => {
        if (!autoReconnect || !isLoggedIn || !walletAddress) {
            return;
        }

        // Clear any existing timeout
        if (reconnectTimeoutRef.current) {
            clearTimeout(reconnectTimeoutRef.current);
            reconnectTimeoutRef.current = null;
        }

        // Don't reconnect if user rejected signing
        const isUserRejectedError =
            error &&
            (error.toLowerCase().includes('user rejected') ||
                error.toLowerCase().includes('user denied') ||
                error.toLowerCase().includes('user cancelled'));

        if (!isConnected && !isConnecting && error && !isUserRejectedError) {
            // Only reconnect if this is a new error or after delay
            if (lastErrorRef.current !== error) {
                lastErrorRef.current = error;
                console.log(`🔄 Scheduling reconnect in ${reconnectDelay}ms due to error:`, error);

                reconnectTimeoutRef.current = setTimeout(() => {
                    console.log('🔄 Attempting auto-reconnect...');
                    connect(walletAddress).catch((connectError) => {
                        console.error('❌ Auto-reconnect failed:', connectError);
                    });
                }, reconnectDelay);
            }
        }

        // Clean up on unmount
        return () => {
            if (reconnectTimeoutRef.current) {
                clearTimeout(reconnectTimeoutRef.current);
                reconnectTimeoutRef.current = null;
            }
        };
    }, [isConnected, isConnecting, error, isLoggedIn, walletAddress, autoReconnect, reconnectDelay, connect]);

    const handleManualReconnect = () => {
        if (walletAddress && !isConnecting) {
            console.log('🔄 Manual reconnect triggered');
            
            let retryCount = 0;
            const maxRetries = 3;
            
            const attemptReconnect = (delay: number) => {
                setTimeout(() => {
                    connect(walletAddress).catch((connectError) => {
                        console.error(`❌ Manual reconnect failed (attempt ${retryCount + 1}/${maxRetries}):`, connectError);
                        
                        // If it's a wallet not ready error, retry with progressive delay
                        if (connectError.message.includes('Privy wallet not ready') && retryCount < maxRetries - 1) {
                            retryCount++;
                            const nextDelay = 2000 * retryCount; // 2s, 4s, 6s
                            console.log(`🔄 Wallet not ready, retrying manual reconnect in ${nextDelay / 1000} seconds... (${retryCount}/${maxRetries})`);
                            attemptReconnect(nextDelay);
                        } else {
                            console.error('❌ Manual reconnect max retries reached or different error occurred');
                        }
                    });
                }, delay);
            };
            
            attemptReconnect(1000); // Initial delay of 1 second for manual reconnect
        }
    };

    const handleLogin = async () => {
        try {
            console.log('🔐 Login triggered');
            await login();
        } catch (loginError) {
            console.error('❌ Login failed:', loginError);
        }
    };

    const getStatusColor = () => {
        if (!isLoggedIn) return 'text-orange-400';
        if (isConnected) return 'text-green-400';
        if (isConnecting) return 'text-yellow-400';
        if (error) return 'text-red-400';
        return 'text-gray-400';
    };

    const getStatusText = () => {
        if (!isLoggedIn) return 'Login required';
        if (isConnected) return 'Connected to Yellow';
        if (isConnecting) return 'Connecting to Yellow...';
        if (error) return 'Failed to connect to Yellow';
        return 'Disconnected from Yellow';
    };

    const getStatusIcon = () => {
        if (!isLoggedIn) return '🔐';
        if (isConnected) return '🟢';
        if (isConnecting) return '🟡';
        if (error) return '🔴';
        return '⚫';
    };

    const isUserRejectedError =
        error &&
        (error.toLowerCase().includes('user rejected') ||
            error.toLowerCase().includes('user denied') ||
            error.toLowerCase().includes('user cancelled'));

    const canReconnect = !isConnected && !isConnecting && walletAddress && !isUserRejectedError;

    return (
        <div className={`flex items-center space-x-2 ${className}`}>
            <span className="text-sm">{getStatusIcon()}</span>
            <span className={`text-xs font-medium ${getStatusColor()}`}>{getStatusText()}</span>

            {/* Login button */}
            {!isLoggedIn && (
                <button
                    onClick={handleLogin}
                    className="text-xs px-2 py-1 bg-orange-500 hover:bg-orange-600 text-white rounded transition-colors"
                    title="Click to login with Privy">
                    Login
                </button>
            )}

            {/* Manual reconnect button */}
            {canReconnect && (
                <button
                    onClick={handleManualReconnect}
                    className="text-xs px-2 py-1 bg-blue-500 hover:bg-blue-600 text-white rounded transition-colors"
                    title="Click to reconnect">
                    Reconnect
                </button>
            )}

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
                    {autoReconnect && !isConnected && !isConnecting && error && !isUserRejectedError && (
                        <div className="mt-1 text-yellow-400">
                            Auto-reconnecting in {Math.ceil(reconnectDelay / 1000)}s...
                        </div>
                    )}
                </div>
            )}
        </div>
    );
};
