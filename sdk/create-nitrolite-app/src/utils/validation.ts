import { execSync } from 'child_process';

/**
 * Validates a project name and returns a sanitized package name
 */
export function validateProjectName(name: string): string | null {
  if (!name || name.trim() === '') {
    return null;
  }

  // Remove leading/trailing whitespace
  const trimmed = name.trim();
  
  // Check for valid characters (letters, numbers, hyphens, underscores)
  const validNameRegex = /^[a-zA-Z0-9_-]+$/;
  if (!validNameRegex.test(trimmed)) {
    return null;
  }

  // Check length (npm package name limits)
  if (trimmed.length > 214) {
    return null;
  }

  // Convert to lowercase for npm package name
  const packageName = trimmed.toLowerCase();
  
  // Check for reserved names
  const reservedNames = [
    'node_modules',
    'favicon.ico',
    'package',
    'package.json',
    'npm',
    'con',
    'prn',
    'aux',
    'nul',
    'com1',
    'com2',
    'com3',
    'com4',
    'com5',
    'com6',
    'com7',
    'com8',
    'com9',
    'lpt1',
    'lpt2',
    'lpt3',
    'lpt4',
    'lpt5',
    'lpt6',
    'lpt7',
    'lpt8',
    'lpt9'
  ];
  
  if (reservedNames.includes(packageName)) {
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
 * Validates if a directory is safe to create
 */
export function validateProjectPath(projectPath: string): {
  isValid: boolean;
  error?: string;
} {
  // Check for absolute paths
  if (projectPath.startsWith('/') || projectPath.includes('..')) {
    return {
      isValid: false,
      error: 'Project path must be relative and cannot contain ".."'
    };
  }

  // Check for invalid characters
  const invalidChars = /[<>:"|?*]/;
  if (invalidChars.test(projectPath)) {
    return {
      isValid: false,
      error: 'Project path contains invalid characters'
    };
  }

  return { isValid: true };
}