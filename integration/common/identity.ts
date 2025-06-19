import { Address, createWalletClient, Hex, http, keccak256, stringToBytes } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';
import { chain } from './setup';
import { ethers } from 'ethers';
import { createECDSAMessageSigner } from '@erc7824/nitrolite';

export class Identity {
    public walletClient = null;
    public stateWalletClient = null;
    public walletAddress: Address;
    public sessionAddress: Address;
    public messageSigner = null;

    constructor(privateWalletPrivateKey: Hex, sessionPrivateKey: Hex) {
        const walletAccount = privateKeyToAccount(privateWalletPrivateKey);
        this.walletAddress = walletAccount.address;

        this.walletClient = createWalletClient({
            account: walletAccount,
            chain,
            transport: http(),
        });

        const sessionAccount = privateKeyToAccount(sessionPrivateKey);
        this.sessionAddress = sessionAccount.address;

        this.stateWalletClient = {
            ...this.walletClient,
            account: {
                address: this.sessionAddress,
            },
            signMessage: async ({ message: { raw } }: { message: { raw: string } }) => {
                // const messageBytes = keccak256(stringToBytes(JSON.stringify(raw)));
                // const flatSignature = await sessionAccount.sign({ hash: messageBytes });

                // return flatSignature as Hex;

                const wallet = new ethers.Wallet(sessionPrivateKey);

                const flatSignature = await wallet._signingKey().signDigest(raw);

                const signature = ethers.utils.joinSignature(flatSignature);

                return signature as Hex;
            },
        };

        this.messageSigner = createECDSAMessageSigner(sessionPrivateKey);
    }
}
