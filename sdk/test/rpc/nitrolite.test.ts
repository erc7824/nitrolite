import { describe, test, expect, jest } from "@jest/globals";
import { Address, Hex } from "viem";
import { NitroliteRPC } from "../../src/rpc/nitrolite";
import { NitroliteRPCMessage, MessageSigner, SingleMessageVerifier, MultiMessageVerifier, RequestData, ResponsePayload } from "../../src/rpc/types";

describe("NitroliteRPC", () => {
    beforeAll(() => {
        jest.spyOn(console, "error").mockImplementation(() => {});
    });
    afterAll(() => {
        (console.error as jest.Mock).mockRestore();
    });
    describe("createRequest", () => {
        test("should create a valid request message without intent", () => {
            const requestId = 12345;
            const method = "test_method";
            const params = ["param1", "param2"];
            const timestamp = 1619876543210;

            const result = NitroliteRPC.createRequest(requestId, method, params, timestamp);

            expect(result).toEqual({
                req: [requestId, method, params, timestamp],
            });
        });

        test("should create a valid request message with intent", () => {
            const requestId = 12345;
            const method = "test_method";
            const params = ["param1", "param2"];
            const timestamp = 1619876543210;
            const intent = [1, 2, 3];

            const result = NitroliteRPC.createRequest(requestId, method, params, timestamp, intent);

            expect(result).toEqual({
                req: [requestId, method, params, timestamp],
                int: intent,
            });
        });

        test("should use default values when not provided", () => {
            jest.spyOn(global.Date, "now").mockReturnValue(1619876543210);
            const result = NitroliteRPC.createRequest(undefined, "test_method");

            expect(result.req).toBeDefined();
            expect(result.req![0]).toBeGreaterThan(0);
            expect(result.req![1]).toBe("test_method");
            expect(result.req![2]).toEqual([]);
            expect(result.req![3]).toBe(1619876543210);
        });
    });

    describe("createAppRequest", () => {
        test("should create a valid application request message", () => {
            const requestId = 12345;
            const method = "test_method";
            const params = ["param1", "param2"];
            const timestamp = 1619876543210;
            const accountId = "0xaccountId" as Hex;

            const result = NitroliteRPC.createAppRequest(requestId, method, params, timestamp, accountId);

            expect(result).toEqual({
                req: [requestId, method, params, timestamp],
                acc: accountId,
            });
        });

        test("should create a valid application request message with intent", () => {
            const requestId = 12345;
            const method = "test_method";
            const params = ["param1", "param2"];
            const timestamp = 1619876543210;
            const accountId = "0xaccountId" as Hex;
            const intent = [1, 2, 3];

            const result = NitroliteRPC.createAppRequest(requestId, method, params, timestamp, accountId, intent);

            expect(result).toEqual({
                req: [requestId, method, params, timestamp],
                acc: accountId,
                int: intent,
            });
        });
    });

    describe("parseResponse", () => {
        test("should parse a valid response message string", () => {
            const responseStr = JSON.stringify({
                res: [12345, "test_method", ["result1", "result2"], 1619876543210],
            });

            const result = NitroliteRPC.parseResponse(responseStr);

            expect(result).toEqual({
                isValid: true,
                isError: false,
                requestId: 12345,
                method: "test_method",
                data: ["result1", "result2"],
                timestamp: 1619876543210,
            });
        });

        test("should parse a valid response message object", () => {
            const responseObj = {
                res: [12345, "test_method", ["result1", "result2"], 1619876543210],
            };

            const result = NitroliteRPC.parseResponse(responseObj);

            expect(result).toEqual({
                isValid: true,
                isError: false,
                requestId: 12345,
                method: "test_method",
                data: ["result1", "result2"],
                timestamp: 1619876543210,
            });
        });

        test("should parse a valid response message with acc field", () => {
            const responseObj = {
                res: [12345, "test_method", ["result1", "result2"], 1619876543210],
                acc: "0xaccountId" as Hex,
            };

            const result = NitroliteRPC.parseResponse(responseObj);

            expect(result).toEqual({
                isValid: true,
                isError: false,
                requestId: 12345,
                method: "test_method",
                data: ["result1", "result2"],
                acc: "0xaccountId",
                timestamp: 1619876543210,
            });
        });

        test("should parse a valid response message with intent", () => {
            const responseObj = {
                res: [12345, "test_method", ["result1", "result2"], 1619876543210],
                int: [1, 2, 3],
            };

            const result = NitroliteRPC.parseResponse(responseObj);

            expect(result).toEqual({
                isValid: true,
                isError: false,
                requestId: 12345,
                method: "test_method",
                data: ["result1", "result2"],
                int: [1, 2, 3],
                timestamp: 1619876543210,
            });
        });

        test("should handle error responses correctly", () => {
            const errorResponse = {
                res: [12345, "error", [{ error: "Something went wrong" }], 1619876543210],
            };

            const result = NitroliteRPC.parseResponse(errorResponse);

            expect(result).toEqual({
                isValid: true,
                isError: true,
                requestId: 12345,
                method: "error",
                data: { error: "Something went wrong" },
                timestamp: 1619876543210,
            });
        });

        test("should return invalid for malformed JSON", () => {
            const result = NitroliteRPC.parseResponse("invalid json");

            expect(result.isValid).toBe(false);
            expect(result.error).toBe("Message parsing failed");
        });

        test("should return invalid for missing res field", () => {
            const result = NitroliteRPC.parseResponse({ something: "else" });

            expect(result.isValid).toBe(false);
            expect(result.error).toBe("Invalid message structure: Missing or invalid 'res' array.");
        });

        test("should return invalid for incorrectly sized res array", () => {
            const result = NitroliteRPC.parseResponse({ res: [1, 2, 3] });

            expect(result.isValid).toBe(false);
            expect(result.error).toBe("Invalid message structure: Missing or invalid 'res' array.");
        });

        test("should return invalid for incorrect types in res array", () => {
            const result = NitroliteRPC.parseResponse({
                res: ["not-a-number", 123, "not-an-array", "not-a-number"],
            });

            expect(result.isValid).toBe(false);
            expect(result.error).toBe("Invalid 'res' payload structure or types.");
        });

        test("should return invalid for malformed error response", () => {
            const result = NitroliteRPC.parseResponse({
                res: [12345, "error", ["not an error object"], 1619876543210],
            });

            expect(result.isValid).toBe(false);
            expect(result.error).toBe("Malformed error response payload.");
        });
    });

    describe("signRequestMessage", () => {
        test("should sign a request message and add signature to the message", async () => {
            const mockSigner = jest.fn<(data: RequestData | ResponsePayload) => Promise<Hex>>().mockResolvedValue("0xsignature" as Hex);
            const request: NitroliteRPCMessage = {
                req: [12345, "test_method", ["param1", "param2"], 1619876543210],
            };

            const result = await NitroliteRPC.signRequestMessage(request, mockSigner);

            expect(mockSigner).toHaveBeenCalledWith(request.req);
            expect(result).toEqual({
                req: [12345, "test_method", ["param1", "param2"], 1619876543210],
                sig: ["0xsignature"],
            });
            expect(result).toBe(request);
        });

        test("should throw if message doesn't have req field", async () => {
            const mockSigner: MessageSigner = jest.fn<MessageSigner>().mockResolvedValue("0xsignature" as Hex);
            const invalidMessage = { res: [1, 2, 3, 4] } as unknown as NitroliteRPCMessage;

            await expect(NitroliteRPC.signRequestMessage(invalidMessage, mockSigner)).rejects.toThrow(
                "signRequestMessage can only sign request messages containing 'req'."
            );
        });
    });

    describe("verifySingleSignature", () => {
        test("should verify a correctly signed message", async () => {
            const mockVerifier = jest.fn<SingleMessageVerifier>().mockResolvedValue(true);
            const message: NitroliteRPCMessage = {
                req: [12345, "test_method", ["param1", "param2"], 1619876543210],
                sig: ["0xsignature" as Hex],
            };
            const expectedSigner = "0xsigner" as Address;

            const result = await NitroliteRPC.verifySingleSignature(message, expectedSigner, mockVerifier);

            expect(mockVerifier).toHaveBeenCalledWith(message.req, "0xsignature", expectedSigner);
            expect(result).toBe(true);
        });

        test("should return false for message with no signatures", async () => {
            const mockVerifier = jest.fn<SingleMessageVerifier>();
            const message: NitroliteRPCMessage = {
                req: [12345, "test_method", ["param1", "param2"], 1619876543210],
            };
            const expectedSigner = "0xsigner" as Address;

            const result = await NitroliteRPC.verifySingleSignature(message, expectedSigner, mockVerifier);

            expect(mockVerifier).not.toHaveBeenCalled();
            expect(result).toBe(false);
        });

        test("should handle verification errors", async () => {
            const mockVerifier = jest.fn<SingleMessageVerifier>().mockImplementation(() => {
                throw new Error("Verification error");
            });
            const message: NitroliteRPCMessage = {
                req: [12345, "test_method", ["param1", "param2"], 1619876543210],
                sig: ["0xsignature" as Hex],
            };
            const expectedSigner = "0xsigner" as Address;

            const result = await NitroliteRPC.verifySingleSignature(message, expectedSigner, mockVerifier);

            expect(mockVerifier).toHaveBeenCalled();
            expect(result).toBe(false);
        });

        test("should verify a response message", async () => {
            const mockVerifier = jest.fn<SingleMessageVerifier>().mockResolvedValue(true);
            const message: NitroliteRPCMessage = {
                res: [12345, "test_method", ["result1", "result2"], 1619876543210],
                sig: ["0xsignature" as Hex],
            };
            const expectedSigner = "0xsigner" as Address;

            const result = await NitroliteRPC.verifySingleSignature(message, expectedSigner, mockVerifier);

            expect(mockVerifier).toHaveBeenCalledWith(message.res, "0xsignature", expectedSigner);
            expect(result).toBe(true);
        });

        test("should verify only the first signature when multiple are present", async () => {
            console.warn = jest.fn();
            const mockVerifier = jest.fn<SingleMessageVerifier>().mockResolvedValue(true);
            const message: NitroliteRPCMessage = {
                req: [12345, "test_method", ["param1", "param2"], 1619876543210],
                sig: ["0xsignature1" as Hex, "0xsignature2" as Hex],
            };
            const expectedSigner = "0xsigner" as Address;

            const result = await NitroliteRPC.verifySingleSignature(message, expectedSigner, mockVerifier);

            expect(mockVerifier).toHaveBeenCalledWith(message.req, "0xsignature1", expectedSigner);
            expect(console.warn).toHaveBeenCalled();
            expect(result).toBe(true);
        });
    });

    describe("verifyMultipleSignatures", () => {
        test("should verify multiple signatures correctly", async () => {
            const mockVerifier = jest.fn<MultiMessageVerifier>().mockResolvedValue(true);
            const message: NitroliteRPCMessage = {
                req: [12345, "test_method", ["param1", "param2"], 1619876543210],
                sig: ["0xsig1" as Hex, "0xsig2" as Hex],
            };
            const expectedSigners = ["0xsigner1" as Address, "0xsigner2" as Address];

            const result = await NitroliteRPC.verifyMultipleSignatures(message, expectedSigners, mockVerifier);

            expect(mockVerifier).toHaveBeenCalledWith(message.req, message.sig, expectedSigners);
            expect(result).toBe(true);
        });

        test("should return false for message with no signatures", async () => {
            const mockVerifier = jest.fn<MultiMessageVerifier>();
            const message: NitroliteRPCMessage = {
                req: [12345, "test_method", ["param1", "param2"], 1619876543210],
            };
            const expectedSigners = ["0xsigner1" as Address, "0xsigner2" as Address];

            const result = await NitroliteRPC.verifyMultipleSignatures(message, expectedSigners, mockVerifier);

            expect(mockVerifier).not.toHaveBeenCalled();
            expect(result).toBe(false);
        });

        test("should handle verification errors", async () => {
            const mockVerifier = jest.fn<MultiMessageVerifier>().mockImplementation(() => {
                throw new Error("Verification error");
            });
            const message: NitroliteRPCMessage = {
                req: [12345, "test_method", ["param1", "param2"], 1619876543210],
                sig: ["0xsig1" as Hex, "0xsig2" as Hex],
            };
            const expectedSigners = ["0xsigner1" as Address, "0xsigner2" as Address];

            const result = await NitroliteRPC.verifyMultipleSignatures(message, expectedSigners, mockVerifier);

            expect(mockVerifier).toHaveBeenCalled();
            expect(result).toBe(false);
        });
    });
});
