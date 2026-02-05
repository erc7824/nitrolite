import { useState } from 'react';
import type { Client } from '@erc7824/nitrolite';
import type { StatusMessage } from '../types';

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
    <div className="bg-white rounded-lg shadow-md p-6">
      <h2 className="text-2xl font-bold mb-4 text-gray-800">User Queries</h2>
      <p className="text-sm text-gray-600 mb-6">Query user balances and transaction history</p>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
        {/* Balances */}
        <div className="border border-gray-200 rounded p-4">
          <h3 className="font-semibold mb-3 text-gray-700">Get Balances</h3>
          <div className="space-y-3">
            <input
              type="text"
              value={balancesAddress}
              onChange={(e) => setBalancesAddress(e.target.value)}
              placeholder="Wallet address"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <button
              onClick={handleGetBalances}
              disabled={loading || !balancesAddress}
              className="w-full bg-blue-500 hover:bg-blue-600 text-white py-2 rounded font-medium transition disabled:bg-gray-400"
            >
              {loading ? 'Loading...' : 'Get Balances'}
            </button>
          </div>
        </div>

        {/* Transactions */}
        <div className="border border-gray-200 rounded p-4">
          <h3 className="font-semibold mb-3 text-gray-700">Get Transactions</h3>
          <div className="space-y-3">
            <input
              type="text"
              value={txAddress}
              onChange={(e) => setTxAddress(e.target.value)}
              placeholder="Wallet address"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <button
              onClick={handleGetTransactions}
              disabled={loading || !txAddress}
              className="w-full bg-blue-500 hover:bg-blue-600 text-white py-2 rounded font-medium transition disabled:bg-gray-400"
            >
              {loading ? 'Loading...' : 'Get Transactions'}
            </button>
          </div>
        </div>
      </div>

      {/* Results Display */}
      {result && (
        <div className="mt-4 bg-gray-50 rounded p-4 border border-gray-200">
          <h3 className="font-semibold mb-3 text-gray-700">Results:</h3>

          {activeView === 'balances' && (
            <div className="space-y-2">
              {result.length === 0 ? (
                <p className="text-gray-600">No balances found</p>
              ) : (
                result.map((balance: any, idx: number) => (
                  <div key={idx} className="flex justify-between items-center border-b border-gray-200 pb-2">
                    <span className="font-medium">{balance.asset}</span>
                    <span className="text-gray-700">{balance.balance.toString()}</span>
                  </div>
                ))
              )}
            </div>
          )}

          {activeView === 'transactions' && (
            <div className="space-y-3">
              <div className="text-sm text-gray-600 mb-2">
                Showing {result.transactions.length} of {result.metadata.totalCount} transactions
              </div>
              {result.transactions.length === 0 ? (
                <p className="text-gray-600">No transactions found</p>
              ) : (
                result.transactions.map((tx: any, idx: number) => (
                  <div key={idx} className="border-l-4 border-blue-500 pl-3 py-2">
                    <div className="font-medium">{tx.txType}</div>
                    <div className="text-sm text-gray-600">ID: {tx.id}</div>
                    <div className="text-sm text-gray-600">From: {tx.fromAccount}</div>
                    <div className="text-sm text-gray-600">To: {tx.toAccount}</div>
                    <div className="text-sm text-gray-700">Amount: {tx.amount.toString()} {tx.asset}</div>
                    <div className="text-xs text-gray-500">
                      {new Date(tx.createdAt).toLocaleString()}
                    </div>
                  </div>
                ))
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
