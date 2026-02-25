import { Address, Hex, concatHex, encodeAbiParameters, keccak256 } from 'viem';

import { RPCAppStateIntent } from './types';

const WALLET_QUORUM_PREFIX = '0xa1' as Hex;

export interface CreateAppSessionHashParticipant {
    walletAddress: Address | Hex;
    signatureWeight: number;
}

export interface CreateAppSessionHashParams {
    application: string;
    participants: CreateAppSessionHashParticipant[];
    quorum: number;
    nonce: bigint | number;
    sessionData?: string;
}

export interface SubmitAppStateHashAllocation {
    participant: Address | Hex;
    asset: string;
    amount: string;
}

export interface SubmitAppStateHashParams {
    appSessionId: Hex | string;
    intent: RPCAppStateIntent | 'close' | number;
    version: bigint | number;
    allocations: SubmitAppStateHashAllocation[];
    sessionData?: string;
}

function normalizeIntent(intent: SubmitAppStateHashParams['intent']): number {
    if (typeof intent === 'number') return intent;

    switch (intent) {
        case RPCAppStateIntent.Operate:
            return 0;
        case RPCAppStateIntent.Deposit:
            return 1;
        case RPCAppStateIntent.Withdraw:
            return 2;
        case 'close':
            return 3;
        default:
            throw new Error(`Unsupported app state intent: ${intent}`);
    }
}

/**
 * Deterministic hash for app-session creation quorum signatures.
 */
export function packCreateAppSessionHash(params: CreateAppSessionHashParams): Hex {
    return keccak256(
        encodeAbiParameters(
            [
                { type: 'string' },
                {
                    type: 'tuple[]',
                    components: [
                        { name: 'walletAddress', type: 'address' },
                        { name: 'signatureWeight', type: 'uint8' },
                    ],
                },
                { type: 'uint8' },
                { type: 'uint64' },
                { type: 'string' },
            ],
            [
                params.application,
                params.participants.map((participant) => ({
                    walletAddress: participant.walletAddress,
                    signatureWeight: participant.signatureWeight,
                })),
                params.quorum,
                BigInt(params.nonce),
                params.sessionData ?? '',
            ],
        ),
    );
}

/**
 * Deterministic hash for app-state update quorum signatures.
 */
export function packSubmitAppStateHash(params: SubmitAppStateHashParams): Hex {
    return keccak256(
        encodeAbiParameters(
            [
                { type: 'bytes32' },
                { type: 'uint8' },
                { type: 'uint64' },
                {
                    type: 'tuple[]',
                    components: [
                        { name: 'participant', type: 'address' },
                        { name: 'asset', type: 'string' },
                        { name: 'amount', type: 'string' },
                    ],
                },
                { type: 'string' },
            ],
            [
                params.appSessionId as Hex,
                normalizeIntent(params.intent),
                BigInt(params.version),
                params.allocations.map((allocation) => ({
                    participant: allocation.participant,
                    asset: allocation.asset,
                    amount: allocation.amount,
                })),
                params.sessionData ?? '',
            ],
        ),
    );
}

/**
 * Prefixes a wallet EIP-191 signature for quorum_sigs consumption by app sessions.
 */
export function toWalletQuorumSignature(signature: Hex | string): Hex {
    if (!signature.startsWith('0x')) {
        throw new Error('Signature must be a hex string with 0x prefix');
    }

    if (signature.toLowerCase().startsWith(WALLET_QUORUM_PREFIX)) {
        return signature as Hex;
    }

    return concatHex([WALLET_QUORUM_PREFIX, signature as Hex]);
}
