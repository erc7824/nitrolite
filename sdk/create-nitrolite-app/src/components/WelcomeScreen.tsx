import React, { useEffect } from 'react';
import { Text, Box, Newline } from 'ink';
import { useInput } from 'ink';

interface WelcomeScreenProps {
    onComplete: () => void;
}

const WELCOME_TEXTS = {
    title: '🚀 Welcome to create-nitrolite-app!',
    subtitle: 'The fastest way to create Nitrolite applications',
    instruction: 'Press Enter or Space to continue, or wait 2 seconds...'
};

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

    // Calculate the maximum content width
    const titleLength = WELCOME_TEXTS.title.length;
    const subtitleLength = WELCOME_TEXTS.subtitle.length;
    const maxContentWidth = Math.max(titleLength, subtitleLength);
    
    // Add padding (4 characters: 2 spaces + 2 border chars)
    const boxWidth = maxContentWidth + 4;
    
    // Helper function to create horizontal border
    const createHorizontalBorder = () => '─'.repeat(boxWidth - 2);
    
    // Helper function to create empty line
    const createEmptyLine = () => ' '.repeat(boxWidth - 2);

    // Create centered text line with styling
    const createTitleLine = () => {
        const contentWidth = boxWidth - 2;
        const padding = Math.max(0, contentWidth - titleLength);
        const leftPadding = Math.floor(padding / 2);
        const rightPadding = padding - leftPadding;
        
        return (
            <Text color="cyan">
                │{' '.repeat(leftPadding)}
                <Text color="white" bold>{WELCOME_TEXTS.title}</Text>
                {' '.repeat(rightPadding)}│
            </Text>
        );
    };

    const createSubtitleLine = () => {
        const contentWidth = boxWidth - 2;
        const padding = Math.max(0, contentWidth - subtitleLength);
        const leftPadding = Math.floor(padding / 2);
        const rightPadding = padding - leftPadding;
        
        return (
            <Text color="cyan">
                │{' '.repeat(leftPadding)}
                <Text color="gray">{WELCOME_TEXTS.subtitle}</Text>
                {' '.repeat(rightPadding)}│
            </Text>
        );
    };

    return (
        <Box flexDirection="column" padding={1}>
            <Text color="cyan">┌{createHorizontalBorder()}┐</Text>
            <Text color="cyan">│{createEmptyLine()}│</Text>
            {createTitleLine()}
            <Text color="cyan">│{createEmptyLine()}│</Text>
            {createSubtitleLine()}
            <Text color="cyan">│{createEmptyLine()}│</Text>
            <Text color="cyan">└{createHorizontalBorder()}┘</Text>
            <Newline />
            <Text color="gray">
                {WELCOME_TEXTS.instruction}
            </Text>
        </Box>
    );
}
