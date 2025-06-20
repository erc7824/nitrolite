import {
    AuthRequest,
    createEIP712AuthMessageSigner,
    Allowance,
    createAuthRequestMessage,
    parseRPCResponse,
    AuthChallengeRPCResponse,
    createAuthVerifyMessage,
} from '@erc7824/nitrolite';
import { Identity } from './identity';
import { getAuthChallengePredicate, getAuthVerifyPredicate, TestWebSocket } from './ws';

export const createAuthSessionWithClearnode = async (
    ws: TestWebSocket,
    identity: Identity,
    authRequestParams?: AuthRequest
) => {
    authRequestParams = authRequestParams || {
        wallet: identity.walletAddress,
        participant: identity.sessionAddress,
        app_name: 'Test Domain',
        expire: String(Math.floor(Date.now() / 1000) + 3600), // 1 hour expiration
        scope: 'console',
        application: '0x9965507D1a55bcC2695C58ba16FB37d819B0A4dc',
        allowances: [],
    };

    const eip712MessageSigner = createEIP712AuthMessageSigner(
        identity.walletClient,
        {
            scope: authRequestParams.scope,
            application: authRequestParams.application,
            participant: authRequestParams.participant,
            expire: authRequestParams.expire,
            allowances: authRequestParams.allowances.map((a: Allowance) => ({
                asset: a.symbol,
                amount: a.amount,
            })),
        },
        {
            name: authRequestParams.app_name,
        }
    );

    const authRequestMsg = await createAuthRequestMessage(authRequestParams);
    const authRequestResponse = await ws.sendAndWaitForResponse(authRequestMsg, getAuthChallengePredicate(), 1000);

    const authRequestParsedResponse = parseRPCResponse(authRequestResponse) as AuthChallengeRPCResponse;

    const authVerifyMsg = await createAuthVerifyMessage(eip712MessageSigner, authRequestParsedResponse);
    await ws.sendAndWaitForResponse(authVerifyMsg, getAuthVerifyPredicate(), 1000);
};
