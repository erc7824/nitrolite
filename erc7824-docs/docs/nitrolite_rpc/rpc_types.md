---
sidebar_position: 4 # Adjusted sidebar_position if message_creation_api is 3
title: RPC Type Definitions
description: Comprehensive type definitions for the Nitrolite RPC protocol.
keywords: [erc7824, statechannels, state channels, nitrolite, rpc, types, typescript, protocol, api, definitions]
---

# Nitrolite RPC Type Definitions

This page provides a comprehensive reference for all TypeScript types, interfaces, and enums used by the Nitrolite RPC system, as defined in the `@erc7824/nitrolite` SDK. These definitions are crucial for understanding the structure of messages exchanged with the Nitrolite broker.

## Core Types

These are fundamental types used throughout the RPC system.

### `RequestID`
A unique identifier for an RPC request. Typically a number.
```typescript
export type RequestID = number;
```

### `Timestamp`
Represents a Unix timestamp in milliseconds. Used for message ordering and security.
```typescript
export type Timestamp = number;
```

### `AccountID`
A unique identifier for a channel or application session, represented as a hexadecimal string.
```typescript
export type AccountID = Hex; // from 'viem'
```

### `Intent`
Represents the allocation intent change as an array of big integers. This is used to specify how funds should be re-distributed in a state update.
```typescript
export type Intent = bigint[];
```

## Message Payloads

These types define the core data arrays within RPC messages.

### `RequestData`
The structured data payload within a request message.
```typescript
export type RequestData = [RequestID, string, any[], Timestamp?];
```
-   `RequestID`: The unique ID of this request.
-   `string`: The name of the RPC method being called.
-   `any[]`: An array of parameters for the method.
-   `Timestamp?`: An optional timestamp for when the request was created.

### `ResponseData`
The structured data payload within a successful response message.
```typescript
export type ResponseData = [RequestID, string, any[], Timestamp?];
```
-   `RequestID`: The ID of the original request this response is for.
-   `string`: The name of the original RPC method.
-   `any[]`: An array containing the result(s) of the method execution.
-   `Timestamp?`: An optional timestamp for when the response was created.

### `NitroliteRPCErrorDetail`
Defines the structure of the error object within an error response.
```typescript
export interface NitroliteRPCErrorDetail {
    error: string;
}
```
-   `error`: A string describing the error that occurred.

### `ErrorResponseData`
The structured data payload for an error response message.
```typescript
export type ErrorResponseData = [RequestID, "error", [NitroliteRPCErrorDetail], Timestamp?];
```
-   `RequestID`: The ID of the original request this error is for.
-   `"error"`: A literal string indicating this is an error response.
-   `[NitroliteRPCErrorDetail]`: An array containing a single `NitroliteRPCErrorDetail` object.
-   `Timestamp?`: An optional timestamp for when the error response was created.

### `ResponsePayload`
A union type representing the payload of a response, which can be either a success (`ResponseData`) or an error (`ErrorResponseData`).
```typescript
export type ResponsePayload = ResponseData | ErrorResponseData;
```

## Message Envelopes

These interfaces define the overall structure of messages sent over the wire.

### `NitroliteRPCMessage`
The base wire format for Nitrolite RPC messages.
```typescript
export interface NitroliteRPCMessage {
    req?: RequestData;
    res?: ResponsePayload;
    int?: Intent;
    sig?: Hex[];
}
```
-   `req?`: The request payload, if this is a request message.
-   `res?`: The response payload, if this is a response message.
-   `int?`: Optional allocation intent change.
-   `sig?`: Optional array of cryptographic signatures (hex strings).

## Parsing Results

### `ParsedResponse`
Represents the result of parsing an incoming Nitrolite RPC response message.
```typescript
export interface ParsedResponse {
    isValid: boolean;
    error?: string;
    isError?: boolean;
    requestId?: RequestID;
    method?: string;
    data?: any[] | NitroliteRPCErrorDetail;
    acc?: AccountID;
    int?: Intent;
    timestamp?: Timestamp;
}
```
-   `isValid`: `true` if the message was successfully parsed and passed basic structural validation.
-   `error?`: If `isValid` is `false`, contains a description of the parsing or validation error.
-   `isError?`: `true` if the parsed response represents an error (i.e., `method === "error"`). Undefined if `isValid` is `false`.
-   `requestId?`: The `RequestID` from the response payload. Undefined if the structure is invalid.
-   `method?`: The method name from the response payload. Undefined if the structure is invalid.
-   `data?`: The extracted data payload (result array for success, `NitroliteRPCErrorDetail` object for error). Undefined if the structure is invalid or the error payload is malformed.
-   `acc?`: The `AccountID` from the message envelope, if present.
-   `int?`: The `Intent` from the message envelope, if present.
-   `timestamp?`: The `Timestamp` from the response payload. Undefined if the structure is invalid.

## Request Parameter Structures

These interfaces define the expected parameters for specific RPC methods.

### `AppDefinition`
Defines the structure of an application's configuration.
```typescript
export interface AppDefinition {
    protocol: string;
    participants: Hex[];
    weights: number[];
    quorum: number;
    challenge: number;
    nonce?: number;
}
```
-   `protocol`: The protocol identifier or name for the application logic (e.g., `"NitroRPC/0.2"`).
-   `participants`: An array of participant addresses (Ethereum addresses as `Hex`) involved in the application.
-   `weights`: An array representing the relative weights or stakes of participants. Order corresponds to the `participants` array.
-   `quorum`: The number/percentage of participants (based on weights) required to reach consensus.
-   `challenge`: A parameter related to the challenge period or mechanism (e.g., duration in seconds).
-   `nonce?`: An optional unique number (nonce) used to ensure the uniqueness of the application instance and prevent replay attacks.

### `CreateAppSessionRequest`
Parameters for the `create_app_session` RPC method.
```typescript
export interface CreateAppSessionRequest {
    definition: AppDefinition;
    token: Hex;
    allocations: bigint[];
}
```
-   `definition`: The `AppDefinition` object detailing the application being created.
-   `token`: The `Hex` address of the ERC20 token contract used for allocations within this application session.
-   `allocations`: An array of `bigint` representing the initial allocation distribution among participants. The order corresponds to the `participants` array in the `definition`.

Example:
```json
{
  "definition": {
    "protocol": "NitroRPC/0.2",
    "participants": [
      "0xAaBbCcDdEeFf0011223344556677889900aAbBcC",
      "0x00112233445566778899AaBbCcDdEeFf00112233"
    ],
    "weights": [100, 0], // Example: Participant 1 has 100% weight
    "quorum": 100,      // Example: 100% quorum needed
    "challenge": 86400, // Example: 1 day challenge period
    "nonce": 12345
  },
  "token": "0xTokenContractAddress00000000000000000000",
  "allocations": ["1000000000000000000", "0"] // 1 Token for P1, 0 for P2 (as strings for bigint)
}
```

### `CloseAppSessionRequest`
Parameters for the `close_app_session` RPC method.
```typescript
export interface CloseAppSessionRequest {
    app_id: Hex;
    allocations: bigint[];
}
```
-   `app_id`: The unique `AccountID` (as `Hex`) of the application session to be closed.
-   `allocations`: An array of `bigint` representing the final allocation distribution among participants upon closing. Order corresponds to the `participants` array in the application's definition.

### `ResizeChannel`
Parameters for the `resize_channel` RPC method.
```typescript
export interface ResizeChannel {
    channel_id: Hex;
    participant_change: bigint;
    funds_destination: Hex;
}
```
-   `channel_id`: The unique `AccountID` (as `Hex`) of the direct ledger channel to be resized.
-   `participant_change`: The `bigint` amount by which the participant's allocation in the channel should change (positive to add funds, negative to remove).
-   `funds_destination`: The `Hex` address where funds will be sent if `participant_change` is negative (withdrawal), or the source of funds if positive (though typically handled by prior on-chain deposit).

## Function Types (Signers & Verifiers)

These types define the signatures for functions used in cryptographic operations.

### `MessageSigner`
A function that signs a message payload.
```typescript
export type MessageSigner = (payload: RequestData | ResponsePayload) => Promise<Hex>;
```
-   Takes a `RequestData` or `ResponsePayload` object (the array part of the message).
-   Returns a `Promise` that resolves to the cryptographic signature as a `Hex` string.

### `SingleMessageVerifier`
A function that verifies a single message signature.
```typescript
export type SingleMessageVerifier = (
  payload: RequestData | ResponsePayload,
  signature: Hex,
  address: Address // from 'viem'
) => Promise<boolean>;
```
-   Takes the `RequestData` or `ResponsePayload` object, the `Hex` signature, and the expected signer's `Address`.
-   Returns a `Promise` that resolves to `true` if the signature is valid for the given payload and address, `false` otherwise.

## Usage Examples

### Creating Message Payloads and Envelopes
```typescript
// Example Request Payload (for a 'ping' method)
const pingRequestData: RequestData = [1, "ping", []]; // Assuming timestamp is added by sender utility

// Example Request Envelope
const pingRequestMessage: NitroliteRPCMessage = {
  req: pingRequestData,
  // sig: ["0xSignatureIfPreSigned..."] // Signature added by signing utility
};

// Example Application-Specific Request
const appActionData: RequestData = [2, "message", [{ move: "rock" }], Date.now()];
const appActionMessage: ApplicationRPCMessage = {
  sid: "0xAppSessionId...",
  req: appActionData,
  // sig: ["0xSignature..."]
};

// Example Successful Response Payload
const pongResponseData: ResponseData = [1, "ping", ["pong"], Date.now()];

// Example Error Detail
const errorDetail: NitroliteRPCErrorDetail = { error: "Method parameters are invalid." };

// Example Error Response Payload
const errorResponseData: ErrorResponseData = [2, "error", [errorDetail], Date.now()];

// Example Response Envelope (Success)
const successResponseEnvelope: NitroliteRPCMessage = {
  res: pongResponseData,
};
```

### Working with Signers (Conceptual)
```typescript
// Conceptual: How a MessageSigner might be used
async function signAndSend(payload: RequestData, signer: MessageSigner, sendMessageToServer: (msg: string) => void) {
  const signature = await signer(payload);
  const message: NitroliteRPCMessage = {
    req: payload,
    sig: [signature]
  };
  sendMessageToServer(JSON.stringify(message));
}
```

## Implementation Considerations

When working with these types:

1.  **Serialization**: Messages are typically serialized to JSON strings for transmission (e.g., over WebSockets).
2.  **Signing**: Payloads (`req` or `res` arrays) are what get signed, not the entire envelope. The resulting signature is then added to the `sig` field of the `NitroliteRPCMessage` envelope.
3.  **Validation**: Always validate the structure and types of incoming messages against these definitions, preferably using utilities provided by the SDK.
4.  **Error Handling**: Properly check for `isError` in `ParsedResponse` and use `NitroliteErrorCode` to understand the nature of failures.
5.  **BigInts**: Note the use of `bigint` for `Intent` and allocation amounts. Ensure your environment and serialization/deserialization logic handle `bigint` correctly (e.g., converting to/from strings for JSON).
6.  **Hex Strings**: Types like `AccountID`, `Hex` (for signatures, token addresses) imply hexadecimal string format (e.g., `"0x..."`).
