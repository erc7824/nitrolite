import { useState, useEffect } from 'react';
import { Text, Box, Newline } from 'ink';
import { WelcomeScreen } from './WelcomeScreen.js';
import { ProjectSetup } from './ProjectSetup.js';
import { ConfirmationScreen } from './ConfirmationScreen.js';
import { ProjectGenerator } from './ProjectGenerator.js';
import { CompletionScreen } from './CompletionScreen.js';
import { validateProjectName, checkGitAvailability } from '../utils/validation.js';
import { ProjectConfig, AppStep } from '../types/index.js';

interface CreateNitroliteAppProps {
  projectDirectory?: string;
  template?: string;
  skipGit?: boolean;
  skipInstall?: boolean;
  skipPrompts?: boolean;
}

export default function CreateNitroliteApp({
  projectDirectory,
  template = 'nextjs-app',
  skipGit = false,
  skipInstall = false,
  skipPrompts = false,
}: CreateNitroliteAppProps) {
  const [step, setStep] = useState<AppStep>('welcome');
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
    setStep('confirmation');
  };

  const handleConfirmationComplete = () => {
    setStep('generate');
  };

  const handleConfirmationCancel = () => {
    setStep('setup');
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

      case 'confirmation':
        return (
          <ConfirmationScreen
            config={config}
            onConfirm={handleConfirmationComplete}
            onCancel={handleConfirmationCancel}
          />
        );

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
