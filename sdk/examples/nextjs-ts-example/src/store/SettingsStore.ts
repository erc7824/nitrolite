import { Chain } from 'viem';
import { chains } from '@/config/chains';
import { proxy } from 'valtio';

const defaultChain = () => {
    // It doesn't work on linea sepolia
    // const chainByDevMode = process.env.NEXT_PUBLIC_DEV_MODE ? 59141 : 137;
    const chainByDevMode = 137;

    const chainId =
        typeof localStorage !== 'undefined' && Number(localStorage.getItem('chainId'))
            ? Number(localStorage.getItem('chainId'))
            : chainByDevMode;

    const chain = chains.find((chain) => chain.id === chainId);

    if (!chain) {
        return chains.find((chain) => chain.id === 137);
    }

    return chain;
};

/**
 * Types
 */
export interface ISettingsStore {
    testnets: boolean;
    activeChain: Chain | undefined;
    prevChainId: number;
}

/**
 * State
 */
const state = proxy<ISettingsStore>({
    testnets: typeof localStorage !== 'undefined' ? Boolean(localStorage.getItem('testnets')) : true,
    activeChain: defaultChain(),
    prevChainId: defaultChain()?.id ?? 0,
});

/**
 * Store / Actions
 */
const SettingsStore = {
    state,

    setActiveChain(value: Chain | undefined) {
        state.activeChain = value;
        localStorage.setItem('chainId', value ? value.id.toString() : '');
    },

    setPrevChainId(value: number | undefined) {
        state.prevChainId = value ?? 0;
    },

    toggleTestNets() {
        state.testnets = !state.testnets;
        if (state.testnets) {
            localStorage.setItem('testnets', 'true');
        } else {
            localStorage.removeItem('testnets');
        }
    },
};

export default SettingsStore;
