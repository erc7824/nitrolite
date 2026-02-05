import { useState, useCallback, useEffect } from 'react';
import { Client, withBlockchainRPC } from '@erc7824/nitrolite';
import { createWalletClient, custom, type WalletClient } from 'viem';
import { mainnet } from 'viem/chains';
import { WalletStateSigner, WalletTransactionSigner } from './walletSigners';
import SetupSection from './components/SetupSection';
import HighLevelOpsSection from './components/HighLevelOpsSection';
import NodeInfoSection from './components/NodeInfoSection';
import UserQueriesSection from './components/UserQueriesSection';
import LowLevelSection from './components/LowLevelSection';
import AppSessionsSection from './components/AppSessionsSection';
import StatusBar from './components/StatusBar';
import type { AppState, StatusMessage } from './types';

const DEFAULT_NODE_URL = 'wss://clearnode-v1-rc.yellow.org/ws';

// Default public RPC endpoints (no API key required)
const DEFAULT_RPC_CONFIGS: Record<string, string> = {
  '11155111': 'https://rpc.sepolia.org', // Ethereum Sepolia
};

function App() {
  const [appState, setAppState] = useState<AppState>({
    client: null,
    address: null,
    connected: false,
    nodeUrl: DEFAULT_NODE_URL,
    rpcConfigs: DEFAULT_RPC_CONFIGS,
    homeBlockchains: {},
  });
  const [status, setStatus] = useState<StatusMessage | null>(null);
  const [walletClient, setWalletClient] = useState<WalletClient | null>(null);
  const [autoConnecting, setAutoConnecting] = useState(true);

  const showStatus = useCallback((type: StatusMessage['type'], message: string, details?: string) => {
    setStatus({ type, message, details });
    setTimeout(() => setStatus(null), 5000);
  }, []);

  // Auto-reconnect on page load
  useEffect(() => {
    const autoConnect = async () => {
      try {
        // Check if MetaMask was previously connected
        const wasConnected = localStorage.getItem('metamask_connected');
        if (!wasConnected || typeof window.ethereum === 'undefined') {
          setAutoConnecting(false);
          return;
        }

        // Try to reconnect silently
        const accounts = await window.ethereum.request({
          method: 'eth_accounts'
        }) as string[];

        if (accounts && accounts.length > 0) {
          const address = accounts[0];
          const client = createWalletClient({
            account: address as `0x${string}`,
            chain: mainnet,
            transport: custom(window.ethereum),
          });

          setWalletClient(client);
          setAppState(prev => ({ ...prev, address, connected: false }));
          showStatus('success', 'Wallet reconnected', `Address: ${address}`);

          // Auto-connect to node if wallet is reconnected
          try {
            // Create signers
            const stateSigner = new WalletStateSigner(client);
            const txSigner = new WalletTransactionSigner(client);

            // Build options with RPC configs
            const options = Object.entries(DEFAULT_RPC_CONFIGS).map(([chainId, rpcUrl]) =>
              withBlockchainRPC(BigInt(chainId), rpcUrl)
            );

            // Create SDK client
            const sdkClient = await Client.create(
              DEFAULT_NODE_URL,
              stateSigner,
              txSigner,
              ...options
            );

            setAppState(prev => ({ ...prev, client: sdkClient, connected: true }));
            showStatus('success', 'Connected to Clearnode', 'Fully reconnected and ready!');
          } catch (nodeError) {
            console.error('Auto node connection failed:', nodeError);
            showStatus('info', 'Wallet reconnected', 'Click "Connect to Node" to continue');
          }
        }
      } catch (error) {
        console.error('Auto-connect failed:', error);
      } finally {
        setAutoConnecting(false);
      }
    };

    autoConnect();
  }, [showStatus]);

  // Listen for account changes
  useEffect(() => {
    if (typeof window.ethereum === 'undefined') return;

    const handleAccountsChanged = (accounts: string[]) => {
      if (accounts.length === 0) {
        // User disconnected
        localStorage.removeItem('metamask_connected');
        setWalletClient(null);
        setAppState(prev => ({ ...prev, address: null, connected: false, client: null }));
        showStatus('info', 'Wallet disconnected');
      } else if (accounts[0] !== appState.address) {
        // User switched accounts
        const newAddress = accounts[0];
        const client = createWalletClient({
          account: newAddress as `0x${string}`,
          chain: mainnet,
          transport: custom(window.ethereum),
        });
        setWalletClient(client);
        setAppState(prev => ({ ...prev, address: newAddress, connected: false, client: null }));
        showStatus('info', 'Account switched', `New address: ${newAddress}`);
      }
    };

    window.ethereum.on('accountsChanged', handleAccountsChanged);
    return () => {
      window.ethereum.removeListener('accountsChanged', handleAccountsChanged);
    };
  }, [appState.address, showStatus]);

  const connectWallet = useCallback(async () => {
    try {
      if (typeof window.ethereum === 'undefined') {
        showStatus('error', 'MetaMask not detected', 'Please install MetaMask extension');
        return;
      }

      // Request account access
      const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' }) as string[];
      if (!accounts || accounts.length === 0) {
        showStatus('error', 'No accounts found', 'Please unlock MetaMask');
        return;
      }

      const address = accounts[0];

      // Create wallet client
      const client = createWalletClient({
        account: address as `0x${string}`,
        chain: mainnet,
        transport: custom(window.ethereum),
      });

      // Save connection state
      localStorage.setItem('metamask_connected', 'true');

      setWalletClient(client);
      setAppState(prev => ({ ...prev, address, connected: false }));
      showStatus('success', 'Wallet connected', `Address: ${address}`);
    } catch (error) {
      showStatus('error', 'Failed to connect wallet', error instanceof Error ? error.message : String(error));
    }
  }, [showStatus]);

  const connectToNode = useCallback(async () => {
    if (!walletClient || !appState.address) {
      showStatus('error', 'Wallet not connected', 'Please connect MetaMask first');
      return;
    }

    try {
      // Create signers
      const stateSigner = new WalletStateSigner(walletClient);
      const txSigner = new WalletTransactionSigner(walletClient);

      // Build options with RPC configs
      const options = Object.entries(appState.rpcConfigs).map(([chainId, rpcUrl]) =>
        withBlockchainRPC(BigInt(chainId), rpcUrl)
      );

      // Create SDK client
      const client = await Client.create(
        appState.nodeUrl,
        stateSigner,
        txSigner,
        ...options
      );

      // Set home blockchains if configured
      const homeBlockchainErrors: string[] = [];
      for (const [asset, chainId] of Object.entries(appState.homeBlockchains)) {
        try {
          await client.setHomeBlockchain(asset, BigInt(chainId));
          console.log(`✓ Home blockchain set for ${asset} on chain ${chainId}`);
        } catch (error) {
          const errorMsg = error instanceof Error ? error.message : String(error);
          console.error(`✗ Failed to set home blockchain for ${asset}:`, errorMsg);
          homeBlockchainErrors.push(`${asset}: ${errorMsg}`);
        }
      }

      setAppState(prev => ({ ...prev, client, connected: true }));

      if (homeBlockchainErrors.length > 0) {
        showStatus('info', 'Connected to Clearnode (with warnings)',
          `Home blockchain setup failed:\n${homeBlockchainErrors.join('\n')}`);
      } else {
        showStatus('success', 'Connected to Clearnode', appState.nodeUrl);
      }
    } catch (error) {
      showStatus('error', 'Failed to connect to node', error instanceof Error ? error.message : String(error));
    }
  }, [walletClient, appState.address, appState.nodeUrl, appState.rpcConfigs, appState.homeBlockchains, showStatus]);

  const disconnectClient = useCallback(async () => {
    if (appState.client) {
      await appState.client.close();
      setAppState(prev => ({ ...prev, client: null, connected: false }));
      showStatus('info', 'Disconnected from node');
    }
  }, [appState.client, showStatus]);

  const disconnectWallet = useCallback(() => {
    localStorage.removeItem('metamask_connected');
    setWalletClient(null);
    if (appState.client) {
      appState.client.close();
    }
    setAppState(prev => ({
      ...prev,
      address: null,
      connected: false,
      client: null
    }));
    showStatus('info', 'Wallet disconnected');
  }, [appState.client, showStatus]);

  const updateRpcConfig = useCallback((chainId: string, rpcUrl: string) => {
    setAppState(prev => ({
      ...prev,
      rpcConfigs: { ...prev.rpcConfigs, [chainId]: rpcUrl },
    }));
  }, []);

  const removeRpcConfig = useCallback((chainId: string) => {
    setAppState(prev => {
      const newConfigs = { ...prev.rpcConfigs };
      delete newConfigs[chainId];
      return { ...prev, rpcConfigs: newConfigs };
    });
  }, []);

  const addHomeBlockchain = useCallback((asset: string, chainId: string) => {
    setAppState(prev => ({
      ...prev,
      homeBlockchains: { ...prev.homeBlockchains, [asset]: chainId },
    }));
  }, []);

  const removeHomeBlockchain = useCallback((asset: string) => {
    setAppState(prev => {
      const newHomeBlockchains = { ...prev.homeBlockchains };
      delete newHomeBlockchains[asset];
      return { ...prev, homeBlockchains: newHomeBlockchains };
    });
  }, []);

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50">
      <div className="max-w-7xl mx-auto px-4 py-8">
        {/* Header */}
        <div className="mb-8 text-center">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">
            Yellow TS  SDK Demo
          </h1>
          {autoConnecting && (
            <div className="mt-3 inline-flex items-center gap-2 px-4 py-2 bg-blue-50 border border-blue-200 rounded-lg">
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
              <span className="text-sm text-blue-700 font-medium">Reconnecting wallet and node...</span>
            </div>
          )}
        </div>

        {/* Status Bar */}
        {status && <StatusBar status={status} onClose={() => setStatus(null)} />}

        {/* Setup Section */}
        <SetupSection
          appState={appState}
          onConnectWallet={connectWallet}
          onConnectNode={connectToNode}
          onDisconnect={disconnectClient}
          onDisconnectWallet={disconnectWallet}
          onNodeUrlChange={(url) => setAppState(prev => ({ ...prev, nodeUrl: url }))}
          onAddRpc={updateRpcConfig}
          onRemoveRpc={removeRpcConfig}
          onAddHomeBlockchain={addHomeBlockchain}
          onRemoveHomeBlockchain={removeHomeBlockchain}
          showStatus={showStatus}
        />

        {/* Operations Sections - Only show when connected */}
        {appState.connected && appState.client && (
          <div className="space-y-6">
            {/* High-Level Operations */}
            <HighLevelOpsSection client={appState.client} showStatus={showStatus} />

            {/* Node Information */}
            <NodeInfoSection client={appState.client} showStatus={showStatus} />

            {/* User Queries */}
            <UserQueriesSection
              client={appState.client}
              defaultAddress={appState.address || ''}
              showStatus={showStatus}
            />

            {/* Low-Level State */}
            <LowLevelSection
              client={appState.client}
              defaultAddress={appState.address || ''}
              showStatus={showStatus}
            />

            {/* App Sessions */}
            <AppSessionsSection
              client={appState.client}
              defaultAddress={appState.address || ''}
              showStatus={showStatus}
            />
          </div>
        )}

        {/* Footer */}
        <div className="mt-12 text-center text-sm text-gray-500">
          <p>Nitrolite SDK v0.5.2 - Built with React & Tailwind</p>
          <p className="mt-1">
            <a
              href="https://github.com/erc7824/nitrolite"
              target="_blank"
              rel="noopener noreferrer"
              className="text-blue-600 hover:text-blue-800"
            >
              GitHub Repository
            </a>
          </p>
        </div>
      </div>
    </div>
  );
}

// Extend Window interface for TypeScript
declare global {
  interface Window {
    ethereum?: any;
  }
}

export default App;
