import * as fs from 'fs';
import * as path from 'path';

const contractPath = '../../contracts/out/ChannelHub.sol/ChannelHub.json';
const outputPath = 'src/abis/generated.ts';

// Read the contract JSON
const contractJson = JSON.parse(fs.readFileSync(contractPath, 'utf8'));
const abi = contractJson.abi;

// Generate TypeScript file (minified to reduce package size)
// To view the ABI in readable format, see: contracts/out/ChannelHub.sol/ChannelHub.json
const output = `// Auto-generated file. Do not edit manually.
// Minified to reduce package size. View readable ABI at: contracts/out/ChannelHub.sol/ChannelHub.json
export const custodyAbi = ${JSON.stringify(abi)} as const;
`;

// Ensure directory exists
const dir = path.dirname(outputPath);
if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
}

// Write the output
fs.writeFileSync(outputPath, output);
console.log(`Generated ABI types at ${outputPath}`);
