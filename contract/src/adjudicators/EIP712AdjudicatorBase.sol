// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {IERC5267} from "lib/openzeppelin-contracts/contracts/interfaces/IERC5267.sol";
import {Ownable} from "lib/openzeppelin-contracts/contracts/access/Ownable.sol";
import {Ownable2Step} from "lib/openzeppelin-contracts/contracts/access/Ownable2Step.sol";

import {Utils} from "../Utils.sol";

/**
 * @title EIP712 Adjudicator Base
 * @notice Base contract for EIP712 compliant adjudicators when EIP-712 domain specifies channel implementation contract address as a verifying contract.
 * @dev Contains a link to the channel implementation contract and provides a method to retrieve its domain separator.
 * Channel implementation contract must be ERC-5267 compliant.
 */
abstract contract EIP712AdjudicatorBase is Ownable2Step {
    /// Taken from EIP-712 specification
    /// https://eips.ethereum.org/EIPS/eip-712#definition-of-domainseparator
    bytes32 private constant TYPE_HASH =
        keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)");

    IERC5267 public channelImpl;

    /**
     * @notice Constructor for the EIP712AdjudicatorBase.
     * @param owner The owner of the adjudicator contract.
     * @param channelImpl_ The address of the channel implementation contract.
     */
    constructor(address owner, address channelImpl_) Ownable(owner) {
        channelImpl = IERC5267(channelImpl_);
    }

    /**
     * @notice Returns the domain separator for the channel implementation contract.
     * @return The domain separator as bytes32.
     * @dev This function retrieves the domain separator from the channel implementation contract.
     */
    function getChannelImplDomainSeparator() public view returns (bytes32) {
        if (address(channelImpl).code.length == 0) {
            // NOTE: soft failure, if channel implementation contract is not set or does not exist
            return Utils.NO_EIP712_SUPPORT;
        }

        try channelImpl.eip712Domain() returns (
            bytes1,
            string memory name,
            string memory version,
            uint256 chainId,
            address verifyingContract,
            bytes32,
            uint256[] memory
        ) {
            /// Taken from EIP-712 specification
            /// https://eips.ethereum.org/EIPS/eip-712#definition-of-domainseparator
            return keccak256(
                abi.encode(TYPE_HASH, keccak256(bytes(name)), keccak256(bytes(version)), chainId, verifyingContract)
            );
        } catch {
            // NOTE: soft failure, if channel implementation contract does not support EIP-712
            return Utils.NO_EIP712_SUPPORT;
        }
    }

    /**
     * @notice Sets the channel implementation contract address.
     * @dev Callable only by Owner.
     * @param channelImpl_ The address of the channel implementation contract.
     */
    function setChannelImpl(address channelImpl_) external onlyOwner {
        channelImpl = IERC5267(channelImpl_);
    }
}
