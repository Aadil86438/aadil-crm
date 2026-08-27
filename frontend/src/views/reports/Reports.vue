<template>
  <div class="pa-4 pa-md-6">
    <div class="mb-6">
      <h1 class="text-h5 font-weight-bold">Analytics & Reports</h1>
      <p class="text-body-2 grey--text mb-0">Insights across sales pipeline, leads, and team activities</p>
    </div>

    <v-tabs v-model="tab" color="primary" class="mb-4">
      <v-tab>Sales Performance</v-tab>
      <v-tab>Lead Analytics</v-tab>
      <v-tab>Activity Overview</v-tab>
    </v-tabs>

    <v-tabs-items v-model="tab">
      <!-- Sales Tab -->
      <v-tab-item>
        <v-row class="mt-2">
          <v-col cols="12" md="6">
            <v-card outlined elevation="0" class="pa-4">
              <h3 class="text-subtitle-1 font-weight-bold mb-4">Pipeline Summary</h3>
              <v-list dense v-if="salesData && salesData.summary">
                <v-list-item><v-list-item-content><v-list-item-title>Total Deals</v-list-item-title></v-list-item-content><v-list-item-action class="font-weight-bold">{{ salesData.summary.total_deals }}</v-list-item-action></v-list-item>
                <v-list-item><v-list-item-content><v-list-item-title>Total Value</v-list-item-title></v-list-item-content><v-list-item-action class="font-weight-bold primary--text">₹{{ (salesData.summary.total_value || 0).toLocaleString() }}</v-list-item-action></v-list-item>
                <v-list-item><v-list-item-content><v-list-item-title>Won Revenue</v-list-item-title></v-list-item-content><v-list-item-action class="font-weight-bold success--text">₹{{ (salesData.summary.won_revenue || 0).toLocaleString() }}</v-list-item-action></v-list-item>
                <v-list-item><v-list-item-content><v-list-item-title>Win Rate</v-list-item-title></v-list-item-content><v-list-item-action class="font-weight-bold">{{ (salesData.summary.win_rate || 0).toFixed(1) }}%</v-list-item-action></v-list-item>
              </v-list>
            </v-card>
          </v-col>
        </v-row>
      </v-tab-item>

      <!-- Lead Analytics Tab -->
      <v-tab-item>
        <v-row class="mt-2">
          <v-col cols="12" md="6">
            <v-card outlined elevation="0" class="pa-4">
              <h3 class="text-subtitle-1 font-weight-bold mb-4">Leads by Status</h3>
              <div v-for="(val, key) in (leadData ? leadData.leads_by_status : {})" :key="key" class="d-flex justify-space-between py-2 border-b">
                <span class="text-body-2">{{ key }}</span>
                <span class="font-weight-bold">{{ val }}</span>
              </div>
            </v-card>
          </v-col>
          <v-col cols="12" md="6">
            <v-card outlined elevation="0" class="pa-4">
              <h3 class="text-subtitle-1 font-weight-bold mb-4">Leads by Source</h3>
              <div v-for="(val, key) in (leadData ? leadData.leads_by_source : {})" :key="key" class="d-flex justify-space-between py-2 border-b">
                <span class="text-body-2">{{ key }}</span>
                <span class="font-weight-bold">{{ val }}</span>
              </div>
            </v-card>
          </v-col>
        </v-row>
      </v-tab-item>

      <!-- Activity Overview Tab -->
      <v-tab-item>
        <v-row class="mt-2">
          <v-col cols="12" md="6">
            <v-card outlined elevation="0" class="pa-4">
              <h3 class="text-subtitle-1 font-weight-bold mb-4">Activities by Type</h3>
              <div v-for="(val, key) in (activityData ? activityData.activities_by_type : {})" :key="key" class="d-flex justify-space-between py-2 border-b">
                <span class="text-body-2 text-capitalize">{{ key }}</span>
                <span class="font-weight-bold">{{ val }}</span>
              </div>
            </v-card>
          </v-col>
        </v-row>
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>

<script>
import reportService from '../../services/reportService'

export default {
  name: 'Reports',
  data() {
    return {
      tab: 0,
      salesData: null,
      leadData: null,
      activityData: null,
    }
  },
  async mounted() {
    try {
      const [s, l, a] = await Promise.all([
        reportService.sales(),
        reportService.leads(),
        reportService.activities()
      ])
      this.salesData = s.data.data
      this.leadData = l.data.data
      this.activityData = a.data.data
    } catch (e) { /* ignore */ }
  }
}
</script>
<style scoped>.border-b { border-bottom: 1px solid #f0f0f0; }</style>
