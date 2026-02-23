describe('Clearnode Connection', () => {
    // TODO: find a public request that can be used to test the connection
    test('This test will be skipped', () => {
        expect(true).toBe(true);
    });
    // let ws: TestWebSocket;

    // beforeEach(() => {
    //     ws = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
    // });

    // afterEach(() => {
    //     ws.close();
    // });

    // it('should receive pong response from the Clearnode server', async () => {
    //     await ws.connect();

    //     const msg = JSON.stringify({ req: [0, 'ping', [], Date.now()], sig: [] });
    //     const response = await ws.sendAndWaitForResponse(msg, getPongPredicate(), 1000);

    //     expect(response).toBeDefined();
    // });

    // it('should handle connection timeout', async () => {
    //     await ws.connect();

    //     await expect(ws.waitForMessage((data) => data === 'nonexistent', undefined, 500)).rejects.toThrow(
    //         'Timeout waiting for message after 500ms. Request ID: N/A'
    //     );
    // });
});
