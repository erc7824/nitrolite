import { useState } from 'react';
import Decimal from 'decimal.js';
import { Loader2, CheckCircle2, AlertCircle, X } from 'lucide-react';
import type { Client } from '@erc7824/nitrolite';
import type { StatusMessage } from '../types';
import { Button } from './ui/button';
import { Input } from './ui/input';

const MAX_APPROVE_AMOUNT = new Decimal('1e18');

type ModalType = 'deposit' | 'withdraw' | 'transfer' | 'close';

interface ActionModalProps {
  type: ModalType;
  client: Client;
  chainId: string;
  asset: string;
  showStatus: (type: StatusMessage['type'], message: string, details?: string) => void;
  onClose: () => void;
  onComplete: () => void;
}

type StepStatus = 'pending' | 'active' | 'done' | 'error';
interface Step {
  label: string;
  status: StepStatus;
}

const MODAL_CONFIG: Record<ModalType, { title: string; verb: string; description: string }> = {
  deposit: { title: 'Deposit', verb: 'deposit', description: 'Add funds to your channel' },
  withdraw: { title: 'Withdraw', verb: 'withdrawal', description: 'Remove funds from your channel' },
  transfer: { title: 'Send', verb: 'transfer', description: 'Transfer to another address' },
  close: { title: 'Close Channel', verb: 'close', description: 'Finalize and close the channel' },
};

function isAllowanceError(error: unknown): boolean {
  const msg = error instanceof Error ? error.message : String(error);
  return msg.toLowerCase().includes('allowance') && msg.toLowerCase().includes('sufficient');
}

export default function ActionModal({
  type, client, chainId, asset, showStatus, onClose, onComplete,
}: ActionModalProps) {
  const [amount, setAmount] = useState('');
  const [recipient, setRecipient] = useState('');
  const [phase, setPhase] = useState<'form' | 'processing' | 'done' | 'error'>('form');
  const [steps, setSteps] = useState<Step[]>([]);
  const [error, setError] = useState('');
  const [checkpointFailed, setCheckpointFailed] = useState(false);
  const [txHash, setTxHash] = useState('');

  const needsCheckpoint = type === 'deposit' || type === 'withdraw' || type === 'close';
  const config = MODAL_CONFIG[type];

  const autoApprove = async (assetName: string, cid: bigint): Promise<void> => {
    await client.approveToken(cid, assetName, MAX_APPROVE_AMOUNT);
  };

  const handleSubmit = async () => {
    try {
      setPhase('processing');

      const allSteps: Step[] = [
        { label: `Processing ${config.verb}...`, status: 'active' },
      ];
      if (needsCheckpoint) {
        allSteps.push({ label: 'Syncing to chain...', status: 'pending' });
      }
      setSteps([...allSteps]);

      // Execute main operation
      const cid = BigInt(chainId);

      switch (type) {
        case 'deposit': {
          const amt = new Decimal(amount);
          try {
            await client.deposit(cid, asset, amt);
          } catch (e) {
            if (!isAllowanceError(e)) throw e;
            allSteps[0] = { label: 'Approving token...', status: 'active' };
            setSteps([...allSteps]);
            await autoApprove(asset, cid);
            allSteps[0] = { label: `Processing ${config.verb}...`, status: 'active' };
            setSteps([...allSteps]);
            await client.deposit(cid, asset, amt);
          }
          break;
        }
        case 'withdraw': {
          const amt = new Decimal(amount);
          await client.withdraw(cid, asset, amt);
          break;
        }
        case 'transfer': {
          const amt = new Decimal(amount);
          await client.transfer(recipient as `0x${string}`, asset, amt);
          break;
        }
        case 'close': {
          await client.closeHomeChannel(asset);
          break;
        }
      }

      allSteps[0] = { label: `${config.title} successful`, status: 'done' };
      setSteps([...allSteps]);

      // Auto-checkpoint if needed
      if (needsCheckpoint) {
        allSteps[1] = { label: 'Syncing to chain...', status: 'active' };
        setSteps([...allSteps]);

        try {
          let hash: string;
          try {
            hash = await client.checkpoint(asset);
          } catch (e) {
            if (!isAllowanceError(e)) throw e;
            const latestState = await client.getLatestState(
              client.getUserAddress() as `0x${string}`, asset, true,
            );
            const stateChainId = latestState.homeLedger.blockchainId;
            await autoApprove(asset, stateChainId);
            hash = await client.checkpoint(asset);
          }

          allSteps[1] = { label: 'Confirmed on-chain', status: 'done' };
          setSteps([...allSteps]);
          setTxHash(hash);
        } catch (cpErr) {
          // Checkpoint failed but main operation succeeded
          const cpMsg = cpErr instanceof Error ? cpErr.message : String(cpErr);
          allSteps[1] = { label: 'Sync failed', status: 'error' };
          setSteps([...allSteps]);
          setCheckpointFailed(true);
          setError(cpMsg);
        }
      }

      if (!needsCheckpoint) {
        showStatus('success', `${config.title} successful`);
        onComplete();
        return;
      }

      setPhase('done');
    } catch (e) {
      const msg = e instanceof Error ? e.message : String(e);
      setError(msg);
      setPhase('error');
      showStatus('error', `${config.title} failed`, msg);
    }
  };

  const canSubmit = (() => {
    if (type === 'close') return true;
    if (!amount) return false;
    if (type === 'transfer' && !recipient) return false;
    return true;
  })();

  return (
    <div
      className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50"
      onClick={phase === 'form' ? onClose : undefined}
    >
      <div
        className="bg-card border border-border p-6 max-w-md w-full mx-4 animate-scale-in"
        onClick={e => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div>
            <div className="text-lg font-semibold uppercase tracking-tight">{config.title}</div>
            <div className="text-xs text-muted-foreground">{config.description}</div>
          </div>
          {phase === 'form' && (
            <button
              onClick={onClose}
              className="text-muted-foreground hover:text-foreground transition-colors p-1"
            >
              <X className="h-5 w-5" />
            </button>
          )}
        </div>

        {/* Form phase */}
        {phase === 'form' && (
          <div className="space-y-4">
            {type === 'transfer' && (
              <div>
                <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-1.5 block">
                  Recipient
                </label>
                <Input
                  value={recipient}
                  onChange={e => setRecipient(e.target.value)}
                  placeholder="0x..."
                  className="font-mono text-xs"
                />
              </div>
            )}

            {type !== 'close' && (
              <div>
                <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-1.5 block">
                  Amount
                </label>
                <div className="relative">
                  <Input
                    value={amount}
                    onChange={e => setAmount(e.target.value)}
                    placeholder="0.00"
                    className="pr-16"
                  />
                  <span className="absolute right-3 top-1/2 -translate-y-1/2 text-xs font-semibold uppercase text-muted-foreground">
                    {asset}
                  </span>
                </div>
              </div>
            )}

            {type === 'close' && (
              <div className="bg-destructive/10 border border-destructive/30 p-3 text-xs text-destructive">
                This will finalize and close your channel. This action cannot be easily undone.
              </div>
            )}

            {needsCheckpoint && type !== 'close' && (
              <div className="text-xs text-muted-foreground">
                This operation will automatically sync to chain after completion.
              </div>
            )}

            <Button
              onClick={handleSubmit}
              disabled={!canSubmit}
              className="w-full"
              variant={type === 'close' ? 'destructive' : 'default'}
            >
              {type === 'close' ? 'Close Channel' : config.title}
            </Button>
          </div>
        )}

        {/* Processing / Done phase */}
        {(phase === 'processing' || phase === 'done') && (
          <div className="space-y-4">
            {steps.map((step, idx) => (
              <div key={idx} className="flex items-center gap-3">
                {step.status === 'active' && <Loader2 className="h-4 w-4 animate-spin text-accent flex-shrink-0" />}
                {step.status === 'done' && <CheckCircle2 className="h-4 w-4 text-green-400 flex-shrink-0" />}
                {step.status === 'pending' && <div className="h-4 w-4 border border-border rounded-full flex-shrink-0" />}
                {step.status === 'error' && <AlertCircle className="h-4 w-4 text-destructive flex-shrink-0" />}
                <span className={`text-sm ${
                  step.status === 'done' ? 'text-green-400' :
                  step.status === 'active' ? 'text-foreground' :
                  'text-muted-foreground'
                }`}>
                  {step.label}
                </span>
              </div>
            ))}

            {txHash && (
              <div className="mt-2 pt-3 border-t border-border">
                <div className="text-xs text-muted-foreground mb-1">Transaction</div>
                <div className="text-xs font-mono break-all text-muted-foreground">{txHash}</div>
              </div>
            )}

            {phase === 'done' && checkpointFailed && (
              <div className="bg-yellow-500/10 border border-yellow-500/30 p-3">
                <div className="text-xs font-semibold text-yellow-500 mb-1 uppercase tracking-wider">Sync failed</div>
                <div className="text-xs text-yellow-500/80 max-h-16 overflow-y-auto break-all">{error}</div>
                <div className="text-xs text-muted-foreground mt-2">
                  Your {config.verb} succeeded. You can sync to chain later using the Sync button on the dashboard.
                </div>
              </div>
            )}

            {phase === 'done' && (
              <Button onClick={onComplete} className="w-full mt-2">
                Done
              </Button>
            )}
          </div>
        )}

        {/* Error phase */}
        {phase === 'error' && (
          <div className="space-y-4">
            {steps.filter(s => s.status === 'done').map((step, idx) => (
              <div key={idx} className="flex items-center gap-3">
                <CheckCircle2 className="h-4 w-4 text-green-400 flex-shrink-0" />
                <span className="text-sm text-green-400">{step.label}</span>
              </div>
            ))}

            <div className="bg-destructive/10 border border-destructive/30 p-3">
              <div className="text-xs font-semibold text-destructive mb-1 uppercase tracking-wider">Error</div>
              <div className="text-xs text-destructive/80 break-all max-h-24 overflow-y-auto">{error}</div>
            </div>

            <div className="flex gap-3">
              <Button
                onClick={() => { setPhase('form'); setSteps([]); setError(''); setTxHash(''); }}
                variant="outline"
                className="flex-1"
              >
                Try Again
              </Button>
              <Button onClick={onClose} variant="outline" className="flex-1">
                Close
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
