# Nitrolite SDK React Demo

A comprehensive React demo application showcasing the full functionality of the Nitrolite TypeScript SDK with MetaMask integration.

## Features

This demo implements all the functionality from the Cerebro CLI in a user-friendly web interface:

### Setup & Configuration
- **MetaMask Integration**: Connect your wallet with one click
- **Node Configuration**: Configure Clearnode WebSocket URL
- **RPC Management**: Add/remove blockchain RPC endpoints

### High-Level Operations (Smart Client)
- **Deposit**: Deposit funds to payment channel (auto-creates channel if needed)
- **Withdraw**: Withdraw funds from payment channel to blockchain
- **Transfer**: Instant off-chain transfers to other wallets
- **Close Channel**: Finalize and close payment channel

### Node Information
- **Ping**: Test node connectivity
- **Node Info**: View node configuration and version
- **List Chains**: Display supported blockchains
- **List Assets**: Browse available assets and tokens

### User Queries
- **Get Balances**: View balances for any wallet address
- **Get Transactions**: Browse transaction history with pagination

### Low-Level State Management
- **Get Latest State**: Inspect channel state details
- **Get Home Channel**: View home channel information
- **Get Escrow Channel**: Query escrow channel by ID

### App Sessions
- **List App Sessions**: Query multi-party application sessions
- **Filter by Status**: View open or closed sessions

## Technology Stack

- **React 18** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server
- **Tailwind CSS** - Styling
- **Viem** - Ethereum interactions
- **MetaMask** - Wallet provider
- **Nitrolite SDK** - Payment channel operations

## Prerequisites

- Node.js 20 or higher
- MetaMask browser extension
- Access to a Clearnode instance
- RPC endpoints for target blockchains (e.g., Polygon Amoy)

## Installation

```bash
# Install dependencies
npm install
```

## Running the Demo

```bash
# Start development server
npm run dev
```

The application will be available at `http://localhost:3000`

## Usage Guide

### 1. Initial Setup

1. **Connect MetaMask**
   - Click "Connect MetaMask" button
   - Approve the connection in MetaMask popup
   - Your address will be displayed

2. **Configure Node URL**
   - Enter your Clearnode WebSocket URL (default: `wss://clearnode-v1-rc.yellow.org/ws`)
   - Example: `wss://clearnode.example.com/ws`

3. **Add Blockchain RPCs**
   - Enter Chain ID (e.g., `80002` for Polygon Amoy)
   - Enter RPC URL (e.g., `https://polygon-amoy.g.alchemy.com/v2/YOUR_KEY`)
   - Click "Add RPC"
   - Repeat for each blockchain you want to use

4. **Connect to Node**
   - Click "Connect to Node" button
   - Wait for successful connection

### 2. Making a Deposit

1. Navigate to "High-Level Operations" section
2. In the "Deposit" card:
   - Enter Chain ID (e.g., `80002`)
   - Enter Asset symbol (e.g., `usdc`)
   - Enter Amount (e.g., `100`)
3. Click "Deposit" button
4. Approve the transaction in MetaMask
5. Wait for confirmation

**Note**: First deposit will automatically create a new payment channel.

### 3. Making a Transfer

1. In the "Transfer" card:
   - Enter Recipient address (0x...)
   - Enter Asset symbol (e.g., `usdc`)
   - Enter Amount
2. Click "Transfer" button
3. Transaction completes instantly (off-chain)

### 4. Checking Balances

1. Navigate to "User Queries" section
2. In "Get Balances" card:
   - Address is pre-filled with your wallet
   - Or enter any wallet address to query
3. Click "Get Balances"
4. View all asset balances

### 5. Viewing Transaction History

1. In "Get Transactions" card:
   - Enter wallet address
2. Click "Get Transactions"
3. View recent transaction history with details

### 6. Withdrawing Funds

1. In "High-Level Operations" → "Withdraw" card:
   - Enter Chain ID
   - Enter Asset symbol
   - Enter Amount
2. Click "Withdraw"
3. Approve transaction in MetaMask
4. Funds will be transferred to your blockchain wallet

## Configuration Examples

### Polygon Amoy Testnet (Chain ID: 80002)
```
RPC URL: https://polygon-amoy.g.alchemy.com/v2/YOUR_API_KEY
Assets: USDC, WETH
```

### Ethereum Sepolia Testnet (Chain ID: 11155111)
```
RPC URL: https://eth-sepolia.g.alchemy.com/v2/YOUR_API_KEY
Assets: USDC, WETH
```

## Features Showcase

### Real-time Status Updates
- Success/error notifications for all operations
- Transaction hash display for on-chain operations
- Loading states for async operations

### Form Validation
- Required field validation
- Address format validation
- Amount validation

### Responsive Design
- Mobile-friendly interface
- Adaptive grid layouts
- Touch-friendly buttons

### Developer-Friendly
- Clear error messages
- Transaction details display
- State inspection tools

## Architecture

```
src/
├── App.tsx                 # Main application component
├── main.tsx               # Application entry point
├── index.css              # Global styles (Tailwind)
├── types.ts               # TypeScript type definitions
└── components/
    ├── SetupSection.tsx           # Wallet & node configuration
    ├── HighLevelOpsSection.tsx    # Deposit/withdraw/transfer/close
    ├── NodeInfoSection.tsx        # Node information queries
    ├── UserQueriesSection.tsx     # Balances & transactions
    ├── LowLevelSection.tsx        # State & channel queries
    ├── AppSessionsSection.tsx     # App sessions queries
    └── StatusBar.tsx              # Status notification component
```

## MetaMask Integration

The demo uses Viem's wallet client with MetaMask as the provider:

```typescript
// State signer (for off-chain operations)
const stateSigner = new EthereumMsgSigner(walletClient);

// Transaction signer (for on-chain operations)
const txSigner = new EthereumRawSigner(walletClient);

// Create SDK client
const client = await Client.create(
  wsURL,
  stateSigner,
  txSigner,
  withBlockchainRPC(chainId, rpcUrl)
);
```

## Troubleshooting

### MetaMask not detected
- Install MetaMask browser extension
- Refresh the page after installation

### Connection failed
- Check Clearnode URL is correct and accessible
- Verify WebSocket URL starts with `wss://` or `ws://`
- Check browser console for detailed error messages

### Transaction failed
- Ensure you have sufficient balance for gas fees
- Verify token approval for deposits
- Check RPC endpoint is configured for the chain
- Ensure you have funds in the channel for transfers/withdrawals

### Wrong network in MetaMask
- The demo works with any network MetaMask is connected to
- Transactions will be sent to the network currently selected in MetaMask
- Make sure MetaMask is on the correct network for your operations

## Production Considerations

This is a demo application. For production use:

1. **Error Handling**: Add comprehensive error handling and retry logic
2. **State Management**: Consider using Redux or Zustand for complex state
3. **Wallet Abstraction**: Support multiple wallet providers (WalletConnect, Coinbase Wallet)
4. **Security**: Never expose private keys or sensitive data
5. **Testing**: Add unit and integration tests
6. **Performance**: Optimize re-renders with React.memo and useMemo
7. **Accessibility**: Add ARIA labels and keyboard navigation
8. **Analytics**: Add event tracking for user actions

## Resources

- [Nitrolite Documentation](https://erc7824.org/quick_start)
- [GitHub Repository](https://github.com/erc7824/nitrolite)
- [Viem Documentation](https://viem.sh)
- [MetaMask Documentation](https://docs.metamask.io)

## License

MIT License - see LICENSE file for details
