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
    // No longer need the showMetaMaskFlow state as we show it directly
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

        // Generate keys and connect to websocket in a sequential flow
        try {
            // Step 1: Generate keys if not present
            let currentKeyPair = keyPair;
            if (!currentKeyPair) {
                currentKeyPair = await generateKeys();
                if (!currentKeyPair) {
                    throw new Error("Failed to generate keys");
                }
            }
            
            // Step 2: Connect to the broker websocket only after we have keys
            if (status === "disconnected" && currentKeyPair) {
                await connect();
            }
            
            // Step 3: Now that we're connected, we could auto-subscribe to a channel if needed
            // This is optional and depends on your application flow
            // if (status === "connected") {
            //    await subscribeToChannel("your-channel-id");
            // }
        } catch (error) {
            console.error("Error in channel opening sequence:", error);
            // You might want to update the UI to show the error
        }
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
                                        // First disconnect from WebSocket if connected
                                        if (status === "connected") {
                                            disconnect(); // This is the WebSocket disconnect
                                        }
                                        // Then disconnect from MetaMask
                                        const { disconnectWallet } = await import("@/hooks/useMetaMask");
                                        await disconnectWallet();
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
                        <div className="flex gap-3 mb-2 flex-col md:flex-row">
                            <div className="bg-white p-3 rounded-lg border border-[#3531ff]/30 shadow-sm flex-1">
                                <div className="flex items-center justify-between">
                                    <div className="flex items-center">
                                        <span className="text-md font-semibold text-gray-800 mr-2">Channel Status</span>
                                        <span className="px-2 py-0.5 bg-[#3531ff]/20 text-[#3531ff] text-xs rounded">Active</span>
                                    </div>
                                    <div className="flex items-center space-x-3">
                                        <div className="flex items-center">
                                            <div
                                                className={`w-2 h-2 rounded-full mr-1 ${
                                                  status === "connected" ? "bg-green-500" : 
                                                  status === "connecting" ? "bg-yellow-500" : 
                                                  "bg-red-500"
                                                }`}
                                            ></div>
                                            <span className="text-xs text-gray-600">
                                              {status === "connected" ? "Channel Active" : 
                                               status === "connecting" ? "Connecting..." : 
                                               "Disconnected"}
                                            </span>
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
                                        <div className={`w-2 h-2 rounded-full mr-1 ${
                                          status === "connected" ? "bg-green-500" : 
                                          status === "connecting" ? "bg-yellow-500" : 
                                          "bg-red-500"
                                        }`}></div>
                                        <span className="text-xs text-gray-600">
                                          {status === "connected" ? "Connected to Broker" : 
                                           status === "connecting" ? "Connecting..." : 
                                           "Disconnected"}
                                        </span>
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
                        
                        <div className="mt-4 flex flex-col md:flex-row space-y-4 md:space-y-0 md:space-x-4">
                            <div className="bg-white p-4 rounded-lg border border-gray-200 shadow-sm w-full md:w-1/2">
                                <h2 className="text-lg font-semibold mb-3 text-[#3531ff]">About Nitrolite</h2>
                                <p className="text-gray-600 text-sm mb-3">
                                    Nitrolite provides secure state channels with cryptographic authentication for fast, low-cost transactions without on-chain delays.
                                </p>
                                <div className="mt-4">
                                    <a href="https://erc7824.org/" target="_blank" rel="noopener noreferrer" className="text-[#3531ff] text-sm hover:underline cursor-pointer">
                                        Learn more about Nitrolite
                                    </a>
                                </div>
                            </div>
                            
                            <div className="bg-white p-4 rounded-lg border border-gray-200 shadow-sm w-full md:w-1/2">
                                <h2 className="text-lg font-semibold mb-3 text-[#3531ff]">FAQ</h2>
                                <div className="space-y-2">
                                    <details className="group">
                                        <summary className="flex justify-between items-center font-medium cursor-pointer text-sm text-gray-700">
                                            <span>What are state channels?</span>
                                            <span className="transition group-open:rotate-180">
                                                <svg fill="none" height="12" width="12" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24"><path d="M6 9l6 6 6-6"></path></svg>
                                            </span>
                                        </summary>
                                        <p className="text-xs text-gray-600 mt-1 group-open:animate-fadeIn">
                                            State channels allow for off-chain transactions that are later settled on-chain, reducing gas costs and increasing speed.
                                        </p>
                                    </details>
                                    
                                    <details className="group">
                                        <summary className="flex justify-between items-center font-medium cursor-pointer text-sm text-gray-700">
                                            <span>How secure are these channels?</span>
                                            <span className="transition group-open:rotate-180">
                                                <svg fill="none" height="12" width="12" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24"><path d="M6 9l6 6 6-6"></path></svg>
                                            </span>
                                        </summary>
                                        <p className="text-xs text-gray-600 mt-1 group-open:animate-fadeIn">
                                            All transactions are cryptographically signed and verified, ensuring the same security guarantees as on-chain transactions.
                                        </p>
                                    </details>
                                    
                                    <details className="group">
                                        <summary className="flex justify-between items-center font-medium cursor-pointer text-sm text-gray-700">
                                            <span>Can I close my channel?</span>
                                            <span className="transition group-open:rotate-180">
                                                <svg fill="none" height="12" width="12" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24"><path d="M6 9l6 6 6-6"></path></svg>
                                            </span>
                                        </summary>
                                        <p className="text-xs text-gray-600 mt-1 group-open:animate-fadeIn">
                                            Yes, you can close your channel at any time, which will settle the final state on-chain.
                                        </p>
                                    </details>
                                    
                                    <details className="group">
                                        <summary className="flex justify-between items-center font-medium cursor-pointer text-sm text-gray-700">
                                            <span>What tokens are supported?</span>
                                            <span className="transition group-open:rotate-180">
                                                <svg fill="none" height="12" width="12" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24"><path d="M6 9l6 6 6-6"></path></svg>
                                            </span>
                                        </summary>
                                        <p className="text-xs text-gray-600 mt-1 group-open:animate-fadeIn">
                                            Nitrolite supports a wide range of ERC-20 tokens across multiple blockchain networks.
                                        </p>
                                    </details>
                                </div>
                            </div>
                        </div>
                    </>
                ) : (
                    <MetaMaskConnect onChannelOpen={handleOpenChannel} />
                )}
            </div>
        </div>
    );
}
