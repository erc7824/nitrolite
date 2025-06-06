import { useState, useEffect } from "react";
import { Button } from "./ui/button";
import { Loader2, AlertCircle, Coins, ArrowRight } from "lucide-react";
import { cn } from "../lib/utils";
// Since we're importing from viem which isn't configured yet, let's declare the types here
// type Address = `0x${string}`;
type Hex = `0x${string}`;
import { useChannel } from "../hooks/useChannel";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "./ui/dialog";

// USDC token address on the testnet
const USDC_ADDRESS: Hex = "0x07865c6E87B9F70255377e024ace6630C1Eaa37F";

interface ChannelRequiredModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSuccess: (action: "join" | "create", roomId?: string) => void;
    mode: "join" | "create";
    roomId?: string;
}

export function ChannelRequiredModal({ isOpen, onClose, onSuccess, mode, roomId }: ChannelRequiredModalProps) {
    const amount = "0.0001"; // Fixed amount
    const [step, setStep] = useState<"info" | "create" | "success">("info");
    const { createChannel, depositToChannel, isLoading, error, isChannelOpen } = useChannel();

    // Check if the modal should stay open when there's no channel
    useEffect(() => {
        // If modal is open, check if channel ID exists
        if (isOpen) {
            const channelId = localStorage.getItem("nitrolite_channel_id");

            // If channel is open in state and channel ID exists, proceed to success
            if (isChannelOpen && channelId) {
                console.log("Channel is already open and ID exists, closing modal without calling onSuccess");
                // Just close the modal without calling onSuccess to prevent double join message
                onClose();
            }
        }
    }, [isChannelOpen, isOpen, onClose]);

    const handleCreateChannel = async () => {
        // Fixed amount, no validation needed

        // No validation needed for fixed amount

        try {
            console.log("Starting channel creation process with fixed amount:", amount);
            // Immediately show the creating screen
            setStep("create");

            try {
                // Step 1: Deposit to channel
                console.log("Starting deposit with amount:", amount);
                await depositToChannel(USDC_ADDRESS, amount);
                console.log("Deposit successful, creating channel...");

                // Step 2: Create channel
                const result = await createChannel(USDC_ADDRESS, amount);
                console.log("Channel creation successful:", result);

                // Only proceed to success if we have a valid result
                if (result && result.channelId) {
                    // Make sure the channel ID exists in localStorage
                    localStorage.setItem("nitrolite_channel_id", result.channelId);
                    console.log("Saved channel ID to localStorage:", result.channelId);

                    // Show success screen
                    setStep("success");

                    // Auto proceed after 1.5 seconds to allow the user to see the success message
                    setTimeout(() => {
                        // Log the room action before proceeding
                        console.log(`Proceeding with ${mode} action after channel creation, roomId:`, roomId);
                        onSuccess(mode, roomId);
                        onClose();
                    }, 1500);
                } else {
                    throw new Error("Channel creation successful but no channel ID returned");
                }
            } catch (innerErr) {
                console.error("Process failed during channel creation:", innerErr);
                throw innerErr; // Re-throw to be caught by the outer catch
            }
        } catch (err) {
            console.error("Channel creation failed:", err);
            // Error state is handled by the useChannel hook
            // Stay in the "create" step for 1.5 seconds so user sees the error before going back to info
            setTimeout(() => {
                setStep("info");
            }, 1500);
        }
    };

    // No handleAmountChange needed for fixed amount

    return (
        <Dialog open={isOpen} onOpenChange={(open) => {
            if (!open) {
                onClose();
            }
        }}>
            <DialogContent className="max-w-md w-full border-gray-700 shadow-2xl relative overflow-hidden">
                {/* Background gradient */}
                <div className={cn("absolute inset-0 bg-gradient-to-b", "from-cyan-900/30 to-gray-900/90", "z-0")}></div>

                {/* Particle effects */}
                <div className="absolute inset-0 overflow-hidden">
                    <div className="absolute w-full h-[200%] top-[-50%] left-0 bg-[radial-gradient(circle,_white_1px,_transparent_1px)] bg-[length:20px_20px] opacity-[0.03] animate-sparkle"></div>
                </div>

                {/* Content */}
                <div className="relative z-10">
                    <DialogHeader>
                        <DialogTitle className="text-2xl font-bold text-viper-green mb-2">
                            {step === "info" && "Channel Required"}
                            {step === "create" && "Creating Channel..."}
                            {step === "success" && "Channel Created!"}
                        </DialogTitle>
                        <DialogDescription>
                            {step === "info" && "To play Viper Duel, you need to create a payment channel by depositing USDC."}
                            {step === "create" && "Please confirm the transaction in your wallet."}
                            {step === "success" && "Your payment channel has been created successfully!"}
                        </DialogDescription>
                    </DialogHeader>

                    <div className="py-4">
                        {step === "info" && (
                            <div className="space-y-4">
                                <div className="bg-black/20 rounded-lg p-4 text-sm text-gray-300 border border-viper-green/30">
                                    <p className="mb-2 flex items-center text-viper-green">
                                        <Coins className="h-4 w-4 mr-2" />
                                        <span>Why deposit USDC?</span>
                                    </p>
                                    <p className="mb-2">
                                        Viper Duel uses payment channels to enable instant, secure gameplay without gas fees for each move.
                                    </p>
                                    <p>Your funds remain fully under your control and can be withdrawn at any time.</p>
                                </div>

                                <div className="space-y-2">
                                    <label htmlFor="amount" className="block text-sm font-medium text-gray-300">
                                        Deposit Amount (USDC)
                                    </label>
                                    <div
                                        className="w-full flex items-center px-3 py-2 bg-gray-900 border border-cyan-900/50 rounded-md cursor-not-allowed h-10"
                                        id="amount"
                                    >
                                        <span className="text-cyan-400 mr-1">$</span>
                                        <span className="text-gray-400">0.0001</span>
                                    </div>
                                    <p className="text-xs text-gray-500 mt-1">Fixed deposit amount: 0.0001 USDC</p>
                                </div>

                                {error && (
                                    <div className="text-sm text-red-400 p-3 bg-red-900/20 border border-red-900/30 rounded-md flex items-start space-x-2">
                                        <AlertCircle className="h-5 w-5 flex-shrink-0 mt-0.5" />
                                        <div>
                                            <p className="font-medium">Error</p>
                                            <p>{error}</p>
                                        </div>
                                    </div>
                                )}
                            </div>
                        )}

                        {step === "create" && (
                            <div className="flex flex-col items-center justify-center py-8 space-y-4">
                                <Loader2 className="h-12 w-12 text-cyan-400 animate-spin" />
                                <p className="text-gray-300">Creating your channel...</p>
                                <p className="text-sm text-gray-500">Check your wallet for transaction confirmation</p>
                            </div>
                        )}

                        {step === "success" && (
                            <div className="flex flex-col items-center justify-center py-8 space-y-4">
                                <div className="h-16 w-16 bg-green-500/20 rounded-full flex items-center justify-center">
                                    <svg className="h-10 w-10 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                                    </svg>
                                </div>
                                <p className="text-lg font-medium text-green-400">Deposit Successful!</p>
                                <p className="text-sm text-gray-400 text-center">
                                    Your payment channel is now active.
                                    <br />
                                    You'll be redirected to the game momentarily.
                                </p>
                            </div>
                        )}
                    </div>

                    <DialogFooter>
                        {step === "info" && (
                            <>
                                <Button 
                                    onClick={() => {
                                        console.log("Cancel button clicked");
                                        onClose();
                                    }} 
                                    variant="outline" 
                                    className="w-full sm:w-auto"
                                >
                                    Cancel
                                </Button>
                                <Button onClick={handleCreateChannel} variant="viperGreen" className="w-full sm:w-auto" disabled={isLoading}>
                                    {isLoading ? (
                                        <>
                                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                            Processing...
                                        </>
                                    ) : (
                                        <>
                                            Create Channel
                                            <ArrowRight className="ml-2 h-4 w-4" />
                                        </>
                                    )}
                                </Button>
                            </>
                        )}

                        {step === "create" && (
                            <Button variant="outline" className="w-full sm:w-auto" disabled>
                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                Waiting for confirmation...
                            </Button>
                        )}

                        {step === "success" && (
                            <Button
                                onClick={() => {
                                    onSuccess(mode, roomId);
                                    onClose();
                                }}
                                variant="viperGreen"
                                className="w-full sm:w-auto"
                            >
                                Continue to Game
                                <ArrowRight className="ml-2 h-4 w-4" />
                            </Button>
                        )}
                    </DialogFooter>
                </div>
            </DialogContent>
        </Dialog>
    );
}
