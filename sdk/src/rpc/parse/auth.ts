import { z } from 'zod';
import { Address } from 'viem';
import { RPCMethod, AuthChallengeResponseParams, AuthVerifyResponseParams, AuthRequestResponseParams } from '../types';
import { addressSchema, ParamsParser } from './common';

const AuthChallengeParamsSchema = z
    .array(
        z
            .object({ challenge_message: z.string() })
            .transform((raw) => ({ challengeMessage: raw.challenge_message }) as AuthChallengeResponseParams),
    )
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0]);

const AuthVerifyParamsSchema = z
    .array(
        z
            .object({
                address: addressSchema,
                session_key: addressSchema,
                success: z.boolean(),
                jwt_token: z.string().optional(),
            })
            .transform(
                (raw) =>
                    ({
                        address: raw.address as Address,
                        sessionKey: raw.session_key as Address,
                        success: raw.success,
                        jwtToken: raw.jwt_token,
                    }) as AuthVerifyResponseParams,
            ),
    )
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0]);

const AuthRequestParamsSchema = z
    .array(
        z
            .object({ challenge_message: z.string() })
            .transform((raw) => ({ challengeMessage: raw.challenge_message }) as AuthRequestResponseParams),
    )
    .refine((arr) => arr.length === 1)
    .transform((arr) => arr[0]);

export const authParamsParsers: Record<string, ParamsParser<unknown>> = {
    [RPCMethod.AuthChallenge]: (params) => AuthChallengeParamsSchema.parse(params),
    [RPCMethod.AuthVerify]: (params) => AuthVerifyParamsSchema.parse(params),
    [RPCMethod.AuthRequest]: (params) => AuthRequestParamsSchema.parse(params),
};
