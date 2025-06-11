import { describe, test, expect, jest, beforeEach } from '@jest/globals';
import { createDeterministicPrivateKey } from '../../src/helpers/createDeterministicPrivateKey';
import { type WalletClient, type Account, type ParseAccount, type Transport, type Chain, keccak256 } from 'viem';

jest.mock('viem', () => ({
    keccak256: jest.fn(() => '0xmockedprivatekey'),
}));

describe('createDeterministicPrivateKey', () => {
    let mockWalletClient: WalletClient<Transport, Chain, ParseAccount<Account>>;
    const mockSignMessage = jest.fn() as jest.MockedFunction<any>;
    const adjudicatorAddress = '0xAdjudicator123';
    const appAddress = '0xApp456';
    const nonce = 42;

    beforeEach(() => {
        jest.clearAllMocks();
        mockWalletClient = {
            account: {
                address: '0xUser789',
            },
            signMessage: mockSignMessage,
        } as any;
        mockSignMessage.mockResolvedValue('0xmockedsignature');
    });

    test('successfully creates deterministic private key', async () => {
        const result = await createDeterministicPrivateKey(mockWalletClient, adjudicatorAddress, appAddress, nonce);

        const expectedMessage = ['nitrolite_state_wallet_v1', adjudicatorAddress, appAddress, '0xUser789', nonce].join(
            '/',
        );

        expect(mockSignMessage).toHaveBeenCalledWith({
            message: expectedMessage,
        });
        expect(keccak256).toHaveBeenCalledWith('0xmockedsignature');
        expect(result).toBe('0xmockedprivatekey');
    });

    test('throws error when wallet client has no account', async () => {
        const walletClientNoAccount = {
            ...mockWalletClient,
            account: undefined,
        } as any;

        await expect(
            createDeterministicPrivateKey(walletClientNoAccount, adjudicatorAddress, appAddress, nonce),
        ).rejects.toThrow('WalletClient must have an account to sign the message.');
    });

    test('creates same key for same inputs', async () => {
        const result1 = await createDeterministicPrivateKey(mockWalletClient, adjudicatorAddress, appAddress, nonce);

        const result2 = await createDeterministicPrivateKey(mockWalletClient, adjudicatorAddress, appAddress, nonce);

        expect(result1).toBe(result2);
        expect(mockSignMessage).toHaveBeenCalledTimes(2);
    });

    test('creates different keys for different nonces', async () => {
        mockSignMessage.mockResolvedValueOnce('0xsignature1').mockResolvedValueOnce('0xsignature2');

        const viemMock = jest.requireMock('viem') as any;
        (viemMock.keccak256 as jest.Mock).mockReturnValueOnce('0xprivatekey1').mockReturnValueOnce('0xprivatekey2');

        const result1 = await createDeterministicPrivateKey(mockWalletClient, adjudicatorAddress, appAddress, nonce);

        const result2 = await createDeterministicPrivateKey(
            mockWalletClient,
            adjudicatorAddress,
            appAddress,
            nonce + 1,
        );

        expect(result1).toBe('0xprivatekey1');
        expect(result2).toBe('0xprivatekey2');
        expect(result1).not.toBe(result2);
    });

    test('creates different keys for different adjudicator addresses', async () => {
        mockSignMessage.mockResolvedValueOnce('0xsignature1').mockResolvedValueOnce('0xsignature2');

        const viemMock = jest.requireMock('viem') as any;
        (viemMock.keccak256 as jest.Mock).mockReturnValueOnce('0xprivatekey1').mockReturnValueOnce('0xprivatekey2');

        const result1 = await createDeterministicPrivateKey(mockWalletClient, adjudicatorAddress, appAddress, nonce);

        const result2 = await createDeterministicPrivateKey(
            mockWalletClient,
            '0xDifferentAdjudicator',
            appAddress,
            nonce,
        );

        expect(result1).toBe('0xprivatekey1');
        expect(result2).toBe('0xprivatekey2');
        expect(result1).not.toBe(result2);
    });

    test('creates different keys for different app addresses', async () => {
        mockSignMessage.mockResolvedValueOnce('0xsignature1').mockResolvedValueOnce('0xsignature2');

        const viemMock = jest.requireMock('viem') as any;
        (viemMock.keccak256 as jest.Mock).mockReturnValueOnce('0xprivatekey1').mockReturnValueOnce('0xprivatekey2');

        const result1 = await createDeterministicPrivateKey(mockWalletClient, adjudicatorAddress, appAddress, nonce);

        const result2 = await createDeterministicPrivateKey(
            mockWalletClient,
            adjudicatorAddress,
            '0xDifferentApp',
            nonce,
        );

        expect(result1).toBe('0xprivatekey1');
        expect(result2).toBe('0xprivatekey2');
        expect(result1).not.toBe(result2);
    });

    test('creates different keys for different user addresses', async () => {
        const differentWalletClient = {
            ...mockWalletClient,
            account: {
                address: '0xDifferentUser',
            },
        } as any;

        mockSignMessage.mockResolvedValueOnce('0xsignature1').mockResolvedValueOnce('0xsignature2');

        const viemMock = jest.requireMock('viem') as any;
        (viemMock.keccak256 as jest.Mock).mockReturnValueOnce('0xprivatekey1').mockReturnValueOnce('0xprivatekey2');

        const result1 = await createDeterministicPrivateKey(mockWalletClient, adjudicatorAddress, appAddress, nonce);

        const result2 = await createDeterministicPrivateKey(
            differentWalletClient,
            adjudicatorAddress,
            appAddress,
            nonce,
        );

        expect(result1).toBe('0xprivatekey1');
        expect(result2).toBe('0xprivatekey2');
        expect(result1).not.toBe(result2);
    });

    test('handles signing errors', async () => {
        mockSignMessage.mockRejectedValue(new Error('Signing failed'));

        await expect(
            createDeterministicPrivateKey(mockWalletClient, adjudicatorAddress, appAddress, nonce),
        ).rejects.toThrow('Signing failed');
    });

    test('constructs message with correct format', async () => {
        await createDeterministicPrivateKey(mockWalletClient, adjudicatorAddress, appAddress, nonce);

        const expectedMessage = 'nitrolite_state_wallet_v1/0xAdjudicator123/0xApp456/0xUser789/42';
        expect(mockSignMessage).toHaveBeenCalledWith({
            message: expectedMessage,
        });
    });
});
