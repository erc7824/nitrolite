import { usePrivy } from '@privy-io/react-auth';
import { useEffect, useState } from 'react';

export function useLoginState() {
    const { ready, authenticated, user, login, logout } = usePrivy();
    const [isReady, setIsReady] = useState(false);
    const [isWalletReady, setIsWalletReady] = useState(false);

    useEffect(() => {
        if (ready) {
            setIsReady(true);
        }
    }, [ready]);

    // Track when wallet is ready
    useEffect(() => {
        if (authenticated && user?.wallet?.address) {
            setIsWalletReady(true);
        } else {
            setIsWalletReady(false);
        }
    }, [authenticated, user?.wallet?.address]);

    const walletAddress = user?.wallet?.address;

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
