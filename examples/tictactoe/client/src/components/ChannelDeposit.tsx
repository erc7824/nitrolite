import { useState } from "react";
import { useChannel } from "../hooks/useChannel";
import { WalletStore } from "../store";
import { useStore } from "../store/storeUtils";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "./ui/card";
import { Button } from "./ui/button";
import { cn } from "../lib/utils";

// Default USDC token address on Polygon
const USDC_ADDRESS = "0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359";

/**
 * Component for depositing funds and creating a channel
 */
export function ChannelDeposit() {
    const amount = "0.0001"; // Fixed amount
    const [isProcessing, setIsProcessing] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [successMessage, setSuccessMessage] = useState<string | null>(null);

    const { createChannel, depositToChannel, isChannelOpen } = useChannel();
    const wallet = useStore(WalletStore.state);

    // Fixed amount, no change handler needed

    const handleCreateChannel = async () => {
        if (!amount || parseFloat(amount) <= 0) {
            setError("Please enter a valid amount");
            return;
        }

        setIsProcessing(true);
        setError(null);
        setSuccessMessage(null);

        try {
            // Then deposit to it
            await depositToChannel(USDC_ADDRESS, amount);

            // First create a channel
            await createChannel(USDC_ADDRESS, amount);

            setSuccessMessage(`Successfully created channel and deposited ${amount} USDC`);
        } catch (err) {
            console.error("Channel creation/deposit error:", err);
            setError(err instanceof Error ? err.message : String(err));
        } finally {
            setIsProcessing(false);
        }
    };

    if (isChannelOpen) {
        return (
            <Card className="mb-6">
                <CardHeader className="pb-2">
                    <CardTitle className="text-xl text-cyan-400">Channel Active</CardTitle>
                    <CardDescription>
                        You have an active channel with {wallet.channelAmount ? Number(wallet.channelAmount) / 1000000 : "0"} USDC deposited.
                    </CardDescription>
                </CardHeader>
                <CardFooter className="justify-end">
                    <Button
                        className="bg-gradient-to-r from-cyan-500 to-blue-600 hover:opacity-90"
                        onClick={() => {
                            // In a real app you might want a separate component for this
                            // This is just a placeholder
                            alert("Close channel functionality would go here");
                        }}
                    >
                        Close Channel
                    </Button>
                </CardFooter>
            </Card>
        );
    }

    return (
        <Card className="mb-6">
            <CardHeader className="pb-2">
                <CardTitle className="text-xl text-cyan-400">Create Channel</CardTitle>
                <CardDescription>Create a new channel by depositing USDC. This is required to play games.</CardDescription>
            </CardHeader>
            <CardContent>
                {error && <div className="mb-4 p-3 bg-red-900/30 border border-red-800 text-red-200 rounded-md text-sm">{error}</div>}

                {successMessage && (
                    <div className="mb-4 p-3 bg-green-900/30 border border-green-800 text-green-200 rounded-md text-sm">{successMessage}</div>
                )}

                <div className="mb-4">
                    <label htmlFor="amount" className="block text-sm font-medium text-gray-300 mb-1">
                        USDC Amount
                    </label>
                    <div className="relative">
                        <div 
                            className="w-full flex items-center px-3 py-2 bg-gray-900 border border-gray-700 rounded-md cursor-not-allowed h-10" 
                            aria-describedby="amount-currency"
                        >
                            <span className="text-gray-400">0.0001</span>
                            <div className="absolute inset-y-0 right-0 flex items-center pr-3">
                                <span className="text-gray-400 sm:text-sm" id="amount-currency">
                                    USDC
                                </span>
                            </div>
                        </div>
                    </div>
                    <p className="mt-1 text-xs text-gray-400">Fixed deposit amount: 0.0001 USDC</p>
                </div>
            </CardContent>
            <CardFooter className="justify-end">
                <Button
                    className={cn(
                        "bg-gradient-to-r from-cyan-500 to-blue-600 hover:opacity-90",
                        isProcessing &&
                            "bg-gray-700 text-gray-400 opacity-50 cursor-not-allowed hover:opacity-50"
                    )}
                    onClick={handleCreateChannel}
                    disabled={isProcessing}
                >
                    {isProcessing ? "Processing..." : "Create & Deposit"}
                </Button>
            </CardFooter>
        </Card>
    );
}
