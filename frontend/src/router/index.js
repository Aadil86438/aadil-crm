import Vue from 'vue'
import VueRouter from 'vue-router'
import store from '../store'

Vue.use(VueRouter)

const routes = [
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/auth/Register.vue'),
    meta: { public: true }
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/auth/Login.vue'),
    meta: { public: true }
  },
  {
    path: '/admin-panel',
    name: 'AdminPanel',
    component: () => import('../views/auth/AdminPanel.vue'),
    meta: { public: true }
  },
  {
    path: '/',
    component: () => import('../layouts/MainLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      { path: '', redirect: 'dashboard' },
      { path: 'dashboard', name: 'Dashboard', component: () => import('../views/Dashboard.vue') },
      { path: 'tasks', name: 'Tasks', component: () => import('../views/tasks/TaskList.vue') },
      { path: 'activities', name: 'Activities', component: () => import('../views/activities/ActivityList.vue') },
      { path: 'calendar', name: 'Calendar', component: () => import('../views/calendar/CalendarView.vue') },
      { path: 'users', name: 'Users', component: () => import('../views/users/UserList.vue') },
      { path: 'audit-log', name: 'AuditLog', component: () => import('../views/admin/AuditLog.vue'), meta: { requiresAdmin: true } },
    ]
  },
  { path: '*', redirect: '/register' }
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes,
  scrollBehavior() {
    return { x: 0, y: 0 }
  }
})

router.beforeEach((to, from, next) => {
  const isAuthenticated = store.getters['auth/isAuthenticated']
  const isAdmin = store.getters['auth/isAdmin']

  if (to.meta.public) {
    if (isAuthenticated && (to.path === '/login' || to.path === '/register')) {
      return next('/dashboard')
    }
    return next()
  }

  if (to.matched.some(record => record.meta.requiresAuth) && !isAuthenticated) {
    return next('/register')
  }

  if (to.meta.requiresAdmin && !isAdmin) {
    return next('/dashboard')
  }

  next()
})

export default router
