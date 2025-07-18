import { useState, useEffect } from 'react';
import { Text, Box, Newline } from 'ink';
import { WelcomeScreen } from './WelcomeScreen.js';
import { ProjectSetup } from './ProjectSetup.js';
import { TemplateSelector } from './TemplateSelector.js';
import { ProjectGenerator } from './ProjectGenerator.js';
import { CompletionScreen } from './CompletionScreen.js';
import { validateProjectName, checkGitAvailability } from '../utils/validation.js';

interface CreateNitroliteAppProps {
  projectDirectory?: string;
  template?: string;
  skipGit?: boolean;
  skipInstall?: boolean;
  skipPrompts?: boolean;
}

type Step = 'welcome' | 'setup' | 'template' | 'generate' | 'complete';

interface ProjectConfig {
  projectPath: string;
  projectName: string;
  template: string;
  initGit: boolean;
  installDeps: boolean;
  gitAvailable: boolean;
}

export default function CreateNitroliteApp({
  projectDirectory,
  template = 'react-vite',
  skipGit = false,
  skipInstall = false,
  skipPrompts = false,
}: CreateNitroliteAppProps) {
  const [step, setStep] = useState<Step>('welcome');
  const [config, setConfig] = useState<ProjectConfig>({
    projectPath: projectDirectory || '',
    projectName: '',
    template,
    initGit: !skipGit,
    installDeps: !skipInstall,
    gitAvailable: false,
  });
  const [error, setError] = useState<string>('');

  useEffect(() => {
    // Check git availability on startup
    checkGitAvailability().then((available) => {
      setConfig((prev) => ({ ...prev, gitAvailable: available }));
    });
  }, []);

  // If all required options are provided via CLI, skip to generation
  useEffect(() => {
    if (skipPrompts && projectDirectory) {
      const projectName = validateProjectName(projectDirectory);
      if (projectName) {
        setConfig((prev) => ({
          ...prev,
          projectName,
          projectPath: projectDirectory,
        }));
        setStep('generate');
      } else {
        setError('Invalid project directory name');
      }
    }
  }, [skipPrompts, projectDirectory]);

  const handleWelcomeComplete = () => {
    setStep('setup');
  };

  const handleSetupComplete = (setupConfig: Partial<ProjectConfig>) => {
    setConfig((prev) => ({ ...prev, ...setupConfig }));
    setStep('template');
  };

  const handleTemplateSelected = (selectedTemplate: string) => {
    setConfig((prev) => ({ ...prev, template: selectedTemplate }));
    setStep('generate');
  };

  const handleGenerationComplete = () => {
    setStep('complete');
  };

  const handleError = (errorMessage: string) => {
    setError(errorMessage);
  };

  if (error) {
    return (
      <Box flexDirection="column" padding={1}>
        <Text color="red">❌ Error: {error}</Text>
        <Newline />
        <Text color="gray">Please try again with different options.</Text>
      </Box>
    );
  }

  const renderCurrentStep = () => {
    switch (step) {
      case 'welcome':
        return <WelcomeScreen onComplete={handleWelcomeComplete} />;

      case 'setup':
        return (
          <ProjectSetup
            initialPath={config.projectPath}
            gitAvailable={config.gitAvailable}
            onComplete={handleSetupComplete}
            onError={handleError}
          />
        );

      case 'template':
        return <TemplateSelector currentTemplate={config.template} onSelect={handleTemplateSelected} />;

      case 'generate':
        return <ProjectGenerator config={config} onComplete={handleGenerationComplete} onError={handleError} />;

      case 'complete':
        return <CompletionScreen config={config} />;

      default:
        return <Text>Unknown step</Text>;
    }
  };

  // Show welcome screen once, then keep it persistent with prompts below
  if (step === 'welcome') {
    return renderCurrentStep();
  }

  return (
    <Box flexDirection="column">
      <WelcomeScreen onComplete={() => { }} interactive={false} />
      {renderCurrentStep()}
    </Box>
  );
}
