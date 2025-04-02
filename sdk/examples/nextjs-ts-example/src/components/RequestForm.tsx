import { useState, ChangeEvent } from 'react';
import { Channel } from '@/types';
import { useMessageService } from '@/hooks/useMessageService';

type RequestFormProps = {
  isConnected: boolean;
  currentChannel: Channel | null;
  onSendRequest: (methodName: string, methodParams: string) => void;
  onSendMessage: (message: string) => void;
  onSubscribeToChannel: (channel: Channel) => void;
  onSendPing: () => void;
  onCheckBalance: () => void;
};

export function RequestForm({ 
  isConnected, 
  currentChannel,
  onSendRequest, 
  onSendMessage,
  onSubscribeToChannel,
  onSendPing,
  onCheckBalance
}: RequestFormProps) {
  // States for form inputs
  const [methodName, setMethodName] = useState<string>("ping");
  const [methodParams, setMethodParams] = useState<string>("");
  const [message, setMessage] = useState<string>("");
  const [selectedChannel, setSelectedChannel] = useState<Channel>("public");
  
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
      setMessage("");
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
              <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              Ping Server
            </button>
            
            <button 
              onClick={onCheckBalance}
              disabled={!isConnected}
              className="bg-[#3531ff] hover:bg-[#2b28cc] disabled:bg-gray-200 disabled:text-gray-400 text-white font-medium py-2 px-4 rounded transition-colors flex items-center cursor-pointer disabled:cursor-not-allowed"
            >
              <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              Check Balance
            </button>
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
              placeholder='e.g. [42, 23]'
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