import { useState, useEffect } from 'react';
import { Text, Box, Newline } from 'ink';
import { useInput } from 'ink';
import path from 'path';
import fs from 'fs-extra';
import { validateProjectName } from '../utils/validation.js';
import { TemplateSelector } from './TemplateSelector.js';
import { ProjectConfig, SetupStep } from '../types/index.js';

interface ProjectSetupProps {
  initialPath: string;
  gitAvailable: boolean;
  onComplete: (config: { projectPath: string; projectName: string; initGit: boolean; template: string }) => void;
  onError: (error: string) => void;
}

export function ProjectSetup({ initialPath, gitAvailable, onComplete, onError }: ProjectSetupProps) {
  const [step, setStep] = useState<SetupStep>('path');
  const [projectPath, setProjectPath] = useState(initialPath || '');
  const [projectName, setProjectName] = useState('');
  const [initGit, setInitGit] = useState(gitAvailable);
  const [template, setTemplate] = useState('nextjs-app');
  const [inputBuffer, setInputBuffer] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    if (initialPath) {
      const name = validateProjectName(initialPath);
      if (name) {
        setProjectName(name);
        setStep(gitAvailable ? 'git' : 'template');
      }
    }
  }, [initialPath, gitAvailable]);

  useInput((input, key) => {
    if (key.ctrl && input === 'c') {
      process.exit(0);
    } else if (step === 'path') {
      if (key.return) {
        handleEnter();
      } else if (key.backspace || key.delete) {
        setInputBuffer((prev) => prev.slice(0, -1));
      } else if (input && !key.ctrl) {
        setInputBuffer((prev) => prev + input);
      }
    } else if (step === 'git' && (input === 'y' || input === 'n')) {
      setInitGit(input === 'y');
      setStep('template');
    }
  });

  const handleEnter = () => {
    if (step === 'path') {
      const pathInput = inputBuffer.trim() || 'my-nitrolite-app';
      const validatedName = validateProjectName(pathInput);

      if (!validatedName) {
        setError('Invalid project name. Use only letters, numbers, hyphens, and underscores.');
        return;
      }

      const fullPath = path.resolve(process.cwd(), pathInput);

      if (fs.existsSync(fullPath)) {
        setError(`Directory "${pathInput}" already exists. Please choose a different name.`);
        return;
      }

      setProjectPath(pathInput);
      setProjectName(validatedName);
      setError('');
      setStep(gitAvailable ? 'git' : 'template');
    }
  };

  const handleTemplateSelect = (selectedTemplate: string) => {
    setTemplate(selectedTemplate);
    onComplete({
      projectPath,
      projectName,
      initGit,
      template: selectedTemplate,
    });
  };

  const handleComplete = () => {
    onComplete({
      projectPath,
      projectName,
      initGit,
      template,
    });
  };

  const renderPathInput = () => (
    <Box flexDirection="column">
      <Text color="cyan">📁 Project Setup</Text>
      <Newline />
      <Text>What is your project directory name?</Text>
      <Text color="gray">(Press Enter to use default: my-nitrolite-app)</Text>
      <Newline />
      <Box>
        <Text color="green">❯ </Text>
        <Text>{inputBuffer}</Text>
        <Text color="gray">█</Text>
      </Box>
      {error && (
        <>
          <Newline />
          <Text color="red">❌ {error}</Text>
        </>
      )}
      <Newline />
      <Text color="gray">
        Press <Text color="white">Ctrl+C</Text> to exit
      </Text>
    </Box>
  );

  const renderGitInput = () => (
    <Box flexDirection="column">
      <Text color="cyan">🔧 Git Configuration</Text>
      <Newline />
      <Text>Initialize a git repository?</Text>
      <Text color="gray">(Git is available on your system)</Text>
      <Newline />
      <Box>
        <Text color="green">❯ </Text>
        <Text>{initGit ? 'Yes' : 'No'}</Text>
        <Text color="gray"> (y/n)</Text>
      </Box>
      <Newline />
      <Text color="gray">
        Press <Text color="white">y</Text> for yes, <Text color="white">n</Text> for no
      </Text>
    </Box>
  );

  switch (step) {
    case 'path':
      return renderPathInput();
    case 'git':
      return renderGitInput();
    case 'template':
      return <TemplateSelector currentTemplate={template} onSelect={handleTemplateSelect} />;
    default:
      return <Text>Unknown step</Text>;
  }
}
