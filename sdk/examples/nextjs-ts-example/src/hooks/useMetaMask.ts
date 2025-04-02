import { useState, useEffect } from 'react';
import { useSnapshot } from 'valtio';
import WalletStore from '@/store/WalletStore';
import SettingsStore from '@/store/SettingsStore';
import { Address } from 'viem';

export function useMetaMask() {
  const [isMetaMaskInstalled, setIsMetaMaskInstalled] = useState<boolean>(false);
  const walletSnap = useSnapshot(WalletStore.state);
  const settingsSnap = useSnapshot(SettingsStore.state);

  // Check if MetaMask is installed
  useEffect(() => {
    if (typeof window !== 'undefined') {
      setIsMetaMaskInstalled(!!window.ethereum?.isMetaMask);
    }
  }, []);

  // Connect to MetaMask
  const connect = async () => {
    if (!window.ethereum) {
      WalletStore.setError('MetaMask is not installed');
      return;
    }

    try {
      // Request accounts
      const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
      const chainIdHex = await window.ethereum.request({ method: 'eth_chainId' });
      const chainId = parseInt(chainIdHex, 16);
      
      // Update store
      WalletStore.connect(accounts[0] as Address, chainId);
      
      // Check if we need to switch networks
      if (settingsSnap.activeChain && settingsSnap.activeChain.id !== chainId) {
        await switchNetwork(settingsSnap.activeChain.id);
      }
    } catch (error) {
      console.error('Error connecting to MetaMask:', error);
      WalletStore.setError('Failed to connect to MetaMask');
    }
  };

  // Disconnect from MetaMask
  const disconnect = async () => {
    // Note: MetaMask doesn't have a direct method to disconnect via the provider API
    // The best practice is to clear the local state
    WalletStore.disconnect();
    
    // MetaMask may implement a disconnect method in the future
    // For now, we'll let the user know they need to disconnect manually if they want to connect a different account
    if (window.ethereum && window.ethereum.isMetaMask) {
      console.log("MetaMask connection state cleared");
      // You can also show a notification to the user if needed
    }
  };

  // Switch network
  const switchNetwork = async (chainId: number) => {
    if (!window.ethereum) return;
    
    try {
      // Try switching to the network
      await window.ethereum.request({
        method: 'wallet_switchEthereumChain',
        params: [{ chainId: `0x${chainId.toString(16)}` }],
      });
      
      // Update the store with the new chain ID
      WalletStore.setChainId(chainId);
    } catch (error: any) {
      // If the network doesn't exist in MetaMask, we need to add it
      if (error.code === 4902) {
        // This would need chain-specific data like name, RPC URLs, etc.
        // Consider adding logic to add networks if needed
        console.error('Network not available in MetaMask');
        WalletStore.setError('Network not available in MetaMask');
      } else {
        console.error('Error switching network:', error);
        WalletStore.setError('Failed to switch network');
      }
    }
  };

  // Listen for account changes
  useEffect(() => {
    if (window.ethereum) {
      const handleAccountsChanged = (accounts: string[]) => {
        if (accounts.length === 0) {
          // User disconnected their wallet
          WalletStore.disconnect();
        } else if (accounts[0] !== walletSnap.account) {
          // Account changed
          WalletStore.connect(accounts[0] as Address, walletSnap.chainId || 1);
        }
      };

      const handleChainChanged = (chainIdHex: string) => {
        const chainId = parseInt(chainIdHex, 16);
        WalletStore.setChainId(chainId);
      };

      window.ethereum.on('accountsChanged', handleAccountsChanged);
      window.ethereum.on('chainChanged', handleChainChanged);

      return () => {
        window.ethereum.removeListener('accountsChanged', handleAccountsChanged);
        window.ethereum.removeListener('chainChanged', handleChainChanged);
      };
    }
  }, [walletSnap.account, walletSnap.chainId]);

  return {
    isMetaMaskInstalled,
    isConnected: walletSnap.connected,
    account: walletSnap.account,
    chainId: walletSnap.chainId,
    error: walletSnap.error,
    connect,
    disconnect,
    switchNetwork,
  };
}