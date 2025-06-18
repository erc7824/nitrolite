import { CONFIG } from '@/setup';
import { PongPredicate, TestWebSocket } from '@/ws';

describe('Clearnode Connection', () => {
    let ws: TestWebSocket;

    beforeEach(() => {
        ws = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
    });

    afterEach(() => {
        ws.close();
    });

    it('should receive pong response from the Clearnode server', async () => {
        await ws.connect();

        const msg = JSON.stringify({ req: [0, 'ping', [], Date.now()], sig: [] });
        const response = await ws.sendAndWaitForResponse(msg, PongPredicate, 1000);

        expect(response).toBeDefined();
    });

    it('should handle connection timeout', async () => {
        await ws.connect();

        await expect(ws.waitForMessage((data) => data === 'nonexistent', 500)).rejects.toThrow(
            'Timeout waiting for message after 500ms'
        );
    });
});
