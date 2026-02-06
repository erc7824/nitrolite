import { useState } from 'react';
import Decimal from 'decimal.js';
import { ArrowDownToLine, ArrowUpFromLine, Send, XCircle } from 'lucide-react';
import type { Client } from '@erc7824/nitrolite';
import type { StatusMessage } from '../types';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { Button } from './ui/button';
import { Input } from './ui/input';

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
      showStatus('success', 'Deposit completed', `Transaction: ${txHash}`);
      setDepositAmount('');
    } catch (error) {
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
      showStatus('success', 'Withdrawal completed', `Transaction: ${txHash}`);
      setWithdrawAmount('');
    } catch (error) {
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
      showStatus('success', 'Transfer completed', `Transaction ID: ${txId}`);
      setTransferAmount('');
    } catch (error) {
      showStatus('error', 'Transfer failed', error instanceof Error ? error.message : String(error));
    } finally {
      setLoading(null);
    }
  };

  const handleCloseChannel = async () => {
    try {
      setLoading('close');
      const txHash = await client.closeHomeChannel(closeAsset);
      showStatus('success', 'Channel closed', `Transaction: ${txHash}`);
    } catch (error) {
      showStatus('error', 'Close channel failed', error instanceof Error ? error.message : String(error));
    } finally {
      setLoading(null);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>High-Level Operations</CardTitle>
        <CardDescription>Smart client operations with automatic state management</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* Deposit */}
          <div className="border border-border p-4 space-y-3">
            <div className="flex items-center gap-2">
              <ArrowDownToLine className="h-4 w-4" />
              <h3 className="text-sm font-semibold uppercase tracking-wider">Deposit</h3>
            </div>
            <Input
              type="text"
              value={depositChainId}
              onChange={(e) => setDepositChainId(e.target.value)}
              placeholder="Chain ID"
            />
            <Input
              type="text"
              value={depositAsset}
              onChange={(e) => setDepositAsset(e.target.value)}
              placeholder="Asset"
            />
            <Input
              type="text"
              value={depositAmount}
              onChange={(e) => setDepositAmount(e.target.value)}
              placeholder="Amount"
            />
            <Button
              onClick={handleDeposit}
              disabled={loading === 'deposit' || !depositChainId || !depositAsset || !depositAmount}
              className="w-full"
            >
              {loading === 'deposit' ? 'Processing...' : 'Deposit'}
            </Button>
          </div>

          {/* Withdraw */}
          <div className="border border-border p-4 space-y-3">
            <div className="flex items-center gap-2">
              <ArrowUpFromLine className="h-4 w-4" />
              <h3 className="text-sm font-semibold uppercase tracking-wider">Withdraw</h3>
            </div>
            <Input
              type="text"
              value={withdrawChainId}
              onChange={(e) => setWithdrawChainId(e.target.value)}
              placeholder="Chain ID"
            />
            <Input
              type="text"
              value={withdrawAsset}
              onChange={(e) => setWithdrawAsset(e.target.value)}
              placeholder="Asset"
            />
            <Input
              type="text"
              value={withdrawAmount}
              onChange={(e) => setWithdrawAmount(e.target.value)}
              placeholder="Amount"
            />
            <Button
              onClick={handleWithdraw}
              disabled={loading === 'withdraw' || !withdrawChainId || !withdrawAsset || !withdrawAmount}
              className="w-full"
            >
              {loading === 'withdraw' ? 'Processing...' : 'Withdraw'}
            </Button>
          </div>

          {/* Transfer */}
          <div className="border border-border p-4 space-y-3">
            <div className="flex items-center gap-2">
              <Send className="h-4 w-4" />
              <h3 className="text-sm font-semibold uppercase tracking-wider">Transfer</h3>
            </div>
            <Input
              type="text"
              value={transferRecipient}
              onChange={(e) => setTransferRecipient(e.target.value)}
              placeholder="Recipient (0x...)"
              className="font-mono text-xs"
            />
            <Input
              type="text"
              value={transferAsset}
              onChange={(e) => setTransferAsset(e.target.value)}
              placeholder="Asset"
            />
            <Input
              type="text"
              value={transferAmount}
              onChange={(e) => setTransferAmount(e.target.value)}
              placeholder="Amount"
            />
            <Button
              onClick={handleTransfer}
              disabled={loading === 'transfer' || !transferRecipient || !transferAsset || !transferAmount}
              className="w-full"
            >
              {loading === 'transfer' ? 'Processing...' : 'Transfer'}
            </Button>
          </div>

          {/* Close Channel */}
          <div className="border border-border p-4 space-y-3">
            <div className="flex items-center gap-2">
              <XCircle className="h-4 w-4" />
              <h3 className="text-sm font-semibold uppercase tracking-wider">Close Channel</h3>
            </div>
            <Input
              type="text"
              value={closeAsset}
              onChange={(e) => setCloseAsset(e.target.value)}
              placeholder="Asset"
            />
            <Button
              onClick={handleCloseChannel}
              disabled={loading === 'close' || !closeAsset}
              variant="destructive"
              className="w-full"
            >
              {loading === 'close' ? 'Processing...' : 'Close Channel'}
            </Button>
            <p className="text-xs text-muted-foreground">Warning: This will finalize and close the channel</p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
