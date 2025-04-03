import { Address, encodeAbiParameters, Hex, keccak256 } from 'viem';
import { getChannelId } from '../../utils';
import type { NitroliteClient } from '../NitroliteClient';
import { AppLogic, Signature } from '../../types';
import { ChannelId, State, Role, Allocation, Channel } from '../types';

/**
 * Channel context for managing application state
 */
export class ChannelContext<T = unknown> {
    readonly channel: Channel;
    readonly channelId: ChannelId;
    readonly client: NitroliteClient;
    readonly appLogic: AppLogic<T>;

    private states: State[] = [];
    private role: Role;

    /**
     * Create a new channel context
     */
    constructor(
        client: NitroliteClient,
        guest: Address,
        appLogic: AppLogic<T>,
        signerAttorneyAddress?: Address
    ) {
        this.client = client;
        this.channel = {
            participants: [client.account?.address as Address, guest],
            adjudicator: appLogic.getAdjudicatorAddress(),
            challenge: 100500n, // TODO:
            nonce: 0n, // TODO:
        } as Channel;
        this.channelId = getChannelId(this.channel);
        this.appLogic = appLogic;

        // If there is no signer attorney, use the client's address
        let participantAddress = (signerAttorneyAddress ||
            client.account?.address) as Address | undefined;
        if (!participantAddress) {
            throw new Error('Channel participant is not provided');
        }

        // TODO:
        // if (participantAddress === channel.participants[0]) {
        this.role = Role.HOST;
        // } else if (participantAddress === channel.participants[1]) {
        // this.role = Role.GUEST;
        // } else {
        // throw new Error('Account is not a participant in this channel');
        // }
    }

    /**
     * Get the channel configuration
     */
    getChannel(): Channel {
        return this.channel;
    }

    /**
     * Get the channel ID
     */
    getChannelId(): ChannelId {
        return this.channelId;
    }

    /**
     * Get the current state
     */
    getCurrentState(): State | undefined {
        return this.states[this.states.length - 1];
    }

    /**
     * Get the current application state
     */
    getCurrentAppState(): T | undefined {
        const currentState = this.getCurrentState();

        return currentState?.data && this.appLogic.decode(currentState.data);
    }

    /**
     * Get the role of the current account
     */
    getRole(): Role {
        return this.role;
    }

    /**
     * Get the other participant's address
     */
    getOtherParticipant(): Address {
        return this.channel.participants[
            this.role === Role.HOST ? Role.GUEST : Role.HOST
        ];
    }

    /**
     * Create a channel state based on application state
     */
    createChannelState(
        appState: T,
        tokenAddress: Address,
        amounts: [bigint, bigint],
        signatures: Signature[] = []
    ): State {
        // Encode the app state
        const data = this.appLogic.encode(appState);

        // Create allocations
        const allocations = this.getAllocations(tokenAddress, amounts);

        return {
            data,
            allocations,
            sigs: signatures,
        };
    }

    /**
     * Open a channel with initial funding
     */
    async open(
        appState: T,
        tokenAddress: Address,
        amounts: [bigint, bigint],
        signatures: Signature[] = []
    ): Promise<ChannelId> {
        // Create initial state
        const initialState = await this.createChannelState(
            appState,
            tokenAddress,
            amounts,
            signatures
        );

        // Save as current state
        this.states.push(initialState);

        // Open the channel
        return this.client.openChannel(this.channel, initialState, this.role);
    }

    /**
     * Append the application state
     */
    appendAppState(newAppState: T): State {
        const currentState = this.getCurrentState();
        const currentAppState = this.getCurrentAppState();

        if (!currentAppState || !currentState) {
            throw new Error(
                'No current app state to update, open channel first'
            );
        }

        // Validate state transition if the app logic provides a validator
        if (this.appLogic.validateTransition) {
            const isValid = this.appLogic.validateTransition(
                this.channel,
                currentAppState,
                newAppState
            );

            if (!isValid) {
                throw new Error('Invalid state transition');
            }
        }

        // Encode the app state
        const data = this.appLogic.encode(newAppState);

        // Create new state with existing allocations
        const newState: State = {
            data,
            allocations: currentState.allocations,
            sigs: [], // Will be filled when signing
        };

        // Append the channel state
        this.states.push(newState);

        return newState;
    }

    /**
     * Check if the current state is final
     */
    isFinal(): boolean {
        const currentAppState = this.getCurrentAppState();
        if (!currentAppState || !this.appLogic.isFinal) {
            return false;
        }

        return this.appLogic.isFinal(currentAppState);
    }

    /**
     * Close the channel with the current state
     */
    async close(): Promise<void> {
        const currentState = this.getCurrentState();
        const currentAppState = this.getCurrentAppState();

        if (!this.isFinal() || !currentState || !currentAppState) {
            throw new Error('No current state to close with');
        }

        let proofs: State[] = [];
        if (this.appLogic.provideProofs) {
            proofs =
                this.appLogic.provideProofs(
                    this.channel,
                    currentAppState,
                    this.states
                ) || [];
        }

        return this.client.closeChannel(this.channelId, currentState, proofs);
    }

    /**
     * Challenge the channel with the current state
     */
    async challenge(): Promise<void> {
        const currentState = this.getCurrentState();
        const currentAppState = this.getCurrentAppState();

        if (!currentState || !currentAppState) {
            throw new Error('No current state to challenge with');
        }

        let proofs: State[] = [];
        if (this.appLogic.provideProofs) {
            proofs =
                this.appLogic.provideProofs(
                    this.channel,
                    currentAppState,
                    this.states
                ) || [];
        }

        return this.client.challengeChannel(
            this.channelId,
            currentState,
            proofs
        );
    }

    /**
     * Checkpoint the current state
     */
    async checkpoint(): Promise<void> {
        const currentState = this.getCurrentState();
        const currentAppState = this.getCurrentAppState();

        if (!currentState || !currentAppState) {
            throw new Error('No current state to checkpoint');
        }

        let proofs: State[] = [];
        if (this.appLogic.provideProofs) {
            proofs =
                this.appLogic.provideProofs(
                    this.channel,
                    currentAppState,
                    this.states
                ) || [];
        }

        return this.client.checkpointChannel(
            this.channelId,
            currentState,
            proofs
        );
    }

    async deposit(tokenAddress: Address, amount: bigint): Promise<void> {
        return this.client.deposit(tokenAddress, amount);
    }

    async withdraw(tokenAddress: Address, amount: bigint): Promise<void> {
        return this.client.withdraw(tokenAddress, amount);
    }

    async getAvailableBalance(tokenAddress: Address): Promise<bigint> {
        if (!this.client.account?.address) {
            throw new Error('Account address is not provided');
        }

        return this.client.getAvailableBalance(
            this.client.account.address,
            tokenAddress
        );
    }

    async getAccountChannels(tokenAddress: Address): Promise<ChannelId[]> {
        if (!this.client.account?.address) {
            throw new Error('Account address is not provided');
        }

        return this.client.getAccountChannels(
            this.client.account.address,
            tokenAddress
        );
    }

    /**
     * Reclaim funds after challenge period expires
     */
    async reclaim(): Promise<void> {
        return this.client.reclaimChannel(this.channelId);
    }

    getStateHash(
        appState: T,
        tokenAddress: Address,
        amounts: [bigint, bigint]
    ): ChannelId {

        // return '0xa552b3021984e63e48bc2ecc66a088d3e67abfe53bfc320fe0ab7063a4e6c235';
        
        const data = this.appLogic.encode(appState);
        const allocations = this.getAllocations(tokenAddress, amounts);
        
        console.log("state hash", this.channelId, data, allocations);

        const encoded = encodeAbiParameters(
            [
                { type: 'bytes32' },
                { type: 'bytes' },
                {
                    type: 'tuple[2]',
                    components: [
                        { name: 'destination', type: 'address' },
                        { name: 'token', type: 'address' },
                        { name: 'amount', type: 'uint256' },
                    ],
                },
            ],
            [this.channelId, data, allocations]
        );

        const stateHash = keccak256(encoded);

        console.log("state has", stateHash);

        return stateHash;
    }

    private getAllocations(
        tokenAddress: Address,
        amounts: [bigint, bigint]
    ): [Allocation, Allocation] {
        return [
            {
                destination: this.channel.participants[0],
                token: tokenAddress,
                amount: amounts[0],
            },
            {
                destination: this.channel.participants[1],
                token: tokenAddress,
                amount: amounts[1],
            },
        ];
    }
}
