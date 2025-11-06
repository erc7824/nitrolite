import { createAuthSessionWithClearnode } from '@/auth';
import { DatabaseUtils } from '@/databaseUtils';
import { Identity } from '@/identity';
import { CONFIG } from '@/setup';
import { getTransferPredicate, TestWebSocket } from '@/ws';
import {
    createTransferMessage,
    parseTransferResponse,
    parseAnyRPCResponse,
    RPCTransferAllocation,
    TransferRequestParams,
    RPCMethod,
} from '@erc7824/nitrolite';

describe('Transfer Integration', () => {
    let ws: TestWebSocket;
    let senderIdentity: Identity;
    let recipientIdentity: Identity;
    let databaseUtils: DatabaseUtils;

    beforeAll(async () => {
        // Setup database utils for cleanup
        databaseUtils = new DatabaseUtils();

        // Create identities
        senderIdentity = new Identity(CONFIG.IDENTITIES[0].WALLET_PK, CONFIG.IDENTITIES[0].SESSION_PK);
        recipientIdentity = new Identity(CONFIG.IDENTITIES[1].WALLET_PK, CONFIG.IDENTITIES[1].SESSION_PK);

        // Create WebSocket connection
        ws = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
        await ws.connect();

        // Create authenticated session for sender
        await createAuthSessionWithClearnode(ws, senderIdentity);

        // Fund sender's account
        await databaseUtils.seedLedger(senderIdentity.walletAddress, senderIdentity.walletAddress, 0, 'usdc', 1000);
    });

    afterAll(async () => {
        if (ws) {
            ws.close();
        }

        // Clean up database
        await databaseUtils.resetClearnodeState();
        await databaseUtils.close();
    });

    describe('Transfer Operations', () => {
        it('should successfully transfer funds to another wallet', async () => {
            const transferParams: TransferRequestParams = {
                destination: recipientIdentity.walletAddress,
                allocations: [
                    {
                        asset: 'usdc',
                        amount: '100',
                    },
                ]
            };

            const transferMsg = await createTransferMessage(senderIdentity.messageSKSigner, transferParams);
            const response = await ws.sendAndWaitForResponse(transferMsg, getTransferPredicate(), 5000);

            expect(response).toBeDefined();

            const parsedResponse = parseTransferResponse(response);
            expect(parsedResponse).toBeDefined();
            expect(parsedResponse.params).toBeDefined();
            expect(parsedResponse.params.transactions).toBeDefined();
            expect(Array.isArray(parsedResponse.params.transactions)).toBe(true);
            expect(parsedResponse.params.transactions.length).toBe(1);

            // Verify transaction details
            const transaction = parsedResponse.params.transactions[0];
            expect(transaction.fromAccount).toBe(senderIdentity.walletAddress);
            expect(transaction.toAccount).toBe(recipientIdentity.walletAddress);
            expect(transaction.asset).toBe('usdc');
            expect(transaction.amount).toBe('100');
            expect(transaction.txType).toBe('transfer');
        });

        it('should reject duplicate transfer request', async () => {
            const transferParams: TransferRequestParams = {
                destination: recipientIdentity.walletAddress,
                allocations: [
                    {
                        asset: 'usdc',
                        amount: '50',
                    },
                ]
            };

            // First transfer - should succeed
            const transferMsg1 = await createTransferMessage(
                senderIdentity.messageSKSigner,
                transferParams,
            );
            const response1 = await ws.sendAndWaitForResponse(transferMsg1, getTransferPredicate(), 5000);

            expect(response1).toBeDefined();

            // Check if it's an error or success
            const parsed1 = parseAnyRPCResponse(response1);
            if (parsed1.method === RPCMethod.Error) {
                const errorParams = parsed1.params as { error: string };
                throw new Error(`First transfer failed: ${errorParams.error}`);
            }

            const parsedResponse1 = parseTransferResponse(response1);
            expect(parsedResponse1).toBeDefined();
            expect(parsedResponse1.params).toBeDefined();
            expect(parsedResponse1.params.transactions).toBeDefined();
            expect(parsedResponse1.params.transactions.length).toBe(1);

            // Verify first transaction succeeded
            const transaction1 = parsedResponse1.params.transactions[0];
            expect(transaction1.amount).toBe('50');

            try {
                await ws.sendAndWaitForResponse(transferMsg1, getTransferPredicate(), 5000);
            } catch (error) {
                // Expecting an error for duplicate transfer
                const err = error as Error;
                expect(err.message).toMatch(/RPC Error.*operation denied.*the request has already been processed/i);
                return;
            }

            throw new Error('Duplicate transfer request was not rejected as expected.');
        });
    });
});
