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

/**
 * Format a balance value for display with commas and decimal places
 */
export function formatBalance(value: any): string {
  if (!value) return '0.00';
  const str = value.toString();
  const parts = str.split('.');
  parts[0] = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, ',');
  if (parts.length === 1) parts.push('00');
  return parts.join('.');
}

/**
 * Format a date/time as a relative time string (e.g., "2m ago")
 */
export function timeAgo(date: Date | string | number): string {
  const now = Date.now();
  const then = typeof date === 'number' ? date : new Date(date).getTime();
  const diff = Math.max(0, now - then);

  if (diff < 60000) return 'just now';
  if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
  if (diff < 604800000) return `${Math.floor(diff / 86400000)}d ago`;
  return new Date(then).toLocaleDateString();
}

/**
 * Format a transaction type enum to a user-friendly label
 */
export function formatTxType(
  txType: any,
  opts?: { fromAccount?: string; userAddress?: string },
): { label: string; isIncoming: boolean } {
  const str = String(txType).toLowerCase();
  if (str === '10' || str.includes('deposit'))
    return { label: 'Deposit', isIncoming: true };
  if (str === '11' || str.includes('withdraw'))
    return { label: 'Withdrawal', isIncoming: false };
  if (str === '30' || str === 'transfer' || str.includes('send')) {
    // For transfers, check from_account to determine direction
    if (opts?.userAddress && opts?.fromAccount) {
      const isSender = opts.fromAccount.toLowerCase() === opts.userAddress.toLowerCase();
      return isSender
        ? { label: 'Sent', isIncoming: false }
        : { label: 'Received', isIncoming: true };
    }
    return { label: 'Sent', isIncoming: false };
  }
  if (str === '31' || str.includes('receive'))
    return { label: 'Received', isIncoming: true };
  if (str === '200' || str.includes('finalize'))
    return { label: 'Finalize', isIncoming: false };
  return { label: String(txType), isIncoming: false };
}
