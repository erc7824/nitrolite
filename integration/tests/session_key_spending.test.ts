import { BlockchainUtils } from '@/blockchainUtils';
import { DatabaseUtils } from '@/databaseUtils';
import { Identity } from '@/identity';
import { TestNitroliteClient } from '@/nitroliteClient';
import { TestWebSocket } from '@/ws';
import { RPCAppStateIntent, RPCProtocolVersion } from '@erc7824/nitrolite';
import { Hex } from 'viem';
import { setupTestIdentitiesAndConnections } from '@/testSetup';
import {
    createTestChannels,
    authenticateAppWithAllowances,
    createTestAppSession,
    toRaw,
    getLedgerBalances,
} from '@/testHelpers';
import { submitAppStateUpdate_v04 } from '@/testAppSessionHelpers';

describe('Session Key Spending Caps', () => {
    const onChainDepositAmount = BigInt(1000);
    const spendingCapAmount = BigInt(500); // Session key limited to 500 USDC
    const initialDepositAmount = BigInt(100);

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

    let appSessionId: string;

    let currentVersion = 1;

    const SESSION_DATA = { gameType: 'chess', gameState: 'waiting' };

    beforeAll(async () => {
        blockUtils = new BlockchainUtils();
        databaseUtils = new DatabaseUtils();

        const { Identity: IdentityClass } = await import('@/identity');
        const { TestWebSocket: TestWebSocketClass } = await import('@/ws');
        const { TestNitroliteClient: TestNitroliteClientClass } = await import('@/nitroliteClient');
        const { createAuthSessionWithClearnode } = await import('@/auth');
        const { CONFIG } = await import('@/setup');

        alice = new IdentityClass(CONFIG.IDENTITIES[0].WALLET_PK, CONFIG.IDENTITIES[0].SESSION_PK);
        aliceWS = new TestWebSocketClass(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
        aliceClient = new TestNitroliteClientClass(alice);
        await aliceWS.connect();
        await createAuthSessionWithClearnode(aliceWS, alice);

        aliceAppIdentity = new IdentityClass(CONFIG.IDENTITIES[0].WALLET_PK, CONFIG.IDENTITIES[0].APP_SESSION_PK);
        aliceAppWS = new TestWebSocketClass(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
        await aliceAppWS.connect();

        bob = new IdentityClass(CONFIG.IDENTITIES[1].WALLET_PK, CONFIG.IDENTITIES[1].SESSION_PK);
        bobAppIdentity = new IdentityClass(CONFIG.IDENTITIES[1].WALLET_PK, CONFIG.IDENTITIES[1].APP_SESSION_PK);
        bobWS = new TestWebSocketClass(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
        bobClient = new TestNitroliteClientClass(bob);
        await bobWS.connect();
        await createAuthSessionWithClearnode(bobWS, bob);
    });

    beforeEach(async () => {
        await blockUtils.makeSnapshot();

        // Create channels for both Alice and Bob
        await createTestChannels(
            [
                { client: aliceClient, ws: aliceWS },
                { client: bobClient, ws: bobWS },
            ],
            toRaw(onChainDepositAmount)
        );

        // Authenticate with spending cap of 500 USDC
        await authenticateAppWithAllowances(aliceAppWS, aliceAppIdentity, spendingCapAmount);

        currentVersion = 1;
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

    describe('Initial deposit within cap', () => {
        it('should allow deposit within spending cap', async () => {
            appSessionId = await createTestAppSession(
                aliceAppIdentity,
                bobAppIdentity,
                aliceAppWS,
                RPCProtocolVersion.NitroRPC_0_4,
                initialDepositAmount,
                SESSION_DATA
            );

            expect(appSessionId).toBeDefined();

            // Verify ledger balance decreased
            const ledgerBalances = await getLedgerBalances(aliceAppIdentity, aliceAppWS);
            expect(ledgerBalances[0].amount).toBe((onChainDepositAmount - initialDepositAmount).toString());
        });

        it('should reject deposit exceeding spending cap', async () => {
            const excessiveAmount = spendingCapAmount + BigInt(100); // 600 USDC (exceeds 500 cap)

            await expect(
                createTestAppSession(
                    aliceAppIdentity,
                    bobAppIdentity,
                    aliceAppWS,
                    RPCProtocolVersion.NitroRPC_0_4,
                    excessiveAmount,
                    SESSION_DATA
                )
            ).rejects.toThrow(/session key spending validation failed.*insufficient session key allowance/i);
        });
    });

    describe('Cumulative spending tracking', () => {
        beforeEach(async () => {
            // Create initial app session with 100 USDC deposit
            appSessionId = await createTestAppSession(
                aliceAppIdentity,
                bobAppIdentity,
                aliceAppWS,
                RPCProtocolVersion.NitroRPC_0_4,
                initialDepositAmount,
                SESSION_DATA
            );
        });

        it('should allow additional deposit within remaining cap', async () => {
            const additionalDeposit = BigInt(200); // Total: 100 + 200 = 300 (within 500 cap)

            const allocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: 'usdc',
                    amount: (initialDepositAmount + additionalDeposit).toString(),
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: 'usdc',
                    amount: '0',
                },
            ];

            await submitAppStateUpdate_v04(
                aliceAppWS,
                aliceAppIdentity,
                appSessionId,
                RPCAppStateIntent.Deposit,
                ++currentVersion,
                allocations,
                SESSION_DATA
            );

            // Verify ledger balance
            const ledgerBalances = await getLedgerBalances(aliceAppIdentity, aliceAppWS);
            expect(ledgerBalances[0].amount).toBe(
                (onChainDepositAmount - initialDepositAmount - additionalDeposit).toString()
            );
        });

        it('should reject additional deposit exceeding remaining cap', async () => {
            const excessiveAdditionalDeposit = BigInt(450); // Total: 100 + 450 = 550 (exceeds 500 cap)

            const allocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: 'usdc',
                    amount: (initialDepositAmount + excessiveAdditionalDeposit).toString(),
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: 'usdc',
                    amount: '0',
                },
            ];

            await expect(
                submitAppStateUpdate_v04(
                    aliceAppWS,
                    aliceAppIdentity,
                    appSessionId,
                    RPCAppStateIntent.Deposit,
                    ++currentVersion,
                    allocations,
                    SESSION_DATA
                )
            ).rejects.toThrow(/session key spending validation failed.*insufficient session key allowance/i);
        });

        it('should track cumulative spending across multiple deposits', async () => {
            // First additional deposit: 150 USDC (total: 250)
            let allocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: 'usdc',
                    amount: (initialDepositAmount + BigInt(150)).toString(),
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: 'usdc',
                    amount: '0',
                },
            ];

            await submitAppStateUpdate_v04(
                aliceAppWS,
                aliceAppIdentity,
                appSessionId,
                RPCAppStateIntent.Deposit,
                ++currentVersion,
                allocations,
                SESSION_DATA
            );

            // Second additional deposit: 200 USDC (total: 450, within 500 cap)
            allocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: 'usdc',
                    amount: (initialDepositAmount + BigInt(150) + BigInt(200)).toString(),
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: 'usdc',
                    amount: '0',
                },
            ];

            await submitAppStateUpdate_v04(
                aliceAppWS,
                aliceAppIdentity,
                appSessionId,
                RPCAppStateIntent.Deposit,
                ++currentVersion,
                allocations,
                SESSION_DATA
            );

            // Verify total spent is 450 USDC
            const ledgerBalances = await getLedgerBalances(aliceAppIdentity, aliceAppWS);
            expect(ledgerBalances[0].amount).toBe((onChainDepositAmount - BigInt(450)).toString());

            // Third deposit attempting 100 more (total would be 550, exceeds cap)
            allocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: 'usdc',
                    amount: (initialDepositAmount + BigInt(150) + BigInt(200) + BigInt(100)).toString(),
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: 'usdc',
                    amount: '0',
                },
            ];

            await expect(
                submitAppStateUpdate_v04(
                    aliceAppWS,
                    aliceAppIdentity,
                    appSessionId,
                    RPCAppStateIntent.Deposit,
                    ++currentVersion,
                    allocations,
                    SESSION_DATA
                )
            ).rejects.toThrow(/session key spending validation failed.*insufficient session key allowance/i);
        });
    });

    describe('Withdrawals do not affect spending cap', () => {
        beforeEach(async () => {
            // Create initial app session with 300 USDC deposit
            appSessionId = await createTestAppSession(
                aliceAppIdentity,
                bobAppIdentity,
                aliceAppWS,
                RPCProtocolVersion.NitroRPC_0_4,
                BigInt(300),
                SESSION_DATA
            );
        });

        it('should not restore spending cap after withdrawal', async () => {
            // Withdraw 100 USDC
            const allocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: 'usdc',
                    amount: '200', // Withdraw 100 from 300
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: 'usdc',
                    amount: '0',
                },
            ];

            await submitAppStateUpdate_v04(
                aliceAppWS,
                aliceAppIdentity,
                appSessionId,
                RPCAppStateIntent.Withdraw,
                ++currentVersion,
                allocations,
                SESSION_DATA
            );

            // Verify ledger balance increased by 100
            const ledgerBalances = await getLedgerBalances(aliceAppIdentity, aliceAppWS);
            expect(ledgerBalances[0].amount).toBe((onChainDepositAmount - BigInt(200)).toString());

            // Try to deposit 300 more (total spent would be 300 + 300 = 600, exceeds cap)
            const depositAllocations = [
                {
                    participant: aliceAppIdentity.walletAddress,
                    asset: 'usdc',
                    amount: '500', // 200 + 300 = 500
                },
                {
                    participant: bobAppIdentity.walletAddress,
                    asset: 'usdc',
                    amount: '0',
                },
            ];

            await expect(
                submitAppStateUpdate_v04(
                    aliceAppWS,
                    aliceAppIdentity,
                    appSessionId,
                    RPCAppStateIntent.Deposit,
                    ++currentVersion,
                    depositAllocations,
                    SESSION_DATA
                )
            ).rejects.toThrow(/session key spending validation failed.*insufficient session key allowance/i);
        });
    });

    describe('Multi-asset spending caps', () => {
        it('should enforce spending cap per asset independently', async () => {
            // Note: This test would require multi-asset support in the test helpers
            // For now, we document the expected behavior:
            // - Each asset in allowances has its own independent spending cap
            // - Spending on USDC doesn't affect ETH cap and vice versa
            // - Session key tracks cumulative spending per asset
        });
    });
});
