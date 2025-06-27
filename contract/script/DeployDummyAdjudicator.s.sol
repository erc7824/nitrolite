// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import {Script, console} from "forge-std/Script.sol";
import {Dummy} from "../src/adjudicators/Dummy.sol";

contract DeployDummyAdjudicatorScript is Script {
    Dummy public dummy;

    function setUp() public {}

    function run(uint32 deployerIndex, string memory mnemonic) public {
        (address gasProvider,) = deriveRememberKey(mnemonic, 0);
        vm.startBroadcast(gasProvider);

        (address deployer,) = deriveRememberKey(mnemonic, deployerIndex);

        payable(deployer).transfer(0.1 ether);

        vm.stopBroadcast();
        vm.startBroadcast(deployer);

        dummy = new Dummy();

        vm.stopBroadcast();

        console.log("Deployed Dummy Adjudicator at:", address(dummy));
    }
}
