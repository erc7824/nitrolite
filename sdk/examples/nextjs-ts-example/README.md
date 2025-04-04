# Nitrolite NextJS TypeScript Example

This example demonstrates how to integrate the Nitrolite SDK with a NextJS application. It showcases a simple state channel application that uses the Nitrolite client to create, manage, and interact with state channels.

## Features

- Integration with MetaMask for wallet connection
- Nitrolite client initialization and configuration 
- State channel creation and management
- Channel state updates and visualization
- Counter application example with custom application logic
- WebSocket connection for real-time updates

## Getting Started

### Prerequisites

- Node.js 18+ and npm
- MetaMask extension installed in your browser

### Installation

1. Clone the repository:

    ```bash
    git clone https://github.com/erc7824/nitrolite.git
    cd nitrolite/sdk/examples/nextjs-ts-example
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

1. **Connect Wallet**: Connect your MetaMask wallet to the application.
2. **Open a Channel**: Specify a token address and amount to open a state channel.
3. **View Channel Status**: See the current state of your open channels.
4. **Update State**: Interact with the counter application to update the channel state.
5. **Close Channel**: Close the channel when you're done.

## Configuration

You can configure the application in the following files:

- `src/config/app.ts`: Application configuration including WebSocket URL and default addresses
- `src/config/contracts/index.ts`: Contract addresses for the Nitrolite framework 
- `src/config/chains/index.ts`: Blockchain network configuration

## Project Structure

```
src/
├── app/                  # Next.js app directory
│   ├── globals.css       # Global styles
│   ├── layout.tsx        # Root layout component
│   └── page.tsx          # Main page component
├── components/           # React components
│   ├── ChannelStatus.tsx # Channel status display
│   ├── MessageList.tsx   # Display messages and state updates
│   └── MetaMaskConnect.tsx # Wallet connection component
├── services/             # Application services
│   └── apps/             # Custom applications
│       └── counter.ts    # Counter application logic
├── hooks/                # Custom React hooks
│   ├── useChannelOpening.ts # Hook for opening channels
│   ├── useMessageService.ts # Hook for message handling
│   └── useNitroliteClient.ts # Hook for Nitrolite client
├── store/                # State management
│   ├── NitroliteStore.ts # Store for Nitrolite client state
│   └── WalletStore.ts    # Store for wallet connection state
├── types/                # TypeScript type definitions
└── config/               # Application configuration
    ├── app.ts            # Global app configuration
    ├── chains/           # Blockchain network configuration
    └── contracts/        # Contract addresses
```

## Integration with Nitrolite SDK

This example demonstrates key features of the Nitrolite SDK:

1. **Client Initialization**: Setup of the Nitrolite client with proper configuration
2. **Channel Management**: Opening and managing state channels
3. **Custom Application Logic**: Implementation of a counter application with custom state transition rules
4. **State Updates**: Handling state updates and signatures
5. **Multi-chain Support**: Configurable for different EVM chains

## Customizing the Example

To adapt this example for your own application:

1. Create a custom application logic class by implementing the `AppLogic` interface
2. Configure the contract addresses in `src/config/contracts/index.ts`
3. Update the UI components to match your application's requirements
4. Modify the state management to handle your application's state

## License

This project is licensed under the MIT License - see the LICENSE file for details.