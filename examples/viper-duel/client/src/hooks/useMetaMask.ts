import { useState, useEffect, useCallback } from 'react';
import { ethers } from 'ethers';

interface MetaMaskState {
  isConnected: boolean;
  address: string;
  provider: ethers.BrowserProvider | null;
  isConnecting: boolean;
  error: string | null;
}

export function useMetaMask() {
  const [state, setState] = useState<MetaMaskState>({
    isConnected: false,
    address: '',
    provider: null,
    isConnecting: false,
    error: null,
  });

  // Check if MetaMask is installed
  const checkIfMetaMaskInstalled = useCallback((): boolean => {
    const { ethereum } = window as any;
    return Boolean(ethereum && ethereum.isMetaMask);
  }, []);

  // Connect to MetaMask
  const connectWallet = useCallback(async () => {
    try {
      setState(prev => ({ ...prev, isConnecting: true, error: null }));
      
      if (!checkIfMetaMaskInstalled()) {
        throw new Error('MetaMask is not installed');
      }

      const { ethereum } = window as any;
      const provider = new ethers.BrowserProvider(ethereum);
      
      // Request accounts access
      const accounts = await provider.send("eth_requestAccounts", []);
      const address = ethers.getAddress(accounts[0]);
      
      // Update state with connection details
      setState({
        isConnected: true,
        address,
        provider,
        isConnecting: false,
        error: null,
      });

      return address;
    } catch (error: any) {
      console.error('Error connecting to MetaMask:', error);
      setState(prev => ({
        ...prev,
        isConnecting: false,
        isConnected: false,
        error: error.message || 'Failed to connect to MetaMask',
      }));
      return null;
    }
  }, [checkIfMetaMaskInstalled]);

  // Disconnect from MetaMask
  const disconnectWallet = useCallback(() => {
    setState({
      isConnected: false,
      address: '',
      provider: null,
      isConnecting: false,
      error: null,
    });
  }, []);

  // Listen for account changes
  useEffect(() => {
    if (!checkIfMetaMaskInstalled()) return;

    const { ethereum } = window as any;
    
    const handleAccountsChanged = (accounts: string[]) => {
      if (accounts.length === 0) {
        // User disconnected their wallet
        disconnectWallet();
      } else if (ethers.getAddress(accounts[0]) !== state.address) {
        // Account changed, update state
        setState(prev => ({
          ...prev,
          address: ethers.getAddress(accounts[0]),
          isConnected: true,
        }));
      }
    };

    const handleChainChanged = () => {
      // Handle chain change by refreshing the page
      window.location.reload();
    };

    // Subscribe to events
    ethereum.on('accountsChanged', handleAccountsChanged);
    ethereum.on('chainChanged', handleChainChanged);

    // Check if already connected
    ethereum.request({ method: 'eth_accounts' })
      .then((accounts: string[]) => {
        if (accounts.length > 0) {
          const provider = new ethers.BrowserProvider(ethereum);
          setState({
            isConnected: true,
            address: ethers.getAddress(accounts[0]),
            provider,
            isConnecting: false,
            error: null,
          });
        }
      })
      .catch((err: Error) => console.error('Error checking accounts:', err));

    // Cleanup listeners on unmount
    return () => {
      if (ethereum.removeListener) {
        ethereum.removeListener('accountsChanged', handleAccountsChanged);
        ethereum.removeListener('chainChanged', handleChainChanged);
      }
    };
  }, [checkIfMetaMaskInstalled, disconnectWallet, state.address]);

  return {
    ...state,
    connectWallet,
    disconnectWallet,
    isMetaMaskInstalled: checkIfMetaMaskInstalled(),
  };
}