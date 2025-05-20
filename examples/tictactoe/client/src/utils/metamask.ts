/**
 * MetaMask utility functions
 */

/**
 * Check if the MetaMask extension is installed
 * @returns boolean indicating whether MetaMask is installed
 */
export function isMetaMaskInstalled(): boolean {
  return typeof window !== 'undefined' && 
         typeof window.ethereum !== 'undefined' && 
         window.ethereum.isMetaMask === true;
}

/**
 * Format an Ethereum address for display
 * @param address Full Ethereum address
 * @returns Shortened address with ellipsis (e.g., 0x1234...5678)
 */
export function formatEthAddress(address: string): string {
  if (!address) return '';
  if (address.length < 10) return address;
  
  return `${address.substring(0, 6)}...${address.substring(address.length - 4)}`;
}