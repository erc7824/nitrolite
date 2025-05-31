import { type Hex } from "viem";
import { ethers } from "ethers";
import { type MessageSigner, type RequestData, type ResponsePayload } from "@erc7824/nitrolite";

/**
 * Interface for a cryptographic keypair
 */
export interface CryptoKeypair {
    /** Private key in hexadecimal format */
    privateKey: string;
    /** Optional Ethereum address derived from the public key */
    address?: string;
}

/**
 * Interface for a wallet signer that can sign messages
 */
export interface WalletSigner {
    /** Optional Ethereum address derived from the public key */
    address: Hex;
    /** Function to sign a message and return a hex signature */
    sign: MessageSigner;
}

/**
 * Creates a signer from a private key using ethers.js v6
 *
 * @param privateKey - The private key to create the signer from
 * @returns A WalletSigner object that can sign messages
 * @throws Error if signer creation fails
 */
export const createEthersSigner = (privateKey: string): WalletSigner => {
    try {
        // Create ethers wallet from private key
        const wallet = new ethers.Wallet(privateKey);

        return {
            address: ethers.getAddress(wallet.address) as Hex,
            sign: async (payload: RequestData | ResponsePayload): Promise<Hex> => {
                try {
                    const message = JSON.stringify(payload);
                    console.log("Signing message in Sign function:", message);
                    const digestHex = ethers.id(message);
                    console.log("Digest Hex:", digestHex);
                    const messageBytes = ethers.getBytes(digestHex);

                    const { serialized: signature } = wallet.signingKey.sign(messageBytes);

                    return signature as Hex;
                } catch (error) {
                    console.error("Error signing message:", error);
                    throw error;
                }
            },
        };
    } catch (error) {
        console.error("Error creating ethers signer:", error);
        throw error;
    }
};

/**
 * Generates a random keypair using ethers v6
 *
 * @returns A Promise resolving to a CryptoKeypair object
 */
export const generateKeyPair = async (): Promise<CryptoKeypair> => {
    try {
        // Create random wallet
        const wallet = ethers.Wallet.createRandom();

        // Hash the private key with Keccak256 for additional security
        const privateKeyHash = ethers.keccak256(wallet.privateKey as string);

        // Derive public key from hashed private key to create a new wallet
        const walletFromHashedKey = new ethers.Wallet(privateKeyHash);

        return {
            privateKey: privateKeyHash,
            address: ethers.getAddress(walletFromHashedKey.address),
        };
    } catch (error) {
        console.error("Error generating keypair, using fallback:", error);
        // Fallback implementation
        const randomHex = ethers.randomBytes(32);
        const privateKey = ethers.keccak256(randomHex);
        const wallet = new ethers.Wallet(privateKey);

        return {
            privateKey: privateKey,
            address: ethers.getAddress(wallet.address),
        };
    }
};
