import { z } from 'zod';
import { RPCChannelStatus, RPCMethod } from '../types';

// --- Shared Interfaces & Classes ---

export interface ParamsParser<T> {
    (params: object[]): T;
}

export class ParserParamsMissingError extends Error {
    constructor(method: RPCMethod) {
        super(`Missing params for ${method} parser`);
        this.name = 'ParserParamsMissingError';
    }
}

// --- Shared Zod Schemas ---

export const hexSchema = z.string().refine((val) => /^0x[0-9a-fA-F]*$/.test(val), {
    message: 'Must be a 0x-prefixed hex string',
});

export const addressSchema = z.string().refine((val) => /^0x[0-9a-fA-F]{40}$/.test(val), {
    message: 'Must be a 0x-prefixed hex string of 40 hex chars (EVM address)',
});

export const statusEnum = z.enum(Object.values(RPCChannelStatus) as [string, ...string[]]);

// --- Shared Parser Functions ---

export const noop: ParamsParser<unknown> = (_) => {
    return {};
};
