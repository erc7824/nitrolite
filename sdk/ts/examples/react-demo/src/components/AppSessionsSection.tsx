import { useState } from 'react';
import type { Client } from '@erc7824/nitrolite';
import type { StatusMessage } from '../types';

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
    <div className="bg-white rounded-lg shadow-md p-6">
      <h2 className="text-2xl font-bold mb-4 text-gray-800">App Sessions</h2>
      <p className="text-sm text-gray-600 mb-6">Query multi-party application sessions</p>

      <div className="border border-gray-200 rounded p-4 mb-6">
        <h3 className="font-semibold mb-3 text-gray-700">Get App Sessions</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
          <input
            type="text"
            value={wallet}
            onChange={(e) => setWallet(e.target.value)}
            placeholder="Participant wallet address"
            className="px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">All Statuses</option>
            <option value="open">Open</option>
            <option value="closed">Closed</option>
          </select>
          <button
            onClick={handleGetAppSessions}
            disabled={loading || !wallet}
            className="bg-blue-500 hover:bg-blue-600 text-white py-2 rounded font-medium transition disabled:bg-gray-400"
          >
            {loading ? 'Loading...' : 'Get Sessions'}
          </button>
        </div>
      </div>

      {/* Results Display */}
      {result && (
        <div className="mt-4 bg-gray-50 rounded p-4 border border-gray-200">
          <h3 className="font-semibold mb-3 text-gray-700">
            Results: {result.sessions.length} sessions (Total: {result.metadata.totalCount})
          </h3>

          {result.sessions.length === 0 ? (
            <p className="text-gray-600">No app sessions found</p>
          ) : (
            <div className="space-y-4">
              {result.sessions.map((session: any, idx: number) => (
                <div key={idx} className="border-l-4 border-purple-500 pl-4 py-3 bg-white rounded">
                  <div className="font-medium text-lg mb-2">Session {session.appSessionId}</div>
                  <div className="grid grid-cols-2 gap-2 text-sm">
                    <div><span className="font-medium">Version:</span> {session.version?.toString()}</div>
                    <div><span className="font-medium">Nonce:</span> {session.nonce?.toString()}</div>
                    <div><span className="font-medium">Quorum:</span> {session.quorum}</div>
                    <div><span className="font-medium">Closed:</span> {session.isClosed ? 'Yes' : 'No'}</div>
                    <div><span className="font-medium">Participants:</span> {session.participants?.length || 0}</div>
                    <div><span className="font-medium">Allocations:</span> {session.allocations?.length || 0}</div>
                  </div>

                  {session.participants && session.participants.length > 0 && (
                    <div className="mt-2">
                      <div className="text-sm font-medium mb-1">Participants:</div>
                      {session.participants.map((p: any, pidx: number) => (
                        <div key={pidx} className="text-xs ml-3">
                          • {p.walletAddress} (Weight: {p.signatureWeight})
                        </div>
                      ))}
                    </div>
                  )}

                  {session.allocations && session.allocations.length > 0 && (
                    <div className="mt-2">
                      <div className="text-sm font-medium mb-1">Allocations:</div>
                      {session.allocations.map((a: any, aidx: number) => (
                        <div key={aidx} className="text-xs ml-3">
                          • {a.participant}: {a.amount?.toString()} {a.asset}
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
