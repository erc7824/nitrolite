// import { useMessageService } from "@/hooks/useMessageService";
import { useSnapshot } from 'valtio';
import WalletStore from '@/store/WalletStore';
import NitroliteStore from '@/store/NitroliteStore';
import { useMemo } from 'react';
import { useChannelClosing } from '@/hooks/channel/useChannelClosing';

interface ChannelStatusProps {
    status: string;
}

export function ChannelStatus({ status }: ChannelStatusProps) {
    const walletSnap = useSnapshot(WalletStore.state);
    const nitroliteSnap = useSnapshot(NitroliteStore.state);

    const channelId = useMemo(() => {
        return nitroliteSnap.channelContext.getChannelId();
    }, [nitroliteSnap]);

    const { handleCloseChannel } = useChannelClosing();

    return (
        <div className="bg-white p-3 rounded-lg border border-[#3531ff]/30 shadow-sm flex-1">
            <div className="flex items-center justify-between">
                <div className="flex items-center">
                    <span className="text-md font-semibold text-gray-800 mr-2">Channel Status</span>
                    <span className="px-2 py-0.5 bg-[#3531ff]/20 text-[#3531ff] text-xs rounded">Active</span>
                </div>
                <div className="flex items-center space-x-3">
                    <div className="flex items-center">
                        <div
                            className={`w-2 h-2 rounded-full mr-1 ${
                                status === 'connected'
                                    ? 'bg-green-500'
                                    : status === 'connecting' || status === 'authenticating'
                                      ? 'bg-yellow-500'
                                      : 'bg-red-500'
                            }`}
                        />
                        <span className="text-xs text-gray-600">
                            {status === 'connected'
                                ? 'Channel Active'
                                : status === 'connecting'
                                  ? 'Connecting...'
                                  : status === 'authenticating'
                                    ? 'Authenticating...'
                                    : 'Disconnected'}
                        </span>
                    </div>
                    <div className="text-xs text-gray-600 font-mono">
                        <span className="px-2 py-0.5 bg-gray-100 rounded-sm">
                            {walletSnap.selectedTokenAddress?.substring(0, 6)}...
                            {walletSnap.selectedTokenAddress?.substring(38)}
                        </span>
                    </div>
                    <div className="text-xs text-gray-600">
                        Amount: <span className="font-mono text-gray-800">{walletSnap.selectedAmount}</span>
                    </div>
                </div>
            </div>

            <div className="mt-3 flex items-center justify-between border-t pt-3">
                <div className="flex items-center space-x-2">
                    <span className="text-sm text-gray-600">Channel ID:</span>
                    <div className="flex items-center space-x-2">
                        <code className="px-2 py-1 bg-gray-100 rounded text-sm font-mono">
                            {channelId?.substring(0, 6)}...{channelId?.substring(channelId.length - 4)}
                        </code>
                        <button
                            onClick={() => navigator.clipboard.writeText(channelId)}
                            className="p-1 hover:bg-gray-100 rounded">
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                className="h-4 w-4 text-gray-600"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor">
                                <path
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    strokeWidth={2}
                                    d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                                />
                            </svg>
                        </button>
                    </div>
                </div>
                <button
                    onClick={() =>
                        handleCloseChannel(channelId, walletSnap.selectedTokenAddress, [walletSnap.selectedAmount, '0'])
                    }
                    className="px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors pointer">
                    Close
                </button>
            </div>
        </div>
    );
}
