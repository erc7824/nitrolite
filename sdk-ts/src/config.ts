/**
 * Configuration for Nitrolite SDK
 * 
 * This module centralizes configuration options and default values
 * for the entire SDK, making settings consistent and configurable.
 */

/**
 * SDK-wide configuration constants
 */
export const DEFAULT_CONFIG = {
  /**
   * Default timeout for requests in milliseconds (30 seconds)
   */
  REQUEST_TIMEOUT_MS: 30000,
  
  /**
   * Default maximum number of retries for failed requests
   */
  MAX_REQUEST_RETRIES: 3,
  
  /**
   * Default channel challenge period in seconds (1 day)
   */
  DEFAULT_CHALLENGE_PERIOD_SEC: 86400n,
  
  /**
   * Default Zero Address
   */
  ZERO_ADDRESS: '0x0000000000000000000000000000000000000000' as const,
  
  /**
   * Default retry backoff multiplier (for exponential backoff)
   */
  RETRY_BACKOFF_MULTIPLIER: 1.5,
  
  /**
   * Default initial retry delay in milliseconds
   */
  INITIAL_RETRY_DELAY_MS: 1000,
  
  /**
   * Maximum retry delay in milliseconds (30 seconds)
   */
  MAX_RETRY_DELAY_MS: 30000,
  
  /**
   * Default message signing deadline in milliseconds (5 minutes)
   */
  DEFAULT_SIGNATURE_DEADLINE_MS: 300000,
  
  /**
   * Maximum size for RPC message payloads in bytes (1MB)
   */
  MAX_RPC_MESSAGE_SIZE_BYTES: 1048576,
  
  /**
   * Default virtual channel nonce refresh interval in seconds (1 hour)
   */
  VIRTUAL_CHANNEL_NONCE_REFRESH_SEC: 3600,
  
  /**
   * Default virtual channel maximum hop count
   */
  MAX_VIRTUAL_CHANNEL_HOPS: 5,
  
  /**
   * Default buffer for LVCI path lookup (to avoid path expiration)
   */
  LVCI_PATH_BUFFER_MS: 5000,
};

/**
 * RPC standard error codes (from JSON-RPC spec)
 */
export const RPC_ERROR_CODES = {
  /**
   * Parse error (-32700)
   * 
   * Invalid JSON was received by the server.
   * An error occurred on the server while parsing the JSON text.
   */
  PARSE_ERROR: -32700,
  
  /**
   * Invalid Request (-32600)
   * 
   * The JSON sent is not a valid Request object.
   */
  INVALID_REQUEST: -32600,
  
  /**
   * Method not found (-32601)
   * 
   * The method does not exist / is not available.
   */
  METHOD_NOT_FOUND: -32601,
  
  /**
   * Invalid params (-32602)
   * 
   * Invalid method parameter(s).
   */
  INVALID_PARAMS: -32602,
  
  /**
   * Internal error (-32603)
   * 
   * Internal JSON-RPC error.
   */
  INTERNAL_ERROR: -32603,
  
  /**
   * Server error (-32000 to -32099)
   * 
   * Reserved for implementation-defined server-errors.
   */
  SERVER_ERROR: -32000,
  
  /**
   * Unauthorized (-32001)
   * 
   * The caller is not authorized to call this method.
   */
  UNAUTHORIZED: -32001,
  
  /**
   * Invalid state (-32002)
   * 
   * The state transition is invalid.
   */
  INVALID_STATE: -32002,
  
  /**
   * Channel not found (-32003)
   * 
   * The requested channel does not exist.
   */
  CHANNEL_NOT_FOUND: -32003,
  
  /**
   * Invalid signature (-32004)
   * 
   * The provided signature is invalid.
   */
  INVALID_SIGNATURE: -32004,
  
  /**
   * Invalid transition (-32005)
   * 
   * The state transition is invalid.
   */
  INVALID_TRANSITION: -32005,
  
  /**
   * Virtual channel error (-32006)
   * 
   * An error occurred with a virtual channel.
   */
  VIRTUAL_CHANNEL_ERROR: -32006,
  
  /**
   * Timeout error (-32007)
   * 
   * The operation timed out.
   */
  TIMEOUT: -32007,
};

/**
 * Global SDK configuration options
 */
export interface SDKConfig {
  /**
   * Request timeout in milliseconds
   * Default: 30000 (30 seconds)
   */
  requestTimeoutMs?: number;
  
  /**
   * Maximum number of request retries
   * Default: 3
   */
  maxRequestRetries?: number;
  
  /**
   * Default challenge period for channels in seconds
   * Default: 86400 (1 day)
   */
  defaultChallengePeriodSec?: bigint;
  
  /**
   * Logger instance for SDK logs
   * Default: console
   */
  logger?: Logger;
  
  /**
   * Log level for SDK logs
   * Default: 'info'
   */
  logLevel?: LogLevel;
  
  /**
   * Maximum virtual channel hop count
   * Default: 5
   */
  maxVirtualChannelHops?: number;
}

/**
 * Log levels
 */
export type LogLevel = 'debug' | 'info' | 'warn' | 'error' | 'none';

/**
 * Logger interface
 */
export interface Logger {
  debug(message: string, ...args: any[]): void;
  info(message: string, ...args: any[]): void;
  warn(message: string, ...args: any[]): void;
  error(message: string, ...args: any[]): void;
}

/**
 * Default logger implementation
 */
export const defaultLogger: Logger = {
  debug: (message: string, ...args: any[]) => console.debug(`[Nitrolite:debug] ${message}`, ...args),
  info: (message: string, ...args: any[]) => console.info(`[Nitrolite:info] ${message}`, ...args),
  warn: (message: string, ...args: any[]) => console.warn(`[Nitrolite:warn] ${message}`, ...args),
  error: (message: string, ...args: any[]) => console.error(`[Nitrolite:error] ${message}`, ...args),
};

/**
 * Create a filtered logger based on log level
 * @param level Minimum log level to output
 * @param baseLogger Base logger to use
 * @returns Filtered logger
 */
export function createFilteredLogger(level: LogLevel, baseLogger: Logger = defaultLogger): Logger {
  const levels: Record<LogLevel, number> = {
    debug: 0,
    info: 1,
    warn: 2,
    error: 3,
    none: 4,
  };
  
  const minLevel = levels[level];
  
  return {
    debug: (message: string, ...args: any[]) => 
      minLevel <= levels.debug && baseLogger.debug(message, ...args),
    info: (message: string, ...args: any[]) => 
      minLevel <= levels.info && baseLogger.info(message, ...args),
    warn: (message: string, ...args: any[]) => 
      minLevel <= levels.warn && baseLogger.warn(message, ...args),
    error: (message: string, ...args: any[]) => 
      minLevel <= levels.error && baseLogger.error(message, ...args),
  };
}

/**
 * Get configuration with default values applied
 * @param config User-provided configuration
 * @returns Configuration with defaults applied
 */
export function getConfigWithDefaults(config: SDKConfig = {}): Required<SDKConfig> {
  return {
    requestTimeoutMs: config.requestTimeoutMs ?? DEFAULT_CONFIG.REQUEST_TIMEOUT_MS,
    maxRequestRetries: config.maxRequestRetries ?? DEFAULT_CONFIG.MAX_REQUEST_RETRIES,
    defaultChallengePeriodSec: config.defaultChallengePeriodSec ?? DEFAULT_CONFIG.DEFAULT_CHALLENGE_PERIOD_SEC,
    logger: config.logger ?? defaultLogger,
    logLevel: config.logLevel ?? 'info',
    maxVirtualChannelHops: config.maxVirtualChannelHops ?? DEFAULT_CONFIG.MAX_VIRTUAL_CHANNEL_HOPS,
  };
}