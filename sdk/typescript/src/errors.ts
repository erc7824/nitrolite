import { Address } from 'viem';

/**
 * Base class for all Nitrolite SDK errors
 */
export class NitroliteError extends Error {
    public readonly code: string;
    public readonly statusCode: number;
    public readonly suggestion: string;
    public readonly details?: Record<string, any>;
    public readonly cause?: Error;

    constructor(
        message: string,
        code: string,
        statusCode: number,
        suggestion: string,
        details?: Record<string, any>,
        cause?: Error,
    ) {
        super(message);
        this.name = this.constructor.name;
        this.code = code;
        this.statusCode = statusCode;
        this.suggestion = suggestion;
        this.details = details;
        this.cause = cause;

        Object.setPrototypeOf(this, new.target.prototype);
    }

    toJSON(): Record<string, any> {
        return {
            name: this.name,
            message: this.message,
            code: this.code,
            statusCode: this.statusCode,
            suggestion: this.suggestion,
            details: this.details,
            cause: this.cause
                ? { name: this.cause.name, message: this.cause.message, stack: this.cause.stack }
                : undefined,
        };
    }

    toString(): string {
        return `${this.name} [${this.code}]: ${this.message}`;
    }
}

// --- Base Error Categories ---

/** Base class for validation errors */
export class ValidationError extends NitroliteError {
    constructor(
        message: string,
        code: string = 'VALIDATION_ERROR',
        statusCode: number = 400,
        suggestion: string = 'Check input parameters',
        details?: Record<string, any>,
        cause?: Error,
    ) {
        super(message, code, statusCode, suggestion, details, cause);
    }
}

/** Base class for authentication/authorization errors */
export class AuthenticationError extends NitroliteError {
    constructor(
        message: string,
        code: string = 'AUTHENTICATION_ERROR',
        statusCode: number = 401,
        suggestion: string = 'Check credentials, permissions, or signatures',
        details?: Record<string, any>,
        cause?: Error,
    ) {
        super(message, code, statusCode, suggestion, details, cause);
    }
}

/** Base class for contract interaction errors */
export class ContractError extends NitroliteError {
    constructor(
        message: string,
        code: string = 'CONTRACT_ERROR',
        statusCode: number = 500,
        suggestion: string = 'Verify contract addresses, interactions, and network status',
        details?: Record<string, any>,
        cause?: Error,
    ) {
        super(message, code, statusCode, suggestion, details, cause);
    }
}

/** Base class for state-related errors */ // Added definition
export class StateError extends NitroliteError {
    constructor(
        message: string,
        code: string = 'STATE_ERROR',
        statusCode: number = 400,
        suggestion: string = 'Check application state or channel status',
        details?: Record<string, any>,
        cause?: Error,
    ) {
        super(message, code, statusCode, suggestion, details, cause);
    }
}

// --- Specific Error Types ---

// Validation Errors
export class InvalidParameterError extends ValidationError {
    constructor(message: string, details?: Record<string, any>, cause?: Error) {
        super(
            message,
            'INVALID_PARAMETER',
            400,
            'Check the parameter value against the expected type or format',
            details,
            cause,
        );
    }
}

export class MissingParameterError extends ValidationError {
    constructor(parameter: string, details?: Record<string, any>, cause?: Error) {
        super(
            `Required parameter '${parameter}' is missing`,
            'MISSING_PARAMETER',
            400,
            `Provide the required '${parameter}' parameter`,
            details,
            cause,
        );
    }
}

// Authentication Errors
export class InvalidSignatureError extends AuthenticationError {
    constructor(message: string = 'Invalid signature', details?: Record<string, any>, cause?: Error) {
        super(
            message,
            'INVALID_SIGNATURE',
            401,
            'Ensure the correct data was signed with the correct key',
            details,
            cause,
        );
    }
}

export class UnauthorizedError extends AuthenticationError {
    constructor(message: string = 'Unauthorized operation', details?: Record<string, any>, cause?: Error) {
        super(message, 'UNAUTHORIZED', 403, 'You do not have permission to perform this operation', details, cause);
    }
}

export class NotParticipantError extends UnauthorizedError {
    constructor(address?: string, channelId?: string, details?: Record<string, any>, cause?: Error) {
        const addressStr = address ? ` ${address}` : '';
        const channelStr = channelId ? ` in channel ${channelId}` : '';
        const message = `Address${addressStr} is not a participant${channelStr}`;
        const combinedDetails = {
            // Combine details before passing
            ...details,
            address,
            channelId,
        };
        super(message, combinedDetails, cause);
        Object.defineProperty(this, 'code', { value: 'NOT_PARTICIPANT', writable: false });
        Object.defineProperty(this, 'suggestion', {
            value: 'Only channel participants can perform this operation',
            writable: false,
        });
    }
}

export class WalletClientRequiredError extends AuthenticationError {
    constructor(details?: Record<string, any>, cause?: Error) {
        super(
            'WalletClient instance is required for this operation',
            'WALLET_CLIENT_REQUIRED',
            400,
            'Provide a valid WalletClient instance during service initialization',
            details,
            cause,
        );
    }
}

export class AccountRequiredError extends AuthenticationError {
    constructor(details?: Record<string, any>, cause?: Error) {
        super(
            'Account is required for this operation',
            'ACCOUNT_REQUIRED',
            400,
            'Ensure an account is associated with the WalletClient or provided explicitly',
            details,
            cause,
        );
    }
}

// Contract Errors
export class ContractNotFoundError extends ContractError {
    constructor(contractType: string = 'Contract', address?: Address, details?: Record<string, any>, cause?: Error) {
        const addressStr = address ? ` at ${address}` : '';
        super(
            `${contractType}${addressStr} not found`,
            'CONTRACT_NOT_FOUND',
            404,
            `Verify the ${contractType.toLowerCase()} address in the configuration and ensure it's deployed on the correct network`,
            details,
            cause,
        );
    }
}

export class ContractReadError extends ContractError {
    constructor(functionName: string, cause: Error, details?: Record<string, any>) {
        super(
            `Failed to read from contract function '${functionName}'`,
            'CONTRACT_READ_FAILED',
            500,
            'Check contract address, network connection, and function arguments',
            details,
            cause,
        );
    }
}

export class ContractCallError extends ContractError {
    constructor(functionName: string, cause: Error, details?: Record<string, any>) {
        super(
            `Contract call simulation failed for function '${functionName}'`,
            'CONTRACT_CALL_FAILED',
            400,
            'Check contract call parameters, account balance/allowance, and contract state',
            details,
            cause,
        );
    }
}

export class TransactionError extends ContractError {
    constructor(operationName: string, cause: Error, details?: Record<string, any>) {
        super(
            `Transaction failed during operation '${operationName}'`,
            'TRANSACTION_FAILED',
            500,
            'Verify transaction parameters, gas limits, nonce, and ensure sufficient funds/allowance. Check network status.',
            details,
            cause,
        );
    }
}

export class TokenError extends ContractError {
    constructor(
        message: string = 'Token operation failed',
        code: string = 'TOKEN_ERROR',
        statusCode: number = 400,
        suggestion: string = 'Check token address, balance, and allowance',
        details?: Record<string, any>,
        cause?: Error,
    ) {
        super(message, code, statusCode, suggestion, details, cause);
    }
}

export class InsufficientBalanceError extends TokenError {
    constructor(
        tokenAddress?: Address,
        required?: bigint,
        actual?: bigint,
        details?: Record<string, any>,
        cause?: Error,
    ) {
        super(
            'Insufficient token balance',
            'INSUFFICIENT_BALANCE',
            400,
            'Ensure the account has enough tokens/ETH to complete this operation',
            {
                ...details,
                tokenAddress,
                required,
                actual,
            },
            cause,
        );
    }
}

export class InsufficientAllowanceError extends TokenError {
    constructor(
        tokenAddress?: Address,
        spender?: Address,
        required?: bigint,
        actual?: bigint,
        details?: Record<string, any>,
        cause?: Error,
    ) {
        super(
            'Insufficient token allowance',
            'INSUFFICIENT_ALLOWANCE',
            400,
            'Approve the spender for the required token amount before continuing',
            {
                ...details,
                tokenAddress,
                spender,
                required,
                actual,
            },
            cause,
        );
    }
}

// State Errors
export class InvalidStateTransitionError extends StateError {
    constructor(message: string = 'Invalid state transition', details?: Record<string, any>, cause?: Error) {
        super(
            message,
            'INVALID_STATE_TRANSITION',
            400,
            'Ensure the state transition follows the application rules',
            details,
            cause,
        );
    }
}

export class StateNotFoundError extends StateError {
    constructor(entity: string = 'State', id?: string, details?: Record<string, any>, cause?: Error) {
        const idStr = id ? ` with ID ${id}` : '';
        super(
            `${entity}${idStr} not found`,
            'STATE_NOT_FOUND',
            404,
            `Verify that the ${entity.toLowerCase()} exists and is accessible`,
            details,
            cause,
        );
    }
}

export class ChannelNotFoundError extends StateNotFoundError {
    constructor(channelId?: string, details?: Record<string, any>, cause?: Error) {
        super('Channel', channelId, details, cause);
        Object.defineProperty(this, 'code', { value: 'CHANNEL_NOT_FOUND' });
        Object.defineProperty(this, 'suggestion', {
            value: 'Verify the channel ID and ensure the channel exists on-chain',
        });
    }
}

export const Errors = {
    NitroliteError,
    ValidationError,
    AuthenticationError,
    ContractError,
    StateError,

    InvalidParameterError,
    MissingParameterError,
    InvalidSignatureError,
    UnauthorizedError,
    NotParticipantError,
    WalletClientRequiredError,
    AccountRequiredError,

    ContractNotFoundError,
    ContractReadError,
    ContractCallError,
    TransactionError,
    TokenError,
    InsufficientBalanceError,
    InsufficientAllowanceError,

    InvalidStateTransitionError,
    StateNotFoundError,
    ChannelNotFoundError,
};

export default Errors;
