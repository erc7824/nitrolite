import { z } from 'zod';
import { RPCProtocolVersion, RPCChannelStatus, RPCMethod } from '../types';
import { Address, Hex } from 'viem';

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

export const hexSchema = z
    .string()
    .refine((val) => /^0x[0-9a-fA-F]*$/.test(val), {
        message: 'Must be a 0x-prefixed hex string',
    })
    .transform((v: string) => v as Hex);

export const addressSchema = z
    .string()
    .refine((val) => /^0x[0-9a-fA-F]{40}$/.test(val), {
        message: 'Must be a 0x-prefixed hex string of 40 hex chars (EVM address)',
    })
    .transform((v: string) => v as Address);

// TODO: add more validation for bigints if needed
export const bigIntSchema = z.string();

export const dateSchema = z.union([z.string(), z.date()]).transform((v) => new Date(v));

export const decimalSchema = z
    .union([z.string(), z.number()])
    .transform((v) => v.toString())
    .refine((val) => /^[+-]?((\d+(\.\d*)?)|(\.\d+))$/.test(val), {
        message: 'Must be a valid decimal string',
    });

export const statusEnum = z.nativeEnum(RPCChannelStatus);

export const protocolVersionEnum = z.nativeEnum(RPCProtocolVersion);

// --- Shared Parser Functions ---

export const noop: ParamsParser<unknown> = (_) => {
    return {};
};
