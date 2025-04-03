import { useState, ChangeEvent } from 'react';
import { NitroliteRPC } from '@erc7824/nitrolite';
import { Channel } from '@/types';
import { useMessageService } from '@/hooks/useMessageService';

interface RequestFormProps {
    isConnected: boolean;
    currentChannel: Channel | null;
    onSendRequest: (methodName: string, methodParams: string) => void;
    onSendMessage: (message: string) => void;
    onSubscribeToChannel: (channel: Channel) => void;
    onSendPing: () => void;
    onCheckBalance: () => void;
}

// Common NitroRPC methods for quick access
const COMMON_RPC_METHODS = [
    { name: 'ping', description: 'Simple ping to check connectivity' },
    { name: 'subscribe', description: 'Subscribe to a channel' },
    { name: 'publish', description: 'Publish message to a channel' },
    { name: 'balance', description: 'Check token balance' },
];

export function RequestForm({
    isConnected,
    currentChannel,
    onSendRequest,
    onSendMessage,
    onSubscribeToChannel,
    onSendPing,
    onCheckBalance,
}: RequestFormProps) {
    // States for form inputs
    const [methodName, setMethodName] = useState<string>('ping');
    const [methodParams, setMethodParams] = useState<string>('');
    const [message, setMessage] = useState<string>('');
    const [selectedChannel, setSelectedChannel] = useState<Channel>('public');
    const [showMethodList, setShowMethodList] = useState<boolean>(false);

    // Use our message service hook
    const { activeChannel, addSystemMessage } = useMessageService();

    // Event handlers
    const handleMethodNameChange = (e: ChangeEvent<HTMLInputElement>) => {
        setMethodName(e.target.value);
    };

    const handleMethodParamsChange = (e: ChangeEvent<HTMLTextAreaElement>) => {
        setMethodParams(e.target.value);
    };

    const handleSendRequest = () => {
        try {
            // Create a formatted request using NitroliteRPC to validate format
            let params: unknown[] = [];

            if (methodParams.trim()) {
                try {
                    params = JSON.parse(methodParams);
                    if (!Array.isArray(params)) params = [params];
                } catch (e) {
                    addSystemMessage(`Error parsing params: ${e instanceof Error ? e.message : String(e)}`);
                    return;
                }
            }

            // Create request format for display purposes
            const request = NitroliteRPC.createRequest(methodName, params);

            addSystemMessage(`Sending NitroRPC request: ${JSON.stringify(request)}`);

            // Send the actual request
            onSendRequest(methodName, methodParams);
        } catch (error) {
            addSystemMessage(`RPC request error: ${error instanceof Error ? error.message : String(error)}`);
        }
    };

    const selectPredefinedMethod = (method: string) => {
        setMethodName(method);
        setShowMethodList(false);

        // Set default parameters based on method
        if (method === 'subscribe' && currentChannel) {
            setMethodParams(JSON.stringify([currentChannel]));
        } else if (method === 'publish' && currentChannel) {
            setMethodParams(JSON.stringify([currentChannel, 'Hello from NitroRPC']));
        } else if (method === 'balance') {
            setMethodParams(JSON.stringify(['0xToken...']));
        } else {
            setMethodParams('[]');
        }
    };

    const handleChannelSelect = (e: ChangeEvent<HTMLSelectElement>) => {
        setSelectedChannel(e.target.value as Channel);
    };

    const handleSubscribe = () => {
        onSubscribeToChannel(selectedChannel);
    };

    const handleMessageChange = (e: ChangeEvent<HTMLInputElement>) => {
        setMessage(e.target.value);
    };

    const handleSendMessage = () => {
        if (message.trim()) {
            onSendMessage(message);
            setMessage('');
        }
    };

    const handleKeyPress = (e: React.KeyboardEvent<HTMLInputElement>) => {
        if (e.key === 'Enter') {
            handleSendMessage();
        }
    };

    return (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            {/* Channel Panel */}
            <div className="md:col-span-3 p-4 bg-white rounded-lg border border-gray-200 shadow-sm">
                <h2 className="text-lg font-semibold mb-3 text-[#3531ff]">Channel: {activeChannel}</h2>

                <div className="flex mb-3 space-x-2">
                    <div className="flex-grow">
                        <select
                            value={selectedChannel}
                            onChange={handleChannelSelect}
                            disabled={!isConnected}
                            className="w-full bg-white text-gray-700 rounded border border-gray-200 focus:border-[#3531ff] focus:ring focus:ring-[#3531ff] focus:ring-opacity-30 py-2 px-4 disabled:bg-gray-100 disabled:text-gray-400"
                        >
                            <option value="public">Public</option>
                            <option value="game">Game</option>
                            <option value="trade">Trade</option>
                            <option value="private">Private</option>
                        </select>
                    </div>
                    <button
                        onClick={handleSubscribe}
                        disabled={!isConnected}
                        className="bg-[#3531ff] hover:bg-[#2b28cc] disabled:bg-gray-200 disabled:text-gray-400 text-white font-medium py-2 px-4 rounded transition-colors cursor-pointer disabled:cursor-not-allowed"
                    >
                        Subscribe
                    </button>
                </div>

                <div className="flex space-x-2">
                    <input
                        type="text"
                        value={message}
                        onChange={handleMessageChange}
                        onKeyPress={handleKeyPress}
                        placeholder="Type your message..."
                        disabled={!isConnected || !currentChannel}
                        className="flex-grow bg-white text-gray-700 rounded border border-gray-200 focus:border-[#3531ff] focus:outline-none focus:ring focus:ring-[#3531ff] focus:ring-opacity-30 py-2 px-4 disabled:bg-gray-100 disabled:text-gray-400"
                    />
                    <button
                        onClick={handleSendMessage}
                        disabled={!isConnected || !currentChannel || !message.trim()}
                        className="bg-[#3531ff] hover:bg-[#2b28cc] disabled:bg-gray-200 disabled:text-gray-400 text-white font-medium py-2 px-4 rounded transition-colors cursor-pointer disabled:cursor-not-allowed"
                    >
                        Send
                    </button>
                </div>
            </div>

            {/* NitroRPC Operations Panel */}
            <div className="md:col-span-1 p-4 bg-white rounded-lg border border-gray-200 shadow-sm">
                <h2 className="text-lg font-semibold mb-3 text-[#3531ff]">NitroRPC</h2>

                <div className="space-y-4">
                    <div className="flex flex-wrap gap-2">
                        <button
                            onClick={onSendPing}
                            disabled={!isConnected}
                            className="bg-[#3531ff] hover:bg-[#2b28cc] disabled:bg-gray-200 disabled:text-gray-400 text-white font-medium py-2 px-4 rounded transition-colors flex items-center cursor-pointer disabled:cursor-not-allowed"
                        >
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                className="h-5 w-5 mr-2"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                            >
                                <path
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    strokeWidth="2"
                                    d="M8 12h.01M12 12h.01M16 12h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                                />
                            </svg>
                            Ping Server
                        </button>

                        <button
                            onClick={onCheckBalance}
                            disabled={!isConnected}
                            className="bg-[#3531ff] hover:bg-[#2b28cc] disabled:bg-gray-200 disabled:text-gray-400 text-white font-medium py-2 px-4 rounded transition-colors flex items-center cursor-pointer disabled:cursor-not-allowed"
                        >
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                className="h-5 w-5 mr-2"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                            >
                                <path
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    strokeWidth="2"
                                    d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                                />
                            </svg>
                            Check Balance
                        </button>
                    </div>

                    <div className="relative">
                        <label className="block text-sm text-gray-600 mb-1">NitroRPC Method</label>
                        <div className="flex">
                            <input
                                type="text"
                                value={methodName}
                                onChange={handleMethodNameChange}
                                placeholder="e.g. ping, subscribe, publish"
                                disabled={!isConnected}
                                className="flex-grow p-2 bg-white text-gray-700 rounded-l border border-gray-200 disabled:bg-gray-100 disabled:text-gray-400"
                            />
                            <button
                                onClick={() => setShowMethodList(!showMethodList)}
                                disabled={!isConnected}
                                className="p-2 bg-gray-100 text-gray-700 rounded-r border border-l-0 border-gray-200 hover:bg-gray-200"
                            >
                                <svg
                                    xmlns="http://www.w3.org/2000/svg"
                                    className="h-5 w-5"
                                    viewBox="0 0 20 20"
                                    fill="currentColor"
                                >
                                    <path
                                        fillRule="evenodd"
                                        d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
                                        clipRule="evenodd"
                                    />
                                </svg>
                            </button>
                        </div>

                        {showMethodList && (
                            <div className="absolute z-10 w-full mt-1 bg-white border border-gray-200 rounded-md shadow-lg">
                                {COMMON_RPC_METHODS.map((method) => (
                                    <div
                                        key={method.name}
                                        onClick={() => selectPredefinedMethod(method.name)}
                                        className="p-2 hover:bg-gray-100 cursor-pointer border-b border-gray-100"
                                    >
                                        <div className="font-medium">{method.name}</div>
                                        <div className="text-xs text-gray-500">{method.description}</div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>

                    <div>
                        <label className="block text-sm text-gray-600 mb-1">Parameters (JSON array)</label>
                        <textarea
                            value={methodParams}
                            onChange={handleMethodParamsChange}
                            placeholder="e.g. [42, 23]"
                            disabled={!isConnected}
                            rows={3}
                            className="w-full p-2 bg-white text-gray-700 rounded border border-gray-200 font-mono text-sm disabled:bg-gray-100 disabled:text-gray-400"
                        />
                    </div>

                    <button
                        onClick={handleSendRequest}
                        disabled={!isConnected || !methodName.trim()}
                        className="w-full bg-[#3531ff] hover:bg-[#2b28cc] disabled:bg-gray-200 disabled:text-gray-400 text-white font-medium py-2 px-4 rounded transition-colors cursor-pointer disabled:cursor-not-allowed"
                    >
                        Send NitroRPC Request
                    </button>
                </div>
            </div>
        </div>
    );
}
