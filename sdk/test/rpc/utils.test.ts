import { describe, test, expect, jest } from "@jest/globals";
import {
    getCurrentTimestamp,
    generateRequestId,
    getRequestId,
    getMethod,
    getParams,
    getResult,
    getTimestamp,
    getError,
    toBytes,
    isValidResponseTimestamp,
    isValidResponseRequestId,
} from "../../src/rpc/utils";
import { NitroliteRPCMessage } from "../../src/rpc/types";

describe("RPC Utils", () => {
    describe("getCurrentTimestamp", () => {
        test("should return the current timestamp", () => {
            jest.spyOn(Date, "now").mockReturnValue(1234567890);
            expect(getCurrentTimestamp()).toBe(1234567890);
        });
    });

    describe("generateRequestId", () => {
        test("should generate a unique request ID", () => {
            jest.spyOn(Date, "now").mockReturnValue(1234567890);
            jest.spyOn(Math, "random").mockReturnValue(0.5);
            expect(generateRequestId()).toBe(1234567890 + 5000);
        });
    });

    describe("getRequestId", () => {
        test("should extract request ID from req field", () => {
            const message = { req: [123, "method", [], 456] };
            expect(getRequestId(message)).toBe(123);
        });

        test("should extract request ID from res field", () => {
            const message = { res: [123, "method", [], 456] };
            expect(getRequestId(message)).toBe(123);
        });

        test("should extract request ID from err field", () => {
            const message = { err: [123, "error", "message", 456] };
            expect(getRequestId(message)).toBe(123);
        });

        test("should return undefined if no ID found", () => {
            const message = { other: "value" };
            expect(getRequestId(message)).toBeUndefined();
        });
    });

    describe("getMethod", () => {
        test("should extract method from req field", () => {
            const message = { req: [123, "method", [], 456] };
            expect(getMethod(message)).toBe("method");
        });

        test("should extract method from res field", () => {
            const message = { res: [123, "method", [], 456] };
            expect(getMethod(message)).toBe("method");
        });

        test("should return undefined if no method found", () => {
            const message = { other: "value" };
            expect(getMethod(message)).toBeUndefined();
        });
    });

    describe("getParams", () => {
        test("should extract params from req field", () => {
            const params = ["param1", "param2"];
            const message = { req: [123, "method", params, 456] };
            expect(getParams(message)).toBe(params);
        });

        test("should return empty array if no params found", () => {
            const message = { req: [123, "method", null, 456] };
            expect(getParams(message)).toEqual([]);
        });

        test("should return empty array if no req field", () => {
            const message = { other: "value" };
            expect(getParams(message)).toEqual([]);
        });
    });

    describe("getResult", () => {
        test("should extract result from res field", () => {
            const result = ["result1", "result2"];
            const message = { res: [123, "method", result, 456] };
            expect(getResult(message)).toBe(result);
        });

        test("should return empty array if no result found", () => {
            const message = { res: [123, "method", null, 456] };
            expect(getResult(message)).toEqual([]);
        });

        test("should return empty array if no res field", () => {
            const message = { other: "value" };
            expect(getResult(message)).toEqual([]);
        });
    });

    describe("getTimestamp", () => {
        test("should extract timestamp from req field", () => {
            const message = { req: [123, "method", [], 456] };
            expect(getTimestamp(message)).toBe(456);
        });

        test("should extract timestamp from res field", () => {
            const message = { res: [123, "method", [], 456] };
            expect(getTimestamp(message)).toBe(456);
        });

        test("should extract timestamp from err field", () => {
            const message = { err: [123, "error", "message", 456] };
            expect(getTimestamp(message)).toBe(456);
        });

        test("should return undefined if no timestamp found", () => {
            const message = { other: "value" };
            expect(getTimestamp(message)).toBeUndefined();
        });
    });

    describe("getError", () => {
        test("should extract error details from err field", () => {
            const message = { err: [123, 400, "Bad Request", 456] };
            expect(getError(message)).toEqual({
                code: 400,
                message: "Bad Request",
            });
        });

        test("should return undefined if no err field", () => {
            const message = { other: "value" };
            expect(getError(message)).toBeUndefined();
        });
    });

    describe("toBytes", () => {
        test("should convert string values to bytes", () => {
            const values = ["value1", "value2"];
            const result = toBytes(values);
            expect(result[0]).toBe("0x76616c756531");
            expect(result[1]).toBe("0x76616c756532");
        });

        test("should convert non-string values to JSON bytes", () => {
            const values = [{ key: "value" }, 123];
            const result = toBytes(values);
            expect(result[0]).toBe("0x7b226b6579223a2276616c7565227d");
            expect(result[1]).toBe("0x313233");
        });
    });

    describe("isValidResponseTimestamp", () => {
        test("should return true if response timestamp is greater than request timestamp", () => {
            const request: NitroliteRPCMessage = { req: [123, "method", [], 100] };
            const response: NitroliteRPCMessage = { res: [123, "method", [], 200] };
            expect(isValidResponseTimestamp(request, response)).toBe(true);
        });

        test("should return false if response timestamp is less than or equal to request timestamp", () => {
            const request: NitroliteRPCMessage = { req: [123, "method", [], 200] };
            const response: NitroliteRPCMessage = { res: [123, "method", [], 200] };
            expect(isValidResponseTimestamp(request, response)).toBe(false);

            const response2: NitroliteRPCMessage = { res: [123, "method", [], 100] };
            expect(isValidResponseTimestamp(request, response2)).toBe(false);
        });

        test("should return false if timestamps are missing", () => {
            const request: NitroliteRPCMessage = { req: [123, "method", []] };
            const response: NitroliteRPCMessage = { res: [123, "method", [], 200] };
            expect(isValidResponseTimestamp(request, response)).toBe(false);

            const request2: NitroliteRPCMessage = { req: [123, "method", [], 100] };
            const response2: NitroliteRPCMessage = { res: [123, "method", []] };
            expect(isValidResponseTimestamp(request2, response2)).toBe(false);
        });
    });

    describe("isValidResponseRequestId", () => {
        test("should return true if response request ID matches request ID", () => {
            const request: NitroliteRPCMessage = { req: [123, "method", [], 100] };
            const response: NitroliteRPCMessage = { res: [123, "method", [], 200] };
            expect(isValidResponseRequestId(request, response)).toBe(true);
        });

        test("should return false if response request ID does not match request ID", () => {
            const request: NitroliteRPCMessage = { req: [123, "method", [], 100] };
            const response: NitroliteRPCMessage = { res: [456, "method", [], 200] };
            expect(isValidResponseRequestId(request, response)).toBe(false);

            const request2: NitroliteRPCMessage = { req: [undefined, "method", [], 100] as any };
            const response2: NitroliteRPCMessage = { res: [123, "method", [], 200] };
            expect(isValidResponseRequestId(request2, response)).toBe(false);

            const request3: NitroliteRPCMessage = { req: [123, "method", [], 100] };
            const response3: NitroliteRPCMessage = { res: [undefined, "method", [], 200] as any };
            expect(isValidResponseRequestId(request3 as any, response3)).toBe(false);
        });
    });
});
