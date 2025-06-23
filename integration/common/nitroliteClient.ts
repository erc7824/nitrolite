import {
    Allocation,
    ChannelsUpdateRPCResponse,
    CloseChannelRPCResponse,
    createCloseChannelMessage,
    NitroliteClient,
    parseRPCResponse,
    RPCChannelStatus,
} from '@erc7824/nitrolite';
import { Identity } from './identity';
import { Address, createPublicClient, Hex, http } from 'viem';
import { chain, CONFIG } from './setup';
import { getChannelUpdatePredicateWithStatus, getCloseChannelPredicate, TestWebSocket } from './ws';
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
    constructor(private identity: Identity) {
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
        const openChannelPromise = ws.waitForMessage(
            getChannelUpdatePredicateWithStatus(RPCChannelStatus.Open),
            undefined,
            5000
        );

        const { initialState } = await this.depositAndCreateChannel(tokenAddress, amount, {
            initialAllocationAmounts: [amount, BigInt(0)],
            stateData: '0x',
        });

        const openResponse = await openChannelPromise;

        const openParsedResponse = parseRPCResponse(openResponse) as ChannelsUpdateRPCResponse;
        const responseChannel = openParsedResponse.params[0];

        return { params: responseChannel, initialState };
    };

    closeAndWithdrawChannel = async (ws: TestWebSocket, channelId: Hex) => {
        const msg = await createCloseChannelMessage(
            this.identity.messageSigner,
            channelId,
            this.identity.walletAddress
        );

        const closeResponse = await ws.sendAndWaitForResponse(msg, getCloseChannelPredicate(), 1000);
        const closeParsedResponse = parseRPCResponse(closeResponse) as CloseChannelRPCResponse;

        const closeChannelUpdateChannelPromise = ws.waitForMessage(
            getChannelUpdatePredicateWithStatus(RPCChannelStatus.Closed),
            undefined,
            5000
        );

        await this.closeChannel({
            finalState: {
                channelId: closeParsedResponse.params.channel_id,
                stateData: closeParsedResponse.params.state_data as Hex,
                allocations: [
                    {
                        destination: closeParsedResponse.params.allocations[0].destination as Address,
                        token: closeParsedResponse.params.allocations[0].token as Address,
                        amount: BigInt(closeParsedResponse.params.allocations[0].amount),
                    },
                    {
                        destination: closeParsedResponse.params.allocations[1].destination as Address,
                        token: closeParsedResponse.params.allocations[1].token as Address,
                        amount: BigInt(closeParsedResponse.params.allocations[1].amount),
                    },
                ] as [Allocation, Allocation],
                version: BigInt(closeParsedResponse.params.version),
                serverSignature: {
                    v: +closeParsedResponse.params.server_signature.v,
                    r: closeParsedResponse.params.server_signature.r as Hex,
                    s: closeParsedResponse.params.server_signature.s as Hex,
                },
            },
            stateData: closeParsedResponse.params.state_data as Hex,
        });

        const closeChannelUpdateResponse = await closeChannelUpdateChannelPromise;
        const closeChannelUpdateParsedResponse = parseRPCResponse(
            closeChannelUpdateResponse
        ) as ChannelsUpdateRPCResponse;
        const responseChannel = closeChannelUpdateParsedResponse.params[0];

        return { params: responseChannel };
    };
}
