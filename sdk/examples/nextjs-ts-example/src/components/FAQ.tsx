export function FAQ() {
    return (
        <div className="bg-white p-4 rounded-lg border border-gray-200 shadow-sm w-full md:w-1/2">
            <h2 className="text-lg font-semibold mb-3 text-[#3531ff]">FAQ</h2>
            <div className="space-y-2">
                <details className="group">
                    <summary className="flex justify-between items-center font-medium cursor-pointer text-sm text-gray-700">
                        <span>What are state channels?</span>
                        <span className="transition group-open:rotate-180">
                            <svg fill="none" height="12" width="12" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24"><path d="M6 9l6 6 6-6"></path></svg>
                        </span>
                    </summary>
                    <p className="text-xs text-gray-600 mt-1 group-open:animate-fadeIn">
                        State channels allow for off-chain transactions that are later settled on-chain, reducing gas costs and increasing speed.
                    </p>
                </details>
                
                <details className="group">
                    <summary className="flex justify-between items-center font-medium cursor-pointer text-sm text-gray-700">
                        <span>How secure are these channels?</span>
                        <span className="transition group-open:rotate-180">
                            <svg fill="none" height="12" width="12" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24"><path d="M6 9l6 6 6-6"></path></svg>
                        </span>
                    </summary>
                    <p className="text-xs text-gray-600 mt-1 group-open:animate-fadeIn">
                        All transactions are cryptographically signed and verified, ensuring the same security guarantees as on-chain transactions.
                    </p>
                </details>
                
                <details className="group">
                    <summary className="flex justify-between items-center font-medium cursor-pointer text-sm text-gray-700">
                        <span>Can I close my channel?</span>
                        <span className="transition group-open:rotate-180">
                            <svg fill="none" height="12" width="12" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24"><path d="M6 9l6 6 6-6"></path></svg>
                        </span>
                    </summary>
                    <p className="text-xs text-gray-600 mt-1 group-open:animate-fadeIn">
                        Yes, you can close your channel at any time, which will settle the final state on-chain.
                    </p>
                </details>
                
                <details className="group">
                    <summary className="flex justify-between items-center font-medium cursor-pointer text-sm text-gray-700">
                        <span>What tokens are supported?</span>
                        <span className="transition group-open:rotate-180">
                            <svg fill="none" height="12" width="12" stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24"><path d="M6 9l6 6 6-6"></path></svg>
                        </span>
                    </summary>
                    <p className="text-xs text-gray-600 mt-1 group-open:animate-fadeIn">
                        Nitrolite supports a wide range of ERC-20 tokens across multiple blockchain networks.
                    </p>
                </details>
            </div>
        </div>
    );
}