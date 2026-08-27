import api from './api'

export default {
  list(params = {}) { return api.get('/api/tasks', { params }) },
  get(id) { return api.get(`/api/tasks/${id}`) },
  create(data) { return api.post('/api/tasks', data) },
  update(id, data) { return api.put(`/api/tasks/${id}`, data) },
  delete(id) { return api.delete(`/api/tasks/${id}`) }
}
