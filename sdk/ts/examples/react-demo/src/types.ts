import { Client } from '@erc7824/nitrolite';

export interface SessionKeyState {
  privateKey: string;   // hex private key
  address: string;      // derived session key address
  metadataHash: string; // from registration (empty if not registered)
  authSig: string;      // from registration (empty if not registered)
  active: boolean;      // whether client currently uses this signer
}

export interface AppState {
  client: Client | null;
  address: string | null;
  connected: boolean;
  nodeUrl: string;
  rpcConfigs: Record<string, string>; // chainId -> rpc url
  homeBlockchains: Record<string, string>; // asset -> chainId
  sessionKey: SessionKeyState | null;
}

export interface StatusMessage {
  type: 'success' | 'error' | 'info';
  message: string;
  details?: string;
}
