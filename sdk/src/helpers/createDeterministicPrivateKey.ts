import {
    type WalletClient,
    type Transport,
    type Chain,
    type Account,
    type ParseAccount,
    toHex,
    keccak256,
    Hex,
} from 'viem';

const DERIVATION_PATH_PREFIX = 'nitrolite_state_wallet_v1_';

/**
 * Generates a deterministic private key from a seed and optional salt
 */
export const createDeterministicPrivateKey = async (
    walletClient: WalletClient<Transport, Chain, ParseAccount<Account>>,
    derivationIndex: number = 0,
    salt: string = ''
): Promise<Hex> => {
    const nonceMessage = `Nitrolite derivation index: ${derivationIndex}`;

    const seed = await walletClient.signMessage({
        message: nonceMessage,
    });

    const input = DERIVATION_PATH_PREFIX + seed + salt;

    const privateKey = keccak256(toHex(input));

    return privateKey;
};
