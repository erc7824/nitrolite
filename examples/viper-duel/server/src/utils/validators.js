/**
 * Validators for game inputs
 */

/**
 * Validates Ethereum address format
 * @param {string} address - Ethereum address to validate
 * @returns {boolean} True if the address is valid
 */
export function isValidEthereumAddress(address) {
  // Check if it's a string and matches Ethereum address pattern (0x followed by 40 hex chars)
  return typeof address === 'string' 
    && /^0x[a-fA-F0-9]{40}$/.test(address);
}

/**
 * Validates room ID format
 * @param {string} roomId - Room ID to validate
 * @returns {boolean} True if the room ID is valid
 */
export function isValidRoomId(roomId) {
  // Basic UUID v4 format check
  return typeof roomId === 'string'
    && /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(roomId);
}

/**
 * Validates snake direction format
 * @param {string} direction - Direction ('UP', 'DOWN', 'LEFT', 'RIGHT')
 * @returns {boolean} True if the direction is valid
 */
export function isValidDirection(direction) {
  return typeof direction === 'string' && ['UP', 'DOWN', 'LEFT', 'RIGHT'].includes(direction);
}

/**
 * Validates bet amount
 * @param {number} betAmount - Bet amount to validate
 * @returns {boolean} True if the bet amount is valid
 */
export function isValidBetAmount(betAmount) {
  const validAmounts = [0, 0.01, 0.1, 1, 2];
  return typeof betAmount === 'number' && validAmounts.includes(betAmount);
}

/**
 * Validates join room payload
 * @param {object} payload - The payload to validate
 * @param {string} payload.roomId - Room ID
 * @param {string} payload.eoa - Ethereum address
 * @param {number} payload.betAmount - Bet amount
 * @returns {object} Validation result with success flag and optional error message
 */
export function validateJoinRoomPayload(payload) {
  if (!payload || typeof payload !== 'object') {
    return { success: false, error: 'Invalid payload format' };
  }

  // EOA validation
  if (!payload.eoa) {
    return { success: false, error: 'Ethereum address is required' };
  }

  if (!isValidEthereumAddress(payload.eoa)) {
    return { success: false, error: 'Invalid Ethereum address format' };
  }

  // Bet amount validation
  if (payload.betAmount !== undefined && !isValidBetAmount(payload.betAmount)) {
    return { success: false, error: 'Invalid bet amount. Must be 0, 0.01, 0.1, 1, or 2' };
  }

  // Room ID validation
  // If roomId is undefined, we're creating a new room
  // If roomId is provided, we're joining an existing room
  if (payload.roomId === undefined) {
    // Creating a new room - no further validation needed
    console.log("Creating new room with bet amount:", payload.betAmount || 0);
    return { success: true, isCreating: true };
  } else {
    // Joining a room - validate room ID 
    if (!isValidRoomId(payload.roomId)) {
      return { success: false, error: 'Invalid room ID format' };
    }
    console.log("Joining existing room:", payload.roomId, "with bet amount:", payload.betAmount || 0);
    return { success: true, isJoining: true };
  }
}

/**
 * Validates direction change payload
 * @param {object} payload - The payload to validate
 * @param {string} payload.roomId - Room ID
 * @param {string} payload.direction - Direction ('UP', 'DOWN', 'LEFT', 'RIGHT')
 * @returns {object} Validation result with success flag and optional error message
 */
export function validateDirectionPayload(payload) {
  if (!payload || typeof payload !== 'object') {
    return { success: false, error: 'Invalid payload format' };
  }

  if (!payload.roomId) {
    return { success: false, error: 'Room ID is required' };
  }

  if (!isValidRoomId(payload.roomId)) {
    return { success: false, error: 'Invalid room ID format' };
  }

  if (!payload.direction) {
    return { success: false, error: 'Direction is required' };
  }

  if (!isValidDirection(payload.direction)) {
    return { success: false, error: 'Invalid direction format (must be UP, DOWN, LEFT, or RIGHT)' };
  }

  return { success: true };
}