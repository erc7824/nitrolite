import { Hex, keccak256, stringToBytes } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';

export const signRawECDSAMessage = async (message: string, privateKey: Hex): Promise<Hex> => {
    const messageBytes = keccak256(stringToBytes(message));
    const flatSignature = await privateKeyToAccount(privateKey).sign({ hash: messageBytes });

    return flatSignature;
};
