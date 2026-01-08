// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ECDSA} from "lib/openzeppelin-contracts/contracts/utils/cryptography/ECDSA.sol";
import {MessageHashUtils} from "lib/openzeppelin-contracts/contracts/utils/cryptography/MessageHashUtils.sol";

import {Definition, CrossChainState, State} from "./interfaces/Types.sol";

library Utils {
    using ECDSA for bytes32;
    using MessageHashUtils for bytes;

    function getChannelId(Definition memory def) internal pure returns (bytes32) {
        return keccak256(abi.encode(def.challengeDuration, def.participant, def.node, def.nonce));
    }

    // ========== Cross-Chain State ==========

    function pack(CrossChainState memory ccs, bytes32 channelId) internal pure returns (bytes memory) {
        return abi.encode(
            channelId,
            ccs.version,
            ccs.homeChainId,
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
        address participant,
        address node
    ) internal pure {
        bytes32 ethSignedHash = pack(ccs, channelId).toEthSignedMessageHash();

        address recoveredParticipant = ethSignedHash.recover(ccs.participantSig);
        address recoveredNode = ethSignedHash.recover(ccs.nodeSig);

        require(recoveredParticipant == participant, "invalid participant signature");
        require(recoveredNode == node, "invalid node signature");
    }

    // ========== State ==========

    function isEmpty(State memory state) internal pure returns (bool) {
        return state.chainId == 0;
    }

}
