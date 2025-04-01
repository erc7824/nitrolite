import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { RPCClient, MemoryRPCProvider, RPCMethod } from '../../src/rpc';
import { createPublicClient, http, Address, Hex } from 'viem';
import { hardhat } from 'viem/chains';

describe('RPCClient', () => {
  // Test addresses
  const aliceAddress = '0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266' as Address; // Hardhat #0
  const bobAddress = '0x70997970C51812dc3A010C7d01b50e0d17dc79C8' as Address; // Hardhat #1
  
  // Mock signer function
  const mockSigner = vi.fn().mockImplementation((message: Hex) => Promise.resolve('0xsignature' as Hex));
  
  // Create providers and clients
  let aliceProvider: MemoryRPCProvider;
  let bobProvider: MemoryRPCProvider;
  let aliceClient: RPCClient;
  let bobClient: RPCClient;
  
  beforeEach(async () => {
    // Reset network
    MemoryRPCProvider.resetNetwork();
    
    // Create providers
    aliceProvider = new MemoryRPCProvider(aliceAddress);
    bobProvider = new MemoryRPCProvider(bobAddress);
    
    // Create clients
    aliceClient = new RPCClient({
      provider: aliceProvider,
      address: aliceAddress,
      signer: mockSigner
    });
    
    bobClient = new RPCClient({
      provider: bobProvider,
      address: bobAddress,
      signer: mockSigner
    });
    
    // Connect clients
    await aliceClient.connect();
    await bobClient.connect();
  });
  
  afterEach(async () => {
    // Disconnect clients
    await aliceClient.disconnect();
    await bobClient.disconnect();
  });
  
  it('should connect and disconnect successfully', async () => {
    expect(MemoryRPCProvider.getConnectedAddresses()).toHaveLength(2);
    expect(MemoryRPCProvider.getConnectedAddresses()).toContain(aliceAddress.toLowerCase());
    expect(MemoryRPCProvider.getConnectedAddresses()).toContain(bobAddress.toLowerCase());
    
    await aliceClient.disconnect();
    expect(MemoryRPCProvider.getConnectedAddresses()).toHaveLength(1);
    expect(MemoryRPCProvider.getConnectedAddresses()).not.toContain(aliceAddress.toLowerCase());
    
    await bobClient.disconnect();
    expect(MemoryRPCProvider.getConnectedAddresses()).toHaveLength(0);
  });
  
  it('should send and receive requests', async () => {
    // Register a method handler on Bob's client
    bobClient.registerMethod('test_method', async (params) => {
      return [params[0] * 2];
    });
    
    // Send a request from Alice to Bob
    const result = await aliceClient.sendRequest<[number]>(bobAddress, 'test_method', [42]);
    
    // Check the result
    expect(result).toEqual([84]);
  });
  
  it('should handle errors in requests', async () => {
    // Register a method handler on Bob's client that throws an error
    bobClient.registerMethod('error_method', async () => {
      throw new Error('Test error');
    });
    
    // Send a request from Alice to Bob and expect it to fail
    await expect(
      aliceClient.sendRequest(bobAddress, 'error_method', [])
    ).rejects.toThrow(/RPC Error/);
  });
  
  it('should handle method not found', async () => {
    // Send a request for a non-existent method
    await expect(
      aliceClient.sendRequest(bobAddress, 'non_existent_method', [])
    ).rejects.toThrow(/Method/);
  });
  
  it('should send and receive notifications', async () => {
    // Set up a notification handler on Bob's client
    const notificationHandler = vi.fn();
    bobClient.on('test_notification', notificationHandler);
    
    // Send a notification from Alice to Bob
    await aliceClient.sendNotification(bobAddress, 'test_notification', [42, 'test']);
    
    // Wait for the notification to be processed
    await new Promise(resolve => setTimeout(resolve, 10));
    
    // Check that Bob received the notification
    expect(notificationHandler).toHaveBeenCalledWith([42, 'test'], aliceAddress);
  });
  
  it('should handle the ping method', async () => {
    // Send a ping request from Alice to Bob
    const result = await aliceClient.sendRequest<[string]>(bobAddress, RPCMethod.PING, []);
    
    // Check the result
    expect(result).toEqual(['pong']);
  });
  
  it('should handle the get_time method', async () => {
    // Send a get_time request from Alice to Bob
    const result = await aliceClient.sendRequest<[number]>(bobAddress, RPCMethod.GET_TIME, []);
    
    // Check the result is a number
    expect(typeof result[0]).toBe('number');
  });
  
  it('should register and unregister methods', async () => {
    // Register a method handler
    const handler = async () => [42];
    aliceClient.registerMethod('test_method', handler);
    
    // Send a request from Bob to Alice
    const result = await bobClient.sendRequest<[number]>(aliceAddress, 'test_method', []);
    expect(result).toEqual([42]);
    
    // Unregister the method
    aliceClient.unregisterMethod('test_method');
    
    // Send another request, should fail
    await expect(
      bobClient.sendRequest(aliceAddress, 'test_method', [])
    ).rejects.toThrow(/Method/);
  });
});