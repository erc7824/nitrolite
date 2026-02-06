import { useState } from 'react';
import { Wallet, History, Coins, ArrowRightLeft } from 'lucide-react';
import type { Client } from '@erc7824/nitrolite';
import type { StatusMessage } from '../types';
import { CollapsibleCard } from './ui/collapsible-card';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Separator } from './ui/separator';

interface UserQueriesSectionProps {
  client: Client;
  defaultAddress: string;
  showStatus: (type: StatusMessage['type'], message: string, details?: string) => void;
}

export default function UserQueriesSection({ client, defaultAddress, showStatus }: UserQueriesSectionProps) {
  const [balancesAddress, setBalancesAddress] = useState(defaultAddress);
  const [txAddress, setTxAddress] = useState(defaultAddress);
  const [result, setResult] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [activeView, setActiveView] = useState<string | null>(null);

  const handleGetBalances = async () => {
    try {
      setLoading(true);
      const balances = await client.getBalances(balancesAddress as `0x${string}`);
      setResult(balances);
      setActiveView('balances');
      showStatus('success', `Retrieved balances for ${balancesAddress}`);
    } catch (error) {
      showStatus('error', 'Failed to get balances', error instanceof Error ? error.message : String(error));
    } finally {
      setLoading(false);
    }
  };

  const handleGetTransactions = async () => {
    try {
      setLoading(true);
      const { transactions, metadata } = await client.getTransactions(txAddress as `0x${string}`, {
        page: 1,
        pageSize: 20,
      });
      setResult({ transactions, metadata });
      setActiveView('transactions');
      showStatus('success', `Retrieved ${transactions.length} transactions`);
    } catch (error) {
      showStatus('error', 'Failed to get transactions', error instanceof Error ? error.message : String(error));
    } finally {
      setLoading(false);
    }
  };

  return (
    <CollapsibleCard
      title="User Queries"
      description="Query user balances and transaction history"
      defaultOpen={false}
    >
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
        {/* Balances */}
        <div className="border border-border p-4 space-y-3">
          <div className="flex items-center gap-2">
            <Coins className="h-4 w-4" />
            <h3 className="text-sm font-semibold uppercase tracking-wider">Get Balances</h3>
          </div>
          <Separator />
          <Input
            type="text"
            value={balancesAddress}
            onChange={(e) => setBalancesAddress(e.target.value)}
            placeholder="Wallet address (0x...)"
            className="font-mono text-xs"
          />
          <Button
            onClick={handleGetBalances}
            disabled={loading || !balancesAddress}
            className="w-full"
          >
            {loading ? 'Loading...' : 'Get Balances'}
          </Button>
        </div>

        {/* Transactions */}
        <div className="border border-border p-4 space-y-3">
          <div className="flex items-center gap-2">
            <History className="h-4 w-4" />
            <h3 className="text-sm font-semibold uppercase tracking-wider">Get Transactions</h3>
          </div>
          <Separator />
          <Input
            type="text"
            value={txAddress}
            onChange={(e) => setTxAddress(e.target.value)}
            placeholder="Wallet address (0x...)"
            className="font-mono text-xs"
          />
          <Button
            onClick={handleGetTransactions}
            disabled={loading || !txAddress}
            className="w-full"
          >
            {loading ? 'Loading...' : 'Get Transactions'}
          </Button>
        </div>
      </div>

      {/* Results Display */}
      {result && (
        <div className="mt-4 bg-muted p-4 space-y-3">
          <div className="flex items-center gap-2">
            <Wallet className="h-4 w-4" />
            <h3 className="text-sm font-semibold uppercase tracking-wider">Results</h3>
          </div>
          <Separator />

          {activeView === 'balances' && (
            <div className="space-y-2">
              {result.length === 0 ? (
                <p className="text-xs text-muted-foreground">No balances found</p>
              ) : (
                result.map((balance: any, idx: number) => (
                  <div key={idx} className="flex justify-between items-center bg-background p-3 border-l-2 border-accent">
                    <span className="text-sm font-semibold uppercase tracking-wider">{balance.asset}</span>
                    <code className="text-sm text-muted-foreground">{balance.balance.toString()}</code>
                  </div>
                ))
              )}
            </div>
          )}

          {activeView === 'transactions' && (
            <div className="space-y-3">
              <div className="text-xs text-muted-foreground mb-2">
                Showing {result.transactions.length} of {result.metadata.totalCount} transactions
              </div>
              {result.transactions.length === 0 ? (
                <p className="text-xs text-muted-foreground">No transactions found</p>
              ) : (
                result.transactions.map((tx: any, idx: number) => (
                  <div key={idx} className="bg-background p-4 border-l-2 border-accent space-y-2">
                    <div className="flex items-center gap-2">
                      <ArrowRightLeft className="h-3 w-3 text-accent" />
                      <span className="text-sm font-semibold uppercase tracking-wider">{tx.txType}</span>
                    </div>
                    <div className="grid grid-cols-1 gap-1 text-xs">
                      <div className="flex gap-2">
                        <span className="text-muted-foreground uppercase tracking-wider">ID:</span>
                        <code className="font-mono">{tx.id}</code>
                      </div>
                      <div className="flex gap-2">
                        <span className="text-muted-foreground uppercase tracking-wider">From:</span>
                        <code className="font-mono truncate">{tx.fromAccount}</code>
                      </div>
                      <div className="flex gap-2">
                        <span className="text-muted-foreground uppercase tracking-wider">To:</span>
                        <code className="font-mono truncate">{tx.toAccount}</code>
                      </div>
                      <div className="flex gap-2">
                        <span className="text-muted-foreground uppercase tracking-wider">Amount:</span>
                        <code className="font-mono font-semibold">{tx.amount.toString()} {tx.asset}</code>
                      </div>
                      <div className="flex gap-2">
                        <span className="text-muted-foreground uppercase tracking-wider">Date:</span>
                        <code className="font-mono text-muted-foreground">
                          {new Date(tx.createdAt).toLocaleString()}
                        </code>
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
          )}
        </div>
      )}
    </CollapsibleCard>
  );
}
