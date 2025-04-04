// Re-export wallet-related hooks
export * from './useMetaMask';
export * from './useWalletConnection';
export { getBalances } from './getTokenBalances';

// Default export for convenience
import { useMetaMask } from './useMetaMask';
export default useMetaMask;