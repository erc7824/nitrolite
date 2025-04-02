# Nitrolite SDK Examples

This directory contains example applications that demonstrate the usage of the Nitrolite SDK.

## NextJS TypeScript Example

[NextJS TypeScript Example](./nextjs-ts-example) - A complete frontend application demonstrating how to use the Nitrolite SDK with a React-based web application.

### Features Demonstrated

- **Client-Side WebSocket Integration**
  - WebSocket connection with automatic reconnection
  - Secure authentication using cryptographic keys
  - Real-time message handling and display

- **NitroliteRPC Protocol**
  - Message creation and signing
  - Request/response handling
  - Channel subscription and messaging

- **UI Components**
  - Connection status management
  - Message display with sender information
  - Form for sending custom RPC requests

- **React Hooks**
  - Custom hooks for WebSocket connection
  - State management for messages and connection status
  - Local storage integration for persistent keys

### Running the Example

```bash
# Navigate to the example directory
cd examples/nextjs-ts-example

# Install dependencies
npm install

# Start the development server
npm run dev
```

Then open [http://localhost:3000](http://localhost:3000) in your browser to see the application.

## Upcoming Examples

The following examples will be implemented in the future:

- **Broker Client Example** - A complete example demonstrating how to use the Nitrolite SDK to interact with the Virtual Ledger Broker with on-chain channel management, off-chain communication via WebSockets, and state updates through RPC.

- **RPC Counter Example** - A simple counter application demonstrating the use of the RPC protocol for off-chain state synchronization.

- **Virtual Channel Example** - An example demonstrating how to create and use a virtual channel through an intermediary.

- **Multi-party applications** - Examples showing how to handle multiple channels and participants.

- **Cross-chain operations** - Examples demonstrating cross-chain functionality.

- **Poker game with custom adjudicator** - A complete game implementation with a custom adjudicator.

- **Streaming data with micropayments** - A demonstration of real-time data streaming with per-use payments.


## Nitrolite RPC Example

[nitrolite-rpc-example.ts](./nitrolite-rpc-example.ts) - A simple example demonstrating how to use the NitroliteRPC protocol with WebSockets.

### Features Demonstrated

- Setting up a WebSocket connection to a broker
- Creating and signing RPC requests using NitroliteRPC
- Handling incoming RPC messages (requests, responses, errors)
- Managing pending requests with promises
- Implementing simple RPC methods (add, subtract, multiply)
- Error handling and proper connection cleanup

### Running the Example

```bash
# Install dependencies
npm install

# Run the example
npx ts-node examples/nitrolite-rpc-example.ts
```

Note: The examples are set up to use the default Hardhat addresses and private keys. In a real application, you would use your own keys and contract addresses.


