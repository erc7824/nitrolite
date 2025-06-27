import { Address, createWalletClient, Hex, http } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';
import { chain } from './setup';
import { createECDSAMessageSigner } from '@erc7824/nitrolite';

export class Identity {
    public walletClient = null;
    public stateWalletClient = null;
    public walletAddress: Address;
    public sessionAddress: Address;
    public messageSigner = null;

    constructor(walletPrivateKey: Hex, sessionPrivateKey: Hex) {
        const walletAccount = privateKeyToAccount(walletPrivateKey);
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
                const flatSignature = await sessionAccount.sign({ hash: raw as Hex });

                return flatSignature as Hex;
            },
        };

        this.messageSigner = createECDSAMessageSigner(sessionPrivateKey);
    }
}
