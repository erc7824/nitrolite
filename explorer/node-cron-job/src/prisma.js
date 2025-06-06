const { PrismaClient } = require('@prisma/client');
const prisma = new PrismaClient();

async function upsertLedgerEntry(data) {
    console.log('Upserting ledger entry:', data);
    try {
        // Ensure accountId is always provided with a default if missing
        const create = {
            id: Number(data.create.id),
            accountId: data.create.accountId || "default", // Default value if undefined
            accountType: data.create.accountType,
            asset: data.create.asset,
            participant: data.create.participant,
            credit: data.create.credit,
            debit: data.create.debit,
            createdAt: data.create.createdAt || new Date(), // Default to current date
        };

        const update = {
            accountId: data.update.accountId || "default", // Default value if undefined
            accountType: data.update.accountType,
            asset: data.update.asset,
            participant: data.update.participant,
            credit: data.update.credit,
            debit: data.update.debit,
            createdAt: data.update.createdAt,
        };

        return await prisma.ledgerEntry.upsert({
            where: { id: Number(data.where.id) },
            create,
            update,
        });
    } catch (error) {
        console.error('Error in upsertLedgerEntry:', error);
        throw error;
    }
}

async function upsertChannel(inputData) {
    try {
        // Handle both array and single object inputs
        const dataArray = Array.isArray(inputData) ? inputData : [inputData];
        const results = [];
        
        for (const data of dataArray) {
            // Check if channel_id or channelId exists
            const channelId = data.where?.channelId || data.channel_id || data.where?.channel_id || data.channelId;
            
            if (!channelId) {
                console.error('Missing channel_id in data:', JSON.stringify(data, (key, value) => 
                    typeof value === 'bigint' ? value.toString() : value));
                throw new Error('Missing required field: channel_id or channelId');
            }

            // Transform snake_case to camelCase if needed
            const transformedData = {
                where: {
                    channelId: channelId
                },
                create: {
                    channelId: channelId,
                    adjudicator: data.create?.adjudicator || data.adjudicator || "",
                    amount: BigInt(data.create?.amount || data.amount || 0),
                    chainId: data.create?.chainId || data.chain_id || "",
                    challenge: data.create?.challenge || data.challenge || "",
                    createdAt: data.create?.createdAt || new Date(data.created_at) || new Date(),
                    nonce: BigInt(data.create?.nonce || data.nonce || 0),
                    participant: data.create?.participant || data.participant || "",
                    status: data.create?.status || data.status || "",
                    token: data.create?.token || data.token || "",
                    updatedAt: data.create?.updatedAt || new Date(data.updated_at) || new Date(),
                    version: data.create?.version || data.version || 0,
                    wallet: data.create?.wallet || data.wallet || "",
                },
                update: {}
            };
            
            // Safely handle the update data (from either nested update or top-level)
            const sourceUpdateData = data.update || data;
            for (const [key, value] of Object.entries(sourceUpdateData)) {
                if (value !== null && value !== undefined) {
                    // Convert snake_case to camelCase for specific fields
                    if (key === 'channel_id') transformedData.update.channelId = value;
                    else if (key === 'chain_id') transformedData.update.chainId = value;
                    else if (key === 'created_at') transformedData.update.createdAt = new Date(value);
                    else if (key === 'updated_at') transformedData.update.updatedAt = new Date(value);
                    else transformedData.update[key] = value;
                }
            }
            
            console.log('Processing channel:', channelId);
            
            const result = await prisma.channel.upsert({
                where: { channelId: transformedData.where.channelId },
                create: transformedData.create,
                update: transformedData.update,
            });
            
            results.push(result);
        }
        
        // Return single result or array based on input type
        return Array.isArray(inputData) ? results : results[0];
    } catch (error) {
        console.error('Error in upsertChannel:', error);
        throw error;
    }
}

// Add a convenience function specifically for handling arrays
async function upsertChannels(channelsArray) {
    if (!Array.isArray(channelsArray)) {
        throw new Error('upsertChannels expects an array input');
    }
    return await upsertChannel(channelsArray);
}

module.exports = { 
    prisma, 
    upsertLedgerEntry, 
    upsertChannel,
    upsertChannels 
};