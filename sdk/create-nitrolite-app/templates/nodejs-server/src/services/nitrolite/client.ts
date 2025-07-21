import WebSocket from 'ws';
import { ethers } from 'ethers';
import {
  NitroliteRPC,
  parseAnyRPCResponse,
  RPCMethod,
  RPCResponse,
  MessageSigner,
  RequestData,
  ResponsePayload
} from '@erc7824/nitrolite';
import { config } from '../../config/index.js';
import { logger } from '../../utils/logger.js';
import type { WalletSigner } from './types.js';

let brokerWebSocket: WebSocket | null = null;
let isAuthenticated = false;
let jwtToken: string | null = null;

// Store pending requests
const pendingRequests = new Map<string, {
  resolve: (value: any) => void;
  reject: (error: Error) => void;
  timeout: NodeJS.Timeout;
}>();

/**
 * Creates a signer from a private key using ethers.js
 */
export function createEthersSigner(privateKey: string): WalletSigner {
  try {
    const wallet = new ethers.Wallet(privateKey);

    return {
      publicKey: wallet.address,
      address: wallet.address as `0x${string}`,
      sign: async (data: RequestData | ResponsePayload): Promise<`0x${string}`> => {
        try {
          const messageStr = JSON.stringify(data);
          const digestHex = ethers.id(messageStr);
          const messageBytes = ethers.getBytes(digestHex);
          const { serialized: signature } = wallet.signingKey.sign(messageBytes);
          return signature as `0x${string}`;
        } catch (error) {
          logger.error('Error signing message:', error);
          throw error;
        }
      },
    };
  } catch (error) {
    logger.error('Error creating ethers signer:', error);
    throw error;
  }
}

/**
 * Initialize Nitrolite RPC client and connect to broker
 */
export async function initializeNitroliteClient(): Promise<void> {
  if (!config.walletPrivateKey) {
    throw new Error('WALLET_PRIVATE_KEY is required');
  }

  try {
    await connectToBroker();
    logger.info('Nitrolite client initialized and connected to broker');
  } catch (error) {
    logger.error('Failed to initialize Nitrolite client:', error);
    throw error;
  }
}

/**
 * Connect to the Nitrolite broker
 */
async function connectToBroker(): Promise<void> {
  if (brokerWebSocket && (brokerWebSocket.readyState === WebSocket.OPEN || brokerWebSocket.readyState === WebSocket.CONNECTING)) {
    logger.info('WebSocket already connected or connecting');
    return;
  }

  logger.info(`Connecting to Nitrolite broker at ${config.yellowWsUrl}`);
  brokerWebSocket = new WebSocket(config.yellowWsUrl);
  isAuthenticated = false;

  brokerWebSocket.on('open', async () => {
    logger.info('Connected to Nitrolite broker');
    try {
      await authenticateWithBroker();
      logger.info('Successfully authenticated with broker');
    } catch (error) {
      logger.error('Authentication with broker failed:', error);
    }
  });

  brokerWebSocket.on('message', (data) => {
    try {
      const message = JSON.parse(data.toString());
      handleBrokerMessage(data.toString());
    } catch (error) {
      logger.error('Error parsing message from broker:', error);
    }
  });

  brokerWebSocket.on('close', (code, reason) => {
    logger.warn(`Disconnected from Nitrolite broker: ${code} ${reason.toString()}`);
    isAuthenticated = false;
    jwtToken = null;
    // Reconnect after 5 seconds
    setTimeout(connectToBroker, 5000);
  });

  brokerWebSocket.on('error', (error) => {
    logger.error('Error in broker WebSocket connection:', error.message);
  });
}

/**
 * Authenticate with the broker
 */
async function authenticateWithBroker(): Promise<void> {
  if (!brokerWebSocket || brokerWebSocket.readyState !== WebSocket.OPEN) {
    throw new Error('WebSocket not connected');
  }

  const signer = createEthersSigner(config.walletPrivateKey);
  const serverAddress = signer.address;

  const expire = String(Math.floor(Date.now() / 1000) + 24 * 60 * 60);

  const authMessage = {
    wallet: serverAddress,
    participant: serverAddress,
    app_name: config.vApp.name,
    expire: expire,
    scope: config.vApp.scope,
    application: serverAddress,
    allowances: [{
      asset: config.asset,
      amount: '0',
    }],
  };

  return new Promise(async (resolve, reject) => {
    const authTimeout = setTimeout(() => {
      reject(new Error('Authentication timeout'));
    }, 15000);

    const authMessageHandler = async (data: WebSocket.RawData) => {
      try {
        const message = parseAnyRPCResponse(data.toString());
        
        if (message.method === RPCMethod.AuthChallenge) {
          logger.info('Received auth challenge, sending verification...');
          // Handle auth challenge - implement EIP-712 signing here
          // For now, we'll use a simplified approach
        } else if (message.method === RPCMethod.AuthVerify) {
          logger.info('Authentication successful');
          if (message.params.jwtToken) {
            jwtToken = message.params.jwtToken;
          }
          isAuthenticated = true;
          clearTimeout(authTimeout);
          brokerWebSocket?.removeListener('message', authMessageHandler);
          resolve();
        } else if (message.method === RPCMethod.Error) {
          const errorMsg = message.params.error || 'Authentication failed';
          logger.error('Authentication failed:', errorMsg);
          clearTimeout(authTimeout);
          brokerWebSocket?.removeListener('message', authMessageHandler);
          reject(new Error(String(errorMsg)));
        }
      } catch (error) {
        logger.error('Error handling auth response:', error);
        clearTimeout(authTimeout);
        brokerWebSocket?.removeListener('message', authMessageHandler);
        reject(error);
      }
    };

    brokerWebSocket.on('message', authMessageHandler);

    try {
      // Create and send auth request
      const authRequest = await NitroliteRPC.createRequest(
        Date.now(),
        RPCMethod.AuthRequest,
        [authMessage]
      );
      const signedMessage = await NitroliteRPC.signRequestMessage(authRequest, signer.sign);
      brokerWebSocket.send(JSON.stringify(signedMessage));
    } catch (error) {
      logger.error('Error creating auth request:', error);
      clearTimeout(authTimeout);
      reject(error);
    }
  });
}

/**
 * Handle messages from the broker
 */
function handleBrokerMessage(raw: string): void {
  try {
    const message = parseAnyRPCResponse(raw);
    
    // Handle ping messages
    if (message.method === RPCMethod.Ping) {
      logger.debug('Received ping from broker, sending pong');
      if (brokerWebSocket && brokerWebSocket.readyState === WebSocket.OPEN) {
        brokerWebSocket.send(JSON.stringify({ type: 'pong' }));
      }
      return;
    }

    // Handle error messages
    if (message.method === RPCMethod.Error) {
      logger.error('Received error from broker:', message.params.error);
      
      if (message.requestId) {
        const pendingRequest = pendingRequests.get(message.requestId.toString());
        if (pendingRequest) {
          const { reject, timeout } = pendingRequest;
          clearTimeout(timeout);
          pendingRequests.delete(message.requestId.toString());
          reject(new Error(message.params.error));
        }
      }
      return;
    }

    // Handle other response messages
    if (message.requestId) {
      const pendingRequest = pendingRequests.get(message.requestId.toString());
      if (pendingRequest) {
        const { resolve, timeout } = pendingRequest;
        clearTimeout(timeout);
        pendingRequests.delete(message.requestId.toString());
        resolve(message.params || []);
        return;
      }
    }

    logger.debug('Received message from broker:', message);
  } catch (error) {
    logger.error('Error handling broker message:', error);
  }
}

/**
 * Send a request to the broker
 */
export async function sendToBroker(request: any): Promise<any> {
  if (!isAuthenticated && !(request.req && request.req[1] === 'auth_request')) {
    throw new Error('Not authenticated with broker');
  }

  if (!brokerWebSocket || brokerWebSocket.readyState !== WebSocket.OPEN) {
    throw new Error('Not connected to broker');
  }

  return new Promise((resolve, reject) => {
    const requestId = request.req?.[0] || Date.now().toString();
    
    const timeout = setTimeout(() => {
      pendingRequests.delete(requestId.toString());
      reject(new Error('Request timeout'));
    }, 10000);

    pendingRequests.set(requestId.toString(), { resolve, reject, timeout });
    brokerWebSocket!.send(JSON.stringify(request));
  });
}

/**
 * Check if authenticated with broker
 */
export function isAuthenticatedWithBroker(): boolean {
  return isAuthenticated;
}

/**
 * Get broker WebSocket connection
 */
export function getBrokerWebSocket(): WebSocket | null {
  return brokerWebSocket;
}