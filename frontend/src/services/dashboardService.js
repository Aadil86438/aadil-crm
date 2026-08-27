import api from './api'

export default {
  getStats() {
    return api.get('/api/dashboard')
  }
}
