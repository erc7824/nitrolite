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

const ITEMS_PER_PAGE = 10;
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
    const whereConditions: Prisma.ChannelWhereInput = searchTerm ? {
      OR: [
        // Exact match for channel ID
        { channelId: searchTerm },
        // Case-insensitive contains for channel ID
        { channelId: { contains: searchTerm, mode: Prisma.QueryMode.insensitive } },
        // Case-insensitive contains for participant
        { participant: { contains: searchTerm, mode: Prisma.QueryMode.insensitive } },
        // Exact match for participant (for full address searches)
        { participant: searchTerm },
        // Case-insensitive contains for wallet
        { wallet: { contains: searchTerm, mode: Prisma.QueryMode.insensitive } },
        // Exact match for wallet (for full address searches)
        { wallet: searchTerm },
      ],
    } : {};

    try {
      const [ channels, total ] = await Promise.all([
        prisma.channel.findMany({
          where: whereConditions,
          skip,
          take: ITEMS_PER_PAGE,
          orderBy: [
            // Show exact matches first
            {
              channelId: searchTerm ? 'asc' : 'desc',
            },
            // Then sort by creation date
            {
              createdAt: 'desc',
            },
          ],
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
        }),
        prisma.channel.count({
          where: whereConditions,
        }),
      ]);

      // Convert BigInt values to strings and format dates
      const serializedChannels = channels.map((channel) => ({
        ...channel,
        amount: channel.amount.toString(),
        nonce: channel.nonce.toString(),
        createdAt: channel.createdAt.toISOString(),
        updatedAt: channel.updatedAt.toISOString(),
      }));

      // If this is a search request, return in the format expected by the search bar
      if (searchTerm) {
        const searchResults = serializedChannels.map(channel => ({
          type: 'channel',
          data: {
            channelId: channel.channelId,
            participant: channel.participant,
            wallet: channel.wallet,
            amount: channel.amount,
            status: channel.status,
            adjudicator: channel.adjudicator,
            chainId: channel.chainId,
            challenge: channel.challenge,
            nonce: channel.nonce,
            token: channel.token,
            version: channel.version,
            createdAt: channel.createdAt,
            updatedAt: channel.updatedAt,
          },
        }));

        return res.status(200).json(searchResults);
      }

      // Otherwise return the regular paginated response
      return res.status(200).json({
        channels: serializedChannels,
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
    console.error('Error in channels API:', error);
    return res.status(500).json({ error: 'Internal server error' });
  }
}
