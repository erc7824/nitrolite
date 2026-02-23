import { useState, useEffect } from 'react';
import { Text, Box, Newline } from 'ink';
import { generateProject } from '../utils/projectGenerator.js';
import { ProjectConfig, GenerationStep } from '../types/index.js';

interface ProjectGeneratorProps {
  config: ProjectConfig;
  onComplete: () => void;
  onError: (error: string) => void;
}

export function ProjectGenerator({ config, onComplete, onError }: ProjectGeneratorProps) {
  const [currentStep, setCurrentStep] = useState<GenerationStep>('copying');
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState('');

  useEffect(() => {
    const runGeneration = async () => {
      try {
        await generateProject(config, {
          onStep: (step) => {
            setCurrentStep(step);
            setProgress(0);
          },
          onProgress: (percent) => {
            setProgress(percent);
          },
          onError: (err) => {
            setError(err);
            onError(err);
          },
        });

        setCurrentStep('complete');
        setTimeout(onComplete, 1000);
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : 'Unknown error occurred';
        setError(errorMessage);
        onError(errorMessage);
      }
    };

    runGeneration();
  }, [config, onComplete, onError]);

  const getStepIcon = (step: GenerationStep) => {
    if (error) return '❌';
    if (step === currentStep) return '⏳';
    if (getStepOrder(step) < getStepOrder(currentStep)) return '✅';
    return '⏸️';
  };

  const getStepOrder = (step: GenerationStep): number => {
    const order = {
      copying: 0,
      templating: 1,
      git: 2,
      installing: 3,
      complete: 4,
    };
    return order[step];
  };

  const getStepText = (step: GenerationStep) => {
    switch (step) {
      case 'copying':
        return 'Copying template files';
      case 'templating':
        return 'Processing template variables';
      case 'git':
        return 'Initializing git repository';
      case 'installing':
        return 'Installing dependencies';
      case 'complete':
        return 'Project created successfully!';
      default:
        return 'Unknown step';
    }
  };

  const renderProgressBar = (percent: number) => {
    const width = 30;
    const filled = Math.round((percent / 100) * width);
    const empty = width - filled;

    return (
      <Box>
        <Text color="green">{'█'.repeat(filled)}</Text>
        <Text color="gray">{'░'.repeat(empty)}</Text>
        <Text color="white"> {percent}%</Text>
      </Box>
    );
  };

  const steps: GenerationStep[] = ['copying', 'templating'];
  if (config.initGit && config.gitAvailable) {
    steps.push('git');
  }
  if (config.installDeps) {
    steps.push('installing');
  }

  return (
    <Box flexDirection="column" padding={1}>
      <Text color="cyan">🚀 Generating Project</Text>
      <Newline />
      <Text>
        Creating <Text color="green">{config.projectName}</Text> with{' '}
        <Text color="blue">{config.template}</Text> template...
      </Text>
      <Newline />

      {steps.map((step) => (
        <Box key={step} flexDirection="column" marginBottom={1}>
          <Box>
            <Text>{getStepIcon(step)} </Text>
            <Text color={step === currentStep ? 'cyan' : 'gray'}>{getStepText(step)}</Text>
          </Box>
          {step === currentStep && !error && <Box paddingLeft={2}>{renderProgressBar(progress)}</Box>}
        </Box>
      ))}

      {error && (
        <>
          <Newline />
          <Text color="red">❌ Error: {error}</Text>
        </>
      )}

      {currentStep === 'complete' && (
        <>
          <Newline />
          <Text color="green">✨ Project created successfully!</Text>
        </>
      )}
    </Box>
  );
}
