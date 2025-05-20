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
 * Validates move position format
 * @param {number} pos - Position on the board (0-8)
 * @returns {boolean} True if the position is valid
 */
export function isValidPosition(pos) {
  return Number.isInteger(pos) && pos >= 0 && pos <= 8;
}

/**
 * Validates join room payload
 * @param {object} payload - The payload to validate
 * @param {string} payload.roomId - Room ID
 * @param {string} payload.eoa - Ethereum address
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

  // Room ID validation
  // If roomId is undefined, we're creating a new room
  // If roomId is provided, we're joining an existing room
  if (payload.roomId === undefined) {
    // Creating a new room - no further validation needed
    console.log("Creating new room");
    return { success: true, isCreating: true };
  } else {
    // Joining a room - validate room ID 
    if (!isValidRoomId(payload.roomId)) {
      return { success: false, error: 'Invalid room ID format' };
    }
    console.log("Joining existing room:", payload.roomId);
    return { success: true, isJoining: true };
  }
}

/**
 * Validates move payload
 * @param {object} payload - The payload to validate
 * @param {string} payload.roomId - Room ID
 * @param {number} payload.pos - Position on the board (0-8)
 * @returns {object} Validation result with success flag and optional error message
 */
export function validateMovePayload(payload) {
  if (!payload || typeof payload !== 'object') {
    return { success: false, error: 'Invalid payload format' };
  }

  if (!payload.roomId) {
    return { success: false, error: 'Room ID is required' };
  }

  if (!isValidRoomId(payload.roomId)) {
    return { success: false, error: 'Invalid room ID format' };
  }

  if (payload.pos === undefined) {
    return { success: false, error: 'Position is required' };
  }

  if (!isValidPosition(payload.pos)) {
    return { success: false, error: 'Invalid position format (must be 0-8)' };
  }

  return { success: true };
}