import {
    Account,
    Call,
    encodeFunctionData,
    ExactPartial,
    Hex,
    numberToHex,
    pad,
    publicActions,
    PublicClient,
    RpcUserOperation,
    SignedAuthorization,
    toEventSelector,
} from 'viem';
import { CallsDetails, ContractCallParams, ContractWriter, WriteResult } from './types';
import { entryPoint06Address, SmartAccount, UserOperation } from 'viem/account-abstraction';
import { AAExecuteAbi } from '../../abis/aa/execute';
import { BundlerClientV06, PartialUserOperationV06 } from './aa_v06_types';
import { UserOpEventAbi } from '../../abis/aa/user_op_event';
import Errors from '../../errors';

export type AAV06ContractWriterConfig = {
    publicClient: PublicClient;
    smartAccount: SmartAccount;
    bundlerClient: BundlerClientV06;

    pollingInterval?: number;
    pollingTimeout?: number;
};

export class AAV06ContractWriter implements ContractWriter {
    public readonly publicClient: PublicClient;
    public readonly smartAccount: SmartAccount;
    public readonly bundlerClient: BundlerClientV06;

    private readonly pollingInterval: number;
    private readonly pollingTimeout: number;

    constructor(config: AAV06ContractWriterConfig) {
        if (!config.publicClient) throw new Errors.MissingParameterError('publicClient');
        if (!config.smartAccount) throw new Errors.MissingParameterError('smartAccount');
        if (!config.bundlerClient) throw new Errors.MissingParameterError('bundlerClient');

        this.publicClient = config.publicClient;
        this.smartAccount = config.smartAccount;
        this.bundlerClient = config.bundlerClient;

        this.pollingInterval = config.pollingInterval ?? 5000;
        this.pollingTimeout = config.pollingTimeout ?? 120000;
    }

    async write(callsDetails: CallsDetails): Promise<WriteResult> {
        const calls = callsDetails.calls.map((call) => this._prepareCalldata(call));

        const txHash = await this._writeCalls(calls);
        return { txHashes: [txHash] };
    }

    getAccount(): Account {
        return this.smartAccount;
    }

    private _prepareCalldata(callParams: ContractCallParams): Call {
        const encoded = encodeFunctionData({
            abi: callParams.abi,
            functionName: callParams.functionName,
            args: callParams.args,
        });

        return {
            to: callParams.address,
            value: callParams.value ?? 0n,
            data: encoded,
        };
    }

    private async _writeCalls(calls: Call[]): Promise<Hex> {
        const chainId = await this.publicClient.getChainId();

        const partialUserOperation = await this._callsToPartialUserOperation(calls);
        const gasParameters = await this.bundlerClient.estimateUserOperation(chainId, partialUserOperation);

        const userOperation = this._formatUserOperation(
            // @ts-ignore
            {
                ...partialUserOperation,
                nonce: ('0x' + partialUserOperation.nonce.toString(16)) as Hex,
                ...gasParameters,
            },
        ) as Required<UserOperation<'0.6'>>;

        userOperation.signature = await this.smartAccount.signUserOperation({
            chainId,
            ...userOperation,
        });

        const userOperationSerialized = this._formatUserOperationRequest(userOperation);
        const userOpHash = await this.bundlerClient.sendUserOperation(chainId, userOperationSerialized);

        return await this._waitForUserOperationReceipt(userOpHash);
    }

    private async _callsToPartialUserOperation(calls: Call[]): Promise<PartialUserOperationV06> {
        const senderAddress = this.smartAccount.address;
        const nonce = await this.smartAccount.getNonce();

        let initCode: Hex = '0x';

        if (!(await this.smartAccount.isDeployed())) {
            // NOTE: for EntryPoint v0.6, the initCode is the factoryData
            const { factory, factoryData } = await this.smartAccount.getFactoryArgs();

            if (factory && factoryData) {
                initCode = (factory + factoryData.substring(2)) as Hex;
            } else {
                throw new Error('SmartAccount factory is not configured properly');
            }
        }

        const partialUserOperation: PartialUserOperationV06 = {
            sender: senderAddress,
            nonce,
            initCode,
            // NOTE: not using SmartWallet interface `encodeCalls`, as we are not fully satisfied with Kernel's implementation
            // Please, change if we change SW provider
            callData: this._encodeExecuteBatchCall(calls),
            paymasterAndData: '0x',
            signature: '0x',
        };

        partialUserOperation.signature = await this.smartAccount.getStubSignature(partialUserOperation);

        return partialUserOperation;
    }

    private async _waitForUserOperationReceipt(userOpHash: Hex): Promise<Hex> {
        const startTime = Date.now();

        return new Promise<Hex>((resolve, reject) => {
            const intervalId = setInterval(async () => {
                try {
                    const chainId = this.publicClient.chain?.id;
                    if (!chainId) {
                        clearInterval(intervalId);
                        reject(new Error('PublicClient chain is not configured'));
                        return;
                    }

                    const logs = await this.bundlerClient.fetchLogs(
                        this.publicClient.chain!.id,
                        [entryPoint06Address],
                        [[toEventSelector(UserOpEventAbi)], [userOpHash]],
                    );

                    if (logs.length > 0) {
                        const txHash = logs[logs.length - 1].transactionHash;

                        if (txHash) {
                            clearInterval(intervalId);
                            resolve(txHash);
                            return;
                        }
                    }

                    if (Date.now() - startTime >= this.pollingTimeout) {
                        clearInterval(intervalId);
                        const waitTimeoutError = new Error(
                            `Timeout for waiting UserOperationEvent. Waited for: ` + this.pollingTimeout + ' ms',
                        );

                        reject(waitTimeoutError);
                        return;
                    }
                } catch (error) {
                    clearInterval(intervalId);
                    reject(error);
                }
            }, this.pollingInterval);
        });
    }

    private _encodeExecuteBatchCall = (args: readonly Call[]) => {
        return encodeFunctionData({
            abi: AAExecuteAbi,
            functionName: 'executeBatch',
            args: [
                args.map((arg) => {
                    return {
                        to: arg.to,
                        value: arg.value || 0n,
                        data: arg.data || '0x',
                    };
                }),
            ],
        });
    };

    private _formatUserOperation = (parameters: RpcUserOperation): UserOperation => {
        const userOperation = { ...parameters } as unknown as UserOperation;

        if (parameters.callGasLimit) userOperation.callGasLimit = BigInt(parameters.callGasLimit);
        if (parameters.maxFeePerGas) userOperation.maxFeePerGas = BigInt(parameters.maxFeePerGas);
        if (parameters.maxPriorityFeePerGas)
            userOperation.maxPriorityFeePerGas = BigInt(parameters.maxPriorityFeePerGas);
        if (parameters.nonce) userOperation.nonce = BigInt(parameters.nonce);
        if (parameters.paymasterPostOpGasLimit)
            userOperation.paymasterPostOpGasLimit = BigInt(parameters.paymasterPostOpGasLimit);
        if (parameters.paymasterVerificationGasLimit)
            userOperation.paymasterVerificationGasLimit = BigInt(parameters.paymasterVerificationGasLimit);
        if (parameters.preVerificationGas) userOperation.preVerificationGas = BigInt(parameters.preVerificationGas);
        if (parameters.verificationGasLimit)
            userOperation.verificationGasLimit = BigInt(parameters.verificationGasLimit);

        return userOperation;
    };

    private _formatUserOperationRequest = (request: ExactPartial<UserOperation>) => {
        const rpcRequest = {} as RpcUserOperation;

        if (typeof request.callData !== 'undefined') rpcRequest.callData = request.callData;
        if (typeof request.callGasLimit !== 'undefined') rpcRequest.callGasLimit = numberToHex(request.callGasLimit);
        if (typeof request.factory !== 'undefined') rpcRequest.factory = request.factory;
        if (typeof request.factoryData !== 'undefined') rpcRequest.factoryData = request.factoryData;
        if (typeof request.initCode !== 'undefined') rpcRequest.initCode = request.initCode;
        if (typeof request.maxFeePerGas !== 'undefined') rpcRequest.maxFeePerGas = numberToHex(request.maxFeePerGas);
        if (typeof request.maxPriorityFeePerGas !== 'undefined')
            rpcRequest.maxPriorityFeePerGas = numberToHex(request.maxPriorityFeePerGas);
        if (typeof request.nonce !== 'undefined') rpcRequest.nonce = numberToHex(request.nonce);
        if (typeof request.paymaster !== 'undefined') rpcRequest.paymaster = request.paymaster;
        if (typeof request.paymasterAndData !== 'undefined')
            rpcRequest.paymasterAndData = request.paymasterAndData || '0x';
        if (typeof request.paymasterData !== 'undefined') rpcRequest.paymasterData = request.paymasterData;
        if (typeof request.paymasterPostOpGasLimit !== 'undefined')
            rpcRequest.paymasterPostOpGasLimit = numberToHex(request.paymasterPostOpGasLimit);
        if (typeof request.paymasterVerificationGasLimit !== 'undefined')
            rpcRequest.paymasterVerificationGasLimit = numberToHex(request.paymasterVerificationGasLimit);
        if (typeof request.preVerificationGas !== 'undefined')
            rpcRequest.preVerificationGas = numberToHex(request.preVerificationGas);
        if (typeof request.sender !== 'undefined') rpcRequest.sender = request.sender;
        if (typeof request.signature !== 'undefined') rpcRequest.signature = request.signature;
        if (typeof request.verificationGasLimit !== 'undefined')
            rpcRequest.verificationGasLimit = numberToHex(request.verificationGasLimit);
        if (typeof request.authorization !== 'undefined')
            rpcRequest.eip7702Auth = this._formatAuthorization(request.authorization);

        return rpcRequest;
    };

    private _formatAuthorization = (authorization: SignedAuthorization) => {
        return {
            address: authorization.address,
            chainId: numberToHex(authorization.chainId),
            nonce: numberToHex(authorization.nonce),
            r: authorization.r ? numberToHex(BigInt(authorization.r), { size: 32 }) : pad('0x', { size: 32 }),
            s: authorization.s ? numberToHex(BigInt(authorization.s), { size: 32 }) : pad('0x', { size: 32 }),
            yParity: authorization.yParity ? numberToHex(authorization.yParity, { size: 1 }) : pad('0x', { size: 32 }),
        };
    };
}
