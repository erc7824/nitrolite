# Viper Duel Server

A WebSocket server for the Viper Duel 1-vs-1 Snake game. This server handles game rooms, player connections, and game state management.

## Quick Start

```bash
# Install dependencies
npm install

# Start the server (production)
npm start

# Start the server with hot reload (development)
npm run dev

# Run linting
npm run lint
```

Server runs on port 8080 by default.

## Protocol

### Client → Server Messages

```js
// Join a room
{
  "type": "joinRoom",
  "payload": {
    "roomId": "uuid-string",  // Optional - system will create if not provided
    "eoa": "0x..."            // Ethereum address
  }
}

// Make a move
{
  "type": "move",
  "payload": {
    "roomId": "uuid-string",  // Required
    "pos": 0                  // Position on board (0-8)
  }
}
```

### Server → Client Messages

```js
// Room state update (after join or move)
{
  "type": "room:state",
  "roomId": "uuid-string",
  "board": [null,null,null,null,"X",null,null,null,null],
  "nextTurn": "O",
  "players": {
    "X": "0x...", // Host EOA
    "O": "0x..."  // Guest EOA
  }
}

// Room ready notification (2 players joined)
{
  "type": "room:ready",
  "roomId": "uuid-string"
}

// Game over notification
{
  "type": "game:over",
  "winner": "X",  // "X", "O", or null (draw)
  "board": ["X","O","X","X","O","O","X",null,null]
}

// Error message
{
  "type": "error",
  "code": "ERROR_CODE",
  "msg": "Error description"
}
```

## Sequence Diagram

```
┌──────┐                   ┌─────────┐                   ┌──────┐
│Client│                   │  Server │                   │Client│
└──┬───┘                   └────┬────┘                   └──┬───┘
   │                            │                           │
   │ {"type":"joinRoom",...}    │                           │
   ├────────────────────────────►                           │
   │                            │                           │
   │ {"type":"room:state",...}  │                           │
   ◄├───────────────────────────┤                           │
   │                            │                           │
   │                            │ {"type":"joinRoom",...}   │
   │                            ◄───────────────────────────┤
   │                            │                           │
   │   {"type":"room:state",...}│                           │
   ◄├───────────────────────────┼───────────────────────────►
   │                            │                           │
   │   {"type":"room:ready",...}│                           │
   ◄├───────────────────────────┼───────────────────────────►
   │                            │                           │
   │ {"type":"move",...}        │                           │
   ├────────────────────────────►                           │
   │                            │                           │
   │   {"type":"room:state",...}│                           │
   ◄├───────────────────────────┼───────────────────────────►
   │                            │                           │
   │                            │ {"type":"move",...}       │
   │                            ◄───────────────────────────┤
   │                            │                           │
   │   {"type":"room:state",...}│                           │
   ◄├───────────────────────────┼───────────────────────────►
   │                            │                           │
   │ (moves continue...)        │                           │
   ├────────────────────────────►                           │
   │                            │                           │
   │   {"type":"game:over",...} │                           │
   ◄├───────────────────────────┼───────────────────────────►
   │                            │                           │
```

## Technical Notes

- Uses in-memory storage for rooms and game state (will be replaced with @erc7824/nitrolite in future)
- First player is assigned as host (X), second as guest (O)
- Rooms are automatically created if a non-existent room ID is provided
- Rooms are automatically cleaned up after games complete
- Game rules enforce alternating turns and valid move placement
- Validation for Ethereum addresses and game actions