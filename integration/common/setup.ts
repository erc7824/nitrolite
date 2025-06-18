import { NitroliteClient } from '@erc7824/nitrolite';
import { Hex } from 'viem';

export const CONFIG = {
    CLEARNODE_URL: 'ws://localhost:8000/ws',
    DEBUG_MODE: false,

    IDENTITIES: [
        {
            walletPrivateKey: '0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d' as Hex,
            sessionPrivateKey: '0x6ad7995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d' as Hex,
        },
    ],

    CHAIN_ID: 31337,
};
