import { describe, test, expect, jest, afterEach } from '@jest/globals';
import { Address, Hex } from 'viem';
import {
    createAuthRequestMessage,
    createAuthVerifyMessageFromChallenge,
    createAuthVerifyMessage,
    createAuthVerifyMessageWithJWT,
    createPingMessage,
    createGetConfigMessage,
    createGetLedgerBalancesMessage,
    createGetLedgerTransactionsMessage,
    createGetAppDefinitionMessage,
    createAppSessionMessage,
    createCloseAppSessionMessage,
    createApplicationMessage,
    createCloseChannelMessage,
    createResizeChannelMessage,
    createGetChannelsMessage,
    createTransferMessage,
    createECDSAMessageSigner,
} from '../../../src/rpc/api';
import {
    CreateAppSessionRequest,
    MessageSigner,
    AuthChallengeResponse,
    RPCMethod,
    RPCChannelStatus,
    RequestData,
    TransferAllocation,
    ResizeChannelRequestParams,
    AuthRequestParams,
    CloseAppSessionRequestParams,
    TxType,
} from '../../../src/rpc/types';

describe('API message creators', () => {
    const signer: MessageSigner = jest.fn(async () => '0xsig' as Hex);
    const requestId = 42;
    const timestamp = 1000;
    const clientAddress = '0x000000000000000000000000000000000000abcd' as Hex;
    const channelId = '0x000000000000000000000000000000000000cdef' as Hex;
    const appId = '0x000000000000000000000000000000000000ffff' as Hex;
    const fundDestination = '0x0000000000000000000000000000000000000000' as Address;

    afterEach(() => {
        jest.clearAllMocks();
    });

    test('createAuthRequestMessage', async () => {
        const authRequest: AuthRequestParams = {
            wallet: clientAddress,
            participant: clientAddress,
            app_name: 'test-app',
            allowances: [],
            expire: '',
            scope: '',
            application: clientAddress,
        };
        const msgStr = await createAuthRequestMessage(authRequest, requestId, timestamp);
        expect(signer).not.toHaveBeenCalled();
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [
                requestId,
                RPCMethod.AuthRequest,
                [clientAddress, clientAddress, 'test-app', [], '', '', clientAddress],
                timestamp,
            ],
            sig: [''],
        });
    });

    test('createAuthVerifyMessageFromChallenge', async () => {
        const challenge = 'challenge123';
        const msgStr = await createAuthVerifyMessageFromChallenge(signer, challenge, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, RPCMethod.AuthVerify, [[{ challenge }]], timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, RPCMethod.AuthVerify, [[{ challenge }]], timestamp],
            sig: ['0xsig'],
        });
    });

    describe('createAuthVerifyMessage', () => {
        const rawResponse: AuthChallengeResponse = {
            method: RPCMethod.AuthChallenge,
            requestId: 1750865059076,
            timestamp: 1750865059117,
            params: {
                challengeMessage: 'c8261773-2619-4fbe-9514-96392f87e7b2',
            },
            signatures: [
                '0xddf03239f12089da25362dd3799edb3c6e7c1bc558f3475298b9dbe94d43137204ad9f37bad1e620d68c6a73b8ef908788f8538b41e49c857c41e6568a8fa76a00',
            ],
        };

        test('successful challenge flow', async () => {
            const msgStr = await createAuthVerifyMessage(signer, rawResponse, requestId, timestamp);
            const challenge = 'c8261773-2619-4fbe-9514-96392f87e7b2';
            expect(signer).toHaveBeenCalledWith([requestId, RPCMethod.AuthVerify, [{ challenge }], timestamp]);
            const parsed = JSON.parse(msgStr);
            expect(parsed).toEqual({
                req: [requestId, RPCMethod.AuthVerify, [{ challenge }], timestamp],
                sig: ['0xsig'],
            });
        });
    });

    test('createPingMessage', async () => {
        const msgStr = await createPingMessage(signer, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, RPCMethod.Ping, [], timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, RPCMethod.Ping, [], timestamp],
            sig: ['0xsig'],
        });
    });

    test('createGetConfigMessage', async () => {
        const msgStr = await createGetConfigMessage(signer, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, RPCMethod.GetConfig, [], timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, RPCMethod.GetConfig, [], timestamp],
            sig: ['0xsig'],
        });
    });

    test('createGetLedgerBalancesMessage', async () => {
        const participant = '0x0123124124124100000000000000000000000000' as Address;
        const ledgerParams = [{ participant }];
        const msgStr = await createGetLedgerBalancesMessage(signer, participant, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, RPCMethod.GetLedgerBalances, ledgerParams, timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, RPCMethod.GetLedgerBalances, ledgerParams, timestamp],
            sig: ['0xsig'],
        });
    });

    test('createGetAppDefinitionMessage', async () => {
        const appParams = [{ app_session_id: appId }];
        const msgStr = await createGetAppDefinitionMessage(signer, appId, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, RPCMethod.GetAppDefinition, appParams, timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, RPCMethod.GetAppDefinition, appParams, timestamp],
            sig: ['0xsig'],
        });
    });

    test('createAppSessionMessage', async () => {
        const params = [
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
                        participant: '0xAaBbCcDdEeFf0011223344556677889900aAbBcC' as Address,
                        asset: 'usdc',
                        amount: '0.0',
                    },
                    {
                        participant: '0x00112233445566778899AaBbCcDdEeFf00112233' as Address,
                        asset: 'usdc',
                        amount: '200.0',
                    },
                ],
            },
        ];
        const msgStr = await createAppSessionMessage(signer, params, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, RPCMethod.CreateAppSession, params, timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, RPCMethod.CreateAppSession, params, timestamp],
            sig: ['0xsig'],
        });
    });

    test('createCloseAppSessionMessage', async () => {
        const closeParams: CloseAppSessionRequestParams[] = [{ app_session_id: appId, allocations: [] }];
        const msgStr = await createCloseAppSessionMessage(signer, closeParams, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, RPCMethod.CloseAppSession, closeParams, timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, RPCMethod.CloseAppSession, closeParams, timestamp],
            sig: ['0xsig'],
        });
    });

    test('createApplicationMessage', async () => {
        const messageParams = ['hello'];
        const msgStr = await createApplicationMessage(signer, appId, messageParams, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, RPCMethod.Message, messageParams, timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, RPCMethod.Message, messageParams, timestamp],
            sid: appId,
            sig: ['0xsig'],
        });
    });

    test('createCloseChannelMessage', async () => {
        const msgStr = await createCloseChannelMessage(signer, channelId, fundDestination, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([
            requestId,
            RPCMethod.CloseChannel,
            [{ channel_id: channelId, funds_destination: fundDestination }],
            timestamp,
        ]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [
                requestId,
                RPCMethod.CloseChannel,
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
            req: [requestId, RPCMethod.AuthVerify, [{ jwt: jwtToken }], timestamp],
            sig: undefined,
        });
    });

    test('createResizeChannelMessage', async () => {
        const resizeParams: ResizeChannelRequestParams[] = [
            {
                channel_id: channelId,
                funds_destination: fundDestination,
                resize_amount: 1000n,
            },
        ];
        const msgStr = await createResizeChannelMessage(signer, resizeParams, requestId, timestamp);
        // The signer should be called with the original bigint value
        expect(signer).toHaveBeenCalledWith([requestId, RPCMethod.ResizeChannel, resizeParams, timestamp]);
        const parsed = JSON.parse(msgStr);
        // The parsed message should have the stringified bigint
        const resizeParamsExpected = resizeParams.map((param) => ({
            ...param,
            resize_amount: param.resize_amount?.toString(),
        }));
        expect(parsed).toEqual({
            req: [requestId, RPCMethod.ResizeChannel, resizeParamsExpected, timestamp],
            sig: ['0xsig'],
        });
    });

    test('createGetChannelsMessage', async () => {
        const participant = '0x0123124124124131000000000000000000000000' as Address;
        const msgStr = await createGetChannelsMessage(signer, participant, RPCChannelStatus.Open, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([
            requestId,
            RPCMethod.GetChannels,
            [{ participant, status: RPCChannelStatus.Open }],
            timestamp,
        ]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, RPCMethod.GetChannels, [{ participant, status: RPCChannelStatus.Open }], timestamp],
            sig: ['0xsig'],
        });
    });

    test('createTransferMessage', async () => {
        const destination = '0x1234567890123456789012345678901234567890' as Address;
        const allocations: TransferAllocation[] = [
            {
                asset: 'USDC',
                amount: '100.5',
            },
            {
                asset: 'ETH',
                amount: '0.25',
            },
        ];
        const transferParams = [{ destination, allocations }];
        const msgStr = await createTransferMessage(signer, destination, allocations, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, RPCMethod.Transfer, transferParams, timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, RPCMethod.Transfer, transferParams, timestamp],
            sig: ['0xsig'],
        });
    });

    test('createGetLedgerTransactionsMessage with no filters', async () => {
        const accountId = 'test-account';
        const expectedParams = [{ account_id: accountId }];
        const msgStr = await createGetLedgerTransactionsMessage(signer, accountId, undefined, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, RPCMethod.GetLedgerTransactions, expectedParams, timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, RPCMethod.GetLedgerTransactions, expectedParams, timestamp],
            sig: ['0xsig'],
        });
    });

    test('createGetLedgerTransactionsMessage with all filters', async () => {
        const accountId = 'test-account';
        const filters = {
            asset: 'USDC',
            tx_type: TxType.Transfer,
            offset: 10,
            limit: 20,
            sort: 'desc' as const,
        };
        const expectedParams = [
            {
                account_id: accountId,
                asset: 'USDC',
                tx_type: TxType.Transfer,
                offset: 10,
                limit: 20,
                sort: 'desc',
            },
        ];
        const msgStr = await createGetLedgerTransactionsMessage(signer, accountId, filters, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, RPCMethod.GetLedgerTransactions, expectedParams, timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, RPCMethod.GetLedgerTransactions, expectedParams, timestamp],
            sig: ['0xsig'],
        });
    });

    test('createGetLedgerTransactionsMessage with partial filters', async () => {
        const accountId = 'test-account';
        const filters = {
            asset: 'ETH',
            limit: 5,
        };
        const expectedParams = [
            {
                account_id: accountId,
                asset: 'ETH',
                limit: 5,
            },
        ];
        const msgStr = await createGetLedgerTransactionsMessage(signer, accountId, filters, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, RPCMethod.GetLedgerTransactions, expectedParams, timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, RPCMethod.GetLedgerTransactions, expectedParams, timestamp],
            sig: ['0xsig'],
        });
    });

    test('createGetLedgerTransactionsMessage filters out null/undefined/empty values', async () => {
        const accountId = 'test-account';
        const filters = {
            asset: '',
            tx_type: TxType.Transfer,
            offset: 0,
            limit: undefined,
            sort: null as any,
        };
        const expectedParams = [
            {
                account_id: accountId,
                tx_type: TxType.Transfer,
                offset: 0,
            },
        ];
        const msgStr = await createGetLedgerTransactionsMessage(signer, accountId, filters, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, RPCMethod.GetLedgerTransactions, expectedParams, timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, RPCMethod.GetLedgerTransactions, expectedParams, timestamp],
            sig: ['0xsig'],
        });
    });

    test('createECDSAMessageSigner', async () => {
        const privateKey = '0xb482c8fa261c29eaaa646703948e2cc2a2ff54411cc42d8fce9a161035dfb3dc';
        const payload = [42, 'ping', [{ p1: 4337, p2: 7702 }], 20] as unknown as RequestData;
        const signer = createECDSAMessageSigner(privateKey);
        const signature = await signer(payload);
        expect(signature).toBeDefined();
        expect(signature).toEqual(
            '0xebf96c7d3d64ab9195a341d3c922e2cb88ea592d2e229aa64d27e024f895e5720e68786c8b34a61d34a0b6f5e0f65dbe95f0a46dee9b7055df3e33f3209ea0d21b',
        );
    });
});
