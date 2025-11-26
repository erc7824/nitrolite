const { PrismaClient } = require('@prisma/client');
const prisma = new PrismaClient();

async function checkAndAddRecord(model, data) {
    const existingRecord = await prisma[model].findUnique({
        where: {
            id: data.id, // Assuming 'id' is the unique identifier
        },
    });

    if (!existingRecord) {
        await prisma[model].create({
            data: data,
        });
    }
}

module.exports = {
    checkAndAddRecord,
};