import { createAuthSessionWithClearnode } from '@/auth';
import { BlockchainUtils } from '@/blockchainUtils';
import { DatabaseUtils } from '@/databaseUtils';
import { Identity } from '@/identity';
import { TestNitroliteClient } from '@/nitroliteClient';
import { CONFIG } from '@/setup';
import { getChannelUpdatePredicateWithStatus, TestWebSocket, getGetLedgerEntriesPredicate } from '@/ws';
import { createGetLedgerEntriesMessage, RPCChannelStatus, rpcResponseParser } from '@erc7824/nitrolite';
import { parseUnits, GetTxpoolContentReturnType, Hash } from 'viem';

describe('Close channel', () => {
    const depositAmount = parseUnits('100', 6); // 100 USDC (decimals = 6)
    const decimalDepositAmount = 100;

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

        await blockUtils.resumeMining();
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

    afterAll(async () => {
        databaseUtils.close();

        await blockUtils.resumeMining();
    });

    it('should create nitrolite client to challenge channels', async () => {
        client = new TestNitroliteClient(identity);

        expect(client).toBeDefined();
        expect(client).toHaveProperty('depositAndCreateChannel');
        expect(client).toHaveProperty('challengeChannel');
    });

    it('should challenge channel in joining state', async () => {
        const joiningChannelPromise = ws.waitForMessage(
            getChannelUpdatePredicateWithStatus(RPCChannelStatus.Joining),
            undefined,
            5000
        );

        const hash = await client.approveTokens(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS, depositAmount);
        await blockUtils.waitForTransaction(hash);

        await blockUtils.pauseMining();

        const { channelId, txHash: createTxHash } = await client.depositAndCreateChannel(
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            depositAmount,
            {
                initialAllocationAmounts: [depositAmount, BigInt(0)],
                stateData: '0x',
            }
        );

        // Mine exactly one block to ensure the transaction is processed and join is not mined
        const depositTxPromise = blockUtils.waitForTransaction(createTxHash);
        await blockUtils.mineBlock();
        await depositTxPromise;

        const { lastValidState } = await client.getChannelData(channelId);
        const poolWithJoin: GetTxpoolContentReturnType = await new Promise((resolve, reject) => {
            const timeout = setTimeout(() => {
                clearInterval(interval);
                reject(new Error('Timed out waiting for pending transaction in txpool'));
            }, 5000);

            const interval = setInterval(async () => {
                const pool = await blockUtils.readTxPool();
                if (Object.keys(pool.pending).length > 0) {
                    clearInterval(interval);
                    clearTimeout(timeout);
                    resolve(pool);
                }
            }, 200);
        });

        // TODO: this approach is very brittle, and could fail if there are multiple pending transactions
        // which usually doesn't happen in tests, but still
        const txKey = Object.keys(poolWithJoin.pending)[0];
        const txIndex = Object.keys(poolWithJoin.pending[txKey])[0];
        const joinTx = poolWithJoin.pending[txKey][txIndex];

        await blockUtils.dropTxFromPool(joinTx.hash as Hash);

        const challengeTxHash = await client.challengeChannel({
            channelId,
            candidateState: lastValidState,
        });

        const challengeTxPromise = blockUtils.waitForTransaction(challengeTxHash);
        await blockUtils.mineBlock();
        await challengeTxPromise;

        const joinTxHash = await blockUtils.sendRawTransactionAs(
            CONFIG.IDENTITIES[0].WALLET_PK,
            {
                chainId: Number(BigInt(joinTx.chainId)),
                nonce: Number(BigInt(joinTx.nonce)),
                gasPrice: BigInt(joinTx.gasPrice),
                gas: BigInt(joinTx.gas),
                to: joinTx.to,
                value: BigInt(joinTx.value),
                data: joinTx.input,
            },
            {
                v: BigInt(joinTx.v),
                r: joinTx.r,
                s: joinTx.s,
            }
        );

        const joinTxPromise = blockUtils.waitForTransaction(joinTxHash);
        await blockUtils.mineBlock();
        await joinTxPromise;

        const channelData = await client.getChannelData(channelId);
        expect(channelData).toBeDefined();

        const joiningResponse = await joiningChannelPromise;
        expect(joiningResponse).toBeDefined();

        const msg = await createGetLedgerEntriesMessage(identity.messageSigner, channelId);
        const response = await ws.sendAndWaitForResponse(msg, getGetLedgerEntriesPredicate(), 5000);

        const { params: parsedResponseParams } = rpcResponseParser.getLedgerEntries(response);
        expect(parsedResponseParams).toBeDefined();

        expect(parsedResponseParams).toHaveLength(2);
        expect(+parsedResponseParams[0].debit + +parsedResponseParams[1].debit).toEqual(decimalDepositAmount);
        expect(+parsedResponseParams[0].credit + +parsedResponseParams[1].credit).toEqual(decimalDepositAmount);
    });
});
