/** @type {import('ts-jest').JestConfigWithTsJest} */
module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'node',
  testMatch: ['**/test/**/*.test.ts'],
  collectCoverageFrom: [
    'src/**/*.ts',
    '!src/**/index.ts',
    '!src/**/*.d.ts'
  ],
  collectCoverage: true,
  coverageReporters: ['text'],
  transform: {
    '^.+\\.ts$': 'ts-jest'
  }
};