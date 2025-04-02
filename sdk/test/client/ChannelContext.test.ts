import { ChannelContext } from '../../src/client/channels/ChannelContext';
import { Channel, AppLogic, Role } from '../../src/types';

// Mock dependencies
const mockClient = {
  account: {
    address: '0x1111111111111111111111111111111111111111'
  },
  openChannel: jest.fn(),
  closeChannel: jest.fn(),
  challengeChannel: jest.fn(),
  checkpointChannel: jest.fn(),
  reclaimChannel: jest.fn()
};

// Test channel
const testChannel: Channel = {
  participants: [
    '0x1111111111111111111111111111111111111111', // Host (matches mockClient.account.address)
    '0x2222222222222222222222222222222222222222'  // Guest
  ],
  adjudicator: '0x3333333333333333333333333333333333333333',
  challenge: BigInt(86400),
  nonce: BigInt(123456)
};

// Test app logic
const testAppLogic: AppLogic<{ value: number }> = {
  encode: jest.fn((data) => '0x' + data.value.toString(16)),
  decode: jest.fn((encoded) => ({ value: parseInt(encoded.slice(2), 16) })),
  validateTransition: jest.fn((prev, next) => next.value > prev.value),
  isFinal: jest.fn((state) => state.value >= 100),
  getAdjudicatorAddress: () => '0x3333333333333333333333333333333333333333'
};

describe('ChannelContext', () => {
  let context: ChannelContext<{ value: number }>;
  
  beforeEach(() => {
    jest.clearAllMocks();
    
    context = new ChannelContext(
      mockClient as any,
      testChannel,
      testAppLogic,
      { value: 0 }
    );
  });
  
  describe('constructor', () => {
    it('should initialize with provided parameters', () => {
      expect(context.getChannel()).toBe(testChannel);
      expect(context.getCurrentAppState()).toEqual({ value: 0 });
      expect(context.getRole()).toBe(Role.HOST);
    });
    
    it('should determine role based on account address', () => {
      // Client is the guest
      const guestClient = {
        ...mockClient,
        account: {
          address: '0x2222222222222222222222222222222222222222' // Guest address
        }
      };
      
      const guestContext = new ChannelContext(
        guestClient as any,
        testChannel,
        testAppLogic,
        { value: 0 }
      );
      
      expect(guestContext.getRole()).toBe(Role.GUEST);
    });
    
    it('should throw if account is not a participant', () => {
      const nonParticipantClient = {
        ...mockClient,
        account: {
          address: '0x9999999999999999999999999999999999999999' // Not a participant
        }
      };
      
      expect(() => {
        new ChannelContext(
          nonParticipantClient as any,
          testChannel,
          testAppLogic,
          { value: 0 }
        );
      }).toThrow();
    });
  });
  
  describe('getters', () => {
    it('should return channel properties', () => {
      expect(context.getChannel()).toBe(testChannel);
      expect(context.getChannelId()).toBeTruthy();
      expect(context.getCurrentAppState()).toEqual({ value: 0 });
      expect(context.getRole()).toBe(Role.HOST);
      expect(context.getOtherParticipant()).toBe('0x2222222222222222222222222222222222222222');
    });
  });
  
  describe('channel operations', () => {
    it('should open a channel', async () => {
      mockClient.openChannel.mockResolvedValue('0xCHANNEL_ID');
      
      const tokenAddress = '0x4444444444444444444444444444444444444444';
      const amounts: [bigint, bigint] = [BigInt(100), BigInt(200)];
      
      await context.open(tokenAddress, amounts);
      
      expect(mockClient.openChannel).toHaveBeenCalledWith(
        testChannel,
        expect.objectContaining({
          data: '0x0', // Encoded value: 0
          allocations: expect.arrayContaining([
            expect.objectContaining({
              destination: '0x1111111111111111111111111111111111111111',
              token: tokenAddress,
              amount: BigInt(100)
            }),
            expect.objectContaining({
              destination: '0x2222222222222222222222222222222222222222',
              token: tokenAddress,
              amount: BigInt(200)
            })
          ])
        })
      );
    });
    
    it('should update app state', async () => {
      const newState = { value: 42 };
      const result = await context.updateAppState(newState);
      
      expect(testAppLogic.validateTransition).toHaveBeenCalledWith(
        { value: 0 }, // old state
        { value: 42 }, // new state
        '0x1111111111111111111111111111111111111111' // signer
      );
      
      expect(context.getCurrentAppState()).toEqual(newState);
      expect(result).toBeDefined();
      expect(result.data).toBe('0x2a'); // Hex for 42
    });
    
    it('should throw if update is invalid', async () => {
      (testAppLogic.validateTransition as jest.Mock).mockReturnValueOnce(false);
      
      await expect(context.updateAppState({ value: 42 })).rejects.toThrow();
    });
    
    it('should close a channel', async () => {
      await context.updateAppState({ value: 42 });
      await context.close();
      
      expect(mockClient.closeChannel).toHaveBeenCalledWith(
        context.getChannelId(),
        expect.objectContaining({ data: '0x2a' }),
        []
      );
    });
    
    it('should challenge a channel', async () => {
      await context.updateAppState({ value: 42 });
      await context.challenge();
      
      expect(mockClient.challengeChannel).toHaveBeenCalledWith(
        context.getChannelId(),
        expect.objectContaining({ data: '0x2a' }),
        []
      );
    });
    
    it('should checkpoint a channel', async () => {
      await context.updateAppState({ value: 42 });
      await context.checkpoint();
      
      expect(mockClient.checkpointChannel).toHaveBeenCalledWith(
        context.getChannelId(),
        expect.objectContaining({ data: '0x2a' }),
        []
      );
    });
    
    it('should reclaim a channel', async () => {
      await context.reclaim();
      
      expect(mockClient.reclaimChannel).toHaveBeenCalledWith(
        context.getChannelId()
      );
    });
  });
  
  describe('state management', () => {
    it('should check if state is final', () => {
      context.updateAppState({ value: 50 });
      expect(context.isFinal()).toBe(false);
      
      context.updateAppState({ value: 100 });
      expect(context.isFinal()).toBe(true);
    });
    
    it('should process received state', async () => {
      const receivedState = {
        data: '0x2a', // Hex for 42
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
      
      const result = await context.processReceivedState(receivedState);
      
      expect(result).toBe(true);
      expect(context.getCurrentAppState()).toEqual({ value: 42 });
      expect(context.getCurrentState()).toBe(receivedState);
    });
    
    it('should reject invalid received state', async () => {
      // First initialize with a state
      await context.updateAppState({ value: 50 });
      
      // Then create a state that would be rejected
      const receivedState = {
        data: '0x1e', // Hex for 30 (which is < 50)
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
      
      (testAppLogic.validateTransition as jest.Mock).mockReturnValueOnce(false);
      
      const result = await context.processReceivedState(receivedState);
      
      expect(result).toBe(false);
      expect(context.getCurrentAppState()).toEqual({ value: 50 }); // Unchanged
    });
  });
});