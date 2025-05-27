<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import clearNetService from '../services/ClearNetService';
import { ethers } from 'ethers';
import { CONTRACT_ADDRESSES } from '../config';
import { waitForTransaction } from '@erc7824/nitrolite';

// ERC20 ABI for decimals
const ERC20_ABI = [
    "function decimals() view returns (uint8)"
];

const props = defineProps<{
    isWalletConnected: boolean;
    roomId: string;
    roomCreator: boolean;
}>();

// Metamask state
const hasMetamask = ref(false);
const metamaskInfo = ref<{
    address: string;
    balance: any;
    provider: ethers.providers.Web3Provider;
    signer: ethers.providers.JsonRpcSigner;
    chainId: number;
} | null>(null);

const emit = defineEmits(['channel-created', 'error']);

const depositAmount = ref<string>('0.00001');
const errorMessage = ref('');
const isCreating = ref(false);
const showAdvanced = ref(false);
const channelData = ref<ChannelResponse | null>(null);
const tokenDecimals = ref<number>(18); // will be updated with value from contract

interface ChannelResponse {
    channelId: string;
    state?: any;
    initialState?: {
        allocations: Array<{
            amount: bigint;
            destination: string;
            token: string;
        }>;
        data: string;
        intent: number;
        sigs: Array<{
            r: string;
            s: string;
            v: number;
        }>;
        version: bigint;
    };
    txHash?: string;
}

// Convert ETH to Wei using contract decimals
const depositAmountWei = computed(() => {
    const amountFloat = parseFloat(depositAmount.value);
    if (isNaN(amountFloat)) return 0n;

    const decimalMultiplier = Math.pow(10, tokenDecimals.value);
    return BigInt(Math.floor(amountFloat * decimalMultiplier));
});

// Validate that the amount is a valid number
const isValidAmount = computed(() => {
    const amount = parseFloat(depositAmount.value);
    return !isNaN(amount) && amount > 0;
});

// Initialize Metamask connection and get balance
async function checkMetamaskBalance() {
    try {
        // Check if Metamask is available
        const { ethereum } = window as any;
        if (!ethereum) {
            console.log('Metamask not detected');
            hasMetamask.value = false;
            return null;
        }

        hasMetamask.value = true;

        // Request accounts from Metamask
        const accounts = await ethereum.request({ method: 'eth_requestAccounts' });

        if (!accounts || accounts.length === 0) {
            console.log('No accounts found');
            metamaskInfo.value = null;
            return null;
        }

        // Create a provider - ethers v5 provider
        const provider = new ethers.providers.Web3Provider(ethereum);

        // Get network information
        const network = await provider.getNetwork();
        console.log(`Connected to network: ${network.name} (${network.chainId})`);

        // Get token decimals from contract
        const tokenContract = new ethers.Contract(CONTRACT_ADDRESSES.tokenAddress, ERC20_ABI, provider);
        try {
            const decimals = await tokenContract.decimals();
            tokenDecimals.value = decimals;
            console.log(`Token decimals: ${decimals}`);
        } catch (error) {
            console.error('Failed to get token decimals:', error);
            // Keep default decimals if contract call fails
        }

        // Get balance in wei
        const balanceWei = await provider.getBalance(accounts[0]);

        // Get signer for transactions
        const signer = provider.getSigner();

        const info = {
            address: accounts[0],
            balance: balanceWei,
            provider,
            signer,
            chainId: network.chainId
        };

        metamaskInfo.value = info;

        // Set up listeners for account and network changes
        ethereum.on('accountsChanged', async (newAccounts: string[]) => {
            console.log('Metamask accounts changed:', newAccounts);
            if (newAccounts.length === 0) {
                metamaskInfo.value = null;
            } else {
                try {
                    const newSigner = provider.getSigner();
                    const newBalance = await provider.getBalance(newAccounts[0]);
                    const network = await provider.getNetwork();

                    metamaskInfo.value = {
                        address: newAccounts[0],
                        balance: newBalance,
                        provider,
                        signer: newSigner,
                        chainId: network.chainId
                    };
                } catch (err) {
                    console.error('Error updating account info:', err);
                    metamaskInfo.value = null;
                }
            }
        });

        ethereum.on('chainChanged', async () => {
            console.log('Metamask network changed, refreshing provider');
            try {
                // Need to refresh provider on chain change
                const updatedProvider = new ethers.providers.Web3Provider(ethereum);
                const network = await updatedProvider.getNetwork();
                const updatedSigner = updatedProvider.getSigner();
                const address = await updatedSigner.getAddress();
                const newBalance = await updatedProvider.getBalance(address);

                metamaskInfo.value = {
                    address,
                    balance: newBalance,
                    provider: updatedProvider,
                    signer: updatedSigner,
                    chainId: network.chainId
                };
            } catch (err) {
                console.error('Error updating network info:', err);
            }
        });

        return info;
    } catch (error) {
        console.error('Error checking Metamask balance:', error);
        metamaskInfo.value = null;
        return null;
    }
}

// Create a new channel
async function createChannel() {
    if (!props.isWalletConnected) {
        errorMessage.value = 'Please connect your wallet first';
        emit('error', errorMessage.value);
        return;
    }

    if (!isValidAmount.value) {
        errorMessage.value = 'Please enter a valid deposit amount';
        emit('error', errorMessage.value);
        return;
    }

    isCreating.value = true;
    errorMessage.value = '';

    try {
        // Check Metamask balance first
        const info = await checkMetamaskBalance();
        if (info) {
            // Convert deposit amount to BigNumber for comparison
            const depositAmountBN = ethers.utils.parseEther(depositAmount.value.toString());

            // Check if user has enough balance
            if (info.balance.lt(depositAmountBN)) {
                errorMessage.value = `Insufficient balance in Metamask. You have ${ethers.utils.formatEther(info.balance)} ETH`;
                emit('error', errorMessage.value);
                return;
            }

            console.log(`Metamask address: ${info.address}`);
            console.log(`Metamask balance: ${ethers.utils.formatEther(info.balance)} ETH`);
        } else {
            console.log('Could not check Metamask balance, proceeding anyway');
        }

        try {
            console.log("Start Create Channel");
            let nitroChannelId = localStorage.getItem("nitro_channel_id");

            if (!nitroChannelId) {
                // Fetch available channels
                const channels = await clearNetService.getAccountChannels();
                console.log("Channels:", channels);
                if (channels.length > 0) {
                    nitroChannelId = channels[0];
                    localStorage.setItem("nitro_channel_id", nitroChannelId);
                    console.log('Fetched channel id from RPC:', nitroChannelId, channels);
                }
            }
            if (!nitroChannelId) {
                // Create the channel using the proper Nitrolite client
                const depositAmount = depositAmountWei.value;
                console.log("Start Deposit", depositAmount);
                const depositResponse = await clearNetService.client.deposit(depositAmount);
                console.log("Deposit response:", depositResponse);
                await waitForTransaction(clearNetService.client.publicClient, depositResponse);
                const createChannelResponse = await clearNetService.client.createChannel({
                    initialAllocationAmounts: [depositAmount, BigInt(0)],
                    stateData: "0x",
                });
                console.log("Create channel response:", createChannelResponse);
                if (createChannelResponse && createChannelResponse.channelId) {
                    nitroChannelId = createChannelResponse.channelId;
                }
            }

            // If we have a channel id, we can create the channel
            if (nitroChannelId) {
                localStorage.setItem("nitro_channel_id", nitroChannelId);
                const newChannelData = { channelId: nitroChannelId };
                channelData.value = newChannelData;
                emit('channel-created', newChannelData);
            } else {
                errorMessage.value = 'Failed to create channel';
                emit('error', errorMessage.value);
            }
        } catch (contractError) {
            console.error('Contract error during channel creation:', contractError);
            let errorMsg = 'Error creating channel';

            if (String(contractError).includes('Invalid address')) {
                errorMsg = 'Contract address configuration error. Please check network settings.';
            } else if (String(contractError).includes('user rejected')) {
                errorMsg = 'Transaction was rejected by user.';
            } else {
                errorMsg = String(contractError);
            }

            errorMessage.value = errorMsg;
            emit('error', errorMessage.value);
        }
    } catch (error) {
        console.error('Error creating channel:', error);
        errorMessage.value = 'Error creating channel: ' + (error instanceof Error ? error.message : String(error));
        emit('error', errorMessage.value);
    } finally {
        isCreating.value = false;
    }
}

// Toggle advanced options
function toggleAdvanced() {
    showAdvanced.value = !showAdvanced.value;
}

// Format Ethereum balance
function formatEtherBalance(balance: any): string {
    if (!balance) return '0';
    try {
        return ethers.utils.formatEther(balance);
    } catch (error) {
        console.error('Error formatting balance:', error);
        return '0';
    }
}

// Format Ethereum address
function formatAddress(address: string): string {
    if (!address) return 'Not connected';
    try {
        return `${address.substring(0, 6)}...${address.substring(address.length - 4)}`;
    } catch (error) {
        console.error('Error formatting address:', error);
        return address || 'Invalid address';
    }
}

// Get friendly network name from chain ID
function getNetworkName(chainId: number): string {
    const networks: Record<number, string> = {
        1: 'Ethereum Mainnet',
        5: 'Goerli Testnet',
        11155111: 'Sepolia Testnet',
        137: 'Polygon Mainnet',
        80001: 'Mumbai Testnet',
        42220: 'Celo Mainnet',
        44787: 'Celo Alfajores Testnet',
        1337: 'Local Development'
    };

    return networks[chainId] || `Unknown Network (${chainId})`;
}

// Check for Metamask on component mount
onMounted(async () => {
    // Check if Metamask is available
    const { ethereum } = window as any;
    hasMetamask.value = !!ethereum;

    if (hasMetamask.value) {
        // Check if already connected - don't prompt for connection yet
        try {
            const accounts = await ethereum.request({ method: 'eth_accounts' });
            if (accounts && accounts.length > 0) {
                // Create a provider
                const provider = new ethers.providers.Web3Provider(ethereum);

                // Get network information
                const network = await provider.getNetwork();

                // Get balance
                const balanceWei = await provider.getBalance(accounts[0]);

                // Get signer for transactions
                const signer = provider.getSigner();

                metamaskInfo.value = {
                    address: accounts[0],
                    balance: balanceWei,
                    provider,
                    signer,
                    chainId: network.chainId
                };
            }
        } catch (error) {
            console.error('Error checking Metamask accounts:', error);
        }
    }
});
</script>

<template>
    <div class="channel-setup">
        <h3>Create Channel</h3>

        <div class="metamask-status" v-if="hasMetamask">
            <div class="metamask-info">
                <img src="https://upload.wikimedia.org/wikipedia/commons/3/36/MetaMask_Fox.svg" alt="Metamask"
                    class="metamask-icon" />
                <div v-if="metamaskInfo" class="metamask-details">
                    <div class="metamask-balance">Balance: {{ formatEtherBalance(metamaskInfo.balance) }} ETH</div>
                    <div class="metamask-address">{{ formatAddress(metamaskInfo.address) }}</div>
                    <div class="metamask-network" :class="{ 'network-testnet': metamaskInfo.chainId !== 1 }">
                        Network: {{ getNetworkName(metamaskInfo.chainId) }}
                    </div>
                </div>
                <div v-else>
                    <div class="metamask-balance">Not connected</div>
                    <button @click="checkMetamaskBalance" class="connect-metamask-btn" :disabled="isCreating">
                        Connect Metamask
                    </button>
                </div>
            </div>
        </div>

        <div class="form-group">
            <label for="depositAmount">Deposit Amount (ETH):</label>
            <input id="depositAmount" v-model="depositAmount" type="number" step="0.00001" min="0.00001"
                :disabled="isCreating" />
            <small>This amount will be used to fund your game channel.</small>
        </div>

        <div class="actions">
            <button @click="createChannel" class="create-btn"
                :disabled="!isWalletConnected || isCreating || !isValidAmount">
                {{ isCreating ? 'Creating...' : 'Create & Fund Channel' }}
            </button>
        </div>

        <div class="advanced-toggle" @click="toggleAdvanced">
            {{ showAdvanced ? 'Hide' : 'Show' }} Advanced Options
        </div>

        <div v-if="showAdvanced" class="advanced-options">
            <div class="form-group">
                <label for="roomId">Room ID:</label>
                <input id="roomId" type="text" :value="roomId" disabled />
            </div>

            <div class="form-group">
                <label for="role">Your Role:</label>
                <input id="role" type="text" value="Room Creator" disabled />
            </div>
        </div>

        <div v-if="errorMessage" class="error-message">
            {{ errorMessage }}
        </div>

        <div v-if="channelData" class="channel-info">
            <div class="info-item">
                <span class="label">Channel ID:</span>
                <span class="value">{{ channelData.channelId }}</span>
            </div>
            <div class="info-item">
                <span class="label">Status:</span>
                <span class="value success">Created</span>
            </div>
        </div>
    </div>
</template>

<style scoped>
.channel-setup {
    background-color: #f8f9fa;
    border-radius: 8px;
    padding: 20px;
    margin-bottom: 20px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
    width: 100%;
    max-width: 500px;
}

h3 {
    margin-top: 0;
    margin-bottom: 20px;
    color: #333;
}

.form-group {
    margin-bottom: 15px;
}

label {
    display: block;
    margin-bottom: 6px;
    font-weight: 600;
    color: #555;
}

input {
    width: 100%;
    padding: 10px 12px;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 16px;
}

input:focus {
    border-color: #4CAF50;
    outline: none;
}

small {
    display: block;
    color: #888;
    margin-top: 4px;
    font-size: 0.85em;
}

.actions {
    margin-top: 20px;
}

.create-btn {
    width: 100%;
    padding: 12px;
    border: none;
    border-radius: 4px;
    font-size: 16px;
    font-weight: 600;
    cursor: pointer;
    transition: background-color 0.2s;
    background-color: #4CAF50;
    color: white;
}

.create-btn:hover:not(:disabled) {
    background-color: #388E3C;
}

.create-btn:disabled {
    background-color: #9e9e9e;
    cursor: not-allowed;
}

.advanced-toggle {
    text-align: center;
    margin-top: 15px;
    color: #2196F3;
    cursor: pointer;
    font-size: 0.9em;
}

.advanced-toggle:hover {
    text-decoration: underline;
}

.advanced-options {
    margin-top: 15px;
    padding-top: 15px;
    border-top: 1px solid #eee;
}

.error-message {
    background-color: #ffebee;
    color: #c62828;
    padding: 10px;
    border-radius: 4px;
    margin-top: 20px;
    text-align: center;
}

.channel-info {
    margin-top: 20px;
    padding: 15px;
    background-color: #e8f5e9;
    border-radius: 4px;
}

.info-item {
    display: flex;
    justify-content: space-between;
    margin-bottom: 8px;
}

.info-item:last-child {
    margin-bottom: 0;
}

.label {
    font-weight: 600;
    color: #555;
}

.value {
    font-family: monospace;
}

.success {
    color: #4CAF50;
    font-weight: bold;
}

/* Metamask styles */
.metamask-status {
    margin-bottom: 20px;
    padding: 12px 15px;
    border-radius: 8px;
    background-color: #fffbf5;
    border: 1px solid #f5a623;
}

.metamask-info {
    display: flex;
    align-items: center;
    gap: 12px;
}

.metamask-icon {
    width: 30px;
    height: 30px;
}

.metamask-balance {
    font-weight: 600;
    margin-bottom: 4px;
}

.metamask-address {
    font-family: monospace;
    color: #666;
    font-size: 0.9em;
}

.metamask-network {
    font-size: 0.85em;
    color: #1976D2;
    margin-top: 3px;
}

.network-testnet {
    color: #f57c00;
}

.metamask-details {
    display: flex;
    flex-direction: column;
    gap: 2px;
}

.connect-metamask-btn {
    background-color: #f5a623;
    color: white;
    border: none;
    padding: 8px 12px;
    border-radius: 4px;
    font-weight: 600;
    cursor: pointer;
    margin-top: 6px;
    transition: background-color 0.2s;
}

.connect-metamask-btn:hover:not(:disabled) {
    background-color: #e09216;
}

.connect-metamask-btn:disabled {
    background-color: #ccc;
    cursor: not-allowed;
}
</style>
