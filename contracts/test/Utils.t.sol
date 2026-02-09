// SPDX-License-Identifier: MIT
pragma solidity 0.8.30;

import {Test, console} from "lib/forge-std/src/Test.sol";
import {Utils} from "../src/Utils.sol";
import {ChannelDefinition, State, Ledger, StateIntent} from "../src/interfaces/Types.sol";
import {ISignatureValidator} from "../src/interfaces/Types.sol";

contract UtilsTest is Test {
    function test_channelId_forDifferentVersions_differ() public pure {
        ChannelDefinition memory def = ChannelDefinition({
            challengeDuration: 86400,
            user: 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045,
            node: 0x435d4B6b68e1083Cc0835D1F971C4739204C1d2a,
            nonce: 42,
            signatureValidator: ISignatureValidator(address(123)),
            metadata: 0x13730b0d8e1bdbdc000000000000000000000000000000000000000000000000
        });

        bytes32 channelIdV1 = Utils.getChannelId(def, 1);
        bytes32 channelIdV2 = Utils.getChannelId(def, 2);
        bytes32 channelIdV255 = Utils.getChannelId(def, 255);

        // Channel IDs must differ for different versions
        assert(channelIdV1 != channelIdV2);
        assert(channelIdV1 != channelIdV255);
        assert(channelIdV2 != channelIdV255);

        // First byte should match the version
        assert(uint8(uint256(channelIdV1) >> 248) == 1);
        assert(uint8(uint256(channelIdV2) >> 248) == 2);
        assert(uint8(uint256(channelIdV255) >> 248) == 255);

        // All other bytes should be the same (derived from the same base hash)
        bytes32 mask = 0x00ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff;
        assert(channelIdV1 & mask == channelIdV2 & mask);
        assert(channelIdV1 & mask == channelIdV255 & mask);
    }

    // ========== Packing Tests ==========

    function test_log_packingState() public pure {
        Ledger memory homeLedger = Ledger({
            chainId: 42,
            token: 0x90b7E285ab6cf4e3A2487669dba3E339dB8a3320,
            decimals: 8,
            userAllocation: 300_000_000,
            userNetFlow: 200_000_001,
            nodeAllocation: 0,
            nodeNetFlow: -99_999_999
        });

        Ledger memory nonHomeLedger = Ledger({
            chainId: 4242,
            token: 0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2,
            decimals: 14,
            userAllocation: 300_000_000_000_000,
            userNetFlow: 200_000_001_000_000,
            nodeAllocation: 0,
            nodeNetFlow: -99_999_999_000_000
        });

        State memory state = State({
            version: 24,
            intent: StateIntent.FINALIZE_ESCROW_WITHDRAWAL,
            metadata: hex"dbf80153432e3e0c221112f69a7d20e80980ee5bc48b5684d3b47a6cb75192bd",
            homeState: homeLedger,
            nonHomeState: nonHomeLedger,
            userSig: hex"36954bf8e670eba9044f0f9eccd3c36871b12ca209f033190bbf378747906d697a521dd4a05faa0ddf3183900df6191ee276055d6d8bf39d8eb8a27e71d2b8b11b",
            nodeSig: hex"2c0648f47bbf3d580dd56acf74662d7d984b6f4abefa1a02ffbd561e0e463761462984ac6dbedac5f679ee29ef58bc9db7f0ac7792d9992832af99a9950039a21b"
        });

        bytes32 channelId = 0x3e9dd25a843e3a234c278c6f3fab3983949e2404b276cacb3c47ada06e00f74b;

        bytes memory packed = Utils.pack(state, channelId);

        // 0x3e9dd25a843e3a234c278c6f3fab3983949e2404b276cacb3c47ada06e00f74b0000000000000000000000000000000000000000000000000000000000000018000000000000000000000000000000000000000000000000000000000000000002af862655bc9c16cbd4753515bd77f3c33d1e3a68c9d4995f6e6f72c01e0eb0000000000000000000000000000000000000000000000000000000000000002a00000000000000000000000090b7e285ab6cf4e3a2487669dba3e339db8a332000000000000000000000000000000000000000000000000000000000000004120000000000000000000000000000000000000000000000000000000000002c460000000000000000000000000000000000000000000000000000000000009de8ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffebe90000000000000000000000000000000000000000000000000000000000001092000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc200000000000000000000000000000000000000000000000000000000000004d2fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffde10000000000000000000000000000000000000000000000000000000000000237fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc86
        console.logBytes(packed);
    }

    function test_log_packingState_emptyNonHome() public pure {
        Ledger memory homeLedger = Ledger({
            chainId: 42,
            token: 0x90b7E285ab6cf4e3A2487669dba3E339dB8a3320,
            decimals: 8,
            userAllocation: 300_000_000,
            userNetFlow: 200_000_001,
            nodeAllocation: 0,
            nodeNetFlow: -99_999_999
        });

        // Empty nonHomeLedger (no escrow)
        Ledger memory nonHomeLedger = Ledger({
            chainId: 0,
            token: address(0),
            decimals: 0,
            userAllocation: 0,
            userNetFlow: 0,
            nodeAllocation: 0,
            nodeNetFlow: 0
        });

        State memory state = State({
            version: 24,
            intent: StateIntent.DEPOSIT,
            metadata: hex"6d621872dd3d14fe6f6ddb415d586e62fb584ffda861ac379bf0d0a0e6410bd6",
            homeState: homeLedger,
            nonHomeState: nonHomeLedger,
            userSig: hex"36954bf8e670eba9044f0f9eccd3c36871b12ca209f033190bbf378747906d697a521dd4a05faa0ddf3183900df6191ee276055d6d8bf39d8eb8a27e71d2b8b11b",
            nodeSig: hex"2c0648f47bbf3d580dd56acf74662d7d984b6f4abefa1a02ffbd561e0e463761462984ac6dbedac5f679ee29ef58bc9db7f0ac7792d9992832af99a9950039a21b"
        });

        bytes32 channelId = 0x3e9dd25a843e3a234c278c6f3fab3983949e2404b276cacb3c47ada06e00f74b;

        bytes memory packed = Utils.pack(state, channelId);

        console.log("Packed state with empty nonHomeLedger:");
        console.logBytes(packed);
    }

    function test_log_calculateChannelId() public pure {
        // Generate metadata from asset: first 8 bytes of keccak256("ether"), padded to 32 bytes
        bytes32 assetHash = keccak256("ether");
        bytes32 metadata;
        assembly {
            metadata := shl(192, shr(192, assetHash))
        }

        // 0x13730b0d8e1bdbdc293b62ba010b1eede56b412ea2980defabe3d0b6c7844c3a
        console.logBytes32(assetHash);
        // 0x13730b0d8e1bdbdc000000000000000000000000000000000000000000000000
        console.logBytes32(metadata);

        ChannelDefinition memory def = ChannelDefinition({
            challengeDuration: 86400,
            user: 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045,
            node: 0x435d4B6b68e1083Cc0835D1F971C4739204C1d2a,
            nonce: 42,
            signatureValidator: ISignatureValidator(address(123)),
            metadata: metadata
        });

        bytes32 channelId = Utils.getChannelId(def, 1);

        // 0x01827da2e3e6aac5385f51934491b6b1bc338a4b4222860943b1d1e6519659ee
        console.logBytes32(channelId);
    }

    function test_log_calculateEscrowId() public pure {
        bytes32 channelId = 0xeac2bed767671a8ab77527e1e2fff00bb2e62de5467d9ba3a4105dad5c6e3d66;
        uint64 version = 42;

        bytes32 escrowId = Utils.getEscrowId(channelId, version);

        // 0xe4d925dcf63add647f25c757d6ff0e74ba31401da91d8c7bafa4846c97a92ac2
        console.logBytes32(escrowId);
    }
}
