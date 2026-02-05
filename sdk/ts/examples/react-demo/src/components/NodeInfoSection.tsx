import { useState } from 'react';
import type { Client } from '@erc7824/nitrolite';
import type { StatusMessage } from '../types';
import { safeStringify } from '../utils';

interface NodeInfoSectionProps {
  client: Client;
  showStatus: (type: StatusMessage['type'], message: string, details?: string) => void;
}

export default function NodeInfoSection({ client, showStatus }: NodeInfoSectionProps) {
  const [result, setResult] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [activeView, setActiveView] = useState<string | null>(null);
  const [chainIdFilter, setChainIdFilter] = useState('');

  const handlePing = async () => {
    try {
      setLoading(true);
      await client.ping();
      showStatus('success', 'Ping successful', 'Node is responding');
      setResult(null);
      setActiveView(null);
    } catch (error) {
      showStatus('error', 'Ping failed', error instanceof Error ? error.message : String(error));
    } finally {
      setLoading(false);
    }
  };

  const handleNodeInfo = async () => {
    try {
      setLoading(true);
      const config = await client.getConfig();
      setResult(config);
      setActiveView('nodeInfo');
      showStatus('success', 'Node info retrieved');
    } catch (error) {
      showStatus('error', 'Failed to get node info', error instanceof Error ? error.message : String(error));
    } finally {
      setLoading(false);
    }
  };

  const handleListChains = async () => {
    try {
      setLoading(true);
      const chains = await client.getBlockchains();
      setResult(chains);
      setActiveView('chains');
      showStatus('success', `Found ${chains.length} blockchains`);
    } catch (error) {
      showStatus('error', 'Failed to list chains', error instanceof Error ? error.message : String(error));
    } finally {
      setLoading(false);
    }
  };

  const handleListAssets = async () => {
    try {
      setLoading(true);
      const chainId = chainIdFilter ? BigInt(chainIdFilter) : undefined;
      const assets = await client.getAssets(chainId);
      setResult(assets);
      setActiveView('assets');
      showStatus('success', `Found ${assets.length} assets`);
    } catch (error) {
      showStatus('error', 'Failed to list assets', error instanceof Error ? error.message : String(error));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      <h2 className="text-2xl font-bold mb-4 text-gray-800">Node Information</h2>
      <p className="text-sm text-gray-600 mb-6">Query node configuration and capabilities</p>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-6">
        <button
          onClick={handlePing}
          disabled={loading}
          className="bg-blue-500 hover:bg-blue-600 text-white py-2 px-4 rounded font-medium transition disabled:bg-gray-400"
        >
          Ping
        </button>
        <button
          onClick={handleNodeInfo}
          disabled={loading}
          className="bg-blue-500 hover:bg-blue-600 text-white py-2 px-4 rounded font-medium transition disabled:bg-gray-400"
        >
          Node Info
        </button>
        <button
          onClick={handleListChains}
          disabled={loading}
          className="bg-blue-500 hover:bg-blue-600 text-white py-2 px-4 rounded font-medium transition disabled:bg-gray-400"
        >
          List Chains
        </button>
        <button
          onClick={handleListAssets}
          disabled={loading}
          className="bg-blue-500 hover:bg-blue-600 text-white py-2 px-4 rounded font-medium transition disabled:bg-gray-400"
        >
          List Assets
        </button>
      </div>

      {/* Assets Filter */}
      {activeView === 'assets' && (
        <div className="mb-4">
          <input
            type="text"
            value={chainIdFilter}
            onChange={(e) => setChainIdFilter(e.target.value)}
            placeholder="Filter by Chain ID (optional)"
            className="px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
      )}

      {/* Results Display */}
      {result && (
        <div className="mt-4 bg-gray-50 rounded p-4 border border-gray-200">
          <h3 className="font-semibold mb-2 text-gray-700">Results:</h3>
          <details className="mb-2 text-xs">
            <summary className="cursor-pointer text-gray-500 hover:text-gray-700">View raw JSON</summary>
            <pre className="mt-2 p-2 bg-gray-100 rounded overflow-auto max-h-40">
              {safeStringify(result, 2)}
            </pre>
          </details>
          {activeView === 'nodeInfo' && (
            <div className="space-y-2">
              <div><span className="font-medium">Address:</span> {result.nodeAddress || result.node_address}</div>
              <div><span className="font-medium">Version:</span> {result.nodeVersion || result.node_version}</div>
              <div><span className="font-medium">Chains:</span> {result.blockchains?.length || 0}</div>
              {result.blockchains && result.blockchains.length > 0 && (
                <div className="mt-3">
                  <div className="font-medium mb-1">Supported Blockchains:</div>
                  {result.blockchains.map((bc: any, idx: number) => (
                    <div key={bc.id || bc.blockchain_id || idx} className="ml-4 text-sm">
                      • {bc.name} (ID: {(bc.id || bc.blockchain_id)?.toString()}) - Contract: {bc.contractAddress || bc.contract_address}
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {activeView === 'chains' && (
            <div className="space-y-3">
              {result.map((chain: any, idx: number) => (
                <div key={chain.id || chain.blockchain_id || idx} className="border-l-4 border-blue-500 pl-3">
                  <div className="font-medium">{chain.name}</div>
                  <div className="text-sm text-gray-600">Chain ID: {(chain.id || chain.blockchain_id)?.toString()}</div>
                  <div className="text-sm text-gray-600">Contract: {chain.contractAddress || chain.contract_address}</div>
                </div>
              ))}
            </div>
          )}

          {activeView === 'assets' && (
            <div className="space-y-3">
              {result.map((asset: any, idx: number) => (
                <div key={asset.symbol || idx} className="border-l-4 border-purple-500 pl-3">
                  <div className="font-medium">{asset.name} ({asset.symbol})</div>
                  <div className="text-sm text-gray-600">Decimals: {asset.decimals}</div>
                  <div className="text-sm text-gray-600">Tokens: {asset.tokens?.length || 0} connected</div>
                  {asset.tokens && asset.tokens.length > 0 && (
                    <div className="mt-1 ml-3 text-xs">
                      {asset.tokens.map((token: any, tidx: number) => (
                        <div key={tidx}>
                          Chain {(token.blockchainId || token.blockchain_id)?.toString()}: {token.address} (decimals: {token.decimals})
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
