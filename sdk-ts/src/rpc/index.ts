/**
 * RPC Protocol Layer
 * 
 * This module provides the RPC protocol implementation for Nitrolite SDK.
 * It enables transport-agnostic communication between participants in
 * a state channel network, including support for virtual channels.
 */

// Export main RPC types
export * from './types';

// Export the RPC client
export { RPCClient, RPCClientConfig } from './client';

// Export the virtual channel implementation
export { LVCI } from './virtual';

// Export providers
export { MemoryRPCProvider, MemoryProviderConfig } from './providers/memory';

// Export utility functions
export { createPayload, verifySignature, validateConnection, validateRequiredParams } from './utils';

// Re-export handlers for advanced usage
export * from './handlers';

// Export helper classes
export { VirtualChannelHelper } from './channels/virtual';