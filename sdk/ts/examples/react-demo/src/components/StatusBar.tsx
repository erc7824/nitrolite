import { X, CheckCircle, AlertCircle, Info } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { StatusMessage } from '../types';

interface StatusBarProps {
  status: StatusMessage;
  onClose: () => void;
}

export default function StatusBar({ status, onClose }: StatusBarProps) {
  const variants = {
    success: {
      bg: 'bg-accent/10 border-accent',
      text: 'text-foreground',
      icon: CheckCircle,
    },
    error: {
      bg: 'bg-destructive/10 border-destructive',
      text: 'text-destructive',
      icon: AlertCircle,
    },
    info: {
      bg: 'bg-muted border-border',
      text: 'text-foreground',
      icon: Info,
    },
  };

  const { bg, text, icon: Icon } = variants[status.type];

  return (
    <div
      className={cn(
        'fixed top-6 right-6 max-w-md border-l-4 border p-4 z-50 animate-in slide-in-from-right-5',
        bg
      )}
    >
      <div className="flex items-start gap-3">
        <Icon className={cn('h-5 w-5 flex-shrink-0 mt-0.5', text)} />
        <div className="flex-1 space-y-1">
          <p className="text-sm font-semibold uppercase tracking-wide">{status.message}</p>
          {status.details && (
            <p className="text-xs text-muted-foreground font-mono break-all">{status.details}</p>
          )}
        </div>
        <button
          onClick={onClose}
          className="flex-shrink-0 hover:bg-muted p-1 transition-colors"
        >
          <X className="h-4 w-4" />
        </button>
      </div>
    </div>
  );
}
