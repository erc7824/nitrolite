import { proxy, subscribe } from 'valtio';

import SettingsStore from './SettingsStore';
import axios from 'axios';
import { Address, Chain } from 'viem';
import { getBalances } from '@/hooks/wallet/getTokenBalances';

export interface TAsset {
    name: string;
    address: Address;
    symbol: string;
    decimals: number;
    chainId: number;
    logoURI?: string;
    precision?: number;
    extensions?: Record<string, unknown>;
}

export interface TBalance {
    symbol: string;
    balance: string;
}

export interface IAssetsState {
    assets: TAsset[] | null;
    balances: TBalance[] | null;
    assetsLoading: boolean;
    balancesLoading: boolean;
}

const state = proxy<IAssetsState>({
    assets: [] as TAsset[] | null,
    balances: [] as TBalance[] | null,
    assetsLoading: false,
    balancesLoading: false,
});

const AssetsStore = {
    state,
    setAssets(assets: TAsset[] | null) {
        state.assets = assets;
        state.assetsLoading = false;
    },
    setAssetsLoading(loading: boolean) {
        state.assetsLoading = loading;
    },
    setBalances(balances: TBalance[] | null) {
        state.balances = balances;
        state.balancesLoading = false;
    },
    setBalancesLoading(loading: boolean) {
        state.balancesLoading = loading;
    },
};

export const fetchAssets = async () => {
    const assetsUrl = process.env.NEXT_PUBLIC_ASSETS_URL?.replace(
        '{chain_id}',
        String(SettingsStore.state.activeChain?.id ?? ''),
    );

    if (!assetsUrl) {
        AssetsStore.setAssets([
            {
                name: 'Test Token',
                address: '0xe7f1725e7734ce288f8367e1bb143e90bb3f0512',
                symbol: 'TST',
                decimals: 8,
                chainId: 1337,
            },
        ]);
        return;
    }

    AssetsStore.setAssetsLoading(true);
    let assets: TAsset[] = [];

    try {
        const response = await axios(assetsUrl);

        assets = response.data.tokens || [];
    } catch (error) {
        console.error('Error fetching assets:', error);
    } finally {
        AssetsStore.setAssetsLoading(false);
    }

    AssetsStore.setAssets(assets);
};

export const fetchBalances = async (accountAddress?: string, activeChain?: Chain) => {
    if (!AssetsStore.state.assets || !accountAddress || !activeChain) {
        AssetsStore.setBalances([]);

        return;
    }

    AssetsStore.setBalancesLoading(true);

    try {
        const newBalances = await getBalances(AssetsStore.state.assets, accountAddress, activeChain);

        AssetsStore.setBalances(newBalances);
    } catch (error) {
        console.error('Error fetching balances:', error);
    } finally {
        AssetsStore.setBalancesLoading(false);
    }
};

let prevActiveChain = SettingsStore.state.activeChain;

subscribe(SettingsStore.state, () => {
    if (AssetsStore.state.assets?.length === 0) {
        fetchAssets();
    } else if (SettingsStore.state.activeChain !== prevActiveChain) {
        prevActiveChain = SettingsStore.state.activeChain;
        fetchAssets();
    }
});

export default AssetsStore;
