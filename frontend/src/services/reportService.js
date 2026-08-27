import api from './api'

export default {
  sales() { return api.get('/api/reports/sales') },
  leads() { return api.get('/api/reports/leads') },
  activities() { return api.get('/api/reports/activities') },
  auditLogs(params = {}) { return api.get('/api/audit-logs', { params }) }
}
