import { PrismaClient, Prisma } from '@prisma/client';
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

const ITEMS_PER_PAGE = 20;
const MIN_SEARCH_LENGTH = 3;

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method !== 'GET') {
    return res.status(405).json({ message: 'Method not allowed' });
  }

  try {
    const page = parseInt(req.query.page as string) || 1;
    const skip = (page - 1) * ITEMS_PER_PAGE;
    const searchTerm = (req.query.q as string || '').trim();

    // Validate search term length
    if (searchTerm && searchTerm.length < MIN_SEARCH_LENGTH) {
      return res.status(400).json({ error: `Search term must be at least ${ MIN_SEARCH_LENGTH } characters long` });
    }

    // Build search conditions if search term is provided
    const whereConditions: Prisma.LedgerEntryWhereInput = searchTerm ? {
      OR: [
        { accountId: { contains: searchTerm, mode: Prisma.QueryMode.insensitive } },
        { participant: { contains: searchTerm, mode: Prisma.QueryMode.insensitive } },
        { asset: { contains: searchTerm, mode: Prisma.QueryMode.insensitive } },
      ],
    } : {};

    try {
      const [ entries, total ] = await Promise.all([
        prisma.ledgerEntry.findMany({
          where: whereConditions,
          skip,
          take: ITEMS_PER_PAGE,
          orderBy: [
            {
              createdAt: 'desc',
            },
          ],
        }),
        prisma.ledgerEntry.count({
          where: whereConditions,
        }),
      ]);

      // Convert Decimal values to strings for JSON serialization
      const serializedEntries = entries.map((entry) => ({
        ...entry,
        credit: entry.credit.toString(),
        debit: entry.debit.toString(),
        createdAt: entry.createdAt.toISOString(),
      }));

      // If this is a search request, return in the format expected by the search bar
      if (searchTerm) {
        const searchResults = serializedEntries.map(entry => ({
          type: 'ledger_entry',
          data: entry,
        }));

        return res.status(200).json(searchResults);
      }

      // Otherwise return the regular paginated response
      return res.status(200).json({
        entries: serializedEntries,
        total,
        hasMore: skip + ITEMS_PER_PAGE < total,
      });
    } catch (dbError) {
      // eslint-disable-next-line no-console
      console.error('Database error:', dbError);
      return res.status(500).json({ error: 'Database error occurred' });
    }
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error('Error in ledger entries API:', error);
    return res.status(500).json({ error: 'Internal server error' });
  }
} 