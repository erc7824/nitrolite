import { Address, Hex, keccak256, encodeAbiParameters, toHex } from 'viem';
import { LightVirtualChannelIdentifier } from './types';

/**
 * LVCI Identifier Type
 */
export type LVCIId = Hex;

/**
 * Class for managing Light Virtual Channel Identifiers
 */
export class LVCI {
  /**
   * Create a new LVCI
   * @param origin The address of the origin participant
   * @param destination The address of the final destination participant
   * @param intermediaries Array of intermediary addresses (empty for direct channels)
   * @param nonce Unique nonce for this channel
   * @returns A Light Virtual Channel Identifier
   */
  static create(origin: Address, destination: Address, intermediaries: Address[] = [], nonce: bigint = BigInt(0)): LightVirtualChannelIdentifier {
    return {
      origin,
      destination,
      intermediaries,
      nonce
    };
  }

  /**
   * Generate a unique identifier for an LVCI
   * @param lvci The LVCI to get ID for
   * @returns A unique hash representing this LVCI
   */
  static getId(lvci: LightVirtualChannelIdentifier): LVCIId {
    // Concatenate all addresses and nonce to create a unique identifier
    const encoded = encodeAbiParameters(
      [
        { name: 'origin', type: 'address' },
        { name: 'destination', type: 'address' },
        { name: 'intermediaries', type: 'address[]' },
        { name: 'nonce', type: 'uint256' }
      ],
      [
        lvci.origin,
        lvci.destination,
        lvci.intermediaries,
        lvci.nonce
      ]
    );
    
    return keccak256(encoded);
  }

  /**
   * Get the routing path as an array of addresses
   * @param lvci The LVCI to get the path for
   * @returns Array of addresses in the routing path
   */
  static getPath(lvci: LightVirtualChannelIdentifier): Address[] {
    return [lvci.origin, ...lvci.intermediaries, lvci.destination];
  }

  /**
   * Check if an address is part of the LVCI path
   * @param lvci The LVCI to check
   * @param address The address to check
   * @returns True if the address is part of the path
   */
  static isParticipant(lvci: LightVirtualChannelIdentifier, address: Address): boolean {
    if (lvci.origin.toLowerCase() === address.toLowerCase()) return true;
    if (lvci.destination.toLowerCase() === address.toLowerCase()) return true;
    return lvci.intermediaries.some(intermediary => intermediary.toLowerCase() === address.toLowerCase());
  }

  /**
   * Get the next hop in the path from a given address
   * @param lvci The LVCI to check
   * @param from The current address
   * @param direction The direction (forward = true, backward = false)
   * @returns The next address in the path or null if at the end
   */
  static getNextHop(lvci: LightVirtualChannelIdentifier, from: Address, direction: boolean = true): Address | null {
    const path = this.getPath(lvci);
    const index = path.findIndex(addr => addr.toLowerCase() === from.toLowerCase());
    
    if (index === -1) return null;
    
    const nextIndex = direction ? index + 1 : index - 1;
    return nextIndex >= 0 && nextIndex < path.length ? path[nextIndex] : null;
  }

  /**
   * Get the position in the path (origin = 0, intermediary = 1...n-1, destination = n)
   * @param lvci The LVCI to check
   * @param address The address to get position for
   * @returns The position in the path or -1 if not found
   */
  static getPosition(lvci: LightVirtualChannelIdentifier, address: Address): number {
    const path = this.getPath(lvci);
    return path.findIndex(addr => addr.toLowerCase() === address.toLowerCase());
  }

  /**
   * Serialize LVCI to string format
   * @param lvci The LVCI to serialize
   * @returns String representation
   */
  static toString(lvci: LightVirtualChannelIdentifier): string {
    const path = this.getPath(lvci).join('>');
    return `${path}#${lvci.nonce.toString()}`;
  }

  /**
   * Create a sub-path LVCI between two participants
   * @param lvci The original LVCI
   * @param from Starting address
   * @param to Ending address
   * @returns A new LVCI representing the sub-path or null if invalid
   */
  static createSubPath(lvci: LightVirtualChannelIdentifier, from: Address, to: Address): LightVirtualChannelIdentifier | null {
    const path = this.getPath(lvci);
    const fromIndex = path.findIndex(addr => addr.toLowerCase() === from.toLowerCase());
    const toIndex = path.findIndex(addr => addr.toLowerCase() === to.toLowerCase());
    
    // Make sure both addresses exist and from comes before to
    if (fromIndex === -1 || toIndex === -1 || fromIndex >= toIndex) {
      return null;
    }
    
    // Extract the sub-path
    const subPath = path.slice(fromIndex + 1, toIndex);
    
    return {
      origin: from,
      destination: to,
      intermediaries: subPath,
      nonce: lvci.nonce
    };
  }
}
