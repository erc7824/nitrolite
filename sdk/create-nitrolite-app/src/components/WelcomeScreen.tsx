import React from 'react';
import { Text, Box, Newline, useInput } from 'ink';

interface WelcomeScreenProps {
    onComplete: () => void;
    interactive?: boolean;
}

const ART = `

███████╗██████╗  ██████╗███████╗ █████╗ ██████╗ ██╗  ██╗
██╔════╝██╔══██╗██╔════╝╚════██║██╔══██╗╚════██╗██║  ██║
█████╗  ██████╔╝██║         ██╔╝╚█████╔╝ █████╔╝███████║
██╔══╝  ██╔══██╗██║        ██╔╝ ██╔══██╗██╔═══╝ ╚════██║
███████╗██║  ██║╚██████╗   ██║  ╚█████╔╝███████╗     ██║
╚══════╝╚═╝  ╚═╝ ╚═════╝   ╚═╝   ╚════╝ ╚══════╝     ╚═╝
`;

const WELCOME_TEXTS = {
    title: '🚀 Welcome to Nitrolite!',
    subtitle: 'The fastest way to create Nitrolite applications',
    instruction: 'Press Enter or Space to continue...',
};

export function WelcomeScreen({ onComplete, interactive = true }: WelcomeScreenProps) {
    // Handle input for interactive mode
    useInput((input, key) => {
        if (interactive && (key.return || input === ' ')) {
            onComplete();
        }
    });

    // Calculate the maximum content width
    const titleLength = WELCOME_TEXTS.title.length;
    const subtitleLength = WELCOME_TEXTS.subtitle.length;
    
    // Get the width of the ASCII art (find the longest line)
    const artLines = ART.trim().split('\n');
    const artWidth = Math.max(...artLines.map(line => line.length));
    
    const maxContentWidth = Math.max(titleLength, subtitleLength, artWidth);

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
                <Text color="white" bold>
                    {WELCOME_TEXTS.title}
                </Text>
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

    const createArtLines = () => {
        return artLines.map((line, index) => {
            const contentWidth = boxWidth - 2;
            const padding = Math.max(0, contentWidth - line.length);
            const leftPadding = Math.floor(padding / 2);
            const rightPadding = padding - leftPadding;

            return (
                <Text key={index} color="cyan">
                    │{' '.repeat(leftPadding)}
                    <Text color="magenta">{line}</Text>
                    {' '.repeat(rightPadding)}│
                </Text>
            );
        });
    };

    return (
        <Box>
            <Box flexDirection="column" padding={1}>
                <Text color="cyan">┌{createHorizontalBorder()}┐</Text>
                <Text color="cyan">│{createEmptyLine()}│</Text>
                {createArtLines()}
                <Text color="cyan">│{createEmptyLine()}│</Text>
                {createTitleLine()}
                <Text color="cyan">│{createEmptyLine()}│</Text>
                {createSubtitleLine()}
                <Text color="cyan">│{createEmptyLine()}│</Text>
                <Text color="cyan">└{createHorizontalBorder()}┘</Text>
                {interactive && (
                    <>
                        <Newline />
                        <Text color="gray">{WELCOME_TEXTS.instruction}</Text>
                    </>
                )}
            </Box>
        </Box>
    );
}
