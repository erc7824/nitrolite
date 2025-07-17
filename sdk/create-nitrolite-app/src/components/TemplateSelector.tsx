import React, { useState } from 'react';
import { Text, Box, Newline } from 'ink';
import { useInput } from 'ink';

interface Template {
  id: string;
  name: string;
  description: string;
  features: string[];
}

const TEMPLATES: Template[] = [
  {
    id: 'react-vite',
    name: 'React + Vite',
    description: 'Modern React setup with Vite, TypeScript, and TailwindCSS',
    features: ['React 18', 'Vite', 'TypeScript', 'TailwindCSS', 'WebSocket integration']
  },
  {
    id: 'vue-composition',
    name: 'Vue 3 + Composition API',
    description: 'Vue 3 with Composition API, Vite, and TypeScript',
    features: ['Vue 3', 'Composition API', 'Vite', 'TypeScript', 'WebSocket integration']
  },
  {
    id: 'nextjs-app',
    name: 'Next.js App Router',
    description: 'Next.js with App Router, TypeScript, and TailwindCSS',
    features: ['Next.js 14', 'App Router', 'TypeScript', 'TailwindCSS', 'SSR support']
  },
  {
    id: 'minimal-sdk',
    name: 'Minimal SDK Integration',
    description: 'Minimal setup with just the Nitrolite SDK',
    features: ['TypeScript', 'Minimal setup', 'WebSocket client', 'SDK only']
  }
];

interface TemplateSelectorProps {
  currentTemplate: string;
  onSelect: (template: string) => void;
}

export function TemplateSelector({ currentTemplate, onSelect }: TemplateSelectorProps) {
  const [selectedIndex, setSelectedIndex] = useState(
    TEMPLATES.findIndex(t => t.id === currentTemplate) || 0
  );

  useInput((input, key) => {
    if (key.upArrow) {
      setSelectedIndex(prev => prev > 0 ? prev - 1 : TEMPLATES.length - 1);
    } else if (key.downArrow) {
      setSelectedIndex(prev => prev < TEMPLATES.length - 1 ? prev + 1 : 0);
    } else if (key.return) {
      onSelect(TEMPLATES[selectedIndex].id);
    } else if (key.ctrl && input === 'c') {
      process.exit(0);
    }
  });

  const selectedTemplate = TEMPLATES[selectedIndex];

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
      <Text color="gray">Press <Text color="white">Ctrl+C</Text> to exit</Text>
    </Box>
  );
}