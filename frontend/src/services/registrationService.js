import api from './api'
import axios from 'axios'

const baseURL = process.env.VUE_APP_API_URL || ''

export default {
  register(data) { return api.post('/api/auth/register', data) },
  submitPayment(data) { return api.post('/api/auth/submit-payment', data) },
  checkStatus(id) { return api.get(`/api/auth/registration-status/${id}`) },
  
  // Admin panel endpoints (uses admin token, not user token)
  verifyAdminCode(code) { return api.post('/api/admin/verify', { code }) },
  getPending(token) {
    return axios.get(`${baseURL}/api/admin/pending`, {
      headers: { Authorization: `Bearer ${token}` }
    })
  },
  getAll(token) {
    return axios.get(`${baseURL}/api/admin/all`, {
      headers: { Authorization: `Bearer ${token}` }
    })
  },
  approveRequest(id, token) {
    return axios.post(`${baseURL}/api/admin/approve/${id}`, {}, {
      headers: { Authorization: `Bearer ${token}` }
    })
  },
  rejectRequest(id, token) {
    return axios.post(`${baseURL}/api/admin/reject/${id}`, {}, {
      headers: { Authorization: `Bearer ${token}` }
    })
  },
  getRedisData(token) {
    return axios.get(`${baseURL}/api/admin/redis`, {
      headers: { Authorization: `Bearer ${token}` }
    })
  },
  deleteRedisKey(key, token) {
    return axios.delete(`${baseURL}/api/admin/redis?key=${encodeURIComponent(key)}`, {
      headers: { Authorization: `Bearer ${token}` }
    })
  }
}
