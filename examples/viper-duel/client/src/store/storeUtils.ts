import { useState, useEffect } from 'react';

/**
 * Simple store implementation for state management
 */
export class Store<T> {
  private state: T;
  private listeners: (() => void)[] = [];

  constructor(initialState: T) {
    this.state = initialState;
  }

  /**
   * Get the current state
   */
  getState(): T {
    return this.state;
  }

  /**
   * Update the state
   */
  setState(newState: Partial<T>): void {
    this.state = { ...this.state, ...newState };
    this.notifyListeners();
  }

  /**
   * Set a specific key in the state
   */
  set<K extends keyof T>(key: K, value: T[K]): void {
    this.state = { ...this.state, [key]: value };
    this.notifyListeners();
  }

  /**
   * Subscribe to changes
   */
  subscribe(listener: () => void): () => void {
    this.listeners.push(listener);
    return () => {
      this.listeners = this.listeners.filter(l => l !== listener);
    };
  }

  /**
   * Notify all listeners of state changes
   */
  private notifyListeners(): void {
    this.listeners.forEach(listener => listener());
  }
}

/**
 * Custom hook to use a store
 */
export function useStore<T>(store: Store<T>): T {
  const [state, setState] = useState<T>(store.getState());

  useEffect(() => {
    const unsubscribe = store.subscribe(() => {
      setState(store.getState());
    });
    return unsubscribe;
  }, [store]);

  return state;
}