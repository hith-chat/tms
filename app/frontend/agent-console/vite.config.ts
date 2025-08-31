import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test-setup.ts'],
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@shared': path.resolve(__dirname, '../shared/src')
    }
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '/v1')
      }
    }
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: {
          // React and core libraries
          vendor: ['react', 'react-dom', 'react-router-dom'],
          // UI components and styling
          ui: ['@radix-ui/react-avatar', '@radix-ui/react-dialog', '@radix-ui/react-dropdown-menu', 
               '@radix-ui/react-popover', '@radix-ui/react-select', '@radix-ui/react-switch', 
               '@radix-ui/react-tabs', '@radix-ui/react-toast', '@radix-ui/react-tooltip',
               'lucide-react', 'class-variance-authority', 'clsx'],
          // Data fetching and state management
          data: ['@tanstack/react-query', '@tanstack/react-virtual', 'axios'],
          // Form handling and utilities
          forms: ['react-hook-form', '@hookform/resolvers', 'zod'],
          // Date and formatting utilities
          utils: ['date-fns', 'react-markdown', 'react-syntax-highlighter'],
          // Shared components
          shared: ['@tms/shared']
        }
      }
    },
    // Increase chunk size warning limit
    chunkSizeWarningLimit: 1000,
    // Enable minification optimizations
    minify: 'esbuild',
    target: 'es2020'
  }
})
