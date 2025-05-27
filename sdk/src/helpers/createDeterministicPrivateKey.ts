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
 * Generates a deterministic private key for wallet clients based on the provided adjudicator and app addresses, and a nonce.
 */
export const createDeterministicPrivateKey = async (
    walletClient: WalletClient<Transport, Chain, ParseAccount<Account>>,
    adjudicatorAddress: string,
    appAddress: string,
    nonce: number = 0,
): Promise<Hex> => {
    // NOTE: keep in mind that this key should be used only for the respective adjudicator and app address
    const nonceMessage = `${walletClient.account.address}/${adjudicatorAddress}/${appAddress}/0/${nonce}`

    const seed = await walletClient.signMessage({
        message: nonceMessage,
    });

    const input = DERIVATION_PATH_PREFIX + seed;

    const privateKey = keccak256(toHex(input));

    return privateKey;
};
