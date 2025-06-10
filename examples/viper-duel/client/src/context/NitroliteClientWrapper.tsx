"use client";

import React, { createContext, useCallback, useContext, useEffect, useState } from "react";
import { createPublicClient, http, type Hex } from "viem";
import { NitroliteStore, WalletStore } from "../store";
import { NitroliteClient, type ContractAddresses } from "@erc7824/nitrolite";

import { ethers } from "ethers";
import { generateKeyPair } from "./createSigner";
import { polygon } from "viem/chains";
import APP_CONFIG from "./app";
import { useAccount, useWalletClient } from 'wagmi';

const CRYPTO_KEYPAIR_KEY = "crypto_keypair";

export const CHAINS = polygon;

export const USDC_ADDRESS = APP_CONFIG.TOKENS[polygon.id] as Hex;

// Create context for the Nitrolite client
interface NitroliteContextType {
    client: NitroliteClient | null;
    loading: boolean;
    error: string | null;
}

const NitroliteContext = createContext<NitroliteContextType>({
    client: null,
    loading: true,
    error: null,
});

// Hook for components to use the Nitrolite client
export const useNitrolite = () => useContext(NitroliteContext);

interface NitroliteClientWrapperProps {
    children?: React.ReactNode;
}

export function NitroliteClientWrapper({ children }: NitroliteClientWrapperProps) {
    const [clientState, setClientState] = useState<NitroliteContextType>({
        client: null,
        loading: true,
        error: null,
    });

    // Use wagmi hooks for wallet connection
    const { address, isConnected } = useAccount();
    const { data: walletClient } = useWalletClient();

    const initializeKeys = useCallback(async (): Promise<{ keyPair: unknown; stateWalletClient: unknown }> => {
        try {
            let keyPair = null;
            const savedKeys = localStorage.getItem(CRYPTO_KEYPAIR_KEY);

            if (savedKeys) {
                try {
                    keyPair = JSON.parse(savedKeys);
                } catch (error) {
                    keyPair = null;
                    console.error("Failed to parse saved keys:", error);
                }
            }

            if (!keyPair) {
                keyPair = await generateKeyPair();
                if (typeof window !== "undefined") {
                    localStorage.setItem(CRYPTO_KEYPAIR_KEY, JSON.stringify(keyPair));
                }
            }

            const wallet = new ethers.Wallet(keyPair.privateKey);

            const stateWalletClient = {
                ...wallet,
                account: {
                    address: wallet.address,
                },
                signMessage: async ({ message: { raw } }: { message: { raw: string } }) => {
                    const { serialized: signature } = wallet.signingKey.sign(raw as ethers.BytesLike);

                    return signature as Hex;
                },
            };

            return { keyPair, stateWalletClient };
        } catch (error) {
            console.error("Failed to initialize keys:", error);
            return { keyPair: null, stateWalletClient: null };
        }
    }, []);

    useEffect(() => {
        const initializeNitrolite = async () => {
            try {
                setClientState((prev) => ({ ...prev, loading: true, error: null }));

                // Only proceed if wallet is connected
                if (!isConnected || !address || !walletClient) {
                    setClientState((prev) => ({
                        ...prev,
                        loading: false,
                        error: "Wallet not connected. Please connect your wallet.",
                    }));
                    return;
                }

                const keyInitResult = await initializeKeys();

                if (!keyInitResult || !keyInitResult.stateWalletClient) {
                    throw new Error("Failed to initialize state wallet client keys.");
                }
                const { stateWalletClient } = keyInitResult;

                const publicClient = createPublicClient({
                    transport: http(),
                    chain: polygon,
                });

                // Use wagmi wallet client
                console.log("Using wagmi wallet client...");
                WalletStore.setWalletClient(walletClient);
                console.log("Wallet client set successfully:", walletClient.account);

                const addresses: ContractAddresses = {
                    custody: APP_CONFIG.CUSTODIES[polygon.id],
                    adjudicator: APP_CONFIG.ADJUDICATORS[polygon.id],
                    guestAddress: APP_CONFIG.CHANNEL.DEFAULT_GUEST as Hex,
                };

                const challengeDuration = APP_CONFIG.CHANNEL.CHALLENGE_PERIOD;

                console.log("Creating Nitrolite client with params:", {
                    publicClientAvailable: !!publicClient,
                    walletClientAvailable: !!walletClient,
                    stateWalletClientAvailable: !!stateWalletClient,
                    account: walletClient.account,
                    chainId: polygon.id,
                    challengeDuration: challengeDuration.toString(),
                    addresses: {
                        custody: addresses.custody,
                        adjudicator: addresses.adjudicator,
                        guestAddress: addresses.guestAddress,
                    },
                });

                // Create the Nitrolite client
                const client = new NitroliteClient({
                    publicClient,
                    walletClient,
                    // @ts-ignore
                    stateWalletClient: stateWalletClient,
                    account: walletClient.account,
                    chainId: polygon.id,
                    challengeDuration: challengeDuration,
                    addresses,
                });

                // Check if client was created successfully
                if (!client) {
                    throw new Error("Nitrolite client creation failed - client is null");
                }

                console.log("Nitrolite client initialized successfully!");

                // Store the client in the global store for access elsewhere
                NitroliteStore.setClient(client);

                setClientState({
                    client,
                    loading: false,
                    error: null,
                });
            } catch (error: unknown) {
                console.error("Failed to initialize Nitrolite client:", error);

                // Provide more specific error messages based on the error
                let errorMessage = "Failed to initialize Nitrolite client";

                if (error instanceof Error) {
                    if (error.message.includes("provider")) {
                        errorMessage = "MetaMask provider error. Please refresh the page and try again.";
                    } else if (error.message.includes("wallet")) {
                        errorMessage = "Wallet client creation failed. Please ensure MetaMask is connected properly.";
                    } else {
                        // Include the actual error message for debugging
                        errorMessage = `Nitrolite client error: ${error.message}`;
                    }

                    // Log additional details for debugging
                    console.debug("Error details:", {
                        message: error.message,
                        stack: error.stack,
                        address: address || "not available",
                    });
                }

                setClientState({
                    client: null,
                    loading: false,
                    error: errorMessage,
                });
            }
        };

        initializeNitrolite();
    }, [initializeKeys, address, isConnected, walletClient]);

    // Provide the client through context
    return <NitroliteContext.Provider value={clientState}>{children}</NitroliteContext.Provider>;
}
