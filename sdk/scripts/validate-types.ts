import { readFileSync, existsSync } from 'fs';
import { join } from 'path';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

interface ValidationResult {
    success: boolean;
    message: string;
    details?: string;
}

/**
 * Validates that generated.ts exists and is valid
 */
async function validateGeneratedFile(): Promise<ValidationResult> {
    const generatedPath = join(__dirname, '../src/abis/generated.ts');

    if (!existsSync(generatedPath)) {
        return {
            success: false,
            message: 'Generated types file does not exist',
            details: 'Run `npm run codegen` to generate types from smart contracts',
        };
    }

    try {
        const content = readFileSync(generatedPath, 'utf-8');

        // Check if file contains expected exports (look for ABI exports)
        const abiExportMatches = content.match(/export const \w+Abi/g);

        if (!abiExportMatches || abiExportMatches.length === 0) {
            return {
                success: false,
                message: 'Generated file does not contain valid ABI exports',
                details: 'The generated.ts file exists but may be malformed or empty',
            };
        }

        // Count total exports
        const exportCount = abiExportMatches.length;

        return {
            success: true,
            message: `Generated types are valid with ${exportCount} contract ABIs`,
        };
    } catch (error) {
        return {
            success: false,
            message: 'Error reading generated types file',
            details: error instanceof Error ? error.message : 'Unknown error',
        };
    }
}

/**
 * Validates TypeScript compilation
 */
async function validateTypeScriptCompilation(): Promise<ValidationResult> {
    try {
        const { stdout, stderr } = await execAsync('npx tsc --noEmit', {
            cwd: join(__dirname, '..'),
        });

        if (stderr && stderr.includes('error')) {
            return {
                success: false,
                message: 'TypeScript compilation errors found',
                details: stderr,
            };
        }

        return {
            success: true,
            message: 'TypeScript compilation successful',
        };
    } catch (error) {
        return {
            success: false,
            message: 'TypeScript compilation failed',
            details: error instanceof Error ? error.message : 'Unknown compilation error',
        };
    }
}

/**
 * Validates that contract ABIs are in sync with source
 */
async function validateContractSync(): Promise<ValidationResult> {
    try {
        // Check if contract build artifacts are newer than generated types
        const generatedPath = join(__dirname, '../src/abis/generated.ts');
        const contractOutPath = join(__dirname, '../../contract/out');

        if (!existsSync(contractOutPath)) {
            return {
                success: false,
                message: 'Contract artifacts not found',
                details: 'Run `forge build` in the contract directory first',
            };
        }

        if (!existsSync(generatedPath)) {
            return {
                success: false,
                message: 'Generated types not found',
                details: 'Run `npm run codegen` to generate types',
            };
        }

        const contractStat = require('fs').statSync(contractOutPath);
        const generatedStat = require('fs').statSync(generatedPath);

        if (contractStat.mtime > generatedStat.mtime) {
            return {
                success: false,
                message: 'Generated types are out of sync with contracts',
                details: 'Contract artifacts are newer than generated types. Run `npm run codegen`',
            };
        }

        return {
            success: true,
            message: 'Contract types are in sync',
        };
    } catch (error) {
        return {
            success: false,
            message: 'Error checking contract sync',
            details: error instanceof Error ? error.message : 'Unknown sync error',
        };
    }
}

/**
 * Validates SDK exports and structure
 */
async function validateSDKStructure(): Promise<ValidationResult> {
    try {
        const indexPath = join(__dirname, '../src/index.ts');
        const content = readFileSync(indexPath, 'utf-8');

        // Check for essential exports
        const essentialExports = [
            "export * from './types'",
            "export * from './utils'",
            "export * from './client'",
            "export * from './abis'",
        ];

        const missingExports = essentialExports.filter((exp) => !content.includes(exp));

        if (missingExports.length > 0) {
            return {
                success: false,
                message: 'Missing essential SDK exports',
                details: `Missing: ${missingExports.join(', ')}`,
            };
        }

        return {
            success: true,
            message: 'SDK structure is valid',
        };
    } catch (error) {
        return {
            success: false,
            message: 'Error validating SDK structure',
            details: error instanceof Error ? error.message : 'Unknown structure error',
        };
    }
}

/**
 * Validates package.json configuration
 */
async function validatePackageConfig(): Promise<ValidationResult> {
    try {
        const packagePath = join(__dirname, '../package.json');
        const packageContent = JSON.parse(readFileSync(packagePath, 'utf-8'));

        // Check essential scripts
        const requiredScripts = ['build', 'codegen', 'typecheck'];
        const missingScripts = requiredScripts.filter((script) => !packageContent.scripts[script]);

        if (missingScripts.length > 0) {
            return {
                success: false,
                message: 'Missing required npm scripts',
                details: `Missing scripts: ${missingScripts.join(', ')}`,
            };
        }

        // Check essential dependencies
        const requiredDeps = ['viem', 'abitype'];
        const missingDeps = requiredDeps.filter(
            (dep) => !packageContent.dependencies[dep] && !packageContent.devDependencies[dep],
        );

        if (missingDeps.length > 0) {
            return {
                success: false,
                message: 'Missing required dependencies',
                details: `Missing: ${missingDeps.join(', ')}`,
            };
        }

        return {
            success: true,
            message: 'Package configuration is valid',
        };
    } catch (error) {
        return {
            success: false,
            message: 'Error validating package configuration',
            details: error instanceof Error ? error.message : 'Unknown config error',
        };
    }
}

/**
 * Main validation function
 */
async function main() {
    console.log('Running SDK validation checks...\n');

    const validations = [
        { name: 'Generated Types', fn: validateGeneratedFile },
        { name: 'TypeScript Compilation', fn: validateTypeScriptCompilation },
        { name: 'Contract Sync', fn: validateContractSync },
        { name: 'SDK Structure', fn: validateSDKStructure },
        { name: 'Package Configuration', fn: validatePackageConfig },
    ];

    let allPassed = true;
    const results: { name: string; result: ValidationResult }[] = [];

    for (const validation of validations) {
        console.log(`Validating ${validation.name}...`);
        const result = await validation.fn();
        results.push({ name: validation.name, result });

        if (result.success) {
            console.log(`✅ ${validation.name}: ${result.message}`);
        } else {
            console.log(`❌ ${validation.name}: ${result.message}`);
            if (result.details) {
                console.log(`   Details: ${result.details}`);
            }
            allPassed = false;
        }
        console.log('');
    }

    // Summary
    console.log('📊 Validation Summary:');
    console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');

    results.forEach(({ name, result }) => {
        const status = result.success ? '✅' : '❌';
        console.log(`${status} ${name}: ${result.message}`);
    });

    console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');

    if (allPassed) {
        console.log('🎉 All validation checks passed! SDK is reliable and ready.');
        process.exit(0);
    } else {
        console.log('💥 Some validation checks failed. Please fix the issues above.');
        process.exit(1);
    }
}

if (require.main === module) {
    main();
}
