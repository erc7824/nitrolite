import { useRef, useEffect, useState } from 'react';
import { useMessageStyles, useMessageService } from '@/hooks/ui';
import { FullscreenMessages } from './FullscreenMessages';

export function MessageList() {
    const [isFullscreen, setIsFullscreen] = useState(false);

    const messagesEndRef = useRef<HTMLDivElement>(null);
    // const containerRef = useRef<HTMLDivElement>(null);
    const messageStyles = useMessageStyles();
    const shouldScrollRef = useRef(true);

    // Use our optimized message service
    const { messages, clearMessages, status } = useMessageService();

    // Monitor messages for changes
    useEffect(() => {
        // Messages state has changed
    }, [messages]);

    // Handle fullscreen toggle event
    useEffect(() => {
        const handleToggleFullscreen = () => {
            setIsFullscreen((prev) => !prev);
        };

        window.addEventListener('toggle-fullscreen-messages', handleToggleFullscreen);
        return () => window.removeEventListener('toggle-fullscreen-messages', handleToggleFullscreen);
    }, []);

    // Handle scroll event to detect if user has scrolled up
    useEffect(() => {
        const container = document.getElementById('message-container');

        if (!container) return;

        const handleScroll = () => {
            const isAtBottom = container.scrollHeight - container.scrollTop - container.clientHeight < 50;

            shouldScrollRef.current = isAtBottom;
        };

        container.addEventListener('scroll', handleScroll);
        return () => container.removeEventListener('scroll', handleScroll);
    }, []);

    // Auto-scroll only if user hasn't scrolled up
    useEffect(() => {
        if (shouldScrollRef.current && messagesEndRef.current) {
            const container = document.getElementById('message-container');

            if (container) {
                container.scrollTo({
                    top: container.scrollHeight,
                    behavior: 'smooth',
                });
            }
        }
    }, [messages]);

    // Format timestamp
    const formatTime = (timestamp?: number) => {
        if (!timestamp) return '';
        return new Date(timestamp).toLocaleTimeString();
    };

    // Close fullscreen handler
    const handleCloseFullscreen = () => {
        setIsFullscreen(false);
    };

    if (isFullscreen) {
        return <FullscreenMessages messages={messages} status={status} onClose={handleCloseFullscreen} />;
    }

    return (
        <div className="mb-4">
            <div className="flex justify-between items-center mb-2">
                <h2 className="text-lg font-semibold text-[#3531ff]">Messages</h2>
                <div className="flex space-x-2">
                    <button
                        onClick={() => setIsFullscreen(true)}
                        className="px-3 py-1 bg-[#3531ff] text-white text-sm rounded hover:bg-[#2b28cc] flex items-center cursor-pointer transition-colors"
                    >
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            className="h-4 w-4 mr-1"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                        >
                            <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth="2"
                                d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5v-4m0 4h-4m4 0l-5-5"
                            />
                        </svg>
                        Full View
                    </button>
                    <button
                        onClick={clearMessages}
                        className="px-3 py-1 bg-white border border-gray-200 text-gray-600 text-sm rounded hover:bg-gray-50 flex items-center cursor-pointer"
                    >
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            className="h-4 w-4 mr-1"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                        >
                            <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth="2"
                                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                            />
                        </svg>
                        Clear
                    </button>
                </div>
            </div>

            <div
                className="bg-white rounded-lg p-4 border border-gray-200 h-[calc(100vh-350px)] min-h-[300px] overflow-y-auto scrollbar-thin shadow-sm"
                id="message-container"
            >
                {messages.length === 0 ? (
                    <div className="text-gray-500 text-center py-10">No messages yet</div>
                ) : (
                    <div>
                        {messages.map((message, index) => (
                            <div
                                key={index}
                                className={`p-1 rounded mb-1.5 ${messageStyles[message.type] || messageStyles.info}`}
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
        </div>
    );
}
