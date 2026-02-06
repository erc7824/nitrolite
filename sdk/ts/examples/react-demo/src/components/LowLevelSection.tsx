import { useState } from 'react';
import { Database, Home, Lock } from 'lucide-react';
import type { Client } from '@erc7824/nitrolite';
import type { StatusMessage } from '../types';
import { CollapsibleCard } from './ui/collapsible-card';
import { Button } from './ui/button';
import { Input } from './ui/input';

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
    <CollapsibleCard
      title="Low-Level State Management"
      description="Direct access to channel states and information"
      defaultOpen={false}
    >
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        {/* Get Latest State */}
        <div className="border border-border p-4 space-y-3">
          <div className="flex items-center gap-2">
            <Database className="h-4 w-4" />
            <h3 className="text-sm font-semibold uppercase tracking-wider">Get Latest State</h3>
          </div>
          <Input
            type="text"
            value={stateWallet}
            onChange={(e) => setStateWallet(e.target.value)}
            placeholder="Wallet address"
            className="font-mono text-xs"
          />
          <Input
            type="text"
            value={stateAsset}
            onChange={(e) => setStateAsset(e.target.value)}
            placeholder="Asset"
          />
          <Button
            onClick={handleGetState}
            disabled={loading || !stateWallet || !stateAsset}
            className="w-full"
          >
            {loading ? 'Loading...' : 'Get State'}
          </Button>
        </div>

        {/* Get Home Channel */}
        <div className="border border-border p-4 space-y-3">
          <div className="flex items-center gap-2">
            <Home className="h-4 w-4" />
            <h3 className="text-sm font-semibold uppercase tracking-wider">Get Home Channel</h3>
          </div>
          <Input
            type="text"
            value={homeChannelWallet}
            onChange={(e) => setHomeChannelWallet(e.target.value)}
            placeholder="Wallet address"
            className="font-mono text-xs"
          />
          <Input
            type="text"
            value={homeChannelAsset}
            onChange={(e) => setHomeChannelAsset(e.target.value)}
            placeholder="Asset"
          />
          <Button
            onClick={handleGetHomeChannel}
            disabled={loading || !homeChannelWallet || !homeChannelAsset}
            className="w-full"
          >
            {loading ? 'Loading...' : 'Get Channel'}
          </Button>
        </div>

        {/* Get Escrow Channel */}
        <div className="border border-border p-4 space-y-3">
          <div className="flex items-center gap-2">
            <Lock className="h-4 w-4" />
            <h3 className="text-sm font-semibold uppercase tracking-wider">Get Escrow Channel</h3>
          </div>
          <Input
            type="text"
            value={escrowChannelId}
            onChange={(e) => setEscrowChannelId(e.target.value)}
            placeholder="Escrow Channel ID"
            className="font-mono text-xs"
          />
          <Button
            onClick={handleGetEscrowChannel}
            disabled={loading || !escrowChannelId}
            className="w-full"
          >
            {loading ? 'Loading...' : 'Get Channel'}
          </Button>
        </div>
      </div>

      {/* Results Display */}
      {result && (
        <div className="mt-4 bg-muted p-4 border border-border max-h-96 overflow-y-auto">
          <h3 className="text-sm font-semibold uppercase tracking-wider mb-3">Results:</h3>

          {activeView === 'state' && (
            <div className="space-y-2 text-sm">
              <div className="flex gap-2">
                <span className="font-semibold uppercase tracking-wider text-xs text-muted-foreground min-w-[120px]">Version:</span>
                <span className="font-mono text-xs">{result.version?.toString()}</span>
              </div>
              <div className="flex gap-2">
                <span className="font-semibold uppercase tracking-wider text-xs text-muted-foreground min-w-[120px]">Epoch:</span>
                <span className="font-mono text-xs">{result.epoch?.toString()}</span>
              </div>
              <div className="flex gap-2">
                <span className="font-semibold uppercase tracking-wider text-xs text-muted-foreground min-w-[120px]">State ID:</span>
                <span className="font-mono text-xs break-all">{result.id}</span>
              </div>
              {result.homeChannelId && (
                <div className="flex gap-2">
                  <span className="font-semibold uppercase tracking-wider text-xs text-muted-foreground min-w-[120px]">Channel:</span>
                  <span className="font-mono text-xs break-all">{result.homeChannelId}</span>
                </div>
              )}
              <div className="mt-4 pt-3 border-t border-border">
                <div className="font-semibold uppercase tracking-wider text-xs mb-2">Home Ledger:</div>
                <div className="ml-4 space-y-1.5 text-xs">
                  <div className="flex gap-2">
                    <span className="font-semibold uppercase tracking-wider text-muted-foreground min-w-[140px]">Chain:</span>
                    <span className="font-mono">{(result.homeLedger?.blockchainId || result.homeLedger?.blockchain_id)?.toString()}</span>
                  </div>
                  <div className="flex gap-2">
                    <span className="font-semibold uppercase tracking-wider text-muted-foreground min-w-[140px]">Token:</span>
                    <span className="font-mono break-all">{result.homeLedger?.tokenAddress || result.homeLedger?.token_address}</span>
                  </div>
                  <div className="flex gap-2">
                    <span className="font-semibold uppercase tracking-wider text-muted-foreground min-w-[140px]">User Balance:</span>
                    <span className="font-mono">{result.homeLedger?.userBalance?.toString() || result.homeLedger?.user_balance?.toString()}</span>
                  </div>
                  <div className="flex gap-2">
                    <span className="font-semibold uppercase tracking-wider text-muted-foreground min-w-[140px]">Node Balance:</span>
                    <span className="font-mono">{result.homeLedger?.nodeBalance?.toString() || result.homeLedger?.node_balance?.toString()}</span>
                  </div>
                  <div className="flex gap-2">
                    <span className="font-semibold uppercase tracking-wider text-muted-foreground min-w-[140px]">User NetFlow:</span>
                    <span className="font-mono">{result.homeLedger?.userNetFlow?.toString() || result.homeLedger?.user_net_flow?.toString()}</span>
                  </div>
                  <div className="flex gap-2">
                    <span className="font-semibold uppercase tracking-wider text-muted-foreground min-w-[140px]">Node NetFlow:</span>
                    <span className="font-mono">{result.homeLedger?.nodeNetFlow?.toString() || result.homeLedger?.node_net_flow?.toString()}</span>
                  </div>
                </div>
              </div>
              {result.transitions && result.transitions.length > 0 && (
                <div className="mt-4 pt-3 border-t border-border">
                  <div className="font-semibold uppercase tracking-wider text-xs mb-2">
                    Transitions: {result.transitions.length}
                  </div>
                  <div className="ml-4 space-y-1">
                    {result.transitions.map((t: any, idx: number) => (
                      <div key={idx} className="text-xs font-mono">
                        {idx + 1}. {t.type} (Amount: {t.amount?.toString()})
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}

          {(activeView === 'homeChannel' || activeView === 'escrowChannel') && (
            <div className="space-y-2 text-sm">
              <div className="flex gap-2">
                <span className="font-semibold uppercase tracking-wider text-xs text-muted-foreground min-w-[160px]">Channel ID:</span>
                <span className="font-mono text-xs break-all">{result.channelId || result.channel_id}</span>
              </div>
              <div className="flex gap-2">
                <span className="font-semibold uppercase tracking-wider text-xs text-muted-foreground min-w-[160px]">User Wallet:</span>
                <span className="font-mono text-xs break-all">{result.userWallet || result.user_wallet}</span>
              </div>
              <div className="flex gap-2">
                <span className="font-semibold uppercase tracking-wider text-xs text-muted-foreground min-w-[160px]">Type:</span>
                <span className="font-mono text-xs">{result.type}</span>
              </div>
              <div className="flex gap-2">
                <span className="font-semibold uppercase tracking-wider text-xs text-muted-foreground min-w-[160px]">Status:</span>
                <span className="font-mono text-xs">{result.status}</span>
              </div>
              <div className="flex gap-2">
                <span className="font-semibold uppercase tracking-wider text-xs text-muted-foreground min-w-[160px]">Version:</span>
                <span className="font-mono text-xs">{(result.stateVersion || result.state_version)?.toString()}</span>
              </div>
              <div className="flex gap-2">
                <span className="font-semibold uppercase tracking-wider text-xs text-muted-foreground min-w-[160px]">Nonce:</span>
                <span className="font-mono text-xs">{result.nonce?.toString()}</span>
              </div>
              <div className="flex gap-2">
                <span className="font-semibold uppercase tracking-wider text-xs text-muted-foreground min-w-[160px]">Chain ID:</span>
                <span className="font-mono text-xs">{(result.blockchainId || result.blockchain_id)?.toString()}</span>
              </div>
              <div className="flex gap-2">
                <span className="font-semibold uppercase tracking-wider text-xs text-muted-foreground min-w-[160px]">Token:</span>
                <span className="font-mono text-xs break-all">{result.tokenAddress || result.token_address}</span>
              </div>
              <div className="flex gap-2">
                <span className="font-semibold uppercase tracking-wider text-xs text-muted-foreground min-w-[160px]">Challenge Duration:</span>
                <span className="font-mono text-xs">{(result.challengeDuration || result.challenge_duration)?.toString()}s</span>
              </div>
            </div>
          )}
        </div>
      )}
    </CollapsibleCard>
  );
}
