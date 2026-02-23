import { execSync } from 'child_process';
import { 
  DEFAULTS, 
  RESERVED_NAMES, 
  VALIDATION_PATTERNS, 
  ERROR_MESSAGES 
} from '../constants/defaults.js';
import { isSafePath } from './pathResolver.js';
import { Result, createSuccess, createError, isErrorResult } from './errorHandler.js';

/**
 * Validates a project name and returns a sanitized package name
 */
export function validateProjectName(name: string): string | null {
  if (!name || name.trim() === '') {
    return null;
  }

  // Remove leading/trailing whitespace
  const trimmed = name.trim();
  
  // Check for valid characters
  if (!VALIDATION_PATTERNS.VALID_NAME.test(trimmed)) {
    return null;
  }

  // Check length (npm package name limits)
  if (trimmed.length > DEFAULTS.MAX_PACKAGE_NAME_LENGTH) {
    return null;
  }

  // Convert to lowercase for npm package name
  const packageName = trimmed.toLowerCase();
  
  // Check for reserved names
  if (RESERVED_NAMES.includes(packageName)) {
    return null;
  }

  // Check if name starts with . or _
  if (packageName.startsWith('.') || packageName.startsWith('_')) {
    return null;
  }

  return packageName;
}

/**
 * Checks if git is available on the system
 */
export async function checkGitAvailability(): Promise<boolean> {
  try {
    execSync('git --version', { stdio: 'ignore' });
    return true;
  } catch {
    return false;
  }
}

/**
 * Validates if a directory is safe to create (improved version with Result type)
 */
export function validateProjectPath(projectPath: string): Result<string> {
  // Check for safe path structure
  if (!isSafePath(projectPath)) {
    return createError(ERROR_MESSAGES.PATH_VALIDATION);
  }

  // Check for invalid characters
  if (VALIDATION_PATTERNS.INVALID_PATH_CHARS.test(projectPath)) {
    return createError(ERROR_MESSAGES.INVALID_PATH_CHARS);
  }

  return createSuccess(projectPath);
}

/**
 * Comprehensive project validation that combines name and path validation
 */
export function validateProject(input: string): Result<{
  projectName: string;
  projectPath: string;
}> {
  // Validate the project name
  const projectName = validateProjectName(input);
  if (!projectName) {
    return createError(ERROR_MESSAGES.INVALID_PROJECT_NAME);
  }

  // Validate the project path
  const pathResult = validateProjectPath(input);
  if (isErrorResult(pathResult)) {
    return createError(pathResult.error);
  }

  return createSuccess({
    projectName,
    projectPath: pathResult.data,
  });
}