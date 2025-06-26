// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import {Script, console} from "forge-std/Script.sol";
import {Custody} from "../src/Custody.sol";

contract DeployCustodyScript is Script {
    Custody public custody;

    function setUp() public {}

    function run(uint32 deployerIndex, string memory mnemonic) public {
        (address gasProvider,) = deriveRememberKey(mnemonic, 0);
        vm.startBroadcast(gasProvider);

        (address deployer,) = deriveRememberKey(mnemonic, deployerIndex);

        payable(deployer).transfer(0.1 ether);

        vm.stopBroadcast();
        vm.startBroadcast(deployer);

        custody = new Custody();

        vm.stopBroadcast();

        console.log("Deployed Custody at:", address(custody));
    }
}
