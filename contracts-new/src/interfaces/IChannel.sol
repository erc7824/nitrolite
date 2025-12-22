// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IVault} from "./IVault.sol";

contract ChannelsHub is IVault {
    // User and Node have different logic in the channel
    // only 2 participants in the channel

    // Funds:
    // When depositing to or withdrawing from, source of User funds is ERC20 contract, and not the Vault
    // Node's funds do not leave the Contract

    // ========= Channel Types ==========

    struct Definition {
        uint32 challengeDuration;
        address participant;
        address node;
        uint256 nonce;
        // to be added later:
        // address executionModule;
        // address signatureValidator;
    }

    enum StateIntent {
        OPERATE,
        CREATE,
        CLOSE,
        DEPOSIT,
        WITHDRAW,
        MIGRATE_HOME,
        LOCK,
        UNLOCK
    }

    struct CrossChainState {
        uint256 version;

        uint64 homeChainId;
        StateIntent intent;

        // to be added for custom executors, e.g. fees, later:
        // bytes executionData;

        State homeState;
        State nonHomeState;

        bytes participantSig;
        bytes nodeSig;
    }

    struct State {
        uint64 chainId;
        address token;

        Balance participantBalance;
        Balance nodeBalance;
    }

    struct Balance {
        uint256 allocation;
        int256 netFlow;
    }

    // ======== Contract Storage ==========

    struct Metadata {
        Definition definition;
        CrossChainState lastState;
        uint256 participantLockedFunds;
        uint256 nodeLockedFunds;
        uint64 challengeExpiry; // timestamp when challenge period ends
    }

    // ========== Getters ==========

    // *** IVault ***

    function getAccountsBalances(address[] calldata accounts, address[] calldata tokens)
        external
        view
        returns (uint256[][] memory) {}

    // ******

    function getNodeBalance(address node, address token) external view returns (uint256) {
        // node's deposit + escrowDepositUnlocked funds + checkpointed Node's net flow (receive - send)
    }

    function getOpenChannels(address participant) external view returns (bytes32[] memory) {
        // return list of open channelIds between participant and node
    }

    function getChannelData(bytes32 channelId)
        external
        view
        returns (
            Definition memory definition,
            CrossChainState memory lastState,
            uint256 challengeExpiry,
            uint256 participantLockedFunds
        ) {
        // return channel Metadata
    }

    function getEscrowDepositData(bytes32 escrowId)
        external
        view
        returns (
            Definition memory definition,
            CrossChainState memory lastState,
            uint256 participantLockedFunds,
            uint64 unlockExpiry,
            uint64 challengeExpiry
        ) {
        // return escrow deposit Metadata
    }

    function getEscrowWithdrawalData(bytes32 escrowId)
        external
        view
        returns (
            Definition memory definition,
            CrossChainState memory lastState,
            uint256 participantLockedFunds
        ) {
        // return escrow withdrawal Metadata
    }

    // *** IVault ***

    function deposit(address account, address token, uint256 amount) external payable {}

    function withdraw(address account, address token, uint256 amount) external {}

    // ******

    // ========== Channel lifecycle ==========

    // usage:
    // - open a new channel with User deposit funds ("entry to Clearnet")
    // - open a new channel with Node transfer funds to User (User joins Clearnet after already receiving funds)
    // - open and close channel in one go for atomic for User receiving and withdrawing funds
    function create(Definition calldata def, CrossChainState calldata initCCS) external payable {
        // -- checks --
        // require(initCCS.homeChainId == block.chainid)
        // require(node balance == 0)
        // validate signatures over (channelId, CrossChainState)

        // -- effects --
        // store channel definition
        // store initial state
        // add user funding amount to "channel locked" amount

        // -- interactions --
        // pull funds from participant
    }

    event ChannelCreated(bytes32 indexed channelId, address indexed participant, address indexed node, Definition definition, CrossChainState initialState);

    // usage:
    // - "open" channel on new chain in case of chain migration
    function migrateChannelHere(Definition calldata def, CrossChainState calldata candidate, CrossChainState[] calldata proof) external payable {
        // -- checks --
        // require(candidate.homeChainId == block.chainid)
        // validate signatures over (channelId, CrossChainState)
        // validate proof:

        // -- effects --
        // store channel definition
        // store initial state
        // move `nodeLockedFunds` to `participantLockedFunds` in Metadata

        // -- interactions --
        // pull funds from participant
    }

    event ChannelMigrated(bytes32 indexed channelId, Definition definition, CrossChainState initialState);

    // usage:
    // - "close" channel on old chain in case of chain migration (basically, a "checkpoint" in a common sense)
    // - acknowledge latest funds change on Home chain
    // - resolve a challenge

    // - deposit User funds into a home chain from ERC20 (here)
    // - withdraw User funds from a home chain to ERC20 (here)
    // - release Node's funds from a channel to the Node's internal vault balance (here)
    // - Node deposit funds into a home chain channel during bridging-deposit (here)
    function checkpoint(bytes32 channelId, CrossChainState calldata candidate, CrossChainState[] calldata proofs) external payable {}

    event Checkpointed(bytes32 indexed channelId, CrossChainState candidate);

    // usage:
    // - close a channel in case either party is unresponsive
    function challenge(bytes32 channelId, CrossChainState calldata candidate, CrossChainState[] calldata proofs, bytes calldata challengerSig) external payable {}

    event Challenged(bytes32 indexed channelId, CrossChainState candidate, uint256 challengeExpiry);

    // usage:
    // - unilaterally close the channel withdrawing all funds
    function close(bytes32 channelId, CrossChainState calldata candidate, CrossChainState[] calldata proofs) external payable {}

    event Closed(bytes32 indexed channelId, CrossChainState finalState);

    // ========= Escrow ==========

    struct EscrowDepositMetadata {
        Definition definition;
        CrossChainState lastState;
        uint256 participantLockedFunds;
        uint64 unlockExpiry;
        uint64 challengeExpiry;
    }

    // include Node address for simplicity?
    // EnumerableMap(bytes32 escrowId => EscrowDepositMetadata) escrowDeposits;

    // usage:
    // - lock user funds during bridging-deposit
    function initiateEscrowDeposit(Definition calldata def, CrossChainState calldata initCCS) external payable {
        // -- checks --
        // require(ccs.homeChainId != block.chainid)
        // validate signatures over (escrowId, Definition, CrossChainState)

        // -- effects --
        // store escrow data
        // add user funding amount to "escrow locked" amount
        // start `unlockTimer` (a constant on the Contract level)

        // -- interactions --
        // pull funds from user
    }

    event EscrowDepositInitiated(bytes32 indexed escrowId, address indexed participant, address indexed node, Definition definition, CrossChainState initialState);

    // usage:
    // - challenge bridging-deposit process
    function challengeEscrowDeposit(bytes32 escrowId, CrossChainState calldata candidate, CrossChainState[] calldata proof) external payable {}

    event EscrowDepositChallenged(bytes32 indexed escrowId, CrossChainState candidate, uint256 challengeExpiry);

    // usage: (optional)
    // - unlock Node's funds after bridging-deposit for better funds efficiency
    // - resolve a challenge during a bridging-deposit process
    function finalizeEscrowDeposit(bytes32 escrowId, CrossChainState calldata candidate, CrossChainState[2] calldata proof) external payable {}

    event EscrowDepositFinalized(bytes32 indexed escrowId, CrossChainState finalState);

    struct EscrowWithdrawalMetadata {
        Definition definition;
        CrossChainState lastState;
        uint256 participantLockedFunds;
    }

    // usage:
    // - lock Node's funds during bridging-withdrawal
    function initiateEscrowWithdrawal(Definition calldata def, CrossChainState calldata initCCS) external payable {}

    event EscrowWithdrawalInitiated(bytes32 indexed escrowId, address indexed participant, address indexed node, Definition definition, CrossChainState initialState);

    // usage:
    // - challenge bridging-withdrawal process
    function challengeEscrowWithdrawal(bytes32 escrowId, CrossChainState calldata candidate, CrossChainState[] calldata proof) external payable {}

    event EscrowWithdrawalChallenged(bytes32 indexed escrowId, CrossChainState candidate, uint256 challengeExpiry);

    // usage:
    // - unlock user funds during bridging-withdrawal
    function finalizeEscrowWithdrawal(bytes32 escrowId, CrossChainState calldata candidate) external payable {}

    event EscrowWithdrawalFinalized(bytes32 indexed escrowId, CrossChainState finalState);
}
