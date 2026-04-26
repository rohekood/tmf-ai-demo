import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';


// More info at: https://storybook.js.org/docs/next/writing-tests/integrations/vitest-addon
export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    globals: true,
    clearMocks: true,
    setupFiles: './src/test/setup.ts',
    exclude: ['**/*.stories.tsx', '**/*.stories.ts', 'node_modules', 'dist'],
  }
});