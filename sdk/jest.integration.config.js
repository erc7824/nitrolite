/** @type {import('jest').Config} */
module.exports = {
    preset: 'ts-jest',
    testEnvironment: 'node',
    roots: ['<rootDir>/test/integration'],
    testMatch: ['**/test/integration/**/*.test.ts', '**/test/integration/**/*.test.js'],
    transform: {
        '^.+\\.ts$': 'ts-jest',
    },
    collectCoverageFrom: ['src/**/*.{ts,js}', '!src/**/*.d.ts', '!src/**/*.test.{ts,js}', '!src/**/__tests__/**'],
    coverageDirectory: 'coverage',
    coverageReporters: ['text', 'lcov', 'html'],
    testTimeout: 30000, // 30 seconds for integration tests
    maxWorkers: 1, // Run tests sequentially for integration testing
    verbose: true,
    // setupFilesAfterEnv removed - each test file handles its own setup now
    moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json', 'node'],
    moduleNameMapper: {
        '^@/(.*)$': '<rootDir>/src/$1',
    },
    testPathIgnorePatterns: ['/node_modules/', '/dist/', '/build/'],
    coveragePathIgnorePatterns: ['/node_modules/', '/test/', '/dist/', '/build/'],
    // Global test setup
    globalSetup: undefined,
    globalTeardown: undefined,
    // Environment variables for tests
    testEnvironmentOptions: {
        NODE_ENV: 'test',
    },
};
