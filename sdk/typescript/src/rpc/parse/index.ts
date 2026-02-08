import { RPCMethod } from '../types';
import { ParamsParser, noop } from './common';
import { authParamsParsers } from './session_keys';
import { ledgerParamsParsers } from './ledger';
import { appParamsParsers } from './app';
import { channelParamsParsers } from './channel';
import { assetParamsParsers } from './asset';
import { miscParamsParsers } from './misc';

export const paramsParsers = {
    ...authParamsParsers,
    ...ledgerParamsParsers,
    ...appParamsParsers,
    ...channelParamsParsers,
    ...assetParamsParsers,
    ...miscParamsParsers,

    // Methods with no params
    [RPCMethod.Ping]: noop,
} as Record<RPCMethod, ParamsParser<unknown>>;
