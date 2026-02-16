# Nitro RPC

NitroRPC is an asynchronous rpc message format designed for state channels,
The Requests are signed by the initiator and the Response are countersigned.

The use of a server side universal millisecond timestamp builds a temper proof history.

## General definition

Every message is formatted in a fixed JSON array (size is always 4 elements) as follow:

`[request_id, method, params, ts]`

**request_id:** is an unsigned integer unique identifier of the request, response must copy the same request_id
**method**: remote method name to be invoked as string
**params**: method params is a json array, it can be empty or contain flexible mix of json types.
**ts**: an unsigned integer 64 bit representing server timestamp in millisecond

## NitroRPC Request

In this example, the client is calling a method named `"subtract"` with positional parameters:

```json
{
  "req": [1001, "subtract", [42, 23], 1741344819012],
  cid: "0x3ce9bc2e2053488bde06de257541e05fc58fd8c4f149cca44f1c702fcbdbde0aa"
  out: ["Serialized array of Allocations"]
  "sig": ["0xa0ad67f51cc73aee5b874ace9bc2e2053488bde06de257541e05fc58fd8c4f149cca44f1c702fcbdbde0aa09bcd24456f465e5c3002c011a3bc0f317df7777d2"]
}
```

- req: rpc message payload `[request_id, method, params, ts]`
- cid: ChannelId from channel creation
- out: Outcome allocations can be optional
- sig: payload client stateHash signature

The millisecond timestamp was returned previously from the server, it is used as a height for proof of history

## NitroRPC Response

For a successful invocation, the server might respond like this:

```json
{
  "res": [1001, "subtract", [19], 1741344819814],
  cid: "0x3ce9bc2e2053488bde06de257541e05fc58fd8c4f149cca44f1c702fcbdbde0aa"
  out: ["Serialized array of Allocations"]
  "sig": ["0xd73268362b04516451ec52170f5c8ca189d35d9ac5e9041c156c9f0faf9aebd2891309e3b2b5d8788578ab3449c96f7aa81aefb25482b53f02bac42c65f806e5"]
}
```

- res: rpc message payload `[request_id, method, params, ts]`
- cid: ChannelId from channel creation
- out: Outcome allocations can be optional
- sig: payload server response stateHash signature

**ts**: server response with latest timestamp

## Go Types

```go
type RPCMessage struct {
  RequestID uint64
  Method    string
  Params    []any
  Result    []any
  Timestamp uint64
}
```

## Solidity types

```solidity
struct RPCMessage {
    uint64 requestID;
    string method;
    bytes params;
    bytes result;
    uint64 timestamp;
}

struct RPCError {
  uint32 code;
  string message;
}
```

### NitroRPC StateHash

Requester and Responder must sign the RPCStateHash build in the following way:

```solidity
keccak256(
  abi.encode(
    reqId,      // Request unique Id act as a nounce
    method,     // Method name
    params,     // Client method params
    result,     // Server response
    timestamp,  // timestamp from the request
  )
);

// Alternatively from RPCMessage type
keccak256(
  abi.encode(
    rpcMessage
  )
)
```

### Signatures

RPCStateHash are signed using curve25519, without EIP-191 prefix as the protocol intend to be chain agnostic.
In case the server rpc handler return an error in result field, it is NOT recommended to countersign the request.

## Adjudication rules

RPCAdjudicator does not provide a complete trustless security between parties,
but rather functionalities which makes it easy to use and reasonably secure for creating audit trails.

The intention that those audit trails can then be processed by a DAO or Consortium of auditors to resolve dispute. This works in the very same way most electronic trade disputes is working today on internet through arbitration courts and online dispute resolution forums (ODR).

### Rules enforced

- Request RPCStateHash must be signed by the Client
- Response RPCStateHash must be signed by the Server
- Request MUST contain the most recent and accurate timestamp provided by the middleman, server or NTP.
- Response stateHash must include the Request timestamp and provide a new one for the next requests.
- Client can choose the timestamp provided as long as it's accurate.
- Response MUST contains the same RequestId nonce
- Client SHOULD use bitmask for ordering unique nonce between browser tabs or websocket connections.
- Adjudication state is only a Pair of Request/Response with respective signatures
- A Request cannot be accepted as a valid state, since Server cannot confirm having received the command.
- At least one proofs must be provided with most recent timestamp
- When we are validating the first RPC Request/Response pair, proof candidate.data MUST contains the magic number OPENCHAN = 7879 (0x1EC7)

### State storage example

```sql
-- Table to store RPC messages
CREATE TABLE rpc_states (
    id SERIAL PRIMARY KEY,
    ts BIGINT,
    req_id INTEGER NOT NULL,
    method VARCHAR(255) NOT NULL,
    params JSONB NOT NULL,
    result JSONB NOT NULL,
    client_sig VARCHAR(128),
    server_sig VARCHAR(128),
    UNIQUE (request_id)
);
```

#### Markdown Table Example

| ID   | Timestamp (ts)     | Method         | Request Arguments                  | Response Arguments                |
|------|--------------------|----------------|------------------------------------|-----------------------------------|
| 1003 | 1741344821000      | add    | [50]              | [50]                      |
| 1004 | 1741344822000      | mul   | [2]                      | [100]                  |
| 1005 | 1741344823000      | sub | [10]                     | [90]        |
