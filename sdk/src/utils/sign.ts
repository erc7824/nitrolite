import { Hex, keccak256 } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';

export const signRawECDSAMessage = async (message: Hex, privateKey: Hex): Promise<Hex> => {
    const hash = keccak256(message);
    const flatSignature = await privateKeyToAccount(privateKey).sign({ hash });

    return flatSignature;
};
