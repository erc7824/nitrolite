import { Address, defineChain, Hex } from 'viem';
import { localhost } from 'viem/chains';

export const CONFIG = {
    CLEARNODE_URL: 'ws://localhost:8000/ws',
    DEBUG_MODE: false,

    IDENTITIES: [
        {
            // 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
            WALLET_PK: '0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d' as Hex,
            // 0xCFEA8Fd97A33D4cb68aC61eDE98Bc18D07eff004
            SESSION_PK: '0x60d7995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d' as Hex,
            // 0x831cA6915B827C87caec7AbaB3b27b1357cB9a6A
            APP_SESSION_PK: '0x71e8995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d' as Hex,
        },
        {
            // 0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC
            WALLET_PK: '0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a' as Hex,
            // 0x1c7678FfCBA8f1573201156D1de213e9e085C95f
            SESSION_PK: '0x6ef5111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a' as Hex,
            // 0x9265cF9A5Fe3eF3DD20c9318b64C9108Ef3c0255
            APP_SESSION_PK: '0x7f06111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a' as Hex,
        },
    ],

    CHAIN_ID: 31337,
    TOKEN_SYMBOL: 'yintegration.usd',
    ADDRESSES: {
        CUSTODY_ADDRESS: '0x8658501c98C3738026c4e5c361c6C3fa95DfB255' as Address,
        DUMMY_ADJUDICATOR_ADDRESS: '0xcbbc03a873c11beeFA8D99477E830be48d8Ae6D7' as Address,
        CLEARNODE_ADDRESS: '0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266' as Address,
        USDC_TOKEN_ADDRESS: '0xbD24c53072b9693A35642412227043Ffa5fac382' as Address,
        // Wrapped Ether token for multi-asset testing
        WETH_TOKEN_ADDRESS: '0xAf119209932D7EDe63055E60854E81acC4063a12' as Address,
    },
    DEFAULT_CHALLENGE_TIMEOUT: 3600,

    DATABASE_NAME: 'postgres',
    DATABASE_USER: 'postgres',
    DATABASE_PASSWORD: 'postgres',
    DATABASE_HOST: 'localhost',
    DATABASE_PORT: 5432,
};

export const chain = defineChain({ ...localhost, id: CONFIG.CHAIN_ID });
