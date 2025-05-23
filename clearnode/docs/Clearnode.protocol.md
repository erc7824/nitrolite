# Clearnode Protocol Specification

## Overview
The Clearnode protocol is a system for managing payment channels and virtual applications between participants. It provides a secure, efficient way to conduct transactions off-chain while retaining the ability to settle on-chain when required, with support for multiple blockchain networks.

## Protocol Flow

### 1. Blockchain Channels and Credit
- The protocol accepts blockchain channels to credit participants' balances in the database ledger
- Participants create on-chain channels through custody contracts (supported on multiple chains including Polygon and Celo)
- Channel creation events from the blockchain are received through webhooks and processed by the `EventHandler`
- These events credit participants' balances in the internal ledger system
- Each participant has an `Account` in the ledger tied to their address.

### 2. Virtual Application Creation
- After being credited from on-chain channels, participants can create virtual applications with other participants
- Virtual applications allow participants to allocate a portion of their balance for peer-to-peer transactions without requiring on-chain operations
- The broker validates that:
  - All participants have channels with the broker
  - All participants have sufficient funds in their respective accounts
  - The requested allocation amounts are available
- Participants must provide signatures to authorize application creation
- The application can designate specific signers who will have authority over application closure
- Funds are transferred from participants' channel accounts to the new virtual application
- The broker sets up message routing between participants

### 3. Virtual Application Operations
- Participants send both requests and responses to each other through virtual applications using WebSocket connections
- Any message (request or response) with an AppID specified is forwarded to all other participants
- The broker maintains a real-time bidirectional communication layer for message routing

### 4. Virtual Application Closure and Settlement
- When participants wish to close a virtual application, all designated signers must provide signatures to authorize the closure
- The broker validates the signatures against the list of authorized signers registered during application creation
- The broker validates the final allocation of funds between participants
- The broker ensures the total allocated amount matches the total funds in the application
- Funds are transferred from the virtual application back to the participants' channels according to the final allocations
- The virtual application is marked as closed and message routing is discontinued
- When participants wish to materialize their balances on-chain, they can request the broker to re-open or update on-chain channels
- Settlement is only performed when requested by participants, allowing most transactions to remain off-chain
- Participants can have channels on multiple blockchain networks simultaneously

## Security Features

### Authentication and Authorization
- All operations are authenticated using cryptographic signatures
- The system uses ECDSA signatures compatible with Ethereum accounts
- Virtual applications implement a multi-signature scheme:
  - Application creation requires signatures from participating parties
  - Application closure requires signatures from all designated signers
- Weight-based quorum signatures are supported for application governance:
  - Each signer can be assigned a weight
  - A quorum threshold determines the minimum total weight required for valid decisions
  - This enables flexible governance models (m-of-n, third-party arbitration, etc.)
- The broker maintains persistent connections with participants through WebSockets
- Authentication uses a challenge-response mechanism to verify address ownership

### Multi-Chain Support
- The system supports multiple blockchain networks (currently Polygon and Celo)
- Each network has its own custody contract address and connection details
- Chain IDs are tracked with channels to ensure proper chain association
- Asset models track tokens per chain, with appropriate decimals for each token
- Channels are created, resized, and closed on their original blockchain network
- Participants can manage channels across multiple chains simultaneously
- The broker maintains separate custody contract instances for each supported network
- Event listeners monitor each blockchain network independently
- The `get_channels` method returns all channels for a participant across all supported chains

## Benefits
- Efficient, low-cost transactions by keeping most operations off-chain
- Security guarantees of blockchain when needed
- Participants can freely transact within their allocated funds in virtual applications
- On-chain settlement only occurs when participants choose to materialize their balances
- Cross-chain compatibility allows users to select their preferred blockchain network
- Multi-chain support provides resilience against network-specific issues
- Diverse token support across various blockchain ecosystems
- Flexibility to leverage the unique benefits of each supported blockchain

## RPC Message Format

All messages exchanged between clients and clearnodes follow this standardized format:

### Request Message

```json
{
  "req": [REQUEST_ID, METHOD, [PARAMETERS], TIMESTAMP],
  "sid": "APP_SESSION_ID", // AppId for Virtual Ledgers for Internal Communication
  "sig": ["SIGNATURE"]  // Client's signature of the entire "req" object
}
```

- The `sid` field serves as both the subject and destination pubsub topic for the message. There is a one-to-one mapping between topics and ledger accounts.
- The `sig` field contains one or more signatures, of the `req` data.

### Response Message

```json
{
  "res": [REQUEST_ID, METHOD, [RESPONSE_DATA], TIMESTAMP],
  "sid": "APP_SESSION_ID", // AppId for Virtual Ledgers for Internal Communication
  "int": [INTENT],// Allocation intent change
  "sig": ["SIGNATURE"]
}
```

- The `sid` field serves as both the subject and destination pubsub topic for the message. There is a one-to-one mapping between topics and ledger accounts.
- The `sig` field contains one or more signatures, of the `res` data.


The structure breakdown:

- `REQUEST_ID`: A unique identifier for the request/response pair (`uint64`)
- `METHOD`: The name of the method being called (`string`)
- `PARAMETERS`/`RESPONSE_DATA`: An array of parameters/response data (`[]any`)
- `TIMESTAMP`: Unix timestamp of the request/response in milliseconds (`uint64`)
- `APP_SESSION_ID` (`sid`): If specified, the message gets forwarded to all participants of a virtual app with thosn AppSessionID.
- `SIGNATURE`: Cryptographic signatures of the message (`[]string`). Multiple signatures may be required for certain operations.

## Data Types

### App Definition

```json
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
```


## Authentication Flow

The authentication process uses a challenge-response mechanism based on Ethereum signatures to verify that a client owns a particular Ethereum address.

### 1. Authentication Initialization

The client initiates authentication by sending an `auth_request` request with their address.

### 2. Challenge Response from Server

The server responds with a random string challenge token.

### 3. Authentication Verification

The client sends a verification request with the challenge token signed by the client's private key.

The server verifies that:

1. The challenge string is valid and not expired
2. The challenge was issued for the claimed address
3. The RPC message is signed by the address's private key

### 4. Authentication Success Response

If authentication is successful, the server responds with a success confirmation.

## Broker Configuration

The broker provides configuration information through the `get_config` method, including:

- Broker's Ethereum address
- Supported blockchain networks with:
  - Network name
  - Chain ID
  - Custody contract address

This allows clients to understand which networks are supported and how to interact with them.

## Peer-to-Peer Messaging

The broker supports bi-directional peer-to-peer messaging between participants in a virtual application. Both requests and responses can be forwarded between participants when they include AppID.

## Error Handling

When an error occurs, the server responds with an error message:

```json
{
  "res": [REQUEST_ID, "error", [{
    "error": "Error message describing what went wrong"
  }], TIMESTAMP],
  "sig": ["SIGNATURE"]
}
```

## Security Considerations

1. **Challenge Expiration**: Authentication challenges expire after 5 minutes
2. **One-time Use**: Each challenge can only be used once
3. **Rate Limiting**: The server limits the number of active challenges
4. **Signature Verification**: All RPC messages must be properly signed by the sender
5. **Session Management**: Sessions expire after a configurable period (default 24 hours)
6. **Address Binding**: Each challenge is stored with the address that requested it
7. **Random Challenge Strings**: Secure, random strings are used as challenge tokens
8. **Quorum Signatures**: Application closure requires signatures meeting or exceeding the quorum threshold
9. **Chain Association**: Each channel is firmly associated with its originating blockchain network

## Client Implementation Guidelines

1. **Authentication Flow**:
   - Begin by sending an `auth_request` request with your Ethereum address
   - Store the challenge string received from the server
   - Send an `auth_verify` request with your address and the challenge string
   - Store the session token and maintain it with regular activity

2. **Message Signing**:
   - Sign all RPC request messages with your private key
   - The signature proves ownership of the address
   - Verify signatures on all server responses for security

3. **Error Handling**:
   - Be prepared to handle session expiration
   - Implement reconnection and re-authentication logic
   - Handle rate limiting errors by implementing backoff strategies
   - Implement timeouts for all requests

4. **Multi-Chain Awareness**:
   - Check the broker's supported networks using `get_config`
   - Track channel `chain_id` values to ensure operations target the correct network
   - Be aware that channel operations (close, resize) must be performed on the blockchain where the channel was created

5. **Security Best Practices**:
   - Never reuse signatures across different sessions or services
   - Verify all message signatures from the server before processing
   - Ensure your private key is securely stored and never exposed
   - Generate a fresh unique identifier client-side for each request ID