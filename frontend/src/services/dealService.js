import api from './api'

export default {
  list(params = {}) { return api.get('/api/deals', { params }) },
  get(id) { return api.get(`/api/deals/${id}`) },
  create(data) { return api.post('/api/deals', data) },
  update(id, data) { return api.put(`/api/deals/${id}`, data) },
  updateStage(id, stage) { return api.patch(`/api/deals/${id}/stage`, { stage }) },
  delete(id) { return api.delete(`/api/deals/${id}`) },
  getPipeline() { return api.get('/api/deals/pipeline') }
}
