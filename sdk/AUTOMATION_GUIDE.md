# 🚀 Nitrolite SDK Automation Guide

> **Complete automation system for TypeScript code generation and documentation**

This guide explains the comprehensive automation system that ensures the Nitrolite SDK maintains perfect synchronization between smart contracts and TypeScript types, with automatic documentation generation and validation.

## 🎯 What This Automation Solves

### Before Automation

- ❌ Manual ABI maintenance and potential desynchronization
- ❌ No type safety between contracts and TypeScript
- ❌ Manual documentation updates that get out of sync
- ❌ Error-prone manual processes
- ❌ No validation of SDK integrity

### After Automation

- ✅ **Auto-generated ABIs** - Always synchronized with contract changes
- ✅ **Full type safety** - Catch errors at compile time, not runtime
- ✅ **Auto-generated documentation** - Always up-to-date API docs
- ✅ **Automated validation** - Comprehensive SDK integrity checks
- ✅ **Zero maintenance** - Contract changes automatically update everything

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Solidity      │    │   Wagmi CLI     │    │   TypeScript    │
│   Contracts     │───▶│   Generator     │───▶│   SDK Types     │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Forge Build   │    │   Auto Docs     │    │   Full Type     │
│   Artifacts     │    │   Generation    │    │   Safety        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Git Hooks     │    │   Validation    │    │   CI/CD Ready   │
│   Automation    │    │   Checks        │    │   Deployment    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🛠️ Core Components

### 1. Type Generation (`wagmi.config.ts`)

**Purpose:** Automatically generates TypeScript types from smart contracts

```typescript
export default defineConfig({
  out: "src/generated.ts",
  plugins: [
    foundry({
      project: "../contract",
      include: [
        "Custody.sol/**",
        "Dummy.sol/**",
        "Consensus.sol/**",
        "Counter.sol/**",
        "Remittance.sol/**",
      ],
      exclude: [
        "*.t.sol/**",
        "*.s.sol/**",
        "forge-std/**",
        "openzeppelin-contracts/**",
      ],
      forge: {
        build: true,
        rebuild: true,
      },
    }),
  ],
});
```

**Generated Output:** `src/generated.ts` with fully typed ABIs

### 2. Documentation Generation (`scripts/generate-docs.ts`)

**Purpose:** Automatically generates comprehensive documentation from contracts

**Features:**

- Extracts contract information from generated types
- Creates detailed function documentation with examples
- Generates event and error documentation
- Produces type-safe usage examples
- Creates SDK overview with all contracts

**Generated Output:**

- `docs/README.md` - SDK overview
- `docs/contracts/[ContractName].md` - Individual contract docs

### 3. Validation System (`scripts/validate-types.ts`)

**Purpose:** Comprehensive SDK integrity validation

**Validation Checks:**

- ✅ **Generated Types** - Validates auto-generated contract types
- ✅ **TypeScript Compilation** - Ensures code compiles without errors
- ✅ **Contract Sync** - Verifies types are in sync with contracts
- ✅ **SDK Structure** - Validates exports and module structure
- ✅ **Package Configuration** - Checks dependencies and scripts

### 4. Git Hooks (`scripts/setup-hooks.sh`)

**Purpose:** Automated validation and regeneration

**Pre-commit Hook:**

- Validates TypeScript types before commits
- Prevents broken code from being committed
- Ensures SDK integrity

**Post-merge Hook:**

- Automatically rebuilds contracts after merges
- Regenerates TypeScript types
- Keeps everything in sync

## 📋 Available Scripts

| Script               | Description                              | Use Case                |
| -------------------- | ---------------------------------------- | ----------------------- |
| `npm run codegen`    | Generate TypeScript types from contracts | After contract changes  |
| `npm run validate`   | Run comprehensive SDK validation         | Before commits/releases |
| `npm run docs`       | Generate auto-updated documentation      | After type changes      |
| `npm run build:full` | Complete build with validation and docs  | Production builds       |
| `npm run dev`        | Development mode with type generation    | Active development      |
| `npm run test:types` | Test type safety and compilation         | CI/CD pipelines         |

## 🔄 Development Workflow

### Daily Development

```bash
# 1. Make contract changes
vim contract/src/Custody.sol

# 2. Build contracts (generates ABIs)
cd contract && forge build

# 3. Regenerate types and validate
cd ../sdk && npm run codegen
npm run validate

# 4. Generate updated documentation
npm run docs

# 5. Build everything
npm run build:full
```

### Automated Workflow (with Git hooks)

```bash
# 1. Make contract changes
vim contract/src/Custody.sol

# 2. Build contracts
cd contract && forge build

# 3. Commit changes (pre-commit hook validates automatically)
git add . && git commit -m "Update contract"

# 4. Merge/pull (post-merge hook regenerates types automatically)
git pull origin main
```

## 🎯 Type Safety Benefits

### Before vs After Comparison

| Manual ABIs                  | Auto-Generated (Wagmi CLI)     |
| ---------------------------- | ------------------------------ |
| ❌ Manual sync required      | ✅ Automatic sync              |
| ❌ Risk of desynchronization | ✅ Always in sync              |
| ❌ No type safety            | ✅ Full type safety            |
| ❌ Manual maintenance        | ✅ Zero maintenance            |
| ❌ Error-prone               | ✅ Compile-time error catching |
| ❌ Manual documentation      | ✅ Auto-generated docs         |

### Type Safety Example

```typescript
import { custodyAbi } from "@erc7824/nitrolite";

// ✅ Full type safety and autocomplete
const result = await publicClient.readContract({
  address: CUSTODY_ADDRESS,
  abi: custodyAbi,
  functionName: "getAccountInfo", // ✅ Auto-complete available
  args: [userAddress, tokenAddress], // ✅ Args type-checked
});

// ✅ Return types are fully typed
const { available, channelCount } = result;
```

## 🔍 Validation Features

### Automated Checks

The validation system performs comprehensive checks:

1. **Generated Types Validation**

   - Verifies `src/generated.ts` exists and contains valid ABIs
   - Counts and validates contract exports
   - Ensures proper TypeScript format

2. **TypeScript Compilation**

   - Runs `tsc --noEmit` to check for compilation errors
   - Validates all type definitions
   - Ensures SDK compiles cleanly

3. **Contract Synchronization**

   - Compares contract build timestamps with generated types
   - Ensures types are up-to-date with latest contracts
   - Prevents stale type usage

4. **SDK Structure Integrity**

   - Validates essential exports in `src/index.ts`
   - Ensures proper module structure
   - Checks for missing dependencies

5. **Package Configuration**
   - Validates required npm scripts exist
   - Checks for essential dependencies
   - Ensures proper package.json setup

### Validation Output Example

```bash
🔍 Running SDK validation checks...

✅ Generated Types: Generated types are valid with 5 contract ABIs
✅ TypeScript Compilation: TypeScript compilation successful
✅ Contract Sync: Contract types are in sync
✅ SDK Structure: SDK structure is valid
✅ Package Configuration: Package configuration is valid

🎉 All validation checks passed! SDK is reliable and ready.
```

## 📚 Documentation Features

### Auto-Generated Documentation

The documentation system creates:

1. **SDK Overview** (`docs/README.md`)

   - Quick start guide
   - Available contracts summary
   - Type safety features
   - Development workflow

2. **Contract Documentation** (`docs/contracts/[ContractName].md`)
   - Complete function reference
   - Parameter documentation
   - Return value documentation
   - Type-safe usage examples
   - Event and error documentation

### Documentation Example

````markdown
#### `getAccountInfo`

**State Mutability:** `view`

**Parameters:**

- `user` (`address`): address
- `token` (`address`): address

**Returns:**

- `available` (`uint256`): uint256
- `channelCount` (`uint256`): uint256

**Usage Example:**

```typescript
import { getAccountInfo } from "@erc7824/nitrolite";

const result = await publicClient.readContract({
  address: contractAddress,
  abi: contractAbi,
  functionName: "getAccountInfo",
  args: [userAddress, tokenAddress],
});
```
````

## 🚀 CI/CD Integration

### GitHub Actions Example

```yaml
name: SDK Validation
on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: "20"

      - name: Install dependencies
        run: |
          cd contract && forge install
          cd ../sdk && npm install

      - name: Build contracts
        run: cd contract && forge build

      - name: Validate SDK
        run: cd sdk && npm run validate

      - name: Build SDK
        run: cd sdk && npm run build:full
```

## 🔧 Setup Instructions

### Initial Setup

1. **Install Dependencies**

   ```bash
   cd sdk && npm install
   ```

2. **Build Contracts**

   ```bash
   cd contract && forge build
   ```

3. **Generate Types**

   ```bash
   cd sdk && npm run codegen
   ```

4. **Setup Git Hooks**

   ```bash
   cd sdk && ./scripts/setup-hooks.sh
   ```

5. **Validate Everything**
   ```bash
   npm run validate
   ```

### Configuration Files

- `wagmi.config.ts` - Type generation configuration
- `package.json` - Scripts and dependencies
- `tsconfig.json` - TypeScript configuration
- `.gitignore` - Excludes generated files

## 🎉 Success Metrics

The automation system provides:

- **5+ contract ABIs** with full type safety
- **Complex types** (State, Allocation, Signature) handled perfectly
- **Live sync** - contract changes appear immediately in types
- **Zero manual work** required after setup
- **Build integration** - types update automatically
- **Comprehensive validation** - 5 different validation checks
- **Auto-generated documentation** - Always up-to-date API docs

## 🤝 Contributing

When contributing to the SDK:

1. Make contract changes in `/contract`
2. Run `forge build` to compile contracts
3. Run `npm run codegen` to regenerate types
4. Run `npm run validate` to ensure everything works
5. Run `npm run docs` to update documentation
6. Commit changes (pre-commit hook validates automatically)

The automation ensures your changes are properly typed and documented!

## 🔮 Future Enhancements

Potential improvements to the automation system:

- **Watch mode** for automatic regeneration during development
- **API documentation hosting** with automated deployment
- **Type coverage reporting** to ensure complete type safety
- **Performance monitoring** for build and validation times
- **Integration with more contract frameworks** beyond Foundry

---

**🚀 This automation system provides exactly what TypeChain offers, but specifically tailored for the Viem + Foundry stack with enhanced reliability features!**
