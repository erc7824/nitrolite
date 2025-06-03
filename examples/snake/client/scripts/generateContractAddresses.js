import { readFileSync, writeFileSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';
import { execSync } from 'child_process';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

function getRepoRoot() {
  return execSync('git rev-parse --show-toplevel', { encoding: 'utf8' }).trim();
}

function getContractAddress(contractName, chainId = 137) {
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
    
    const broadcastData = JSON.parse(readFileSync(broadcastPath, 'utf8'));
    const transaction = broadcastData.transactions.find(tx => tx.contractName === contractName);
    
    if (!transaction) {
      throw new Error(`Contract ${contractName} not found in broadcast file`);
    }
    
    return transaction.contractAddress;
  } catch (error) {
    console.error(`Failed to read contract address for ${contractName}:`, error.message);
    console.error(`Tried to read from: ${broadcastPath}`);
    throw new Error(`Contract address for ${contractName} must be read from broadcast file`);
  }
}

// Generate the contract addresses file
const custodyAddress = getContractAddress('Custody');
const adjudicatorAddress = getContractAddress('Dummy');

const content = `// Auto-generated file - do not edit manually
// Generated from contract broadcast files at build time

import type { Hex } from 'viem';

export function getCustodyAddress(): Hex {
  return '${custodyAddress}' as Hex;
}

export function getAdjudicatorAddress(): Hex {
  return '${adjudicatorAddress}' as Hex;
}
`;

const outputPath = join(__dirname, '..', 'src', 'config', 'contractAddresses.ts');
writeFileSync(outputPath, content, 'utf8');

console.log('✅ Contract addresses generated:');
console.log(`   Custody: ${custodyAddress}`);
console.log(`   Adjudicator: ${adjudicatorAddress}`);