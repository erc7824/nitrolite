import { Copy, Check } from 'lucide-react';
import { Badge } from './ui/badge';
import { Button } from './ui/button';
import { useState, useEffect } from 'react';
import { cn } from '../lib/utils';

interface RoomInfoProps {
  roomId: string;
}

export function RoomInfo({ roomId }: RoomInfoProps) {
  const [copied, setCopied] = useState(false);
  
  // Copy room ID to clipboard
  const copyToClipboard = async () => {
    try {
      await navigator.clipboard.writeText(roomId);
      setCopied(true);
    } catch (err) {
      console.error('Failed to copy room ID', err);
    }
  };
  
  // Reset copied state after a delay
  useEffect(() => {
    if (copied) {
      const timeout = setTimeout(() => setCopied(false), 2000);
      return () => clearTimeout(timeout);
    }
  }, [copied]);

  // Don't show anything if roomId is not set yet
  if (!roomId) {
    return null;
  }
  
  return (
    <div className="flex items-center justify-center w-full text-sm">
      <Badge 
        variant="default" 
        className={cn(
          "px-3 py-1.5 h-8 border border-gray-800/50 shadow-sm group transition-colors duration-300",
          copied ? "bg-gray-800/70 border-gray-700/50" : "bg-gray-800/50 hover:bg-gray-800/70"
        )}
      >
        <span className="text-gray-400 mr-2">Room:</span>
        <span className={cn(
          "font-mono transition-colors",
          copied ? "text-viper-green" : "text-gray-300"
        )}>
          {roomId.length > 12 ? `${roomId.substring(0, 8)}...` : roomId}
        </span>
        <Button 
          variant="ghost" 
          size="icon"
          className={cn(
            "h-5 w-5 ml-2 hover:bg-transparent transition-colors",
            copied ? "text-viper-green hover:text-viper-green-light" : "text-viper-green hover:text-viper-green-light"
          )}
          onClick={copyToClipboard}
        >
          {copied ? (
            <Check className="h-3.5 w-3.5 animate-fadeIn" />
          ) : (
            <Copy className="h-3.5 w-3.5" />
          )}
          <span className="sr-only">{copied ? "Copied" : "Copy room ID"}</span>
        </Button>
      </Badge>
    </div>
  );
}