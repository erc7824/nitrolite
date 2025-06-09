import { WalletButton } from "@rainbow-me/rainbowkit";
import { Button } from "./button";
import { Wallet, Loader2 } from "lucide-react";

export function StyledWalletButton() {
  return (
    <WalletButton.Custom wallet="metaMask">
      {({ ready, connect, connecting }) => {
        return (
          <Button 
            onClick={connect}
            disabled={!ready || connecting}
            className="w-full bg-gradient-to-r from-cyan-600 via-cyan-500 to-blue-600 hover:from-cyan-500 hover:via-cyan-400 hover:to-blue-500 text-white border-0 font-semibold py-3 px-6 rounded-lg shadow-lg hover:shadow-cyan-500/25 transition-all duration-300 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {connecting ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Connecting...
              </>
            ) : (
              <>
                <Wallet className="h-4 w-4 mr-2" />
                Connect MetaMask
              </>
            )}
          </Button>
        );
      }}
    </WalletButton.Custom>
  );
}