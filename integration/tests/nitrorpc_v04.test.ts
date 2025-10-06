import { BlockchainUtils } from '@/blockchainUtils';
import { DatabaseUtils } from '@/databaseUtils';
import { Identity } from '@/identity';
import { TestNitroliteClient } from '@/nitroliteClient';
import { TestWebSocket } from '@/ws';
import { RPCAppStateIntent, RPCProtocolVersion } from '@erc7824/nitrolite';
import { Hex } from 'viem';
import { fetchAndParseAppSessions, setupTestIdentitiesAndConnections } from '@/testSetup';
import {
    createTestChannels,
    authenticateAppWithAllowances,
    createTestAppSession,
    toRaw,
} from '@/testHelpers';
import { submitAppStateUpdate_v04 } from '@/testAppSessionHelpers';
import { createAuthSessionWithClearnode } from '@/auth';

describe('App session state v0.4 error cases', () => {
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

    let currentVersion = 1;

    const START_SESSION_DATA = { gameType: 'chess', gameState: 'waiting' };

    let START_ALLOCATIONS;

    beforeAll(async () => {
        blockUtils = new BlockchainUtils();
        databaseUtils = new DatabaseUtils();

        ({alice, aliceWS, aliceClient, aliceAppIdentity, aliceAppWS, bob, bobWS, bobClient, bobAppIdentity} = await setupTestIdentitiesAndConnections());

        START_ALLOCATIONS = [
            {
                participant: aliceAppIdentity.walletAddress,
                asset: 'USDC',
                amount: (appSessionDepositAmount).toString(),
            },
            {
                participant: bobAppIdentity.walletAddress,
                asset: 'USDC',
                amount: '0',
            },
        ];

        await blockUtils.makeSnapshot();
    });

    beforeEach(async () => {
        [aliceChannelId, bobChannelId] = await createTestChannels([{client: aliceClient, ws: aliceWS}, {client: bobClient, ws: bobWS}], toRaw(onChainDepositAmount));

        await authenticateAppWithAllowances(aliceAppWS, aliceAppIdentity, appSessionDepositAmount);

        appSessionId = await createTestAppSession(
            aliceAppIdentity,
            bobAppIdentity,
            aliceAppWS,
            RPCProtocolVersion.NitroRPC_0_4,
            appSessionDepositAmount,
            START_SESSION_DATA
        );

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

    it('should fail on skipping version number', async () => {
        let allocations = structuredClone(START_ALLOCATIONS);

        try {
            await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Operate, currentVersion + 42, allocations, { state: 'blah'});
        } catch (e) {
            expect((e as Error).message).toMatch(
                `RPC Error: incorrect app state: incorrect version: expected ${
                    currentVersion + 1
                }, got ${currentVersion + 42}`
            );

            const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
            expect(sessionData).toEqual(START_SESSION_DATA);
            return;
        }

        throw new Error('Expected error was not thrown');
    });

    it('should fail on operate intent and positive delta', async () => {
        let allocations = structuredClone(START_ALLOCATIONS);

        allocations[0].amount = (BigInt(allocations[0].amount) + BigInt(10)).toString(); // 110 USDC - more than deposited

        try {
            await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Operate, currentVersion + 1, allocations, { state: 'test' });
        } catch (e) {
            expect((e as Error).message).toMatch(/RPC Error.*incorrect operate request.*non-zero allocations sum delta/i);

            const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
            expect(sessionData).toEqual(START_SESSION_DATA);
            return;
        }

        throw new Error('Expected error was not thrown');
    });

    it('should fail on operate intent and negative delta', async () => {
        let allocations = structuredClone(START_ALLOCATIONS);
        allocations[0].amount = (BigInt(allocations[0].amount) - BigInt(10)).toString(); // 90 USDC - less than deposited

        try {
            await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Operate, currentVersion + 1, allocations, { state: 'test' });
        } catch (e) {
            expect((e as Error).message).toMatch(/RPC Error.*incorrect operate request.*non-zero allocations sum delta/i);

            const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
            expect(sessionData).toEqual(START_SESSION_DATA);
            return;
        }

        throw new Error('Expected error was not thrown');
    });

    describe('deposit intent', () => {
        it('should fail on zero delta', async () => {
            let allocations = structuredClone(START_ALLOCATIONS); // same as deposited, zero delta

            try {
                await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Deposit, currentVersion + 1, allocations, { state: 'test' });
            } catch (e) {
                expect((e as Error).message).toMatch(/RPC Error.*incorrect deposit request.*non-positive allocations sum delta/i);

                const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
                expect(sessionData).toEqual(START_SESSION_DATA);
                return;
            }

            throw new Error('Expected error was not thrown');
        });

        it('should fail on negative delta', async () => {
            let allocations = structuredClone(START_ALLOCATIONS);
            allocations[0].amount = (BigInt(allocations[0].amount) - BigInt(10)).toString(); // 90 USDC - less than deposited

            try {
                await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Deposit, currentVersion + 1, allocations, { state: 'test' });
            } catch (e) {
                expect((e as Error).message).toMatch(/RPC Error.*incorrect deposit request.*decreased allocation for participant/i);

                const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
                expect(sessionData).toEqual(START_SESSION_DATA);
                return;
            }

            throw new Error('Expected error was not thrown');
        });

        it('should fail on positive and negative allocation deltas', async () => {
            let allocations = structuredClone(START_ALLOCATIONS);
            allocations[0].amount = (BigInt(allocations[0].amount) + BigInt(20)).toString(); // 120 USDC - more than deposited
            allocations[1].amount = (BigInt(allocations[1].amount) - BigInt(10)).toString(); // 90 USDC - less than deposited

            try {
                await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Deposit, currentVersion + 1, allocations, { state: 'test' });
            } catch (e) {
                expect((e as Error).message).toMatch(/RPC Error.*incorrect deposit request.*decreased allocation for participant/i);

                const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
                expect(sessionData).toEqual(START_SESSION_DATA);
                return;
            }

            throw new Error('Expected error was not thrown');
        });

        it('should fail on insufficient unified balance', async () => {
            // Try to deposit more than Alice has in ledger (she has 1000, already deposited 100, so has 900 available)
            let allocations = structuredClone(START_ALLOCATIONS);
            const hugeAmount = onChainDepositAmount * BigInt(10); // 10,000 USDC
            allocations[0].amount = hugeAmount.toString(); // 10,000 USDC - way more than available

            try {
                await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Deposit, currentVersion + 1, allocations, { state: 'test' });
            } catch (e) {
                expect((e as Error).message).toMatch(/RPC Error.*incorrect deposit request.*insufficient unified balance/i);

                const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
                expect(sessionData).toEqual(START_SESSION_DATA);
                return;
            }

            throw new Error('Expected error was not thrown');
        });

        it('should fail on depositing to v0.2 app session', async () => {
            // Create a v0.2 app session (which doesn't support deposits/withdrawals)
            const v02AppSessionId = await createTestAppSession(
                aliceAppIdentity,
                bobAppIdentity,
                aliceAppWS,
                RPCProtocolVersion.NitroRPC_0_2,
                appSessionDepositAmount,
                START_SESSION_DATA
            );

            let allocations = structuredClone(START_ALLOCATIONS);
            allocations[0].amount = (BigInt(allocations[0].amount) + BigInt(10)).toString(); // 110 USDC - more than deposited

            try {
                await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, v02AppSessionId, RPCAppStateIntent.Deposit, 2, allocations, { state: 'test' });
            } catch (e) {
                expect((e as Error).message).toMatch(/RPC Error.*incorrect request.*specified parameters are not supported in this protocol/i);
                return;
            }

            throw new Error('Expected error was not thrown');
        });

        it('should fail on quorum reached but without depositor', async () => {
            let allocations = structuredClone(START_ALLOCATIONS);
            // Bob is depositing
            allocations[1].amount = (BigInt(allocations[1].amount) + BigInt(10)).toString(); // 10 USDC - more than deposited

            try {
                // Alice signs and constitutes 100% of quorum, but is not a depositor
                await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Deposit, currentVersion + 1, allocations, { state: 'test' });
            } catch (e) {
                expect((e as Error).message).toMatch(/RPC Error.*incorrect deposit request.*depositor signature is required/i);

                const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
                expect(sessionData).toEqual(START_SESSION_DATA);
                return;
            }

            throw new Error('Expected error was not thrown');
        });

        it('should fail on depositor signature but no quorum', async () => {
            // authenticate Bob's app identity, so that it can sign
            await createAuthSessionWithClearnode(bobWS, bobAppIdentity);

            let allocations = structuredClone(START_ALLOCATIONS);
            // Bob is depositing
            allocations[1].amount = (BigInt(allocations[1].amount) + BigInt(10)).toString(); // 10 USDC - more than deposited

            try {
                // Bob signs and constitutes 0% of quorum, but is a depositor
                await submitAppStateUpdate_v04(bobWS, bobAppIdentity, appSessionId, RPCAppStateIntent.Deposit, currentVersion + 1, allocations, { state: 'test' });
            } catch (e) {
                expect((e as Error).message).toMatch(/RPC Error.*incorrect deposit request.*quorum not reached/i);

                const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, bobAppIdentity, appSessionId);
                expect(sessionData).toEqual(START_SESSION_DATA);
                return;
            }

            throw new Error('Expected error was not thrown');
        });
    });

    describe('withdraw intent', () => {
        it('should fail on zero delta', async () => {
            let allocations = structuredClone(START_ALLOCATIONS); // same as deposited, zero delta

            try {
                await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Withdraw, currentVersion + 1, allocations, { state: 'test' });
            } catch (e) {
                expect((e as Error).message).toMatch(/RPC Error.*incorrect withdrawal request.*non-negative allocations sum delta/i);

                const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
                expect(sessionData).toEqual(START_SESSION_DATA);
                return;
            }

            throw new Error('Expected error was not thrown');
        });

        it('should fail on positive delta', async () => {
            let allocations = structuredClone(START_ALLOCATIONS);
            allocations[0].amount = (BigInt(allocations[0].amount) + BigInt(10)).toString(); // 110 USDC - more than deposited

            try {
                await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Withdraw, currentVersion + 1, allocations, { state: 'test' });
            } catch (e) {
                expect((e as Error).message).toMatch(/RPC Error.*incorrect withdrawal request.*increased allocation for participant/i);

                const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
                expect(sessionData).toEqual(START_SESSION_DATA);
                return;
            }

            throw new Error('Expected error was not thrown');
        });

        it('should fail on positive and negative allocation deltas', async () => {
            let allocations = structuredClone(START_ALLOCATIONS);
            allocations[0].amount = (BigInt(allocations[0].amount) + BigInt(10)).toString(); // 110 USDC - more than deposited
            allocations[1].amount = (BigInt(allocations[1].amount) - BigInt(20)).toString(); // 80 USDC - less than deposited

            try {
                await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Withdraw, currentVersion + 1, allocations, { state: 'test' });
            } catch (e) {
                expect((e as Error).message).toMatch(/RPC Error.*incorrect withdrawal request.*increased allocation for participant/i);

                const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, aliceAppIdentity, appSessionId);
                expect(sessionData).toEqual(START_SESSION_DATA);
                return;
            }

            throw new Error('Expected error was not thrown');
        });

        it('should fail on withdrawing from v0.2 app session', async () => {
            // Create a v0.2 app session (which doesn't support deposits/withdrawals)
            const v02AppSessionId = await createTestAppSession(
                aliceAppIdentity,
                bobAppIdentity,
                aliceAppWS,
                RPCProtocolVersion.NitroRPC_0_2,
                appSessionDepositAmount,
                START_SESSION_DATA
            );

            let allocations = structuredClone(START_ALLOCATIONS);
            allocations[0].amount = (BigInt(allocations[0].amount) - BigInt(10)).toString(); // 90 USDC - less than deposited

            try {
                await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, v02AppSessionId, RPCAppStateIntent.Withdraw, 2, allocations, { state: 'test' });
            } catch (e) {
                expect((e as Error).message).toMatch(/RPC Error.*incorrect request.*specified parameters are not supported in this protocol/i);
                return;
            }

            throw new Error('Expected error was not thrown');
        });

        it('should fail on no quorum reached', async () => {
            // authenticate Bob's app identity, so that it can sign
            await createAuthSessionWithClearnode(bobWS, bobAppIdentity);

            // for Bob to withdraw, he needs to get a balance first
            let allocations = structuredClone(START_ALLOCATIONS);
            allocations[0].amount = (BigInt(allocations[0].amount) - BigInt(50)).toString(); // 50 USDC
            allocations[1].amount = (BigInt(allocations[1].amount) + BigInt(50)).toString(); // 50 USDC

            const INTERMEDIATE_SESSION_DATA = { state: 'intermediate' };
            await submitAppStateUpdate_v04(aliceAppWS, aliceAppIdentity, appSessionId, RPCAppStateIntent.Operate, currentVersion + 1, allocations, INTERMEDIATE_SESSION_DATA);

            // Bob is withdrawing
            allocations[1].amount = (BigInt(allocations[1].amount) - BigInt(10)).toString(); // 40 USDC - less than before

            try {
                // Bob signs and constitutes 0% of quorum
                await submitAppStateUpdate_v04(aliceAppWS, bobAppIdentity, appSessionId, RPCAppStateIntent.Withdraw, currentVersion + 2, allocations, { state: 'test' });
            } catch (e) {
                expect((e as Error).message).toMatch(/RPC Error.*incorrect withdrawal request.*quorum not reached/i);

                const { sessionData } = await fetchAndParseAppSessions(aliceAppWS, bobAppIdentity, appSessionId);
                expect(sessionData).toEqual(INTERMEDIATE_SESSION_DATA);
                return;
            }

            throw new Error('Expected error was not thrown');
        });
    });
});
