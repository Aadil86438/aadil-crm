import api from './api'

export default {
  search(q) { return api.get('/api/search', { params: { q } }) }
}
