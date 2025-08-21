import { Account, Address, Chain, Hex, ParseAccount, toHex, Transport, WalletClient } from 'viem';
import { State, UnsignedState } from './types';
import { getPackedState, getStateHash } from '../utils';
import { signRawECDSAMessage } from '../utils/sign';
import { privateKeyToAccount } from 'viem/accounts';

// TODO: perhaps extend this interface with rpc signing methods and use it as universal signer interface

/**
 * Interface for signing protocol states.
 * This interface is used to abstract the signing logic for state updates in the Nitrolite SDK.
 * It allows for different implementations, such as using a wallet client or a session key.
 * Also implementation could include data packing/encoding, which is crucial for some signatures (EIP-712, EIP-191)
 */
export interface StateSigner {
    /**
     * Get the address of the signer.
     * @returns The address of the signer.
     */
    getAddress(): Address;
    /**
     * Sign a state for a given channel ID.
     * @param channelId The ID of the channel.
     * @param state The state to sign.
     * @returns A Promise that resolves to the signature as a Hex string.
     */
    signState(channelId: Hex, state: UnsignedState): Promise<Hex>;
    /**
     * Sign a raw message.
     * @param message The message to sign as a Hex string.
     * @returns A Promise that resolves to the signature as a Hex string.
     * @dev use viem's `toHex` to convert the message to Hex if needed.
     */
    signRawMessage(message: Hex): Promise<Hex>;
}

/**
 * Implementation of the StateSigner interface using a viem WalletClient.
 * This class uses the wallet client to sign states and raw messages.
 * It is suitable for use in scenarios where the wallet client is available and can sign messages,
 * e.g. signing with MetaMask or other wallet providers.
 */
export class WalletStateSigner implements StateSigner {
    private readonly walletClient: WalletClient<Transport, Chain, ParseAccount<Account>>;

    constructor(walletClient: WalletClient<Transport, Chain, ParseAccount<Account>>) {
        this.walletClient = walletClient;
    }

    getAddress(): Address {
        return this.walletClient.account.address;
    }

    async signState(channelId: Hex, state: State): Promise<Hex> {
        const packedState = getPackedState(channelId, state)

        return this.walletClient.signMessage({ message: { raw: packedState } });
    }

    async signRawMessage(message: Hex): Promise<Hex> {
        return this.walletClient.signMessage({ message: { raw: message } });
    }
}

/**
 * Implementation of the StateSigner interface using a session key.
 * This class uses a session key to sign states and raw messages.
 * It is suitable for scenarios where a session key is used for signing and private key could be exposed to application.
 */
export class SessionKeyStateSigner implements StateSigner {
    private readonly sessionKey: Hex;
    private readonly account: Account;

    constructor(sessionKey: Hex) {
        this.sessionKey = sessionKey;
        this.account = privateKeyToAccount(sessionKey);
    }

    getAddress(): Address {
        return this.account.address;
    }

    async signState(channelId: Hex, state: State): Promise<Hex> {
        const packedState = getPackedState(channelId, state);

        return signRawECDSAMessage(packedState, this.sessionKey);
    }

    async signRawMessage(message: Hex): Promise<Hex> {
        return signRawECDSAMessage(message, this.sessionKey);
    }
}
