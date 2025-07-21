import { useState } from 'react';
import { Text, Box, Newline } from 'ink';
import { useInput } from 'ink';
import { TEMPLATES, Template } from '../constants/templates.js';

interface TemplateSelectorProps {
  currentTemplate: string;
  onSelect: (template: string) => void;
}

export function TemplateSelector({ currentTemplate, onSelect }: TemplateSelectorProps) {
  const [selectedIndex, setSelectedIndex] = useState(() => {
    const index = TEMPLATES.findIndex((t) => t.id === currentTemplate);
    return index >= 0 ? index : 0;
  });

  useInput((input, key) => {
    if (key.upArrow) {
      setSelectedIndex((prev) => (prev > 0 ? prev - 1 : TEMPLATES.length - 1));
    } else if (key.downArrow) {
      setSelectedIndex((prev) => (prev < TEMPLATES.length - 1 ? prev + 1 : 0));
    } else if (key.return) {
      onSelect(TEMPLATES[selectedIndex].id);
    } else if (key.ctrl && input === 'c') {
      process.exit(0);
    }
  });

  return (
    <Box flexDirection="column" padding={1}>
      <Text color="cyan">🎨 Select Template</Text>
      <Newline />
      <Text>Choose a template for your Nitrolite application:</Text>
      <Newline />

      {TEMPLATES.map((template, index) => (
        <Box key={template.id} flexDirection="column" marginBottom={1}>
          <Box>
            <Text color={index === selectedIndex ? 'green' : 'gray'}>
              {index === selectedIndex ? '❯ ' : '  '}
            </Text>
            <Text color={index === selectedIndex ? 'green' : 'white'} bold={index === selectedIndex}>
              {template.name}
            </Text>
          </Box>
          {index === selectedIndex && (
            <Box flexDirection="column" paddingLeft={2}>
              <Text color="gray">{template.description}</Text>
              <Box flexDirection="row" flexWrap="wrap" gap={1}>
                {template.features.map((feature, featureIndex) => (
                  <Text key={featureIndex} color="blue">
                    • {feature}
                  </Text>
                ))}
              </Box>
            </Box>
          )}
        </Box>
      ))}

      <Newline />
      <Text color="gray">
        Use <Text color="white">↑↓</Text> arrows to navigate, <Text color="white">Enter</Text> to select
      </Text>
      <Text color="gray">
        Press <Text color="white">Ctrl+C</Text> to exit
      </Text>
    </Box>
  );
}
