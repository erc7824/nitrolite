// import { useMessageService } from "@/hooks/useMessageService";
import { useSnapshot } from 'valtio';
import WalletStore from '@/store/WalletStore';

interface ChannelStatusProps {
    status: string;
}

export function ChannelStatus({ status }: ChannelStatusProps) {
    const walletSnap = useSnapshot(WalletStore.state);

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
        </div>
    );
}
