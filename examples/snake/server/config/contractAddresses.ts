import { readFileSync } from 'fs';
import { join } from 'path';
import { execSync } from 'child_process';
import { Hex } from 'viem';

interface BroadcastTransaction {
  contractAddress: string;
  contractName: string;
}

interface BroadcastFile {
  transactions: BroadcastTransaction[];
}

function getRepoRoot(): string {
  return execSync('git rev-parse --show-toplevel', { encoding: 'utf8' }).trim();
}

function getContractAddressFromBroadcast(contractName: string, chainId: number = 137): Hex {
  try {
    const repoRoot = getRepoRoot();
    const broadcastPath = join(
      repoRoot,
      'contract',
      'broadcast',
      `${contractName}.s.sol`,
      chainId.toString(),
      'run-latest.json'
    );
    
    const broadcastData: BroadcastFile = JSON.parse(readFileSync(broadcastPath, 'utf8'));
    const transaction = broadcastData.transactions.find(tx => tx.contractName === contractName);
    
    if (!transaction) {
      throw new Error(`Contract ${contractName} not found in broadcast file`);
    }
    
    return transaction.contractAddress as Hex;
  } catch (error) {
    console.error(`Failed to read contract address for ${contractName} from broadcast file:`, error);
    throw new Error(`Contract address for ${contractName} must be read from broadcast file`);
  }
}

export function getCustodyAddress(chainId: number = 137): Hex {
  return getContractAddressFromBroadcast('Custody', chainId);
}

export function getAdjudicatorAddress(chainId: number = 137): Hex {
  return getContractAddressFromBroadcast('Dummy', chainId);
}