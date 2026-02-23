// Default values and magic numbers centralized

export const DEFAULTS = {
  // Project defaults
  PROJECT_NAME: 'my-nitrolite-app',
  TEMPLATE: 'nextjs-app',
  
  // Timeouts
  WELCOME_TIMEOUT: 1000,
  COMPLETION_TIMEOUT: 1000,
  
  // Progress percentages
  PROGRESS: {
    QUARTER: 25,
    HALF: 50,
    THREE_QUARTERS: 75,
    COMPLETE: 100,
  },
  
  // UI dimensions
  PROGRESS_BAR_WIDTH: 30,
  BORDER_PADDING: 4,
  BORDER_THICKNESS: 2,
  COMPONENT_PADDING: 1,
  INDENT_PADDING: 2,
  
  // Validation limits
  MAX_PACKAGE_NAME_LENGTH: 214,
} as const;

// Step configurations
export const STEP_CONFIG = {
  ORDER: {
    copying: 0,
    templating: 1,
    git: 2,
    installing: 3,
    complete: 4,
  },
  ICONS: {
    ERROR: '❌',
    LOADING: '⏳', 
    COMPLETE: '✅',
    PENDING: '⏸️',
  },
  MESSAGES: {
    copying: 'Copying template files',
    templating: 'Processing template variables',
    git: 'Initializing git repository',
    installing: 'Installing dependencies', 
    complete: 'Project created successfully!',
  },
} as const;

// Error messages
export const ERROR_MESSAGES = {
  INVALID_PROJECT_NAME: 'Invalid project name. Use only letters, numbers, hyphens, and underscores.',
  DIRECTORY_EXISTS: (name: string) => `Directory "${name}" already exists. Please choose a different name.`,
  TEMPLATE_NOT_FOUND: (template: string) => `Template "${template}" not found`,
  GIT_INIT_FAILED: (error: string) => `Failed to initialize git repository: ${error}`,
  DEPENDENCY_INSTALL_FAILED: (error: string) => `Failed to install dependencies: ${error}`,
  NO_PACKAGE_JSON: 'No package.json found in project',
  UNKNOWN_ERROR: 'Unknown error occurred',
  PATH_VALIDATION: 'Project path must be relative and cannot contain ".."',
  INVALID_PATH_CHARS: 'Project path contains invalid characters',
} as const;

// UI text constants
export const UI_TEXT = {
  WELCOME: {
    TITLE: '🚀 Welcome to Nitrolite!',
    SUBTITLE: 'The fastest way to create Nitrolite applications',
    INSTRUCTION: 'Press Enter or Space to continue...',
  },
  PROJECT_SETUP: {
    TITLE: '📁 Project Setup',
    DIRECTORY_PROMPT: 'What is your project directory name?',
    DEFAULT_HINT: '(Press Enter to use default: my-nitrolite-app)',
  },
  GIT_CONFIG: {
    TITLE: '🔧 Git Configuration', 
    PROMPT: 'Initialize a git repository?',
    HINT: '(Git is available on your system)',
    INSTRUCTIONS: 'Press y for yes, n for no',
  },
  GENERATION: {
    TITLE: '🚀 Generating Project',
    CREATING: 'Creating',
    WITH_TEMPLATE: 'with',
    TEMPLATE_SUFFIX: 'template...',
    SUCCESS: '✨ Project created successfully!',
  },
  NAVIGATION: {
    EXIT: 'Press Ctrl+C to exit',
    ARROWS: 'Use ↑↓ arrows to navigate, Enter to select',
    YES_NO: '(y/n)',
    PROMPT: '❯ ',
    CURSOR: '█',
  },
} as const;

// File skip patterns
export const SKIP_FILES = [
  'node_modules',
  '.git',
  '.next',
  'dist',
  'build',
  '.template.json',
  '.DS_Store',
  'Thumbs.db',
];

// Template file extensions to process
export const TEMPLATE_EXTENSIONS = [
  '.json',
  '.md', 
  '.ts',
  '.tsx',
  '.js',
  '.jsx',
  '.vue',
  '.html',
];

// Reserved package names (extracted from validation.ts)
export const RESERVED_NAMES = [
  'node_modules',
  'favicon.ico', 
  'package',
  'package.json',
  'npm',
  'con',
  'prn',
  'aux',
  'nul',
  'com1', 'com2', 'com3', 'com4', 'com5', 'com6', 'com7', 'com8', 'com9',
  'lpt1', 'lpt2', 'lpt3', 'lpt4', 'lpt5', 'lpt6', 'lpt7', 'lpt8', 'lpt9',
];

// Validation patterns
export const VALIDATION_PATTERNS = {
  VALID_NAME: /^[a-zA-Z0-9_-]+$/,
  INVALID_PATH_CHARS: /[<>:"|?*]/,
} as const;