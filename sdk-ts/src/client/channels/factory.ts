import {
  Address,
  Hex,
  encodeAbiParameters,
  Abi
} from 'viem';
import { 
  AppDataTypes,
  AppLogic
} from '../../types';
import { createAppLogic } from '../../utils';
import { ChannelContext } from './ChannelContext';
import type { HachiClient } from '../HachiClient';

/**
 * Create a numeric value application
 * @param client The Hachi client
 * @param params Application parameters
 * @returns A new channel context with a numeric value app
 */
export function createNumericChannel(
  client: HachiClient,
  params: {
    participants: [Address, Address];
    adjudicatorAddress?: Address;
    adjudicatorAbi?: Abi;
    challenge?: bigint;
    nonce?: bigint;
    initialValue?: bigint;
    finalValue?: bigint;
  }
): ChannelContext<AppDataTypes.NumericState> {
  const adjudicatorType = 'numeric';
  
  // Get adjudicator address
  let adjudicatorAddress: Address;
  try {
    // First try to use explicitly provided address
    if (params.adjudicatorAddress) {
      adjudicatorAddress = params.adjudicatorAddress;
    } else {
      // Otherwise try to get from registered adjudicators
      adjudicatorAddress = client.getAdjudicatorAddress(adjudicatorType);
    }
  } catch (error) {
    // If no adjudicator found and we have an ABI, throw a more helpful error
    if (params.adjudicatorAbi) {
      throw new Error(
        `No adjudicator address found for type '${adjudicatorType}'. ` +
        `You provided an ABI but need to also provide an address via adjudicatorAddress parameter ` +
        `or by registering an address for this type in the client configuration.`
      );
    }
    // Otherwise re-throw the original error
    throw error;
  }
  
  // Register custom adjudicator ABI if provided
  if (params.adjudicatorAbi) {
    client.registerAdjudicatorAbi(adjudicatorType, params.adjudicatorAbi);
  }
  
  // Create the app logic
  const appLogic = createAppLogic<AppDataTypes.NumericState>({
    adjudicatorAddress,
    adjudicatorType,
    encode: (data) => {
      return encodeAbiParameters([{ type: 'uint256', name: 'value' }], [data.value]);
    },
    decode: (encoded) => {
      // In a real implementation, this would properly decode the data
      // This is a simplified placeholder
      return { value: BigInt(0) };
    },
    validateTransition: (prevState, nextState) => {
      // Value can only increase
      return nextState.value > prevState.value;
    },
    isFinal: params.finalValue ? 
      (state) => state.value >= (params.finalValue || BigInt(0)) : 
      undefined
  });
  
  // Create the channel
  const channel = {
    participants: params.participants,
    adjudicator: adjudicatorAddress,
    challenge: params.challenge || BigInt(86400), // Default: 1 day
    nonce: params.nonce || BigInt(Date.now())
  };
  
  // Create and return the channel context
  return new ChannelContext<AppDataTypes.NumericState>(
    client,
    channel,
    appLogic,
    { value: params.initialValue || BigInt(0) }
  );
}

/**
 * Create a sequential state application
 * @param client The Hachi client
 * @param params Application parameters
 * @returns A new channel context with a sequential state app
 */
export function createSequentialChannel(
  client: HachiClient,
  params: {
    participants: [Address, Address];
    adjudicatorAddress?: Address;
    adjudicatorAbi?: Abi;
    challenge?: bigint;
    nonce?: bigint;
    initialValue?: bigint;
  }
): ChannelContext<AppDataTypes.SequentialState> {
  const adjudicatorType = 'sequential';
  
  // Get adjudicator address
  let adjudicatorAddress: Address;
  try {
    // First try to use explicitly provided address
    if (params.adjudicatorAddress) {
      adjudicatorAddress = params.adjudicatorAddress;
    } else {
      // Otherwise try to get from registered adjudicators
      adjudicatorAddress = client.getAdjudicatorAddress(adjudicatorType);
    }
  } catch (error) {
    // If no adjudicator found and we have an ABI, throw a more helpful error
    if (params.adjudicatorAbi) {
      throw new Error(
        `No adjudicator address found for type '${adjudicatorType}'. ` +
        `You provided an ABI but need to also provide an address via adjudicatorAddress parameter ` +
        `or by registering an address for this type in the client configuration.`
      );
    }
    // Otherwise re-throw the original error
    throw error;
  }
  
  // Register custom adjudicator ABI if provided
  if (params.adjudicatorAbi) {
    client.registerAdjudicatorAbi(adjudicatorType, params.adjudicatorAbi);
  }
  
  // Create the app logic
  const appLogic = createAppLogic<AppDataTypes.SequentialState>({
    adjudicatorAddress,
    adjudicatorType,
    encode: (data) => {
      return encodeAbiParameters(
        [
          { type: 'uint256', name: 'sequence' },
          { type: 'uint256', name: 'value' }
        ],
        [data.sequence, data.value]
      );
    },
    decode: (encoded) => {
      // In a real implementation, this would properly decode the data
      // This is a simplified placeholder
      return { sequence: BigInt(0), value: BigInt(0) };
    },
    validateTransition: (prevState, nextState) => {
      // Sequence must increase, value cannot decrease
      return (
        nextState.sequence > prevState.sequence && 
        nextState.value >= prevState.value
      );
    }
  });
  
  // Create the channel
  const channel = {
    participants: params.participants,
    adjudicator: adjudicatorAddress,
    challenge: params.challenge || BigInt(86400), // Default: 1 day
    nonce: params.nonce || BigInt(Date.now())
  };
  
  // Create and return the channel context
  return new ChannelContext<AppDataTypes.SequentialState>(
    client,
    channel,
    appLogic,
    { sequence: BigInt(0), value: params.initialValue || BigInt(0) }
  );
}

/**
 * Create a custom application
 * @param client The Hachi client
 * @param params Custom application parameters
 * @returns A new channel context with custom app logic
 */
export function createCustomChannel<T = unknown>(
  client: HachiClient,
  params: {
    participants: [Address, Address];
    challenge?: bigint;
    nonce?: bigint;
    encode: (data: T) => Hex;
    decode: (encoded: Hex) => T;
    validateTransition?: (prevState: T, nextState: T, signer: Address) => boolean;
    isFinal?: (state: T) => boolean;
    adjudicatorAddress?: Address;
    adjudicatorType?: string;
    adjudicatorAbi?: Abi;
    initialState?: T;
  }
): ChannelContext<T> {
  // Determine the adjudicator address
  let adjudicatorAddress: Address;
  const adjudicatorType = params.adjudicatorType || 'base';
  
  try {
    if (params.adjudicatorAddress) {
      // Use explicitly provided address
      adjudicatorAddress = params.adjudicatorAddress;
    } else if (params.adjudicatorType) {
      // Look up address by type
      adjudicatorAddress = client.getAdjudicatorAddress(params.adjudicatorType);
    } else {
      // Use default "base" adjudicator
      adjudicatorAddress = client.getAdjudicatorAddress('base');
    }
  } catch (error) {
    // If no adjudicator found and we have an ABI, throw a more helpful error
    if (params.adjudicatorAbi) {
      throw new Error(
        `No adjudicator address found for type '${adjudicatorType}'. ` +
        `You provided an ABI but need to also provide an address via adjudicatorAddress parameter ` +
        `or by registering an address for this type in the client configuration.`
      );
    }
    // Otherwise re-throw the original error
    throw error;
  }
  
  // Register custom adjudicator ABI if provided
  if (params.adjudicatorAbi && params.adjudicatorType) {
    client.registerAdjudicatorAbi(params.adjudicatorType, params.adjudicatorAbi);
  }
  
  // Create the app logic
  const appLogic = createAppLogic<T>({
    adjudicatorAddress,
    adjudicatorType: params.adjudicatorType,
    encode: params.encode,
    decode: params.decode,
    validateTransition: params.validateTransition,
    isFinal: params.isFinal
  });
  
  // Create the channel
  const channel = {
    participants: params.participants,
    adjudicator: adjudicatorAddress,
    challenge: params.challenge || BigInt(86400), // Default: 1 day
    nonce: params.nonce || BigInt(Date.now())
  };
  
  // Create and return the channel context
  return new ChannelContext<T>(
    client,
    channel,
    appLogic,
    params.initialState
  );
}