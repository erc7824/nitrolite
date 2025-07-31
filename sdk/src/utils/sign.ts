import { Hex } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';

export const signRawECDSAMessage = async (message: Hex, privateKey: Hex): Promise<Hex> => {
    const flatSignature = await privateKeyToAccount(privateKey).sign({ hash: message });

    return flatSignature;
};
