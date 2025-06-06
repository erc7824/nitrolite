import { Express, RequestHandler } from 'express';
import { ethers } from 'ethers';
import { randomBytes } from 'crypto';
import { SERVER_PRIVATE_KEY } from '../config/index.ts';
import { getRoom, getAllRooms } from '../services/stateService.ts';
import { clearNetRPC } from '../services/gameService.ts';
import {
  signStateData,
  verifySignature,
  isAuthenticatedWithBroker
} from '../services/brokerService.ts';
import {
  generateChallenge,
  verifyChallengeSignature
} from '../middlewares/authMiddleware.ts';
import { Room } from '../interfaces/index.ts';
import { Hex } from 'viem';

// Setup API routes for the Express app
export function setupApiRoutes(app: Express): void {
  // Get active rooms
  app.get('/api/rooms', ((_, res) => {
    const activeRooms = Array.from(getAllRooms().entries())
      .filter(([_, room]) => room.players.size < 2) // Only return rooms that aren't full
      .map(([id, room]) => ({
        id,
        playerCount: room.players.size,
        createdAt: room.createdAt
      }));

    res.json(activeRooms);
  }) as RequestHandler);

  // Get room details
  app.get('/api/rooms/:roomId', ((req, res) => {
    const { roomId } = req.params;
    const room = getRoom(roomId);

    if (!room) {
      return res.status(404).json({ error: 'Room not found' });
    }

    res.json({
      id: roomId,
      playerCount: room.players.size,
      isGameOver: room.isGameOver || false,
      channelIds: Array.from(room.channelIds),
      createdAt: room.createdAt
    });
  }) as RequestHandler);

  // Room channel endpoint
  app.get('/api/rooms/:roomId/channel', ((req, res) => {
    const roomId = req.params.roomId;
    const room = getRoom(roomId);

    if (!room) {
      return res.status(404).json({ error: 'Room not found' });
    }

    // Get the first channelId in the room
    const channelIds = Array.from(room.channelIds);
    if (channelIds.length === 0) {
      return res.status(404).json({ error: 'No channel found for this room' });
    }

    res.json({
      channelId: channelIds[0],
      roomId
    });
  }) as RequestHandler);

  // Game sessions endpoint
  app.post('/api/game-sessions', (async (req, res) => {
    const { channelId } = req.body;

    if (!channelId) {
      return res.status(400).json({ error: 'Channel ID is required' });
    }

    try {
      // Check if channel exists in any room
      let foundRoom: Room | null = null;
      for (const room of getAllRooms().values()) {
        if (room.channelIds.has(channelId)) {
          foundRoom = room;
          break;
        }
      }

      if (!foundRoom) {
        return res.status(404).json({ error: 'Channel not found in any active room' });
      }

      // Create a session ID
      const sessionId = `session_${randomBytes(8).toString('hex')}`;

      // Get channel info from ClearNet RPC
      const channelInfo = await clearNetRPC.getChannelInfo(channelId);

      res.json({
        sessionId,
        channelId,
        channelInfo
      });
    } catch (error) {
      console.error('Error creating game session:', error);
      res.status(500).json({ error: 'Failed to create game session' });
    }
  }) as RequestHandler);

  // Get game session channel state
  app.get('/api/game-sessions/:sessionId/state', (async (req, res) => {
    const { sessionId } = req.params;

    try {
      const channelId = req.query.channelId as Hex;

      if (!channelId) {
        return res.status(400).json({ error: 'Channel ID is required as a query parameter' });
      }

      // Get channel info from ClearNet RPC
      const channelInfo = await clearNetRPC.getChannelInfo(channelId);

      if (!channelInfo) {
        return res.status(404).json({ error: 'Channel not found' });
      }

      // Find the associated room
      let foundRoom: Room | null = null;
      for (const room of getAllRooms().values()) {
        if (room.channelIds.has(channelId)) {
          foundRoom = room;
          break;
        }
      }

      if (!foundRoom) {
        return res.status(404).json({ error: 'Room not found for this channel' });
      }

      // Return the current state
      res.json({
        sessionId,
        channelId,
        roomId: foundRoom.id,
        state: foundRoom.currentState,
        isGameOver: foundRoom.isGameOver || false,
        stateVersion: foundRoom.stateVersion,
        players: Array.from(foundRoom.players.values()).map(p => ({
          id: p.id,
          nickname: p.nickname,
          score: p.score,
          isDead: p.isDead || false
        }))
      });
    } catch (error) {
      console.error('Error getting game session state:', error);
      res.status(500).json({ error: 'Failed to get game session state' });
    }
  }) as RequestHandler);

  // Authentication status endpoint
  app.get('/api/auth-status', ((_, res) => {
    res.json({
      authenticated: isAuthenticatedWithBroker(),
      serverAddress: new ethers.Wallet(SERVER_PRIVATE_KEY).address
    });
  }) as RequestHandler);

  // Sign state endpoint
  app.post('/api/sign-state', (async (req, res) => {
    const { stateData } = req.body;

    if (!stateData) {
      return res.status(400).json({ error: 'State data is required' });
    }

    try {
      const result = await signStateData(typeof stateData === 'string' ? stateData : JSON.stringify(stateData));
      res.json({
        success: true,
        signature: result.signature,
        signerAddress: result.address
      });
    } catch (error) {
      console.error('Error signing state:', error);
      res.status(500).json({ error: 'Failed to sign state data' });
    }
  }) as RequestHandler);

  // Verify signature endpoint
  app.post('/api/verify-signature', ((req, res) => {
    const { message, signature, address } = req.body;

    if (!message || !signature || !address) {
      return res.status(400).json({
        error: 'Message, signature, and address are required',
        received: { message: !!message, signature: !!signature, address: !!address }
      });
    }

    try {
      const isValid = verifySignature(message, signature, address);
      res.json({
        success: true,
        valid: isValid
      });
    } catch (error) {
      console.error('Error verifying signature:', error);
      res.status(500).json({ error: 'Failed to verify signature' });
    }
  }) as RequestHandler);

  // Generate challenge for client authentication
  app.post('/api/auth/challenge', ((req, res) => {
    const { address } = req.body;

    if (!address || !ethers.isAddress(address)) {
      return res.status(400).json({ error: 'Valid Ethereum address is required' });
    }

    try {
      // Use the middleware function to generate and store a challenge
      const { challenge, timestamp } = generateChallenge(address);

      res.json({
        success: true,
        challenge,
        timestamp
      });
    } catch (error) {
      console.error('Error generating challenge:', error);
      res.status(500).json({ error: 'Failed to generate authentication challenge' });
    }
  }) as RequestHandler);

  // Verify client auth response
  app.post('/api/auth/verify', ((req, res) => {
    const { address, signature } = req.body;

    if (!address || !signature) {
      return res.status(400).json({
        error: 'Address and signature are required'
      });
    }

    try {
      // Use the middleware function to verify the signature against the stored challenge
      const isValid = verifyChallengeSignature(address, signature);

      if (isValid) {
        // In a real implementation, you would issue a JWT token here
        // For demo, we'll create a simple token as address:signature
        const token = `${address}:${signature}`;

        res.json({
          success: true,
          authenticated: true,
          address,
          token
        });
      } else {
        res.status(401).json({
          error: 'Invalid signature or expired challenge',
          authenticated: false
        });
      }
    } catch (error) {
      console.error('Error verifying authentication:', error);
      res.status(500).json({ error: 'Failed to verify authentication' });
    }
  }) as RequestHandler);
}
