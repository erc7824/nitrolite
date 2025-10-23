import { DatabaseUtils } from '@/databaseUtils';
import { Identity } from '@/identity';
import { CONFIG } from '@/setup';
import { getAuthChallengePredicate, getAuthVerifyPredicate, TestWebSocket } from '@/ws';
import {
    AuthChallengeResponse,
    AuthRequestParams,
    createAuthRequestMessage,
    createAuthVerifyMessage,
    createAuthVerifyMessageWithJWT,
    createEIP712AuthMessageSigner,
    parseAuthChallengeResponse,
    parseAuthVerifyResponse,
} from '@erc7824/nitrolite';

describe('Clearnode Authentication', () => {
    let ws: TestWebSocket;

    afterAll(() => {
        ws.close();
        const databaseUtils = new DatabaseUtils();
        databaseUtils.resetClearnodeState();
        databaseUtils.close();
    });

    const identity = new Identity(CONFIG.IDENTITIES[0].WALLET_PK, CONFIG.IDENTITIES[0].SESSION_PK);

    const authRequestParams: AuthRequestParams = {
        address: identity.walletAddress,
        session_key: identity.sessionAddress,
        application: 'clearnode',
        expire: String(Math.floor(Date.now() / 1000) + 3600), // 1 hour expiration
        scope: 'console',
        allowances: [],
    };

    const eip712MessageSigner = createEIP712AuthMessageSigner(
        identity.walletClient,
        {
            scope: authRequestParams.scope,
            application: authRequestParams.application,
            participant: authRequestParams.session_key,
            expire: authRequestParams.expire,
            allowances: authRequestParams.allowances,
        },
        {
            name: 'clearnode',
        }
    );

    let parsedChallengeResponse: AuthChallengeResponse;
    let jwtToken: string;

    it('should receive challenge', async () => {
        ws = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
        await ws.connect();

        const msg = await createAuthRequestMessage(authRequestParams);
        const response = await ws.sendAndWaitForResponse(msg, getAuthChallengePredicate(), 1000);
        expect(response).toBeDefined();

        parsedChallengeResponse = parseAuthChallengeResponse(response);
        expect(parsedChallengeResponse.params.challengeMessage).toBeDefined();
    });

    // TODO: there are some issues with createAuthVerifyMessageFromChallenge, fix it
    // it('should verify identity with EIP712 signature from challenge string', async () => {
    //     const msg = await createAuthVerifyMessageFromChallenge(
    //         eip712MessageSigner,
    //         parsedChallengeResponse.params.challengeMessage
    //     );
    //     const response = await ws.sendAndWaitForResponse(msg, AuthVerifyPredicate, 1000);
    //     expect(response).toBeDefined();

    //     const parsedAuthVerifyResponse = parseRPCResponse(response) as AuthVerifyRPCResponse;
    //     expect(parsedAuthVerifyResponse.params.success).toBeTruthy();
    // });

    it('should verify identity with EIP712 signature from challenge response', async () => {
        const msg = await createAuthVerifyMessage(eip712MessageSigner, parsedChallengeResponse);
        const response = await ws.sendAndWaitForResponse(msg, getAuthVerifyPredicate(), 1000);
        expect(response).toBeDefined();

        const parsedAuthVerifyResponse = parseAuthVerifyResponse(response);

        expect(parsedAuthVerifyResponse.params.success).toBe(true);
        expect(parsedAuthVerifyResponse.params.sessionKey).toBe(authRequestParams.session_key);
        expect(parsedAuthVerifyResponse.params.address).toBe(authRequestParams.address);
        expect(parsedAuthVerifyResponse.params.jwtToken).toBeDefined();

        jwtToken = parsedAuthVerifyResponse.params.jwtToken;
    });

    it('should verify identity with JWT token', async () => {
        // Recreate the WebSocket connection to simulate a new session
        ws.close();
        ws = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
        await ws.connect();

        const msg = await createAuthVerifyMessageWithJWT(jwtToken);
        const response = await ws.sendAndWaitForResponse(msg, getAuthVerifyPredicate(), 1000);
        expect(response).toBeDefined();

        const parsedAuthVerifyResponse = parseAuthVerifyResponse(response);

        expect(parsedAuthVerifyResponse.params.success).toBe(true);
        expect(parsedAuthVerifyResponse.params.sessionKey).toBe(authRequestParams.session_key);
        expect(parsedAuthVerifyResponse.params.address).toBe(authRequestParams.address);
        expect(parsedAuthVerifyResponse.params.jwtToken).toBeUndefined();
    });
});
