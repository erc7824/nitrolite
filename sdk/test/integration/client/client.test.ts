import { describe, test, expect, beforeAll, afterAll, beforeEach, afterEach } from '@jest/globals';
import { Account, Address, zeroAddress, parseEther, Quantity, Hex } from 'viem';
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

describe('Client Integration Tests', () => {
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

        // Setup contracts and fund accounts (same as before)
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
                const flatSignature = await wallet._signingKey().signDigest(raw);
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

    afterEach(async () => {
        // Cleanup if needed
    });

    afterAll(async () => {
        await resetTestEnvironment(snapshotId);
    }, 30000);

    describe('Client Initialization', () => {
        test('should initialize client successfully', async () => {
            expect(client).toBeDefined();
            expect(client.chainId).toBe(TEST_CONSTANTS.CHAIN_ID);
            expect(client.account.address).toBeDefined();
        });

        test('should have valid addresses configured', async () => {
            expect(client.addresses.custody).toBeDefined();
            expect(client.addresses.adjudicator).toBeDefined();
            expect(client.addresses.guestAddress).toBeDefined();
        });
    });

    describe('Account Management', () => {
        test('should have connected account', async () => {
            expect(client.account).toBeDefined();
            expect(client.account.address).toBe(alice.address);
        });

        test('should get account balance', async () => {
            const accountBalance = await client.getAccountBalance(tokenAddress);
            expect(accountBalance).toBeDefined();
            expect(accountBalance).toBeGreaterThanOrEqual(0n);
        });

        test('should get token balance', async () => {
            const balance = await client.getTokenBalance(tokenAddress);
            expect(balance).toBeGreaterThanOrEqual(0n);
        });

        test('should get token allowance', async () => {
            const allowance = await client.getTokenAllowance(tokenAddress);
            expect(allowance).toBeGreaterThanOrEqual(0n);
        });
    });
});
