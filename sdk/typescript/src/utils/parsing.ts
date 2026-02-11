import { Address, Hex } from 'viem';
import { ChannelDefinition, Ledger, State } from '../client/types';

/**
 * Type representing raw channel definition data as returned from the contract.
 * This matches the structure of the ChannelDefinition struct in Solidity.
 */
export interface RawChannelDefinition {
    challengeDuration: number;
    user: Address;
    node: Address;
    nonce: bigint;
    metadata: Hex;
}

/**
 * Type representing raw ledger data as returned from the contract.
 * This matches the structure of the Ledger struct in Solidity.
 */
export interface RawLedger {
    chainId: bigint;
    token: Address;
    decimals: number;
    userAllocation: bigint;
    userNetFlow: bigint;
    nodeAllocation: bigint;
    nodeNetFlow: bigint;
}

/**
 * Type representing raw state data as returned from the contract.
 * This matches the structure of the State struct in Solidity.
 */
export interface RawState {
    version: bigint;
    intent: number;
    metadata: Hex;
    homeState: RawLedger;
    nonHomeState: RawLedger;
    userSig: Hex;
    nodeSig: Hex;
}

/**
 * Parse raw channel definition data from contract into SDK type.
 * Useful when reading channel data directly from the contract.
 * @param raw Raw channel definition from contract call.
 * @returns Parsed ChannelDefinition object.
 */
export function parseChannelDefinition(raw: RawChannelDefinition): ChannelDefinition {
    return {
        challengeDuration: raw.challengeDuration,
        user: raw.user,
        node: raw.node,
        nonce: raw.nonce,
        metadata: raw.metadata,
    };
}

/**
 * Parse raw ledger data from contract into SDK type.
 * Useful when reading state data directly from the contract.
 * @param raw Raw ledger from contract call.
 * @returns Parsed Ledger object.
 */
export function parseLedger(raw: RawLedger): Ledger {
    return {
        chainId: raw.chainId,
        token: raw.token,
        decimals: raw.decimals,
        userAllocation: raw.userAllocation,
        userNetFlow: raw.userNetFlow,
        nodeAllocation: raw.nodeAllocation,
        nodeNetFlow: raw.nodeNetFlow,
    };
}

/**
 * Parse raw state data from contract into SDK type.
 * Useful when reading state data directly from the contract.
 * @param raw Raw state from contract call.
 * @returns Parsed State object.
 */
export function parseState(raw: RawState): State {
    return {
        version: raw.version,
        intent: raw.intent,
        metadata: raw.metadata,
        homeState: parseLedger(raw.homeState),
        nonHomeState: parseLedger(raw.nonHomeState),
        userSig: raw.userSig,
        nodeSig: raw.nodeSig,
    };
}
