/**
 * Error handling system for Nitrolite SDK
 *
 * This module provides a standardized error system with categorized
 * errors, unique error codes, and helpful troubleshooting information.
 */

import { Address } from "viem";

/**
 * Base class for all Nitrolite SDK errors
 */
export class NitroliteError extends Error {
    /** Unique error code */
    public readonly code: string;
    /** HTTP-like status code */
    public readonly statusCode: number;
    /** Troubleshooting suggestion */
    public readonly suggestion: string;
    /** Additional details about the error */
    public readonly details?: Record<string, any>;

    /**
     * Create a new NitroliteError
     * @param message Error message
     * @param code Error code
     * @param statusCode HTTP-like status code
     * @param suggestion Troubleshooting suggestion
     * @param details Additional details about the error
     */
    constructor(message: string, code: string, statusCode: number, suggestion: string, details?: Record<string, any>) {
        super(message);
        this.name = this.constructor.name;
        this.code = code;
        this.statusCode = statusCode;
        this.suggestion = suggestion;
        this.details = details;

        // Ensure instanceof works correctly
        Object.setPrototypeOf(this, new.target.prototype);
    }

    /**
     * Convert error to a plain object for serialization
     */
    toJSON(): Record<string, any> {
        return {
            name: this.name,
            message: this.message,
            code: this.code,
            statusCode: this.statusCode,
            suggestion: this.suggestion,
            details: this.details,
        };
    }

    /**
     * Get error name and code
     */
    toString(): string {
        return `${this.name} [${this.code}]: ${this.message}`;
    }
}

// ----------------------------------------------------------------------------
// Base error categories
// ----------------------------------------------------------------------------

/**
 * Base class for validation errors
 */
export class ValidationError extends NitroliteError {
    constructor(
        message: string,
        code: string = "VALIDATION_ERROR",
        statusCode: number = 400,
        suggestion: string = "Check input parameters against the API documentation",
        details?: Record<string, any>
    ) {
        super(message, code, statusCode, suggestion, details);
    }
}

/**
 * Base class for network/connectivity errors
 */
export class NetworkError extends NitroliteError {
    constructor(
        message: string,
        code: string = "NETWORK_ERROR",
        statusCode: number = 500,
        suggestion: string = "Check your network connection and try again",
        details?: Record<string, any>
    ) {
        super(message, code, statusCode, suggestion, details);
    }
}

/**
 * Base class for timeout errors
 */
export class TimeoutError extends NitroliteError {
    constructor(
        message: string,
        code: string = "TIMEOUT_ERROR",
        statusCode: number = 408,
        suggestion: string = "The operation timed out. Consider increasing the timeout value or try again later",
        details?: Record<string, any>
    ) {
        super(message, code, statusCode, suggestion, details);
    }
}

/**
 * Base class for authentication errors
 */
export class AuthenticationError extends NitroliteError {
    constructor(
        message: string,
        code: string = "AUTHENTICATION_ERROR",
        statusCode: number = 401,
        suggestion: string = "Check your credentials and permissions",
        details?: Record<string, any>
    ) {
        super(message, code, statusCode, suggestion, details);
    }
}

/**
 * Base class for state transition errors
 */
export class StateError extends NitroliteError {
    constructor(
        message: string,
        code: string = "STATE_ERROR",
        statusCode: number = 400,
        suggestion: string = "Verify the current state and your attempted transition",
        details?: Record<string, any>
    ) {
        super(message, code, statusCode, suggestion, details);
    }
}

/**
 * Base class for contract interaction errors
 */
export class ContractError extends NitroliteError {
    constructor(
        message: string,
        code: string = "CONTRACT_ERROR",
        statusCode: number = 500,
        suggestion: string = "Verify contract addresses and interactions",
        details?: Record<string, any>
    ) {
        super(message, code, statusCode, suggestion, details);
    }
}

// ----------------------------------------------------------------------------
// Specific error types
// ----------------------------------------------------------------------------

// --- Validation errors ---

/**
 * Error thrown when input parameters are invalid
 */
export class InvalidParameterError extends ValidationError {
    constructor(message: string, details?: Record<string, any>) {
        super(message, "INVALID_PARAMETER", 400, "Check the parameter value against the API documentation", details);
    }
}

/**
 * Error thrown when a required parameter is missing
 */
export class MissingParameterError extends ValidationError {
    constructor(parameter: string, details?: Record<string, any>) {
        super(
            `Required parameter '${parameter}' is missing`,
            "MISSING_PARAMETER",
            400,
            `Make sure to provide the required '${parameter}' parameter`,
            details
        );
    }
}

// --- Authentication errors ---

/**
 * Error thrown when a signature is invalid
 */
export class InvalidSignatureError extends AuthenticationError {
    constructor(message: string = "Invalid signature", details?: Record<string, any>) {
        super(message, "INVALID_SIGNATURE", 401, "Ensure you are using the correct signing key and data", details);
    }
}

/**
 * Error thrown when an operation is unauthorized
 */
export class UnauthorizedError extends AuthenticationError {
    constructor(message: string = "Unauthorized operation", details?: Record<string, any>) {
        super(message, "UNAUTHORIZED", 403, "You do not have permission to perform this operation", details);
    }
}

// --- Network errors ---

/**
 * Error thrown when a request times out
 */
export class RequestTimeoutError extends TimeoutError {
    constructor(message: string = "Request timed out", retries?: number, details?: Record<string, any>) {
        const retryInfo = retries !== undefined ? ` after ${retries} retries` : "";
        super(`${message}${retryInfo}`, "REQUEST_TIMEOUT", 408, "Consider increasing the timeout or retries in the configuration", details);
    }
}

/**
 * Error thrown when a connection fails
 */
export class ConnectionError extends NetworkError {
    constructor(message: string = "Connection failed", details?: Record<string, any>) {
        super(message, "CONNECTION_FAILED", 503, "Check that the remote endpoint is available and reachable", details);
    }
}

/**
 * Error thrown when a provider is not connected
 */
export class ProviderNotConnectedError extends NetworkError {
    constructor(providerType: string = "Provider", details?: Record<string, any>) {
        super(
            `${providerType} is not connected`,
            "PROVIDER_NOT_CONNECTED",
            503,
            `Call connect() on the ${providerType.toLowerCase()} before using it`,
            details
        );
    }
}

// --- State errors ---

/**
 * Error thrown when a state transition is invalid
 */
export class InvalidStateTransitionError extends StateError {
    constructor(message: string = "Invalid state transition", details?: Record<string, any>) {
        super(message, "INVALID_STATE_TRANSITION", 400, "Ensure the state transition follows the application rules", details);
    }
}

/**
 * Error thrown when a state is not found
 */
export class StateNotFoundError extends StateError {
    constructor(entity: string = "State", id?: string, details?: Record<string, any>) {
        const idStr = id ? ` with ID ${id}` : "";
        super(`${entity}${idStr} not found`, "STATE_NOT_FOUND", 404, `Verify that the ${entity.toLowerCase()} exists and is accessible`, details);
    }
}

/**
 * Error thrown when a state is not initialized
 */
export class StateNotInitializedError extends StateError {
    constructor(stateType: string = "State", details?: Record<string, any>) {
        super(`${stateType} is not initialized`, "STATE_NOT_INITIALIZED", 400, `Initialize the ${stateType.toLowerCase()} before using it`, details);
    }
}

// --- Channel errors ---

/**
 * Error thrown when a channel is not found
 */
export class ChannelNotFoundError extends StateNotFoundError {
    constructor(channelId?: string, details?: Record<string, any>) {
        super("Channel", channelId, details);
        Object.defineProperty(this, 'code', { value: "CHANNEL_NOT_FOUND" });
        Object.defineProperty(this, 'suggestion', { value: "Verify that the channel exists and is registered" });
    }
}

/**
 * Error thrown when a participant is not in a channel
 */
export class NotParticipantError extends UnauthorizedError {
    constructor(address?: string, channelId?: string, details?: Record<string, any>) {
        const addressStr = address ? ` ${address}` : "";
        const channelStr = channelId ? ` in channel ${channelId}` : "";
        super(`Address${addressStr} is not a participant${channelStr}`, {
            ...details,
            address,
            channelId,
        });

        Object.defineProperty(this, 'code', { value: "NOT_PARTICIPANT" });
        Object.defineProperty(this, 'suggestion', { value: "Only participants can perform operations on a channel" });
    }
}

// --- RPC errors ---

/**
 * Error thrown when an RPC method is not found
 */
export class MethodNotFoundError extends ValidationError {
    constructor(method?: string, details?: Record<string, any>) {
        const methodStr = method ? ` '${method}'` : "";
        super(`Method${methodStr} not found`, "METHOD_NOT_FOUND", 404, "Verify the method name and that it is registered on the server", details);
    }
}

/**
 * Error thrown when an RPC error occurs
 */
export class RPCError extends NitroliteError {
    constructor(message: string, rpcCode: number, details?: Record<string, any>) {
        super(message, `RPC_ERROR_${rpcCode}`, 500, "Check the RPC error code for more information", details);
    }
}

/**
 * Error thrown when RPC request parameters are invalid
 */
export class InvalidRPCParamsError extends ValidationError {
    constructor(message: string = "Invalid RPC parameters", details?: Record<string, any>) {
        super(message, "INVALID_RPC_PARAMS", 400, "Check the parameters against the RPC method documentation", details);
    }
}

// --- Virtual channel errors ---

/**
 * Error thrown when a virtual channel operation fails
 */
export class VirtualChannelError extends NitroliteError {
    constructor(
        message: string,
        code: string = "VIRTUAL_CHANNEL_ERROR",
        statusCode: number = 400,
        suggestion: string = "Check the virtual channel configuration",
        details?: Record<string, any>
    ) {
        super(message, code, statusCode, suggestion, details);
    }
}

/**
 * Error thrown when no next hop is found in a virtual channel
 */
export class NoNextHopError extends VirtualChannelError {
    constructor(address?: string, details?: Record<string, any>) {
        const addressStr = address ? ` for ${address}` : "";
        super(`No next hop found${addressStr}`, "NO_NEXT_HOP", 404, "Verify that the LVCI path is configured correctly", details);
    }
}

/**
 * Error thrown when a relay operation fails
 */
export class RelayError extends VirtualChannelError {
    constructor(message: string = "Failed to relay message", details?: Record<string, any>) {
        super(message, "RELAY_FAILED", 500, "Check that all intermediaries in the path are available", details);
    }
}

// --- Contract errors ---

/**
 * Error thrown when a contract is not found
 */
export class ContractNotFoundError extends ContractError {
    constructor(contractType: string = "Contract", details?: Record<string, any>) {
        super(
            `${contractType} not found`,
            "CONTRACT_NOT_FOUND",
            404,
            `Verify the ${contractType.toLowerCase()} address in the configuration`,
            details
        );
    }
}

/**
 * Error thrown when a contract operation fails
 */
export class ContractCallError extends ContractError {
    constructor(message: string = "Contract call failed", details?: Record<string, any>) {
        super(message, "CONTRACT_CALL_FAILED", 500, "Check the contract call parameters and transaction settings", details);
    }
}

/**
 * Error thrown when a transaction fails
 */
export class TransactionError extends ContractError {
    constructor(message: string = "Transaction failed", details?: Record<string, any>) {
        super(message, "TRANSACTION_FAILED", 500, "Verify transaction parameters and ensure sufficient funds", details);
    }
}

/**
 * Error thrown when a token operation fails
 */
export class TokenError extends ContractError {
    constructor(
        message: string = "Token operation failed",
        code: string = "TOKEN_ERROR",
        statusCode: number = 400,
        suggestion: string = "Check token balance and allowance",
        details?: Record<string, any>
    ) {
        super(message, code, statusCode, suggestion, details);
    }
}

/**
 * Error thrown when token balance is insufficient
 */
export class InsufficientBalanceError extends TokenError {
    constructor(tokenAddress?: Address, required?: bigint, actual?: bigint, details?: Record<string, any>) {
        super("Insufficient token balance", "INSUFFICIENT_BALANCE", 400, "Ensure you have enough tokens to complete this operation", {
            ...details,
            tokenAddress,
            required,
            actual,
        });
    }
}

/**
 * Error thrown when token allowance is insufficient
 */
export class InsufficientAllowanceError extends TokenError {
    constructor(tokenAddress?: Address, spender?: Address, required?: bigint, actual?: bigint, details?: Record<string, any>) {
        super("Insufficient token allowance", "INSUFFICIENT_ALLOWANCE", 400, "Approve the token for the required amount before continuing", {
            ...details,
            tokenAddress,
            spender,
            required,
            actual,
        });
    }
}

// Create namespace object containing all error types
export const Errors = {
    NitroliteError,
    ValidationError,
    NetworkError,
    TimeoutError,
    AuthenticationError,
    StateError,
    ContractError,

    InvalidParameterError,
    MissingParameterError,
    InvalidSignatureError,
    UnauthorizedError,
    RequestTimeoutError,
    ConnectionError,
    ProviderNotConnectedError,
    InvalidStateTransitionError,
    StateNotFoundError,
    StateNotInitializedError,
    ChannelNotFoundError,
    NotParticipantError,
    MethodNotFoundError,
    RPCError,
    InvalidRPCParamsError,
    VirtualChannelError,
    NoNextHopError,
    RelayError,
    ContractNotFoundError,
    ContractCallError,
    TransactionError,
    TokenError,
    InsufficientBalanceError,
    InsufficientAllowanceError,
};

export default Errors;
