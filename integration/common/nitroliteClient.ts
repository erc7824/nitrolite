import {
    Allocation,
    convertRPCToClientChannel,
    convertRPCToClientState,
    createCloseChannelMessage,
    createCreateChannelMessage,
    NitroliteClient,
    parseChannelUpdateResponse,
    parseCloseChannelResponse,
    parseCreateChannelResponse,
    RPCChannelStatus,
} from '@erc7824/nitrolite';
import { Identity } from './identity';
import { Address, createPublicClient, Hex, http } from 'viem';
import { chain, CONFIG } from './setup';
import {
    getChannelUpdatePredicateWithStatus,
    getCloseChannelPredicate,
    getCreateChannelPredicate,
    TestWebSocket,
} from './ws';

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
        { tokenAddress, amount, depositAmount }: { tokenAddress: Address; amount: bigint; depositAmount?: bigint }
    ) => {
        const msg = await createCreateChannelMessage(this.identity.messageSigner, {
            chain_id: chain.id,
            token: tokenAddress,
            amount,
        });
        const createResponse = await ws.sendAndWaitForResponse(msg, getCreateChannelPredicate(), 5000);
        expect(createResponse).toBeDefined();

        const { params: createParsedResponseParams } = parseCreateChannelResponse(createResponse);

        const openChannelPromise = ws.waitForMessage(
            getChannelUpdatePredicateWithStatus(RPCChannelStatus.Open),
            undefined,
            5000
        );

        depositAmount = depositAmount ?? amount;
        const { initialState } = await this.depositAndCreateChannel(tokenAddress, depositAmount, {
            unsignedInitialState: convertRPCToClientState(
                createParsedResponseParams.state,
                createParsedResponseParams.serverSignature
            ),
            channel: convertRPCToClientChannel(createParsedResponseParams.channel),
            serverSignature: createParsedResponseParams.serverSignature,
        });

        const openResponse = await openChannelPromise;

        const openParsedResponse = parseChannelUpdateResponse(openResponse);
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
        const closeParsedResponse = parseCloseChannelResponse(closeResponse);

        const closeChannelUpdateChannelPromise = ws.waitForMessage(
            getChannelUpdatePredicateWithStatus(RPCChannelStatus.Closed),
            undefined,
            5000
        );

        await this.closeChannel({
            finalState: {
                intent: closeParsedResponse.params.state.intent,
                channelId: closeParsedResponse.params.channelId,
                data: closeParsedResponse.params.state.stateData,
                allocations: [
                    {
                        destination: closeParsedResponse.params.state.allocations[0].destination as Address,
                        token: closeParsedResponse.params.state.allocations[0].token as Address,
                        amount: closeParsedResponse.params.state.allocations[0].amount,
                    },
                    {
                        destination: closeParsedResponse.params.state.allocations[1].destination as Address,
                        token: closeParsedResponse.params.state.allocations[1].token as Address,
                        amount: closeParsedResponse.params.state.allocations[1].amount,
                    },
                ] as [Allocation, Allocation],
                version: BigInt(closeParsedResponse.params.state.version),
                serverSignature: closeParsedResponse.params.serverSignature,
            },
            stateData: closeParsedResponse.params.state.stateData,
        });

        const closeChannelUpdateResponse = await closeChannelUpdateChannelPromise;
        const closeChannelUpdateParsedResponse = parseChannelUpdateResponse(closeChannelUpdateResponse);
        const responseChannel = closeChannelUpdateParsedResponse.params;

        return { params: responseChannel };
    };
}
