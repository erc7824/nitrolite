import { useState, useEffect } from "react";
import type { JoinRoomPayload, AvailableRoom } from "../types";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "./ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./ui/tabs";
import { Input } from "./ui/input";
import { Button } from "./ui/button";
import { cn } from "../lib/utils";
import { Wallet, Users, Loader2, KeyRound, GamepadIcon, RefreshCw, Clock, AlertCircle } from "lucide-react";
import { useMetaMask } from "../hooks/useMetaMask";
import { ChannelRequiredModal } from "./ChannelRequiredModal";
import { OnlinePlayersCounter } from "./OnlinePlayersCounter";

interface LobbyProps {
    onJoinRoom: (payload: JoinRoomPayload) => void;
    isConnected: boolean;
    error: string | null;
    availableRooms?: AvailableRoom[];
    onGetAvailableRooms: () => void;
    onlineUsers?: number;
}

export function Lobby({ onJoinRoom, isConnected, error, availableRooms = [], onGetAvailableRooms, onlineUsers = 1 }: LobbyProps) {
    const [roomId, setRoomId] = useState("");
    const [roomIdError, setRoomIdError] = useState("");
    const [mode, setMode] = useState<"create" | "join">("create");
    const [loadingRooms, setLoadingRooms] = useState(false);
    const [showChannelModal, setShowChannelModal] = useState(false);
    const [pendingRoomAction, setPendingRoomAction] = useState<{ mode: "create" | "join"; roomId?: string } | null>(null);

    // Use channel hook to check if channel exists
    const { isChannelOpen } = { isChannelOpen: true };

    // Use MetaMask hook for wallet connection
    const { address, isConnected: isWalletConnected, connectWallet, isConnecting, error: metamaskError, isMetaMaskInstalled } = useMetaMask();

    // Fetch available rooms when tab changes to 'join'
    useEffect(() => {
        if (mode === "join" && isConnected && isWalletConnected) {
            setLoadingRooms(true);
            onGetAvailableRooms();
            // Set a timeout to hide the loading indicator after 2 seconds
            // (in case the server doesn't respond or takes too long)
            const timeoutId = setTimeout(() => {
                setLoadingRooms(false);
            }, 2000);

            return () => clearTimeout(timeoutId);
        }
    }, [mode, isConnected, isWalletConnected, onGetAvailableRooms]);

    // When availableRooms changes, stop the loading indicator
    useEffect(() => {
        setLoadingRooms(false);
    }, [availableRooms]);

    // Validate Room ID format for joining
    const validateRoomId = (id: string): boolean => {
        // For joining, room ID is required
        if (mode === "join") {
            if (!id.trim()) {
                setRoomIdError("Room ID is required when joining a game");
                return false;
            }

            // Should be a valid UUID format (8-4-4-4-12 hex chars)
            const isValid = /^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$/.test(id);
            setRoomIdError(isValid ? "" : "Please enter a valid room ID (UUID format)");
            return isValid;
        }

        // For creating, room ID is ignored
        return true;
    };

    // Handle tab change
    const handleTabChange = (value: string) => {
        setMode(value as "create" | "join");
        setRoomIdError("");

        // Fetch available rooms when switching to join tab
        if (value === "join" && isConnected && isWalletConnected) {
            onGetAvailableRooms();
        }
    };

    // Helper function to format time since creation
    const formatTimeAgo = (timestamp: number): string => {
        const now = Date.now();
        const secondsAgo = Math.floor((now - timestamp) / 1000);

        if (secondsAgo < 60) {
            return `${secondsAgo} second${secondsAgo !== 1 ? "s" : ""} ago`;
        }

        const minutesAgo = Math.floor(secondsAgo / 60);
        if (minutesAgo < 60) {
            return `${minutesAgo} minute${minutesAgo !== 1 ? "s" : ""} ago`;
        }

        const hoursAgo = Math.floor(minutesAgo / 60);
        return `${hoursAgo} hour${hoursAgo !== 1 ? "s" : ""} ago`;
    };

    // Handle joining a specific available room
    const handleJoinAvailableRoom = (selectedRoomId: string) => {
        if (!isWalletConnected || !address) {
            return;
        }

        // Check if channel exists first
        if (!isChannelOpen) {
            setPendingRoomAction({ mode: "join", roomId: selectedRoomId });
            setShowChannelModal(true);
            return;
        }

        // Use MetaMask wallet address for app session participants
        console.log("Joining available room with MetaMask address:", address, "and roomId:", selectedRoomId);
        onJoinRoom({ eoa: address, roomId: selectedRoomId });
    };

    // Handle manual refresh of available rooms
    const handleRefreshRooms = () => {
        if (isConnected && isWalletConnected) {
            setLoadingRooms(true);
            onGetAvailableRooms();
        }
    };

    // Handle form submission
    const handleSubmit = (event: React.FormEvent) => {
        event.preventDefault();

        if (!isWalletConnected || !address) {
            return;
        }

        const isRoomIdValid = validateRoomId(roomId);

        if (!isRoomIdValid && mode === "join") {
            return;
        }

        // Check if channel exists first
        if (!isChannelOpen) {
            if (mode === "create") {
                setPendingRoomAction({ mode: "create" });
            } else {
                setPendingRoomAction({ mode: "join", roomId: roomId.trim() });
            }
            setShowChannelModal(true);
            return;
        }

        // Use MetaMask wallet address for app session participants
        if (mode === "create") {
            // When creating a room, always pass undefined for roomId
            console.log("Creating a room with MetaMask address:", address);
            onJoinRoom({ eoa: address, roomId: undefined });
        } else {
            // When joining, use the entered roomId
            console.log("Joining a room with MetaMask address:", address, "and roomId:", roomId.trim());
            onJoinRoom({ eoa: address, roomId: roomId.trim() });
        }
    };

    // Handle successful channel creation
    const handleChannelSuccess = (action: "join" | "create", roomIdParam?: string) => {
        if (!address) return;

        // Use MetaMask wallet address for app session participants

        // Add a debounce mechanism to prevent duplicate calls
        const now = Date.now();
        const lastCallTime = window.localStorage.getItem("last_join_call_time");
        const throttleTime = 2000; // 2 seconds

        if (lastCallTime && now - parseInt(lastCallTime) < throttleTime) {
            console.log("Throttling joinRoom call - too soon after previous call");
            return;
        }

        // Store this call time
        window.localStorage.setItem("last_join_call_time", now.toString());

        if (action === "create") {
            console.log("Creating a room with MetaMask address after channel creation:", address);
            onJoinRoom({ eoa: address, roomId: undefined });
        } else {
            console.log("Joining a room with MetaMask address after channel creation:", address, "and roomId:", roomIdParam);
            onJoinRoom({ eoa: address, roomId: roomIdParam });
        }
    };

    // Handle connect wallet button click
    const handleConnectWallet = async () => {
        await connectWallet();
    };

    return (
        <div className="flex flex-col items-center relative">
            {/* Channel creation modal */}
            <ChannelRequiredModal
                isOpen={showChannelModal}
                onClose={() => setShowChannelModal(false)}
                onSuccess={handleChannelSuccess}
                mode={pendingRoomAction?.mode || "create"}
                roomId={pendingRoomAction?.roomId}
            />
            {/* Animated glow behind card */}
            <div className="absolute -inset-8 bg-gradient-to-br from-cyan-500/5 via-transparent to-fuchsia-500/5 rounded-full blur-2xl"></div>

            <Card className="w-full max-w-md mx-auto shadow-[0_0_30px_rgba(0,229,255,0.15)] border-gray-800/50 bg-gray-900/80 backdrop-blur-sm overflow-hidden">
                {/* Header with particle effect */}
                <CardHeader className="text-center relative">
                    <div className="absolute inset-0 opacity-10 overflow-hidden">
                        <div className="absolute w-full h-[200%] top-[-50%] left-0 bg-[radial-gradient(circle,_white_1px,_transparent_1px)] bg-[length:20px_20px] opacity-20 animate-sparkle"></div>
                    </div>

                    <CardTitle className="text-5xl font-bold mb-2 relative z-10">
                        <span className="text-glow-cyan">Nitro</span>
                        <span className="text-glow-magenta ml-1">Aura</span>
                    </CardTitle>
                    <CardDescription className="text-base">Every move leaves an aura.</CardDescription>
                </CardHeader>

                <CardContent className="relative z-10">
                    {!isMetaMaskInstalled ? (
                        <div className="space-y-6 p-2">
                            <div className="bg-amber-900/20 border border-amber-800/30 rounded-lg p-4 text-amber-300 flex flex-col items-center text-center space-y-4">
                                <AlertCircle className="h-10 w-10 text-amber-400" />
                                <div>
                                    <h3 className="font-medium text-lg mb-2">MetaMask Not Detected</h3>
                                    <p className="text-sm text-amber-200/80 mb-2">
                                        To play Nitro Aura, you need to install the MetaMask browser extension.
                                    </p>
                                    <a
                                        href="https://metamask.io/download/"
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        className="text-sm font-medium text-amber-200 hover:text-amber-100 underline underline-offset-4"
                                    >
                                        Install MetaMask
                                    </a>
                                </div>
                            </div>
                        </div>
                    ) : !isWalletConnected ? (
                        <div className="space-y-6 p-2">
                            <div className="bg-gradient-to-br from-cyan-950/30 to-fuchsia-950/30 border border-gray-700/40 rounded-lg p-8 text-center space-y-6">
                                <Wallet className="h-12 w-12 mx-auto text-cyan-400 opacity-80" />
                                <div>
                                    <h3 className="font-medium text-lg mb-2 bg-gradient-to-r from-cyan-400 to-fuchsia-400 bg-clip-text text-transparent">
                                        Connect Your Wallet
                                    </h3>
                                    <p className="text-sm text-gray-400 mb-6">Connect your MetaMask wallet to create or join games.</p>
                                    <Button onClick={handleConnectWallet} variant="glowCyan" size="xl" className="w-full" disabled={isConnecting}>
                                        {isConnecting ? (
                                            <>
                                                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                                                Connecting...
                                            </>
                                        ) : (
                                            <>
                                                <Wallet className="h-4 w-4 mr-2" />
                                                Connect MetaMask
                                            </>
                                        )}
                                    </Button>

                                    {metamaskError && <p className="mt-4 text-sm text-red-400">{metamaskError}</p>}
                                </div>
                            </div>
                        </div>
                    ) : (
                        <Tabs defaultValue="create" onValueChange={handleTabChange}>
                            <TabsList className="grid grid-cols-2 p-1 mb-5">
                                <TabsTrigger
                                    value="create"
                                    className="data-[state=active]:bg-cyan-950/50 data-[state=active]:text-cyan-400 data-[state=active]:shadow-[0_0_10px_rgba(0,229,255,0.2)]"
                                    disabled={!isConnected}
                                >
                                    <GamepadIcon className="w-4 h-4 mr-2" />
                                    Create Game
                                </TabsTrigger>
                                <TabsTrigger
                                    value="join"
                                    className="data-[state=active]:bg-fuchsia-950/50 data-[state=active]:text-fuchsia-400 data-[state=active]:shadow-[0_0_10px_rgba(255,73,225,0.2)]"
                                    disabled={!isConnected}
                                >
                                    <Users className="w-4 h-4 mr-2" />
                                    Join Game
                                </TabsTrigger>
                            </TabsList>

                            <form onSubmit={handleSubmit} className="space-y-6 mt-2">
                                {/* Wallet address display */}
                                <div className="space-y-1.5">
                                    <label className="block text-sm font-medium text-gray-300 flex items-center">
                                        <Wallet className="h-4 w-4 mr-1.5 text-gray-500" />
                                        Connected Wallet
                                    </label>
                                    <div className="flex items-center p-2 bg-gray-900/50 border border-gray-800/50 rounded-md">
                                        <div className="flex-1 font-mono text-sm text-gray-300 truncate px-2">{address}</div>
                                        <div className="flex-shrink-0 ml-2">
                                            <div className="flex items-center space-x-1 bg-green-950/50 text-green-400 text-xs py-1 px-2 rounded-full border border-green-900/50">
                                                <span className="h-2 w-2 rounded-full bg-green-500"></span>
                                                <span>Connected</span>
                                            </div>
                                        </div>
                                    </div>
                                </div>

                                {/* Join tab content */}
                                <TabsContent value="join" className="space-y-5 mt-4 mb-0">
                                    {/* Betting amount notice */}
                                    <div className="rounded-md bg-amber-950/20 p-4 text-sm border border-amber-900/30 shadow-inner">
                                        <p className="mb-2 text-amber-400 font-medium flex items-center">Game Stakes</p>
                                        <p className="text-amber-200 text-sm opacity-90">
                                            Each player bets <span className="font-bold text-amber-300">$0.01</span>. Winner takes all!
                                        </p>
                                    </div>

                                    {/* Available games list */}
                                    <div className="rounded-md bg-fuchsia-950/10 p-4 border border-fuchsia-900/20 shadow-inner space-y-3">
                                        <div className="flex items-center justify-between mb-2">
                                            <h3 className="text-fuchsia-400 font-medium flex items-center">
                                                <Users className="h-4 w-4 mr-1.5" />
                                                Available Games
                                            </h3>
                                            <Button
                                                variant="ghost"
                                                size="sm"
                                                className="h-8 px-2 text-gray-400 hover:text-fuchsia-400"
                                                onClick={handleRefreshRooms}
                                                disabled={!isConnected || loadingRooms}
                                            >
                                                <RefreshCw className={cn("h-4 w-4 mr-1", loadingRooms && "animate-spin")} />
                                                Refresh
                                            </Button>
                                        </div>

                                        {loadingRooms ? (
                                            <div className="flex items-center justify-center py-8">
                                                <div className="flex flex-col items-center space-y-2">
                                                    <Loader2 className="h-6 w-6 text-fuchsia-400 animate-spin" />
                                                    <p className="text-sm text-gray-400">Loading available games...</p>
                                                </div>
                                            </div>
                                        ) : availableRooms.length > 0 ? (
                                            <div className="space-y-2">
                                                {availableRooms.map((room) => (
                                                    <div
                                                        key={room.roomId}
                                                        className="bg-gray-800/50 rounded-md p-3 border border-gray-700/40 hover:border-fuchsia-800/30 group transition-all hover:bg-gray-800/70"
                                                    >
                                                        <div className="flex items-center justify-between">
                                                            <div className="flex-1">
                                                                <div className="flex items-center">
                                                                    <span className="text-gray-300 font-mono text-sm truncate max-w-[200px]">
                                                                        {room.roomId}
                                                                    </span>
                                                                </div>
                                                                <div className="flex items-center mt-1 text-xs text-gray-500">
                                                                    <Clock className="h-3 w-3 mr-1" />
                                                                    <span>Created {formatTimeAgo(room.createdAt)}</span>
                                                                </div>
                                                            </div>
                                                            <Button
                                                                variant="glowMagenta"
                                                                size="sm"
                                                                className="ml-2 whitespace-nowrap"
                                                                disabled={!isConnected}
                                                                onClick={() => handleJoinAvailableRoom(room.roomId)}
                                                            >
                                                                Join
                                                            </Button>
                                                        </div>
                                                    </div>
                                                ))}
                                            </div>
                                        ) : (
                                            <div className="flex flex-col items-center justify-center py-6 space-y-2 text-gray-400">
                                                <Users className="h-10 w-10 opacity-20" />
                                                <p className="text-sm">No games available</p>
                                                <p className="text-xs opacity-70">Create a new game or try again later</p>
                                            </div>
                                        )}
                                    </div>

                                    {/* Manual room ID entry */}
                                    <div className="space-y-1.5 pt-2">
                                        <div className="flex items-center justify-between">
                                            <label htmlFor="roomId" className="block text-sm font-medium text-gray-300 flex items-center">
                                                <KeyRound className="h-4 w-4 mr-1.5 text-gray-500" />
                                                Room ID
                                            </label>
                                            <p className="text-xs text-gray-500">Or enter room ID manually</p>
                                        </div>
                                        <Input
                                            id="roomId"
                                            type="text"
                                            value={roomId}
                                            onChange={(e) => setRoomId(e.target.value)}
                                            placeholder="Enter the room ID to join"
                                            icon={<KeyRound className="h-4 w-4" />}
                                            variant="magenta"
                                            className={cn(roomIdError && "border-red-500 focus-visible:ring-red-500")}
                                        />
                                        {roomIdError && (
                                            <p className="mt-1 text-sm text-red-400 flex items-center">
                                                <span className="inline-block h-1.5 w-1.5 rounded-full bg-red-500 mr-1.5"></span>
                                                {roomIdError}
                                            </p>
                                        )}
                                        <p className="text-xs text-gray-500 mt-1 pl-1">Format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx</p>
                                    </div>
                                </TabsContent>

                                {/* Create tab content */}
                                <TabsContent value="create" className="space-y-4 mt-4 mb-0">
                                    {/* Betting amount notice */}
                                    <div className="rounded-md bg-amber-950/20 p-4 text-sm border border-amber-900/30 shadow-inner">
                                        <p className="mb-2 text-amber-400 font-medium flex items-center">Game Stakes</p>
                                        <p className="text-amber-200 text-sm opacity-90">
                                            Each player bets <span className="font-bold text-amber-300">$0.01</span>. Winner takes all!
                                        </p>
                                    </div>

                                    <div className="rounded-md bg-cyan-950/20 p-4 text-sm text-gray-300 border border-cyan-900/30 shadow-inner">
                                        <p className="mb-2 text-cyan-400 font-medium flex items-center">
                                            <GamepadIcon className="h-4 w-4 mr-1.5" />
                                            Host a New Game
                                        </p>
                                        <p className="text-sm opacity-90">You'll create a room and get a Room ID to share with your opponent.</p>
                                    </div>
                                </TabsContent>

                                {/* Error message */}
                                {error && (
                                    <div className="text-sm text-red-400 p-3 bg-red-900/20 border border-red-900/30 rounded-md shadow-inner animate-pulse">
                                        <p className="font-medium mb-1">Error</p>
                                        <p>{error}</p>
                                    </div>
                                )}

                                {/* Submit button */}
                                <Button
                                    type="submit"
                                    disabled={!isConnected}
                                    variant={mode === "create" ? "glowCyan" : "glowMagenta"}
                                    size="xxl"
                                    className="w-full mt-6"
                                    leftIcon={!isConnected ? <Loader2 className="animate-spin" /> : undefined}
                                >
                                    {!isConnected ? "Connecting..." : mode === "create" ? "Create Game" : "Join Game"}
                                </Button>
                            </form>
                        </Tabs>
                    )}
                </CardContent>

                {/* Card footer with tagline and online counter */}
                <CardFooter className="justify-between py-3 border-t border-gray-800/30">
                    <p className="text-xs text-gray-500">Light speed, neon bleed.</p>
                    <div className="flex items-center gap-2">
                        <OnlinePlayersCounter className="ml-auto" count={onlineUsers} />
                    </div>
                </CardFooter>
            </Card>
        </div>
    );
}
