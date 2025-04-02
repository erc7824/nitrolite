# Nitrolite NextJS TypeScript Example

This example demonstrates how to integrate the Nitrolite SDK (@erc7824/nitrolite) with a NextJS application. It showcases a simple WebSocket client that connects to a Nitrolite broker and enables real-time communication using the NitroliteRPC protocol.

## Features

- WebSocket connection to a Nitrolite broker server
- Cryptographic authentication using Ethereum-compatible keys
- Channel subscription and messaging
- Custom RPC requests using the NitroliteRPC protocol
- React hooks for easy state management
- Responsive UI with connection status indicators

## Getting Started

### Prerequisites

- Node.js 18+ and npm
- A running Nitrolite broker server (see main repository for setup instructions)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/erc7824/nitrolite.git
   cd nitrolite/sdk-ts/examples/nextjs-ts-example
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Start the development server:
   ```bash
   npm run dev
   ```

4. Open [http://localhost:3000](http://localhost:3000) in your browser to see the application.

## Usage

1. **Generate Keys**: Click the "Generate Keys" button to create a new key pair.
2. **Connect**: Connect to the WebSocket server using the generated keys.
3. **Subscribe to a Channel**: Select a channel from the dropdown and click "Subscribe".
4. **Send Messages**: Send messages to the subscribed channel.
5. **Custom RPC**: Send custom RPC requests to the server.

## Configuration

You can configure the WebSocket server URL in the `src/app/page.tsx` file:

```tsx
const { /* ... */ } = useWebSocket("ws://localhost:8000/ws");
```

## Project Structure

```
src/
├── app/                # Next.js app directory
│   ├── globals.css     # Global styles
│   ├── layout.tsx      # Root layout component
│   └── page.tsx        # Main page component
├── components/         # React components
│   ├── About.tsx               # Information about the app
│   ├── ConnectionStatus.tsx    # WebSocket connection status display
│   ├── MessageList.tsx         # Display received messages
│   └── RequestForm.tsx         # Form for sending requests
├── hooks/              # Custom React hooks
│   ├── useConnectionStatus.ts  # Hook for connection status styling
│   ├── useMessageStyles.ts     # Hook for message styling
│   └── useWebSocket.ts         # WebSocket connection hook
├── types/              # TypeScript type definitions
│   └── index.ts                # Common type definitions
└── utils/              # Utility functions
    └── wsClient.ts             # WebSocket client implementation
```

## Integration with Nitrolite SDK

This example uses the `@erc7824/nitrolite` SDK to handle WebSocket communication with the Nitrolite broker. Key features include:

1. **NitroliteRPC Protocol**: Uses the NitroliteRPC format for structured communication.
2. **Signing and Verification**: Implements cryptographic message signing and verification.
3. **Channel Management**: Demonstrates subscription to channels and message publication.
4. **Error Handling**: Implements robust error handling and reconnection logic.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
