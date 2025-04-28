import { describe, test, expect, jest } from "@jest/globals";
import { encodeAbiParameters } from "viem";
import { encoders, AppStatus, createAppLogic, StateValidators } from "../../src/utils/app";
import type { Address } from "viem";

// Mock viem.encodeAbiParameters to return a predictable value
jest.mock("viem", () => ({
    encodeAbiParameters: jest.fn(() => "0xencoded"),
}));

describe("encoders", () => {
    test("numeric calls encodeAbiParameters with uint256", () => {
        const result = encoders.numeric(5n);
        expect(encodeAbiParameters).toHaveBeenCalledWith([{ type: "uint256", name: "value" }], [5n]);
        expect(result).toBe("0xencoded");
    });

    test("sequential calls encodeAbiParameters with sequence and value", () => {
        const result = encoders.sequential(1n, 2n);
        expect(encodeAbiParameters).toHaveBeenCalledWith(
            [
                { type: "uint256", name: "sequence" },
                { type: "uint256", name: "value" },
            ],
            [1n, 2n]
        );
        expect(result).toBe("0xencoded");
    });

    test("turnBased calls encodeAbiParameters with data, turn, status, isComplete", () => {
        const data = { foo: "bar" };
        const result = encoders.turnBased(data, 0, 1, true);
        expect(encodeAbiParameters).toHaveBeenCalledWith(
            [
                { type: "bytes", name: "data" },
                { type: "uint8", name: "turn" },
                { type: "uint8", name: "status" },
                { type: "bool", name: "isComplete" },
            ],
            ["0x", 0, 1, true]
        );
        expect(result).toBe("0xencoded");
    });

    test("empty returns '0x'", () => {
        expect(encoders.empty()).toBe("0x");
    });
});

describe("AppStatus enum", () => {
    test("values are correct", () => {
        expect(AppStatus.PENDING).toBe(0);
        expect(AppStatus.ACTIVE).toBe(1);
        expect(AppStatus.COMPLETE).toBe(2);
    });
});

describe("createAppLogic", () => {
    const dummyData = { x: 1 };
    const encode = jest.fn((d) => "0xenc");
    const decode = jest.fn((h) => dummyData);
    const validateTransition = jest.fn((channel, prev, next) => true);
    const provideProofs = jest.fn((channel, state, prevStates) => ["proof"]);
    const isFinal = jest.fn((state) => false);
    const adjudicatorAddress = "0xADJ" as Address;
    const adjudicatorType = "custom";

    const logic = createAppLogic({
        adjudicatorAddress,
        adjudicatorType,
        // @ts-ignore
        encode,
        decode,
        validateTransition,
        provideProofs,
        isFinal,
    });

    test("encode and decode are passed through", () => {
        expect(logic.encode(dummyData)).toBe("0xenc");
        expect(encode).toHaveBeenCalledWith(dummyData);
        expect(logic.decode("0x")).toEqual(dummyData);
        expect(decode).toHaveBeenCalledWith("0x");
    });

    test("getAdjudicatorAddress and type", () => {
        expect(logic.getAdjudicatorAddress()).toBe(adjudicatorAddress);
        expect(logic.getAdjudicatorType && logic.getAdjudicatorType()).toBe(adjudicatorType);
    });

    test("custom validators are set", () => {
        expect(logic.validateTransition).toBe(validateTransition);
        expect(logic.provideProofs).toBe(provideProofs);
        expect(logic.isFinal).toBe(isFinal);
    });
});

describe("StateValidators.turnBased", () => {
    const roles: [Address, Address] = ["0xA" as Address, "0xB" as Address];
    const validator = StateValidators.turnBased<{ turn: number }>((s) => s.turn);
    const prev = { turn: 0 };
    const next = { turn: 1 };

    test("valid transition and signer", () => {
        expect(validator(prev, next, roles[0], roles)).toBe(true);
    });

    test("invalid turn increment", () => {
        expect(validator(prev, { turn: 0 }, roles[0], roles)).toBe(false);
    });

    test("invalid signer", () => {
        expect(validator(prev, next, "0xC" as Address, roles)).toBe(false);
    });
});

describe("StateValidators.sequential", () => {
    const initiator = "0xINIT" as Address;
    const validator = StateValidators.sequential<{ seq: bigint; val: bigint }>(
        (s) => s.seq,
        (s) => s.val
    );
    const prev = { seq: 1n, val: 2n };
    const nextValid = { seq: 2n, val: 3n };
    const nextLowSeq = { seq: 1n, val: 3n };
    const nextLowVal = { seq: 2n, val: 1n };

    test("valid when initiator and sequence/value increasing", () => {
        expect(validator(prev, nextValid, initiator, initiator)).toBe(true);
    });

    test("invalid signer", () => {
        expect(validator(prev, nextValid, "0xOTHER" as Address, initiator)).toBe(false);
    });

    test("invalid sequence", () => {
        expect(validator(prev, nextLowSeq, initiator, initiator)).toBe(false);
    });

    test("invalid value", () => {
        expect(validator(prev, nextLowVal, initiator, initiator)).toBe(false);
    });
});
