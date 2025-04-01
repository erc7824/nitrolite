import { Address, Hex } from "viem";

/**
 * NitroliteRPC Message Types
 */
export interface NitroliteRPCMessage {
    /**
     * For requests: [requestId, method, params, timestamp]
     */
    req?: [number, string, any[], number];

    /**
     * For responses: [requestId, method, result, timestamp]
     */
    res?: [number, string, any[], number];

    /**
     * For errors: [requestId, errorCode, errorMessage, timestamp]
     */
    err?: [number, number, string, number];

    /**
     * Message signature
     */
    sig?: Hex;
}

/**
 * Standard NitroliteRPC error codes
 */
export enum NitroliteErrorCode {
    // Standard JSON-RPC error codes
    PARSE_ERROR = -32700, // Invalid JSON
    INVALID_REQUEST = -32600, // Invalid Request object
    METHOD_NOT_FOUND = -32601, // Method doesn't exist
    INVALID_PARAMS = -32602, // Invalid method parameters
    INTERNAL_ERROR = -32603, // Internal JSON-RPC error

    // Nitro-specific error codes
    INVALID_STATE = -32001, // Invalid state transition
    CHANNEL_NOT_FOUND = -32002, // Channel not found
    INVALID_SIGNATURE = -32003, // Invalid signature
}

/**
 * Message signer function type
 */
export type MessageSigner = (payload: Hex) => Promise<Hex>;

/**
 * Message verifier function type
 */
export type MessageVerifier = (payload: Hex, signature: Hex, address: Address) => Promise<boolean>;
