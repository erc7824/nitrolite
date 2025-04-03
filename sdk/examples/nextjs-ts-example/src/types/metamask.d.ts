interface Window {
    ethereum?: {
        isMetaMask?: boolean;
        request: (request: { method: string; params?: unknown[] }) => Promise<unknown>;
        on: (eventName: string, listener: (...args: unknown[]) => void) => void;
        removeListener: (eventName: string, listener: (...args: unknown[]) => void) => void;
        chainId: string;
        // Internal MetaMask state - may not be stable API but helps with disconnection
        _state?: {
            accounts?: string[];
            initialized?: boolean;
            isConnected?: boolean;
            isPermanentlyDisconnected?: boolean;
            isUnlocked?: boolean;
            resetState?: () => void;
        };
    };
}
