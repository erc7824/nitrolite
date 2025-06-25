import { describe, test, expect, beforeAll, afterAll, beforeEach } from '@jest/globals';
import { Account, Address, parseEther, Quantity, Hex } from 'viem';
import { ethers } from 'ethers';
import {
    getTestEnvironment,
    resetTestEnvironment,
    deployContract,
    getContractArtifacts,
    getDeployedContractAddresses,
    TEST_CONSTANTS,
    TestEnvironment,
    createWalletClientForAccount,
    fundAccountWithNative,
    fundAccountWithERC20,
    TEST_PRIVATE_KEYS,
} from '../setup';
import { NitroliteClient } from '../../../src/client';
import { CreateChannelParams } from '../../../src/client/types';

describe('State Channel Integration Tests', () => {
    let testEnv: TestEnvironment;
    let client: NitroliteClient;
    let alice: Account;
    let bob: Account;
    let snapshotId: Quantity;
    let tokenAddress: Address;

    beforeAll(async () => {
        testEnv = await getTestEnvironment();
        alice = testEnv.accounts.alice;
        bob = testEnv.accounts.bob;
        snapshotId = await testEnv.testClient.snapshot();

        const deployedAddresses = getDeployedContractAddresses();
        if (deployedAddresses.custody && deployedAddresses.adjudicator && deployedAddresses.token) {
            testEnv.deployedContracts.custody = deployedAddresses.custody;
            testEnv.deployedContracts.adjudicator = deployedAddresses.adjudicator;
            testEnv.deployedContracts.token = deployedAddresses.token;
        } else {
            const artifacts = getContractArtifacts();
            testEnv.deployedContracts.custody = await deployContract(
                testEnv,
                artifacts.custody.abi,
                artifacts.custody.bytecode,
            );
            testEnv.deployedContracts.adjudicator = await deployContract(
                testEnv,
                artifacts.adjudicator.abi,
                artifacts.adjudicator.bytecode,
            );
            testEnv.deployedContracts.token = await deployContract(
                testEnv,
                artifacts.token.abi,
                artifacts.token.bytecode,
                ['Nitrolite Token', 'NTL', 18, `${2n ** 256n - 1n}`],
            );
        }

        tokenAddress = testEnv.deployedContracts.token;
        await fundAccountWithNative(testEnv, alice.address, TEST_CONSTANTS.INITIAL_BALANCE);
        await fundAccountWithERC20(testEnv, alice.address, TEST_CONSTANTS.INITIAL_BALANCE);
        await fundAccountWithNative(testEnv, bob.address, TEST_CONSTANTS.INITIAL_BALANCE);
        await fundAccountWithERC20(testEnv, bob.address, TEST_CONSTANTS.INITIAL_BALANCE);
    }, 30000);

    beforeEach(async () => {
        const aliceWalletClient = createWalletClientForAccount(testEnv, alice);
        const wallet = new ethers.Wallet(TEST_PRIVATE_KEYS.alice);
        const stateWalletClient = {
            ...wallet,
            account: { address: wallet.address as Address },
            signMessage: async ({ message: { raw } }: { message: { raw: string } }) => {
                const flatSignature = wallet._signingKey().signDigest(raw);
                return ethers.utils.joinSignature(flatSignature) as Hex;
            },
        };

        client = new NitroliteClient({
            publicClient: testEnv.publicClient,
            walletClient: aliceWalletClient,
            chainId: TEST_CONSTANTS.CHAIN_ID,
            challengeDuration: BigInt(TEST_CONSTANTS.CHALLENGE_PERIOD),
            addresses: {
                custody: testEnv.deployedContracts.custody!,
                adjudicator: testEnv.deployedContracts.adjudicator!,
                guestAddress: bob.address,
            },
            // @ts-ignore
            stateWalletClient,
        });
    });

    afterAll(async () => {
        await resetTestEnvironment(snapshotId);
    }, 30000);

    test('should create state channel', async () => {
        const depositAmount = parseEther('20');
        await client.approveTokens(tokenAddress, depositAmount);
        await client.deposit(tokenAddress, depositAmount);

        const channelParams: CreateChannelParams = {
            initialAllocationAmounts: [parseEther('5'), parseEther('5')],
            stateData: '0x',
        };

        const result = await client.createChannel(tokenAddress, channelParams);
        expect(result).toBeDefined();
        expect(result.channelId).toBeDefined();
        expect(result.initialState).toBeDefined();
        expect(result.txHash).toMatch(/^0x[a-fA-F0-9]{64}$/);
    });

    test('should deposit and create channel in one operation', async () => {
        const depositAmount = parseEther('10');
        await client.approveTokens(tokenAddress, depositAmount);

        const channelParams: CreateChannelParams = {
            initialAllocationAmounts: [parseEther('5'), parseEther('5')],
            stateData: '0x',
        };

        const result = await client.depositAndCreateChannel(tokenAddress, depositAmount, channelParams);
        expect(result).toBeDefined();
        expect(result.channelId).toBeDefined();
        expect(result.initialState).toBeDefined();
        expect(result.txHash).toMatch(/^0x[a-fA-F0-9]{64}$/);
    });

    test('should get open channels', async () => {
        const channels = await client.getOpenChannels();
        expect(Array.isArray(channels)).toBe(true);
    });
});
