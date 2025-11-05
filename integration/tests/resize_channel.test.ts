import { createAuthSessionWithClearnode } from '@/auth';
import { BlockchainUtils } from '@/blockchainUtils';
import { DatabaseUtils } from '@/databaseUtils';
import { Identity } from '@/identity';
import { TestNitroliteClient } from '@/nitroliteClient';
import { CONFIG } from '@/setup';
import { composeResizeChannelParams, getLedgerBalances, toRaw } from '@/testHelpers';
import { getChannelUpdatePredicateWithStatus, getResizeChannelPredicate, TestWebSocket } from '@/ws';
import { createResizeChannelMessage, parseResizeChannelResponse, RPCChannelStatus } from '@erc7824/nitrolite';
import { Hex, parseUnits } from 'viem';

describe('Resize channel', () => {
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
        await databaseUtils.resetClearnodeState();
        await blockUtils.resetSnapshot();
    });

    afterAll(async () => {
        await databaseUtils.close();
    });

    it('should create nitrolite client to resize channels', async () => {
        client = new TestNitroliteClient(identity);

        expect(client).toBeDefined();
        expect(client).toHaveProperty('depositAndCreateChannel');
        expect(client).toHaveProperty('resizeChannel');
    });

    it('should resize channel by adding funds from deposit to channel', async () => {
        const { params: createResponseParams, state: createResponseState } = await client.createAndWaitForChannel(ws, {
            tokenAddress: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            amount: depositAmount * BigInt(5),
            depositAmount: depositAmount * BigInt(10), // depositing more than initial amount to have resize buffer
        });

        const channelBalance = await client.getChannelBalance(
            createResponseParams.channelId,
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS
        );
        expect(channelBalance).toBe(depositAmount * BigInt(5)); // 500

        const preResizeAccountBalance = await client.getAccountBalance(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
        expect(preResizeAccountBalance).toBe(depositAmount * BigInt(5)); // 1000 - 500

        const preResizeChannelBalance = await client.getChannelBalance(
            createResponseParams.channelId,
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS
        );
        expect(preResizeChannelBalance).toBe(depositAmount * BigInt(5)); // 500

        const msg = await createResizeChannelMessage(identity.messageSKSigner, {
            channel_id: createResponseParams.channelId,
            resize_amount: depositAmount,
            allocate_amount: parseUnits('0', 6),
            funds_destination: identity.walletAddress,
        });

        const resizeResponse = await ws.sendAndWaitForResponse(msg, getResizeChannelPredicate(), 1000);
        const { params: resizeResponseParams } = parseResizeChannelResponse(resizeResponse);
        expect(resizeResponseParams.channelId).toBe(createResponseParams.channelId);
        expect(resizeResponseParams.state.stateData).toBeDefined();
        expect(resizeResponseParams.state.intent).toBe(2); // StateIntent.RESIZE // TODO: add enum to sdk
        expect(resizeResponseParams.state.version).toBe(Number(createResponseState.version) + 1);

        expect(resizeResponseParams.serverSignature).toBeDefined();

        expect(resizeResponseParams.state.allocations).toBeDefined();
        expect(resizeResponseParams.state.allocations).toHaveLength(2);
        expect(String(resizeResponseParams.state.allocations[0].destination)).toBe(identity.walletAddress);
        expect(String(resizeResponseParams.state.allocations[0].token)).toBe(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
        expect(String(resizeResponseParams.state.allocations[0].amount)).toBe(
            (depositAmount * BigInt(6)).toString() // 500 + 100
        );
        expect(String(resizeResponseParams.state.allocations[1].destination)).toBe(CONFIG.ADDRESSES.GUEST_ADDRESS);
        expect(String(resizeResponseParams.state.allocations[1].token)).toBe(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
        expect(String(resizeResponseParams.state.allocations[1].amount)).toBe('0');

        const {txHash: resizeChannelTxHash} = await client.resizeChannel({
            ...composeResizeChannelParams(
                resizeResponseParams.channelId as Hex,
                resizeResponseParams,
                createResponseState
            ),
        });
        expect(resizeChannelTxHash).toBeDefined();

        const resizeReceipt = await blockUtils.waitForTransaction(resizeChannelTxHash);
        expect(resizeReceipt).toBeDefined();

        const postResizeAccountBalance = await client.getAccountBalance(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
        expect(postResizeAccountBalance).toBe(depositAmount * BigInt(4)); // 1000 - 500 - 100

        const postResizeChannelBalance = await client.getChannelBalance(
            createResponseParams.channelId,
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS
        );
        expect(postResizeChannelBalance).toBe(depositAmount * BigInt(6)); // 500 + 100
    });

    it('should resize channel by withdrawing funds from channel to deposit', async () => {
        const { params: createResponseParams, state: createResponseState } = await client.createAndWaitForChannel(ws, {
            tokenAddress: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            amount: depositAmount * BigInt(5),
            depositAmount: depositAmount * BigInt(10), // depositing more than initial amount to have resize buffer
        });

        const preResizeAccountBalance = await client.getAccountBalance(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
        expect(preResizeAccountBalance).toBe(depositAmount * BigInt(5)); // 1000 - 500

        const preResizeChannelBalance = await client.getChannelBalance(
            createResponseParams.channelId,
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS
        );
        expect(preResizeChannelBalance).toBe(depositAmount * BigInt(5)); // 500

        const msg = await createResizeChannelMessage(identity.messageSKSigner, {
            channel_id: createResponseParams.channelId,
            resize_amount: -depositAmount,
            allocate_amount: parseUnits('0', 6),
            funds_destination: identity.walletAddress,
        });

        const resizeResponse = await ws.sendAndWaitForResponse(msg, getResizeChannelPredicate(), 1000);

        const { params: resizeResponseParams } = parseResizeChannelResponse(resizeResponse);
        expect(resizeResponseParams.state.allocations).toBeDefined();
        expect(resizeResponseParams.state.allocations).toHaveLength(2);
        expect(String(resizeResponseParams.state.allocations[0].destination)).toBe(identity.walletAddress);
        expect(String(resizeResponseParams.state.allocations[0].amount)).toBe(
            (depositAmount * BigInt(4)).toString() // 500 - 100
        );
        expect(String(resizeResponseParams.state.allocations[1].destination)).toBe(CONFIG.ADDRESSES.GUEST_ADDRESS);
        expect(String(resizeResponseParams.state.allocations[1].amount)).toBe('0');

        const {txHash: resizeChannelTxHash} = await client.resizeChannel({
            ...composeResizeChannelParams(
                resizeResponseParams.channelId as Hex,
                resizeResponseParams,
                createResponseState
            ),
        });
        expect(resizeChannelTxHash).toBeDefined();

        const resizeReceipt = await blockUtils.waitForTransaction(resizeChannelTxHash);
        expect(resizeReceipt).toBeDefined();

        const postResizeAccountBalance = await client.getAccountBalance(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
        expect(postResizeAccountBalance).toBe(depositAmount * BigInt(6)); // 1000 - 500 + 100

        const postResizeChannelBalance = await client.getChannelBalance(
            createResponseParams.channelId,
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS
        );
        expect(postResizeChannelBalance).toBe(depositAmount * BigInt(4)); // 500 - 100
    });

    it('should resize channel by allocating funds from channel to virtual ledger', async () => {
        const { params: createResponseParams, state: createResponseState } = await client.createAndWaitForChannel(ws, {
            tokenAddress: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            amount: depositAmount * BigInt(5),
            depositAmount: depositAmount * BigInt(10), // depositing more than initial amount to have resize buffer
        });

        const preResizeAccountBalance = await client.getAccountBalance(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
        expect(preResizeAccountBalance).toBe(depositAmount * BigInt(5)); // 1000 - 500

        const preResizeChannelBalance = await client.getChannelBalance(
            createResponseParams.channelId,
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS
        );
        expect(preResizeChannelBalance).toBe(depositAmount * BigInt(5)); // 500

        const msg = await createResizeChannelMessage(identity.messageSKSigner, {
            channel_id: createResponseParams.channelId,
            resize_amount: parseUnits('0', 6),
            allocate_amount: -depositAmount,
            funds_destination: identity.walletAddress,
        });

        const resizeResponse = await ws.sendAndWaitForResponse(msg, getResizeChannelPredicate(), 1000);
        const { params: resizeResponseParams } = parseResizeChannelResponse(resizeResponse);
        expect(resizeResponseParams.state.allocations).toBeDefined();
        expect(resizeResponseParams.state.allocations).toHaveLength(2);
        expect(String(resizeResponseParams.state.allocations[0].destination)).toBe(identity.walletAddress);
        expect(String(resizeResponseParams.state.allocations[0].amount)).toBe(
            (depositAmount * BigInt(4)).toString() // 500 - 100
        );
        expect(String(resizeResponseParams.state.allocations[1].destination)).toBe(CONFIG.ADDRESSES.GUEST_ADDRESS);
        expect(String(resizeResponseParams.state.allocations[1].amount)).toBe('0');

        const {txHash: resizeChannelTxHash} = await client.resizeChannel({
            ...composeResizeChannelParams(
                resizeResponseParams.channelId as Hex,
                resizeResponseParams,
                createResponseState
            )
        });
        expect(resizeChannelTxHash).toBeDefined();

        const resizeReceipt = await blockUtils.waitForTransaction(resizeChannelTxHash);
        expect(resizeReceipt).toBeDefined();

        const postResizeAccountBalance = await client.getAccountBalance(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
        expect(postResizeAccountBalance).toBe(depositAmount * BigInt(5)); // 1000 - 500

        const postResizeChannelBalance = await client.getChannelBalance(
            createResponseParams.channelId,
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS
        );
        expect(postResizeChannelBalance).toBe(depositAmount * BigInt(4)); // 1000 - 500 - 100
    });

    it('should subtract resize amount from unified balance after withdrawal resize request', async () => {
        const DEPOSIT_AMOUNT = BigInt(10)
        const WITHDRAWAL_AMOUNT = BigInt(1)

        const { params: createResponseParams, state: createResponseState } = await client.createAndWaitForChannel(ws, {
            tokenAddress: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            amount: toRaw(DEPOSIT_AMOUNT)
        });

        const preResizeUnifiedBalance = await getLedgerBalances(identity, ws);
        expect(preResizeUnifiedBalance.length).toBe(1);
        expect(preResizeUnifiedBalance[0].amount).toBe(DEPOSIT_AMOUNT.toString());

        const msg = await createResizeChannelMessage(identity.messageSKSigner, {
            channel_id: createResponseParams.channelId,
            resize_amount: -toRaw(WITHDRAWAL_AMOUNT),
            allocate_amount: -toRaw(WITHDRAWAL_AMOUNT),
            funds_destination: identity.walletAddress,
        });

        const resizeResponse = await ws.sendAndWaitForResponse(msg, getResizeChannelPredicate(), 1000);
        const { params: resizeResponseParams } = parseResizeChannelResponse(resizeResponse);

        // after resize withdrawal is requested, the unified balance should decrease by resize amount
        const postResizeReqUnifiedBalance = await getLedgerBalances(identity, ws);
        expect(postResizeReqUnifiedBalance.length).toBe(1);
        expect(postResizeReqUnifiedBalance[0].amount).toBe((DEPOSIT_AMOUNT - WITHDRAWAL_AMOUNT).toString());

        const {txHash: resizeChannelTxHash} = await client.resizeChannel({
            ...composeResizeChannelParams(
                resizeResponseParams.channelId as Hex,
                resizeResponseParams,
                createResponseState
            ),
        });
        expect(resizeChannelTxHash).toBeDefined();

        const resizeReceipt = await blockUtils.waitForTransaction(resizeChannelTxHash);
        expect(resizeReceipt).toBeDefined();

        const postResizeUnifiedBalance = await getLedgerBalances(identity, ws);
        expect(postResizeUnifiedBalance.length).toBe(1);
        expect(postResizeUnifiedBalance[0].amount).toBe((DEPOSIT_AMOUNT - WITHDRAWAL_AMOUNT).toString());
    });

    it('should NOT subtract resize amount from unified balance after top-up resize request', async () => {
        const DEPOSIT_AMOUNT = BigInt(10)
        const TOP_UP_AMOUNT = BigInt(1)

        const { params: createResponseParams, state: createResponseState } = await client.createAndWaitForChannel(ws, {
            tokenAddress: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            amount: toRaw(DEPOSIT_AMOUNT),
            depositAmount: toRaw(DEPOSIT_AMOUNT + TOP_UP_AMOUNT), // deposit more to have top-up buffer
        });

        const preResizeUnifiedBalance = await getLedgerBalances(identity, ws);
        expect(preResizeUnifiedBalance.length).toBe(1);
        expect(preResizeUnifiedBalance[0].amount).toBe(DEPOSIT_AMOUNT.toString());

        const msg = await createResizeChannelMessage(identity.messageSKSigner, {
            channel_id: createResponseParams.channelId,
            resize_amount: toRaw(TOP_UP_AMOUNT),
            funds_destination: identity.walletAddress,
        });

        const resizeResponse = await ws.sendAndWaitForResponse(msg, getResizeChannelPredicate(), 1000);
        const { params: resizeResponseParams } = parseResizeChannelResponse(resizeResponse);

        // after resize deposit is requested, the unified balance should NOT change
        const postResizeReqUnifiedBalance = await getLedgerBalances(identity, ws);
        expect(postResizeReqUnifiedBalance.length).toBe(1);
        expect(postResizeReqUnifiedBalance[0].amount).toBe(DEPOSIT_AMOUNT.toString());

        const {txHash: resizeChannelTxHash} = await client.resizeChannel({
            ...composeResizeChannelParams(
                resizeResponseParams.channelId as Hex,
                resizeResponseParams,
                createResponseState
            ),
        });
        expect(resizeChannelTxHash).toBeDefined();

        const resizeReceipt = await blockUtils.waitForTransaction(resizeChannelTxHash);
        expect(resizeReceipt).toBeDefined();

        const postResizeUnifiedBalance = await getLedgerBalances(identity, ws);
        expect(postResizeUnifiedBalance.length).toBe(1);
        expect(postResizeUnifiedBalance[0].amount).toBe((DEPOSIT_AMOUNT + TOP_UP_AMOUNT).toString());
    });

    it('fail on requesting resize after resize was already requested', async () => {
        const DEPOSIT_AMOUNT = BigInt(10)
        const WITHDRAW_AMOUNT = BigInt(1)

        const { params: createResponseParams } = await client.createAndWaitForChannel(ws, {
            tokenAddress: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            amount: toRaw(DEPOSIT_AMOUNT),
        });

        const preResizeUnifiedBalance = await getLedgerBalances(identity, ws);
        expect(preResizeUnifiedBalance.length).toBe(1);
        expect(preResizeUnifiedBalance[0].amount).toBe(DEPOSIT_AMOUNT.toString());

        const msg = await createResizeChannelMessage(identity.messageSKSigner, {
            channel_id: createResponseParams.channelId,
            resize_amount: -toRaw(WITHDRAW_AMOUNT),
            funds_destination: identity.walletAddress,
        });

        await ws.sendAndWaitForResponse(msg, getResizeChannelPredicate(), 1000);

        // do NOT perform the resize again, just send the request and expect error

        const msg2 = await createResizeChannelMessage(identity.messageSKSigner, {
            channel_id: createResponseParams.channelId,
            resize_amount: -toRaw(WITHDRAW_AMOUNT),
            funds_destination: identity.walletAddress,
        });

        try {
            await ws.sendAndWaitForResponse(msg2, getResizeChannelPredicate(), 1000);
        } catch (e) {
            expect(e).toBeDefined();
            expect(e.message).toMatch(/RPC Error.*operation denied: resize already ongoing/);
        }
    });

    it('should release locked funds after close if resize was requested, but not performed', async () => {
        const DEPOSIT_AMOUNT = BigInt(10)
        const WITHDRAWAL_AMOUNT = BigInt(1)

        const { params: createResponseParams } = await client.createAndWaitForChannel(ws, {
            tokenAddress: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            amount: toRaw(DEPOSIT_AMOUNT)
        });

        const preResizeUnifiedBalance = await getLedgerBalances(identity, ws);
        expect(preResizeUnifiedBalance.length).toBe(1);
        expect(preResizeUnifiedBalance[0].amount).toBe(DEPOSIT_AMOUNT.toString());

        const msg = await createResizeChannelMessage(identity.messageSKSigner, {
            channel_id: createResponseParams.channelId,
            resize_amount: -toRaw(WITHDRAWAL_AMOUNT),
            allocate_amount: -toRaw(WITHDRAWAL_AMOUNT),
            funds_destination: identity.walletAddress,
        });

        await ws.sendAndWaitForResponse(msg, getResizeChannelPredicate(), 1000);

        // do NOT perform the resize, just close the channel
        const resizeChannelUpdatePromise = ws.waitForMessage(
            getChannelUpdatePredicateWithStatus(RPCChannelStatus.Closed),
            undefined,
            5000
        );

        const response = await client.closeAndWithdrawChannel(ws, createResponseParams.channelId);
        await resizeChannelUpdatePromise;
        console.log(response.params.amount);

        // after channel is closed, the locked funds for resize should be released back to unified balance
        const postResizeUnifiedBalance = await getLedgerBalances(identity, ws);
        expect(postResizeUnifiedBalance.length).toBe(1);
        // NOTE: this will become equal to "DEPOSIT_AMOUNT" after zero-allocation channel enforcement is implemented
        expect(postResizeUnifiedBalance[0].amount).toBe((WITHDRAWAL_AMOUNT).toString());
    });
});
