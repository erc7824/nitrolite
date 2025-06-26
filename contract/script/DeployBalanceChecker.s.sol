// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Script, console} from "forge-std/Script.sol";
import {BalanceChecker} from "../test/BalanceChecker.sol";

contract DeployBalanceCheckerScript is Script {
    BalanceChecker public balanceChecker;

    function setUp() public {}

    function run(uint32 deployerIndex, string memory mnemonic) public {
        (address gasProvider,) = deriveRememberKey(mnemonic, 0);
        vm.startBroadcast(gasProvider);

        (address deployer,) = deriveRememberKey(mnemonic, deployerIndex);

        payable(deployer).transfer(0.1 ether);

        vm.stopBroadcast();
        vm.startBroadcast(deployer);

        balanceChecker = new BalanceChecker();

        vm.stopBroadcast();

        console.log("Deployed BalanceChecker at:", address(balanceChecker));
    }
}
