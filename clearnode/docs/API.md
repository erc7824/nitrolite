# Clearnode API Reference

## API Endpoints

| Method | Description |
|--------|-------------|
| `auth_request` | Initiates authentication with the server |
| `auth_challenge` | Server response with authentication challenge |
| `auth_verify` | Completes authentication with a challenge response |
| `ping` | Simple connectivity check |
| `get_config` | Retrieves broker configuration and supported networks |
| `get_assets` | Retrieves all supported assets (optionally filtered by chain_id) |
| `get_app_definition` | Retrieves application definition for a ledger account |
| `get_app_sessions` | Lists virtual applications for a participant with optional status filter |
| `get_ledger_balances` | Lists participants and their balances for a ledger account |
| `get_ledger_entries` | Retrieves detailed ledger entries for a participant |
| `get_channels` | Lists all channels for a participant with their status across all chains |
| `get_rpc_history` | Retrieves all RPC message history for a participant |
| `create_app_session` | Creates a new virtual application on a ledger |
| `close_app_session` | Closes a virtual application |
| `close_channel` | Closes a payment channel |
| `resize_channel` | Adjusts channel capacity |

## Authentication

### Authentication Request

Initiates authentication with the server.

**Request:**

```json
{
  "req": [1, "auth_request", ["0x1234567890abcdef..."], 1619123456789],
  "sig": ["0x5432abcdef..."] // Client's signature of the entire 'req' object
}
```

### Authentication Challenge

Server response with a challenge token for the client to sign.

**Response:**

```json
{
  "res": [1, "auth_challenge", [{
    "challenge_message": "550e8400-e29b-41d4-a716-446655440000"
  }], 1619123456789],
  "sig": ["0x9876fedcba..."] // Server's signature of the entire 'res' object
}
```

### Authentication Verification

Completes authentication with a challenge response.

**Request:**

```json
{
  "req": [2, "auth_verify", [{
    "address": "0x1234567890abcdef...",
    "challenge": "550e8400-e29b-41d4-a716-446655440000"
  }], 1619123456789],
  "sig": ["0x2345bcdef..."] // Client's signature of the entire 'req' object
}
```

**Response:**

```json
{
  "res": [2, "auth_verify", [{
    "address": "0x1234567890abcdef...",
    "success": true
  }], 1619123456789],
  "sig": ["0xabcd1234..."] // Server's signature of the entire 'res' object
}
```

## Ledger Management

### Get App Definition

Retrieves the application definition for a specific ledger account.

**Request:**

```json
{
  "req": [1, "get_app_definition", [{
    "app_session_id": "0x1234567890abcdef..."
  }], 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

**Response:**

```json
{
  "res": [1, "get_app_definition", [
    {
      "protocol": "NitroRPC/0.2",
      "participants": [
        "0xAaBbCcDdEeFf0011223344556677889900aAbBcC",
        "0x00112233445566778899AaBbCcDdEeFf00112233"
      ],
      "weights": [50, 50],
      "quorum": 100,
      "challenge": 86400,
      "nonce": 1
    }
  ], 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

### Get App Sessions

Lists all virtual applications for a participant. Optionally, you can filter the results by status (open, closed).

**Request:**

```json
{
  "req": [1, "get_app_sessions", [{
    "participant": "0x1234567890abcdef...",
    "status": "open"  // Optional: filter by status
  }], 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

**Response:**

```json
{
  "res": [1, "get_app_sessions", [[
    {
      "app_session_id": "0x3456789012abcdef...",
      "status": "open",
      "participants": [
        "0x1234567890abcdef...",
        "0x00112233445566778899AaBbCcDdEeFf00112233"
      ],
      "protocol": "NitroAura",
      "challenge": 86400,
      "weights": [50, 50],
      "quorum": 100,
      "version": 1,
      "nonce": 123456789
    },
    {
      "app_session_id": "0x7890123456abcdef...",
      "status": "open",
      "participants": [
        "0x1234567890abcdef...",
        "0xAaBbCcDdEeFf0011223344556677889900aAbBcC"
      ],
      "protocol": "NitroSnake",
      "challenge": 86400,
      "weights": [70, 30],
      "quorum": 100,
      "version": 1,
      "nonce": 123456790
    }
  ]], 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

### Get Ledger Balances

Retrieves the balances of all participants in a specific ledger account.

**Request:**

```json
{
  "req": [1, "get_ledger_balances", [{
    "participant": "0x1234567890abcdef..."
  }], 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

**Response:**

```json
{
  "res": [1, "get_ledger_balances", [[
    {
      "asset": "usdc",
      "amount": "100.0"
    },
    {
      "asset": "eth",
      "amount": "0.5"
    }
  ]], 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

### Get Ledger Entries

Retrieves the detailed ledger entries for an account, providing a complete transaction history. This can be used to audit all deposits, withdrawals, and transfers.

**Request:**

```json
{
  "req": [1, "get_ledger_entries", [{
    "account_id": "0x1234567890abcdef...",
    "asset": "usdc"  // Optional: filter by asset
  }], 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

**Response:**

```json
{
  "res": [1, "get_ledger_entries", [[
    {
      "id": 123,
      "account_id": "0x1234567890abcdef...",
      "account_type": 0,
      "asset": "usdc",
      "participant": "0x1234567890abcdef...",
      "credit": "100.0",
      "debit": "0.0",
      "created_at": "2023-05-01T12:00:00Z"
    },
    {
      "id": 124,
      "account_id": "0x1234567890abcdef...",
      "account_type": 0,
      "asset": "usdc",
      "participant": "0x1234567890abcdef...",
      "credit": "0.0",
      "debit": "25.0",
      "created_at": "2023-05-01T14:30:00Z"
    }
  ]], 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

### Get Channels

Retrieves all channels for a participant (both open, closed, and joining), ordered by creation date (newest first). This method returns channels across all supported chains.

**Request:**

```json
{
  "req": [1, "get_channels", [{
    "participant": "0x1234567890abcdef..."
  }], 1619123456789],
  "sig": []
}
```

**Response:**

```json
{
  "res": [1, "get_channels", [[
    {
      "channel_id": "0xfedcba9876543210...",
      "participant": "0x1234567890abcdef...",
      "status": "open",
      "token": "0xeeee567890abcdef...",
      "amount": "100000",
      "chain_id": 137,
      "adjudicator": "0xAdjudicatorContractAddress...",
      "challenge": 86400,
      "nonce": 1,
      "version": 2,
      "created_at": "2023-05-01T12:00:00Z",
      "updated_at": "2023-05-01T12:30:00Z"
    },
    {
      "channel_id": "0xabcdef1234567890...",
      "participant": "0x1234567890abcdef...",
      "status": "closed",
      "token": "0xeeee567890abcdef...",
      "amount": "50000",
      "chain_id": 42220,
      "adjudicator": "0xAdjudicatorContractAddress...",
      "challenge": 86400,
      "nonce": 1,
      "version": 3,
      "created_at": "2023-04-15T10:00:00Z",
      "updated_at": "2023-04-20T14:30:00Z"
    }
  ]], 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

The signature in the request must be from the participant's private key, verifying they own the address. This prevents unauthorized access to channel information.

Each channel response includes:
- `channel_id`: Unique identifier for the channel
- `participant`: The participant's address
- `status`: Current status ("open", "closed", or "joining")
- `token`: The token address for the channel
- `amount`: Total channel capacity
- `chain_id`: The blockchain network ID where the channel exists (e.g., 137 for Polygon, 42220 for Celo, 8453 for Base)
- `adjudicator`: The address of the adjudicator contract
- `challenge`: Challenge period duration in seconds
- `nonce`: Current nonce value for the channel
- `version`: Current version of the channel state
- `created_at`: When the channel was created (ISO 8601 format)
- `updated_at`: When the channel was last updated (ISO 8601 format)

### Get RPC History

Retrieves all RPC messages history for a participant, ordered by timestamp (newest first).

**Request:**

```json
{
  "req": [4, "get_rpc_history", [], 1619123456789],
  "sig": []
}
```

**Response:**

```json
{
  "res": [4, "get_rpc_history", [[
    {
      "id": 123,
      "sender": "0x1234567890abcdef...",
      "req_id": 42,
      "method": "get_channels",
      "params": "[{\"participant\":\"0x1234567890abcdef...\"}]",
      "timestamp": 1619123456789,
      "req_sig": ["0x9876fedcba..."],
      "response": "{\"res\":[42,\"get_channels\",[[...]],1619123456799]}",
      "res_sig": ["0xabcd1234..."]
    },
    {
      "id": 122,
      "sender": "0x1234567890abcdef...",
      "req_id": 41,
      "method": "ping",
      "params": "[null]",
      "timestamp": 1619123446789,
      "req_sig": ["0x8765fedcba..."],
      "response": "{\"res\":[41,\"pong\",[],1619123446799]}",
      "res_sig": ["0xdcba4321..."]
    }
  ]], 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

## Virtual Application Management

### Create Virtual Application

Creates a virtual application between participants.

**Request:**

```json
{
  "req": [1, "create_app_session", [{
    "definition": {
      "protocol": "NitroRPC/0.2",
      "participants": [
        "0xAaBbCcDdEeFf0011223344556677889900aAbBcC",
        "0x00112233445566778899AaBbCcDdEeFf00112233"
      ],
      "weights": [50, 50],
      "quorum": 100,
      "challenge": 86400,
      "nonce": 1
    },
    "allocations": [
      {
        "participant": "0xAaBbCcDdEeFf0011223344556677889900aAbBcC",
        "asset": "usdc",
        "amount": "100.0"
      },
      {
        "participant": "0x00112233445566778899AaBbCcDdEeFf00112233",
        "asset": "usdc",
        "amount": "100.0"
      }
    ]
  }], 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

**Response:**

```json
{
  "res": [1, "create_app_session", [{
    "app_session_id": "0x3456789012abcdef...",
    "status": "open"
  }], 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

### Close Virtual Application

Closes a virtual application and redistributes funds.

**Request:**

```json
{
  "req": [1, "close_app_session", [{
    "app_session_id": "0x3456789012abcdef...",
    "allocations": [
      {
        "participant": "0xAaBbCcDdEeFf0011223344556677889900aAbBcC",
        "asset": "usdc",
        "amount": "0.0"
      },
      {
        "participant": "0x00112233445566778899AaBbCcDdEeFf00112233",
        "asset": "usdc",
        "amount": "200.0"
      }
    ]
  }], 1619123456789],
  "sig": ["0x9876fedcba...", "0x8765fedcba..."]
}
```

**Response:**

```json
{
  "res": [1, "close_app_session", [{
    "app_session_id": "0x3456789012abcdef...",
    "status": "closed"
  }], 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

### Close Channel

Closes a channel between a participant and the broker.

**Request:**

```json
{
  "req": [1, "close_channel", [{
    "channel_id": "0x4567890123abcdef...",
    "funds_destination": "0x1234567890abcdef..."
  }], 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

**Response:**

```json
{
  "res": [1, "close_channel", [{
    "channel_id": "0x4567890123abcdef...",
    "intent": 1,
    "version": 123,
    "state_data": "0x0000000000000000000000000000000000000000000000000000000000001ec7",
    "allocations": [
      {
        "destination": "0x1234567890abcdef...",
        "token": "0xeeee567890abcdef...",
        "amount": "50000"
      },
      {
        "destination": "0xbbbb567890abcdef...", // Broker address
        "token": "0xeeee567890abcdef...",
        "amount": "50000"
      }
    ],
    "state_hash": "0xLedgerStateHash",
    "server_signature": {
      "v": "27",
      "r": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
      "s": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
    }
  }], 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

### Resize Channel

Adjusts the capacity of a channel.

**Request:**

```json
{
  "req": [1, "resize_channel", [{
    "channel_id": "0x4567890123abcdef...",
    "allocate_amount": "20.0",
    "resize_amount": "100.0",
    "funds_destination": "0x1234567890abcdef..."
  }], 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

`allocate_amount` is how much more token user wants to allocate to this token-network specific channel from his unified balance.
`resize_amount` is how much user wants to deposit or withdraw from a token-network speecific channel.

Example:
- Initial state: user an open channel on Polygon with 20 usdc, and a channel on Celo with 5 usdc.
- User wants to deposit 75 usdc on Celo. User calls `resize_channel`, with `allocate_amount=0` and `resize_amount=75`.
- Now user's unified balance is 100 usdc (20 on Polygon and 80 on Celo).
- Now user wants wo withdraw all 100 usdc on Polygon. To withdraw, user must allocate 80 on this specific channel (`allocate_amount=80`), and resize it (`resize_amount=-100`). Also it is recommended to deallocate the channel on Celo (optional, but we may make this required in the future).

**Response:**

```json
{
  "res": [1, "resize_channel", [{
    "channel_id": "0x4567890123abcdef...",
    "state_data": "0x0000000000000000000000000000000000000000000000000000000000002ec7",
    "intent": 2,
    "version": 5,
    "allocations": [
      {
        "destination": "0x1234567890abcdef...",
        "token": "0xeeee567890abcdef...",
        "amount": "100000"
      },
      {
        "destination": "0xbbbb567890abcdef...", // Broker address
        "token": "0xeeee567890abcdef...",
        "amount": "0"
      }
    ],
    "state_hash": "0xLedgerStateHash",
    "server_signature": {
      "v": "28",
      "r": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
      "s": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
    }
  }], 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

The channel will be resized on the blockchain network where it was originally opened, as identified by the `chain_id` associated with the channel. The `new_amount` parameter specifies the desired capacity for the channel.

## Messaging

### Send Message in Virtual Application

Sends a message to all participants in a virtual app session.

**Request:**

```json
{
  "req": [1, "your_custom_method", [{
    "your_custom_field": "Hello, application participants!"
  }], 1619123456789],
  "sid": "0x3456789012abcdef...", // Virtual App Session ID
  "sig": ["0x9876fedcba..."]
}
```

### Send Response in Virtual Application

Responses can also be forwarded to all participants in a virtual application by including the AppSessionID `sid`:

```json
{
  "res": [1, "your_custom_method", [{
    "your_custom_field": "I confirm that I have received your message!"
  }], 1619123456789],
  "sid": "0x3456789012abcdef...", // Virtual App Session ID
  "sig": ["0x9876fedcba..."]
}
```

## Utility Methods

### Ping

Simple ping to check connectivity.

**Request:**

```json
{
  "req": [1, "ping", [], 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

**Response:**

```json
{
  "res": [1, "pong", [], 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

### Balance Updates

The server automatically sends balance updates to clients in these scenarios:
1. After successful authentication (as a welcome message)
2. After channel operations (open, close, resize)
3. After application operations (create, close)

Balance updates are sent as unsolicited server messages with the "bu" method:

```json
{
  "res": [1234567890123, "bu", [[
    {
      "asset": "usdc",
      "amount": "100.0"
    },
    {
      "asset": "eth",
      "amount": "0.5"
    }
  ]], 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

The balance update provides the latest balances for all assets in the participant's unified ledger, allowing clients to maintain an up-to-date view of available funds without explicitly requesting them.

### Open Channels

The server automatically sends all open channels as a batch update to clients after successful authentication.

```json
{
  "res": [1234567890123, "channels", [[
    {
      "channel_id": "0xfedcba9876543210...",
      "participant": "0x1234567890abcdef...",
      "status": "open",
      "token": "0xeeee567890abcdef...",
      "amount": "100000",
      "chain_id": 137,
      "adjudicator": "0xAdjudicatorContractAddress...",
      "challenge": 86400,
      "nonce": 1,
      "version": 2,
      "created_at": "2023-05-01T12:00:00Z",
      "updated_at": "2023-05-01T12:30:00Z"
    },
    {
      "channel_id": "0xabcdef1234567890...",
      "participant": "0x1234567890abcdef...",
      "status": "open",
      "token": "0xeeee567890abcdef...",
      "amount": "50000",
      "chain_id": 42220,
      "adjudicator": "0xAdjudicatorContractAddress...",
      "challenge": 86400,
      "nonce": 1,
      "version": 3,
      "created_at": "2023-04-15T10:00:00Z",
      "updated_at": "2023-04-20T14:30:00Z"
    }
  ]], 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

### Channel Updates

For channel updates, the server sends them in these scenarios:
1. When a channel is created
2. When a channel's status changes (open, joined, closed)
3. When a channel is resized

Individual channel updates are sent as unsolicited server messages with the "cu" method:

```json
{
  "res": [1234567890123, "cu", [{
    "channel_id": "0xfedcba9876543210...",
    "participant": "0x1234567890abcdef...",
    "status": "open",
    "token": "0xeeee567890abcdef...",
    "amount": "100000",
    "chain_id": 137,
    "adjudicator": "0xAdjudicatorContractAddress...",
    "challenge": 86400,
    "nonce": 1,
    "version": 2,
    "created_at": "2023-05-01T12:00:00Z",
    "updated_at": "2023-05-01T12:30:00Z"
  }], 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

The channel update contains the complete current state of a specific channel, allowing clients to maintain an up-to-date view of their channels without explicitly requesting them through the `get_channels` method.

### Get Configuration

Retrieves broker configuration information including supported networks.

**Request:**

```json
{
  "req": [1, "get_config", [], 1619123456789],
  "sig": []
}
```

**Response:**

```json
{
  "res": [1, "get_config", [{
    "broker_address": "0xbbbb567890abcdef...",
    "networks": [
      {
        "name": "polygon",
        "chain_id": 137,
        "custody_address": "0xCustodyContractAddress1..."
      },
      {
        "name": "celo",
        "chain_id": 42220,
        "custody_address": "0xCustodyContractAddress2..."
      },
      {
        "name": "base",
        "chain_id": 8453,
        "custody_address": "0xCustodyContractAddress3..."
      }
    ]
  }], 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

### Get Assets

Retrieves all supported assets. Optionally, you can filter the assets by chain_id.

**Request without filter:**

```json
{
  "req": [1, "get_assets", [], 1619123456789],
  "sig": []
}
```

**Request with chain_id filter:**

```json
{
  "req": [1, "get_assets", [{
    "chain_id": 137
  }], 1619123456789],
  "sig": []
}
```

**Response:**

```json
{
  "res": [1, "get_assets", [[{
    "token": "0xeeee567890abcdef...",
    "chain_id": 137,
    "symbol": "usdc",
    "decimals": 6
  },
  {
    "token": "0xffff567890abcdef...",
    "chain_id": 137,
    "symbol": "weth",
    "decimals": 18
  },
  {
    "token": "0xaaaa567890abcdef...",
    "chain_id": 42220,
    "symbol": "celo",
    "decimals": 18
  }]], 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

## Error Handling

When an error occurs, the server responds with an error message:

```json
{
  "res": [REQUEST_ID, "error", [{
    "error": "Error message describing what went wrong"
  }], 1619123456789],
  "sig": ["0xabcd1234..."]
}
```
