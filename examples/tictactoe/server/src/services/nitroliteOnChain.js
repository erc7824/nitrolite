/**
 * Nitrolite on-chain operations (separate from WebSocket RPC)
 * This file handles all interactions with the blockchain
 */
import { createPublicClient, createWalletClient, http } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';
import { polygon } from 'viem/chains';
import { NitroliteClient } from '@erc7824/nitrolite';
import dotenv from 'dotenv';
import logger from '../utils/logger.js';

// Load environment variables
dotenv.config();

// Singleton instance of on-chain client
let nitroliteOnChainClient = null;

/**
 * Initialize a new Nitrolite on-chain client
 * @param {string} privateKey - Private key for the server wallet
 * @returns {NitroliteClient} The initialized Nitrolite client
 */
export async function initializeNitroliteOnChain(privateKey) {
  try {
    if (nitroliteOnChainClient) {
      logger.nitro('Nitrolite on-chain client already initialized');
      return nitroliteOnChainClient;
    }

    logger.nitro('Initializing Nitrolite on-chain client...');
    
    // Create wallet from private key
    const wallet = privateKeyToAccount(privateKey);
    logger.system('Wallet:', wallet);
    const address = wallet.address;
    
    logger.system(`Server wallet initialized with address: ${address}`);
    
    // Create client instances required by Nitrolite
    const publicClient = createPublicClient({
      transport: http(process.env.POLYGON_RPC_URL),
      chain: polygon,
    });

    // Create wallet client
    const walletClient = createWalletClient({
      transport: http(process.env.POLYGON_RPC_URL),
      chain: polygon,
      account: wallet,
    });

    // Create state wallet client
    const stateWalletClient = {
      account: {
        address: address,
      },
      signMessage: async ({ message: { raw }}) => {
        console.log('Signing message:', raw);
        const signature = await wallet.sign({ hash: raw });
        console.log('Signature:', signature);
        return signature;
    },
    };
    
    // Contract addresses from environment variables
    const addresses = {
      custody: process.env.CUSTODY_ADDRESS,
      adjudicator: process.env.ADJUDICATOR_ADDRESS,
      guestAddress: process.env.DEFAULT_GUEST_ADDRESS,
      tokenAddress: process.env.USDC_TOKEN_ADDRESS,
    };
    
    // Initialize Nitrolite client
    nitroliteOnChainClient = new NitroliteClient({
      publicClient,
      walletClient,
      stateWalletClient,
      account: address,
      chainId: Number(process.env.CHAIN_ID),
      challengeDuration: BigInt(1), // Use the same value as the client
      addresses,
    });
    
    logger.nitro('Nitrolite on-chain client initialized successfully');
    return nitroliteOnChainClient;
    
  } catch (error) {
    logger.error('Error initializing Nitrolite on-chain client:', error);
    throw error;
  }
}

/**
 * Get the existing Nitrolite on-chain client or initialize a new one
 * @param {string} privateKey - Private key for the server wallet
 * @returns {NitroliteClient} The Nitrolite client instance
 */
export async function getNitroliteOnChainClient(privateKey) {
  if (!nitroliteOnChainClient && privateKey) {
    await initializeNitroliteOnChain(privateKey);
  }
  return nitroliteOnChainClient;
}

/**
 * Create a new channel without deposit
 * @param {NitroliteClient} client - The Nitrolite client instance
 * @returns {Promise<Object>} The created channel result
 */
export async function createChannel(client) {
  try {
    if (!client) {
      throw new Error('Nitrolite client is required for channel creation');
    }
    
    logger.nitro('Creating new channel without deposit...', client);
    
    const result = await client.createChannel({
      initialAllocationAmounts: [0, 0],
      stateData: '0x', // Empty state data
    });
    
    logger.nitro('Channel created successfully');
    logger.data('Channel result', result);
    
    return result;
  } catch (error) {
    logger.error('Channel creation failed:', error);
    throw error;
  }
}

// End of exports
