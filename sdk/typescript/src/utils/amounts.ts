/**
 * Utility functions for converting between human-readable token amounts and bigint values.
 * These functions handle decimal precision without external dependencies.
 */

/**
 * Convert a bigint token amount to a human-readable string.
 * @param amount The amount in smallest token unit (wei-equivalent).
 * @param decimals The number of decimals for the token (e.g., 18 for most ERC20s).
 * @returns Human-readable string representation (e.g., "1.5").
 * @example
 * formatTokenAmount(1500000000000000000n, 18) // "1.5"
 * formatTokenAmount(1000000n, 6) // "1.0"
 */
export function formatTokenAmount(amount: bigint, decimals: number): string {
    if (decimals < 0 || decimals > 77) {
        throw new Error(`Invalid decimals: ${decimals}. Must be between 0 and 77.`);
    }

    if (amount < 0n) {
        return '-' + formatTokenAmount(-amount, decimals);
    }

    const divisor = 10n ** BigInt(decimals);
    const integerPart = amount / divisor;
    const fractionalPart = amount % divisor;

    if (fractionalPart === 0n) {
        return integerPart.toString();
    }

    // Pad fractional part with leading zeros if needed
    const fractionalStr = fractionalPart.toString().padStart(decimals, '0');
    // Remove trailing zeros
    const trimmedFractional = fractionalStr.replace(/0+$/, '');

    return `${integerPart}.${trimmedFractional}`;
}

/**
 * Convert a human-readable token amount string to bigint.
 * @param amount The human-readable amount (e.g., "1.5", "0.001").
 * @param decimals The number of decimals for the token.
 * @returns The amount in smallest token unit as bigint.
 * @throws Error if the amount format is invalid or has too many decimal places.
 * @example
 * parseTokenAmount("1.5", 18) // 1500000000000000000n
 * parseTokenAmount("1", 6) // 1000000n
 */
export function parseTokenAmount(amount: string, decimals: number): bigint {
    if (decimals < 0 || decimals > 77) {
        throw new Error(`Invalid decimals: ${decimals}. Must be between 0 and 77.`);
    }

    // Handle negative amounts
    if (amount.startsWith('-')) {
        return -parseTokenAmount(amount.slice(1), decimals);
    }

    // Remove whitespace
    const trimmed = amount.trim();

    // Validate format
    if (!/^[0-9]*\.?[0-9]*$/.test(trimmed)) {
        throw new Error(`Invalid amount format: ${amount}. Must be a valid number.`);
    }

    if (trimmed === '' || trimmed === '.') {
        throw new Error(`Invalid amount: ${amount}. Cannot be empty or just a decimal point.`);
    }

    const parts = trimmed.split('.');

    // Integer part (default to "0" if not present)
    const integerPart = parts[0] || '0';

    // Fractional part (default to "" if not present)
    const fractionalPart = parts[1] || '';

    // Check if fractional part has too many decimals
    if (fractionalPart.length > decimals) {
        throw new Error(
            `Amount ${amount} has too many decimal places. Maximum ${decimals} decimals allowed.`,
        );
    }

    // Pad fractional part to match decimals
    const paddedFractional = fractionalPart.padEnd(decimals, '0');

    // Combine and convert to bigint
    const combined = integerPart + paddedFractional;

    return BigInt(combined);
}
