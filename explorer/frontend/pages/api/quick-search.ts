import { PrismaClient } from '@prisma/client';
import type { NextApiRequest, NextApiResponse } from 'next';

const globalForPrisma = global as unknown as { prisma: PrismaClient };

export const prisma =
  globalForPrisma.prisma ||
  new PrismaClient();

if (process.env.NODE_ENV !== 'production') {
  globalForPrisma.prisma = prisma;
}

const SEARCH_LIMIT = 5;

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method !== 'GET') {
    return res.status(405).json({ message: 'Method not allowed' });
  }

  try {
    const searchTerm = req.query.q as string;

    if (!searchTerm || searchTerm.trim().length === 0) {
      return res.status(200).json([]);
    }

    const channels = await prisma.channel.findMany({
      where: {
        OR: [
          { channelId: { contains: searchTerm } },
          { participant: { contains: searchTerm } },
          { wallet: { contains: searchTerm } },
        ],
      },
      take: SEARCH_LIMIT,
      orderBy: {
        createdAt: 'desc',
      },
      select: {
        channelId: true,
        participant: true,
        wallet: true,
        amount: true,
        status: true,
      },
    });

    // Convert BigInt values to strings
    const searchResults = channels.map(channel => ({
      type: 'channel',
      data: {
        ...channel,
        amount: channel.amount.toString(),
      },
    }));

    res.status(200).json(searchResults);
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error('Error searching channels:', error);
    res.status(500).json({ error: 'Failed to search channels' });
  }
} 