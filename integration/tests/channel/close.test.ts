import { createAuthSessionWithClearnode } from '@/auth';
import { BlockchainUtils } from '@/blockchainUtils';
import { Identity } from '@/identity';
import { TestNitroliteClient } from '@/nitroliteClient';
import { CONFIG } from '@/setup';
import { getCloseChannelPredicate, getResizeChannelPredicate, TestWebSocket } from '@/ws';
import {
    Allocation,
    createCloseChannelMessage,
    createResizeChannelMessage,
    parseRPCResponse,
    ResizeChannelRPCResponse,
} from '@erc7824/nitrolite';
import { Address, Hex } from 'viem';

describe('Close channel', () => {
    const depositAmount = BigInt(100 * 10 ** 6); // 100 USDC

    let ws: TestWebSocket;
    let identity: Identity;
    let client: TestNitroliteClient;
    let blockUtils: BlockchainUtils;

    beforeAll(async () => {
        blockUtils = new BlockchainUtils();

        ws = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
        await ws.connect();

        identity = new Identity(CONFIG.IDENTITIES[0].WALLET_PK, CONFIG.IDENTITIES[0].SESSION_PK);

        await createAuthSessionWithClearnode(ws, identity);
    });

    afterAll(() => {
        ws.close();
    });

    it('should create nitrolite client to close channels', async () => {
        client = new TestNitroliteClient(identity);

        expect(client).toBeDefined();
        expect(client).toHaveProperty('depositAndCreateChannel');
        expect(client).toHaveProperty('closeChannel');
        expect(client).toHaveProperty('withdrawal');
    });

    it('should close channel and withdraw funds', async () => {
        const { params } = await client.createAndWaitForChannel(ws, {
            tokenAddress: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            amount: depositAmount,
        });

        const resizeMessage = await createResizeChannelMessage(identity.messageSigner, [
            {
                channel_id: params.channel_id,
                // @ts-ignore
                resize_amount: Number(-depositAmount),
                // @ts-ignore
                allocate_amount: 0,
                funds_destination: identity.walletAddress,
            },
        ]);
        const resizeResponse = await ws.sendAndWaitForResponse(resizeMessage, getResizeChannelPredicate(), 1000);
        expect(resizeResponse).toBeDefined();

        const resizeParsedResponse = parseRPCResponse(resizeResponse) as ResizeChannelRPCResponse;

        console.log(resizeParsedResponse);

        const resizeTxHash = await client.resizeChannel({
            resizeState: {
                channelId: resizeParsedResponse.params.channel_id,
                stateData: resizeParsedResponse.params.state_data as Hex,
                // @ts-ignore
                allocations: resizeParsedResponse.params.allocations.map((a) => ({
                    destination: a.destination as Address,
                    token: a.token as Address,
                    amount: a.amount,
                })) as [Allocation, Allocation],
                // @ts-ignore
                version: resizeParsedResponse.params.version,
                intent: resizeParsedResponse.params.intent,
                serverSignature: {
                    v: +resizeParsedResponse.params.server_signature.v,
                    r: resizeParsedResponse.params.server_signature.r as Hex,
                    s: resizeParsedResponse.params.server_signature.s as Hex,
                },
            },
            proofStates: [],
        });

        const resizeReceipt = await blockUtils.waitForTransaction(resizeTxHash);
        expect(resizeReceipt).toBeDefined();

        const msg = await createCloseChannelMessage(identity.messageSigner, params.channel_id, identity.walletAddress);
        const response = await ws.sendAndWaitForResponse(msg, getCloseChannelPredicate(), 1000);
        expect(response).toBeDefined();

        console.log(response);

        // client.closeChannel({
        //     finalState: {
        //         channelId: params.channel_id,
        //         stateData: '0x',
        //         allocations: initialState.allocations as [Allocation, Allocation],
        //         version: initialState.version,
        //         serverSignature: initialState.,
        //     },
        // });

        // stateData?: Hex;
        // finalState: {
        //     channelId: ChannelId;
        //     stateData: Hex;
        //     allocations: [Allocation, Allocation];
        //     version: bigint;
        //     serverSignature: Signature;
        // };
    });
});
