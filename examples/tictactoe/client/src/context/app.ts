import { type Address } from "viem";

/**
 * Application configuration
 *
 * This file contains configuration settings for the application,
 * including network endpoints and default values.
 */
export const APP_CONFIG = {
    // WebSocket configuration for real-time communication
    WEBSOCKET: {
        URL: "wss://clearnet.yellow.com/ws",
    },

    CHANNEL: {
        DEFAULT_GUEST: "0x3c93C321634a80FB3657CFAC707718A11cA57cBf",
        CHALLENGE_PERIOD: BigInt(3600),
    },

    TOKENS: {
        137: "0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359" as Address,
    },

    CUSTODIES: {
        137: "0x4C8Bd8877C3b403BA9f9ECfaAD910AF0d8CA2c4D" as Address,
    },

    DEFAULT_ADJUDICATOR: "dummy",

    ADJUDICATORS: {
        137: "0x5F4A4B1D293A973a1Bc0daD3BB3692Bd51058FCF" as Address,
    },
};

export default APP_CONFIG;
