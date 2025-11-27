import { proxy } from 'valtio';
import type { Address } from 'viem';

export interface TAsset {
    token: Address;
    chainId: number;
    symbol: string;
    decimals: number;
}

export interface LedgerBalance {
    asset: string;
    amount: string;
}

export interface IAssetsState {
    assets: TAsset[] | null;
    ledgerBalances: LedgerBalance[] | null;
    assetsLoading: boolean;
    ledgerBalancesLoading: boolean;
    isFirstBalanceLoad: boolean;
}

const state = proxy<IAssetsState>({
    assets: [] as TAsset[] | null,
    ledgerBalances: null, // Start with null instead of empty array
    assetsLoading: false,
    ledgerBalancesLoading: false,
    isFirstBalanceLoad: true,
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
    setLedgerBalances(balances: LedgerBalance[]): void {
        if (Array.isArray(balances)) {
            state.ledgerBalances = [...balances];
        }

        state.isFirstBalanceLoad = false;
    },
    setLedgerBalancesLoading(loading: boolean) {
        state.ledgerBalancesLoading = loading;
    },
    setLedgerBalancesFirstLoading() {
        state.isFirstBalanceLoad = false;
    },
};

export default AssetsStore;
