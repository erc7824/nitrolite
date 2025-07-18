import { useEffect } from 'react';
import { Text, Box, Newline } from 'ink';

interface ProjectConfig {
  projectPath: string;
  projectName: string;
  template: string;
  initGit: boolean;
  installDeps: boolean;
}

interface CompletionScreenProps {
  config: ProjectConfig;
}

export function CompletionScreen({ config }: CompletionScreenProps) {
  useEffect(() => {
    // Exit the process after showing the completion message
    const timer = setTimeout(() => {
      process.exit(0);
    }, 100); // Small delay to ensure the message is displayed

    return () => clearTimeout(timer);
  }, []);

  const getNextSteps = () => {
    const steps = [];

    steps.push(`cd ${config.projectPath}`);

    if (!config.installDeps) {
      steps.push('npm install');
    }

    // Add template-specific commands
    switch (config.template) {
      case 'react-vite':
      case 'vue-composition':
        steps.push('npm run dev');
        break;
      case 'nextjs-app':
        steps.push('npm run dev');
        break;
      case 'minimal-sdk':
        steps.push('npm run start');
        break;
    }

    return steps;
  };

  const getTemplateInfo = () => {
    switch (config.template) {
      case 'react-vite':
        return {
          name: 'React + Vite',
          port: '5173',
          features: ['Hot Module Replacement', 'TypeScript', 'TailwindCSS', 'WebSocket integration'],
        };
      case 'vue-composition':
        return {
          name: 'Vue 3 + Composition API',
          port: '5173',
          features: ['Composition API', 'TypeScript', 'WebSocket integration', 'Vite'],
        };
      case 'nextjs-app':
        return {
          name: 'Next.js App Router',
          port: '3000',
          features: ['App Router', 'TypeScript', 'TailwindCSS', 'SSR support'],
        };
      case 'minimal-sdk':
        return {
          name: 'Minimal SDK Integration',
          port: null,
          features: ['TypeScript', 'WebSocket client', 'SDK only'],
        };
      default:
        return {
          name: 'Unknown template',
          port: null,
          features: [],
        };
    }
  };

  const nextSteps = getNextSteps();
  const templateInfo = getTemplateInfo();

  return (
    <Box flexDirection="column" padding={1}>
      <Text color="green">┌─────────────────────────────────────────────────────────────┐</Text>
      <Text color="green">│ │</Text>
      <Text color="green">
        │{' '}
        <Text color="white" bold>
          🎉 Success! Your Nitrolite app is ready!
        </Text>{' '}
        │ │
      </Text>
      <Text color="green">│ │</Text>
      <Text color="green">└─────────────────────────────────────────────────────────────┘</Text>

      <Newline />

      <Box flexDirection="column">
        <Text color="cyan">📋 Project Details:</Text>
        <Newline />
        <Text>
          {' '}
          📁 Project: <Text color="green">{config.projectName}</Text>
        </Text>
        <Text>
          {' '}
          🎨 Template: <Text color="blue">{templateInfo.name}</Text>
        </Text>
        <Text>
          {' '}
          🔧 Git:{' '}
          <Text color={config.initGit ? 'green' : 'gray'}>
            {config.initGit ? 'Initialized' : 'Not initialized'}
          </Text>
        </Text>
        <Text>
          {' '}
          📦 Dependencies:{' '}
          <Text color={config.installDeps ? 'green' : 'gray'}>
            {config.installDeps ? 'Installed' : 'Not installed'}
          </Text>
        </Text>
        {templateInfo.port && (
          <Text>
            {' '}
            🌐 Dev server: <Text color="yellow">http://localhost:{templateInfo.port}</Text>
          </Text>
        )}
      </Box>

      <Newline />

      <Box flexDirection="column">
        <Text color="cyan">✨ Features included:</Text>
        <Newline />
        {templateInfo.features.map((feature, index) => (
          <Text key={index}> • {feature}</Text>
        ))}
      </Box>

      <Newline />

      <Box flexDirection="column">
        <Text color="cyan">🚀 Next steps:</Text>
        <Newline />
        {nextSteps.map((step, index) => (
          <Box key={index} flexDirection="row">
            <Text color="gray">{index + 1}. </Text>
            <Text color="white" backgroundColor="gray" bold>
              {' '}
              {step}{' '}
            </Text>
          </Box>
        ))}
      </Box>

      <Newline />

      <Box flexDirection="column">
        <Text color="cyan">📚 Documentation:</Text>
        <Newline />
        <Text>
          {' '}
          🔗 Nitrolite SDK: <Text color="blue">https://github.com/erc7824/nitrolite</Text>
        </Text>
        <Text>
          {' '}
          📖 Examples: <Text color="blue">./examples/</Text>
        </Text>
        <Text>
          {' '}
          🐛 Issues: <Text color="blue">https://github.com/erc7824/nitrolite/issues</Text>
        </Text>
      </Box>

      <Newline />

      <Text color="gray">Happy coding! 🚀</Text>
    </Box>
  );
}
