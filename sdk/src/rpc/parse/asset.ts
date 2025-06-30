import { z } from 'zod';
import { Address } from 'viem';
import { RPCMethod, GetAssetsResponseParams } from '../types';
import { addressSchema, ParamsParser } from './common';

const GetAssetsParamsSchema = z
    .array(
        z.array(
            z
                .object({ token: addressSchema, chain_id: z.number(), symbol: z.string(), decimals: z.number() })
                .transform(
                    (a) =>
                        ({
                            token: a.token as Address,
                            chainId: a.chain_id,
                            symbol: a.symbol,
                            decimals: a.decimals,
                        }) as GetAssetsResponseParams,
                ),
        ),
    )
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0])
    .transform((arr) => arr as GetAssetsResponseParams[]);

export const assetParamsParsers: Record<string, ParamsParser<unknown>> = {
    [RPCMethod.GetAssets]: (params) => GetAssetsParamsSchema.parse(params),
    [RPCMethod.Assets]: (params) => GetAssetsParamsSchema.parse(params), // Alias
};
