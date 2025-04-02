import { Hex } from "viem";

/**
 * Utility functions for NitroRPC
 */

/**
 * Create a payload hex string for signing
 *
 * @param data The data to convert to a hex string
 * @returns A hex string representation of the data
 */
export function createPayload(data: any): string {
    // Optimize performance by using Buffer directly
    // This is more efficient than the TextEncoder approach
    return JSON.stringify(data);
}

/**
 * Get the current time in milliseconds
 *
 * @returns The current timestamp in milliseconds
 */
export function getCurrentTimestamp(): number {
    return Date.now();
}

/**
 * Generate a unique request ID
 *
 * @returns A unique request ID
 */
export function generateRequestId(): number {
    return Math.floor(Date.now() + Math.random() * 10000);
}

/**
 * Extract the request ID from a message
 *
 * @param message The message to extract from
 * @returns The request ID, or undefined if not found
 */
export function getRequestId(message: any): number | undefined {
    if (message.req) return message.req[0];
    if (message.res) return message.res[0];
    if (message.err) return message.err[0];
    return undefined;
}

/**
 * Extract the method name from a request or response
 *
 * @param message The message to extract from
 * @returns The method name, or undefined if not found
 */
export function getMethod(message: any): string | undefined {
    if (message.req) return message.req[1];
    if (message.res) return message.res[1];
    return undefined;
}

/**
 * Extract parameters from a request
 *
 * @param message The request message
 * @returns The parameters, or an empty array if not found
 */
export function getParams(message: any): any[] {
    if (message.req) return message.req[2] || [];
    return [];
}

/**
 * Extract result from a response
 *
 * @param message The response message
 * @returns The result, or an empty array if not found
 */
export function getResult(message: any): any[] {
    if (message.res) return message.res[2] || [];
    return [];
}

/**
 * Extract error details from an error message
 *
 * @param message The error message
 * @returns The error details, or undefined if not found
 */
export function getError(message: any): { code: number; message: string } | undefined {
    if (message.err) {
        return {
            code: message.err[1],
            message: message.err[2],
        };
    }
    return undefined;
}
