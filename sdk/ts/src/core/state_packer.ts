import { Address, encodeAbiParameters } from 'viem';
import { State } from './types';
import { AssetStore, StatePacker } from './interface';
import { getStateTransitionHash, transitionToIntent, decimalToBigInt } from './utils';

/**
 * ContractLedger matches Solidity's Ledger struct for ABI encoding
 */
interface ContractLedger {
  chainId: bigint; // uint64
  token: Address;
  decimals: number; // uint8
  userAllocation: bigint; // uint256
  userNetFlow: bigint; // int256
  nodeAllocation: bigint; // uint256
  nodeNetFlow: bigint; // int256
}

/**
 * StatePackerV1 encodes states into ABI-packed bytes for on-chain submission
 */
export class StatePackerV1 implements StatePacker {
  private assetStore: AssetStore;

  constructor(assetStore: AssetStore) {
    this.assetStore = assetStore;
  }

  /**
   * PackState encodes a channel ID and state into ABI-packed bytes for on-chain submission.
   * This matches the Solidity contract's two-step encoding:
   *
   *   signingData = abi.encode(version, intent, metadata, homeLedger, nonHomeLedger)
   *   message = abi.encode(channelId, signingData)
   *
   * The signingData is encoded as dynamic bytes inside the outer abi.encode.
   * @param state - State to pack
   * @returns Packed bytes as hex string
   */
  async packState(state: State): Promise<`0x${string}`> {
    // Ensure HomeChannelID is present
    if (!state.homeChannelId) {
      throw new Error('state.homeChannelId is required for packing');
    }

    // Convert HomeChannelID to bytes32
    const channelId = state.homeChannelId as `0x${string}`;

    // Generate metadata from state transition
    const metadata = getStateTransitionHash(state.transition);

    // Get home ledger decimals
    const homeDecimals = await this.assetStore.getTokenDecimals(
      state.homeLedger.blockchainId,
      state.homeLedger.tokenAddress
    );

    // Convert decimal amounts to bigint scaled to the token's smallest unit
    const userBalanceBi = decimalToBigInt(state.homeLedger.userBalance, homeDecimals);
    const userNetFlowBi = decimalToBigInt(state.homeLedger.userNetFlow, homeDecimals);
    const nodeBalanceBi = decimalToBigInt(state.homeLedger.nodeBalance, homeDecimals);
    const nodeNetFlowBi = decimalToBigInt(state.homeLedger.nodeNetFlow, homeDecimals);

    const homeLedger: ContractLedger = {
      chainId: state.homeLedger.blockchainId,
      token: state.homeLedger.tokenAddress,
      decimals: homeDecimals,
      userAllocation: userBalanceBi,
      userNetFlow: userNetFlowBi,
      nodeAllocation: nodeBalanceBi,
      nodeNetFlow: nodeNetFlowBi,
    };

    // For nonHomeState, use escrow ledger if available, otherwise use zero values
    let nonHomeLedger: ContractLedger;

    if (state.escrowLedger) {
      const escrowDecimals = await this.assetStore.getTokenDecimals(
        state.escrowLedger.blockchainId,
        state.escrowLedger.tokenAddress
      );

      const escrowUserBalanceBi = decimalToBigInt(state.escrowLedger.userBalance, escrowDecimals);
      const escrowUserNetFlowBi = decimalToBigInt(state.escrowLedger.userNetFlow, escrowDecimals);
      const escrowNodeBalanceBi = decimalToBigInt(state.escrowLedger.nodeBalance, escrowDecimals);
      const escrowNodeNetFlowBi = decimalToBigInt(state.escrowLedger.nodeNetFlow, escrowDecimals);

      nonHomeLedger = {
        chainId: state.escrowLedger.blockchainId,
        token: state.escrowLedger.tokenAddress,
        decimals: escrowDecimals,
        userAllocation: escrowUserBalanceBi,
        userNetFlow: escrowUserNetFlowBi,
        nodeAllocation: escrowNodeBalanceBi,
        nodeNetFlow: escrowNodeNetFlowBi,
      };
    } else {
      nonHomeLedger = {
        chainId: 0n,
        token: '0x0000000000000000000000000000000000000000' as Address,
        decimals: 0,
        userAllocation: 0n,
        userNetFlow: 0n,
        nodeAllocation: 0n,
        nodeNetFlow: 0n,
      };
    }

    // Determine intent based on transition
    const intent = transitionToIntent(state.transition);

    // Define the Ledger tuple type matching Solidity
    const ledgerComponents = [
      { name: 'chainId', type: 'uint64' },
      { name: 'token', type: 'address' },
      { name: 'decimals', type: 'uint8' },
      { name: 'userAllocation', type: 'uint256' },
      { name: 'userNetFlow', type: 'int256' },
      { name: 'nodeAllocation', type: 'uint256' },
      { name: 'nodeNetFlow', type: 'int256' },
    ] as const;

    // Step 1: Pack signingData = abi.encode(version, intent, metadata, homeLedger, nonHomeLedger)
    const signingData = encodeAbiParameters(
      [
        { type: 'uint64' }, // version
        { type: 'uint8' }, // intent
        { type: 'bytes32' }, // metadata
        { type: 'tuple', components: ledgerComponents }, // homeState
        { type: 'tuple', components: ledgerComponents }, // nonHomeState
      ],
      [
        state.version,
        intent,
        metadata as `0x${string}`,
        {
          chainId: homeLedger.chainId,
          token: homeLedger.token,
          decimals: homeLedger.decimals,
          userAllocation: homeLedger.userAllocation,
          userNetFlow: homeLedger.userNetFlow,
          nodeAllocation: homeLedger.nodeAllocation,
          nodeNetFlow: homeLedger.nodeNetFlow,
        },
        {
          chainId: nonHomeLedger.chainId,
          token: nonHomeLedger.token,
          decimals: nonHomeLedger.decimals,
          userAllocation: nonHomeLedger.userAllocation,
          userNetFlow: nonHomeLedger.userNetFlow,
          nodeAllocation: nonHomeLedger.nodeAllocation,
          nodeNetFlow: nonHomeLedger.nodeNetFlow,
        },
      ]
    );

    // Step 2: Pack message = abi.encode(channelId, signingData)
    // This matches Solidity: Utils.pack(channelId, signingData) = abi.encode(channelId, signingData)
    // where signingData is dynamic bytes
    const packed = encodeAbiParameters(
      [
        { type: 'bytes32' }, // channelId
        { type: 'bytes' },   // signingData (dynamic bytes)
      ],
      [
        channelId,
        signingData,
      ]
    );

    return packed;
  }
}

/**
 * NewStatePackerV1 creates a new state packer instance
 * @param assetStore - Asset store for retrieving token metadata
 * @returns StatePackerV1 instance
 */
export function newStatePackerV1(assetStore: AssetStore): StatePackerV1 {
  return new StatePackerV1(assetStore);
}

/**
 * PackState is a convenience function that creates a StatePackerV1 and packs the state.
 * For production use, create a StatePackerV1 instance and reuse it.
 * @param state - State to pack
 * @param assetStore - Asset store for retrieving token metadata
 * @returns Packed bytes as hex string
 */
export async function packState(state: State, assetStore: AssetStore): Promise<`0x${string}`> {
  const packer = newStatePackerV1(assetStore);
  return packer.packState(state);
}
