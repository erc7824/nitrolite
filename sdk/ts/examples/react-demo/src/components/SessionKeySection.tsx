import { useState } from 'react';
import { generatePrivateKey } from 'viem/accounts';
import { privateKeyToAccount } from 'viem/accounts';
import { Key, KeyRound, Trash2, Upload } from 'lucide-react';
import { getChannelSessionKeyAuthMetadataHashV1 } from '@erc7824/nitrolite';
import type { Client } from '@erc7824/nitrolite';
import type { AppState, SessionKeyState, StatusMessage } from '../types';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { formatAddress } from '../utils';

interface SessionKeySectionProps {
  client: Client;
  appState: AppState;
  onSetSessionKey: (sk: SessionKeyState) => void;
  onActivateSessionKey: (sk: SessionKeyState) => Promise<void>;
  onClearSessionKey: () => Promise<void>;
  showStatus: (type: StatusMessage['type'], message: string, details?: string) => void;
}

export default function SessionKeySection({
  client,
  appState,
  onSetSessionKey,
  onActivateSessionKey,
  onClearSessionKey,
  showStatus,
}: SessionKeySectionProps) {
  const [loading, setLoading] = useState<string | null>(null);
  const [importKey, setImportKey] = useState('');
  const [showImport, setShowImport] = useState(false);

  // Registration form
  const [expiresHours, setExpiresHours] = useState('24');
  const [assets, setAssets] = useState('usdc');

  const sk = appState.sessionKey;
  const isRegistered = sk && sk.metadataHash && sk.authSig;

  const handleGenerate = () => {
    try {
      const privateKey = generatePrivateKey();
      const account = privateKeyToAccount(privateKey);
      const newSk: SessionKeyState = {
        privateKey,
        address: account.address,
        metadataHash: '',
        authSig: '',
        active: false,
      };
      onSetSessionKey(newSk);
      showStatus('success', 'Session key generated', `Address: ${account.address}`);
    } catch (error) {
      showStatus('error', 'Failed to generate session key', error instanceof Error ? error.message : String(error));
    }
  };

  const handleImport = () => {
    try {
      let key = importKey.trim();
      if (!key.startsWith('0x')) {
        key = `0x${key}`;
      }
      const account = privateKeyToAccount(key as `0x${string}`);
      const newSk: SessionKeyState = {
        privateKey: key,
        address: account.address,
        metadataHash: '',
        authSig: '',
        active: false,
      };
      onSetSessionKey(newSk);
      setImportKey('');
      setShowImport(false);
      showStatus('success', 'Session key imported', `Address: ${account.address}`);
    } catch (error) {
      showStatus('error', 'Invalid private key', error instanceof Error ? error.message : String(error));
    }
  };

  const handleRegister = async () => {
    if (!sk || !appState.address) return;

    try {
      setLoading('register');

      const walletAddress = appState.address;
      const sessionKeyAddr = sk.address;
      const assetList = assets.split(',').map(a => a.trim()).filter(Boolean);
      const hours = parseInt(expiresHours, 10);
      if (isNaN(hours) || hours <= 0) {
        showStatus('error', 'Invalid expiration hours');
        return;
      }

      // Determine version
      let version = 1n;
      try {
        const existing = await client.getLastChannelKeyStates(walletAddress, sessionKeyAddr);
        if (existing && existing.length > 0) {
          version = BigInt(existing[0].version) + 1n;
        }
      } catch {
        // No existing keys, version=1
      }

      const expiresAt = BigInt(Math.floor(Date.now() / 1000) + hours * 3600);

      // Build state
      const state = {
        user_address: walletAddress,
        session_key: sessionKeyAddr,
        version: version.toString(),
        assets: assetList,
        expires_at: expiresAt.toString(),
        user_sig: '',
      };

      // Sign with MetaMask (current client uses default signer)
      const sig = await client.signChannelSessionKeyState(state);
      state.user_sig = sig;

      // Submit to clearnode
      await client.submitChannelSessionKeyState(state);

      // Compute metadata hash
      const metadataHash = getChannelSessionKeyAuthMetadataHashV1(version, assetList, expiresAt);

      // Activate: store full data and reconnect
      const activeSk: SessionKeyState = {
        ...sk,
        metadataHash,
        authSig: sig,
        active: true,
      };

      await onActivateSessionKey(activeSk);
      showStatus('success', 'Session key registered and activated', `Version: ${version}, Assets: ${assetList.join(', ')}`);
    } catch (error) {
      showStatus('error', 'Registration failed', error instanceof Error ? error.message : String(error));
    } finally {
      setLoading(null);
    }
  };

  const handleClear = async () => {
    try {
      setLoading('clear');
      await onClearSessionKey();
      showStatus('success', 'Session key cleared', 'Reverted to MetaMask signing');
    } catch (error) {
      showStatus('error', 'Failed to clear session key', error instanceof Error ? error.message : String(error));
    } finally {
      setLoading(null);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Session Key</CardTitle>
        <CardDescription>Delegate state signing to a session key for seamless operations</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* Status / Generate */}
          <div className="border border-border p-4 space-y-3">
            <div className="flex items-center gap-2">
              <Key className="h-4 w-4" />
              <h3 className="text-sm font-semibold uppercase tracking-wider">Session Key</h3>
            </div>

            {!sk ? (
              <>
                <p className="text-xs text-muted-foreground">No session key configured</p>
                <div className="flex gap-2">
                  <Button onClick={handleGenerate} className="flex-1">
                    Generate New
                  </Button>
                  <Button variant="outline" onClick={() => setShowImport(!showImport)} className="flex-1">
                    <Upload className="h-3 w-3 mr-1" />
                    Import
                  </Button>
                </div>
                {showImport && (
                  <div className="space-y-2">
                    <Input
                      type="password"
                      value={importKey}
                      onChange={(e) => setImportKey(e.target.value)}
                      placeholder="Private key (0x...)"
                      className="font-mono text-xs"
                    />
                    <Button onClick={handleImport} disabled={!importKey.trim()} className="w-full">
                      Import Key
                    </Button>
                  </div>
                )}
              </>
            ) : (
              <>
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-muted-foreground">Address:</span>
                    <span className="font-mono text-xs">{formatAddress(sk.address)}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-muted-foreground">Status:</span>
                    {sk.active ? (
                      <span className="text-xs font-medium text-green-500">Active</span>
                    ) : isRegistered ? (
                      <span className="text-xs font-medium text-yellow-500">Registered</span>
                    ) : (
                      <span className="text-xs font-medium text-muted-foreground">Not registered</span>
                    )}
                  </div>
                  {isRegistered && (
                    <div className="flex items-center gap-2">
                      <span className="text-xs text-muted-foreground">Metadata:</span>
                      <span className="font-mono text-xs">{formatAddress(sk.metadataHash)}</span>
                    </div>
                  )}
                </div>
                <Button
                  variant="destructive"
                  onClick={handleClear}
                  disabled={loading === 'clear'}
                  className="w-full"
                >
                  <Trash2 className="h-3 w-3 mr-1" />
                  {loading === 'clear' ? 'Clearing...' : 'Clear Session Key'}
                </Button>
              </>
            )}
          </div>

          {/* Register on Clearnode */}
          <div className="border border-border p-4 space-y-3">
            <div className="flex items-center gap-2">
              <KeyRound className="h-4 w-4" />
              <h3 className="text-sm font-semibold uppercase tracking-wider">Register on Clearnode</h3>
            </div>

            {!sk ? (
              <p className="text-xs text-muted-foreground">Generate or import a session key first</p>
            ) : isRegistered ? (
              <p className="text-xs text-muted-foreground">Session key is already registered and {sk.active ? 'active' : 'ready to activate'}</p>
            ) : (
              <>
                <Input
                  type="text"
                  value={expiresHours}
                  onChange={(e) => setExpiresHours(e.target.value)}
                  placeholder="Expires in (hours)"
                />
                <Input
                  type="text"
                  value={assets}
                  onChange={(e) => setAssets(e.target.value)}
                  placeholder="Assets (comma-separated)"
                />
                <Button
                  onClick={handleRegister}
                  disabled={loading === 'register' || !expiresHours || !assets}
                  className="w-full"
                >
                  {loading === 'register' ? 'Registering...' : 'Register & Activate'}
                </Button>
                <p className="text-xs text-muted-foreground">
                  MetaMask will sign to authorize the session key. The client will reconnect with the new signer.
                </p>
              </>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
