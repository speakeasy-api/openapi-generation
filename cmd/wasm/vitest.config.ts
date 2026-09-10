import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    // Set the test timeout to 30 seconds (default is 5000ms)
    // The WASM tests should take a little over 5 seconds to run
    testTimeout: 30_000,
    environment: 'jsdom',
    watch: false,
  },
});
