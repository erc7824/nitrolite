import { Address } from 'viem';

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
        // Default address for the second participant
        DEFAULT_GUEST: '0x70997970C51812dc3A010C7d01b50e0d17dc79C8',
        GUEST_KEY: '0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d',
        // Challenge period in seconds (1 day)
        CHALLENGE_PERIOD: 86400,
    },

    CUSTODIES: {
        1337: '0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE' as Address,
    },

    ADJUDICATORS: {
        flag: '0x5fbdb2315678afecb367f032d93f642f64180aa3' as Address,
    },
};

export default APP_CONFIG;
