import { describe, test, expect } from '@jest/globals';
import Errors, {
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
} from '../../../src/errors';
import { Address } from 'viem';

describe('Error classes', () => {
    test('NitroliteError properties, toJSON and toString', () => {
        const cause = new Error('root cause');
        const details = { foo: 'bar' };
        const err = new NitroliteError('message', 'CODE', 123, 'suggestion', details, cause);
        expect(err.name).toBe('NitroliteError');
        expect(err.message).toBe('message');
        expect(err.code).toBe('CODE');
        expect(err.statusCode).toBe(123);
        expect(err.suggestion).toBe('suggestion');
        expect(err.details).toBe(details);
        expect(err.cause).toBe(cause);

        const json = err.toJSON();
        expect(json).toMatchObject({
            name: 'NitroliteError',
            message: 'message',
            code: 'CODE',
            statusCode: 123,
            suggestion: 'suggestion',
            details,
        });
        expect(json.cause).toEqual({
            name: cause.name,
            message: cause.message,
            stack: cause.stack,
        });

        expect(err.toString()).toBe('NitroliteError [CODE]: message');
    });

    test('ValidationError defaults', () => {
        const err = new ValidationError('oops');
        expect(err.code).toBe('VALIDATION_ERROR');
        expect(err.statusCode).toBe(400);
        expect(err.suggestion).toBe('Check input parameters');
    });

    test('AuthenticationError defaults', () => {
        const err = new AuthenticationError('auth failed');
        expect(err.code).toBe('AUTHENTICATION_ERROR');
        expect(err.statusCode).toBe(401);
        expect(err.suggestion).toBe('Check credentials, permissions, or signatures');
    });

    test('ContractError defaults', () => {
        const err = new ContractError('contract broken');
        expect(err.code).toBe('CONTRACT_ERROR');
        expect(err.statusCode).toBe(500);
        expect(err.suggestion).toBe('Verify contract addresses, interactions, and network status');
    });

    test('StateError defaults', () => {
        const err = new StateError('state bad');
        expect(err.code).toBe('STATE_ERROR');
        expect(err.statusCode).toBe(400);
        expect(err.suggestion).toBe('Check application state or channel status');
    });

    test('InvalidParameterError', () => {
        const err = new InvalidParameterError('param bad');
        expect(err.code).toBe('INVALID_PARAMETER');
        expect(err.suggestion).toContain('expected type or format');
    });

    test('MissingParameterError', () => {
        const err = new MissingParameterError('x');
        expect(err.message).toBe("Required parameter 'x' is missing");
        expect(err.code).toBe('MISSING_PARAMETER');
        expect(err.suggestion).toContain("Provide the required 'x' parameter");
    });

    test('InvalidSignatureError', () => {
        const err = new InvalidSignatureError();
        expect(err.code).toBe('INVALID_SIGNATURE');
        expect(err.statusCode).toBe(401);
    });

    test('UnauthorizedError', () => {
        const err = new UnauthorizedError();
        expect(err.code).toBe('UNAUTHORIZED');
        expect(err.statusCode).toBe(403);
    });

    test('NotParticipantError', () => {
        const addr = '0x123' as Address;
        const cid = '0xcid';
        const err = new NotParticipantError(addr, cid);
        expect(err.code).toBe('NOT_PARTICIPANT');
        expect(err.suggestion).toBe('Only channel participants can perform this operation');
        expect(err.message).toContain(addr);
        expect(err.message).toContain(cid);
        expect(err.details).toMatchObject({ address: addr, channelId: cid });
    });

    test('WalletClientRequiredError', () => {
        const err = new WalletClientRequiredError();
        expect(err.code).toBe('WALLET_CLIENT_REQUIRED');
        expect(err.statusCode).toBe(400);
    });

    test('AccountRequiredError', () => {
        const err = new AccountRequiredError();
        expect(err.code).toBe('ACCOUNT_REQUIRED');
        expect(err.statusCode).toBe(400);
    });

    test('ContractNotFoundError', () => {
        const addr = '0xabc' as Address;
        const err = new ContractNotFoundError('Foo', addr);
        expect(err.code).toBe('CONTRACT_NOT_FOUND');
        expect(err.message).toContain('Foo at');
        expect(err.message).toContain(addr);
    });

    test('ContractReadError wraps cause', () => {
        const cause = new Error('read fail');
        const err = new ContractReadError('fn', cause);
        expect(err.code).toBe('CONTRACT_READ_FAILED');
        expect(err.cause).toBe(cause);
    });

    test('ContractCallError wraps cause', () => {
        const cause = new Error('call fail');
        const err = new ContractCallError('fn', cause);
        expect(err.code).toBe('CONTRACT_CALL_FAILED');
        expect(err.cause).toBe(cause);
    });

    test('TransactionError wraps cause', () => {
        const cause = new Error('tx fail');
        const err = new TransactionError('op', cause);
        expect(err.code).toBe('TRANSACTION_FAILED');
        expect(err.cause).toBe(cause);
    });

    test('TokenError defaults', () => {
        const err = new TokenError();
        expect(err.code).toBe('TOKEN_ERROR');
        expect(err.suggestion).toContain('Check token address, balance, and allowance');
    });

    test('InsufficientBalanceError details', () => {
        const addr = '0xtoken' as Address;
        const err = new InsufficientBalanceError(addr, 10n, 5n);
        expect(err.code).toBe('INSUFFICIENT_BALANCE');
        expect(err.details).toMatchObject({ tokenAddress: addr, required: 10n, actual: 5n });
    });

    test('InsufficientAllowanceError details', () => {
        const addr = '0xtoken' as Address;
        const spender = '0xspender' as Address;
        const err = new InsufficientAllowanceError(addr, spender, 20n, 0n);
        expect(err.code).toBe('INSUFFICIENT_ALLOWANCE');
        expect(err.details).toMatchObject({ tokenAddress: addr, spender, required: 20n, actual: 0n });
    });

    test('InvalidStateTransitionError defaults', () => {
        const err = new InvalidStateTransitionError();
        expect(err.code).toBe('INVALID_STATE_TRANSITION');
        expect(err.suggestion).toContain('application rules');
    });

    test('StateNotFoundError message and code', () => {
        const err = new StateNotFoundError('Item', 'id123');
        expect(err.code).toBe('STATE_NOT_FOUND');
        expect(err.message).toContain('Item with ID id123 not found');
    });

    test('ChannelNotFoundError overrides code and suggestion', () => {
        const err = new ChannelNotFoundError('cid123');
        expect(err.code).toBe('CHANNEL_NOT_FOUND');
        expect(err.suggestion).toContain('channel exists on-chain');
    });
});

describe('Errors namespace', () => {
    test('Errors object contains all error constructors', () => {
        expect(Errors.InvalidParameterError).toBe(InvalidParameterError);
        expect(Errors.MissingParameterError).toBe(MissingParameterError);
        expect(Errors.TransactionError).toBe(TransactionError);
        expect(typeof Errors.NitroliteError).toBe('function');
    });
});
