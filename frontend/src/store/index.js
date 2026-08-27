import Vue from 'vue'
import Vuex from 'vuex'
import api from '../services/api'
import router from '../router'

Vue.use(Vuex)

export default new Vuex.Store({
  modules: {
    auth: {
      namespaced: true,
      state: {
        token: localStorage.getItem('crm_token') || null,
        user: JSON.parse(localStorage.getItem('crm_user') || 'null'),
      },
      getters: {
        token: state => state.token,
        user: state => state.user,
        isAuthenticated: state => !!state.token,
        isAdmin: state => state.user && state.user.role === 'admin',
        isManager: state => state.user && ['admin', 'manager'].includes(state.user.role),
      },
      mutations: {
        SET_AUTH(state, { token, user }) {
          state.token = token
          state.user = user
          localStorage.setItem('crm_token', token)
          localStorage.setItem('crm_user', JSON.stringify(user))
        },
        CLEAR_AUTH(state) {
          state.token = null
          state.user = null
          localStorage.removeItem('crm_token')
          localStorage.removeItem('crm_user')
        },
        UPDATE_USER(state, user) {
          state.user = user
          localStorage.setItem('crm_user', JSON.stringify(user))
        }
      },
      actions: {
        async login({ commit }, credentials) {
          const response = await api.post('/api/auth/login', credentials)
          const { token, user } = response.data.data
          commit('SET_AUTH', { token, user })
          return user
        },
        async logout({ commit }) {
          try {
            await api.post('/api/auth/logout')
          } catch (e) { /* ignore */ }
          commit('CLEAR_AUTH')
          router.push('/login')
        },
        async fetchCurrentUser({ commit }) {
          const response = await api.get('/api/auth/me')
          commit('UPDATE_USER', response.data.data)
          return response.data.data
        }
      }
    },
    snackbar: {
      namespaced: true,
      state: {
        show: false,
        message: '',
        color: 'success',
        timeout: 3000,
      },
      mutations: {
        SHOW(state, { message, color = 'success', timeout = 3000 }) {
          state.message = message
          state.color = color
          state.timeout = timeout
          state.show = true
        },
        HIDE(state) {
          state.show = false
        }
      },
      actions: {
        show({ commit }, payload) {
          commit('SHOW', payload)
        },
        success({ commit }, message) {
          commit('SHOW', { message, color: 'success' })
        },
        error({ commit }, message) {
          commit('SHOW', { message, color: 'error', timeout: 5000 })
        },
        warning({ commit }, message) {
          commit('SHOW', { message, color: 'warning' })
        }
      }
    }
  }
})
