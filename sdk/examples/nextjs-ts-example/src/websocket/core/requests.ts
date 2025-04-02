import { NitroliteRPC } from "@erc7824/nitrolite";
import { Channel } from "@/types";
import { WebSocketConnection } from "./connection";
import { WalletSigner } from "../crypto";

/**
 * Class for handling WebSocket requests
 */
export class WSRequests {
  constructor(
    private connection: WebSocketConnection,
    private signer: WalletSigner
  ) {}

  /**
   * Sends a request to the server
   * 
   * @param method - The method to call
   * @param params - The parameters to pass
   * @returns A Promise that resolves with the response
   */
  async sendRequest(method: string, params: any[] = []): Promise<any> {
    if (!this.connection.isConnected) throw new Error("WebSocket not connected");
    return this.sendSignedRequest(NitroliteRPC.createRequest(method, params));
  }

  /**
   * Subscribes to a channel
   * 
   * @param channel - The channel to subscribe to
   * @returns A Promise that resolves when subscribed
   */
  async subscribe(channel: Channel): Promise<void> {
    if (!this.connection.isConnected) throw new Error("WebSocket not connected");

    const request = NitroliteRPC.createRequest("subscribe", [channel]);
    await this.sendSignedRequest(request);
    this.connection.setCurrentChannel(channel);
  }

  /**
   * Publishes a message to the current channel
   * 
   * @param message - The message to publish
   * @returns A Promise that resolves when the message is published
   */
  async publishMessage(message: string): Promise<void> {
    if (!this.connection.isConnected) throw new Error("WebSocket not connected");
    if (!this.connection.currentSubscribedChannel) throw new Error("Not subscribed to any channel");

    const shortenedKey = this.connection.getShortenedPublicKey();
    const request = NitroliteRPC.createRequest(
      "publish", 
      [this.connection.currentSubscribedChannel, message, shortenedKey]
    );
    await this.sendRequestDirect(await NitroliteRPC.signMessage(request, this.signer.sign));
  }

  /**
   * Sends a ping request to the server
   * 
   * @returns A Promise that resolves with the response
   */
  async ping(): Promise<any> {
    return this.sendSignedRequest(NitroliteRPC.createRequest("ping", []));
  }

  /**
   * Checks the balance of a token
   * 
   * @param tokenAddress - The address of the token
   * @returns A Promise that resolves with the balance
   */
  async checkBalance(tokenAddress: string = "0xSHIB..."): Promise<any> {
    return this.sendSignedRequest(NitroliteRPC.createRequest("balance", [tokenAddress]));
  }

  /**
   * Sends multiple requests in batch
   * 
   * @param requests - The requests to send
   * @returns A Promise that resolves with the responses
   */
  async sendBatch(requests: { method: string; params: any[] }[]): Promise<any[]> {
    return Promise.all(requests.map((req) => this.sendRequest(req.method, req.params)));
  }

  /**
   * Helper method to sign and send a request
   * 
   * @param request - The request to send
   * @returns A Promise that resolves with the response
   * @private
   */
  private async sendSignedRequest(request: any): Promise<any> {
    const signedRequest = await NitroliteRPC.signMessage(request, this.signer.sign);
    return this.sendRequestDirect(signedRequest);
  }

  /**
   * Helper method to send a pre-constructed request
   * 
   * @param signedRequest - The signed request to send
   * @returns A Promise that resolves with the response
   * @private
   */
  private async sendRequestDirect(signedRequest: any): Promise<any> {
    if (!this.connection.isConnected) throw new Error("WebSocket not connected");
    const ws = this.connection.webSocket;
    if (!ws) throw new Error("WebSocket not connected");
    
    const pendingRequests = this.connection.getPendingRequests();
    const timeout = this.connection.getRequestTimeout();

    return new Promise((resolve, reject) => {
      const requestId = signedRequest.req[0];
      const requestTimeout = setTimeout(() => {
        if (pendingRequests.has(requestId)) {
          pendingRequests.delete(requestId);
          reject(new Error(`Request timeout: ${signedRequest.req[1]}`));
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

      ws.send(JSON.stringify(signedRequest));
    });
  }
}