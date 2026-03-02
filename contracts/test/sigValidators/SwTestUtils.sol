// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {IERC1271} from "@openzeppelin/contracts/interfaces/IERC1271.sol";

/**
 * @title MockSmartWallet
 * @notice Mock smart contract wallet that implements ERC-1271
 */
contract MockSmartWallet is IERC1271 {
    bytes4 private constant ERC1271_MAGIC_VALUE = 0x1626ba7e;

    address public owner;
    bool public shouldReturnValid;

    constructor(address _owner) {
        owner = _owner;
        shouldReturnValid = true;
    }

    function isValidSignature(bytes32 hash, bytes memory signature) external view override returns (bytes4) {
        if (!shouldReturnValid) {
            return 0xffffffff;
        }

        // Check signature length
        if (signature.length != 65) {
            return 0xffffffff;
        }

        // Recover signer from signature
        (uint8 v, bytes32 r, bytes32 s) = _splitSignature(signature);
        address recovered = ecrecover(hash, v, r, s);

        if (recovered == owner) {
            return ERC1271_MAGIC_VALUE;
        }

        return 0xffffffff;
    }

    function setValidation(bool _shouldReturnValid) external {
        shouldReturnValid = _shouldReturnValid;
    }

    function _splitSignature(bytes memory sig) private pure returns (uint8 v, bytes32 r, bytes32 s) {
        assembly {
            r := mload(add(sig, 32))
            s := mload(add(sig, 64))
            v := byte(0, mload(add(sig, 96)))
        }
    }
}

/**
 * @title MockSmartWalletFactory
 * @notice Mock factory for deploying smart wallets (used for ERC-6492 testing)
 */
contract MockSmartWalletFactory {
    function deploy(address owner, bytes32 salt) external payable returns (address) {
        MockSmartWallet wallet = new MockSmartWallet{salt: salt}(owner);
        return address(wallet);
    }

    function getAddress(address owner, bytes32 salt) external view returns (address) {
        bytes32 hash = keccak256(
            abi.encodePacked(
                bytes1(0xff),
                address(this),
                salt,
                keccak256(abi.encodePacked(type(MockSmartWallet).creationCode, abi.encode(owner)))
            )
        );
        return address(uint160(uint256(hash)));
    }
}

/**
 * @title FailingFactory
 * @notice Mock factory that always fails deployment
 */
contract FailingFactory {
    function deploy(address, bytes32) external pure {
        revert("Deployment always fails");
    }
}

/**
 * @title EmptyFactory
 * @notice Mock factory that succeeds but doesn't actually deploy anything
 */
contract EmptyFactory {
    function deploy(address, bytes32) external pure returns (address) {
        // Succeeds but doesn't deploy anything
        return address(0);
    }
}

bytes32 constant ERC6492_MAGIC_SUFFIX = 0x6492649264926492649264926492649264926492649264926492649264926492;

library SwTestUtils {
    function createERC6492Signature(
        address factoryAddress,
        bytes memory factoryCalldata,
        bytes memory signature
    ) internal pure returns (bytes memory) {
        bytes memory wrappedData = abi.encode(factoryAddress, factoryCalldata, signature);
        return abi.encodePacked(wrappedData, ERC6492_MAGIC_SUFFIX);
    }
}
