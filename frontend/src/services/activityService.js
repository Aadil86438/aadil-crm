import api from './api'

export default {
  list(params = {}) { return api.get('/api/activities', { params }) },
  create(data) { return api.post('/api/activities', data) },
  delete(id) { return api.delete(`/api/activities/${id}`) },
  getCalendar() { return api.get('/api/calendar') }
}
