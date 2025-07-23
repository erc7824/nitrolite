---
sidebar_position: 3
title: NitroliteRPC
description: Overview of the NitroliteRPC, its core logic, and links to detailed API and type definitions.
keywords: [erc7824, statechannels, state channels, nitrolite, rpc, websockets, messaging, protocol]
---

import { Card, CardGrid } from '@site/src/components/Card';
import MethodDetails from '@site/src/components/MethodDetails';

# NitroliteRPC

The NitroliteRPC provides a secure, reliable real-time communication protocol for state channel applications. It enables off-chain message exchange, state updates, and channel management. This system is built around the `NitroliteRPC` class, which provides the foundational methods for message construction, signing, parsing, and verification.

<CardGrid cols={2}>
  <Card
    title="Message Creation API"
    description="Detailed reference for functions that create specific RPC request messages."
    to="./message_creation_api"
  />
  <Card
    title="RPC Type Definitions"
    description="Comprehensive documentation of all TypeScript types and interfaces used by the RPC system."
    to="./rpc_types"
  />
</CardGrid>

## Core Logic: The `NitroliteRPC` Class

The `NitroliteRPC` class is central to the RPC system. It offers a suite of static methods to handle the low-level details of the NitroliteRPC protocol.

### Message Creation

<MethodDetails
  name="createRequest"
  description="Constructs a standard RPC request object, forming the base structure for any request sent via the NitroliteRPC protocol."
  params={[
    { name: "requestId", type: "RequestID", description: "A unique identifier for the request." },
    { name: "method", type: "string", description: "The RPC method name." },
    { name: "params", type: "T (unknown[])", description: "An array of parameters for the RPC method." },
    { name: "timestamp", type: "Timestamp", description: "The current timestamp in milliseconds." }
  ]}
  returns="NitroliteRPCRequest<T>"
  example={`const request = NitroliteRPC.createRequest(1, "get_config", [], Date.now());`}
/>

<MethodDetails
  name="createAppRequest"
  description="Constructs an RPC request object specifically scoped to an application, including application ID and optional intent."
  params={[
    { name: "requestId", type: "RequestID", description: "A unique identifier for the request." },
    { name: "method", type: "string", description: "The RPC method name." },
    { name: "params", type: "T (unknown[])", description: "An array of parameters for the RPC method." },
    { name: "timestamp", type: "Timestamp", description: "The current timestamp in milliseconds." },
    { name: "appId", type: "AccountID", description: "The identifier of the target application or channel." },
    { name: "intent", type: "Intent", description: "Optional intent associated with the application request.", optional: true }
  ]}
  returns="NitroliteRPCRequest<T>"
  example={`const appRequest = NitroliteRPC.createAppRequest(1, "app_update", [{data: "0x123"}], Date.now(), "app_0xabc", Intent.UPDATE_STATE);`}
/>

<MethodDetails
  name="createResponse"
  description="Constructs a successful RPC response object, used to reply to a received request."
  params={[
    { name: "requestId", type: "RequestID", description: "The ID of the request this response corresponds to." },
    { name: "method", type: "string", description: "The original RPC method name." },
    { name: "data", type: "D (unknown)", description: "The payload/data of the response." },
    { name: "timestamp", type: "Timestamp", description: "The current timestamp in milliseconds." },
    { name: "appId", type: "AccountID", description: "Optional application/channel scope.", optional: true },
    { name: "intent", type: "Intent", description: "Optional intent associated with the response.", optional: true }
  ]}
  returns="NitroliteRPCResponse<D>"
  example={`const response = NitroliteRPC.createResponse(1, "get_config", { version: "1.0" }, Date.now());`}
/>

<MethodDetails
  name="createErrorResponse"
  description="Constructs an RPC error response object, used to indicate a failure in processing a request."
  params={[
    { name: "requestId", type: "RequestID", description: "The ID of the request this error response corresponds to." },
    { name: "method", type: "string", description: "The original RPC method name." },
    { name: "error", type: "RPCError", description: "An error object detailing the failure." },
    { name: "timestamp", type: "Timestamp", description: "The current timestamp in milliseconds." },
    { name: "appId", type: "AccountID", description: "Optional application/channel scope.", optional: true }
  ]}
  returns="NitroliteRPCResponse<null>"
  example={`const errorResponse = NitroliteRPC.createErrorResponse(1, "get_config", { code: -32601, message: "Method not found" }, Date.now());`}
/>

### Message Signing

<MethodDetails
  name="signRequestMessage"
  description="Signs an RPC request message using the provided signer function, adding a signature to the request object."
  params={[
    { name: "request", type: "NitroliteRPCRequest<T>", description: "The RPC request object to sign." },
    { name: "signer", type: "MessageSigner", description: "An async function that takes a message string and returns a Promise<Hex> (signature)." }
  ]}
  returns="Promise<SignedNitroliteRPCRequest<T>>"
  example={`// Assuming 'request' is a NitroliteRPCRequest and 'signer' is a MessageSigner
const signedRequest = await NitroliteRPC.signRequestMessage(request, signer);`}
/>

<MethodDetails
  name="signResponseMessage"
  description="Signs an RPC response message. This is used in specific scenarios where responses also require authentication."
  params={[
    { name: "response", type: "NitroliteRPCResponse<D>", description: "The RPC response object to sign." },
    { name: "signer", type: "MessageSigner", description: "An async function that takes a message string and returns a Promise<Hex> (signature)." }
  ]}
  returns="Promise<SignedNitroliteRPCResponse<D>>"
  example={`// Assuming 'response' is a NitroliteRPCResponse and 'signer' is a MessageSigner
const signedResponse = await NitroliteRPC.signResponseMessage(response, signer);`}
/>

### Message Parsing & Validation

<MethodDetails
  name="parseResponse"
  description="Parses and validates an incoming RPC response string or object, ensuring it conforms to the NitroliteRPC structure."
  params={[
    { name: "message", type: "string | object", description: "The raw RPC response message (JSON string or pre-parsed object)." }
  ]}
  returns="ParsedResponse<D, E>"
  example={`try {
  const parsed = NitroliteRPC.parseResponse(incomingMessageString);
  if (parsed.success) {
    // process parsed.result
  } else {
    // process parsed.error
  }
} catch (e) {
  // handle malformed message
}`}
/>

These methods ensure that all communication adheres to the defined RPC structure and security requirements.

## Generic Message Structure

The `NitroliteRPC` class operates on messages adhering to the following general structures. For precise details on each field and for specific message types, please refer to the [RPC Type Definitions](./rpc_types).

```typescript
// Generic Request message structure
{
  "req": [requestId, method, params, timestamp], // Core request tuple
  "int"?: Intent,          // Optional intent for state changes
  "acc"?: AccountID,       // Optional account scope (channel/app ID)
  "sig": [signature]       // Array of signatures
}

// Generic Response message structure
{
  "res": [requestId, method, dataPayload, timestamp], // Core response tuple
  "acc"?: AccountID,       // Optional account scope
  "int"?: Intent,          // Optional intent
  "sig"?: [signature]      // Optional signatures for certain response types
}
```

## Next Steps

Dive deeper into the specifics of the RPC system:

<CardGrid cols={2}>
  <Card
    title="Message Creation API Details"
    description="Explore the high-level functions for constructing various RPC requests."
    to="./message_creation_api"
  />
  <Card
    title="Browse All RPC Types"
    description="Examine the detailed data structures and type definitions used in NitroliteRPC."
    to="./rpc_types"
  />
</CardGrid>