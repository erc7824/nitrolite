import { useState, ChangeEvent } from 'react';
import { Channel as AppChannel } from '@/types';
import { useMessageService } from '@/hooks/ui';
import { Channel } from '@erc7824/nitrolite';

interface RequestFormProps {
    isConnected: boolean;
    currentChannel: Channel | null;
    onSendRequest: (methodName: string, methodParams: string) => void;
    onSendMessage: (message: string, channelOverride?: AppChannel) => void;
    onSendPing: () => void;
}

export function RequestForm({
    isConnected,
    currentChannel,
    onSendRequest,
    onSendMessage,
    onSendPing,
}: RequestFormProps) {
    // States for form inputs
    const [methodName, setMethodName] = useState<string>('ping');
    const [methodParams, setMethodParams] = useState<string>('');
    const [message, setMessage] = useState<string>('');
    const [selectedChannel, setSelectedChannel] = useState<AppChannel>('public');

    // Use our message service hook
    const { activeChannel } = useMessageService();

    // Event handlers
    const handleMethodNameChange = (e: ChangeEvent<HTMLInputElement>) => {
        setMethodName(e.target.value);
    };

    const handleMethodParamsChange = (e: ChangeEvent<HTMLTextAreaElement>) => {
        setMethodParams(e.target.value);
    };

    const handleSendRequest = () => {
        onSendRequest(methodName, methodParams);
    };

    const handleChannelSelect = (e: ChangeEvent<HTMLSelectElement>) => {
        setSelectedChannel(e.target.value as AppChannel);
    };

    // Channel selection is now just for display/message override

    const handleMessageChange = (e: ChangeEvent<HTMLInputElement>) => {
        setMessage(e.target.value);
    };

    const handleSendMessage = () => {
        if (message.trim()) {
            // Pass the selected channel as an override
            onSendMessage(message, selectedChannel);
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
                    {/* Channel selection is used for message routing */}
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

            {/* Operations Panel */}
            <div className="md:col-span-1 p-4 bg-white rounded-lg border border-gray-200 shadow-sm">
                <h2 className="text-lg font-semibold mb-3 text-[#3531ff]">Operations</h2>

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

                        {/* Only ping button remains */}
                    </div>

                    <div>
                        <label className="block text-sm text-gray-600 mb-1">Custom RPC Method</label>
                        <input
                            type="text"
                            value={methodName}
                            onChange={handleMethodNameChange}
                            placeholder="e.g. ping, add, subtract"
                            disabled={!isConnected}
                            className="w-full p-2 bg-white text-gray-700 rounded border border-gray-200 disabled:bg-gray-100 disabled:text-gray-400 mb-2"
                        />
                    </div>

                    <div>
                        <label className="block text-sm text-gray-600 mb-1">Parameters (JSON)</label>
                        <textarea
                            value={methodParams}
                            onChange={handleMethodParamsChange}
                            placeholder="e.g. [42, 23]"
                            disabled={!isConnected}
                            rows={2}
                            className="w-full p-2 bg-white text-gray-700 rounded border border-gray-200 font-mono text-sm disabled:bg-gray-100 disabled:text-gray-400 mb-2"
                        />
                    </div>

                    <button
                        onClick={handleSendRequest}
                        disabled={!isConnected || !methodName.trim()}
                        className="w-full bg-[#3531ff] hover:bg-[#2b28cc] disabled:bg-gray-200 disabled:text-gray-400 text-white font-medium py-2 px-4 rounded transition-colors cursor-pointer disabled:cursor-not-allowed"
                    >
                        Send Custom Request
                    </button>
                </div>
            </div>
        </div>
    );
}
