import React, { useState, useEffect } from 'react';
import { useSnapshot } from 'valtio';
import SettingsStore from '@/store/SettingsStore';
import AssetsStore, { fetchAssets, fetchBalances, TAsset } from '@/store/AssetsStore';
import WalletStore from '@/store/WalletStore';
import { useMetaMask } from '@/hooks/useMetaMask';
import { chains } from '@/config/chains';
import { useNitroliteClient } from '@/hooks/useNitroliteClient';

interface MetaMaskConnectProps {
    onChannelOpen: (tokenAddress: string, amount: string) => void;
}

const MetaMaskConnect: React.FC<MetaMaskConnectProps> = ({ onChannelOpen }) => {
    const [step, setStep] = useState<1 | 2>(1);
    const [selectedToken, setSelectedToken] = useState<TAsset | null>(null);
    const [amount, setAmount] = useState<string>('');

    const settingsSnapshot = useSnapshot(SettingsStore.state);
    const assetsSnapshot = useSnapshot(AssetsStore.state);

    const {
        isMetaMaskInstalled,
        isConnected,
        account,
        error,
        connect: connectMetaMask,
        disconnect: disconnectMetaMask,
        switchNetwork,
    } = useMetaMask();

    // Connect to MetaMask
    const connectWallet = async () => {
        if (!isMetaMaskInstalled) {
            alert('Please install MetaMask to use this feature');
            return;
        }

        try {
            await connectMetaMask();
            setStep(2);
        } catch (error) {
            // Show error message to user
            alert(`Error connecting to MetaMask: ${error instanceof Error ? error.message : String(error)}`);
        }
    };

    // Disconnect from MetaMask
    const disconnectWallet = async () => {
        await disconnectMetaMask();
        setStep(1);
        setSelectedToken(null);
        setAmount('');

        // Display a message to the user about disconnecting in MetaMask if needed
        // We could add a toast notification here in a real app
    };

    // Handle chain change in dropdown
    const handleChainChange = (chainId: number) => {
        const selectedChain = chains.find((chain) => chain.id === chainId);

        if (selectedChain) {
            SettingsStore.setActiveChain(selectedChain);
            switchNetwork(chainId);
        }
    };

    // Open channel with selected token and amount
    const handleOpenChannel = () => {
        if (selectedToken && amount) {
            onChannelOpen(selectedToken.address, amount);
        }
    };

    // Load balances when connected or chain changes
    useEffect(() => {
        if (isConnected && account && settingsSnapshot.activeChain) {
            fetchBalances(account, settingsSnapshot.activeChain);
        }
    }, [isConnected, account, settingsSnapshot.activeChain]);

    // Load assets when component mounts or chain changes
    useEffect(() => {
        fetchAssets();
    }, [settingsSnapshot.activeChain]);

    // Find balance for selected token
    const getBalanceForToken = (token: TAsset): string => {
        if (!assetsSnapshot.balances) return '0';
        const tokenBalance = assetsSnapshot.balances.find((b) => b.symbol === token.symbol);

        return tokenBalance?.balance || '0';
    };

    // Display error message if there's an error
    useEffect(() => {
        if (error) {
            alert(error);
            WalletStore.clearError();
        }
    }, [error]);

    if (step === 1) {
        return (
            <div className="bg-white p-6 rounded-lg shadow-md border border-gray-200">
                <div className="text-center mb-6">
                    <h2 className="text-2xl font-bold mb-4">Welcome to Nitrolite</h2>
                    <p className="text-gray-600 mb-6 max-w-md mx-auto">
                        Experience secure state channels with fast, low-cost transactions without on-chain delays.
                    </p>
                </div>

                <div className="max-w-md mx-auto">
                    <h3 className="text-xl font-bold mb-4 text-[#3531ff]">Connect Your Wallet</h3>
                    <p className="text-gray-600 mb-4">
                        Connect your MetaMask wallet to open a channel and start using Nitrolite.
                    </p>
                    <button
                        onClick={connectWallet}
                        className="w-full bg-[#3531ff] hover:bg-[#2b28cc] text-white font-bold py-3 px-6 rounded transition-colors cursor-pointer shadow-sm">
                        Connect MetaMask
                    </button>

                    {!isMetaMaskInstalled && (
                        <p className="text-red-400 mt-4 text-center">
                            MetaMask is not installed. Please install the MetaMask extension to continue.
                        </p>
                    )}
                </div>
            </div>
        );
    }

    return (
        <div className="bg-white p-6 rounded-lg shadow-md border border-gray-200">
            <div className="flex justify-between items-center mb-4">
                <h2 className="text-xl font-bold">Select Token & Amount</h2>
                <div className="flex items-center space-x-2">
                    <span className="text-sm bg-white border border-gray-200 py-1 px-2 rounded font-mono text-gray-700 shadow-sm">
                        {account?.substring(0, 6)}...{account?.substring(38)}
                    </span>
                    <button
                        onClick={disconnectWallet}
                        className="bg-red-600 hover:bg-red-700 text-white text-sm py-1 px-2 rounded transition-colors cursor-pointer">
                        Disconnect
                    </button>
                </div>
            </div>

            <div className="mb-4">
                <label className="block text-gray-700 mb-2">Select Network</label>
                <select
                    className="w-full p-2 bg-white border border-gray-300 rounded text-gray-700"
                    value={settingsSnapshot.activeChain?.id || ''}
                    onChange={(e) => handleChainChange(Number(e.target.value))}>
                    {chains.map((chain) => (
                        <option key={chain.id} value={chain.id}>
                            {chain.name}
                        </option>
                    ))}
                </select>
            </div>

            <div className="mb-4">
                <label className="block text-gray-700 mb-2">Select Token</label>
                {assetsSnapshot.assetsLoading ? (
                    <div className="text-gray-500">Loading tokens...</div>
                ) : (
                    <select
                        className="w-full p-2 bg-white border border-gray-300 rounded text-gray-700"
                        value={selectedToken?.address || ''}
                        onChange={(e) => {
                            const token = assetsSnapshot.assets?.find((a) => a.address === e.target.value) || null;

                            setSelectedToken(token);
                        }}>
                        <option value="">Select a token</option>
                        {assetsSnapshot.assets?.map((asset) => (
                            <option key={asset.address} value={asset.address}>
                                {asset.symbol} - Balance: {getBalanceForToken(asset)}
                            </option>
                        ))}
                    </select>
                )}
            </div>

            <div className="mb-6">
                <label className="block text-gray-700 mb-2">Amount</label>
                <input
                    type="text"
                    className="w-full p-2 bg-white border border-gray-300 rounded text-gray-700"
                    value={amount}
                    onChange={(e) => setAmount(e.target.value)}
                    placeholder="Enter amount"
                />
                {selectedToken && (
                    <p className="text-sm text-gray-500 mt-1">
                        Available: {getBalanceForToken(selectedToken)} {selectedToken.symbol}
                    </p>
                )}
            </div>

            <button
                onClick={handleOpenChannel}
                disabled={!selectedToken || !amount || assetsSnapshot.balancesLoading}
                className={`w-full py-2 px-4 rounded font-bold ${
                    !selectedToken || !amount || assetsSnapshot.balancesLoading
                        ? 'bg-gray-200 text-gray-400 cursor-not-allowed'
                        : 'bg-[#3531ff] hover:bg-[#2b28cc] transition-colors cursor-pointer text-white'
                }`}>
                {assetsSnapshot.balancesLoading ? 'Loading Balances...' : 'Open Channel'}
            </button>
        </div>
    );
};

export default MetaMaskConnect;
