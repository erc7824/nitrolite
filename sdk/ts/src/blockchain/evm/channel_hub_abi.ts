/**
 * ChannelHub contract ABI
 * Generated from contracts/src/ChannelHub.sol
 */

import { Abi } from 'abitype';

export const ChannelHubAbi = [
  // Constants
  {
    type: 'function',
    name: 'MIN_CHALLENGE_DURATION',
    inputs: [],
    outputs: [{ name: '', type: 'uint32' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    name: 'ESCROW_DEPOSIT_UNLOCK_DELAY',
    inputs: [],
    outputs: [{ name: '', type: 'uint32' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    name: 'MAX_DEPOSIT_ESCROW_PURGE',
    inputs: [],
    outputs: [{ name: '', type: 'uint32' }],
    stateMutability: 'view',
  },

  // IVault functions
  {
    type: 'function',
    name: 'depositToVault',
    inputs: [
      { name: 'node', type: 'address' },
      { name: 'token', type: 'address' },
      { name: 'amount', type: 'uint256' },
    ],
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    name: 'withdrawFromVault',
    inputs: [
      { name: 'node', type: 'address' },
      { name: 'token', type: 'address' },
      { name: 'amount', type: 'uint256' },
    ],
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    name: 'getAccountBalance',
    inputs: [
      { name: 'account', type: 'address' },
      { name: 'token', type: 'address' },
    ],
    outputs: [{ name: '', type: 'uint256' }],
    stateMutability: 'view',
  },

  // Channel lifecycle functions
  {
    type: 'function',
    name: 'createChannel',
    inputs: [
      {
        name: 'definition',
        type: 'tuple',
        components: [
          { name: 'challengeDuration', type: 'uint32' },
          { name: 'user', type: 'address' },
          { name: 'node', type: 'address' },
          { name: 'nonce', type: 'uint64' },
          { name: 'metadata', type: 'bytes32' },
        ],
      },
      {
        name: 'state',
        type: 'tuple',
        components: [
          { name: 'version', type: 'uint64' },
          { name: 'intent', type: 'uint8' },
          { name: 'metadata', type: 'bytes32' },
          {
            name: 'homeState',
            type: 'tuple',
            components: [
              { name: 'chainId', type: 'uint64' },
              { name: 'token', type: 'address' },
              { name: 'decimals', type: 'uint8' },
              { name: 'userAllocation', type: 'uint256' },
              { name: 'userNetFlow', type: 'int256' },
              { name: 'nodeAllocation', type: 'uint256' },
              { name: 'nodeNetFlow', type: 'int256' },
            ],
          },
          {
            name: 'nonHomeState',
            type: 'tuple',
            components: [
              { name: 'chainId', type: 'uint64' },
              { name: 'token', type: 'address' },
              { name: 'decimals', type: 'uint8' },
              { name: 'userAllocation', type: 'uint256' },
              { name: 'userNetFlow', type: 'int256' },
              { name: 'nodeAllocation', type: 'uint256' },
              { name: 'nodeNetFlow', type: 'int256' },
            ],
          },
          { name: 'userSig', type: 'bytes' },
          { name: 'nodeSig', type: 'bytes' },
        ],
      },
    ],
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    name: 'checkpointChannel',
    inputs: [
      { name: 'channelId', type: 'bytes32' },
      {
        name: 'candidate',
        type: 'tuple',
        components: [
          { name: 'version', type: 'uint64' },
          { name: 'intent', type: 'uint8' },
          { name: 'metadata', type: 'bytes32' },
          {
            name: 'homeState',
            type: 'tuple',
            components: [
              { name: 'chainId', type: 'uint64' },
              { name: 'token', type: 'address' },
              { name: 'decimals', type: 'uint8' },
              { name: 'userAllocation', type: 'uint256' },
              { name: 'userNetFlow', type: 'int256' },
              { name: 'nodeAllocation', type: 'uint256' },
              { name: 'nodeNetFlow', type: 'int256' },
            ],
          },
          {
            name: 'nonHomeState',
            type: 'tuple',
            components: [
              { name: 'chainId', type: 'uint64' },
              { name: 'token', type: 'address' },
              { name: 'decimals', type: 'uint8' },
              { name: 'userAllocation', type: 'uint256' },
              { name: 'userNetFlow', type: 'int256' },
              { name: 'nodeAllocation', type: 'uint256' },
              { name: 'nodeNetFlow', type: 'int256' },
            ],
          },
          { name: 'userSig', type: 'bytes' },
          { name: 'nodeSig', type: 'bytes' },
        ],
      },
      { name: 'proofs', type: 'tuple[]', components: [] }, // State[] - simplified
    ],
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    name: 'depositToChannel',
    inputs: [
      { name: 'channelId', type: 'bytes32' },
      {
        name: 'candidate',
        type: 'tuple',
        components: [
          { name: 'version', type: 'uint64' },
          { name: 'intent', type: 'uint8' },
          { name: 'metadata', type: 'bytes32' },
          {
            name: 'homeState',
            type: 'tuple',
            components: [
              { name: 'chainId', type: 'uint64' },
              { name: 'token', type: 'address' },
              { name: 'decimals', type: 'uint8' },
              { name: 'userAllocation', type: 'uint256' },
              { name: 'userNetFlow', type: 'int256' },
              { name: 'nodeAllocation', type: 'uint256' },
              { name: 'nodeNetFlow', type: 'int256' },
            ],
          },
          {
            name: 'nonHomeState',
            type: 'tuple',
            components: [
              { name: 'chainId', type: 'uint64' },
              { name: 'token', type: 'address' },
              { name: 'decimals', type: 'uint8' },
              { name: 'userAllocation', type: 'uint256' },
              { name: 'userNetFlow', type: 'int256' },
              { name: 'nodeAllocation', type: 'uint256' },
              { name: 'nodeNetFlow', type: 'int256' },
            ],
          },
          { name: 'userSig', type: 'bytes' },
          { name: 'nodeSig', type: 'bytes' },
        ],
      },
    ],
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    name: 'withdrawFromChannel',
    inputs: [
      { name: 'channelId', type: 'bytes32' },
      {
        name: 'candidate',
        type: 'tuple',
        components: [
          { name: 'version', type: 'uint64' },
          { name: 'intent', type: 'uint8' },
          { name: 'metadata', type: 'bytes32' },
          {
            name: 'homeState',
            type: 'tuple',
            components: [
              { name: 'chainId', type: 'uint64' },
              { name: 'token', type: 'address' },
              { name: 'decimals', type: 'uint8' },
              { name: 'userAllocation', type: 'uint256' },
              { name: 'userNetFlow', type: 'int256' },
              { name: 'nodeAllocation', type: 'uint256' },
              { name: 'nodeNetFlow', type: 'int256' },
            ],
          },
          {
            name: 'nonHomeState',
            type: 'tuple',
            components: [
              { name: 'chainId', type: 'uint64' },
              { name: 'token', type: 'address' },
              { name: 'decimals', type: 'uint8' },
              { name: 'userAllocation', type: 'uint256' },
              { name: 'userNetFlow', type: 'int256' },
              { name: 'nodeAllocation', type: 'uint256' },
              { name: 'nodeNetFlow', type: 'int256' },
            ],
          },
          { name: 'userSig', type: 'bytes' },
          { name: 'nodeSig', type: 'bytes' },
        ],
      },
    ],
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    name: 'challengeChannel',
    inputs: [
      { name: 'channelId', type: 'bytes32' },
      {
        name: 'candidate',
        type: 'tuple',
        components: [
          { name: 'version', type: 'uint64' },
          { name: 'intent', type: 'uint8' },
          { name: 'metadata', type: 'bytes32' },
          {
            name: 'homeState',
            type: 'tuple',
            components: [
              { name: 'chainId', type: 'uint64' },
              { name: 'token', type: 'address' },
              { name: 'decimals', type: 'uint8' },
              { name: 'userAllocation', type: 'uint256' },
              { name: 'userNetFlow', type: 'int256' },
              { name: 'nodeAllocation', type: 'uint256' },
              { name: 'nodeNetFlow', type: 'int256' },
            ],
          },
          {
            name: 'nonHomeState',
            type: 'tuple',
            components: [
              { name: 'chainId', type: 'uint64' },
              { name: 'token', type: 'address' },
              { name: 'decimals', type: 'uint8' },
              { name: 'userAllocation', type: 'uint256' },
              { name: 'userNetFlow', type: 'int256' },
              { name: 'nodeAllocation', type: 'uint256' },
              { name: 'nodeNetFlow', type: 'int256' },
            ],
          },
          { name: 'userSig', type: 'bytes' },
          { name: 'nodeSig', type: 'bytes' },
        ],
      },
      { name: 'proofs', type: 'tuple[]', components: [] }, // State[]
      { name: 'challengerSig', type: 'bytes' },
    ],
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    name: 'closeChannel',
    inputs: [
      { name: 'channelId', type: 'bytes32' },
      {
        name: 'candidate',
        type: 'tuple',
        components: [
          { name: 'version', type: 'uint64' },
          { name: 'intent', type: 'uint8' },
          { name: 'metadata', type: 'bytes32' },
          {
            name: 'homeState',
            type: 'tuple',
            components: [
              { name: 'chainId', type: 'uint64' },
              { name: 'token', type: 'address' },
              { name: 'decimals', type: 'uint8' },
              { name: 'userAllocation', type: 'uint256' },
              { name: 'userNetFlow', type: 'int256' },
              { name: 'nodeAllocation', type: 'uint256' },
              { name: 'nodeNetFlow', type: 'int256' },
            ],
          },
          {
            name: 'nonHomeState',
            type: 'tuple',
            components: [
              { name: 'chainId', type: 'uint64' },
              { name: 'token', type: 'address' },
              { name: 'decimals', type: 'uint8' },
              { name: 'userAllocation', type: 'uint256' },
              { name: 'userNetFlow', type: 'int256' },
              { name: 'nodeAllocation', type: 'uint256' },
              { name: 'nodeNetFlow', type: 'int256' },
            ],
          },
          { name: 'userSig', type: 'bytes' },
          { name: 'nodeSig', type: 'bytes' },
        ],
      },
      { name: 'proofs', type: 'tuple[]', components: [] }, // State[]
    ],
    outputs: [],
    stateMutability: 'nonpayable',
  },

  // Getter functions
  {
    type: 'function',
    name: 'getOpenChannels',
    inputs: [{ name: 'user', type: 'address' }],
    outputs: [{ name: '', type: 'bytes32[]' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    name: 'getChannelData',
    inputs: [{ name: 'channelId', type: 'bytes32' }],
    outputs: [
      {
        name: '',
        type: 'tuple',
        components: [
          {
            name: 'definition',
            type: 'tuple',
            components: [
              { name: 'challengeDuration', type: 'uint32' },
              { name: 'user', type: 'address' },
              { name: 'node', type: 'address' },
              { name: 'nonce', type: 'uint64' },
              { name: 'metadata', type: 'bytes32' },
            ],
          },
          {
            name: 'lastState',
            type: 'tuple',
            components: [
              { name: 'version', type: 'uint64' },
              { name: 'intent', type: 'uint8' },
              { name: 'metadata', type: 'bytes32' },
              {
                name: 'homeState',
                type: 'tuple',
                components: [
                  { name: 'chainId', type: 'uint64' },
                  { name: 'token', type: 'address' },
                  { name: 'decimals', type: 'uint8' },
                  { name: 'userAllocation', type: 'uint256' },
                  { name: 'userNetFlow', type: 'int256' },
                  { name: 'nodeAllocation', type: 'uint256' },
                  { name: 'nodeNetFlow', type: 'int256' },
                ],
              },
              {
                name: 'nonHomeState',
                type: 'tuple',
                components: [
                  { name: 'chainId', type: 'uint64' },
                  { name: 'token', type: 'address' },
                  { name: 'decimals', type: 'uint8' },
                  { name: 'userAllocation', type: 'uint256' },
                  { name: 'userNetFlow', type: 'int256' },
                  { name: 'nodeAllocation', type: 'uint256' },
                  { name: 'nodeNetFlow', type: 'int256' },
                ],
              },
              { name: 'userSig', type: 'bytes' },
              { name: 'nodeSig', type: 'bytes' },
            ],
          },
          { name: 'challengeExpiry', type: 'uint64' },
        ],
      },
    ],
    stateMutability: 'view',
  },
] as const satisfies Abi;
