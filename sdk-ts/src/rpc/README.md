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

## Building a WebSocket Client

For WebSocket communication, you'll need to implement a client that can:

1. **Connect to the broker server**: Establish a WebSocket connection
2. **Authenticate**: Sign and send an authentication message using your Ethereum key
3. **Process messages**: Handle incoming requests, responses, and errors
4. **Send messages**: Create, sign, and send NitroliteRPC messages

Here's a simplified example:

```typescript
import { NitroliteRPC } from '@erc7824/nitrolite';
import { Hex } from 'viem';

// Interface for the wallet signer
interface WalletSigner {
  publicKey: string;
  address?: string;
  sign: (message: string) => Promise<Hex>;
}

class WebSocketClient {
  private ws: WebSocket | null = null;
  private signer: WalletSigner;
  
  constructor(url: string, signer: WalletSigner) {
    this.signer = signer;
  }
  
  async connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.ws = new WebSocket(url);
      
      this.ws.onopen = async () => {
        try {
          // Authenticate upon connection
          await this.authenticate();
          resolve();
        } catch (error) {
          reject(error);
        }
      };
      
      // Set up message handling
      this.ws.onmessage = (event) => {
        const response = JSON.parse(event.data);
        // Handle messages...
      };
    });
  }
  
  private async authenticate(): Promise<void> {
    // Create and send authentication request
    const authRequest = NitroliteRPC.createRequest(
      'auth', 
      [this.signer.publicKey]
    );
    
    const signedRequest = await NitroliteRPC.signMessage(
      authRequest,
      this.signer.sign
    );
    
    this.ws!.send(JSON.stringify(signedRequest));
    
    // Wait for auth response...
  }
  
  async sendRequest(method: string, params: any[] = []): Promise<any> {
    const request = NitroliteRPC.createRequest(method, params);
    const signedRequest = await NitroliteRPC.signMessage(request, this.signer.sign);
    
    // Send and wait for response...
    this.ws!.send(JSON.stringify(signedRequest));
  }
}
```

For a complete implementation, see our [NextJS example](https://github.com/erc7824/nitrolite/tree/main/sdk-ts/examples/nextjs-ts-example).

## Common Message Types

Nitrolite applications typically use these standard RPC method types:

| Method | Description | Parameters |
|--------|-------------|------------|
| `auth` | Authenticate with the broker | `[publicKey]` |
| `subscribe` | Subscribe to a channel | `[channelName]` |
| `publish` | Publish a message to a channel | `[channelName, messageData]` |
| `ping` | Check connection latency | `[timestamp]` |
| `get_balance` | Check token balance | `[tokenAddress]` |
| `state_update` | Update channel state | `[channelId, stateData, signature]` |

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

## Security Considerations

1. **Timestamp Validation**: Always check that incoming messages have a timestamp within an acceptable range to prevent replay attacks.
2. **Signature Verification**: Always verify signatures on received messages.
3. **Request IDs**: Track request IDs to associate responses with their original requests.
4. **Error Handling**: Implement robust error handling for network issues and invalid messages.