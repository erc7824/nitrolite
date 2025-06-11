import { type WalletClient, type Transport, type Chain, type Account, type ParseAccount, keccak256, Hex } from 'viem';

/**
 * A prefix used to ensure the generated private key is unique to this specific derivation scheme.
 * This prevents collisions with other signature-based derivation methods.
 */
const DERIVATION_PATH_PREFIX = 'nitrolite_state_wallet_v1';

/**
 * Generates a deterministic private key for a session based on the wallet client,
 * adjudicator address, app address, and a unique nonce. This key can be used
 * for actions within a specific application context (e.g., a state channel) without
 * requiring the user to sign every action with their main wallet.
 *
 * @warning This function generates a key that is valid across all chains where the
 * adjudicator and app contracts are deployed at the same addresses. This provides a
 * consistent session address for multi-chain dApps but removes chain-specific replay
 * protection from the key generation itself.
 *
 * @param walletClient The Viem WalletClient instance of the user's primary wallet.
 * @param adjudicatorAddress The contract address of the adjudicator for the state channel.
 * @param appAddress The contract address of the application logic for the state channel.
 * @param nonce A unique identifier for the channel or session, preventing key reuse.
 * @returns A promise that resolves to a securely derived private key as a Hex string.
 */
export const createDeterministicPrivateKey = async (
    walletClient: WalletClient<Transport, Chain, ParseAccount<Account>>,
    adjudicatorAddress: string,
    appAddress: string,
    nonce: number,
): Promise<Hex> => {
    if (!walletClient.account) {
        throw new Error('WalletClient must have an account to sign the message.');
    }

    // The message to be signed by the user's main wallet.
    // It includes the necessary context to derive a unique key for the application session.
    // By omitting the chainId, the same signature will be produced for the same
    // context on any chain, resulting in the same private key and address.
    // Format: DERIVATION_PATH_PREFIX/adjudicatorAddress/appAddress/userAddress/nonce
    const messageToSign = [
        DERIVATION_PATH_PREFIX,
        adjudicatorAddress,
        appAddress,
        walletClient.account.address,
        nonce,
    ].join('/');

    // The user signs the structured message with their main wallet.
    // This signature acts as the seed for the deterministic private key.
    // It proves ownership of the main account and consent to generate the session key.
    const seedSignature = await walletClient.signMessage({
        message: messageToSign,
    });

    const privateKey = keccak256(seedSignature);

    return privateKey;
};
