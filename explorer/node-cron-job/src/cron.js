const cron = require('node-cron');
const WebSocket = require('ws');
const { upsertLedgerEntry, upsertChannel } = require('./prisma');

//console.log('upsertLedgerEntry:', upsertLedgerEntry);
//console.log('upsertChannel:', upsertChannel);

const WSS_URL = 'wss://clearnet.yellow.com/ws';

let ws;

// Function to initialize WebSocket connection
function initializeWebSocket() {
    ws = new WebSocket(WSS_URL);

    ws.on('open', () => {
        console.log('Connected to WSS');
    });

    ws.on('message', (data) => {
        try {
            const message = JSON.parse(data.toString());
            console.log('Message received from WSS:', message);
        } catch (error) {
            console.error('Error parsing message from WSS:', error);
        }
    });

    ws.on('error', (error) => {
        console.error('WebSocket error:', error);
    });

    ws.on('close', () => {
        console.log('WebSocket connection closed. Reconnecting...');
        setTimeout(initializeWebSocket, 5000); // Reconnect after 5 seconds
    });
}

// Function to send a message through WebSocket
function sendMessage(message, callback) {
    if (ws && ws.readyState === WebSocket.OPEN) {
        console.log('Sending message:', message);
        ws.send(JSON.stringify(message));

        ws.once('message', (data) => {
            //console.log('Response received for message:', data.toString());
            try {
                const parsedData = JSON.parse(data.toString());
                console.log('Parsed response:', parsedData);
                callback(parsedData);
            } catch (error) {
                console.error('Error parsing response from WSS:', error);
            }
        });
    } else {
        console.error('WebSocket is not connected. Message not sent.');
    }
}

function processLedgerEntries(data) {
    if (data && data.res && Array.isArray(data.res[2]) && Array.isArray(data.res[2][0])) {
        const ledgerEntries = data.res[2][0];
        ledgerEntries.forEach(async (entry) => {
            console.log('Ledger entry being upserted:', entry);

            // Ensure the entry has a valid `id` field
            if (!entry.id) {
                console.error('Ledger entry is missing a unique identifier:', entry);
                return; // Skip this entry
            }

            // Convert id to the correct type (number)
            const entryId = Number(entry.id);
            if (isNaN(entryId)) {
                console.error('Invalid id format for ledger entry:', entry);
                return;
            }

            try {
                // Format the data to match your Prisma model
                const processedEntry = {
                    id: entryId,
                    accountId: entry.account_id || "default", // Provide default value if missing
                    accountType: entry.account_type,
                    asset: entry.asset,
                    participant: entry.participant,
                    credit: entry.credit,
                    debit: entry.debit,
                    createdAt: entry.created_at ? new Date(entry.created_at) : new Date(),
                };

                console.log('Processed ledger entry:', processedEntry);

                const result = await upsertLedgerEntry({
                    where: { id: entryId },
                    create: processedEntry,
                    update: processedEntry,
                });

                console.log('Upsert result:', result);
            } catch (error) {
                console.error('Error upserting ledger entry:', error);
                console.error('Error details:', error.message);
                // Print full error stack for debugging
                console.error(error.stack);
            }
        });
    } else {
        console.error('Unexpected ledger entries structure:', JSON.stringify(data, null, 2));
    }
}

async function processChannels(data) {
    console.log('Raw channels data received:', JSON.stringify(data, null, 2));
    
    // Initialize channels array
    let channels = [];
    
    // Handle various possible data structures
    try {
        if (data && data.res) {
            console.log('Processing data.res structure...');
            console.log('data.res length:', data.res.length);
            console.log('data.res[2] type:', typeof data.res[2]);
            console.log('data.res[2] sample:', Array.isArray(data.res[2]) ? data.res[2].slice(0, 1) : data.res[2]);
            
            if (Array.isArray(data.res[2])) {
                // Check if data.res[2] contains another array (nested structure)
                if (Array.isArray(data.res[2][0])) {
                    channels = data.res[2][0];
                    console.log('Found channels in data.res[2][0], length:', channels.length);
                } else {
                    // data.res[2] is directly the channels array
                    channels = data.res[2];
                    console.log('Found channels in data.res[2], length:', channels.length);
                }
            } else if (typeof data.res[2] === 'string') {
                // Handle string that might be a stringified JSON array
                try {
                    const parsed = JSON.parse(data.res[2]);
                    if (Array.isArray(parsed)) {
                        channels = parsed;
                        console.log('Parsed channels from string, length:', channels.length);
                    }
                } catch (jsonError) {
                    console.error('Error parsing string data:', jsonError);
                }
            }
        } else if (Array.isArray(data)) {
            // Directly handle array of channels
            channels = data;
            console.log('Data is directly an array, length:', channels.length);
        }
        
        console.log('Number of channels extracted:', channels.length);
        console.log('Sample channel data:', channels[0]);
        
        // Modified validation to properly check for valid channel objects
        console.log('Before validation - channels length:', channels.length);
        console.log('First channel before validation:', channels[0]);
        
        channels = channels.filter((channel, index) => {
            console.log(`Validating channel ${index}:`, typeof channel, channel?.channel_id);
            
            if (!channel || typeof channel !== 'object') {
                console.warn(`Skipping invalid channel ${index} (not an object):`, channel);
                return false;
            }
            
            // Check if the channel has a channel_id property
            if (!channel.hasOwnProperty('channel_id') || !channel.channel_id) {
                console.warn(`Skipping channel ${index} missing channel_id:`, JSON.stringify(channel, null, 2));
                return false;
            }
            
            console.log(`✅ Channel ${index} passed validation:`, channel.channel_id);
            return true;
        });
        
        console.log('Number of valid channels after filtering:', channels.length);
        
        // Process each channel
        for (const [index, channel] of channels.entries()) {
            try {
                console.log(`Processing channel ${index + 1}/${channels.length}:`, channel.channel_id);
                
                // Format the data to match your Prisma model
                const processedChannel = {
                    channelId: channel.channel_id,
                    participant: channel.participant || "",
                    status: channel.status || "",
                    token: channel.token || "",
                    wallet: channel.wallet || "",
                    amount: channel.amount !== undefined ? BigInt(String(channel.amount)) : BigInt(0),
                    chainId: parseInt(channel.chain_id) || 0,  // Convert to integer to match Prisma schema
                    adjudicator: channel.adjudicator || "",
                    challenge: parseInt(channel.challenge) || 0,  // Ensure this is also an integer
                    nonce: channel.nonce ? BigInt(String(channel.nonce)) : BigInt(0),
                    version: parseInt(channel.version) || 0,  // Ensure this is an integer
                    createdAt: channel.created_at ? new Date(channel.created_at) : new Date(),
                    updatedAt: channel.updated_at ? new Date(channel.updated_at) : new Date()
                };

                console.log(`Processed channel data [${index}]:`, processedChannel);

                const result = await upsertChannel({
                    where: { channelId: channel.channel_id },
                    create: processedChannel,
                    update: processedChannel,
                });

                console.log(`✅ Successfully upserted channel ${channel.channel_id}`);
            } catch (error) {
                console.error(`❌ Error upserting channel at index ${index}:`, error);
                console.error('Error details:', error.message);
                console.error('Failed channel data:', JSON.stringify(channel, null, 2));
                console.error('Full error stack:', error.stack);
            }
        }
    } catch (error) {
        console.error('Error processing channels:', error);
        console.error('Original data structure:', data);
    }
}

// Messages for the WebSocket
const ledgerMessage = {
    req: [1, 'get_ledger_entries', [], Date.now()],
    sig: ['0xd2efd06ffa63037547b897a4590db52307e8de45d961df1ab6796e321e37a13e7dc42bf4885d72ce1a2ff52186bc3be25d814a73859b4644d8ea368948249b3d00'],
};

const channelMessage = {
    req: [1, 'get_channels', [], Date.now()],
    sig: ['0x853b49719ccd142296dc3b3f215ec6a3c4d93f3719fc1f62b18fc9031375d4200db3855d1b749f2e74839c2236bc6158776e2564d2942240aad2ed48655c977e00'],
};

// Schedule cron jobs
cron.schedule('*/1 * * * *', () => {
  console.log('Running cron job for ledger entries...');
  sendMessage(ledgerMessage, processLedgerEntries);
});


cron.schedule('*/1 * * * *', () => {
    console.log('Running cron job for channel entries...');
    sendMessage(channelMessage, processChannels);
});

// Initialize WebSocket connection
initializeWebSocket();

// Ping WebSocket every 30 seconds
setInterval(() => {
    console.log('Pinging WebSocket for ledger entries...');
    //sendMessage(ledgerMessage, processLedgerEntries);
}, 30000); // 30 seconds in milliseconds