/**
 * Safely stringify objects that may contain BigInt values
 */
export function safeStringify(obj: any, space?: number): string {
  try {
    return JSON.stringify(
      obj,
      (_, value) => (typeof value === 'bigint' ? value.toString() : value),
      space
    );
  } catch (error) {
    return `Error serializing: ${error instanceof Error ? error.message : String(error)}`;
  }
}

/**
 * Format address for display (0x1234...5678)
 */
export function formatAddress(address: string): string {
  if (!address || address.length < 10) return address;
  return `${address.slice(0, 6)}...${address.slice(-4)}`;
}

/**
 * Format large numbers with commas
 */
export function formatNumber(num: number | string | bigint): string {
  const numStr = num.toString();
  return numStr.replace(/\B(?=(\d{3})+(?!\d))/g, ',');
}
