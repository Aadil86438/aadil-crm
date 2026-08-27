import api from './api'

export default {
  list(params = {}) { return api.get('/api/contacts', { params }) },
  listSimple() { return api.get('/api/contacts/simple') },
  get(id) { return api.get(`/api/contacts/${id}`) },
  create(data) { return api.post('/api/contacts', data) },
  update(id, data) { return api.put(`/api/contacts/${id}`, data) },
  delete(id) { return api.delete(`/api/contacts/${id}`) }
}
