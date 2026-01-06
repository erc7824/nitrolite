import { describe, test, expect, beforeEach, jest } from '@jest/globals';
import { Erc20Service } from '../../../../src/client/services/Erc20Service';
import { Errors } from '../../../../src/errors';
import { Address, SimulateContractReturnType } from 'viem';

describe('Erc20Service', () => {
    const tokenAddress = '0x0000000000000000000000000000000000000001' as Address;
    const owner = '0x0000000000000000000000000000000000000002' as Address;
    const spender = '0x0000000000000000000000000000000000000003' as Address;
    const account = '0x0000000000000000000000000000000000000004' as Address;

    let mockPublicClient: any;
    let mockWalletClient: any;
    let service: Erc20Service;

    beforeEach(() => {
        mockPublicClient = {
            readContract: jest.fn(),
            simulateContract: jest.fn(),
        };
        mockWalletClient = {
            writeContract: jest.fn(),
            account: account,
        };
        service = new Erc20Service(mockPublicClient as any, mockWalletClient as any, account);
    });

    describe('constructor', () => {
        test('should throw if publicClient is missing', () => {
            expect(() => new Erc20Service(undefined as any)).toThrow(Errors.MissingParameterError);
        });
    });

    describe('getTokenBalance', () => {
        test('should return balance', async () => {
            (mockPublicClient.readContract as any).mockResolvedValue(100n);
            const result = await service.getTokenBalance(tokenAddress, owner);
            expect(mockPublicClient.readContract).toHaveBeenCalledWith({
                address: tokenAddress,
                abi: expect.anything(),
                functionName: 'balanceOf',
                args: [owner],
            });
            expect(result).toBe(100n);
        });

        test('should throw ContractReadError on error', async () => {
            (mockPublicClient.readContract as any).mockRejectedValue(new Error('fail'));
            await expect(service.getTokenBalance(tokenAddress, owner)).rejects.toThrow(Errors.ContractReadError);
        });

        test('should rethrow NitroliteError', async () => {
            const nitroErr = new Errors.MissingParameterError('test');
            (mockPublicClient.readContract as any).mockRejectedValue(nitroErr);
            await expect(service.getTokenBalance(tokenAddress, owner)).rejects.toThrow(nitroErr);
        });
    });

    describe('getTokenAllowance', () => {
        test('should return allowance', async () => {
            (mockPublicClient.readContract as any).mockResolvedValue(50n);
            const result = await service.getTokenAllowance(tokenAddress, owner, spender);
            expect(mockPublicClient.readContract).toHaveBeenCalledWith({
                address: tokenAddress,
                abi: expect.anything(),
                functionName: 'allowance',
                args: [owner, spender],
            });
            expect(result).toBe(50n);
        });

        test('should throw ContractReadError on error', async () => {
            (mockPublicClient.readContract as any).mockRejectedValue(new Error('fail'));
            await expect(service.getTokenAllowance(tokenAddress, owner, spender)).rejects.toThrow(
                Errors.ContractReadError,
            );
        });

        test('should rethrow NitroliteError', async () => {
            const nitroErr = new Errors.WalletClientRequiredError();
            (mockPublicClient.readContract as any).mockRejectedValue(nitroErr);
            await expect(service.getTokenAllowance(tokenAddress, owner, spender)).rejects.toThrow(nitroErr);
        });
    });

    describe('prepareApprove', () => {
        test('should return request', async () => {
            const requestObj = { to: '0x', data: '0x' } as unknown as SimulateContractReturnType['request'];
            // simulateContract returns { request, result }
            (mockPublicClient.simulateContract as any).mockResolvedValue({
                request: requestObj,
                result: undefined as any, // Add missing 'result' property
            });
            const result = await service.prepareApprove(tokenAddress, spender, 123n);
            expect(mockPublicClient.simulateContract).toHaveBeenCalledWith({
                address: tokenAddress,
                abi: expect.anything(),
                functionName: 'approve',
                args: [spender, 123n],
                account,
            });
            expect(result).toBe(requestObj);
        });
    });

    describe('approve', () => {
        test('should execute and return tx hash', async () => {
            const requestObj = { to: '0x', data: '0x' } as unknown as SimulateContractReturnType['request'];
            // simulateContract returns { request, result }
            (mockPublicClient.simulateContract as any).mockResolvedValue({
                request: requestObj,
                result: undefined as any, // Add missing 'result' property
            });
            (mockWalletClient.writeContract as any).mockResolvedValue('0xhash');
            const result = await service.approve(tokenAddress, spender, 123n);
            expect(mockWalletClient.writeContract).toHaveBeenCalledWith({
                ...requestObj,
                account,
            });
            expect(result).toBe('0xhash');
        });

        test('should throw WalletClientRequiredError if walletClient missing', async () => {
            const svc = new Erc20Service(mockPublicClient as any, undefined, account);
            await expect(svc.approve(tokenAddress, spender, 123n)).rejects.toThrow(Errors.WalletClientRequiredError);
        });

        test('should throw TransactionError on error', async () => {
            const requestObj = { to: '0x', data: '0x' } as unknown as SimulateContractReturnType['request'];
            (mockPublicClient.simulateContract as any).mockResolvedValue({
                request: requestObj,
                result: undefined as any,
            });
            (mockWalletClient.writeContract as any).mockRejectedValue(new Error('fail'));
            await expect(service.approve(tokenAddress, spender, 123n)).rejects.toThrow(Errors.TransactionError);
        });
    });
});
