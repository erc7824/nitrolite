import { useState, useEffect } from 'react';
import { type Address } from 'viem';
import { useAccountModal, useConnectModal } from '@rainbow-me/rainbowkit';
import { useAccount, useConnect, useSwitchChain } from 'wagmi';
import { SettingsStore, WalletStore } from '../store';
import { useStore } from '../store/storeUtils';

// Storage key for wallet connection
const WALLET_CONNECTION_KEY = 'wallet_connection';

export function useMetaMask() {
    const [isMetaMaskInstalled, setIsMetaMaskInstalled] =
        useState<boolean>(false);
    const walletState = useStore(WalletStore.state);
    const settingsState = useStore(SettingsStore.state);

    const { openConnectModal } = useConnectModal();
    const { openAccountModal } = useAccountModal();

    const { address, isConnected, isConnecting, chain } = useAccount();
    const { connect, connectors } = useConnect();
    const { switchChain } = useSwitchChain();

    useEffect(() => {
        if (typeof window !== 'undefined') {
            setIsMetaMaskInstalled(!!window.ethereum?.isMetaMask);
        }
    }, []);

    useEffect(() => {
        if (isConnected && address) {
            if (
                !walletState.isConnected ||
                walletState.walletAddress !== address
            ) {
                WalletStore.connect(address as Address);
                localStorage.setItem(WALLET_CONNECTION_KEY, 'true');
            }

            if (chain && walletState.chainId !== chain.id) {
                WalletStore.setChainId(chain.id);
            }
        } else if (!isConnected && walletState.isConnected) {
            localStorage.removeItem(WALLET_CONNECTION_KEY);
        }
    }, [
        isConnected,
        address,
        chain,
        walletState.isConnected,
        walletState.walletAddress,
        walletState.chainId,
    ]);

    // Connect using RainbowKit
    const connectWallet = async () => {
        try {
            // Find MetaMask connector
            const metamaskConnector = connectors.find((connector) =>
                connector.name.toLowerCase().includes('metamask')
            );

            if (metamaskConnector) {
                connect({ connector: metamaskConnector });
            } else {
                openConnectModal?.();
            }
        } catch (error) {
            console.error('Error connecting wallet:', error);
            WalletStore.setError('Failed to connect wallet');
        }
    };

    const switchNetwork = async (chainId: number) => {
        try {
            await switchChain({ chainId });
        } catch (error) {
            console.error('Error switching network:', error);
            WalletStore.setError('Failed to switch network');
        }
    };

    useEffect(() => {
        if (
            settingsState.activeChain &&
            chain &&
            settingsState.activeChain !== chain.id &&
            isConnected
        ) {
            WalletStore.setChainId(chain.id);
        }
    }, [settingsState.activeChain, chain, isConnected]);

    return {
        isMetaMaskInstalled,
        isConnected: walletState.isConnected,
        account: walletState.walletAddress,
        address: walletState.walletAddress,
        chainId: walletState.chainId,
        error: walletState.error,
        connect: connectWallet,
        connectWallet,
        openConnectModal,
        openAccountModal,
        switchNetwork,
        isConnecting,
    };
}