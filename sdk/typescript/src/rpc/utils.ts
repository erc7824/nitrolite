import { Hex, stringToHex } from 'viem';
import { RPCMessage, RPCMessageType } from './types';

/**
 * Get the current time in milliseconds
 */
export function getCurrentTimestamp(): number {
    return Date.now();
}

/**
 * Generate a unique request ID
 */
export function generateRequestId(): number {
    return Math.floor(Date.now() + Math.random() * 10000);
}

/**
 * Extract the message type from a wire format message.
 * Wire format: [type, requestId, method, params, timestamp]
 */
export function getMessageType(message: RPCMessage): RPCMessageType {
    return message[0];
}

/**
 * Extract the request ID from a wire format message.
 * Wire format: [type, requestId, method, params, timestamp]
 */
export function getRequestId(message: RPCMessage): number {
    return message[1];
}

/**
 * Extract the method name from a wire format message.
 * Wire format: [type, requestId, method, params, timestamp]
 */
export function getMethod(message: RPCMessage): string {
    return message[2];
}

/**
 * Extract parameters from a wire format message.
 * Wire format: [type, requestId, method, params, timestamp]
 */
export function getParams(message: RPCMessage): Record<string, unknown> {
    return message[3];
}

/**
 * Extract timestamp from a wire format message.
 * Wire format: [type, requestId, method, params, timestamp]
 */
export function getTimestamp(message: RPCMessage): number {
    return message[4];
}

/**
 * Check if a message is a request.
 */
export function isRequest(message: RPCMessage): boolean {
    return message[0] === RPCMessageType.Request;
}

/**
 * Check if a message is a response.
 */
export function isResponse(message: RPCMessage): boolean {
    return message[0] === RPCMessageType.Response;
}

/**
 * Check if a message is an event.
 */
export function isEvent(message: RPCMessage): boolean {
    return message[0] === RPCMessageType.Event;
}

/**
 * Check if a message is an error response.
 */
export function isErrorResponse(message: RPCMessage): boolean {
    return message[0] === RPCMessageType.ErrorResponse;
}

/**
 * Convert parameters or results to bytes format for smart contract interaction.
 */
export function toBytes(values: any[]): Hex[] {
    return values.map((v) => (typeof v === 'string' ? stringToHex(v) : stringToHex(JSON.stringify(v))));
}

/**
 * Validates that a response timestamp is greater than the request timestamp.
 */
export function isValidResponseTimestamp(request: RPCMessage, response: RPCMessage): boolean {
    const requestTimestamp = getTimestamp(request);
    const responseTimestamp = getTimestamp(response);

    if (requestTimestamp === undefined || responseTimestamp === undefined) {
        return false;
    }

    return responseTimestamp > requestTimestamp;
}

/**
 * Validates that a response request ID matches the request.
 */
export function isValidResponseRequestId(request: RPCMessage, response: RPCMessage): boolean {
    const requestId = getRequestId(request);
    const responseId = getRequestId(response);

    if (requestId === undefined || responseId === undefined) {
        return false;
    }

    return responseId === requestId;
}
