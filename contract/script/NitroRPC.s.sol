// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import {Script, console} from "forge-std/Script.sol";
import {NitroRPC} from "../src/adjudicators/NitroRPC.sol";

contract NitroRPCScript is Script {
    NitroRPC public nitroRPC;

    function setUp() public {}

    function run() public {
        vm.startBroadcast();

        nitroRPC = new NitroRPC();

        vm.stopBroadcast();
    }
}
