import axios from 'axios'

// Admin endpoints target the GCP Admin Service (port 8081 or GCP VM IP)
// If VUE_APP_ADMIN_API_URL is set, use it; otherwise fallback to VUE_APP_API_URL or relative
const adminBaseURL = (process.env.VUE_APP_ADMIN_API_URL || process.env.VUE_APP_API_URL || '').replace(/\/$/, '')
const coreBaseURL = (process.env.VUE_APP_API_URL || '').replace(/\/$/, '')

export default {
  register(data) { 
    return axios.post(`${coreBaseURL}/api/auth/register`, data) 
  },

  // Admin panel endpoints (uses admin token, targets admin-service-gcp)
  verifyAdminCode(code) { 
    return axios.post(`${adminBaseURL}/api/admin/verify`, { code }) 
  },
  
  // Cross-Cloud Proxy endpoints (calls admin-service-gcp -> core-service-aws)
  getCoreStats(token) {
    return axios.get(`${adminBaseURL}/api/admin/core/stats`, {
      headers: { Authorization: `Bearer ${token}` }
    })
  },
  getCoreUsers(token) {
    return axios.get(`${adminBaseURL}/api/admin/core/users`, {
      headers: { Authorization: `Bearer ${token}` }
    })
  },
  getCoreAuditLogs(page = 1, pageSize = 20, token) {
    return axios.get(`${adminBaseURL}/api/admin/core/audit-logs?page=${page}&page_size=${pageSize}`, {
      headers: { Authorization: `Bearer ${token}` }
    })
  },

  // Legacy compatibility: getAll returns users proxied from core-service-aws
  getAll(token) {
    return axios.get(`${adminBaseURL}/api/admin/all`, {
      headers: { Authorization: `Bearer ${token}` }
    })
  },
  getRedisData(token) {
    return axios.get(`${adminBaseURL}/api/admin/redis`, {
      headers: { Authorization: `Bearer ${token}` }
    })
  },
  deleteRedisKey(key, token) {
    return axios.delete(`${adminBaseURL}/api/admin/redis?key=${encodeURIComponent(key)}`, {
      headers: { Authorization: `Bearer ${token}` }
    })
  },
  getK8sStatus(token) {
    return axios.get(`${adminBaseURL}/api/admin/k8s/status`, {
      headers: { Authorization: `Bearer ${token}` }
    })
  },
  killK8sPod(podName, token) {
    return axios.post(`${adminBaseURL}/api/admin/k8s/kill-pod`, { podName }, {
      headers: { Authorization: `Bearer ${token}` }
    })
  },

  // Log Viewer endpoints (targets admin-service-gcp)
  getLogFiles(token) {
    return axios.get(`${adminBaseURL}/api/admin/logs`, {
      headers: { Authorization: `Bearer ${token}` }
    })
  },
  viewLogFile(date, token) {
    return axios.get(`${adminBaseURL}/api/admin/logs/view?date=${date}`, {
      headers: { Authorization: `Bearer ${token}` }
    })
  },
  getLogDownloadURL(date, token) {
    return `${adminBaseURL}/api/admin/logs/download?date=${date}`
  },

  // User Approval & Rejection (calls admin proxy or core service directly)
  approveUser(id, token) {
    const url = adminBaseURL ? `${adminBaseURL}/api/admin/users/${id}/approve` : `${coreBaseURL}/api/users/${id}/approve`
    return axios.post(url, {}, {
      headers: { Authorization: `Bearer ${token}` }
    })
  },
  rejectUser(id, token) {
    const url = adminBaseURL ? `${adminBaseURL}/api/admin/users/${id}/reject` : `${coreBaseURL}/api/users/${id}/reject`
    return axios.post(url, {}, {
      headers: { Authorization: `Bearer ${token}` }
    })
  },

  // Real-time live log stream SSE URL
  getLiveLogsStreamURL(token) {
    if (adminBaseURL) {
      return `${adminBaseURL}/api/admin/logs/live?token=${encodeURIComponent(token || '')}`
    }
    return `${coreBaseURL}/api/logs/live`
  }
}
