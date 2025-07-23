import { z } from 'zod';
import { RPCMethod, GetAssetsResponseParams, RPCAsset, AssetsResponseParams } from '../types';
import { addressSchema, ParamsParser } from './common';

const AssetObjectSchema = z
    .object({ token: addressSchema, chain_id: z.number(), symbol: z.string(), decimals: z.number() })
    .transform(
        (raw): RPCAsset => ({
            token: raw.token,
            chainId: raw.chain_id,
            symbol: raw.symbol,
            decimals: raw.decimals,
        }),
    );

const GetAssetsParamsSchema = z
    .object({
        assets: z.array(AssetObjectSchema),
    })
    .transform(
        (raw): GetAssetsResponseParams => ({
            assets: raw.assets,
        }),
    );

const AssetsParamsSchema = z
    .object({
        assets: z.array(AssetObjectSchema),
    })
    .transform(
        (raw): AssetsResponseParams => ({
            assets: raw.assets,
        }),
    );

export const assetParamsParsers: Record<string, ParamsParser<unknown>> = {
    [RPCMethod.GetAssets]: (params) => GetAssetsParamsSchema.parse(params),
    [RPCMethod.Assets]: (params) => AssetsParamsSchema.parse(params),
};
