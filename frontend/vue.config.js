const { defineConfig } = require('@vue/cli-service')

module.exports = defineConfig({
  transpileDependencies: ['vuetify'],
  lintOnSave: false,
  devServer: {
    port: 3000,
    proxy: {
      '/api/admin': {
        target: process.env.VUE_APP_ADMIN_API_URL || 'http://localhost:8081',
        changeOrigin: true,
      },
      '/api': {
        target: process.env.VUE_APP_API_URL || 'http://localhost:8080',
        changeOrigin: true,
      },
      '/health': {
        target: process.env.VUE_APP_API_URL || 'http://localhost:8080',
        changeOrigin: true,
      }
    }
  }
})
