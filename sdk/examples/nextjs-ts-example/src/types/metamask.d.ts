interface Window {
  ethereum?: {
    isMetaMask?: boolean;
    request: (request: { method: string; params?: any[] }) => Promise<any>;
    on: (eventName: string, listener: (...args: any[]) => void) => void;
    removeListener: (eventName: string, listener: (...args: any[]) => void) => void;
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