import {
    Allocation,
    createCloseChannelMessage,
    NitroliteClient,
    RPCChannelStatus,
    rpcResponseParser,
} from '@erc7824/nitrolite';
import { Identity } from './identity';
import { Address, createPublicClient, Hex, http } from 'viem';
import { chain, CONFIG } from './setup';
import { getChannelUpdatePredicateWithStatus, getCloseChannelPredicate, TestWebSocket } from './ws';

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
            stateSigner: identity.stateSigner,
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
        { tokenAddress, amount, depositAmount }: { tokenAddress: Address; amount: bigint, depositAmount?: bigint }
    ) => {
        depositAmount = depositAmount ?? amount;

        const openChannelPromise = ws.waitForMessage(
            getChannelUpdatePredicateWithStatus(RPCChannelStatus.Open),
            undefined,
            5000
        );

        const { initialState } = await this.depositAndCreateChannel(tokenAddress, depositAmount, {
            initialAllocationAmounts: [amount, BigInt(0)],
            stateData: '0x',
        });

        const openResponse = await openChannelPromise;

        const openParsedResponse = rpcResponseParser.channelUpdate(openResponse);
        const responseChannel = openParsedResponse.params;

        return { params: responseChannel, initialState };
    };

    closeAndWithdrawChannel = async (ws: TestWebSocket, channelId: Hex) => {
        const msg = await createCloseChannelMessage(
            this.identity.messageSigner,
            channelId,
            this.identity.walletAddress
        );

        const closeResponse = await ws.sendAndWaitForResponse(msg, getCloseChannelPredicate(), 1000);
        const closeParsedResponse = rpcResponseParser.closeChannel(closeResponse);

        const closeChannelUpdateChannelPromise = ws.waitForMessage(
            getChannelUpdatePredicateWithStatus(RPCChannelStatus.Closed),
            undefined,
            5000
        );

        await this.closeChannel({
            finalState: {
                intent: closeParsedResponse.params.intent,
                channelId: closeParsedResponse.params.channelId as Hex,
                data: closeParsedResponse.params.stateData as Hex,
                allocations: [
                    {
                        destination: closeParsedResponse.params.allocations[0].destination as Address,
                        token: closeParsedResponse.params.allocations[0].token as Address,
                        amount: closeParsedResponse.params.allocations[0].amount,
                    },
                    {
                        destination: closeParsedResponse.params.allocations[1].destination as Address,
                        token: closeParsedResponse.params.allocations[1].token as Address,
                        amount: closeParsedResponse.params.allocations[1].amount,
                    },
                ] as [Allocation, Allocation],
                version: BigInt(closeParsedResponse.params.version),
                serverSignature: closeParsedResponse.params.serverSignature,
            },
            stateData: closeParsedResponse.params.stateData as Hex,
        });

        const closeChannelUpdateResponse = await closeChannelUpdateChannelPromise;
        const closeChannelUpdateParsedResponse = rpcResponseParser.channelUpdate(closeChannelUpdateResponse);
        const responseChannel = closeChannelUpdateParsedResponse.params;

        return { params: responseChannel };
    };
}
