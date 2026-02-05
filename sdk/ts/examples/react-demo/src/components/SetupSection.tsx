import { useState } from 'react';
import type { AppState, StatusMessage } from '../types';
import { formatAddress } from '../utils';

interface SetupSectionProps {
  appState: AppState;
  onConnectWallet: () => void;
  onConnectNode: () => void;
  onDisconnect: () => void;
  onDisconnectWallet: () => void;
  onNodeUrlChange: (url: string) => void;
  onAddRpc: (chainId: string, rpcUrl: string) => void;
  onRemoveRpc: (chainId: string) => void;
  onAddHomeBlockchain: (asset: string, chainId: string) => void;
  onRemoveHomeBlockchain: (asset: string) => void;
  showStatus: (type: StatusMessage['type'], message: string, details?: string) => void;
}

export default function SetupSection({
  appState,
  onConnectWallet,
  onConnectNode,
  onDisconnect,
  onDisconnectWallet,
  onNodeUrlChange,
  onAddRpc,
  onRemoveRpc,
  onAddHomeBlockchain,
  onRemoveHomeBlockchain,
  showStatus,
}: SetupSectionProps) {
  const [chainId, setChainId] = useState('11155111');
  const [rpcUrl, setRpcUrl] = useState('');
  const [homeAsset, setHomeAsset] = useState('');
  const [homeChainId, setHomeChainId] = useState('11155111');

  const handleAddRpc = () => {
    if (!chainId || !rpcUrl) {
      showStatus('error', 'Both Chain ID and RPC URL are required');
      return;
    }
    onAddRpc(chainId, rpcUrl);
    showStatus('success', `RPC configured for chain ${chainId}`);
    setRpcUrl('');
  };

  const handleAddHomeBlockchain = () => {
    if (!homeAsset || !homeChainId) {
      showStatus('error', 'Both Asset and Chain ID are required');
      return;
    }
    onAddHomeBlockchain(homeAsset.toLowerCase(), homeChainId);
    showStatus('success', `Home blockchain set for ${homeAsset} to chain ${homeChainId}`);
    setHomeAsset('');
  };

  return (
    <div className="bg-white rounded-lg shadow-md p-6 mb-6">
      <h2 className="text-2xl font-bold mb-4 text-gray-800">Setup & Configuration</h2>

      {/* Wallet Connection */}
      <div className="mb-6 p-4 bg-gray-50 rounded border border-gray-200">
        <h3 className="font-semibold mb-3 text-gray-700">1. Connect Wallet</h3>
        <div className="flex items-center gap-4">
          {!appState.address ? (
            <button
              onClick={onConnectWallet}
              className="bg-orange-500 hover:bg-orange-600 text-white px-6 py-2 rounded font-medium transition"
            >
              Connect MetaMask
            </button>
          ) : (
            <div className="flex items-center gap-2">
              <span className="text-sm text-gray-600">Connected:</span>
              <code className="bg-gray-200 px-3 py-1 rounded text-sm font-mono">
                {formatAddress(appState.address)}
              </code>
              <span className="w-2 h-2 bg-green-500 rounded-full"></span>
              <button
                onClick={onDisconnectWallet}
                className="text-red-500 hover:text-red-700 text-sm px-2 py-1 border border-red-300 rounded"
              >
                Disconnect
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Node URL Configuration */}
      <div className="mb-6 p-4 bg-gray-50 rounded border border-gray-200">
        <h3 className="font-semibold mb-3 text-gray-700">2. Configure Clearnode URL</h3>
        <div className="flex gap-2">
          <input
            type="text"
            value={appState.nodeUrl}
            onChange={(e) => onNodeUrlChange(e.target.value)}
            disabled={appState.connected}
            placeholder="wss://clearnode.example.com/ws"
            className="flex-1 px-4 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100"
          />
        </div>
      </div>

      {/* RPC Configuration */}
      <div className="mb-6 p-4 bg-gray-50 rounded border border-gray-200">
        <h3 className="font-semibold mb-3 text-gray-700">3. Blockchain RPCs (Pre-configured)</h3>
        <p className="text-xs text-gray-500 mb-3">
          ✓ Public RPCs already configured for Polygon Amoy, Ethereum Sepolia, Base Sepolia, and Arbitrum Sepolia
        </p>
        <div className="space-y-3">
          <div className="flex gap-2">
            <input
              type="text"
              value={chainId}
              onChange={(e) => setChainId(e.target.value)}
              placeholder="Chain ID (e.g., 80002)"
              className="w-32 px-4 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <input
              type="text"
              value={rpcUrl}
              onChange={(e) => setRpcUrl(e.target.value)}
              placeholder="RPC URL (e.g., https://polygon-amoy.g.alchemy.com/v2/KEY)"
              className="flex-1 px-4 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <button
              onClick={handleAddRpc}
              className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded font-medium transition"
            >
              Add RPC
            </button>
          </div>

          {Object.entries(appState.rpcConfigs).length > 0 && (
            <div className="mt-2 space-y-1">
              <p className="text-sm text-gray-600 mb-2">Configured RPCs:</p>
              {Object.entries(appState.rpcConfigs).map(([id, url]) => (
                <div key={id} className="flex items-center gap-2 text-sm bg-white p-2 rounded">
                  <span className="font-semibold text-gray-700">Chain {id}:</span>
                  <code className="flex-1 text-gray-600 truncate">{url}</code>
                  <button
                    onClick={() => onRemoveRpc(id)}
                    className="text-red-500 hover:text-red-700 px-2"
                  >
                    Remove
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Home Blockchain Configuration (Optional) */}
      <div className="mb-6 p-4 bg-gray-50 rounded border border-gray-200">
        <h3 className="font-semibold mb-3 text-gray-700">Home Blockchain Selection (Optional)</h3>
        <p className="text-xs text-gray-500 mb-3">
          Configure which blockchain to use as the "home" blockchain for each asset. This determines where deposits/withdrawals settle. If not set, you'll be prompted when needed.
        </p>
        <div className="space-y-3">
          <div className="flex gap-2">
            <input
              type="text"
              value={homeAsset}
              onChange={(e) => setHomeAsset(e.target.value)}
              placeholder="Asset (e.g., usdc)"
              className="w-32 px-4 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <input
              type="text"
              value={homeChainId}
              onChange={(e) => setHomeChainId(e.target.value)}
              placeholder="Chain ID (e.g., 11155111)"
              className="flex-1 px-4 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <button
              onClick={handleAddHomeBlockchain}
              disabled={appState.connected}
              className="bg-purple-500 hover:bg-purple-600 text-white px-4 py-2 rounded font-medium transition disabled:bg-gray-400"
            >
              Set Home
            </button>
          </div>

          {Object.entries(appState.homeBlockchains).length > 0 && (
            <div className="mt-2 space-y-1">
              <p className="text-sm text-gray-600 mb-2">Configured Home Blockchains:</p>
              {Object.entries(appState.homeBlockchains).map(([asset, chainId]) => (
                <div key={asset} className="flex items-center gap-2 text-sm bg-white p-2 rounded">
                  <span className="font-semibold text-gray-700">{asset}:</span>
                  <code className="flex-1 text-gray-600">Chain {chainId}</code>
                  <button
                    onClick={() => onRemoveHomeBlockchain(asset)}
                    disabled={appState.connected}
                    className="text-red-500 hover:text-red-700 px-2 disabled:text-gray-400"
                  >
                    Remove
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Connect/Disconnect */}
      <div className="flex gap-3">
        {!appState.connected ? (
          <button
            onClick={onConnectNode}
            disabled={!appState.address}
            className="bg-green-500 hover:bg-green-600 text-white px-8 py-3 rounded font-medium transition disabled:bg-gray-400 disabled:cursor-not-allowed"
          >
            4. Connect to Node
          </button>
        ) : (
          <>
            <div className="flex items-center gap-2 bg-green-100 px-4 py-2 rounded">
              <span className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
              <span className="text-green-800 font-medium">Connected to Node</span>
            </div>
            <button
              onClick={onDisconnect}
              className="bg-red-500 hover:bg-red-600 text-white px-6 py-2 rounded font-medium transition"
            >
              Disconnect
            </button>
          </>
        )}
      </div>
    </div>
  );
}
