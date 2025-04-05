import { Address, encodeAbiParameters, Hex, keccak256 } from "viem";
import { getChannelId } from "../../utils";
import type { NitroliteClient } from "../NitroliteClient";
import { AppLogic, Signature } from "../../types";
import { ChannelId, State, Role, Allocation, Channel } from "../types";

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
     * Create a new channel context a
     * @param client client to use
     * @param channel channel to create or join
     * @param initialState initial provided state
     * @param appLogic application logic to use
     */
    constructor(client: NitroliteClient, channel: Channel, initialState: State, appLogic: AppLogic<T>) {
        this.client = client;
        this.channel = channel;
        this.channelId = getChannelId(channel);
        this.appLogic = appLogic;
        this.states.push(initialState);

        this.role = Role.UNDEFINED;
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
        return this.channel.participants[this.role === Role.CREATOR ? Role.GUEST : Role.CREATOR];
    }

    /**
     * Create a channel state based on application state
     */
    createChannelState(appState: T, tokenAddress: Address, amounts: [bigint, bigint], signatures: Signature[] = []): State {
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
    async create(): Promise<void> {
        const initialState = this.getCurrentState();

        if (!initialState) {
            throw new Error("No initial state to create channel");
        }

        await this.client.createChannel(this.channel, initialState);

        this.role = Role.CREATOR;
    }

    /**
     * Join an existing channel
     */
    async join(): Promise<void> {
        const initialState = this.getCurrentState();

        if (!initialState) {
            throw new Error("No initial state to join channel");
        }

        // Assuming, that the channel is already created and consists of two participants,
        // in this case the id of paticipant 0 is the creator and participant 1 is the guest
        await this.client.joinChannel(this.channelId, 1, initialState.sigs[1]);

        // Set the role
        this.role = Role.GUEST;
    }

    /**
     * Append the application state
     */
    appendAppState(newAppState: T, tokenAddress: Address, amounts: [bigint, bigint], signatures: Signature[] = []): State {
        const currentState = this.getCurrentState();
        const currentAppState = this.getCurrentAppState();

        if (!currentAppState || !currentState) {
            throw new Error("No current app state to update, open channel first");
        }

        // Validate state transition if the app logic provides a validator
        if (this.appLogic.validateTransition) {
            const isValid = this.appLogic.validateTransition(this.channel, currentAppState, newAppState);

            if (!isValid) {
                throw new Error("Invalid state transition");
            }
        }

        // Create new state with existing allocations
        const newState: State = this.createChannelState(newAppState, tokenAddress, amounts, signatures);

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
     * Close the channel with the provided state
     */
    async close(newAppState: T, tokenAddress: Address, amounts: [bigint, bigint], signatures: Signature[] = []): Promise<void> {
        if (!this.appLogic.isFinal || !this.appLogic.isFinal(newAppState)) {
            throw new Error("Provided state is not final");
        }

        const finalState = this.createChannelState(newAppState, tokenAddress, amounts, signatures);

        let proofs: State[] = [];
        if (this.appLogic.provideProofs) {
            proofs = this.appLogic.provideProofs(this.channel, newAppState, this.states) || [];
        }

        this.states.push(finalState);

        return this.client.closeChannel(this.channelId, finalState, proofs);
    }

    /**
     * Challenge the channel with the current state
     */
    async challenge(): Promise<void> {
        const currentState = this.getCurrentState();
        const currentAppState = this.getCurrentAppState();

        if (!currentState || !currentAppState) {
            throw new Error("No current state to challenge with");
        }

        let proofs: State[] = [];
        if (this.appLogic.provideProofs) {
            proofs = this.appLogic.provideProofs(this.channel, currentAppState, this.states) || [];
        }

        return this.client.challengeChannel(this.channelId, currentState, proofs);
    }

    /**
     * Checkpoint the current state
     */
    async checkpoint(): Promise<void> {
        const currentState = this.getCurrentState();
        const currentAppState = this.getCurrentAppState();

        if (!currentState || !currentAppState) {
            throw new Error("No current state to checkpoint");
        }

        let proofs: State[] = [];
        if (this.appLogic.provideProofs) {
            proofs = this.appLogic.provideProofs(this.channel, currentAppState, this.states) || [];
        }

        return this.client.checkpointChannel(this.channelId, currentState, proofs);
    }

    async deposit(tokenAddress: Address, amount: bigint): Promise<void> {
        return this.client.deposit(tokenAddress, amount);
    }

    async withdraw(tokenAddress: Address, amount: bigint): Promise<void> {
        return this.client.withdraw(tokenAddress, amount);
    }

    async getAvailableBalance(tokenAddress: Address): Promise<bigint> {
        if (!this.client.account?.address) {
            throw new Error("Account address is not provided");
        }

        return this.client.getAvailableBalance(this.client.account.address, tokenAddress);
    }

    async getAccountChannels(tokenAddress: Address): Promise<ChannelId[]> {
        if (!this.client.account?.address) {
            throw new Error("Account address is not provided");
        }

        return this.client.getAccountChannels(this.client.account.address, tokenAddress);
    }

    getStateHash(state: State): ChannelId {
        const encoded = encodeAbiParameters(
            [
                { type: "bytes32" },
                { type: "bytes" },
                {
                    type: "tuple[]",
                    components: [
                        { name: "destination", type: "address" },
                        { name: "token", type: "address" },
                        { name: "amount", type: "uint256" },
                    ],
                },
            ],
            [this.channelId, state.data, state.allocations]
        );

        const stateHash = keccak256(encoded);

        return stateHash;
    }

    private getAllocations(tokenAddress: Address, amounts: [bigint, bigint]): [Allocation, Allocation] {
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
