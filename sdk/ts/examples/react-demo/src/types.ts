import { Client } from '@erc7824/nitrolite';

export interface AppState {
  client: Client | null;
  address: string | null;
  connected: boolean;
  nodeUrl: string;
  rpcConfigs: Record<string, string>; // chainId -> rpc url
  homeBlockchains: Record<string, string>; // asset -> chainId
}

export interface StatusMessage {
  type: 'success' | 'error' | 'info';
  message: string;
  details?: string;
}
