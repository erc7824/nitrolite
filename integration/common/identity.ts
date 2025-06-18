import { Address, createWalletClient, defineChain, Hex, http, WalletClient } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';
import { localhost } from 'viem/chains';
import { CONFIG } from './setup';
import { createEIP712AuthMessageSigner, MessageSigner, RequestData, ResponsePayload } from '@erc7824/nitrolite';
import { ethers } from 'ethers';

export class Identity {
    public walletClient = null;
    public walletAddress: Address;
    public sessionAddress: Address;

    constructor(privateWalletPrivateKey: Hex, private sessionPrivateKey: Hex) {
        const walletAccount = privateKeyToAccount(privateWalletPrivateKey);

        this.walletClient = createWalletClient({
            account: walletAccount,
            chain: defineChain({ ...localhost, id: CONFIG.CHAIN_ID }),
            transport: http(),
        });

        this.walletAddress = walletAccount.address;

        const sessionAccount = privateKeyToAccount(sessionPrivateKey);
        this.sessionAddress = sessionAccount.address;
    }

    getMessageSigner(): MessageSigner {
        return async (payload: RequestData | ResponsePayload): Promise<Hex> => {
            try {
                const wallet = new ethers.Wallet(this.sessionPrivateKey);
                const messageBytes = ethers.utils.arrayify(ethers.utils.id(JSON.stringify(payload)));
                const flatSignature = await wallet._signingKey().signDigest(messageBytes);
                const signature = ethers.utils.joinSignature(flatSignature);

                return signature as Hex;
            } catch (error) {
                console.error('Error signing message:', error);
                throw error;
            }
        };
    }
}
