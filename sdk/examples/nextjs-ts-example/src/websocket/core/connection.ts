import { WSStatus, Channel } from "@/types";
import { WebSocketReadyState, WebSocketClientOptions } from "./types";
import { WalletSigner, shortenPublicKey } from "../crypto";
import { handleMessage } from "./messageHandler";
import { authenticate } from "./authentication";

/**
 * Client for WebSocket connection management
 */
export class WebSocketConnection {
  private ws: WebSocket | null = null;
  private pendingRequests = new Map<number, { resolve: Function; reject: Function }>();
  private reconnectTimeout: NodeJS.Timeout | null = null;
  private reconnectAttempts = 0;
  private onStatusChangeCallback?: (status: WSStatus) => void;
  private onMessageCallback?: (message: any) => void;
  private onErrorCallback?: (error: Error) => void;
  private currentChannel: Channel | null = null;

  /**
   * Creates a new WebSocketConnection
   * 
   * @param url - The WebSocket URL to connect to
   * @param signer - The WalletSigner to use for authentication
   * @param options - Configuration options
   */
  constructor(
    private url: string,
    private signer: WalletSigner,
    private options: WebSocketClientOptions = {
      autoReconnect: true,
      reconnectDelay: 1000,
      maxReconnectAttempts: 5,
      requestTimeout: 10000,
    }
  ) {}

  /**
   * Sets a callback to be called when the connection status changes
   */
  onStatusChange(cb: (status: WSStatus) => void): void {
    this.onStatusChangeCallback = cb;
  }

  /**
   * Sets a callback to be called when a message is received
   */
  onMessage(cb: (message: any) => void): void {
    this.onMessageCallback = cb;
  }

  /**
   * Sets a callback to be called when an error occurs
   */
  onError(cb: (error: Error) => void): void {
    this.onErrorCallback = cb;
  }

  /**
   * Gets the current ready state of the WebSocket
   */
  get readyState(): WebSocketReadyState {
    return this.ws ? this.ws.readyState : WebSocketReadyState.CLOSED;
  }

  /**
   * Gets whether the WebSocket is currently connected
   */
  get isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocketReadyState.OPEN;
  }

  /**
   * Gets the currently subscribed channel, if any
   */
  get currentSubscribedChannel(): Channel | null {
    return this.currentChannel;
  }

  /**
   * Gets a shortened version of the signer's public key
   */
  getShortenedPublicKey(): string {
    return shortenPublicKey(this.signer.publicKey);
  }

  /**
   * Connects to the WebSocket server
   */
  connect(): Promise<void> {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }

    return new Promise((resolve, reject) => {
      try {
        if (this.isConnected) return resolve();

        this.ws = new WebSocket(this.url);
        this.onStatusChangeCallback?.("connecting");

        this.ws.onopen = async () => {
          try {
            this.onStatusChangeCallback?.("authenticating");
            await authenticate(this.ws, this.signer, this.options.requestTimeout);
            this.onStatusChangeCallback?.("connected");
            this.reconnectAttempts = 0;
            resolve();
          } catch (error) {
            this.onStatusChangeCallback?.("auth_failed");
            this.onErrorCallback?.(error instanceof Error ? error : new Error(String(error)));
            reject(error);
            this.close();
            this.handleReconnect();
          }
        };

        this.ws.onmessage = (event) => {
          handleMessage(
            event,
            this.pendingRequests,
            (channel) => this.currentChannel = channel,
            this.onMessageCallback,
            this.onErrorCallback
          );
        };

        this.ws.onerror = () => {
          this.onErrorCallback?.(new Error("WebSocket connection error"));
          reject(new Error("WebSocket connection error"));
        };

        this.ws.onclose = () => {
          this.onStatusChangeCallback?.("disconnected");
          this.ws = null;
          this.currentChannel = null;

          this.pendingRequests.forEach(({ reject }) => 
            reject(new Error("WebSocket connection closed"))
          );
          this.pendingRequests.clear();

          this.handleReconnect();
        };
      } catch (error) {
        reject(error);
        this.handleReconnect();
      }
    });
  }

  /**
   * Handles reconnection logic
   */
  private handleReconnect(): void {
    if (!this.options.autoReconnect || 
        this.reconnectAttempts >= this.options.maxReconnectAttempts) {
      if (this.reconnectAttempts >= this.options.maxReconnectAttempts) {
        this.onStatusChangeCallback?.("reconnect_failed");
      }
      return;
    }

    if (this.reconnectTimeout) clearTimeout(this.reconnectTimeout);

    this.reconnectAttempts++;
    const delay = this.options.reconnectDelay * Math.pow(1.5, this.reconnectAttempts - 1);
    this.onStatusChangeCallback?.("reconnecting");

    this.reconnectTimeout = setTimeout(() => {
      this.connect().catch(() => {});
    }, delay);
  }

  /**
   * Closes the WebSocket connection
   */
  close(): void {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }

    if (this.ws && [WebSocketReadyState.OPEN, WebSocketReadyState.CONNECTING].includes(this.ws.readyState)) {
      try {
        this.ws.close(1000, "Normal closure");
      } catch (err) {
        console.error("Error while closing WebSocket:", err);
      }
    }
    this.ws = null;
    this.currentChannel = null;

    this.pendingRequests.forEach(({ reject }) => 
      reject(new Error("WebSocket connection closed by client"))
    );
    this.pendingRequests.clear();
    this.onStatusChangeCallback?.("disconnected");
  }

  // Expose the WebSocket and pendingRequests for use by other modules
  get webSocket(): WebSocket | null {
    return this.ws;
  }

  getPendingRequests(): Map<number, { resolve: Function; reject: Function }> {
    return this.pendingRequests;
  }

  setCurrentChannel(channel: Channel): void {
    this.currentChannel = channel;
  }

  getRequestTimeout(): number {
    return this.options.requestTimeout;
  }
}