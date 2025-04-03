import {
    Address,
    Hex,
    decodeAbiParameters,
    encodeAbiParameters,
    encodePacked,
} from 'viem';
import { AppLogic, Channel, State } from '../../src/types';

export interface AndAppState {
    isFinal: boolean;
    flag: boolean;
}

export class AndApp implements AppLogic<AndAppState> {
    private adjudicatorAddress: Address;

    constructor(adjudicatorAddress: Address) {
        this.adjudicatorAddress = adjudicatorAddress;
    }

    encode(data: AndAppState): Hex {
        return encodeAbiParameters(
            [
                {
                    name: 'isFinal',
                    type: 'bool',
                    internalType: 'bool',
                },
                {
                    name: 'flag',
                    type: 'bool',
                    internalType: 'bool',
                },
            ],
            [data.isFinal, data.flag]
        );
    }

    decode(encoded: Hex): AndAppState {
        const [isFinal, flag] = decodeAbiParameters(
            [
                {
                    name: 'isFinal',
                    type: 'bool',
                    internalType: 'bool',
                },
                {
                    name: 'flag',
                    type: 'bool',
                    internalType: 'bool',
                },
            ],
            encoded
        );

        return {
            isFinal,
            flag,
        };
    }

    validateTransition(
        _channel: Channel,
        prevState: AndAppState,
        _nextState: AndAppState
    ): boolean {
        // Can't transition from final state
        if (prevState.isFinal) return false;

        return true;
    }

    isFinal(state: AndAppState): boolean {
        return state.isFinal;
    }

    getAdjudicatorAddress(): Address {
        return this.adjudicatorAddress;
    }

    getAdjudicatorType(): string {
        return 'and';
    }

    provideProofs(
        _channel: Channel,
        state: AndAppState,
        previousStates: State[]
    ): State[] {
        if (state.isFinal) {
            // Provide proofs for the nearest previous state if the current state is final
            return previousStates.slice(-2, -1);
        }

        return [];
    }
}
