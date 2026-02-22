import type { NitroliteClient } from './client';
import type { LedgerChannel, LedgerBalance, ClearNodeAsset } from './types';

export interface EventPollerCallbacks {
    onChannelUpdate?: (channels: LedgerChannel[]) => void;
    onBalanceUpdate?: (balances: LedgerBalance[]) => void;
    onAssetsUpdate?: (assets: ClearNodeAsset[]) => void;
    onError?: (error: Error) => void;
}

/**
 * Polls the v1.0.0 Client for state changes and dispatches synthetic events
 * that match the v0.5.3 push event shapes.
 */
export class EventPoller {
    private intervalId: ReturnType<typeof setInterval> | null = null;
    private running = false;

    constructor(
        private client: NitroliteClient,
        private callbacks: EventPollerCallbacks,
        private intervalMs = 5000,
    ) {}

    start(): void {
        if (this.running) return;
        this.running = true;
        this.poll();
        this.intervalId = setInterval(() => this.poll(), this.intervalMs);
    }

    stop(): void {
        this.running = false;
        if (this.intervalId) {
            clearInterval(this.intervalId);
            this.intervalId = null;
        }
    }

    setInterval(ms: number): void {
        this.intervalMs = ms;
        if (this.running) {
            this.stop();
            this.start();
        }
    }

    private async poll(): Promise<void> {
        try {
            const [channels, balances, assets] = await Promise.allSettled([
                this.client.getChannels(),
                this.client.getBalances(),
                this.client.getAssetsList(),
            ]);

            if (channels.status === 'fulfilled' && this.callbacks.onChannelUpdate) {
                this.callbacks.onChannelUpdate(channels.value);
            }
            if (balances.status === 'fulfilled' && this.callbacks.onBalanceUpdate) {
                this.callbacks.onBalanceUpdate(balances.value);
            }
            if (assets.status === 'fulfilled' && this.callbacks.onAssetsUpdate) {
                this.callbacks.onAssetsUpdate(assets.value);
            }
        } catch (err) {
            this.callbacks.onError?.(err instanceof Error ? err : new Error(String(err)));
        }
    }
}
