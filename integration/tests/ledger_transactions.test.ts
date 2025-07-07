import { createAuthSessionWithClearnode } from "@/auth";
import { DatabaseUtils } from "@/databaseUtils";
import { Identity } from "@/identity";
import { CONFIG } from "@/setup";
import { getGetLedgerTransactionsPredicate, TestWebSocket } from "@/ws";
import { createGetLedgerTransactionsMessage, rpcResponseParser, GetLedgerTransactionsFilters, TxType } from "@erc7824/nitrolite";

describe("Ledger Transactions Integration", () => {
    let ws: TestWebSocket;
    let identity: Identity;
    let databaseUtils: DatabaseUtils;

    beforeAll(async () => {
        // Setup database utils for cleanup
        databaseUtils = new DatabaseUtils();

        // Create identity
        identity = new Identity(CONFIG.IDENTITIES[0].WALLET_PK, CONFIG.IDENTITIES[0].SESSION_PK);

        // Create WebSocket connection
        ws = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
        await ws.connect();

        // Create authenticated session
        await createAuthSessionWithClearnode(ws, identity);
    });

    afterAll(async () => {
        if (ws) {
            ws.close();
        }

        // Clean up database
        databaseUtils.cleanupDatabaseData();
        databaseUtils.close();
    });

    describe("createGetLedgerTransactionsMessage", () => {
        it("should successfully request ledger transactions with no filters", async () => {
            const accountId = identity.walletAddress;
            const msg = await createGetLedgerTransactionsMessage(identity.messageSigner, accountId);

            const response = await ws.sendAndWaitForResponse(msg, getGetLedgerTransactionsPredicate(), 5000);

            expect(response).toBeDefined();

            const parsedResponse = rpcResponseParser.getLedgerTransactions(response);
            expect(parsedResponse).toBeDefined();
            expect(parsedResponse.params).toBeDefined();
            expect(Array.isArray(parsedResponse.params)).toBe(true);
        });

        it("should successfully request ledger transactions with asset filter", async () => {
            const accountId = identity.walletAddress;
            const filters: GetLedgerTransactionsFilters = {
                asset: "usdc",
            };

            const msg = await createGetLedgerTransactionsMessage(identity.messageSigner, accountId, filters);

            const response = await ws.sendAndWaitForResponse(msg, getGetLedgerTransactionsPredicate(), 5000);

            expect(response).toBeDefined();

            const parsedResponse = rpcResponseParser.getLedgerTransactions(response);
            expect(parsedResponse).toBeDefined();
            expect(parsedResponse.params).toBeDefined();
            expect(Array.isArray(parsedResponse.params)).toBe(true);

            // If there are transactions, they should all be for usdc
            if (parsedResponse.params.length > 0) {
                parsedResponse.params.forEach((transaction) => {
                    expect(transaction.asset).toBe("usdc");
                });
            }
        });

        it("should successfully request ledger transactions with tx_type filter", async () => {
            const accountId = identity.walletAddress;
            const filters: GetLedgerTransactionsFilters = {
                tx_type: TxType.Deposit,
            };

            const msg = await createGetLedgerTransactionsMessage(identity.messageSigner, accountId, filters);

            const response = await ws.sendAndWaitForResponse(msg, getGetLedgerTransactionsPredicate(), 5000);

            expect(response).toBeDefined();

            const parsedResponse = rpcResponseParser.getLedgerTransactions(response);
            expect(parsedResponse).toBeDefined();
            expect(parsedResponse.params).toBeDefined();
            expect(Array.isArray(parsedResponse.params)).toBe(true);

            // If there are transactions, they should all be of type 'deposit'
            if (parsedResponse.params.length > 0) {
                parsedResponse.params.forEach((transaction) => {
                    expect(transaction.txType).toBe(TxType.Deposit);
                });
            }
        });

        it("should successfully request ledger transactions with pagination", async () => {
            const accountId = identity.walletAddress;
            const filters: GetLedgerTransactionsFilters = {
                limit: 5,
                offset: 0,
            };

            const msg = await createGetLedgerTransactionsMessage(identity.messageSigner, accountId, filters);

            const response = await ws.sendAndWaitForResponse(msg, getGetLedgerTransactionsPredicate(), 5000);

            expect(response).toBeDefined();

            const parsedResponse = rpcResponseParser.getLedgerTransactions(response);
            expect(parsedResponse).toBeDefined();
            expect(parsedResponse.params).toBeDefined();
            expect(Array.isArray(parsedResponse.params)).toBe(true);

            // Should not return more than the limit
            expect(parsedResponse.params.length).toBeLessThanOrEqual(5);
        });

        it("should successfully request ledger transactions with sort order", async () => {
            const accountId = identity.walletAddress;
            const filters: GetLedgerTransactionsFilters = {
                sort: "desc",
                limit: 10,
            };

            const msg = await createGetLedgerTransactionsMessage(identity.messageSigner, accountId, filters);

            const response = await ws.sendAndWaitForResponse(msg, getGetLedgerTransactionsPredicate(), 5000);

            expect(response).toBeDefined();

            const parsedResponse = rpcResponseParser.getLedgerTransactions(response);
            expect(parsedResponse).toBeDefined();
            expect(parsedResponse.params).toBeDefined();
            expect(Array.isArray(parsedResponse.params)).toBe(true);

            // If there are multiple transactions, they should be sorted by createdAt in descending order
            if (parsedResponse.params.length > 1) {
                for (let i = 0; i < parsedResponse.params.length - 1; i++) {
                    const currentDate = parsedResponse.params[i].createdAt;
                    const nextDate = parsedResponse.params[i + 1].createdAt;
                    expect(currentDate.getTime()).toBeGreaterThanOrEqual(nextDate.getTime());
                }
            }
        });

        it("should successfully request ledger transactions with all filters", async () => {
            const accountId = identity.walletAddress;
            const filters: GetLedgerTransactionsFilters = {
                asset: "usdc",
                tx_type: TxType.Deposit,
                offset: 0,
                limit: 3,
                sort: "desc",
            };

            const msg = await createGetLedgerTransactionsMessage(identity.messageSigner, accountId, filters);

            const response = await ws.sendAndWaitForResponse(msg, getGetLedgerTransactionsPredicate(), 5000);

            expect(response).toBeDefined();

            const parsedResponse = rpcResponseParser.getLedgerTransactions(response);
            expect(parsedResponse).toBeDefined();
            expect(parsedResponse.params).toBeDefined();
            expect(Array.isArray(parsedResponse.params)).toBe(true);

            // Should not return more than the limit
            expect(parsedResponse.params.length).toBeLessThanOrEqual(3);

            // All transactions should match the filters
            parsedResponse.params.forEach((transaction) => {
                expect(transaction.asset).toBe("usdc");
                expect(transaction.txType).toBe(TxType.Deposit);
            });
        });

        it("should handle empty results gracefully", async () => {
            const accountId = identity.walletAddress;
            const filters: GetLedgerTransactionsFilters = {
                asset: "NONEXISTENT_ASSET",
            };

            const msg = await createGetLedgerTransactionsMessage(identity.messageSigner, accountId, filters);

            const response = await ws.sendAndWaitForResponse(msg, getGetLedgerTransactionsPredicate(), 5000);

            expect(response).toBeDefined();

            const parsedResponse = rpcResponseParser.getLedgerTransactions(response);
            expect(parsedResponse).toBeDefined();
            expect(parsedResponse.params).toBeDefined();
            expect(Array.isArray(parsedResponse.params)).toBe(true);
            expect(parsedResponse.params.length).toBe(0);
        });

        it("should handle invalid account ID gracefully", async () => {
            const invalidAccountId = "0x0000000000000000000000000000000000000000";

            const msg = await createGetLedgerTransactionsMessage(identity.messageSigner, invalidAccountId);

            const response = await ws.sendAndWaitForResponse(msg, getGetLedgerTransactionsPredicate(), 5000);

            expect(response).toBeDefined();

            const parsedResponse = rpcResponseParser.getLedgerTransactions(response);
            expect(parsedResponse).toBeDefined();
            expect(parsedResponse.params).toBeDefined();
            expect(Array.isArray(parsedResponse.params)).toBe(true);
            // Should return empty array for invalid/non-existent account
            expect(parsedResponse.params.length).toBe(0);
        });
    });
});
