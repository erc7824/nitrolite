# Viper Duel Visual Effects Guide

This document provides an overview of the particle and aura visual effects implemented in the Viper Duel game.

## Core Components

### 1. CellAuraEffect

Located in `CellAuraEffect.tsx`, this component creates dynamic canvas-based particle effects that appear around cells when a player marks X or O. The effects are customized based on the player symbol (cyan for X, magenta for O).

Key features:
- Canvas-based rendering for performance
- Dynamic particle generation and movement
- Particle fading and lifecycle management
- Explosion effect when a cell is first marked

### 2. CellMarkEffect

Located in `CellMarkEffect.tsx`, this component adds dramatic visual effects when a player marks a cell, including:
- Center burst animation
- Orbiting particles
- Radiating lines
- Random floating particles

The component uses CSS animations and transforms for these effects.

### 3. BackgroundAnimation

Located in `BackgroundAnimation.tsx`, this component renders the global background particle effect. Features include:
- Floating particles across the entire game background
- Interactive effects that respond to mouse movement
- Connection lines between nearby particles
- Burst effects triggered by cell marks via custom events

## CSS and Animation

### Custom Animations

The following custom animations are defined in `tailwind.config.js`:
- `sparkle`: Particles moving upward with rotation
- `float`: Gentle vertical floating motion
- `pulse-cyan` and `pulse-magenta`: Glowing pulsation effects
- `explode-cyan` and `explode-magenta`: Explosion animations
- `orbit`: Circular movement for particles

Additional CSS animations in `App.css`:
- Particle movement
- Glow effects
- Line growth animations

## Event Communication

Components communicate through custom events:
- `cellMarked`: Dispatched when a cell is marked with X or O, includes position and player data
- This event triggers background particle bursts and sound effects

## Sound Effects

Sound feedback is provided through the `useSoundEffects` hook, which:
- Generates programmatic audio for different game events
- Plays appropriate sounds for marking cells and game outcomes
- Adjusts volume and timing for optimal user experience

## Integration Points

The effects are integrated at these key points:
1. In `Cell.tsx` when a player marks a cell
2. In `GameOver.tsx` for win/lose/draw animations
3. Throughout the game interface with consistent cyan/magenta theming

## Extending the Effects

To add new visual effects:
1. Use the existing color scheme (cyan for X, magenta for O)
2. Follow the pattern of component-based effects
3. Utilize custom events for communication between components
4. Add new animations to tailwind.config.js
5. Consider performance implications for mobile devices

---

*"Every move leaves an aura" - Viper Duel*