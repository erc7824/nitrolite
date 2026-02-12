import { useState } from 'react';
import { Activity, Info, Link, Coins } from 'lucide-react';
import type { Client } from '@erc7824/nitrolite';
import type { StatusMessage } from '../types';
import { safeStringify } from '../utils';
import { CollapsibleCard } from './ui/collapsible-card';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Separator } from './ui/separator';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from './ui/accordion';

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
    <CollapsibleCard
      title="Node Information"
      description="Query node configuration and capabilities"
      defaultOpen={true}
    >
      <div className="space-y-6">
        {/* Action Buttons */}
        <div className="space-y-3">
          <div className="text-sm font-semibold uppercase tracking-wider">
            Operations
          </div>
          <Separator />
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            <Button
              onClick={handlePing}
              disabled={loading}
              variant="outline"
              className="gap-2"
            >
              <Activity className="h-4 w-4" />
              Ping
            </Button>
            <Button
              onClick={handleNodeInfo}
              disabled={loading}
              variant="outline"
              className="gap-2"
            >
              <Info className="h-4 w-4" />
              Node Info
            </Button>
            <Button
              onClick={handleListChains}
              disabled={loading}
              variant="outline"
              className="gap-2"
            >
              <Link className="h-4 w-4" />
              List Chains
            </Button>
            <Button
              onClick={handleListAssets}
              disabled={loading}
              variant="outline"
              className="gap-2"
            >
              <Coins className="h-4 w-4" />
              List Assets
            </Button>
          </div>
        </div>

        {/* Assets Filter */}
        {activeView === 'assets' && (
          <div className="space-y-3">
            <div className="text-sm font-semibold uppercase tracking-wider">
              Filter Options
            </div>
            <Separator />
            <Input
              type="text"
              value={chainIdFilter}
              onChange={(e) => setChainIdFilter(e.target.value)}
              placeholder="Filter by Chain ID (optional)"
            />
          </div>
        )}

        {/* Results Display */}
        {result && (
          <div className="space-y-3">
            <div className="text-sm font-semibold uppercase tracking-wider">
              Results
            </div>
            <Separator />

            <Accordion type="single" collapsible className="w-full">
              <AccordionItem value="raw-json">
                <AccordionTrigger className="text-xs uppercase tracking-wider">
                  View Raw JSON
                </AccordionTrigger>
                <AccordionContent>
                  <pre className="p-3 bg-muted text-xs overflow-auto max-h-40 font-mono border">
                    {safeStringify(result, 2)}
                  </pre>
                </AccordionContent>
              </AccordionItem>
            </Accordion>

            {activeView === 'nodeInfo' && (
              <div className="space-y-4 mt-4">
                <div className="space-y-2 bg-muted p-4">
                  <div className="flex items-start gap-2">
                    <span className="text-xs font-semibold uppercase tracking-wider min-w-20">Address:</span>
                    <code className="text-xs font-mono flex-1">{result.nodeAddress || result.node_address}</code>
                  </div>
                  <div className="flex items-start gap-2">
                    <span className="text-xs font-semibold uppercase tracking-wider min-w-20">Version:</span>
                    <code className="text-xs font-mono flex-1">{result.nodeVersion || result.node_version}</code>
                  </div>
                  <div className="flex items-start gap-2">
                    <span className="text-xs font-semibold uppercase tracking-wider min-w-20">Chains:</span>
                    <span className="text-xs">{result.blockchains?.length || 0}</span>
                  </div>
                </div>

                {result.blockchains && result.blockchains.length > 0 && (
                  <div className="space-y-2">
                    <div className="text-xs font-semibold uppercase tracking-wider">
                      Supported Blockchains
                    </div>
                    <div className="space-y-2">
                      {result.blockchains.map((bc: any, idx: number) => (
                        <div key={bc.id || bc.blockchain_id || idx} className="border-l-2 border-accent pl-3 py-1 bg-muted">
                          <div className="text-sm font-medium">{bc.name}</div>
                          <div className="text-xs text-muted-foreground">
                            ID: {(bc.id || bc.blockchain_id)?.toString()}
                          </div>
                          <div className="text-xs text-muted-foreground font-mono truncate">
                            Contract: {bc.channelHubAddress || bc.channel_hub_address}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}

            {activeView === 'chains' && (
              <div className="space-y-2 mt-4">
                {result.map((chain: any, idx: number) => (
                  <div key={chain.id || chain.blockchain_id || idx} className="border-l-2 border-accent pl-3 py-2 bg-muted">
                    <div className="text-sm font-medium">{chain.name}</div>
                    <div className="text-xs text-muted-foreground">
                      Chain ID: {(chain.id || chain.blockchain_id)?.toString()}
                    </div>
                    <div className="text-xs text-muted-foreground font-mono truncate">
                      Contract: {chain.channelHubAddress || chain.channel_hub_address}
                    </div>
                  </div>
                ))}
              </div>
            )}

            {activeView === 'assets' && (
              <div className="space-y-2 mt-4">
                {result.map((asset: any, idx: number) => (
                  <div key={asset.symbol || idx} className="border-l-2 border-accent pl-3 py-2 bg-muted">
                    <div className="text-sm font-medium">
                      {asset.name} ({asset.symbol})
                    </div>
                    <div className="text-xs text-muted-foreground">
                      Decimals: {asset.decimals}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      Tokens: {asset.tokens?.length || 0} connected
                    </div>
                    {asset.tokens && asset.tokens.length > 0 && (
                      <div className="mt-2 ml-3 space-y-1">
                        {asset.tokens.map((token: any, tidx: number) => (
                          <div key={tidx} className="text-xs bg-background p-2 font-mono">
                            <div className="text-muted-foreground">
                              Chain {(token.blockchainId || token.blockchain_id)?.toString()}
                            </div>
                            <div className="truncate">{token.address}</div>
                            <div className="text-muted-foreground">
                              Decimals: {token.decimals}
                            </div>
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
    </CollapsibleCard>
  );
}
