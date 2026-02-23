import { Text, Box, Newline } from 'ink';
import { useInput } from 'ink';
import { ProjectConfig } from '../types/index.js';
import { TEMPLATES } from '../constants/templates.js';

interface ConfirmationScreenProps {
  config: ProjectConfig;
  onConfirm: () => void;
  onCancel: () => void;
}

function getTemplateName(templateId: string): string {
  const template = TEMPLATES.find(t => t.id === templateId);
  return template?.name || templateId;
}

export function ConfirmationScreen({ config, onConfirm, onCancel }: ConfirmationScreenProps) {
  useInput((input, key) => {
    if (key.ctrl && input === 'c') {
      process.exit(0);
    } else if (input === 'y') {
      onConfirm();
    } else if (input === 'n') {
      onCancel();
    }
  });

  return (
    <Box flexDirection="column" padding={1}>
      <Text color="cyan">✅ Confirm Project Setup</Text>
      <Newline />
      <Text>
        Project directory: <Text color="green">{config.projectPath}</Text>
      </Text>
      <Text>
        Package name: <Text color="green">{config.projectName}</Text>
      </Text>
      <Text>
        Template: <Text color="green">{getTemplateName(config.template)}</Text>
      </Text>
      <Text>
        Initialize git: <Text color={config.initGit ? 'green' : 'red'}>{config.initGit ? 'Yes' : 'No'}</Text>
      </Text>
      <Newline />
      <Text>Create project with these settings? (y/n)</Text>
      <Newline />
      <Text color="gray">
        Press <Text color="white">y</Text> to continue, <Text color="white">n</Text> to go back
      </Text>
    </Box>
  );
}