# Nitrolite Example Tutorials

Learn Nitrolite through hands-on examples and comprehensive tutorials.

## Available Tutorials

### [Hello Nitrolite: State Channels Made Simple](hello-nitrolite.md)

Learn state channels by building your first Nitrolite application - clearer than any quick start guide

**Difficulty:** beginner | **Time:** 15 minutes

**What you'll learn:**

- Complete setup from zero to working state channel
- Core concepts: ClearNode, authentication, channel lifecycle
- Instant state updates with zero gas fees
- Production-ready patterns and best practices

## Getting Started

1. Start with [Hello Nitrolite](hello-nitrolite.md) to understand the fundamentals
2. Make sure you have the [prerequisites](../README.md#prerequisites) installed
3. Follow the step-by-step instructions
4. Experiment and build upon the examples

## Tutorial Format

Each tutorial follows a consistent structure:

- **Tutorial Info** - Difficulty, time estimate, and technologies
- **Prerequisites** - What you need to know beforehand
- **What You'll Learn** - Key concepts covered
- **Architecture Overview** - High-level system design
- **Step-by-step Guide** - Detailed implementation walkthrough with actual commands
- **Working Code** - Real examples with line references
- **Concept Deep Dives** - Detailed explanations of important topics
- **Production Guidance** - Security, UX, and scaling considerations

## Why These Tutorials Are Different

Our tutorials use **literate programming** - they're generated from working TypeScript code with comprehensive comments. This means:

- **Always Current**: Generated from working code, so types and APIs never get outdated
- **Complete Setup**: Every command, config file, and dependency is included
- **Actionable Steps**: Concrete instructions you can follow immediately
- **Zero Maintenance**: Documentation updates automatically when code changes

## Contributing New Tutorials

Want to create a new tutorial? Here's how our literate programming system works:

### 1. Create Your Example Project

Create a new directory in `examples/` with a working application:

```bash
mkdir examples/my-tutorial
cd examples/my-tutorial
npm init -y
# ... build your working example
```

### 2. Add Tutorial Comments

Add special comments to your TypeScript files using these patterns:

```typescript
/**tutorial-meta
title: "My Amazing Tutorial"
description: "Learn how to build something amazing"
difficulty: "beginner"
estimatedTime: "20 minutes"
technologies: ["TypeScript", "React", "Nitrolite"]
concepts: ["State Channels", "Real-time Updates"]
prerequisites: ["Basic TypeScript", "Node.js installed"]
*/

/**tutorial:architecture
# System Architecture

Explain your app's high-level design with diagrams:

```

[Client] <-> [WebSocket] <-> [Server] <-> [Nitrolite]

````

This helps users understand the big picture before diving into code.
*/

/**tutorial:step Setup the Project
Each step explains what you're implementing and why.
Include actual commands and code examples.

```bash
mkdir my-project
cd my-project
npm install @erc7824/nitrolite
````

This creates the basic project structure we'll need.
\*/

/\*_tutorial:concept Understanding State Channels
Concept blocks provide detailed explanations of important topics.
Use these to teach underlying principles, not just mechanics.
_/

class MyExampleClass {
// Your working code here - it will be automatically extracted
// and included in the generated tutorial with line references
}

````

### 3. Generate the Tutorial

Run the generation command to create professional markdown:

```bash
npm run docs:tutorials
````

This automatically:

- Extracts your tutorial comments
- Generates a professional markdown file with table of contents
- Includes code examples with file references and line numbers
- Creates consistent formatting and structure

### 4. Review and Submit

- Check the generated markdown in `docs/tutorials/`
- Test that all code examples work
- Submit a pull request

### Comment Types Reference

- **`/**tutorial-meta`\*\* - Tutorial metadata (title, difficulty, time, etc.)
- **`/**tutorial:architecture`\*\* - High-level system design explanations
- **`/**tutorial:step`\*\* - Step-by-step implementation instructions
- **`/**tutorial:concept`\*\* - Deep dives into important concepts

### Best Practices

- **Start with Architecture**: Help users understand the big picture first
- **Explain Why**: Don't just show what to do, explain why it matters
- **Use Real Examples**: Base tutorials on working applications that actually run
- **Progress Gradually**: Each step builds naturally on the previous
- **Include Concepts**: Teach underlying principles, not just mechanics

## Next Steps

After completing the tutorials:

- Continue with the [ERC-7824 Quick Start Guide](https://erc7824.org/quick_start/) for advanced features
- Explore the [SDK documentation](../README.md)
- Check out the [example projects](../../examples/)
- Report issues or contribute at [GitHub](https://github.com/erc7824/nitrolite)

The goal: make Nitrolite the easiest state channel framework to learn and use.
