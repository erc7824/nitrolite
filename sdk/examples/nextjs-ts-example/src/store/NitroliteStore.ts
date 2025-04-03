import { ChannelContext, NitroliteClient } from '@erc7824/nitrolite';
import { proxy } from 'valtio';

export interface IWalletState {
    client: typeof NitroliteClient;

    channelContext: { [key: string]: ChannelContext };
}

const state = proxy<IWalletState>({
    client: NitroliteClient,
    channelContext: {},
});

const NitroliteStore = {
    state,
};

export default NitroliteStore;
