import { Address, Hex } from 'viem';

// ============================================================================
// Base Event Types
// ============================================================================

/**
 * Base channel event structure
 */
export interface ChannelEvent {
  channelId: string;
  stateVersion: bigint; // uint64
}

/**
 * Channel challenged event structure (includes challenge expiry)
 */
export interface ChannelChallengedEvent {
  channelId: string;
  stateVersion: bigint; // uint64
  challengeExpiry: bigint; // uint64
}

/**
 * Blockchain event metadata
 */
export interface BlockchainEvent {
  contractAddress: Address;
  blockchainId: bigint; // uint64
  name: string;
  blockNumber: bigint; // uint64
  transactionHash: Hex;
  logIndex: number; // uint32
}

// ============================================================================
// Home Channel Events
// ============================================================================

/**
 * HomeChannelCreatedEvent represents the ChannelCreated event
 */
export type HomeChannelCreatedEvent = ChannelEvent;

/**
 * HomeChannelMigratedEvent represents the ChannelMigrated event
 */
export type HomeChannelMigratedEvent = ChannelEvent;

/**
 * HomeChannelCheckpointedEvent represents the Checkpointed event
 */
export type HomeChannelCheckpointedEvent = ChannelEvent;

/**
 * HomeChannelChallengedEvent represents the Challenged event
 */
export type HomeChannelChallengedEvent = ChannelChallengedEvent;

/**
 * HomeChannelClosedEvent represents the Closed event
 */
export type HomeChannelClosedEvent = ChannelEvent;

// ============================================================================
// Escrow Deposit Events
// ============================================================================

/**
 * EscrowDepositInitiatedEvent represents the EscrowDepositInitiated event
 */
export type EscrowDepositInitiatedEvent = ChannelEvent;

/**
 * EscrowDepositChallengedEvent represents the EscrowDepositChallenged event
 */
export type EscrowDepositChallengedEvent = ChannelChallengedEvent;

/**
 * EscrowDepositFinalizedEvent represents the EscrowDepositFinalized event
 */
export type EscrowDepositFinalizedEvent = ChannelEvent;

// ============================================================================
// Escrow Withdrawal Events
// ============================================================================

/**
 * EscrowWithdrawalInitiatedEvent represents the EscrowWithdrawalInitiated event
 */
export type EscrowWithdrawalInitiatedEvent = ChannelEvent;

/**
 * EscrowWithdrawalChallengedEvent represents the EscrowWithdrawalChallenged event
 */
export type EscrowWithdrawalChallengedEvent = ChannelChallengedEvent;

/**
 * EscrowWithdrawalFinalizedEvent represents the EscrowWithdrawalFinalized event
 */
export type EscrowWithdrawalFinalizedEvent = ChannelEvent;
