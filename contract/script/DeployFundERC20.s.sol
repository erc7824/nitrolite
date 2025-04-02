// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {Script, console} from "forge-std/Script.sol";
import {TestERC20} from "../src/test/TestERC20.sol";

contract DeployFundERC20Script is Script {
    TestERC20 public token;

    function setUp() public {}

    function run(string memory mnemonic) public {
        (address deployer,) = deriveRememberKey(mnemonic, 0);
        vm.startBroadcast(deployer);

        token = new TestERC20("Test Token", "TST", 8, type(uint256).max);

        for (uint32 i = 0; i < 15; i++) {
            address mintTo = vm.createWallet(vm.deriveKey(mnemonic, i)).addr;
            token.mint(mintTo, type(uint128).max);
        }

        vm.stopBroadcast();
    }
}
