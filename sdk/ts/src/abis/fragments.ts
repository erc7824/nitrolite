import { AbiParameter } from 'abitype';

/**
 * Common ABI fragments that can be reused across different ABIs
 */

// Channel tuple structure
export const ChannelParamFragment: AbiParameter[] = [
    { name: 'participants', type: 'address[2]' },
    { name: 'adjudicator', type: 'address' },
    { name: 'challenge', type: 'uint64' },
    { name: 'nonce', type: 'uint64' },
];

// Allocation tuple structure
export const AllocationParamFragment: AbiParameter[] = [
    { name: 'destination', type: 'address' },
    { name: 'token', type: 'address' },
    { name: 'amount', type: 'uint256' },
];

// Signature tuple structure
export const SignatureParamFragment: AbiParameter[] = [
    { name: 'v', type: 'uint8' },
    { name: 'r', type: 'bytes32' },
    { name: 's', type: 'bytes32' },
];

// State tuple structure
export const StateParamFragment: AbiParameter[] = [
    { name: 'data', type: 'bytes' },
    {
        name: 'allocations',
        type: 'tuple[2]',
        components: AllocationParamFragment,
    },
    {
        name: 'sigs',
        type: 'tuple[]',
        components: SignatureParamFragment,
    },
];
