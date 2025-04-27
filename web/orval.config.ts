import { defineConfig } from 'orval';

export default defineConfig({
  inboxWhisperer: {
    input: '../api/openapi.yaml',
    output: {
      mode: 'tags-split',
      target: './src/api/generated/',
      client: 'axios',
      prettier: true,
    },
  },
});
