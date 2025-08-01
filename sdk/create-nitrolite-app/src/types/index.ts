export interface ProjectConfig {
  projectPath: string;
  projectName: string;
  template: string;
  initGit: boolean;
  installDeps: boolean;
  gitAvailable: boolean;
}

export type GenerationStep = 'copying' | 'templating' | 'git' | 'installing' | 'complete';

export type SetupStep = 'path' | 'git' | 'template';

export type AppStep = 'welcome' | 'setup' | 'confirmation' | 'generate' | 'complete';