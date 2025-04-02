export function About() {
  return (
    <div className="p-4 bg-gray-800 rounded-lg">
      <h2 className="text-lg font-semibold mb-4 text-primary-300">About</h2>
      
      <div className="space-y-4">
        <p>
          This is a secure WebSocket client with cryptographic authentication using the NitroRPC protocol.
        </p>
        
        <div className="bg-gray-900 bg-opacity-50 p-3 rounded-lg border-l-2 border-primary-500">
          <h3 className="text-sm font-semibold mb-2 text-primary-400">How to use this client:</h3>
          <ol className="list-decimal list-inside space-y-1 text-sm text-gray-300">
            <li>Generate a cryptographic key pair for authentication</li>
            <li>Connect to the WebSocket server</li>
            <li>Subscribe to a channel to send and receive messages</li>
            <li>Try the utility functions like Ping and Check Balance</li>
          </ol>
        </div>
        
        <div className="grid grid-cols-3 gap-4 text-center text-xs">
          <div className="bg-gray-900 p-3 rounded-lg">
            <div className="mb-2">
              <svg className="w-8 h-8 mx-auto mb-1 text-primary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" xmlns="http://www.w3.org/2000/svg">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
              </svg>
              <span className="block">Secure</span>
            </div>
            <p className="text-gray-400">All messages signed with your private key</p>
          </div>
          
          <div className="bg-gray-900 p-3 rounded-lg">
            <div className="mb-2">
              <svg className="w-8 h-8 mx-auto mb-1 text-secondary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" xmlns="http://www.w3.org/2000/svg">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span className="block">Efficient</span>
            </div>
            <p className="text-gray-400">Using NitroRPC for fast, lightweight communication</p>
          </div>
          
          <div className="bg-gray-900 p-3 rounded-lg">
            <div className="mb-2">
              <svg className="w-8 h-8 mx-auto mb-1 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" xmlns="http://www.w3.org/2000/svg">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
              </svg>
              <span className="block">Versatile</span>
            </div>
            <p className="text-gray-400">Supports multiple channels and custom RPC methods</p>
          </div>
        </div>
        
        <p className="text-xs text-center text-gray-500 pt-2">
          Made with ethers.js and Next.js for ETHGlobal Hackathon. Source available on GitHub.
        </p>
      </div>
    </div>
  );
}