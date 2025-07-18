export const isDevelopment = import.meta.env.DEV;
export const isProduction = import.meta.env.PROD;

export const config = {
  isDev: isDevelopment,
  isProd: isProduction,
  yellowWsUrl: import.meta.env.VITE_YELLOW_WS_URL || 'wss://clearnet.yellow.com/ws',
  asset: import.meta.env.VITE_ASSET || 'usdc',
  vApp: {
    name: import.meta.env.VITE_VAPP_NAME || '',
    scope: import.meta.env.VITE_VAPP_SCOPE || '',
  },
} as const;
