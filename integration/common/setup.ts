import { Address, defineChain, Hex } from 'viem';
import { localhost } from 'viem/chains';

export const CONFIG = {
    CLEARNODE_URL: 'ws://localhost:8000/ws',
    DEBUG_MODE: false,

    IDENTITIES: [
        {
            // 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
            WALLET_PK: '0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d' as Hex,
            // 0xf24b3419C0f9aB9cCD9447340232FA4763F1718c
            SESSION_PK: '0x6ad7995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d' as Hex,
        },
    ],

    CHAIN_ID: 31337,
    ADDRESSES: {
        CUSTODY_ADDRESS: '0x8658501c98C3738026c4e5c361c6C3fa95DfB255' as Address,
        DUMMY_ADJUDICATOR_ADDRESS: '0xcbbc03a873c11beeFA8D99477E830be48d8Ae6D7' as Address,
        GUEST_ADDRESS: '0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266' as Address,
        USDC_TOKEN_ADDRESS: '0xbD24c53072b9693A35642412227043Ffa5fac382' as Address,
    },
    DEFAULT_CHALLENGE_TIMEOUT: 3600,
};

export const chain = defineChain({ ...localhost, id: CONFIG.CHAIN_ID });
