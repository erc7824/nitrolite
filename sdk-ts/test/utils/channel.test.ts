import { getChannelId, getStateHash } from '../../src/utils';
import { Channel, State, Allocation } from '../../src/types';

describe('Channel Utilities', () => {
  // Test data
  const testChannel: Channel = {
    participants: ['0x1111111111111111111111111111111111111111', '0x2222222222222222222222222222222222222222'],
    adjudicator: '0x3333333333333333333333333333333333333333',
    challenge: BigInt(86400),
    nonce: BigInt(123456)
  };
  
  const testAllocations: [Allocation, Allocation] = [
    {
      destination: '0x1111111111111111111111111111111111111111',
      token: '0x4444444444444444444444444444444444444444',
      amount: BigInt(100)
    },
    {
      destination: '0x2222222222222222222222222222222222222222',
      token: '0x4444444444444444444444444444444444444444',
      amount: BigInt(200)
    }
  ];
  
  const testState: State = {
    data: '0x0123456789abcdef',
    allocations: testAllocations,
    sigs: []
  };
  
  describe('getChannelId', () => {
    it('should generate a deterministic channel ID', () => {
      const channelId = getChannelId(testChannel);
      
      // Channel ID should be a hex string
      expect(channelId).toMatch(/^0x[0-9a-f]{64}$/i);
      
      // Same channel config should give same ID
      const duplicateId = getChannelId(testChannel);
      expect(duplicateId).toEqual(channelId);
    });
    
    it('should generate different IDs for different channels', () => {
      const channelId = getChannelId(testChannel);
      
      // Different nonce should give different ID
      const differentChannel: Channel = {
        ...testChannel,
        nonce: BigInt(654321)
      };
      const differentId = getChannelId(differentChannel);
      expect(differentId).not.toEqual(channelId);
      
      // Different participants should give different ID
      const differentParticipants: Channel = {
        ...testChannel,
        participants: ['0x5555555555555555555555555555555555555555', '0x6666666666666666666666666666666666666666']
      };
      const differentParticipantsId = getChannelId(differentParticipants);
      expect(differentParticipantsId).not.toEqual(channelId);
    });
  });
  
  describe('getStateHash', () => {
    it('should generate a deterministic state hash', () => {
      const stateHash = getStateHash(testChannel, testState);
      
      // State hash should be a hex string
      expect(stateHash).toMatch(/^0x[0-9a-f]{64}$/i);
      
      // Same state should give same hash
      const duplicateHash = getStateHash(testChannel, testState);
      expect(duplicateHash).toEqual(stateHash);
    });
    
    it('should generate different hashes for different states', () => {
      const stateHash = getStateHash(testChannel, testState);
      
      // Different data should give different hash
      const differentData: State = {
        ...testState,
        data: '0xabcdef0123456789'
      };
      const differentDataHash = getStateHash(testChannel, differentData);
      expect(differentDataHash).not.toEqual(stateHash);
      
      // Different allocations should give different hash
      const differentAllocations: State = {
        ...testState,
        allocations: [
          {
            destination: '0x1111111111111111111111111111111111111111',
            token: '0x4444444444444444444444444444444444444444',
            amount: BigInt(300) // different amount
          },
          {
            destination: '0x2222222222222222222222222222222222222222',
            token: '0x4444444444444444444444444444444444444444',
            amount: BigInt(200)
          }
        ]
      };
      const differentAllocationsHash = getStateHash(testChannel, differentAllocations);
      expect(differentAllocationsHash).not.toEqual(stateHash);
    });
    
    it('should generate different hashes for different channels with same state', () => {
      const stateHash = getStateHash(testChannel, testState);
      
      // Different channel should give different hash even with same state
      const differentChannel: Channel = {
        ...testChannel,
        nonce: BigInt(654321)
      };
      const differentChannelHash = getStateHash(differentChannel, testState);
      expect(differentChannelHash).not.toEqual(stateHash);
    });
  });
});