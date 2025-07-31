// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Test} from "lib/forge-std/src/Test.sol";
import {Vm} from "lib/forge-std/src/Vm.sol";

import {ECDSA} from "lib/openzeppelin-contracts/contracts/utils/cryptography/ECDSA.sol";

import {TestUtils} from "../TestUtils.sol";
import {MockERC20} from "../mocks/MockERC20.sol";
import {MockEIP712} from "../mocks/MockEIP712.sol";

import {IAdjudicator} from "../../src/interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, StateIntent, STATE_TYPEHASH} from "../../src/interfaces/Types.sol";
import {SimpleConsensus} from "../../src/adjudicators/SimpleConsensus.sol";
import {Utils} from "../../src/Utils.sol";

contract SimpleConsensusTest is Test {
    using ECDSA for bytes32;

    SimpleConsensus public adjudicator;
    MockEIP712 public mockedChannelImpl;

    // Mockup constructor parameters
    address mockedOwner = address(0x456);

    address public host;
    address public guest;
    uint256 public hostPrivateKey;
    uint256 public guestPrivateKey;

    Channel public channel;
    MockERC20 public token;

    uint256 private constant HOST = 0;
    uint256 private constant GUEST = 1;

    function setUp() public {
        mockedChannelImpl = new MockEIP712("TestChannelImpl", "1.0");
        adjudicator = new SimpleConsensus(mockedOwner, address(mockedChannelImpl));

        hostPrivateKey = 0x1;
        guestPrivateKey = 0x2;
        host = vm.addr(hostPrivateKey);
        guest = vm.addr(guestPrivateKey);

        token = new MockERC20("Test Token", "TEST", 18);

        address[] memory participants = new address[](2);
        participants[HOST] = host;
        participants[GUEST] = guest;
        channel = Channel({
            participants: participants,
            adjudicator: address(adjudicator),
            challenge: 3600, // 1 hour challenge period
            nonce: 1
        });
    }

    function createAllocations(uint256 hostAmount, uint256 guestAmount) internal view returns (Allocation[2] memory) {
        Allocation[2] memory allocations;

        allocations[HOST] = Allocation({destination: host, token: address(token), amount: hostAmount});
        allocations[GUEST] = Allocation({destination: guest, token: address(token), amount: guestAmount});

        return allocations;
    }

    function _createInitialState(string memory data) internal view returns (State memory) {
        return _createState(data, 0, StateIntent.INITIALIZE);
    }

    function _createOperateState(string memory data, uint256 version) internal view returns (State memory) {
        return _createState(data, version, StateIntent.OPERATE);
    }

    function _createState(string memory data, uint256 version, StateIntent intent)
        internal
        view
        returns (State memory)
    {
        Allocation[2] memory allocations = createAllocations(50, 50);

        State memory state;
        state.intent = intent;
        state.version = version;
        state.data = bytes(data);
        state.allocations = new Allocation[](2);
        state.allocations[HOST] = allocations[HOST];
        state.allocations[GUEST] = allocations[GUEST];
        state.sigs = new bytes[](0);

        return state;
    }

    function _createResizeState(string memory data, uint256 version, int256[] memory resizeAmounts)
        internal
        view
        returns (State memory)
    {
        State memory state = _createState(data, version, StateIntent.RESIZE);
        state.data = abi.encode(resizeAmounts, bytes(data));

        return state;
    }

    function _signState(State memory state, uint256 privateKey) internal view returns (bytes memory) {
        bytes memory packedState = Utils.getPackedState(Utils.getChannelId(channel), state);
        return TestUtils.sign(vm, privateKey, packedState);
    }

    function _signStateEIP191(State memory state, uint256 privateKey) internal view returns (bytes memory) {
        bytes memory packedState = Utils.getPackedState(Utils.getChannelId(channel), state);
        return TestUtils.signEIP191(vm, privateKey, packedState);
    }

    function _signStateEIP712(State memory state, uint256 privateKey) internal view returns (bytes memory) {
        bytes32 channelId = Utils.getChannelId(channel);
        bytes32 domainSeparator = mockedChannelImpl.domainSeparator();
        return TestUtils.signStateEIP712(vm, channelId, state, STATE_TYPEHASH, domainSeparator, privateKey);
    }

    function test_adjudicate_firstState_valid_withRawECDSASignatures() public {
        State memory initialState = _createInitialState("initial state");
        initialState.sigs = new bytes[](2);
        initialState.sigs[HOST] = _signState(initialState, hostPrivateKey);
        initialState.sigs[GUEST] = _signState(initialState, guestPrivateKey);

        bool valid = adjudicator.adjudicate(channel, initialState, new State[](0));
        assertTrue(valid, "Valid first state transition should be accepted");
    }

    function test_adjudicate_firstState_valid_withEIP191Signatures() public {
        State memory initialState = _createInitialState("initial state");
        initialState.sigs = new bytes[](2);
        initialState.sigs[HOST] = _signStateEIP191(initialState, hostPrivateKey);
        initialState.sigs[GUEST] = _signStateEIP191(initialState, guestPrivateKey);

        bool valid = adjudicator.adjudicate(channel, initialState, new State[](0));
        assertTrue(valid, "Valid first state transition with EIP191 signatures should be accepted");
    }

    function test_adjudicate_firstState_valid_withEIP712Signatures() public {
        State memory initialState = _createInitialState("initial state");
        initialState.sigs = new bytes[](2);
        initialState.sigs[HOST] = _signStateEIP712(initialState, hostPrivateKey);
        initialState.sigs[GUEST] = _signStateEIP712(initialState, guestPrivateKey);

        bool valid = adjudicator.adjudicate(channel, initialState, new State[](0));
        assertTrue(valid, "Valid first state transition with EIP712 signatures should be accepted");
    }

    function test_adjudicate_firstState_revert_whenMissingParticipantSignature() public {
        State memory initialState = _createInitialState("initial state");
        initialState.sigs = new bytes[](1);
        initialState.sigs[HOST] = _signState(initialState, hostPrivateKey);

        bool valid = adjudicator.adjudicate(channel, initialState, new State[](0));
        assertFalse(valid, "First state without both signatures should be rejected");
    }

    function test_adjudicate_firstState_revert_whenIncorrectIntent() public {
        State memory initialState = _createInitialState("initial state");
        initialState.intent = StateIntent.OPERATE; // Incorrect intent, should be INITIALIZE
        initialState.sigs = new bytes[](2);
        initialState.sigs[HOST] = _signState(initialState, hostPrivateKey);
        initialState.sigs[GUEST] = _signState(initialState, guestPrivateKey);

        bool valid = adjudicator.adjudicate(channel, initialState, new State[](0));
        assertFalse(valid, "First state with incorrect intent should be rejected");
    }

    function test_adjudicate_firstState_revert_whenIncorrectVersion() public {
        State memory initialState = _createInitialState("initial state");
        initialState.version = 1; // Incorrect version, should be 0
        initialState.sigs = new bytes[](2);
        initialState.sigs[HOST] = _signState(initialState, hostPrivateKey);
        initialState.sigs[GUEST] = _signState(initialState, guestPrivateKey);

        bool valid = adjudicator.adjudicate(channel, initialState, new State[](0));
        assertFalse(valid, "First state with incorrect version should be rejected");
    }

    function test_adjudicate_laterState_valid() public {
        State memory state1 = _createOperateState("state 42", 42);
        state1.sigs = new bytes[](2);
        state1.sigs[HOST] = _signState(state1, hostPrivateKey);
        state1.sigs[GUEST] = _signState(state1, guestPrivateKey);

        bool valid = adjudicator.adjudicate(channel, state1, new State[](0));
        assertTrue(valid, "Valid state transition from 1 to 2 should be accepted");
    }

    function test_adjudicate_revert_whenTooManyProofs() public {
        State memory state1 = _createOperateState("state 1", 1);
        state1.sigs = new bytes[](2);
        state1.sigs[HOST] = _signState(state1, hostPrivateKey);
        state1.sigs[GUEST] = _signState(state1, guestPrivateKey);

        State memory state2 = _createOperateState("state 2", 2);
        state2.sigs = new bytes[](2);
        state2.sigs[HOST] = _signState(state2, hostPrivateKey);
        state2.sigs[GUEST] = _signState(state2, guestPrivateKey);

        State[] memory tooManyProofs = new State[](1);
        tooManyProofs[0] = state1;

        bool valid = adjudicator.adjudicate(channel, state2, tooManyProofs);
        assertFalse(valid, "State with too many proofs should be rejected");
    }

    // Test signature validation using a non-corrupt signature but wrong signer
    function test_adjudicate_revert_wrongSigner() public {
        // Create state with signatures from wrong participants
        State memory state = _createOperateState("state 13", 13);
        state.sigs = new bytes[](2);

        state.sigs[HOST] = _signState(state, guestPrivateKey); // Should be host, but using guest
        state.sigs[GUEST] = _signState(state, guestPrivateKey);

        // Verify that the signature from wrong participant is rejected
        bool valid = adjudicator.adjudicate(channel, state, new State[](0));
        assertFalse(valid, "State with wrong signer should be rejected");
    }
}
