import { useState } from 'react';
import { Search, Users, Database, CheckCircle2, XCircle } from 'lucide-react';
import type { Client } from '@erc7824/nitrolite';
import type { StatusMessage } from '../types';
import { CollapsibleCard } from './ui/collapsible-card';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Separator } from './ui/separator';

interface AppSessionsSectionProps {
  client: Client;
  defaultAddress: string;
  showStatus: (type: StatusMessage['type'], message: string, details?: string) => void;
}

export default function AppSessionsSection({ client, defaultAddress, showStatus }: AppSessionsSectionProps) {
  const [wallet, setWallet] = useState(defaultAddress);
  const [status, setStatus] = useState<string>('');
  const [result, setResult] = useState<any>(null);
  const [loading, setLoading] = useState(false);

  const handleGetAppSessions = async () => {
    try {
      setLoading(true);
      const { sessions, metadata } = await client.getAppSessions({
        wallet: wallet as `0x${string}`,
        status: status || undefined,
        page: 1,
        pageSize: 20,
      });
      setResult({ sessions, metadata });
      showStatus('success', `Retrieved ${sessions.length} app sessions`);
    } catch (error) {
      showStatus('error', 'Failed to get app sessions', error instanceof Error ? error.message : String(error));
    } finally {
      setLoading(false);
    }
  };

  return (
    <CollapsibleCard
      title="App Sessions"
      description="Query multi-party application sessions"
      defaultOpen={false}
    >
      <div className="space-y-6">
        {/* Query Form */}
        <div className="space-y-3">
          <div className="flex items-center gap-2 text-sm font-semibold uppercase tracking-wider">
            <Search className="h-4 w-4" />
            <span>Query Sessions</span>
          </div>
          <Separator />

          <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
            <Input
              type="text"
              value={wallet}
              onChange={(e) => setWallet(e.target.value)}
              placeholder="Participant wallet address"
              className="font-mono text-xs"
            />

            <select
              value={status}
              onChange={(e) => setStatus(e.target.value)}
              className="flex h-10 w-full border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
            >
              <option value="">All Statuses</option>
              <option value="open">Open</option>
              <option value="closed">Closed</option>
            </select>

            <Button
              onClick={handleGetAppSessions}
              disabled={loading || !wallet}
              className="gap-2"
            >
              <Search className="h-4 w-4" />
              {loading ? 'Loading...' : 'Get Sessions'}
            </Button>
          </div>
        </div>

        {/* Results Display */}
        {result && (
          <div className="space-y-3">
            <Separator />
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2 text-sm font-semibold uppercase tracking-wider">
                <Database className="h-4 w-4" />
                <span>Results</span>
              </div>
              <div className="text-xs text-muted-foreground">
                {result.sessions.length} of {result.metadata.totalCount} sessions
              </div>
            </div>

            {result.sessions.length === 0 ? (
              <div className="text-sm text-muted-foreground bg-muted p-4 text-center">
                No app sessions found
              </div>
            ) : (
              <div className="space-y-4">
                {result.sessions.map((session: any, idx: number) => (
                  <div
                    key={idx}
                    className="border border-border bg-muted/30 p-4 space-y-3"
                  >
                    {/* Session Header */}
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <div className="text-sm font-semibold uppercase tracking-wider">
                          Session {session.appSessionId}
                        </div>
                      </div>
                      {session.isClosed ? (
                        <div className="flex items-center gap-1 text-xs text-muted-foreground">
                          <XCircle className="h-3 w-3" />
                          <span>Closed</span>
                        </div>
                      ) : (
                        <div className="flex items-center gap-1 text-xs" style={{ color: '#FCD000' }}>
                          <CheckCircle2 className="h-3 w-3" />
                          <span>Open</span>
                        </div>
                      )}
                    </div>

                    {/* Session Details Grid */}
                    <div className="grid grid-cols-2 md:grid-cols-3 gap-3 text-xs">
                      <div className="space-y-1">
                        <div className="text-muted-foreground uppercase tracking-wider">Version</div>
                        <div className="font-mono">{session.version?.toString()}</div>
                      </div>
                      <div className="space-y-1">
                        <div className="text-muted-foreground uppercase tracking-wider">Nonce</div>
                        <div className="font-mono">{session.nonce?.toString()}</div>
                      </div>
                      <div className="space-y-1">
                        <div className="text-muted-foreground uppercase tracking-wider">Quorum</div>
                        <div className="font-mono">{session.quorum}</div>
                      </div>
                      <div className="space-y-1">
                        <div className="text-muted-foreground uppercase tracking-wider">Participants</div>
                        <div className="font-mono">{session.participants?.length || 0}</div>
                      </div>
                      <div className="space-y-1">
                        <div className="text-muted-foreground uppercase tracking-wider">Allocations</div>
                        <div className="font-mono">{session.allocations?.length || 0}</div>
                      </div>
                    </div>

                    {/* Participants Details */}
                    {session.participants && session.participants.length > 0 && (
                      <div className="space-y-2 pt-2 border-t border-border">
                        <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider">
                          <Users className="h-3 w-3" />
                          <span>Participants</span>
                        </div>
                        <div className="space-y-1 pl-5">
                          {session.participants.map((p: any, pidx: number) => (
                            <div key={pidx} className="text-xs font-mono text-muted-foreground flex items-center gap-2">
                              <span className="text-[#FCD000]">•</span>
                              <span className="flex-1 truncate">{p.walletAddress}</span>
                              <span className="text-[10px] uppercase">Weight: {p.signatureWeight}</span>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {/* Allocations Details */}
                    {session.allocations && session.allocations.length > 0 && (
                      <div className="space-y-2 pt-2 border-t border-border">
                        <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider">
                          <Database className="h-3 w-3" />
                          <span>Allocations</span>
                        </div>
                        <div className="space-y-1 pl-5">
                          {session.allocations.map((a: any, aidx: number) => (
                            <div key={aidx} className="text-xs font-mono text-muted-foreground flex items-center gap-2">
                              <span className="text-[#FCD000]">•</span>
                              <span className="flex-1 truncate">{a.participant}</span>
                              <span className="uppercase">{a.amount?.toString()} {a.asset}</span>
                            </div>
                          ))}
                        </div>
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
