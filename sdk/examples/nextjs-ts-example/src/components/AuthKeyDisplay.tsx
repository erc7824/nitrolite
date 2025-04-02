import { CryptoKeypair } from "@/websocket";

interface AuthKeyDisplayProps {
    keyPair: CryptoKeypair | null;
    status: string;
}

export function AuthKeyDisplay({ keyPair, status }: AuthKeyDisplayProps) {
    if (!keyPair) return null;

    return (
        <div className="bg-white p-3 rounded-lg border border-gray-200 shadow-sm">
            <div className="flex items-center justify-between mb-2">
                <span className="text-xs font-medium text-gray-700">Authentication Keys</span>
                <div className="flex items-center">
                    <div className={`w-2 h-2 rounded-full mr-1 ${
                      status === "connected" ? "bg-green-500" : 
                      status === "connecting" || status === "authenticating" ? "bg-yellow-500" : 
                      "bg-red-500"
                    }`}></div>
                    <span className="text-xs text-gray-600">
                      {status === "connected" ? "Connected to Broker" : 
                       status === "connecting" ? "Connecting..." :
                       status === "authenticating" ? "Authenticating..." :
                       "Disconnected"}
                    </span>
                </div>
            </div>

            <div className="grid grid-cols-1 gap-1 text-xs">
                <div className="flex items-center">
                    <span className="text-gray-500 w-20">Address:</span>
                    <span className="font-mono bg-gray-100 px-1 py-0.5 rounded-sm text-gray-800 overflow-hidden text-ellipsis flex-1">
                        {keyPair.address}
                    </span>
                    <button
                        className="ml-1 px-1 py-0.5 text-[#3531ff] hover:bg-[#3531ff]/10 rounded cursor-pointer transition-colors"
                        onClick={() => {
                            navigator.clipboard.writeText(keyPair.address || "");
                        }}
                        title="Copy address"
                    >
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            className="h-3 w-3"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                        >
                            <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth="2"
                                d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3"
                            />
                        </svg>
                    </button>
                </div>
                <div className="flex items-center">
                    <span className="text-gray-500 w-20">Public Key:</span>
                    <span className="font-mono bg-gray-100 px-1 py-0.5 rounded-sm text-gray-800 overflow-hidden text-ellipsis flex-1">{`${keyPair.publicKey.substring(
                        0,
                        16
                    )}...${keyPair.publicKey.substring(keyPair.publicKey.length - 16)}`}</span>
                    <button
                        className="ml-1 px-1 py-0.5 text-[#3531ff] hover:bg-[#3531ff]/10 rounded cursor-pointer transition-colors"
                        onClick={() => {
                            navigator.clipboard.writeText(keyPair.publicKey);
                        }}
                        title="Copy public key"
                    >
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            className="h-3 w-3"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                        >
                            <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth="2"
                                d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3"
                            />
                        </svg>
                    </button>
                </div>
                <div className="flex items-center">
                    <span className="text-gray-500 w-20">Private Key:</span>
                    <span className="font-mono bg-gray-100 px-1 py-0.5 rounded-sm text-gray-800 overflow-hidden text-ellipsis flex-1">{`${keyPair.privateKey.substring(
                        0,
                        16
                    )}...${keyPair.privateKey.substring(keyPair.privateKey.length - 16)}`}</span>
                    <button
                        className="ml-1 px-1 py-0.5 text-[#3531ff] hover:bg-[#3531ff]/10 rounded cursor-pointer transition-colors"
                        onClick={() => {
                            navigator.clipboard.writeText(keyPair.privateKey);
                        }}
                        title="Copy private key"
                    >
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            className="h-3 w-3"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                        >
                            <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth="2"
                                d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3"
                            />
                        </svg>
                    </button>
                </div>
            </div>
        </div>
    );
}