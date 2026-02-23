import path from 'path';
import { fileURLToPath } from 'url';
import { dirname } from 'path';

// Get current directory for ES modules
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

/**
 * Resolves a project path relative to the current working directory
 */
export function resolveProjectPath(projectPath: string): string {
  return path.resolve(process.cwd(), projectPath);
}

/**
 * Resolves the template directory path
 */
export function resolveTemplatePath(template: string): string {
  return path.join(__dirname, '../../templates', template);
}

/**
 * Gets the relative path from source to target
 */
export function getRelativePath(sourcePath: string, targetPath: string): string {
  return path.relative(sourcePath, targetPath);
}

/**
 * Resolves a file path within a project directory
 */
export function resolveProjectFile(projectPath: string, filePath: string): string {
  const projectDir = resolveProjectPath(projectPath);
  return path.join(projectDir, filePath);
}

/**
 * Gets the current directory path for ES modules
 */
export function getCurrentDirname(): string {
  return __dirname;
}

/**
 * Checks if a path is safe (relative and doesn't contain ..)
 */
export function isSafePath(projectPath: string): boolean {
  return !projectPath.startsWith('/') && !projectPath.includes('..');
}