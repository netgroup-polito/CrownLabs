import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react-swc';
import svgr from 'vite-plugin-svgr';
import tailwindcss from '@tailwindcss/vite';

const base = process.env.PUBLIC_URL || '/';

if (base[base.length - 1] !== '/') {
  throw new Error('PUBLIC_URL must end with a slash');
}

// https://vite.dev/config/
export default defineConfig({
  base,
  server: {
    port: 3000,
  },
  plugins: [react(), svgr(), tailwindcss()],
  build: {
    sourcemap: true,
    rollupOptions: {
      output: {
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
