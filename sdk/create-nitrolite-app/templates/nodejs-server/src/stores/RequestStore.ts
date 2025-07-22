import type { RequestInfo } from '../types/index.js';
import { logger } from '../utils/logger.js';

interface PendingRequest {
  requestInfo: RequestInfo;
  resolve: (response: any) => void;
  reject: (error: Error) => void;
  timeout?: NodeJS.Timeout;
}

class RequestStoreClass {
  private pendingRequests = new Map<string, PendingRequest>();

  registerRequest(
    requestInfo: RequestInfo,
    options: { timeout?: number } = {}
  ): Promise<any> {
    const { timeout = 30000 } = options;

    return new Promise((resolve, reject) => {
      const timeoutId = setTimeout(() => {
        this.pendingRequests.delete(requestInfo.requestId);
        reject(new Error(`Request ${requestInfo.requestId} timed out after ${timeout}ms`));
      }, timeout);

      this.pendingRequests.set(requestInfo.requestId, {
        requestInfo,
        resolve,
        reject,
        timeout: timeoutId,
      });

      logger.debug(`Request registered: ${requestInfo.method} [ID: ${requestInfo.requestId}]`);
    });
  }

  handleResponse(requestId: string, response: any): boolean {
    const pendingRequest = this.pendingRequests.get(requestId);
    
    if (!pendingRequest) {
      logger.warn(`Request not found [ID: ${requestId}]`);
      return false;
    }

    this.pendingRequests.delete(requestId);
    
    if (pendingRequest.timeout) {
      clearTimeout(pendingRequest.timeout);
    }

    logger.debug(`Request response received [ID: ${requestId}]`);
    pendingRequest.resolve(response);
    return true;
  }

  handleError(requestId: string, error: Error): boolean {
    const pendingRequest = this.pendingRequests.get(requestId);
    
    if (!pendingRequest) {
      logger.warn(`Request not found for error handling [ID: ${requestId}]`);
      return false;
    }

    this.pendingRequests.delete(requestId);
    
    if (pendingRequest.timeout) {
      clearTimeout(pendingRequest.timeout);
    }

    logger.debug(`Request error handled [ID: ${requestId}]: ${error.message}`);
    pendingRequest.reject(error);
    return true;
  }

  getPendingRequestsCount(): number {
    return this.pendingRequests.size;
  }

  getPendingRequestIds(): string[] {
    return Array.from(this.pendingRequests.keys());
  }

  clearRequest(requestId: string): boolean {
    const pendingRequest = this.pendingRequests.get(requestId);
    
    if (!pendingRequest) {
      return false;
    }

    this.pendingRequests.delete(requestId);
    
    if (pendingRequest.timeout) {
      clearTimeout(pendingRequest.timeout);
    }

    pendingRequest.reject(new Error(`Request ${requestId} was cancelled`));
    return true;
  }

  clearAllRequests(): void {
    for (const [requestId, pendingRequest] of this.pendingRequests) {
      if (pendingRequest.timeout) {
        clearTimeout(pendingRequest.timeout);
      }
      pendingRequest.reject(new Error('All requests cleared'));
    }
    
    this.pendingRequests.clear();
    logger.info(`Cleared ${this.pendingRequests.size} pending requests`);
  }
}

const RequestStore = new RequestStoreClass();

export default RequestStore;
export type { RequestInfo };