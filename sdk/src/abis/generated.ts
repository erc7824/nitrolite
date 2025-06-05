//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Custody
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export const custodyAbi = [
  {
    type: 'function',
    inputs: [
      { name: 'channelId', internalType: 'bytes32', type: 'bytes32' },
      {
        name: 'candidate',
        internalType: 'struct State',
        type: 'tuple',
        components: [
          { name: 'intent', internalType: 'enum StateIntent', type: 'uint8' },
          { name: 'version', internalType: 'uint256', type: 'uint256' },
          { name: 'data', internalType: 'bytes', type: 'bytes' },
          {
            name: 'allocations',
            internalType: 'struct Allocation[]',
            type: 'tuple[]',
            components: [
              { name: 'destination', internalType: 'address', type: 'address' },
              { name: 'token', internalType: 'address', type: 'address' },
              { name: 'amount', internalType: 'uint256', type: 'uint256' },
            ],
          },
          {
            name: 'sigs',
            internalType: 'struct Signature[]',
            type: 'tuple[]',
            components: [
              { name: 'v', internalType: 'uint8', type: 'uint8' },
              { name: 'r', internalType: 'bytes32', type: 'bytes32' },
              { name: 's', internalType: 'bytes32', type: 'bytes32' },
            ],
          },
        ],
      },
      {
        name: 'proofs',
        internalType: 'struct State[]',
        type: 'tuple[]',
        components: [
          { name: 'intent', internalType: 'enum StateIntent', type: 'uint8' },
          { name: 'version', internalType: 'uint256', type: 'uint256' },
          { name: 'data', internalType: 'bytes', type: 'bytes' },
          {
            name: 'allocations',
            internalType: 'struct Allocation[]',
            type: 'tuple[]',
            components: [
              { name: 'destination', internalType: 'address', type: 'address' },
              { name: 'token', internalType: 'address', type: 'address' },
              { name: 'amount', internalType: 'uint256', type: 'uint256' },
            ],
          },
          {
            name: 'sigs',
            internalType: 'struct Signature[]',
            type: 'tuple[]',
            components: [
              { name: 'v', internalType: 'uint8', type: 'uint8' },
              { name: 'r', internalType: 'bytes32', type: 'bytes32' },
              { name: 's', internalType: 'bytes32', type: 'bytes32' },
            ],
          },
        ],
      },
    ],
    name: 'challenge',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      { name: 'channelId', internalType: 'bytes32', type: 'bytes32' },
      {
        name: 'candidate',
        internalType: 'struct State',
        type: 'tuple',
        components: [
          { name: 'intent', internalType: 'enum StateIntent', type: 'uint8' },
          { name: 'version', internalType: 'uint256', type: 'uint256' },
          { name: 'data', internalType: 'bytes', type: 'bytes' },
          {
            name: 'allocations',
            internalType: 'struct Allocation[]',
            type: 'tuple[]',
            components: [
              { name: 'destination', internalType: 'address', type: 'address' },
              { name: 'token', internalType: 'address', type: 'address' },
              { name: 'amount', internalType: 'uint256', type: 'uint256' },
            ],
          },
          {
            name: 'sigs',
            internalType: 'struct Signature[]',
            type: 'tuple[]',
            components: [
              { name: 'v', internalType: 'uint8', type: 'uint8' },
              { name: 'r', internalType: 'bytes32', type: 'bytes32' },
              { name: 's', internalType: 'bytes32', type: 'bytes32' },
            ],
          },
        ],
      },
      {
        name: 'proofs',
        internalType: 'struct State[]',
        type: 'tuple[]',
        components: [
          { name: 'intent', internalType: 'enum StateIntent', type: 'uint8' },
          { name: 'version', internalType: 'uint256', type: 'uint256' },
          { name: 'data', internalType: 'bytes', type: 'bytes' },
          {
            name: 'allocations',
            internalType: 'struct Allocation[]',
            type: 'tuple[]',
            components: [
              { name: 'destination', internalType: 'address', type: 'address' },
              { name: 'token', internalType: 'address', type: 'address' },
              { name: 'amount', internalType: 'uint256', type: 'uint256' },
            ],
          },
          {
            name: 'sigs',
            internalType: 'struct Signature[]',
            type: 'tuple[]',
            components: [
              { name: 'v', internalType: 'uint8', type: 'uint8' },
              { name: 'r', internalType: 'bytes32', type: 'bytes32' },
              { name: 's', internalType: 'bytes32', type: 'bytes32' },
            ],
          },
        ],
      },
    ],
    name: 'checkpoint',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      { name: 'channelId', internalType: 'bytes32', type: 'bytes32' },
      {
        name: 'candidate',
        internalType: 'struct State',
        type: 'tuple',
        components: [
          { name: 'intent', internalType: 'enum StateIntent', type: 'uint8' },
          { name: 'version', internalType: 'uint256', type: 'uint256' },
          { name: 'data', internalType: 'bytes', type: 'bytes' },
          {
            name: 'allocations',
            internalType: 'struct Allocation[]',
            type: 'tuple[]',
            components: [
              { name: 'destination', internalType: 'address', type: 'address' },
              { name: 'token', internalType: 'address', type: 'address' },
              { name: 'amount', internalType: 'uint256', type: 'uint256' },
            ],
          },
          {
            name: 'sigs',
            internalType: 'struct Signature[]',
            type: 'tuple[]',
            components: [
              { name: 'v', internalType: 'uint8', type: 'uint8' },
              { name: 'r', internalType: 'bytes32', type: 'bytes32' },
              { name: 's', internalType: 'bytes32', type: 'bytes32' },
            ],
          },
        ],
      },
      {
        name: '',
        internalType: 'struct State[]',
        type: 'tuple[]',
        components: [
          { name: 'intent', internalType: 'enum StateIntent', type: 'uint8' },
          { name: 'version', internalType: 'uint256', type: 'uint256' },
          { name: 'data', internalType: 'bytes', type: 'bytes' },
          {
            name: 'allocations',
            internalType: 'struct Allocation[]',
            type: 'tuple[]',
            components: [
              { name: 'destination', internalType: 'address', type: 'address' },
              { name: 'token', internalType: 'address', type: 'address' },
              { name: 'amount', internalType: 'uint256', type: 'uint256' },
            ],
          },
          {
            name: 'sigs',
            internalType: 'struct Signature[]',
            type: 'tuple[]',
            components: [
              { name: 'v', internalType: 'uint8', type: 'uint8' },
              { name: 'r', internalType: 'bytes32', type: 'bytes32' },
              { name: 's', internalType: 'bytes32', type: 'bytes32' },
            ],
          },
        ],
      },
    ],
    name: 'close',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      {
        name: 'ch',
        internalType: 'struct Channel',
        type: 'tuple',
        components: [
          {
            name: 'participants',
            internalType: 'address[]',
            type: 'address[]',
          },
          { name: 'adjudicator', internalType: 'address', type: 'address' },
          { name: 'challenge', internalType: 'uint64', type: 'uint64' },
          { name: 'nonce', internalType: 'uint64', type: 'uint64' },
        ],
      },
      {
        name: 'initial',
        internalType: 'struct State',
        type: 'tuple',
        components: [
          { name: 'intent', internalType: 'enum StateIntent', type: 'uint8' },
          { name: 'version', internalType: 'uint256', type: 'uint256' },
          { name: 'data', internalType: 'bytes', type: 'bytes' },
          {
            name: 'allocations',
            internalType: 'struct Allocation[]',
            type: 'tuple[]',
            components: [
              { name: 'destination', internalType: 'address', type: 'address' },
              { name: 'token', internalType: 'address', type: 'address' },
              { name: 'amount', internalType: 'uint256', type: 'uint256' },
            ],
          },
          {
            name: 'sigs',
            internalType: 'struct Signature[]',
            type: 'tuple[]',
            components: [
              { name: 'v', internalType: 'uint8', type: 'uint8' },
              { name: 'r', internalType: 'bytes32', type: 'bytes32' },
              { name: 's', internalType: 'bytes32', type: 'bytes32' },
            ],
          },
        ],
      },
    ],
    name: 'create',
    outputs: [{ name: 'channelId', internalType: 'bytes32', type: 'bytes32' }],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      { name: 'token', internalType: 'address', type: 'address' },
      { name: 'amount', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'deposit',
    outputs: [],
    stateMutability: 'payable',
  },
  {
    type: 'function',
    inputs: [{ name: 'account', internalType: 'address', type: 'address' }],
    name: 'getAccountChannels',
    outputs: [{ name: '', internalType: 'bytes32[]', type: 'bytes32[]' }],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'user', internalType: 'address', type: 'address' },
      { name: 'token', internalType: 'address', type: 'address' },
    ],
    name: 'getAccountInfo',
    outputs: [
      { name: 'available', internalType: 'uint256', type: 'uint256' },
      { name: 'channelCount', internalType: 'uint256', type: 'uint256' },
    ],
    stateMutability: 'view',
  },
  {
    type: 'function',
    inputs: [
      { name: 'channelId', internalType: 'bytes32', type: 'bytes32' },
      { name: 'index', internalType: 'uint256', type: 'uint256' },
      {
        name: 'sig',
        internalType: 'struct Signature',
        type: 'tuple',
        components: [
          { name: 'v', internalType: 'uint8', type: 'uint8' },
          { name: 'r', internalType: 'bytes32', type: 'bytes32' },
          { name: 's', internalType: 'bytes32', type: 'bytes32' },
        ],
      },
    ],
    name: 'join',
    outputs: [{ name: '', internalType: 'bytes32', type: 'bytes32' }],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      { name: 'channelId', internalType: 'bytes32', type: 'bytes32' },
      {
        name: 'candidate',
        internalType: 'struct State',
        type: 'tuple',
        components: [
          { name: 'intent', internalType: 'enum StateIntent', type: 'uint8' },
          { name: 'version', internalType: 'uint256', type: 'uint256' },
          { name: 'data', internalType: 'bytes', type: 'bytes' },
          {
            name: 'allocations',
            internalType: 'struct Allocation[]',
            type: 'tuple[]',
            components: [
              { name: 'destination', internalType: 'address', type: 'address' },
              { name: 'token', internalType: 'address', type: 'address' },
              { name: 'amount', internalType: 'uint256', type: 'uint256' },
            ],
          },
          {
            name: 'sigs',
            internalType: 'struct Signature[]',
            type: 'tuple[]',
            components: [
              { name: 'v', internalType: 'uint8', type: 'uint8' },
              { name: 'r', internalType: 'bytes32', type: 'bytes32' },
              { name: 's', internalType: 'bytes32', type: 'bytes32' },
            ],
          },
        ],
      },
      {
        name: 'proofs',
        internalType: 'struct State[]',
        type: 'tuple[]',
        components: [
          { name: 'intent', internalType: 'enum StateIntent', type: 'uint8' },
          { name: 'version', internalType: 'uint256', type: 'uint256' },
          { name: 'data', internalType: 'bytes', type: 'bytes' },
          {
            name: 'allocations',
            internalType: 'struct Allocation[]',
            type: 'tuple[]',
            components: [
              { name: 'destination', internalType: 'address', type: 'address' },
              { name: 'token', internalType: 'address', type: 'address' },
              { name: 'amount', internalType: 'uint256', type: 'uint256' },
            ],
          },
          {
            name: 'sigs',
            internalType: 'struct Signature[]',
            type: 'tuple[]',
            components: [
              { name: 'v', internalType: 'uint8', type: 'uint8' },
              { name: 'r', internalType: 'bytes32', type: 'bytes32' },
              { name: 's', internalType: 'bytes32', type: 'bytes32' },
            ],
          },
        ],
      },
    ],
    name: 'resize',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'function',
    inputs: [
      { name: 'token', internalType: 'address', type: 'address' },
      { name: 'amount', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'withdraw',
    outputs: [],
    stateMutability: 'nonpayable',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'channelId',
        internalType: 'bytes32',
        type: 'bytes32',
        indexed: true,
      },
      {
        name: 'expiration',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'Challenged',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'channelId',
        internalType: 'bytes32',
        type: 'bytes32',
        indexed: true,
      },
    ],
    name: 'Checkpointed',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'channelId',
        internalType: 'bytes32',
        type: 'bytes32',
        indexed: true,
      },
      {
        name: 'finalState',
        internalType: 'struct State',
        type: 'tuple',
        components: [
          { name: 'intent', internalType: 'enum StateIntent', type: 'uint8' },
          { name: 'version', internalType: 'uint256', type: 'uint256' },
          { name: 'data', internalType: 'bytes', type: 'bytes' },
          {
            name: 'allocations',
            internalType: 'struct Allocation[]',
            type: 'tuple[]',
            components: [
              { name: 'destination', internalType: 'address', type: 'address' },
              { name: 'token', internalType: 'address', type: 'address' },
              { name: 'amount', internalType: 'uint256', type: 'uint256' },
            ],
          },
          {
            name: 'sigs',
            internalType: 'struct Signature[]',
            type: 'tuple[]',
            components: [
              { name: 'v', internalType: 'uint8', type: 'uint8' },
              { name: 'r', internalType: 'bytes32', type: 'bytes32' },
              { name: 's', internalType: 'bytes32', type: 'bytes32' },
            ],
          },
        ],
        indexed: false,
      },
    ],
    name: 'Closed',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'channelId',
        internalType: 'bytes32',
        type: 'bytes32',
        indexed: true,
      },
      {
        name: 'wallet',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'channel',
        internalType: 'struct Channel',
        type: 'tuple',
        components: [
          {
            name: 'participants',
            internalType: 'address[]',
            type: 'address[]',
          },
          { name: 'adjudicator', internalType: 'address', type: 'address' },
          { name: 'challenge', internalType: 'uint64', type: 'uint64' },
          { name: 'nonce', internalType: 'uint64', type: 'uint64' },
        ],
        indexed: false,
      },
      {
        name: 'initial',
        internalType: 'struct State',
        type: 'tuple',
        components: [
          { name: 'intent', internalType: 'enum StateIntent', type: 'uint8' },
          { name: 'version', internalType: 'uint256', type: 'uint256' },
          { name: 'data', internalType: 'bytes', type: 'bytes' },
          {
            name: 'allocations',
            internalType: 'struct Allocation[]',
            type: 'tuple[]',
            components: [
              { name: 'destination', internalType: 'address', type: 'address' },
              { name: 'token', internalType: 'address', type: 'address' },
              { name: 'amount', internalType: 'uint256', type: 'uint256' },
            ],
          },
          {
            name: 'sigs',
            internalType: 'struct Signature[]',
            type: 'tuple[]',
            components: [
              { name: 'v', internalType: 'uint8', type: 'uint8' },
              { name: 'r', internalType: 'bytes32', type: 'bytes32' },
              { name: 's', internalType: 'bytes32', type: 'bytes32' },
            ],
          },
        ],
        indexed: false,
      },
    ],
    name: 'Created',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'wallet',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'token',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'amount',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'Deposited',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'channelId',
        internalType: 'bytes32',
        type: 'bytes32',
        indexed: true,
      },
      {
        name: 'index',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'Joined',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'channelId',
        internalType: 'bytes32',
        type: 'bytes32',
        indexed: true,
      },
    ],
    name: 'Opened',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'channelId',
        internalType: 'bytes32',
        type: 'bytes32',
        indexed: true,
      },
      {
        name: 'deltaAllocations',
        internalType: 'int256[]',
        type: 'int256[]',
        indexed: false,
      },
    ],
    name: 'Resized',
  },
  {
    type: 'event',
    anonymous: false,
    inputs: [
      {
        name: 'wallet',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'token',
        internalType: 'address',
        type: 'address',
        indexed: true,
      },
      {
        name: 'amount',
        internalType: 'uint256',
        type: 'uint256',
        indexed: false,
      },
    ],
    name: 'Withdrawn',
  },
  { type: 'error', inputs: [], name: 'ChallengeNotExpired' },
  { type: 'error', inputs: [], name: 'ChannelNotFinal' },
  {
    type: 'error',
    inputs: [{ name: 'channelId', internalType: 'bytes32', type: 'bytes32' }],
    name: 'ChannelNotFound',
  },
  { type: 'error', inputs: [], name: 'ECDSAInvalidSignature' },
  {
    type: 'error',
    inputs: [{ name: 'length', internalType: 'uint256', type: 'uint256' }],
    name: 'ECDSAInvalidSignatureLength',
  },
  {
    type: 'error',
    inputs: [{ name: 's', internalType: 'bytes32', type: 'bytes32' }],
    name: 'ECDSAInvalidSignatureS',
  },
  {
    type: 'error',
    inputs: [
      { name: 'available', internalType: 'uint256', type: 'uint256' },
      { name: 'required', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'InsufficientBalance',
  },
  { type: 'error', inputs: [], name: 'InvalidAdjudicator' },
  { type: 'error', inputs: [], name: 'InvalidAllocations' },
  { type: 'error', inputs: [], name: 'InvalidAmount' },
  { type: 'error', inputs: [], name: 'InvalidChallengePeriod' },
  { type: 'error', inputs: [], name: 'InvalidParticipant' },
  { type: 'error', inputs: [], name: 'InvalidState' },
  { type: 'error', inputs: [], name: 'InvalidStateSignatures' },
  { type: 'error', inputs: [], name: 'InvalidStatus' },
  { type: 'error', inputs: [], name: 'InvalidValue' },
  {
    type: 'error',
    inputs: [{ name: 'token', internalType: 'address', type: 'address' }],
    name: 'SafeERC20FailedOperation',
  },
  {
    type: 'error',
    inputs: [
      { name: 'token', internalType: 'address', type: 'address' },
      { name: 'to', internalType: 'address', type: 'address' },
      { name: 'amount', internalType: 'uint256', type: 'uint256' },
    ],
    name: 'TransferFailed',
  },
] as const
