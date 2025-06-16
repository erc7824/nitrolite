// Auto-generated test artifact. Do not edit manually.
// Generated from: Dummy.sol/Dummy
export const DummyArtifacts = {
    abi: [
    {
        "type": "function",
        "name": "adjudicate",
        "inputs": [
            {
                "name": "",
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
                "name": "",
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
        "outputs": [
            {
                "name": "valid",
                "type": "bool",
                "internalType": "bool"
            }
        ],
        "stateMutability": "pure"
    },
    {
        "type": "function",
        "name": "compare",
        "inputs": [
            {
                "name": "",
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
                "name": "result",
                "type": "int8",
                "internalType": "int8"
            }
        ],
        "stateMutability": "pure"
    }
],
    bytecode: '0x60808060405234601557610163908161001a8239f35b5f80fdfe6080806040526004361015610012575f80fd5b5f3560e01c90816305b959ef14610092575063cc2a842d14610032575f80fd5b3461008e57604060031936011261008e5760043567ffffffffffffffff811161008e5760031960a0913603011261008e5760243567ffffffffffffffff811161008e5760031960a0913603011261008e57602060405160018152f35b5f80fd5b3461008e57606060031936011261008e5760043567ffffffffffffffff811161008e576003196080913603011261008e5760243567ffffffffffffffff811161008e5760031960a0913603011261008e5760443567ffffffffffffffff811161008e573660238201121561008e5780600401359067ffffffffffffffff821161008e57602490369260051b01011161008e5780600160209252f3fea26469706673582212207fd5d20e7dd7761c133fbaab73e38d483611bafbaeb8c16c6fae0963900a7c4a64736f6c634300081d0033' as `0x${string}`,
};
