# Hachi SDK Examples

This directory contains example applications that demonstrate the usage of the Hachi SDK.

## Broker Client Example

[broker-client.ts](./broker-client.ts) - A complete example demonstrating how to use the Hachi SDK to interact with the Virtual Ledger Broker. It showcases on-chain channel management, off-chain communication via WebSockets, and state updates through RPC.

### Features Demonstrated

- **WebSocket Provider Integration**
  - Custom implementation of the `RPCProvider` interface for WebSocket communication
  - Connection handling and message parsing
  - Error handling and reconnection logic

- **Channel Management**
  - Creating direct on-chain channels with the broker
  - Establishing virtual channels through the broker
  - Managing channel state and signatures

- **NitroRPC Protocol**
  - Implementing the signed request/response protocol
  - Handling method registration and message routing
  - Proper timestamp handling for state ordering

- **Error Handling**
  - Comprehensive error checking and recovery
  - Proper validation before operations
  - Type-safe error handling with the SDK's error hierarchy

- **Token Operations**
  - Balance checking
  - Allowance management
  - Automatic approval when needed

### Running the Example

1. Start the broker service:
   ```bash
   cd ../broker
   go run main.go
   ```

2. In a new terminal, run the client example:
   ```bash
   cd hachi-sdk-ts
   npm run build
   npx ts-node examples/broker-client.ts
   ```

3. To run with a real private key (replace with your own):
   ```bash
   PRIVATE_KEY=0x... npx ts-node examples/broker-client.ts
   ```

### Configuration

Edit the example configuration at the bottom of the file to match your environment:

```typescript
// Configuration
const PRIVATE_KEY = '0x0000000000000000000000000000000000000000000000000000000000000001' as Hex; // Replace with real private key
const BROKER_URL = 'ws://localhost:8080';
const BROKER_ADDRESS = '0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266' as Address;
const TOKEN_ADDRESS = '0xc778417E063141139Fce010982780140Aa0cD5Ab' as Address; // WETH on Sepolia
const CONTRACT_ADDRESSES = {
  custody: '0x5FbDB2315678afecb367f032d93F642f64180aa3' as Address,
  adjudicators: {
    sequential: '0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512' as Address,
    numeric: '0x9fE46736679d2D9a65F0992F2272dE9f3c7fa6e0' as Address,
    trivial: '0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9' as Address
  }
};
```

## RPC Counter Example

[rpc-counter.ts](./rpc-counter.ts) - A simple counter application that demonstrates the use of the RPC protocol for off-chain state synchronization.

### Features Demonstrated

- Setting up RPC clients and providers
- Creating a numeric state channel
- Enhancing a channel with RPC communication
- Exchanging state updates between participants
- Handling state updates and displaying counter values
- Closing the channel

### Running the Example

To run the example, you'll need a local Ethereum node running (like Hardhat):

```bash
# Start a local Hardhat node in a separate terminal
npx hardhat node

# Run the example
npx ts-node examples/rpc-counter.ts
```

## Virtual Channel Example

[virtual-channel.ts](./virtual-channel.ts) - An example demonstrating how to create and use a virtual channel through an intermediary.

### Features Demonstrated

- Setting up direct channels between participants
- Creating a Light Virtual Channel Identifier (LVCI)
- Using an intermediary (Bob) to connect Alice and Charlie
- Sending state updates through the virtual channel
- Closing the virtual channel

### Running the Example

```bash
# Start a local Hardhat node in a separate terminal
npx hardhat node

# Run the example
npx ts-node examples/virtual-channel.ts
```

Note: The examples are set up to use the default Hardhat addresses and private keys. In a real application, you would use your own keys and contract addresses.

## Micropayment Streaming Example

[micropayment-streaming.ts](./micropayment-streaming.ts) - A comprehensive example of implementing a pay-per-second streaming service using state channels.

### Features Demonstrated

- Creating and funding a sequential payment channel
- Implementing per-second micropayments
- Real-time off-chain state updates
- Comprehensive error handling with recovery strategies
- RPC communication between participants
- Complete channel lifecycle management and proper cleanup

### Running the Example

```bash
# Start a local Hardhat node in a separate terminal
npx hardhat node

# Run the example
npx ts-node examples/micropayment-streaming.ts
```

### Application Architecture

The example implements a `StreamingService` class that demonstrates how to structure a real-world application:

- **Error handling**: Uses a custom `ErrorHandler` with retry logic and error-specific handling
- **Graceful state management**: Properly manages streaming state and recovery
- **Clean separation of concerns**: Client setup, channel management, and streaming logic are separated
- **Resource cleanup**: Ensures all resources are properly cleaned up, even when errors occur

## Tic-Tac-Toe Game Example

[tic-tac-toe.ts](./tic-tac-toe.ts) - A turn-based game example that follows the state channel pattern in the README sequence diagram.

### Features Demonstrated

- Custom application state encoding/decoding
- Turn-based state validation rules
- Custom game logic implementation
- State transitions with proper validation
- Visual game state representation
- Full game lifecycle from setup to settlement

### Running the Example

```bash
# Start a local Hardhat node in a separate terminal
npx hardhat node

# Run the example
npx ts-node examples/tic-tac-toe.ts
```

### Application Architecture

This example demonstrates:

- **Custom Application Logic**: The game implements rules for a turn-based tic-tac-toe game
- **State Verification**: Each move is validated to ensure game rules are followed
- **Off-Chain State Updates**: Players exchange moves off-chain
- **Win Detection**: The game detects when a player has won
- **On-Chain Settlement**: The final game state is settled on-chain when complete

This example directly implements the sequence diagram in the README, showing how Alice and Bob (Player 1 and Player 2) exchange off-chain state updates, with the blockchain as the final arbiter.

## Upcoming Examples

More examples planned for the future:

- Multi-party applications with multiple channels
- Cross-chain operations
- Poker game with custom adjudicator
- Streaming data with micropayments