export class CompatError extends Error {
    readonly code: string;
    constructor(message: string, code: string) {
        super(message);
        this.name = 'CompatError';
        this.code = code;
    }
}

export class AllowanceError extends CompatError {
    constructor(message: string) { super(message, 'ALLOWANCE_INSUFFICIENT'); this.name = 'AllowanceError'; }
}

export class UserRejectedError extends CompatError {
    constructor(message: string) { super(message, 'USER_REJECTED'); this.name = 'UserRejectedError'; }
}

export class InsufficientFundsError extends CompatError {
    constructor(message: string) { super(message, 'INSUFFICIENT_FUNDS'); this.name = 'InsufficientFundsError'; }
}

export class NotInitializedError extends CompatError {
    constructor(message: string) { super(message, 'NOT_INITIALIZED'); this.name = 'NotInitializedError'; }
}

export function getUserFacingMessage(error: unknown): string {
    if (error instanceof AllowanceError) return 'Transaction was rejected. Please approve the transaction in your wallet to continue.';
    if (error instanceof UserRejectedError) return 'You rejected the transaction. Please try again and approve in your wallet.';
    if (error instanceof InsufficientFundsError) return 'Insufficient funds for gas fees plus the requested amount.';
    if (error instanceof NotInitializedError) return 'Wallet connection issue. Please reconnect your wallet and try again.';
    if (error instanceof CompatError) return `Operation failed: ${error.message}`;
    return 'An unexpected error occurred. Please try again.';
}
