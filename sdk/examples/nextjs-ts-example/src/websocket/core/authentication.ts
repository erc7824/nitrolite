import { NitroliteRPC } from "@erc7824/nitrolite";
import { WalletSigner } from "../crypto";

/**
 * Authenticates with the WebSocket server
 *
 * @param ws - The WebSocket connection
 * @param signer - The signer to use for authentication
 * @param timeout - Timeout in milliseconds
 * @returns A Promise that resolves when authenticated
 */
export async function authenticate(ws: WebSocket, signer: WalletSigner, timeout: number): Promise<void> {
    if (!ws) throw new Error("WebSocket not connected");

    const authRequest = NitroliteRPC.createRequest("auth", [signer.publicKey]);
    console.log("authRequest", authRequest);
    const signedAuthRequest = await NitroliteRPC.signMessage(authRequest, signer.sign);
    console.log("signedAuthRequest", signedAuthRequest);

    return new Promise((resolve, reject) => {
        if (!ws) return reject(new Error("WebSocket not connected"));

        const authTimeout = setTimeout(() => {
            ws.removeEventListener("message", handleAuthResponse);
            reject(new Error("Authentication timeout"));
        }, timeout);

        const handleAuthResponse = (event: MessageEvent) => {
            let response;

            try {
                response = JSON.parse(event.data);
            } catch (error) {
                console.error("Error parsing auth response:", error);
                console.log("Raw auth message:", event.data);
                return; // Continue waiting for valid responses
            }

            try {
                if ((response.res && response.res[1] === "auth") || response.type === "auth_success") {
                    clearTimeout(authTimeout);
                    ws.removeEventListener("message", handleAuthResponse);
                    resolve();
                } else if ((response.err && response.err[1]) || response.type === "auth_error") {
                    clearTimeout(authTimeout);
                    ws.removeEventListener("message", handleAuthResponse);
                    const errorMsg = response.err ? response.err[2] : response.error || "Authentication failed";
                    reject(new Error(errorMsg));
                }
            } catch (error) {
                console.error("Error handling auth response:", error);
                clearTimeout(authTimeout);
                ws.removeEventListener("message", handleAuthResponse);
                reject(new Error(`Authentication error: ${error instanceof Error ? error.message : String(error)}`));
            }
        };

        ws.addEventListener("message", handleAuthResponse);
        ws.send(JSON.stringify(signedAuthRequest));
    });
}
