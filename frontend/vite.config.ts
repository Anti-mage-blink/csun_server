import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// 开发期通过 proxy 将 /api 转发到统一 Gin 后端，与生产期 Nginx 反代保持一致
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    // 单个 chunk 超过 500KB 时给出警告，这里放宽到 1500KB 以避免 antd 等大依赖触发告警
    chunkSizeWarningLimit: 1500,
    rollupOptions: {
      output: {
        // 将第三方依赖拆分到独立 chunk，便于浏览器缓存
        manualChunks: {
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          'antd-vendor': ['antd', '@ant-design/icons'],
          'utils-vendor': ['axios', 'zustand'],
        },
      },
    },
  },
})
