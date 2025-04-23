# Nitrolite SDK Error Handling Guide

The Nitrolite SDK provides a comprehensive error handling system that helps you identify and address issues that occur during development and production. All errors extend from the `NitroliteError` base class and provide detailed information to assist with troubleshooting.

## Table of Contents

- [Error Structure](#error-structure)
- [Error Categories](#error-categories)
  - [Validation Errors](#validation-errors)
  - [Authentication Errors](#authentication-errors)
  - [Contract Errors](#contract-errors)
  - [Token Errors](#token-errors)
  - [State Errors](#state-errors)
- [Error Handling Examples](#error-handling-examples)
  - [Basic Error Handling](#basic-error-handling)
  - [Advanced Error Handling](#advanced-error-handling)
- [Common Error Scenarios](#common-error-scenarios)
  - [Deposit Failures](#deposit-failures)
  - [Channel Creation Issues](#channel-creation-issues)
  - [Checkpoint and Challenge Errors](#checkpoint-and-challenge-errors)
- [Error Prevention Best Practices](#error-prevention-best-practices)

## Error Structure

Each error in the SDK includes:

| Property | Description |
|----------|-------------|
| `name` | The class name of the error (e.g., `ContractCallError`) |
| `code` | A unique string identifier (e.g., `CONTRACT_CALL_FAILED`) |
| `message` | A detailed description of what went wrong |
| `statusCode` | An HTTP-like status code (e.g., `400`, `500`) |
| `suggestion` | A human-readable suggestion for resolving the error |
| `details` | Optional context-specific data about the error |
| `cause` | Optional original error that triggered this one |

## Error Categories

### Validation Errors

Errors related to invalid inputs or parameters:

| Error Class | Code | Description |
|-------------|------|-------------|
| `ValidationError` | `VALIDATION_ERROR` | Base validation error |
| `InvalidParameterError` | `INVALID_PARAMETER` | An input parameter is invalid |
| `MissingParameterError` | `MISSING_PARAMETER` | A required parameter is missing |

### Authentication Errors

Errors related to authentication and authorization:

| Error Class | Code | Description |
|-------------|------|-------------|
| `AuthenticationError` | `AUTHENTICATION_ERROR` | Base authentication error |
| `InvalidSignatureError` | `INVALID_SIGNATURE` | A signature is invalid |
| `UnauthorizedError` | `UNAUTHORIZED` | Operation not authorized |
| `NotParticipantError` | `NOT_PARTICIPANT` | Address is not a channel participant |
| `WalletClientRequiredError` | `WALLET_CLIENT_REQUIRED` | Operation requires a wallet client |
| `AccountRequiredError` | `ACCOUNT_REQUIRED` | Operation requires an account |

### Contract Errors

Errors related to blockchain and smart contract interactions:

| Error Class | Code | Description |
|-------------|------|-------------|
| `ContractError` | `CONTRACT_ERROR` | Base contract error |
| `ContractNotFoundError` | `CONTRACT_NOT_FOUND` | Contract not found at address |
| `ContractReadError` | `CONTRACT_READ_FAILED` | Reading from contract failed |
| `ContractCallError` | `CONTRACT_CALL_FAILED` | Contract call simulation failed |
| `TransactionError` | `TRANSACTION_FAILED` | On-chain transaction failed |

### Token Errors

Errors related to ERC20 tokens:

| Error Class | Code | Description |
|-------------|------|-------------|
| `TokenError` | `TOKEN_ERROR` | Base token operation error |
| `InsufficientBalanceError` | `INSUFFICIENT_BALANCE` | Insufficient token balance |
| `InsufficientAllowanceError` | `INSUFFICIENT_ALLOWANCE` | Insufficient token allowance |

### State Errors

Errors related to channel state:

| Error Class | Code | Description |
|-------------|------|-------------|
| `StateError` | `STATE_ERROR` | Base state error |
| `InvalidStateTransitionError` | `INVALID_STATE_TRANSITION` | Invalid state transition |
| `StateNotFoundError` | `STATE_NOT_FOUND` | State not found |
| `ChannelNotFoundError` | `CHANNEL_NOT_FOUND` | Channel not found |

## Error Handling Examples

### Basic Error Handling

```typescript
import { 
  NitroliteError, 
  TokenError,
  ContractError,
  InsufficientAllowanceError
} from '@erc7824/nitrolite';

try {
  await client.deposit(amount);
} catch (error) {
  if (error instanceof TokenError) {
    // Handle token errors
    console.error(`Token error: ${error.message}`);
    console.error(`Suggestion: ${error.suggestion}`);
    
    if (error instanceof InsufficientAllowanceError) {
      console.log("Approving tokens and retrying...");
      await client.approveTokens(amount);
      await client.deposit(amount);
    }
  } else if (error instanceof ContractError) {
    // Handle contract errors
    console.error(`Contract error: ${error.message}`);
  } else if (error instanceof NitroliteError) {
    // Handle other SDK errors
    console.error(`Error: ${error.code} - ${error.message}`);
    console.error(`Suggestion: ${error.suggestion}`);
  } else {
    // Handle unexpected errors
    console.error("Unexpected error:", error);
  }
}
```

### Advanced Error Handling

Handle specific error situations with custom logic:

```typescript
try {
  await client.createChannel(params);
} catch (error) {
  if (error instanceof InsufficientBalanceError) {
    const { required, actual } = error.details || {};
    console.error(`Insufficient balance. Required: ${required}, Available: ${actual}`);
    
    // Display UI for depositing more funds
    showDepositUI(required - actual);
  } else if (error instanceof ContractCallError) {
    console.error(`Contract call failed: ${error.message}`);
    
    // Check for gas-related issues
    if (error.message.includes('gas')) {
      console.log("Try again with higher gas limit");
    }
  } else if (error instanceof TransactionError) {
    // Get information about the failed transaction
    const { hash } = error.details || {};
    if (hash) {
      console.log(`Transaction failed. Check explorer: ${getExplorerUrl(hash)}`);
    }
  } else if (error instanceof MissingParameterError) {
    console.error(`Missing parameter: ${error.message}`);
    
    // Highlight the field in the UI
    highlightMissingField(error.message);
  }
}
```

## Common Error Scenarios

### Deposit Failures

```typescript
try {
  await client.deposit(amount);
} catch (error) {
  if (error instanceof InsufficientBalanceError) {
    console.error("Not enough tokens in wallet");
    // Show current balance vs required amount
    console.log(`Required: ${error.details?.required}, Available: ${error.details?.actual}`);
  } else if (error instanceof InsufficientAllowanceError) {
    // Token needs approval
    console.log(`Current allowance: ${error.details?.actual}, Required: ${error.details?.required}`);
    await client.approveTokens(amount);
    await client.deposit(amount); // Retry
  }
}
```

### Channel Creation Issues

```typescript
try {
  await client.createChannel(params);
} catch (error) {
  if (error.code === "CONTRACT_CALL_FAILED") {
    console.error("Failed to create channel:", error.message);
    console.log("Suggestion:", error.suggestion);
  } else if (error.code === "INVALID_PARAMETER") {
    console.error("Invalid parameters:", error.message);
    
    // Check for specific parameter issues
    if (error.message.includes("participants")) {
      console.log("Please check the participant addresses");
    } else if (error.message.includes("allocation")) {
      console.log("Please check the allocation amounts");
    }
  } else if (error.code === "MISSING_PARAMETER") {
    // This handles cases where required configuration is missing
    if (error.message.includes("adjudicator")) {
      console.error("The adjudicator address is missing in the configuration");
    }
  }
}
```

### Checkpoint and Challenge Errors

```typescript
try {
  await client.checkpointChannel(params);
} catch (error) {
  if (error.code === "INVALID_SIGNATURE") {
    console.error("Invalid signatures on state");
    // Check if states are properly signed
    if (params.candidateState.sigs.length < 2) {
      console.log("State must be signed by both participants");
    }
  } else if (error.code === "CHANNEL_NOT_FOUND") {
    console.error("Channel does not exist on-chain");
    console.log("Verify the channel ID:", params.channelId);
  } else if (error instanceof ContractCallError) {
    console.error("Contract call failed during checkpoint");
    // Check if the channel is in a valid state for checkpointing
    console.log("Verify that the channel is active and not closed");
  }
}
```

## Error Prevention Best Practices

1. **Validate inputs**: Check parameters before sending to methods
   ```typescript
   if (!channelId) {
     throw new Errors.MissingParameterError('channelId');
   }
   ```

2. **Check balances**: Verify sufficient funds before deposits/channel creation
   ```typescript
   const balance = await client.getTokenBalance();
   if (balance < amount) {
     throw new Errors.InsufficientBalanceError(tokenAddress, amount, balance);
   }
   ```

3. **Verify signatures**: Ensure all required signatures are present
   ```typescript
   if (!state.sigs || state.sigs.length < 2) {
     throw new Errors.InvalidSignatureError('State must be signed by both participants');
   }
   ```

4. **Handle network issues**: Implement retry logic for network-related errors
   ```typescript
   const MAX_RETRIES = 3;
   let attempt = 0;
   
   while (attempt < MAX_RETRIES) {
     try {
       return await client.deposit(amount);
     } catch (error) {
       if (error instanceof TransactionError && error.message.includes('network')) {
         attempt++;
         await new Promise(r => setTimeout(r, 1000 * attempt));
         continue;
       }
       throw error;
     }
   }
   ```

5. **Use proper error handling**: Leverage the error hierarchy for targeted handling
   ```typescript
   try {
     // Your code here
   } catch (error) {
     if (error instanceof TokenError) {
       // Handle token errors
     } else if (error instanceof StateError) {
       // Handle state errors
     } else if (error instanceof ContractError) {
       // Handle contract errors
     } else if (error instanceof NitroliteError) {
       // Handle other SDK errors
     } else {
       // Handle unexpected errors
     }
   }
   ```

By following these guidelines and leveraging the SDK's structured error system, you can create more robust applications that gracefully handle errors and provide clear feedback to users.