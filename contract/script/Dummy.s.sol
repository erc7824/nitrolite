// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import {Script, console} from "forge-std/Script.sol";
import {Dummy} from "../src/adjudicators/Dummy.sol";

contract DummyScript is Script {
    Dummy public dummy;

    function setUp() public {}

    function run() public {
        vm.startBroadcast();

        dummy = new Dummy();

        vm.stopBroadcast();
    }
}