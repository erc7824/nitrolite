# Session Keys

Session keys are delegated keys that enable applications to perform operations on behalf of a user's wallet with specified spending limits, permissions, and expiration times. They provide a secure way to grant limited access to applications without exposing the main wallet's private key.

> **Important:** Session keys are **no longer used as on-chain channel participant addresses** for new channels created after the v0.5.0 release. For all new channels, the wallet address is used directly as the participant address. However, session keys still function correctly for channels that were created before v0.5.0, ensuring backward compatibility.

## Core Concepts

### Applications

Each session key is associated with a specific **application name**, which identifies the application or service that will use the session key. This association serves several purposes:

- **Application Isolation**: Different applications get separate session keys, preventing one application from using another's delegated access
- **Access Control**: Operations performed with a session key are validated against the application specified during registration
- **Single Active Key**: Only one session key can be active per wallet+application combination. Registering a new session key for the same application automatically invalidates any existing session key for that application

> **Important:** After registering a session key, subsequent authentication requests with the same session key don't require you to specify all parameters again. The session key configuration is already stored in the system and will be used automatically.

#### Special Application: "clearnode"

Session keys registered with the application name `"clearnode"` receive special treatment:

- **Root Access**: These session keys bypass spending allowance validation and application restrictions
- **Full Permissions**: They can perform any operation the wallet itself could perform
- **Backward Compatibility**: This special behavior facilitates migration from older versions
- **Expiration Still Applies**: Even with root access, the session key expires according to its `expires_at` timestamp

> **Note:** The "clearnode" application name is primarily for backward compatibility and will be deprecated after a migration period for developers.

### Expiration

All session keys must have an **expiration timestamp** (`expires_at`) that defines when the session key becomes invalid:

- **Future Timestamp Required**: The expiration time must be set to a future date when registering a session key
- **Automatic Invalidation**: Once the expiration time passes, the session key can no longer be used for any operations
- **Grace Period**: After expiration, the session key is removed from the active keys list
- **Applies to All Keys**: Even "clearnode" application session keys must respect the expiration timestamp

### Allowances

Allowances define **spending limits** for session keys, specifying which assets the session key can spend and how much:

```json
{
  "allowances": [
    {
      "asset": "usdc",
      "amount": "100.0"
    },
    {
      "asset": "eth",
      "amount": "0.5"
    }
  ]
}
```

#### Allowance Validation

- **Supported Assets Only**: All assets specified in allowances must be supported by the system. Unsupported assets cause authentication to fail
- **Usage Tracking**: The system tracks spending per session key by recording which session key was used for each ledger debit operation
- **Spending Limits**: Once a session key reaches its spending cap for an asset, further operations requiring that asset are rejected with: `"operation denied: insufficient session key allowance: X required, Y available"`
- **Empty Allowances**: Providing an empty `allowances` array (`[]`) means zero spending allowed for all assets—any operation attempting to spend funds will be rejected

> **Important:** Only one session key is allowed per wallet+application combination. If you register a new session key for the same application, the old one is automatically invalidated and removed from the database.

#### Allowances for "clearnode" Application

Session keys with `application: "clearnode"` are exempt from allowance enforcement:

- **No Spending Limits**: Allowance checks are bypassed entirely
- **Full Financial Access**: These keys can spend any amount of any supported asset
- **Expiration Still Matters**: Even without allowance restrictions, the session key still expires according to its `expires_at` timestamp

## How to Manage Session Keys

### Create and Configure

To create a session key, use the `auth_request` method during authentication. This registers the session key with its configuration:

**Request:**

```json
{
  "req": [
    1,
    "auth_request",
    {
      "address": "0x1234567890abcdef...",
      "session_key": "0x9876543210fedcba...",
      "application": "Chess Game",
      "allowances": [
        {
          "asset": "usdc",
          "amount": "100.0"
        },
        {
          "asset": "eth",
          "amount": "0.5"
        }
      ],
      "scope": "app.create",
      "expires_at": 1762417328
    },
    1619123456789
  ],
  "sig": ["0x5432abcdef..."]
}
```

**Parameters:**

- `address` (required): The wallet address that owns this session key
- `session_key` (required): The address of the session key to register
- `application` (optional): Name of the application using this session key (defaults to "clearnode" if not provided)
- `allowances` (optional): Array of asset allowances specifying spending limits
- `scope` (optional): Permission scope (e.g., "app.create", "ledger.readonly")
- `expires_at` (required): Unix timestamp (in seconds) when this session key expires

> **Note:** When authenticating with an already registered session key, you must still provide all fields in the request. However, the values you provide will be ignored—the system uses the configuration stored during initial registration. This behavior will be improved in future versions.

### List Active Session Keys

Use the `get_session_keys` method to retrieve all active (non-expired) session keys for the authenticated user:

**Request:**

```json
{
  "req": [1, "get_session_keys", {}, 1619123456789],
  "sig": ["0x9876fedcba..."]
}
```

**Response:**

```json
{
  "res": [
    1,
    "get_session_keys",
    {
      "session_keys": [
        {
          "id": 1,
          "session_key": "0xabcdef1234567890...",
          "application": "Chess Game",
          "allowances": [
            {
              "asset": "usdc",
              "allowance": "100.0",
              "used": "45.0"
            },
            {
              "asset": "eth",
              "allowance": "0.5",
              "used": "0.0"
            }
          ],
          "scope": "app.create",
          "expires_at": "2024-12-31T23:59:59Z",
          "created_at": "2024-01-01T00:00:00Z"
        }
      ]
    },
    1619123456789
  ],
  "sig": ["0xabcd1234..."]
}
```

**Response Fields:**

- `id`: Unique identifier for the session key record
- `session_key`: The address of the session key
- `application`: Application name this session key is authorized for
- `allowances`: Array of allowances with usage tracking:
  - `asset`: Symbol of the asset (e.g., "usdc", "eth")
  - `allowance`: Maximum amount the session key can spend
  - `used`: Amount already spent by this session key
- `scope`: Permission scope (omitted if empty)
- `expires_at`: When this session key expires (ISO 8601 format)
- `created_at`: When the session key was created (ISO 8601 format)

### Revoke a Session Key

To immediately invalidate a session key, use the `revoke_session_key` method:

**Request:**

```json
{
  "req": [
    1,
    "revoke_session_key",
    {
      "session_key": "0xabcdef1234567890..."
    },
    1619123456789
  ],
  "sig": ["0x9876fedcba..."]
}
```

**Response:**

```json
{
  "res": [
    1,
    "revoke_session_key",
    {
      "session_key": "0xabcdef1234567890..."
    },
    1619123456789
  ],
  "sig": ["0xabcd1234..."]
}
```

**Permission Rules:**

- A wallet can revoke any of its session keys
- A session key can revoke itself
- A session key with `application: "clearnode"` can revoke other session keys belonging to the same wallet
- A non-"clearnode" session key cannot revoke other session keys (only itself)

**Important Notes:**

- Revocation is **immediate and cannot be undone**
- After revocation, any operations attempted with the revoked session key will fail with a validation error
- The revoked session key will no longer appear in the `get_session_keys` response
- Revocation is useful for security purposes when a session key may have been compromised

**Error Cases:**

- Session key does not exist, belongs to another wallet, or is expired: `"operation denied: provided address is not an active session key of this user"`
- Non-"clearnode" session key attempting to revoke another session key: `"operation denied: insufficient permissions for the active session key"`
