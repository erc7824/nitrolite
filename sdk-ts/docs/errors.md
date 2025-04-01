# Nitrolite SDK Error Handling Guide

The Nitrolite SDK provides a comprehensive error handling system to help developers identify and address issues that may arise during development and production. This guide covers best practices for handling errors, understanding error categories, and troubleshooting common issues.

## Error Architecture

All errors in the Nitrolite SDK extend from the base `NitroliteError` class, which provides:

- **code**: A unique string identifier for the error
- **statusCode**: An HTTP-like status code
- **suggestion**: A human-readable suggestion for resolving the error
- **details**: Additional context-specific information about the error

Errors are organized into categories to help with error handling and debugging.

## Error Categories

### Validation Errors

Errors related to invalid inputs or parameters:

| Error Class | Code | Status | Description |
|-------------|------|--------|-------------|
| `ValidationError` | `VALIDATION_ERROR` | 400 | Base validation error |
| `InvalidParameterError` | `INVALID_PARAMETER` | 400 | An input parameter is invalid |
| `MissingParameterError` | `MISSING_PARAMETER` | 400 | A required parameter is missing |
| `MethodNotFoundError` | `METHOD_NOT_FOUND` | 404 | The requested method doesn't exist |
| `InvalidRPCParamsError` | `INVALID_RPC_PARAMS` | 400 | RPC parameters are invalid |

### Authentication Errors

Errors related to authentication and authorization:

| Error Class | Code | Status | Description |
|-------------|------|--------|-------------|
| `AuthenticationError` | `AUTHENTICATION_ERROR` | 401 | Base authentication error |
| `InvalidSignatureError` | `INVALID_SIGNATURE` | 401 | A signature is invalid |
| `UnauthorizedError` | `UNAUTHORIZED` | 403 | Operation not authorized |
| `NotParticipantError` | `NOT_PARTICIPANT` | 403 | Address is not a channel participant |

### Network Errors

Errors related to network connectivity and timeouts:

| Error Class | Code | Status | Description |
|-------------|------|--------|-------------|
| `NetworkError` | `NETWORK_ERROR` | 500 | Base network error |
| `ConnectionError` | `CONNECTION_FAILED` | 503 | Connection to server failed |
| `ProviderNotConnectedError` | `PROVIDER_NOT_CONNECTED` | 503 | Provider not connected |
| `TimeoutError` | `TIMEOUT_ERROR` | 408 | Base timeout error |
| `RequestTimeoutError` | `REQUEST_TIMEOUT` | 408 | Request timed out |

### State Errors

Errors related to application state:

| Error Class | Code | Status | Description |
|-------------|------|--------|-------------|
| `StateError` | `STATE_ERROR` | 400 | Base state error |
| `InvalidStateTransitionError` | `INVALID_STATE_TRANSITION` | 400 | State transition is invalid |
| `StateNotFoundError` | `STATE_NOT_FOUND` | 404 | State not found |
| `StateNotInitializedError` | `STATE_NOT_INITIALIZED` | 400 | State not initialized |
| `ChannelNotFoundError` | `CHANNEL_NOT_FOUND` | 404 | Channel not found |

### Virtual Channel Errors

Errors related to virtual channels:

| Error Class | Code | Status | Description |
|-------------|------|--------|-------------|
| `VirtualChannelError` | `VIRTUAL_CHANNEL_ERROR` | 400 | Base virtual channel error |
| `NoNextHopError` | `NO_NEXT_HOP` | 404 | No next hop found in channel path |
| `RelayError` | `RELAY_FAILED` | 500 | Message relay failed |

### Contract Errors

Errors related to blockchain contracts:

| Error Class | Code | Status | Description |
|-------------|------|--------|-------------|
| `ContractError` | `CONTRACT_ERROR` | 500 | Base contract error |
| `ContractNotFoundError` | `CONTRACT_NOT_FOUND` | 404 | Contract not found |
| `ContractCallError` | `CONTRACT_CALL_FAILED` | 500 | Contract call failed |
| `TransactionError` | `TRANSACTION_FAILED` | 500 | Transaction failed |

### RPC Errors

Errors related to RPC protocol:

| Error Class | Code | Status | Description |
|-------------|------|--------|-------------|
| `RPCError` | `RPC_ERROR_*` | 500 | Generic RPC error |

## Standard RPC Error Codes

The SDK uses standard JSON-RPC error codes:

| Code | Constant | Description |
|------|----------|-------------|
| -32700 | `PARSE_ERROR` | Invalid JSON received |
| -32600 | `INVALID_REQUEST` | JSON not a valid Request object |
| -32601 | `METHOD_NOT_FOUND` | Method does not exist |
| -32602 | `INVALID_PARAMS` | Invalid method parameters |
| -32603 | `INTERNAL_ERROR` | Internal JSON-RPC error |
| -32000 | `SERVER_ERROR` | Generic server error |
| -32001 | `UNAUTHORIZED` | Not authorized to call method |
| -32002 | `INVALID_STATE` | Invalid state transition |
| -32003 | `CHANNEL_NOT_FOUND` | Channel not found |
| -32004 | `INVALID_SIGNATURE` | Invalid signature |
| -32005 | `INVALID_TRANSITION` | Invalid state transition |
| -32006 | `VIRTUAL_CHANNEL_ERROR` | Virtual channel error |
| -32007 | `TIMEOUT` | Operation timed out |

## Error Handling Best Practices

### Using Instanceof Checks for Specific Handling

```typescript
import { 
  NitroliteError, 
  TokenError, 
  TransactionError,
  NetworkError,
  ValidationError 
} from '@ethtaipei/Nitrolite-sdk-ts';

try {
  await client.openChannel(channel, initialState);
} catch (error) {
  if (error instanceof TokenError) {
    // Handle token-specific errors
    console.error(`Token error: ${error.message}, Suggestion: ${error.suggestion}`);
    if (error.code === 'INSUFFICIENT_BALANCE') {
      console.log('Please add more tokens to your wallet');
    } else if (error.code === 'INSUFFICIENT_ALLOWANCE') {
      console.log('Approve token spending and try again');
    }
  } else if (error instanceof TransactionError) {
    // Handle transaction errors
    console.error(`Transaction failed: ${error.message}`);
    console.log('Receipt details:', error.details?.receipt);
  } else if (error instanceof NetworkError) {
    // Handle any network-related error
    console.error(`Network issue: ${error.message}`);
    // Wait and retry later
  } else if (error instanceof ValidationError) {
    // Handle validation issues
    console.error(`Invalid input: ${error.message}`);
    // Fix parameters
  } else if (error instanceof NitroliteError) {
    // Handle general SDK errors
    console.error(`Error: ${error.message}, Code: ${error.code}`);
    console.error(`Suggestion: ${error.suggestion}`);
  } else {
    // Handle unexpected errors
    console.error('Unknown error:', error);
  }
}
```

### Implementing Retry Logic for Transient Errors

```typescript
import { NetworkError, TimeoutError } from '@ethtaipei/Nitrolite-sdk-ts';

async function withRetry(operation, maxRetries = 3) {
  let retries = 0;
  while (true) {
    try {
      return await operation();
    } catch (error) {
      if (
        error instanceof NetworkError || 
        error instanceof TimeoutError
      ) {
        if (retries >= maxRetries) {
          throw error;
        }
        retries++;
        console.warn(`Retrying operation (${retries}/${maxRetries})...`);
        await new Promise(resolve => setTimeout(resolve, 1000 * retries));
      } else {
        throw error;
      }
    }
  }
}

// Usage
await withRetry(() => client.openChannel(channel, initialState));
```

### Creating Context-Aware Error Handlers

```typescript
import { NitroliteError, ChannelNotFoundError, InvalidStateTransitionError } from '@ethtaipei/Nitrolite-sdk-ts';

// Create specialized error handlers for different operations
const channelErrorHandler = (operation) => async (...args) => {
  try {
    return await operation(...args);
  } catch (error) {
    if (error instanceof ChannelNotFoundError) {
      console.error(`Channel not found: ${error.message}`);
      // Attempt recovery - maybe create a new channel
      return await createNewChannel(...args);
    } else if (error instanceof InvalidStateTransitionError) {
      console.error(`Invalid state transition: ${error.message}`);
      // Maybe fetch latest state and retry
      return await retryWithLatestState(...args);
    } else if (error instanceof NitroliteError) {
      console.error(`Channel operation error: ${error.code} - ${error.message}`);
      console.error(`Suggestion: ${error.suggestion}`);
    }
    throw error; // Re-throw if not handled
  }
};

// Use the handler to wrap operations
const safeUpdateState = channelErrorHandler(channel.updateAppState.bind(channel));
await safeUpdateState(newState);
```

### Extracting and Logging Detailed Error Data

```typescript
import { NitroliteError, VirtualChannelError } from '@ethtaipei/Nitrolite-sdk-ts';

function logDetailedError(error) {
  if (error instanceof NitroliteError) {
    console.error(`
      Error: ${error.name} [${error.code}]
      Message: ${error.message}
      Status: ${error.statusCode}
      Suggestion: ${error.suggestion}
    `);
    
    // Log specific details based on error type
    if (error instanceof VirtualChannelError) {
      console.error('Virtual Channel Details:');
      if (error.details?.lvci) {
        console.error(`- Channel ID: ${error.details.lvci}`);
      }
      if (error.details?.position !== undefined) {
        console.error(`- Position: ${error.details.position}`);
      }
      if (error.details?.nextHop) {
        console.error(`- Failed at hop: ${error.details.nextHop}`);
      }
    }
    
    // Log other details
    if (error.details?.cause) {
      console.error('Root cause:', error.details.cause);
    }
  } else {
    console.error('Unexpected error:', error);
  }
}
```

## Common Error Scenarios and Solutions

### Channel Opening Errors

Common issues when opening channels:

```typescript
try {
  const channelId = await client.openChannel(channel, initialState);
  console.log(`Channel opened with ID: ${channelId}`);
} catch (error) {
  if (error instanceof TokenError) {
    // Token issues
    if (error.code === 'INSUFFICIENT_BALANCE') {
      console.error('Not enough tokens to fund the channel');
      console.log('Current balance:', error.details?.balance);
      console.log('Required amount:', error.details?.requiredAmount);
    } else if (error.code === 'INSUFFICIENT_ALLOWANCE') {
      console.error('Need to approve token spending');
      // Approve tokens and retry
      const amount = error.details?.requiredAmount || BigInt(1000);
      await client.approveTokens(error.details?.token, amount, client.custodyAddress);
      // Retry opening the channel
      await client.openChannel(channel, initialState);
    }
  } else if (error instanceof ContractCallError) {
    console.error('Contract interaction failed:', error.message);
    console.error('Suggestion:', error.suggestion);
    // Check gas limit, contract parameters, etc.
  } else if (error instanceof ValidationError) {
    console.error('Invalid channel parameters:', error.message);
    // Fix channel configuration and retry
  }
}
```

### State Update Errors

Handling errors during state updates:

```typescript
try {
  await rpcChannel.updateAppState(newState);
} catch (error) {
  if (error instanceof InvalidStateTransitionError) {
    console.error('Invalid state transition:', error.message);
    
    // Get the latest state and merge changes
    const currentState = rpcChannel.getCurrentAppState();
    console.log('Current state:', currentState);
    
    // Create a corrected state based on the current one
    const correctedState = {
      ...currentState,
      // Apply only valid changes
      value: Math.max(currentState.value, newState.value),
      sequence: currentState.sequence + 1n
    };
    
    // Retry with corrected state
    await rpcChannel.updateAppState(correctedState);
    
  } else if (error instanceof ConnectionError) {
    console.error('Connection lost during update:', error.message);
    // Try to reconnect RPC client
    await rpcClient.connect();
    // Then retry update
    await rpcChannel.updateAppState(newState);
  }
}
```

### Virtual Channel Routing Errors

Resolving issues with virtual channels:

```typescript
try {
  await rpcClient.relayStateUpdate(lvci, state);
} catch (error) {
  if (error instanceof NoNextHopError) {
    console.error('No valid next hop found:', error.message);
    console.log('Current position:', error.details?.position);
    console.log('Path:', error.details?.path);
    
    // Check all participants are connected
    const connectedAddresses = await getConnectedParticipants();
    const missingParticipants = lvci.path.filter(addr => !connectedAddresses.includes(addr));
    
    if (missingParticipants.length > 0) {
      console.error('Missing participants:', missingParticipants);
      // Notify users about missing participants
    }
  } else if (error instanceof RelayError) {
    console.error('Failed to relay message:', error.message);
    console.log('Failed at hop:', error.details?.nextHop);
    
    // Check specific participant connection
    const isHopConnected = await checkParticipantConnection(error.details?.nextHop);
    if (!isHopConnected) {
      console.error('Intermediary is offline:', error.details?.nextHop);
      // Maybe try an alternative path
    }
  }
}
```

### Token Approval and Transaction Errors

Handling blockchain transaction issues:

```typescript
try {
  await client.approveTokens(tokenAddress, amount, spender);
} catch (error) {
  if (error instanceof TransactionError) {
    console.error('Transaction failed:', error.message);
    
    // Check transaction receipt for more details
    if (error.details?.receipt) {
      console.log('Gas used:', error.details.receipt.gasUsed);
      console.log('Status:', error.details.receipt.status);
    }
    
    // Check if gas price is too low
    if (error.message.includes('underpriced')) {
      console.log('Transaction underpriced, retrying with higher gas...');
      // Retry with higher gas price
      await client.approveTokens(tokenAddress, amount, spender, {
        gasPrice: increasedGasPrice
      });
    }
  } else if (error instanceof ContractCallError) {
    console.error('Contract call simulation failed:', error.message);
    // This usually means the transaction would fail if sent
    console.log('Reason:', error.details?.cause?.message);
  }
}

## Advanced Error Handling Techniques

### Error Configuration Options

You can adjust error-related settings in the SDK configuration:

```typescript
import { getConfigWithDefaults } from '@ethtaipei/Nitrolite-sdk-ts';

const config = getConfigWithDefaults({
  // Increase request timeout
  requestTimeoutMs: 60000,
  
  // Increase maximum retries
  maxRequestRetries: 5,
  
  // Enable detailed logging for debugging
  logLevel: 'debug'
});

// Pass config to SDK components
const client = new RPCClient({
  provider,
  address,
  signer,
  ...config
});
```

### Implementing a Global Error Handler

For applications with many channel operations, a global error handler can be useful:

```typescript
import { 
  NitroliteError, 
  NetworkError,
  TimeoutError,
  ContractError,
  StateError,
  VirtualChannelError
} from '@ethtaipei/Nitrolite-sdk-ts';

// Create a global error handler
class NitroliteErrorHandler {
  constructor(options = {}) {
    this.options = {
      maxRetries: 3,
      retryDelay: 1000,
      shouldLogErrors: true,
      onNetworkError: null,
      onContractError: null,
      ...options
    };
  }
  
  // Wrap an operation with error handling
  async handle(operation, context = {}) {
    let retries = 0;
    
    while (true) {
      try {
        return await operation();
      } catch (error) {
        // Log the error if enabled
        if (this.options.shouldLogErrors) {
          this.logError(error, context);
        }
        
        // Handle retryable errors
        if (this.isRetryableError(error) && retries < this.options.maxRetries) {
          retries++;
          console.warn(`Retrying operation (${retries}/${this.options.maxRetries})...`);
          await new Promise(resolve => setTimeout(resolve, this.options.retryDelay * retries));
          continue;
        }
        
        // Handle specific error categories
        if (error instanceof NetworkError && this.options.onNetworkError) {
          await this.options.onNetworkError(error, context);
        } else if (error instanceof ContractError && this.options.onContractError) {
          await this.options.onContractError(error, context);
        } else if (error instanceof StateError && this.options.onStateError) {
          await this.options.onStateError(error, context);
        } else if (error instanceof VirtualChannelError && this.options.onVirtualChannelError) {
          await this.options.onVirtualChannelError(error, context);
        }
        
        // Re-throw the error for the caller to handle
        throw error;
      }
    }
  }
  
  // Determine if an error is retryable
  isRetryableError(error) {
    return (
      error instanceof NetworkError ||
      error instanceof TimeoutError ||
      (error instanceof ContractError && 
        (error.message.includes('nonce') || error.message.includes('underpriced')))
    );
  }
  
  // Detailed error logging
  logError(error, context) {
    if (error instanceof NitroliteError) {
      console.error(`
        [${new Date().toISOString()}] ${error.name} [${error.code}]
        Message: ${error.message}
        Status: ${error.statusCode}
        Suggestion: ${error.suggestion}
        Context: ${JSON.stringify(context)}
        ${error.details ? `Details: ${JSON.stringify(error.details, null, 2)}` : ''}
      `);
    } else {
      console.error(`
        [${new Date().toISOString()}] Unexpected Error
        Message: ${error.message}
        Stack: ${error.stack}
        Context: ${JSON.stringify(context)}
      `);
    }
  }
}

// Usage
const errorHandler = new NitroliteErrorHandler({
  onNetworkError: async (error, context) => {
    // Reconnect logic
    if (context.client) {
      await context.client.connect();
    }
  },
  onContractError: async (error, context) => {
    // Notify user about blockchain issues
    notifyUser('Blockchain operation failed: ' + error.message);
  }
});

// Wrap operations with the handler
try {
  await errorHandler.handle(
    () => client.openChannel(channel, initialState),
    { client, channel, operation: 'openChannel' }
  );
} catch (error) {
  // Handle unrecoverable errors
  console.error('Operation failed after all recovery attempts:', error.message);
}
```