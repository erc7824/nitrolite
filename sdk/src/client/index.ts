import { Account, Hex, PublicClient, WalletClient, Chain, Transport, ParseAccount, Hash, zeroAddress } from "viem";

import { NitroliteService, Erc20Service } from "./services";
import {
    Channel,
    State,
    ChannelId,
    NitroliteClientConfig,
    CreateChannelParams,
    CheckpointChannelParams,
    ChallengeChannelParams,
    CloseChannelParams,
    AccountInfo,
} from "./types";
import { getStateHash, generateChannelNonce, getChannelId, encoders, removeQuotesFromRS, signState } from "../utils";
import * as Errors from "../errors";
import { ContractAddresses } from "../abis";
import { MAGIC_NUMBERS } from "../config";

/**
 * The main client class for interacting with the Nitrolite SDK.
 * Provides high-level methods for managing state channels and funds.
 */
export class NitroliteClient {
    public readonly publicClient: PublicClient;
    public readonly walletClient: WalletClient<Transport, Chain, ParseAccount<Account>>;
    public readonly addresses: ContractAddresses;
    public readonly challengeDuration: bigint;
    private readonly nitroliteService: NitroliteService;
    private readonly erc20Service: Erc20Service;
    private readonly account: ParseAccount<Account>;

    constructor(config: NitroliteClientConfig) {
        if (!config.publicClient) throw new Errors.MissingParameterError("publicClient");
        if (!config.walletClient) throw new Errors.MissingParameterError("walletClient");
        if (!config.walletClient.account) throw new Errors.MissingParameterError("walletClient.account");
        if (!config.addresses?.custody) throw new Errors.MissingParameterError("addresses.custody");
        if (!config.addresses?.adjudicators) throw new Errors.MissingParameterError("addresses.adjudicators");
        if (!config.addresses?.guestAddress) throw new Errors.MissingParameterError("addresses.guestAddress");
        if (!config.addresses?.tokenAddress) throw new Errors.MissingParameterError("addresses.tokenAddress");

        this.publicClient = config.publicClient;
        this.walletClient = config.walletClient;
        this.account = config.walletClient.account;
        this.addresses = config.addresses;
        this.challengeDuration = config.challengeDuration ?? 0n;

        this.nitroliteService = new NitroliteService(this.publicClient, this.addresses, this.walletClient, this.account);
        this.erc20Service = new Erc20Service(this.publicClient, this.walletClient);
    }

    /**
     * Deposits tokens or ETH into the custody contract.
     * Handles ERC20 approval if necessary.
     * @param amount The amount of tokens/ETH to deposit.
     * @returns The transaction hash.
     */
    async deposit(amount: bigint): Promise<Hash> {
        const owner = this.account.address;
        const spender = this.addresses.custody;
        const tokenAddress = this.addresses.tokenAddress;

        if (tokenAddress !== zeroAddress) {
            const allowance = await this.erc20Service.getTokenAllowance(tokenAddress, owner, spender);
            if (allowance < amount) {
                try {
                    await this.erc20Service.approve(tokenAddress, spender, amount);
                } catch (err) {
                    const error = new Errors.TokenError("Failed to approve tokens for deposit");
                    throw error;
                }
            }
        }

        try {
            return await this.nitroliteService.deposit(tokenAddress, amount);
        } catch (err) {
            throw new Errors.ContractError("Failed to execute deposit on contract", undefined, undefined, undefined, undefined, err as Error);
        }
    }

    /**
     * Creates a new state channel on-chain.
     * Constructs the initial state, signs it, and calls the custody contract.
     * @param params Parameters for channel creation.
     * @returns The channel ID, the signed initial state, and the transaction hash.
     */
    async createChannel(params: CreateChannelParams): Promise<{ channelId: ChannelId; initialState: State; txHash: Hash }> {
        const { initialAllocationAmounts, stateData } = params;

        const channelNonce = generateChannelNonce();
        const participants: [Hex, Hex] = [this.account.address, this.addresses.guestAddress];
        const tokenAddress = this.addresses.tokenAddress;
        const adjudicatorAddress = this.addresses.adjudicators["default"];
        const challengeDuration = this.challengeDuration;

        if (!participants || participants.length !== 2) {
            throw new Errors.InvalidParameterError("Channel must have two participants.");
        }

        const channel: Channel = {
            participants,
            adjudicator: adjudicatorAddress,
            challenge: challengeDuration,
            nonce: channelNonce,
        };

        const initialAppData = stateData ?? encoders["numeric"](MAGIC_NUMBERS.OPEN);

        const channelId = getChannelId(channel);

        const initialState: State = {
            data: initialAppData,
            allocations: [
                { destination: participants[0], token: tokenAddress, amount: initialAllocationAmounts[0] },
                { destination: participants[1], token: tokenAddress, amount: initialAllocationAmounts[1] },
            ],
            sigs: [],
        };

        const stateHash = getStateHash(channelId, initialState);
        const accountSignature = await signState(stateHash, this.walletClient.signMessage);

        initialState.sigs = [accountSignature];

        try {
            const txHash = await this.nitroliteService.createChannel(channel, initialState);

            return { channelId, initialState, txHash };
        } catch (err) {
            throw new Errors.ContractError("Failed to execute createChannel on contract", undefined, undefined, undefined, undefined, err as Error);
        }
    }

    async depositAndCreateChannel(
        depositAmount: bigint,
        params: CreateChannelParams
    ): Promise<{ channelId: ChannelId; initialState: State; txHash: Hash }> {
        const depositTxHash = await this.deposit(depositAmount);
        const { channelId, initialState, txHash } = await this.createChannel(params);

        return { channelId, initialState, txHash: depositTxHash };
    }

    /**
     * Checkpoints a state on-chain.
     * Requires the state to be signed by both participants.
     * @param params Parameters for checkpointing the state.
     * @returns The transaction hash.
     */
    async checkpointChannel(params: CheckpointChannelParams): Promise<Hash> {
        const { channelId, candidateState, proofStates = [] } = params;

        if (!candidateState.sigs || candidateState.sigs.length < 2) {
            throw new Errors.InvalidParameterError("Candidate state for checkpoint must be signed by both participants.");
        }

        try {
            return await this.nitroliteService.checkpoint(channelId, candidateState, proofStates);
        } catch (err) {
            throw new Errors.ContractError(
                "Failed to execute checkpointChannel on contract",
                undefined,
                undefined,
                undefined,
                undefined,
                err as Error
            );
        }
    }

    /**
     * Challenges a channel on-chain with a candidate state.
     * Used when the counterparty is unresponsive. Requires the candidate state to be signed by the challenger.
     * @param params Parameters for challenging the channel.
     * @returns The transaction hash.
     */
    async challengeChannel(params: ChallengeChannelParams): Promise<Hash> {
        const { channelId, candidateState, proofStates = [] } = params;

        try {
            return await this.nitroliteService.challenge(channelId, candidateState, proofStates);
        } catch (err) {
            throw new Errors.ContractError(
                "Failed to execute challengeChannel on contract",
                undefined,
                undefined,
                undefined,
                undefined,
                err as Error
            );
        }
    }

    /**
     * Closes a channel on-chain using a mutually agreed final state.
     * Requires the final state signed by both participants.
     * @param params Parameters for closing the channel.
     * @returns The transaction hash.
     */
    async closeChannel(params: CloseChannelParams): Promise<Hash> {
        const { finalState } = params;
        const finalSignatures = removeQuotesFromRS(finalState.server_signature)["server_signature"];
        const appState = MAGIC_NUMBERS.CLOSE;

        const state: State = {
            data: encoders["numeric"](appState),
            allocations: finalState.allocations,
            sigs: [],
        };

        const stateHash = getStateHash(finalState.channel_id, state); // Pass channelId if required by util

        const accountSignature = await signState(stateHash, this.walletClient.signMessage);

        state.sigs = [accountSignature, ...finalSignatures];

        try {
            return await this.nitroliteService.close(finalState.channel_id, state);
        } catch (err) {
            throw new Errors.ContractError("Failed to execute closeChannel on contract", undefined, undefined, undefined, undefined, err as Error);
        }
    }

    /**
     * Withdraws tokens previously deposited into the custody contract.
     * This does not withdraw funds locked in active channels.
     * @param amount The amount of tokens/ETH to withdraw.
     * @returns The transaction hash.
     */
    async withdrawal(amount: bigint): Promise<Hash> {
        const tokenAddress = this.addresses.tokenAddress;

        try {
            return await this.nitroliteService.withdraw(tokenAddress, amount);
        } catch (err) {
            throw new Errors.ContractError("Failed to execute withdrawDeposit on contract", undefined, undefined, undefined, undefined, err as Error);
        }
    }

    /**
     * Retrieves a list of channel IDs associated with a specific account.
     * @returns An array of Channel IDs.
     */
    async getAccountChannels(): Promise<ChannelId[]> {
        try {
            return await this.nitroliteService.getAccountChannels(this.account.address);
        } catch (err) {
            throw err;
        }
    }

    /**
     * Retrieves deposit and lock information for an account regarding a specific token.
     * @returns Account info including available, locked amounts and channel count.
     */
    async getAccountInfo(): Promise<AccountInfo> {
        try {
            return await this.nitroliteService.getAccountInfo(this.account.address, this.addresses.tokenAddress);
        } catch (err) {
            throw err;
        }
    }

    /**
     * Approves the custody contract to spend a specified amount of an ERC20 token.
     * @returns The transaction hash.
     */
    async approveTokens(amount: bigint): Promise<Hash> {
        const spender = this.addresses.custody;
        const tokenAddress = this.addresses.tokenAddress;

        try {
            return await this.erc20Service.approve(tokenAddress, spender, amount);
        } catch (err) {
            throw new Errors.TokenError("Failed to approve tokens", undefined, undefined, undefined, undefined, err as Error);
        }
    }

    /**
     * Gets the current allowance granted by an owner to a spender for a specific ERC20 token.
     * @returns The allowance amount as a bigint.
     */
    async getTokenAllowance(): Promise<bigint> {
        const tokenAddress = this.addresses.tokenAddress;
        const targetOwner = this.account.address;
        const targetSpender = this.addresses.custody;

        try {
            return await this.erc20Service.getTokenAllowance(tokenAddress, targetOwner, targetSpender);
        } catch (err) {
            throw new Errors.TokenError("Failed to get token allowance", undefined, undefined, undefined, undefined, err as Error);
        }
    }

    /**
     * Gets the balance of a specific ERC20 token for an account.
     * @returns The token balance as a bigint.
     */
    async getTokenBalance(): Promise<bigint> {
        const tokenAddress = this.addresses.tokenAddress;
        const targetAccount = this.account.address;
        try {
            return await this.erc20Service.getTokenBalance(tokenAddress, targetAccount);
        } catch (err) {
            throw new Errors.TokenError("Failed to get token balance", undefined, undefined, undefined, undefined, err as Error);
        }
    }
}
