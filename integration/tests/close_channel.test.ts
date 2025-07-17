import { createAuthSessionWithClearnode } from '@/auth';
import { BlockchainUtils } from '@/blockchainUtils';
import { DatabaseUtils } from '@/databaseUtils';
import { Identity } from '@/identity';
import { TestNitroliteClient } from '@/nitroliteClient';
import { CONFIG } from '@/setup';
import { getCloseChannelPredicate, TestWebSocket } from '@/ws';
import { createCloseChannelMessage, parseCloseChannelResponse } from '@erc7824/nitrolite';
import { Hex, parseUnits } from 'viem';

describe('Close channel', () => {
    const depositAmount = parseUnits('100', 6); // 100 USDC (decimals = 6)

    let ws: TestWebSocket;
    let identity: Identity;
    let client: TestNitroliteClient;
    let blockUtils: BlockchainUtils;
    let databaseUtils: DatabaseUtils;

    beforeAll(async () => {
        blockUtils = new BlockchainUtils();
        databaseUtils = new DatabaseUtils();
        identity = new Identity(CONFIG.IDENTITIES[0].WALLET_PK, CONFIG.IDENTITIES[0].SESSION_PK);
        ws = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
    });

    beforeEach(async () => {
        await ws.connect();
        await createAuthSessionWithClearnode(ws, identity);
        await blockUtils.makeSnapshot();
    });

    afterEach(async () => {
        ws.close();
        await databaseUtils.cleanupDatabaseData();
        await blockUtils.resetSnapshot();
    });

    afterAll(() => {
        databaseUtils.close();
    });

    it('should create nitrolite client to close channels', async () => {
        client = new TestNitroliteClient(identity);

        expect(client).toBeDefined();
        expect(client).toHaveProperty('depositAndCreateChannel');
        expect(client).toHaveProperty('closeChannel');
        expect(client).toHaveProperty('withdrawal');
    });

    it('should close channel and withdraw funds', async () => {
        const preFundBalance = await blockUtils.getErc20Balance(
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            identity.walletAddress
        );

        const { params } = await client.createAndWaitForChannel(ws, {
            tokenAddress: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            amount: depositAmount,
        });

        const postFundBalance = await blockUtils.getErc20Balance(
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            identity.walletAddress
        );

        expect(postFundBalance.rawBalance).toBe(preFundBalance.rawBalance - depositAmount);

        const msg = await createCloseChannelMessage(identity.messageSigner, params.channelId, identity.walletAddress);
        const closeResponse = await ws.sendAndWaitForResponse(msg, getCloseChannelPredicate(), 1000);
        expect(closeResponse).toBeDefined();

        const closeParsedResponse = parseCloseChannelResponse(closeResponse);

        const closeChannelTxHash = await client.closeChannel({
            finalState: {
                intent: closeParsedResponse.params.state.intent,
                channelId: closeParsedResponse.params.channelId,
                data: closeParsedResponse.params.state.stateData as Hex,
                allocations: closeParsedResponse.params.state.allocations,
                version: BigInt(closeParsedResponse.params.state.version),
                serverSignature: closeParsedResponse.params.serverSignature,
            },
            stateData: closeParsedResponse.params.state.stateData as Hex,
        });
        expect(closeChannelTxHash).toBeDefined();

        const closeReceipt = await blockUtils.waitForTransaction(closeChannelTxHash);
        expect(closeReceipt).toBeDefined();

        // Close should not change wallet balance
        expect(postFundBalance.rawBalance).toBe(preFundBalance.rawBalance - depositAmount);

        const withdrawalTxHash = await client.withdrawal(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS, depositAmount);
        expect(withdrawalTxHash).toBeDefined();

        const withdrawalReceipt = await blockUtils.waitForTransaction(withdrawalTxHash);
        expect(withdrawalReceipt).toBeDefined();

        const postWithdrawalBalance = await blockUtils.getErc20Balance(
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            identity.walletAddress
        );
        expect(postWithdrawalBalance.rawBalance).toBe(preFundBalance.rawBalance);
    });
});
