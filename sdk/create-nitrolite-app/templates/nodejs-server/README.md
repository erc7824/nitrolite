# {{projectName}}

A Node.js server application built with the Nitrolite SDK for state channel applications.

## Features

- 🚀 **Express.js HTTP Server** - RESTful API endpoints
- 📡 **WebSocket Server** - Real-time bidirectional communication
- ⚡ **Nitrolite SDK Integration** - State channel functionality
- 🔐 **EIP-712 Authentication** - Wallet-based authentication
- 📝 **TypeScript** - Full type safety
- 🔥 **Hot Reload** - Development-friendly auto-restart
- 🛡️ **Error Handling** - Comprehensive error handling and logging
- 📊 **Health Checks** - Production-ready health monitoring

## Getting Started

### Prerequisites

- Node.js >= 20.0.0
- npm or yarn
- A wallet with some ETH for testing (if using mainnet/testnet)

### Installation

1. **Clone and install dependencies:**
   ```bash
   cd {{projectName}}
   npm install
   ```

2. **Configure environment variables:**
   ```bash
   cp .env.example .env
   ```
   
   Edit `.env` and configure the required variables:
   ```env
   # Server Configuration
   PORT=3001
   NODE_ENV=development
   
   # Nitrolite Configuration
   YELLOW_WS_URL=wss://clearnet.yellow.com/ws
   ASSET=usdc
   
   # Wallet Configuration (REQUIRED in production)
   WALLET_PRIVATE_KEY=0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef
   
   # App Configuration
   VAPP_NAME={{projectName}}
   VAPP_SCOPE={{packageName}}
   ```

### Development

Start the development server with hot reload:

```bash
npm run dev
```

The server will start on `http://localhost:3001` (or your configured PORT).

### Production

Build and start the production server:

```bash
npm run build
npm start
```

## Project Structure

```
{{projectName}}/
├── src/
│   ├── config/
│   │   └── index.ts          # Environment configuration
│   ├── services/
│   │   ├── nitrolite/
│   │   │   ├── client.ts     # Nitrolite WebSocket client
│   │   │   ├── auth.ts       # Authentication utilities
│   │   │   └── types.ts      # Nitrolite type definitions
│   │   └── websocket.ts      # WebSocket message handling
│   ├── middleware/
│   │   └── auth.ts           # Authentication middleware
│   ├── routes/
│   │   ├── health.ts         # Health check endpoints
│   │   └── nitrolite.ts      # Nitrolite API endpoints
│   ├── stores/
│   │   ├── RequestStore.ts   # Request/response tracking
│   │   └── MessageStore.ts   # Message history storage
│   ├── types/
│   │   └── index.ts          # Shared type definitions
│   ├── utils/
│   │   ├── crypto.ts         # Cryptographic utilities
│   │   ├── logger.ts         # Logging utility
│   │   └── shutdown.ts       # Graceful shutdown handling
│   └── server.ts             # Main server entry point
├── dist/                     # Compiled JavaScript (after build)
├── .env.example              # Environment variables template
├── package.json
├── tsconfig.json
└── README.md
```

## API Endpoints

### Health Checks

- `GET /health` - Basic health check
- `GET /health/detailed` - Detailed health check with service status
- `GET /health/ready` - Readiness probe for deployments
- `GET /health/live` - Liveness probe for deployments

### Nitrolite API

- `GET /api/nitrolite/status` - Get Nitrolite connection status
- `POST /api/nitrolite/ping` - Send ping to Nitrolite network (requires auth)
- `POST /api/nitrolite/send` - Send message to Nitrolite network (requires auth)
- `GET /api/nitrolite/session` - Get session information (requires auth)
- `GET /api/nitrolite/ws/clients` - Get WebSocket client information
- `POST /api/nitrolite/ws/broadcast` - Broadcast message to all WebSocket clients
- `POST /api/nitrolite/ws/disconnect/:clientId` - Disconnect a specific WebSocket client

## WebSocket API

The server provides a WebSocket API for real-time communication. Connect to `ws://localhost:3001`.

### Message Format

All WebSocket messages follow this format:

```typescript
interface WebSocketMessage {
  type: string;
  payload?: any;
  timestamp?: number;
}
```

### Supported Message Types

#### Client Messages

1. **Ping**
   ```json
   {
     "type": "ping"
   }
   ```

2. **Forward to Nitrolite**
   ```json
   {
     "type": "nitrolite_message",
     "payload": {
       "method": "assets",
       "params": {...}
     }
   }
   ```

3. **Get Status**
   ```json
   {
     "type": "status"
   }
   ```

#### Server Messages

1. **Welcome**
   ```json
   {
     "type": "welcome",
     "clientId": "client_123...",
     "timestamp": 1234567890,
     "nitroliteStatus": {
       "connected": true,
       "status": "connected"
     }
   }
   ```

2. **Nitrolite Messages**
   ```json
   {
     "type": "nitrolite_message",
     "data": {...},
     "timestamp": 1234567890
   }
   ```

3. **Status Updates**
   ```json
   {
     "type": "nitrolite_status",
     "status": "connected",
     "timestamp": 1234567890
   }
   ```

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PORT` | Server port | `3001` | No |
| `NODE_ENV` | Environment | `development` | No |
| `YELLOW_WS_URL` | Nitrolite broker WebSocket URL | `wss://clearnet.yellow.com/ws` | No |
| `ASSET` | Asset type for transactions | `usdc` | No |
| `WALLET_PRIVATE_KEY` | Server wallet private key | - | Yes (production) |
| `VAPP_NAME` | Virtual app name | `{{projectName}}` | No |
| `VAPP_SCOPE` | Virtual app scope | `{{packageName}}` | No |

## Scripts

- `npm run dev` - Start development server with hot reload
- `npm run build` - Build for production
- `npm start` - Start production server
- `npm run lint` - Run ESLint

## Nitrolite Integration

This server integrates with the Nitrolite SDK to provide state channel functionality:

- **RPC Client**: Connects to the Yellow network broker
- **Authentication**: EIP-712 wallet authentication
- **State Channels**: Support for creating and managing app sessions
- **Real-time Updates**: WebSocket integration for real-time state updates

## Development Tips

1. **Environment Setup**: Always use the development environment for testing
2. **Wallet Security**: Never commit your private key to version control
3. **Error Handling**: Check the logs for detailed error information
4. **Authentication**: Test wallet authentication flows before production
5. **State Management**: Understand the state channel lifecycle

## Deployment

### Docker

Create a `Dockerfile`:

```dockerfile
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY dist/ ./dist/
EXPOSE 3001
CMD ["node", "dist/server.js"]
```

### Environment Variables in Production

Ensure these are set in production:

- `NODE_ENV=production`
- `WALLET_PRIVATE_KEY` (secure wallet with appropriate permissions)
- Other configuration as needed

## License

ISC License - see the main Nitrolite repository for details.

## Support

- [Nitrolite Documentation](https://github.com/erc7824/nitrolite)
- [Issue Tracker](https://github.com/erc7824/nitrolite/issues)

---

Generated with [create-nitrolite-app](https://github.com/erc7824/nitrolite/tree/main/sdk/create-nitrolite-app)