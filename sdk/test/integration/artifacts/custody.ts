// Auto-generated test artifact. Do not edit manually.
// Generated from: Custody.sol/Custody
export const CustodyArtifacts = {
    abi: [
    {
        "type": "function",
        "name": "challenge",
        "inputs": [
            {
                "name": "channelId",
                "type": "bytes32",
                "internalType": "bytes32"
            },
            {
                "name": "candidate",
                "type": "tuple",
                "internalType": "struct State",
                "components": [
                    {
                        "name": "intent",
                        "type": "uint8",
                        "internalType": "enum StateIntent"
                    },
                    {
                        "name": "version",
                        "type": "uint256",
                        "internalType": "uint256"
                    },
                    {
                        "name": "data",
                        "type": "bytes",
                        "internalType": "bytes"
                    },
                    {
                        "name": "allocations",
                        "type": "tuple[]",
                        "internalType": "struct Allocation[]",
                        "components": [
                            {
                                "name": "destination",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "token",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "amount",
                                "type": "uint256",
                                "internalType": "uint256"
                            }
                        ]
                    },
                    {
                        "name": "sigs",
                        "type": "tuple[]",
                        "internalType": "struct Signature[]",
                        "components": [
                            {
                                "name": "v",
                                "type": "uint8",
                                "internalType": "uint8"
                            },
                            {
                                "name": "r",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            },
                            {
                                "name": "s",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            }
                        ]
                    }
                ]
            },
            {
                "name": "proofs",
                "type": "tuple[]",
                "internalType": "struct State[]",
                "components": [
                    {
                        "name": "intent",
                        "type": "uint8",
                        "internalType": "enum StateIntent"
                    },
                    {
                        "name": "version",
                        "type": "uint256",
                        "internalType": "uint256"
                    },
                    {
                        "name": "data",
                        "type": "bytes",
                        "internalType": "bytes"
                    },
                    {
                        "name": "allocations",
                        "type": "tuple[]",
                        "internalType": "struct Allocation[]",
                        "components": [
                            {
                                "name": "destination",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "token",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "amount",
                                "type": "uint256",
                                "internalType": "uint256"
                            }
                        ]
                    },
                    {
                        "name": "sigs",
                        "type": "tuple[]",
                        "internalType": "struct Signature[]",
                        "components": [
                            {
                                "name": "v",
                                "type": "uint8",
                                "internalType": "uint8"
                            },
                            {
                                "name": "r",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            },
                            {
                                "name": "s",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            }
                        ]
                    }
                ]
            },
            {
                "name": "challengerSig",
                "type": "tuple",
                "internalType": "struct Signature",
                "components": [
                    {
                        "name": "v",
                        "type": "uint8",
                        "internalType": "uint8"
                    },
                    {
                        "name": "r",
                        "type": "bytes32",
                        "internalType": "bytes32"
                    },
                    {
                        "name": "s",
                        "type": "bytes32",
                        "internalType": "bytes32"
                    }
                ]
            }
        ],
        "outputs": [],
        "stateMutability": "nonpayable"
    },
    {
        "type": "function",
        "name": "checkpoint",
        "inputs": [
            {
                "name": "channelId",
                "type": "bytes32",
                "internalType": "bytes32"
            },
            {
                "name": "candidate",
                "type": "tuple",
                "internalType": "struct State",
                "components": [
                    {
                        "name": "intent",
                        "type": "uint8",
                        "internalType": "enum StateIntent"
                    },
                    {
                        "name": "version",
                        "type": "uint256",
                        "internalType": "uint256"
                    },
                    {
                        "name": "data",
                        "type": "bytes",
                        "internalType": "bytes"
                    },
                    {
                        "name": "allocations",
                        "type": "tuple[]",
                        "internalType": "struct Allocation[]",
                        "components": [
                            {
                                "name": "destination",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "token",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "amount",
                                "type": "uint256",
                                "internalType": "uint256"
                            }
                        ]
                    },
                    {
                        "name": "sigs",
                        "type": "tuple[]",
                        "internalType": "struct Signature[]",
                        "components": [
                            {
                                "name": "v",
                                "type": "uint8",
                                "internalType": "uint8"
                            },
                            {
                                "name": "r",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            },
                            {
                                "name": "s",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            }
                        ]
                    }
                ]
            },
            {
                "name": "proofs",
                "type": "tuple[]",
                "internalType": "struct State[]",
                "components": [
                    {
                        "name": "intent",
                        "type": "uint8",
                        "internalType": "enum StateIntent"
                    },
                    {
                        "name": "version",
                        "type": "uint256",
                        "internalType": "uint256"
                    },
                    {
                        "name": "data",
                        "type": "bytes",
                        "internalType": "bytes"
                    },
                    {
                        "name": "allocations",
                        "type": "tuple[]",
                        "internalType": "struct Allocation[]",
                        "components": [
                            {
                                "name": "destination",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "token",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "amount",
                                "type": "uint256",
                                "internalType": "uint256"
                            }
                        ]
                    },
                    {
                        "name": "sigs",
                        "type": "tuple[]",
                        "internalType": "struct Signature[]",
                        "components": [
                            {
                                "name": "v",
                                "type": "uint8",
                                "internalType": "uint8"
                            },
                            {
                                "name": "r",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            },
                            {
                                "name": "s",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            }
                        ]
                    }
                ]
            }
        ],
        "outputs": [],
        "stateMutability": "nonpayable"
    },
    {
        "type": "function",
        "name": "close",
        "inputs": [
            {
                "name": "channelId",
                "type": "bytes32",
                "internalType": "bytes32"
            },
            {
                "name": "candidate",
                "type": "tuple",
                "internalType": "struct State",
                "components": [
                    {
                        "name": "intent",
                        "type": "uint8",
                        "internalType": "enum StateIntent"
                    },
                    {
                        "name": "version",
                        "type": "uint256",
                        "internalType": "uint256"
                    },
                    {
                        "name": "data",
                        "type": "bytes",
                        "internalType": "bytes"
                    },
                    {
                        "name": "allocations",
                        "type": "tuple[]",
                        "internalType": "struct Allocation[]",
                        "components": [
                            {
                                "name": "destination",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "token",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "amount",
                                "type": "uint256",
                                "internalType": "uint256"
                            }
                        ]
                    },
                    {
                        "name": "sigs",
                        "type": "tuple[]",
                        "internalType": "struct Signature[]",
                        "components": [
                            {
                                "name": "v",
                                "type": "uint8",
                                "internalType": "uint8"
                            },
                            {
                                "name": "r",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            },
                            {
                                "name": "s",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            }
                        ]
                    }
                ]
            },
            {
                "name": "",
                "type": "tuple[]",
                "internalType": "struct State[]",
                "components": [
                    {
                        "name": "intent",
                        "type": "uint8",
                        "internalType": "enum StateIntent"
                    },
                    {
                        "name": "version",
                        "type": "uint256",
                        "internalType": "uint256"
                    },
                    {
                        "name": "data",
                        "type": "bytes",
                        "internalType": "bytes"
                    },
                    {
                        "name": "allocations",
                        "type": "tuple[]",
                        "internalType": "struct Allocation[]",
                        "components": [
                            {
                                "name": "destination",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "token",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "amount",
                                "type": "uint256",
                                "internalType": "uint256"
                            }
                        ]
                    },
                    {
                        "name": "sigs",
                        "type": "tuple[]",
                        "internalType": "struct Signature[]",
                        "components": [
                            {
                                "name": "v",
                                "type": "uint8",
                                "internalType": "uint8"
                            },
                            {
                                "name": "r",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            },
                            {
                                "name": "s",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            }
                        ]
                    }
                ]
            }
        ],
        "outputs": [],
        "stateMutability": "nonpayable"
    },
    {
        "type": "function",
        "name": "create",
        "inputs": [
            {
                "name": "ch",
                "type": "tuple",
                "internalType": "struct Channel",
                "components": [
                    {
                        "name": "participants",
                        "type": "address[]",
                        "internalType": "address[]"
                    },
                    {
                        "name": "adjudicator",
                        "type": "address",
                        "internalType": "address"
                    },
                    {
                        "name": "challenge",
                        "type": "uint64",
                        "internalType": "uint64"
                    },
                    {
                        "name": "nonce",
                        "type": "uint64",
                        "internalType": "uint64"
                    }
                ]
            },
            {
                "name": "initial",
                "type": "tuple",
                "internalType": "struct State",
                "components": [
                    {
                        "name": "intent",
                        "type": "uint8",
                        "internalType": "enum StateIntent"
                    },
                    {
                        "name": "version",
                        "type": "uint256",
                        "internalType": "uint256"
                    },
                    {
                        "name": "data",
                        "type": "bytes",
                        "internalType": "bytes"
                    },
                    {
                        "name": "allocations",
                        "type": "tuple[]",
                        "internalType": "struct Allocation[]",
                        "components": [
                            {
                                "name": "destination",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "token",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "amount",
                                "type": "uint256",
                                "internalType": "uint256"
                            }
                        ]
                    },
                    {
                        "name": "sigs",
                        "type": "tuple[]",
                        "internalType": "struct Signature[]",
                        "components": [
                            {
                                "name": "v",
                                "type": "uint8",
                                "internalType": "uint8"
                            },
                            {
                                "name": "r",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            },
                            {
                                "name": "s",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            }
                        ]
                    }
                ]
            }
        ],
        "outputs": [
            {
                "name": "channelId",
                "type": "bytes32",
                "internalType": "bytes32"
            }
        ],
        "stateMutability": "nonpayable"
    },
    {
        "type": "function",
        "name": "deposit",
        "inputs": [
            {
                "name": "account",
                "type": "address",
                "internalType": "address"
            },
            {
                "name": "token",
                "type": "address",
                "internalType": "address"
            },
            {
                "name": "amount",
                "type": "uint256",
                "internalType": "uint256"
            }
        ],
        "outputs": [],
        "stateMutability": "payable"
    },
    {
        "type": "function",
        "name": "depositAndCreate",
        "inputs": [
            {
                "name": "token",
                "type": "address",
                "internalType": "address"
            },
            {
                "name": "amount",
                "type": "uint256",
                "internalType": "uint256"
            },
            {
                "name": "ch",
                "type": "tuple",
                "internalType": "struct Channel",
                "components": [
                    {
                        "name": "participants",
                        "type": "address[]",
                        "internalType": "address[]"
                    },
                    {
                        "name": "adjudicator",
                        "type": "address",
                        "internalType": "address"
                    },
                    {
                        "name": "challenge",
                        "type": "uint64",
                        "internalType": "uint64"
                    },
                    {
                        "name": "nonce",
                        "type": "uint64",
                        "internalType": "uint64"
                    }
                ]
            },
            {
                "name": "initial",
                "type": "tuple",
                "internalType": "struct State",
                "components": [
                    {
                        "name": "intent",
                        "type": "uint8",
                        "internalType": "enum StateIntent"
                    },
                    {
                        "name": "version",
                        "type": "uint256",
                        "internalType": "uint256"
                    },
                    {
                        "name": "data",
                        "type": "bytes",
                        "internalType": "bytes"
                    },
                    {
                        "name": "allocations",
                        "type": "tuple[]",
                        "internalType": "struct Allocation[]",
                        "components": [
                            {
                                "name": "destination",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "token",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "amount",
                                "type": "uint256",
                                "internalType": "uint256"
                            }
                        ]
                    },
                    {
                        "name": "sigs",
                        "type": "tuple[]",
                        "internalType": "struct Signature[]",
                        "components": [
                            {
                                "name": "v",
                                "type": "uint8",
                                "internalType": "uint8"
                            },
                            {
                                "name": "r",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            },
                            {
                                "name": "s",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            }
                        ]
                    }
                ]
            }
        ],
        "outputs": [
            {
                "name": "",
                "type": "bytes32",
                "internalType": "bytes32"
            }
        ],
        "stateMutability": "payable"
    },
    {
        "type": "function",
        "name": "getAccountsBalances",
        "inputs": [
            {
                "name": "accounts",
                "type": "address[]",
                "internalType": "address[]"
            },
            {
                "name": "tokens",
                "type": "address[]",
                "internalType": "address[]"
            }
        ],
        "outputs": [
            {
                "name": "",
                "type": "uint256[][]",
                "internalType": "uint256[][]"
            }
        ],
        "stateMutability": "view"
    },
    {
        "type": "function",
        "name": "getChannelBalances",
        "inputs": [
            {
                "name": "channelId",
                "type": "bytes32",
                "internalType": "bytes32"
            },
            {
                "name": "tokens",
                "type": "address[]",
                "internalType": "address[]"
            }
        ],
        "outputs": [
            {
                "name": "balances",
                "type": "uint256[]",
                "internalType": "uint256[]"
            }
        ],
        "stateMutability": "view"
    },
    {
        "type": "function",
        "name": "getChannelData",
        "inputs": [
            {
                "name": "channelId",
                "type": "bytes32",
                "internalType": "bytes32"
            }
        ],
        "outputs": [
            {
                "name": "channel",
                "type": "tuple",
                "internalType": "struct Channel",
                "components": [
                    {
                        "name": "participants",
                        "type": "address[]",
                        "internalType": "address[]"
                    },
                    {
                        "name": "adjudicator",
                        "type": "address",
                        "internalType": "address"
                    },
                    {
                        "name": "challenge",
                        "type": "uint64",
                        "internalType": "uint64"
                    },
                    {
                        "name": "nonce",
                        "type": "uint64",
                        "internalType": "uint64"
                    }
                ]
            },
            {
                "name": "status",
                "type": "uint8",
                "internalType": "enum ChannelStatus"
            },
            {
                "name": "wallets",
                "type": "address[]",
                "internalType": "address[]"
            },
            {
                "name": "challengeExpiry",
                "type": "uint256",
                "internalType": "uint256"
            },
            {
                "name": "lastValidState",
                "type": "tuple",
                "internalType": "struct State",
                "components": [
                    {
                        "name": "intent",
                        "type": "uint8",
                        "internalType": "enum StateIntent"
                    },
                    {
                        "name": "version",
                        "type": "uint256",
                        "internalType": "uint256"
                    },
                    {
                        "name": "data",
                        "type": "bytes",
                        "internalType": "bytes"
                    },
                    {
                        "name": "allocations",
                        "type": "tuple[]",
                        "internalType": "struct Allocation[]",
                        "components": [
                            {
                                "name": "destination",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "token",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "amount",
                                "type": "uint256",
                                "internalType": "uint256"
                            }
                        ]
                    },
                    {
                        "name": "sigs",
                        "type": "tuple[]",
                        "internalType": "struct Signature[]",
                        "components": [
                            {
                                "name": "v",
                                "type": "uint8",
                                "internalType": "uint8"
                            },
                            {
                                "name": "r",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            },
                            {
                                "name": "s",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            }
                        ]
                    }
                ]
            }
        ],
        "stateMutability": "view"
    },
    {
        "type": "function",
        "name": "getOpenChannels",
        "inputs": [
            {
                "name": "accounts",
                "type": "address[]",
                "internalType": "address[]"
            }
        ],
        "outputs": [
            {
                "name": "",
                "type": "bytes32[][]",
                "internalType": "bytes32[][]"
            }
        ],
        "stateMutability": "view"
    },
    {
        "type": "function",
        "name": "join",
        "inputs": [
            {
                "name": "channelId",
                "type": "bytes32",
                "internalType": "bytes32"
            },
            {
                "name": "index",
                "type": "uint256",
                "internalType": "uint256"
            },
            {
                "name": "sig",
                "type": "tuple",
                "internalType": "struct Signature",
                "components": [
                    {
                        "name": "v",
                        "type": "uint8",
                        "internalType": "uint8"
                    },
                    {
                        "name": "r",
                        "type": "bytes32",
                        "internalType": "bytes32"
                    },
                    {
                        "name": "s",
                        "type": "bytes32",
                        "internalType": "bytes32"
                    }
                ]
            }
        ],
        "outputs": [
            {
                "name": "",
                "type": "bytes32",
                "internalType": "bytes32"
            }
        ],
        "stateMutability": "nonpayable"
    },
    {
        "type": "function",
        "name": "resize",
        "inputs": [
            {
                "name": "channelId",
                "type": "bytes32",
                "internalType": "bytes32"
            },
            {
                "name": "candidate",
                "type": "tuple",
                "internalType": "struct State",
                "components": [
                    {
                        "name": "intent",
                        "type": "uint8",
                        "internalType": "enum StateIntent"
                    },
                    {
                        "name": "version",
                        "type": "uint256",
                        "internalType": "uint256"
                    },
                    {
                        "name": "data",
                        "type": "bytes",
                        "internalType": "bytes"
                    },
                    {
                        "name": "allocations",
                        "type": "tuple[]",
                        "internalType": "struct Allocation[]",
                        "components": [
                            {
                                "name": "destination",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "token",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "amount",
                                "type": "uint256",
                                "internalType": "uint256"
                            }
                        ]
                    },
                    {
                        "name": "sigs",
                        "type": "tuple[]",
                        "internalType": "struct Signature[]",
                        "components": [
                            {
                                "name": "v",
                                "type": "uint8",
                                "internalType": "uint8"
                            },
                            {
                                "name": "r",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            },
                            {
                                "name": "s",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            }
                        ]
                    }
                ]
            },
            {
                "name": "proofs",
                "type": "tuple[]",
                "internalType": "struct State[]",
                "components": [
                    {
                        "name": "intent",
                        "type": "uint8",
                        "internalType": "enum StateIntent"
                    },
                    {
                        "name": "version",
                        "type": "uint256",
                        "internalType": "uint256"
                    },
                    {
                        "name": "data",
                        "type": "bytes",
                        "internalType": "bytes"
                    },
                    {
                        "name": "allocations",
                        "type": "tuple[]",
                        "internalType": "struct Allocation[]",
                        "components": [
                            {
                                "name": "destination",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "token",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "amount",
                                "type": "uint256",
                                "internalType": "uint256"
                            }
                        ]
                    },
                    {
                        "name": "sigs",
                        "type": "tuple[]",
                        "internalType": "struct Signature[]",
                        "components": [
                            {
                                "name": "v",
                                "type": "uint8",
                                "internalType": "uint8"
                            },
                            {
                                "name": "r",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            },
                            {
                                "name": "s",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            }
                        ]
                    }
                ]
            }
        ],
        "outputs": [],
        "stateMutability": "nonpayable"
    },
    {
        "type": "function",
        "name": "withdraw",
        "inputs": [
            {
                "name": "token",
                "type": "address",
                "internalType": "address"
            },
            {
                "name": "amount",
                "type": "uint256",
                "internalType": "uint256"
            }
        ],
        "outputs": [],
        "stateMutability": "nonpayable"
    },
    {
        "type": "event",
        "name": "Challenged",
        "inputs": [
            {
                "name": "channelId",
                "type": "bytes32",
                "indexed": true,
                "internalType": "bytes32"
            },
            {
                "name": "state",
                "type": "tuple",
                "indexed": false,
                "internalType": "struct State",
                "components": [
                    {
                        "name": "intent",
                        "type": "uint8",
                        "internalType": "enum StateIntent"
                    },
                    {
                        "name": "version",
                        "type": "uint256",
                        "internalType": "uint256"
                    },
                    {
                        "name": "data",
                        "type": "bytes",
                        "internalType": "bytes"
                    },
                    {
                        "name": "allocations",
                        "type": "tuple[]",
                        "internalType": "struct Allocation[]",
                        "components": [
                            {
                                "name": "destination",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "token",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "amount",
                                "type": "uint256",
                                "internalType": "uint256"
                            }
                        ]
                    },
                    {
                        "name": "sigs",
                        "type": "tuple[]",
                        "internalType": "struct Signature[]",
                        "components": [
                            {
                                "name": "v",
                                "type": "uint8",
                                "internalType": "uint8"
                            },
                            {
                                "name": "r",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            },
                            {
                                "name": "s",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            }
                        ]
                    }
                ]
            },
            {
                "name": "expiration",
                "type": "uint256",
                "indexed": false,
                "internalType": "uint256"
            }
        ],
        "anonymous": false
    },
    {
        "type": "event",
        "name": "Checkpointed",
        "inputs": [
            {
                "name": "channelId",
                "type": "bytes32",
                "indexed": true,
                "internalType": "bytes32"
            },
            {
                "name": "state",
                "type": "tuple",
                "indexed": false,
                "internalType": "struct State",
                "components": [
                    {
                        "name": "intent",
                        "type": "uint8",
                        "internalType": "enum StateIntent"
                    },
                    {
                        "name": "version",
                        "type": "uint256",
                        "internalType": "uint256"
                    },
                    {
                        "name": "data",
                        "type": "bytes",
                        "internalType": "bytes"
                    },
                    {
                        "name": "allocations",
                        "type": "tuple[]",
                        "internalType": "struct Allocation[]",
                        "components": [
                            {
                                "name": "destination",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "token",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "amount",
                                "type": "uint256",
                                "internalType": "uint256"
                            }
                        ]
                    },
                    {
                        "name": "sigs",
                        "type": "tuple[]",
                        "internalType": "struct Signature[]",
                        "components": [
                            {
                                "name": "v",
                                "type": "uint8",
                                "internalType": "uint8"
                            },
                            {
                                "name": "r",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            },
                            {
                                "name": "s",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            }
                        ]
                    }
                ]
            }
        ],
        "anonymous": false
    },
    {
        "type": "event",
        "name": "Closed",
        "inputs": [
            {
                "name": "channelId",
                "type": "bytes32",
                "indexed": true,
                "internalType": "bytes32"
            },
            {
                "name": "finalState",
                "type": "tuple",
                "indexed": false,
                "internalType": "struct State",
                "components": [
                    {
                        "name": "intent",
                        "type": "uint8",
                        "internalType": "enum StateIntent"
                    },
                    {
                        "name": "version",
                        "type": "uint256",
                        "internalType": "uint256"
                    },
                    {
                        "name": "data",
                        "type": "bytes",
                        "internalType": "bytes"
                    },
                    {
                        "name": "allocations",
                        "type": "tuple[]",
                        "internalType": "struct Allocation[]",
                        "components": [
                            {
                                "name": "destination",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "token",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "amount",
                                "type": "uint256",
                                "internalType": "uint256"
                            }
                        ]
                    },
                    {
                        "name": "sigs",
                        "type": "tuple[]",
                        "internalType": "struct Signature[]",
                        "components": [
                            {
                                "name": "v",
                                "type": "uint8",
                                "internalType": "uint8"
                            },
                            {
                                "name": "r",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            },
                            {
                                "name": "s",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            }
                        ]
                    }
                ]
            }
        ],
        "anonymous": false
    },
    {
        "type": "event",
        "name": "Created",
        "inputs": [
            {
                "name": "channelId",
                "type": "bytes32",
                "indexed": true,
                "internalType": "bytes32"
            },
            {
                "name": "wallet",
                "type": "address",
                "indexed": true,
                "internalType": "address"
            },
            {
                "name": "channel",
                "type": "tuple",
                "indexed": false,
                "internalType": "struct Channel",
                "components": [
                    {
                        "name": "participants",
                        "type": "address[]",
                        "internalType": "address[]"
                    },
                    {
                        "name": "adjudicator",
                        "type": "address",
                        "internalType": "address"
                    },
                    {
                        "name": "challenge",
                        "type": "uint64",
                        "internalType": "uint64"
                    },
                    {
                        "name": "nonce",
                        "type": "uint64",
                        "internalType": "uint64"
                    }
                ]
            },
            {
                "name": "initial",
                "type": "tuple",
                "indexed": false,
                "internalType": "struct State",
                "components": [
                    {
                        "name": "intent",
                        "type": "uint8",
                        "internalType": "enum StateIntent"
                    },
                    {
                        "name": "version",
                        "type": "uint256",
                        "internalType": "uint256"
                    },
                    {
                        "name": "data",
                        "type": "bytes",
                        "internalType": "bytes"
                    },
                    {
                        "name": "allocations",
                        "type": "tuple[]",
                        "internalType": "struct Allocation[]",
                        "components": [
                            {
                                "name": "destination",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "token",
                                "type": "address",
                                "internalType": "address"
                            },
                            {
                                "name": "amount",
                                "type": "uint256",
                                "internalType": "uint256"
                            }
                        ]
                    },
                    {
                        "name": "sigs",
                        "type": "tuple[]",
                        "internalType": "struct Signature[]",
                        "components": [
                            {
                                "name": "v",
                                "type": "uint8",
                                "internalType": "uint8"
                            },
                            {
                                "name": "r",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            },
                            {
                                "name": "s",
                                "type": "bytes32",
                                "internalType": "bytes32"
                            }
                        ]
                    }
                ]
            }
        ],
        "anonymous": false
    },
    {
        "type": "event",
        "name": "Deposited",
        "inputs": [
            {
                "name": "wallet",
                "type": "address",
                "indexed": true,
                "internalType": "address"
            },
            {
                "name": "token",
                "type": "address",
                "indexed": true,
                "internalType": "address"
            },
            {
                "name": "amount",
                "type": "uint256",
                "indexed": false,
                "internalType": "uint256"
            }
        ],
        "anonymous": false
    },
    {
        "type": "event",
        "name": "Joined",
        "inputs": [
            {
                "name": "channelId",
                "type": "bytes32",
                "indexed": true,
                "internalType": "bytes32"
            },
            {
                "name": "index",
                "type": "uint256",
                "indexed": false,
                "internalType": "uint256"
            }
        ],
        "anonymous": false
    },
    {
        "type": "event",
        "name": "Opened",
        "inputs": [
            {
                "name": "channelId",
                "type": "bytes32",
                "indexed": true,
                "internalType": "bytes32"
            }
        ],
        "anonymous": false
    },
    {
        "type": "event",
        "name": "Resized",
        "inputs": [
            {
                "name": "channelId",
                "type": "bytes32",
                "indexed": true,
                "internalType": "bytes32"
            },
            {
                "name": "deltaAllocations",
                "type": "int256[]",
                "indexed": false,
                "internalType": "int256[]"
            }
        ],
        "anonymous": false
    },
    {
        "type": "event",
        "name": "Withdrawn",
        "inputs": [
            {
                "name": "wallet",
                "type": "address",
                "indexed": true,
                "internalType": "address"
            },
            {
                "name": "token",
                "type": "address",
                "indexed": true,
                "internalType": "address"
            },
            {
                "name": "amount",
                "type": "uint256",
                "indexed": false,
                "internalType": "uint256"
            }
        ],
        "anonymous": false
    },
    {
        "type": "error",
        "name": "ChallengeNotExpired",
        "inputs": []
    },
    {
        "type": "error",
        "name": "ChannelNotFinal",
        "inputs": []
    },
    {
        "type": "error",
        "name": "ChannelNotFound",
        "inputs": [
            {
                "name": "channelId",
                "type": "bytes32",
                "internalType": "bytes32"
            }
        ]
    },
    {
        "type": "error",
        "name": "DepositAlreadyFulfilled",
        "inputs": []
    },
    {
        "type": "error",
        "name": "DepositsNotFulfilled",
        "inputs": [
            {
                "name": "expectedFulfilled",
                "type": "uint256",
                "internalType": "uint256"
            },
            {
                "name": "actualFulfilled",
                "type": "uint256",
                "internalType": "uint256"
            }
        ]
    },
    {
        "type": "error",
        "name": "ECDSAInvalidSignature",
        "inputs": []
    },
    {
        "type": "error",
        "name": "ECDSAInvalidSignatureLength",
        "inputs": [
            {
                "name": "length",
                "type": "uint256",
                "internalType": "uint256"
            }
        ]
    },
    {
        "type": "error",
        "name": "ECDSAInvalidSignatureS",
        "inputs": [
            {
                "name": "s",
                "type": "bytes32",
                "internalType": "bytes32"
            }
        ]
    },
    {
        "type": "error",
        "name": "InsufficientBalance",
        "inputs": [
            {
                "name": "available",
                "type": "uint256",
                "internalType": "uint256"
            },
            {
                "name": "required",
                "type": "uint256",
                "internalType": "uint256"
            }
        ]
    },
    {
        "type": "error",
        "name": "InvalidAdjudicator",
        "inputs": []
    },
    {
        "type": "error",
        "name": "InvalidAllocations",
        "inputs": []
    },
    {
        "type": "error",
        "name": "InvalidAmount",
        "inputs": []
    },
    {
        "type": "error",
        "name": "InvalidChallengePeriod",
        "inputs": []
    },
    {
        "type": "error",
        "name": "InvalidChallengerSignature",
        "inputs": []
    },
    {
        "type": "error",
        "name": "InvalidParticipant",
        "inputs": []
    },
    {
        "type": "error",
        "name": "InvalidState",
        "inputs": []
    },
    {
        "type": "error",
        "name": "InvalidStateSignatures",
        "inputs": []
    },
    {
        "type": "error",
        "name": "InvalidStatus",
        "inputs": []
    },
    {
        "type": "error",
        "name": "InvalidValue",
        "inputs": []
    },
    {
        "type": "error",
        "name": "SafeERC20FailedOperation",
        "inputs": [
            {
                "name": "token",
                "type": "address",
                "internalType": "address"
            }
        ]
    },
    {
        "type": "error",
        "name": "TransferFailed",
        "inputs": [
            {
                "name": "token",
                "type": "address",
                "internalType": "address"
            },
            {
                "name": "to",
                "type": "address",
                "internalType": "address"
            },
            {
                "name": "amount",
                "type": "uint256",
                "internalType": "uint256"
            }
        ]
    }
],
    bytecode: '0x608080604052346015576149a0908161001a8239f35b5f80fdfe60806040526004361015610011575f80fd5b5f3560e01c8063259311c9146118b05780632f33c4d6146116f75780635a9eb80e146116275780638340f549146115f2578063925bc4791461157c578063a22b823d14611054578063bc7b456f14610fc7578063d0cce1e814610b9f578063d37ff7b514610b33578063d710e92f146109a5578063de22731f146103a1578063e617208c1461023b5763f3fef3a3146100a8575f80fd5b34610237576040600319360112610237576100c161211f565b60243590335f52600160205260405f20906001600160a01b0381165f528160205260405f205491838310610207576001600160a01b0392508282165f5260205260405f20610110848254612691565b90551690816101b5575f80808084335af13d156101b0573d61013181612313565b9061013f60405192836120e4565b81525f60203d92013e5b1561017d575b6040519081527fd1c19fbcd4551a5edfb66d43d2e337c04837afda3482b42bdf569a8fccdae5fb60203392a3005b907fbf182be8000000000000000000000000000000000000000000000000000000005f526004523360245260445260645ffd5b610149565b6102026040517fa9059cbb000000000000000000000000000000000000000000000000000000006020820152336024820152826044820152604481526101fc6064826120e4565b83614666565b61014f565b50507fcf479181000000000000000000000000000000000000000000000000000000005f5260045260245260445ffd5b5f80fd5b34610237576020600319360112610237575f606060405161025b81612090565b818152826020820152826040820152015261027461269e565b506004355f525f60205260405f2061028b81612602565b60ff60038301541691604051906102a36060836120e4565b600282526040366020840137600481015f5b600281106103735750506102d0600f600e8301549201612bdb565b6040519460a0865267ffffffffffffffff60606102fa8751608060a08b01526101208a01906121ae565b966001600160a01b0360208201511660c08a01528260408201511660e08a0152015116610100870152600581101561035f57859461035b9461034892602088015286820360408801526121ae565b9160608501528382036080850152612219565b0390f35b634e487b7160e01b5f52602160045260245ffd5b806001600160a01b0361038860019385612dcf565b90549060031b1c1661039a82876126f9565b52016102b5565b34610237576103af3661201d565b5050815f525f60205260405f2060ff600382015416600581101561035f57801561099257600281036106c8575081356004811015610237576103f0816121ea565b600381036106b957602083013580156106b95760808401916002610414848761258c565b9050036106915761043761042785612602565b6104313688612439565b90613fdc565b156106915761044990600f85016127ef565b60108301556011820161045f604085018561265e565b9067ffffffffffffffff821161067d576104838261047d8554612808565b85612856565b5f90601f8311600114610619576104b192915f918361060e575b50508160011b915f199060031b1c19161790565b90555b601282016104c5606085018561258c565b91906104d1838361289b565b905f5260205f205f915b8383106105a857505050506104f460138301918461258c565b906104ff828461291d565b915f5260205f205f925b82841061056857505050507f3646844802330633cc652490829391a0e9ddb82143a86a7e39ca148dfb05c9109161054f6105496012610563945b01612b5a565b85614283565b604051918291602083526020830190612e13565b0390a2005b80359060ff8216820361023757606060039160ff6001941660ff198654161785556020810135848601556040810135600286015501920193019290610509565b60036060826001600160a01b036105c0600195612909565b166001600160a01b03198654161785556105dc60208201612909565b6001600160a01b0385870191166001600160a01b031982541617905560408101356002860155019201920191906104db565b01359050888061049d565b601f19831691845f5260205f20925f5b818110610665575090846001959493921061064c575b505050811b0190556104b4565b01355f19600384901b60f8161c1916905587808061063f565b91936020600181928787013581550195019201610629565b634e487b7160e01b5f52604160045260245ffd5b7f773a750f000000000000000000000000000000000000000000000000000000005f5260045ffd5b63baf3f0f760e01b5f5260045ffd5b60030361096a57600e810180544210156109365782356004811015610237576106f0816121ea565b600381036106b95761070e61070484612602565b6104313687612439565b15610691575f6107229255600f83016127ef565b602082013560108201556011810161073d604084018461265e565b9067ffffffffffffffff821161067d5761075b8261047d8554612808565b5f90601f83116001146108d25761078892915f91836108c75750508160011b915f199060031b1c19161790565b90555b6012810161079c606084018461258c565b91906107a8838361289b565b905f5260205f205f915b8383106108615750505050601381016107ce608084018461258c565b906107d9828461291d565b915f5260205f205f925b82841061082157505050507f3646844802330633cc652490829391a0e9ddb82143a86a7e39ca148dfb05c9109161054f610549601261056394610543565b80359060ff8216820361023757606060039160ff6001941660ff1986541617855560208101358486015560408101356002860155019201930192906107e3565b60036060826001600160a01b03610879600195612909565b166001600160a01b031986541617855561089560208201612909565b6001600160a01b0385870191166001600160a01b031982541617905560408101356002860155019201920191906107b2565b01359050878061049d565b601f19831691845f5260205f20925f5b81811061091e5750908460019594939210610905575b505050811b01905561078b565b01355f19600384901b60f8161c191690558680806108f8565b919360206001819287870135815501950192016108e2565b507f3646844802330633cc652490829391a0e9ddb82143a86a7e39ca148dfb05c9109161054f610549601261056394610543565b7ff525e320000000000000000000000000000000000000000000000000000000005f5260045ffd5b836379c1d89f60e11b5f5260045260245ffd5b346102375760206003193601126102375760043567ffffffffffffffff8111610237576109d6903690600401612149565b8051906109fb6109e583612107565b926109f360405194856120e4565b808452612107565b90610a0e601f196020850193018361298b565b5f5b8151811015610aa2576001600160a01b03610a2b82846126f9565b51165f526001602052600160405f20016040519081602082549182815201915f5260205f20905f905b808210610a8a5750505090610a6e816001949303826120e4565b610a7882876126f9565b52610a8381866126f9565b5001610a10565b90919260016020819286548152019401920190610a54565b50509060405191829160208301906020845251809152604083019060408160051b85010192915f905b828210610ada57505050500390f35b9193909294603f19908203018252845190602080835192838152019201905f905b808210610b1b575050506020806001929601920192018594939192610acb565b90919260208060019286518152019401920190610afb565b346102375760406003193601126102375760043567ffffffffffffffff811161023757608060031982360301126102375760243567ffffffffffffffff81116102375760a0600319823603011261023757602091610b9791600401906004016137da565b604051908152f35b3461023757610bad3661201d565b835f525f60205260405f2091600383019160ff83541690600582101561035f578115610fb4576004821461096a57853592600484101561023757610bf0846121ea565b836106b957600f86019260ff84541690600181145f14610c32577ff525e320000000000000000000000000000000000000000000000000000000005f5260045ffd5b600203610f0157506001600160a01b0360018701541691610c66610c56368a612439565b610c5f86612bdb565b9085614482565b156106b957610c90926020926040518095819482936305b959ef60e01b84528d8d60048601612f8f565b03915afa908115610ef6575f91610ec7575b50156106b957610cbd925b600260ff198254161790556127ef565b6020820135601082015560118101610cd8604084018461265e565b9067ffffffffffffffff821161067d57610cf68261047d8554612808565b5f90601f8311600114610e6357610d2392915f91836108c75750508160011b915f199060031b1c19161790565b90555b60128101610d37606084018461258c565b9190610d43838361289b565b905f5260205f205f915b838310610dfd57868660138701610d67608083018361258c565b90610d72828461291d565b915f5260205f205f925b828410610dbd57857fa876bb57c3d3b4b0363570fd7443e30dfe18d4b422fe9898358262d78485325d61056387604051918291602083526020830190612e13565b80359060ff8216820361023757606060039160ff6001941660ff198654161785556020810135848601556040810135600286015501920193019290610d7c565b60036060826001600160a01b03610e15600195612909565b166001600160a01b0319865416178555610e3160208201612909565b6001600160a01b0385870191166001600160a01b03198254161790556040810135600286015501920192019190610d4d565b601f19831691845f5260205f20925f5b818110610eaf5750908460019594939210610e96575b505050811b019055610d26565b01355f19600384901b60f8161c19169055868080610e89565b91936020600181928787013581550195019201610e73565b610ee9915060203d602011610eef575b610ee181836120e4565b81019061270d565b87610ca2565b503d610ed7565b6040513d5f823e3d90fd5b90916020610f35916001600160a01b0360018a0154169460405193849283926305b959ef60e01b84528d8d60048601612f8f565b0381865afa908115610ef6575f91610f95575b50156106b957610f57816121ea565b15610f6d575b50610cbd925f600e860155610cad565b610f8a90610f7b3688612439565b610f8484612bdb565b91614482565b156106b95786610f5d565b610fae915060203d602011610eef57610ee181836120e4565b89610f48565b866379c1d89f60e11b5f5260045260245ffd5b346102375760c06003193601126102375760243567ffffffffffffffff81116102375760a060031982360301126102375760443567ffffffffffffffff811161023757611018903690600401611fec565b9060607fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff9c3601126102375761105292600401600435613027565b005b346102375760a060031936011261023757600435606060431936011261023757805f525f60205260405f2090600382019160ff835416600581101561035f578015611569575f190161096a5760016024350361154157600d81015461151957600f8101906110d36110c482612602565b6110cd84612bdb565b9061406d565b6001600160a01b03611100816110e885612d3a565b90549060031b1c16926110fa366123c5565b906146d3565b16036106915761110f82612bdb565b936040519461111f6060876120e4565b6002865260405f5b8181106114f0575050608081019561113f87516126c8565b51611149826126c8565b52611153816126c8565b5061115d366123c5565b611166826126e9565b52611170816126e9565b50865260018301546001600160a01b031660405190602061119181846120e4565b5f8352601f19015f5b8181106114d95750506111c79160209160405180809581946305b959ef60e01b8352888b6004850161276c565b03915afa908115610ef6575f916114ba575b50156106b9576112456111ee60088501612d67565b9461122086600c87019060206001916001600160a01b0380825116166001600160a01b03198554161784550151910155565b6005850180546001600160a01b03191633179055825190611240826121ea565b6127ef565b6020810151601084015560118301604082015180519067ffffffffffffffff821161067d576112788261047d8554612808565b602090601f8311600114611457576112a692915f918361144c5750508160011b915f199060031b1c19161790565b90555b6060601284019101519060208251926112c2848461289b565b01905f5260205f205f915b8383106113e9575050505060138201945160208151916112ed838961291d565b01955f5260205f20955f905b8282106113b157835460ff19166002178455602087611359886001600160a01b036113238a612d3a565b90549060031b1c165f526001845261134183600160405f20016146fc565b50836001600160a01b03825116910151908333614154565b807fe8e915db7b3549b9e9e9b3e2ec2dc3edd1f76961504366998824836401f6846a8360405160018152a260405190807fd087f17acc177540af5f382bc30c65363705b90855144d285a822536ee11fdd15f80a28152f35b60036020826040600194518c60ff1960ff8084511616915416178d558c8685830151910155015160028c0155019801910190966112f9565b60036020826040600194516001600160a01b0380825116166001600160a01b03198854161787556001600160a01b0384820151166001600160a01b0387890191166001600160a01b031982541617905501516002860155019201920191906112cd565b015190508a8061049d565b90601f19831691845f52815f20925f5b8181106114a2575090846001959493921061148a575b505050811b0190556112a9565b01515f1960f88460031b161c1916905589808061147d565b92936020600181928786015181550195019301611467565b6114d3915060203d602011610eef57610ee181836120e4565b876111d9565b6020906114e461269e565b8282870101520161119a565b6020906040516114ff81612074565b5f81525f838201525f604082015282828b01015201611127565b7f1b136079000000000000000000000000000000000000000000000000000000005f5260045ffd5b7fa145c43e000000000000000000000000000000000000000000000000000000005f5260045ffd5b826379c1d89f60e11b5f5260045260245ffd5b60806003193601126102375761159061211f565b6044359067ffffffffffffffff82116102375760806003198336030112610237576064359167ffffffffffffffff83116102375760a06003198436030112610237576020926115e6610b979360243590336129e8565b600401906004016137da565b60606003193601126102375761160661211f565b602435906001600160a01b03821682036102375761105291604435916129e8565b346102375760406003193601126102375760043560243567ffffffffffffffff81116102375761165b903690600401612149565b61166581516129a7565b5f5b82518110156116b257600190845f525f602052601460405f20016001600160a01b038061169484886126f9565b5116165f5260205260405f20546116ab82856126f9565b5201611667565b506040518091602082016020835281518091526020604084019201905f5b8181106116de575050500390f35b82518452859450602093840193909201916001016116d0565b346102375760406003193601126102375760043567ffffffffffffffff811161023757611728903690600401611fec565b60243567ffffffffffffffff811161023757611748903690600401611fec565b91909261175482612107565b9361176260405195866120e4565b82855261176e83612107565b93611781601f196020880196018661298b565b5f5b84811061181a57858760405191829160208301906020845251809152604083019060408160051b85010192915f905b8282106117c157505050500390f35b9193909294603f19908203018252845190602080835192838152019201905f905b8082106118025750505060208060019296019201920185949391926117b2565b909192602080600192865181520194019201906117e2565b611823826129a7565b61182d82896126f9565b5261183881886126f9565b505f5b82811061184b5750600101611783565b6001906001600160a01b03611869611864858a8a6129d8565b612909565b165f528160205260405f206001600160a01b0361188a61186484888a6129d8565b165f5260205260405f20546118a9826118a3868d6126f9565b516126f9565b520161183b565b34610237576118be3661201d565b9190835f525f60205260405f209360ff600386015416600581101561035f578015611fd9577ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0161096a5783156106b957823592600484101561023757611924846121ea565b600284036106b957823593609e198436030194858112156102375761194c9085013690612439565b966020830135602089015160018101809111611b405781036106b95760608901936119778551613fa6565b606081019161199861199361198c858561258c565b369161232f565b613fa6565b6119ae6119a485612602565b6104313685612439565b15610691576119c0604083018361265e565b810195906020818803126102375780359067ffffffffffffffff821161023757019686601f89011215610237578735966119f988612107565b98611a076040519a8b6120e4565b888a5260208a01906020829a60051b82010192831161023757602001905b828210611fc95750505051611a3d61198c868661258c565b9060028951036106b957611a6f816040611a6681611a5d611aa6966126c8565b510151926126e9565b5101519061257f565b91611aa0611a90611a7f8c6126c8565b51611a898d6126e9565b5190614052565b916040611a6681611a5d846126c8565b92614052565b03611fa1575f198b019b8b8d11611b4057611ac08d612107565b9c806040519e8f90611ad290826120e4565b52611adc90612107565b601f19018d5f5b828110611f885750505060015b8c811015611b54578060051b8b01358c811215610237575f19820191908c01818311611b40578f92600193611b29611b39933690612439565b611b3383836126f9565b526126f9565b5001611af0565b634e487b7160e01b5f52601160045260245ffd5b5060208d611b88926001600160a01b0360018a015416906040518095819482936305b959ef60e01b84528d6004850161276c565b03915afa908115610ef6575f91611f69575b50156106b957611bad90600f86016127ef565b601084015560118301611bc3604083018361265e565b9067ffffffffffffffff821161067d57611be18261047d8554612808565b5f90601f8311600114611f0557611c0e92915f9183611efa5750508160011b915f199060031b1c19161790565b90555b60128301611c1f838361258c565b9190611c2b838361289b565b905f5260205f205f915b838310611e94575050505060138301611c51608083018361258c565b90611c5c828461291d565b915f5260205f205f925b828410611e545750505050611c7e9161198c9161258c565b9060068101916001600160a01b038354165f5b60028110611dfb57505f5b60028110611d5b575050600a5f9201915b60028110611d18575050505060405191602083019060208452518091526040830191905f5b818110611d0257857ff3b6c524f73df7344d9fcf2f960a57aba7fba7e292d8b79ed03d786f7b2b112f86860387a2005b8251845260209384019390920191600101611cd2565b806040611d27600193856126f9565b51015182611d358388612b47565b5001556040611d4482856126f9565b51015182611d528387612b47565b50015501611cad565b5f611d6982899796976126f9565b5112611d7b575b600101939293611c9c565b6001600160a01b03611d908260048801612dcf565b90549060031b1c1690611da381896126f9565b517f80000000000000000000000000000000000000000000000000000000000000008114611b4057600192611df49160405191611ddf83612074565b82528560208301525f0360408201528a6145b3565b9050611d70565b805f611e0c6001938a9897986126f9565b5113611e1c575b01939293611c91565b611e4f6001600160a01b03611e348360048a01612dcf565b90549060031b1c16848b611e48858d6126f9565b5192614154565b611e13565b80359060ff8216820361023757606060039160ff6001941660ff198654161785556020810135848601556040810135600286015501920193019290611c66565b60036060826001600160a01b03611eac600195612909565b166001600160a01b0319865416178555611ec860208201612909565b6001600160a01b0385870191166001600160a01b03198254161790556040810135600286015501920192019190611c35565b013590508a8061049d565b601f19831691845f5260205f20925f5b818110611f515750908460019594939210611f38575b505050811b019055611c11565b01355f19600384901b60f8161c19169055898080611f2b565b91936020600181928787013581550195019201611f15565b611f82915060203d602011610eef57610ee181836120e4565b89611b9a565b6020918282611f9561269e565b92010152018e90611ae3565b7f52e4cb1c000000000000000000000000000000000000000000000000000000005f5260045ffd5b8135815260209182019101611a25565b506379c1d89f60e11b5f5260045260245ffd5b9181601f840112156102375782359167ffffffffffffffff8311610237576020808501948460051b01011161023757565b6060600319820112610237576004359160243567ffffffffffffffff81116102375760a0600319828503011261023757600401916044359067ffffffffffffffff82116102375761207091600401611fec565b9091565b6060810190811067ffffffffffffffff82111761067d57604052565b6080810190811067ffffffffffffffff82111761067d57604052565b60a0810190811067ffffffffffffffff82111761067d57604052565b6040810190811067ffffffffffffffff82111761067d57604052565b90601f601f19910116810190811067ffffffffffffffff82111761067d57604052565b67ffffffffffffffff811161067d5760051b60200190565b600435906001600160a01b038216820361023757565b35906001600160a01b038216820361023757565b9080601f8301121561023757813561216081612107565b9261216e60405194856120e4565b81845260208085019260051b82010192831161023757602001905b8282106121965750505090565b602080916121a384612135565b815201910190612189565b90602080835192838152019201905f5b8181106121cb5750505090565b82516001600160a01b03168452602093840193909201916001016121be565b6004111561035f57565b90601f19601f602080948051918291828752018686015e5f8582860101520116010190565b8051612224816121ea565b825260208101516020830152612249604082015160a0604085015260a08401906121f4565b906060810151918381036060850152602080845192838152019301905f5b8181106122cb5750505060800151916080818303910152602080835192838152019201905f5b81811061229a5750505090565b909192602060606001926040875160ff8151168352848101518584015201516040820152019401910191909161228d565b909193602061230960019287519060406060926001600160a01b0381511683526001600160a01b036020820151166020840152015160408201520190565b9501929101612267565b67ffffffffffffffff811161067d57601f01601f191660200190565b92919261233b82612107565b9361234960405195866120e4565b606060208685815201930282019181831161023757925b82841061236d5750505050565b60608483031261023757602060609160405161238881612074565b61239187612135565b815261239e838801612135565b8382015260408701356040820152815201930192612360565b359060ff8216820361023757565b604319606091011261023757604051906123de82612074565b8160443560ff8116810361023757815260643560208201526040608435910152565b91908260609103126102375760405161241881612074565b6040808294612426816123b7565b8452602081013560208501520135910152565b919060a08382031261023757604051612451816120ac565b80938035600481101561023757825260208101356020830152604081013567ffffffffffffffff811161023757810183601f8201121561023757803561249681612313565b916124a460405193846120e4565b818352856020838301011161023757815f92602080930183860137830101526040830152606081013567ffffffffffffffff811161023757810183601f8201121561023757838160206124f99335910161232f565b606083015260808101359067ffffffffffffffff8211610237570182601f8201121561023757803561252a81612107565b9361253860405195866120e4565b8185526020606081870193028401019281841161023757602001915b838310612565575050505060800152565b60206060916125748486612400565b815201920191612554565b91908201809211611b4057565b903590601e1981360301821215610237570180359067ffffffffffffffff82116102375760200191606082023603831361023757565b90602082549182815201915f5260205f20905f5b8181106125e35750505090565b82546001600160a01b03168452602090930192600192830192016125d6565b9060405161260f81612090565b606067ffffffffffffffff600283956040516126368161262f81856125c2565b03826120e4565b85528260018201546001600160a01b038116602088015260a01c166040860152015416910152565b903590601e1981360301821215610237570180359067ffffffffffffffff82116102375760200191813603831361023757565b91908203918211611b4057565b604051906126ab826120ac565b60606080835f81525f602082015282604082015282808201520152565b8051156126d55760200190565b634e487b7160e01b5f52603260045260245ffd5b8051600110156126d55760400190565b80518210156126d55760209160051b010190565b90816020910312610237575180151581036102375790565b9060808152606067ffffffffffffffff600261274460808501866125c2565b948260018201546001600160a01b038116602088015260a01c16604086015201541691015290565b9161278261279092606085526060850190612725565b908382036020850152612219565b906040818303910152815180825260208201916020808360051b8301019401925f915b8383106127c257505050505090565b90919293946020806127e083601f1986600196030187528951612219565b970193019301919392906127b3565b906127f9816121ea565b60ff60ff198354169116179055565b90600182811c92168015612836575b602083101461282257565b634e487b7160e01b5f52602260045260245ffd5b91607f1691612817565b81811061284b575050565b5f8155600101612840565b9190601f811161286557505050565b61288f925f5260205f20906020601f840160051c83019310612891575b601f0160051c0190612840565b565b9091508190612882565b9068010000000000000000811161067d578154918181558282106128be57505050565b82600302926003840403611b405781600302916003830403611b40575f5260205f2091820191015b8181106128f1575050565b805f600392555f60018201555f6002820155016128e6565b356001600160a01b03811681036102375790565b9068010000000000000000811161067d5781549181815582821061294057505050565b82600302926003840403611b405781600302916003830403611b40575f5260205f2091820191015b818110612973575050565b805f600392555f60018201555f600282015501612968565b5f5b82811061299957505050565b60608282015260200161298d565b906129b182612107565b6129be60405191826120e4565b828152601f196129ce8294612107565b0190602036910137565b91908110156126d55760051b0190565b908215612b1f576001600160a01b0316918215918215612ae857813403612ac0577f8752a472e571a816aea92eec8dae9baf628e840f4929fbcc2d155e6233ff68a7916001600160a01b036020925b1693845f526001835260405f20865f52835260405f20612a5883825461257f565b905515612a69575b604051908152a3565b612abb6040517f23b872dd000000000000000000000000000000000000000000000000000000008482015233602482015230604482015282606482015260648152612ab56084826120e4565b86614666565b612a60565b7faa7feadc000000000000000000000000000000000000000000000000000000005f5260045ffd5b34612ac0577f8752a472e571a816aea92eec8dae9baf628e840f4929fbcc2d155e6233ff68a7916001600160a01b03602092612a37565b7f2c5211c6000000000000000000000000000000000000000000000000000000005f5260045ffd5b9060028110156126d55760011b01905f90565b908154612b6681612107565b92612b7460405194856120e4565b81845260208401905f5260205f205f915b838310612b925750505050565b60036020600192604051612ba581612074565b6001600160a01b0386541681526001600160a01b0385870154168382015260028601546040820152815201920192019190612b85565b90604051612be8816120ac565b809260ff815416612bf8816121ea565b8252600181015460208301526040516002820180545f91612c1882612808565b8085529160018116908115612d135750600114612ccf575b505090612c42816004949303826120e4565b6040840152612c5360038201612b5a565b606084015201908154612c6581612107565b92612c7360405194856120e4565b81845260208401905f5260205f205f915b838310612c95575050505060800152565b60036020600192604051612ca881612074565b60ff8654168152848601548382015260028601546040820152815201920192019190612c84565b5f9081526020812094939250905b808210612cf75750919250908101602001612c4282612c30565b9192936001816020925483858801015201910190939291612cdd565b60ff191660208087019190915292151560051b85019092019250612c429150839050612c30565b8054600110156126d5575f52600160205f2001905f90565b80548210156126d5575f5260205f2001905f90565b90604051612d74816120c8565b6020600182946001600160a01b0381541684520154910152565b9190612dbc576020816001600160a01b03806001945116166001600160a01b03198554161784550151910155565b634e487b7160e01b5f525f60045260245ffd5b60028210156126d55701905f90565b9035601e198236030181121561023757016020813591019167ffffffffffffffff821161023757606082023603831361023757565b8035600481101561023757612e27816121ea565b8252602081013560208301526040810135601e198236030181121561023757810160208135910167ffffffffffffffff82116102375781360381136102375781601f1992601f9260a060408801528160a088015260c08701375f60c08287010152011682019060c082019160e0612ea16060840184612dde565b86840360c0016060880152948590529101925f5b818110612f2757505050612ecf8160806020930190612dde565b92909360808183039101528281520191905f5b818110612eef5750505090565b90919260608060019260ff612f03886123b7565b16815260208701356020820152604087013560408201520194019101919091612ee2565b9091936060806001926001600160a01b03612f4189612135565b1681526001600160a01b03612f5860208a01612135565b16602082015260408881013590820152019501929101612eb5565b929190612f8a602091604086526040860190612e13565b930152565b91612fa5612fb392606085526060850190612725565b908382036020850152612e13565b906040818303910152828152602081019260208160051b83010193835f91609e1982360301945b848410612feb575050505050505090565b90919293949596601f1982820301835287358781121561023757602061301660019387839401612e13565b990193019401929195949390612fda565b91939290825f525f60205260405f20600381019560ff87541690600582101561035f5781156136e6575f946003831480156136d9575b61096a578435916004831015968761023757613078846121ea565b600384146106b95761309661308c87612602565b6110cd368a612439565b6040516130a2816120c8565b8754156126d557875f526001600160a01b0360205f20541681526001600160a01b0380613142816130d28c612d3a565b90549060031b1c1694602085019586526130ed366064612400565b9060405160208101918252604080820152600960608201527f6368616c6c656e6765000000000000000000000000000000000000000000000060808201526080815261313a60a0826120e4565b5190206146d3565b92511691169081141591826136c4575b505061369c57600f86019460ff8654169161035f576001146135f5578790613179816121ea565b600181036134a257506102375761318f836121ea565b6001830361340e5750506131b56131a63686612439565b6131af84612bdb565b9061420e565b156106b9575b6131d667ffffffffffffffff600185015460a01c164261257f565b9485600e850155610237576131ea916127ef565b6020820135601082015560118101613205604084018461265e565b9067ffffffffffffffff821161067d576132238261047d8554612808565b5f90601f83116001146133aa5761325092915f918361339f5750508160011b915f199060031b1c19161790565b90555b60128101613264606084018461258c565b9190613270838361289b565b905f5260205f205f915b838310613339575050505060130194613296608083018361258c565b906132a1828961291d565b965f5260205f205f975b8289106132f9575050507f2cce3a04acfb5f7911860de30611c13af2df5880b4a1f829fa7b4f2a26d0375693949550600360ff198254161790556132f460405192839283612f73565b0390a2565b80359060ff8216820361023757606060039160ff6001941660ff1986541617855560208101358486015560408101356002860155019201980197906132ab565b60036060826001600160a01b03613351600195612909565b166001600160a01b031986541617855561336d60208201612909565b6001600160a01b0385870191166001600160a01b0319825416179055604081013560028601550192019201919061327a565b013590505f8061049d565b601f19831691845f5260205f20925f5b8181106133f657509084600195949392106133dd575b505050811b019055613253565b01355f19600384901b60f8161c191690555f80806133d0565b919360206001819287870135815501950192016133ba565b6001600160a01b036001860154169161343361342a3689612439565b610c5f87612bdb565b156106b95761345d926020926040518095819482936305b959ef60e01b84528c8c60048601612f8f565b03915afa908115610ef6575f91613483575b506131bb5763baf3f0f760e01b5f5260045ffd5b61349c915060203d602011610eef57610ee181836120e4565b5f61346f565b90506134ad816121ea565b8061356e575086610237576134c1836121ea565b826106b9576134dc6134d33688612439565b6131af86612bdb565b156134ea575b50505b6131bb565b6001600160a01b036001860154169161350661342a3689612439565b156106b957613530926020926040518095819482936305b959ef60e01b84528c8c60048601612f8f565b03915afa908115610ef6575f9161354f575b50156106b9575f806134e2565b613568915060203d602011610eef57610ee181836120e4565b5f613542565b600291975061357c816121ea565b036106b95761358a826121ea565b600182146106b9575f9561359d836121ea565b826135be576001600160a01b036001860154169161343361342a3689612439565b505093505f936135cd816121ea565b600281036106b9576135e26131a63686612439565b6134e55763baf3f0f760e01b5f5260045ffd5b5050505091939495505061361791506131af6136113685612439565b91612bdb565b156106b9576132f48161365461054961198c60607f3646844802330633cc652490829391a0e9ddb82143a86a7e39ca148dfb05c91096018461258c565b837f2cce3a04acfb5f7911860de30611c13af2df5880b4a1f829fa7b4f2a26d0375660405180613685428683612f73565b0390a2604051918291602083526020830190612e13565b7f61a44f6e000000000000000000000000000000000000000000000000000000005f5260045ffd5b516001600160a01b0316141590505f80613152565b505f95506004831461305d565b856379c1d89f60e11b5f5260045260245ffd5b903590601e1981360301821215610237570180359067ffffffffffffffff821161023757602001918160051b3603831361023757565b3567ffffffffffffffff811681036102375790565b359067ffffffffffffffff8216820361023757565b919091608081840312610237576040519061377382612090565b819381359067ffffffffffffffff8211610237578261379b606094926137c594869401612149565b85526137a960208201612135565b60208601526137ba60408201613744565b604086015201613744565b910152565b91908110156126d5576060020190565b9060026137e783806136f9565b905014801590613f7e575b8015613f50575b8015613f07575b61154157602082016001600160a01b0361381982612909565b1615613edf576040830192610e1067ffffffffffffffff6138398661372f565b1610613eb7578235600481101561023757613853816121ea565b600181036106b957602084013594856106b9576138786138733685613759565b614547565b95865f525f60205260ff600360405f20015416600581101561035f5761096a576138af6138a53686613759565b6110cd3689612439565b90608087019160016138c1848a61258c565b905003610691576138d2838961258c565b919091156126d5576138e487806136f9565b919091156126d55761390e6001600160a01b0392916110fa6139068594612909565b953690612400565b92169116036106915760608701916002613928848a61258c565b905003611fa157885f525f60205260405f209161394587806136f9565b9067ffffffffffffffff821161067d5768010000000000000000821161067d578454828655808310613e9b575b50845f5260205f205f5b838110613e805750505050600183016001600160a01b0361399c8a612909565b166001600160a01b03198254161781556139b58661372f565b7fffffffff0000000000000000ffffffffffffffffffffffffffffffffffffffff7bffffffffffffffff000000000000000000000000000000000000000083549260a01b169116179055613a736002840196606089019767ffffffffffffffff613a1e8a61372f565b82547fffffffffffffffffffffffffffffffffffffffffffffffff000000000000000016911617905560038501805460ff191660011790556004850180546001600160a01b03191633179055600f85016127ef565b601083015560118201613a8960408a018a61265e565b9067ffffffffffffffff821161067d57613aa78261047d8554612808565b5f90601f8311600114613e1c57613ad492915f918361339f5750508160011b915f199060031b1c19161790565b90555b60128201613ae5848a61258c565b9190613af1838361289b565b905f5260205f205f915b838310613db65750505050613b1460138301918961258c565b90613b1f828461291d565b915f5260205f205f925b828410613d7657505050505f91600a600683019201925b60028110613ce9575050613b56613b8191612d67565b80929060206001916001600160a01b0380825116166001600160a01b03198554161784550151910155565b613b8b84806136f9565b919091156126d5576001600160a01b03613ba7613bd893612909565b165f526001602052613bbf88600160405f20016146fc565b5060206001600160a01b03825116910151908833614154565b604051936040855260c08501938035601e198236030181121561023757016020813591019467ffffffffffffffff8211610237578160051b36038613610237576080604088015281905260e0860195889590949392915f5b818110613cb5575050509467ffffffffffffffff613c9a859482613c8f613caf966001600160a01b03613c847f7044488f9b947dc40d596a71992214b1050317a18ab1dced28e9d22320c398429c9d612135565b1660608a0152613744565b166080870152613744565b1660a084015282810360208401523396612e13565b0390a390565b91965091929394966020806001926001600160a01b03613cd48b612135565b16815201970191019189969795949392613c30565b80613d4e8a6040613d1f84613d0d88613d196020613d1360019b613d0d858b61258c565b906137ca565b01612909565b9561258c565b01356001600160a01b0360405192613d36846120c8565b1682526020820152613d488387612b47565b90612d8e565b613d70604051613d5d816120c8565b5f81525f6020820152613d488388612b47565b01613b40565b80359060ff8216820361023757606060039160ff6001941660ff198654161785556020810135848601556040810135600286015501920193019290613b29565b60036060826001600160a01b03613dce600195612909565b166001600160a01b0319865416178555613dea60208201612909565b6001600160a01b0385870191166001600160a01b03198254161790556040810135600286015501920192019190613afb565b601f19831691845f5260205f20925f5b818110613e685750908460019594939210613e4f575b505050811b019055613ad7565b01355f19600384901b60f8161c191690555f8080613e42565b91936020600181928787013581550195019201613e2c565b6001906020613e8e85612909565b940193818401550161397c565b613eb190865f528360205f209182019101612840565b5f613972565b7fb4e12433000000000000000000000000000000000000000000000000000000005f5260045ffd5b7fea9e70ce000000000000000000000000000000000000000000000000000000005f5260045ffd5b50613f1282806136f9565b156126d557613f2090612909565b613f2a83806136f9565b600110156126d5576001600160a01b03613f476020829301612909565b16911614613800565b50613f5b82806136f9565b600110156126d557613f7760206001600160a01b039201612909565b16156137f9565b50613f8982806136f9565b156126d557613f9f6001600160a01b0391612909565b16156137f2565b60028151036106b9576001600160a01b036020613fd18282613fc7866126c8565b51015116936126e9565b51015116036106b957565b906080613fe9828461406d565b91019160028351510361404b575f5b600281106140095750505050600190565b6140148185516126f9565b516001600160a01b036140358161402c8587516126f9565b511692866146d3565b160361404357600101613ff8565b505050505f90565b5050505f90565b9190915f8382019384129112908015821691151617611b4057565b61407690614547565b90805190614083826121ea565b6020810151916140c86060604084015193015192604051948593602085019788526140ad816121ea565b6040850152606084015260a0608084015260c08301906121f4565b91601f198284030160a0830152602080825194858152019101925f5b818110614106575050614100925003601f1981018352826120e4565b51902090565b91600191935061414560209186519060406060926001600160a01b0381511683526001600160a01b036020820151166020840152015160408201520190565b940191019184929391936140e4565b8315614208576001600160a01b03165f52600160205260405f206001600160a01b0383165f528060205260405f20548481106141d8578461419491612691565b906001600160a01b0384165f5260205260405f20555f525f6020526001600160a01b03601460405f200191165f526020526141d460405f2091825461257f565b9055565b84907fcf479181000000000000000000000000000000000000000000000000000000005f5260045260245260445ffd5b50505050565b6040516142398161422b6020820194602086526040830190612219565b03601f1981018352826120e4565b5190209060405161425a8161422b6020820194602086526040830190612219565b5190201490565b60048101905b818110614272575050565b5f8082556001820155600201614267565b90815f525f60205260405f209060038201600460ff1982541617905560028151036106b9575f5b600281106144655750505f5b6002811061442c5750505f525f60205260405f2080545f825580614412575b505f60018201555f60028201555f6003820155614301600682016142fc8160048501612840565b614261565b61430d600a8201614261565b5f600e8201555f600f8201555f60108201556011810161432d8154612808565b90816143cf575b5050601281018054905f815581614396575b50506013018054905f81558161435a575050565b81600302916003830403611b40575f5260205f20908101905b81811061437e575050565b805f600392555f60018201555f600282015501614373565b81600302916003830403611b40575f5260205f20908101905b8181101561434657805f600392555f60018201555f6002820155016143af565b81601f5f93116001146143e65750555b5f80614334565b8183526020832061440291601f0160051c810190600101612840565b80825281602081209155556143df565b61442690825f5260205f2090810190612840565b5f6142d5565b806001600160a01b0361444160019385612d52565b90549060031b1c165f528160205261445e848360405f2001614764565b50016142b6565b8061447c614475600193856126f9565b51866145b3565b016142aa565b602060405180927fcc2a842d00000000000000000000000000000000000000000000000000000000825260406004830152816001600160a01b03816144df6144cd604483018a612219565b6003198382030160248401528a612219565b0392165afa5f918161450a575b506144ff57506020809101519101511090565b90505f8092500b1390565b9091506020813d60201161453f575b81614526602093836120e4565b810103126102375751805f0b810361023757905f6144ec565b3d9150614519565b8051906141006001600160a01b036020830151169167ffffffffffffffff606081604084015116920151169260405193849261458f602085019760a0895260c08601906121ae565b926040850152606084015260808301524660a083015203601f1981018352826120e4565b906040810191825115614661575f525f602052601460405f20019160208201916001600160a01b0380845116165f528360205260405f205493841561465a576001600160a01b0392518086115f1461464f57614610908096612691565b908380865116165f5260205260405f205551165f5260016020526001600160a01b038060405f20925116165f526020526141d460405f2091825461257f565b506146108580612691565b5050505050565b505050565b905f602091828151910182855af115610ef6575f513d6146ca57506001600160a01b0381163b155b6146955750565b6001600160a01b03907f5274afe7000000000000000000000000000000000000000000000000000000005f521660045260245ffd5b6001141561468e565b6146f0906146f99260ff8151166040602083015192015192614821565b909291926148a3565b90565b6001810190825f528160205260405f2054155f1461404b5780546801000000000000000081101561067d5761475161473b826001879401855584612d52565b819391549060031b91821b915f19901b19161790565b905554915f5260205260405f2055600190565b906001820191815f528260205260405f20548015155f14614043575f198101818111611b405782545f19810191908211611b40578181036147ec575b505050805480156147d8575f1901906147b98282612d52565b8154905f199060031b1b19169055555f526020525f6040812055600190565b634e487b7160e01b5f52603160045260245ffd5b61480c6147fc61473b9386612d52565b90549060031b1c92839286612d52565b90555f528360205260405f20555f80806147a0565b91907f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a08411614898579160209360809260ff5f9560405194855216868401526040830152606082015282805260015afa15610ef6575f516001600160a01b0381161561488e57905f905f90565b505f906001905f90565b5050505f9160039190565b6148ac816121ea565b806148b5575050565b6148be816121ea565b600181036148ee577ff645eedf000000000000000000000000000000000000000000000000000000005f5260045ffd5b6148f7816121ea565b6002810361492b57507ffce698f7000000000000000000000000000000000000000000000000000000005f5260045260245ffd5b600390614937816121ea565b1461493f5750565b7fd78bce0c000000000000000000000000000000000000000000000000000000005f5260045260245ffdfea264697066735822122002f6b139bbcbd14ae961e893cc2e778df6bf9bd3ee0d21952c11bd8f219e04f964736f6c634300081d0033' as `0x${string}`,
};
