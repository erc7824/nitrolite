import { z } from 'zod';
import { RPCMethod, AuthChallengeResponseParams, AuthVerifyResponseParams, AuthRequestResponseParams } from '../types';
import { addressSchema, ParamsParser } from './common';

const AuthChallengeParamsSchema = z
    .object({ challenge_message: z.string() })
    .transform((raw): AuthChallengeResponseParams => ({ challengeMessage: raw.challenge_message }));

const AuthVerifyParamsSchema = z
    .object({
        address: addressSchema,
        session_key: addressSchema,
        success: z.boolean(),
        jwt_token: z.string().optional(),
    })
    .transform(
        (raw): AuthVerifyResponseParams => ({
            address: raw.address,
            sessionKey: raw.session_key,
            success: raw.success,
            jwtToken: raw.jwt_token,
        }),
    );

const AuthRequestParamsSchema = z
    .object({ challenge_message: z.string() })
    .transform((raw): AuthRequestResponseParams => ({ challengeMessage: raw.challenge_message }));

export const authParamsParsers: Record<string, ParamsParser<unknown>> = {
    [RPCMethod.AuthChallenge]: (params) => AuthChallengeParamsSchema.parse(params),
    [RPCMethod.AuthVerify]: (params) => AuthVerifyParamsSchema.parse(params),
    [RPCMethod.AuthRequest]: (params) => AuthRequestParamsSchema.parse(params),
};
