import { useEffect, useRef } from 'react';
import { useMessageStyles } from '@/hooks/useMessageStyles';
import { useSnapshot } from 'valtio';
import WalletStore from '@/store/WalletStore';
import { useMessageService } from '@/hooks/useMessageService';

interface FullscreenMessagesProps {
    onClose: () => void;
}

export function FullscreenMessages({ onClose }: FullscreenMessagesProps) {
    const messageStyles = useMessageStyles();
    const messagesEndRef = useRef<HTMLDivElement>(null);
    const containerRef = useRef<HTMLDivElement>(null);
    const walletSnap = useSnapshot(WalletStore.state);
    const shouldScrollRef = useRef(true);

    // Use our message service hook
    const { messages, status } = useMessageService();

    // Auto-scroll to bottom when messages change
    useEffect(() => {
        if (shouldScrollRef.current && messagesEndRef.current && containerRef.current) {
            containerRef.current.scrollTo({
                top: containerRef.current.scrollHeight,
                behavior: 'smooth',
            });
        }
    }, [messages]);

    // Handle scroll events to detect if user has scrolled up
    useEffect(() => {
        const container = containerRef.current;

        if (!container) return;

        const handleScroll = () => {
            const isAtBottom = container.scrollHeight - container.scrollTop - container.clientHeight < 50;

            shouldScrollRef.current = isAtBottom;
        };

        container.addEventListener('scroll', handleScroll);
        return () => container.removeEventListener('scroll', handleScroll);
    }, []);

    // Format timestamp
    const formatTime = (timestamp?: number) => {
        if (!timestamp) return '';
        return new Date(timestamp).toLocaleTimeString();
    };

    return (
        <div className="fixed inset-0 bg-white z-50 flex flex-col">
            <div className="flex justify-between items-center py-2 px-4 border-b border-gray-200">
                <h1 className="text-lg font-bold text-[#3531ff]">Nitrolite Messages</h1>
                <button
                    onClick={onClose}
                    className="bg-[#3531ff] hover:bg-[#2b28cc] text-white px-3 py-1 rounded flex items-center transition-colors"
                >
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        className="h-4 w-4 mr-1"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                    >
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                    Exit Full View
                </button>
            </div>

            <div ref={containerRef} className="flex-grow overflow-y-auto p-3" id="fullscreen-message-container">
                {messages.length === 0 ? (
                    <div className="text-gray-500 text-center py-10">No messages yet</div>
                ) : (
                    <div className="w-full mx-auto">
                        {messages.map((message, index) => (
                            <div
                                key={index}
                                className={`p-1 rounded mb-0 ${messageStyles[message.type] || messageStyles.info}`}
                            >
                                {message.type === 'sent' && (
                                    <svg
                                        className="inline-block w-4 h-4 mr-1 -mt-1"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        xmlns="http://www.w3.org/2000/svg"
                                    >
                                        <path
                                            d="M5 13l4 4L19 7"
                                            stroke="currentColor"
                                            strokeWidth="2"
                                            strokeLinecap="round"
                                            strokeLinejoin="round"
                                        />
                                    </svg>
                                )}
                                {message.type === 'received' && (
                                    <svg
                                        className="inline-block w-4 h-4 mr-1 -mt-1"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        xmlns="http://www.w3.org/2000/svg"
                                    >
                                        <path
                                            d="M20 12h-9.5m0 0l3.5 3.5m-3.5-3.5l3.5-3.5M4 12h1.5"
                                            stroke="currentColor"
                                            strokeWidth="2"
                                            strokeLinecap="round"
                                        />
                                    </svg>
                                )}
                                {message.type === 'system' && (
                                    <svg
                                        className="inline-block w-4 h-4 mr-1 -mt-1"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        xmlns="http://www.w3.org/2000/svg"
                                    >
                                        <path
                                            d="M13 16h-2v-6h2v6zm0-8h-2V6h2v2zm1-5H8c-1.1 0-2 .9-2 2v16c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2z"
                                            fill="currentColor"
                                        />
                                    </svg>
                                )}
                                {message.type === 'error' && (
                                    <svg
                                        className="inline-block w-4 h-4 mr-1 -mt-1"
                                        viewBox="0 0 24 24"
                                        fill="none"
                                        xmlns="http://www.w3.org/2000/svg"
                                    >
                                        <path
                                            d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                                            stroke="currentColor"
                                            strokeWidth="2"
                                            strokeLinecap="round"
                                        />
                                    </svg>
                                )}
                                <span className="text-xs text-gray-500">{formatTime(message.timestamp)}</span>
                                {message.sender && <span className="font-medium"> {message.sender}:</span>}{' '}
                                {message.text}
                            </div>
                        ))}
                        <div ref={messagesEndRef} />
                    </div>
                )}
            </div>

            <div className="p-2 border-t border-gray-200 bg-gray-50 flex items-center justify-between text-xs text-gray-600">
                <div className="flex items-center">
                    <div
                        className={`w-2 h-2 rounded-full mr-1 ${status === 'connected' ? 'bg-green-500' : 'bg-yellow-500'}`}
                    />
                    <span>{status === 'connected' ? 'Connected' : 'Connecting...'}</span>
                </div>

                {walletSnap.channelOpen && (
                    <div className="flex items-center">
                        <span className="mr-2">Channel:</span>
                        <span className="font-mono bg-white px-1 py-0.5 rounded-sm border border-gray-200 text-gray-700">
                            {walletSnap.selectedTokenAddress?.substring(0, 6)}...
                            {walletSnap.selectedTokenAddress?.substring(38)}
                        </span>
                    </div>
                )}
            </div>
        </div>
    );
}
