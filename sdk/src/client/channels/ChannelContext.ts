import { Address, Hex } from 'viem';
import { getChannelId } from '../../utils';
import type { NitroliteClient } from '../NitroliteClient';
import { AppLogic } from '../../types';
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
        channel: Channel,
        appLogic: AppLogic<T>,
        signerAttorneyAddress?: Address
    ) {
        this.client = client;
        this.channel = channel;
        this.channelId = getChannelId(channel);
        this.appLogic = appLogic;

        // If there is no signer attorney, use the client's address
        let participantAddress = (signerAttorneyAddress ||
            client.account?.address) as Address | undefined;
        if (!participantAddress) {
            throw new Error('Channel participant is not provided');
        }

        if (participantAddress === channel.participants[0]) {
            this.role = Role.HOST;
        } else if (participantAddress === channel.participants[1]) {
            this.role = Role.GUEST;
        } else {
            throw new Error('Account is not a participant in this channel');
        }
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
    private async createChannelState(
        appState: T,
        tokenAddress: Address,
        amounts: [bigint, bigint]
    ): Promise<State> {
        // Encode the app state
        const data = this.appLogic.encode(appState);

        // Create allocations
        const allocations: [Allocation, Allocation] = [
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

        // Create state (without signatures)
        return {
            data,
            allocations,
            sigs: [],
        };
    }

    /**
     * Open a channel with initial funding
     */
    async open(
        appState: T,
        tokenAddress: Address,
        amounts: [bigint, bigint]
    ): Promise<ChannelId> {
        // Create initial state
        const initialState = await this.createChannelState(
            appState,
            tokenAddress,
            amounts
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

    async deposit(
        tokenAddress: Address,
        amount: bigint
    ): Promise<void> {
        return this.client.deposit(tokenAddress, amount);
    }

    async withdraw(
        tokenAddress: Address,
        amount: bigint
    ): Promise<void> {
        return this.client.withdraw(tokenAddress, amount);
    }

    async getAvailableBalance(
        tokenAddress: Address
    ): Promise<bigint> {
        if (!this.client.account?.address) {
            throw new Error('Account address is not provided');
        }

        return this.client.getAvailableBalance(this.client.account.address, tokenAddress);
    }

    async getAccountChannels(
        tokenAddress: Address
    ): Promise<ChannelId[]> {
        if (!this.client.account?.address) {
            throw new Error('Account address is not provided');
        }

        return this.client.getAccountChannels(this.client.account.address, tokenAddress);
    }

    /**
     * Reclaim funds after challenge period expires
     */
    async reclaim(): Promise<void> {
        return this.client.reclaimChannel(this.channelId);
    }
}
