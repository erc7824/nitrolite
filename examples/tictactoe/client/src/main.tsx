import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.tsx";
import { NitroliteClientWrapper } from "./context/NitroliteClientWrapper.tsx";
import { WebSocketProvider } from "./context/WebSocketContext.tsx";

createRoot(document.getElementById("root")!).render(
    <StrictMode>
        <NitroliteClientWrapper>
            <WebSocketProvider>
                <App />
            </WebSocketProvider>
        </NitroliteClientWrapper>
    </StrictMode>
);
