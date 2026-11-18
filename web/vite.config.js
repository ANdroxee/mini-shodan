import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// En développement, /api est relayé vers l'API Go (make dev-api).
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: { '/api': process.env.VITE_API_URL || 'http://localhost:8080' },
  },
});
