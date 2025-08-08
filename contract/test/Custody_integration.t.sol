// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {Test} from "lib/forge-std/src/Test.sol";
import {Vm} from "lib/forge-std/src/Vm.sol";

import {TestUtils} from "./TestUtils.sol";
import {MockERC20} from "./mocks/MockERC20.sol";
import {MockFlagERC1271} from "./mocks/MockFlagERC1271.sol";

import {Custody} from "../src/Custody.sol";
import {SimpleConsensus} from "../src/adjudicators/SimpleConsensus.sol";
import {Utils} from "../src/Utils.sol";
import {ChannelStatus, Channel, State, Allocation, StateIntent, STATE_TYPEHASH} from "../src/interfaces/Types.sol";

/// @dev used to mock the deployment of ERC-4337 accounts to specific addresses
contract CheatERC6492Factory is Test {
    function createAccount(address to, bool flag) external {
        deployCodeTo("MockFlagERC1271", abi.encode(flag), to);
    }
}

contract CustodyIntegrationTest_Signatures is Test {
    Custody public custody;
    SimpleConsensus public adjudicator;
    MockERC20 public token;
    CheatERC6492Factory public factory;

    bytes32 custodyDomainSeparator;

    // Test participants
    address public participant1;
    address public participant2;
    uint256 public participant1PrivateKey;
    uint256 public participant2PrivateKey;

    // Test parameters
    uint256 constant DEPOSIT_AMOUNT = 1000;
    uint256 constant INITIAL_BALANCE = 10000;
    uint64 constant CHALLENGE_DURATION = 3600; // 1 hour
    uint64 constant NONCE = 1;

    // Channel and state tracking
    Channel public channel;
    bytes32 public channelId;

    // Constants for participant ordering
    uint256 private constant PARTICIPANT_1 = 0;
    uint256 private constant PARTICIPANT_2 = 1;

    function setUp() public {
        // Deploy contracts
        custody = new Custody();
        adjudicator = new SimpleConsensus(address(this), address(custody));
        token = new MockERC20("Test Token", "TEST", 18);
        factory = new CheatERC6492Factory();

        custodyDomainSeparator = TestUtils.buildDomainSeparatorForContract(custody);

        // Set up participants
        participant1PrivateKey = vm.createWallet("participant1").privateKey;
        participant2PrivateKey = vm.createWallet("participant2").privateKey;
        participant1 = vm.addr(participant1PrivateKey);
        participant2 = vm.addr(participant2PrivateKey);

        // Fund participants
        token.mint(participant1, INITIAL_BALANCE);
        token.mint(participant2, INITIAL_BALANCE);

        // Approve token transfers
        vm.prank(participant1);
        token.approve(address(custody), INITIAL_BALANCE);

        vm.prank(participant2);
        token.approve(address(custody), INITIAL_BALANCE);

        // Create channel
        address[] memory participants = new address[](2);
        participants[PARTICIPANT_1] = participant1;
        participants[PARTICIPANT_2] = participant2;

        channel = Channel({
            participants: participants,
            adjudicator: address(adjudicator),
            challenge: CHALLENGE_DURATION,
            nonce: NONCE
        });

        channelId = Utils.getChannelId(channel);
    }

    // ==================== SIGNATURE HELPERS ====================

    function _signStateEIP6492(address signer, State memory) internal view returns (bytes memory) {
        bytes memory signature = "dummy signature";
        bool flag = true; // meaning each EIP-1271 signature is valid

        bytes memory factoryCalldata = abi.encodeWithSelector(CheatERC6492Factory.createAccount.selector, signer, flag);

        bytes memory erc6492Sig = abi.encode(address(factory), factoryCalldata, signature);
        return abi.encodePacked(erc6492Sig, Utils.ERC6492_DETECTION_SUFFIX);
    }

    function _signChallenge(State memory state, uint256 privateKey) internal view returns (bytes memory) {
        bytes memory packedChallengeState = abi.encodePacked(Utils.getPackedState(channelId, state), "challenge");
        return TestUtils.sign(vm, privateKey, packedChallengeState);
    }

    // ==================== STATE CREATION HELPERS ====================

    function _createState(StateIntent intent, uint256 version, bytes memory data, uint256 amount1, uint256 amount2)
        internal
        view
        returns (State memory)
    {
        Allocation[] memory allocations = new Allocation[](2);
        allocations[PARTICIPANT_1] = Allocation({destination: participant1, token: address(token), amount: amount1});
        allocations[PARTICIPANT_2] = Allocation({destination: participant2, token: address(token), amount: amount2});

        return State({intent: intent, version: version, data: data, allocations: allocations, sigs: new bytes[](0)});
    }

    function _createInitialState() internal view returns (State memory) {
        return _createState(StateIntent.INITIALIZE, 0, bytes(""), DEPOSIT_AMOUNT, DEPOSIT_AMOUNT);
    }

    function _createOperateState(uint256 version, bytes memory data) internal view returns (State memory) {
        return _createState(StateIntent.OPERATE, version, data, DEPOSIT_AMOUNT, DEPOSIT_AMOUNT);
    }

    function _createFinalState(uint256 version) internal view returns (State memory) {
        return _createState(StateIntent.FINALIZE, version, bytes(""), DEPOSIT_AMOUNT, DEPOSIT_AMOUNT);
    }

    // ==================== MAIN INTEGRATION TEST ====================

    function test_fullChannelLifecycle_withMixedSignatures() public {
        // ==================== 1. CREATE CHANNEL ====================

        // Create initial state - participant1 uses EIP191
        State memory initialState = _createInitialState();
        initialState.sigs = new bytes[](1);
        initialState.sigs[0] = TestUtils.signStateEIP191(vm, channelId, initialState, participant1PrivateKey);

        vm.prank(participant1);
        custody.depositAndCreate(address(token), DEPOSIT_AMOUNT, channel, initialState);

        (, ChannelStatus status,,,) = custody.getChannelData(channelId);

        assertTrue(status == ChannelStatus.INITIAL, "Channel should be in INITIAL status");

        // ==================== 2. JOIN CHANNEL ====================

        vm.prank(participant2);
        custody.deposit(participant2, address(token), DEPOSIT_AMOUNT);

        // Participant2 joins using raw ECDSA signature
        bytes memory participant2JoinSig = TestUtils.signStateRaw(vm, channelId, initialState, participant2PrivateKey);

        vm.prank(participant2);
        custody.join(channelId, 1, participant2JoinSig);

        (, status,,,) = custody.getChannelData(channelId);

        assertTrue(status == ChannelStatus.ACTIVE, "Channel should be in ACTIVE status");

        // ==================== 3. CHALLENGE CHANNEL ====================

        // Create challenge state - participant1 uses EIP712, participant2 uses raw ECDSA
        State memory challengeState = _createOperateState(1, bytes("challenge data"));
        challengeState.sigs = new bytes[](2);
        challengeState.sigs[PARTICIPANT_1] = TestUtils.signStateEIP712(
            vm, channelId, challengeState, STATE_TYPEHASH, custodyDomainSeparator, participant1PrivateKey
        );
        challengeState.sigs[PARTICIPANT_2] =
            TestUtils.signStateRaw(vm, channelId, challengeState, participant2PrivateKey);

        bytes memory challengerSig = _signChallenge(challengeState, participant1PrivateKey);

        vm.prank(participant1);
        custody.challenge(channelId, challengeState, new State[](0), challengerSig);

        uint256 challengeExpiry;
        (, status,, challengeExpiry,) = custody.getChannelData(channelId);

        assertTrue(status == ChannelStatus.DISPUTE, "Channel should be in DISPUTE status");
        assertTrue(challengeExpiry > block.timestamp, "Channel should have challengeExpiry set in future");

        // ==================== 4. CHECKPOINT TO RESOLVE CHALLENGE ====================

        // Create checkpoint state with higher version - participant1 uses raw ECDSA, participant2 uses raw ECDSA
        State memory checkpointState = _createOperateState(2, bytes("checkpoint data"));
        checkpointState.sigs = new bytes[](2);
        checkpointState.sigs[PARTICIPANT_1] =
            TestUtils.signStateRaw(vm, channelId, checkpointState, participant1PrivateKey);
        checkpointState.sigs[PARTICIPANT_2] =
            TestUtils.signStateRaw(vm, channelId, checkpointState, participant2PrivateKey);

        vm.prank(participant2);
        custody.checkpoint(channelId, checkpointState, new State[](0));

        (, status,, challengeExpiry,) = custody.getChannelData(channelId);

        assertTrue(status == ChannelStatus.ACTIVE, "Channel should be back to ACTIVE status after checkpoint");
        assertEq(challengeExpiry, 0, "Channel should have no challengeExpiry after checkpoint");

        // ==================== 5. CHECKPOINT AGAIN ====================

        // Create checkpoint state with higher version - participant1 uses EIP-6492, participant2 uses raw ECDSA
        checkpointState = _createOperateState(3, bytes("checkpoint data"));
        checkpointState.sigs = new bytes[](2);
        checkpointState.sigs[PARTICIPANT_1] = _signStateEIP6492(participant1, checkpointState);
        checkpointState.sigs[PARTICIPANT_2] =
            TestUtils.signStateRaw(vm, channelId, checkpointState, participant2PrivateKey);

        vm.prank(participant2);
        custody.checkpoint(channelId, checkpointState, new State[](0));

        (, status,, challengeExpiry,) = custody.getChannelData(channelId);

        assertTrue(status == ChannelStatus.ACTIVE, "Channel should still be in an ACTIVE status after checkpoint");

        // ==================== 6. CLOSE CHANNEL ====================

        // Create final state - participant1 uses EIP1271, participant2 uses raw ECDSA
        State memory finalState = _createFinalState(4);
        finalState.sigs = new bytes[](2);
        // as participant1 already has a contract at its address, we assume this contract expects EIP-191 signature
        finalState.sigs[PARTICIPANT_1] = TestUtils.signStateEIP191(vm, channelId, finalState, participant1PrivateKey);
        finalState.sigs[PARTICIPANT_2] = TestUtils.signStateRaw(vm, channelId, finalState, participant2PrivateKey);

        vm.prank(participant1);
        custody.close(channelId, finalState, new State[](0));

        (, status,,,) = custody.getChannelData(channelId);

        assertTrue(status == ChannelStatus.VOID, "Channel should have VOID status after close (channel data deleted)");

        // ==================== 7. VERIFY FINAL BALANCES ====================

        address[] memory users = new address[](2);
        users[0] = participant1;
        users[1] = participant2;

        address[] memory tokens = new address[](1);
        tokens[0] = address(token);

        uint256[][] memory balances = custody.getAccountsBalances(users, tokens);

        assertEq(balances[0][0], DEPOSIT_AMOUNT, "Participant1 should have deposit amount available");
        assertEq(balances[1][0], DEPOSIT_AMOUNT, "Participant2 should have deposit amount available");

        bytes32[][] memory channels = custody.getOpenChannels(users);
        assertEq(channels[0].length, 0, "Participant1 should have no open channels");
        assertEq(channels[1].length, 0, "Participant2 should have no open channels");
    }

    // ==================== HELPER FUNCTION FOR VERIFICATION ====================

    function skipChallengeTime() internal {
        skip(CHALLENGE_DURATION + 1);
    }
}
