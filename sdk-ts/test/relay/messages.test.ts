import {
    MessageType,
    ProposeStateMessage,
    AcceptStateMessage,
    RejectStateMessage,
    SignStateMessage,
    ChallengeNotificationMessage,
    ClosureNotificationMessage,
    NitroliteMessage,
} from "../../src/relay";
import { State } from "../../src/types";

describe("Message Types", () => {
    const testChannelId = "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef";
    const timestamp = Date.now();

    const testState: State = {
        data: "0x1234",
        allocations: [
            {
                destination: "0x1111111111111111111111111111111111111111",
                token: "0x2222222222222222222222222222222222222222",
                amount: BigInt(100),
            },
            {
                destination: "0x3333333333333333333333333333333333333333",
                token: "0x2222222222222222222222222222222222222222",
                amount: BigInt(200),
            },
        ],
        sigs: [],
    };

    const testStateHash = "0x5678567856785678567856785678567856785678567856785678567856785678";

    describe("ProposeStateMessage", () => {
        it("should create a valid message", () => {
            const message: ProposeStateMessage = {
                type: MessageType.PROPOSE_STATE,
                channelId: testChannelId,
                timestamp,
                state: testState,
                stateHash: testStateHash,
            };

            expect(message.type).toBe(MessageType.PROPOSE_STATE);
            expect(message.channelId).toBe(testChannelId);
            expect(message.timestamp).toBe(timestamp);
            expect(message.state).toBe(testState);
            expect(message.stateHash).toBe(testStateHash);
        });

        it("can be assigned to NitroliteMessage type", () => {
            const message: ProposeStateMessage = {
                type: MessageType.PROPOSE_STATE,
                channelId: testChannelId,
                timestamp,
                state: testState,
                stateHash: testStateHash,
            };

            const genericMessage: NitroliteMessage = message;
            expect(genericMessage.type).toBe(MessageType.PROPOSE_STATE);
        });
    });

    describe("AcceptStateMessage", () => {
        it("should create a valid message", () => {
            const message: AcceptStateMessage = {
                type: MessageType.ACCEPT_STATE,
                channelId: testChannelId,
                timestamp,
                stateHash: testStateHash,
                signature: {
                    v: 27,
                    r: "0x1234123412341234123412341234123412341234123412341234123412341234",
                    s: "0x5678567856785678567856785678567856785678567856785678567856785678",
                },
            };

            expect(message.type).toBe(MessageType.ACCEPT_STATE);
            expect(message.channelId).toBe(testChannelId);
            expect(message.timestamp).toBe(timestamp);
            expect(message.stateHash).toBe(testStateHash);
            expect(message.signature.v).toBe(27);
        });
    });

    describe("RejectStateMessage", () => {
        it("should create a valid message", () => {
            const message: RejectStateMessage = {
                type: MessageType.REJECT_STATE,
                channelId: testChannelId,
                timestamp,
                stateHash: testStateHash,
                reason: "Invalid state transition",
            };

            expect(message.type).toBe(MessageType.REJECT_STATE);
            expect(message.channelId).toBe(testChannelId);
            expect(message.timestamp).toBe(timestamp);
            expect(message.stateHash).toBe(testStateHash);
            expect(message.reason).toBe("Invalid state transition");
        });
    });

    describe("SignStateMessage", () => {
        it("should create a valid message", () => {
            const message: SignStateMessage = {
                type: MessageType.SIGN_STATE,
                channelId: testChannelId,
                timestamp,
                state: testState,
            };

            expect(message.type).toBe(MessageType.SIGN_STATE);
            expect(message.channelId).toBe(testChannelId);
            expect(message.timestamp).toBe(timestamp);
            expect(message.state).toBe(testState);
        });
    });

    describe("ChallengeNotificationMessage", () => {
        it("should create a valid message", () => {
            const expirationTime = Date.now() + 86400000; // 1 day from now

            const message: ChallengeNotificationMessage = {
                type: MessageType.CHALLENGE_NOTIFICATION,
                channelId: testChannelId,
                timestamp,
                expirationTime,
                challengeState: testState,
            };

            expect(message.type).toBe(MessageType.CHALLENGE_NOTIFICATION);
            expect(message.channelId).toBe(testChannelId);
            expect(message.timestamp).toBe(timestamp);
            expect(message.expirationTime).toBe(expirationTime);
            expect(message.challengeState).toBe(testState);
        });
    });

    describe("ClosureNotificationMessage", () => {
        it("should create a valid message", () => {
            const message: ClosureNotificationMessage = {
                type: MessageType.CLOSURE_NOTIFICATION,
                channelId: testChannelId,
                timestamp,
                finalState: testState,
            };

            expect(message.type).toBe(MessageType.CLOSURE_NOTIFICATION);
            expect(message.channelId).toBe(testChannelId);
            expect(message.timestamp).toBe(timestamp);
            expect(message.finalState).toBe(testState);
        });
    });
});
