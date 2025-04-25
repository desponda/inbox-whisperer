module.exports = {
  inboxWhisperer: {
    input: '../api/openapi.yaml',
    output: {
      mode: 'tags-split',
      target: './src/api/generated/',
      // schemas: './src/api/generated/model.ts',
      client: 'swr',
      mock: false,
    },
    hooks: {
      afterAllFilesWrite: 'prettier --write ./src/api/generated/**/*.{ts,js}',
    },
  },
};
