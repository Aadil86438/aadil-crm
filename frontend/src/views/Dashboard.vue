<template>
  <div class="pa-4 pa-md-6">
    <!-- Page Header -->
    <div class="d-flex align-center mb-6">
      <div>
        <h1 class="text-h5 font-weight-bold">Dashboard</h1>
        <p class="text-body-2 grey--text mb-0">Welcome back, {{ user && user.name }}! Here's your CRM overview.</p>
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
      <p class="text-body-2 grey--text mt-4">Loading dashboard...</p>
    </div>

    <template v-else>
      <!-- KPI Cards Row 1 -->
      <v-row class="mb-4">
        <v-col cols="6" sm="4" md="3" lg="2" v-for="kpi in kpiCards" :key="kpi.label">
          <v-card class="kpi-card" :to="kpi.to" hover>
            <v-card-text class="pa-4">
              <div class="d-flex align-center justify-space-between mb-2">
                <v-icon :color="kpi.color" size="28">{{ kpi.icon }}</v-icon>
                <v-chip x-small :color="kpi.trend > 0 ? 'success' : 'error'" text-color="white" v-if="kpi.trend !== undefined">
                  <v-icon x-small left>{{ kpi.trend > 0 ? 'mdi-trending-up' : 'mdi-trending-down' }}</v-icon>
                  {{ Math.abs(kpi.trend) }}%
                </v-chip>
              </div>
              <div class="text-h5 font-weight-bold">{{ kpi.value }}</div>
              <div class="text-caption grey--text">{{ kpi.label }}</div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <!-- Charts Row -->
      <v-row class="mb-4">
        <!-- Leads by Status Doughnut -->
        <v-col cols="12" sm="6" md="4">
          <v-card class="chart-card h-100">
            <v-card-title class="pb-0 text-body-1 font-weight-bold">Leads by Status</v-card-title>
            <v-card-text>
              <div v-if="hasLeadsData" class="chart-container">
                <canvas ref="leadsStatusChart"></canvas>
              </div>
              <div v-else class="text-center py-8 grey--text">No lead data yet</div>
            </v-card-text>
          </v-card>
        </v-col>

        <!-- Deals by Stage -->
        <v-col cols="12" sm="6" md="4">
          <v-card class="chart-card h-100">
            <v-card-title class="pb-0 text-body-1 font-weight-bold">Deals by Stage</v-card-title>
            <v-card-text>
              <div v-if="hasDealsData" class="chart-container">
                <canvas ref="dealStageChart"></canvas>
              </div>
              <div v-else class="text-center py-8 grey--text">No deal data yet</div>
            </v-card-text>
          </v-card>
        </v-col>

        <!-- Monthly Revenue -->
        <v-col cols="12" md="4">
          <v-card class="chart-card h-100">
            <v-card-title class="pb-0 text-body-1 font-weight-bold">Monthly Revenue</v-card-title>
            <v-card-text>
              <div v-if="hasRevenueData" class="chart-container">
                <canvas ref="revenueChart"></canvas>
              </div>
              <div v-else class="text-center py-8 grey--text">No revenue data yet</div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <!-- Recent Data & Upcoming -->
      <v-row>
        <!-- Recent Leads -->
        <v-col cols="12" md="6">
          <v-card>
            <v-card-title class="text-body-1 font-weight-bold d-flex align-center">
              <v-icon left color="orange">mdi-account-arrow-right</v-icon>
              Recent Leads
              <v-spacer />
              <v-btn text x-small color="primary" to="/leads">View All</v-btn>
            </v-card-title>
            <v-divider />
            <v-list dense>
              <template v-if="recentLeads.length">
                <v-list-item v-for="lead in recentLeads" :key="lead.id" :to="`/leads/${lead.id}`" class="py-1">
                  <v-list-item-avatar size="32" :color="statusColor(lead.lead_status)">
                    <span class="white--text text-caption">{{ leadInitials(lead) }}</span>
                  </v-list-item-avatar>
                  <v-list-item-content>
                    <v-list-item-title class="text-body-2">{{ lead.first_name }} {{ lead.last_name }}</v-list-item-title>
                    <v-list-item-subtitle class="text-caption">{{ lead.company }}</v-list-item-subtitle>
                  </v-list-item-content>
                  <v-list-item-action>
                    <v-chip x-small :color="statusColor(lead.lead_status)" dark>{{ lead.lead_status }}</v-chip>
                  </v-list-item-action>
                </v-list-item>
              </template>
              <v-list-item v-else>
                <v-list-item-content class="text-center grey--text text-caption py-4">No recent leads</v-list-item-content>
              </v-list-item>
            </v-list>
          </v-card>
        </v-col>

        <!-- Recent Deals -->
        <v-col cols="12" md="6">
          <v-card>
            <v-card-title class="text-body-1 font-weight-bold d-flex align-center">
              <v-icon left color="purple">mdi-briefcase</v-icon>
              Recent Deals
              <v-spacer />
              <v-btn text x-small color="primary" to="/deals">View All</v-btn>
            </v-card-title>
            <v-divider />
            <v-list dense>
              <template v-if="recentDeals.length">
                <v-list-item v-for="deal in recentDeals" :key="deal.id" :to="`/deals/${deal.id}`" class="py-1">
                  <v-list-item-icon class="mr-3">
                    <v-icon :color="stageColor(deal.stage)" size="20">mdi-circle</v-icon>
                  </v-list-item-icon>
                  <v-list-item-content>
                    <v-list-item-title class="text-body-2">{{ deal.name }}</v-list-item-title>
                    <v-list-item-subtitle class="text-caption">{{ deal.stage }}</v-list-item-subtitle>
                  </v-list-item-content>
                  <v-list-item-action>
                    <span class="text-caption font-weight-bold">{{ formatAmount(deal.amount) }}</span>
                  </v-list-item-action>
                </v-list-item>
              </template>
              <v-list-item v-else>
                <v-list-item-content class="text-center grey--text text-caption py-4">No recent deals</v-list-item-content>
              </v-list-item>
            </v-list>
          </v-card>
        </v-col>
      </v-row>
    </template>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import dashboardService from '../services/dashboardService'
import Chart from 'chart.js/auto'

export default {
  name: 'Dashboard',
  data() {
    return {
      loading: true,
      stats: null,
      recentLeads: [],
      recentDeals: [],
      charts: [],
    }
  },
  computed: {
    ...mapGetters('auth', ['user']),
    today() {
      return new Date().toLocaleDateString('en-IN', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })
    },
    kpiCards() {
      if (!this.stats) return []
      const s = this.stats
      return [
        { label: 'Total Leads', value: s.total_leads || 0, icon: 'mdi-account-arrow-right', color: 'orange', to: '/leads', trend: 12 },
        { label: 'New Leads', value: s.new_leads || 0, icon: 'mdi-account-plus', color: 'blue', to: '/leads?status=New' },
        { label: 'Qualified', value: s.qualified_leads || 0, icon: 'mdi-star-check', color: 'green', to: '/leads?status=Qualified' },
        { label: 'Contacts', value: s.total_contacts || 0, icon: 'mdi-contacts', color: 'cyan', to: '/contacts' },
        { label: 'Accounts', value: s.total_accounts || 0, icon: 'mdi-office-building', color: 'teal', to: '/accounts' },
        { label: 'Open Deals', value: s.open_deals || 0, icon: 'mdi-briefcase-outline', color: 'purple', to: '/deals' },
        { label: 'Won Deals', value: s.won_deals || 0, icon: 'mdi-trophy', color: 'success', to: '/deals?stage=Closed Won', trend: 8 },
        { label: 'Lost Deals', value: s.lost_deals || 0, icon: 'mdi-close-circle', color: 'error', to: '/deals?stage=Closed Lost' },
        { label: 'Revenue', value: this.formatAmount(s.total_revenue), icon: 'mdi-currency-inr', color: 'green darken-2', to: '/reports', trend: 15 },
        { label: 'Pipeline', value: this.formatAmount(s.pipeline_value), icon: 'mdi-chart-line', color: 'indigo', to: '/deals' },
        { label: 'Due Today', value: s.activities_today || 0, icon: 'mdi-bell-ring', color: 'red', to: '/activities' },
        { label: 'Upcoming Tasks', value: s.upcoming_tasks || 0, icon: 'mdi-checkbox-marked-circle', color: 'blue-grey', to: '/tasks' },
      ]
    },
    hasLeadsData() {
      return this.stats && this.stats.leads_by_status && Object.keys(this.stats.leads_by_status).length > 0
    },
    hasDealsData() {
      return this.stats && this.stats.deals_by_stage && Object.keys(this.stats.deals_by_stage).length > 0
    },
    hasRevenueData() {
      return this.stats && this.stats.monthly_revenue && this.stats.monthly_revenue.length > 0
    }
  },
  methods: {
    formatAmount(val) {
      if (!val && val !== 0) return '—'
      if (val >= 10000000) return '₹' + (val / 10000000).toFixed(1) + 'Cr'
      if (val >= 100000) return '₹' + (val / 100000).toFixed(1) + 'L'
      if (val >= 1000) return '₹' + (val / 1000).toFixed(0) + 'K'
      return '₹' + val
    },
    leadInitials(l) {
      return ((l.first_name || '')[0] || '') + ((l.last_name || '')[0] || '')
    },
    statusColor(s) {
      const m = { New: 'blue', Contacted: 'orange', Qualified: 'green', Unqualified: 'grey', Converted: 'purple' }
      return m[s] || 'grey'
    },
    stageColor(s) {
      const m = { Qualification: 'blue', 'Needs Analysis': 'cyan', Proposal: 'orange', Negotiation: 'purple', 'Closed Won': 'green', 'Closed Lost': 'red' }
      return m[s] || 'grey'
    },
    destroyCharts() {
      this.charts.forEach(c => c.destroy())
      this.charts = []
    },
    renderCharts() {
      this.destroyCharts()
      this.$nextTick(() => {
        if (this.hasLeadsData && this.$refs.leadsStatusChart) {
          const data = this.stats.leads_by_status
          this.charts.push(new Chart(this.$refs.leadsStatusChart, {
            type: 'doughnut',
            data: {
              labels: Object.keys(data),
              datasets: [{ data: Object.values(data), backgroundColor: ['#1565C0','#FF8F00','#2E7D32','#757575','#6A1B9A'] }]
            },
            options: { responsive: true, maintainAspectRatio: true, plugins: { legend: { position: 'bottom', labels: { font: { size: 11 } } } } }
          }))
        }
        if (this.hasDealsData && this.$refs.dealStageChart) {
          const data = this.stats.deals_by_stage
          this.charts.push(new Chart(this.$refs.dealStageChart, {
            type: 'bar',
            data: {
              labels: Object.keys(data),
              datasets: [{ label: 'Deals', data: Object.values(data), backgroundColor: '#1565C0', borderRadius: 6 }]
            },
            options: { responsive: true, maintainAspectRatio: true, plugins: { legend: { display: false } }, scales: { y: { beginAtZero: true, ticks: { stepSize: 1 } }, x: { ticks: { font: { size: 10 } } } } }
          }))
        }
        if (this.hasRevenueData && this.$refs.revenueChart) {
          const data = this.stats.monthly_revenue
          this.charts.push(new Chart(this.$refs.revenueChart, {
            type: 'line',
            data: {
              labels: data.map(d => d.month),
              datasets: [{ label: 'Revenue', data: data.map(d => d.revenue), borderColor: '#1565C0', backgroundColor: 'rgba(21,101,192,0.1)', fill: true, tension: 0.4 }]
            },
            options: { responsive: true, maintainAspectRatio: true, plugins: { legend: { display: false } }, scales: { y: { beginAtZero: true } } }
          }))
        }
      })
    },
    async loadDashboard() {
      this.loading = true
      try {
        const res = await dashboardService.getStats()
        const data = res.data.data
        this.stats = data
        this.recentLeads = data.recent_leads || []
        this.recentDeals = data.recent_deals || []
        this.$nextTick(() => this.renderCharts())
      } catch (err) {
        this.$store.dispatch('snackbar/error', 'Failed to load dashboard data')
      } finally {
        this.loading = false
      }
    }
  },
  mounted() {
    this.loadDashboard()
  },
  beforeDestroy() {
    this.destroyCharts()
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
