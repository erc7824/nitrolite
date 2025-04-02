// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import {Script, console} from "forge-std/Script.sol";
import {Custody} from "../src/Custody.sol";

contract CustodyScript is Script {
    Custody public custody;

    function setUp() public {}

    function run() public {
        vm.startBroadcast();

        custody = new Custody();

        vm.stopBroadcast();
    }
}
