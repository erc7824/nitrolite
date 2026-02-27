/**
 * Configuration for the Nitrolite SDK Client
 */

/**
 * Config holds the configuration options for the Nitrolite client.
 */
export interface Config {
  /** WebSocket URL of the Nitrolite server */
  url: string;

  /** Maximum time to wait for initial connection (in milliseconds) */
  handshakeTimeout?: number;

  /** Called when connection errors occur */
  errorHandler?: (error: Error) => void;

  /** Maps blockchain IDs to their RPC endpoints */
  blockchainRPCs?: Map<bigint, string>;
}

/**
 * Option is a functional option for configuring the Client.
 */
export type Option = (config: Config) => void;

/**
 * Default error handler logs errors to console.
 */
function defaultErrorHandler(err: Error): void {
  if (err) {
    console.error('[nitrolite]', 'connection error:', err);
  }
}

/**
 * DefaultConfig returns the default configuration with sensible defaults.
 */
export const DefaultConfig: Partial<Config> = {
  handshakeTimeout: 5000, // 5 seconds
  errorHandler: defaultErrorHandler,
  blockchainRPCs: new Map(),
};

/**
 * WithHandshakeTimeout sets the maximum time to wait for initial connection.
 */
export function withHandshakeTimeout(timeout: number): Option {
  return (config: Config) => {
    config.handshakeTimeout = timeout;
  };
}

/**
 * WithErrorHandler sets a custom error handler for connection errors.
 * The handler is called when the connection encounters an error or is closed.
 */
export function withErrorHandler(handler: (error: Error) => void): Option {
  return (config: Config) => {
    config.errorHandler = handler;
  };
}

/**
 * WithBlockchainRPC configures the RPC endpoint for a specific blockchain.
 * This is required for operations that interact with the blockchain (deposit, withdraw, etc.).
 *
 * @param chainId - The blockchain ID (e.g., 80002n for Polygon Amoy)
 * @param rpcUrl - The RPC endpoint URL
 */
export function withBlockchainRPC(chainId: bigint, rpcUrl: string): Option {
  return (config: Config) => {
    if (!config.blockchainRPCs) {
      config.blockchainRPCs = new Map();
    }
    config.blockchainRPCs.set(chainId, rpcUrl);
  };
}
