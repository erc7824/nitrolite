import { createAuthSessionWithClearnode } from '@/auth';
import { BlockchainUtils } from '@/blockchainUtils';
import { DatabaseUtils } from '@/databaseUtils';
import { Identity } from '@/identity';
import { TestNitroliteClient } from '@/nitroliteClient';
import { CONFIG } from '@/setup';
import { getResizeChannelPredicate, TestWebSocket } from '@/ws';
import { createResizeChannelMessage, parseResizeChannelResponse } from '@erc7824/nitrolite';
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
        const { params: createResponseParams } = await client.createAndWaitForChannel(ws, {
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

        const msg = await createResizeChannelMessage(identity.messageSigner, {
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
        expect(resizeResponseParams.state.version).toBe(createResponseParams.version + 1);

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

        const resizeChannelTxHash = await client.resizeChannel({
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
                    version: BigInt(createResponseParams.version),
                    data: '0x',
                    allocations: [
                        {
                            destination: identity.walletAddress,
                            token: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
                            amount: depositAmount * BigInt(5),
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

        const postResizeAccountBalance = await client.getAccountBalance(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
        expect(postResizeAccountBalance).toBe(depositAmount * BigInt(4)); // 1000 - 500 - 100

        const postResizeChannelBalance = await client.getChannelBalance(
            createResponseParams.channelId,
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS
        );
        expect(postResizeChannelBalance).toBe(depositAmount * BigInt(6)); // 500 + 100
    });

    it('should resize channel by withdrawing funds from channel to deposit', async () => {
        const { params: createResponseParams } = await client.createAndWaitForChannel(ws, {
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

        const msg = await createResizeChannelMessage(identity.messageSigner, {
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

        const resizeChannelTxHash = await client.resizeChannel({
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
                    version: BigInt(createResponseParams.version),
                    data: '0x',
                    allocations: [
                        {
                            destination: identity.walletAddress,
                            token: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
                            amount: depositAmount * BigInt(5), // 500
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

        const postResizeAccountBalance = await client.getAccountBalance(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
        expect(postResizeAccountBalance).toBe(depositAmount * BigInt(6)); // 1000 - 500 + 100

        const postResizeChannelBalance = await client.getChannelBalance(
            createResponseParams.channelId,
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS
        );
        expect(postResizeChannelBalance).toBe(depositAmount * BigInt(4)); // 500 - 100
    });

    it('should resize channel by allocating funds from channel to virtual ledger', async () => {
        const { params: createResponseParams } = await client.createAndWaitForChannel(ws, {
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

        const msg = await createResizeChannelMessage(identity.messageSigner, {
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

        const resizeChannelTxHash = await client.resizeChannel({
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
                    version: BigInt(createResponseParams.version),
                    data: '0x',
                    allocations: [
                        {
                            destination: identity.walletAddress,
                            token: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
                            amount: depositAmount * BigInt(5), // 500
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

        const postResizeAccountBalance = await client.getAccountBalance(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
        expect(postResizeAccountBalance).toBe(depositAmount * BigInt(5)); // 1000 - 500

        const postResizeChannelBalance = await client.getChannelBalance(
            createResponseParams.channelId,
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS
        );
        expect(postResizeChannelBalance).toBe(depositAmount * BigInt(4)); // 1000 - 500 - 100
    });
});
