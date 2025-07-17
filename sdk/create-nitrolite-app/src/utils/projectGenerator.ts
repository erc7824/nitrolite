import fs from 'fs-extra';
import path from 'path';
import { execSync } from 'child_process';
import mustache from 'mustache';
import { fileURLToPath } from 'url';
import { dirname } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

interface ProjectConfig {
  projectPath: string;
  projectName: string;
  template: string;
  initGit: boolean;
  installDeps: boolean;
  gitAvailable: boolean;
}

interface GenerationCallbacks {
  onStep: (step: 'copying' | 'templating' | 'git' | 'installing' | 'complete') => void;
  onProgress: (percent: number) => void;
  onError: (error: string) => void;
}

const SKIP_FILES = [
  'node_modules',
  '.git',
  '.next',
  'dist',
  'build',
  '.template.json',
  '.DS_Store',
  'Thumbs.db'
];

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
    const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
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
  const templatePath = path.join(__dirname, '../../templates', template);
  const targetPath = path.resolve(process.cwd(), projectPath);
  
  // Check if template exists
  if (!fs.existsSync(templatePath)) {
    throw new Error(`Template "${template}" not found`);
  }
  
  // Create target directory
  await fs.ensureDir(targetPath);
  
  // Get all files recursively
  const files = await getAllFiles(templatePath);
  const totalFiles = files.length;
  
  // Copy files with progress updates
  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    const relativePath = path.relative(templatePath, file);
    const targetFilePath = path.join(targetPath, relativePath);
    
    // Skip certain files
    if (SKIP_FILES.some(skip => relativePath.includes(skip))) {
      continue;
    }
    
    // Ensure target directory exists
    await fs.ensureDir(path.dirname(targetFilePath));
    
    // Copy file
    await fs.copy(file, targetFilePath);
    
    // Update progress
    const progress = Math.round(((i + 1) / totalFiles) * 100);
    onProgress(progress);
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
  const targetPath = path.resolve(process.cwd(), projectPath);
  const templateVariables = {
    projectName,
    packageName: projectName.toLowerCase(),
    year: new Date().getFullYear(),
    date: new Date().toISOString().split('T')[0]
  };
  
  // Get all files that might contain template variables
  const templateFiles = await getAllFiles(targetPath);
  const filesToProcess = templateFiles.filter(file => {
    const ext = path.extname(file);
    return ['.json', '.md', '.ts', '.tsx', '.js', '.jsx', '.vue', '.html'].includes(ext);
  });
  
  const totalFiles = filesToProcess.length;
  
  for (let i = 0; i < filesToProcess.length; i++) {
    const file = filesToProcess[i];
    
    try {
      // Read file content
      const content = await fs.readFile(file, 'utf-8');
      
      // Process mustache templates
      const processedContent = mustache.render(content, templateVariables);
      
      // Write back to file
      await fs.writeFile(file, processedContent, 'utf-8');
      
      // Update progress
      const progress = Math.round(((i + 1) / totalFiles) * 100);
      onProgress(progress);
    } catch (error) {
      // Skip files that can't be processed
      console.warn(`Could not process template variables in ${file}:`, error);
    }
  }
}

/**
 * Initialize git repository
 */
async function initializeGitRepository(
  projectPath: string,
  onProgress: (percent: number) => void
): Promise<void> {
  const targetPath = path.resolve(process.cwd(), projectPath);
  
  try {
    // Initialize git repository
    onProgress(25);
    execSync('git init', { cwd: targetPath, stdio: 'ignore' });
    
    // Create .gitignore if it doesn't exist
    onProgress(50);
    const gitignorePath = path.join(targetPath, '.gitignore');
    if (!fs.existsSync(gitignorePath)) {
      const gitignoreContent = `
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
      
      await fs.writeFile(gitignorePath, gitignoreContent, 'utf-8');
    }
    
    // Add all files
    onProgress(75);
    execSync('git add -A', { cwd: targetPath, stdio: 'ignore' });
    
    // Create initial commit
    onProgress(100);
    execSync('git commit -m "Initial commit from create-nitrolite-app"', { 
      cwd: targetPath, 
      stdio: 'ignore' 
    });
    
  } catch (error) {
    throw new Error(`Failed to initialize git repository: ${error instanceof Error ? error.message : 'Unknown error'}`);
  }
}

/**
 * Install dependencies
 */
async function installDependencies(
  projectPath: string,
  onProgress: (percent: number) => void
): Promise<void> {
  const targetPath = path.resolve(process.cwd(), projectPath);
  
  try {
    onProgress(25);
    
    // Check if package.json exists
    const packageJsonPath = path.join(targetPath, 'package.json');
    if (!fs.existsSync(packageJsonPath)) {
      throw new Error('No package.json found in project');
    }
    
    onProgress(50);
    
    // Install dependencies
    execSync('npm install', { 
      cwd: targetPath, 
      stdio: 'ignore'
    });
    
    onProgress(100);
    
  } catch (error) {
    throw new Error(`Failed to install dependencies: ${error instanceof Error ? error.message : 'Unknown error'}`);
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