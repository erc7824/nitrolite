// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

import {IERC1271} from "lib/openzeppelin-contracts/contracts/interfaces/IERC1271.sol";

contract MockFlagERC1271 is IERC1271 {
    bool private _flag;
    bytes4 private constant ERC1271_SUCCESS = 0x1626ba7e;
    bytes4 private constant ERC1271_FAILURE = 0xffffffff;

    constructor(bool flag) {
        _flag = flag;
    }

    function isValidSignature(bytes32, bytes memory) external view override returns (bytes4) {
        return _flag ? ERC1271_SUCCESS : ERC1271_FAILURE;
    }

    function setFlag(bool flag) external {
        _flag = flag;
    }

    function getFlag() external view returns (bool) {
        return _flag;
    }
}
