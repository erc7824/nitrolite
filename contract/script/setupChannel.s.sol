// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import {Script, console} from "forge-std/Script.sol";
import {IERC20} from "@openzeppelin/contracts/interfaces/IERC20.sol";

import {Custody} from "../src/Custody.sol";
import {Utils} from "../src/Utils.sol";
import {Channel, State, Allocation, ChannelStatus, StateIntent, Amount} from "../src/interfaces/Types.sol";
import {TestUtils} from "../test/TestUtils.sol";

contract SetupChannelScript is Script {
    uint64 constant CHALLENGE_DURATION = 1 days;
    uint64 constant NONCE = 0;
    uint256 constant CHANNEL_DEPOSIT_AMOUNT = 10;

    // TODO: move this out of the file
    uint256 constant USER_SESSION_KEY = 66797303920176115455777297298802822546720088673606171092140646898562007298987;
    address constant USER_SESSION_KEY_ADDRESS = 0xfAea0C9Dc921b8040DFa5Dd617014c957cf46455;

    function setUp() public {}

    function run(
        string memory mnemonic,
        address custody,
        address adjudicator,
        address token,
        uint256 custodyDepositAmount
    ) public {
        address[] memory addresses = new address[](3);

        for (uint32 i = 0; i < 3; i++) {
            (address wallet,) = deriveRememberKey(mnemonic, i);
            addresses[i] = wallet;

            vm.startBroadcast(wallet);
            IERC20(token).approve(address(custody), type(uint256).max);
            Custody(custody).deposit(wallet, token, custodyDepositAmount);
            vm.stopBroadcast();
        }

        Channel memory channel = createChannel(USER_SESSION_KEY_ADDRESS, addresses[0], adjudicator);
        State memory initialState = createInitialState(token, USER_SESSION_KEY_ADDRESS, addresses[0]);
        bytes memory userSig = signState(channel, initialState, USER_SESSION_KEY);
        initialState.sigs = new bytes[](1);
        initialState.sigs[0] = userSig;

        vm.broadcast(addresses[1]);
        Custody(custody).create(channel, initialState);
    }

    function createChannel(address userSessionKey, address broker, address adjudicator)
        internal
        pure
        returns (Channel memory)
    {
        address[] memory participants = new address[](2);
        participants[0] = userSessionKey;
        participants[1] = broker;

        return
            Channel({participants: participants, adjudicator: adjudicator, challenge: CHALLENGE_DURATION, nonce: NONCE});
    }

    function createInitialState(address token, address userSessionKey, address broker)
        internal
        pure
        returns (State memory)
    {
        Allocation[] memory allocations = new Allocation[](2);

        allocations[0] =
            Allocation({destination: userSessionKey, token: address(token), amount: CHANNEL_DEPOSIT_AMOUNT});

        allocations[1] = Allocation({destination: broker, token: address(token), amount: 0});

        return State({
            intent: StateIntent.INITIALIZE,
            version: 0,
            data: bytes(""),
            allocations: allocations,
            sigs: new bytes[](0)
        });
    }

    function signState(Channel memory chan, State memory state, uint256 privateKey)
        internal
        view
        returns (bytes memory)
    {
        bytes memory packedState = Utils.getPackedState(Utils.getChannelId(chan), state);
        return TestUtils.sign(vm, privateKey, packedState);
    }
}
