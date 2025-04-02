import { TAsset } from '@/store/AssetsStore';
import balanceCheckerAbi from '@/utils/balance.checker.abi.json';
import { Address, Chain, createPublicClient, formatUnits, http } from 'viem';

const BALANCE_CHECKER_ADDRESSES = {
    1: '0xb1f8e55c7f64d203c1400b9d8555d050f94adf39',
    11155111: '0xBfbCed302deD369855fc5f7668356e123ca4B329',
    10: '0xB1c568e9C3E6bdaf755A60c7418C269eb11524FC',
    56: '0x2352c63A83f9Fd126af8676146721Fa00924d7e4',
    137: '0x2352c63A83f9Fd126af8676146721Fa00924d7e4',
    324: '0x458fEd3144680a5b8bcfaa0F9594aa19B4Ea2D34',
    8453: '0x6AA75276052D96696134252587894ef5FFA520af',
    42161: '0x151E24A486D7258dd7C33Fb67E4bB01919B7B32c',
    43114: '0xD023D153a0DFa485130ECFdE2FAA7e612EF94818',
    59141: '0x424A079B89571a515Fd6fe0ba614060D5Fd8E16A',
    59144: '0xF62e6a41561b3650a69Bb03199C735e3E3328c0D',
    534352: '0x3192B36822DAe9069b41311D909D116F96EB876d',
} as { [chainId: number]: Address };

export const getBalances = async (
    assets: TAsset[],
    walletAddress: string,
    chain: Chain | undefined,
): Promise<any[]> => {
    if (!chain) {
        return [];
    }

    const publicClient = createPublicClient({
        transport: http(chain.rpcUrls.default.http[0]),
        chain,
    });

    const addresses = assets.map((asset) => asset.address);

    const balancesBigInt = (await publicClient.readContract({
        abi: balanceCheckerAbi,
        address: BALANCE_CHECKER_ADDRESSES[chain.id] as Address,
        functionName: 'balances',
        args: [[walletAddress], addresses],
    })) as bigint[];

    const balances = balancesBigInt.map((balanceBigInt: bigint, index: number) => {
        const asset = assets[index];

        return {
            symbol: asset.symbol,
            balance: formatUnits(balanceBigInt, asset.decimals),
        };
    });

    return balances;
};
