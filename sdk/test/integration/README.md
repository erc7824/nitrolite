# Integration Tests

This directory contains integration tests that test the full Nitrolite SDK with real blockchain interactions.

## Structure

Simple structure similar to unit tests:

- `client/` - Client integration tests
    - `client.test.ts` - Client initialization and account management
    - `deposits.test.ts` - Deposit operations (ERC20 and ETH)
    - `channels.test.ts` - State channel creation and management
- `setup.ts` - Test environment setup and utilities
- `artifacts/` - Auto-generated contract artifacts

## Running Tests

```bash
# Run all integration tests
npm run test:integration

# Run specific test files
npx jest --config jest.integration.config.js client/client.test.ts
npx jest --config jest.integration.config.js client/deposits.test.ts
npx jest --config jest.integration.config.js client/channels.test.ts
```

## Test Environment

Integration tests use:

- **Real blockchain**: Anvil local blockchain
- **Real contracts**: Deployed test contracts
- **Real transactions**: Actual on-chain interactions
- **Test accounts**: Pre-funded test accounts

## Performance

- **Timeout**: 30 seconds per test
- **Execution**: Sequential (maxWorkers: 1)
- **Setup**: Shared test environment with snapshots
- **Cleanup**: Automatic blockchain state reset
