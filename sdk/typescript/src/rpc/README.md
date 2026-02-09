# Nitrolite RPC Module

This module provides RPC communication capabilities for the Nitrolite SDK, allowing clients to interact with Clearnode through WebSocket connections using the v1 API specification.

## Overview

The Nitrolite RPC protocol uses a flat array wire format for all messages. Messages are formatted as 5-element JSON arrays:

```typescript
[type, requestId, method, params, timestamp]
```

Where:
- `type`: Message type (1=Request, 2=Response, 3=Event, 4=ErrorResponse)
- `requestId`: Unique request identifier (number)
- `method`: RPC method name (string)
- `params`: Method parameters (object)
- `timestamp`: Unix timestamp in milliseconds (number)

## API Functions

### Node Operations

| Function | Description | Parameters |
|----------|-------------|------------|
| `createGetConfigMessage` | Get node configuration and supported networks | `requestId?`, `timestamp?` |
| `createGetAssetsMessage` | Get available assets | `chainId?`, `requestId?`, `timestamp?` |
| `createPingMessage` | Check connection health | `requestId?`, `timestamp?` |

### Session Key Management

| Function | Description | Parameters |
|----------|-------------|------------|
| `createRegisterMessage` | Register a new session key | `signer`, `address`, `options?`, `requestId?`, `timestamp?` |
| `createGetSessionKeysMessage` | List session keys for a wallet | `wallet`, `requestId?`, `timestamp?` |
| `createRevokeSessionKeyMessage` | Revoke a session key | `signer`, `sessionKey`, `requestId?`, `timestamp?` |

### User Operations

| Function | Description | Parameters |
|----------|-------------|------------|
| `createGetBalancesMessage` | Get user balances | `wallet`, `requestId?`, `timestamp?` |
| `createGetTransactionsMessage` | Get transaction history | `wallet`, `options?`, `requestId?`, `timestamp?` |

### Channel Operations

| Function | Description | Parameters |
|----------|-------------|------------|
| `createGetHomeChannelMessage` | Get home channel for user/asset | `wallet`, `asset`, `requestId?`, `timestamp?` |
| `createGetEscrowChannelMessage` | Get escrow channel by ID | `escrowChannelId`, `requestId?`, `timestamp?` |
| `createGetChannelsMessage` | List channels | `wallet`, `options?`, `requestId?`, `timestamp?` |
| `createGetLatestStateMessage` | Get latest state | `wallet`, `asset`, `onlySigned?`, `requestId?`, `timestamp?` |
| `createGetStatesMessage` | Get state history | `wallet`, `asset`, `onlySigned`, `options?`, `requestId?`, `timestamp?` |
| `createCreateChannelMessage` | Request channel creation | `signer`, `state`, `channelDefinition`, `requestId?`, `timestamp?` |
| `createSubmitStateMessage` | Submit state transition | `signer`, `state`, `requestId?`, `timestamp?` |

### App Session Operations

| Function | Description | Parameters |
|----------|-------------|------------|
| `createGetAppDefinitionMessage` | Get app definition | `signer`, `appSessionId`, `requestId?`, `timestamp?` |
| `createGetAppSessionsMessage` | List app sessions | `signer`, `options?`, `requestId?`, `timestamp?` |
| `createCreateAppSessionMessage` | Create app session | `signer`, `definition`, `quorumSigs`, `sessionData?`, `requestId?`, `timestamp?` |
| `createSubmitAppStateMessage` | Submit app state update | `signer`, `appStateUpdate`, `quorumSigs`, `requestId?`, `timestamp?` |
| `createSubmitDepositStateMessage` | Submit deposit state | `signer`, `appStateUpdate`, `quorumSigs`, `userState`, `requestId?`, `timestamp?` |
| `createRebalanceAppSessionsMessage` | Rebalance app sessions | `signer`, `signedUpdates`, `requestId?`, `timestamp?` |

## Message Format

### Request Example

```json
[1, 12345, "node.v1.get_config", {}, 1741344819012]
```

- Type: `1` (Request)
- Request ID: `12345`
- Method: `"node.v1.get_config"`
- Params: `{}`
- Timestamp: `1741344819012`

### Response Example

```json
[2, 12345, "node.v1.get_config", {"brokerAddress": "0x...", "networks": [...]}, 1741344819814]
```

- Type: `2` (Response)
- Request ID: `12345` (matches request)
- Method: `"node.v1.get_config"`
- Params: `{"brokerAddress": "0x...", "networks": [...]}`
- Timestamp: `1741344819814`

### Event Example

```json
[3, 0, "bu", {"balanceUpdates": [{"asset": "usdc", "amount": "1000"}]}, 1741344820000]
```

- Type: `3` (Event)
- Request ID: `0` (server-initiated)
- Method: `"bu"` (balance update)
- Params: `{"balanceUpdates": [...]}`
- Timestamp: `1741344820000`

### Error Response Example

```json
[4, 12345, "node.v1.get_config", {"error": "Server error"}, 1741344819500]
```

- Type: `4` (ErrorResponse)
- Request ID: `12345`
- Method: `"node.v1.get_config"`
- Params: `{"error": "Server error"}`
- Timestamp: `1741344819500`

## Architecture

The RPC module is organized into several layers:

1. **High-level API (`api.ts`)**: User-friendly functions for creating specific RPC messages. Each function returns a JSON-stringified message ready to be sent over WebSocket.

2. **Core Protocol (`nitrolite.ts`)**: Contains the `NitroliteRPC` class with core functionality for creating and signing messages.

3. **Type System (`types/`)**: TypeScript interfaces for requests, responses, and common types.

4. **Parsers (`parse/`)**: Zod-based parsers for validating and transforming responses from snake_case (wire format) to camelCase (TypeScript).

5. **Utilities (`utils.ts`)**: Helper functions for working with messages (extracting fields, validation, etc.).

## Using the RPC Module

### Basic Example

```typescript
import { createRegisterMessage, createGetBalancesMessage } from '@erc7824/nitrolite/rpc';
import { privateKeyToAccount } from 'viem/accounts';

// Create a signer from private key
const account = privateKeyToAccount('0x...');
const signer = async (hash: Hex) => account.signMessage({ message: { raw: hash } });

// Register a session key
const registerMsg = await createRegisterMessage(
  signer,
  account.address,
  {
    application: "my-app",
    allowances: [{ asset: "usdc", allowance: "1000000" }]
  }
);

// Get user balances (no signing required for queries)
const balancesMsg = createGetBalancesMessage(account.address);

// Send via WebSocket
websocket.send(registerMsg);
websocket.send(balancesMsg);
```

### Parsing Responses

```typescript
import { parseAnyRPCResponse, parseGetBalancesResponse } from '@erc7824/nitrolite/rpc';

websocket.onmessage = (event) => {
  // Parse any response
  const response = parseAnyRPCResponse(event.data);
  console.log('Method:', response.method);
  console.log('Params:', response.params);

  // Or parse specific response for type safety
  const balances = parseGetBalancesResponse(event.data);
  console.log('Balances:', balances.params.balances);
};
```

### Working with States

```typescript
import { createGetLatestStateMessage, createSubmitStateMessage } from '@erc7824/nitrolite/rpc';

// Get the latest state
const latestStateMsg = createGetLatestStateMessage(
  account.address,
  "usdc",
  true  // only signed states
);

// Submit a new state
const submitMsg = await createSubmitStateMessage(
  signer,
  {
    id: "state_id",
    transitions: [...],
    asset: "usdc",
    userWallet: account.address,
    epoch: 1,
    version: 2,
    homeLedger: {...},
    // ... other state fields
  }
);
```

## Response Parsers

The module provides typed parsers for all RPC methods:

```typescript
// Node operations
parseGetConfigResponse(raw: string)
parseGetAssetsResponse(raw: string)
parsePingResponse(raw: string)

// Session keys
parseRegisterResponse(raw: string)
parseGetSessionKeysResponse(raw: string)
parseRevokeSessionKeyResponse(raw: string)

// User operations
parseGetBalancesResponse(raw: string)
parseGetTransactionsResponse(raw: string)

// Channel operations
parseGetHomeChannelResponse(raw: string)
parseGetEscrowChannelResponse(raw: string)
parseGetChannelsResponse(raw: string)
parseGetLatestStateResponse(raw: string)
parseGetStatesResponse(raw: string)
parseCreateChannelResponse(raw: string)
parseSubmitStateResponse(raw: string)

// App session operations
parseGetAppDefinitionResponse(raw: string)
parseGetAppSessionsResponse(raw: string)
parseCreateAppSessionResponse(raw: string)
parseSubmitAppStateResponse(raw: string)
parseSubmitDepositStateResponse(raw: string)
parseRebalanceAppSessionsResponse(raw: string)

// Events
parseMessageResponse(raw: string)
parseAssetsResponse(raw: string)
parseBalanceUpdateResponse(raw: string)
parseTransferNotificationResponse(raw: string)
parseChannelUpdateResponse(raw: string)
parseChannelsUpdateResponse(raw: string)
parseAppSessionUpdateResponse(raw: string)
```

## Utility Functions

```typescript
import {
  getMessageType,
  getRequestId,
  getMethod,
  getParams,
  getTimestamp,
  isRequest,
  isResponse,
  isEvent,
  isErrorResponse,
  isValidResponseTimestamp,
  isValidResponseRequestId
} from '@erc7824/nitrolite/rpc';

const message: RPCMessage = JSON.parse(rawMessage);

// Extract fields
const type = getMessageType(message);       // 1, 2, 3, or 4
const requestId = getRequestId(message);    // number
const method = getMethod(message);          // string
const params = getParams(message);          // Record<string, unknown>
const timestamp = getTimestamp(message);    // number

// Type checks
if (isResponse(message)) {
  console.log('This is a response');
}

// Validation
if (isValidResponseRequestId(requestMsg, responseMsg)) {
  console.log('Response matches request');
}
```

## Breaking Changes from v0.5.x

### Wire Format

**Before (v0.5.x):**
```json
{
  "req": [requestId, "method", params, timestamp],
  "sig": ["0x..."]
}
```

**After (v1.0.0):**
```json
[RPCMessageType, requestId, "method", params, timestamp]
```

### Removed Methods

The following methods are no longer supported:

- `auth_request`, `auth_challenge`, `auth_verify` → Use `register`
- `get_ledger_balances` → Use `get_balances`
- `get_ledger_transactions` → Use `get_transactions`
- `get_ledger_entries` → Removed
- `resize_channel`, `close_channel` → Use channel state operations
- `close_app_session` → Use channel state operations
- `transfer` → Use state transitions
- `get_user_tag`, `get_rpc_history` → Removed
- `cleanup_session_key_cache`, `pong` → Removed

### Removed Functions

```typescript
// These functions no longer exist:
createAuthRequestMessage()
createAuthVerifyMessage()
createAuthVerifyMessageFromChallenge()
createGetLedgerBalancesMessage()
createGetLedgerTransactionsMessage()
createCloseChannelMessage()
createCloseAppSessionMessage()
createTransferMessage()
createCleanupSessionKeyCacheMessage()
```

### New Functions

```typescript
// New session key functions:
createRegisterMessage()
createGetSessionKeysMessage()
createRevokeSessionKeyMessage()

// New user functions:
createGetBalancesMessage()
createGetTransactionsMessage()

// New channel functions:
createGetHomeChannelMessage()
createGetEscrowChannelMessage()
createGetLatestStateMessage()
createGetStatesMessage()
createSubmitStateMessage()

// New app session functions:
createSubmitDepositStateMessage()
createRebalanceAppSessionsMessage()
```

### Type Changes

**Removed:**
- `NitroliteRPCMessage` → Use `RPCMessage`
- `ApplicationRPCMessage` → Use `RPCRequest`
- `MessageSigner` → Use `(hash: Hex) => Promise<Hex>`

**Changed:**
- `RPCTransaction` structure changed.

**New:**
- `RPCChannel`, `RPCState`, `RPCLedger`, `RPCTransition`
- `RPCBalanceEntry`, `RPCSessionKey`, `RPCAllowanceUsage`
- `PaginationMetadata`

## Method Name Reference

| Method Name | Description |
|-------------|-------------|
| `node.v1.get_config` | Get node configuration |
| `node.v1.get_assets` | Get available assets |
| `node.v1.ping` | Health check |
| `session_keys.v1.register` | Register session key |
| `session_keys.v1.get_session_keys` | List session keys |
| `session_keys.v1.revoke_session_key` | Revoke session key |
| `user.v1.get_balances` | Get user balances |
| `user.v1.get_transactions` | Get transaction history |
| `channels.v1.get_channels` | List channels |
| `channels.v1.get_home_channel` | Get home channel |
| `channels.v1.get_escrow_channel` | Get escrow channel |
| `channels.v1.get_latest_state` | Get latest state |
| `channels.v1.get_states` | Get state history |
| `channels.v1.request_creation` | Create channel |
| `channels.v1.submit_state` | Submit state |
| `app_sessions.v1.get_app_definition` | Get app definition |
| `app_sessions.v1.get_app_sessions` | List app sessions |
| `app_sessions.v1.create_app_session` | Create app session |
| `app_sessions.v1.submit_app_state` | Submit app state |
| `app_sessions.v1.submit_deposit_state` | Submit deposit state |
| `app_sessions.v1.rebalance_app_sessions` | Rebalance sessions |

## Server Push Events

| Event Code | Method | Description |
|------------|--------|-------------|
| `bu` | `balance_update` | Balance changed |
| `tr` | `transfer_notification` | Transfer received |
| `cu` | `channel_update` | Channel state changed |
| `channels` | `channels_update` | Multiple channels updated |
| `asu` | `app_session_update` | App session changed |
| `assets` | `assets` | Available assets changed |
| `message` | `message` | Virtual app message |

## Security Considerations

1. **Message Signing**: All state-modifying operations require message signing with a valid session key or wallet.

2. **Timestamp Validation**: Verify response timestamps are greater than request timestamps to prevent replay attacks.

3. **Request ID Matching**: Always verify response request IDs match the original request.

4. **Session Key Management**: Session keys should have expiration times. Revoke keys when no longer needed.

5. **State Validation**: When receiving states, verify signatures before using the data.

## Error Handling

Errors are returned as ErrorResponse messages (type 4):

```typescript
[4, requestId, method, {"error": "Error description"}, timestamp]
```

Parse errors using `parseErrorResponse()`:

```typescript
import { parseErrorResponse } from '@erc7824/nitrolite/rpc';

try {
  const response = parseAnyRPCResponse(rawMessage);
  // Handle response
} catch (error) {
  if (isErrorResponse(JSON.parse(rawMessage))) {
    const errorResponse = parseErrorResponse(rawMessage);
    console.error('RPC Error:', errorResponse.params.error);
  }
}
```
