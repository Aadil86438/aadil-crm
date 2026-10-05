import api from './api'

export default {
  auditLogs(params = {}) { return api.get('/api/audit-logs', { params }) }
}
