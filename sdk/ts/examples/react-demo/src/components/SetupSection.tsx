import { useState } from 'react';
import { Wallet, Link2, Server, Globe, X } from 'lucide-react';
import type { AppState, StatusMessage } from '../types';
import { formatAddress } from '../utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Badge } from './ui/badge';
import { Separator } from './ui/separator';

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
  const [homeAsset, setHomeAsset] = useState('usdc');
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
    <Card>
      <CardHeader>
        <CardTitle>Setup & Configuration</CardTitle>
        <CardDescription>Connect your wallet and configure network settings</CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Wallet Connection */}
        <div className="space-y-3">
          <div className="flex items-center gap-2 text-sm font-semibold uppercase tracking-wider">
            <Wallet className="h-4 w-4" />
            <span>1. Wallet Connection</span>
          </div>
          <Separator />
          <div className="flex items-center gap-3">
            {!appState.address ? (
              <Button onClick={onConnectWallet} className="gap-2">
                <Wallet className="h-4 w-4" />
                Connect MetaMask
              </Button>
            ) : (
              <>
                <Badge variant="outline" className="px-3 py-1.5 gap-2">
                  <div className="h-2 w-2 rounded-full bg-accent animate-pulse" />
                  <code className="text-xs font-mono">{formatAddress(appState.address)}</code>
                </Badge>
                <Button
                  onClick={onDisconnectWallet}
                  variant="ghost"
                  size="sm"
                  className="gap-2"
                >
                  <X className="h-3 w-3" />
                  Disconnect
                </Button>
              </>
            )}
          </div>
        </div>

        {/* Node URL Configuration */}
        <div className="space-y-3">
          <div className="flex items-center gap-2 text-sm font-semibold uppercase tracking-wider">
            <Link2 className="h-4 w-4" />
            <span>2. Clearnode URL</span>
          </div>
          <Separator />
          <Input
            type="text"
            value={appState.nodeUrl}
            onChange={(e) => onNodeUrlChange(e.target.value)}
            disabled={appState.connected}
            placeholder="wss://clearnode.example.com/ws"
          />
        </div>

        {/* RPC Configuration */}
        <div className="space-y-3">
          <div className="flex items-center gap-2 text-sm font-semibold uppercase tracking-wider">
            <Server className="h-4 w-4" />
            <span>3. Blockchain RPCs</span>
          </div>
          <Separator />
          <p className="text-xs text-muted-foreground">
            Pre-configured for Polygon Amoy, Ethereum Sepolia, Base Sepolia, and Arbitrum Sepolia
          </p>
          <div className="flex gap-2">
            <Input
              type="text"
              value={chainId}
              onChange={(e) => setChainId(e.target.value)}
              placeholder="Chain ID"
              className="w-32"
            />
            <Input
              type="text"
              value={rpcUrl}
              onChange={(e) => setRpcUrl(e.target.value)}
              placeholder="RPC URL"
              className="flex-1"
            />
            <Button onClick={handleAddRpc} variant="secondary" size="sm">
              Add
            </Button>
          </div>

          {Object.entries(appState.rpcConfigs).length > 0 && (
            <div className="space-y-2">
              {Object.entries(appState.rpcConfigs).map(([id, url]) => (
                <div key={id} className="flex items-center gap-2 bg-muted p-2 text-xs">
                  <span className="font-semibold uppercase">Chain {id}</span>
                  <code className="flex-1 text-muted-foreground truncate">{url}</code>
                  <Button
                    onClick={() => onRemoveRpc(id)}
                    variant="ghost"
                    size="sm"
                    className="h-6 px-2"
                  >
                    <X className="h-3 w-3" />
                  </Button>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Home Blockchain Configuration */}
        <div className="space-y-3">
          <div className="flex items-center gap-2 text-sm font-semibold uppercase tracking-wider">
            <Globe className="h-4 w-4" />
            <span>Home Blockchain (Optional)</span>
          </div>
          <Separator />
          <p className="text-xs text-muted-foreground">
            Configure which blockchain to use as the "home" blockchain for each asset
          </p>
          <div className="flex gap-2">
            <Input
              type="text"
              value={homeAsset}
              onChange={(e) => setHomeAsset(e.target.value)}
              placeholder="Asset"
              className="w-32"
            />
            <Input
              type="text"
              value={homeChainId}
              onChange={(e) => setHomeChainId(e.target.value)}
              placeholder="Chain ID"
              className="flex-1"
            />
            <Button
              onClick={handleAddHomeBlockchain}
              disabled={appState.connected}
              variant="secondary"
              size="sm"
            >
              Set
            </Button>
          </div>

          {Object.entries(appState.homeBlockchains).length > 0 && (
            <div className="space-y-2">
              {Object.entries(appState.homeBlockchains).map(([asset, chainId]) => (
                <div key={asset} className="flex items-center gap-2 bg-muted p-2 text-xs">
                  <span className="font-semibold uppercase">{asset}</span>
                  <code className="flex-1 text-muted-foreground">Chain {chainId}</code>
                  <Button
                    onClick={() => onRemoveHomeBlockchain(asset)}
                    disabled={appState.connected}
                    variant="ghost"
                    size="sm"
                    className="h-6 px-2"
                  >
                    <X className="h-3 w-3" />
                  </Button>
                </div>
              ))}
            </div>
          )}
        </div>

        <Separator />

        {/* Connect/Disconnect */}
        <div className="flex items-center gap-3">
          {!appState.connected ? (
            <Button
              onClick={onConnectNode}
              disabled={!appState.address}
              className="gap-2"
              size="lg"
            >
              <Server className="h-4 w-4" />
              4. Connect to Node
            </Button>
          ) : (
            <>
              <Badge className="px-4 py-2 gap-2">
                <div className="h-2 w-2 rounded-full bg-secondary-foreground animate-pulse" />
                <span>Connected to Node</span>
              </Badge>
              <Button onClick={onDisconnect} variant="destructive" size="sm">
                Disconnect
              </Button>
            </>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
