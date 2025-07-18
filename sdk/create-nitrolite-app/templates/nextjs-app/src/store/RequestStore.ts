import { proxy } from 'valtio';
import { RPCMethod, type RequestID } from '@erc7824/nitrolite';

const STORE_LIMIT = 100;
const DEFAULT_TIMEOUT = 30000; // 30 sec

export interface RequestInfo {
    requestId: RequestID;
    method: RPCMethod;
    params: any;
    timestamp: number;
}

export interface PendingRequest {
    requestInfo: RequestInfo;
    resolve: (response: any) => void;
    reject: (error: Error) => void;
    timeout?: number;
    timeoutId?: number;
}

export interface RequestState {
    status: 'pending' | 'success' | 'failed';
    requestInfo: RequestInfo;
    response?: any;
    error?: string;
    timestamp: number;
}

export interface IRequestStoreState {
    pendingRequests: Map<RequestID, PendingRequest>;
    completedRequests: Map<RequestID, RequestState>;
}

const state = proxy<IRequestStoreState>({
    pendingRequests: new Map(),
    completedRequests: new Map(),
});

const RequestStore = {
    state,

    /**
     * Registers a pending request and returns a Promise that resolves when response arrives
     */
    registerRequest(requestInfo: RequestInfo, options: { timeout?: number } = {}): Promise<any> {
        const pendingRequest: PendingRequest = {
            requestInfo,
            resolve: () => {},
            reject: () => {},
            timeout: options.timeout || DEFAULT_TIMEOUT,
        };

        if (pendingRequest.timeout) {
            pendingRequest.timeoutId = window.setTimeout(() => {
                this.handleTimeout(requestInfo.requestId);
            }, pendingRequest.timeout);
        }

        state.pendingRequests.set(requestInfo.requestId, pendingRequest);

        return new Promise((resolve, reject) => {
            pendingRequest.resolve = resolve;
            pendingRequest.reject = reject;

            // Poll for results
            const checkResult = () => {
                const completed = state.completedRequests.get(requestInfo.requestId);
                if (completed) {
                    if (completed.status === 'success') {
                        resolve(completed.response);
                    } else {
                        reject(new Error(completed.error));
                    }
                    return;
                }

                // Check if still pending
                if (state.pendingRequests.has(requestInfo.requestId)) {
                    setTimeout(checkResult, 100); // Poll every 100ms
                }
            };

            checkResult();
        });
    },

    /**
     * Handles incoming response by matching requestId
     */
    handleResponse(requestId: RequestID, response: any): void {
        const isError = response.method === RPCMethod.Error;

        const pendingRequest = state.pendingRequests.get(requestId);
        if (!pendingRequest) {
            if (import.meta.env.DEV) {
                console.warn(`Received response for unknown requestId: ${requestId}`);
            }
            return;
        }

        if (pendingRequest.timeoutId) {
            window.clearTimeout(pendingRequest.timeoutId);
        }

        const completedState: RequestState = {
            status: isError ? 'failed' : 'success',
            requestInfo: pendingRequest.requestInfo,
            response: isError ? undefined : response,
            error: isError ? response?.params?.error || 'Unknown error' : undefined,
            timestamp: Date.now(),
        };

        state.completedRequests.set(requestId, completedState);
        state.pendingRequests.delete(requestId);
        this.cleanup();

        pendingRequest.resolve(response);
    },

    /**
     * Handles request timeout
     */
    handleTimeout(requestId: RequestID): void {
        const pendingRequest = state.pendingRequests.get(requestId);
        if (!pendingRequest) {
            return;
        }

        const completedState: RequestState = {
            status: 'failed',
            requestInfo: pendingRequest.requestInfo,
            error: `Request timeout after ${pendingRequest.timeout}ms`,
            timestamp: Date.now(),
        };

        state.completedRequests.set(requestId, completedState);
        state.pendingRequests.delete(requestId);
        this.cleanup();

        pendingRequest.reject(new Error(completedState.error));
    },

    /**
     * Gets the current status of a request
     */
    getRequestStatus(requestId: RequestID): 'pending' | 'success' | 'failed' | 'not_found' {
        if (state.pendingRequests.has(requestId)) {
            return 'pending';
        }

        const completed = state.completedRequests.get(requestId);
        return completed ? completed.status : 'not_found';
    },

    /**
     * Gets the completed request state
     */
    getRequestState(requestId: RequestID): RequestState | undefined {
        return state.completedRequests.get(requestId);
    },

    /**
     * Cleanup old completed requests (keep last N requests)
     */
    cleanup(): void {
        const entries = Array.from(state.completedRequests.entries());
        if (entries.length > STORE_LIMIT) {
            // Sort by timestamp and keep most recent only
            entries.sort((a, b) => b[1].timestamp - a[1].timestamp);
            const toKeep = entries.slice(0, STORE_LIMIT);

            state.completedRequests.clear();
            toKeep.forEach(([id, state]) => {
                this.state.completedRequests.set(id, state);
            });
        }
    },
};

export default RequestStore;
