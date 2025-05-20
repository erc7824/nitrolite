# Tic Tac Toe with Nitrolite

This example demonstrates a multiplayer Tic Tac Toe game using Nitrolite for state channels.

## Structure

- `/client` - React frontend application
- `/server` - WebSocket server for game coordination

## Prerequisites

- Node.js >= 20.0.0
- npm or yarn

## Quick Start

1. Install dependencies and start the server:

```bash
# Start the server
cd server
npm install
npm run dev
```

2. In a new terminal, start the client:

```bash
# Start the client
cd client
npm install
npm run dev
```

3. Open your browser to the URL shown in the client terminal (usually http://localhost:5173)

## Development

- Client development server: `npm run dev` in the `/client` directory
- Server development with auto-restart: `npm run dev` in the `/server` directory

## Building for Production

```bash
# Build the client
cd client
npm run build

# Start the server in production
cd ../server
npm start
```

See individual README files in the `/client` and `/server` directories for more details.