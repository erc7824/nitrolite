import {
  Channel,
  State,
  Allocation,
  Signature,
  Metadata,
  AdjudicatorStatus,
  Role,
  AppLogic,
  AppConfig,
  AppDataTypes
} from '../../src/types';

describe('Types', () => {
  describe('Channel', () => {
    it('should allow creating a valid channel object', () => {
      const channel: Channel = {
        participants: [
          '0x1111111111111111111111111111111111111111',
          '0x2222222222222222222222222222222222222222'
        ],
        adjudicator: '0x3333333333333333333333333333333333333333',
        challenge: BigInt(86400),
        nonce: BigInt(123456)
      };
      
      expect(channel.participants.length).toBe(2);
      expect(channel.adjudicator).toBe('0x3333333333333333333333333333333333333333');
      expect(channel.challenge).toBe(BigInt(86400));
      expect(channel.nonce).toBe(BigInt(123456));
    });
  });
  
  describe('State', () => {
    it('should allow creating a valid state object', () => {
      const allocations: [Allocation, Allocation] = [
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
      
      const signatures: Signature[] = [
        {
          v: 27,
          r: '0x1234123412341234123412341234123412341234123412341234123412341234',
          s: '0x5678567856785678567856785678567856785678567856785678567856785678'
        }
      ];
      
      const state: State = {
        data: '0x1234',
        allocations,
        sigs: signatures
      };
      
      expect(state.data).toBe('0x1234');
      expect(state.allocations).toBe(allocations);
      expect(state.sigs).toBe(signatures);
    });
  });
  
  describe('Metadata', () => {
    it('should allow creating valid metadata', () => {
      const channel: Channel = {
        participants: [
          '0x1111111111111111111111111111111111111111',
          '0x2222222222222222222222222222222222222222'
        ],
        adjudicator: '0x3333333333333333333333333333333333333333',
        challenge: BigInt(86400),
        nonce: BigInt(123456)
      };
      
      const state: State = {
        data: '0x1234',
        allocations: [
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
        ],
        sigs: []
      };
      
      const metadata: Metadata = {
        chan: channel,
        challengeExpire: BigInt(0),
        lastValidState: state
      };
      
      expect(metadata.chan).toBe(channel);
      expect(metadata.challengeExpire).toBe(BigInt(0));
      expect(metadata.lastValidState).toBe(state);
    });
  });
  
  describe('AppLogic', () => {
    it('should allow creating valid app logic', () => {
      const appLogic: AppLogic<{ value: number }> = {
        encode: (data) => '0x' + data.value.toString(16),
        decode: (encoded) => ({ value: parseInt(encoded.slice(2), 16) }),
        validateTransition: (prev, next) => next.value > prev.value,
        isFinal: (state) => state.value >= 100,
        getAdjudicatorAddress: () => '0x3333333333333333333333333333333333333333',
        getAdjudicatorType: () => 'numeric'
      };
      
      expect(typeof appLogic.encode).toBe('function');
      expect(typeof appLogic.decode).toBe('function');
      expect(typeof appLogic.validateTransition).toBe('function');
      expect(typeof appLogic.isFinal).toBe('function');
      expect(typeof appLogic.getAdjudicatorAddress).toBe('function');
      expect(typeof appLogic.getAdjudicatorType).toBe('function');
      
      // Test encode/decode
      const data = { value: 42 };
      const encoded = appLogic.encode(data);
      const decoded = appLogic.decode(encoded);
      expect(decoded).toEqual(data);
      
      // Test validation
      expect(appLogic.validateTransition!({ value: 10 }, { value: 20 }, '0x')).toBe(true);
      expect(appLogic.validateTransition!({ value: 20 }, { value: 10 }, '0x')).toBe(false);
      
      // Test finality
      expect(appLogic.isFinal!({ value: 50 })).toBe(false);
      expect(appLogic.isFinal!({ value: 100 })).toBe(true);
    });
  });
  
  describe('Enums', () => {
    it('should have correct AdjudicatorStatus values', () => {
      expect(AdjudicatorStatus.VOID).toBe(0);
      expect(AdjudicatorStatus.PARTIAL).toBe(1);
      expect(AdjudicatorStatus.ACTIVE).toBe(2);
      expect(AdjudicatorStatus.INVALID).toBe(3);
      expect(AdjudicatorStatus.FINAL).toBe(4);
    });
    
    it('should have correct Role values', () => {
      expect(Role.HOST).toBe(0);
      expect(Role.GUEST).toBe(1);
    });
  });
  
  describe('Generic App Data Types', () => {
    it('should allow creating NumericState', () => {
      const state: AppDataTypes.NumericState = {
        value: BigInt(42)
      };
      
      expect(state.value).toBe(BigInt(42));
    });
    
    it('should allow creating SequentialState', () => {
      const state: AppDataTypes.SequentialState = {
        sequence: BigInt(1),
        value: BigInt(42)
      };
      
      expect(state.sequence).toBe(BigInt(1));
      expect(state.value).toBe(BigInt(42));
    });
    
    it('should allow creating TurnBasedState', () => {
      const state: AppDataTypes.TurnBasedState = {
        data: { boardState: [0, 0, 0, 0, 0, 0, 0, 0, 0] },
        turn: 0,
        status: 1,
        isComplete: false
      };
      
      expect(state.turn).toBe(0);
      expect(state.status).toBe(1);
      expect(state.isComplete).toBe(false);
    });
  });
});