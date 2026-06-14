import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import Components from 'unplugin-vue-components/vite'

export default defineConfig({
  server: {
    proxy: {
      '/api/': {
        target: 'http://localhost:8085',
        changeOrigin: true,
        secure: false,
      },
      '/lang': {
        target: 'http://localhost:8085',
        changeOrigin: true,
        secure: false,
      }
    },
  },
  plugins: [
    vue(),
    Components({
      dirs: ['node_modules/picocrank/vue/components'],
      extensions: ['vue'],
      deep: true,
      dts: false,
    }),
  ],
})
