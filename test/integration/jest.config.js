/** @type {import('ts-jest').JestConfigWithTsJest} */
module.exports = {
    preset: 'ts-jest',
    testEnvironment: 'node',
    roots: ['<rootDir>/tests'],
    testMatch: ['./**/*.test.ts'],
    transform: {
        '^.+\\.ts$': 'ts-jest',
    },
    testTimeout: 20000,
    modulePaths: ['<rootDir>/common'],
    moduleNameMapper: {
        '^@/(.*)$': '<rootDir>/common/$1',
    },
};
