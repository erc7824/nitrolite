/**
 * Services index - exports all service modules
 */

// Nitrolite RPC (WebSocket) client
export { 
  initializeRPCClient, 
  getRPCClient,
  NitroliteRPCClient, 
  WSStatus 
} from './nitroliteRPC.js';


// App sessions for game rooms
export {
  createAppSession,
  closeAppSession,
  getAppSession,
  hasAppSession,
  getAllAppSessions,
  generateAppSessionMessage,
  getPendingAppSessionMessage,
  addAppSessionSignature,
  createAppSessionWithSignatures
} from './appSessions.js';

// Room management
export { createRoomManager } from './roomManager.js';

// Snake game logic
export { createGame, changeDirection, updateGame, formatGameState, formatGameOverMessage } from './snake.js';