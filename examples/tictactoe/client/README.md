# Tic Tac Toe Client

React-based frontend for the Tic Tac Toe game with Nitrolite integration.

## Setup

```bash
# Install dependencies
npm install
```

## Development

```bash
# Start development server
npm run dev
```

The application will be available at http://localhost:5173 (or another port if 5173 is in use).

## Build

```bash
# Build for production
npm run build

# Preview production build
npm run preview
```

The production build will be generated in the `dist` directory.

## Environment Setup

Create a `.env` file in the client directory with the following variables if needed:

```
VITE_WS_URL=ws://localhost:3000
```

Adjust the WebSocket URL according to your server configuration.