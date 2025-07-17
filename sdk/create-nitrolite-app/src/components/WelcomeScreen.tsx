import React, { useEffect } from 'react';
import { Text, Box, Newline } from 'ink';
import { useInput } from 'ink';

interface WelcomeScreenProps {
  onComplete: () => void;
}

export function WelcomeScreen({ onComplete }: WelcomeScreenProps) {
  useInput((input, key) => {
    if (key.return || input === ' ') {
      onComplete();
    }
  });

  useEffect(() => {
    // Auto-advance after 2 seconds or wait for user input
    const timer = setTimeout(() => {
      onComplete();
    }, 2000);

    return () => clearTimeout(timer);
  }, [onComplete]);

  return (
    <Box flexDirection="column" padding={1}>
      <Text color="cyan">
        ┌─────────────────────────────────────────────────────────────┐
      </Text>
      <Text color="cyan">
        │                                                             │
      </Text>
      <Text color="cyan">
        │  <Text color="white" bold>🚀 Welcome to create-nitrolite-app!</Text>                    │
      </Text>
      <Text color="cyan">
        │                                                             │
      </Text>
      <Text color="cyan">
        │  <Text color="gray">The fastest way to create Nitrolite applications</Text>       │
      </Text>
      <Text color="cyan">
        │                                                             │
      </Text>
      <Text color="cyan">
        └─────────────────────────────────────────────────────────────┘
      </Text>
      <Newline />
      <Text color="gray">Press <Text color="white">Enter</Text> or <Text color="white">Space</Text> to continue, or wait 2 seconds...</Text>
    </Box>
  );
}