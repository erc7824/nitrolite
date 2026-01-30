import { AbiEvent } from 'viem';

export const UserOpEventAbi: AbiEvent = {
    anonymous: false,
    inputs: [
        {
            indexed: true,
            internalType: 'bytes32',
            name: 'userOpHash',
            type: 'bytes32',
        },
        {
            indexed: true,
            internalType: 'address',
            name: 'sender',
            type: 'address',
        },
        {
            indexed: true,
            internalType: 'address',
            name: 'paymaster',
            type: 'address',
        },
        {
            indexed: false,
            internalType: 'uint256',
            name: 'nonce',
            type: 'uint256',
        },
        {
            indexed: false,
            internalType: 'bool',
            name: 'success',
            type: 'bool',
        },
        {
            indexed: false,
            internalType: 'uint256',
            name: 'actualGasCost',
            type: 'uint256',
        },
        {
            indexed: false,
            internalType: 'uint256',
            name: 'actualGasUsed',
            type: 'uint256',
        },
    ],
    name: 'UserOperationEvent',
    type: 'event',
};
