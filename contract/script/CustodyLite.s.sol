// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import {Script, console} from "forge-std/Script.sol";
import {CustodyLite} from "../src/CustodyLite.sol";

contract CustodyLiteScript is Script {
    CustodyLite public custodyLite;

    function setUp() public {}

    function run() public {
        vm.startBroadcast();

        custodyLite = new CustodyLite();

        vm.stopBroadcast();
    }
}