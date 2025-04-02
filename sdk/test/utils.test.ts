// This file is kept for backward compatibility
// The tests have been reorganized into a directory structure
// See the test directory for the full test suite

import { getChannelId } from '../src/utils';
import { Channel } from '../src/types';

describe('Legacy Utils Tests', () => {
  describe('getChannelId', () => {
    it('should generate a deterministic channel ID', () => {
      const channel: Channel = {
        participants: ['0x1111111111111111111111111111111111111111', '0x2222222222222222222222222222222222222222'],
        adjudicator: '0x3333333333333333333333333333333333333333',
        challenge: BigInt(86400),
        nonce: BigInt(123456)
      };
      
      const channelId = getChannelId(channel);
      
      // Channel ID should be a hex string
      expect(channelId).toMatch(/^0x[0-9a-f]{64}$/i);
      
      // Same channel config should give same ID
      const duplicateId = getChannelId(channel);
      expect(duplicateId).toEqual(channelId);
      
      // Different nonce should give different ID
      const differentChannel: Channel = {
        ...channel,
        nonce: BigInt(654321)
      };
      const differentId = getChannelId(differentChannel);
      expect(differentId).not.toEqual(channelId);
    });
  });
});