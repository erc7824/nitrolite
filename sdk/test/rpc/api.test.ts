import { describe, test, expect, jest } from "@jest/globals";
import { Address, Hex } from "viem";
import {
    createAuthRequestMessage,
    createAuthVerifyMessageFromChallenge,
    createAuthVerifyMessage,
    createPingMessage,
    createGetConfigMessage,
    createGetLedgerBalancesMessage,
    createGetAppDefinitionMessage,
    createAppSessionMessage,
    createCloseAppSessionMessage,
    createApplicationMessage,
    createCloseChannelMessage,
} from "../../src/rpc/api";
import { CreateAppSessionRequest, MessageSigner } from "../../src/rpc/types";

describe("API message creators", () => {
    const signer: MessageSigner = jest.fn(async () => "0xsig" as Hex);
    const requestId = 42;
    const timestamp = 1000;
    const clientAddress = "0x000000000000000000000000000000000000abcd" as Hex;
    const channelId = "0x000000000000000000000000000000000000cdef" as Hex;
    const appId = "0x000000000000000000000000000000000000ffff" as Hex;
    const fundDestination = "0x" as Address;
    const sampleIntent = [1, 2, 3];

    afterEach(() => {
        jest.clearAllMocks();
    });

    test("createAuthRequestMessage", async () => {
        const msgStr = await createAuthRequestMessage(signer, clientAddress, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, "auth_request", [clientAddress], timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, "auth_request", [clientAddress], timestamp],
            sig: ["0xsig"],
        });
    });

    test("createAuthVerifyMessageFromChallenge", async () => {
        const challenge = "challenge123";
        const msgStr = await createAuthVerifyMessageFromChallenge(signer, clientAddress, challenge, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, "auth_verify", [{ address: clientAddress, challenge }], timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, "auth_verify", [{ address: clientAddress, challenge }], timestamp],
            sig: ["0xsig"],
        });
    });

    describe("createAuthVerifyMessage", () => {
        const rawResponse = JSON.stringify({
            res: [999, "auth_challenge", [{ challenge_message: "msg" }], 200],
        });

        test("successful challenge flow", async () => {
            const msgStr = await createAuthVerifyMessage(signer, rawResponse, clientAddress, requestId, timestamp);
            expect(signer).toHaveBeenCalledWith([requestId, "auth_verify", [{ address: clientAddress, challenge: "msg" }], timestamp]);
            const parsed = JSON.parse(msgStr);
            expect(parsed).toEqual({
                req: [requestId, "auth_verify", [{ address: clientAddress, challenge: "msg" }], timestamp],
                sig: ["0xsig"],
            });
        });

        test("throws on invalid response", async () => {
            await expect(createAuthVerifyMessage(signer, "{}", clientAddress, requestId, timestamp)).rejects.toThrow(
                "Invalid auth_challenge response"
            );
        });

        test("throws on wrong method", async () => {
            const wrong = JSON.stringify({
                res: [100, "other", [{ challenge_message: "msg" }], 200],
            });
            await expect(createAuthVerifyMessage(signer, wrong, clientAddress, requestId, timestamp)).rejects.toThrow(
                "Expected 'auth_challenge' method"
            );
        });
    });

    test("createPingMessage", async () => {
        const msgStr = await createPingMessage(signer, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, "ping", [], timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, "ping", [], timestamp],
            sig: ["0xsig"],
        });
    });

    test("createGetConfigMessage", async () => {
        const msgStr = await createGetConfigMessage(signer, channelId, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, "get_config", [], timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, "get_config", [], timestamp],
            sig: ["0xsig"],
        });
    });

    test("createGetLedgerBalancesMessage", async () => {
        const ledgerParams = [{ acc: channelId }];
        const msgStr = await createGetLedgerBalancesMessage(signer, channelId, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, "get_ledger_balances", ledgerParams, timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, "get_ledger_balances", ledgerParams, timestamp],
            sig: ["0xsig"],
        });
    });

    test("createGetAppDefinitionMessage", async () => {
        const appParams = [{ acc: appId }];
        const msgStr = await createGetAppDefinitionMessage(signer, appId, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, "get_app_definition", appParams, timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, "get_app_definition", appParams, timestamp],
            sig: ["0xsig"],
        });
    });

    test("createAppSessionMessage", async () => {
        const params: CreateAppSessionRequest[] = [
            {
                definition: {
                    protocol: "p",
                    participants: [],
                    weights: [],
                    quorum: 0,
                    challenge: 0,
                    nonce: 0,
                },
                token: "0x",
                // @ts-ignore
                allocation: [100, 0],
            },
        ];
        // @ts-ignore
        const msgStr = await createAppSessionMessage(signer, params, sampleIntent, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, "create_app_session", params, timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, "create_app_session", params, timestamp],
            int: sampleIntent,
            sig: ["0xsig"],
        });
    });

    test("createCloseAppSessionMessage", async () => {
        const closeParams = [{ appId, allocation: [] }];
        // @ts-ignore
        const msgStr = await createCloseAppSessionMessage(signer, closeParams, sampleIntent, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, "close_app_session", closeParams, timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, "close_app_session", closeParams, timestamp],
            int: sampleIntent,
            sig: ["0xsig"],
        });
    });

    test("createApplicationMessage", async () => {
        const messageParams = ["hello"];
        const msgStr = await createApplicationMessage(signer, appId, messageParams, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, "message", messageParams, timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, "message", messageParams, timestamp],
            acc: appId,
            sig: ["0xsig"],
        });
    });

    test("createCloseChannelMessage", async () => {
        const msgStr = await createCloseChannelMessage(signer, channelId, fundDestination, requestId, timestamp);
        expect(signer).toHaveBeenCalledWith([requestId, "close_channel", [{ channel_id: channelId, funds_destination: fundDestination }], timestamp]);
        const parsed = JSON.parse(msgStr);
        expect(parsed).toEqual({
            req: [requestId, "close_channel", [{ channel_id: channelId, funds_destination: fundDestination }], timestamp],
            sig: ["0xsig"],
        });
    });
});
