import { useState, useCallback } from 'react';
import { CryptoKeypair, generateKeyPair, getAddressFromPublicKey } from '@/websocket';

// Custom hook to load keypair from localStorage
export function useKeyPair() {
    const [keyPair, setKeyPair] = useState<CryptoKeypair | null>(() => {
        if (typeof window === 'undefined') return null;
        
        const savedKeys = localStorage.getItem('crypto_keypair');
        if (!savedKeys) return null;
        
        try {
            const parsed = JSON.parse(savedKeys) as CryptoKeypair;
            
            if (parsed.publicKey && !parsed.address) {
                parsed.address = getAddressFromPublicKey(parsed.publicKey);
                localStorage.setItem('crypto_keypair', JSON.stringify(parsed));
            }
            
            return parsed;
        } catch (e) {
            console.error('Failed to parse saved keys:', e);
            return null;
        }
    });
    
    // Function to clear saved keys
    const clearKeys = useCallback(() => {
        if (typeof window !== 'undefined') {
            localStorage.removeItem('crypto_keypair');
        }
        setKeyPair(null);
    }, []);
    
    // Generate a new key pair
    const generateKeys = useCallback(async () => {
        try {
            const newKeyPair = await generateKeyPair();
            setKeyPair(newKeyPair);
            
            if (typeof window !== 'undefined') {
                localStorage.setItem('crypto_keypair', JSON.stringify(newKeyPair));
            }
            
            return newKeyPair;
        } catch (error) {
            console.error(`Error generating keys:`, error);
            return null;
        }
    }, []);
    
    return { keyPair, clearKeys, generateKeys, hasKeys: !!keyPair };
}