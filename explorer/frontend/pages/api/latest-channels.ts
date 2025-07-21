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

const LATEST_CHANNELS_LIMIT = 5;

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method !== 'GET') {
    return res.status(405).json({ message: 'Method not allowed' });
  }

  try {
    const channels = await prisma.channel.findMany({
      take: LATEST_CHANNELS_LIMIT,
      orderBy: {
        createdAt: 'desc',
      },
      select: {
        channelId: true,
        adjudicator: true,
        amount: true,
        chainId: true,
        challenge: true,
        createdAt: true,
        nonce: true,
        participant: true,
        status: true,
        token: true,
        updatedAt: true,
        version: true,
        wallet: true,
      },
    });

    // Convert BigInt values to strings and format dates
    const serializedChannels = channels.map((channel) => ({
      ...channel,
      amount: channel.amount.toString(),
      nonce: channel.nonce.toString(),
      createdAt: channel.createdAt.toISOString(),
      updatedAt: channel.updatedAt.toISOString(),
    }));

    res.status(200).json(serializedChannels);
  } catch (error) {
    // eslint-disable-next-line no-console
    console.error('Error fetching latest channels:', error);
    res.status(500).json({ error: 'Failed to fetch latest channels' });
  }
}
