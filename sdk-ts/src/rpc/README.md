# NitroliteRPC - Simple RPC for State Channels

A minimalist implementation of the NitroliteRPC protocol for communicating with state channel brokers.

## Overview

NitroliteRPC is a lightweight RPC protocol designed for state channels. Messages are formatted as fixed JSON arrays with a standard structure:

```
[request_id, method, params, timestamp]
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

### Errors

```json
{
  "err": [1001, -32601, "Method not found", 1741344819814],
  "sig": "0xd73268362b04516451ec52170f5c8ca189d35d9ac5e9041c156c9f0faf9aebd2891309e3b2b5d8788578ab3449c96f7aa81aefb25482b53f02bac42c65f806e5"
}
```

- `err`: RPC message payload `[request_id, error_code, error_message, timestamp]`
- `sig`: Payload signature

## Using NitroliteRPC

The `NitroliteRPC` class provides utilities for creating and signing these messages. It's designed to be simple and straightforward:

```typescript
import { NitroliteRPC, NitroliteRPCMessage } from '@erc7824/nitrolite';

// Create a request message
const request = NitroliteRPC.createRequest(
  'subtract',     // Method name
  [42, 23],       // Method parameters
  1001            // Optional: Request ID
);

// Sign the request with your own signer function
const signedRequest = await NitroliteRPC.signMessage(
  request,
  (message) => yourSigningFunction(message)
);

// Send the message via your own WebSocket connection
ws.send(JSON.stringify(signedRequest));
```

## Integrating with Your Application

NitroliteRPC is transport-agnostic, meaning you can use any WebSocket library or other transport mechanism. Here's how to integrate it:

1. **Create Messages**:
   ```typescript
   const request = NitroliteRPC.createRequest(method, params, requestId);
   const response = NitroliteRPC.createResponse(requestId, method, result);
   const error = NitroliteRPC.createError(requestId, errorCode, errorMessage);
   ```

2. **Sign Messages**:
   ```typescript
   const signedMessage = await NitroliteRPC.signMessage(message, yourSigningFunction);
   ```

3. **Verify Messages**:
   ```typescript
   const isValid = await NitroliteRPC.verifyMessage(
     message,
     expectedSignerAddress,
     yourVerifyFunction
   );
   ```

4. **Send and Receive**:
   ```typescript
   // Sending
   ws.send(JSON.stringify(signedMessage));
   
   // Receiving
   ws.onmessage = (event) => {
     const message = JSON.parse(event.data);
     // Check if it's a request, response, or error
     if (message.req) { /* ... */ }
     else if (message.res) { /* ... */ }
     else if (message.err) { /* ... */ }
   };
   ```

## Error Codes

Standard JSON-RPC error codes:

- `-32700`: Parse Error - Invalid JSON
- `-32600`: Invalid Request - Not a valid request object
- `-32601`: Method Not Found - Method doesn't exist
- `-32602`: Invalid Params - Invalid method parameters
- `-32603`: Internal Error - Internal JSON-RPC error

Nitrolite-specific error codes:

- `-32001`: Invalid State - Invalid state transition
- `-32002`: Channel Not Found - Referenced channel doesn't exist
- `-32003`: Invalid Signature - Signature verification failed