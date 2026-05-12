import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import { resolve } from 'path'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const backendUrl = env.VITE_DEV_PROXY_TARGET || 'http://localhost:8080'
  const devPort = Number(env.VITE_DEV_PORT || 3000)
  const outDir = env.VITE_BUILD_OUT_DIR || 'dist'

  return {
    plugins: [react()],
    resolve: {
      alias: {
        '@': resolve(__dirname, 'src')
      }
    },
    build: {
      outDir,
      emptyOutDir: true,
      rollupOptions: {
        output: {
          manualChunks(id) {
            if (!id.includes('node_modules')) return
            if (id.includes('/i18next') || id.includes('/react-i18next')) return 'vendor-i18n'
            if (id.includes('/@tanstack/')) return 'vendor-query'
            // react / react-dom / react-router stay in the main vendor chunk to avoid
            // cross-chunk circular references with their transitive deps (scheduler, etc.)
          }
        }
      }
    },
    server: {
      host: '0.0.0.0',
      port: devPort,
      proxy: {
        '/api': { target: backendUrl, changeOrigin: true },
        '/v1': { target: backendUrl, changeOrigin: true },
        '/setup': { target: backendUrl, changeOrigin: true }
      }
    }
  }
})
