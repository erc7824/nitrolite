// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

contract IContract {
    // User and Node have different logic in the channel
    // only 2 participants in the channel

    // Funds:
    // When depositing to or withdrawing from, source of User funds is ERC20 contract, and not the Vault
    // Node's funds do not leave the Contract

    // ========= Channel Types ==========

    struct Definition {
        uint32 challengeDuration;
        address participant;
        address node; // TODO: move to contract level for now, until we don't have other nodes. Are there any disadvantages?
        uint256 nonce;
        // to be added later:
        // address executionModule;
        // address signatureValidator;
    }

    struct CrossChainState {
        uint256 version;

        uint64 homeChainId;
        State[] states;

        bool isFinal;

        bytes participantSig;
        bytes nodeSig;

        // intent?
    }

    struct State {
        // intent?
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

    // ========== Channel lifecycle ==========

    // usage:
    // - open a new channel with User deposit funds ("entry to Clearnet")
    // - open a new channel with Node transfer funds to User (User joins Clearnet after already receiving funds)
    // - open and close channel in one go for atomic for User receiving and withdrawing funds
    function create(Definition def, CrossChainState initCCS) external payable {
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

    event ChannelCreated(bytes32 indexed channelId, address indexed participant, Definition definition, CrossChainState initialState);

    // usage:
    // - "open" channel on new chain in case of chain migration
    function migrateChannelHere(Definition def, CrossChainState candidate, CrossChainState[] proof) external payable {
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

    event ChannelMigrated(bytes32 indexed channelId, address indexed participant, Definition definition, CrossChainState initialState);

    // usage:
    // - deposit User funds into a home chain from ERC20 (here)
    function deposit() external payable;

    // usage:
    // - withdraw User funds from a home chain to ERC20 (here)
    function withdraw() external payable;

    // usage:
    // - release Node's funds from a channel to the Node's internal vault balance (here)
    function releaseNodeFunds() external payable;

    // usage:
    // - "close" channel on old chain in case of chain migration (basically, a "checkpoint" in a common sense)
    // - acknowledge latest funds change on Home chain
    // - resolve a challenge
    // - Node deposit funds into a home chain channel during bridging-deposit
    function checkpoint() external payable {

    }

    event Checkpointed(bytes32 indexed channelId, CrossChainState candidate);

    // usage:
    // - close a channel in case either party is unresponsive
    function challenge() external payable;

    event Challenged(bytes32 indexed channelId, CrossChainState candidate, uint256 challengeExpiry);

    // usage:
    // - unilaterally close the channel withdrawing all funds
    function close() external payable;

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
    EnumerableMap(bytes32 escrowId => EscrowDepositMetadata) escrowDeposits;

    // usage:
    // - lock user funds during bridging-deposit
    function initiateEscrowDeposit(Definition def, CrossChainState initCCS) external payable {
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

    event EscrowDepositInitiated(bytes32 indexed escrowId, address indexed participant, Definition definition, CrossChainState initialState);

    // usage:
    // - challenge bridging-deposit process
    function challengeEscrowDeposit(bytes32 escrowId, CrossChainState candidate, CrossChainState[] proof) external payable;

    event EscrowDepositChallenged(bytes32 indexed escrowId, CrossChainState candidate, uint256 challengeExpiry);

    // usage:
    // - resolve a challenge during a bridging-deposit process
    function checkpointEscrowDeposit(bytes32 escrowId, CrossChainState candidate, CrossChainState[] proof) external payable;

    event EscrowDepositCheckpointed(bytes32 indexed escrowId, CrossChainState candidate);

    // usage: (optional)
    // - unlock Node's funds after bridging-deposit for better funds efficiency
    function finalizeEscrowDeposit(bytes32 escrowId, CrossChainState candidate, CrossChainState[2] proof) external payable;

    event EscrowDepositFinalized(bytes32 indexed escrowId, CrossChainState finalState);

    struct EscrowWithdrawalMetadata {
        Definition definition;
        CrossChainState lastState;
        uint256 participantLockedFunds;
    }

    // usage:
    // - lock Node's funds during bridging-withdrawal
    function initiateEscrowWithdrawal(Definition def, CrossChainState initCCS) external payable;

    event EscrowWithdrawalInitiated(bytes32 indexed escrowId, address indexed participant, Definition definition, CrossChainState initialState);

    // usage:
    // - unlock user funds during bridging-withdrawal
    function finalizeEscrowWithdrawal(bytes32 escrowId, CrossChainState candidate) external payable;

    event EscrowWithdrawalFinalized(bytes32 indexed escrowId, CrossChainState finalState);
}
