/**
 * Nitrolite SDK for TypeScript
 *
 * A comprehensive SDK for building state channel applications
 * with the Nitrolite framework.
 */

// Core types
export * from "./types";

// Utils
export * from "./utils";

// Error types
export * from "./errors";

// Client (without re-exporting conflicting types)
export {
  NitroliteClient
} from "./client/NitroliteClient";
export {
  NitroliteClientConfig
} from "./client/config";
export {
  ChannelOperations
} from "./client/operations";
export {
  createNumericChannel,
  createSequentialChannel,
  createCustomChannel,
  ChannelContext
} from "./client/channels";

// Export from base config (avoiding conflicts)
export {
  DEFAULT_CONFIG,
  RPC_ERROR_CODES,
  SDKConfig,
  LogLevel,
  Logger,
  defaultLogger,
  createFilteredLogger,
  getConfigWithDefaults
} from "./config";

// RPC Relay
export {
  Message,
  MessageHandler,
  MessageProcessor,
  WebSocketOptions,
  WebSocketRelayConfig
} from "./relay";

// RPC Protocol
export * from "./rpc";

// Contract ABIs
export * from "./abis";
