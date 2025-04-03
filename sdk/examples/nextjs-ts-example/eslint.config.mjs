import typescriptEslint from '@typescript-eslint/eslint-plugin';
import _import from 'eslint-plugin-import';
import react from 'eslint-plugin-react';
import jsxA11Y from 'eslint-plugin-jsx-a11y';
import reactHooks from 'eslint-plugin-react-hooks';
import nextPlugin from '@next/eslint-plugin-next';
import { fixupPluginRules } from '@eslint/compat';
import globals from 'globals';
import tsParser from '@typescript-eslint/parser';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import js from '@eslint/js';
import { FlatCompat } from '@eslint/eslintrc';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const compat = new FlatCompat({
    baseDirectory: __dirname,
    recommendedConfig: js.configs.recommended,
    allConfig: js.configs.all,
});

export default [
    {
        ignores: [
            '**/node_modules',
            '**/.next',
            '**/styles',
            '**/public',
            '**/migrations',
            '**/docs',
            '**/data',
            '**/__mocks__',
            '**/dist',
        ],
    },
    ...compat.extends(
        'eslint:recommended',
        'plugin:@typescript-eslint/eslint-recommended',
        'plugin:react/recommended',
        'plugin:react/jsx-runtime',
        'plugin:@next/next/recommended',
    ),
    {
        plugins: {
            '@typescript-eslint': typescriptEslint,
            import: fixupPluginRules(_import),
            react,
            next: nextPlugin,
            'jsx-a11y': jsxA11Y,
            'react-hooks': fixupPluginRules(reactHooks),
        },

        languageOptions: {
            globals: {
                ...globals.browser,
                ...globals.commonjs,
                ...globals.jest,
                ...globals.node,
            },

            parser: tsParser,
        },

        settings: {
            react: {
                pragma: 'React',
                version: 'detect',
            },
        },

        rules: {
            quotes: [
                'error',
                'single',
                {
                    avoidEscape: true,
                },
            ],

            'no-duplicate-imports': ['error'],
            'no-useless-escape': 'off',
            'no-unused-vars': 'off',
            'no-case-declarations': 'off',
            'react/prop-types': 'off',
            'no-debugger': 'warn',
            '@next/next/no-duplicate-head': 'off',
            '@next/next/no-page-custom-font': 'off',

            '@typescript-eslint/no-unused-vars': [
                'warn',
                {
                    argsIgnorePattern: '^_',
                    args: 'after-used',
                    ignoreRestSiblings: true,
                },
            ],

            'react-hooks/rules-of-hooks': 'error',
            'react-hooks/exhaustive-deps': 'off',

            'padding-line-between-statements': [
                'error',
                {
                    blankLine: 'always',
                    prev: ['const', 'let', 'var'],
                    next: '*',
                },
                {
                    blankLine: 'any',
                    prev: ['const', 'let', 'var'],
                    next: ['const', 'let', 'var'],
                },
            ],

            '@typescript-eslint/consistent-type-definitions': ['warn', 'interface'],

            'react/jsx-curly-brace-presence': [
                'error',
                {
                    props: 'never',
                    children: 'never',
                },
            ],

            'react/self-closing-comp': 'error',
            'react/jsx-boolean-value': ['error', 'never'],
            'react/display-name': 'off',
        },
    },
];
