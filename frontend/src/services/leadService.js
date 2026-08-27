import api from './api'

export default {
  list(params = {}) {
    return api.get('/api/leads', { params })
  },
  get(id) {
    return api.get(`/api/leads/${id}`)
  },
  create(data) {
    return api.post('/api/leads', data)
  },
  update(id, data) {
    return api.put(`/api/leads/${id}`, data)
  },
  delete(id) {
    return api.delete(`/api/leads/${id}`)
  },
  convert(id, data) {
    return api.post(`/api/leads/${id}/convert`, data)
  }
}
