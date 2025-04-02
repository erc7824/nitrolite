import { useSnapshot } from "valtio";
import WalletStore from "@/store/WalletStore";

interface HeaderProps {
    onDisconnect: () => void;
    wsConnected: boolean;
}

export function Header({ onDisconnect, wsConnected }: HeaderProps) {
    const walletSnap = useSnapshot(WalletStore.state);

    return (
        <header className="mb-8">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold text-[#3531ff]">Nitrolite</h1>
                </div>
                {walletSnap.connected && (
                    <div className="flex items-center space-x-2">
                        <span className="text-sm bg-white border border-gray-200 py-1 px-2 rounded font-mono text-gray-700 shadow-sm">
                            {walletSnap.account?.substring(0, 6)}...{walletSnap.account?.substring(38)}
                        </span>
                        <button
                            onClick={onDisconnect}
                            className="bg-red-600 hover:bg-red-700 text-white text-sm py-1 px-2 rounded transition-colors cursor-pointer"
                        >
                            Disconnect
                        </button>
                    </div>
                )}
            </div>
        </header>
    );
}