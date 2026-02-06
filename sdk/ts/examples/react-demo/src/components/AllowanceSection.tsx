import { useState } from 'react';
import Decimal from 'decimal.js';
import type { Client } from '@erc7824/nitrolite';
import type { StatusMessage } from '../types';

interface AllowanceSectionProps {
  client: Client;
  defaultAddress: string;
  showStatus: (type: StatusMessage['type'], message: string, details?: string) => void;
}

export default function AllowanceSection({ client, defaultAddress, showStatus }: AllowanceSectionProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [chainId, setChainId] = useState('11155111');
  const [tokenAddress, setTokenAddress] = useState('');
  const [amount, setAmount] = useState('');
  const [currentAllowance, setCurrentAllowance] = useState<string | null>(null);
  const [loading, setLoading] = useState<string | null>(null);

  // Check current allowance
  const checkAllowance = async () => {
    if (!tokenAddress || !chainId) {
      showStatus('error', 'Missing fields', 'Please enter token address and chain ID');
      return;
    }

    try {
      setLoading('check');
      const allowance = await client.checkTokenAllowance(
        BigInt(chainId),
        tokenAddress,
        defaultAddress
      );

      setCurrentAllowance(allowance.toString());
      console.log('Allowance checked:', allowance.toString());
      showStatus('success', 'Allowance checked', `Current allowance: ${allowance.toString()}`);
    } catch (error) {
      console.error('Failed to check allowance:', error);
      showStatus('error', 'Check failed', error instanceof Error ? error.message : String(error));
      setCurrentAllowance('Error');
    } finally {
      setLoading(null);
    }
  };

  const handleApprove = async () => {
    if (!tokenAddress || !amount || !chainId) {
      showStatus('error', 'Missing fields', 'Please fill in all fields');
      return;
    }

    try {
      setLoading('approve');
      const amountBig = BigInt(amount);

      const hash = await client.approveToken(
        BigInt(chainId),
        tokenAddress,
        amountBig
      );

      console.log('Approval successful. Transaction hash:', hash);
      showStatus('success', 'Allowance approved', `Transaction: ${hash}`);

      await checkAllowance();
      setAmount('');
    } catch (error) {
      console.error('Approval failed:', error);
      showStatus('error', 'Approval failed', error instanceof Error ? error.message : String(error));
    } finally {
      setLoading(null);
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h2 className="text-2xl font-bold text-gray-800">Token Allowance</h2>
          <p className="text-sm text-gray-600 mt-1">
            Approve the contract to spend your tokens. <strong>Required before deposits!</strong>
          </p>
        </div>
        <button
          onClick={() => setIsExpanded(!isExpanded)}
          className="ml-4 px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded font-medium transition"
        >
          {isExpanded ? '▲ Hide' : '▼ Show'}
        </button>
      </div>

      {isExpanded && (
        <div className="grid grid-cols-1 gap-6">
        {/* Approve Section */}
        <div className="border border-blue-200 rounded p-4 bg-blue-50">
          <h3 className="font-semibold mb-3 text-blue-700">Approve Token Spending</h3>
          <div className="space-y-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Chain ID
              </label>
              <input
                type="text"
                value={chainId}
                onChange={(e) => setChainId(e.target.value)}
                placeholder="11155111 (Sepolia)"
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Token Address
              </label>
              <input
                type="text"
                value={tokenAddress}
                onChange={(e) => setTokenAddress(e.target.value)}
                placeholder="0x... (ERC20 token contract address)"
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <p className="text-xs text-gray-500 mt-1">
                Example USDC on Sepolia: 0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Amount (in smallest unit - e.g., 1000000 = 1 USDC with 6 decimals)
              </label>
              <input
                type="text"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                placeholder="1000000 (for 6 decimal tokens) or very large number"
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <p className="text-xs text-gray-500 mt-1">
                💡 Tip: Use 115792089237316195423570985008687907853269984665640564039457584007913129639935 for unlimited
              </p>
            </div>

            {currentAllowance !== null && (
              <div className="bg-white border border-blue-300 rounded p-3">
                <p className="text-sm">
                  <span className="font-medium text-blue-900">Current Allowance:</span>{' '}
                  <span className="font-mono text-blue-700">{currentAllowance}</span>
                </p>
              </div>
            )}

            <div className="flex gap-2">
              <button
                onClick={handleApprove}
                disabled={loading === 'approve' || !tokenAddress || !amount || !chainId}
                className="flex-1 bg-blue-600 hover:bg-blue-700 text-white py-2 px-4 rounded font-medium transition disabled:bg-gray-400 disabled:cursor-not-allowed"
              >
                {loading === 'approve' ? '⏳ Approving...' : '✅ Approve'}
              </button>
              <button
                onClick={checkAllowance}
                disabled={loading === 'check' || !tokenAddress || !chainId}
                className="flex-1 bg-gray-600 hover:bg-gray-700 text-white py-2 px-4 rounded font-medium transition disabled:bg-gray-400 disabled:cursor-not-allowed"
              >
                {loading === 'check' ? '⏳ Checking...' : '🔍 Check Allowance'}
              </button>
            </div>
          </div>
        </div>

        {/* Info Section */}
        <div className="bg-yellow-50 border border-yellow-200 rounded p-4">
          <h4 className="font-semibold text-yellow-900 mb-2">ℹ️ About Token Allowances</h4>
          <ul className="text-sm text-yellow-800 space-y-1">
            <li>✓ Allowance lets the contract spend your tokens on your behalf</li>
            <li>✓ <strong>Required before making deposits</strong> to the channel</li>
            <li>✓ You can approve unlimited tokens or a specific amount</li>
            <li>✓ Check your current allowance before approving more</li>
            <li>✓ Amount must be in smallest unit (e.g., 1 USDC = 1000000 if 6 decimals)</li>
          </ul>
        </div>
      </div>
      )}
    </div>
  );
}
