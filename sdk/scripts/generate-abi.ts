import * as fs from 'fs';
import * as path from 'path';

const contractPath = '../contract/out/Custody.sol/Custody.json';
const outputPath = 'src/abis/generated.ts';

// Read the contract JSON
const contractJson = JSON.parse(fs.readFileSync(contractPath, 'utf8'));
const abi = contractJson.abi;

// Generate TypeScript file
const output = `// Auto-generated file. Do not edit manually.
export const custodyAbi = ${JSON.stringify(abi, null, 2)} as const;
`;

// Ensure directory exists
const dir = path.dirname(outputPath);
if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
}

// Write the output
fs.writeFileSync(outputPath, output);
console.log(`Generated ABI types at ${outputPath}`);
