/**
 * Message interfaces for Nitrolite SDK
 *
 * This module provides type definitions for message formats,
 * but intentionally does NOT provide any implementation.
 *
 * Developers have complete freedom to implement communication
 * between participants in whatever way suits their application.
 */

import { Address, Hex } from "viem";
import { State, ChannelId, StateHash, Signature } from "../types";

/**
 * Message types for Nitrolite protocol messages
 */
export enum MessageType {
    PROPOSE_STATE = "propose_state",
    ACCEPT_STATE = "accept_state",
    REJECT_STATE = "reject_state",
    SIGN_STATE = "sign_state",
    CHALLENGE_NOTIFICATION = "challenge_notification",
    CLOSURE_NOTIFICATION = "closure_notification",
}

/**
 * Base message interface
 */
export interface BaseMessage {
    type: MessageType;
    channelId: ChannelId;
    timestamp: number;
}

/**
 * Message proposing a new state
 */
export interface ProposeStateMessage extends BaseMessage {
    type: MessageType.PROPOSE_STATE;
    state: State;
    stateHash: StateHash;
}

/**
 * Message accepting a proposed state
 */
export interface AcceptStateMessage extends BaseMessage {
    type: MessageType.ACCEPT_STATE;
    stateHash: StateHash;
    signature: Signature;
}

/**
 * Message rejecting a proposed state
 */
export interface RejectStateMessage extends BaseMessage {
    type: MessageType.REJECT_STATE;
    stateHash: StateHash;
    reason: string;
}

/**
 * Message with a signed state
 */
export interface SignStateMessage extends BaseMessage {
    type: MessageType.SIGN_STATE;
    state: State;
}

/**
 * Message notifying of an on-chain challenge
 */
export interface ChallengeNotificationMessage extends BaseMessage {
    type: MessageType.CHALLENGE_NOTIFICATION;
    expirationTime: number;
    challengeState: State;
}

/**
 * Message notifying of channel closure
 */
export interface ClosureNotificationMessage extends BaseMessage {
    type: MessageType.CLOSURE_NOTIFICATION;
    finalState: State;
}

/**
 * Union type for all message types
 */
export type NitroliteMessage =
    | ProposeStateMessage
    | AcceptStateMessage
    | RejectStateMessage
    | SignStateMessage
    | ChallengeNotificationMessage
    | ClosureNotificationMessage;
