import { describe, test, expect } from '@jest/globals';
import { Address, Hex } from 'viem';
import {
    createPingMessage,
    createGetConfigMessage,
    createGetAssetsMessage,
    createGetBalancesMessage,
    createGetTransactionsMessage,
    createGetHomeChannelMessage,
    createGetEscrowChannelMessage,
    createGetChannelsMessage,
    createGetLatestStateMessage,
    createGetStatesMessage,
    createCreateChannelMessage,
    createSubmitStateMessage,
    createGetAppDefinitionMessage,
    createGetAppSessionsMessage,
    createCreateAppSessionMessage,
    createSubmitAppStateMessage,
    createSubmitDepositStateMessage,
    createRebalanceAppSessionsMessage,
    createRegisterMessage,
    createGetSessionKeysMessage,
    createRevokeSessionKeyMessage,
    createApplicationMessage,
} from '../../../src/rpc/api';
import {
    RPCMethod,
    RPCAppStateIntent,
    RPCAppParticipant,
    RPCAppDefinition,
    RPCState,
    RPCAppSessionAllocation,
    RPCAppStateUpdate,
    RPCSignedAppStateUpdate,
    RPCTransition,
    RPCTransitionType,
    RPCLedger,
} from '../../../src/rpc/types';
import { RPCMessageType } from '../../../src/rpc/types';

describe('API v1 message creators', () => {
    const requestId = 42;
    const timestamp = 1000;
    const walletAddress = '0x000000000000000000000000000000000000abcd' as Address;
    const tokenAddress = '0x000000000000000000000000000000000000cdef' as Address;
    const appSessionId = '0x000000000000000000000000000000000000ffff' as Hex;

    describe('Node methods', () => {
        test('createPingMessage', () => {
            const msgStr = createPingMessage(requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed).toEqual([
                RPCMessageType.Request,
                requestId,
                RPCMethod.Ping,
                {},
                timestamp,
            ]);
        });

        test('createGetConfigMessage', () => {
            const msgStr = createGetConfigMessage(requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed).toEqual([
                RPCMessageType.Request,
                requestId,
                RPCMethod.GetConfig,
                {},
                timestamp,
            ]);
        });

        test('createGetAssetsMessage without chainId', () => {
            const msgStr = createGetAssetsMessage(undefined, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed).toEqual([
                RPCMessageType.Request,
                requestId,
                RPCMethod.GetAssets,
                {},
                timestamp,
            ]);
        });

        test('createGetAssetsMessage with chainId', () => {
            const msgStr = createGetAssetsMessage(1, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed).toEqual([
                RPCMessageType.Request,
                requestId,
                RPCMethod.GetAssets,
                { chain_id: 1 },
                timestamp,
            ]);
        });
    });

    describe('User methods', () => {
        test('createGetBalancesMessage', () => {
            const msgStr = createGetBalancesMessage(walletAddress, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed).toEqual([
                RPCMessageType.Request,
                requestId,
                RPCMethod.GetBalances,
                { wallet: walletAddress },
                timestamp,
            ]);
        });

        test('createGetTransactionsMessage with no filters', () => {
            const msgStr = createGetTransactionsMessage(walletAddress, undefined, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed).toEqual([
                RPCMessageType.Request,
                requestId,
                RPCMethod.GetTransactions,
                { wallet: walletAddress },
                timestamp,
            ]);
        });

        test('createGetTransactionsMessage with all filters', () => {
            const options = {
                asset: 'usdc',
                tx_type: 'transfer',
                from_time: 1000,
                to_time: 2000,
                pagination: { offset: 0, limit: 10, sort: 'desc' as const },
            };
            const msgStr = createGetTransactionsMessage(walletAddress, options, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed[3]).toMatchObject({
                wallet: walletAddress,
                asset: 'usdc',
                tx_type: 'transfer',
                from_time: 1000,
                to_time: 2000,
                pagination: { offset: 0, limit: 10, sort: 'desc' },
            });
        });
    });

    describe('Channel methods', () => {
        test('createGetHomeChannelMessage', () => {
            const msgStr = createGetHomeChannelMessage(walletAddress, 'usdc', requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed).toEqual([
                RPCMessageType.Request,
                requestId,
                RPCMethod.GetHomeChannel,
                { wallet: walletAddress, asset: 'usdc' },
                timestamp,
            ]);
        });

        test('createGetEscrowChannelMessage', () => {
            const channelId = 'channel123';
            const msgStr = createGetEscrowChannelMessage(channelId, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed).toEqual([
                RPCMessageType.Request,
                requestId,
                RPCMethod.GetEscrowChannel,
                { escrow_channel_id: channelId },
                timestamp,
            ]);
        });

        test('createGetChannelsMessage with no filters', () => {
            const msgStr = createGetChannelsMessage(walletAddress, undefined, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed).toEqual([
                RPCMessageType.Request,
                requestId,
                RPCMethod.GetChannels,
                { wallet: walletAddress },
                timestamp,
            ]);
        });

        test('createGetChannelsMessage with filters', () => {
            const options = {
                asset: 'usdc',
                status: 'open',
                pagination: { offset: 0, limit: 10, sort: 'asc' as const },
            };
            const msgStr = createGetChannelsMessage(walletAddress, options, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed[3]).toMatchObject({
                wallet: walletAddress,
                asset: 'usdc',
                status: 'open',
                pagination: { offset: 0, limit: 10, sort: 'asc' },
            });
        });

        test('createGetLatestStateMessage', () => {
            const msgStr = createGetLatestStateMessage(walletAddress, 'usdc', true, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed).toEqual([
                RPCMessageType.Request,
                requestId,
                RPCMethod.GetLatestState,
                { wallet: walletAddress, asset: 'usdc', only_signed: true },
                timestamp,
            ]);
        });

        test('createGetStatesMessage', () => {
            const options = {
                epoch: 1,
                channel_id: 'channel123',
                pagination: { offset: 0, limit: 10, sort: 'desc' as const },
            };
            const msgStr = createGetStatesMessage(walletAddress, 'usdc', false, options, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed[3]).toMatchObject({
                wallet: walletAddress,
                asset: 'usdc',
                only_signed: false,
                epoch: 1,
                channel_id: 'channel123',
                pagination: { offset: 0, limit: 10, sort: 'desc' },
            });
        });

        test('createCreateChannelMessage', () => {
            const homeLedger: RPCLedger = {
                tokenAddress,
                blockchainId: 1,
                userBalance: '100',
                userNetFlow: '0',
                nodeBalance: '100',
                nodeNetFlow: '0',
            };

            const state: RPCState = {
                id: 'state123',
                transitions: [],
                asset: 'usdc',
                userWallet: walletAddress,
                epoch: 0,
                version: 0,
                homeLedger,
            };

            const channelDef = { nonce: 1, challenge: 3600 };
            const msgStr = createCreateChannelMessage(state, channelDef, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed[2]).toBe(RPCMethod.CreateChannel);
            expect(parsed[3]).toMatchObject({
                state,
                channel_definition: channelDef,
            });
        });

        test('createSubmitStateMessage', () => {
            const homeLedger: RPCLedger = {
                tokenAddress,
                blockchainId: 1,
                userBalance: '100',
                userNetFlow: '0',
                nodeBalance: '100',
                nodeNetFlow: '0',
            };

            const state: RPCState = {
                id: 'state123',
                transitions: [],
                asset: 'usdc',
                userWallet: walletAddress,
                epoch: 0,
                version: 1,
                homeLedger,
            };

            const msgStr = createSubmitStateMessage(state, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed).toEqual([
                RPCMessageType.Request,
                requestId,
                RPCMethod.SubmitState,
                { state },
                timestamp,
            ]);
        });
    });

    describe('App Session methods', () => {
        test('createGetAppDefinitionMessage', () => {
            const msgStr = createGetAppDefinitionMessage(appSessionId, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed).toEqual([
                RPCMessageType.Request,
                requestId,
                RPCMethod.GetAppDefinition,
                { app_session_id: appSessionId },
                timestamp,
            ]);
        });

        test('createGetAppSessionsMessage with no filters', () => {
            const msgStr = createGetAppSessionsMessage(undefined, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed).toEqual([
                RPCMessageType.Request,
                requestId,
                RPCMethod.GetAppSessions,
                {},
                timestamp,
            ]);
        });

        test('createGetAppSessionsMessage with filters', () => {
            const options = {
                app_session_id: 'session123',
                participant: walletAddress,
                status: 'open',
                pagination: { offset: 0, limit: 10, sort: 'asc' as const },
            };
            const msgStr = createGetAppSessionsMessage(options, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed[3]).toMatchObject(options);
        });

        test('createCreateAppSessionMessage', () => {
            const participants: RPCAppParticipant[] = [
                { walletAddress, signatureWeight: 1 },
            ];
            const definition: RPCAppDefinition = {
                application: 'test-app',
                participants,
                quorum: 1,
                nonce: 1,
            };
            const quorumSigs: Hex[] = ['0xsig1' as Hex];
            const sessionData = '{"data":"test"}';

            const msgStr = createCreateAppSessionMessage(definition, quorumSigs, sessionData, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed[2]).toBe(RPCMethod.CreateAppSession);
            expect(parsed[3]).toMatchObject({
                definition,
                quorum_sigs: quorumSigs,
                session_data: sessionData,
            });
        });

        test('createSubmitAppStateMessage', () => {
            const allocations: RPCAppSessionAllocation[] = [
                { participant: walletAddress, asset: 'usdc', amount: '100' },
            ];
            const appStateUpdate = {
                app_session_id: appSessionId,
                intent: RPCAppStateIntent.Operate,
                version: 1,
                allocations,
            };
            const quorumSigs: Hex[] = ['0xsig1' as Hex];

            const msgStr = createSubmitAppStateMessage(appStateUpdate, quorumSigs, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed[2]).toBe(RPCMethod.SubmitAppState);
            expect(parsed[3]).toMatchObject({
                app_state_update: appStateUpdate,
                quorum_sigs: quorumSigs,
            });
        });

        test('createSubmitDepositStateMessage', () => {
            const allocations: RPCAppSessionAllocation[] = [
                { participant: walletAddress, asset: 'usdc', amount: '100' },
            ];
            const appStateUpdate: RPCAppStateUpdate = {
                app_session_id: appSessionId,
                intent: RPCAppStateIntent.Deposit,
                version: 1,
                allocations,
            };
            const quorumSigs: Hex[] = ['0xsig1' as Hex];
            const homeLedger: RPCLedger = {
                tokenAddress,
                blockchainId: 1,
                userBalance: '100',
                userNetFlow: '0',
                nodeBalance: '100',
                nodeNetFlow: '0',
            };
            const userState: RPCState = {
                id: 'state123',
                transitions: [],
                asset: 'usdc',
                userWallet: walletAddress,
                epoch: 0,
                version: 1,
                homeLedger,
            };

            const msgStr = createSubmitDepositStateMessage(appStateUpdate, quorumSigs, userState, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed[2]).toBe(RPCMethod.SubmitDepositState);
            expect(parsed[3]).toMatchObject({
                app_state_update: appStateUpdate,
                quorum_sigs: quorumSigs,
                user_state: userState,
            });
        });

        test('createRebalanceAppSessionsMessage', () => {
            const allocations: RPCAppSessionAllocation[] = [
                { participant: walletAddress, asset: 'usdc', amount: '100' },
            ];
            const signedUpdates: RPCSignedAppStateUpdate[] = [
                {
                    app_state_update: {
                        app_session_id: appSessionId,
                        intent: RPCAppStateIntent.Rebalance,
                        version: 1,
                        allocations,
                    },
                    quorum_sigs: ['0xsig1' as Hex],
                },
            ];

            const msgStr = createRebalanceAppSessionsMessage(signedUpdates, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed).toEqual([
                RPCMessageType.Request,
                requestId,
                RPCMethod.RebalanceAppSessions,
                { signed_updates: signedUpdates },
                timestamp,
            ]);
        });
    });

    describe('Session Key methods', () => {
        test('createRegisterMessage with minimal params', () => {
            const msgStr = createRegisterMessage(walletAddress, undefined, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed).toEqual([
                RPCMessageType.Request,
                requestId,
                RPCMethod.Register,
                { address: walletAddress },
                timestamp,
            ]);
        });

        test('createRegisterMessage with all options', () => {
            const sessionKey = '0x0000000000000000000000000000000000001234' as Address;
            const options = {
                session_key: sessionKey,
                application: 'test-app',
                allowances: [{ asset: 'usdc', allowance: '1000' }],
                scope: 'full',
                expires_at: 9999999999,
            };
            const msgStr = createRegisterMessage(walletAddress, options, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed[3]).toMatchObject({
                address: walletAddress,
                ...options,
            });
        });

        test('createGetSessionKeysMessage', () => {
            const msgStr = createGetSessionKeysMessage(walletAddress, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed).toEqual([
                RPCMessageType.Request,
                requestId,
                RPCMethod.GetSessionKeys,
                { wallet: walletAddress },
                timestamp,
            ]);
        });

        test('createRevokeSessionKeyMessage', () => {
            const sessionKey = '0x0000000000000000000000000000000000001234' as Address;
            const msgStr = createRevokeSessionKeyMessage(sessionKey, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed).toEqual([
                RPCMessageType.Request,
                requestId,
                RPCMethod.RevokeSessionKey,
                { session_key: sessionKey },
                timestamp,
            ]);
        });
    });

    describe('Application Message', () => {
        test('createApplicationMessage', () => {
            const params = { type: 'custom', data: 'hello' };
            const msgStr = createApplicationMessage(params, requestId, timestamp);
            const parsed = JSON.parse(msgStr);
            expect(parsed).toEqual([
                RPCMessageType.Request,
                requestId,
                RPCMethod.Message,
                params,
                timestamp,
            ]);
        });
    });
});
