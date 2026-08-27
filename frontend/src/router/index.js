import Vue from 'vue'
import VueRouter from 'vue-router'
import store from '../store'

Vue.use(VueRouter)

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/auth/Login.vue'),
    meta: { public: true }
  },
  {
    path: '/',
    component: () => import('../layouts/MainLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      { path: '', name: 'Dashboard', component: () => import('../views/Dashboard.vue') },
      { path: 'leads', name: 'Leads', component: () => import('../views/leads/LeadList.vue') },
      { path: 'leads/new', name: 'LeadCreate', component: () => import('../views/leads/LeadForm.vue') },
      { path: 'leads/:id', name: 'LeadDetail', component: () => import('../views/leads/LeadDetail.vue') },
      { path: 'leads/:id/edit', name: 'LeadEdit', component: () => import('../views/leads/LeadForm.vue') },
      { path: 'leads/:id/convert', name: 'LeadConvert', component: () => import('../views/leads/LeadConvert.vue') },
      { path: 'contacts', name: 'Contacts', component: () => import('../views/contacts/ContactList.vue') },
      { path: 'contacts/new', name: 'ContactCreate', component: () => import('../views/contacts/ContactForm.vue') },
      { path: 'contacts/:id', name: 'ContactDetail', component: () => import('../views/contacts/ContactDetail.vue') },
      { path: 'contacts/:id/edit', name: 'ContactEdit', component: () => import('../views/contacts/ContactForm.vue') },
      { path: 'accounts', name: 'Accounts', component: () => import('../views/accounts/AccountList.vue') },
      { path: 'accounts/new', name: 'AccountCreate', component: () => import('../views/accounts/AccountForm.vue') },
      { path: 'accounts/:id', name: 'AccountDetail', component: () => import('../views/accounts/AccountDetail.vue') },
      { path: 'accounts/:id/edit', name: 'AccountEdit', component: () => import('../views/accounts/AccountForm.vue') },
      { path: 'deals', name: 'Deals', component: () => import('../views/deals/DealList.vue') },
      { path: 'deals/pipeline', name: 'DealPipeline', component: () => import('../views/deals/DealKanban.vue') },
      { path: 'deals/new', name: 'DealCreate', component: () => import('../views/deals/DealForm.vue') },
      { path: 'deals/:id', name: 'DealDetail', component: () => import('../views/deals/DealDetail.vue') },
      { path: 'deals/:id/edit', name: 'DealEdit', component: () => import('../views/deals/DealForm.vue') },
      { path: 'tasks', name: 'Tasks', component: () => import('../views/tasks/TaskList.vue') },
      { path: 'activities', name: 'Activities', component: () => import('../views/activities/ActivityList.vue') },
      { path: 'calendar', name: 'Calendar', component: () => import('../views/calendar/CalendarView.vue') },
      { path: 'reports', name: 'Reports', component: () => import('../views/reports/Reports.vue') },
      { path: 'search', name: 'Search', component: () => import('../views/search/SearchResults.vue') },
      { path: 'users', name: 'Users', component: () => import('../views/users/UserList.vue'), meta: { requiresManager: true } },
      { path: 'audit-log', name: 'AuditLog', component: () => import('../views/admin/AuditLog.vue'), meta: { requiresAdmin: true } },
    ]
  },
  { path: '*', redirect: '/' }
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
  const isManager = store.getters['auth/isManager']

  if (to.meta.public) {
    if (isAuthenticated && to.path === '/login') {
      return next('/')
    }
    return next()
  }

  if (to.meta.requiresAuth && !isAuthenticated) {
    return next('/login')
  }

  if (to.meta.requiresAdmin && !isAdmin) {
    return next('/')
  }

  if (to.meta.requiresManager && !isManager) {
    return next('/')
  }

  next()
})

export default router
