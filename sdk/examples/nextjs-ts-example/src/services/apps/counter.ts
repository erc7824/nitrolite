import APP_CONFIG from '@/config/app';
import { Message } from '@/types';
import { AppLogic, Channel, State } from '@erc7824/nitrolite';
import { Address, encodeAbiParameters, Hex, decodeAbiParameters } from 'viem';

/**
 * CounterApp: A simple state channel application that implements a counter
 * 
 * This application demonstrates the basic concepts of implementing a custom
 * state channel application using the Nitrolite SDK. The counter can be
 * incremented, decremented, or reset, and demonstrates custom validation logic.
 */
export class CounterApp implements AppLogic<Message> {
    /**
     * Encode application state to bytes for on-chain use
     * @param data The application state to encode
     * @returns Hex-encoded data
     */
    public encode(data: Message): Hex {
        return encodeAbiParameters(
            [
                {
                    name: 'counter',
                    type: 'uint256',
                    internalType: 'uint256',
                },
                {
                    name: 'sequence',
                    type: 'uint256',
                    internalType: 'uint256',
                }
            ],
            [BigInt(data.text || '0'), BigInt(data.sequence || '0')],
        );
    }

    /**
     * Decode bytes back to application state
     * @param encoded The encoded state
     * @returns Decoded application state
     */
    public decode(encoded: Hex): Message {
        const [counter, sequence] = decodeAbiParameters(
            [
                {
                    name: 'counter',
                    type: 'uint256',
                    internalType: 'uint256',
                },
                {
                    name: 'sequence',
                    type: 'uint256',
                    internalType: 'uint256',
                }
            ],
            encoded,
        );

        return {
            text: counter.toString(),
            type: 'system',
            sequence: sequence.toString(),
        };
    }

    /**
     * Validate a state transition
     * @param _channel The channel context
     * @param prevState Previous application state
     * @param nextState New application state
     * @returns Whether the transition is valid
     */
    public validateTransition(_channel: Channel, prevState: Message, nextState: Message): boolean {
        // State transitions must have increasing sequence numbers
        const prevSequence = BigInt(prevState.sequence || '0');
        const nextSequence = BigInt(nextState.sequence || '0');
        
        if (nextSequence <= prevSequence) {
            return false;
        }
        
        // Counter can increment, decrement, or reset to 0
        const prevValue = BigInt(prevState.text || '0');
        const nextValue = BigInt(nextState.text || '0');
        
        return (
            nextValue === prevValue + BigInt(1) || // Increment
            nextValue === prevValue - BigInt(1) || // Decrement
            nextValue === BigInt(0)                // Reset
        );
    }

    /**
     * Provide proof states for dispute resolution
     * @param _channel The channel context
     * @param _state Current state
     * @param _previousStates Previous states
     * @returns Array of proof states
     */
    public provideProofs(_channel: Channel, _state: Message, _previousStates: State[]): State[] {
        // No additional proofs needed for this simple application
        return [];
    }

    /**
     * Determine if a state is final
     * @param state The application state to check
     * @returns Whether the state is final
     */
    public isFinal(state: Message): boolean {
        // The application is considered final when counter is 0
        return state.text === '0';
    }

    /**
     * Get the address of the adjudicator contract
     * @returns Adjudicator contract address
     */
    public getAdjudicatorAddress(): Address {
        // TODO:
        return APP_CONFIG.ADJUDICATORS.flag as Address;
    }

    /**
     * Get the type of adjudicator
     * @returns Adjudicator type string
     */
    public getAdjudicatorType(): string {
        return 'counter';
    }
}