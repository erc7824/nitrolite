import { Abi } from 'abitype';

/**
 * ABI for the Adjudicator interface
 * Used to validate state transitions in state channels
 */
export const AdjudicatorAbi: Abi = [
    {
        type: 'function',
        name: 'adjudicate',
        inputs: [
            {
                name: 'chan',
                type: 'tuple',
                components: [
                    { name: 'participants', type: 'address[2]' },
                    { name: 'adjudicator', type: 'address' },
                    { name: 'challenge', type: 'uint64' },
                    { name: 'nonce', type: 'uint64' },
                ],
            },
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
                            { name: 'amount', type: 'uint256' },
                        ],
                    },
                    {
                        name: 'sigs',
                        type: 'tuple[]',
                        components: [
                            { name: 'v', type: 'uint8' },
                            { name: 'r', type: 'bytes32' },
                            { name: 's', type: 'bytes32' },
                        ],
                    },
                ],
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
                            { name: 'amount', type: 'uint256' },
                        ],
                    },
                    {
                        name: 'sigs',
                        type: 'tuple[]',
                        components: [
                            { name: 'v', type: 'uint8' },
                            { name: 'r', type: 'bytes32' },
                            { name: 's', type: 'bytes32' },
                        ],
                    },
                ],
            },
        ],
        outputs: [
            { name: 'outcome', type: 'uint8' },
            {
                name: 'allocations',
                type: 'tuple[2]',
                components: [
                    { name: 'destination', type: 'address' },
                    { name: 'token', type: 'address' },
                    { name: 'amount', type: 'uint256' },
                ],
            },
        ],
        stateMutability: 'view',
    },
];
