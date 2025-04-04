import { Hex } from 'viem';

/**
 * Application configuration
 *
 * This file contains configuration settings for the application,
 * including network endpoints and default values.
 */
export const APP_CONFIG = {
    // WebSocket configuration for real-time communication
    WEBSOCKET: {
        URL: 'ws://localhost:8000/ws',
    },

    // Channel configuration
    CHANNEL: {
        // Default counterparty address (for demo purposes)
        DEFAULT_ADDRESS: '0x70997970C51812dc3A010C7d01b50e0d17dc79C8' as Hex,

        // Default private key (for demo purposes only - NEVER use in production)
        DEFAULT_PRIVATE_KEY: '0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80' as Hex,

        // Challenge period in seconds (1 day)
        CHALLENGE_PERIOD: 86400,
    },

    // Default token configuration
    TOKENS: {
        // Default ERC20 token address
        DEFAULT_TOKEN: '0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2' as Hex,
    },
};

export default APP_CONFIG;
