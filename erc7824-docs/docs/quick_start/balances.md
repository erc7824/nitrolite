---
sidebar_position: 6
title: Channel Asset Management
description: Monitor off-chain balances in your active state channels using NitroliteRPC.
keywords: [erc7824, nitrolite, balances, off-chain, ledger balances, clearnode]
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Channel Asset Management

After connecting to a ClearNode, you'll need to monitor the off-chain balances in your state channels. This guide explains how to retrieve and work with off-chain balance information using the NitroliteRPC protocol.

## Understanding Off-Chain Balances

Off-chain balances in Nitrolite represent:

- Your current funds in the state channel
- Balances that update in real-time as transactions occur
- The source of truth for application operations
- Assets that are backed by on-chain deposits

## Checking Off-Chain Balances

To monitor your channel funds, you need to retrieve the current off-chain balances from the ClearNode.

## Understanding the Ledger Balances Request

The `get_ledger_balances` request is used to retrieve the current off-chain balances for a specific participant from the ClearNode:

- **Request params**: `[{ participant: "0xAddress" }]` where `0xAddress` is the participant's address
- **Response**: Array containing the balances for different assets held by the participant

The response contains a list of assets and their amounts for the specified participant. The balances are represented as strings with decimal precision, making it easier to display them directly without additional conversion.

```javascript
// Example response format for get_ledger_balances
{
  "res": [1, "get_ledger_balances", [[  // The nested array contains balance data
    {
      "asset": "usdc",  // Asset identifier
      "amount": "100.0"  // Amount as a string with decimal precision
    },
    {
      "asset": "eth",
      "amount": "0.5"
    }
  ]], 1619123456789],  // Timestamp
  "sig": ["0xabcd1234..."]
}
```

To retrieve these balances, use the `get_ledger_balances` request with the ClearNode:

<Tabs>
  <TabItem value="using-helper" label="Using SDK Helper">

```javascript
import { createGetLedgerBalancesMessage, parseRPCMessage, RPCMethod } from '@erc7824/nitrolite';
import { ethers } from 'ethers';

// Your message signer function (same as in auth flow)
const messageSigner = async (payload) => {
  const message = JSON.stringify(payload);
  const digestHex = ethers.id(message);
  const messageBytes = ethers.getBytes(digestHex);
  const { serialized: signature } = stateWallet.signingKey.sign(messageBytes);
  return signature;
};

// Function to get ledger balances
async function getLedgerBalances(ws, participant) {
  return new Promise((resolve, reject) => {
    // Create a unique handler for this specific request
    const handleMessage = (event) => {
      const message = parseRPCMessage(event.data);
      
      // Check if this is a response to our get_ledger_balances request
      if (message.method === RPCMethod.GetLedgerBalances) {
        // Remove the message handler to avoid memory leaks
        ws.removeEventListener('message', handleMessage);
        
        // Resolve with the balances data
        resolve(message.params);
      }
    };
    
    // Add the message handler
    ws.addEventListener('message', handleMessage);
    
    // Create and send the ledger balances request
    createGetLedgerBalancesMessage(messageSigner, participant)
      .then(message => {
        ws.send(message);
      })
      .catch(error => {
        // Remove the message handler on error
        ws.removeEventListener('message', handleMessage);
        reject(error);
      });
      
    // Set a timeout to prevent hanging indefinitely
    setTimeout(() => {
      ws.removeEventListener('message', handleMessage);
      reject(new Error('Timeout waiting for ledger balances'));
    }, 10000); // 10 second timeout
  });
}

// Usage example
const participantAddress = '0x1234567890abcdef1234567890abcdef12345678';

try {
  const balances = await getLedgerBalances(ws, participantAddress);
  
  console.log('Channel ledger balances:', balances);
  // Example output:
  // [
  //   [
  //     { "asset": "usdc", "amount": "100.0" },
  //     { "asset": "eth", "amount": "0.5" }
  //   ]
  // ]
  
  // Process your balances
  if (balances.length > 0) {
    // Display each asset balance
    balances.forEach(balance => {
      console.log(`${balance.asset.toUpperCase()} balance: ${balance.amount}`);
    });
    
    // Example: find a specific asset
    const usdcBalance = balances.find(b => b.asset.toLowerCase() === 'usdc');
    if (usdcBalance) {
      console.log(`USDC balance: ${usdcBalance.amount}`);
    }
  } else {
    console.log('No balance data available');
  }
} catch (error) {
  console.error('Failed to get ledger balances:', error);
}
```

  </TabItem>
  <TabItem value="manual" label="Manual Request">

```javascript
import { ethers } from 'ethers';
import { generateRequestId, getCurrentTimestamp, generateRequestId, parseRPCMessage, RPCMethod } from '@erc7824/nitrolite';

// Function to create a signed ledger balances request
async function createLedgerBalancesRequest(signer, participant) {
  const requestId = generateRequestId();
  const method = RPCMethod.GetLedgerBalances; // Use the RPC method enum for clarity
  const params = [{ participant }]; // Note: updated parameter name to 'participant'
  const timestamp = getCurrentTimestamp();
  
  // Create the request structure
  const requestData = [requestId, method, params, timestamp];
  const request = { req: requestData };
  
  // Sign the request
  const message = JSON.stringify(request);
  const digestHex = ethers.id(message);
  const messageBytes = ethers.getBytes(digestHex);
  const { serialized: signature } = signer.wallet.signingKey.sign(messageBytes);
  
  // Add signature to the request
  request.sig = [signature];
  
  return { stringified: JSON.stringify(request), requestId };
}

// Function to get ledger balances
async function getLedgerBalances(ws, participant, signer) {
  return new Promise((resolve, reject) => {
    createLedgerBalancesRequest(signer, participant)
      .then(({ stringified, requestId }) => {
        // Set up message handler
        const handleMessage = (event) => {
          try {
            const message = parseRPCMessage(event.data);
            
            // Check if this is our response
            if (message.requestId === requestId && 
                message.method === RPCMethod.GetLedgerBalances) {
              
              // Remove the listener
              ws.removeEventListener('message', handleMessage);
              
              // Resolve with the balances data
              resolve(message.params);
            }
          } catch (error) {
            console.error('Error parsing message:', error);
          }
        };
        
        // Add message listener
        ws.addEventListener('message', handleMessage);
        
        // Send the request
        ws.send(stringified);
        
        // Set timeout
        setTimeout(() => {
          ws.removeEventListener('message', handleMessage);
          reject(new Error('Timeout waiting for ledger balances'));
        }, 10000);
      })
      .catch(reject);
  });
}

// Usage example
const participantAddress = '0x1234567890abcdef1234567890abcdef12345678';

try {
  const balances = await getLedgerBalances(ws, participantAddress, stateWallet);
  
  console.log('Channel ledger balances:', balances);
  // Process and display balances
  // ...
  
} catch (error) {
  console.error('Failed to get ledger balances:', error);
}
```

  </TabItem>
</Tabs>

## Checking Balances for a Participant

To retrieve off-chain balances for a participant, use the `createGetLedgerBalancesMessage` helper function:

```javascript
import { createGetLedgerBalancesMessage, parseRPCResponse, RPCMethod } from '@erc7824/nitrolite';
import { ethers } from 'ethers';

// Function to get ledger balances for a participant
async function getLedgerBalances(ws, participant, messageSigner) {
  return new Promise((resolve, reject) => {
    // Message handler for the response
    const handleMessage = (event) => {
      try {
        const message = parseRPCResponse(event.data);
        
        // Check if this is a response to our get_ledger_balances request
        if (message.method === RPCMethod.GetLedgerBalances) {
          // Clean up by removing the event listener
          ws.removeEventListener('message', handleMessage);
          
          // Resolve with the balance data
          resolve(message.params);
        }
      } catch (error) {
        console.error('Error parsing message:', error);
      }
    };
    
    // Set up timeout to avoid hanging indefinitely
    const timeoutId = setTimeout(() => {
      ws.removeEventListener('message', handleMessage);
      reject(new Error('Timeout waiting for ledger balances'));
    }, 10000); // 10 second timeout
    
    // Add the message handler
    ws.addEventListener('message', handleMessage);
    
    // Create and send the balance request
    createGetLedgerBalancesMessage(messageSigner, participant)
      .then(message => {
        ws.send(message);
      })
      .catch(error => {
        clearTimeout(timeoutId);
        ws.removeEventListener('message', handleMessage);
        reject(error);
      });
  });
}

// Example usage
const participantAddress = '0x1234567890abcdef1234567890abcdef12345678';

getLedgerBalances(ws, participantAddress, messageSigner)
  .then(balances => {
    console.log('Channel balances:', balances);
    
    // Process and display your balances
    if (balances.length > 0) {
      console.log('My balances:');
      balances.forEach(balance => {
        console.log(`- ${balance.asset.toUpperCase()}: ${balance.amount}`);
      });
    } else {
      console.log('No balance data available');
    }
  })
  .catch(error => {
    console.error('Failed to get ledger balances:', error);
  });
```

## Processing Balance Data

When you receive balance data from the ClearNode, it's helpful to format it for better readability:

```javascript
// Simple function to format your balance data for display
function formatMyBalances(balances) {
  // Return formatted balance information
  return balances.map(balance => ({
    asset: balance.asset.toUpperCase(),
    amount: balance.amount,
    // You can add additional formatting here if needed
    displayAmount: `${balance.amount} ${balance.asset.toUpperCase()}`
  }));
}

// Example usage
const myFormattedBalances = formatMyBalances(balancesFromClearNode);

if (myFormattedBalances && myFormattedBalances.length > 0) {
  console.log('My balances:');
  myFormattedBalances.forEach(balance => {
    console.log(`- ${balance.displayAmount}`);
  });
} else {
  console.log('No balance data available');
}
```

## Best Practices for Balance Checking

When working with off-chain balances, follow these best practices:

### Regular Balance Polling

Set up a regular interval to check balances, especially in active applications:

```javascript
// Simple balance monitoring function
function startBalanceMonitoring(ws, participantAddress, messageSigner, intervalMs = 30000) {
  // Check immediately on start
  getLedgerBalances(ws, participantAddress, messageSigner)
    .then(displayBalances)
    .catch(err => console.error('Initial balance check failed:', err));
  
  // Set up interval for regular checks
  const intervalId = setInterval(() => {
    getLedgerBalances(ws, participantAddress, messageSigner)
      .then(displayBalances)
      .catch(err => console.error('Balance check failed:', err));
  }, intervalMs); // Check every 30 seconds by default
  
  // Return function to stop monitoring
  return () => clearInterval(intervalId);
}

// Simple display function
function displayBalances(balances) {
  console.log(`Balance update at ${new Date().toLocaleTimeString()}:`);
  
  // Format and display your balances
  if (balances.length > 0) {
    console.log('My balances:');
    balances.forEach(balance => {
      console.log(`- ${balance.asset.toUpperCase()}: ${balance.amount}`);
    });
  } else {
    console.log('No balance data available');
  }
}
```

## Common Errors and Troubleshooting

When retrieving off-chain balances, you might encounter these common issues:

| Error Type | Description | Solution |
|------------|-------------|----------|
| **Authentication errors** | WebSocket connection loses authentication | Re-authenticate before requesting balances again |
| **Channel not found** | The channel ID is invalid or the channel has been closed | Verify the channel ID and check if the channel is still active |
| **Connection issues** | WebSocket disconnects during a balance request | Implement reconnection logic with exponential backoff |
| **Timeout** | The ClearNode does not respond in a timely manner | Set appropriate timeouts and implement retry logic |

## Next Steps

Now that you understand how to monitor off-chain balances in your channels, you can:

1. [Create an application session](application_session) to start transacting off-chain 
2. Learn about [channel closing](close_session) when you're done with the channel