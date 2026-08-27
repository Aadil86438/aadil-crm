<template>
  <div class="pa-4 pa-md-6">
    <div class="mb-6">
      <h1 class="text-h5 font-weight-bold">Audit Log</h1>
      <p class="text-body-2 grey--text mb-0">System activity trail for compliance and security auditing</p>
    </div>

    <v-card elevation="0" outlined>
      <v-data-table
        :headers="headers"
        :items="logs"
        :loading="loading"
        :server-items-length="total"
        :options.sync="options"
        :footer-props="{ itemsPerPageOptions: [20, 50, 100] }"
        @update:options="load"
        id="audit-logs-table"
      >
        <template v-slot:item.action="{ item }">
          <v-chip x-small :color="actionColor(item.action)" dark class="text-uppercase font-weight-bold">{{ item.action }}</v-chip>
        </template>

        <template v-slot:item.entity="{ item }">
          <span class="text-capitalize text-body-2 font-weight-medium">{{ item.entity }}</span>
        </template>

        <template v-slot:item.created_at="{ item }">
          <span class="text-caption">{{ formatDate(item.created_at) }}</span>
        </template>
      </v-data-table>
    </v-card>
  </div>
</template>

<script>
import reportService from '../../services/reportService'

export default {
  name: 'AuditLog',
  data() {
    return {
      logs: [], total: 0, loading: false,
      options: { page: 1, itemsPerPage: 20 },
      headers: [
        { text: 'Timestamp', value: 'created_at' },
        { text: 'User Email', value: 'user_email' },
        { text: 'Action', value: 'action' },
        { text: 'Entity', value: 'entity' },
        { text: 'Entity ID', value: 'entity_id', sortable: false },
        { text: 'IP Address', value: 'ip_address', sortable: false },
      ]
    }
  },
  methods: {
    async load() {
      this.loading = true
      try {
        const res = await reportService.auditLogs({ page: this.options.page, page_size: this.options.itemsPerPage })
        this.logs = res.data.data || []
        this.total = res.data.total || 0
      } finally { this.loading = false }
    },
    actionColor(a) {
      const m = { login: 'blue', logout: 'blue-grey', create: 'green', update: 'orange', delete: 'red', convert: 'purple', stage_change: 'indigo' }
      return m[a] || 'grey'
    },
    formatDate(d) { return d ? new Date(d).toLocaleString('en-IN') : '—' }
  }
}
</script>
