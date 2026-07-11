import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

const env = (globalThis as { process?: { env?: Record<string, string | undefined> } }).process?.env;
const apiProxyTarget = env?.VITE_API_PROXY_TARGET ?? 'http://localhost:8080';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: apiProxyTarget,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },
});
