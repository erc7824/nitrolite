// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import {Script, console} from "forge-std/Script.sol";
import {IERC20} from "@openzeppelin/contracts/interfaces/IERC20.sol";

import {Custody} from "../src/Custody.sol";
import {Channel, State, Allocation, Signature, ChannelStatus, StateIntent, Amount} from "../src/interfaces/Types.sol";

contract DeployCustodyScript is Script {
    Custody public custody;

    function setUp() public {}

    function run() public {
        vm.broadcast();
        custody = new Custody();
    }
}
