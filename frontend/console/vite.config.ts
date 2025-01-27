import react from '@vitejs/plugin-react'
import { visualizer } from 'rollup-plugin-visualizer'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [
    react(),
    process.env.ANALYZE === 'true' &&
      visualizer({
        open: true, // Automatically open the visualization in your browser
        gzipSize: true, // Show gzipped sizes
        brotliSize: true, // Show brotli sizes
        filename: 'dist/stats.html', // Output file
      }),
  ].filter(Boolean),
  css: {
    modules: {
      localsConvention: 'camelCaseOnly',
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules/lodash')) {
            return 'vendor-lodash'
          }
          if (id.includes('node_modules/react/') || id.includes('node_modules/react-dom') || id.includes('node_modules/react-router-dom')) {
            return 'vendor-react'
          }
          if (id.includes('node_modules/@headlessui/react') || id.includes('node_modules/@tanstack/react-query')) {
            return 'vendor-ui'
          }
          if (id.includes('node_modules/json-schema-faker')) {
            return 'vendor-json-schema-faker'
          }
          if (id.includes('node_modules/json-schema-library')) {
            return 'vendor-json-schema-library'
          }
          if (id.includes('node_modules/yaml')) {
            return 'vendor-yaml'
          }
          if (id.includes('node_modules/fuse.js')) {
            return 'vendor-fuse'
          }
          if (id.includes('node_modules/reactflow') || id.includes('node_modules/@reactflow')) {
            return 'vendor-reactflow'
          }
          if (id.includes('node_modules/@codemirror')) {
            return 'vendor-codemirror'
          }
          if (id.includes('/protos/xyz/block/ftl/')) {
            return 'vendor-protos'
          }
        },
      },
    },
  },
})
