import { Hex, stringToHex } from 'viem';
import { NitroliteRPCMessage, RPCMethod, RPCResponseParamsByMethod, RPCResponse } from './types';

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

/**
 * Parses a raw RPC response string into a structured RPCResponse object
 * @param response The raw RPC response string to parse
 * @returns An RPCResponse object containing the parsed data
 */
export function parseRPCResponse(response: string): RPCResponse {
    // TODO: Add support for other rpc protocols besides websocket
    try {
        const parsed = JSON.parse(response);

        if (!Array.isArray(parsed.res) || parsed.res.length !== 4) {
            throw new Error('Invalid RPC response format');
        }

        const method = parsed.res[1] as keyof RPCResponseParamsByMethod;
        const responseObj = {
            method: method as RPCMethod,
            requestId: parsed.res[0],
            timestamp: parsed.res[3],
            signatures: parsed.sig || [],
            params: parseRPCParameters(method, parsed.res[2]),
        } as RPCResponse;

        return responseObj;
    } catch (e) {
        throw new Error(`Failed to parse RPC response: ${e instanceof Error ? e.message : e}`);
    }
}

function parseRPCParameters<M extends keyof RPCResponseParamsByMethod>(
    _: M,
    params: Array<any>,
): RPCResponseParamsByMethod[M][] {
    const result: RPCResponseParamsByMethod[M][] = [];

    if (Array.isArray(params) && params.length > 0) {
        const data = params[0];
        if (Array.isArray(data)) {
            // If data is an array, push each element
            for (const item of data) {
                if (typeof item === 'object' && item !== null) {
                    result.push(item as RPCResponseParamsByMethod[M]);
                }
            }
        } else if (typeof data === 'object' && data !== null) {
            // If data is an object, push it directly
            result.push(data as RPCResponseParamsByMethod[M]);
        }
    }

    return result;
}
