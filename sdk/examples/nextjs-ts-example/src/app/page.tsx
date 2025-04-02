"use client";

import { ConnectionStatus } from "@/components/ConnectionStatus";
import { RequestForm } from "@/components/RequestForm";
import { MessageList } from "@/components/MessageList";
import { About } from "@/components/About";
import { useWebSocket } from "@/hooks/useWebSocket";

export default function Home() {
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
        clearKeys
    } = useWebSocket("ws://localhost:8000/ws");

    return (
        <div className="min-h-screen bg-gradient-to-br from-gray-900 to-gray-800 text-white p-6">
            <div className="max-w-6xl mx-auto">
                <header className="mb-8">
                    <div className="flex items-center justify-between">
                        <div>
                            <h1 className="text-3xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-primary-400 to-secondary-500">Broker WebSocket</h1>
                            <p className="text-gray-400">Secure communication with cryptographic authentication</p>
                        </div>
                    </div>
                </header>

                <ConnectionStatus 
                    status={status} 
                    keyPair={keyPair}
                    onGenerateKeys={generateKeys}
                    onConnect={connect} 
                    onDisconnect={disconnect}
                    onClearKeys={clearKeys} 
                />

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

                <About />
            </div>
        </div>
    );
}
