import { useState } from 'react';
import type { Client } from '@erc7824/nitrolite';
import type { StatusMessage } from '../types';

interface LowLevelSectionProps {
  client: Client;
  defaultAddress: string;
  showStatus: (type: StatusMessage['type'], message: string, details?: string) => void;
}

export default function LowLevelSection({ client, defaultAddress, showStatus }: LowLevelSectionProps) {
  const [stateWallet, setStateWallet] = useState(defaultAddress);
  const [stateAsset, setStateAsset] = useState('usdc');

  const [homeChannelWallet, setHomeChannelWallet] = useState(defaultAddress);
  const [homeChannelAsset, setHomeChannelAsset] = useState('usdc');

  const [escrowChannelId, setEscrowChannelId] = useState('');

  const [result, setResult] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [activeView, setActiveView] = useState<string | null>(null);

  const handleGetState = async () => {
    try {
      setLoading(true);
      const state = await client.getLatestState(stateWallet as `0x${string}`, stateAsset, false);
      setResult(state);
      setActiveView('state');
      showStatus('success', 'State retrieved');
    } catch (error) {
      showStatus('error', 'Failed to get state', error instanceof Error ? error.message : String(error));
    } finally {
      setLoading(false);
    }
  };

  const handleGetHomeChannel = async () => {
    try {
      setLoading(true);
      const channel = await client.getHomeChannel(homeChannelWallet as `0x${string}`, homeChannelAsset);
      setResult(channel);
      setActiveView('homeChannel');
      showStatus('success', 'Home channel retrieved');
    } catch (error) {
      showStatus('error', 'Failed to get home channel', error instanceof Error ? error.message : String(error));
    } finally {
      setLoading(false);
    }
  };

  const handleGetEscrowChannel = async () => {
    try {
      setLoading(true);
      const channel = await client.getEscrowChannel(escrowChannelId);
      setResult(channel);
      setActiveView('escrowChannel');
      showStatus('success', 'Escrow channel retrieved');
    } catch (error) {
      showStatus('error', 'Failed to get escrow channel', error instanceof Error ? error.message : String(error));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      <h2 className="text-2xl font-bold mb-4 text-gray-800">Low-Level State Management</h2>
      <p className="text-sm text-gray-600 mb-6">Direct access to channel states and information</p>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        {/* Get Latest State */}
        <div className="border border-gray-200 rounded p-4">
          <h3 className="font-semibold mb-3 text-gray-700">Get Latest State</h3>
          <div className="space-y-3">
            <input
              type="text"
              value={stateWallet}
              onChange={(e) => setStateWallet(e.target.value)}
              placeholder="Wallet address"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
            />
            <input
              type="text"
              value={stateAsset}
              onChange={(e) => setStateAsset(e.target.value)}
              placeholder="Asset"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
            />
            <button
              onClick={handleGetState}
              disabled={loading || !stateWallet || !stateAsset}
              className="w-full bg-blue-500 hover:bg-blue-600 text-white py-2 rounded font-medium transition disabled:bg-gray-400 text-sm"
            >
              {loading ? 'Loading...' : 'Get State'}
            </button>
          </div>
        </div>

        {/* Get Home Channel */}
        <div className="border border-gray-200 rounded p-4">
          <h3 className="font-semibold mb-3 text-gray-700">Get Home Channel</h3>
          <div className="space-y-3">
            <input
              type="text"
              value={homeChannelWallet}
              onChange={(e) => setHomeChannelWallet(e.target.value)}
              placeholder="Wallet address"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
            />
            <input
              type="text"
              value={homeChannelAsset}
              onChange={(e) => setHomeChannelAsset(e.target.value)}
              placeholder="Asset"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
            />
            <button
              onClick={handleGetHomeChannel}
              disabled={loading || !homeChannelWallet || !homeChannelAsset}
              className="w-full bg-blue-500 hover:bg-blue-600 text-white py-2 rounded font-medium transition disabled:bg-gray-400 text-sm"
            >
              {loading ? 'Loading...' : 'Get Channel'}
            </button>
          </div>
        </div>

        {/* Get Escrow Channel */}
        <div className="border border-gray-200 rounded p-4">
          <h3 className="font-semibold mb-3 text-gray-700">Get Escrow Channel</h3>
          <div className="space-y-3">
            <input
              type="text"
              value={escrowChannelId}
              onChange={(e) => setEscrowChannelId(e.target.value)}
              placeholder="Escrow Channel ID"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
            />
            <button
              onClick={handleGetEscrowChannel}
              disabled={loading || !escrowChannelId}
              className="w-full bg-blue-500 hover:bg-blue-600 text-white py-2 rounded font-medium transition disabled:bg-gray-400 text-sm"
            >
              {loading ? 'Loading...' : 'Get Channel'}
            </button>
          </div>
        </div>
      </div>

      {/* Results Display */}
      {result && (
        <div className="mt-4 bg-gray-50 rounded p-4 border border-gray-200 max-h-96 overflow-y-auto">
          <h3 className="font-semibold mb-3 text-gray-700">Results:</h3>

          {activeView === 'state' && (
            <div className="space-y-2 text-sm">
              <div><span className="font-medium">Version:</span> {result.version?.toString()}</div>
              <div><span className="font-medium">Epoch:</span> {result.epoch?.toString()}</div>
              <div><span className="font-medium">State ID:</span> {result.id}</div>
              {result.homeChannelId && <div><span className="font-medium">Channel:</span> {result.homeChannelId}</div>}
              <div className="mt-3 font-medium">Home Ledger:</div>
              <div className="ml-4 space-y-1">
                <div>Chain: {(result.homeLedger?.blockchainId || result.homeLedger?.blockchain_id)?.toString()}</div>
                <div>Token: {result.homeLedger?.tokenAddress || result.homeLedger?.token_address}</div>
                <div>User Balance: {result.homeLedger?.userBalance?.toString() || result.homeLedger?.user_balance?.toString()}</div>
                <div>Node Balance: {result.homeLedger?.nodeBalance?.toString() || result.homeLedger?.node_balance?.toString()}</div>
                <div>User NetFlow: {result.homeLedger?.userNetFlow?.toString() || result.homeLedger?.user_net_flow?.toString()}</div>
                <div>Node NetFlow: {result.homeLedger?.nodeNetFlow?.toString() || result.homeLedger?.node_net_flow?.toString()}</div>
              </div>
              {result.transitions && result.transitions.length > 0 && (
                <>
                  <div className="mt-3 font-medium">Transitions: {result.transitions.length}</div>
                  {result.transitions.map((t: any, idx: number) => (
                    <div key={idx} className="ml-4 text-xs">
                      {idx + 1}. {t.type} (Amount: {t.amount?.toString()})
                    </div>
                  ))}
                </>
              )}
            </div>
          )}

          {(activeView === 'homeChannel' || activeView === 'escrowChannel') && (
            <div className="space-y-2 text-sm">
              <div><span className="font-medium">Channel ID:</span> {result.channelId || result.channel_id}</div>
              <div><span className="font-medium">User Wallet:</span> {result.userWallet || result.user_wallet}</div>
              <div><span className="font-medium">Type:</span> {result.type}</div>
              <div><span className="font-medium">Status:</span> {result.status}</div>
              <div><span className="font-medium">Version:</span> {(result.stateVersion || result.state_version)?.toString()}</div>
              <div><span className="font-medium">Nonce:</span> {result.nonce?.toString()}</div>
              <div><span className="font-medium">Chain ID:</span> {(result.blockchainId || result.blockchain_id)?.toString()}</div>
              <div><span className="font-medium">Token:</span> {result.tokenAddress || result.token_address}</div>
              <div><span className="font-medium">Challenge Duration:</span> {(result.challengeDuration || result.challenge_duration)?.toString()}s</div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
