import { createAuthSessionWithClearnode } from '@/auth';
import { BlockchainUtils } from '@/blockchainUtils';
import { DatabaseUtils } from '@/databaseUtils';
import { Identity } from '@/identity';
import { TestNitroliteClient } from '@/nitroliteClient';
import { CONFIG } from '@/setup';
import {
    getCloseAppSessionPredicate,
    getCreateAppSessionPredicate,
    getGetLedgerBalancesPredicate,
    getGetAppSessionsPredicate,
    getSubmitStatePredicate,
    TestWebSocket,
} from '@/ws';
import {
    AppDefinition,
    createAppSessionMessage,
    createCloseAppSessionMessage,
    createGetAppSessionsMessage,
    createGetLedgerBalancesMessage,
    createSubmitStateMessage,
    RPCChannelStatus,
    rpcResponseParser,
} from '@erc7824/nitrolite';
import { Hex, parseUnits } from 'viem';

describe('Close channel', () => {
    const depositAmount = parseUnits('100', 6); // 100 USDC (decimals = 6)
    const decimalDepositAmount = BigInt(100);

    let ws: TestWebSocket;
    let identity: Identity;
    let client: TestNitroliteClient;

    let appWS: TestWebSocket;
    let appIdentity: Identity;

    let cpWS: TestWebSocket;
    let cpIdentity: Identity;
    let appCPIdentity: Identity;
    let cpClient: TestNitroliteClient;

    let blockUtils: BlockchainUtils;
    let databaseUtils: DatabaseUtils;

    let channelId: Hex;
    let cpChannelId: Hex;
    let appSessionId: string;

    const fetchAndParseAppSessions = async () => {
        const getAppSessionsMsg = await createGetAppSessionsMessage(
            appIdentity.messageSigner,
            appIdentity.walletAddress
        );
        const getAppSessionsResponse = await appWS.sendAndWaitForResponse(
            getAppSessionsMsg,
            getGetAppSessionsPredicate(),
            1000
        );

        const getAppSessionsParsedResponse = rpcResponseParser.getAppSessions(getAppSessionsResponse);
        expect(getAppSessionsParsedResponse).toBeDefined();
        expect(getAppSessionsParsedResponse.params).toHaveLength(1);
        
        const appSession = getAppSessionsParsedResponse.params[0];
        expect(appSession.appSessionId).toBe(appSessionId);
        expect(appSession.sessionData).toBeDefined();
        
        return {
            appSession,
            sessionData: JSON.parse(appSession.sessionData!)
        };
    };

    beforeAll(async () => {
        blockUtils = new BlockchainUtils();
        databaseUtils = new DatabaseUtils();

        // Here we need to simulate difference between channel and app session
        identity = new Identity(CONFIG.IDENTITIES[0].WALLET_PK, CONFIG.IDENTITIES[0].SESSION_PK);
        ws = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
        client = new TestNitroliteClient(identity);

        appIdentity = new Identity(CONFIG.IDENTITIES[0].WALLET_PK, CONFIG.IDENTITIES[0].APP_SESSION_PK);
        appWS = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);

        cpWS = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
        cpIdentity = new Identity(CONFIG.IDENTITIES[1].WALLET_PK, CONFIG.IDENTITIES[1].SESSION_PK);
        appCPIdentity = new Identity(CONFIG.IDENTITIES[1].WALLET_PK, CONFIG.IDENTITIES[1].APP_SESSION_PK);
        cpClient = new TestNitroliteClient(cpIdentity);

        await ws.connect();
        await appWS.connect();
        await cpWS.connect();

        await createAuthSessionWithClearnode(ws, identity);
        await createAuthSessionWithClearnode(cpWS, cpIdentity);
        await blockUtils.makeSnapshot();
    });

    afterAll(async () => {
        ws.close();
        appWS.close();
        cpWS.close();

        await databaseUtils.cleanupDatabaseData();
        await blockUtils.resetSnapshot();

        await databaseUtils.close();
    });

    it('should create and init two channels', async () => {
        const { params } = await client.createAndWaitForChannel(ws, {
            tokenAddress: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            amount: depositAmount * BigInt(10), // 10 times the deposit amount
        });

        channelId = params.channelId;

        const { params: cpParams } = await cpClient.createAndWaitForChannel(cpWS, {
            tokenAddress: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            amount: depositAmount * BigInt(10), // 10 times the deposit amount
        });

        cpChannelId = cpParams.channelId;
    });

    it('should create app session with allowance for participant to deposit', async () => {
        await createAuthSessionWithClearnode(appWS, appIdentity, {
            wallet: appIdentity.walletAddress,
            participant: appIdentity.sessionAddress,
            app_name: 'App Domain',
            expire: String(Math.floor(Date.now() / 1000) + 3600), // 1 hour expiration
            scope: 'console',
            application: '0x9965507D1a55bcC2695C58ba16FB37d819B0A4dc', // random address, no use for now
            allowances: [
                {
                    asset: 'usdc',
                    amount: decimalDepositAmount.toString(),
                },
            ],
        });
    });

    it('should take snapshot of ledger balances', async () => {
        const getLedgerBalancesMsg = await createGetLedgerBalancesMessage(
            appIdentity.messageSigner,
            appIdentity.walletAddress
        );
        const getLedgerBalancesResponse = await appWS.sendAndWaitForResponse(
            getLedgerBalancesMsg,
            getGetLedgerBalancesPredicate(),
            1000
        );

        const getLedgerBalancesParsedResponse = rpcResponseParser.getLedgerBalances(getLedgerBalancesResponse);
        expect(getLedgerBalancesParsedResponse).toBeDefined();
        expect(getLedgerBalancesParsedResponse.params).toHaveLength(1);
        expect(getLedgerBalancesParsedResponse.params).toHaveLength(1);
        expect(getLedgerBalancesParsedResponse.params[0].amount).toBe(
            (decimalDepositAmount * BigInt(10)).toString()
        );
        expect(getLedgerBalancesParsedResponse.params[0].asset).toBe('USDC');
    });

    it('should create app session', async () => {
        const definition: AppDefinition = {
            protocol: 'nitroliterpc',
            participants: [appIdentity.walletAddress, appCPIdentity.walletAddress],
            weights: [100, 0],
            quorum: 100,
            challenge: 0,
            nonce: Date.now(),
        };

        const allocations = [
            {
                participant: appIdentity.walletAddress,
                asset: 'USDC',
                amount: decimalDepositAmount.toString(),
            },
            {
                participant: appCPIdentity.walletAddress,
                asset: 'USDC',
                amount: '0',
            },
        ];

        const createAppSessionMsg = await createAppSessionMessage(appIdentity.messageSigner, [
            {
                definition,
                allocations,
                session_data: JSON.stringify({
                    gameType: "chess",
                    timeControl: { initial: 600, increment: 5 },
                    gameState: "waiting"
                })
            },
        ]);
        const createAppSessionResponse = await appWS.sendAndWaitForResponse(
            createAppSessionMsg,
            getCreateAppSessionPredicate(),
            1000
        );

        const createAppSessionParsedResponse = rpcResponseParser.createAppSession(createAppSessionResponse);

        expect(createAppSessionParsedResponse).toBeDefined();
        expect(createAppSessionParsedResponse.params.appSessionId).toBeDefined();
        expect(createAppSessionParsedResponse.params.status).toBe(RPCChannelStatus.Open);
        expect(createAppSessionParsedResponse.params.version).toBeDefined();

        appSessionId = createAppSessionParsedResponse.params.appSessionId;
    });

    it('should submit state with updated session_data', async () => {
        const submitStateMsg = await createSubmitStateMessage(appIdentity.messageSigner, [
            {
                app_session_id: appSessionId as Hex,
                allocations: [
                    {
                        participant: appIdentity.walletAddress,
                        asset: 'USDC',
                        amount: (decimalDepositAmount / BigInt(2)).toString(), // 50 USDC
                    },
                    {
                        participant: appCPIdentity.walletAddress,
                        asset: 'USDC',
                        amount: (decimalDepositAmount / BigInt(2)).toString(), // 50 USDC
                    },
                ],
                session_data: JSON.stringify({
                    gameType: "chess",
                    timeControl: { initial: 600, increment: 5 },
                    gameState: "active",
                    currentMove: "e2e4",
                    moveCount: 1
                })
            },
        ]);

        const submitStateResponse = await appWS.sendAndWaitForResponse(
            submitStateMsg,
            getSubmitStatePredicate(),
            1000
        );

        const submitStateParsedResponse = rpcResponseParser.submitState(submitStateResponse);
        expect(submitStateParsedResponse).toBeDefined();
        expect(submitStateParsedResponse.params.appSessionId).toBe(appSessionId);
        expect(submitStateParsedResponse.params.status).toBe(RPCChannelStatus.Open);
        expect(submitStateParsedResponse.params.version).toBe(2);
    });

    it('should verify sessionData changes after updates', async () => {
        const { sessionData } = await fetchAndParseAppSessions();
        
        expect(sessionData.gameState).toBe('active');
        expect(sessionData.currentMove).toBe('e2e4');
        expect(sessionData.moveCount).toBe(1);
        expect(sessionData.gameType).toBe('chess');
        expect(sessionData.timeControl).toEqual({ initial: 600, increment: 5 });
    });

    it('should close app session', async () => {
        const closeAppSessionMsg = await createCloseAppSessionMessage(appIdentity.messageSigner, [
            {
                app_session_id: appSessionId as Hex,
                allocations: [
                    {
                        participant: appIdentity.walletAddress,
                        asset: 'USDC',
                        amount: '0',
                    },
                    {
                        participant: appCPIdentity.walletAddress,
                        asset: 'USDC',
                        amount: decimalDepositAmount.toString(),
                    },
                ],
                session_data: JSON.stringify({
                    gameType: "chess",
                    timeControl: { initial: 600, increment: 5 },
                    gameState: "finished",
                    winner: "white",
                    endCondition: "checkmate"
                })
            },
        ]);

        const closeAppSessionResponse = await appWS.sendAndWaitForResponse(
            closeAppSessionMsg,
            getCloseAppSessionPredicate(),
            1000
        );
        expect(closeAppSessionResponse).toBeDefined();

        const closeAppSessionParsedResponse = rpcResponseParser.closeAppSession(closeAppSessionResponse);
        expect(closeAppSessionParsedResponse).toBeDefined();

        expect(closeAppSessionParsedResponse.params.appSessionId).toBe(appSessionId);
        expect(closeAppSessionParsedResponse.params.status).toBe(RPCChannelStatus.Closed);
        expect(closeAppSessionParsedResponse.params.version).toBe(3);
    });

    it('should verify sessionData changes after closing', async () => {
        const { appSession, sessionData } = await fetchAndParseAppSessions();
        
        expect(appSession.status).toBe(RPCChannelStatus.Closed);
        expect(sessionData.gameState).toBe('finished');
        expect(sessionData.winner).toBe('white');
        expect(sessionData.endCondition).toBe('checkmate');
        expect(sessionData.gameType).toBe('chess');
        expect(sessionData.timeControl).toEqual({ initial: 600, increment: 5 });
    });

    it('should update ledger balances', async () => {
        const getLedgerBalancesMsg = await createGetLedgerBalancesMessage(
            appIdentity.messageSigner,
            appIdentity.walletAddress
        );
        const getLedgerBalancesResponse = await appWS.sendAndWaitForResponse(
            getLedgerBalancesMsg,
            getGetLedgerBalancesPredicate(),
            1000
        );

        const getLedgerBalancesParsedResponse = rpcResponseParser.getLedgerBalances(getLedgerBalancesResponse);
        expect(getLedgerBalancesParsedResponse).toBeDefined();
        expect(getLedgerBalancesParsedResponse.params).toHaveLength(1);
        expect(getLedgerBalancesParsedResponse.params[0].amount).toBe((decimalDepositAmount * BigInt(9)).toString());
        expect(getLedgerBalancesParsedResponse.params[0].asset).toBe('USDC');
    });

    // TODO: fix multiple ws connection and add resize
    // it('should close and withdraw both channels', async () => {
    //     // TODO: connect to ws to overwrite app ws session
    //     await createAuthSessionWithClearnode(ws, identity);

    //     await client.closeAndWithdrawChannel(ws, channelId);
    //     await cpClient.closeAndWithdrawChannel(cpWS, cpChannelId);
    // });
});
