import 'dotenv/config';

export const isDevelopment = process.env.NODE_ENV === 'development';
export const isProduction = process.env.NODE_ENV === 'production';

export const config = {
  isDev: isDevelopment,
  isProd: isProduction,
  port: parseInt(process.env.PORT || '3001'),
  yellowWsUrl: process.env.YELLOW_WS_URL || 'wss://clearnet.yellow.com/ws',
  asset: process.env.ASSET || 'usdc',
  walletPrivateKey: process.env.WALLET_PRIVATE_KEY || '',
  vApp: {
    name: process.env.VAPP_NAME || '{{projectName}}',
    scope: process.env.VAPP_SCOPE || '{{packageName}}',
  },
} as const;

// Validate required environment variables
if (!config.walletPrivateKey && !isDevelopment) {
  throw new Error('WALLET_PRIVATE_KEY environment variable is required in production');
}

if (config.walletPrivateKey && !config.walletPrivateKey.startsWith('0x')) {
  throw new Error('WALLET_PRIVATE_KEY must start with 0x');
}