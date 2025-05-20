import { WebSocket } from 'ws';
import { Room, PendingRequest } from '../interfaces';

// Global state
export const rooms = new Map<string, Room>();
let brokerWs: WebSocket | null = null;
export const pendingRequests = new Map<string, PendingRequest>();

// Set broker WebSocket connection
export function setBrokerWebSocket(ws: WebSocket) {
  brokerWs = ws;
}

// Get broker WebSocket connection
export function getBrokerWebSocket(): WebSocket | null {
  return brokerWs;
}

// Add a room
export function addRoom(roomId: string, room: Room): void {
  console.log(`[stateService] Adding room ${roomId}`);
  rooms.set(roomId, room);
  console.log(`[stateService] Current rooms:`, Array.from(rooms.keys()));
}

// Get a room
export function getRoom(roomId: string): Room | undefined {
  console.log(`[stateService] Getting room ${roomId}`);
  console.log(`[stateService] Available rooms:`, Array.from(rooms.keys()));
  return rooms.get(roomId);
}

// Remove a room
export function removeRoom(roomId: string): boolean {
  return rooms.delete(roomId);
}

// Get all rooms
export function getAllRooms(): Map<string, Room> {
  return rooms;
}

// Add a pending request
export function addPendingRequest(
  requestId: string,
  resolve: Function,
  reject: Function,
  timeout: NodeJS.Timeout
): void {
  pendingRequests.set(requestId, { resolve, reject, timeout });
}

// Get and remove a pending request
export function getPendingRequest(requestId: string): PendingRequest | undefined {
  const request = pendingRequests.get(requestId);
  pendingRequests.delete(requestId);
  return request;
}

// Clear a pending request
export function clearPendingRequest(requestId: string): void {
  if (pendingRequests.has(requestId)) {
    const { timeout } = pendingRequests.get(requestId)!;
    clearTimeout(timeout);
    pendingRequests.delete(requestId);
  }
}
