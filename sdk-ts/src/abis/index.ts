/**
 * ABIs for Hachi contracts
 * 
 * This file exports all ABIs used by the SDK as well as common fragments and types
 */

// Export main ABIs
export { CustodyAbi } from './custody';
export { AdjudicatorAbi } from './adjudicator';
export { Erc20Abi } from './token';

// Export ABI fragments
export * from './fragments';

// Export ABI types
export * from './types';

// Type demonstration for properly configuring addresses
// IMPORTANT: This is only a type example, not actual implementation
// Developers must provide their own addresses for the contracts
export const defaultAbiConfig = {
  // This is just an example structure
  // The SDK will require proper addresses to be provided by the user
  chainId: 1, // Example chain ID
  addresses: {
    custody: '0x0000000000000000000000000000000000000000', // Placeholder - not for use
    adjudicators: {
      base: '0x0000000000000000000000000000000000000000', // Placeholder - not for use
      numeric: '0x0000000000000000000000000000000000000000', // Placeholder - not for use
      sequential: '0x0000000000000000000000000000000000000000', // Placeholder - not for use
    }
  }
};