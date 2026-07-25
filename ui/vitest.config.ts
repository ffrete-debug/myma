/// <reference types="vitest" />
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import { resolve } from 'node:path'

// Vitest config kept separate from next.config.ts so that running unit tests
// does not invoke the Next.js compiler. We only need the React plugin for the
// JSX in component tests; the rest of the TS pipeline comes from esbuild
// which understands the tsconfig path alias.
//
// CSS is stubbed out (`css: false`) so vitest doesn't try to load the
// project's postcss.config.mjs — that config uses Next.js-specific plugin
// shorthand that vite's standalone PostCSS pipeline can't parse.
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  css: {
    // Disable PostCSS discovery: postcss.config.mjs uses a Next.js-specific
    // plugin string shorthand (`"@tailwindcss/postcss"`) that vite's
    // standalone PostCSS pipeline refuses to load — and we don't need CSS
    // processing for unit tests anyway.
    postcss: {
      plugins: [],
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    include: ['src/**/*.{test,spec}.{ts,tsx}'],
  },
})
