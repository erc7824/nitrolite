import fs from 'fs-extra';
import path from 'path';
import { execSync } from 'child_process';
import mustache from 'mustache';
import { SDK_VERSION } from '../constants/version.js';
import { ProjectConfig, GenerationStep } from '../types/index.js';
import { 
  DEFAULTS, 
  SKIP_FILES, 
  TEMPLATE_EXTENSIONS,
  ERROR_MESSAGES 
} from '../constants/defaults.js';
import { 
  resolveProjectPath, 
  resolveTemplatePath, 
  getRelativePath,
  resolveProjectFile 
} from './pathResolver.js';
import { createProgressUpdater } from './progressCalculator.js';
import { 
  wrapAsync, 
  createTemplateError, 
  createGitError, 
  createInstallError,
  getErrorMessage,
  isErrorResult 
} from './errorHandler.js';

interface GenerationCallbacks {
  onStep: (step: GenerationStep) => void;
  onProgress: (percent: number) => void;
  onError: (error: string) => void;
}

// Git ignore template content
const DEFAULT_GITIGNORE = `
# Dependencies
node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*

# Production builds
/dist
/build
/.next

# Environment variables
.env
.env.local
.env.development.local
.env.test.local
.env.production.local

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Logs
*.log
`.trim();

/**
 * Main project generation function
 */
export async function generateProject(
  config: ProjectConfig,
  callbacks: GenerationCallbacks
): Promise<void> {
  const { projectPath, projectName, template, initGit, installDeps, gitAvailable } = config;
  
  try {
    // Step 1: Copy template files
    callbacks.onStep('copying');
    await copyTemplateFiles(projectPath, template, callbacks.onProgress);
    
    // Step 2: Process template variables
    callbacks.onStep('templating');
    await processTemplateVariables(projectPath, projectName, callbacks.onProgress);
    
    // Step 3: Initialize git repository
    if (initGit && gitAvailable) {
      callbacks.onStep('git');
      await initializeGitRepository(projectPath, callbacks.onProgress);
    }
    
    // Step 4: Install dependencies
    if (installDeps) {
      callbacks.onStep('installing');
      await installDependencies(projectPath, callbacks.onProgress);
    }
    
    callbacks.onStep('complete');
    
  } catch (error) {
    const errorMessage = getErrorMessage(error);
    callbacks.onError(errorMessage);
    throw error;
  }
}

/**
 * Copy template files to the project directory
 */
async function copyTemplateFiles(
  projectPath: string,
  template: string,
  onProgress: (percent: number) => void
): Promise<void> {
  const templatePath = resolveTemplatePath(template);
  const targetPath = resolveProjectPath(projectPath);
  
  // Check if template exists
  if (!fs.existsSync(templatePath)) {
    throw createTemplateError(template);
  }
  
  // Create target directory
  await fs.ensureDir(targetPath);
  
  // Get all files recursively
  const files = await getAllFiles(templatePath);
  const updateProgress = createProgressUpdater(files.length, onProgress);
  
  // Copy files with progress updates
  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    const relativePath = getRelativePath(templatePath, file);
    const targetFilePath = path.join(targetPath, relativePath);
    
    // Skip certain files
    if (shouldSkipFile(relativePath)) {
      continue;
    }
    
    // Ensure target directory exists
    await fs.ensureDir(path.dirname(targetFilePath));
    
    // Copy file
    await fs.copy(file, targetFilePath);
    
    // Update progress
    updateProgress(i);
  }
}

/**
 * Process template variables using mustache
 */
async function processTemplateVariables(
  projectPath: string,
  projectName: string,
  onProgress: (percent: number) => void
): Promise<void> {
  const targetPath = resolveProjectPath(projectPath);
  
  const templateVariables = {
    projectName,
    packageName: projectName.toLowerCase(),
    year: new Date().getFullYear(),
    date: new Date().toISOString().split('T')[0],
    sdkVersion: SDK_VERSION
  };
  
  // Get all files that might contain template variables
  const templateFiles = await getAllFiles(targetPath);
  const filesToProcess = templateFiles.filter(shouldProcessFile);
  const updateProgress = createProgressUpdater(filesToProcess.length, onProgress);
  
  for (let i = 0; i < filesToProcess.length; i++) {
    const file = filesToProcess[i];
    
    const result = await wrapAsync(async () => {
      const content = await fs.readFile(file, 'utf-8');
      const processedContent = mustache.render(content, templateVariables);
      await fs.writeFile(file, processedContent, 'utf-8');
    });
    
    if (isErrorResult(result)) {
      // Skip files that can't be processed but log the warning
      console.warn(`Could not process template variables in ${file}:`, result.error);
    }
    
    updateProgress(i);
  }
}

/**
 * Initialize git repository
 */
async function initializeGitRepository(
  projectPath: string,
  onProgress: (percent: number) => void
): Promise<void> {
  const targetPath = resolveProjectPath(projectPath);
  
  try {
    // Initialize git repository
    onProgress(DEFAULTS.PROGRESS.QUARTER);
    execSync('git init', { cwd: targetPath, stdio: 'ignore' });
    
    // Create .gitignore if it doesn't exist
    onProgress(DEFAULTS.PROGRESS.HALF);
    await ensureGitignoreExists(targetPath);
    
    // Add all files
    onProgress(DEFAULTS.PROGRESS.THREE_QUARTERS);
    execSync('git add -A', { cwd: targetPath, stdio: 'ignore' });
    
    // Create initial commit
    onProgress(DEFAULTS.PROGRESS.COMPLETE);
    execSync('git commit -m "Initial commit from create-nitrolite-app"', { 
      cwd: targetPath, 
      stdio: 'ignore' 
    });
    
  } catch (error) {
    throw createGitError(getErrorMessage(error));
  }
}

/**
 * Install dependencies
 */
async function installDependencies(
  projectPath: string,
  onProgress: (percent: number) => void
): Promise<void> {
  const targetPath = resolveProjectPath(projectPath);
  
  try {
    onProgress(DEFAULTS.PROGRESS.QUARTER);
    
    // Check if package.json exists
    const packageJsonPath = resolveProjectFile(projectPath, 'package.json');
    if (!fs.existsSync(packageJsonPath)) {
      throw new Error(ERROR_MESSAGES.NO_PACKAGE_JSON);
    }
    
    onProgress(DEFAULTS.PROGRESS.HALF);
    
    // Install dependencies
    execSync('npm install', { 
      cwd: targetPath, 
      stdio: 'ignore'
    });
    
    onProgress(DEFAULTS.PROGRESS.COMPLETE);
    
  } catch (error) {
    throw createInstallError(getErrorMessage(error));
  }
}

/**
 * Get all files recursively from a directory
 */
async function getAllFiles(dir: string): Promise<string[]> {
  const files: string[] = [];
  
  const items = await fs.readdir(dir, { withFileTypes: true });
  
  for (const item of items) {
    const fullPath = path.join(dir, item.name);
    
    if (item.isDirectory()) {
      if (!SKIP_FILES.includes(item.name)) {
        const subFiles = await getAllFiles(fullPath);
        files.push(...subFiles);
      }
    } else {
      files.push(fullPath);
    }
  }
  
  return files;
}

/**
 * Checks if a file should be skipped during copying
 */
function shouldSkipFile(relativePath: string): boolean {
  return SKIP_FILES.some(skip => relativePath.includes(skip));
}

/**
 * Checks if a file should be processed for template variables
 */
function shouldProcessFile(file: string): boolean {
  const ext = path.extname(file);
  return TEMPLATE_EXTENSIONS.includes(ext);
}

/**
 * Ensures .gitignore file exists with default content
 */
async function ensureGitignoreExists(targetPath: string): Promise<void> {
  const gitignorePath = path.join(targetPath, '.gitignore');
  if (!fs.existsSync(gitignorePath)) {
    await fs.writeFile(gitignorePath, DEFAULT_GITIGNORE, 'utf-8');
  }
}