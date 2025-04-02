import { keccak256, encodeAbiParameters, toHex, fromHex, Address, Hex, recoverMessageAddress, stringToBytes } from "viem";
import { Channel, State, Signature, ChannelId, StateHash, AppLogic } from "../types";

/**
 * Compute the unique identifier for a channel
 * @param channel The channel configuration
 * @returns The channel identifier as Hex
 */
export function getChannelId(channel: Channel): ChannelId {
    const encoded = encodeAbiParameters(
        [
            { name: "participants", type: "address[2]" },
            { name: "adjudicator", type: "address" },
            { name: "challenge", type: "uint64" },
            { name: "nonce", type: "uint64" },
        ],
        [channel.participants, channel.adjudicator, channel.challenge, channel.nonce]
    );

    return keccak256(encoded);
}

/**
 * Compute the hash of a channel state in a canonical way (ignoring the signature)
 * @param channel The channel configuration
 * @param state The state struct
 * @returns The state hash as Hex
 */
export function getStateHash(channel: Channel, state: State): StateHash {
    const channelId = getChannelId(channel);

    const encoded = encodeAbiParameters(
        [
            { name: "channelId", type: "bytes32" },
            { name: "data", type: "bytes" },
            { name: "allocations", type: "tuple(address destination, address token, uint256 amount)[2]" },
        ],
        [channelId, state.data, state.allocations]
    );

    return keccak256(encoded);
}

/**
 * Create a signature for a state
 * @param stateHash The hash of the state to sign
 * @param privateKey The private key to sign with
 * @returns The signature
 */
export async function signState(stateHash: StateHash, signer: any): Promise<Signature> {
    // Implementation depends on the signer type
    // This is a placeholder that would be replaced with actual signing logic
    throw new Error("Not implemented - depends on signing methodology");
}

/**
 * Verifies that a state is signed by the specified participant
 * @param stateHash The hash of the state to verify
 * @param signature The signature to verify
 * @param signer The address of the expected signer
 * @returns True if the signature is valid, false otherwise
 */
export async function verifySignature(stateHash: StateHash, signature: Signature, signer: Address): Promise<boolean> {
    try {
        // Convert signature parts to proper format
        // Use a simplified approach to avoid fromHex spreading issues
        const r = fromHex(signature.r, { to: "bytes", size: 32 });
        const s = fromHex(signature.s, { to: "bytes", size: 32 });

        // Concatenate the bytes
        const bytes = new Uint8Array([...r, ...s, signature.v]);
        const sigString = toHex(bytes);

        // Recover address from signature
        const recoveredAddress = await recoverMessageAddress({
            message: { raw: stateHash },
            signature: sigString,
        });

        return recoveredAddress.toLowerCase() === signer.toLowerCase();
    } catch (error) {
        console.error("Signature verification failed:", error);
        return false;
    }
}

/**
 * Generate a robust nonce for channel creation
 * This mitigates collision risks by combining:
 * - Current timestamp (milliseconds)
 * - A random component
 * - An address-derived component (optional)
 *
 * @param address Optional address to mix into the nonce
 * @returns A unique BigInt nonce for channel creation
 */
export function generateChannelNonce(address?: Address): bigint {
    // Get current timestamp in milliseconds
    const timestamp = BigInt(Date.now());

    // Add random component (32 bits of randomness)
    const randomComponent = BigInt(Math.floor(Math.random() * 0xffffffff));

    // Combine with timestamp
    let nonce = (timestamp << 32n) | randomComponent;

    // If address is provided, incorporate its last 8 bytes into the nonce
    if (address) {
        // Take the last 8 bytes of the address (excluding 0x prefix)
        const addressComponent = BigInt(`0x${address.slice(-16)}`);
        // XOR the address component with the nonce
        nonce = nonce ^ addressComponent;
    }

    return nonce;
}

/**
 * Prepares a message for signing by:
 * 2. Hashing with keccak256
 * 3. Returning the hash bytes (without 0x prefix)
 *
 * @param message The message to prepare for signing
 * @returns Uint8Array of the hash bytes (without 0x prefix)
 */
export function prepareMessageForSigning(message: string): Hex {
    return keccak256(Buffer.from(message));
}

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
 * Simple status constants for application states
 */
export enum AppStatus {
    PENDING = 0,
    ACTIVE = 1,
    COMPLETE = 2,
}

/**
 * Helper function to create a custom application logic
 * @param config Configuration for the app logic
 * @returns AppLogic implementation
 */
export function createAppLogic<T>(config: {
    adjudicatorAddress: Address;
    adjudicatorType?: string;
    encode: (data: T) => Hex;
    decode: (encoded: Hex) => T;
    validateTransition?: (prevState: T, nextState: T, signer: Address) => boolean;
    isFinal?: (state: T) => boolean;
}): AppLogic<T> {
    const appLogic: AppLogic<T> = {
        encode: config.encode,
        decode: config.decode,
        validateTransition: config.validateTransition
            ? (channel, prevState, nextState) =>
                  config.validateTransition!(prevState, nextState, "0x0000000000000000000000000000000000000000" as Address)
            : undefined,
        isFinal: config.isFinal,
        getAdjudicatorAddress: () => config.adjudicatorAddress,
    };

    // Add adjudicator type if provided
    if (config.adjudicatorType) {
        appLogic.getAdjudicatorType = () => config.adjudicatorType as string;
    }

    return appLogic;
}

/**
 * State validators for common application patterns
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
