// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import {Script, console} from "forge-std/Script.sol";
import {IERC20} from "@openzeppelin/contracts/interfaces/IERC20.sol";

import {Custody} from "../src/Custody.sol";
import {Dummy} from "../src/adjudicators/Dummy.sol";
import {TestERC20} from "../test/TestERC20.sol";
import {Channel, State, Allocation, Signature, ChannelStatus, StateIntent, Amount} from "../src/interfaces/Types.sol";

contract DeployScript is Script {
    bytes32 public constant CUSTODY_SALT = keccak256("NITROLITE_CUSTODY_V1");
    bytes32 public constant ADJUDICATOR_SALT =
        keccak256("NITROLITE_ADJUDICATOR_V1");
    bytes32 public constant TOKEN_SALT = keccak256("NITROLITE_TEST_TOKEN_V1");

    Custody public custody;
    Dummy public adjudicator;
    TestERC20 public testToken;

    address public custodyAddress;
    address public adjudicatorAddress;
    address public testTokenAddress;

    function setUp() public {}

    function run() public {
        console.log("Starting deployment of Nitrolite contracts...");
        console.log("Deployer address:", msg.sender);
        console.log("Chain ID:", block.chainid);

        vm.startBroadcast();

        deployContracts();

        logDeployedAddresses();

        setupContracts();

        vm.stopBroadcast();

        console.log("Deployment completed successfully!");
    }

    function deployContracts() internal {
        console.log("Deploying contracts with deterministic addresses...");

        custody = new Custody{salt: CUSTODY_SALT}();
        custodyAddress = address(custody);
        console.log("Custody deployed at:", custodyAddress);

        adjudicator = new Dummy{salt: ADJUDICATOR_SALT}();
        adjudicatorAddress = address(adjudicator);
        console.log("Adjudicator deployed at:", adjudicatorAddress);

        testToken = new TestERC20{salt: TOKEN_SALT}(
            "Nitrolite Test Token",
            "NTL",
            18,
            type(uint256).max // Max supply for testing
        );
        testTokenAddress = address(testToken);
        console.log("Test Token deployed at:", testTokenAddress);
    }

    function logDeployedAddresses() internal view {
        console.log("=== DEPLOYED CONTRACTS ===");
        console.log("Custody Address:", custodyAddress);
        console.log("Adjudicator Address:", adjudicatorAddress);
        console.log("Test Token Address:", testTokenAddress);
        console.log("========================");
    }

    function setupContracts() internal {
        console.log("Setting up contracts...");

        // Standard Anvil test accounts
        address alice = 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266;
        address bob = 0x70997970C51812dc3A010C7d01b50e0d17dc79C8;
        address charlie = 0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC;

        uint256 initialBalance = 1000 * 10 ** 18; // 1000 tokens

        testToken.mint(alice, initialBalance);
        testToken.mint(bob, initialBalance);
        testToken.mint(charlie, initialBalance);

        console.log(
            "Minted",
            initialBalance / 10 ** 18,
            "tokens to test accounts"
        );
    }

    function getDeploymentAddresses()
        external
        view
        returns (
            address predictedCustody,
            address predictedAdjudicator,
            address predictedToken
        )
    {
        predictedCustody = computeCreate2AddressCustom(
            CUSTODY_SALT,
            keccak256(type(Custody).creationCode)
        );

        predictedAdjudicator = computeCreate2AddressCustom(
            ADJUDICATOR_SALT,
            keccak256(type(Dummy).creationCode)
        );

        bytes memory tokenCreationCode = abi.encodePacked(
            type(TestERC20).creationCode,
            abi.encode("Nitrolite Test Token", "NTL", 18, type(uint256).max)
        );

        predictedToken = computeCreate2AddressCustom(
            TOKEN_SALT,
            keccak256(tokenCreationCode)
        );
    }

    function computeCreate2AddressCustom(
        bytes32 salt,
        bytes32 bytecodeHash
    ) internal view returns (address) {
        return
            address(
                uint160(
                    uint256(
                        keccak256(
                            abi.encodePacked(
                                bytes1(0xff),
                                address(this),
                                salt,
                                bytecodeHash
                            )
                        )
                    )
                )
            );
    }
}
