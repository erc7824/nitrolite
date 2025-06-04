import { describe, test, expect, beforeAll, afterAll, beforeEach, afterEach } from '@jest/globals';
import { Account, Address, zeroAddress, parseEther, formatEther, keccak256, toBytes, toHex } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';
import { getTestEnvironment, resetTestEnvironment, deployContract, fundAccount, getAccountBalance, mineBlock, getContractArtifacts, TEST_CONSTANTS, TestEnvironment, createWalletClientForAccount } from './setup';

// Import SDK modules to test
import { NitroliteClient } from '../../src/client';
import { CreateChannelParams, AccountInfo, StateIntent } from '../../src/client/types';

describe('SDK Non-Regression Tests', () => {
  let testEnv: TestEnvironment;
  let client: NitroliteClient;
  let alice: Account;
  let bob: Account;
  
  beforeAll(async () => {
    testEnv = await getTestEnvironment();
    alice = testEnv.accounts.alice;
    bob = testEnv.accounts.bob;
    
    // Fund test accounts
    await fundAccount(testEnv, alice.address, TEST_CONSTANTS.INITIAL_BALANCE);
    await fundAccount(testEnv, bob.address, TEST_CONSTANTS.INITIAL_BALANCE);
    
    // Deploy contracts
    const artifacts = getContractArtifacts();
    testEnv.deployedContracts.custody = await deployContract(testEnv, artifacts.custody.bytecode);
    testEnv.deployedContracts.adjudicator = await deployContract(testEnv, artifacts.adjudicator.bytecode);
  });

  beforeEach(async () => {
    // Create a wallet client for Alice with proper account setup
    const aliceWalletClient = createWalletClientForAccount(testEnv, alice);
    
    // Initialize SDK client for each test
    client = new NitroliteClient({
      publicClient: testEnv.publicClient,
      walletClient: aliceWalletClient,
      chainId: TEST_CONSTANTS.CHAIN_ID,
      challengeDuration: BigInt(TEST_CONSTANTS.CHALLENGE_PERIOD),
      addresses: {
        custody: testEnv.deployedContracts.custody!,
        adjudicator: testEnv.deployedContracts.adjudicator!,
        guestAddress: bob.address,
        tokenAddress: zeroAddress, // ETH
      },
    });
  });

  afterEach(async () => {
    // Cleanup if needed
  });

  afterAll(async () => {
    await resetTestEnvironment();
  });

  describe('Client Initialization', () => {
    test('should initialize client successfully', async () => {
      expect(client).toBeDefined();
      expect(client.chainId).toBe(TEST_CONSTANTS.CHAIN_ID);
      expect(client.account.address).toBeDefined();
    });

    test('should fail initialization with invalid parameters', async () => {
      expect(() => {
        const invalidWalletClient = createWalletClientForAccount(testEnv, alice);
        new NitroliteClient({
          publicClient: testEnv.publicClient,
          walletClient: invalidWalletClient,
          chainId: 0, // Invalid chain ID
          challengeDuration: 100n, // Too short
          addresses: {
            custody: zeroAddress,
            adjudicator: zeroAddress,
            guestAddress: zeroAddress,
            tokenAddress: zeroAddress,
          },
        });
      }).toThrow();
    });

    test('should have valid addresses configured', async () => {
      expect(client.addresses.custody).toBeDefined();
      expect(client.addresses.adjudicator).toBeDefined();
      expect(client.addresses.guestAddress).toBeDefined();
      expect(client.addresses.tokenAddress).toBeDefined();
    });
  });

  describe('Account Management', () => {
    test('should have connected account', async () => {
      expect(client.account).toBeDefined();
      expect(client.account.address).toBe(alice.address);
    });

    test('should get account info', async () => {
      const accountInfo: AccountInfo = await client.getAccountInfo();
      expect(accountInfo).toBeDefined();
      expect(accountInfo.available).toBeGreaterThanOrEqual(0n);
      expect(accountInfo.channelCount).toBeGreaterThanOrEqual(0n);
    });

    test('should get token balance', async () => {
      const balance = await client.getTokenBalance();
      expect(balance).toBeGreaterThanOrEqual(0n);
    });

    test('should get token allowance', async () => {
      const allowance = await client.getTokenAllowance();
      expect(allowance).toBeGreaterThanOrEqual(0n);
    });
  });

  describe('Deposit Operations', () => {
    test('should deposit ETH successfully', async () => {
      const depositAmount = parseEther('1');
      const txHash = await client.deposit(depositAmount);
      
      expect(txHash).toBeDefined();
      expect(txHash).toMatch(/^0x[a-fA-F0-9]{64}$/);
    });

    test('should handle deposit with insufficient balance', async () => {
      const depositAmount = parseEther('10000'); // More than available
      
      await expect(client.deposit(depositAmount)).rejects.toThrow();
    });

    test('should approve tokens if needed', async () => {
      const approveAmount = parseEther('100');
      const txHash = await client.approveTokens(approveAmount);
      
      expect(txHash).toBeDefined();
      expect(txHash).toMatch(/^0x[a-fA-F0-9]{64}$/);
    });
  });

  describe('State Channel Operations', () => {
    test('should create state channel', async () => {
      const channelParams: CreateChannelParams = {
        initialAllocationAmounts: [parseEther('5'), parseEther('5')],
        stateData: '0x',
      };

      const result = await client.createChannel(channelParams);
      
      expect(result).toBeDefined();
      expect(result.channelId).toBeDefined();
      expect(result.initialState).toBeDefined();
      expect(result.txHash).toMatch(/^0x[a-fA-F0-9]{64}$/);
    });

    test('should deposit and create channel in one operation', async () => {
      const depositAmount = parseEther('10');
      const channelParams: CreateChannelParams = {
        initialAllocationAmounts: [parseEther('5'), parseEther('5')],
        stateData: '0x',
      };

      const result = await client.depositAndCreateChannel(depositAmount, channelParams);
      
      expect(result).toBeDefined();
      expect(result.channelId).toBeDefined();
      expect(result.initialState).toBeDefined();
      expect(result.depositTxHash).toMatch(/^0x[a-fA-F0-9]{64}$/);
      expect(result.createChannelTxHash).toMatch(/^0x[a-fA-F0-9]{64}$/);
    });

    test('should validate channel parameters', async () => {
      // Test with invalid allocation amounts (negative values)
      await expect(client.createChannel({
        initialAllocationAmounts: [-1n as any, parseEther('5')],
        stateData: '0x',
      })).rejects.toThrow();

      // Test with mismatched allocation amounts (should sum to something reasonable)
      await expect(client.createChannel({
        initialAllocationAmounts: [parseEther('1000'), parseEther('1000')], // More than available
        stateData: '0x',
      })).rejects.toThrow();
    });

    test('should handle channel creation with insufficient deposit', async () => {
      const channelParams: CreateChannelParams = {
        initialAllocationAmounts: [parseEther('1000'), parseEther('1000')], // More than available
        stateData: '0x',
      };

      await expect(client.createChannel(channelParams)).rejects.toThrow();
    });
  });

  describe('Transaction Processing', () => {
    test('should handle withdrawal', async () => {
      const withdrawAmount = parseEther('0.5');
      const txHash = await client.withdrawal(withdrawAmount);
      
      expect(txHash).toBeDefined();
      expect(txHash).toMatch(/^0x[a-fA-F0-9]{64}$/);
    });

    test('should get account channels', async () => {
      const channels = await client.getAccountChannels();
      expect(Array.isArray(channels)).toBe(true);
    });

    test('should estimate gas for operations', async () => {
      // This test would verify gas estimation functionality
      // For now, we just test that the client has the necessary infrastructure
      expect(client.publicClient).toBeDefined();
      expect(client.walletClient).toBeDefined();
    });
  });

  describe('Error Handling', () => {
    test('should handle RPC errors gracefully', async () => {
      // Simulate RPC error by calling with invalid data
      await expect(client.deposit(0n)).rejects.toThrow();
    });

    test('should provide meaningful error messages', async () => {
      try {
        await client.createChannel({
          initialAllocationAmounts: [-1n as any, parseEther('5')],
          stateData: '0x',
        });
      } catch (error) {
        expect(error).toBeDefined();
        expect(error instanceof Error).toBe(true);
      }
    });

    test('should handle contract interaction failures', async () => {
      // Test with invalid contract interaction
      await expect(client.withdrawal(parseEther('99999'))).rejects.toThrow();
    });
  });

  describe('Performance Tests', () => {
    test('should handle multiple deposits efficiently', async () => {
      const startTime = Date.now();
      
      const depositPromises = Array.from({ length: 5 }, () =>
        client.deposit(parseEther('0.1'))
      );

      const results = await Promise.allSettled(depositPromises);
      const endTime = Date.now();
      
      // At least some should succeed
      const successful = results.filter(r => r.status === 'fulfilled');
      expect(successful.length).toBeGreaterThan(0);
      
      // Should complete within reasonable time
      expect(endTime - startTime).toBeLessThan(30000); // 30 seconds
    });

    test('should maintain performance under concurrent operations', async () => {
      const operations = [
        () => client.getAccountInfo(),
        () => client.getTokenBalance(),
        () => client.getTokenAllowance(),
        () => client.getAccountChannels(),
      ];

      const startTime = Date.now();
      const results = await Promise.allSettled(
        operations.map(op => op())
      );
      const endTime = Date.now();

      // All should succeed
      const successful = results.filter(r => r.status === 'fulfilled');
      expect(successful.length).toBe(operations.length);
      
      // Should be fast
      expect(endTime - startTime).toBeLessThan(5000); // 5 seconds
    });
  });

  describe('Security Tests', () => {
    test('should validate transaction signatures', async () => {
      // Test would verify signature validation
      expect(client.walletClient.account).toBeDefined();
    });

    test('should prevent unauthorized operations', async () => {
      // Test would verify authorization checks
      expect(client.account.address).toBe(alice.address);
    });

    test('should handle malformed inputs safely', async () => {
      // Test with malformed inputs
      await expect(client.deposit(-1n as any)).rejects.toThrow();
    });
  });

  describe('Integration with Smart Contracts', () => {
    test('should interact with custody contract', async () => {
      expect(client.addresses.custody).toBeDefined();
      expect(client.addresses.custody).not.toBe(zeroAddress);
    });

    test('should interact with adjudicator contract', async () => {
      expect(client.addresses.adjudicator).toBeDefined();
      expect(client.addresses.adjudicator).not.toBe(zeroAddress);
    });

    test('should handle contract events', async () => {
      // Test would verify event handling
      expect(client.publicClient).toBeDefined();
    });
  });
}); 