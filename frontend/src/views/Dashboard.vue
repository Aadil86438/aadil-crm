<template>
  <div class="pa-4 pa-md-6">
    <!-- Page Header -->
    <div class="d-flex align-center mb-6">
      <div>
        <h1 class="text-h5 font-weight-bold">Employee Dashboard</h1>
        <p class="text-body-2 grey--text mb-0">Welcome back, {{ user && user.name }}! Here is your Employee Management overview.</p>
      </div>
      <v-spacer />
      <v-chip small color="primary" outlined>
        <v-icon left x-small>mdi-calendar</v-icon>
        {{ today }}
      </v-chip>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="text-center py-16">
      <v-progress-circular indeterminate color="primary" size="56" />
      <p class="text-body-2 grey--text mt-4">Loading employee dashboard...</p>
    </div>

    <template v-else>
      <!-- KPI Cards Row 1 -->
      <v-row class="mb-4">
        <v-col cols="12" sm="6" md="3" v-for="kpi in kpiCards" :key="kpi.label">
          <v-card class="kpi-card" :to="kpi.to" hover>
            <v-card-text class="pa-4">
              <div class="d-flex align-center justify-space-between mb-2">
                <v-icon :color="kpi.color" size="28">{{ kpi.icon }}</v-icon>
              </div>
              <div class="text-h4 font-weight-bold">{{ kpi.value }}</div>
              <div class="text-caption grey--text">{{ kpi.label }}</div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </template>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import dashboardService from '../services/dashboardService'
import userService from '../services/userService'

export default {
  name: 'Dashboard',
  data() {
    return {
      loading: true,
      stats: null,
      usersCount: 0,
    }
  },
  computed: {
    ...mapGetters('auth', ['user']),
    today() {
      return new Date().toLocaleDateString('en-IN', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })
    },
    kpiCards() {
      const s = this.stats || {}
      return [
        { label: 'Total Employees', value: this.usersCount || 1, icon: 'mdi-account-group', color: 'primary', to: '/users' },
        { label: 'Active Tasks', value: s.upcoming_tasks || 0, icon: 'mdi-checkbox-marked-circle', color: 'indigo', to: '/tasks' },
        { label: 'Today Activities', value: s.activities_today || 0, icon: 'mdi-timeline', color: 'teal', to: '/activities' },
        { label: 'My Role', value: (this.user && this.user.role) ? this.user.role.toUpperCase() : 'EMPLOYEE', icon: 'mdi-shield-account', color: 'purple', to: '/users' },
      ]
    }
  },
  methods: {
    async loadDashboard() {
      this.loading = true
      try {
        const [dashRes, userRes] = await Promise.allSettled([
          dashboardService.getStats(),
          userService.list()
        ])
        if (dashRes.status === 'fulfilled') {
          this.stats = dashRes.value.data.data
        }
        if (userRes.status === 'fulfilled') {
          this.usersCount = (userRes.value.data.data || []).length
        }
      } catch (err) {
        this.$store.dispatch('snackbar/error', 'Failed to load dashboard data')
      } finally {
        this.loading = false
      }
    }
  },
  mounted() {
    this.loadDashboard()
  }
}
</script>

<style scoped>
.kpi-card {
  border-radius: 12px !important;
  transition: transform 0.2s, box-shadow 0.2s;
  cursor: pointer;
  border-left: 3px solid transparent;
}
.kpi-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0,0,0,0.12) !important;
}
.chart-card { border-radius: 12px !important; }
.chart-container { position: relative; height: 220px; }
.h-100 { height: 100%; }
</style>
