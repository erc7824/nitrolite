import { describe, it, expect, beforeEach } from 'vitest';
import { LVCI } from '../../src/rpc';
import { Address } from 'viem';

describe('LVCI (Light Virtual Channel Identifier)', () => {
  // Test addresses
  const alice = '0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266' as Address; // Hardhat #0
  const bob = '0x70997970C51812dc3A010C7d01b50e0d17dc79C8' as Address; // Hardhat #1
  const charlie = '0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC' as Address; // Hardhat #2
  const denis = '0x90F79bf6EB2c4f870365E785982E1f101E93b906' as Address; // Hardhat #3
  
  // LVCIs for testing
  let directLVCI: ReturnType<typeof LVCI.create>; // Alice -> Bob
  let singleHopLVCI: ReturnType<typeof LVCI.create>; // Alice -> Bob -> Charlie
  let multiHopLVCI: ReturnType<typeof LVCI.create>; // Alice -> Bob -> Charlie -> Denis
  
  beforeEach(() => {
    // Create LVCI instances
    directLVCI = LVCI.create(alice, bob);
    singleHopLVCI = LVCI.create(alice, charlie, [bob]);
    multiHopLVCI = LVCI.create(alice, denis, [bob, charlie]);
  });
  
  describe('create', () => {
    it('should create a direct LVCI', () => {
      expect(directLVCI.origin).toBe(alice);
      expect(directLVCI.destination).toBe(bob);
      expect(directLVCI.intermediaries).toEqual([]);
      expect(directLVCI.nonce).toBe(BigInt(0));
    });
    
    it('should create a single-hop LVCI', () => {
      expect(singleHopLVCI.origin).toBe(alice);
      expect(singleHopLVCI.destination).toBe(charlie);
      expect(singleHopLVCI.intermediaries).toEqual([bob]);
      expect(singleHopLVCI.nonce).toBe(BigInt(0));
    });
    
    it('should create a multi-hop LVCI', () => {
      expect(multiHopLVCI.origin).toBe(alice);
      expect(multiHopLVCI.destination).toBe(denis);
      expect(multiHopLVCI.intermediaries).toEqual([bob, charlie]);
      expect(multiHopLVCI.nonce).toBe(BigInt(0));
    });
    
    it('should create an LVCI with a custom nonce', () => {
      const customNonceLVCI = LVCI.create(alice, bob, [], BigInt(42));
      expect(customNonceLVCI.nonce).toBe(BigInt(42));
    });
  });
  
  describe('getId', () => {
    it('should generate a unique ID for each LVCI', () => {
      const id1 = LVCI.getId(directLVCI);
      const id2 = LVCI.getId(singleHopLVCI);
      const id3 = LVCI.getId(multiHopLVCI);
      
      expect(id1).not.toBe(id2);
      expect(id1).not.toBe(id3);
      expect(id2).not.toBe(id3);
    });
    
    it('should generate the same ID for identical LVCIs', () => {
      const lvci1 = LVCI.create(alice, bob);
      const lvci2 = LVCI.create(alice, bob);
      
      expect(LVCI.getId(lvci1)).toBe(LVCI.getId(lvci2));
    });
    
    it('should generate different IDs for LVCIs with different nonces', () => {
      const lvci1 = LVCI.create(alice, bob, [], BigInt(1));
      const lvci2 = LVCI.create(alice, bob, [], BigInt(2));
      
      expect(LVCI.getId(lvci1)).not.toBe(LVCI.getId(lvci2));
    });
  });
  
  describe('getPath', () => {
    it('should return the correct path for a direct LVCI', () => {
      const path = LVCI.getPath(directLVCI);
      expect(path).toEqual([alice, bob]);
    });
    
    it('should return the correct path for a single-hop LVCI', () => {
      const path = LVCI.getPath(singleHopLVCI);
      expect(path).toEqual([alice, bob, charlie]);
    });
    
    it('should return the correct path for a multi-hop LVCI', () => {
      const path = LVCI.getPath(multiHopLVCI);
      expect(path).toEqual([alice, bob, charlie, denis]);
    });
  });
  
  describe('isParticipant', () => {
    it('should correctly identify participants', () => {
      // Direct LVCI
      expect(LVCI.isParticipant(directLVCI, alice)).toBe(true);
      expect(LVCI.isParticipant(directLVCI, bob)).toBe(true);
      expect(LVCI.isParticipant(directLVCI, charlie)).toBe(false);
      
      // Single-hop LVCI
      expect(LVCI.isParticipant(singleHopLVCI, alice)).toBe(true);
      expect(LVCI.isParticipant(singleHopLVCI, bob)).toBe(true);
      expect(LVCI.isParticipant(singleHopLVCI, charlie)).toBe(true);
      expect(LVCI.isParticipant(singleHopLVCI, denis)).toBe(false);
      
      // Multi-hop LVCI
      expect(LVCI.isParticipant(multiHopLVCI, alice)).toBe(true);
      expect(LVCI.isParticipant(multiHopLVCI, bob)).toBe(true);
      expect(LVCI.isParticipant(multiHopLVCI, charlie)).toBe(true);
      expect(LVCI.isParticipant(multiHopLVCI, denis)).toBe(true);
    });
  });
  
  describe('getNextHop', () => {
    it('should return the next hop in the forward direction', () => {
      // Direct LVCI
      expect(LVCI.getNextHop(directLVCI, alice)).toBe(bob);
      expect(LVCI.getNextHop(directLVCI, bob)).toBe(null);
      
      // Single-hop LVCI
      expect(LVCI.getNextHop(singleHopLVCI, alice)).toBe(bob);
      expect(LVCI.getNextHop(singleHopLVCI, bob)).toBe(charlie);
      expect(LVCI.getNextHop(singleHopLVCI, charlie)).toBe(null);
      
      // Multi-hop LVCI
      expect(LVCI.getNextHop(multiHopLVCI, alice)).toBe(bob);
      expect(LVCI.getNextHop(multiHopLVCI, bob)).toBe(charlie);
      expect(LVCI.getNextHop(multiHopLVCI, charlie)).toBe(denis);
      expect(LVCI.getNextHop(multiHopLVCI, denis)).toBe(null);
    });
    
    it('should return the next hop in the backward direction', () => {
      // Direct LVCI
      expect(LVCI.getNextHop(directLVCI, alice, false)).toBe(null);
      expect(LVCI.getNextHop(directLVCI, bob, false)).toBe(alice);
      
      // Single-hop LVCI
      expect(LVCI.getNextHop(singleHopLVCI, alice, false)).toBe(null);
      expect(LVCI.getNextHop(singleHopLVCI, bob, false)).toBe(alice);
      expect(LVCI.getNextHop(singleHopLVCI, charlie, false)).toBe(bob);
      
      // Multi-hop LVCI
      expect(LVCI.getNextHop(multiHopLVCI, alice, false)).toBe(null);
      expect(LVCI.getNextHop(multiHopLVCI, bob, false)).toBe(alice);
      expect(LVCI.getNextHop(multiHopLVCI, charlie, false)).toBe(bob);
      expect(LVCI.getNextHop(multiHopLVCI, denis, false)).toBe(charlie);
    });
    
    it('should return null for non-participants', () => {
      expect(LVCI.getNextHop(directLVCI, charlie)).toBe(null);
      expect(LVCI.getNextHop(singleHopLVCI, denis)).toBe(null);
    });
  });
  
  describe('getPosition', () => {
    it('should return the correct position in the path', () => {
      // Direct LVCI
      expect(LVCI.getPosition(directLVCI, alice)).toBe(0);
      expect(LVCI.getPosition(directLVCI, bob)).toBe(1);
      expect(LVCI.getPosition(directLVCI, charlie)).toBe(-1);
      
      // Single-hop LVCI
      expect(LVCI.getPosition(singleHopLVCI, alice)).toBe(0);
      expect(LVCI.getPosition(singleHopLVCI, bob)).toBe(1);
      expect(LVCI.getPosition(singleHopLVCI, charlie)).toBe(2);
      expect(LVCI.getPosition(singleHopLVCI, denis)).toBe(-1);
      
      // Multi-hop LVCI
      expect(LVCI.getPosition(multiHopLVCI, alice)).toBe(0);
      expect(LVCI.getPosition(multiHopLVCI, bob)).toBe(1);
      expect(LVCI.getPosition(multiHopLVCI, charlie)).toBe(2);
      expect(LVCI.getPosition(multiHopLVCI, denis)).toBe(3);
    });
  });
  
  describe('toString', () => {
    it('should return a formatted string representation', () => {
      expect(LVCI.toString(directLVCI)).toBe(`${alice}>${bob}#0`);
      expect(LVCI.toString(singleHopLVCI)).toBe(`${alice}>${bob}>${charlie}#0`);
      expect(LVCI.toString(multiHopLVCI)).toBe(`${alice}>${bob}>${charlie}>${denis}#0`);
      
      const customNonceLVCI = LVCI.create(alice, bob, [], BigInt(42));
      expect(LVCI.toString(customNonceLVCI)).toBe(`${alice}>${bob}#42`);
    });
  });
  
  describe('createSubPath', () => {
    it('should create a sub-path between two participants', () => {
      // Create a sub-path from Alice to Charlie in the multi-hop LVCI
      const subPath = LVCI.createSubPath(multiHopLVCI, alice, charlie);
      
      expect(subPath).not.toBeNull();
      if (subPath) {
        expect(subPath.origin).toBe(alice);
        expect(subPath.destination).toBe(charlie);
        expect(subPath.intermediaries).toEqual([bob]);
        expect(subPath.nonce).toBe(multiHopLVCI.nonce);
      }
    });
    
    it('should create a sub-path between two intermediaries', () => {
      // Create a sub-path from Bob to Denis in the multi-hop LVCI
      const subPath = LVCI.createSubPath(multiHopLVCI, bob, denis);
      
      expect(subPath).not.toBeNull();
      if (subPath) {
        expect(subPath.origin).toBe(bob);
        expect(subPath.destination).toBe(denis);
        expect(subPath.intermediaries).toEqual([charlie]);
        expect(subPath.nonce).toBe(multiHopLVCI.nonce);
      }
    });
    
    it('should return null for invalid sub-paths', () => {
      // Trying to create a sub-path in the wrong direction
      expect(LVCI.createSubPath(multiHopLVCI, charlie, bob)).toBeNull();
      
      // Trying to create a sub-path with a non-participant
      expect(LVCI.createSubPath(directLVCI, alice, charlie)).toBeNull();
      
      // Trying to create a sub-path from a participant to itself
      expect(LVCI.createSubPath(multiHopLVCI, alice, alice)).toBeNull();
    });
  });
});