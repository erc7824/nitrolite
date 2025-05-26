<script setup lang="ts">
import { ref, onMounted } from "vue";
import { ethers } from "ethers";
import { type NitroliteClientConfig } from "@erc7824/nitrolite";
import { generateKeyPair } from "../crypto";
import { createPublicClient, createWalletClient, custom, http } from "viem";
import type { Hex } from "viem";
import clearNetService from "../services/ClearNetService";
import { CONTRACT_ADDRESSES } from "../config";
import { polygon } from "viem/chains";
import { createEthersSigner } from "../crypto";

// Component state
const isConnected = ref(false);
const isConnecting = ref(false);
const isAuthenticated = ref(false);
const walletAddress = ref("");
const balance = ref(BigInt(0));
const walletError = ref("");

// Storage keys
const KEY_PAIR = "crypto_keypair";

// Event emitters
const emit = defineEmits<{
    "wallet-connected": [{ address: string; balance: bigint }];
    "wallet-disconnected": [];
    error: [string];
}>();

onMounted(() => { connectWallet(); });

// Wallet signer interface following server implementation
interface WalletSigner {
    publicKey: string;
    address: string;
    sign: (payload: any) => Promise<string>;
}

/**
 * Gets or creates a wallet signer with a private key stored in localStorage
 */
async function getOrCreateWalletSigner(): Promise<WalletSigner> {
    let keyPair = null;
    const savedKeys = localStorage.getItem(KEY_PAIR);

    if (savedKeys) {
        try {
            keyPair = JSON.parse(savedKeys);
        } catch (error) {
            keyPair = null;
        }
    }

    if (!keyPair) {
        keyPair = await generateKeyPair();
        if (typeof window !== "undefined") {
            localStorage.setItem(KEY_PAIR, JSON.stringify(keyPair));
        }
    }

    return createEthersSigner(keyPair.privateKey);
}

/**
 * Main connect function that establishes connection and authenticates
 */
async function connectWallet() {
    if (isConnecting.value) return;

    isConnecting.value = true;
    walletError.value = "";

    try {
        // Get or create wallet signer
        const signer = await getOrCreateWalletSigner();
        walletAddress.value = signer.address;

        console.log("Using wallet with address:", signer.address);

        // Set fake balance for display purposes
        balance.value = BigInt(1000000000000000000); // 1 ETH

        // Update connection state
        isConnected.value = true;

        // Initialize ClearNetService with proper wallet configuration
        try {
            // Generate or retrieve state keys for signing
            const CRYPTO_KEYPAIR_KEY = "crypto_keypair";
            let keyPair = null;
            const savedKeys = localStorage.getItem(CRYPTO_KEYPAIR_KEY);

            if (savedKeys) {
                try {
                    keyPair = JSON.parse(savedKeys);
                } catch (error) {
                    keyPair = null;
                }
            }

            if (!keyPair) {
                keyPair = await generateKeyPair();
                localStorage.setItem(CRYPTO_KEYPAIR_KEY, JSON.stringify(keyPair));
            }

            // Create a dedicated state wallet for signing
            const stateWallet = new ethers.Wallet(keyPair.privateKey);

            // Create a proper stateWalletClient for signing state updates
            const stateWalletClient = {
                account: {
                    address: stateWallet.address,
                },
                signMessage: async ({ message: { raw } }: { message: { raw: string } }) => {
                    try {
                        const flatSignature = stateWallet._signingKey().signDigest(raw);
                        const signature = ethers.utils.joinSignature(flatSignature);
                        return signature as Hex;
                    } catch (error) {
                        console.error("Error signing with state wallet:", error);
                        throw error;
                    }
                },
            };

            const ethereum = (window as any).ethereum;

            const accounts = await ethereum.request({
                method: "eth_requestAccounts",
            });
            const address = accounts[0];

            // Create the wallet client using the ethereum provider
            const walletClient = createWalletClient({
                transport: custom(ethereum),
                chain: polygon,
                account: address as Hex,
            });

            const publicClient = createPublicClient({
                transport: http(),
                chain: polygon,
            });


            // Create the actual Nitrolite configuration
            const nitroConfig: NitroliteClientConfig = {
                publicClient,
                walletClient,
                // @ts-ignore
                stateWalletClient,
                addresses: CONTRACT_ADDRESSES,
                chainId: polygon.id,
                challengeDuration: BigInt(86400), // 1 day in seconds
            };

            // Initialize the ClearNetService
            console.log("Initializing ClearNetService with config:", nitroConfig);
            const initialized = await clearNetService.initialize(nitroConfig);
            console.log("ClearNetService initialized:", initialized);
            if (!initialized) {
                throw new Error("Failed to initialize ClearNetService");
            }

            console.log("ClearNetService initialized successfully");
        } catch (error) {
            console.error("Error initializing ClearNetService:", error);
            walletError.value = "Failed to initialize payment channel services";
            emit("error", walletError.value);
        }

        // Emit success event
        emit("wallet-connected", {
            address: signer.address,
            balance: balance.value,
        });
    } catch (error) {
        console.error("Failed to connect:", error);
        walletError.value = error instanceof Error ? error.message : "Unknown error";
        emit("error", walletError.value);
    } finally {
        isConnecting.value = false;
    }
}

/**
 * Disconnects from the broker
 */
function disconnectWallet() {
    // Reset all state variables
    isConnected.value = false;
    isAuthenticated.value = false;
    walletAddress.value = "";
    balance.value = BigInt(0);

    // Private key is preserved in localStorage
    // to maintain identity consistency across sessions

    // Emit disconnected event
    emit("wallet-disconnected");
}

/**
 * Formats an address for display (e.g., 0x1234...5678)
 */
function formatAddress(address: string): string {
    if (!address) return "";
    return `${address.substring(0, 6)}...${address.substring(address.length - 4)}`;
}

/**
 * Formats a balance in wei to ETH with 4 decimal places
 */
function formatBalance(balanceWei: bigint): string {
    return (Number(balanceWei) / 1e18).toFixed(4);
}
</script>

<template>
    <div class="wallet-connect">
        <div v-if="!isConnected" class="connect-container">
            <button class="connect-btn" :disabled="isConnecting">
                {{ isConnecting ? "Connecting..." : "Connect to Broker" }}
            </button>
            <div v-if="walletError" class="error-message">{{ walletError }}</div>
        </div>

        <div v-else class="wallet-info">
            <div class="address">
                <span class="address-label">Wallet Address:</span>
                <span class="address-value">{{ formatAddress(walletAddress) }}</span>
            </div>

            <div class="connection-status">
                <span class="status-label">Status:</span>
                <span class="status-value" :class="{ 'status-authenticated': isAuthenticated }">
                    {{ isAuthenticated ? "Authenticated" : "Connected" }}
                </span>
            </div>

            <div class="balance">
                <span class="balance-label">Balance:</span>
                <span class="balance-value">{{ formatBalance(balance) }} ETH</span>
            </div>

            <button @click="disconnectWallet" class="disconnect-btn">Disconnect</button>
        </div>
    </div>
</template>

<style scoped>
.wallet-connect {
    padding: 15px;
    border-radius: 8px;
    background-color: #f8f9fa;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
    margin-bottom: 20px;
    width: 100%;
    max-width: 500px;
}

.connect-container {
    display: flex;
    flex-direction: column;
    align-items: center;
}

.connect-btn {
    background-color: #4caf50;
    color: white;
    border: none;
    padding: 10px 20px;
    border-radius: 4px;
    font-weight: 600;
    cursor: pointer;
    transition: background-color 0.2s;
    width: 100%;
    max-width: 250px;
}

.connect-btn:hover {
    background-color: #388e3c;
}

.connect-btn:disabled {
    background-color: #9e9e9e;
    cursor: not-allowed;
}

.wallet-info {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.address,
.balance,
.connection-status {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 0;
}

.address-label,
.balance-label,
.status-label {
    font-weight: 600;
    color: #666;
}

.address-value,
.balance-value,
.status-value {
    font-family: monospace;
    background-color: #e9ecef;
    padding: 4px 8px;
    border-radius: 4px;
}

.status-authenticated {
    background-color: #d4edda;
    color: #155724;
}

.disconnect-btn {
    background-color: #f44336;
    color: white;
    border: none;
    padding: 8px 16px;
    border-radius: 4px;
    font-weight: 600;
    cursor: pointer;
    transition: background-color 0.2s;
    margin-top: 10px;
    align-self: flex-end;
}

.disconnect-btn:hover {
    background-color: #d32f2f;
}

.error-message {
    color: #f44336;
    margin-top: 10px;
    text-align: center;
}
</style>
