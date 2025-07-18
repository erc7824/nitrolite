export const isDevelopment = process.env.NODE_ENV === 'development';
export const isProduction = process.env.NODE_ENV === 'production';

export const config = {
  isDev: isDevelopment,
  isProd: isProduction,
  yellowWsUrl: process.env.NEXT_PUBLIC_YELLOW_WS_URL || 'wss://clearnet.yellow.com/ws',
  asset: process.env.NEXT_PUBLIC_ASSET || 'usdc',
  vApp: {
    name: process.env.NEXT_PUBLIC_VAPP_NAME || '',
    scope: process.env.NEXT_PUBLIC_VAPP_SCOPE || '',
  },
} as const;
