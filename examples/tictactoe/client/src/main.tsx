import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import './index.css';
import '@rainbow-me/rainbowkit/styles.css';
import App from './App.tsx';
import { NitroliteClientWrapper } from './context/NitroliteClientWrapper.tsx';
import { WebSocketProvider } from './context/WebSocketContext.tsx';
import { RainbowKitConnectProvider } from './providers/RainbowkitConnectProvider.tsx';

createRoot(document.getElementById('root')!).render(
    <StrictMode>
        <RainbowKitConnectProvider>
            <NitroliteClientWrapper>
                <WebSocketProvider>
                    <App />
                </WebSocketProvider>
            </NitroliteClientWrapper>
        </RainbowKitConnectProvider>
    </StrictMode>
);
