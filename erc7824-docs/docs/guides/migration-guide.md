---
sidebar_position: 2
title: Migration Guide
description: Guide to migrate to newer versions of Nitrolite
keywords: [migration, upgrade, breaking changes, nitrolite, erc7824]
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Migration Guide

If you are coming from an earlier version of Nitrolite, you will need to account for the following breaking changes.

## 0.5.x Breaking changes

The 0.5.x release includes fundamental protocol changes affecting session keys, channel operations, state signatures, and channel resize rules. The main objective of these changes is to enhance security, and provide better experience for developers and users by ability to limit allowances for specific applications.

**Not ready to migrate?** Unfortunately, at this time Yellow Network does not provide ClearNodes running the previous version of the protocol, so you will need to migrate to the latest version to continue using the Network.

### Protocol Changes

These protocol-level changes affect all implementations and integrations with the Yellow Network.

#### Session Keys: Applications, Allowances, and Expiration

Session keys now have enhanced properties that define their access levels and capabilities:

- **Application field**: Determines the scope of session key permissions. Setting this to an application name (e.g., "My Trading App") grants application-scoped access with enforced allowances. Setting it to "clearnode" grants root access equivalent to the wallet itself.

- **Allowances field**: Defines spending limits for application-scoped session keys. These limits are tracked cumulatively across all operations and are enforced by the protocol.

- **Expires_at field**: Uses a bigint timestamp (seconds since epoch). Once expired, session keys are permanently frozen and cannot be reactivated. This is particularly critical for root access keys (application set to "clearnode") - if they expire, you lose the ability to perform channel operations.

#### Channel Creation: Separate Create and Fund Steps

Clearnode no longer supports creating channels with an initial deposit. All channels must be created with zero balance and funded separately through a resize operation. This two-step process ensures cleaner state management and prevents edge cases in channel initialization.

#### State Signatures: Wallet vs Session Key Signing

A fundamental change in how channel states are signed:

- **Channels created before v0.5.0**: The participant address is the session key, and all states must be signed by that session key.

- **Channels created after v0.5.0**: The participant address is the wallet address, and all states must be signed by the wallet.

This change improves security and aligns with standard practices, but requires careful handling during the transition period.

#### Resize Operations: Strict Channel Balance Rules

The protocol now enforces strict rules about channel balances and their impact on other operations:

- **Blocked operations**: Users with any channel containing non-zero amounts cannot perform transfers, submit app states with deposit intent, or create app sessions with non-zero allocations.

- **Resizing state**: After a resize request, channels enter a "resizing" state with locked funds until the on-chain transaction is confirmed. If a channel remains stuck in this state for an extended period, the recommended action is to close the channel and create a new one.

- **Allocate amount semantics**: The resize operation uses `allocate_amount` where negative values withdraw from the channel to unified balance, and positive values deposit to the channel.

:::warning
**Legacy channel migration**: Users with existing channels containing non-zero amounts must either resize them to zero (by providing "resize_amount" as 0 and "allocate_amount" as your **negative** on-chain balance) or close them to enable full protocol functionality. If you are unsure how to adjust resize parameters, the safe option is to close the old on-chain channel entirely, and open a new one.
:::

#### Non-Zero Channel Allocations: Operation Restrictions

The following operations will return errors if the user has any channel with non-zero amount:

- **Transfer**: Returns error code indicating blocked due to non-zero channel balance
- **Submit App State** (with deposit intent): Rejected if attempting to deposit
- **Create App Session** (with allocations): Rejected if attempting to allocate

The returned error has the following format: `operation denied: non-zero allocation in <count> channel(s) detected owned by wallet <address>"`

### Nitrolite SDK

You should definitely read this section if you are using the Nitrolite SDK.

#### Update Authentication

Implementing the new session key protocol changes:

<Tabs>
  <TabItem value="application" label="Application Session Key">

  ```typescript
  const authRequest = {
    address: '0x...',
    session_key: '0x...',
    application: 'My Trading App', // Application name for confined access
    allowances: [
      { asset: 'usdc', amount: '1000.0' },
      { asset: 'eth', amount: '0.5' }
    ],
    scope: 'app.create',
    expires_at: BigInt(Math.floor(Date.now() / 1000) + 7 * 24 * 60 * 60) // 7 days
  };
  ```

  </TabItem>
  <TabItem value="root" label="Root Access (Clearnode)">

  ```typescript
  const authRequest = {
    address: '0x...',
    session_key: '0x...',
    application: 'clearnode', // Special value for root access
    allowances: [], // Not enforced for root access
    scope: 'app.create',
    expires_at: BigInt(Math.floor(Date.now() / 1000) + 365 * 24 * 60 * 60) // Long expiration recommended
  };
  ```

  </TabItem>
</Tabs>

**Important considerations:**
- Root access keys (application: "clearnode") cannot perform channel operations after expiration
- Plan expiration times based on your operational needs
- Application-scoped keys track cumulative spending against allowances

#### Migrate Channel Creation

Channels must now be created with zero initial deposit and funded separately via the `resizeChannel` method:

```typescript
const { channelId } = await client.createChannel({
  chain_id: 1,
  token: tokenAddress,
  // remove-next-line
  amount: BigInt(1000000), // Initial deposit
  // remove-next-line
  session_key: '0x...' // Optional
});

// add-start
// Step 2: Fund the channel separately
await client.resizeChannel({
  channel_id: channelId,
  amount: BigInt(1000000),
});
// add-end
```

#### Resize correctly

Channel resizing must be negotiated with the ClearNode through WebSocket. Use `resize_amount` and `allocate_amount` with correct sign convention (`resize_amount = -allocate_amount`) and help users with non-zero channel balances migrate by resizing to zero or reopening channels.

Channel resize can be requested as follows:

```typescript
const resizeMessage = await createResizeChannelMessage(messageSigner, {
  channel_id: channelId,
  resize_amount: BigInt(50), // Positive = deposit to channel, negative = withdraw from channel to custody ledger
  allocate_amount: BigInt(-50), // Negative = deposit to unified balance, negative = withdraw from unified balance to channel
  funds_destination: walletAddress,
});

const resizeResponse = {}; // send the message and wait for Clearnode's response

const { params: resizeResponseParams } = parseResizeChannelResponse(resizeResponse);
const resizeParams = {
  resizeState: {
      channelId,
      ...resizeResponseParams.state,
      serverSignature: resizeResponseParams.serverSignature,
      data: resizeResponseParams.state.stateData as Hex,
      version: BigInt(resizeResponseParams.state.version),
  },
  // `previousState` is either initial or previous resizing state, depending on which has higher version number
  // can be obtained with `await (client.getChannelData(channelId)).lastValidState`
  proofStates: [previousState],
}

const {txHash} = await client.resizeChannel(resizeParams);
```

Here is how you can migrate your channels:

```typescript
// Check and migrate channels with non-zero amounts
const channels = await client.getOpenChannels();

for (const channel of channels) {
  if (channel.amount > 0) {
    // Must empty channel to enable transfers/app operations
    const resizeMessage = await createResizeChannelMessage(messageSigner, {
      channel_id: channel.channelId,
      resize_amount: BigInt(0),
      allocate_amount: -BigInt(channel.amount),
      funds_destination: walletAddress,
    });
    
    // perform the resize as shown above
  }
}
```


**Critical:** Operations blocked when any channel has non-zero amount:
- Off-chain transfers
- App state submissions with deposit intent
- Creating app sessions with allocations

#### Test State Signatures

If you plan to work with on-chain channels opened PRIOR to v0.5.0, then on NitroliteClient initialization the `stateSigner` you specify must be based on a Session Key used in the channel as participant. Even if this session key is or will expire, you still need to provide a `stateSigner` based on it.

On the other hand, if you plan to work with channels created SINCE v0.5.0, you can specify the `stateSigner` based on the `walletClient` you have specified.

#### Manage Session Keys

New methods have been added for comprehensive session key management, including retrieval and revocation.

```typescript
// Get all active session keys
const sessionKeys = await client.getSessionKeys();

// Revoke a specific session key
await client.revokeSessionKey({
  session_key: '0x...'
});

// Session key data structure
interface RPCSessionKey {
  id: string;
  sessionKey: Address;
  application: string;
  allowances: RPCAllowanceUsage[]; // Includes usage tracking
  scope: string;
  expiresAt: bigint;
  createdAt: bigint;
}
```

#### EIP-712 Signatures: String-based Amounts

EIP-712 signature types now use string values for amounts instead of numeric types to support better precision with decimal values.

```typescript
const types = {
  Allowance: [
    { name: 'asset', type: 'string' },
    // remove-next-line
    { name: 'amount', type: 'uint256' },
    // add-next-line
    { name: 'amount', type: 'string' },
  ]
};
```

### ClearNode API

You should read this section only if you are using the ClearNode API directly.

#### Update Authentication

Use the new session key parameters with proper `application`, `allowances`, and `expires_at` fields:

<Tabs>
  <TabItem value="application" label="Application Auth">

  ```json
  {
    "req": [1, "auth_request", {
      "address": "0x1234567890abcdef...",
      "session_key": "0x9876543210fedcba...",
      "application": "My Trading App",
      "allowances": [
        { "asset": "usdc", "amount": "1000.0" },
        { "asset": "eth", "amount": "0.5" }
      ],
      "scope": "app.create",
      "expires_at": 1719123456789
    }, 1619123456789],
    "sig": ["0x..."]
  }
  ```

  </TabItem>
  <TabItem value="root" label="Root Auth (Clearnode)">

  ```json
  {
    "req": [1, "auth_request", {
      "address": "0x1234567890abcdef...",
      "session_key": "0x9876543210fedcba...",
      "application": "clearnode",
      "allowances": [],
      "scope": "app.create",
      "expires_at": 1750659456789
    }, 1619123456789],
    "sig": ["0x..."]
  }
  ```

  </TabItem>
</Tabs>

#### Migrate Channel Creation

Implement the two-step process (create empty, then resize to fund)

The `create_channel` method no longer accepts `amount` and `session_key` parameters:

```json
{
  "req": [1, "create_channel", {
    "chain_id": 137,
    "token": "0xeeee567890abcdef...",
    // remove-next-line
    "amount": "100000000",
    // remove-next-line
    "session_key": "0x1234567890abcdef..."
  }, 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

#### Manage Session Keys

New methods for session key operations have been added.

##### Get Session Keys

Request:
```json
{
  "req": [1, "get_session_keys", {}, 1619123456789],
  "sig": ["0x..."]
}
```

Response:
```json
{
  "res": [1, "get_session_keys", {
    "session_keys": [{
      "id": "sk_123",
      "session_key": "0x9876543210fedcba...",
      "application": "My Trading App",
      "allowances": [
        { "asset": "usdc", "amount": "1000.0", "used": "250.0" }
      ],
      "scope": "app.create",
      "expires_at": 1719123456789,
      "created_at": 1619123456789
    }]
  }, 1619123456789],
  "sig": ["0x..."]
}
```

##### Revoke Session Key Request

Request:
```json
{
  "req": [1, "revoke_session_key", {
    "session_key": "0x1234567890abcdef..."
  }, 1619123456789],
  "sig": ["0x..."]
}
```

Response:
```json
{
  "res": [1, "revoke_session_key", {
    "session_key": "0x1234567890abcdef..."
  }, 1619123456789],
  "sig": ["0x..."]
}
```

## 0.3.x Breaking changes

The 0.3.x release includes breaking changes to the SDK architecture, smart contract interfaces, and Clearnode API enhancements listed below.

**Not ready to migrate?** Unfortunately, at this time Yellow Network does not provide ClearNodes running the previous version of the protocol, so you will need to migrate to the latest version to continue using the Network.

### Nitrolite SDK

You should definitely read this section if you are using the Nitrolite SDK.

#### Client: Replaced `stateWalletClient` with `StateSigner`

The `stateWalletClient` parameter of `NitroliteClient` has been replaced with a required `stateSigner` parameter that implements the `StateSigner` interface.

When initializing the client, you should use either `WalletStateSigner` or `SessionKeyStateSigner` to handle state signing.

```typescript
// remove-next-line
import { createNitroliteClient } from '@erc7824/nitrolite';
// add-start
import { 
  createNitroliteClient,
  WalletStateSigner
} from '@erc7824/nitrolite';
// add-end

const client = createNitroliteClient({
  publicClient,
  walletClient,
  // remove-next-line
  stateWalletClient: sessionWalletClient,
  // add-next-line
  stateSigner: new WalletStateSigner(walletClient),
  addresses,
});
```

**For session key signing:**

```typescript
import { SessionKeyStateSigner } from '@erc7824/nitrolite';

const stateSigner = new SessionKeyStateSigner('0x...' as Hex);
```

#### Actions: Modified `createChannel` Parameters

The `CreateChannelParams` interface has been fully restructured for better clarity.

You should use the new [`CreateChannel` ClearNode API endpoint](#added-create_channel-method) to get the response, that fully resembles the channel creation parameters.

```typescript
// remove-start
const { channelId, initialState, txHash } = await client.createChannel(
  tokenAddress,
  {
    initialAllocationAmounts: [amount1, amount2],
    stateData: '0x...',
  }
);
// remove-end
// add-start
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
// add-end
```

#### Actions: Structured Typed RPC Request Parameters

RPC requests now use endpoint-specific object-based parameters instead of untyped arrays for improved type safety.

You should update your RPC request creation code to use the new structured format and RPC types.

```typescript
// remove-start
const request = NitroliteRPC.createRequest(
  requestId,
  RPCMethod.GetChannels,
  [participant, status],
  timestamp
);
// remove-end
// add-start
const request = NitroliteRPC.createRequest({
  method: RPCMethod.GetChannels,
  params: {
    participant,
    status,
  },
  requestId,
  timestamp,
});
// add-end
```

#### Actions: Standardized Channel Operations Responses

The responses for `CloseChannel` and `ResizeChannel` methods have been aligned with newly added `CreateChannel` endpoint for consistency.

Update your response handling code to use the new `RPCChannelOperation` type.

```typescript
// remove-start
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
// remove-end
// add-start
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
// add-end
```

#### Actions: Modified `Signature` Type

The `Signature` struct has been replaced with a simple `Hex` type to support EIP-1271 and EIP-6492 signatures.

Update your signature-handling code to use the new `Hex` type. Still, if using Nitrolite utils correctly, you will not need to change anything, as the utils will handle the conversion for you.

```typescript
// remove-start
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
// remove-end
// add-start
type Signature = Hex;

const sig: Signature = '0x...';
// add-end
```

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

```json
{
  // remove-next-line
  "req": [1, "auth_request", [{
  // add-next-line
  "req": [1, "auth_request", {
    "address": "0x1234567890abcdef...",
    "session_key": "0x9876543210fedcba...",
    "app_name": "Example App",
    // remove-next-line
    "allowances": [ "usdc", "100.0" ],
    // add-start
    "allowances": [
      {
        "asset": "usdc",
        "amount": "100.0"
      }
    ],
    // add-end
    "scope": "app.create",
    "expire": "3600",
    "application": "0xApp1234567890abcdef..."
  // remove-next-line
  }], 1619123456789],
  // add-next-line
  }, 1619123456789],
  "sig": ["0x5432abcdef..."]
}
```

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

```json
// remove-start
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
// remove-end
// add-start
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
// add-end
```

#### Added: Pagination Metadata

Pagination-supporting endpoints now include a `metadata` struct in their responses with pagination information.

Update your response handling for `get_channels`, `get_app_sessions`, `get_ledger_entries`, and `get_ledger_transactions` to use the new metadata structure.

```json
// remove-start
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
// remove-end
// add-start
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
// add-end
```

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

```solidity
// remove-start
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
// remove-end
// add-start
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
// add-end
```

#### Actions: Updated `State` Signature Array

The `State` struct now uses `bytes[]` for signatures instead of `Signature[]`.

```solidity
struct State {
  uint8 intent;
  uint256 version;
  bytes data;
  Allocation[] allocations;
  // remove-next-line
  Signature[] sigs;
  // add-next-line
  bytes[] sigs;
}
```

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

```solidity
// remove-start
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
// remove-end
// add-start
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
// add-end
```

#### Added: Enhanced Signature Support

Smart contracts now support EIP-191, EIP-712, EIP-1271, and EIP-6492 signature formats for greater compatibility.

The contracts automatically detect and verify the appropriate signature format:
- **Raw ECDSA**: Traditional `(r, s, v)` signatures
- **EIP-191**: Personal message signatures (`\x19Ethereum Signed Message:\n`)  
- **EIP-712**: Typed structured data signatures
- **EIP-1271**: Smart contract wallet signatures
- **EIP-6492**: Signatures for undeployed contracts

No changes are needed in your contract calls - the signature verification is handled automatically by the contract.
