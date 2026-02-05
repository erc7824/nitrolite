import { Address, encodeAbiParameters, keccak256, toHex, pad, slice } from 'viem';
import Decimal from 'decimal.js';
import {
  Transition,
  TransitionType,
  INTENT_OPERATE,
  INTENT_CLOSE,
  INTENT_DEPOSIT,
  INTENT_WITHDRAW,
  INTENT_INITIATE_ESCROW_DEPOSIT,
  INTENT_FINALIZE_ESCROW_DEPOSIT,
  INTENT_INITIATE_ESCROW_WITHDRAWAL,
  INTENT_FINALIZE_ESCROW_WITHDRAWAL,
  INTENT_INITIATE_MIGRATION,
} from './types';

// Configure Decimal.js for high precision arithmetic
Decimal.set({ precision: 50 });

// ============================================================================
// Intent Conversion
// ============================================================================

/**
 * TransitionToIntent maps a transition type to its on-chain intent value
 * @param transition - The transition to convert
 * @returns Intent value (uint8)
 * @throws Error if transition is null or has unexpected type
 */
export function transitionToIntent(transition: Transition | null | undefined): number {
  if (!transition) {
    throw new Error('at least one transition is expected');
  }

  switch (transition.type) {
    case TransitionType.TransferSend:
    case TransitionType.TransferReceive:
    case TransitionType.Commit:
    case TransitionType.Release:
      return INTENT_OPERATE;
    case TransitionType.Finalize:
      return INTENT_CLOSE;
    case TransitionType.HomeDeposit:
      return INTENT_DEPOSIT;
    case TransitionType.HomeWithdrawal:
      return INTENT_WITHDRAW;
    case TransitionType.MutualLock:
      return INTENT_INITIATE_ESCROW_DEPOSIT;
    case TransitionType.EscrowDeposit:
      return INTENT_FINALIZE_ESCROW_DEPOSIT;
    case TransitionType.EscrowLock:
      return INTENT_INITIATE_ESCROW_WITHDRAWAL;
    case TransitionType.EscrowWithdraw:
      return INTENT_FINALIZE_ESCROW_WITHDRAWAL;
    case TransitionType.Migrate:
      return INTENT_INITIATE_MIGRATION;
    // TODO: Add FINALIZE_MIGRATION
    default:
      throw new Error(`unexpected transition type: ${transition.type}`);
  }
}

// ============================================================================
// Decimal Validation & Conversion
// ============================================================================

/**
 * ValidateDecimalPrecision validates that an amount doesn't exceed the maximum allowed decimal places
 * @param amount - The decimal amount to validate
 * @param maxDecimals - Maximum allowed decimal places (uint8)
 * @throws Error if amount exceeds precision
 */
export function validateDecimalPrecision(amount: Decimal, maxDecimals: number): void {
  const exponent = amount.decimalPlaces();
  if (exponent > maxDecimals) {
    throw new Error(
      `amount exceeds maximum decimal precision: max ${maxDecimals} decimals allowed, got ${exponent}`
    );
  }
}

/**
 * DecimalToBigInt converts a decimal.Decimal amount to bigint scaled to the token's smallest unit.
 * For example, 1.23 USDC (6 decimals) becomes 1230000n.
 * This is used when preparing amounts for smart contract calls.
 * @param amount - The decimal amount
 * @param decimals - Number of decimals for the token (uint8)
 * @returns Scaled bigint amount
 * @throws Error if amount has too many decimal places
 */
export function decimalToBigInt(amount: Decimal, decimals: number): bigint {
  // Calculate the multiplier (e.g., 10^6)
  const multiplier = new Decimal(10).pow(decimals);

  // Scale the amount
  const scaled = amount.mul(multiplier);

  // Check if it's an integer
  if (!scaled.isInteger()) {
    throw new Error(
      `amount ${amount.toString()} exceeds maximum decimal precision: max ${decimals} decimals allowed`
    );
  }

  // Convert to bigint
  return BigInt(scaled.toFixed(0));
}

// ============================================================================
// Channel ID Generation
// ============================================================================

/**
 * GetHomeChannelID generates a unique identifier for a primary channel based on its definition.
 * This matches the Solidity getChannelId function which computes keccak256(abi.encode(ChannelDefinition)).
 * The metadata is derived from the asset: first 8 bytes of keccak256(asset) padded to 32 bytes.
 * @param node - Node address
 * @param user - User wallet address
 * @param asset - Asset symbol
 * @param nonce - Channel nonce (uint64)
 * @param challengeDuration - Challenge period in seconds (uint32)
 * @returns Channel ID as hex string
 */
export function getHomeChannelId(
  node: Address,
  user: Address,
  asset: string,
  nonce: bigint,
  challengeDuration: number
): string {
  // Generate metadata from asset
  const metadata = generateChannelMetadata(asset);

  // Define the channel definition struct matching Solidity
  // struct ChannelDefinition {
  //   uint32 challengeDuration;
  //   address user;
  //   address node;
  //   uint64 nonce;
  //   bytes32 metadata;
  // }
  const packed = encodeAbiParameters(
    [
      {
        name: 'channelDefinition',
        type: 'tuple',
        components: [
          { name: 'challengeDuration', type: 'uint32' },
          { name: 'user', type: 'address' },
          { name: 'node', type: 'address' },
          { name: 'nonce', type: 'uint64' },
          { name: 'metadata', type: 'bytes32' },
        ],
      },
    ],
    [
      {
        challengeDuration: challengeDuration,
        user: user,
        node: node,
        nonce: nonce,
        metadata: metadata,
      },
    ]
  );

  return keccak256(packed);
}

/**
 * GetEscrowChannelID derives an escrow-specific channel ID based on a home channel and state version.
 * This matches the Solidity getEscrowId function which computes keccak256(abi.encode(channelId, version)).
 * @param homeChannelId - Home channel ID (bytes32)
 * @param stateVersion - State version (uint64)
 * @returns Escrow channel ID as hex string
 */
export function getEscrowChannelId(homeChannelId: string, stateVersion: bigint): string {
  const packed = encodeAbiParameters(
    [{ type: 'bytes32' }, { type: 'uint64' }],
    [homeChannelId as `0x${string}`, stateVersion]
  );

  return keccak256(packed);
}

// ============================================================================
// State ID Generation
// ============================================================================

/**
 * GetStateID creates a unique hash representing a specific snapshot of a user's wallet and asset state.
 * @param userWallet - User wallet address
 * @param asset - Asset symbol
 * @param epoch - User epoch index (uint64)
 * @param version - State version (uint64)
 * @returns State ID as hex string
 */
export function getStateId(
  userWallet: Address,
  asset: string,
  epoch: bigint,
  version: bigint
): string {
  const packed = encodeAbiParameters(
    [{ type: 'address' }, { type: 'string' }, { type: 'uint256' }, { type: 'uint256' }],
    [userWallet, asset, epoch, version]
  );

  return keccak256(packed);
}

/**
 * GetStateTransitionsHash hashes a list of transitions into metadata
 * @param transitions - Array of transitions
 * @returns Hash as bytes32 (hex string)
 */
export function getStateTransitionsHash(transitions: Transition[]): string {
  // Convert transitions to contract format
  const contractTransitions = transitions.map((t) => ({
    type: t.type,
    txId: t.txId,
    accountId: t.accountId,
    amount: t.amount.toString(),
  }));

  const packed = encodeAbiParameters(
    [
      {
        type: 'tuple[]',
        components: [
          { name: 'type', type: 'uint8' },
          { name: 'txId', type: 'string' },
          { name: 'accountId', type: 'string' },
          { name: 'amount', type: 'string' },
        ],
      },
    ],
    [contractTransitions]
  );

  return keccak256(packed);
}

// ============================================================================
// Transaction ID Generation
// ============================================================================

/**
 * GetSenderTransactionID calculates and returns a unique transaction ID reference for actions initiated by user.
 * @param toAccount - Recipient account
 * @param senderNewStateId - Sender's new state ID
 * @returns Transaction ID as hex string
 */
export function getSenderTransactionId(toAccount: string, senderNewStateId: string): string {
  return getTransactionId(toAccount, senderNewStateId);
}

/**
 * GetReceiverTransactionID calculates and returns a unique transaction ID reference for actions initiated by node.
 * @param fromAccount - Sender account
 * @param receiverNewStateId - Receiver's new state ID
 * @returns Transaction ID as hex string
 */
export function getReceiverTransactionId(fromAccount: string, receiverNewStateId: string): string {
  return getTransactionId(fromAccount, receiverNewStateId);
}

function getTransactionId(account: string, newStateId: string): string {
  const packed = encodeAbiParameters(
    [{ type: 'string' }, { type: 'bytes32' }],
    [account, newStateId as `0x${string}`]
  );

  return keccak256(packed);
}

// ============================================================================
// Metadata Generation
// ============================================================================

/**
 * GenerateChannelMetadata creates metadata from an asset by taking the first 8 bytes of keccak256(asset)
 * and padding the rest with zeros to make a 32-byte array.
 * @param asset - Asset symbol
 * @returns 32-byte metadata as hex string
 */
export function generateChannelMetadata(asset: string): `0x${string}` {
  // Hash the asset
  const assetHash = keccak256(toHex(asset));

  // Take first 8 bytes and pad with zeros to 32 bytes (pad on the right)
  const first8Bytes = slice(assetHash, 0, 8);
  const metadata = pad(first8Bytes, { dir: 'right', size: 32 });

  return metadata;
}
