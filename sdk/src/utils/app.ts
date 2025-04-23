import { encodeAbiParameters, Address, Hex } from "viem";
import { AppLogic } from "../types"; // Updated import path
import { Channel } from "../client/types"; // Import Channel if needed by validators

/**
 * Common encoders for standard application data patterns.
 */
export const encoders = {
    /**
     * Encode a single numeric value
     * @param value The value to encode
     * @returns Hex-encoded value
     */
    numeric: (value: bigint): Hex => {
        return encodeAbiParameters([{ type: "uint256", name: "value" }], [value]);
    },

    /**
     * Encode a sequential state
     * @param sequence Sequence number
     * @param value Associated value
     * @returns Hex-encoded state
     */
    sequential: (sequence: bigint, value: bigint): Hex => {
        return encodeAbiParameters(
            [
                { type: "uint256", name: "sequence" },
                { type: "uint256", name: "value" },
            ],
            [sequence, value]
        );
    },

    /**
     * Encode a turn-based state
     * @param data Game-specific data
     * @param turn Current turn index
     * @param status Game status
     * @param isComplete Whether the game is complete
     * @returns Hex-encoded state
     */
    turnBased: (data: unknown, turn: number, status: number, isComplete: boolean): Hex => {
        // This is a simplified implementation - real implementation would need to encode game-specific data
        return encodeAbiParameters(
            [
                { type: "bytes", name: "data" },
                { type: "uint8", name: "turn" },
                { type: "uint8", name: "status" },
                { type: "bool", name: "isComplete" },
            ],
            ["0x", turn, status, isComplete]
        );
    },

    /**
     * Create an empty state
     * @returns Empty hex state
     */
    empty: (): Hex => {
        return "0x";
    },
};

/**
 * Simple status constants for application states.
 */
export enum AppStatus {
    PENDING = 0,
    ACTIVE = 1,
    COMPLETE = 2,
}

/**
 * Helper function to create a custom application logic object.
 * @param config Configuration for the app logic.
 * @returns AppLogic implementation.
 */
export function createAppLogic<T>(config: {
    adjudicatorAddress: Address;
    adjudicatorType?: string;
    encode: (data: T) => Hex;
    decode: (encoded: Hex) => T;
    // Note: validateTransition signature in AppLogic now takes Channel, T, T
    // We adapt the config function signature here for simplicity, but the AppLogic implementation needs the channel.
    validateTransition?: (channel: Channel, prevState: T, nextState: T) => boolean;
    provideProofs?: (channel: Channel, state: T, previousStates: any[]) => any[]; // Use correct State type
    isFinal?: (state: T) => boolean;
}): AppLogic<T> {
    const appLogic: AppLogic<T> = {
        encode: config.encode,
        decode: config.decode,
        validateTransition: config.validateTransition, // Pass directly if signature matches AppLogic
        provideProofs: config.provideProofs,
        isFinal: config.isFinal,
        getAdjudicatorAddress: () => config.adjudicatorAddress,
        getAdjudicatorType: config.adjudicatorType ? () => config.adjudicatorType as string : undefined,
    };
    return appLogic;
}

/**
 * State validators for common application patterns.
 * These validators now receive the full Channel object as per the AppLogic interface.
 */
export const StateValidators = {
    /**
     * Create a turn-based validator
     * @param getTurn Function to extract the turn from state
     * @returns A validator function
     */
    turnBased<T>(getTurn: (state: T) => number): (prevState: T, nextState: T, signer: Address, roles: [Address, Address]) => boolean {
        return (prevState: T, nextState: T, signer: Address, roles: [Address, Address]): boolean => {
            const prevTurn = getTurn(prevState);
            const nextTurn = getTurn(nextState);

            // Turn must increment
            if (nextTurn !== (prevTurn + 1) % 2) {
                return false;
            }

            // Signer must be the player whose turn it was
            const expectedSigner = roles[prevTurn];
            return signer.toLowerCase() === expectedSigner.toLowerCase();
        };
    },

    /**
     * Create a sequential state validator
     * @param getSequence Function to extract sequence number from state
     * @param getValue Function to extract value from state
     * @returns A validator function
     */
    sequential<T>(
        getSequence: (state: T) => bigint,
        getValue: (state: T) => bigint
    ): (prevState: T, nextState: T, signer: Address, initiator: Address) => boolean {
        return (prevState: T, nextState: T, signer: Address, initiator: Address): boolean => {
            // Only initiator can update state
            if (signer.toLowerCase() !== initiator.toLowerCase()) {
                return false;
            }

            // Sequence must increase
            if (getSequence(nextState) <= getSequence(prevState)) {
                return false;
            }

            // Value cannot decrease
            if (getValue(nextState) < getValue(prevState)) {
                return false;
            }

            return true;
        };
    },
};
