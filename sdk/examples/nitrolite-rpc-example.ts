import { NitroliteRPC, NitroliteRPCMessage, NitroliteErrorCode } from "../src/rpc";
import { Hex, Address } from "viem";
import { privateKeyToAccount } from "viem/accounts";
import WebSocket from "ws"; // Or any WebSocket library

/**
 * Simple example of using NitroliteRPC with WebSockets
 */
class NitroliteExample {
    private ws: WebSocket | null = null;
    private address: Address;
    private signer: (message: Hex) => Promise<Hex>;
    private pendingRequests: Map<number, { resolve: Function; reject: Function }> = new Map();
    private nextRequestId: number = 1;

    /**
     * Initialize the example
     */
    constructor(privateKey: Hex, brokerUrl: string) {
        // Create account from private key
        const account = privateKeyToAccount(privateKey);
        this.address = account.address;

        // Create signer function
        this.signer = (message) => account.signMessage({ message });

        // Set up WebSocket
        this.setupWebSocket(brokerUrl);
    }

    /**
     * Set up WebSocket connection
     */
    private setupWebSocket(url: string): void {
        // Create WebSocket connection
        this.ws = new WebSocket(url);

        // Set up event handlers
        this.ws.on("open", () => {
            console.log("Connected to broker");

            // Send initial connection message
            this.ws?.send(
                JSON.stringify({
                    type: "connect",
                    address: this.address,
                })
            );
        });

        this.ws.on("message", (data) => {
            this.handleMessage(data.toString());
        });

        this.ws.on("error", (error) => {
            console.error("WebSocket error:", error);
        });

        this.ws.on("close", () => {
            console.log("Disconnected from broker");
        });
    }

    /**
     * Handle incoming messages
     */
    private async handleMessage(data: string): Promise<void> {
        try {
            const message = JSON.parse(data) as NitroliteRPCMessage;

            // Handle response messages
            if (message.res) {
                const [requestId, method, result] = message.res;
                const pending = this.pendingRequests.get(requestId);

                if (pending) {
                    this.pendingRequests.delete(requestId);
                    pending.resolve(result);
                }
            }

            // Handle error messages
            else if (message.err) {
                const [requestId, code, errorMessage] = message.err;
                const pending = this.pendingRequests.get(requestId);

                if (pending) {
                    this.pendingRequests.delete(requestId);
                    pending.reject(new Error(`Error ${code}: ${errorMessage}`));
                }
            }

            // Handle request messages
            else if (message.req) {
                const [requestId, method, params] = message.req;

                // Handle method calls
                try {
                    let result: any[];

                    if (method === "add") {
                        result = [this.add(params[0], params[1])];
                    } else if (method === "subtract") {
                        result = [this.subtract(params[0], params[1])];
                    } else if (method === "multiply") {
                        result = [this.multiply(params[0], params[1])];
                    } else {
                        await this.sendError(requestId, NitroliteErrorCode.METHOD_NOT_FOUND, `Method '${method}' not found`);
                        return;
                    }

                    await this.sendResponse(requestId, method, result);
                } catch (error: any) {
                    await this.sendError(requestId, NitroliteErrorCode.INTERNAL_ERROR, error.message || "Internal error");
                }
            }
        } catch (error) {
            console.error("Error handling message:", error);
        }
    }

    /**
     * Send a request to the broker
     */
    async sendRequest<T>(method: string, params: any[] = []): Promise<T> {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            throw new Error("Not connected to broker");
        }

        const requestId = this.nextRequestId++;

        // Create request message
        const request = NitroliteRPC.createRequest(method, params, requestId);

        // Sign the request
        const signedRequest = await NitroliteRPC.signMessage(request, this.signer);

        // Send the request
        return new Promise<T>((resolve, reject) => {
            // Store the pending request
            this.pendingRequests.set(requestId, { resolve, reject });

            // Send the request
            this.ws?.send(JSON.stringify(signedRequest));
        });
    }

    /**
     * Send a response to a request
     */
    private async sendResponse(requestId: number, method: string, result: any[]): Promise<void> {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            throw new Error("Not connected to broker");
        }

        // Create response message
        const response = NitroliteRPC.createResponse(requestId, method, result);

        // Sign the response
        const signedResponse = await NitroliteRPC.signMessage(response, this.signer);

        // Send the response
        this.ws.send(JSON.stringify(signedResponse));
    }

    /**
     * Send an error response to a request
     */
    private async sendError(requestId: number, code: number, message: string): Promise<void> {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            throw new Error("Not connected to broker");
        }

        // Create error message
        const error = NitroliteRPC.createError(requestId, code, message);

        // Sign the error
        const signedError = await NitroliteRPC.signMessage(error, this.signer);

        // Send the error
        this.ws.send(JSON.stringify(signedError));
    }

    // Example methods

    private add(a: number, b: number): number {
        return a + b;
    }

    private subtract(a: number, b: number): number {
        return a - b;
    }

    private multiply(a: number, b: number): number {
        return a * b;
    }

    /**
     * Call add method on remote server
     */
    async remoteAdd(a: number, b: number): Promise<number> {
        const result = await this.sendRequest<[number]>("add", [a, b]);
        return result[0];
    }

    /**
     * Call subtract method on remote server
     */
    async remoteSubtract(a: number, b: number): Promise<number> {
        const result = await this.sendRequest<[number]>("subtract", [a, b]);
        return result[0];
    }

    /**
     * Call multiply method on remote server
     */
    async remoteMultiply(a: number, b: number): Promise<number> {
        const result = await this.sendRequest<[number]>("multiply", [a, b]);
        return result[0];
    }

    /**
     * Disconnect from the broker
     */
    disconnect(): void {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }
}

/**
 * Run the example
 */
async function main() {
    // Replace with real values
    const PRIVATE_KEY = "0x0000000000000000000000000000000000000000000000000000000000000001" as Hex;
    const BROKER_URL = "ws://localhost:8080";

    const example = new NitroliteExample(PRIVATE_KEY, BROKER_URL);

    try {
        // Wait a bit for connection to establish
        await new Promise((resolve) => setTimeout(resolve, 1000));

        // Call remote methods
        const sum = await example.remoteAdd(10, 20);
        console.log("10 + 20 =", sum);

        const difference = await example.remoteSubtract(50, 30);
        console.log("50 - 30 =", difference);

        const product = await example.remoteMultiply(5, 7);
        console.log("5 * 7 =", product);

        // Keep the connection alive for a while
        await new Promise((resolve) => setTimeout(resolve, 5000));
    } catch (error) {
        console.error("Example failed:", error);
    } finally {
        // Clean up
        example.disconnect();
    }
}

// Run the example if executed directly
if (require.main === module) {
    main().catch(console.error);
}

// Export for importing in other files
export { NitroliteExample };
