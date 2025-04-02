import { useState, useEffect, useRef, useCallback, useMemo } from 'react';
import { 
  createWebSocketClient,
  createEthersSigner, 
  generateKeyPair, 
  WalletSigner, 
  CryptoKeypair,
  getAddressFromPublicKey
} from '@/utils/wsClient';
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

  // Effect to generate keys if none exist on component mount
  useEffect(() => {
    if (!keyPair && typeof window !== 'undefined') {
      // Generate keys on component mount if none exist
      generateKeyPair().then(newKeyPair => {
        if (newKeyPair) {
          setKeyPair(newKeyPair);
          localStorage.setItem('crypto_keypair', JSON.stringify(newKeyPair));
          try {
            setCurrentSigner(createEthersSigner(newKeyPair.privateKey));
            addMessage("Generated new key pair", "success");
            addMessage(`Ethereum Address: ${newKeyPair.address || 'unknown'}`, "info");
          } catch (e) {
            console.error('Failed to create signer from new keys:', e);
          }
        }
      });
    } else if (keyPair?.privateKey) {
      // Initialize signer from saved keys if available
      try {
        setCurrentSigner(createEthersSigner(keyPair.privateKey));
      } catch (e) {
        console.error('Failed to create signer from saved keys:', e);
      }
    }
  }, []);
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
  
  const addMessage = useCallback((text: string, type: MessageType = "info", sender?: string) => {
    setMessages(prev => [...prev, { 
      text, 
      type, 
      sender,
      timestamp: Date.now()
    }]);
  }, []);
  
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
  
  // This will store the auto-connect intent
  const [shouldAutoConnect, setShouldAutoConnect] = useState(false);
  
  // Check if we should auto-connect
  useEffect(() => {
    // Only mark for auto-connect once on startup if we have keys
    if (keyPair && status === "disconnected" && !shouldAutoConnect) {
      setShouldAutoConnect(true);
    }
  }, [keyPair, status, shouldAutoConnect]);
  
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
      addMessage(`Private key: ${newKeyPair.privateKey}`, "info");
      addMessage(`Public key: ${newKeyPair.publicKey}`, "info");
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
      return addMessage("Please generate a key pair first", "error");
    }
    
    try {
      await clientRef.current.connect();
    } catch (error) {
      console.error("Connection error:", error);
      addMessage(`Connection error: ${error instanceof Error ? error.message : String(error)}`, "error");
    }
  }, [keyPair, addMessage]);
  
  // Handle auto-connection after connect function is fully defined
  useEffect(() => {
    if (shouldAutoConnect && connect) {
      const timer = setTimeout(() => {
        addMessage("Attempting auto-connection with saved keys...", "system");
        connect();
        setShouldAutoConnect(false); // Reset flag to prevent multiple attempts
      }, 1500); // Small delay to allow UI to initialize
      
      return () => clearTimeout(timer);
    }
  }, [shouldAutoConnect, connect, addMessage]);
  
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
    // Generate new keys immediately
    generateKeyPair().then(newKeyPair => {
      if (newKeyPair) {
        setKeyPair(newKeyPair);
        localStorage.setItem('crypto_keypair', JSON.stringify(newKeyPair));
        try {
          setCurrentSigner(createEthersSigner(newKeyPair.privateKey));
          addMessage("Generated new key pair", "success");
          addMessage(`Ethereum Address: ${newKeyPair.address || 'unknown'}`, "info");
        } catch (e) {
          console.error('Failed to create signer from new keys:', e);
        }
      }
    });
  }, [addMessage, generateKeyPair]);
  
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