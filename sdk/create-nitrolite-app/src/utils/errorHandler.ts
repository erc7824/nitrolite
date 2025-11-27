import { ERROR_MESSAGES } from '../constants/defaults.js';

/**
 * Standardized error result type
 */
export interface ErrorResult {
  success: false;
  error: string;
}

/**
 * Standardized success result type
 */
export interface SuccessResult<T = void> {
  success: true;
  data: T;
}

/**
 * Combined result type
 */
export type Result<T = void> = SuccessResult<T> | ErrorResult;

/**
 * Creates a success result
 */
export function createSuccess<T = void>(data: T): SuccessResult<T> {
  return { success: true, data };
}

/**
 * Creates an error result
 */
export function createError(error: string): ErrorResult {
  return { success: false, error };
}

/**
 * Safely extracts error message from unknown error
 */
export function getErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }
  if (typeof error === 'string') {
    return error;
  }
  return ERROR_MESSAGES.UNKNOWN_ERROR;
}

/**
 * Wraps an async function to return a standardized result
 */
export async function wrapAsync<T>(
  fn: () => Promise<T>
): Promise<Result<T>> {
  try {
    const data = await fn();
    return createSuccess(data);
  } catch (error) {
    return createError(getErrorMessage(error));
  }
}

/**
 * Wraps a sync function to return a standardized result
 */
export function wrapSync<T>(
  fn: () => T
): Result<T> {
  try {
    const data = fn();
    return createSuccess(data);
  } catch (error) {
    return createError(getErrorMessage(error));
  }
}

/**
 * Custom error class for project operations
 */
export class ProjectError extends Error {
  constructor(message: string, public readonly code?: string) {
    super(message);
    this.name = 'ProjectError';
  }
}

/**
 * Creates a formatted error for template operations
 */
export function createTemplateError(template: string): ProjectError {
  return new ProjectError(ERROR_MESSAGES.TEMPLATE_NOT_FOUND(template), 'TEMPLATE_NOT_FOUND');
}

/**
 * Creates a formatted error for git operations
 */
export function createGitError(error: string): ProjectError {
  return new ProjectError(ERROR_MESSAGES.GIT_INIT_FAILED(error), 'GIT_INIT_FAILED');
}

/**
 * Creates a formatted error for dependency installation
 */
export function createInstallError(error: string): ProjectError {
  return new ProjectError(ERROR_MESSAGES.DEPENDENCY_INSTALL_FAILED(error), 'DEPENDENCY_INSTALL_FAILED');
}

/**
 * Type guard to check if result is an error
 */
export function isErrorResult<T>(result: Result<T>): result is ErrorResult {
  return !result.success;
}