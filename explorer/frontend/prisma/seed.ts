import { PrismaClient } from '@prisma/client';

const prisma = new PrismaClient();

interface ChannelDataEntry {
  adjudicator: string;
  amount: string | number;
  chain_id: number;
  challenge: number;
  channel_id: string;
  created_at: string;
  nonce: string | number;
  participant: string;
  status: string;
  token: string;
  updated_at: string;
  version: number;
  wallet: string;
}

// Generate 500 channels with varied data
const generateChannelData = (): Array<ChannelDataEntry> => {
  const adjudicators = [
    '0x6D3B5EFa1f81f65037cD842F48E44BcBCa48CBEF',
    '0x5F4A4B1D293A973a1Bc0daD3BB3692Bd51058FCF',
    '0x4E3C2B1A0D9F8E7C6B5A4D3C2E1F0A9B8C7D6E5',
  ];

  const tokens: Record<number, string> = {
    '137': '0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359', // Polygon
    '11155111': '0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238', // Sepolia
    '1': '0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2', // Ethereum
  };

  const chainIds = [ 1, 137, 11155111 ];
  const statuses = [ 'open', 'closed', 'disputed' ];
  const baseTime = new Date('2025-05-01T00:00:00Z').getTime();
  const channels: Array<ChannelDataEntry> = [];

  for (let i = 0; i < 500; i++) {
    const chainId = chainIds[Math.floor(Math.random() * chainIds.length)];
    const status = statuses[Math.floor(Math.random() * statuses.length)];
    const createdAt = new Date(baseTime + Math.floor(Math.random() * 30 * 24 * 60 * 60 * 1000));
    const updatedAt = new Date(createdAt.getTime() + Math.floor(Math.random() * 24 * 60 * 60 * 1000));

    channels.push({
      adjudicator: adjudicators[Math.floor(Math.random() * adjudicators.length)],
      amount: Math.floor(Math.random() * 10000000) + 100000, // Random amount between 100k and 10M
      chain_id: chainId,
      challenge: 3600,
      channel_id: '0x' + Array.from({ length: 64 }, () => '0123456789abcdef'[Math.floor(Math.random() * 16)]).join(''),
      created_at: createdAt.toISOString(),
      nonce: BigInt(Math.floor(Math.random() * 1e16)).toString(),
      participant: '0x' + Array.from({ length: 40 }, () => '0123456789abcdef'[Math.floor(Math.random() * 16)]).join(''),
      status,
      token: tokens[chainId],
      updated_at: updatedAt.toISOString(),
      version: status === 'open' ? 1 : Math.floor(Math.random() * 3) + 1,
      wallet: '0x' + Array.from({ length: 40 }, () => '0123456789abcdef'[Math.floor(Math.random() * 16)]).join(''),
    });
  }

  return channels;
};

const channelData: Array<ChannelDataEntry> = generateChannelData();

async function main() {
  // Clear both tables before seeding
  await prisma.ledgerEntry.deleteMany();
  await prisma.channel.deleteMany();
  // eslint-disable-next-line no-console
  console.log('Start seeding ...');

  // Seed Channel data
  for (const entry of channelData) {
    try {
      // eslint-disable-next-line no-console
      console.log(`Creating channel with ID: ${ entry.channel_id }`);
      await prisma.channel.create({
        data: {
          channelId: entry.channel_id,
          adjudicator: entry.adjudicator,
          amount: BigInt(entry.amount),
          chainId: entry.chain_id,
          challenge: entry.challenge,
          createdAt: new Date(entry.created_at),
          nonce: BigInt(entry.nonce),
          participant: entry.participant,
          status: entry.status,
          token: entry.token,
          updatedAt: new Date(entry.updated_at),
          version: entry.version,
          wallet: entry.wallet,
        },
      });
    } catch (error) {
      // eslint-disable-next-line no-console
      console.error(`Failed to create channel ${ entry.channel_id }:`, error);
      throw error; // Re-throw to stop the seeding process
    }
  }

  // eslint-disable-next-line no-console
  console.log('Seeding finished.');
}

main()
  .then(async() => {
    await prisma.$disconnect();
  })
  .catch(async(e) => {
  // eslint-disable-next-line no-console
    console.error(e);

    await prisma.$disconnect();
    process.exit(1);
  });
