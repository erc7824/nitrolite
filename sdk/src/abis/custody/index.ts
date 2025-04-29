import { Abi } from "abitype";

export const ChannelOpenedEvent = "Opened";
export const ChannelClosedEvent = "ChannelClosed";
export const ChannelChallengedEvent = "Challenged";
export const ChannelCheckpointedEvent = "Checkpointed";
export const ChannelCreatedEvent = "Created";
export const ChannelJoinedEvent = "Joined";

/**
 * ABI for the Custody contract
 * Manages the lifecycle of state channels
 */
export const CustodyAbi: Abi = [
    {
        type: "function",
        name: "challenge",
        inputs: [
            { name: "channelId", type: "bytes32", internalType: "bytes32" },
            {
                name: "candidate",
                type: "tuple",
                internalType: "struct State",
                components: [
                    { name: "intent", type: "uint8", internalType: "enum StateIntent" },
                    { name: "version", type: "uint256", internalType: "uint256" },
                    { name: "data", type: "bytes", internalType: "bytes" },
                    {
                        name: "allocations",
                        type: "tuple[]",
                        internalType: "struct Allocation[]",
                        components: [
                            { name: "destination", type: "address", internalType: "address" },
                            { name: "token", type: "address", internalType: "address" },
                            { name: "amount", type: "uint256", internalType: "uint256" },
                        ],
                    },
                    {
                        name: "sigs",
                        type: "tuple[]",
                        internalType: "struct Signature[]",
                        components: [
                            { name: "v", type: "uint8", internalType: "uint8" },
                            { name: "r", type: "bytes32", internalType: "bytes32" },
                            { name: "s", type: "bytes32", internalType: "bytes32" },
                        ],
                    },
                ],
            },
            {
                name: "proofs",
                type: "tuple[]",
                internalType: "struct State[]",
                components: [
                    { name: "intent", type: "uint8", internalType: "enum StateIntent" },
                    { name: "version", type: "uint256", internalType: "uint256" },
                    { name: "data", type: "bytes", internalType: "bytes" },
                    {
                        name: "allocations",
                        type: "tuple[]",
                        internalType: "struct Allocation[]",
                        components: [
                            { name: "destination", type: "address", internalType: "address" },
                            { name: "token", type: "address", internalType: "address" },
                            { name: "amount", type: "uint256", internalType: "uint256" },
                        ],
                    },
                    {
                        name: "sigs",
                        type: "tuple[]",
                        internalType: "struct Signature[]",
                        components: [
                            { name: "v", type: "uint8", internalType: "uint8" },
                            { name: "r", type: "bytes32", internalType: "bytes32" },
                            { name: "s", type: "bytes32", internalType: "bytes32" },
                        ],
                    },
                ],
            },
        ],
        outputs: [],
        stateMutability: "nonpayable",
    },
    {
        type: "function",
        name: "checkpoint",
        inputs: [
            { name: "channelId", type: "bytes32", internalType: "bytes32" },
            {
                name: "candidate",
                type: "tuple",
                internalType: "struct State",
                components: [
                    { name: "intent", type: "uint8", internalType: "enum StateIntent" },
                    { name: "version", type: "uint256", internalType: "uint256" },
                    { name: "data", type: "bytes", internalType: "bytes" },
                    {
                        name: "allocations",
                        type: "tuple[]",
                        internalType: "struct Allocation[]",
                        components: [
                            { name: "destination", type: "address", internalType: "address" },
                            { name: "token", type: "address", internalType: "address" },
                            { name: "amount", type: "uint256", internalType: "uint256" },
                        ],
                    },
                    {
                        name: "sigs",
                        type: "tuple[]",
                        internalType: "struct Signature[]",
                        components: [
                            { name: "v", type: "uint8", internalType: "uint8" },
                            { name: "r", type: "bytes32", internalType: "bytes32" },
                            { name: "s", type: "bytes32", internalType: "bytes32" },
                        ],
                    },
                ],
            },
            {
                name: "proofs",
                type: "tuple[]",
                internalType: "struct State[]",
                components: [
                    { name: "intent", type: "uint8", internalType: "enum StateIntent" },
                    { name: "version", type: "uint256", internalType: "uint256" },
                    { name: "data", type: "bytes", internalType: "bytes" },
                    {
                        name: "allocations",
                        type: "tuple[]",
                        internalType: "struct Allocation[]",
                        components: [
                            { name: "destination", type: "address", internalType: "address" },
                            { name: "token", type: "address", internalType: "address" },
                            { name: "amount", type: "uint256", internalType: "uint256" },
                        ],
                    },
                    {
                        name: "sigs",
                        type: "tuple[]",
                        internalType: "struct Signature[]",
                        components: [
                            { name: "v", type: "uint8", internalType: "uint8" },
                            { name: "r", type: "bytes32", internalType: "bytes32" },
                            { name: "s", type: "bytes32", internalType: "bytes32" },
                        ],
                    },
                ],
            },
        ],
        outputs: [],
        stateMutability: "nonpayable",
    },
    {
        type: "function",
        name: "close",
        inputs: [
            { name: "channelId", type: "bytes32", internalType: "bytes32" },
            {
                name: "candidate",
                type: "tuple",
                internalType: "struct State",
                components: [
                    { name: "intent", type: "uint8", internalType: "enum StateIntent" },
                    { name: "version", type: "uint256", internalType: "uint256" },
                    { name: "data", type: "bytes", internalType: "bytes" },
                    {
                        name: "allocations",
                        type: "tuple[]",
                        internalType: "struct Allocation[]",
                        components: [
                            { name: "destination", type: "address", internalType: "address" },
                            { name: "token", type: "address", internalType: "address" },
                            { name: "amount", type: "uint256", internalType: "uint256" },
                        ],
                    },
                    {
                        name: "sigs",
                        type: "tuple[]",
                        internalType: "struct Signature[]",
                        components: [
                            { name: "v", type: "uint8", internalType: "uint8" },
                            { name: "r", type: "bytes32", internalType: "bytes32" },
                            { name: "s", type: "bytes32", internalType: "bytes32" },
                        ],
                    },
                ],
            },
            {
                name: "",
                type: "tuple[]",
                internalType: "struct State[]",
                components: [
                    { name: "intent", type: "uint8", internalType: "enum StateIntent" },
                    { name: "version", type: "uint256", internalType: "uint256" },
                    { name: "data", type: "bytes", internalType: "bytes" },
                    {
                        name: "allocations",
                        type: "tuple[]",
                        internalType: "struct Allocation[]",
                        components: [
                            { name: "destination", type: "address", internalType: "address" },
                            { name: "token", type: "address", internalType: "address" },
                            { name: "amount", type: "uint256", internalType: "uint256" },
                        ],
                    },
                    {
                        name: "sigs",
                        type: "tuple[]",
                        internalType: "struct Signature[]",
                        components: [
                            { name: "v", type: "uint8", internalType: "uint8" },
                            { name: "r", type: "bytes32", internalType: "bytes32" },
                            { name: "s", type: "bytes32", internalType: "bytes32" },
                        ],
                    },
                ],
            },
        ],
        outputs: [],
        stateMutability: "nonpayable",
    },
    {
        type: "function",
        name: "create",
        inputs: [
            {
                name: "ch",
                type: "tuple",
                internalType: "struct Channel",
                components: [
                    { name: "participants", type: "address[]", internalType: "address[]" },
                    { name: "adjudicator", type: "address", internalType: "address" },
                    { name: "challenge", type: "uint64", internalType: "uint64" },
                    { name: "nonce", type: "uint64", internalType: "uint64" },
                ],
            },
            {
                name: "initial",
                type: "tuple",
                internalType: "struct State",
                components: [
                    { name: "intent", type: "uint8", internalType: "enum StateIntent" },
                    { name: "version", type: "uint256", internalType: "uint256" },
                    { name: "data", type: "bytes", internalType: "bytes" },
                    {
                        name: "allocations",
                        type: "tuple[]",
                        internalType: "struct Allocation[]",
                        components: [
                            { name: "destination", type: "address", internalType: "address" },
                            { name: "token", type: "address", internalType: "address" },
                            { name: "amount", type: "uint256", internalType: "uint256" },
                        ],
                    },
                    {
                        name: "sigs",
                        type: "tuple[]",
                        internalType: "struct Signature[]",
                        components: [
                            { name: "v", type: "uint8", internalType: "uint8" },
                            { name: "r", type: "bytes32", internalType: "bytes32" },
                            { name: "s", type: "bytes32", internalType: "bytes32" },
                        ],
                    },
                ],
            },
        ],
        outputs: [{ name: "channelId", type: "bytes32", internalType: "bytes32" }],
        stateMutability: "nonpayable",
    },
    {
        type: "function",
        name: "deposit",
        inputs: [
            { name: "token", type: "address", internalType: "address" },
            { name: "amount", type: "uint256", internalType: "uint256" },
        ],
        outputs: [],
        stateMutability: "payable",
    },
    {
        type: "function",
        name: "getAccountChannels",
        inputs: [{ name: "account", type: "address", internalType: "address" }],
        outputs: [{ name: "", type: "bytes32[]", internalType: "bytes32[]" }],
        stateMutability: "view",
    },
    {
        type: "function",
        name: "getAccountInfo",
        inputs: [
            { name: "user", type: "address", internalType: "address" },
            { name: "token", type: "address", internalType: "address" },
        ],
        outputs: [
            { name: "available", type: "uint256", internalType: "uint256" },
            { name: "locked", type: "uint256", internalType: "uint256" },
            { name: "channelCount", type: "uint256", internalType: "uint256" },
        ],
        stateMutability: "view",
    },
    {
        type: "function",
        name: "join",
        inputs: [
            { name: "channelId", type: "bytes32", internalType: "bytes32" },
            { name: "index", type: "uint256", internalType: "uint256" },
            {
                name: "sig",
                type: "tuple",
                internalType: "struct Signature",
                components: [
                    { name: "v", type: "uint8", internalType: "uint8" },
                    { name: "r", type: "bytes32", internalType: "bytes32" },
                    { name: "s", type: "bytes32", internalType: "bytes32" },
                ],
            },
        ],
        outputs: [{ name: "", type: "bytes32", internalType: "bytes32" }],
        stateMutability: "nonpayable",
    },
    {
        type: "function",
        name: "resize",
        inputs: [
            { name: "channelId", type: "bytes32", internalType: "bytes32" },
            {
                name: "candidate",
                type: "tuple",
                internalType: "struct State",
                components: [
                    { name: "intent", type: "uint8", internalType: "enum StateIntent" },
                    { name: "version", type: "uint256", internalType: "uint256" },
                    { name: "data", type: "bytes", internalType: "bytes" },
                    {
                        name: "allocations",
                        type: "tuple[]",
                        internalType: "struct Allocation[]",
                        components: [
                            { name: "destination", type: "address", internalType: "address" },
                            { name: "token", type: "address", internalType: "address" },
                            { name: "amount", type: "uint256", internalType: "uint256" },
                        ],
                    },
                    {
                        name: "sigs",
                        type: "tuple[]",
                        internalType: "struct Signature[]",
                        components: [
                            { name: "v", type: "uint8", internalType: "uint8" },
                            { name: "r", type: "bytes32", internalType: "bytes32" },
                            { name: "s", type: "bytes32", internalType: "bytes32" },
                        ],
                    },
                ],
            },
            {
                name: "proofs",
                type: "tuple[]",
                internalType: "struct State[]",
                components: [
                    { name: "intent", type: "uint8", internalType: "enum StateIntent" },
                    { name: "version", type: "uint256", internalType: "uint256" },
                    { name: "data", type: "bytes", internalType: "bytes" },
                    {
                        name: "allocations",
                        type: "tuple[]",
                        internalType: "struct Allocation[]",
                        components: [
                            { name: "destination", type: "address", internalType: "address" },
                            { name: "token", type: "address", internalType: "address" },
                            { name: "amount", type: "uint256", internalType: "uint256" },
                        ],
                    },
                    {
                        name: "sigs",
                        type: "tuple[]",
                        internalType: "struct Signature[]",
                        components: [
                            { name: "v", type: "uint8", internalType: "uint8" },
                            { name: "r", type: "bytes32", internalType: "bytes32" },
                            { name: "s", type: "bytes32", internalType: "bytes32" },
                        ],
                    },
                ],
            },
        ],
        outputs: [],
        stateMutability: "nonpayable",
    },
    {
        type: "function",
        name: "withdraw",
        inputs: [
            { name: "token", type: "address", internalType: "address" },
            { name: "amount", type: "uint256", internalType: "uint256" },
        ],
        outputs: [],
        stateMutability: "nonpayable",
    },
    {
        type: "event",
        name: "Challenged",
        inputs: [
            { name: "channelId", type: "bytes32", indexed: true, internalType: "bytes32" },
            { name: "expiration", type: "uint256", indexed: false, internalType: "uint256" },
        ],
        anonymous: false,
    },
    {
        type: "event",
        name: "Checkpointed",
        inputs: [{ name: "channelId", type: "bytes32", indexed: true, internalType: "bytes32" }],
        anonymous: false,
    },
    {
        type: "event",
        name: "Closed",
        inputs: [{ name: "channelId", type: "bytes32", indexed: true, internalType: "bytes32" }],
        anonymous: false,
    },
    {
        type: "event",
        name: "Created",
        inputs: [
            { name: "channelId", type: "bytes32", indexed: true, internalType: "bytes32" },
            {
                name: "channel",
                type: "tuple",
                indexed: false,
                internalType: "struct Channel",
                components: [
                    { name: "participants", type: "address[]", internalType: "address[]" },
                    { name: "adjudicator", type: "address", internalType: "address" },
                    { name: "challenge", type: "uint64", internalType: "uint64" },
                    { name: "nonce", type: "uint64", internalType: "uint64" },
                ],
            },
            {
                name: "initial",
                type: "tuple",
                indexed: false,
                internalType: "struct State",
                components: [
                    { name: "intent", type: "uint8", internalType: "enum StateIntent" },
                    { name: "version", type: "uint256", internalType: "uint256" },
                    { name: "data", type: "bytes", internalType: "bytes" },
                    {
                        name: "allocations",
                        type: "tuple[]",
                        internalType: "struct Allocation[]",
                        components: [
                            { name: "destination", type: "address", internalType: "address" },
                            { name: "token", type: "address", internalType: "address" },
                            { name: "amount", type: "uint256", internalType: "uint256" },
                        ],
                    },
                    {
                        name: "sigs",
                        type: "tuple[]",
                        internalType: "struct Signature[]",
                        components: [
                            { name: "v", type: "uint8", internalType: "uint8" },
                            { name: "r", type: "bytes32", internalType: "bytes32" },
                            { name: "s", type: "bytes32", internalType: "bytes32" },
                        ],
                    },
                ],
            },
        ],
        anonymous: false,
    },
    {
        type: "event",
        name: "Joined",
        inputs: [
            { name: "channelId", type: "bytes32", indexed: true, internalType: "bytes32" },
            { name: "index", type: "uint256", indexed: false, internalType: "uint256" },
        ],
        anonymous: false,
    },
    {
        type: "event",
        name: "Opened",
        inputs: [{ name: "channelId", type: "bytes32", indexed: true, internalType: "bytes32" }],
        anonymous: false,
    },
    {
        type: "event",
        name: "Resized",
        inputs: [
            { name: "channelId", type: "bytes32", indexed: true, internalType: "bytes32" },
            { name: "deltaAllocations", type: "int256[]", indexed: false, internalType: "int256[]" },
        ],
        anonymous: false,
    },
    { type: "error", name: "ChallengeNotExpired", inputs: [] },
    { type: "error", name: "ChannelNotFinal", inputs: [] },
    { type: "error", name: "ChannelNotFound", inputs: [{ name: "channelId", type: "bytes32", internalType: "bytes32" }] },
    { type: "error", name: "ECDSAInvalidSignature", inputs: [] },
    { type: "error", name: "ECDSAInvalidSignatureLength", inputs: [{ name: "length", type: "uint256", internalType: "uint256" }] },
    { type: "error", name: "ECDSAInvalidSignatureS", inputs: [{ name: "s", type: "bytes32", internalType: "bytes32" }] },
    {
        type: "error",
        name: "InsufficientBalance",
        inputs: [
            { name: "available", type: "uint256", internalType: "uint256" },
            { name: "required", type: "uint256", internalType: "uint256" },
        ],
    },
    { type: "error", name: "InvalidAdjudicator", inputs: [] },
    { type: "error", name: "InvalidAllocations", inputs: [] },
    { type: "error", name: "InvalidAmount", inputs: [] },
    { type: "error", name: "InvalidChallengePeriod", inputs: [] },
    { type: "error", name: "InvalidParticipant", inputs: [] },
    { type: "error", name: "InvalidState", inputs: [] },
    { type: "error", name: "InvalidStateSignatures", inputs: [] },
    { type: "error", name: "InvalidStatus", inputs: [] },
    {
        type: "error",
        name: "TransferFailed",
        inputs: [
            { name: "token", type: "address", internalType: "address" },
            { name: "to", type: "address", internalType: "address" },
            { name: "amount", type: "uint256", internalType: "uint256" },
        ],
    },
];
