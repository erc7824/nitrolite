// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import {Test} from "lib/forge-std/src/Test.sol";
import {IERC5267} from "lib/openzeppelin-contracts/contracts/interfaces/IERC5267.sol";
import {Ownable} from "lib/openzeppelin-contracts/contracts/access/Ownable.sol";

import {EIP712AdjudicatorBase} from "../../src/adjudicators/EIP712AdjudicatorBase.sol";
import {TestEIP712Adjudicator} from "../mocks/TestEIP712Adjudicator.sol";
import {MockERC20} from "../mocks/MockERC20.sol";
import {MockEIP712} from "../mocks/MockEIP712.sol";
import {Utils} from "../../src/Utils.sol";

contract EIP712AdjudicatorBaseTest is Test {
    TestEIP712Adjudicator public adjudicator;
    MockERC20 public token;

    address public owner = address(0x1);
    address public notOwner = address(0x2);
    address public channelImpl = address(0x3);
    address public nonContractAddress = address(0x4);
    address public newChannelImpl = address(0x5);

    function setUp() public {
        token = new MockERC20("Test Token", "TEST", 18);
        adjudicator = new TestEIP712Adjudicator(owner, channelImpl);
    }

    function test_constructor_setsValues() public view {
        assertEq(adjudicator.owner(), owner, "Owner should be set correctly");
        assertEq(address(adjudicator.channelImpl()), channelImpl, "Channel implementation should be set correctly");
    }

    function test_getChannelImplDomainSeparator_returnsCorrectCustodyDomainSeparator() public {
        // Deploy a contract that implements EIP712
        MockEIP712 mockCustody = new MockEIP712("Custody", "1.0");
        TestEIP712Adjudicator custodyAdjudicator = new TestEIP712Adjudicator(owner, address(mockCustody));

        bytes32 result = custodyAdjudicator.getChannelImplDomainSeparator();
        bytes32 expectedDomainSeparator = mockCustody.domainSeparator();

        assertEq(result, expectedDomainSeparator, "Domain separator should match expected value");
    }

    function test_getChannelImplDomainSeparator_returnsNoEip712Support_whenContractDoesNotSupportEip712() public {
        // Set MockERC20 as channel impl (doesn't support EIP712)
        vm.prank(owner);
        adjudicator.setChannelImpl(address(token));

        bytes32 result = adjudicator.getChannelImplDomainSeparator();

        assertEq(result, Utils.NO_EIP712_SUPPORT, "Should return NO_EIP712_SUPPORT for non-EIP712 contract");
    }

    function test_getChannelImplDomainSeparator_returnsNoEip712Support_whenAddressIsNotContract() public {
        // Set non-contract address as channel impl
        vm.prank(owner);
        adjudicator.setChannelImpl(nonContractAddress);

        bytes32 result = adjudicator.getChannelImplDomainSeparator();

        assertEq(result, Utils.NO_EIP712_SUPPORT, "Should return NO_EIP712_SUPPORT for non-contract address");
    }

    function test_setChannelImpl_onlyWorksForOwner() public {
        vm.prank(owner);
        adjudicator.setChannelImpl(newChannelImpl);

        assertEq(address(adjudicator.channelImpl()), newChannelImpl, "Channel impl should be updated by owner");
    }

    function test_setChannelImpl_failsForNotOwner() public {
        vm.prank(notOwner);
        vm.expectRevert(abi.encodeWithSelector(Ownable.OwnableUnauthorizedAccount.selector, notOwner));
        adjudicator.setChannelImpl(newChannelImpl);
    }
}
