// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ECDSA} from "lib/openzeppelin-contracts/contracts/utils/cryptography/ECDSA.sol";
import {MessageHashUtils} from "lib/openzeppelin-contracts/contracts/utils/cryptography/MessageHashUtils.sol";

import {Definition, CrossChainState, State} from "./interfaces/Types.sol";

library Utils {
    using ECDSA for bytes32;
    using MessageHashUtils for bytes;

    function getChannelId(Definition memory def) internal pure returns (bytes32) {
        return keccak256(abi.encode(def));
    }

    // ========== Cross-Chain State ==========

    function pack(CrossChainState memory ccs, bytes32 channelId) internal pure returns (bytes memory) {
        return abi.encode(
            channelId,
            ccs.version,
            ccs.intent,
            ccs.homeState,
            ccs.nonHomeState
            // omit signatures
        );
    }

    // supports only EIP-191 signatures for now
    function validateSignatures(
        CrossChainState memory ccs,
        bytes32 channelId,
        address user,
        address node
    ) internal pure {
        bytes32 ethSignedHash = pack(ccs, channelId).toEthSignedMessageHash();

        address recoveredUser = ethSignedHash.recover(ccs.userSig);
        address recoveredNode = ethSignedHash.recover(ccs.nodeSig);

        require(recoveredUser == user, "invalid user signature");
        require(recoveredNode == node, "invalid node signature");
    }

    function validateChallengerSignature(
        CrossChainState memory ccs,
        bytes32 channelId,
        bytes memory challengerSig,
        address user,
        address node
    ) internal pure {
        bytes memory packedChallengeState = abi.encodePacked(pack(ccs, channelId), "challenge");
        bytes32 ethSignedHash = packedChallengeState.toEthSignedMessageHash();

        address recoveredChallenger = ethSignedHash.recover(challengerSig);

        require(recoveredChallenger == user || recoveredChallenger == node, "challenger must be node or user");
    }

    // ========== State ==========

    function isEmpty(State memory state) internal pure returns (bool) {
        return state.chainId == 0;
    }

}
