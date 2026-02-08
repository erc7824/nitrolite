import { Hex, stringToHex } from 'viem';
import { NitroliteRPCMessage } from './types';

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
 * Extract timestamp from a message
 *
 * @param message The message to extract from
 * @returns The timestamp, or undefined if not found
 */
export function getTimestamp(message: any): number | undefined {
    if (message.req) return message.req[3];
    if (message.res) return message.res[3];
    if (message.err) return message.err[3];
    return undefined;
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

/**
 * Convert parameters or results to bytes format for smart contract interaction
 *
 * @param values Array of values to convert
 * @returns Array of hex strings
 */
export function toBytes(values: any[]): Hex[] {
    return values.map((v) => (typeof v === 'string' ? stringToHex(v) : stringToHex(JSON.stringify(v))));
}

/**
 * Validates that a response timestamp is greater than the request timestamp
 *
 * @param request The request message
 * @param response The response message
 * @returns True if the response timestamp is valid
 */
export function isValidResponseTimestamp(request: NitroliteRPCMessage, response: NitroliteRPCMessage): boolean {
    const requestTimestamp = getTimestamp(request);
    const responseTimestamp = getTimestamp(response);

    if (requestTimestamp === undefined || responseTimestamp === undefined) {
        return false;
    }

    return responseTimestamp > requestTimestamp;
}

/**
 * Validates that a response request ID matches the request
 *
 * @param request The request message
 * @param response The response message
 * @returns True if the response request ID is valid
 */
export function isValidResponseRequestId(request: NitroliteRPCMessage, response: NitroliteRPCMessage): boolean {
    const requestId = getRequestId(request);
    const responseId = getRequestId(response);

    if (requestId === undefined || responseId === undefined) {
        return false;
    }

    return responseId === requestId;
}
