import { parseRPCResponse, RPCChannelStatus, RPCMethod } from '@erc7824/nitrolite';
import { WebSocket } from 'ws';

export class TestWebSocket {
    private socket: WebSocket | null = null;
    private messageListeners: ((data: string) => void)[] = [];

    constructor(private url: string, private debugMode = false) {}

    connect(): Promise<void> {
        return new Promise((resolve, reject) => {
            if (this.socket) {
                return resolve();
            }

            this.socket = new WebSocket(this.url);

            this.socket.on('open', () => {
                if (this.debugMode) {
                    console.log('WebSocket connection established');
                }

                resolve();
            });

            this.socket.on('message', (event) => {
                if (this.debugMode) {
                    console.log('Message received:', event.toString());
                }

                for (const listener of this.messageListeners) {
                    try {
                        listener(event.toString());
                    } catch (error) {
                        console.error('Error in message listener:', error);
                    }
                }
            });

            this.socket.on('error', (error) => {
                if (this.debugMode) {
                    console.error('WebSocket error:', error);
                }

                reject(error);
            });

            this.socket.on('close', () => {
                this.socket = null;
            });
        });
    }

    waitForMessage(predicate: (data: string) => boolean, timeout = 1000): Promise<string> {
        return new Promise((resolve, reject) => {
            const timeoutId = setTimeout(() => {
                this.messageListeners = this.messageListeners.filter((l) => l !== messageHandler);
                reject(new Error(`Timeout waiting for message after ${timeout}ms`));
            }, timeout);

            const messageHandler = (data: string) => {
                if (predicate(data)) {
                    clearTimeout(timeoutId);
                    this.messageListeners = this.messageListeners.filter((l) => l !== messageHandler);
                    resolve(data);
                }
            };

            this.messageListeners.push(messageHandler);
        });
    }

    send(message: string): void {
        if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
            throw new Error('WebSocket is not open. Cannot send message.');
        }
        this.socket.send(message);
    }

    sendAndWaitForResponse(message: string, predicate: (data: string) => boolean, timeout = 1000): Promise<any> {
        const messagePromise = this.waitForMessage(predicate, timeout);
        this.send(message);
        return messagePromise;
    }

    close(): void {
        if (this.socket) {
            this.socket.close();
            this.socket = null;
            this.messageListeners = [];
        }
    }
}

export const getPongPredicate = () => {
    return (data: string): boolean => {
        try {
            const parsedData = parseRPCResponse(data);
            if (parsedData.method === RPCMethod.Pong) {
                return true;
            }
        } catch (error) {
            console.error('Error parsing data for PongPredicate:', error);
        }

        return false;
    };
};

export const getAuthChallengePredicate = () => {
    return (data: string): boolean => {
        try {
            const parsedData = parseRPCResponse(data);
            if (parsedData.method === RPCMethod.AuthChallenge) {
                return true;
            }
        } catch (error) {
            console.error('Error parsing data for AuthChallengePredicate:', error);
        }

        return false;
    };
};

export const getAuthVerifyPredicate = () => {
    return (data: string): boolean => {
        try {
            const parsedData = parseRPCResponse(data);
            if (parsedData.method === RPCMethod.AuthVerify) {
                return true;
            }
        } catch (error) {
            console.error('Error parsing data for AuthVerifyPredicate:', error);
        }

        return false;
    };
};

export const getChannelUpdatePredicateWithStatus = (status: RPCChannelStatus) => {
    return (data: string): boolean => {
        try {
            const parsedData = parseRPCResponse(data);
            if (parsedData.method === RPCMethod.ChannelUpdate && parsedData.params[0].status === status) {
                return true;
            }
        } catch (error) {
            console.error('Error parsing data for ChannelUpdatePredicate:', error);
        }

        return false;
    };
};

export const getCloseChannelPredicate = () => {
    return (data: string): boolean => {
        try {
            const parsedData = parseRPCResponse(data);
            if (parsedData.method === RPCMethod.CloseChannel) {
                return true;
            }
        } catch (error) {
            console.error('Error parsing data for ChannelUpdatePredicate:', error);
        }

        return false;
    };
};

export const getResizeChannelPredicate = () => {
    return (data: string): boolean => {
        try {
            const parsedData = parseRPCResponse(data);
            if (parsedData.method === RPCMethod.ResizeChannel) {
                return true;
            }
        } catch (error) {
            console.error('Error parsing data for ResizeChannelPredicate:', error);
        }

        return false;
    };
};

export const getErrorPredicate = () => {
    return (data: string): boolean => {
        try {
            const parsedData = parseRPCResponse(data);
            if (parsedData.method === RPCMethod.Error) {
                return true;
            }
        } catch (error) {
            console.error('Error parsing data for ErrorPredicate:', error);
        }

        return false;
    };
}
