import api from './api'

export default {
  list(params = {}) { return api.get('/api/users', { params }) },
  listSimple() { return api.get('/api/users/simple') },
  create(data) { return api.post('/api/users', data) },
  update(id, data) { return api.put(`/api/users/${id}`, data) },
  resetPassword(id, password) { return api.put(`/api/users/${id}/password`, { password }) },
  delete(id) { return api.delete(`/api/users/${id}`) }
}
