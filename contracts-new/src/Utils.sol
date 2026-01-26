// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {ECDSA} from "lib/openzeppelin-contracts/contracts/utils/cryptography/ECDSA.sol";
import {MessageHashUtils} from "lib/openzeppelin-contracts/contracts/utils/cryptography/MessageHashUtils.sol";

import {ChannelDefinition, State, Ledger} from "./interfaces/Types.sol";

library Utils {
    using ECDSA for bytes32;
    using MessageHashUtils for bytes;

    function getChannelId(ChannelDefinition memory def) internal pure returns (bytes32) {
        return keccak256(abi.encode(def));
    }

    function getEscrowId(bytes32 channelId, uint64 version) internal pure returns (bytes32) {
        // "channelId, (state-)version" pair is unique as long as participants do not reuse versions
        return keccak256(abi.encode(channelId, version));
    }

    // ========== Cross-Chain State ==========

    function pack(State memory ccs, bytes32 channelId) internal pure returns (bytes memory) {
        return abi.encode(
            channelId,
            ccs.version,
            ccs.intent,
            ccs.metadata,
            ccs.homeState,
            ccs.nonHomeState
            // omit signatures
        );
    }

    // supports only EIP-191 signatures for now
    function validateSignatures(State memory ccs, bytes32 channelId, address user, address node) internal pure {
        bytes32 ethSignedHash = pack(ccs, channelId).toEthSignedMessageHash();

        address recoveredUser = ethSignedHash.recover(ccs.userSig);
        address recoveredNode = ethSignedHash.recover(ccs.nodeSig);

        require(recoveredUser == user, "invalid user signature");
        require(recoveredNode == node, "invalid node signature");
    }

    // supports only EIP-191 signatures for now
    function validateNodeSignature(State memory ccs, bytes32 channelId, address node) internal pure {
        bytes32 ethSignedHash = pack(ccs, channelId).toEthSignedMessageHash();
        address recoveredNode = ethSignedHash.recover(ccs.nodeSig);
        require(recoveredNode == node, "invalid node signature");
    }

    function validateChallengerSignature(
        State memory ccs,
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

    // ========== Ledger ==========

    function isEmpty(Ledger memory state) internal pure returns (bool) {
        return state.chainId == 0;
    }
}
