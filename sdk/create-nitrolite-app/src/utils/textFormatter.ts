import { DEFAULTS } from '../constants/defaults.js';

/**
 * Calculates padding for centering text within a given width
 */
export function calculateCenterPadding(textLength: number, containerWidth: number): {
  leftPadding: number;
  rightPadding: number;
} {
  const totalPadding = Math.max(0, containerWidth - textLength);
  const leftPadding = Math.floor(totalPadding / 2);
  const rightPadding = totalPadding - leftPadding;
  
  return { leftPadding, rightPadding };
}

/**
 * Creates a horizontal border of specified width
 */
export function createHorizontalBorder(width: number, char = '─'): string {
  return char.repeat(Math.max(0, width - DEFAULTS.BORDER_THICKNESS));
}

/**
 * Creates an empty line of specified width
 */
export function createEmptyLine(width: number): string {
  return ' '.repeat(Math.max(0, width - DEFAULTS.BORDER_THICKNESS));
}

/**
 * Calculates the maximum width needed for a box containing multiple text elements
 */
export function calculateBoxWidth(...textLengths: number[]): number {
  const maxContentWidth = Math.max(...textLengths);
  return maxContentWidth + DEFAULTS.BORDER_PADDING;
}

/**
 * Gets the maximum line width from multi-line text
 */
export function getMaxLineWidth(text: string): number {
  const lines = text.trim().split('\n');
  return Math.max(...lines.map(line => line.length));
}

/**
 * Centers text within a container and returns padding information
 */
export function centerText(text: string, containerWidth: number): {
  text: string;
  leftPadding: number;
  rightPadding: number;
} {
  const contentWidth = containerWidth - DEFAULTS.BORDER_THICKNESS;
  const { leftPadding, rightPadding } = calculateCenterPadding(text.length, contentWidth);
  
  return {
    text,
    leftPadding,
    rightPadding,
  };
}

/**
 * Formats multiple lines with consistent centering
 */
export function centerMultipleLines(lines: string[], containerWidth: number): Array<{
  text: string;
  leftPadding: number;
  rightPadding: number;
}> {
  return lines.map(line => centerText(line, containerWidth));
}