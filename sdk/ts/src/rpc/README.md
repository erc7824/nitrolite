# Nitrolite RPC Module

This module provides RPC communication capabilities for the Nitrolite SDK, allowing clients to interact with Nitrolite nodes through WebSocket connections.

## API Functions

| Function                               | Description                                                        | Parameters                                                                    |
| -------------------------------------- | ------------------------------------------------------------------ | ----------------------------------------------------------------------------- |
| `createAuthRequestMessage`             | Creates a signed auth request message for initial authentication   | `signer`, `clientAddress`, `requestId?`, `timestamp?`                         |
| `createAuthVerifyMessageFromChallenge` | Creates a signed auth verify message using an explicit challenge   | `signer`, `clientAddress`, `challenge`, `requestId?`, `timestamp?`            |
| `createAuthVerifyMessage`              | Creates a signed auth verify message from a raw challenge response | `signer`, `rawChallengeResponse`, `clientAddress`, `requestId?`, `timestamp?` |
| `createPingMessage`                    | Creates a signed ping message to check connection                  | `signer`, `requestId?`, `timestamp?`                                          |
| `createGetConfigMessage`               | Creates a signed get_config message                                | `signer`, `channelId`, `requestId?`, `timestamp?`                             |
| `createGetLedgerBalancesMessage`       | Creates a signed get_ledger_balances message                       | `signer`, `channelId`, `requestId?`, `timestamp?`                             |
| `createGetAppDefinitionMessage`        | Creates a signed get_app_definition message                        | `signer`, `appId`, `requestId?`, `timestamp?`                                 |
| `createAppSessionMessage`              | Creates a signed create_app_session message                        | `signer`, `params`, `intent`, `requestId?`, `timestamp?`                      |
| `createCloseAppSessionMessage`         | Creates a signed close_app_session message                         | `signer`, `params`, `intent`, `requestId?`, `timestamp?`                      |
| `createApplicationMessage`             | Creates a signed application message                               | `signer`, `appId`, `messageParams`, `requestId?`, `timestamp?`                |
| `createCloseChannelMessage`            | Creates a signed close_channel message                             | `signer`, `channelId`, `requestId?`, `timestamp?`                             |
| `createTransferMessage`                | Creates a signed transfer message                                  | `signer`, `params`, `requestId?`, `timestamp?`                                |
| `createCleanupSessionKeyCacheMessage`  | Creates a signed cleanup_session_key_cache message                 | `signer`, `requestId?`, `timestamp?`                                          |
| `revokeSessionKeyMessage`              | Creates a signed revoke_session_key message                        | `signer`, `sessionKey`, `requestId?`, `timestamp?`                            |

## NitroliteRPC Class Methods

| Method                     | Description                                      | Parameters                                                     |
| -------------------------- | ------------------------------------------------ | -------------------------------------------------------------- |
| `createRequest`            | Creates a NitroliteRPC request message           | `requestId?`, `method`, `params?`, `timestamp?`, `int?`        |
| `createAppRequest`         | Creates an application-specific RPC message      | `requestId?`, `method`, `params?`, `timestamp?`, `acc`, `int?` |
| `parseResponse`            | Parses and validates a raw response              | `rawMessage`                                                   |
| `signRequestMessage`       | Signs a request message with the provided signer | `message`, `signer`                                            |
| `verifySingleSignature`    | Verifies a single signature on a message         | `message`, `expectedSigner`, `verifier`                        |
| `verifyMultipleSignatures` | Verifies multiple signatures on a message        | `message`, `expectedSigners`, `verifier`                       |

## Overview

NitroliteRPC is a lightweight RPC protocol designed for state channels. Messages are formatted as fixed JSON arrays with a standard structure:

```js
[request_id, method, params, timestamp];
```

## Message Format

### Requests

```json
{
    "req": [1001, "subtract", [42, 23], 1741344819012],
    "sig": "0xa0ad67f51cc73aee5b874ace9bc2e2053488bde06de257541e05fc58fd8c4f149cca44f1c702fcbdbde0aa09bcd24456f465e5c3002c011a3bc0f317df7777d2"
}
```

- `req`: RPC message payload `[request_id, method, params, timestamp]`
- `sig`: Payload signature

### Responses

```json
{
    "res": [1001, "subtract", [19], 1741344819814],
    "sig": "0xd73268362b04516451ec52170f5c8ca189d35d9ac5e9041c156c9f0faf9aebd2891309e3b2b5d8788578ab3449c96f7aa81aefb25482b53f02bac42c65f806e5"
}
```

- `res`: RPC message payload `[request_id, method, result, timestamp]`
- `sig`: Payload signature

## Architecture

The RPC module is organized into two main layers:

1. **High-level API (`api.ts`)**: Provides user-friendly functions for creating specific types of RPC messages. Each function returns a JSON-stringified, signed message ready to be sent over a WebSocket connection.

2. **Low-level Core (`nitrolite.ts`)**: Contains the `NitroliteRPC` class implementing core functionality for creating, signing, and parsing RPC messages according to the Nitrolite protocol specification.

## Using NitroliteRPC

The Nitrolite SDK provides convenient API functions for creating properly formatted and signed messages:

```typescript
import { createAuthRequestMessage, createPingMessage } from '@nitrolite/sdk/rpc';
import { MessageSigner } from '@nitrolite/sdk/rpc/types';

// Assuming you have a signing function that returns a Promise<Hex>
const signer: MessageSigner = async (payload) => {
    // Your signing implementation
    return '0x...'; // Signed message
};

// Create an auth request message
const authRequestMessage = await createAuthRequestMessage(signer, '0xYourAddress');

// Create a ping message
const pingMessage = await createPingMessage(signer);

// Send the messages via WebSocket
websocket.send(authRequestMessage);
websocket.send(pingMessage);
```

## Types

The module defines TypeScript interfaces and types for RPC messages, requests, and responses in `types.ts`. Key types include:

- `NitroliteRPCMessage`: Base interface for all RPC messages
- `ApplicationRPCMessage`: Interface for application-specific RPC messages
- `ParsedResponse`: Interface representing parsed RPC responses
- `MessageSigner`: Function type for signing message payloads
- `NitroliteErrorCode`: Enum of standard error codes for the Nitrolite RPC protocol

## Error Handling

Messages returned from the server can be parsed using `NitroliteRPC.parseResponse()`, which returns a `ParsedResponse` object containing validation status and extracted fields.

Error responses follow the JSON-RPC error format, with specific error codes defined in `NitroliteErrorCode`.

## Security Considerations

1. **Timestamp Validation**: Always check that incoming messages have a timestamp within an acceptable range to prevent replay attacks.
2. **Signature Verification**: Always verify signatures on received messages.
3. **Request IDs**: Track request IDs to associate responses with their original requests.
4. **Error Handling**: Implement robust error handling for network issues and invalid messages.
