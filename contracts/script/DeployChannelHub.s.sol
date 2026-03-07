// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Script} from "forge-std/Script.sol";
import {console} from "forge-std/console.sol";

import {ChannelHub} from "../src/ChannelHub.sol";
import {ISignatureValidator} from "../src/interfaces/ISignatureValidator.sol";
import {ECDSAValidator} from "../src/sigValidators/ECDSAValidator.sol";

/**
 * @title DeployChannelHub
 * @notice Forge script to deploy engine libraries and ChannelHub
 * @dev Foundry automatically deploys unlinked libraries (ChannelEngine,
 *      EscrowDepositEngine, EscrowWithdrawalEngine) before ChannelHub in the
 *      broadcast batch. Their addresses appear in the broadcast JSON output.
 *
 * Usage:
 *   DEFAULT_VALIDATOR_ADDR=<addr>   Address of an already-deployed ISignatureValidator.
 *                                   Leave unset to deploy a fresh ECDSAValidator.
 *
 *   forge script script/DeployChannelHub.s.sol:DeployChannelHub \
 *     --rpc-url <RPC_URL> \
 *     --private-key <DEPLOYER_PK> \
 *     --broadcast \
 *     [-vvvv]
 *
 */
contract DeployChannelHub is Script {
    function run() external {
        // Optional: reuse an existing validator or deploy a fresh ECDSAValidator
        address defaultValidatorAddr = vm.envOr("DEFAULT_VALIDATOR_ADDR", address(0));
        run(defaultValidatorAddr);
    }

    function run(address defaultValidatorAddr) public {
        // msg.sender is set by Foundry to the address derived from --private-key
        address deployer = msg.sender;

        console.log("=== Deploy ChannelHub ===");
        console.log("Deployer:          ", deployer);
        console.log("Chain ID:          ", block.chainid);

        // ----------------------------------------------------------------
        // Predict addresses for informational logging.
        // Foundry auto-deploys unlinked libraries in the order they appear
        // in the dependency graph before broadcasting the script's own txs.
        // The exact deployment order follows the nonce sequence starting at
        // the deployer's current nonce.
        // ----------------------------------------------------------------
        uint64 nonce = vm.getNonce(deployer);

        bool deployValidator = defaultValidatorAddr == address(0);
        if (deployValidator) {
            console.log("ECDSAValidator:    ", vm.computeCreateAddress(deployer, nonce));
            nonce++;
        } else {
            console.log("DefaultValidator:  ", defaultValidatorAddr);
        }

        // Library deployment slots (auto-managed by Foundry)
        console.log("ChannelEngine:     ", vm.computeCreateAddress(deployer, nonce));
        console.log("EscrowDepositEng:  ", vm.computeCreateAddress(deployer, nonce + 1));
        console.log("EscrowWithdrawEng: ", vm.computeCreateAddress(deployer, nonce + 2));
        console.log("ChannelHub:        ", vm.computeCreateAddress(deployer, nonce + 3));

        vm.startBroadcast();

        // 1. Deploy default signature validator if not provided
        if (deployValidator) {
            ECDSAValidator ecdsaValidator = new ECDSAValidator();
            defaultValidatorAddr = address(ecdsaValidator);
            console.log("Deployed ECDSAValidator:", defaultValidatorAddr);
        }

        // 2. Deploy ChannelHub.
        //    Foundry detects unlinked library references (ChannelEngine,
        //    EscrowDepositEngine, EscrowWithdrawalEngine) and inserts their
        //    deployment transactions before this one in the broadcast batch.
        ChannelHub hub = new ChannelHub(ISignatureValidator(defaultValidatorAddr));

        vm.stopBroadcast();

        // ----------------------------------------------------------------
        // Summary
        // ----------------------------------------------------------------
        console.log("");
        console.log("=== Deployment complete ===");
        console.log("DefaultSigValidator:", defaultValidatorAddr);
        console.log("ChannelHub:         ", address(hub));
        console.log("(Library addresses are logged above and in the broadcast JSON)");
    }
}
