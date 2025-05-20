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
        URL: "wss://clearnode-multichain-production.up.railway.app/ws",
    },

    CHANNEL: {
        DEFAULT_GUEST: "0x3c93C321634a80FB3657CFAC707718A11cA57cBf",
        CHALLENGE_PERIOD: BigInt(1),
    },

    TOKENS: {
        137: "0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359" as Address,
    },

    CUSTODIES: {
        137: "0x1096644156Ed58BF596e67d35827Adc97A25D940" as Address,
    },

    DEFAULT_ADJUDICATOR: "dummy",

    ADJUDICATORS: {
        137: "0xa3f2f64455c9f8D68d9dCAeC2605D64680FaF898" as Address,
    },
};

export default APP_CONFIG;
