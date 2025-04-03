import { Channel as ChannelType, Message } from '@/types';
import { AppLogic, ChannelContext, NitroliteClient, Signature } from '@erc7824/nitrolite';
import { proxy } from 'valtio';
import { Address } from 'viem';

export interface IWalletState {
    client: NitroliteClient;

    channelContext: Record<ChannelType, ChannelContext>;
}

const state = proxy<IWalletState>({
    client: null,
    channelContext: {} as Record<ChannelType, ChannelContext<Message>>,
});

const NitroliteStore = {
    state,

    setClient(client: NitroliteClient) {
        state.client = client;
    },

    setChannelContext(channelType: ChannelType, guest: Address, app: AppLogic<Message>) {
        state.channelContext[channelType] = new ChannelContext<Message>(state.client, guest, app);
    },

    async deposit(channel: ChannelType, tokenAddress: Address, amount: string) {
        await state.channelContext[channel].deposit(tokenAddress, BigInt(amount));
    },

    async openChannel(channel: ChannelType, appState: Message, token: Address, allocations: [bigint, bigint], signatures: Signature[] = []) {
        await state.channelContext[channel].open(appState, token, allocations, signatures);
    },
};

export default NitroliteStore;
