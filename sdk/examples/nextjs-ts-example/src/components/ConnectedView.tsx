import { ChannelStatus } from '@/components/ChannelStatus';
import { AuthKeyDisplay } from '@/components/AuthKeyDisplay';
import { ConnectionStatus } from '@/components/ConnectionStatus';
import { MessageList } from '@/components/MessageList';
import { InfoSection } from '@/components/InfoSection';
import { Channel } from '@/types';
import { CryptoKeypair } from '@/websocket/crypto';

interface ConnectedViewProps {
    status: string;
    keyPair: CryptoKeypair | null;
    isConnected: boolean;
    wsChannel?: Channel | null;
    onGenerateKeys?: () => Promise<any>;
    onConnect?: () => void;
    onDisconnect?: () => void;
}

export function ConnectedView({
    status,
    keyPair,
    isConnected,
    wsChannel,
    onGenerateKeys,
    onConnect,
    onDisconnect,
}: ConnectedViewProps) {
    return (
        <>
            <div className="flex gap-3 mb-2 flex-col md:flex-row">
                <ChannelStatus status={status} />
                <AuthKeyDisplay keyPair={keyPair} status={status} />
            </div>
            
            {onGenerateKeys && onConnect && onDisconnect && (
                <ConnectionStatus 
                    status={status as any}
                    keyPair={keyPair}
                    onGenerateKeys={onGenerateKeys}
                    onConnect={onConnect}
                    onDisconnect={onDisconnect}
                />
            )}

            <MessageList />

            <InfoSection />
        </>
    );
}
