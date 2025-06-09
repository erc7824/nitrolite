import { useState, useEffect } from "react";
import { Lobby } from "./Lobby";
// import { ChannelDeposit } from './ChannelDeposit';
import { useChannel } from "../hooks/useChannel";
import { useWebSocketContext } from "../context/WebSocketContext";
import { useNitroliteIntegration } from "../hooks/useNitroliteIntegration";
import { useAccount } from 'wagmi';
import type { JoinRoomPayload, AvailableRoom } from "../types";
import { Card, CardDescription, CardHeader, CardTitle } from "./ui/card";
import { Loader2, AlertCircle } from "lucide-react";

/**
 * Integrated game lobby that handles channel state
 */
interface GameLobbyIntegratedProps {
    onJoinRoom: (payload: JoinRoomPayload) => void;
    availableRooms: AvailableRoom[];
    onGetAvailableRooms: () => void;
    onlineUsers?: number;
}

export function GameLobbyIntegrated({ onJoinRoom, availableRooms = [], onGetAvailableRooms, onlineUsers = 1 }: GameLobbyIntegratedProps) {
    const [isLoading, setIsLoading] = useState(true);
    const { isConnected, status } = useWebSocketContext();
    const { clearStoredChannel } = useChannel();
    const { isConnected: isMetaMaskConnected } = useAccount();
    useNitroliteIntegration(); // Ensure proper integration
    
    // Handle channel initialization on component mount
    useEffect(() => {
        // Check if nitrolite_channel_id exists in localStorage
        const channelId = localStorage.getItem('nitrolite_channel_id');
        if (!channelId) {
            console.log("GameLobbyIntegrated: No nitrolite_channel_id found in localStorage");
            // If we need to force a channel creation, we can clear any stale state
            clearStoredChannel();
        } else {
            console.log("GameLobbyIntegrated: Found nitrolite_channel_id in localStorage:", channelId);
        }
    }, [clearStoredChannel]);

    // Simulate loading state
    useEffect(() => {
        const timer = setTimeout(() => {
            setIsLoading(false);
        }, 1000);

        return () => clearTimeout(timer);
    }, []);

    // Show loading state
    if (isLoading) {
        return (
            <>
                <div className="flex justify-center items-center min-h-[220px]">
                    <Loader2 className="h-12 w-12 animate-spin text-viper-green" />
                </div>
            </>
        );
    }

    // Show Lobby component which has built-in MetaMask connection UI
    if (!isMetaMaskConnected) {
        return (
            <>
                <Lobby
                    onJoinRoom={onJoinRoom}
                    isConnected={isConnected}
                    error={null}
                    availableRooms={availableRooms}
                    onGetAvailableRooms={onGetAvailableRooms}
                    onlineUsers={onlineUsers}
                />
            </>
        );
    }

    // Show WebSocket connection status if MetaMask is connected but WebSocket is not
    if (!isConnected) {
        return (
            <>
                <Card className="mb-6">
                    <CardHeader className="pb-2">
                        <CardTitle className="text-xl text-viper-yellow flex items-center gap-2">
                            <AlertCircle className="h-5 w-5" /> WebSocket Status: {status}
                        </CardTitle>
                        <CardDescription>
                            Connecting to the game server...
                            {status === "reconnect_failed" && (
                                <span className="block mt-2 text-red-400">Connection failed. Please refresh the page and try again.</span>
                            )}
                        </CardDescription>
                    </CardHeader>
                </Card>
            </>
        );
    }

    // The channel creation is now handled on-demand when joining/creating a game
    // We'll keep this code commented for reference
    /*
  if (!isChannelOpen) {
    return (
      <>
        <AppHeader />
        <div className="space-y-4">
          <Card className="mb-2">
            <CardHeader className="pb-3">
              <CardTitle className="text-xl text-cyan-400">Channel Required</CardTitle>
              <CardDescription>
                You need to create a channel and deposit funds to play games.
              </CardDescription>
            </CardHeader>
          </Card>
          
          <ChannelDeposit />
        </div>
      </>
    );
  }
  */

    // Show the game lobby when connected and channel is open
    return (
        <>
            <Lobby
                onJoinRoom={onJoinRoom}
                isConnected={isConnected}
                error={null}
                availableRooms={availableRooms}
                onGetAvailableRooms={onGetAvailableRooms}
                onlineUsers={onlineUsers}
            />
        </>
    );
}
