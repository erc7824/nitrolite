// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Test} from "lib/forge-std/src/Test.sol";
import {Vm} from "lib/forge-std/src/Vm.sol";

import {MessageHashUtils} from "lib/openzeppelin-contracts/contracts/utils/cryptography/MessageHashUtils.sol";
import {ECDSA} from "lib/openzeppelin-contracts/contracts/utils/cryptography/ECDSA.sol";

import {TestUtils} from "../TestUtils.sol";

import {IAdjudicator} from "../../src/interfaces/IAdjudicator.sol";
import {Channel, State, Allocation, StateIntent} from "../../src/interfaces/Types.sol";
import {Counter, Counter as CounterContract} from "../../src/adjudicators/Counter.sol";
import {Utils} from "../../src/Utils.sol";

contract CounterTest is Test {
    CounterContract public counter;

    // Test accounts
    address public host;
    address public guest;
    uint256 public hostPrivateKey;
    uint256 public guestPrivateKey;

    // Channel parameters
    Channel public channel;

    // Constants for participant ordering
    uint256 private constant HOST = 0;
    uint256 private constant GUEST = 1;

    function setUp() public {
        // Deploy the adjudicator
        counter = new CounterContract();

        // Set private keys and corresponding addresses
        hostPrivateKey = 0x1;
        guestPrivateKey = 0x2;
        host = vm.addr(hostPrivateKey);
        guest = vm.addr(guestPrivateKey);

        // Set up the channel with the two participants
        address[] memory participants = new address[](2);
        participants[HOST] = host;
        participants[GUEST] = guest;
        channel = Channel({
            participants: participants,
            adjudicator: address(counter),
            challenge: 3600, // 1-hour challenge period
            nonce: 1
        });
    }

    // -------------------- HELPERS --------------------

    function _createInitialState(uint256 target) internal pure returns (State memory) {
        State memory state;
        state.intent = StateIntent.INITIALIZE;
        state.version = 0;
        state.data = abi.encode(Counter.Data({target: target}));

        // Create dummy allocations
        state.allocations = new Allocation[](2);
        state.allocations[HOST] = Allocation({destination: address(0), token: address(0), amount: 100});
        state.allocations[GUEST] = Allocation({destination: address(0), token: address(0), amount: 100});

        state.sigs = new bytes[](0);
        return state;
    }

    function _createCounterState(uint256 target, uint256 version) internal pure returns (State memory) {
        State memory state;
        state.intent = StateIntent.OPERATE; // Set the proper intent for operation
        state.version = version;

        // Encode the Counter data
        state.data = abi.encode(Counter.Data({target: target}));

        // Create dummy allocations
        state.allocations = new Allocation[](2);
        state.allocations[HOST] = Allocation({destination: address(0), token: address(0), amount: 100});
        state.allocations[GUEST] = Allocation({destination: address(0), token: address(0), amount: 100});

        state.sigs = new bytes[](0);
        return state;
    }

    function _createResizeState(uint256 target, uint256 version, int256[] memory resizeAmounts)
        internal
        pure
        returns (State memory)
    {
        State memory state;
        state.intent = StateIntent.RESIZE;
        state.version = version;

        // Encode resize amounts and the Counter data
        state.data = abi.encode(resizeAmounts, abi.encode(Counter.Data({target: target})));

        // Create dummy allocations
        state.allocations = new Allocation[](2);
        state.allocations[HOST] = Allocation({destination: address(0), token: address(0), amount: 100});
        state.allocations[GUEST] = Allocation({destination: address(0), token: address(0), amount: 100});

        state.sigs = new bytes[](0);
        return state;
    }

    function _signState(State memory state, uint256 privateKey) internal view returns (bytes memory) {
        bytes memory packedState = Utils.getPackedState(Utils.getChannelId(channel), state);
        return TestUtils.sign(vm, privateKey, packedState);
    }

    // -------------------- FIRST STATE TRANSITION TESTS --------------------

    function test_adjudicate_firstState_valid() public {
        // Create initial state with target 10
        State memory initialState = _createInitialState(10);
        initialState.sigs = new bytes[](2);
        initialState.sigs[HOST] = _signState(initialState, hostPrivateKey);
        initialState.sigs[GUEST] = _signState(initialState, guestPrivateKey);

        // Create first state (version 1)
        State memory firstState = _createCounterState(10, 1);
        firstState.sigs = new bytes[](1);
        firstState.sigs[0] = _signState(firstState, hostPrivateKey); // Host signs state 1

        // Provide the initial state as proof
        State[] memory proofs = new State[](1);
        proofs[0] = initialState;

        // Verify that the transition is valid
        bool valid = counter.adjudicate(channel, firstState, proofs);
        assertTrue(valid, "Valid first state transition should be accepted");
    }

    function test_adjudicate_firstState_revert_whenTargetExceeded() public {
        // Create initial state with target 10
        State memory initialState = _createInitialState(10);
        initialState.sigs = new bytes[](2);
        initialState.sigs[HOST] = _signState(initialState, hostPrivateKey);
        initialState.sigs[GUEST] = _signState(initialState, guestPrivateKey);

        // Create first state with version exceeding target
        State memory firstState = _createCounterState(10, 11); // version > target
        firstState.sigs = new bytes[](1);
        firstState.sigs[0] = _signState(firstState, hostPrivateKey);

        // Provide the initial state as proof
        State[] memory proofs = new State[](1);
        proofs[0] = initialState;

        // Verify that the transition is rejected
        bool valid = counter.adjudicate(channel, firstState, proofs);
        assertFalse(valid, "State with version exceeding target should be rejected");
    }

    function test_adjudicate_firstState_revert_whenIncorrectSigner() public {
        // Create initial state with target 10
        State memory initialState = _createInitialState(10);
        initialState.sigs = new bytes[](2);
        initialState.sigs[HOST] = _signState(initialState, hostPrivateKey);
        initialState.sigs[GUEST] = _signState(initialState, guestPrivateKey);

        // Create first state signed by guest (should be host)
        State memory firstState = _createCounterState(10, 1);
        firstState.sigs = new bytes[](1);
        firstState.sigs[0] = _signState(firstState, guestPrivateKey); // Guest signs instead of host

        // Provide the initial state as proof
        State[] memory proofs = new State[](1);
        proofs[0] = initialState;

        // Verify that the transition is rejected
        bool valid = counter.adjudicate(channel, firstState, proofs);
        assertFalse(valid, "First state signed by incorrect participant should be rejected");
    }

    function test_adjudicate_firstState_revert_whenWrongIntent() public {
        // Create initial state with target 10
        State memory initialState = _createInitialState(10);
        initialState.sigs = new bytes[](2);
        initialState.sigs[HOST] = _signState(initialState, hostPrivateKey);
        initialState.sigs[GUEST] = _signState(initialState, guestPrivateKey);

        // Create first state with incorrect intent
        State memory firstState = _createCounterState(10, 1);
        firstState.intent = StateIntent.FINALIZE; // Wrong intent for first operational state
        firstState.sigs = new bytes[](1);
        firstState.sigs[0] = _signState(firstState, hostPrivateKey);

        // Provide the initial state as proof
        State[] memory proofs = new State[](1);
        proofs[0] = initialState;

        // Verify that the transition is rejected
        bool valid = counter.adjudicate(channel, firstState, proofs);
        assertFalse(valid, "State with incorrect intent should be rejected");
    }

    // -------------------- LATER STATE TRANSITION TESTS --------------------

    function test_adjudicate_laterState_valid() public {
        // Create state 1
        State memory state1 = _createCounterState(10, 1);
        state1.sigs = new bytes[](1);
        state1.sigs[0] = _signState(state1, hostPrivateKey);

        // Create state 2 (signed by guest)
        State memory state2 = _createCounterState(10, 2);
        state2.sigs = new bytes[](1);
        state2.sigs[0] = _signState(state2, guestPrivateKey);

        // Provide state 1 as proof for state 2
        State[] memory proofs = new State[](1);
        proofs[0] = state1;

        // Verify that the transition is valid
        bool valid = counter.adjudicate(channel, state2, proofs);
        assertTrue(valid, "Valid state transition from 1 to 2 should be accepted");
    }

    function test_adjudicate_laterState_revert_whenIncorrectVersionIncrement() public {
        // Create state 1
        State memory state1 = _createCounterState(10, 1);
        state1.sigs = new bytes[](1);
        state1.sigs[0] = _signState(state1, hostPrivateKey);

        // Create state 3 (skipping version 2)
        State memory state3 = _createCounterState(10, 3);
        state3.sigs = new bytes[](1);
        state3.sigs[0] = _signState(state3, guestPrivateKey);

        // Provide state 1 as proof for state 3
        State[] memory proofs = new State[](1);
        proofs[0] = state1;

        // Verify that the transition is rejected
        bool valid = counter.adjudicate(channel, state3, proofs);
        assertFalse(valid, "State with non-sequential version should be rejected");
    }

    function test_adjudicate_laterState_revert_whenTargetChanged() public {
        // Create state 1 with target 10
        State memory state1 = _createCounterState(10, 1);
        state1.sigs = new bytes[](1);
        state1.sigs[0] = _signState(state1, hostPrivateKey);

        // Create state 2 with different target
        State memory state2 = _createCounterState(15, 2);
        state2.sigs = new bytes[](1);
        state2.sigs[0] = _signState(state2, guestPrivateKey);

        // Provide state 1 as proof for state 2
        State[] memory proofs = new State[](1);
        proofs[0] = state1;

        // Verify that the transition is rejected
        bool valid = counter.adjudicate(channel, state2, proofs);
        assertFalse(valid, "State with changed target should be rejected");
    }

    function test_adjudicate_laterState_revert_whenAllocationSumChanged() public {
        // Create state 1
        State memory state1 = _createCounterState(10, 1);
        state1.sigs = new bytes[](1);
        state1.sigs[0] = _signState(state1, hostPrivateKey);

        // Create state 2 with different allocation sums
        State memory state2 = _createCounterState(10, 2);
        state2.allocations[HOST].amount = 120; // Changed from 100
        state2.allocations[GUEST].amount = 100;
        state2.sigs = new bytes[](1);
        state2.sigs[0] = _signState(state2, guestPrivateKey);

        // Provide state 1 as proof for state 2
        State[] memory proofs = new State[](1);
        proofs[0] = state1;

        // Verify that the transition is rejected
        bool valid = counter.adjudicate(channel, state2, proofs);
        assertFalse(valid, "State with changed allocation sum should be rejected");
    }

    function test_adjudicate_laterState_revert_whenWrongSigner() public {
        // Create state 1
        State memory state1 = _createCounterState(10, 1);
        state1.sigs = new bytes[](1);
        state1.sigs[0] = _signState(state1, hostPrivateKey);

        // Create state 2 signed by host (should be guest)
        State memory state2 = _createCounterState(10, 2);
        state2.sigs = new bytes[](1);
        state2.sigs[0] = _signState(state2, hostPrivateKey); // Host signs again instead of guest

        // Provide state 1 as proof for state 2
        State[] memory proofs = new State[](1);
        proofs[0] = state1;

        // Verify that the transition is rejected
        bool valid = counter.adjudicate(channel, state2, proofs);
        assertFalse(valid, "State signed by incorrect participant should be rejected");
    }

    function test_adjudicate_revert_whenNoStateProof() public {
        // Create state 2 without providing a proof
        State memory state2 = _createCounterState(10, 2);
        state2.sigs = new bytes[](1);
        state2.sigs[0] = _signState(state2, guestPrivateKey);

        // Provide empty proofs array
        State[] memory emptyProofs = new State[](0);

        // Verify that the adjudication is rejected
        bool valid = counter.adjudicate(channel, state2, emptyProofs);
        assertFalse(valid, "Non-initial state without proof should be rejected");
    }

    function test_adjudicate_revert_whenTooManyProofs() public {
        // Create state 1
        State memory state1 = _createCounterState(10, 1);
        state1.sigs = new bytes[](1);
        state1.sigs[0] = _signState(state1, hostPrivateKey);

        // Create state 2
        State memory state2 = _createCounterState(10, 2);
        state2.sigs = new bytes[](1);
        state2.sigs[0] = _signState(state2, guestPrivateKey);

        // Provide both state 0 and state 1 as proofs (too many)
        State[] memory tooManyProofs = new State[](2);
        tooManyProofs[0] = _createInitialState(10);
        tooManyProofs[1] = state1;

        // Verify that the adjudication is rejected
        bool valid = counter.adjudicate(channel, state2, tooManyProofs);
        assertFalse(valid, "State with too many proofs should be rejected");
    }

    // -------------------- RESIZE STATE TRANSITION TESTS --------------------

    function test_adjudicate_afterResize_valid() public {
        // Create state 1
        State memory state1 = _createCounterState(10, 1);
        state1.sigs = new bytes[](1);
        state1.sigs[0] = _signState(state1, hostPrivateKey);

        // Create resize state 2
        int256[] memory resizeAmounts = new int256[](2);
        resizeAmounts[0] = 20;
        resizeAmounts[1] = -20;
        State memory resizeState = _createResizeState(10, 2, resizeAmounts);
        resizeState.sigs = new bytes[](1);
        resizeState.sigs[0] = _signState(resizeState, guestPrivateKey);

        // Create state 3 after resize (valid operation state)
        State memory state3 = _createCounterState(10, 3);
        state3.sigs = new bytes[](1);
        state3.sigs[0] = _signState(state3, hostPrivateKey);

        // Provide resize state as proof for state 3
        State[] memory proofs = new State[](1);
        proofs[0] = resizeState;

        // Verify that the transition is valid
        bool valid = counter.adjudicate(channel, state3, proofs);
        assertTrue(valid, "Valid state transition after resize should be accepted");
    }

    function test_adjudicate_afterResize_revert_whenTargetChanged() public {
        // Create resize state 2
        int256[] memory resizeAmounts = new int256[](2);
        resizeAmounts[0] = 20;
        resizeAmounts[1] = -20;
        State memory resizeState = _createResizeState(10, 2, resizeAmounts);
        resizeState.sigs = new bytes[](1);
        resizeState.sigs[0] = _signState(resizeState, guestPrivateKey);

        // Create state 3 after resize with different target
        State memory state3 = _createCounterState(15, 3); // Changed target from 10 to 15
        state3.sigs = new bytes[](1);
        state3.sigs[0] = _signState(state3, hostPrivateKey);

        // Provide resize state as proof for state 3
        State[] memory proofs = new State[](1);
        proofs[0] = resizeState;

        // Verify that the transition is rejected
        bool valid = counter.adjudicate(channel, state3, proofs);
        assertFalse(valid, "State with changed target after resize should be rejected");
    }
}
