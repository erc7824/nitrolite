import { useRef, useEffect } from 'react';
import { useMessageStyles, MessageType } from '@/hooks/useMessageStyles';
import { Message } from '@/types';

type MessageListProps = {
  messages: Message[];
  onClear: () => void;
};

export function MessageList({ messages, onClear }: MessageListProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const messageStyles = useMessageStyles();
  const shouldScrollRef = useRef(true);

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
      // Scroll the container element itself, not the whole page
      const container = document.getElementById('message-container');
      if (container) {
        container.scrollTo({
          top: container.scrollHeight,
          behavior: 'smooth'
        });
      }
    }
  }, [messages]);

  // Format timestamp
  const formatTime = (timestamp?: number) => {
    if (!timestamp) return '';
    return new Date(timestamp).toLocaleTimeString();
  };

  return (
    <div className="mb-6">
      <div className="flex justify-between items-center mb-2">
        <h2 className="text-lg font-semibold">Messages</h2>
        <button 
          onClick={onClear}
          className="px-3 py-1 bg-gray-700 text-white text-sm rounded hover:bg-gray-600 flex items-center"
        >
          <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
          Clear
        </button>
      </div>
      
      <div className="bg-gray-800 rounded-lg p-4 h-80 overflow-y-auto scrollbar-thin" id="message-container">
        {messages.length === 0 ? (
          <div className="text-gray-500 text-center py-10">No messages yet</div>
        ) : (
          <div>
            {messages.map((message, index) => (
              <div key={index} className={`p-3 rounded-lg mb-2 ${messageStyles[message.type] || messageStyles.info}`}>
                {message.type === 'sent' && (
                  <svg className="inline-block w-4 h-4 mr-1 -mt-1" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M5 13l4 4L19 7" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  </svg>
                )}
                
                {message.type === 'received' && (
                  <svg className="inline-block w-4 h-4 mr-1 -mt-1" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M20 12h-9.5m0 0l3.5 3.5m-3.5-3.5l3.5-3.5M4 12h1.5" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                  </svg>
                )}
                
                {message.type === 'system' && (
                  <svg className="inline-block w-4 h-4 mr-1 -mt-1" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M13 16h-2v-6h2v6zm0-8h-2V6h2v2zm1-5H8c-1.1 0-2 .9-2 2v16c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2z" fill="currentColor"/>
                  </svg>
                )}
                
                {message.type === 'error' && (
                  <svg className="inline-block w-4 h-4 mr-1 -mt-1" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                  </svg>
                )}
                
                <span className="text-xs text-gray-500">{formatTime(message.timestamp)}</span>
                
                {message.sender && (
                  <span className="font-medium"> {message.sender}:</span>
                )} {message.text}
              </div>
            ))}
            <div ref={messagesEndRef} />
          </div>
        )}
      </div>
    </div>
  );
}