/**
 * Contract type definitions for EVM blockchain interactions
 */

import { Address } from 'viem';

/**
 * ChannelDefinition represents the immutable definition of a state channel
 */
export interface ChannelDefinition {
  challengeDuration: number; // uint32
  user: Address;
  node: Address;
  nonce: bigint; // uint64
  metadata: `0x${string}`; // bytes32
}

/**
 * Ledger represents the balance allocations on a specific blockchain
 */
export interface Ledger {
  chainId: bigint; // uint64
  token: Address;
  decimals: number; // uint8
  userAllocation: bigint; // uint256
  userNetFlow: bigint; // int256
  nodeAllocation: bigint; // uint256
  nodeNetFlow: bigint; // int256
}

/**
 * State represents a state of a payment channel
 */
export interface State {
  version: bigint; // uint64
  intent: number; // uint8
  metadata: `0x${string}`; // bytes32
  homeState: Ledger;
  nonHomeState: Ledger;
  userSig: `0x${string}`; // bytes
  nodeSig: `0x${string}`; // bytes
}
