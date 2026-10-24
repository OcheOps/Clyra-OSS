/// <reference types="vitest/config" />
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  test: {
    include: ['src/**/*.test.{ts,tsx}'],
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('/src/lib/generated/componentCatalog.ts')) return 'component-catalog';
          if (!id.includes('/node_modules/')) return undefined;
          if (id.includes('/@xyflow/')) return 'react-flow';
          if (id.includes('/jspdf/')) return 'export-pdf';
          if (id.includes('/html2canvas/')) return 'export-canvas';
          if (id.includes('/html-to-image/')) return 'export-image';
          return undefined;
        },
      },
    },
  },
});
