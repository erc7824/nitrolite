import { Address } from 'viem';
import { RPCMethodHandler, RPCMethod, RPCErrorCode } from './types';

/**
 * Class for managing RPC method handlers
 */
export class MethodRegistry {
  private methods: Map<string, RPCMethodHandler> = new Map();
  private methodAccess: Map<string, Set<Address>> = new Map();
  
  /**
   * Register a method handler
   * @param method Method name
   * @param handler Handler function
   * @param allowedCallers Optional list of addresses allowed to call this method
   */
  register(method: string, handler: RPCMethodHandler, allowedCallers?: Address[]): void {
    // Register the method handler
    this.methods.set(method, handler);
    
    // Set up access control if provided
    if (allowedCallers) {
      const callerSet = new Set<Address>();
      for (const caller of allowedCallers) {
        callerSet.add(caller.toLowerCase() as Address);
      }
      this.methodAccess.set(method, callerSet);
    }
  }
  
  /**
   * Unregister a method handler
   * @param method Method name
   */
  unregister(method: string): void {
    this.methods.delete(method);
    this.methodAccess.delete(method);
  }
  
  /**
   * Get a method handler
   * @param method Method name
   * @param caller Caller address
   * @returns The handler function or null if not found or not allowed
   */
  getHandler(method: string, caller: Address): RPCMethodHandler | null {
    // Check if the method exists
    const handler = this.methods.get(method);
    if (!handler) {
      return null;
    }
    
    // Check access control if set up
    const allowedCallers = this.methodAccess.get(method);
    if (allowedCallers && !allowedCallers.has(caller.toLowerCase() as Address)) {
      return null;
    }
    
    return handler;
  }
  
  /**
   * Call a method handler
   * @param method Method name
   * @param params Method parameters
   * @param caller Caller address
   * @returns Promise that resolves with the result or rejects with an error
   */
  async callMethod(method: string, params: any[], caller: Address): Promise<any[]> {
    // Get the handler
    const handler = this.getHandler(method, caller);
    
    if (!handler) {
      const allowed = this.methods.has(method);
      if (allowed) {
        throw new Error(`Caller ${caller} is not authorized to call method ${method}`);
      } else {
        throw new Error(`Method ${method} not found`);
      }
    }
    
    // Call the handler
    return handler(params, caller);
  }
  
  /**
   * Get all registered method names
   * @returns Array of method names
   */
  getMethodNames(): string[] {
    return Array.from(this.methods.keys());
  }
  
  /**
   * Check if a method exists
   * @param method Method name
   * @returns True if the method exists
   */
  hasMethod(method: string): boolean {
    return this.methods.has(method);
  }
  
  /**
   * Add a caller to the allowed list for a method
   * @param method Method name
   * @param caller Caller address
   */
  allowCaller(method: string, caller: Address): void {
    let allowedCallers = this.methodAccess.get(method);
    
    if (!allowedCallers) {
      allowedCallers = new Set<Address>();
      this.methodAccess.set(method, allowedCallers);
    }
    
    allowedCallers.add(caller.toLowerCase() as Address);
  }
  
  /**
   * Remove a caller from the allowed list for a method
   * @param method Method name
   * @param caller Caller address
   */
  denyCaller(method: string, caller: Address): void {
    const allowedCallers = this.methodAccess.get(method);
    
    if (allowedCallers) {
      allowedCallers.delete(caller.toLowerCase() as Address);
    }
  }
}

/**
 * Create a registry with standard methods
 * @param handlers Map of method implementations
 * @returns A registry with standard methods
 */
export function createStandardRegistry(handlers: Partial<Record<RPCMethod, RPCMethodHandler>>): MethodRegistry {
  const registry = new MethodRegistry();
  
  // Register provided handlers
  for (const [method, handler] of Object.entries(handlers)) {
    if (handler) {
      registry.register(method, handler);
    }
  }
  
  return registry;
}
