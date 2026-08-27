import api from './api'

export default {
  list(params = {}) { return api.get('/api/accounts', { params }) },
  listSimple() { return api.get('/api/accounts/simple') },
  get(id) { return api.get(`/api/accounts/${id}`) },
  create(data) { return api.post('/api/accounts', data) },
  update(id, data) { return api.put(`/api/accounts/${id}`, data) },
  delete(id) { return api.delete(`/api/accounts/${id}`) }
}
