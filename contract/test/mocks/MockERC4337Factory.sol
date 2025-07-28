// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {MockFlagERC1271} from "./MockFlagERC1271.sol";

contract MockERC4337Factory {
    event AccountCreated(address indexed account, bytes32 salt, bool flag);

    function createAccount(bool flag, bytes32 salt) external returns (address) {
        bytes memory bytecode = abi.encodePacked(type(MockFlagERC1271).creationCode, abi.encode(flag));

        address account;
        assembly {
            account := create2(0, add(bytecode, 0x20), mload(bytecode), salt)
        }

        require(account != address(0), "Account creation failed");

        emit AccountCreated(account, salt, flag);
        return account;
    }

    function getAddress(bool flag, bytes32 salt) external view returns (address) {
        bytes memory bytecode = abi.encodePacked(type(MockFlagERC1271).creationCode, abi.encode(flag));

        bytes32 hash = keccak256(abi.encodePacked(bytes1(0xff), address(this), salt, keccak256(bytecode)));

        return address(uint160(uint256(hash)));
    }
}
