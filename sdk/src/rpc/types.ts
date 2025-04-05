import { Address, Hex } from "viem";

/**
 * Wire format for NitroRPC messages (compact format for transmission)
 */
export interface NitroliteRPCMessage {
    /** For requests: [requestId, method, params, timestamp] */
    req?: [number, string, any[], number];

    /** For responses: [requestId, method, result, timestamp] */
    res?: [number, string, any[], number];

    /** For errors: [requestId, errorCode, errorMessage, timestamp] */
    err?: [number, number, string, number];

    /** Message signature */
    sig?: Hex;
}

/**
 * Smart contract compatible format matching the contract's struct
 */
export interface RPCMessage {
    /** Unique identifier for the request */
    requestID: bigint;
    
    /** Method name to be invoked */
    method: string;
    
    /** Method parameters (serialized as hex) */
    params: Hex[];
    
    /** Method result (serialized as hex) */
    result: Hex[];
    
    /** Timestamp in milliseconds */
    timestamp: bigint;
}

/**
 * Error details for RPC responses
 */
export interface RPCError {
    /** Error code */
    code: number;
    
    /** Error message */
    message: string;
}

/**
 * Standard NitroliteRPC error codes
 */
export enum NitroliteErrorCode {
    // Standard JSON-RPC error codes
    PARSE_ERROR = -32700,
    INVALID_REQUEST = -32600,
    METHOD_NOT_FOUND = -32601,
    INVALID_PARAMS = -32602,
    INTERNAL_ERROR = -32603,

    // Nitro-specific error codes
    INVALID_STATE = -32001,
    CHANNEL_NOT_FOUND = -32002,
    INVALID_SIGNATURE = -32003,
    INVALID_TIMESTAMP = -32004,
    INVALID_REQUEST_ID = -32005,
    INSUFFICIENT_SIGNATURES = -32006
}

/** 
 * Function type for signing messages 
 * @param payload - The data to sign
 * @returns A Promise that resolves to the signature as a Hex string
 */
export type MessageSigner = (payload: string) => Promise<Hex>;

/**
 * Function type for verifying message signatures
 * @param payload - The data that was signed
 * @param signature - The signature to verify
 * @param address - The address of the expected signer
 * @returns A Promise that resolves to true if the signature is valid, false otherwise
 */
export type MessageVerifier = (payload: string, signature: Hex, address: Address) => Promise<boolean>;
