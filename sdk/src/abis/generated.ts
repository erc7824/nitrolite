// Auto-generated file. Do not edit manually.
export const custodyAbi = [
  {
    "type": "constructor",
    "inputs": [],
    "stateMutability": "nonpayable"
  },
  {
    "type": "function",
    "name": "CHALLENGE_STATE_TYPEHASH",
    "inputs": [],
    "outputs": [
      {
        "name": "",
        "type": "bytes32",
        "internalType": "bytes32"
      }
    ],
    "stateMutability": "view"
  },
  {
    "type": "function",
    "name": "MIN_CHALLENGE_PERIOD",
    "inputs": [],
    "outputs": [
      {
        "name": "",
        "type": "uint256",
        "internalType": "uint256"
      }
    ],
    "stateMutability": "view"
  },
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
            "type": "bytes[]",
            "internalType": "bytes[]"
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
            "type": "bytes[]",
            "internalType": "bytes[]"
          }
        ]
      },
      {
        "name": "challengerSig",
        "type": "bytes",
        "internalType": "bytes"
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
            "type": "bytes[]",
            "internalType": "bytes[]"
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
            "type": "bytes[]",
            "internalType": "bytes[]"
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
            "type": "bytes[]",
            "internalType": "bytes[]"
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
            "type": "bytes[]",
            "internalType": "bytes[]"
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
            "type": "bytes[]",
            "internalType": "bytes[]"
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
            "type": "bytes[]",
            "internalType": "bytes[]"
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
    "name": "eip712Domain",
    "inputs": [],
    "outputs": [
      {
        "name": "fields",
        "type": "bytes1",
        "internalType": "bytes1"
      },
      {
        "name": "name",
        "type": "string",
        "internalType": "string"
      },
      {
        "name": "version",
        "type": "string",
        "internalType": "string"
      },
      {
        "name": "chainId",
        "type": "uint256",
        "internalType": "uint256"
      },
      {
        "name": "verifyingContract",
        "type": "address",
        "internalType": "address"
      },
      {
        "name": "salt",
        "type": "bytes32",
        "internalType": "bytes32"
      },
      {
        "name": "extensions",
        "type": "uint256[]",
        "internalType": "uint256[]"
      }
    ],
    "stateMutability": "view"
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
            "type": "bytes[]",
            "internalType": "bytes[]"
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
        "type": "bytes",
        "internalType": "bytes"
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
            "type": "bytes[]",
            "internalType": "bytes[]"
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
            "type": "bytes[]",
            "internalType": "bytes[]"
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
            "type": "bytes[]",
            "internalType": "bytes[]"
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
            "type": "bytes[]",
            "internalType": "bytes[]"
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
            "type": "bytes[]",
            "internalType": "bytes[]"
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
            "type": "bytes[]",
            "internalType": "bytes[]"
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
    "name": "EIP712DomainChanged",
    "inputs": [],
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
    "name": "ERC6492DeploymentFailed",
    "inputs": [
      {
        "name": "factory",
        "type": "address",
        "internalType": "address"
      },
      {
        "name": "calldata_",
        "type": "bytes",
        "internalType": "bytes"
      }
    ]
  },
  {
    "type": "error",
    "name": "ERC6492NoCode",
    "inputs": [
      {
        "name": "expectedSigner",
        "type": "address",
        "internalType": "address"
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
    "name": "InvalidShortString",
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
    "name": "StringTooLong",
    "inputs": [
      {
        "name": "str",
        "type": "string",
        "internalType": "string"
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
] as const;
