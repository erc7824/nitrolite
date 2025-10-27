import { Address, createWalletClient, Hex, http } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';
import { chain } from './setup';
import { createECDSAMessageSigner } from '@erc7824/nitrolite';
import { SessionKeyStateSigner } from '@erc7824/nitrolite/dist/client/signer';

export class Identity {
    public walletClient = null;
    public stateSigner = null;
    public walletStateSigner = null;
    public walletAddress: Address;
    public sessionAddress: Address;
    public messageSigner = null;
    public walletMessageSigner = null;

    constructor(walletPrivateKey: Hex, sessionPrivateKey: Hex) {
        const walletAccount = privateKeyToAccount(walletPrivateKey);
        this.walletAddress = walletAccount.address;

        this.walletClient = createWalletClient({
            account: walletAccount,
            chain,
            transport: http(),
        });

        this.stateSigner = new SessionKeyStateSigner(sessionPrivateKey);
        this.messageSigner = createECDSAMessageSigner(sessionPrivateKey);
        this.sessionAddress = this.stateSigner.getAddress();

        this.walletStateSigner = new SessionKeyStateSigner(walletPrivateKey);
        this.walletMessageSigner = createECDSAMessageSigner(walletPrivateKey);
    }
}
