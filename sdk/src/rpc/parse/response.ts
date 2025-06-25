import { z } from "zod";
import { Address } from "viem";
import { RPCMethod, RPCResponseParamsByMethod, RPCChannelStatus, GetConfigResponseParams, AuthChallengeResponseParams, AuthVerifyResponseParams, AuthRequestResponseParams, CreateAppSessionResponseParams, SubmitStateResponseParams, CloseAppSessionResponseParams, GetAppDefinitionResponseParams, ResizeChannelResponseParams, CloseChannelResponseParams, ChannelUpdateResponseParams, TransferRPCResponseParams, ErrorResponseParams, GetLedgerBalancesResponseParams, GetLedgerEntriesResponseParams, GetAppSessionsResponseParams, GetChannelsResponseParams, GetRPCHistoryResponseParams, GetAssetsResponseParams, BalanceUpdateResponseParams } from "../types";

function snakeToCamel(str: string): string {
  return str.replace(/_([a-z])/g, (_, c) => c.toUpperCase());
}

export function keysToCamel<T>(obj: any): T {
  if (Array.isArray(obj)) {
    return obj.map((v) => keysToCamel(v)) as any;
  } else if (obj && typeof obj === 'object') {
    return Object.fromEntries(
      Object.entries(obj).map(([k, v]) => [snakeToCamel(k), keysToCamel(v)])
    ) as T;
  }
  return obj;
}

export interface ParamsParser<T> {
  (params: object[]): T;
}

export class ParserParamsMissingError extends Error {
  constructor(method: RPCMethod) {
    super(`Missing params for ${method} parser`);
    this.name = 'ParserParamsMissingError';
  }
}

// Zod schemas for Hex and Address (0x-prefixed strings)
const hexSchema = z.string().refine((val) => /^0x[0-9a-fA-F]*$/.test(val), {
  message: "Must be a 0x-prefixed hex string",
});
const addressSchema = z.string().refine((val) => /^0x[0-9a-fA-F]{40}$/.test(val), {
  message: "Must be a 0x-prefixed hex string of 40 hex chars (EVM address)",
});

// Zod schemas for each response type
const NetworkInfoSchema = z.object({
  name: z.string(),
  chain_id: z.number(),
  custody_address: addressSchema,
  adjudicator_address: addressSchema,
});

const GetConfigParamsSchema = z
  .array(
    z.object({
      broker_address: addressSchema,
      networks: z.array(NetworkInfoSchema),
    })
      .strict()
      .transform((raw) => ({
        brokerAddress: raw.broker_address as Address,
        networks: raw.networks.map((n) => ({
          name: n.name,
          chainId: n.chain_id,
          custodyAddress: n.custody_address as Address,
          adjudicatorAddress: n.adjudicator_address as Address,
        })),
      }) as GetConfigResponseParams)
  )
  .refine((arr) => arr.length === 1, { message: 'Expected single object in array' })
  .transform((arr) => arr[0] as GetConfigResponseParams);

const GetLedgerBalancesParamsSchema = z
  .array(
    z.object({
      asset: z.string(),
      amount: z.union([z.string(), z.number()]).transform((a) => BigInt(a)),
    })
  )
  .transform(arr => arr as GetLedgerBalancesResponseParams[]);

const ErrorParamsSchema = z.array(z.object({ error: z.string() })).transform(arr => arr as ErrorResponseParams[]);

const AuthChallengeParamsSchema = z
  .array(
    z.object({ challenge_message: z.string() })
      .transform((raw) => ({ challengeMessage: raw.challenge_message }) as AuthChallengeResponseParams)
  )
  .refine((arr) => arr.length === 1, { message: 'Expected single object in array' })
  .transform((arr) => arr[0] as AuthChallengeResponseParams);

const AuthVerifyParamsSchema = z
  .array(
    z.object({
      address: addressSchema,
      jwt_token: z.string().optional(),
      session_key: addressSchema,
      success: z.boolean(),
    })
      .transform((raw) => ({
        address: raw.address as Address,
        jwtToken: raw.jwt_token,
        sessionKey: raw.session_key as Address,
        success: raw.success,
      }) as AuthVerifyResponseParams)
  )
  .refine((arr) => arr.length === 1, { message: 'Expected single object in array' })
  .transform((arr) => arr[0] as AuthVerifyResponseParams);

const AuthRequestParamsSchema = z
  .array(
    z.object({ challenge_message: z.string() })
      .transform((raw) => ({ challengeMessage: raw.challenge_message }) as AuthRequestResponseParams)
  )
  .refine((arr) => arr.length === 1, { message: 'Expected single object in array' })
  .transform((arr) => arr[0] as AuthRequestResponseParams);

const GetLedgerEntriesParamsSchema = z
  .array(
    z
      .object({
        id: z.number(),
        account_id: z.string(),
        account_type: z.string(),
        asset: z.string(),
        participant: addressSchema,
        credit: z.union([z.string(), z.number()]).transform((v) => BigInt(v)),
        debit: z.union([z.string(), z.number()]).transform((v) => BigInt(v)),
        created_at: z.union([z.string(), z.date()]).transform((v) => new Date(v)),
      })
      .transform((e) => ({
        id: e.id,
        accountId: e.account_id,
        accountType: e.account_type,
        asset: e.asset,
        participant: e.participant as Address,
        credit: e.credit,
        debit: e.debit,
        createdAt: e.created_at,
      }) as GetLedgerEntriesResponseParams)
  )
  .transform(arr => arr as GetLedgerEntriesResponseParams[]);

const statusEnum = z.enum(Object.values(RPCChannelStatus) as [string, ...string[]]);

const CreateAppSessionParamsSchema = z
  .array(
    z.object({
      app_session_id: hexSchema,
      version: z.number(),
      status: statusEnum,
    })
      .transform((raw) => ({
        appSessionId: raw.app_session_id as `0x${string}`,
        version: raw.version,
        status: raw.status as RPCChannelStatus,
      }) as CreateAppSessionResponseParams)
  )
  .refine((arr) => arr.length === 1, { message: 'Expected single object in array' })
  .transform((arr) => arr[0] as CreateAppSessionResponseParams);

const SubmitStateParamsSchema = z
  .array(
    z.object({
      app_session_id: hexSchema,
      version: z.number(),
      status: statusEnum,
    })
      .transform((raw) => ({
        appSessionId: raw.app_session_id as `0x${string}`,
        version: raw.version,
        status: raw.status as RPCChannelStatus,
      }) as SubmitStateResponseParams)
  )
  .refine((arr) => arr.length === 1, { message: 'Expected single object in array' })
  .transform((arr) => arr[0] as SubmitStateResponseParams);

const CloseAppSessionParamsSchema = z
  .array(
    z.object({
      app_session_id: hexSchema,
      version: z.number(),
      status: statusEnum,
    })
      .transform((raw) => ({
        appSessionId: raw.app_session_id as `0x${string}`,
        version: raw.version,
        status: raw.status as RPCChannelStatus,
      }) as CloseAppSessionResponseParams)
  )
  .refine((arr) => arr.length === 1, { message: 'Expected single object in array' })
  .transform((arr) => arr[0] as CloseAppSessionResponseParams);

const GetAppDefinitionParamsSchema = z
  .array(
    z.object({
      protocol: z.string(),
      participants: z.array(addressSchema),
      weights: z.array(z.number()),
      quorum: z.number(),
      challenge: z.number(),
      nonce: z.number(),
    })
      .transform((raw) => ({
        protocol: raw.protocol,
        participants: raw.participants as Address[],
        weights: raw.weights,
        quorum: raw.quorum,
        challenge: raw.challenge,
        nonce: raw.nonce,
      }) as GetAppDefinitionResponseParams)
  )
  .refine((arr) => arr.length === 1, { message: 'Expected single object in array' })
  .transform((arr) => arr[0] as GetAppDefinitionResponseParams);

const GetAppSessionsParamsSchema = z
  .array(
    z
      .object({
        app_session_id: hexSchema,
        status: statusEnum,
        participants: z.array(addressSchema),
        protocol: z.string(),
        challenge: z.number(),
        weights: z.array(z.number()),
        quorum: z.number(),
        version: z.number(),
        nonce: z.number(),
        created_at: z.union([z.string(), z.date()]).transform((v) => new Date(v)),
        updated_at: z.union([z.string(), z.date()]).transform((v) => new Date(v)),
      })
      .transform((s) => ({
        appSessionId: s.app_session_id as `0x${string}`,
        status: s.status as RPCChannelStatus,
        participants: s.participants as Address[],
        protocol: s.protocol,
        challenge: s.challenge,
        weights: s.weights,
        quorum: s.quorum,
        version: s.version,
        nonce: s.nonce,
        createdAt: s.created_at,
        updatedAt: s.updated_at,
      }) as GetAppSessionsResponseParams)
  )
  .transform(arr => arr as GetAppSessionsResponseParams[]);

const RPCAllocationSchema = z.object({
  destination: addressSchema,
  token: addressSchema,
  amount: z.union([z.string(), z.number()]).transform((a) => BigInt(a)),
});

const ServerSignatureSchema = z.object({
  v: z.string(),
  r: z.string(),
  s: z.string(),
});

const ResizeChannelParamsSchema = z
  .array(
    z.object({
      channel_id: hexSchema,
      state_data: z.string(),
      intent: z.number(),
      version: z.number(),
      allocations: z.array(RPCAllocationSchema),
      state_hash: z.string(),
      server_signature: ServerSignatureSchema,
    })
      .transform((raw) => ({
        channelId: raw.channel_id as `0x${string}`,
        stateData: raw.state_data,
        intent: raw.intent,
        version: raw.version,
        allocations: raw.allocations.map((a) => ({
          destination: a.destination as Address,
          token: a.token as Address,
          amount: a.amount,
        })),
        stateHash: raw.state_hash,
        serverSignature: raw.server_signature,
      }) as ResizeChannelResponseParams)
  )
  .refine((arr) => arr.length === 1, { message: 'Expected single object in array' })
  .transform((arr) => arr[0] as ResizeChannelResponseParams);

const CloseChannelParamsSchema = z
  .array(
    z.object({
      channel_id: hexSchema,
      state_data: z.string(),
      intent: z.number(),
      version: z.number(),
      allocations: z.array(RPCAllocationSchema),
      state_hash: z.string(),
      server_signature: ServerSignatureSchema,
    })
      .transform((raw) => ({
        channelId: raw.channel_id as `0x${string}`,
        stateData: raw.state_data,
        intent: raw.intent,
        version: raw.version,
        allocations: raw.allocations.map((a) => ({
          destination: a.destination as Address,
          token: a.token as Address,
          amount: a.amount,
        })),
        stateHash: raw.state_hash,
        serverSignature: raw.server_signature,
      }) as CloseChannelResponseParams)
  )
  .refine((arr) => arr.length === 1, { message: 'Expected single object in array' })
  .transform((arr) => arr[0] as CloseChannelResponseParams);

const GetChannelsParamsSchema = z
  .array(
    z
      .object({
        channel_id: hexSchema,
        participant: addressSchema,
        status: statusEnum,
        token: addressSchema,
        wallet: addressSchema,
        amount: z.union([z.string(), z.number()]).transform((a) => BigInt(a)),
        chain_id: z.number(),
        adjudicator: addressSchema,
        challenge: z.number(),
        nonce: z.number(),
        version: z.number(),
        created_at: z.union([z.string(), z.date()]).transform((v) => new Date(v)),
        updated_at: z.union([z.string(), z.date()]).transform((v) => new Date(v)),
      })
      .transform((c) => ({
        channelId: c.channel_id as `0x${string}`,
        participant: c.participant as Address,
        status: c.status as RPCChannelStatus,
        token: c.token as Address,
        wallet: c.wallet as Address,
        amount: c.amount,
        chainId: c.chain_id,
        adjudicator: c.adjudicator as Address,
        challenge: c.challenge,
        nonce: c.nonce,
        version: c.version,
        createdAt: c.created_at,
        updatedAt: c.updated_at,
      }) as GetChannelsResponseParams)
  )
  .transform(arr => arr as GetChannelsResponseParams[]);

const GetRPCHistoryParamsSchema = z
  .array(
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
      .transform((h) => ({
        id: h.id,
        sender: h.sender as Address,
        reqId: h.req_id,
        method: h.method,
        params: h.params,
        timestamp: h.timestamp,
        reqSig: h.req_sig as any,
        resSig: h.res_sig as any,
        response: h.response,
      }) as GetRPCHistoryResponseParams)
  )
  .transform(arr => arr as GetRPCHistoryResponseParams[]);

const GetAssetsParamsSchema = z
  .array(
    z
      .object({
        token: addressSchema,
        chain_id: z.number(),
        symbol: z.string(),
        decimals: z.number(),
      })
      .transform((a) => ({
        token: a.token as Address,
        chainId: a.chain_id,
        symbol: a.symbol,
        decimals: a.decimals,
      }) as GetAssetsResponseParams)
  )
  .transform(arr => arr as GetAssetsResponseParams[]);

const BalanceUpdateParamsSchema = z
  .array(
    z
      .object({
        asset: z.string(),
        amount: z.union([z.string(), z.number()]).transform((a) => BigInt(a)),
      })
      .transform((b) => ({
        asset: b.asset,
        amount: b.amount,
      }) as BalanceUpdateResponseParams)
  )
  .transform(arr => arr as BalanceUpdateResponseParams[]);

const ChannelUpdateParamsSchema = z
  .array(
    z.object({
      channel_id: hexSchema,
      participant: addressSchema,
      status: statusEnum,
      token: addressSchema,
      amount: z.union([z.string(), z.number()]).transform((a) => BigInt(a)),
      chain_id: z.number(),
      adjudicator: addressSchema,
      challenge: z.number(),
      nonce: z.number(),
      version: z.number(),
      created_at: z.union([z.string(), z.date()]).transform((v) => new Date(v)),
      updated_at: z.union([z.string(), z.date()]).transform((v) => new Date(v)),
    })
      .transform((c) => ({
        channelId: c.channel_id as `0x${string}`,
        participant: c.participant as Address,
        status: c.status as RPCChannelStatus,
        token: c.token as Address,
        amount: c.amount,
        chainId: c.chain_id,
        adjudicator: c.adjudicator as Address,
        challenge: c.challenge,
        nonce: c.nonce,
        version: c.version,
        createdAt: c.created_at,
        updatedAt: c.updated_at,
      }) as ChannelUpdateResponseParams)
  )
  .refine((arr) => arr.length === 1, { message: 'Expected single object in array' })
  .transform((arr) => arr[0] as ChannelUpdateResponseParams);

const ChannelsUpdateParamsSchema = z.array(ChannelUpdateParamsSchema).transform(arr => arr as ChannelUpdateResponseParams[]);

const TransferParamsSchema = z
  .array(
    z.object({
      from: addressSchema,
      to: addressSchema,
      allocations: z.array(
        z.object({
          asset: z.string(),
          amount: z.union([z.string(), z.number()]).transform((a) => a.toString()),
        })
      ),
      created_at: z.union([z.string(), z.date()]).transform((v) => new Date(v)),
    })
      .transform((raw) => ({
        from: raw.from as Address,
        to: raw.to as Address,
        allocations: raw.allocations,
        createdAt: raw.created_at,
      }) as TransferRPCResponseParams)
  )
  .refine((arr) => arr.length === 1, { message: 'Expected single object in array' })
  .transform((arr) => arr[0] as TransferRPCResponseParams);

const parseErrorParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.Error]> = (params) => ErrorParamsSchema.parse(params);
const parseGetLedgerBalancesParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.GetLedgerBalances]> = (params) => GetLedgerBalancesParamsSchema.parse(params);
const parseGetLedgerEntriesParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.GetLedgerEntries]> = (params) => GetLedgerEntriesParamsSchema.parse(params);
const parseGetAppSessionsParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.GetAppSessions]> = (params) => GetAppSessionsParamsSchema.parse(params);
const parseGetChannelsParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.GetChannels]> = (params) => GetChannelsParamsSchema.parse(params);
const parseGetRPCHistoryParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.GetRPCHistory]> = (params) => GetRPCHistoryParamsSchema.parse(params);
const parseGetAssetsParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.GetAssets]> = (params) => GetAssetsParamsSchema.parse(params);
const parseBalanceUpdateParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.BalanceUpdate]> = (params) => BalanceUpdateParamsSchema.parse(params);
const parseChannelsUpdateParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.ChannelsUpdate]> = (params) => ChannelsUpdateParamsSchema.parse(params);
const parseGetConfigParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.GetConfig]> = (params) => GetConfigParamsSchema.parse(params);
const parseAuthChallengeParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.AuthChallenge]> = (params) => AuthChallengeParamsSchema.parse(params);
const parseAuthVerifyParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.AuthVerify]> = (params) => AuthVerifyParamsSchema.parse(params);
const parseAuthRequestParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.AuthRequest]> = (params) => AuthRequestParamsSchema.parse(params);
const parseCreateAppSessionParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.CreateAppSession]> = (params) => CreateAppSessionParamsSchema.parse(params);
const parseSubmitStateParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.SubmitState]> = (params) => SubmitStateParamsSchema.parse(params);
const parseCloseAppSessionParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.CloseAppSession]> = (params) => CloseAppSessionParamsSchema.parse(params);
const parseGetAppDefinitionParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.GetAppDefinition]> = (params) => GetAppDefinitionParamsSchema.parse(params);
const parseResizeChannelParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.ResizeChannel]> = (params) => ResizeChannelParamsSchema.parse(params);
const parseCloseChannelParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.CloseChannel]> = (params) => CloseChannelParamsSchema.parse(params);
const parseMessageParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.Message]> = (params) => {
  if (!Array.isArray(params) || params.length === 0) throw new ParserParamsMissingError(RPCMethod.Message);
  return params[0]; // Message params are app-specific, skip zod
};
const parseChannelUpdateParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.ChannelUpdate]> = (params) => ChannelUpdateParamsSchema.parse(params);
const parseTransferParams: ParamsParser<RPCResponseParamsByMethod[RPCMethod.Transfer]> = (params) => TransferParamsSchema.parse(params);

const noop: ParamsParser<any> = (_) => { return {}; };

export const paramsParsers: Record<RPCMethod, ParamsParser<any>> = {
  [RPCMethod.AuthChallenge]: parseAuthChallengeParams,
  [RPCMethod.AuthVerify]: parseAuthVerifyParams,
  [RPCMethod.AuthRequest]: parseAuthRequestParams,
  [RPCMethod.Error]: parseErrorParams,
  [RPCMethod.GetConfig]: parseGetConfigParams,
  [RPCMethod.GetLedgerBalances]: parseGetLedgerBalancesParams,
  [RPCMethod.GetLedgerEntries]: parseGetLedgerEntriesParams,
  [RPCMethod.CreateAppSession]: parseCreateAppSessionParams,
  [RPCMethod.SubmitState]: parseSubmitStateParams,
  [RPCMethod.CloseAppSession]: parseCloseAppSessionParams,
  [RPCMethod.GetAppDefinition]: parseGetAppDefinitionParams,
  [RPCMethod.GetAppSessions]: parseGetAppSessionsParams,
  [RPCMethod.ResizeChannel]: parseResizeChannelParams,
  [RPCMethod.CloseChannel]: parseCloseChannelParams,
  [RPCMethod.GetChannels]: parseGetChannelsParams,
  [RPCMethod.GetRPCHistory]: parseGetRPCHistoryParams,
  [RPCMethod.GetAssets]: parseGetAssetsParams,
  [RPCMethod.Assets]: parseGetAssetsParams,
  [RPCMethod.Message]: parseMessageParams,
  [RPCMethod.BalanceUpdate]: parseBalanceUpdateParams,
  [RPCMethod.ChannelsUpdate]: parseChannelsUpdateParams,
  [RPCMethod.ChannelUpdate]: parseChannelUpdateParams,
  [RPCMethod.Ping]: noop,
  [RPCMethod.Pong]: noop,
  [RPCMethod.Transfer]: parseTransferParams,
};
