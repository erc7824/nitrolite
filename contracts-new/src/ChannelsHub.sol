// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {IERC20} from "lib/openzeppelin-contracts/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "lib/openzeppelin-contracts/contracts/token/ERC20/utils/SafeERC20.sol";

import {IVault} from "./interfaces/IVault.sol";
import {Definition, CrossChainState, State, StateIntent} from "./interfaces/Types.sol";

import {Utils} from "./Utils.sol";

contract ChannelsHub is IVault {
    using {Utils.validateSignatures} for CrossChainState;
    using {Utils.isEmpty} for State;
    using SafeERC20 for IERC20;

    error InvalidAddress();
    error InvalidAmount();
    error InvalidValue();
    error AddressCollision(address collision);
    error IncorrectChallengeDuration();

    struct Metadata {
        Definition definition;
        CrossChainState lastState;
        uint256 participantLockedFunds;
        uint256 nodeLockedFunds;
        uint64 challengeExpiry; // timestamp when challenge period ends
    }

    // ======== Contract Storage ==========

    uint32 constant MIN_CHALLENGE_DURATION = 1 days;

    mapping(bytes32 channelId => Metadata meta) internal _channels;

    mapping(address node => mapping(address token => uint256 balance)) internal _nodeBalances;

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
    // TODO: replace with user-locked liquidity provision

    // user to deposit funds into node's balance
    function deposit(address node, address token, uint256 amount) external payable {
        require(node != address(0), InvalidAddress());
        require(amount > 0, InvalidAmount());

        _nodeBalances[node][token] += amount;

        _pullFunds(token, amount);

        emit Deposited(node, token, amount);
    }

    function withdraw(address to, address token, uint256 amount) external {
        require(to != address(0), InvalidAddress());
        require(amount > 0, InvalidAmount());

        uint256 currentBalance = _nodeBalances[msg.sender][token];
        require(currentBalance >= amount, "insufficient balance");

        _nodeBalances[msg.sender][token] = currentBalance - amount;

        _pushFunds(to, token, amount);

        emit Withdrawn(msg.sender, token, amount);
    }

    // ******

    // ========== Channel lifecycle ==========

    // usage:
    // - open a new channel with User deposit funds ("entry to Clearnet")
    // - open a new channel with Node transfer funds to User (User joins Clearnet after already receiving funds)
    // - open and close channel in one go for atomic for User receiving and withdrawing funds
    function create(Definition calldata def, CrossChainState calldata initCCS) external payable {
        // -- checks --
        _requireValidDefinition(def);

        require(initCCS.homeChainId == block.chainid, "invalid home chain id for initial state");
        require(initCCS.version == 0, "invalid initial version");
        require(initCCS.intent == StateIntent.CREATE, "invalid initial intent");
        require(!initCCS.homeState.isEmpty(), "home state cannot be empty");
        require(initCCS.nonHomeState.isEmpty(), "non-home state must be empty");
        require(initCCS.homeState.nodeBalance.allocation == 0, "node balance must be zero in initial state");

        bytes32 channelId = Utils.getChannelId(def);

        initCCS.validateSignatures(channelId, def.participant, def.node);

        // -- effects --
        _channels[channelId] = Metadata({
            definition: def,
            lastState: initCCS,
            participantLockedFunds: initCCS.homeState.participantBalance.allocation,
            nodeLockedFunds: 0,
            challengeExpiry: 0
        });

        // -- interactions --
        _pullFunds(
            initCCS.homeState.token,
            initCCS.homeState.participantBalance.allocation
        );

        emit ChannelCreated(channelId, def.participant, def.node, def, initCCS);
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
    function checkpoint(bytes32 channelId, CrossChainState calldata candidate, CrossChainState[] calldata proof) external payable {
        // There is a problem for the contract to differentiate between actions in the same state.
        // a "deposit" state may also contain "unlock to / lock from Node"
        // such action can either be:
        // - included in the same state OR
        // - provided as a proof state, which action must also be performed before processing a candidate ("deposit") state

        // However, the "proof" approach will lead to situations when it is possible to execute such proof several times
    }

    event ChannelCheckpointed(bytes32 indexed channelId, CrossChainState candidate);

    // usage:
    // - close a channel in case either party is unresponsive
    function challenge(bytes32 channelId, CrossChainState calldata candidate, CrossChainState[] calldata proof, bytes calldata challengerSig) external payable {}

    event ChannelChallenged(bytes32 indexed channelId, CrossChainState candidate, uint256 challengeExpiry);

    // usage:
    // - unilaterally close the channel withdrawing all funds
    function close(bytes32 channelId, CrossChainState calldata candidate, CrossChainState[] calldata proof) external payable {}

    event ChannelClosed(bytes32 indexed channelId, CrossChainState finalState);

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

    // ========= Internal ==========

    function _requireValidDefinition(Definition calldata def) internal pure {
        require(def.participant != address(0), InvalidAddress());
        require(def.node != address(0), InvalidAddress());
        require(def.participant != def.node, AddressCollision(def.participant));
        require(def.challengeDuration >= MIN_CHALLENGE_DURATION, IncorrectChallengeDuration());
    }

    function _pullFunds(address token, uint256 amount) internal {
        if (amount == 0) return;

        if (token == address(0)) {
            require(msg.value == amount, InvalidValue());
        } else {
            require(msg.value == 0, InvalidValue());
        }

        if (token != address(0)) {
            IERC20(token).safeTransferFrom(msg.sender, address(this), amount);
        }
    }

    function _pushFunds(address to, address token, uint256 amount) internal {
        if (amount == 0) return;

        if (token == address(0)) {
            payable(to).transfer(amount);
        } else {
            IERC20(token).safeTransfer(to, amount);
        }
    }

    function _getChainId() internal view returns (uint64) {
        return uint64(block.chainid); // narrowing to uint64 is safe according to EIP-2294
    }
}
