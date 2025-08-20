---
sidebar_position: 2
title: Migration Guide
description: Guide to migrate to newer versions of Nitrolite
keywords: [migration, upgrade, breaking changes, nitrolite, erc7824]
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Migration Guide

If you are coming from an earlier version of Nitrolite, you will need to make sure to update the following APIs listed below.

## 0.3.x Breaking changes

The 0.3.x release includes breaking changes to the SDK architecture, smart contract interfaces, and Clearnode API enhancements listed below.

Not ready to migrate? You can pin your dependencies to the previous version.

### Contracts

#### Signature Format Changes

The contract interfaces now use `bytes` for signatures instead of the `Signature` struct to support various signature formats.

**1. Removed `Signature` Struct**

```solidity
struct Signature { // [!code --]
    uint8 v; // [!code --]
    bytes32 r; // [!code --]
    bytes32 s; // [!code --]
} // [!code --]
```

**2. Updated Function Signatures**

<Tabs>
  <TabItem value="before" label="Before">

  ```solidity
  function join(
    bytes32 channelId,
    uint256 index,
    Signature calldata sig // [!code --]
  ) external returns (bytes32);
  
  function challenge(
    bytes32 channelId,
    State calldata candidate,
    State[] calldata proofs,
    Signature calldata challengerSig // [!code --]
  ) external;
  ```

  </TabItem>
  <TabItem value="after" label="After">

  ```solidity
  function join(
    bytes32 channelId,
    uint256 index,
    bytes calldata sig // [!code ++]
  ) external returns (bytes32);
  
  function challenge(
    bytes32 channelId,
    State calldata candidate,
    State[] calldata proofs,
    bytes calldata challengerSig // [!code ++]
  ) external;
  ```

  </TabItem>
</Tabs>

**3. State Signatures Array**

<Tabs>
  <TabItem value="before" label="Before">

  ```solidity
  struct State {
    uint8 intent;
    uint256 version;
    bytes data;
    Allocation[] allocations;
    Signature[] sigs; // [!code --]
  }
  ```

  </TabItem>
  <TabItem value="after" label="After">

  ```solidity
  struct State {
    uint8 intent;
    uint256 version;
    bytes data;
    Allocation[] allocations;
    bytes[] sigs; // [!code ++]
  }
  ```

  </TabItem>
</Tabs>

#### New Create/Join Flow in Custody Contract

The Custody contract has been updated with a new create/join flow that changes how channels are initialized.

<Tabs>
  <TabItem value="before" label="Before">

  ```solidity
  // Single-step channel creation
  function createChannel(
    Channel calldata channel,
    State calldata initialState
  ) external payable returns (bytes32);
  ```

  </TabItem>
  <TabItem value="after" label="After">

  ```solidity
  // Two-step process: create then join
  function create(
    Channel calldata channel,
    State calldata initialState
  ) external payable returns (bytes32);
  
  function join(
    bytes32 channelId,
    uint256 index,
    bytes calldata sig
  ) external returns (bytes32);
  ```

  </TabItem>
</Tabs>

### Clearnode

#### Request/Response Structure Changes

Clearnode API has changed from array-based to object-based request and response structures for better type safety and clarity.

<Tabs>
  <TabItem value="before" label="Before">

  ```json
  {
    "req": [1, "auth_request", [
      "0x1234567890abcdef...",
      "0x1234567890abcdef...",
      "Example App",
      [["usdc", "100.0"]],
      "3600",
      "app.create",
      "0xApp1234567890abcdef..."
    ], 1619123456789],
    "sig": ["0x5432abcdef..."]
  }
  ```

  </TabItem>
  <TabItem value="after" label="After">

  ```json
  {
    "req": [1, "auth_request", {
      "address": "0x1234567890abcdef...",
      "session_key": "0x9876543210fedcba...",
      "app_name": "Example App",
      "allowances": [
        {
          "asset": "usdc",
          "amount": "100.0"
        }
      ],
      "scope": "app.create",
      "expire": "3600",
      "application": "0xApp1234567890abcdef..."
    }, 1619123456789],
    "sig": ["0x5432abcdef..."]
  }
  ```

  </TabItem>
</Tabs>

### Nitrolite SDK

#### Modified Signature Type

The `Signature` interface has been replaced with a simple `Hex` type to support various signature standards including EIP-1271 and EIP-6492.

<Tabs>
  <TabItem value="before" label="Before">

  ```typescript
  interface Signature {
    v: number;
    r: Hex;
    s: Hex;
  }
  
  const sig: Signature = {
    v: 27,
    r: '0x...',
    s: '0x...'
  };
  ```

  </TabItem>
  <TabItem value="after" label="After">

  ```typescript
  type Signature = Hex;
  
  const sig: Signature = '0x...'; // Combined signature
  ```

  </TabItem>
</Tabs>

#### New State Signing Architecture

The SDK now uses a dedicated `StateSigner` interface for all state signing operations, replacing the optional `stateWalletClient` parameter.

[Read more](/nitrolite_client/advanced#state-signing)

<Tabs>
  <TabItem value="before" label="Before">

  ```typescript
  import { createNitroliteClient } from '@erc7824/nitrolite';
  
  const client = createNitroliteClient({
    publicClient,
    walletClient,
    stateWalletClient: sessionWalletClient, // [!code --]
    addresses,
  });
  ```

  </TabItem>
  <TabItem value="after" label="After">

  ```typescript
  import { 
    createNitroliteClient,
    WalletStateSigner
  } from '@erc7824/nitrolite';
  
  const client = createNitroliteClient({
    publicClient,
    walletClient,
    stateSigner: new WalletStateSigner(walletClient), // [!code ++]
    addresses,
  });
  ```

  </TabItem>
</Tabs>

#### Modified `CreateChannelParams` Interface

The `CreateChannelParams` interface has been updated to provide more explicit control over channel initialization due to changes in create/join semantics.

<Tabs>
  <TabItem value="before" label="Before">

  ```typescript
  const params: CreateChannelParams = {
    initialAllocationAmounts: [amount1, amount2],
    stateData: '0x...',
  };
  ```

  </TabItem>
  <TabItem value="after" label="After">

  ```typescript
  const params: CreateChannelParams = {
    channel: {
      participants: [address1, address2],
      weights: [50, 50],
      quorum: 100,
      adjudicator: adjudicatorAddress,
      challenge: 86400n,
      token: tokenAddress,
      nonce: 1n,
    },
    unsignedInitialState: {
      intent: StateIntent.Fund,
      version: 0n,
      data: '0x',
      allocations: [
        { destination: address1, token: tokenAddress, amount: amount1 },
        { destination: address2, token: tokenAddress, amount: amount2 },
      ],
    },
    serverSignature: '0x...',
  };
  ```

  </TabItem>
</Tabs>

#### RPC Request Structure

RPC method creation functions now use structured parameters instead of arrays for better type safety, matching the Clearnode API changes.

<Tabs>
  <TabItem value="before" label="Before">

  ```typescript
  const request = NitroliteRPC.createRequest(
    requestId,
    RPCMethod.GetChannels,
    [participant, status], // [!code --]
    timestamp
  );
  ```

  </TabItem>
  <TabItem value="after" label="After">

  ```typescript
  const request = NitroliteRPC.createRequest({
    method: RPCMethod.GetChannels,
    params: { // [!code ++]
      participant, // [!code ++]
      status, // [!code ++]
    }, // [!code ++]
    requestId,
    timestamp,
  });
  ```

  </TabItem>
</Tabs>
