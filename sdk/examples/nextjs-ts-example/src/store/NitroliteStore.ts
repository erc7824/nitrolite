import { Message } from '@/types';
import { AppLogic, ChannelContext, NitroliteClient, Signature, Channel as NitroliteChannel } from '@erc7824/nitrolite';
import { proxy } from 'valtio';
import { Address } from 'viem';

export interface IWalletState {
    client: NitroliteClient;

    channelContext: Record<string, ChannelContext<Message>>;
}

const state = proxy<IWalletState>({
    client: null,
    channelContext: {} as Record<string, ChannelContext<Message>>,
});

const NitroliteStore = {
    state,

    setClient(client: NitroliteClient) {
        if (!client) {
            console.error('Attempted to set null or undefined Nitrolite client');
            return false;
        }
        state.client = client;
        return true;
    },

    setChannelContext(channelId: NitroliteChannel | string, guest: Address, app: AppLogic<Message>) {
        try {
            if (!state.client) {
                throw new Error('Nitrolite client not initialized');
            }
            
            // Convert channel to string for use as an object key
            const key = typeof channelId === 'string' ? channelId : JSON.stringify(channelId);
            
            state.channelContext[key] = new ChannelContext<Message>(state.client, guest, app);
            return true;
        } catch (error) {
            console.error(`Failed to set channel context for ${typeof channelId === 'string' ? channelId : 'complex channel'}:`, error);
            return false;
        }
    },

    async deposit(channelId: NitroliteChannel | string, tokenAddress: Address, amount: string) {
        try {
            // Convert channel to string for use as an object key
            const key = typeof channelId === 'string' ? channelId : JSON.stringify(channelId);
            
            if (!state.channelContext[key]) {
                throw new Error(`Channel context not found for channel: ${key}`);
            }
            await state.channelContext[key].deposit(tokenAddress, BigInt(amount));
            return true;
        } catch (error) {
            console.error(`Failed to deposit to channel ${typeof channelId === 'string' ? channelId : 'complex channel'}:`, error);
            return false;
        }
    },

    async openChannel(
        channelId: NitroliteChannel | string,
        appState: Message,
        token: Address,
        allocations: [bigint, bigint],
        signatures: Signature[] = [],
    ) {
        try {
            // Convert channel to string for use as an object key
            const key = typeof channelId === 'string' ? channelId : JSON.stringify(channelId);
            
            if (!state.channelContext[key]) {
                throw new Error(`Channel context not found for channel: ${key}`);
            }
            await state.channelContext[key].open(appState, token, allocations, signatures);
            return true;
        } catch (error) {
            console.error(`Failed to open channel ${typeof channelId === 'string' ? channelId : 'complex channel'}:`, error);
            return false;
        }
    },
};

export default NitroliteStore;
