import { useCallback, useEffect } from 'react';
import { NitroliteStore, WalletStore } from '../store';
import { useStore } from '../store/storeUtils';
import { useWebSocketContext } from '../context/WebSocketContext';
import { useChannel } from './useChannel';

/**
 * This hook integrates the WebSocket connections with Nitrolite channels
 * It should be used at app level to handle WebSocket and Channel connections
 */
export function useNitroliteIntegration() {
  const { status, isConnected } = useWebSocketContext();
  const walletState = useStore(WalletStore.state);
  const nitroliteState = useStore(NitroliteStore.state);
  const { currentChannel, clearStoredChannel } = useChannel();

  /**
   * Handle recovered or existing channels
   */
  useEffect(() => {
    // If we're connected to WebSocket and have a client but no channel, try to recover it
    if (isConnected && nitroliteState.client && !currentChannel) {
      const channelId = localStorage.getItem('nitrolite_channel_id');
      const channelState = localStorage.getItem('nitrolite_channel_state');
      
      if (channelId && channelState) {
        try {
          // This is a simplified version - in a real app you would parse the state and reconnect
          console.log('Found saved channel, should reconnect:', channelId);
          
          // Mark as having an open channel
          WalletStore.setChannelOpen(true);
        } catch (error) {
          console.error('Failed to recover channel:', error);
          clearStoredChannel();
        }
      } else {
        console.log('No existing channel data found in localStorage');
        // Ensure channels are marked as closed if no data exists
        WalletStore.setChannelOpen(false);
      }
    }
  }, [isConnected, nitroliteState.client, currentChannel, clearStoredChannel]);

  /**
   * Initialize the Nitrolite client when necessary
   */
  const initializeNitroliteClient = useCallback(async (clientInstance: any) => {
    NitroliteStore.setClient(clientInstance);
  }, []);

  /**
   * Handle WebSocket disconnection
   */
  useEffect(() => {
    if (status === 'disconnected' && walletState.channelOpen) {
      // Websocket was disconnected but we have an open channel
      // This is just logging for now, but you could implement reconnection logic
      console.log('WebSocket disconnected while channel is open - should attempt to reconnect');
    }
  }, [status, walletState.channelOpen]);

  return {
    wsStatus: status,
    isWsConnected: isConnected,
    hasOpenChannel: walletState.channelOpen,
    currentChannelId: currentChannel ? JSON.stringify(currentChannel).substring(0, 20) : null,
    initializeNitroliteClient
  };
}