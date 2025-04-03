import {
    Address,
    createPublicClient,
    createWalletClient,
    http,
    parseEther,
    PublicClient,
    WalletClient,
} from 'viem';
import { privateKeyToAccount } from 'viem/accounts';
import { localhost } from 'viem/chains';
import { Channel, Role } from '../../src/types';
import { AndApp, AndAppState } from './AndApp';
import { ChannelContext, NitroliteClient } from '../../src';
import { describe, beforeEach, expect, test } from '@jest/globals';

describe('ChannelContext', () => {
    const hostPrivateKey =
        '0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d';
    const guestPrivateKey =
        '0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a';

    const hostAccount = privateKeyToAccount(hostPrivateKey);
    const guestAccount = privateKeyToAccount(guestPrivateKey);

    const contractAddresses = {
        custody: '0x3Aa5ebB10DC797CAC828524e59A333d0A371443c' as Address,
        adjudicators: {
            and: '0x68B1D87F95878fE05B998F19b66F4baba5De1aed' as Address,
        },
    };

    const tokenAddress =
        '0x9fe46736679d2d9a65f0992f2272de9f3c7fa6e0' as Address;

    let publicClient: PublicClient;
    let hostWalletClient: WalletClient;
    let hostNitroliteClient: NitroliteClient;
    let andApp: AndApp;
    let channel: Channel;

    beforeEach(() => {
        // Setup clients
        publicClient = createPublicClient({
            chain: localhost,
            transport: http(),
        });

        hostWalletClient = createWalletClient({
            chain: localhost,
            transport: http(),
            account: hostAccount,
        });

        // Setup Nitrolite clients
        hostNitroliteClient = new NitroliteClient({
            publicClient,
            walletClient: hostWalletClient,
            account: hostAccount,
            addresses: contractAddresses,
        });

        // Setup AND app
        andApp = new AndApp(contractAddresses.adjudicators.and);

        // Setup test channel
        channel = {
            participants: [hostAccount.address, guestAccount.address],
            nonce: BigInt(Date.now()),
            adjudicator: andApp.getAdjudicatorAddress(),
            challenge: 0n,
        };
    });

    test('should create channel context correctly', () => {
        const context = new ChannelContext(
            hostNitroliteClient,
            channel,
            andApp
        );
        expect(context.getRole()).toBe(Role.HOST);
        expect(context.getOtherParticipant()).toBe(guestAccount.address);
    });

    test('should open channel with initial state', async () => {
        const context = new ChannelContext(
            hostNitroliteClient,
            channel,
            andApp
        );

        const initialState: AndAppState = {
            isFinal: false,
            flag: false,
        };

        const amounts: [bigint, bigint] = [100n, 100n];

        const channelId = await context.open(
            initialState,
            tokenAddress,
            amounts
        );
        expect(channelId).toBeDefined();

        const currentState = context.getCurrentAppState();
        expect(currentState).toEqual(initialState);
    });

    test('should append valid state transition', async () => {
        const context = new ChannelContext(
            hostNitroliteClient,
            channel,
            andApp
        );

        // Open channel
        const initialState: AndAppState = {
            isFinal: false,
            flag: false,
        };

        const amounts: [bigint, bigint] = [100n, 100n];
        await context.open(initialState, tokenAddress, amounts);

        // Append new state
        const newState: AndAppState = {
            isFinal: false,
            flag: true,
        };

        context.appendAppState(newState);
        const currentState = context.getCurrentAppState();
        expect(currentState).toEqual(newState);
    });

    test('should reject invalid state transition', async () => {
        const context = new ChannelContext(
            hostNitroliteClient,
            channel,
            andApp
        );

        // Open channel
        const initialState: AndAppState = {
            isFinal: true,
            flag: true,
        };

        const amounts: [bigint, bigint] = [100n, 100n];
        await context.open(initialState, tokenAddress, amounts);

        // Try to append invalid state
        const newState: AndAppState = {
            isFinal: false, // Can't append state, after final was added
            flag: false,
        };

        expect(() => context.appendAppState(newState)).toThrow(
            'Invalid state transition'
        );
    });

    test('should close channel with final state', async () => {
        const context = new ChannelContext(
            hostNitroliteClient,
            channel,
            andApp
        );

        // Open channel
        const initialState: AndAppState = {
            isFinal: false,
            flag: false,
        };

        const amounts: [bigint, bigint] = [100n, 100n];
        await context.open(initialState, tokenAddress, amounts);

        // Append final state
        const finalState: AndAppState = {
            isFinal: true,
            flag: false,
        };

        await context.appendAppState(finalState);
        expect(context.isFinal()).toBe(true);

        // Close channel
        await expect(context.close()).resolves.not.toThrow();
    });
});
