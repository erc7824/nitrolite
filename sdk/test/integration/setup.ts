import { createEVM } from '@ethereumjs/evm';
import { createVM } from '@ethereumjs/vm';
import { MerkleStateManager } from '@ethereumjs/statemanager';
import { createBlockchain } from '@ethereumjs/blockchain';
import { Common, Hardfork, Mainnet } from '@ethereumjs/common';
import { createPublicClient, createTestClient, createWalletClient, http, parseEther, Account, type Address, type Hash, type PublicClient, type TestClient, type WalletClient } from 'viem';
import { anvil } from 'viem/chains';
import { privateKeyToAccount } from 'viem/accounts';

export interface TestEnvironment {
  evm: any;
  vm: any;
  stateManager: any;
  blockchain: any;
  publicClient: PublicClient;
  testClient: TestClient;
  walletClient: WalletClient;
  accounts: {
    alice: Account;
    bob: Account;
    charlie: Account;
    deployer: Account;
  };
  deployedContracts: {
    custody?: Address;
    adjudicator?: Address;
    token?: Address;
  };
}

export const TEST_CONSTANTS = {
  INITIAL_BALANCE: parseEther('1000'),
  CHALLENGE_PERIOD: 3600, // 1 hour in seconds
  GAS_LIMIT: 10000000n,
  BLOCK_TIME: 12, // seconds
  CHAIN_ID: 31337, // Anvil default
};

// Test accounts with known private keys
const TEST_PRIVATE_KEYS = {
  alice: '0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80' as `0x${string}`,
  bob: '0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d' as `0x${string}`,
  charlie: '0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a' as `0x${string}`,
  deployer: '0x7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6' as `0x${string}`,
};

let testEnvironment: TestEnvironment | null = null;

export async function getTestEnvironment(): Promise<TestEnvironment> {
  if (testEnvironment) {
    return testEnvironment;
  }

  // Initialize EthereumJS components
  const common = new Common({ chain: Mainnet, hardfork: Hardfork.London });
  const stateManager = new MerkleStateManager();
  const blockchain = await createBlockchain({ common });
  const vm = await createVM({ common, stateManager, blockchain });
  const evm = await createEVM({ common, stateManager, blockchain });

  // Create test accounts
  const accounts = {
    alice: privateKeyToAccount(TEST_PRIVATE_KEYS.alice),
    bob: privateKeyToAccount(TEST_PRIVATE_KEYS.bob),
    charlie: privateKeyToAccount(TEST_PRIVATE_KEYS.charlie),
    deployer: privateKeyToAccount(TEST_PRIVATE_KEYS.deployer),
  };

  // Create viem clients
  const publicClient = createPublicClient({
    chain: anvil,
    transport: http('http://127.0.0.1:8545'),
  });

  const testClient = createTestClient({
    chain: anvil,
    transport: http('http://127.0.0.1:8545'),
    mode: 'anvil',
  });

  const walletClient = createWalletClient({
    chain: anvil,
    transport: http('http://127.0.0.1:8545'),
    account: accounts.alice, // Default to Alice for tests
  });

  testEnvironment = {
    evm,
    vm,
    stateManager,
    blockchain,
    publicClient,
    testClient,
    walletClient,
    accounts,
    deployedContracts: {},
  };

  return testEnvironment;
}

export async function resetTestEnvironment(): Promise<void> {
  if (testEnvironment) {
    // Reset blockchain state
    await testEnvironment.testClient.reset();
    testEnvironment.deployedContracts = {};
  }
}

export async function fundAccount(
  testEnv: TestEnvironment,
  address: Address,
  amount: bigint
): Promise<void> {
  await testEnv.testClient.setBalance({
    address,
    value: amount,
  });
}

export async function getAccountBalance(
  testEnv: TestEnvironment,
  address: Address
): Promise<bigint> {
  return await testEnv.publicClient.getBalance({ address });
}

export async function deployContract(
  testEnv: TestEnvironment,
  bytecode: `0x${string}`,
  args: any[] = []
): Promise<Address> {
  // This is a simplified deployment - in a real implementation,
  // you would use the actual contract artifacts and ABI
  const hash = await testEnv.walletClient.deployContract({
    abi: [], // Would be the actual ABI
    bytecode,
    args,
    account: testEnv.accounts.deployer,
    chain: anvil,
  });

  const receipt = await testEnv.publicClient.waitForTransactionReceipt({ hash });
  
  if (!receipt.contractAddress) {
    throw new Error('Failed to deploy contract');
  }

  return receipt.contractAddress;
}

export async function mineBlock(testEnv: TestEnvironment): Promise<void> {
  await testEnv.testClient.mine({ blocks: 1 });
}

export function getContractArtifacts() {
  // Mock contract artifacts - in a real implementation, these would be loaded
  // from the actual compiled contracts
  return {
    custody: {
      bytecode: '0x608060405234801561001057600080fd5b50' as `0x${string}`,
      abi: [],
    },
    adjudicator: {
      bytecode: '0x608060405234801561001057600080fd5b50' as `0x${string}`,
      abi: [],
    },
    token: {
      bytecode: '0x608060405234801561001057600080fd5b50' as `0x${string}`,
      abi: [],
    },
  };
}

// Utility function to create a wallet client for a specific account
export function createWalletClientForAccount(
  testEnv: TestEnvironment,
  account: Account
): WalletClient<any, any, Account> {
  return createWalletClient({
    chain: anvil,
    transport: http('http://127.0.0.1:8545'),
    account,
  }) as WalletClient<any, any, Account>;
}

// Jest setup
beforeAll(async () => {
  await getTestEnvironment();
});

afterAll(async () => {
  await resetTestEnvironment();
}); 