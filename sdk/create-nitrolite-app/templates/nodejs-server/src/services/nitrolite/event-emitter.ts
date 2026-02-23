import type { WSStatus } from '../../types/index.js';

export class EventEmitter<T = any> {
    private listeners = new Set<(data: T) => void>();

    add(listener: (data: T) => void): () => void {
        this.listeners.add(listener);
        return () => this.listeners.delete(listener);
    }

    emit(data: T): void {
        this.listeners.forEach((listener) => listener(data));
    }

    clear(): void {
        this.listeners.clear();
    }

    get size(): number {
        return this.listeners.size;
    }
}

export class NitroliteEventEmitter {
    readonly status = new EventEmitter<WSStatus>();
    readonly message = new EventEmitter<any>();
    readonly error = new EventEmitter<Error>();

    clear(): void {
        this.status.clear();
        this.message.clear();
        this.error.clear();
    }
}