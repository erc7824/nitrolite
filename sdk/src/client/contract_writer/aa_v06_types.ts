import { Hex, Prettify, PartialBy, Address, Log, RpcUserOperation } from 'viem';
import { UserOperation } from 'viem/account-abstraction';

export type PartialUserOperationV06 = Prettify<
    PartialBy<
        Required<UserOperation<'0.6'>>,
        | 'callGasLimit'
        | 'maxFeePerGas'
        | 'maxPriorityFeePerGas'
        | 'paymasterAndData'
        | 'preVerificationGas'
        | 'verificationGasLimit'
        | 'authorization'
    >
>;

export type GasParametersV06 = {
    callGasLimit: Hex;
    verificationGasLimit: Hex;
    preVerificationGas: Hex;
    paymasterAndData: Hex;
    maxFeePerGas: Hex;
    maxPriorityFeePerGas: Hex;
};

export interface BundlerClientV06 {
    estimateUserOperation(chainId: number, userOp: PartialUserOperationV06): Promise<GasParametersV06>;
    sendUserOperation(chainId: number, userOp: RpcUserOperation<'0.6'>): Promise<Hex>;
    fetchLogs(chainId: number, addresses: Address[], topics: Hex[][]): Promise<Log[]>;
}
