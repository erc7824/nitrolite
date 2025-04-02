export function About() {
    return (
        <div className="p-4 bg-white rounded-lg border border-gray-200 shadow-md">
            <h2 className="text-lg font-semibold mb-4 text-[#3531ff]">About Nitrolite</h2>

            <div className="space-y-4">
                <p>Nitrolite provides secure state channels with cryptographic authentication for fast and secure transactions.</p>

                <div className="bg-gray-50 p-3 rounded-lg border-l-2 border-[#3531ff]">
                    <h3 className="text-sm font-semibold mb-2 text-[#3531ff]">How to use Nitrolite:</h3>
                    <ol className="list-decimal list-inside space-y-1 text-sm text-gray-600">
                        <li>Connect your MetaMask wallet</li>
                        <li>Select token and amount</li>
                        <li>Open channel to enable secure transactions</li>
                        <li>Use channel functions for messaging and transactions</li>
                    </ol>
                </div>

                <div className="grid grid-cols-3 gap-4 text-center text-xs">
                    <div className="bg-white p-3 rounded-lg border border-gray-200 shadow-sm">
                        <div className="mb-2">
                            <svg
                                className="w-8 h-8 mx-auto mb-1 text-[#3531ff]"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                                xmlns="http://www.w3.org/2000/svg"
                            >
                                <path
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    strokeWidth="2"
                                    d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
                                />
                            </svg>
                            <span className="block">Secure</span>
                        </div>
                        <p className="text-gray-600">All transactions cryptographically signed</p>
                    </div>

                    <div className="bg-white p-3 rounded-lg border border-gray-200 shadow-sm">
                        <div className="mb-2">
                            <svg
                                className="w-8 h-8 mx-auto mb-1 text-[#3531ff]"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                                xmlns="http://www.w3.org/2000/svg"
                            >
                                <path
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    strokeWidth="2"
                                    d="M8 12h.01M12 12h.01M16 12h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                                />
                            </svg>
                            <span className="block">Fast</span>
                        </div>
                        <p className="text-gray-600">Instant transactions without on-chain delays</p>
                    </div>

                    <div className="bg-white p-3 rounded-lg border border-gray-200 shadow-sm">
                        <div className="mb-2">
                            <svg
                                className="w-8 h-8 mx-auto mb-1 text-[#3531ff]"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                                xmlns="http://www.w3.org/2000/svg"
                            >
                                <path
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    strokeWidth="2"
                                    d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z"
                                />
                            </svg>
                            <span className="block">Multi-Chain</span>
                        </div>
                        <p className="text-gray-600">Supports multiple blockchain networks</p>
                    </div>
                </div>

                <p className="text-xs text-center text-gray-500 pt-2">Made by the Nitro team.</p>
            </div>
        </div>
    );
}
