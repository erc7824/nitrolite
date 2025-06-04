import { describe, test, expect, jest } from '@jest/globals';
import { Address, Hex } from 'viem';
import {
    createAuthRequestMessage,
    createAuthVerifyMessageFromChallenge,
    createAuthVerifyMessage,
    createAuthVerifyMessageWithJWT,
    createPingMessage,
    createGetConfigMessage,
    createGetLedgerBalancesMessage,
    createGetAppDefinitionMessage,
    createAppSessionMessage,
    createCloseAppSessionMessage,
    createApplicationMessage,
    createCloseChannelMessage,
    createResizeChannelMessage,
    createGetChannelsMessage,
} from '../../src/rpc/api';
import {
    CreateAppSessionRequest,
    CloseAppSessionRequest,
    ResizeChannel,
    MessageSigner,
    AuthRequest,
} from '../../src/rpc/types';

describe('API message creators', () => {
    const signer: MessageSigner = jest.fn(async () => '0xsig' as Hex);
    const requestId = 42;
    const timestamp = 1000;
    const clientAddress = '0x000000000000000000000000000000000000abcd' as Hex;
    const channelId = '0x000000000000000000000000000000000000cdef' as Hex;
    const appId = '0x000000000000000000000000000000000000ffff' as Hex;
    const fundDestination = '0x' as Address;
    const sampleIntent = [1, 2, 3];

    afterEach(() => {
        jest.clearAllMocks();
    });

    test('createAuthRequestMessage', async () => {
        const authRequest: AuthRequest = {
            wallet: clientAddress,
            participant: clientAddress,
            app_name: 'test-app',
            allowances: {},
        };
        const msgStr = await createAuthRequestMessage(authRequest, requestId, timestamp);
        expect(signer).not.toHaveBeenCalled();
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, 'auth_request', [clientAddress, clientAddress, 'test-app', [], '', '', ''], timestamp],
            sig: [''],
        });
    });

    test('createAuthVerifyMessageFromChallenge', async () => {
        const challenge = 'challenge123';
        const msgStr = await createAuthVerifyMessageFromChallenge(signer, challenge, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, 'auth_verify', [[{ challenge }]], timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, 'auth_verify', [[{ challenge }]], timestamp],
            sig: ['0xsig'],
        });
    });

    describe('createAuthVerifyMessage', () => {
        const rawResponse = JSON.stringify({
            res: [999, 'auth_challenge', [{ challenge_message: 'msg' }], 200],
        });

        test('successful challenge flow', async () => {
            const msgStr = await createAuthVerifyMessage(signer, rawResponse, requestId, timestamp);
            expect(signer).toHaveBeenCalledWith([requestId, 'auth_verify', [{ challenge: 'msg' }], timestamp]);
            const parsed = JSON.parse(msgStr);
            expect(parsed).toEqual({
                req: [requestId, 'auth_verify', [{ challenge: 'msg' }], timestamp],
                sig: ['0xsig'],
            });
        });

        test('throws on invalid response', async () => {
            await expect(createAuthVerifyMessage(signer, '{}', requestId, timestamp)).rejects.toThrow(
                'Invalid auth_challenge response',
            );
        });

        test('throws on wrong method', async () => {
            const wrong = JSON.stringify({
                res: [100, 'other', [{ challenge_message: 'msg' }], 200],
            });
            await expect(createAuthVerifyMessage(signer, wrong, requestId, timestamp)).rejects.toThrow(
                "Expected 'auth_challenge' method",
            );
        });
    });

    test('createPingMessage', async () => {
        const msgStr = await createPingMessage(signer, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, 'ping', [], timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, 'ping', [], timestamp],
            sig: ['0xsig'],
        });
    });

    test('createGetConfigMessage', async () => {
        const msgStr = await createGetConfigMessage(signer, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, 'get_config', [], timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, 'get_config', [], timestamp],
            sig: ['0xsig'],
        });
    });

    test('createGetLedgerBalancesMessage', async () => {
        const participant = '0x01231241241241' as Address;
        const ledgerParams = [{ participant }];
        const msgStr = await createGetLedgerBalancesMessage(signer, participant, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, 'get_ledger_balances', ledgerParams, timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, 'get_ledger_balances', ledgerParams, timestamp],
            sig: ['0xsig'],
        });
    });

    test('createGetAppDefinitionMessage', async () => {
        const appParams = [{ app_session_id: appId }];
        const msgStr = await createGetAppDefinitionMessage(signer, appId, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, 'get_app_definition', appParams, timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, 'get_app_definition', appParams, timestamp],
            sig: ['0xsig'],
        });
    });

    test('createAppSessionMessage', async () => {
        const params: CreateAppSessionRequest[] = [
            {
                definition: {
                    protocol: 'p',
                    participants: [],
                    weights: [],
                    quorum: 0,
                    challenge: 0,
                    nonce: 0,
                },
                allocations: [
                    {
                        participant: '0xAaBbCcDdEeFf0011223344556677889900aAbBcC',
                        asset: 'usdc',
                        amount: '0.0',
                    },
                    {
                        participant: '0x00112233445566778899AaBbCcDdEeFf00112233',
                        asset: 'usdc',
                        amount: '200.0',
                    },
                ],
            },
        ];
        // @ts-ignore
        const msgStr = await createAppSessionMessage(signer, params, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, 'create_app_session', params, timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, 'create_app_session', params, timestamp],
            sig: ['0xsig'],
        });
    });

    test('createCloseAppSessionMessage', async () => {
        const closeParams = [{ app_session_id: appId, allocation: [] }];
        // @ts-ignore
        const msgStr = await createCloseAppSessionMessage(signer, closeParams, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, 'close_app_session', closeParams, timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, 'close_app_session', closeParams, timestamp],
            sig: ['0xsig'],
        });
    });

    test('createApplicationMessage', async () => {
        const messageParams = ['hello'];
        const msgStr = await createApplicationMessage(signer, appId, messageParams, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, 'message', messageParams, timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, 'message', messageParams, timestamp],
            sid: appId,
            sig: ['0xsig'],
        });
    });

    test('createCloseChannelMessage', async () => {
        const msgStr = await createCloseChannelMessage(signer, channelId, fundDestination, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([
            requestId,
            'close_channel',
            [{ channel_id: channelId, funds_destination: fundDestination }],
            timestamp,
        ]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [
                requestId,
                'close_channel',
                [{ channel_id: channelId, funds_destination: fundDestination }],
                timestamp,
            ],
            sig: ['0xsig'],
        });
    });

    test('createAuthVerifyMessageWithJWT', async () => {
        const jwtToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...';
        const msgStr = await createAuthVerifyMessageWithJWT(jwtToken, requestId, timestamp);
        expect(signer).not.toHaveBeenCalled();
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, 'auth_verify', [{ jwt: jwtToken }], timestamp],
            sig: undefined,
        });
    });

    test('createResizeChannelMessage', async () => {
        const resizeParams: ResizeChannel[] = [
            {
                channel_id: channelId,
                allocations: [
                    {
                        participant: clientAddress,
                        asset: 'usdc',
                        amount: '100.0',
                    },
                ],
            },
        ];
        const msgStr = await createResizeChannelMessage(signer, resizeParams, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, 'resize_channel', resizeParams, timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, 'resize_channel', resizeParams, timestamp],
            sig: ['0xsig'],
        });
    });

    test('createGetChannelsMessage', async () => {
        const msgStr = await createGetChannelsMessage(signer, '0x0123124124124131', requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([
            requestId,
            'get_channels',
            [{ participant: '0x0123124124124131' }],
            timestamp,
        ]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, 'get_channels', [{ participant: '0x0123124124124131' }], timestamp],
            sig: ['0xsig'],
        });
    });
});
