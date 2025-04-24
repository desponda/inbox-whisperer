// Jest configuration for Vite + React + TypeScript (CommonJS)
module.exports = {
  testEnvironment: 'jsdom',
  transform: {
    '^.+\\.(ts|tsx)$': 'ts-jest',
  },
  moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json', 'node', 'css'],
  setupFilesAfterEnv: ['<rootDir>/setupTests.ts'],

  moduleNameMapper: {
    '^.+\.css$': 'identity-obj-proxy',
    '^.+\.(less|scss|sass)$': 'identity-obj-proxy',
  },
  testMatch: [
    '<rootDir>/src/**/*.(spec|test).(ts|tsx|js)'
  ],
};
