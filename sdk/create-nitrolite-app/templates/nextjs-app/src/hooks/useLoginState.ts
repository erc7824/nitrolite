import { usePrivy, useWallets } from '@privy-io/react-auth';
import { useEffect, useState } from 'react';

export function useLoginState() {
    const { ready, authenticated, user, login, logout } = usePrivy();
    const { wallets } = useWallets();
    const [isReady, setIsReady] = useState(false);
    const [isWalletReady, setIsWalletReady] = useState(false);

    useEffect(() => {
        if (ready) {
            setIsReady(true);
        }
    }, [ready]);

    // Find embedded wallet address
    const embeddedPrivyWallet = wallets.find((wallet) => wallet.walletClientType === 'privy');
    const walletAddress = embeddedPrivyWallet?.address;

    // Track when wallet is ready
    useEffect(() => {
        if (authenticated && walletAddress) {
            setIsWalletReady(true);
        } else {
            setIsWalletReady(false);
        }
    }, [authenticated, walletAddress]);

    return {
        isLoggedIn: authenticated,
        walletAddress,
        login,
        logout,
        isReady,
        user,
        isWalletReady,
    };
}
