# Nitrolite SDK Documentation

> **Auto-generated documentation with real usage examples**

The Nitrolite SDK empowers developers to build high-performance, scalable web3 applications using state channels.

## Quick Start

```bash
npm install @erc7824/nitrolite
```

```typescript
import { custodyAbi, NitroliteClient } from '@erc7824/nitrolite';

// Initialize client with full type safety
const client = new NitroliteClient({ ...config });

// Deposit funds for state channels
await client.deposit(tokenAddress, amount);

// Create a state channel
const { channelId } = await client.createChannel({
    initialAllocationAmounts: [amount1, amount2],
});
```

## Available Contracts

### custody

- **Functions:** 10
- **Events:** 9
- **Errors:** 18
- [View Details](./contracts/custody.md)

## Key Features

✅ **Auto-generated Types** - Always synchronized with contract changes
✅ **Real Usage Examples** - Extracted from actual codebase usage
✅ **Business Context** - Meaningful descriptions from JSDoc comments
✅ **Type Safety** - Full TypeScript support with autocomplete
✅ **Zero Maintenance** - Documentation updates automatically
