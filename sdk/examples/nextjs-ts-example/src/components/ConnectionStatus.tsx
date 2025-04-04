import { useConnectionStatus } from '@/hooks/websocket';
import { WSStatus } from '@/types';
import { CryptoKeypair } from '@/websocket/crypto';

// Utility function to shorten long strings (like keys) for display
const shortenString = (str: string, chars: number = 8): string => {
    if (str.length <= chars * 2) return str;
    return str.substring(0, chars) + '...' + str.substring(str.length - chars);
};

interface ConnectionStatusProps {
    status: WSStatus;
    keyPair: CryptoKeypair | null;
    onGenerateKeys: () => Promise<CryptoKeypair | null>;
    onConnect: () => void;
    onDisconnect: () => void;
    onClearKeys?: () => void;
}

export function ConnectionStatus({
    status,
    keyPair,
    onGenerateKeys,
    onConnect,
    onDisconnect,
    onClearKeys,
}: ConnectionStatusProps) {
    const { statusColor, statusText } = useConnectionStatus(status);
    const isGeneratingKeys = status === 'waiting';

    return (
        <div className="mb-6 p-4 bg-gray-800 rounded-lg">
            <div className="flex items-center justify-between mb-4">
                <div className="flex items-center">
                    <div className={`w-3 h-3 rounded-full mr-2 ${statusColor}`} />
                    <span>Status: {statusText}</span>
                </div>

                <div className="space-x-2">
                    <button
                        onClick={onConnect}
                        disabled={status !== 'disconnected' || !keyPair}
                        className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed"
                    >
                        Connect
                    </button>

                    <button
                        onClick={onDisconnect}
                        disabled={status === 'disconnected'}
                        className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 disabled:bg-gray-600 disabled:cursor-not-allowed"
                    >
                        Disconnect
                    </button>
                </div>
            </div>

            <div className="border-t border-gray-700 pt-4">
                <h3 className="text-sm font-medium text-gray-400 mb-2">Authentication</h3>

                <div className="space-y-3">
                    <button
                        onClick={onGenerateKeys}
                        disabled={isGeneratingKeys}
                        className="w-full flex items-center justify-center bg-green-600 hover:bg-green-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white font-medium py-2 px-4 rounded"
                    >
                        {isGeneratingKeys ? (
                            <>
                                <svg
                                    className="animate-spin -ml-1 mr-2 h-4 w-4 text-white"
                                    xmlns="http://www.w3.org/2000/svg"
                                    fill="none"
                                    viewBox="0 0 24 24"
                                >
                                    <circle
                                        className="opacity-25"
                                        cx="12"
                                        cy="12"
                                        r="10"
                                        stroke="currentColor"
                                        strokeWidth="4"
                                    />
                                    <path
                                        className="opacity-75"
                                        fill="currentColor"
                                        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                                    />
                                </svg>
                                Generating...
                            </>
                        ) : (
                            <>
                                <svg
                                    xmlns="http://www.w3.org/2000/svg"
                                    className="h-5 w-5 mr-2"
                                    fill="none"
                                    viewBox="0 0 24 24"
                                    stroke="currentColor"
                                >
                                    <path
                                        strokeLinecap="round"
                                        strokeLinejoin="round"
                                        strokeWidth="2"
                                        d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"
                                    />
                                </svg>
                                Generate Keys
                            </>
                        )}
                    </button>

                    {keyPair && (
                        <>
                            {keyPair.address && (
                                <div className="space-y-1">
                                    <label className="flex items-center text-xs font-medium text-gray-400">
                                        Ethereum Address
                                        <span className="ml-2 px-2 py-0.5 text-xs bg-green-900 text-green-300 rounded-full">
                                            Saved
                                        </span>
                                    </label>
                                    <div className="p-2 bg-gray-900 rounded text-sm text-green-300 font-mono break-all overflow-x-auto">
                                        {keyPair.address}
                                    </div>
                                </div>
                            )}

                            <div className="space-y-1 mt-2">
                                <label className="block text-xs font-medium text-gray-400">Public Key</label>
                                <div className="p-2 bg-gray-900 rounded text-xs text-gray-300 font-mono break-all overflow-x-auto">
                                    {shortenString(keyPair.publicKey, 20)}
                                </div>
                            </div>

                            <div className="space-y-1 mt-2">
                                <label className="block text-xs font-medium text-gray-400">Private Key</label>
                                <div className="p-2 bg-gray-900 rounded text-xs text-gray-300 font-mono break-all overflow-x-auto">
                                    {shortenString(keyPair.privateKey, 20)}
                                </div>
                            </div>

                            {onClearKeys && (
                                <button
                                    onClick={onClearKeys}
                                    className="w-full flex items-center justify-center bg-red-600 hover:bg-red-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white font-medium py-2 px-4 rounded mt-4"
                                >
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        className="h-5 w-5 mr-2"
                                        fill="none"
                                        viewBox="0 0 24 24"
                                        stroke="currentColor"
                                    >
                                        <path
                                            strokeLinecap="round"
                                            strokeLinejoin="round"
                                            strokeWidth="2"
                                            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                                        />
                                    </svg>
                                    Clear Saved Keys
                                </button>
                            )}
                        </>
                    )}
                </div>
            </div>
        </div>
    );
}
