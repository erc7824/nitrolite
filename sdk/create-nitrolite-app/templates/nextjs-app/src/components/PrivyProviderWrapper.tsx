'use client';

import { PrivyProvider } from '@privy-io/react-auth';
import { config } from '@/utils/env';

interface PrivyProviderWrapperProps {
  children: React.ReactNode;
}

export function PrivyProviderWrapper({ children }: PrivyProviderWrapperProps) {
  const privyAppId = config.privyAppId;

  if (!privyAppId) {
    return (
      <div className="p-4 bg-red-50 border border-red-200 rounded-lg m-4">
        <h2 className="text-red-800 font-bold mb-2">Configuration Required</h2>
        <p className="text-red-700 mb-2">Please configure your Privy App ID in your environment variables.</p>
        <ol className="text-red-700 text-sm list-decimal ml-4 space-y-1">
          <li>
            Go to{' '}
            <a
              href="https://dashboard.privy.io/"
              className="underline"
              target="_blank"
              rel="noopener noreferrer">
              https://dashboard.privy.io/
            </a>
          </li>
          <li>Create a new app or select an existing one</li>
          <li>Copy your App ID</li>
          <li>
            Add <code className="bg-red-100 px-1 rounded">NEXT_PUBLIC_PRIVY_APP_ID=your-app-id-here</code>{' '}
            to your .env.local file
          </li>
          <li>Restart your development server</li>
        </ol>
      </div>
    );
  }

  return (
    <PrivyProvider
      appId={privyAppId}
      config={{
        appearance: {
          theme: 'dark',
          accentColor: '#676FFF',
        },
        embeddedWallets: {
          ethereum: {
            createOnLogin: 'all-users',
          },
        },
      }}>
      {children}
    </PrivyProvider>
  );
}
