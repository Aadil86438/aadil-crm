import Vue from 'vue'
import Vuetify from 'vuetify'
import 'vuetify/dist/vuetify.min.css'

Vue.use(Vuetify)

export default new Vuetify({
  theme: {
    dark: false,
    themes: {
      light: {
        primary: '#1565C0',
        secondary: '#0288D1',
        accent: '#FF6F00',
        error: '#D32F2F',
        warning: '#F57C00',
        info: '#0288D1',
        success: '#2E7D32',
        background: '#F5F7FA',
        surface: '#FFFFFF',
        sidebar: '#0D47A1',
      }
    }
  },
  icons: {
    iconfont: 'mdi',
  },
})
