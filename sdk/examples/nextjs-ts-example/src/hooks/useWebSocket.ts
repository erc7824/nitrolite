import { useState, useEffect, useRef, useCallback, useMemo } from 'react';
import { 
  createWebSocketClient,
  createEthersSigner, 
  generateKeyPair, 
  WalletSigner, 
  CryptoKeypair,
  getAddressFromPublicKey
} from '@/websocket';
import { Message, Channel, WSStatus } from '@/types';
import { MessageType } from '@/hooks/useMessageStyles';

export function useWebSocket(url: string) {
  const [status, setStatus] = useState<WSStatus>("disconnected");
  const [messages, setMessages] = useState<Message[]>([]);
  const [keyPair, setKeyPair] = useState<CryptoKeypair | null>(() => {
    // Try to load keys from localStorage on initial render
    if (typeof window !== 'undefined') {
      const savedKeys = localStorage.getItem('crypto_keypair');
      if (savedKeys) {
        try {
          const parsed = JSON.parse(savedKeys) as CryptoKeypair;
          
          // If missing address property, derive it from the public key
          if (parsed.publicKey && !parsed.address) {
            try {
              parsed.address = getAddressFromPublicKey(parsed.publicKey);
              // Update localStorage with the derived address
              localStorage.setItem('crypto_keypair', JSON.stringify(parsed));
            } catch (addrError) {
              console.error('Failed to derive address from public key:', addrError);
            }
          }
          
          return parsed;
        } catch (e) {
          console.error('Failed to parse saved keys:', e);
        }
      }
    }
    return null;
  });
  const [currentSigner, setCurrentSigner] = useState<WalletSigner | null>(null);

  // Create addMessage function before using it
  const addMessage = useCallback((text: string, type: MessageType = "info", sender?: string) => {
    setMessages(prev => [...prev, { 
      text, 
      type, 
      sender,
      timestamp: Date.now()
    }]);
  }, []);
  
  // Effect to initialize signer from existing keys if available
  useEffect(() => {
    if (keyPair?.privateKey && !currentSigner) {
      // Initialize signer from saved keys if available
      try {
        setCurrentSigner(createEthersSigner(keyPair.privateKey));
        addMessage("Loaded existing keys", "system");
        // Don't generate keys automatically anymore - wait for explicit channel opening
      } catch (e) {
        console.error('Failed to create signer from saved keys:', e);
      }
    }
  }, [keyPair, currentSigner, addMessage]);
  const [currentChannel, setCurrentChannel] = useState<Channel | null>(null);
  
  // Create WebSocket client with current signer
  const client = useMemo(() => {
    if (!currentSigner) return null;
    return createWebSocketClient(
      url, 
      currentSigner,
      { 
        autoReconnect: true, 
        reconnectDelay: 1000, 
        maxReconnectAttempts: 5,
        requestTimeout: 10000
      }
    );
  }, [url, currentSigner]);
  
  const clientRef = useRef<any>(null);
  
  // Update the client reference when the client changes
  useEffect(() => {
    clientRef.current = client;
  }, [client]);
  
  // Initialize WebSocket event listeners
  useEffect(() => {
    const client = clientRef.current;
    if (!client) return;
    
    client.onStatusChange((newStatus) => {
      setStatus(newStatus);
      addMessage(`Connection status: ${newStatus}`, "system");
    });
    
    client.onMessage((message) => {
      console.log("Received message:", message);
      
      // Handle different message types
      if (message.type === "message" && message.data) {
        const messageData = message.data;
        if (messageData.message && messageData.sender) {
          addMessage(`${messageData.message}`, "received", messageData.sender);
        } else {
          addMessage(`Received: ${JSON.stringify(messageData)}`, "received");
        }
      } else if (message.type === "pong" && message.data) {
        addMessage(`Server responded with pong (${message.data.timestamp || 'no timestamp'})`, "received");
      } else if (message.type === "rpc_response" && message.data) {
        if (message.data.method === "get_balance") {
          addMessage(`Balance: ${message.data.result || "0"} tokens`, "success");
        }
      }
    });
    
    client.onError(error => addMessage(`Error: ${error.message}`, "error"));
    
    addMessage("Welcome to Broker WebSocket Client", "system");
    
    // Initialize with saved keys if available
    if (keyPair) {
      addMessage("Keys loaded from browser storage", "success");
      addMessage(`Ethereum Address: ${keyPair.address || 'unknown'}`, "info");
    } else {
      addMessage("Generate a key pair to begin", "system");
    }
    
    return () => client.close();
  }, [addMessage, keyPair]);
  
  // Removed auto-connect functionality as we're now connecting explicitly after channel opening
  
  // Generate a new key pair
  const generateKeys = useCallback(async () => {
    try {
      const newKeyPair = await generateKeyPair();
      setKeyPair(newKeyPair);
      
      // Save to localStorage
      if (typeof window !== 'undefined') {
        localStorage.setItem('crypto_keypair', JSON.stringify(newKeyPair));
      }
      
      // Create a new signer with the generated private key
      const newSigner = createEthersSigner(newKeyPair.privateKey);
      setCurrentSigner(newSigner);
      
      addMessage("Generated new key pair", "success");
      addMessage(`Private key: ${newKeyPair.privateKey.substring(0, 10)}...`, "info");
      addMessage(`Public key: ${newKeyPair.publicKey.substring(0, 10)}...`, "info");
      addMessage("Keys saved to browser storage", "success");
      
      return newKeyPair;
    } catch (error) {
      addMessage(`Error generating keys: ${error instanceof Error ? error.message : String(error)}`, "error");
      return null;
    }
  }, [addMessage]);
  
  // Connect to WebSocket
  const connect = useCallback(async () => {
    if (!keyPair) {
      addMessage("Please generate a key pair first", "error");
      throw new Error("No key pair available for connection");
    }
    
    try {
      addMessage("Connecting to broker...", "system");
      await clientRef.current.connect();
      addMessage("Connected to broker successfully", "success");
      return true;
    } catch (error) {
      console.error("Connection error:", error);
      addMessage(`Connection error: ${error instanceof Error ? error.message : String(error)}`, "error");
      throw error;
    }
  }, [keyPair, addMessage]);
  
  // Disconnect from WebSocket
  const disconnect = useCallback(() => {
    clientRef.current.close();
  }, []);
  
  // Subscribe to a channel
  const subscribeToChannel = useCallback(async (channel: Channel) => {
    if (!clientRef.current.isConnected) {
      return addMessage("Not connected to server", "error");
    }
    
    try {
      addMessage(`Subscribing to channel: ${channel}`, "info");
      await clientRef.current.subscribe(channel);
      setCurrentChannel(channel);
      addMessage(`Subscribed to channel: ${channel}`, "success");
    } catch (error) {
      addMessage(`Subscribe error: ${error instanceof Error ? error.message : String(error)}`, "error");
    }
  }, [addMessage]);
  
  // Send a message to the current channel
  const sendMessage = useCallback(async (message: string) => {
    if (!clientRef.current.isConnected) {
      return addMessage("Not connected to server", "error");
    }
    
    if (!clientRef.current.currentSubscribedChannel) {
      return addMessage("Please subscribe to a channel first", "error");
    }
    
    try {
      await clientRef.current.publishMessage(message);
      
      // Get display name with shortened public key
      const shortenedKey = clientRef.current.getShortenedPublicKey();
      addMessage(message, "sent", `You (${shortenedKey})`);
    } catch (error) {
      addMessage(`Send error: ${error instanceof Error ? error.message : String(error)}`, "error");
    }
  }, [addMessage]);
  
  // Send a ping request
  const sendPing = useCallback(async () => {
    if (!clientRef.current.isConnected) {
      return addMessage("Not connected to server", "error");
    }
    
    try {
      addMessage("Sending ping request", "info");
      await clientRef.current.ping();
    } catch (error) {
      addMessage(`Ping error: ${error instanceof Error ? error.message : String(error)}`, "error");
    }
  }, [addMessage]);
  
  // Check balance
  const checkBalance = useCallback(async (tokenAddress: string = "0xSHIB...") => {
    if (!clientRef.current.isConnected) {
      return addMessage("Not connected to server", "error");
    }
    
    try {
      addMessage(`Requesting balance information for ${tokenAddress}`, "info");
      await clientRef.current.checkBalance(tokenAddress);
    } catch (error) {
      addMessage(`Balance check error: ${error instanceof Error ? error.message : String(error)}`, "error");
    }
  }, [addMessage]);
  
  // Send a generic RPC request
  const sendRequest = useCallback(async (methodName: string, methodParams: string) => {
    if (!clientRef.current.isConnected) {
      return addMessage("Not connected to server", "error");
    }
    
    try {
      let params: any[] = [];
      if (methodParams.trim()) {
        try {
          params = JSON.parse(methodParams);
          if (!Array.isArray(params)) params = [params];
        } catch (e) {
          return addMessage(`Error parsing params: ${e instanceof Error ? e.message : String(e)}`, "error");
        }
      }
      
      addMessage(`Sending ${methodName} request with params: ${JSON.stringify(params)}`, "info");
      const response = await clientRef.current.sendRequest(methodName, params);
      addMessage(`Received response: ${JSON.stringify(response)}`, "success");
      return response;
    } catch (error) {
      addMessage(`Request error: ${error instanceof Error ? error.message : String(error)}`, "error");
    }
  }, [addMessage]);
  
  const clearMessages = useCallback(() => setMessages([]), []);
  
  // Function to clear saved keys
  const clearKeys = useCallback(() => {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('crypto_keypair');
    }
    setKeyPair(null);
    setCurrentSigner(null);
    addMessage("Keys cleared from browser storage", "system");
    
    // Don't auto-generate after clearing - let the channel opening flow handle it
    addMessage("Keys will be generated when you open a channel", "info");
  }, [addMessage]);
  
  return {
    // State
    status,
    messages,
    keyPair,
    currentChannel,
    
    // Computed values
    isConnected: clientRef.current?.isConnected || false,
    hasKeys: !!keyPair,
    
    // Actions
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
  };
}