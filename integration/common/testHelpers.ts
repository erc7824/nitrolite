import { Identity } from '@/identity';
import { TestNitroliteClient } from '@/nitroliteClient';
import { CONFIG } from '@/setup';
import { TestWebSocket, getCreateAppSessionPredicate, getGetLedgerBalancesPredicate } from '@/ws';
import { createAuthSessionWithClearnode } from '@/auth';
import {
    RPCAppDefinition,
    RPCProtocolVersion,
    createAppSessionMessage,
    parseCreateAppSessionResponse,
    RPCChannelStatus,
    createGetLedgerBalancesMessage,
    parseGetLedgerBalancesResponse,
    RPCBalance,
} from '@erc7824/nitrolite';
import { Hex } from 'viem';

export function toRaw(amount: bigint, decimals: number = 6): bigint {
    return amount * BigInt(10 ** decimals);
}

/**
 * Creates test channels with the specified deposit amount.
 */
export async function createTestChannels(
    params: {
        client: TestNitroliteClient;
        ws: TestWebSocket;
    }[],
    depositAmount: bigint
): Promise<Hex[]> {
    const channelIds: Hex[] = [];

    for (const { client, ws } of params) {
        const { params: channelParams } = await client.createAndWaitForChannel(ws, {
            tokenAddress: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            amount: depositAmount,
        });

        channelIds.push(channelParams.channelId);
    }

    return channelIds;
}

/**
 * Authenticates a participant's app identity with allowances for deposits.
 */
export async function authenticateAppWithAllowances(
    participantAppWS: TestWebSocket,
    participantAppIdentity: Identity,
    decimalDepositAmount: bigint
): Promise<void> {
    await createAuthSessionWithClearnode(participantAppWS, participantAppIdentity, {
        address: participantAppIdentity.walletAddress,
        session_key: participantAppIdentity.sessionAddress,
        app_name: 'App Domain',
        expire: String(Math.floor(Date.now() / 1000) + 3600), // 1 hour expiration
        scope: 'console',
        application: '0x9965507D1a55bcC2695C58ba16FB37d819B0A4dc', // random address, no use for now
        allowances: [
            {
                asset: 'usdc',
                amount: decimalDepositAmount.toString(),
            },
        ],
    });
}

/**
 * Creates a test app session between Alice and Bob with the specified protocol version.
 */
export async function createTestAppSession(
    aliceAppIdentity: Identity,
    bobAppIdentity: Identity,
    aliceAppWS: TestWebSocket,
    protocol: RPCProtocolVersion,
    decimalDepositAmount: bigint,
    sessionData: object
): Promise<string> {
    const definition: RPCAppDefinition = {
        protocol,
        participants: [aliceAppIdentity.walletAddress, bobAppIdentity.walletAddress],
        weights: [100, 0],
        quorum: 100,
        challenge: 0,
        nonce: Date.now(),
    };

    const allocations = [
        {
            participant: aliceAppIdentity.walletAddress,
            asset: 'USDC',
            amount: decimalDepositAmount.toString(),
        },
        {
            participant: bobAppIdentity.walletAddress,
            asset: 'USDC',
            amount: '0',
        },
    ];

    const createAppSessionMsg = await createAppSessionMessage(aliceAppIdentity.messageSigner, {
        definition,
        allocations,
        session_data: JSON.stringify(sessionData),
    });

    const createAppSessionResponse = await aliceAppWS.sendAndWaitForResponse(
        createAppSessionMsg,
        getCreateAppSessionPredicate(),
        1000
    );

    const createAppSessionParsedResponse = parseCreateAppSessionResponse(createAppSessionResponse);

    expect(createAppSessionParsedResponse).toBeDefined();
    expect(createAppSessionParsedResponse.params.appSessionId).toBeDefined();
    expect(createAppSessionParsedResponse.params.status).toBe(RPCChannelStatus.Open);
    expect(createAppSessionParsedResponse.params.version).toBeDefined();

    return createAppSessionParsedResponse.params.appSessionId;
}

/**
 * Fetches and returns ledger balances for the given app identity.
 * Expects at least one balance to exist.
 * Returns an array of balances.
 * */
export async function getLedgerBalances(appIdentity: Identity, appWS: TestWebSocket): Promise<RPCBalance[]> {
    const getLedgerBalancesMsg = await createGetLedgerBalancesMessage(
        appIdentity.messageSigner,
        appIdentity.walletAddress
    );
    const getLedgerBalancesResponse = await appWS.sendAndWaitForResponse(
        getLedgerBalancesMsg,
        getGetLedgerBalancesPredicate(),
        1000
    );

    const getLedgerBalancesParsedResponse = parseGetLedgerBalancesResponse(getLedgerBalancesResponse);
    expect(getLedgerBalancesParsedResponse).toBeDefined();

    return getLedgerBalancesParsedResponse.params.ledgerBalances;
}
