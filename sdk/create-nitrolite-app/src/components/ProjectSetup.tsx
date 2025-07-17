import React, { useState, useEffect } from 'react';
import { Text, Box, Newline } from 'ink';
import { useInput } from 'ink';
import path from 'path';
import fs from 'fs-extra';
import { validateProjectName } from '../utils/validation.js';

interface ProjectSetupProps {
  initialPath: string;
  gitAvailable: boolean;
  onComplete: (config: {
    projectPath: string;
    projectName: string;
    initGit: boolean;
  }) => void;
  onError: (error: string) => void;
}

type SetupStep = 'path' | 'git' | 'confirm';

export function ProjectSetup({ 
  initialPath, 
  gitAvailable, 
  onComplete, 
  onError 
}: ProjectSetupProps) {
  const [step, setStep] = useState<SetupStep>('path');
  const [projectPath, setProjectPath] = useState(initialPath || '');
  const [projectName, setProjectName] = useState('');
  const [initGit, setInitGit] = useState(gitAvailable);
  const [inputBuffer, setInputBuffer] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    if (initialPath) {
      const name = validateProjectName(initialPath);
      if (name) {
        setProjectName(name);
        setStep(gitAvailable ? 'git' : 'confirm');
      }
    }
  }, [initialPath, gitAvailable]);

  useInput((input, key) => {
    if (key.return) {
      handleEnter();
    } else if (key.backspace || key.delete) {
      setInputBuffer(prev => prev.slice(0, -1));
    } else if (key.ctrl && input === 'c') {
      process.exit(0);
    } else if (step === 'git' && (input === 'y' || input === 'n')) {
      setInitGit(input === 'y');
      setStep('confirm');
    } else if (step === 'confirm' && (input === 'y' || input === 'n')) {
      if (input === 'y') {
        handleComplete();
      } else {
        // Go back to path input
        setStep('path');
        setInputBuffer('');
        setError('');
      }
    } else if (step === 'path' && input && !key.ctrl) {
      setInputBuffer(prev => prev + input);
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
      setStep(gitAvailable ? 'git' : 'confirm');
    }
  };

  const handleComplete = () => {
    onComplete({
      projectPath,
      projectName,
      initGit
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
      <Text color="gray">Press <Text color="white">Ctrl+C</Text> to exit</Text>
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
      <Text color="gray">Press <Text color="white">y</Text> for yes, <Text color="white">n</Text> for no</Text>
    </Box>
  );

  const renderConfirmation = () => (
    <Box flexDirection="column">
      <Text color="cyan">✅ Confirm Project Setup</Text>
      <Newline />
      <Text>Project directory: <Text color="green">{projectPath}</Text></Text>
      <Text>Package name: <Text color="green">{projectName}</Text></Text>
      <Text>Initialize git: <Text color={initGit ? 'green' : 'red'}>{initGit ? 'Yes' : 'No'}</Text></Text>
      <Newline />
      <Text>Create project with these settings?</Text>
      <Newline />
      <Box>
        <Text color="green">❯ </Text>
        <Text color="gray">(y/n)</Text>
      </Box>
      <Newline />
      <Text color="gray">Press <Text color="white">y</Text> to continue, <Text color="white">n</Text> to go back</Text>
    </Box>
  );

  switch (step) {
    case 'path':
      return renderPathInput();
    case 'git':
      return renderGitInput();
    case 'confirm':
      return renderConfirmation();
    default:
      return <Text>Unknown step</Text>;
  }
}