import type { Address } from 'viem';

// Common token decimals
const TOKEN_DECIMALS: Record<string, number> = {
  // USDC on Polygon
  '0x3c499c542cef5e3811e1192ce70d8cc03d5c3359': 6,
  // USDC on Ethereum
  '0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48': 6,
  // USDC on most testnets (including Goerli)
  '0x07865c6e87b9f70255377e024ace6630c1eaa37f': 6,
  // Add more tokens as needed
};

// Default decimal places if token not found
const DEFAULT_DECIMALS = 18;

/**
 * Get the number of decimal places for a token
 * @param tokenAddress The token address
 * @returns The number of decimal places
 */
export function getTokenDecimals(tokenAddress: Address): number {
  if (!tokenAddress) {
    console.warn('Token address is empty, using default decimals');
    return DEFAULT_DECIMALS;
  }
  
  const normalizedAddress = tokenAddress.toLowerCase();
  const decimals = TOKEN_DECIMALS[normalizedAddress];
  
  if (decimals !== undefined) {
    return decimals;
  }
  
  // If the token appears to be USDC based on its address
  if (normalizedAddress.includes('usdc')) {
    console.log(`Token address ${tokenAddress} appears to be USDC, using 6 decimals`);
    return 6;
  }
  
  console.log(`Token decimals not found for ${tokenAddress}, using default ${DEFAULT_DECIMALS}`);
  return DEFAULT_DECIMALS;
}

/**
 * Parse a token amount from human-readable to contract units
 * @param tokenAddress The token address
 * @param amount The amount as a string
 * @returns The amount in BigInt
 */
export function parseTokenUnits(tokenAddress: Address, amount: string): bigint {
  try {
    const decimals = getTokenDecimals(tokenAddress);
    
    // Handle empty or invalid amounts
    if (!amount || amount === '' || isNaN(Number(amount))) {
      console.warn(`Invalid amount provided: "${amount}", defaulting to 0`);
      return BigInt(0);
    }
    
    // Handle numbers with decimal points
    const parts = amount.split('.');
    const wholePart = parts[0] || '0';
    let fractionalPart = parts[1] || '';
    
    // Pad or truncate fractional part to match decimals
    if (fractionalPart.length > decimals) {
      fractionalPart = fractionalPart.substring(0, decimals);
    } else {
      while (fractionalPart.length < decimals) {
        fractionalPart += '0';
      }
    }
    
    // Remove leading zeros from whole part
    const wholePartWithoutLeadingZeros = wholePart.replace(/^0+/, '') || '0';
    
    // Combine parts without decimal point
    const combinedString = wholePartWithoutLeadingZeros + fractionalPart;
    
    // Convert to BigInt
    return BigInt(combinedString);
  } catch (error) {
    console.error(`Error parsing token units:`, error);
    // Return 0 as a fallback to prevent crashes
    return BigInt(0);
  }
}

/**
 * Format a token amount from contract units to human-readable
 * @param tokenAddress The token address
 * @param amountBigInt The amount as a BigInt
 * @returns The formatted amount as a string
 */
export function formatTokenUnits(tokenAddress: Address, amountBigInt: bigint): string {
  try {
    const decimals = getTokenDecimals(tokenAddress);
    
    // Handle zero amount
    if (amountBigInt === BigInt(0)) {
      return '0';
    }
    
    // Convert to string and pad with leading zeros if needed
    let amountStr = amountBigInt.toString();
    while (amountStr.length <= decimals) {
      amountStr = '0' + amountStr;
    }
    
    // Split into whole and fractional parts
    const wholePart = amountStr.slice(0, -decimals) || '0';
    const fractionalPart = amountStr.slice(-decimals);
    
    // Format with decimal point and trim trailing zeros
    let result = `${wholePart}.${fractionalPart}`;
    
    // Remove trailing zeros after decimal point
    result = result.replace(/\.?0+$/, '');
    if (result.endsWith('.')) {
      result = result.slice(0, -1);
    }
    
    return result || '0';
  } catch (error) {
    console.error(`Error formatting token units:`, error);
    // Return '0' as a fallback
    return '0';
  }
}