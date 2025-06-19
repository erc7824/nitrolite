import { ChannelsUpdateRPCResponse, NitroliteClient, parseRPCResponse, RPCChannelStatus } from '@erc7824/nitrolite';
import { Identity } from './identity';
import { Address, createPublicClient, http } from 'viem';
import { chain, CONFIG } from './setup';
import { getChannelUpdatePredicateWithStatus, TestWebSocket } from './ws';
import { BlockchainUtils } from './blockchainUtils';

// export const createNitroliteClientFromIdentity = (identity: Identity): NitroliteClient => {
//     const publicClient = createPublicClient({
//         chain,
//         transport: http(),
//     });

//     return new NitroliteClient({
//         // @ts-ignore
//         publicClient,
//         walletClient: identity.walletClient,
//         stateWalletClient: identity.stateWalletClient,
//         account: identity.walletClient.account,
//         chainId: chain.id,
//         challengeDuration: BigInt(CONFIG.DEFAULT_CHALLENGE_TIMEOUT), // min
//         addresses: {
//             custody: CONFIG.ADDRESSES.CUSTODY_ADDRESS,
//             adjudicator: CONFIG.ADDRESSES.DUMMY_ADJUDICATOR_ADDRESS,
//             guestAddress: CONFIG.ADDRESSES.GUEST_ADDRESS,
//         },
//     });
// };

export class TestNitroliteClient extends NitroliteClient {
    constructor(identity: Identity) {
        const publicClient = createPublicClient({
            chain,
            transport: http(),
        });

        super({
            // @ts-ignore
            publicClient,
            walletClient: identity.walletClient,
            stateWalletClient: identity.stateWalletClient,
            account: identity.walletClient.account,
            chainId: chain.id,
            challengeDuration: BigInt(CONFIG.DEFAULT_CHALLENGE_TIMEOUT), // min
            addresses: {
                custody: CONFIG.ADDRESSES.CUSTODY_ADDRESS,
                adjudicator: CONFIG.ADDRESSES.DUMMY_ADJUDICATOR_ADDRESS,
                guestAddress: CONFIG.ADDRESSES.GUEST_ADDRESS,
            },
        });
    }

    createAndWaitForChannel = async (
        ws: TestWebSocket,
        { tokenAddress, amount }: { tokenAddress: Address; amount: bigint }
    ) => {
        const openChannelPromise = ws.waitForMessage(getChannelUpdatePredicateWithStatus(RPCChannelStatus.Open), 5000);

        const { initialState } = await this.depositAndCreateChannel(tokenAddress, amount, {
            initialAllocationAmounts: [amount, BigInt(0)],
            stateData: '0x',
        });

        const openResponse = await openChannelPromise;

        const openParsedResponse = parseRPCResponse(openResponse) as ChannelsUpdateRPCResponse;
        const responseChannel = openParsedResponse.params[0];

        return { params: responseChannel, initialState };
    };
}
