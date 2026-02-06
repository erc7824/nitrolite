import { useState } from 'react';
import Decimal from 'decimal.js';
import type { Client } from '@erc7824/nitrolite';
import type { StatusMessage } from '../types';

interface HighLevelOpsSectionProps {
  client: Client;
  showStatus: (type: StatusMessage['type'], message: string, details?: string) => void;
}

export default function HighLevelOpsSection({ client, showStatus }: HighLevelOpsSectionProps) {
  const [depositChainId, setDepositChainId] = useState('11155111');
  const [depositAsset, setDepositAsset] = useState('usdc');
  const [depositAmount, setDepositAmount] = useState('');
  const [loading, setLoading] = useState<string | null>(null);

  const [withdrawChainId, setWithdrawChainId] = useState('11155111');
  const [withdrawAsset, setWithdrawAsset] = useState('usdc');
  const [withdrawAmount, setWithdrawAmount] = useState('');

  const [transferRecipient, setTransferRecipient] = useState('');
  const [transferAsset, setTransferAsset] = useState('usdc');
  const [transferAmount, setTransferAmount] = useState('');

  const [closeAsset, setCloseAsset] = useState('usdc');

  const handleDeposit = async () => {
    try {
      setLoading('deposit');
      const amount = new Decimal(depositAmount);
      const txHash = await client.deposit(BigInt(depositChainId), depositAsset, amount);

      console.log('Deposit successful. Transaction:', txHash);
      showStatus('success', 'Deposit completed', `Transaction: ${txHash}`);
      setDepositAmount('');
    } catch (error) {
      console.error('Deposit failed:', error);
      showStatus('error', 'Deposit failed', error instanceof Error ? error.message : String(error));
    } finally {
      setLoading(null);
    }
  };

  const handleWithdraw = async () => {
    try {
      setLoading('withdraw');
      const amount = new Decimal(withdrawAmount);
      const txHash = await client.withdraw(BigInt(withdrawChainId), withdrawAsset, amount);

      console.log('Withdrawal successful. Transaction:', txHash);
      showStatus('success', 'Withdrawal completed', `Transaction: ${txHash}`);
      setWithdrawAmount('');
    } catch (error) {
      console.error('Withdrawal failed:', error);
      showStatus('error', 'Withdrawal failed', error instanceof Error ? error.message : String(error));
    } finally {
      setLoading(null);
    }
  };

  const handleTransfer = async () => {
    try {
      setLoading('transfer');
      const amount = new Decimal(transferAmount);
      const txId = await client.transfer(transferRecipient as `0x${string}`, transferAsset, amount);

      console.log('Transfer successful. Transaction ID:', txId);
      showStatus('success', 'Transfer completed', `Transaction ID: ${txId}`);
      setTransferAmount('');
    } catch (error) {
      console.error('Transfer failed:', error);
      showStatus('error', 'Transfer failed', error instanceof Error ? error.message : String(error));
    } finally {
      setLoading(null);
    }
  };

  const handleCloseChannel = async () => {
    try {
      setLoading('close');
      const txHash = await client.closeHomeChannel(closeAsset);

      console.log('Channel closed successfully. Transaction:', txHash);
      showStatus('success', 'Channel closed', `Transaction: ${txHash}`);
    } catch (error) {
      console.error('Channel close failed:', error);
      showStatus('error', 'Close channel failed', error instanceof Error ? error.message : String(error));
    } finally {
      setLoading(null);
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      <h2 className="text-2xl font-bold mb-4 text-gray-800">High-Level Operations</h2>
      <p className="text-sm text-gray-600 mb-6">Smart client operations with automatic state management</p>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Deposit */}
        <div className="border border-gray-200 rounded p-4">
          <h3 className="font-semibold mb-3 text-green-700">Deposit</h3>
          <div className="space-y-3">
            <input
              type="text"
              value={depositChainId}
              onChange={(e) => setDepositChainId(e.target.value)}
              placeholder="Chain ID (e.g., 80002)"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-green-500"
            />
            <input
              type="text"
              value={depositAsset}
              onChange={(e) => setDepositAsset(e.target.value)}
              placeholder="Asset (e.g., usdc)"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-green-500"
            />
            <input
              type="text"
              value={depositAmount}
              onChange={(e) => setDepositAmount(e.target.value)}
              placeholder="Amount"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-green-500"
            />
            <button
              onClick={handleDeposit}
              disabled={loading === 'deposit' || !depositChainId || !depositAsset || !depositAmount}
              className="w-full bg-green-500 hover:bg-green-600 text-white py-2 rounded font-medium transition disabled:bg-gray-400"
            >
              {loading === 'deposit' ? 'Processing...' : 'Deposit'}
            </button>
          </div>
        </div>

        {/* Withdraw */}
        <div className="border border-gray-200 rounded p-4">
          <h3 className="font-semibold mb-3 text-blue-700">Withdraw</h3>
          <div className="space-y-3">
            <input
              type="text"
              value={withdrawChainId}
              onChange={(e) => setWithdrawChainId(e.target.value)}
              placeholder="Chain ID (e.g., 80002)"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <input
              type="text"
              value={withdrawAsset}
              onChange={(e) => setWithdrawAsset(e.target.value)}
              placeholder="Asset (e.g., usdc)"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <input
              type="text"
              value={withdrawAmount}
              onChange={(e) => setWithdrawAmount(e.target.value)}
              placeholder="Amount"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <button
              onClick={handleWithdraw}
              disabled={loading === 'withdraw' || !withdrawChainId || !withdrawAsset || !withdrawAmount}
              className="w-full bg-blue-500 hover:bg-blue-600 text-white py-2 rounded font-medium transition disabled:bg-gray-400"
            >
              {loading === 'withdraw' ? 'Processing...' : 'Withdraw'}
            </button>
          </div>
        </div>

        {/* Transfer */}
        <div className="border border-gray-200 rounded p-4">
          <h3 className="font-semibold mb-3 text-purple-700">Transfer</h3>
          <div className="space-y-3">
            <input
              type="text"
              value={transferRecipient}
              onChange={(e) => setTransferRecipient(e.target.value)}
              placeholder="Recipient address (0x...)"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-purple-500"
            />
            <input
              type="text"
              value={transferAsset}
              onChange={(e) => setTransferAsset(e.target.value)}
              placeholder="Asset (e.g., usdc)"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-purple-500"
            />
            <input
              type="text"
              value={transferAmount}
              onChange={(e) => setTransferAmount(e.target.value)}
              placeholder="Amount"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-purple-500"
            />
            <button
              onClick={handleTransfer}
              disabled={loading === 'transfer' || !transferRecipient || !transferAsset || !transferAmount}
              className="w-full bg-purple-500 hover:bg-purple-600 text-white py-2 rounded font-medium transition disabled:bg-gray-400"
            >
              {loading === 'transfer' ? 'Processing...' : 'Transfer'}
            </button>
          </div>
        </div>

        {/* Close Channel */}
        <div className="border border-gray-200 rounded p-4">
          <h3 className="font-semibold mb-3 text-red-700">Close Channel</h3>
          <div className="space-y-3">
            <input
              type="text"
              value={closeAsset}
              onChange={(e) => setCloseAsset(e.target.value)}
              placeholder="Asset (e.g., usdc)"
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-red-500"
            />
            <button
              onClick={handleCloseChannel}
              disabled={loading === 'close' || !closeAsset}
              className="w-full bg-red-500 hover:bg-red-600 text-white py-2 rounded font-medium transition disabled:bg-gray-400"
            >
              {loading === 'close' ? 'Processing...' : 'Close Channel'}
            </button>
            <p className="text-xs text-gray-500">Warning: This will finalize and close the channel</p>
          </div>
        </div>
      </div>
    </div>
  );
}
