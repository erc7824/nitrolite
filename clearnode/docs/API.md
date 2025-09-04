# Clearnode API Reference

## API Endpoints

| Method                    | Description                                                              | Access  |
| ------------------------- | ------------------------------------------------------------------------ | ------- |
| `auth_request`            | Initiates authentication with the server                                 | Public  |
| `auth_challenge`          | Server response with authentication challenge                            | Public  |
| `auth_verify`             | Completes authentication with a challenge response                       | Public  |
| `ping`                    | Simple connectivity check                                                | Public  |
| `get_config`              | Retrieves broker configuration and supported networks                    | Public  |
| `get_assets`              | Retrieves all supported assets (optionally filtered by chain_id)         | Public  |
| `get_channels`            | Lists all channels for a participant with their status across all chains | Public  |
| `get_app_definition`      | Retrieves application definition for a ledger account                    | Public  |
| `get_app_sessions`        | Lists virtual applications for a participant with optional status filter | Public  |
| `get_ledger_entries`      | Retrieves detailed ledger entries for a participant                      | Public  |
| `get_ledger_transactions` | Retrieves transaction history with optional filtering                    | Public  |
| `get_user_tag`            | Retrieves user's tag                                                     | Private |
| `get_rpc_history`         | Retrieves all RPC message history for a participant                      | Private |
| `get_ledger_balances`     | Lists participants and their balances for a ledger account               | Private |
| `transfer`                | Transfers funds from user's unified balance to another account           | Private |
| `create_app_session`      | Creates a new virtual application on a ledger                            | Private |
| `submit_app_state`        | Submits an intermediate state into a virtual application                 | Private |
| `close_app_session`       | Closes a virtual application                                             | Private |
| `create_channel`          | Returns data and Broker signature to open a channel                      | Private |
| `close_channel`           | Returns data and Broker signature to close a channel                     | Private |
| `resize_channel`          | Returns data and Broker signature to adjust channel capacity             | Private |

**Legenda**

| Access | Authentication required |
|--------|-------------------------|
| Private | Yes |
| Public | No |

## Authentication

### Authentication Request

Initiates authentication with the server.

**Request:**

```json
{
  "req": [1, "auth_request", {
    "address": "0x1234567890abcdef...",
    "session_key": "0x9876543210fedcba...", // If specified, enables delegation to this key
    "app_name": "Example App", // Application name for analytics
    "allowances": [ // Asset allowances for the session
      {
        "asset": "usdc", 
        "amount": "100.0"
      }
    ],
    "scope": "app.create", // Permission scope (e.g., "app.create", "ledger.readonly")
    "expire": "3600", //  Session expiration
    "application": "0xApp1234567890abcdef..." // Application public address
  }, 1619123456789],
  "sig": ["0x5432abcdef..."] // Client's signature of the entire 'req' object
}
```

### Authentication Challenge

Server response with a challenge token for the client to sign.

**Response:**

```json
{
  "res": [1, "auth_challenge", {
    "challenge_message": "550e8400-e29b-41d4-a716-446655440000"
  }, 1619123456789],
  "sig": ["0x9876fedcba..."] // Server's signature of the entire 'res' object
}
```

### Authentication Verification

Completes authentication with a challenge response.

**Request:**

```json
{
  "req": [2, "auth_verify", {
    "challenge": "550e8400-e29b-41d4-a716-446655440000"
  }, 1619123456789],
  "sig": ["0x2345bcdef..."] // Client's EIP-712 signatures of the challenge data object
}
```

**Response:**

```json
{
  "res": [2, "auth_verify", {
    "address": "0x1234567890abcdef...",
    "success": true,
    "jwt_token": "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9..." // JWT token for subsequent requests
  }, 1619123456789],
  "sig": ["0xabcd1234..."] // Server's signature of the entire 'res' object
}
```

#### JWT Authentication

After successful authentication, the server provides a JWT token that can be used for subsequent authenticated requests. The JWT contains:

- Policy information with wallet address, participant, scope, and expiration
- Permission scopes (e.g., "app.create", "ledger.readonly")
- Asset allowances (if specified during auth_request)
- Standard JWT claims (issued at, expiration, etc.)

The JWT token has a default validity period of 24 hours and must be refreshed by making a new authentication request before expiration.

## Ledger Management

### Get Channels

Retrieves all channels for a participant (both open, closed, and joining), ordered by creation date (newest first). This method returns channels across all supported chains. If no participant is specified, it returns all channels.
Supports pagination and sorting by providing optional request parameters and metadata fields in response.

> Sorted descending by `created_at` by default.

**Request:**

```json
{
  "req": [1, "get_channels", {}, 1619123456789],
  "sig": []
}
```

**Request with pagination and sorting:**

```json
{
  "req": [1, "get_channels", {
    "participant": "0x1234567890abcdef...", // Optional: filter by participant
    "status":"open", // Optional filter
    "offset": 42, // Optional: pagination offset
    "limit": 10, // Optional: number of channels to return
    "sort": "desc" // Optional: sort asc or desc by created_at
  }, 1619123456789],
  "sig": []
}
```

**Response:**

```json
{
  "res": [1, "get_channels", {
    "channels" : [
      {
        "channel_id": "0xfedcba9876543210...",
        "participant": "0x1234567890abcdef...",
        "wallet": "0x1234567890abcdef...",
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
        "wallet": "0x1234567890abcdef...",
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
    ],
    "metadata": {
      "page": 5,
      "per_page": 10,
      "total_count": 56,
      "page_count": 6
    }
  }, 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

The signature in the request must be from the participant's private key, verifying they own the address. This prevents unauthorized access to channel information.

Each channel response includes:

- `channel_id`: Unique identifier for the channel
- `participant`: The participant's address
- `wallet`: The wallet address associated with this channel (may differ from participant if using delegation)
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
  
Metadata fields provide pagination information:

- `page`: Current page number
- `per_page`: Number of channels per page
- `total_count`: Total number of channels available
- `page_count`: Total number of pages based on the `per_page` limit

### Get App Definition

Retrieves the application definition for a specific ledger account.

**Request:**

```json
{
  "req": [1, "get_app_definition", {
    "app_session_id": "0x1234567890abcdef..."
  }, 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

**Response:**

```json
{
  "res": [1, "get_app_definition", {
    "protocol": "NitroRPC/0.2",
    "participants": [
      "0xAaBbCcDdEeFf0011223344556677889900aAbBcC",
      "0x00112233445566778899AaBbCcDdEeFf00112233"
    ],
    "weights": [50, 50],
    "quorum": 100,
    "challenge": 86400,
    "nonce": 1
  }, 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

### Get App Sessions

Lists all virtual applications for a participant sorted by updated_at from the newest to oldest. Optionally, you can filter the results by status (open, closed).
Supports pagination and sorting.

> Sorted descending by `created_at` by default.

**Request:**

```json
{
  "req": [1, "get_app_sessions", {}, 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

**Request with filtering, pagination, and sorting:**

```json
{
  "req": [1, "get_app_sessions", {
    "participant": "0x1234567890abcdef...",  // Optional: filter by participant
    "status": "open",  // Optional: filter by status
    "offset": 42, // Optional: pagination offset
    "limit": 10, // Optional: number of sessions to return
    "sort": "asc", // Optional: sort asc or desc
  }, 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

**Response:**

```json
{
  "res": [1, "get_app_sessions", {
    "app_sessions" : [
      {
        "app_session_id": "0x3456789012abcdef...",
        "status": "open",
        "participants": [
          "0x1234567890abcdef...",
          "0x00112233445566778899AaBbCcDdEeFf00112233"
        ],
        "session_data": "{\"gameType\":\"rps\",\"rounds\":5,\"currentRound\":3,\"scores\":{\"0x1234567890abcdef\":2,\"0x00112233445566778899AaBbCcDdEeFf00112233\":1}}",
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
        "session_data": "{\"gameType\":\"snake\",\"boardSize\":20,\"snakeLength\":5,\"score\":150,\"level\":3,\"gameState\":\"active\"}",
        "protocol": "NitroSnake",
        "challenge": 86400,
        "weights": [70, 30],
        "quorum": 100,
        "version": 1,
        "nonce": 123456790
      }
    ],
    "metadata": {
      "page": 5,
      "per_page": 10,
      "total_count": 56,
      "page_count": 6
    }
  }, 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

### Get Ledger Balances

Retrieves the balances of all participants in a specific ledger account.

**Request:**

```json
{
  "req": [1, "get_ledger_balances", {
    "participant": "0x1234567890abcdef...", // TO BE DEPRECATED
    // OR
    "account_id": "0x1234567890abcdef..."
  }, 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

If no params passed, it returns the ledger balance of user's wallet.
To get balance in a specific virtual app session, specify `app_session_id` as account_id.

**Response:**

```json
{
  "res": [1, "get_ledger_balances", {
    "ledger_balances": [
      {
        "asset": "usdc",
        "amount": "100.0"
      },
      {
        "asset": "eth",
        "amount": "0.5"
      }
    ],
  }, 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

### Get User Tag

Retrieves the user's tag, which can be used for transfer operations. The tag is a unique identifier for the user.

**Request:**

```json
{
  "req": [1, "get_user_tag", {}, 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

**Response:**

```json
{
  "res": [1, "get_user_tag", {
    "tag": "UX123D",
  }, 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

### Transfer Funds

This method allows a user to transfer assets from their unified balance to another account. The user must have sufficient funds for each asset being transferred. The operation will fail if any of the specified assets have insufficient funds.

User may specify the `destination` (wallet address) or `destination_user_tag` (user tag) to identify the recipient. `destination_user_tag` is used if and only if the `destination` field is empty.

CAUTION: Invalid destination address may result in loss of funds.
Currently, `Transfer` supports ledger account of another user as destination (wallet address is identifier).

**Request:**

```json
{
  "req": [1, "transfer", {
    "destination": "0x9876543210abcdef...",
    "allocations": [
      {
        "asset": "usdc",
        "amount": "50.0"
      },
      {
        "asset": "eth",
        "amount": "0.1"
      }
    ]
  }, 1619123456789],
  "sig": ["0x9876fedcba..."]
}

// OR

{
  "req": [1, "transfer", {
    "destination_user_tag": "UX123D",
    "allocations": [
      {
        "asset": "usdc",
        "amount": "50.0"
      },
      {
        "asset": "eth",
        "amount": "0.1"
      }
    ]
  }, 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

**Response:**

```json
{
  "res": [1, "transfer", {
    "transactions" : [
      {
        "id": "1",
        "tx_type": "transfer",
        "from_account": "0x1234567890abcdef...",
        "from_account_tag": "NQKO7C",
        "to_account": "0x9876543210abcdef...",
        "to_account_tag": "UX123D",
        "asset": "usdc",
        "amount": "50.0",
        "created_at": "2023-05-01T12:00:00Z"
      },
      {
        "id": "2",
        "tx_type": "transfer",
        "from_account": "0x1234567890abcdef...",
        "from_account_tag": "NQKO7C",
        "to_account": "0x9876543210abcdef...",
        "to_account_tag": "UX123D",
        "asset": "eth",
        "amount": "0.1",
        "created_at": "2023-05-01T12:00:00Z"
      }
    ]
  }, 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

The response returns an array of transaction objects, with one transaction for each asset being transferred.

Each transaction includes:

- `id`: Unique transaction reference
- `tx_type`: Transaction type (transfer/deposit/withdrawal/app_deposit/app_withdrawal)
- `from_account`: The account that sent the funds
- `from_account_tag`: The user tag for the sender account (empty if no tag exists or not a wallet account)
- `to_account`: The account that received the funds  
- `to_account_tag`: The user tag for the recipient account (empty if no tag exists or not a wallet account)
- `asset`: The asset symbol that was transferred
- `amount`: The amount transferred for this specific asset
- `created_at`: When the transaction was created (ISO 8601 format)

### Get Ledger Entries

Retrieves the detailed ledger entries for an account, providing a complete transaction history. This can be used to audit all deposits, withdrawals, and transfers. If no filter is specified, returns all entries, otherwise applies one or multiple filters.
Supports pagination and sorting.

> Sorted descending by `created_at` by default.

**Request:**

```json
{
  "req": [1, "get_ledger_entries", {}, 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

**Request with filtering, pagination, and sorting:**

```json
{
  "req": [1, "get_ledger_entries", {
    "account_id": "0x1234567890abcdef...", // Optional: filter by account ID
    "wallet": "0x1234567890abcdef...", // Optional: filter by participant
    "asset": "usdc", // Optional: filter by asset
    "offset": 42, // Optional: pagination offset
    "limit": 10, // Optional: number of entries to return
    "sort": "desc" // Optional: sort asc or desc by created_at
  }, 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

**Response:**

```json
{
  "res": [1, "get_ledger_entries", {
    "ledger_entries": [
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
    ],
    "metadata": {
      "page": 5,
      "per_page": 10,
      "total_count": 56,
      "page_count": 6
    }
  }, 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

### Get Transactions

Retrieves ledger transaction history with optional filtering by asset and transaction type. This endpoint provides a view of transactions where the specified account appears as either the sender or receiver.
Supports pagination and sorting.

> Sorted descending by `created_at` by default.

**Available Transaction Types:**

- `transfer`: Direct transfers between unified accounts
- `deposit`: Funds deposited to a unified account
- `withdrawal`: Funds withdrawn from a unified account
- `app_deposit`: Deposits into to application sessions
- `app_withdrawal`: Withdrawals from application sessions

**Request:**

```json
{
  "req": [1, "get_ledger_transactions", {}, 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

**Request with filtering, pagination, and sorting:**

```json
{
  "req": [1, "get_ledger_transactions", {
    "account_id": "0x1234567890abcdef...", // Optional: filter by account ID
    "asset": "usdc",     // Optional: filter by asset
    "tx_type": "transfer", // Optional: filter by transaction type
    "offset": 42, // Optional: pagination offset
    "limit": 10, // Optional: number of transactions to return
    "sort": "desc" // Optional: sort asc or desc by created_at
  }, 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

**Response:**

```json
{
  "res": [1, "get_ledger_transactions", {
    "ledger_transactions":[
      {
        "id": "1",
        "tx_type": "transfer",
        "from_account": "0x1234567890abcdef...",
        "from_account_tag": "NQKO7C",
        "to_account": "0x9876543210abcdef...",
        "to_account_tag": "UX123D",
        "asset": "usdc",
        "amount": "50.0",
        "created_at": "2023-05-01T12:00:00Z"
      },
      {
        "id": "2",
        "tx_type": "deposit",
        "from_account": "0x9876543210abcdef...", // Channel account
        "from_account_tag": "", // Channel accounts does not have tags
        "to_account": "0x1234567890abcdef...",
        "to_account_tag": "UX123D",
        "asset": "usdc",
        "amount": "25.0",
        "created_at": "2023-05-01T10:30:00Z"
      }
    ],
    "metadata": {
      "page": 5,
      "per_page": 10,
      "total_count": 56,
      "page_count": 6
    }
  }, 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

Each transaction response includes:

- `id`: Unique transaction id reference
- `tx_type`: Transaction type (transfer/deposit/withdrawal/app_deposit/app_withdrawal)
- `from_account`: The account that sent the funds
- `from_account_tag`: The user tag for the sender account (empty if no tag exists or not a wallet account)
- `to_account`: The account that received the funds
- `to_account_tag`: The user tag for the recipient account (empty if no tag exists or not a wallet account)
- `asset`: The asset symbol (e.g., "usdc", "eth")
- `amount`: The transaction amount
- `created_at`: When the transaction was created (ISO 8601 format)

Transactions are ordered by creation date (newest first). If no `account_id` is provided, returns all transactions. The `asset` and `tx_type` filters can be used to narrow results to specific asset types or transaction types.

### Get RPC History

Retrieves all RPC messages history for a participant, ordered by timestamp (newest first).

**Request:**

```json
{
  "req": [4, "get_rpc_history", {}, 1619123456789],
  "sig": []
}
```

**Response:**

```json
{
  "res": [4, "get_rpc_history", {
    "rpc_entries": [
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
    ]
  }, 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

## Virtual Application Management

### Create Virtual Application

Creates a virtual application between participants.
Participants must agree on signature weights and a quorum; this quorum is required to submit an intermediate state or close an app session. The create app session request must be signed by all participants with non-zero allocations.

The optional `session_data` field can be used to store application-specific data that will be preserved throughout the session lifecycle. This enables applications to maintain custom state information such as configuration settings, business logic state, or any other data needed for the application.

**Request:**

```json
{
  "req": [1, "create_app_session", {
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
    ],
    "session_data": "{\"gameType\":\"chess\",\"timeControl\":{\"initial\":600,\"increment\":5},\"maxPlayers\":2,\"gameState\":\"waiting\"}"
  }, 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

**Response:**

```json
{
  "res": [1, "create_app_session", {
    "app_session_id": "0x3456789012abcdef...",
    "version": "1",
    "status": "open"
  }, 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

### Submit Application State

Submits an intermediate state into a virtual application and redistributes funds in an app session.
To submit an intermediate state, participants must reach the signature quorum that they agreed on when creating the app session.
This means that the sum of the weights of signers must reach the specified threshold in the app definition.

The optional `session_data` field can be used to update application-specific data associated with the session, allowing applications to track progress, update configurations, or store any custom state changes during the session lifecycle.

**Request:**

```json
{
  "req": [1, "submit_app_state", {
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
    ],
    "session_data": "{\"gameType\":\"chess\",\"timeControl\":{\"initial\":600,\"increment\":5},\"maxPlayers\":2,\"gameState\":\"finished\",\"winner\":\"0x00112233445566778899AaBbCcDdEeFf00112233\",\"endCondition\":\"checkmate\"}" // Optional
  }, 1619123456789],
  "sig": ["0x9876fedcba...", "0x8765fedcba..."]
}
```

**Response:**

```json
{
  "res": [1, "submit_app_state", {
    "app_session_id": "0x3456789012abcdef...",
    "version": "567",
    "status": "open"
  }, 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

### Close Virtual Application

Closes a virtual application and redistributes funds.
To close the app session, participants must reach the signature quorum that they agreed on when creating the app session.
This means that the sum of the weights of signers must reach the specified threshold in the app definition.

The optional `session_data` field can be used to provide final application-specific data when closing the session, such as final results, completion status, or any other concluding information that should be preserved with the session closure.

**Request:**

```json
{
  "req": [1, "close_app_session", {
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
    ],
    "session_data": "{\"gameType\":\"chess\",\"timeControl\":{\"initial\":600,\"increment\":5},\"maxPlayers\":2,\"gameState\":\"closed\",\"winner\":\"0x00112233445566778899AaBbCcDdEeFf00112233\",\"endCondition\":\"checkmate\",\"moveHistory\":[\"e2e4\",\"e7e5\",\"Nf3\",\"Nc6\"]}"
  }, 1619123456789],
  "sig": ["0x9876fedcba...", "0x8765fedcba..."]
}
```

**Response:**

```json
{
  "res": [1, "close_app_session", {
    "app_session_id": "0x3456789012abcdef...",
    "version": "3",
    "status": "closed"
  }, 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

### Create Channel

Requests opening a channel with a Clearnode broker on a specific network.

**Request:**

```json
{
  "req": [1, "create_channel", [{
    "chain_id": 137,
    "token": "0xeeee567890abcdef...",
    "amount": "100000000",
    "session_key": "0x1234567890abcdef..." // Optional
  }], 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

The request parameters are:
- `chain_id`: The blockchain network ID where the channel should be created
- `token`: The token contract address for the channel
- `amount`: The initial amount to deposit in the channel (in raw token units)
- `session_key`: An optional session key for channel operations

**Response:**

Returns signed initial state with the requested amounts ready to submit on Blockchain.

```json
{
  "res": [1, "create_channel", [{
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
  }], 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

The response includes:
- `channel_id`: Unique identifier for the channel
- `channel`: Channel definition with participants, adjudicator, challenge period, and nonce
- `state`: Complete initial state structure containing intent, version, state_data, and allocations
- `server_signature`: Broker's signature of the state

### Close Channel

To close a channel, the user must request the final state signed by the broker and then submit it to the smart contract.
Only an open channel can be closed. In case the user does not agree with the final state provided by the broker, they can call the `challenge` method directly on the smart contract.

**Request:**

```json
{
  "req": [1, "close_channel", {
    "channel_id": "0x4567890123abcdef...",
    "funds_destination": "0x1234567890abcdef..."
  }, 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

In the request, the user must specify funds destination. After the channel is closed, funds become available to withdraw from the smart contract for the specified address.

**Response:**

```json
{
  "res": [1, "close_channel", {
    "channel_id": "0x4567890123abcdef...",
    "state": {
      "intent": 3, // IntentFINALIZE - constant specifying that this is a final state
      "version": 123,
      "state_data": "0xc0ffee", 
      "allocations": [
        {
          "destination": "0x1234567890abcdef...", // Provided funds address
          "token": "0xeeee567890abcdef...",
          "amount": "50000"
        },
        {
          "destination": "0xbbbb567890abcdef...", // Broker address
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

### Resize Channel

Adjusts the capacity of a channel.

**Request:**

```json
{
  "req": [1, "resize_channel", {
    "channel_id": "0x4567890123abcdef...",
    "allocate_amount": "200000000",
    "resize_amount": "1000000000",
    "funds_destination": "0x1234567890abcdef..."
  }, 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

`allocate_amount` is how much more token user wants to allocate to this token-network specific channel from his unified balance.
`resize_amount` is how much user wants to deposit or withdraw from a token-network specific channel.

Example:

- Initial state: user an open channel on Polygon with 20 usdc, and a channel on Celo with 5 usdc.
- User wants to deposit 75 usdc on Celo. User calls `resize_channel`, with `allocate_amount=0` and `resize_amount=75`.
- Now user's unified balance is 100 usdc (20 on Polygon and 80 on Celo).
- Now user wants wo withdraw all 100 usdc on Polygon. To withdraw, user must allocate 80 on this specific channel (`allocate_amount=80`), and resize it (`resize_amount=-100`). Also it is recommended to deallocate the channel on Celo (optional, but we may make this required in the future).
- Note: all amounts are shown as examples. In your requests, you must specify these amounts in raw format.

**Response:**

```json
{
  "res": [1, "resize_channel", {
    "channel_id": "0x4567890123abcdef...",
    "state": {
      "intent": 2, // IntentRESIZE
      "version": 5,
      "state_data": "0xc0ffee",
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
      ]
    },
    "server_signature": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1c"
  }, 1619123456789],
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
  "req": [1, "your_custom_method", {
    "your_custom_field": "Hello, application participants!"
  }, 1619123456789],
  "sid": "0x3456789012abcdef...", // Virtual App Session ID
  "sig": ["0x9876fedcba..."]
}
```

### Send Response in Virtual Application

Responses can also be forwarded to all participants in a virtual application by including the AppSessionID `sid`:

```json
{
  "res": [1, "your_custom_method", {
    "your_custom_field": "I confirm that I have received your message!"
  }, 1619123456789],
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
  "req": [1, "ping", {}, 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

**Response:**

```json
{
  "res": [1, "pong", {}, 1619123456789],
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
  "res": [1234567890123, "bu", {
    "balance_updates": [
      {
        "asset": "usdc",
        "amount": "100.0"
      },
      {
        "asset": "eth",
        "amount": "0.5"
      }
    ]
  }, 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

The balance update provides the latest balances for all assets in the participant's unified ledger, allowing clients to maintain an up-to-date view of available funds without explicitly requesting them.

### Open Channels

The server automatically sends all open channels as a batch update to clients after successful authentication.

```json
{
  "res": [1234567890123, "channels", {
    "channels": [
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
    ]
  }, 1619123456789],
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
  "res": [1234567890123, "cu", {
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
  }, 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

The channel update contains the complete current state of a specific channel, allowing clients to maintain an up-to-date view of their channels without explicitly requesting them through the `get_channels` method.

### Transfer Notifications

The server automatically sends transfer notifications to clients when funds are transferred to their account. These notifications inform the recipient about incoming transfers.

Transfer notifications are sent as unsolicited server messages with the "transfer" method:

```json
{
  "res": [1234567890123, "tr", {
    "transactions": [
      {
        "id": "1",
        "tx_type": "transfer",
        "from_account": "0x9876543210abcdef...",
        "from_account_tag": "ABC123",
        "to_account": "0x1234567890abcdef...",
        "to_account_tag": "XYZ789",
        "asset": "usdc",
        "amount": "50.0",
        "created_at": "2023-05-01T12:00:00Z"
      },
      {
        "id": "2",
        "tx_type": "transfer",
        "from_account": "0x9876543210abcdef...",
        "from_account_tag": "ABC123",
        "to_account": "0x1234567890abcdef...",
        "to_account_tag": "XYZ789",
        "asset": "weth",
        "amount": "0.1",
        "created_at": "2023-05-01T12:00:00Z"
      }
    ]
  }, 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

The transfer notification contains an array of transaction objects representing the incoming transfers. Each transaction includes:

- `id`: Unique transaction reference
- `tx_type`: Transaction type (currently only "transfer" for these notifications)
- `from_account`: The account that sent the funds
- `from_account_tag`: The user tag for the sender account (empty if no tag exists)
- `to_account`: The account that received the funds (the notification recipient)
- `to_account_tag`: The user tag for the recipient account (empty if no tag exists)
- `asset`: The asset symbol that was transferred
- `amount`: The amount transferred for this specific asset
- `created_at`: When the transaction was created (ISO 8601 format)

### Get Configuration

Retrieves broker configuration information including supported networks.

**Request:**

```json
{
  "req": [1, "get_config", {}, 1619123456789],
  "sig": []
}
```

**Response:**

```json
{
  "res": [1, "get_config", {
    "broker_address": "0xbbbb567890abcdef...",
    "networks": [
      {
        "chain_id": 137,
        "custody_address": "0xCustodyContractAddress1...",
        "adjudicator_address":"0xCustodyContractAddress1..."
      },
      {
        "chain_id": 42220,
        "custody_address": "0xCustodyContractAddress2...",
        "adjudicator_address":"0xCustodyContractAddress1..."
      },
      {
        "chain_id": 8453,
        "custody_address": "0xCustodyContractAddress3...",
        "adjudicator_address":"0xCustodyContractAddress1..."
      }
    ]
  }, 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

### Get Assets

Retrieves all supported assets. Optionally, you can filter the assets by chain_id.

> Sorted ascending by `symbol` by default.

**Request without filter:**

```json
{
  "req": [1, "get_assets", {}, 1619123456789],
  "sig": []
}
```

**Request with chain_id filter:**

```json
{
  "req": [1, "get_assets", {
    "chain_id": 137
  }, 1619123456789],
  "sig": []
}
```

**Response:**

```json
{
  "res": [1, "get_assets", {
    "assets": [
      {
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
      }
    ]
  }, 1619123456789],
  "sig": ["0xabcd1234..."]
}
```

## Error Handling

When an error occurs, the server responds with an error message:

```json
{
  "res": [REQUEST_ID, "error", {
    "error": "Error message describing what went wrong"
  }, 1619123456789],
  "sig": ["0xabcd1234..."]
}
```
