"use client";

import { useState, useEffect } from "react";
import { useSnapshot } from "valtio";
import { ConnectionStatus } from "@/components/ConnectionStatus";
import { RequestForm } from "@/components/RequestForm";
import { MessageList } from "@/components/MessageList";
import { About } from "@/components/About";
import { useWebSocket } from "@/hooks/useWebSocket";
import MetaMaskConnect from "@/components/MetaMaskConnect";
import WalletStore from "@/store/WalletStore";
import { fetchAssets } from "@/store/AssetsStore";
import { Address } from "viem";

export default function Home() {
    const [showMetaMaskFlow, setShowMetaMaskFlow] = useState(false);
    const walletSnap = useSnapshot(WalletStore.state);

    const {
        status,
        messages,
        keyPair,
        currentChannel,
        isConnected,
        generateKeys,
        connect,
        disconnect,
        subscribeToChannel,
        sendMessage,
        sendPing,
        checkBalance,
        sendRequest,
        clearMessages,
        clearKeys,
    } = useWebSocket("ws://localhost:8000/ws");

    // Load assets when component mounts
    useEffect(() => {
        fetchAssets();
    }, []);

    // Function to handle channel opening
    const handleOpenChannel = async (tokenAddress: string, amount: string) => {
        // Here you would integrate with the channel opening logic from your SDK
        console.log(`Opening channel with token ${tokenAddress} and amount ${amount}`);

        // Update wallet store
        WalletStore.openChannel(tokenAddress as Address, amount);

        // Automatically generate keys if not present
        if (!keyPair) {
            await generateKeys();
        }

        // Automatically connect to broker websocket
        setTimeout(() => {
            if (status === "disconnected" && keyPair) {
                connect();
            }
        }, 500);

        // Hide MetaMask flow and show the main app
        setShowMetaMaskFlow(false);
    };

    return (
        <div className="min-h-screen bg-gradient-to-br from-white to-gray-100 text-gray-800 p-6">
            <div className="max-w-6xl mx-auto">
                <header className="mb-8">
                    <div className="flex items-center justify-between">
                        <div>
                            <h1 className="text-3xl font-bold text-[#3531ff]">Nitrolite</h1>
                        </div>
                        {walletSnap.connected && (
                            <div className="flex items-center space-x-2">
                                <span className="text-sm bg-white border border-gray-200 py-1 px-2 rounded font-mono text-gray-700 shadow-sm">
                                    {walletSnap.account?.substring(0, 6)}...{walletSnap.account?.substring(38)}
                                </span>
                                <button
                                    onClick={async () => {
                                        const { disconnect } = await import("@/hooks/useMetaMask");
                                        disconnect();
                                    }}
                                    className="bg-red-600 hover:bg-red-700 text-white text-sm py-1 px-2 rounded transition-colors cursor-pointer"
                                >
                                    Disconnect
                                </button>
                            </div>
                        )}
                    </div>
                </header>

                {walletSnap.channelOpen ? (
                    <>
                        <div className="flex gap-4 mb-2 flex-col md:flex-row">
                            <div className="bg-white p-3 rounded-lg border border-[#3531ff]/30 shadow-sm flex-1">
                                <div className="flex items-center justify-between">
                                    <div className="flex items-center">
                                        <span className="text-md font-semibold text-gray-800 mr-2">Channel Status</span>
                                        <span className="px-2 py-0.5 bg-[#3531ff]/20 text-[#3531ff] text-xs rounded">Active</span>
                                    </div>
                                    <div className="flex items-center space-x-3">
                                        <div className="flex items-center">
                                            <div
                                                className={`w-2 h-2 rounded-full mr-1 ${status === "connected" ? "bg-green-500" : "bg-yellow-500"}`}
                                            ></div>
                                            <span className="text-xs text-gray-600">{status === "connected" ? "Channel Active" : "Initializing"}</span>
                                        </div>
                                        <div className="text-xs text-gray-600 font-mono">
                                            <span className="px-2 py-0.5 bg-gray-100 rounded-sm">
                                                {walletSnap.selectedTokenAddress?.substring(0, 6)}...{walletSnap.selectedTokenAddress?.substring(38)}
                                            </span>
                                        </div>
                                        <div className="text-xs text-gray-600">
                                            Amount: <span className="font-mono text-gray-800">{walletSnap.selectedAmount}</span>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div className="bg-white p-3 rounded-lg border border-gray-200 shadow-sm">
                                <div className="flex items-center justify-between mb-2">
                                    <span className="text-xs font-medium text-gray-700">Authentication Keys</span>
                                    <div className="flex items-center">
                                        <div className={`w-2 h-2 rounded-full mr-1 ${status === "connected" ? "bg-green-500" : "bg-red-500"}`}></div>
                                        <span className="text-xs text-gray-600">{status === "connected" ? "Connected to Broker" : "Connecting..."}</span>
                                    </div>
                                </div>

                                {keyPair && (
                                    <div className="grid grid-cols-1 gap-1 text-xs">
                                        <div className="flex items-center">
                                            <span className="text-gray-500 w-20">Address:</span>
                                            <span className="font-mono bg-gray-100 px-1 py-0.5 rounded-sm text-gray-800 overflow-hidden text-ellipsis flex-1">
                                                {keyPair.address}
                                            </span>
                                            <button
                                                className="ml-1 px-1 py-0.5 text-[#3531ff] hover:bg-[#3531ff]/10 rounded cursor-pointer transition-colors"
                                                onClick={() => {
                                                    navigator.clipboard.writeText(keyPair.address || "");
                                                }}
                                                title="Copy address"
                                            >
                                                <svg
                                                    xmlns="http://www.w3.org/2000/svg"
                                                    className="h-3 w-3"
                                                    fill="none"
                                                    viewBox="0 0 24 24"
                                                    stroke="currentColor"
                                                >
                                                    <path
                                                        strokeLinecap="round"
                                                        strokeLinejoin="round"
                                                        strokeWidth="2"
                                                        d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3"
                                                    />
                                                </svg>
                                            </button>
                                        </div>
                                        <div className="flex items-center">
                                            <span className="text-gray-500 w-20">Public Key:</span>
                                            <span className="font-mono bg-gray-100 px-1 py-0.5 rounded-sm text-gray-800 overflow-hidden text-ellipsis flex-1">{`${keyPair.publicKey.substring(
                                                0,
                                                16
                                            )}...${keyPair.publicKey.substring(keyPair.publicKey.length - 16)}`}</span>
                                            <button
                                                className="ml-1 px-1 py-0.5 text-[#3531ff] hover:bg-[#3531ff]/10 rounded cursor-pointer transition-colors"
                                                onClick={() => {
                                                    navigator.clipboard.writeText(keyPair.publicKey);
                                                }}
                                                title="Copy public key"
                                            >
                                                <svg
                                                    xmlns="http://www.w3.org/2000/svg"
                                                    className="h-3 w-3"
                                                    fill="none"
                                                    viewBox="0 0 24 24"
                                                    stroke="currentColor"
                                                >
                                                    <path
                                                        strokeLinecap="round"
                                                        strokeLinejoin="round"
                                                        strokeWidth="2"
                                                        d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3"
                                                    />
                                                </svg>
                                            </button>
                                        </div>
                                        <div className="flex items-center">
                                            <span className="text-gray-500 w-20">Private Key:</span>
                                            <span className="font-mono bg-gray-100 px-1 py-0.5 rounded-sm text-gray-800 overflow-hidden text-ellipsis flex-1">{`${keyPair.privateKey.substring(
                                                0,
                                                16
                                            )}...${keyPair.privateKey.substring(keyPair.privateKey.length - 16)}`}</span>
                                            <button
                                                className="ml-1 px-1 py-0.5 text-[#3531ff] hover:bg-[#3531ff]/10 rounded cursor-pointer transition-colors"
                                                onClick={() => {
                                                    navigator.clipboard.writeText(keyPair.privateKey);
                                                }}
                                                title="Copy private key"
                                            >
                                                <svg
                                                    xmlns="http://www.w3.org/2000/svg"
                                                    className="h-3 w-3"
                                                    fill="none"
                                                    viewBox="0 0 24 24"
                                                    stroke="currentColor"
                                                >
                                                    <path
                                                        strokeLinecap="round"
                                                        strokeLinejoin="round"
                                                        strokeWidth="2"
                                                        d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3"
                                                    />
                                                </svg>
                                            </button>
                                        </div>
                                    </div>
                                )}
                            </div>
                        </div>

                        <MessageList messages={messages} onClear={clearMessages} />

                        <RequestForm
                            isConnected={isConnected}
                            currentChannel={currentChannel}
                            onSendRequest={sendRequest}
                            onSendMessage={sendMessage}
                            onSubscribeToChannel={subscribeToChannel}
                            onSendPing={sendPing}
                            onCheckBalance={checkBalance}
                        />
                    </>
                ) : showMetaMaskFlow ? (
                    <MetaMaskConnect onChannelOpen={handleOpenChannel} />
                ) : (
                    <div className="flex flex-col items-center justify-center py-12">
                        <div className="text-center mb-8">
                            <h2 className="text-2xl font-bold mb-4">Welcome to Nitrolite</h2>
                            <p className="text-gray-600 max-w-md mx-auto">
                                Connect your MetaMask wallet and open a channel to start using Nitrolite's secure services.
                            </p>
                        </div>
                        <button
                            onClick={() => setShowMetaMaskFlow(true)}
                            className="bg-[#3531ff] hover:bg-[#2b28cc] text-white font-bold py-3 px-6 rounded-lg transition-colors shadow-lg shadow-[#3531ff]/20 cursor-pointer"
                        >
                            Connect MetaMask & Open Channel
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
}
