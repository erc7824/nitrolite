import { Abi, Account, Address, Chain, ContractFunctionParameters, Hex } from 'viem';

export interface ContractWriter {
    write: (callDetails: CallsDetails) => Promise<WriteResult>;
}

export interface ContractCallParams<
    TAbi extends Abi = Abi,
    TFunctionName extends string = string,
    TChain extends Chain | undefined = Chain | undefined,
> {
    // Required parameters
    address: Address;
    abi: TAbi;
    functionName: TFunctionName;

    // Optional parameters
    args?: readonly unknown[];
    account?: Account | Address;
    chain?: TChain;

    // Transaction parameters
    value?: bigint;
    gas?: bigint;
    // gasPrice?: bigint;
    maxFeePerGas?: bigint;
    maxPriorityFeePerGas?: bigint;
    nonce?: number;

    // Additional optional parameters
    dataSuffix?: `0x${string}`;
    // type?: 'legacy' | 'eip2930' | 'eip1559';
    type?: "eip7702";
}

export interface CallsDetails {
    calls: ContractCallParams[];
}

export interface WriteResult {
    txHashes: Hex[];
}
