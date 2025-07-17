#!/usr/bin/env node

import React from 'react';
import { render } from 'ink';
import { program } from 'commander';
import CreateNitroliteApp from './components/CreateNitroliteApp.js';

program
  .name('create-nitrolite-app')
  .description('CLI tool to create new Nitrolite applications')
  .version('1.0.0')
  .argument('[project-directory]', 'directory where the project will be created')
  .option('-t, --template <template>', 'template to use', 'react-vite')
  .option('--no-git', 'skip git repository initialization')
  .option('--no-install', 'skip dependency installation')
  .option('-y, --yes', 'skip prompts and use defaults')
  .action((projectDirectory, options) => {
    render(
      <CreateNitroliteApp
        projectDirectory={projectDirectory}
        template={options.template}
        skipGit={!options.git}
        skipInstall={!options.install}
        skipPrompts={options.yes}
      />
    );
  });

program.parse();