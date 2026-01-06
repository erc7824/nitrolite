import { BlockchainUtils } from '@/blockchainUtils';
import { DatabaseUtils } from '@/databaseUtils';
import { Identity } from '@/identity';
import { TestNitroliteClient } from '@/nitroliteClient';
import { CONFIG, chain } from '@/setup';
import { toRaw } from '@/testHelpers';
import { TestWebSocket, getCreateChannelPredicate } from '@/ws';
import {
    Channel,
    State,
    StateIntent,
    SessionKeyStateSigner,
    getChannelId,
    ChannelStatus,
    createCreateChannelMessage,
    parseCreateChannelResponse,
    convertRPCToClientState,
    convertRPCToClientChannel,
} from '@erc7824/nitrolite';
import { Hex, parseUnits } from 'viem';
import { createAuthSessionWithClearnode } from '@/auth';

// NOTE: make sure this private key matches the one used to run integration tests
const CLEARNODE_PRIVATE_KEY: Hex = '0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80';

describe('Backward compatibility: participant as Wallet or Session Key', () => {
    const depositAmount = parseUnits('1000', 6); // 1000 USDC
    const resizeAmount = parseUnits('500', 6); // 500 USDC

    let ws: TestWebSocket;
    let identity: Identity;
    let blockUtils: BlockchainUtils;
    let databaseUtils: DatabaseUtils;
    let clearnodeSigner: SessionKeyStateSigner;

    beforeAll(async () => {
        blockUtils = new BlockchainUtils();
        databaseUtils = new DatabaseUtils();

        identity = new Identity(CONFIG.IDENTITIES[0].WALLET_PK, CONFIG.IDENTITIES[0].SESSION_PK);
        clearnodeSigner = new SessionKeyStateSigner(CLEARNODE_PRIVATE_KEY);

        ws = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
        await ws.connect();
        await createAuthSessionWithClearnode(ws, identity);
    });

    afterAll(async () => {
        ws.close();
        await databaseUtils.close();
    });

    describe('on-chain operations with Session Key as participant', () => {
        let client: TestNitroliteClient;
        let channelId: Hex;
        let channel: Channel;
        let currentState: State;
        let currentVersion: bigint;

        beforeAll(async () => {
            await blockUtils.makeSnapshot();

            // Create TestNitroliteClient with session key signer
            // This simulates how channels work with session keys
            client = new TestNitroliteClient(identity, identity.stateSKSigner);

            // Manually create a channel with session key as participant as Clearnode would return
            // a signature over a channel with Wallet, not a Session Key, as participant
            // This simulates a channel created before v0.5.0 where participant was a Session Key
            channel = {
                // specify user's Session Key, not Wallet
                participants: [identity.sessionKeyAddress, clearnodeSigner.getAddress()],
                adjudicator: CONFIG.ADDRESSES.DUMMY_ADJUDICATOR_ADDRESS,
                challenge: BigInt(CONFIG.DEFAULT_CHALLENGE_TIMEOUT),
                nonce: BigInt(Date.now()),
            };

            channelId = getChannelId(channel, chain.id);

            // Create initial state (unsigned)
            const unsignedInitialState: State = {
                intent: StateIntent.INITIALIZE,
                version: BigInt(0),
                data: '0x00' as Hex,
                allocations: [
                    {
                        destination: identity.sessionKeyAddress,
                        token: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
                        amount: BigInt(0),
                    },
                    {
                        destination: CONFIG.ADDRESSES.CLEARNODE_ADDRESS,
                        token: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
                        amount: BigInt(0),
                    },
                ],
                sigs: [],
            };

            const userSignature = await identity.stateSKSigner.signState(channelId, unsignedInitialState);
            const clearnodeSignature = await clearnodeSigner.signState(channelId, unsignedInitialState);

            const initialState: State = {
                ...unsignedInitialState,
                sigs: [userSignature, clearnodeSignature],
            };

            const { channelId: createdChannelId, initialState: createdState } = await client.depositAndCreateChannel(
                CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
                depositAmount,
                {
                    channel,
                    unsignedInitialState: initialState,
                    serverSignature: clearnodeSignature,
                }
            );

            expect(createdChannelId).toBe(channelId);
            currentState = createdState;
            currentVersion = createdState.version;

            const channelData = await client.getChannelData(channelId);
            expect(channelData.channel.participants[0]).toBe(identity.sessionKeyAddress);
            expect(channelData.status).toBe(ChannelStatus.ACTIVE);
        });

        afterAll(async () => {
            await databaseUtils.resetClearnodeState();
            await blockUtils.resetSnapshot();
        });

        it('should resize the channel using Clearnode RPC', async () => {
            const { state } = await client.resizeChannelAndWait(
                ws,
                channelId,
                currentState,
                identity.walletAddress,
                resizeAmount
            );

            const channelData = await client.getChannelData(channelId);
            expect(channelData.lastValidState.version).toBe(BigInt(state.version));

            currentState = state;
            currentVersion = BigInt(state.version);

            const accountBalance = await client.getAccountBalance(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
            expect(accountBalance).toBe(resizeAmount);
        });

        it('should challenge the channel with a manually signed state', async () => {
            const challengeVersion = currentVersion + BigInt(1);
            const unsignedChallengeState: State = {
                intent: StateIntent.OPERATE,
                version: challengeVersion,
                data: '0x01' as Hex,
                allocations: currentState.allocations,
                sigs: [],
            };

            const userSignature = await identity.stateSKSigner.signState(channelId, unsignedChallengeState);
            const clearnodeSignature = await clearnodeSigner.signState(channelId, unsignedChallengeState);

            const signedChallengeState: State = {
                ...unsignedChallengeState,
                sigs: [userSignature, clearnodeSignature],
            };

            const txHash = await client.challengeChannel({
                channelId,
                candidateState: signedChallengeState,
                proofStates: [],
            });

            expect(txHash).toBeDefined();

            const receipt = await blockUtils.waitForTransaction(txHash);
            expect(receipt).toBeDefined();

            const channelData = await client.getChannelData(channelId);
            expect(channelData.lastValidState.version).toBe(challengeVersion);

            currentState = signedChallengeState;
            currentVersion = challengeVersion;
        });

        it('should checkpoint the channel with a manually signed state', async () => {
            const checkpointVersion = currentVersion + BigInt(1);
            const unsignedCheckpointState: State = {
                intent: StateIntent.OPERATE,
                version: checkpointVersion,
                data: '0x02' as Hex,
                allocations: currentState.allocations,
                sigs: [],
            };

            const userSignature = await identity.stateSKSigner.signState(channelId, unsignedCheckpointState);
            const clearnodeSignature = await clearnodeSigner.signState(channelId, unsignedCheckpointState);

            const signedCheckpointState: State = {
                ...unsignedCheckpointState,
                sigs: [userSignature, clearnodeSignature],
            };

            const txHash = await client.checkpointChannel({
                channelId,
                candidateState: signedCheckpointState,
            });

            expect(txHash).toBeDefined();

            const receipt = await blockUtils.waitForTransaction(txHash);
            expect(receipt).toBeDefined();

            const channelData = await client.getChannelData(channelId);
            expect(channelData.lastValidState.version).toBe(checkpointVersion);

            currentState = signedCheckpointState;
            currentVersion = checkpointVersion;
        });

        it('should close the channel using Clearnode RPC', async () => {
            const finalVersion = currentVersion + BigInt(1);
            const unsignedFinalState: State = {
                intent: StateIntent.FINALIZE,
                version: finalVersion,
                data: '0x02' as Hex,
                allocations: currentState.allocations,
                sigs: [],
            };

            // NOTE: we do not call the Clearnode as it will reject the operation with "participant * has challenged channels,
            // cannot execute operation". After a corresponding event handler is added to the Clearnode, it can be called directly
            const clearnodeSignature = await clearnodeSigner.signState(channelId, unsignedFinalState);

            const txHash = await client.closeChannel({
                finalState: {
                    ...unsignedFinalState,
                    channelId,
                    serverSignature: clearnodeSignature,
                },
                stateData: unsignedFinalState.data,
            });

            expect(txHash).toBeDefined();

            const receipt = await blockUtils.waitForTransaction(txHash);
            expect(receipt).toBeDefined();

            const channelData = await client.getChannelData(channelId);
            expect(channelData.status).toBe(ChannelStatus.VOID); // channel should have been deleted after close

            // Withdraw funds
            const withdrawalTxHash = await client.withdrawal(
                CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
                toRaw(BigInt(500))
            );

            expect(withdrawalTxHash).toBeDefined();

            const withdrawalReceipt = await blockUtils.waitForTransaction(withdrawalTxHash);
            expect(withdrawalReceipt).toBeDefined();

            // Verify account balance is now zero
            const finalAccountBalance = await client.getAccountBalance(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
            expect(finalAccountBalance).toBe(BigInt(0));
        });
    });

    describe('on-chain operations with Wallet as participant', () => {
        let client: TestNitroliteClient;
        let channelId: Hex;
        let currentState: State;
        let currentVersion: bigint;

        beforeAll(async () => {
            await blockUtils.makeSnapshot();

            // Create TestNitroliteClient still with Session Key signer
            // This is the way developers will use the NitroliteClient to interact with channels created in previous versions
            client = new TestNitroliteClient(identity, identity.stateSKSigner);

            // Request channel creation from Clearnode
            const msg = await createCreateChannelMessage(identity.messageWalletSigner, {
                chain_id: chain.id,
                token: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            });

            const createResponse = await ws.sendAndWaitForResponse(msg, getCreateChannelPredicate(), 5000);
            expect(createResponse).toBeDefined();

            const { params: createParsedResponseParams } = parseCreateChannelResponse(createResponse);

            // Use depositAndCreateChannel with the channel struct from Clearnode
            const { channelId: createdChannelId, initialState } = await client.depositAndCreateChannel(
                CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
                depositAmount,
                {
                    unsignedInitialState: convertRPCToClientState(
                        createParsedResponseParams.state,
                        createParsedResponseParams.serverSignature
                    ),
                    channel: convertRPCToClientChannel(createParsedResponseParams.channel),
                    serverSignature: createParsedResponseParams.serverSignature,
                }
            );

            channelId = createdChannelId;
            currentState = initialState;
            currentVersion = initialState.version;

            const channelData = await client.getChannelData(channelId);
            expect(channelData.channel.participants[0]).toBe(identity.walletAddress);
            expect(channelData.status).toBe(ChannelStatus.ACTIVE);
        });

        afterAll(async () => {
            await databaseUtils.resetClearnodeState();
            await blockUtils.resetSnapshot();
        });

        it('should resize the channel using Clearnode RPC', async () => {
            const resizeAmount = toRaw(BigInt(500)); // Resize by 500 USDC
            const { state } = await client.resizeChannelAndWait(
                ws,
                channelId,
                currentState,
                identity.walletAddress,
                resizeAmount
            );

            const channelData = await client.getChannelData(channelId);
            expect(channelData.lastValidState.version).toBe(BigInt(state.version));

            currentState = state;
            currentVersion = BigInt(state.version);

            const accountBalance = await client.getAccountBalance(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
            expect(accountBalance).toBe(resizeAmount);
        });

        it('should challenge the channel with a manually signed state', async () => {
            const challengeVersion = currentVersion + BigInt(1);
            const unsignedChallengeState: State = {
                intent: StateIntent.OPERATE,
                version: challengeVersion,
                data: '0x01' as Hex,
                allocations: currentState.allocations,
                sigs: [],
            };

            const userSignature = await identity.stateWalletSigner.signState(channelId, unsignedChallengeState);
            const clearnodeSignature = await clearnodeSigner.signState(channelId, unsignedChallengeState);

            const signedChallengeState: State = {
                ...unsignedChallengeState,
                sigs: [userSignature, clearnodeSignature],
            };

            const txHash = await client.challengeChannel({
                channelId,
                candidateState: signedChallengeState,
                proofStates: [],
            });

            expect(txHash).toBeDefined();

            const receipt = await blockUtils.waitForTransaction(txHash);
            expect(receipt).toBeDefined();

            const channelData = await client.getChannelData(channelId);
            expect(channelData.lastValidState.version).toBe(challengeVersion);

            currentState = signedChallengeState;
            currentVersion = challengeVersion;
        });

        it('should checkpoint the channel with a manually signed state', async () => {
            const checkpointVersion = currentVersion + BigInt(1);
            const unsignedCheckpointState: State = {
                intent: StateIntent.OPERATE,
                version: checkpointVersion,
                data: '0x02' as Hex,
                allocations: currentState.allocations,
                sigs: [],
            };

            const userSignature = await identity.stateWalletSigner.signState(channelId, unsignedCheckpointState);
            const clearnodeSignature = await clearnodeSigner.signState(channelId, unsignedCheckpointState);

            const signedCheckpointState: State = {
                ...unsignedCheckpointState,
                sigs: [userSignature, clearnodeSignature],
            };

            const txHash = await client.checkpointChannel({
                channelId,
                candidateState: signedCheckpointState,
            });

            expect(txHash).toBeDefined();

            const receipt = await blockUtils.waitForTransaction(txHash);
            expect(receipt).toBeDefined();

            const channelData = await client.getChannelData(channelId);
            expect(channelData.lastValidState.version).toBe(checkpointVersion);

            currentState = signedCheckpointState;
            currentVersion = checkpointVersion;
        });

        it('should close the channel using Clearnode RPC', async () => {
            const finalVersion = currentVersion + BigInt(1);
            const unsignedFinalState: State = {
                intent: StateIntent.FINALIZE,
                version: finalVersion,
                data: '0x02' as Hex,
                allocations: currentState.allocations,
                sigs: [],
            };

            // NOTE: we do not call the Clearnode as it will reject the operation with "participant * has challenged channels,
            // cannot execute operation". After a corresponding event handler is added to the Clearnode, it can be called directly
            const clearnodeSignature = await clearnodeSigner.signState(channelId, unsignedFinalState);

            const txHash = await client.closeChannel({
                finalState: {
                    ...unsignedFinalState,
                    channelId,
                    serverSignature: clearnodeSignature,
                },
                stateData: unsignedFinalState.data,
            });

            expect(txHash).toBeDefined();

            const receipt = await blockUtils.waitForTransaction(txHash);
            expect(receipt).toBeDefined();

            const channelData = await client.getChannelData(channelId);
            expect(channelData.status).toBe(ChannelStatus.VOID); // channel should have been deleted after close

            // Withdraw funds
            const withdrawalTxHash = await client.withdrawal(
                CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
                toRaw(BigInt(500))
            );

            expect(withdrawalTxHash).toBeDefined();

            const withdrawalReceipt = await blockUtils.waitForTransaction(withdrawalTxHash);
            expect(withdrawalReceipt).toBeDefined();

            // Verify account balance is now zero
            const finalAccountBalance = await client.getAccountBalance(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);
            expect(finalAccountBalance).toBe(BigInt(0));
        });
    });
});
