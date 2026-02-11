import { AbiParameter } from 'abitype';

/**
 * Common ABI fragments that can be reused across different ABIs
 * These match the new contract structures from V1
 */

// Ledger tuple structure - represents token allocations and flows on a chain
export const LedgerParamFragment: AbiParameter[] = [
    { name: 'chainId', type: 'uint64' },
    { name: 'token', type: 'address' },
    { name: 'decimals', type: 'uint8' },
    { name: 'userAllocation', type: 'uint256' },
    { name: 'userNetFlow', type: 'int256' },
    { name: 'nodeAllocation', type: 'uint256' },
    { name: 'nodeNetFlow', type: 'int256' },
];

// State tuple structure - represents the channel state
export const StateParamFragment: AbiParameter[] = [
    { name: 'version', type: 'uint64' },
    { name: 'intent', type: 'uint8' }, // enum StateIntent
    { name: 'metadata', type: 'bytes32' },
    {
        name: 'homeState',
        type: 'tuple',
        components: LedgerParamFragment,
    },
    {
        name: 'nonHomeState',
        type: 'tuple',
        components: LedgerParamFragment,
    },
    { name: 'userSig', type: 'bytes' },
    { name: 'nodeSig', type: 'bytes' },
];

// ChannelDefinition tuple structure - defines channel parameters
export const ChannelDefinitionParamFragment: AbiParameter[] = [
    { name: 'challengeDuration', type: 'uint32' },
    { name: 'user', type: 'address' },
    { name: 'node', type: 'address' },
    { name: 'nonce', type: 'uint64' },
    { name: 'metadata', type: 'bytes32' },
];
