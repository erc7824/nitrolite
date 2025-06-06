import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.tsx";
import { NitroliteClientWrapper } from "./context/NitroliteClientWrapper.tsx";
import { WebSocketProvider } from "./context/WebSocketContext.tsx";
import { RainbowKitProvider } from "@rainbow-me/rainbowkit";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { WagmiProvider } from "wagmi";

import { config } from "./lib/config.ts";

const client = new QueryClient();

createRoot(document.getElementById("root")!).render(
    <StrictMode>
        <WagmiProvider config={config}>
            <QueryClientProvider client={client}>
                <RainbowKitProvider>
                    <NitroliteClientWrapper>
                        <WebSocketProvider>
                            <App />
                        </WebSocketProvider>
                    </NitroliteClientWrapper>
                </RainbowKitProvider>
            </QueryClientProvider>
        </WagmiProvider>
    </StrictMode>
);
