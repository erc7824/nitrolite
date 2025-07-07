import { z } from 'zod';
import { Address } from 'viem';
import {
    RPCMethod,
    GetConfigResponseParams,
    ErrorResponseParams,
    TransferRPCResponseParams,
    GetRPCHistoryResponseParams,
    GetUserTagResponseParams,
} from '../types';
import { hexSchema, addressSchema, ParamsParser, ParserParamsMissingError } from './common';

const NetworkInfoSchema = z.object({
    name: z.string(),
    chain_id: z.number(),
    custody_address: addressSchema,
    adjudicator_address: addressSchema,
});

const GetConfigParamsSchema = z
    .array(
        z
            .object({ broker_address: addressSchema, networks: z.array(NetworkInfoSchema) })
            .strict()
            .transform(
                (raw) =>
                    ({
                        brokerAddress: raw.broker_address as Address,
                        networks: raw.networks.map((n) => ({
                            name: n.name,
                            chainId: n.chain_id,
                            custodyAddress: n.custody_address as Address,
                            adjudicatorAddress: n.adjudicator_address as Address,
                        })),
                    }) as GetConfigResponseParams,
            ),
    )
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0]);

const ErrorParamsSchema = z.array(z.string().transform((raw) => ({ error: raw }) as ErrorResponseParams));

const TransferParamsSchema = z
    .array(
        z
            .object({
                from: addressSchema,
                to: addressSchema,
                allocations: z.array(
                    z.object({
                        asset: z.string(),
                        amount: z.union([z.string(), z.number()]).transform((a) => a.toString()),
                    }),
                ),
                created_at: z.union([z.string(), z.date()]).transform((v) => new Date(v)),
            })
            .transform(
                (raw) =>
                    ({
                        from: raw.from as Address,
                        to: raw.to as Address,
                        allocations: raw.allocations,
                        createdAt: raw.created_at,
                    }) as TransferRPCResponseParams,
            ),
    )
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0]);

const GetRPCHistoryParamsSchema = z
    .array(
        z.array(
            z
                .object({
                    id: z.number(),
                    sender: addressSchema,
                    req_id: z.number(),
                    method: z.string(),
                    params: z.string(),
                    timestamp: z.number(),
                    req_sig: z.array(hexSchema),
                    res_sig: z.array(hexSchema),
                    response: z.string(),
                })
                .transform(
                    (h) =>
                        ({
                            id: h.id,
                            sender: h.sender as Address,
                            reqId: h.req_id,
                            method: h.method,
                            params: h.params,
                            timestamp: h.timestamp,
                            reqSig: h.req_sig as any,
                            resSig: h.res_sig as any,
                            response: h.response,
                        }) as GetRPCHistoryResponseParams,
                ),
        ),
    )
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0])
    .transform((arr) => arr as GetRPCHistoryResponseParams[]);

const GetUserTagParamsSchema = z
    .array(
        z
            .object({
                tag: z.string(),
            })
            .strict()
            .transform((raw) => ({ tag: raw.tag }) as GetUserTagResponseParams),
    )
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0]);

const parseMessageParams: ParamsParser<unknown> = (params) => {
    if (!Array.isArray(params) || params.length === 0) throw new ParserParamsMissingError(RPCMethod.Message);
    return params[0];
};

export const miscParamsParsers: Record<string, ParamsParser<unknown>> = {
    [RPCMethod.GetConfig]: (params) => GetConfigParamsSchema.parse(params),
    [RPCMethod.Error]: (params) => ErrorParamsSchema.parse(params),
    [RPCMethod.Transfer]: (params) => TransferParamsSchema.parse(params),
    [RPCMethod.GetRPCHistory]: (params) => GetRPCHistoryParamsSchema.parse(params),
    [RPCMethod.GetUserTag]: (params) => GetUserTagParamsSchema.parse(params),
    [RPCMethod.Message]: parseMessageParams,
};
