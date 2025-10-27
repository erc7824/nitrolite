import { BlockchainUtils } from '@/blockchainUtils';
import { DatabaseUtils } from '@/databaseUtils';
import { Identity } from '@/identity';
import { TestNitroliteClient } from '@/nitroliteClient';
import { CONFIG } from '@/setup';
import {
    TestWebSocket,
    getResizeChannelPredicate,
    getCloseChannelPredicate,
} from '@/ws';
import {
    createCloseChannelMessage,
    createResizeChannelMessage,
    parseCloseChannelResponse,
    parseResizeChannelResponse,
    RPCProtocolVersion,
} from '@erc7824/nitrolite';
import { Hex } from 'viem';
import {
    setupTestIdentitiesAndConnections,
    fetchAndParseAppSessions,
} from '@/testSetup';
import {
    createTestChannels,
    authenticateAppWithAllowances,
    createTestAppSession,
    getLedgerBalances,
    toRaw,
} from '@/testHelpers';
import {
    submitAppStateUpdate_v02,
    closeAppSessionWithState,
} from '@/testAppSessionHelpers';

describe('nitrorpc_v02 lifecycle', () => {
    const ASSET_SYMBOL = CONFIG.TOKEN_SYMBOL;

    const onChainDepositAmount = BigInt(1000);
    const appSessionDepositAmount = BigInt(100);

    let aliceWS: TestWebSocket;
    let alice: Identity;
    let aliceClient: TestNitroliteClient;

    let aliceAppWS: TestWebSocket;
    let aliceAppIdentity: Identity;

    let bobWS: TestWebSocket;
    let bob: Identity;
    let bobAppIdentity: Identity;
    let bobClient: TestNitroliteClient;

    let blockUtils: BlockchainUtils;
    let databaseUtils: DatabaseUtils;

    let aliceChannelId: Hex;
    let bobChannelId: Hex;
    let appSessionId: string;

    const GAME_TYPE = 'chess';
    const TIME_CONTROL = { initial: 600, increment: 5 };

    const SESSION_DATA_WAITING = {
        gameType: GAME_TYPE,
        timeControl: TIME_CONTROL,
        gameState: 'waiting',
    };

    const SESSION_DATA_ACTIVE = {
        ...SESSION_DATA_WAITING,
        gameState: 'active',
        currentMove: 'e2e4',
        moveCount: 1,
    };

    const SESSION_DATA_FINISHED = {
        ...SESSION_DATA_WAITING,
        gameState: 'finished',
        winner: 'white',
        endCondition: 'checkmate',
    };


    beforeAll(async () => {
        blockUtils = new BlockchainUtils();
        databaseUtils = new DatabaseUtils();

        const setup = await setupTestIdentitiesAndConnections();
        alice = setup.alice;
        aliceWS = setup.aliceWS;
        aliceClient = setup.aliceClient;
        aliceAppIdentity = setup.aliceAppIdentity;
        aliceAppWS = setup.aliceAppWS;
        bob = setup.bob;
        bobAppIdentity = setup.bobAppIdentity;
        bobWS = setup.bobWS;
        bobClient = setup.bobClient;

        await blockUtils.makeSnapshot();
    });

    afterAll(async () => {
        aliceWS.close();
        aliceAppWS.close();
        bobWS.close();

        await databaseUtils.resetClearnodeState();
        await blockUtils.resetSnapshot();

        await databaseUtils.close();
    });

    it('should create and init two channels', async () => {
        [aliceChannelId, bobChannelId] = await createTestChannels([{client: aliceClient, ws: aliceWS}, {client: bobClient, ws: bobWS}], toRaw(onChainDepositAmount));
    });

    it('should create app session with allowance for participant to deposit', async () => {
        await authenticateAppWithAllowances(aliceAppWS, aliceAppIdentity, ASSET_SYMBOL, appSessionDepositAmount);
    });

    it('should take snapshot of ledger balances', async () => {
        const ledgerBalances = await getLedgerBalances(aliceAppIdentity, aliceAppWS);
        expect(ledgerBalances).toHaveLength(1);
        expect(ledgerBalances[0].amount).toBe((onChainDepositAmount).toString());
        expect(ledgerBalances[0].asset).toBe(ASSET_SYMBOL);
    });

    it('should create app session', async () => {
        appSessionId = await createTestAppSession(
            aliceAppIdentity,
            bobAppIdentity,
            aliceAppWS,
            RPCProtocolVersion.NitroRPC_0_2,
            ASSET_SYMBOL,
            appSessionDepositAmount,
            SESSION_DATA_WAITING
        );
    });

    it('should submit state with updated session_data', async () => {
        const allocations = [
            {
                participant: aliceAppIdentity.walletAddress,
                asset:  ASSET_SYMBOL,
                amount: (appSessionDepositAmount / BigInt(2)).toString(), // 50 USDC
            },
            {
                participant: bobAppIdentity.walletAddress,
                asset:  ASSET_SYMBOL,
                amount: (appSessionDepositAmount / BigInt(2)).toString(), // 50 USDC
            },
        ];

        await submitAppStateUpdate_v02(aliceAppWS, aliceAppIdentity, appSessionId, allocations, SESSION_DATA_ACTIVE, 2);
    });

    it('should verify sessionData changes after updates', async () => {
        const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);

        expect(sessionData).toEqual(SESSION_DATA_ACTIVE);
    });

    it('should close app session', async () => {
        const allocations = [
            {
                participant: aliceAppIdentity.walletAddress,
                asset: ASSET_SYMBOL,
                amount: '0',
            },
            {
                participant: bobAppIdentity.walletAddress,
                asset: ASSET_SYMBOL,
                amount: appSessionDepositAmount.toString(),
            },
        ];

        await closeAppSessionWithState(aliceAppWS, aliceAppIdentity, appSessionId, allocations, SESSION_DATA_FINISHED, 3);
    });

    it('should verify sessionData changes after closing', async () => {
        const { appSession, sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);

        expect(appSession.status).toBe('closed');
        expect(sessionData).toEqual(SESSION_DATA_FINISHED);
    });

    it('should update ledger balances for providing side', async () => {
        const ledgerBalances = await getLedgerBalances(aliceAppIdentity, aliceAppWS);
        expect(ledgerBalances).toHaveLength(1);
        expect(ledgerBalances[0].amount).toBe((appSessionDepositAmount * BigInt(9)).toString()); // 1000 - 100
        expect(ledgerBalances[0].asset).toBe(ASSET_SYMBOL);
    });

    it('should update ledger balances for receiving side', async () => {
        const ledgerBalances = await getLedgerBalances(bobAppIdentity, bobWS);
        expect(ledgerBalances).toHaveLength(1);
        expect(ledgerBalances[0].amount).toBe((appSessionDepositAmount * BigInt(11)).toString()); // 1000 + 100
        expect(ledgerBalances[0].asset).toBe(ASSET_SYMBOL);
    });

    it('should close channel and withdraw without app funds', async () => {
        const msg = await createCloseChannelMessage(alice.messageSigner, aliceChannelId, alice.walletAddress);

        const closeResponse = await aliceWS.sendAndWaitForResponse(msg, getCloseChannelPredicate(), 1000);
        expect(closeResponse).toBeDefined();

        const { params: closeResponseParams } = parseCloseChannelResponse(closeResponse);
        const closeChannelTxHash = await aliceClient.closeChannel({
            finalState: {
                intent: closeResponseParams.state.intent,
                channelId: closeResponseParams.channelId,
                data: closeResponseParams.state.stateData as Hex,
                allocations: closeResponseParams.state.allocations,
                version: BigInt(closeResponseParams.state.version),
                serverSignature: closeResponseParams.serverSignature,
            },
            stateData: closeResponseParams.state.stateData as Hex,
        });
        expect(closeChannelTxHash).toBeDefined();

        const closeReceipt = await blockUtils.waitForTransaction(closeChannelTxHash);
        expect(closeReceipt).toBeDefined();

        const postCloseAccountBalance = await aliceClient.getAccountBalance(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
        expect(postCloseAccountBalance).toBe(toRaw(onChainDepositAmount - appSessionDepositAmount)); // 1000 - 100
    });

    it('should resize channel by withdrawing received funds from app to channel', async () => {
        const msg = await createResizeChannelMessage(bob.messageSigner, {
            channel_id: bobChannelId,
            allocate_amount: toRaw(appSessionDepositAmount),
            funds_destination: bob.walletAddress,
        });

        const resizeResponse = await bobWS.sendAndWaitForResponse(msg, getResizeChannelPredicate(), 1000);
        const { params: resizeResponseParams } = parseResizeChannelResponse(resizeResponse);

        expect(resizeResponseParams.state.allocations).toBeDefined();
        expect(resizeResponseParams.state.allocations).toHaveLength(2);
        expect(String(resizeResponseParams.state.allocations[0].destination)).toBe(bob.walletAddress);
        expect(String(resizeResponseParams.state.allocations[0].amount)).toBe(
            (toRaw(onChainDepositAmount + appSessionDepositAmount)).toString() // 1000 + 100
        );
        expect(String(resizeResponseParams.state.allocations[1].destination)).toBe(CONFIG.ADDRESSES.GUEST_ADDRESS);
        expect(String(resizeResponseParams.state.allocations[1].amount)).toBe('0');

        const resizeChannelTxHash = await bobClient.resizeChannel({
            resizeState: {
                channelId: resizeResponseParams.channelId as Hex,
                intent: resizeResponseParams.state.intent,
                version: BigInt(resizeResponseParams.state.version),
                data: resizeResponseParams.state.stateData as Hex,
                allocations: resizeResponseParams.state.allocations,
                serverSignature: resizeResponseParams.serverSignature,
            },
            proofStates: [
                // NOTE: Dummy adjudicator doesn't validate proofs, so we can pass any valid (from Custody POV) state
                {
                    intent: 1, // StateIntent.INITIALIZE
                    version: BigInt(0),
                    data: '0x',
                    allocations: [
                        {
                            destination: bob.walletAddress,
                            token: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
                            amount: toRaw(onChainDepositAmount)
                        },
                        {
                            destination: CONFIG.ADDRESSES.GUEST_ADDRESS,
                            token: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
                            amount: BigInt(0),
                        },
                    ],
                    sigs: [],
                },
            ],
        });
        expect(resizeChannelTxHash).toBeDefined();

        const resizeReceipt = await blockUtils.waitForTransaction(resizeChannelTxHash);
        expect(resizeReceipt).toBeDefined();
    });

    it('should close channel and withdraw with app funds', async () => {
        const msg = await createCloseChannelMessage(bob.messageSigner, bobChannelId, bob.walletAddress);

        const closeResponse = await bobWS.sendAndWaitForResponse(msg, getCloseChannelPredicate(), 1000);
        expect(closeResponse).toBeDefined();

        const { params: closeResponseParams } = parseCloseChannelResponse(closeResponse);
        const closeChannelTxHash = await bobClient.closeChannel({
            finalState: {
                intent: closeResponseParams.state.intent,
                channelId: closeResponseParams.channelId,
                data: closeResponseParams.state.stateData as Hex,
                allocations: closeResponseParams.state.allocations,
                version: BigInt(closeResponseParams.state.version),
                serverSignature: closeResponseParams.serverSignature,
            },
            stateData: closeResponseParams.state.stateData as Hex,
        });
        expect(closeChannelTxHash).toBeDefined();

        const closeReceipt = await blockUtils.waitForTransaction(closeChannelTxHash);
        expect(closeReceipt).toBeDefined();

        const postCloseAccountBalance = await bobClient.getAccountBalance(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
        expect(postCloseAccountBalance).toBe(toRaw(onChainDepositAmount + appSessionDepositAmount)); // 1000 + 100
    });

    it('should withdraw funds from channel for providing side', async () => {
        const preWithdrawalBalance = await blockUtils.getErc20Balance(
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            alice.walletAddress
        );

        const withdrawalTxHash = await aliceClient.withdrawal(
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            toRaw(onChainDepositAmount - appSessionDepositAmount) // 1000 - 100
        );
        expect(withdrawalTxHash).toBeDefined();

        const withdrawalReceipt = await blockUtils.waitForTransaction(withdrawalTxHash);
        expect(withdrawalReceipt).toBeDefined();

        const postWithdrawalBalance = await blockUtils.getErc20Balance(
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            alice.walletAddress
        );
        expect(postWithdrawalBalance.rawBalance - preWithdrawalBalance.rawBalance).toBe(toRaw(onChainDepositAmount - appSessionDepositAmount)); // + 900

        const accountBalance = await aliceClient.getAccountBalance(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
        expect(accountBalance).toBe(BigInt(0));
    });

    it('should withdraw funds from channel for receiving side', async () => {
        const preWithdrawalBalance = await blockUtils.getErc20Balance(
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            bob.walletAddress
        );

        const withdrawalTxHash = await bobClient.withdrawal(
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            toRaw(onChainDepositAmount + appSessionDepositAmount) // 1000 + 100
        );
        expect(withdrawalTxHash).toBeDefined();

        const withdrawalReceipt = await blockUtils.waitForTransaction(withdrawalTxHash);
        expect(withdrawalReceipt).toBeDefined();

        const postWithdrawalBalance = await blockUtils.getErc20Balance(
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            bob.walletAddress
        );
        expect(postWithdrawalBalance.rawBalance - preWithdrawalBalance.rawBalance).toBe(toRaw(onChainDepositAmount + appSessionDepositAmount)); // + 1100

        const accountBalance = await bobClient.getAccountBalance(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
        expect(accountBalance).toBe(BigInt(0));
    });
});
