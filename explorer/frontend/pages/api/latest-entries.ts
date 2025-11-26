import { PrismaClient } from '@prisma/client';
import type { LedgerEntry } from '@prisma/client';
import type { NextApiRequest, NextApiResponse } from 'next';

const prisma = new PrismaClient();

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse<Array<LedgerEntry> | { error: string }>,
) {
  if (req.method === 'GET') {
    try {
      const latestEntries = await prisma.ledgerEntry.findMany({
        orderBy: {
          id: 'desc',
        },
        take: 5,
      });
      res.status(200).json(latestEntries);
    } catch (error) {
      res.status(500).json({ error: 'Failed to fetch latest ledger entries' });
    }
  } else {
    res.setHeader('Allow', [ 'GET' ]);
    res.status(405).end(`Method ${ req.method } Not Allowed`);
  }
}
