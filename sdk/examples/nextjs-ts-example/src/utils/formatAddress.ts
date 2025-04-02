import { Address } from "viem";

/**
 * Formats an Ethereum address in the style of MetaMask
 * Shows first 6 chars, then ... then last 4 chars
 * @param address Ethereum address to format
 * @param short If true, only show first 6 and last 4 chars
 * @returns Formatted address string
 */
export function formatAddress(address: Address | null | undefined, short: boolean = true): string {
  if (!address) {
    return '';
  }
  
  if (short) {
    return `${address.substring(0, 6)}...${address.substring(address.length - 4)}`;
  }
  
  return address;
}