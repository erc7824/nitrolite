import { BlockchainUtils } from '@/blockchainUtils';
import { DatabaseUtils } from '@/databaseUtils';
import { Identity } from '@/identity';
import { TestNitroliteClient } from '@/nitroliteClient';
import { TestWebSocket } from '@/ws';
import { RPCProtocolVersion, State } from '@erc7824/nitrolite';
import { Hex } from 'viem';
import { setupTestIdentitiesAndConnections } from '@/testSetup';
import {
    createTestChannels,
    authenticateAppWithAllowances,
    createTestAppSession,
    toRaw,
} from '@/testHelpers';
import { CONFIG } from '@/setup';

describe('App session v0.2', () => {
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

    let initialStates: State[];

    const START_SESSION_DATA = { gameType: 'chess', gameState: 'waiting' };

    let START_ALLOCATIONS;

    beforeAll(async () => {
        blockUtils = new BlockchainUtils();
        databaseUtils = new DatabaseUtils();

        ({alice, aliceWS, aliceClient, aliceAppIdentity, aliceAppWS, bob, bobWS, bobClient, bobAppIdentity} = await setupTestIdentitiesAndConnections());

        START_ALLOCATIONS = [
            {
                participant: aliceAppIdentity.walletAddress,
                asset: ASSET_SYMBOL,
                amount: (appSessionDepositAmount).toString(),
            },
            {
                participant: bobAppIdentity.walletAddress,
                asset: ASSET_SYMBOL,
                amount: '0',
            },
        ];
    });

    beforeEach(async () => {
        await blockUtils.makeSnapshot();
        ({channelIds: [aliceChannelId, bobChannelId], states: initialStates} = await createTestChannels([{client: aliceClient, ws: aliceWS}, {client: bobClient, ws: bobWS}], toRaw(onChainDepositAmount)));

        await authenticateAppWithAllowances(aliceAppWS, aliceAppIdentity, ASSET_SYMBOL, appSessionDepositAmount.toString());
    });

    afterEach(async () => {
        await blockUtils.resetSnapshot();
        await databaseUtils.resetClearnodeState();
    });

    afterAll(async () => {
        aliceWS.close();
        aliceAppWS.close();
        bobWS.close();

        await databaseUtils.close();
    });

    describe('session creation', () => {
        it('rejects to create app session if one of participants has non-empty channel', async () => {
            const resizeAmount = BigInt(1);
            await aliceClient.resizeChannelAndWait(
                aliceWS,
                aliceChannelId,
                initialStates[0],
                alice.walletAddress,
                BigInt(0), // resize only to the user, so that a channel contains on-chain funds
                resizeAmount
            );

            const balance = await aliceClient.getChannelBalance(aliceChannelId, CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
            expect(balance).toBe(resizeAmount);

            try {
                await createTestAppSession(
                    aliceAppIdentity,
                    bobAppIdentity,
                    aliceAppWS,
                    RPCProtocolVersion.NitroRPC_0_2,
                    ASSET_SYMBOL,
                    appSessionDepositAmount.toString(),
                    START_SESSION_DATA
                );
            } catch (e) {
                expect((e as Error).message).toMatch(/RPC Error.*operation denied.*non-zero allocation.*detected/i);
                return;
            }

            throw new Error('App session creation was not rejected as expected.');
        });
    });
});
