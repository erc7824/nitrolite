// Auto-generated test artifact. Do not edit manually.
// Generated from: TestERC20.sol/TestERC20
export const TestERC20Artifacts = {
    abi: [
        {
            type: 'constructor',
            inputs: [
                {
                    name: 'name_',
                    type: 'string',
                    internalType: 'string',
                },
                {
                    name: 'symbol_',
                    type: 'string',
                    internalType: 'string',
                },
                {
                    name: 'decimals_',
                    type: 'uint8',
                    internalType: 'uint8',
                },
                {
                    name: 'cap_',
                    type: 'uint256',
                    internalType: 'uint256',
                },
            ],
            stateMutability: 'nonpayable',
        },
        {
            type: 'function',
            name: 'allowance',
            inputs: [
                {
                    name: 'owner',
                    type: 'address',
                    internalType: 'address',
                },
                {
                    name: 'spender',
                    type: 'address',
                    internalType: 'address',
                },
            ],
            outputs: [
                {
                    name: '',
                    type: 'uint256',
                    internalType: 'uint256',
                },
            ],
            stateMutability: 'view',
        },
        {
            type: 'function',
            name: 'approve',
            inputs: [
                {
                    name: 'spender',
                    type: 'address',
                    internalType: 'address',
                },
                {
                    name: 'value',
                    type: 'uint256',
                    internalType: 'uint256',
                },
            ],
            outputs: [
                {
                    name: '',
                    type: 'bool',
                    internalType: 'bool',
                },
            ],
            stateMutability: 'nonpayable',
        },
        {
            type: 'function',
            name: 'balanceOf',
            inputs: [
                {
                    name: 'account',
                    type: 'address',
                    internalType: 'address',
                },
            ],
            outputs: [
                {
                    name: '',
                    type: 'uint256',
                    internalType: 'uint256',
                },
            ],
            stateMutability: 'view',
        },
        {
            type: 'function',
            name: 'burn',
            inputs: [
                {
                    name: 'account',
                    type: 'address',
                    internalType: 'address',
                },
                {
                    name: 'amount',
                    type: 'uint256',
                    internalType: 'uint256',
                },
            ],
            outputs: [],
            stateMutability: 'nonpayable',
        },
        {
            type: 'function',
            name: 'cap',
            inputs: [],
            outputs: [
                {
                    name: '',
                    type: 'uint256',
                    internalType: 'uint256',
                },
            ],
            stateMutability: 'view',
        },
        {
            type: 'function',
            name: 'decimals',
            inputs: [],
            outputs: [
                {
                    name: '',
                    type: 'uint8',
                    internalType: 'uint8',
                },
            ],
            stateMutability: 'view',
        },
        {
            type: 'function',
            name: 'mint',
            inputs: [
                {
                    name: 'account',
                    type: 'address',
                    internalType: 'address',
                },
                {
                    name: 'amount',
                    type: 'uint256',
                    internalType: 'uint256',
                },
            ],
            outputs: [],
            stateMutability: 'nonpayable',
        },
        {
            type: 'function',
            name: 'name',
            inputs: [],
            outputs: [
                {
                    name: '',
                    type: 'string',
                    internalType: 'string',
                },
            ],
            stateMutability: 'view',
        },
        {
            type: 'function',
            name: 'symbol',
            inputs: [],
            outputs: [
                {
                    name: '',
                    type: 'string',
                    internalType: 'string',
                },
            ],
            stateMutability: 'view',
        },
        {
            type: 'function',
            name: 'totalSupply',
            inputs: [],
            outputs: [
                {
                    name: '',
                    type: 'uint256',
                    internalType: 'uint256',
                },
            ],
            stateMutability: 'view',
        },
        {
            type: 'function',
            name: 'transfer',
            inputs: [
                {
                    name: 'to',
                    type: 'address',
                    internalType: 'address',
                },
                {
                    name: 'value',
                    type: 'uint256',
                    internalType: 'uint256',
                },
            ],
            outputs: [
                {
                    name: '',
                    type: 'bool',
                    internalType: 'bool',
                },
            ],
            stateMutability: 'nonpayable',
        },
        {
            type: 'function',
            name: 'transferFrom',
            inputs: [
                {
                    name: 'from',
                    type: 'address',
                    internalType: 'address',
                },
                {
                    name: 'to',
                    type: 'address',
                    internalType: 'address',
                },
                {
                    name: 'value',
                    type: 'uint256',
                    internalType: 'uint256',
                },
            ],
            outputs: [
                {
                    name: '',
                    type: 'bool',
                    internalType: 'bool',
                },
            ],
            stateMutability: 'nonpayable',
        },
        {
            type: 'event',
            name: 'Approval',
            inputs: [
                {
                    name: 'owner',
                    type: 'address',
                    indexed: true,
                    internalType: 'address',
                },
                {
                    name: 'spender',
                    type: 'address',
                    indexed: true,
                    internalType: 'address',
                },
                {
                    name: 'value',
                    type: 'uint256',
                    indexed: false,
                    internalType: 'uint256',
                },
            ],
            anonymous: false,
        },
        {
            type: 'event',
            name: 'Transfer',
            inputs: [
                {
                    name: 'from',
                    type: 'address',
                    indexed: true,
                    internalType: 'address',
                },
                {
                    name: 'to',
                    type: 'address',
                    indexed: true,
                    internalType: 'address',
                },
                {
                    name: 'value',
                    type: 'uint256',
                    indexed: false,
                    internalType: 'uint256',
                },
            ],
            anonymous: false,
        },
        {
            type: 'error',
            name: 'ERC20ExceededCap',
            inputs: [
                {
                    name: 'increasedSupply',
                    type: 'uint256',
                    internalType: 'uint256',
                },
                {
                    name: 'cap',
                    type: 'uint256',
                    internalType: 'uint256',
                },
            ],
        },
        {
            type: 'error',
            name: 'ERC20InsufficientAllowance',
            inputs: [
                {
                    name: 'spender',
                    type: 'address',
                    internalType: 'address',
                },
                {
                    name: 'allowance',
                    type: 'uint256',
                    internalType: 'uint256',
                },
                {
                    name: 'needed',
                    type: 'uint256',
                    internalType: 'uint256',
                },
            ],
        },
        {
            type: 'error',
            name: 'ERC20InsufficientBalance',
            inputs: [
                {
                    name: 'sender',
                    type: 'address',
                    internalType: 'address',
                },
                {
                    name: 'balance',
                    type: 'uint256',
                    internalType: 'uint256',
                },
                {
                    name: 'needed',
                    type: 'uint256',
                    internalType: 'uint256',
                },
            ],
        },
        {
            type: 'error',
            name: 'ERC20InvalidApprover',
            inputs: [
                {
                    name: 'approver',
                    type: 'address',
                    internalType: 'address',
                },
            ],
        },
        {
            type: 'error',
            name: 'ERC20InvalidCap',
            inputs: [
                {
                    name: 'cap',
                    type: 'uint256',
                    internalType: 'uint256',
                },
            ],
        },
        {
            type: 'error',
            name: 'ERC20InvalidReceiver',
            inputs: [
                {
                    name: 'receiver',
                    type: 'address',
                    internalType: 'address',
                },
            ],
        },
        {
            type: 'error',
            name: 'ERC20InvalidSender',
            inputs: [
                {
                    name: 'sender',
                    type: 'address',
                    internalType: 'address',
                },
            ],
        },
        {
            type: 'error',
            name: 'ERC20InvalidSpender',
            inputs: [
                {
                    name: 'spender',
                    type: 'address',
                    internalType: 'address',
                },
            ],
        },
    ],
    bytecode:
        '0x60c06040523461036457610cd68038038061001981610368565b9283398101906080818303126103645780516001600160401b038111610364578261004591830161038d565b602082015190926001600160401b0382116103645761006591830161038d565b9060408101519060ff82168203610364576060015183516001600160401b03811161027557600354600181811c9116801561035a575b602082101461025757601f81116102f7575b50602094601f8211600114610294579481929394955f92610289575b50508160011b915f199060031b1c1916176003555b82516001600160401b03811161027557600454600181811c9116801561026b575b602082101461025757601f81116101f4575b506020601f821160011461019157819293945f92610186575b50508160011b915f199060031b1c1916176004555b80156101735760805260a0526040516108f790816103df823960805181818161040501526104b5015260a051816104f10152f35b63392e1e2760e01b5f525f60045260245ffd5b015190505f8061012a565b601f1982169060045f52805f20915f5b8181106101dc575095836001959697106101c4575b505050811b0160045561013f565b01515f1960f88460031b161c191690555f80806101b6565b9192602060018192868b0151815501940192016101a1565b60045f527f8a35acfbc15ff81a39ae7d344fd709f28e8600b4aa8c65c6b64bfe7fe36bd19b601f830160051c8101916020841061024d575b601f0160051c01905b8181106102425750610111565b5f8155600101610235565b909150819061022c565b634e487b7160e01b5f52602260045260245ffd5b90607f16906100ff565b634e487b7160e01b5f52604160045260245ffd5b015190505f806100c9565b601f1982169560035f52805f20915f5b8881106102df575083600195969798106102c7575b505050811b016003556100de565b01515f1960f88460031b161c191690555f80806102b9565b919260206001819286850151815501940192016102a4565b60035f527fc2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b601f830160051c81019160208410610350575b601f0160051c01905b81811061034557506100ad565b5f8155600101610338565b909150819061032f565b90607f169061009b565b5f80fd5b6040519190601f01601f191682016001600160401b0381118382101761027557604052565b81601f82011215610364578051906001600160401b038211610275576103bc601f8301601f1916602001610368565b928284526020838301011161036457815f9260208093018386015e830101529056fe6080806040526004361015610012575f80fd5b5f3560e01c90816306fdde03146106e457508063095ea7b31461066257806318160ddd1461064557806323b872dd14610515578063313ce567146104d8578063355274ea1461049e57806340c10f191461038a57806370a082311461035357806395d89b411461020a5780639dc29fac14610129578063a9059cbb146100f85763dd62ed3e146100a0575f80fd5b346100f45760406003193601126100f4576100b96107e5565b6001600160a01b036100c96107fb565b91165f5260016020526001600160a01b0360405f2091165f52602052602060405f2054604051908152f35b5f80fd5b346100f45760406003193601126100f45761011e6101146107e5565b6024359033610811565b602060405160018152f35b346100f45760406003193601126100f4576101426107e5565b6001600160a01b03602435911680156101de57805f525f60205260405f20548281106101ac576020835f947fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef938587528684520360408620558060025403600255604051908152a3005b907fe450d38c000000000000000000000000000000000000000000000000000000005f5260045260245260445260645ffd5b7f96c6fd1e000000000000000000000000000000000000000000000000000000005f525f60045260245ffd5b346100f4575f6003193601126100f4576040515f600454908160011c60018316928315610349575b6020821084146103355781855284939081156102f35750600114610297575b5003601f01601f191681019067ffffffffffffffff8211818310176102835761027f829182604052826107bb565b0390f35b634e487b7160e01b5f52604160045260245ffd5b60045f90815291507f8a35acfbc15ff81a39ae7d344fd709f28e8600b4aa8c65c6b64bfe7fe36bd19b5b8183106102d75750508101602001601f19610251565b60209193508060019154838588010152019101909183926102c1565b7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff001660208581019190915291151560051b84019091019150601f199050610251565b634e487b7160e01b5f52602260045260245ffd5b90607f1690610232565b346100f45760206003193601126100f4576001600160a01b036103746107e5565b165f525f602052602060405f2054604051908152f35b346100f45760406003193601126100f4576103a36107e5565b6001600160a01b03166024358115610472576002549080820180921161045e5760207fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef915f9360025584845283825260408420818154019055604051908152a37f000000000000000000000000000000000000000000000000000000000000000060025481811161043057005b7f9e79f854000000000000000000000000000000000000000000000000000000005f5260045260245260445ffd5b634e487b7160e01b5f52601160045260245ffd5b7fec442f05000000000000000000000000000000000000000000000000000000005f525f60045260245ffd5b346100f4575f6003193601126100f45760206040517f00000000000000000000000000000000000000000000000000000000000000008152f35b346100f4575f6003193601126100f457602060405160ff7f0000000000000000000000000000000000000000000000000000000000000000168152f35b346100f45760606003193601126100f45761052e6107e5565b6105366107fb565b604435906001600160a01b03831692835f52600160205260405f206001600160a01b0333165f5260205260405f20545f198110610579575b5061011e9350610811565b8381106106115784156105e55733156105b95761011e945f52600160205260405f206001600160a01b0333165f526020528360405f20910390558461056e565b7f94280d62000000000000000000000000000000000000000000000000000000005f525f60045260245ffd5b7fe602df05000000000000000000000000000000000000000000000000000000005f525f60045260245ffd5b83907ffb8f41b2000000000000000000000000000000000000000000000000000000005f523360045260245260445260645ffd5b346100f4575f6003193601126100f4576020600254604051908152f35b346100f45760406003193601126100f45761067b6107e5565b6024359033156105e5576001600160a01b03169081156105b957335f52600160205260405f20825f526020528060405f20556040519081527f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b92560203392a3602060405160018152f35b346100f4575f6003193601126100f4575f600354908160011c600183169283156107b1575b6020821084146103355781855284939081156102f35750600114610755575003601f01601f191681019067ffffffffffffffff8211818310176102835761027f829182604052826107bb565b60035f90815291507fc2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b5b8183106107955750508101602001601f19610251565b602091935080600191548385880101520191019091839261077f565b90607f1690610709565b601f19601f602060409481855280519182918282880152018686015e5f8582860101520116010190565b600435906001600160a01b03821682036100f457565b602435906001600160a01b03821682036100f457565b6001600160a01b03169081156101de576001600160a01b031691821561047257815f525f60205260405f205481811061088f57817fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef92602092855f525f84520360405f2055845f525f825260405f20818154019055604051908152a3565b827fe450d38c000000000000000000000000000000000000000000000000000000005f5260045260245260445260645ffdfea2646970667358221220ef599a504a8ac25355fdaab329c6a2dd26a896974e8ced6f5fc876b4f34048fd64736f6c634300081d0033' as `0x${string}`,
};
