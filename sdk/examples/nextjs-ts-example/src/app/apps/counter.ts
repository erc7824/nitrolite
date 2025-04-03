import { Message } from '@/types';
import { AppLogic, Channel, State } from '@erc7824/nitrolite';
import { Address, encodeAbiParameters, Hex, decodeAbiParameters } from 'viem';

export class CounterApp implements AppLogic<Message> {
    public encode(data: Message): Hex {
        return encodeAbiParameters(
            [
                {
                    name: 'counter',
                    type: 'uint256',
                    internalType: 'uint256',
                },
            ],
            [BigInt(data.text)],
        );
    }

    public decode(encoded: Hex): Message {
        const [counter] = decodeAbiParameters(
            [
                {
                    name: 'counter',
                    type: 'uint256',
                    internalType: 'uint256',
                },
            ],
            encoded,
        );

        return {
            text: counter.toString(),
            type: 'system',
        };
    }

    public validateTransition(channel: Channel, prevState: Message, nextState: Message): boolean {
        return true;
    }

    public provideProofs(channel: Channel, state: Message, previousStates: State[]): State[] {
        return [];
    }

    public isFinal(state: Message): boolean {
        return state.text === '0';
    }

    public getAdjudicatorAddress(): Address {
        return '0x5fbdb2315678afecb367f032d93f642f64180aa3' as Address;
    }

    public getAdjudicatorType(): string {
        return 'counter';
    }
}
