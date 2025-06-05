import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react-swc';
import svgr from 'vite-plugin-svgr';
import tailwindcss from '@tailwindcss/vite';

// https://vite.dev/config/
export default defineConfig({
  base: process.env.PUBLIC_URL || '/',
  server: {
    port: 3000,
  },
  plugins: [react(), svgr(), tailwindcss()],
  build: {
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: {
          oidc: ['react-oidc-context'],
        },
        chunkFileNames: 'assets/[hash].js',
        assetFileNames: 'assets/[hash][extname]',
      },
    },
  },
  css: {
    preprocessorOptions: {
      less: {
        additionalData: '@primary-color: #1c7afd; @secondary-color: #FF7C11;',
      },
    },
  },
});
