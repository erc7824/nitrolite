import { createAuthSessionWithClearnode } from '@/auth';
import { DatabaseUtils } from '@/databaseUtils';
import { Identity } from '@/identity';
import { CONFIG } from '@/setup';
import { getGetUserTagPredicate, TestWebSocket } from "@/ws";
import { createGetUserTagMessage, parseGetUserTagResponse } from "@erc7824/nitrolite";

describe('Get User Tag Integration', () => {
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
        await databaseUtils.resetClearnodeState();
        await databaseUtils.close();
    });

    describe('createGetUserTagMessage', () => {
        it('should successfully request user tag', async () => {
            const msg = await createGetUserTagMessage(identity.messageSKSigner);

            const response = await ws.sendAndWaitForResponse(msg, getGetUserTagPredicate(), 5000);

            expect(response).toBeDefined();

            const parsedResponse = parseGetUserTagResponse(response);
            expect(parsedResponse).toBeDefined();
            expect(parsedResponse.params).toBeDefined();
            expect(parsedResponse.params.tag).toBeDefined();
            expect(typeof parsedResponse.params.tag).toBe('string');
            expect(parsedResponse.params.tag.length).toBeGreaterThan(0);
        });

        it('should return consistent user tag across multiple requests', async () => {
            // First request
            const msg1 = await createGetUserTagMessage(identity.messageSKSigner);
            const response1 = await ws.sendAndWaitForResponse(msg1, getGetUserTagPredicate(), 5000);
            const parsedResponse1 = parseGetUserTagResponse(response1);

            // Second request
            const msg2 = await createGetUserTagMessage(identity.messageSKSigner);
            const response2 = await ws.sendAndWaitForResponse(msg2, getGetUserTagPredicate(), 5000);
            const parsedResponse2 = parseGetUserTagResponse(response2);

            expect(parsedResponse1.params.tag).toBe(parsedResponse2.params.tag);
        });

        it('should return valid tag format', async () => {
            const msg = await createGetUserTagMessage(identity.messageSKSigner);
            const response = await ws.sendAndWaitForResponse(msg, getGetUserTagPredicate(), 5000);
            const parsedResponse = parseGetUserTagResponse(response);

            // Verify the tag format matches expected pattern (e.g., 'UX123D8C')
            expect(parsedResponse.params.tag).toMatch(/^[A-Z0-9]+$/);
            expect(parsedResponse.params.tag.length).toBeGreaterThan(3);
        });
    });
});
