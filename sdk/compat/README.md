# Nitrolite Compat SDK

[![License](https://img.shields.io/npm/l/@erc7824/nitrolite.svg)](https://github.com/erc7824/nitrolite/blob/main/LICENSE)

Compatibility layer that bridges the Nitrolite SDK **v0.5.3 API** to the **v1.0.0 runtime**, letting existing dApps upgrade to the new protocol with minimal code changes.

```text
┌─────────────────────┐
│    Your dApp code    │  ← unchanged v0.5.3 imports
├─────────────────────┤
│  @erc7824/nitrolite  │
│       -compat        │  ← this package (translation layer)
├─────────────────────┤
│  @erc7824/nitrolite  │  ← v1.0.0 SDK (actual runtime)
└─────────────────────┘
```

## Why

The v1.0.0 protocol introduces breaking changes across 14 dimensions — wire format, authentication, WebSocket lifecycle, unit system, asset resolution, and more. A direct migration touches 20+ files per app with deep, scattered rewrites.

The compat layer centralises this complexity into **~1,000 lines** that absorb the protocol differences, reducing per-app integration effort by an estimated **56–70%**.

## Build Size

Measured on **February 24, 2026** from `sdk/compat` using:

```bash
npm run build:prod
npm pack --dry-run --json
```

| Metric | Size |
|---|---:|
| npm tarball (`size`) | 16,503 bytes (16.1 KB) |
| unpacked package (`unpackedSize`) | 73,292 bytes (71.6 KB) |
| compiled JS in `dist/*.js` | 38,146 bytes (37.3 KB) |
| type declarations in `dist/*.d.ts` | 20,293 bytes (19.8 KB) |
| total emitted runtime + types (`.js` + `.d.ts`) | 58,439 bytes (57.1 KB) |

## Migration Guide

Step-by-step guides for migrating from v0.5.3:

- [Overview & Quick Start](./docs/migration-overview.md) — pattern changes, import swaps
- [On-Chain Changes](./docs/migration-onchain.md) — deposits, withdrawals, channels
- [Off-Chain Changes](./docs/migration-offchain.md) — auth, app sessions, transfers, ledger queries

## Installation

```bash
npm install @erc7824/nitrolite-compat
# peer dependencies
npm install @erc7824/nitrolite viem
```

Or with a local `file:` reference (monorepo):

```json
{
  "dependencies": {
    "@erc7824/nitrolite-compat": "file:../sdk/compat",
    "@erc7824/nitrolite": "file:../sdk/ts"
  }
}
```

## Quick Start

### 1. Initialize the client

Replace `new Client(ws, signer)` with `NitroliteClient.create()`:

```typescript
import { NitroliteClient, WalletStateSigner, blockchainRPCsFromEnv } from '@erc7824/nitrolite-compat';

const client = await NitroliteClient.create({
  wsURL: 'wss://clearnode.example.com/ws',
  walletClient,          // viem WalletClient with account
  chainId: 11155111,     // Sepolia
  blockchainRPCs: blockchainRPCsFromEnv(),
});
```

### 2. Deposit & create a channel

In v1.0.0, channel creation is implicit on deposit — no separate `createChannel()` call needed:

```typescript
const tokenAddress = '0x6E2C4707DA119425DF2C722E2695300154652F56'; // USDC on Sepolia
const amount = 11_000_000n; // 11 USDC in raw units (6 decimals)

await client.deposit(tokenAddress as Address, amount);
```

### 3. Query channels, balances, ledger entries

```typescript
const channels = await client.getChannels();
const balances = await client.getBalances();
const entries  = await client.getLedgerEntries();
const sessions = await client.getAppSessionsList();
const assets   = await client.getAssetsList();
```

### 4. Transfer off-chain

```typescript
await client.transfer(recipientAddress, [
  { asset: 'usdc', amount: '5.0' },
]);
```

### 5. Close & clean up

```typescript
await client.closeChannel();
await client.close();
```

## Method Cheat Sheet

### Channel Operations

| Method | Description |
|---|---|
| `deposit(token, amount)` | Deposit to channel (creates if needed) |
| `depositAndCreateChannel(token, amount)` | Alias for `deposit()` |
| `withdrawal(token, amount)` | Withdraw from channel |
| `closeChannel()` | Close all open channels |
| `resizeChannel({ allocate_amount, token })` | Resize an existing channel |
| `challengeChannel({ state })` | Challenge a channel on-chain |

### Queries

| Method | Description |
|---|---|
| `getChannels()` | List all ledger channels (open, closed, etc.) |
| `getBalances(wallet?)` | Get ledger balances |
| `getLedgerEntries(wallet?)` | Get transaction history |
| `getAppSessionsList(wallet?, status?)` | List app sessions (filter by `'open'`/`'closed'`) |
| `getAssetsList()` | List supported assets |
| `getAccountInfo()` | Aggregate balance + channel count |
| `getConfig()` | Node configuration |
| `getChannelData(channelId)` | Full channel + state for a specific channel |

### App Sessions

| Method | Description |
|---|---|
| `createAppSession(definition, allocations)` | Create an app session |
| `closeAppSession(appSessionId, allocations)` | Close an app session |
| `submitAppState(params)` | Submit state update (operate/deposit/withdraw) |
| `getAppDefinition(appSessionId)` | Get the definition for a session |

### Transfers

| Method | Description |
|---|---|
| `transfer(destination, allocations)` | Off-chain transfer to another participant |

### Asset Resolution

| Method | Description |
|---|---|
| `resolveToken(tokenAddress)` | Look up asset info by token address |
| `resolveAsset(symbol)` | Look up asset info by symbol name |
| `resolveAssetDisplay(tokenAddress, chainId?)` | Get display-friendly symbol + decimals |
| `getTokenDecimals(tokenAddress)` | Get decimals for a token |
| `formatAmount(tokenAddress, rawAmount)` | Convert raw bigint → human-readable string |
| `parseAmount(tokenAddress, humanAmount)` | Convert human-readable string → raw bigint |
| `findOpenChannel(tokenAddress, chainId?)` | Find an open channel for a given token |

### Lifecycle

| Method | Description |
|---|---|
| `ping()` | Health check |
| `close()` | Close the WebSocket connection |
| `refreshAssets()` | Re-fetch the asset map from the clearnode |

### Accessing the v1.0.0 SDK Directly

The underlying v1.0.0 `Client` is exposed for advanced use cases not covered by the compat surface:

```typescript
const v1Client = client.innerClient;
await v1Client.getHomeChannel(wallet, 'usdc');
```

## Configuration

### `NitroliteClientConfig`

```typescript
interface NitroliteClientConfig {
  wsURL: string;                           // Clearnode WebSocket URL
  walletClient: WalletClient;              // viem WalletClient with account
  chainId: number;                         // Chain ID (e.g. 11155111 for Sepolia)
  blockchainRPCs?: Record<number, string>; // Optional chain ID → RPC URL map
  addresses?: ContractAddresses;           // Deprecated — ignored, addresses come from get_config
  challengeDuration?: bigint;              // Deprecated — ignored
}
```

### Environment Variables

`blockchainRPCsFromEnv()` reads `NEXT_PUBLIC_BLOCKCHAIN_RPCS`:

```text
NEXT_PUBLIC_BLOCKCHAIN_RPCS=11155111:https://rpc.sepolia.io,1:https://mainnet.infura.io/v3/KEY
```

## Signers

### `WalletStateSigner`

A v0.5.3-compatible signer class that wraps a `WalletClient`. Actual state signing in v1.0.0 is handled internally by `ChannelDefaultSigner`; this class exists so existing store types compile:

```typescript
import { WalletStateSigner } from '@erc7824/nitrolite-compat';

const signer = new WalletStateSigner(walletClient);
```

### `createECDSAMessageSigner`

Creates a `MessageSigner` function from a private key, compatible with the v0.5.3 signing pattern:

```typescript
import { createECDSAMessageSigner } from '@erc7824/nitrolite-compat';

const sign = createECDSAMessageSigner(privateKey);
const signature = await sign(payload);
```

## Error Handling

The compat layer provides typed error classes for common failure modes:

| Error Class | Code | Description |
|---|---|---|
| `CompatError` | *(varies)* | Base class for all compat errors |
| `AllowanceError` | `ALLOWANCE_INSUFFICIENT` | Token approval needed |
| `UserRejectedError` | `USER_REJECTED` | User cancelled in wallet |
| `InsufficientFundsError` | `INSUFFICIENT_FUNDS` | Not enough balance |
| `NotInitializedError` | `NOT_INITIALIZED` | Client not connected |

### `getUserFacingMessage(error)`

Returns a human-friendly string suitable for UI display:

```typescript
import { getUserFacingMessage, AllowanceError } from '@erc7824/nitrolite-compat';

try {
  await client.deposit(token, amount);
} catch (err) {
  showToast(getUserFacingMessage(err));
  // → "Transaction was rejected. Please approve the transaction in your wallet to continue."
}
```

### `NitroliteClient.classifyError(error)`

Converts raw SDK/wallet errors into the appropriate typed error:

```typescript
try {
  await client.deposit(token, amount);
} catch (err) {
  const typed = NitroliteClient.classifyError(err);
  if (typed instanceof AllowanceError) {
    // prompt user to approve
  }
}
```

## Event Polling

v0.5.3 used server-push WebSocket events. v1.0.0 uses a polling model. The `EventPoller` bridges this gap:

```typescript
import { EventPoller } from '@erc7824/nitrolite-compat';

const poller = new EventPoller(client, {
  onChannelUpdate: (channels) => updateUI(channels),
  onBalanceUpdate: (balances) => updateBalances(balances),
  onAssetsUpdate:  (assets)   => updateAssets(assets),
  onError:         (err)      => console.error(err),
}, 5000); // poll every 5 seconds

poller.start();

// Later:
poller.stop();
poller.setInterval(10000); // change interval
```

## RPC Stubs

The following functions exist so that any remaining v0.5.3 `create*Message` / `parse*Response` imports compile. They are intentionally **no-ops** — prefer calling `NitroliteClient` methods directly:

```typescript
// These compile but do nothing meaningful:
createGetChannelsMessage, parseGetChannelsResponse,
createGetLedgerBalancesMessage, parseGetLedgerBalancesResponse,
parseGetLedgerEntriesResponse, parseGetAppSessionsResponse,
createTransferMessage, createAppSessionMessage, parseCreateAppSessionResponse,
createCloseAppSessionMessage, parseCloseAppSessionResponse,
createCreateChannelMessage, parseCreateChannelResponse,
createCloseChannelMessage, parseCloseChannelResponse,
createResizeChannelMessage, parseResizeChannelResponse,
createPingMessage,
convertRPCToClientChannel, convertRPCToClientState,
parseAnyRPCResponse, NitroliteRPC
```

## Auth Stubs

v1.0.0 handles authentication internally — there is no public auth API. These stubs allow existing auth code to compile while doing nothing at runtime:

```typescript
createAuthRequestMessage(params)            // → no-op JSON string
createAuthVerifyMessage(signer, response)   // → no-op JSON string
createAuthVerifyMessageWithJWT(jwt)         // → no-op JSON string
createEIP712AuthMessageSigner(wallet, ...)  // → returns () => '0x'
```

## Types Reference

All types previously imported from `@erc7824/nitrolite` (v0.5.3) are re-exported:

### Enums

- `RPCMethod` — RPC method names (`Ping`, `GetConfig`, `GetChannels`, etc.)
- `RPCChannelStatus` — Channel status values (`Open`, `Closed`, `Resizing`, `Challenged`)

### Wire Types

- `MessageSigner` — `(payload: Uint8Array) => Promise<string>`
- `NitroliteRPCMessage` — `{ req: [number, string, any, number]; sig: string }`
- `RPCResponse` — `{ requestId, method, params }`
- `RPCBalance` — `{ asset, amount }`
- `RPCAsset` — `{ token, chainId, symbol, decimals }`
- `RPCChannelUpdate` — Full channel update payload
- `RPCLedgerEntry` — Ledger transaction entry
- `AccountID` — String alias for account identifiers

### Channel Operation Types

- `ContractAddresses` — `{ custody, adjudicator }`
- `Allocation` — `{ destination, token, amount }`
- `FinalState` — Final channel state with signatures
- `ChannelData` — `{ lastValidState, stateData }`
- `CreateChannelResponseParams`, `CloseChannelResponseParams`
- `ResizeChannelRequestParams`
- `TransferAllocation` — `{ asset, amount }`

### App Session Types

- `RPCAppDefinition` — `{ protocol, participants, weights, quorum, challenge, nonce }`
- `RPCAppSessionAllocation` — `{ participant, asset, amount }`
- `CloseAppSessionRequestParams`

### State Channel Primitives

- `Channel` — Channel metadata (id, participants, adjudicator, challenge, nonce, version)
- `State` — Channel state (channelId, version, data, allocations)
- `AppLogic<T>` — Interface for custom app logic implementations

### Clearnode Response Types

- `AccountInfo` — `{ available: bigint, channelCount: bigint }`
- `LedgerChannel` — Full ledger channel record (id, participant, status, token, amount, chain_id, etc.)
- `LedgerBalance` — `{ asset, amount }`
- `LedgerEntry` — Ledger entry with credit/debit
- `AppSession` — App session record
- `ClearNodeAsset` — `{ token, chainId, symbol, decimals }`

## Advanced Configuration

### `buildClientOptions`

Converts a `CompatClientConfig` into v1.0.0 `Option[]` values passed to `Client.create()`. Useful if you need to customise the underlying SDK client beyond what `NitroliteClient.create()` exposes:

```typescript
import { buildClientOptions, type CompatClientConfig } from '@erc7824/nitrolite-compat';

const opts = buildClientOptions({
  wsURL: 'wss://clearnode.example.com/ws',
  blockchainRPCs: { 11155111: 'https://rpc.sepolia.io' },
});
```

## Next.js Integration Notes

When using the compat package in a Next.js app with Turbopack:

1. **Add to `transpilePackages`** in `next.config.ts`:

```typescript
const nextConfig = {
  transpilePackages: ['@erc7824/nitrolite', '@erc7824/nitrolite-compat'],
};
```

2. **Use `--install-links`** when installing `file:` dependencies to avoid symlink issues:

```bash
npm install --install-links
```

3. The package declares `"sideEffects": false` in its `package.json`, enabling tree-shaking of unused exports.

## Peer Dependencies

| Package | Version |
|---|---|
| `@erc7824/nitrolite` | `>=0.5.3` |
| `viem` | `^2.0.0` |

## License

MIT
