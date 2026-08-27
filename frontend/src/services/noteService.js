import api from './api'

export default {
  list(params = {}) { return api.get('/api/notes', { params }) },
  create(data) { return api.post('/api/notes', data) },
  update(id, data) { return api.put(`/api/notes/${id}`, data) },
  delete(id) { return api.delete(`/api/notes/${id}`) }
}
