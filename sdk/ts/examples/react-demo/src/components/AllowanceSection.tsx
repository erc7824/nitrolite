import { useState } from 'react';
import { ChevronDown, ChevronUp, CheckCircle2, Search, Info } from 'lucide-react';
import type { Client } from '@erc7824/nitrolite';
import type { StatusMessage } from '../types';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Separator } from './ui/separator';

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
      showStatus('success', 'Allowance checked', `Current allowance: ${allowance.toString()}`);
    } catch (error) {
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
      const hash = await client.approveToken(BigInt(chainId), tokenAddress, amountBig);

      showStatus('success', 'Allowance approved', `Transaction: ${hash}`);
      await checkAllowance();
      setAmount('');
    } catch (error) {
      showStatus('error', 'Approval failed', error instanceof Error ? error.message : String(error));
    } finally {
      setLoading(null);
    }
  };

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <CardTitle>Token Allowance</CardTitle>
            <CardDescription>
              Approve the contract to spend your tokens. <span className="text-accent font-semibold">Required before deposits</span>
            </CardDescription>
          </div>
          <Button
            onClick={() => setIsExpanded(!isExpanded)}
            variant="outline"
            size="sm"
            className="gap-2"
          >
            {isExpanded ? (
              <>
                <ChevronUp className="h-4 w-4" />
                Hide
              </>
            ) : (
              <>
                <ChevronDown className="h-4 w-4" />
                Show
              </>
            )}
          </Button>
        </div>
      </CardHeader>

      {isExpanded && (
        <CardContent className="space-y-6">
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-xs uppercase tracking-wider font-semibold">Chain ID</label>
              <Input
                type="text"
                value={chainId}
                onChange={(e) => setChainId(e.target.value)}
                placeholder="11155111 (Sepolia)"
              />
            </div>

            <div className="space-y-2">
              <label className="text-xs uppercase tracking-wider font-semibold">Token Address</label>
              <Input
                type="text"
                value={tokenAddress}
                onChange={(e) => setTokenAddress(e.target.value)}
                placeholder="0x..."
                className="font-mono text-xs"
              />
              <p className="text-xs text-muted-foreground">
                Example USDC on Sepolia: 0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238
              </p>
            </div>

            <div className="space-y-2">
              <label className="text-xs uppercase tracking-wider font-semibold">Amount</label>
              <Input
                type="text"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                placeholder="1000000"
                className="font-mono"
              />
              <p className="text-xs text-muted-foreground">
                Amount in smallest unit (e.g., 1000000 = 1 USDC with 6 decimals)
              </p>
            </div>

            {currentAllowance !== null && (
              <div className="bg-muted p-3 space-y-1">
                <p className="text-xs uppercase tracking-wider font-semibold">Current Allowance</p>
                <code className="text-xs font-mono break-all">{currentAllowance}</code>
              </div>
            )}

            <div className="grid grid-cols-2 gap-2">
              <Button
                onClick={handleApprove}
                disabled={loading === 'approve' || !tokenAddress || !amount || !chainId}
                className="gap-2"
              >
                <CheckCircle2 className="h-4 w-4" />
                {loading === 'approve' ? 'Approving...' : 'Approve'}
              </Button>
              <Button
                onClick={checkAllowance}
                disabled={loading === 'check' || !tokenAddress || !chainId}
                variant="secondary"
                className="gap-2"
              >
                <Search className="h-4 w-4" />
                {loading === 'check' ? 'Checking...' : 'Check'}
              </Button>
            </div>
          </div>

          <Separator />

          <div className="bg-accent/10 border border-accent/20 p-4 space-y-2">
            <div className="flex items-start gap-2">
              <Info className="h-4 w-4 mt-0.5 flex-shrink-0" />
              <div className="space-y-1 text-xs">
                <p className="font-semibold uppercase tracking-wider">About Allowances</p>
                <ul className="space-y-0.5 text-muted-foreground">
                  <li>• Lets the contract spend tokens on your behalf</li>
                  <li>• Required before making deposits</li>
                  <li>• You can approve unlimited or specific amount</li>
                </ul>
              </div>
            </div>
          </div>
        </CardContent>
      )}
    </Card>
  );
}
