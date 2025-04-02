import { Abi } from 'abitype';

export const ChannelOpenedEvent = "ChannelOpened";
export const ChannelClosedEvent = "ChannelClosed";
export const ChannelChallengedEvent = "ChannelChallenged";
export const ChannelCheckpointedEvent = "ChannelCheckpointed";

/**
 * ABI for the Custody contract
 * Manages the lifecycle of state channels
 */
export const CustodyAbi: Abi = [
  // Functions
  {
    type: 'function',
    name: 'open',
    inputs: [
      {
        name: 'ch',
        type: 'tuple',
        components: [
          { name: 'participants', type: 'address[2]' },
          { name: 'adjudicator', type: 'address' },
          { name: 'challenge', type: 'uint64' },
          { name: 'nonce', type: 'uint64' }
        ]
      },
      {
        name: 'deposit',
        type: 'tuple',
        components: [
          { name: 'data', type: 'bytes' },
          {
            name: 'allocations',
            type: 'tuple[2]',
            components: [
              { name: 'destination', type: 'address' },
              { name: 'token', type: 'address' },
              { name: 'amount', type: 'uint256' }
            ]
          },
          {
            name: 'sigs',
            type: 'tuple[]',
            components: [
              { name: 'v', type: 'uint8' },
              { name: 'r', type: 'bytes32' },
              { name: 's', type: 'bytes32' }
            ]
          }
        ]
      }
    ],
    outputs: [{ name: 'channelId', type: 'bytes32' }],
    stateMutability: 'nonpayable'
  },
  {
    type: 'function',
    name: 'close',
    inputs: [
      { name: 'channelId', type: 'bytes32' },
      {
        name: 'candidate',
        type: 'tuple',
        components: [
          { name: 'data', type: 'bytes' },
          {
            name: 'allocations',
            type: 'tuple[2]',
            components: [
              { name: 'destination', type: 'address' },
              { name: 'token', type: 'address' },
              { name: 'amount', type: 'uint256' }
            ]
          },
          {
            name: 'sigs',
            type: 'tuple[]',
            components: [
              { name: 'v', type: 'uint8' },
              { name: 'r', type: 'bytes32' },
              { name: 's', type: 'bytes32' }
            ]
          }
        ]
      },
      {
        name: 'proofs',
        type: 'tuple[]',
        components: [
          { name: 'data', type: 'bytes' },
          {
            name: 'allocations',
            type: 'tuple[2]',
            components: [
              { name: 'destination', type: 'address' },
              { name: 'token', type: 'address' },
              { name: 'amount', type: 'uint256' }
            ]
          },
          {
            name: 'sigs',
            type: 'tuple[]',
            components: [
              { name: 'v', type: 'uint8' },
              { name: 'r', type: 'bytes32' },
              { name: 's', type: 'bytes32' }
            ]
          }
        ]
      }
    ],
    outputs: [],
    stateMutability: 'nonpayable'
  },
  {
    type: 'function',
    name: 'challenge',
    inputs: [
      { name: 'channelId', type: 'bytes32' },
      {
        name: 'candidate',
        type: 'tuple',
        components: [
          { name: 'data', type: 'bytes' },
          {
            name: 'allocations',
            type: 'tuple[2]',
            components: [
              { name: 'destination', type: 'address' },
              { name: 'token', type: 'address' },
              { name: 'amount', type: 'uint256' }
            ]
          },
          {
            name: 'sigs',
            type: 'tuple[]',
            components: [
              { name: 'v', type: 'uint8' },
              { name: 'r', type: 'bytes32' },
              { name: 's', type: 'bytes32' }
            ]
          }
        ]
      },
      {
        name: 'proofs',
        type: 'tuple[]',
        components: [
          { name: 'data', type: 'bytes' },
          {
            name: 'allocations',
            type: 'tuple[2]',
            components: [
              { name: 'destination', type: 'address' },
              { name: 'token', type: 'address' },
              { name: 'amount', type: 'uint256' }
            ]
          },
          {
            name: 'sigs',
            type: 'tuple[]',
            components: [
              { name: 'v', type: 'uint8' },
              { name: 'r', type: 'bytes32' },
              { name: 's', type: 'bytes32' }
            ]
          }
        ]
      }
    ],
    outputs: [],
    stateMutability: 'nonpayable'
  },
  {
    type: 'function',
    name: 'checkpoint',
    inputs: [
      { name: 'channelId', type: 'bytes32' },
      {
        name: 'candidate',
        type: 'tuple',
        components: [
          { name: 'data', type: 'bytes' },
          {
            name: 'allocations',
            type: 'tuple[2]',
            components: [
              { name: 'destination', type: 'address' },
              { name: 'token', type: 'address' },
              { name: 'amount', type: 'uint256' }
            ]
          },
          {
            name: 'sigs',
            type: 'tuple[]',
            components: [
              { name: 'v', type: 'uint8' },
              { name: 'r', type: 'bytes32' },
              { name: 's', type: 'bytes32' }
            ]
          }
        ]
      },
      {
        name: 'proofs',
        type: 'tuple[]',
        components: [
          { name: 'data', type: 'bytes' },
          {
            name: 'allocations',
            type: 'tuple[2]',
            components: [
              { name: 'destination', type: 'address' },
              { name: 'token', type: 'address' },
              { name: 'amount', type: 'uint256' }
            ]
          },
          {
            name: 'sigs',
            type: 'tuple[]',
            components: [
              { name: 'v', type: 'uint8' },
              { name: 'r', type: 'bytes32' },
              { name: 's', type: 'bytes32' }
            ]
          }
        ]
      }
    ],
    outputs: [],
    stateMutability: 'nonpayable'
  },
  {
    type: 'function',
    name: 'reclaim',
    inputs: [{ name: 'channelId', type: 'bytes32' }],
    outputs: [],
    stateMutability: 'nonpayable'
  },
  
  // Events
  {
    type: 'event',
    name: 'ChannelOpened',
    inputs: [
      { indexed: true, name: 'channelId', type: 'bytes32' },
      {
        indexed: false,
        name: 'channel',
        type: 'tuple',
        components: [
          { name: 'participants', type: 'address[2]' },
          { name: 'adjudicator', type: 'address' },
          { name: 'challenge', type: 'uint64' },
          { name: 'nonce', type: 'uint64' }
        ]
      }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'ChannelChallenged',
    inputs: [
      { indexed: true, name: 'channelId', type: 'bytes32' },
      { indexed: false, name: 'expiration', type: 'uint256' }
    ],
    anonymous: false
  },
  {
    type: 'event',
    name: 'ChannelCheckpointed',
    inputs: [{ indexed: true, name: 'channelId', type: 'bytes32' }],
    anonymous: false
  },
  {
    type: 'event',
    name: 'ChannelClosed',
    inputs: [{ indexed: true, name: 'channelId', type: 'bytes32' }],
    anonymous: false
  }
];