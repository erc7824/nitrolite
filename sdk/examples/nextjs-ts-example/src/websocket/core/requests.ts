import { NitroliteRPC } from "@erc7824/nitrolite";
import { Channel } from "@/types";
import { WebSocketConnection } from "./connection";
import { WalletSigner } from "../crypto";
import MessageService from "../services/MessageService";

/**
 * Handles WebSocket requests with NitroliteRPC
 */
export class WSRequests {
    constructor(private connection: WebSocketConnection, private signer: WalletSigner) {}

    /**
     * Sends a request to the server
     */
    async sendRequest(method: string, params: any[] = []): Promise<any> {
        if (!this.connection.isConnected) {
            const errorMsg = "WebSocket not connected";
            MessageService.error(errorMsg);
            throw new Error(errorMsg);
        }
        
        MessageService.system(`Sending request: ${method}`);
        return this.sendSignedRequest(NitroliteRPC.createRequest(method, params));
    }

    /**
     * Subscribes to a channel
     */
    async subscribe(channel: Channel): Promise<void> {
        if (!this.connection.isConnected) {
            const errorMsg = "WebSocket not connected";
            MessageService.error(errorMsg);
            throw new Error(errorMsg);
        }

        MessageService.system(`Subscribing to channel: ${channel}`);
        
        const request = NitroliteRPC.createRequest("subscribe", [channel]);
        await this.sendSignedRequest(request);
        this.connection.setCurrentChannel(channel);
    }

    /**
     * Publishes a message to the current channel
     */
    async publishMessage(message: string): Promise<void> {
        if (!this.connection.isConnected) {
            const errorMsg = "WebSocket not connected";
            MessageService.error(errorMsg);
            throw new Error(errorMsg);
        }
        
        if (!this.connection.currentSubscribedChannel) {
            const errorMsg = "Not subscribed to any channel";
            MessageService.error(errorMsg);
            throw new Error(errorMsg);
        }

        const shortenedKey = this.connection.getShortenedPublicKey();
        MessageService.sent(message, shortenedKey);
        
        const request = NitroliteRPC.createRequest("publish", [
            this.connection.currentSubscribedChannel, 
            message, 
            shortenedKey
        ]);
        
        await this.sendRequestDirect(await NitroliteRPC.signMessage(request, this.signer.sign));
    }

    /**
     * Sends a ping request to the server
     */
    async ping(): Promise<any> {
        MessageService.system("Sending ping request");
        return this.sendSignedRequest(NitroliteRPC.createRequest("ping", []));
    }

    /**
     * Checks the balance of a token
     */
    async checkBalance(tokenAddress: string = "0xSHIB..."): Promise<any> {
        MessageService.system(`Checking balance for token: ${tokenAddress}`);
        return this.sendSignedRequest(NitroliteRPC.createRequest("balance", [tokenAddress]));
    }

    /**
     * Sends multiple requests in batch
     */
    async sendBatch(requests: { method: string; params: any[] }[]): Promise<any[]> {
        MessageService.system(`Sending batch of ${requests.length} requests`);
        return Promise.all(requests.map((req) => this.sendRequest(req.method, req.params)));
    }

    /**
     * Helper method to sign and send a request
     */
    private async sendSignedRequest(request: any): Promise<any> {
        try {
            const signedRequest = await NitroliteRPC.signMessage(request, this.signer.sign);
            return this.sendRequestDirect(signedRequest);
        } catch (error) {
            const errorMsg = `Signing request failed: ${error instanceof Error ? error.message : String(error)}`;
            MessageService.error(errorMsg);
            throw new Error(errorMsg);
        }
    }

    /**
     * Helper method to send a pre-constructed request
     */
    private async sendRequestDirect(signedRequest: any): Promise<any> {
        if (!this.connection.isConnected || !this.connection.webSocket) {
            const errorMsg = "WebSocket not connected";
            MessageService.error(errorMsg);
            throw new Error(errorMsg);
        }

        const pendingRequests = this.connection.getPendingRequests();
        const timeout = this.connection.getRequestTimeout();
        const ws = this.connection.webSocket;

        return new Promise((resolve, reject) => {
            const requestId = signedRequest.req[0];
            const method = signedRequest.req[1];
            
            const requestTimeout = setTimeout(() => {
                if (pendingRequests.has(requestId)) {
                    pendingRequests.delete(requestId);
                    const timeoutMsg = `Request timeout: ${method}`;
                    MessageService.error(timeoutMsg);
                    reject(new Error(timeoutMsg));
                }
            }, timeout);

            pendingRequests.set(requestId, {
                resolve: (result: any) => {
                    clearTimeout(requestTimeout);
                    resolve(result);
                },
                reject: (error: Error) => {
                    clearTimeout(requestTimeout);
                    reject(error);
                },
            });

            try {
                ws.send(JSON.stringify(signedRequest));
            } catch (error) {
                clearTimeout(requestTimeout);
                pendingRequests.delete(requestId);
                
                const errorMsg = `Failed to send message: ${error instanceof Error ? error.message : String(error)}`;
                MessageService.error(errorMsg);
                reject(new Error(errorMsg));
            }
        });
    }
}