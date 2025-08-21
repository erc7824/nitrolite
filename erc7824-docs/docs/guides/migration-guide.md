---
sidebar_position: 2
title: Migration Guide
description: Guide to migrate to newer versions of Nitrolite
keywords: [migration, upgrade, breaking changes, nitrolite, erc7824]
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## 0.3.0 Changelog

This release focuses on streamlining channel creation, improving API consistency, and adding support for modern signature standards.

### Nitrolite SDK

- Replaced an optional NitroliteClient `stateWalletClient` field with a required `StateSigner` interface to handle state signing operations (via `signState(channelId, state)`).
- Added `WalletStateSigner` (using viem's `WalletClient`) and `SessionStateSigner` (using raw private key account) classes to provide convenient implementations of the `StateSigner` interface.
- Updated `CreateChannel(...)` method parameters to support the improved channel creation flow.
- Standardized the request types in all Clearnode API methods, so that now `params` is an object instead of an array. This improves type safety and clarity.
- Standardized `CreateChannel`, `CloseChannel`, and `ResizeChannel` method response types.
- Changed the `Signature` type from `{r, s, v}` struct to `Hex`.
- Added pagination types and parameters for requests to ClearNode pagination-supporting endpoints.

### Clearnode API

- Added `create_channel` method to facilitate the improved, single-transaction channel opening flow. The response includes the `Channel` and `State` structs, alongside with the `serverSignature` for the initial state.
- Standardized the request structure in all Clearnode API methods, so that now `parameters` is an object instead of an array. This improves type safety and clarity.
- Standardized `create_channel`, `close_channel`, and `resize_channel` method responses.
- Added `metadata` struct to the `get_channels`, `get_app_sessions`, `get_ledger_entries`, and `get_ledger_transactions` method responses to provide additional pagination information: current page, number of items per page, total pages, and total items.

The v0.3.0 API documentation can be found in the [API docs of the ClearNode](https://github.com/erc7824/nitrolite/blob/fcbf8a737ea5643234e8d01f14ee9344b3d0198b/clearnode/docs/API.md).

### Smart Contract

- Channels can now become operational immediately after the `create(...)` method call if all participant signatures are provided. In such case, for participants other than sender ("creator") the funds are transferred from the account specified in `participants` array.
- Changed the `Signature` type from `{r, s, v}` struct to `bytes`.
- Added `EIP712AdjudicatorBase` to support EIP-712 typed structured data signatures in Adjudicator contracts.
- Added support for [EIP-191](https://eip.tools/eip/191), [EIP-712](https://eip.tools/eip/712), [EIP-1271](https://eip.tools/eip/1271), and [EIP-6492](https://eip.tools/eip/6492) signatures.

# Migration Guide

If you are coming from an earlier version of Nitrolite, you will need to account for the following breaking changes.

## 0.3.x Breaking changes

The 0.3.x release includes breaking changes to the SDK architecture, smart contract interfaces, and Clearnode API enhancements listed below.

**Not ready to migrate?** Unfortunately, at this time Yellow Network does not provide ClearNodes running the previous version of the protocol, so you will need to migrate to the latest version to continue using the Network.

### Nitrolite SDK

You should definitely read this section if you are using the Nitrolite SDK.

#### Client: Replaced `stateWalletClient` with `StateSigner`

The `stateWalletClient` parameter of `NitroliteClient` has been replaced with a required `stateSigner` parameter that implements the `StateSigner` interface.

When initializing the client, you should use either `WalletStateSigner` or `SessionKeyStateSigner` to handle state signing.

<Tabs>
  <TabItem value="before" label="Before">

  ```typescript
  import { createNitroliteClient } from '@erc7824/nitrolite';
  
  const client = createNitroliteClient({
    publicClient,
    walletClient,
    stateWalletClient: sessionWalletClient,
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
    stateSigner: new WalletStateSigner(walletClient),
    addresses,
  });
  ```

  </TabItem>
</Tabs>

**For session key signing:**

```typescript
import { SessionKeyStateSigner } from '@erc7824/nitrolite';

const stateSigner = new SessionKeyStateSigner('0x...' as Hex);
```

#### Actions: Modified `createChannel` Parameters

The `CreateChannelParams` interface has been fully restructured for better clarity.

You should use the new [`CreateChannel` ClearNode API endpoint](#added-create_channel-method) to get the response, that fully resembles the channel creation parameters.

<Tabs>
  <TabItem value="before" label="Before">

  ```typescript
  const { channelId, initialState, txHash } = await client.createChannel(
    tokenAddress,
    {
      initialAllocationAmounts: [amount1, amount2],
      stateData: '0x...',
    }
  );
  ```

  </TabItem>
  <TabItem value="after" label="After">

  ```typescript
  const { channelId, initialState, txHash } = await client.createChannel({
    channel: {
      participants: [address1, address2],
      adjudicator: adjudicatorAddress,
      challenge: 86400n,
      nonce: 42n,
    },
    unsignedInitialState: {
      intent: StateIntent.Initialize,
      version: 0n,
      data: '0x',
      allocations: [
        { destination: address1, token: tokenAddress, amount: amount1 },
        { destination: address2, token: tokenAddress, amount: amount2 },
      ],
    },
    serverSignature: '0x...',
  });
  ```

  </TabItem>
</Tabs>

#### Actions: Structured Typed RPC Request Parameters

RPC requests now use endpoint-specific object-based parameters instead of untyped arrays for improved type safety.

You should update your RPC request creation code to use the new structured format and RPC types.

<Tabs>
  <TabItem value="before" label="Before">

  ```typescript
  const request = NitroliteRPC.createRequest(
    requestId,
    RPCMethod.GetChannels,
    [participant, status],
    timestamp
  );
  ```

  </TabItem>
  <TabItem value="after" label="After">

  ```typescript
  const request = NitroliteRPC.createRequest({
    method: RPCMethod.GetChannels,
    params: {
      participant,
      status,
    },
    requestId,
    timestamp,
  });
  ```

  </TabItem>
</Tabs>

#### Actions: Standardized Channel Operations Responses

The responses for `CloseChannel` and `ResizeChannel` methods have been aligned with newly added `CreateChannel` endpoint for consistency.

Update your response handling code to use the new `RPCChannelOperation` type.

<Tabs>
  <TabItem value="before" label="Before">

  ```typescript
    export interface ResizeChannelResponseParams {
      channelId: Hex;
      stateData: Hex;
      intent: number;
      version: number;
      allocations: RPCAllocation[];
      stateHash: Hex;
      serverSignature: ServerSignature;
  }

  export interface CloseChannelResponseParams {
      channelId: Hex;
      intent: number;
      version: number;
      stateData: Hex;
      allocations: RPCAllocation[];
      stateHash: Hex;
      serverSignature: ServerSignature;
  }
  ```

  </TabItem>
  <TabItem value="after" label="After">

  ```typescript
  export interface RPCChannelOperation {
    channelId: Hex;
    state: RPCChannelOperationState;
    serverSignature: Hex;
  }

  export interface CreateChannelResponse extends GenericRPCMessage {
    method: RPCMethod.CreateChannel;
    params: RPCChannelOperation & {
        channel: RPCChannel;
    };
  }

  export interface ResizeChannelResponse extends GenericRPCMessage {
      method: RPCMethod.ResizeChannel;
      params: RPCChannelOperation;
  }

  export interface CloseChannelResponse extends GenericRPCMessage {
      method: RPCMethod.CloseChannel;
      params: RPCChannelOperation;
  }
  ```

  </TabItem>
</Tabs>

#### Actions: Modified `Signature` Type

The `Signature` struct has been replaced with a simple `Hex` type to support EIP-1271 and EIP-6492 signatures.

Update your signature-handling code to use the new `Hex` type. Still, if using Nitrolite utils correctly, you will not need to change anything, as the utils will handle the conversion for you.

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

#### Added: Pagination Types and Parameters

To support pagination in ClearNode API requests, new types and parameters have been added.

For now, only `GetLedgerTransactions` request has been updated to include pagination.

```typescript
export interface PaginationFilters {
    /** Pagination offset. */
    offset?: number;
    /** Number of transactions to return. */
    limit?: number;
    /** Sort order by created_at. */
    sort?: 'asc' | 'desc';
}
```

### Clearnode API

You should read this section only if you are using the ClearNode API directly, or if you are using the Nitrolite SDK with custom ClearNode API requests.

#### Actions: Structured Request Parameters

ClearNode API requests have migrated from array-based parameters to structured object parameters for improved type safety and API clarity.

Update all your ClearNode API requests to use object-based parameters instead of arrays.

<Tabs>
  <TabItem value="before" label="Before">

  ```json
  {
    "req": [1, "auth_request", [{
      "address": "0x1234567890abcdef...",
      "session_key": "0x9876543210fedcba...",
      "app_name": "Example App",
      "allowances": [ "usdc", "100.0" ],
      "scope": "app.create",
      "expire": "3600",
      "application": "0xApp1234567890abcdef..."
    }], 1619123456789],
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

#### Added: `create_channel` Method

A new `create_channel` method has been added to facilitate the improved single-transaction channel opening flow.

Use this method to request channel creation parameters from the broker, then submit the returned data to the smart contract via Nitrolite SDK or directly.

**Request:**
```json
{
  "req": [1, "create_channel", {
    "chain_id": 137,
    "token": "0xeeee567890abcdef...",
    "amount": "100000000",
    "session_key": "0x1234567890abcdef..." // Optional
  }, 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

**Response:**
```json
{
  "res": [1, "create_channel", {
    "channel_id": "0x4567890123abcdef...",
    "channel": {
      "participants": ["0x1234567890abcdef...", "0xbbbb567890abcdef..."],
      "adjudicator": "0xAdjudicatorContractAddress...",
      "challenge": 3600,
      "nonce": 1619123456789
    },
    "state": {
      "intent": 1,
      "version": 0,
      "state_data": "0xc0ffee",
      "allocations": [
        {
          "destination": "0x1234567890abcdef...",
          "token": "0xeeee567890abcdef...",
          "amount": "100000000"
        },
        {
          "destination": "0xbbbb567890abcdef...",
          "token": "0xeeee567890abcdef...",
          "amount": "0"
        }
      ]
    },
    "server_signature": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1c"
  }, 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

#### API: Standardized Channel Operation Responses

The responses for `create_channel`, `close_channel`, and `resize_channel` methods have been unified for consistency.

Update your response parsing to handle the new unified structure with `channel_id`, `state`, and `server_signature` fields.

<Tabs>
  <TabItem value="before" label="Before">

  ```json
  {
    "res": [1, "close_channel", {
      "channelId": "0x4567890123abcdef...",
      "intent": 3,
      "version": 123,
      "stateData": "0x0000000000000000000000000000000000000000000000000000000000001ec7",
      "allocations": [...],
      "stateHash": "0x...",
      "serverSignature": "0x..."
    }, 1619123456789],
    "sig": ["0xabcd1234..."]
  }
  ```

  </TabItem>
  <TabItem value="after" label="After">

  ```json
  {
    "res": [1, "close_channel", {
      "channel_id": "0x4567890123abcdef...",
      "state": {
        "intent": 3,
        "version": 123,
        "state_data": "0xc0ffee",
        "allocations": [
          {
            "destination": "0x1234567890abcdef...",
            "token": "0xeeee567890abcdef...",
            "amount": "50000"
          }
        ]
      },
      "server_signature": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1c"
    }, 1619123456789],
    "sig": ["0xabcd1234..."]
  }
  ```

  </TabItem>
</Tabs>

#### Added: Pagination Metadata

Pagination-supporting endpoints now include a `metadata` struct in their responses with pagination information.

Update your response handling for `get_channels`, `get_app_sessions`, `get_ledger_entries`, and `get_ledger_transactions` to use the new metadata structure.

<Tabs>
  <TabItem value="before" label="Before">

  ```json
  {
    "res": [1, "get_channels", [
      [
        {
          "channel_id": "0xfedcba9876543210...",
          "status": "open",
          // ... channel data
        }
      ]
    ], 1619123456789],
    "sig": ["0xabcd1234..."]
  }
  ```

  </TabItem>
  <TabItem value="after" label="After">

  ```json
  {
    "res": [1, "get_channels", {
      "channels": [
        {
          "channel_id": "0xfedcba9876543210...",
          "status": "open",
          // ... channel data
        }
      ],
      "metadata": {
        "page": 1,
        "per_page": 10,
        "total_count": 56,
        "page_count": 6
      }
    }, 1619123456789],
    "sig": ["0xabcd1234..."]
  }
  ```

  </TabItem>
</Tabs>

The metadata fields provide:
- `page`: Current page number
- `per_page`: Number of items per page  
- `total_count`: Total number of items available
- `page_count`: Total number of pages

### Contracts

You should read this section only if you are using the Nitrolite smart contracts directly.

#### Action: Replaced `Signature` Struct with `bytes`

The `Signature` struct has been removed and replaced with `bytes` type to support EIP-1271, EIP-6492, and other signature formats.

Update all contract interactions that use signatures to pass `bytes` instead of the struct.

<Tabs>
  <TabItem value="before" label="Before">

  ```solidity
  struct Signature {
    uint8 v;
    bytes32 r;
    bytes32 s;
  }
  
  function join(
    bytes32 channelId,
    uint256 index,
    Signature calldata sig
  ) external returns (bytes32);
  
  function challenge(
    bytes32 channelId,
    State calldata candidate,
    State[] calldata proofs,
    Signature calldata challengerSig
  ) external;
  ```

  </TabItem>
  <TabItem value="after" label="After">

  ```solidity
  // Signature struct is removed
  
  function join(
    bytes32 channelId,
    uint256 index,
    bytes calldata sig
  ) external returns (bytes32);
  
  function challenge(
    bytes32 channelId,
    State calldata candidate,
    State[] calldata proofs,
    bytes calldata challengerSig
  ) external;
  ```

  </TabItem>
</Tabs>

#### Actions: Updated `State` Signature Array

The `State` struct now uses `bytes[]` for signatures instead of `Signature[]`.

<Tabs>
  <TabItem value="before" label="Before">

  ```solidity
  struct State {
    uint8 intent;
    uint256 version;
    bytes data;
    Allocation[] allocations;
    Signature[] sigs;
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
    bytes[] sigs;
  }
  ```

  </TabItem>
</Tabs>

#### Added: Auto-Join Channel Creation Flow

Channels can now become operational immediately after the `create()` call if all participant signatures are provided.

When calling `create()` with complete signatures from all participants, the channel automatically becomes active without requiring a separate `join()` call.

**Single signature (requires join):**
```solidity
// Create channel with only creator's signature
State memory initialState = State({
    intent: StateIntent.Fund,
    version: 0,
    data: "0x",
    allocations: allocations,
    sigs: [creatorSignature] // Only one signature
});

bytes32 channelId = custody.create(channel, initialState);
// Channel status: JOINING - requires server to call join()
```

**Complete signatures (auto-active):**
```solidity
// Create channel with all participants' signatures
State memory initialState = State({
    intent: StateIntent.Fund,
    version: 0,
    data: "0x",
    allocations: allocations,
    sigs: [creatorSignature, serverSignature] // All signatures
});

bytes32 channelId = custody.create(channel, initialState);
// Channel status: ACTIVE - ready for use immediately
```

#### Actions: Update Adjudicator Contracts for EIP-712 Support

A new `EIP712AdjudicatorBase` base contract has been added to support EIP-712 typed structured data signatures in adjudicator implementations.

The `EIP712AdjudicatorBase` provides:
- **Domain separator retrieval**: Gets EIP-712 domain separator from the channel implementation contract
- **ERC-5267 compliance**: Automatically handles EIP-712 domain data retrieval
- **Ownership management**: Built-in access control for updating channel implementation address
- **Graceful fallbacks**: Returns `NO_EIP712_SUPPORT` constant when EIP-712 is not available

If you have custom adjudicator contracts, inherit from `EIP712AdjudicatorBase` to enable EIP-712 signature verification.

<Tabs>
  <TabItem value="before" label="Before">

  ```solidity
  import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
  import {Channel, State, Allocation, StateIntent} from "../interfaces/Types.sol";

  contract MyAdjudicator is IAdjudicator {
      function adjudicate(
          Channel calldata chan, 
          State calldata candidate, 
          State[] calldata proofs
      ) external view override returns (bool valid) {
          return candidate.validateUnanimousSignatures(chan);
      }
  }
  ```

  </TabItem>
  <TabItem value="after" label="After">

  ```solidity
  import {IAdjudicator} from "../interfaces/IAdjudicator.sol";
  import {Channel, State, Allocation, StateIntent} from "../interfaces/Types.sol";
  import {EIP712AdjudicatorBase} from "./EIP712AdjudicatorBase.sol";

  contract MyAdjudicator is IAdjudicator, EIP712AdjudicatorBase {
      constructor(address owner, address channelImpl) 
          EIP712AdjudicatorBase(owner, channelImpl) {}

      function adjudicate(
          Channel calldata chan, 
          State calldata candidate, 
          State[] calldata proofs
      ) external override returns (bool valid) {
          bytes32 domainSeparator = getChannelImplDomainSeparator();
          return candidate.validateUnanimousStateSignatures(chan, domainSeparator);
      }
  }
  ```

  </TabItem>
</Tabs>

#### Added: Enhanced Signature Support

Smart contracts now support EIP-191, EIP-712, EIP-1271, and EIP-6492 signature formats for greater compatibility.

The contracts automatically detect and verify the appropriate signature format:
- **Raw ECDSA**: Traditional `(r, s, v)` signatures
- **EIP-191**: Personal message signatures (`\x19Ethereum Signed Message:\n`)  
- **EIP-712**: Typed structured data signatures
- **EIP-1271**: Smart contract wallet signatures
- **EIP-6492**: Signatures for undeployed contracts

No changes are needed in your contract calls - the signature verification is handled automatically by the contract.
