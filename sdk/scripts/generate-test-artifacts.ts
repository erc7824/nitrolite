import * as fs from 'fs';
import * as path from 'path';

const CONTRACTS_OUT_DIR = '../contract/out';
const TEST_ARTIFACTS_DIR = 'test/integration/artifacts';

interface ContractJson {
    abi: any[];
    bytecode: {
        object: string;
    };
}

interface ContractMapping {
    name: string;
    solidityFile: string;
    contractName: string;
    outputFileName: string;
}

const contracts: ContractMapping[] = [
    {
        name: 'Custody',
        solidityFile: 'Custody.sol',
        contractName: 'Custody',
        outputFileName: 'custody.ts'
    },
    {
        name: 'Dummy',
        solidityFile: 'Dummy.sol', 
        contractName: 'Dummy',
        outputFileName: 'dummy.ts'
    },
    {
        name: 'TestERC20',
        solidityFile: 'TestERC20.sol',
        contractName: 'TestERC20', 
        outputFileName: 'testERC20.ts'
    }
];

function generateArtifact(contract: ContractMapping): void {
    const contractPath = path.join(CONTRACTS_OUT_DIR, contract.solidityFile, `${contract.contractName}.json`);
    
    if (!fs.existsSync(contractPath)) {
        console.warn(`Contract file not found: ${contractPath}`);
        return;
    }

    const contractJson: ContractJson = JSON.parse(fs.readFileSync(contractPath, 'utf8'));
    
    const bytecode = contractJson.bytecode.object.startsWith('0x') 
        ? contractJson.bytecode.object 
        : `0x${contractJson.bytecode.object}`;

    const output = `// Auto-generated test artifact. Do not edit manually.
// Generated from: ${contract.solidityFile}/${contract.contractName}
export const ${contract.name}Artifacts = {
    abi: ${JSON.stringify(contractJson.abi, null, 4)},
    bytecode: '${bytecode}' as \`0x\${string}\`,
};
`;

    const outputPath = path.join(TEST_ARTIFACTS_DIR, contract.outputFileName);
    
    // Ensure directory exists
    const dir = path.dirname(outputPath);
    if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir, { recursive: true });
    }

    fs.writeFileSync(outputPath, output);
    console.log(`Generated test artifact: ${outputPath}`);
}

// Generate all test artifacts
contracts.forEach(generateArtifact);

console.log('Test artifact generation complete!');