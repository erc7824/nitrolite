import { z } from 'zod';
import { RPCMethod, GetAssetsResponseParams, AssetsResponseParams, RPCAsset, RPCToken } from '../types';
import { addressSchema, ParamsParser } from './common';

// Token schema (token on a specific blockchain)
const TokenSchema = z
    .object({
        name: z.string(),
        symbol: z.string(),
        address: addressSchema,
        blockchain_id: z.number(),
        decimals: z.number(),
    })
    .transform(
        (raw): RPCToken => ({
            name: raw.name,
            symbol: raw.symbol,
            address: raw.address,
            blockchainId: raw.blockchain_id,
            decimals: raw.decimals,
        }),
    );

// Asset schema (asset with tokens across multiple blockchains)
const AssetSchema = z
    .object({
        name: z.string(),
        symbol: z.string(),
        tokens: z.array(TokenSchema),
    })
    .transform(
        (raw): RPCAsset => ({
            name: raw.name,
            symbol: raw.symbol,
            tokens: raw.tokens,
        }),
    );

// get_assets response parser
const GetAssetsParamsSchema = z
    .object({
        assets: z.array(AssetSchema),
    })
    .transform(
        (raw): GetAssetsResponseParams => ({
            assets: raw.assets,
        }),
    );

// assets event parser (server push)
const AssetsParamsSchema = z
    .object({
        assets: z.array(AssetSchema),
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
