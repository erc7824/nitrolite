import { PrismaClient } from '@prisma/client';
import type { NextApiRequest, NextApiResponse } from 'next';

// PrismaClient is attached to the `global` object in development to prevent
// exhausting your database connection limit.
const globalForPrisma = global as unknown as { prisma: PrismaClient };

export const prisma =
  globalForPrisma.prisma ||
  new PrismaClient();

if (process.env.NODE_ENV !== 'production') {
  globalForPrisma.prisma = prisma;
}

const LATEST_ENTRIES_LIMIT = 5;

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method !== 'GET') {
    return res.status(405).json({ message: 'Method not allowed' });
  }

  try {
    const entries = await prisma.ledgerEntry.findMany({
      take: LATEST_ENTRIES_LIMIT,
      orderBy: {
        createdAt: 'desc',
      },
    });

    // Convert Decimal values to strings for JSON serialization
    const serializedEntries = entries.map((entry) => ({
      ...entry,
      credit: entry.credit.toString(),
      debit: entry.debit.toString(),
      createdAt: entry.createdAt.toISOString(),
    }));

    res.status(200).json(serializedEntries);
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error('Error fetching latest ledger entries:', error);
    res.status(500).json({ error: 'Failed to fetch latest ledger entries' });
  }
} 