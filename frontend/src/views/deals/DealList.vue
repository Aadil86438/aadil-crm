<template>
  <div class="pa-4 pa-md-6">
    <div class="d-flex flex-wrap align-center mb-4 gap-2">
      <div>
        <h1 class="text-h5 font-weight-bold">Deals</h1>
        <p class="text-body-2 grey--text mb-0">{{ total }} total deals</p>
      </div>
      <v-spacer />
      <v-btn outlined color="primary" to="/deals/pipeline" class="mr-2">
        <v-icon left small>mdi-view-column</v-icon>Pipeline View
      </v-btn>
      <v-btn color="primary" to="/deals/new" id="create-deal-btn">
        <v-icon left>mdi-plus</v-icon><span class="d-none d-sm-inline">New Deal</span>
      </v-btn>
    </div>

    <v-card class="mb-4 pa-3" elevation="0" outlined>
      <v-row dense>
        <v-col cols="12" sm="5" md="4">
          <v-text-field v-model="search" placeholder="Search deals..." prepend-inner-icon="mdi-magnify" dense outlined hide-details clearable @input="onSearchInput" id="deal-search" />
        </v-col>
        <v-col cols="6" sm="3" md="2">
          <v-select v-model="stageFilter" :items="stages" label="Stage" dense outlined hide-details clearable @change="load" />
        </v-col>
      </v-row>
    </v-card>

    <v-card elevation="0" outlined>
      <v-data-table :headers="headers" :items="deals" :loading="loading" :server-items-length="total" :options.sync="options" :footer-props="{ itemsPerPageOptions: [10, 20, 50] }" @update:options="load" id="deals-table">
        <template v-slot:item.name="{ item }">
          <router-link :to="`/deals/${item.id}`" class="text-body-2 font-weight-medium text-decoration-none primary--text">{{ item.name }}</router-link>
        </template>
        <template v-slot:item.amount="{ item }">
          <span class="font-weight-medium">{{ item.amount ? '₹' + Number(item.amount).toLocaleString() : '—' }}</span>
        </template>
        <template v-slot:item.stage="{ item }">
          <v-chip small :color="stageColor(item.stage)" dark>{{ item.stage }}</v-chip>
        </template>
        <template v-slot:item.probability="{ item }">
          <div class="d-flex align-center">
            <v-progress-linear :value="item.probability || 0" :color="probColor(item.probability)" height="6" rounded class="mr-2" style="width: 60px" />
            <span class="text-caption">{{ item.probability || 0 }}%</span>
          </div>
        </template>
        <template v-slot:item.expected_close_date="{ item }">
          <span class="text-caption" :class="isOverdue(item.expected_close_date) ? 'error--text' : ''">{{ formatDate(item.expected_close_date) }}</span>
        </template>
        <template v-slot:item.actions="{ item }">
          <v-btn icon x-small :to="`/deals/${item.id}`"><v-icon x-small>mdi-eye</v-icon></v-btn>
          <v-btn icon x-small :to="`/deals/${item.id}/edit`"><v-icon x-small>mdi-pencil</v-icon></v-btn>
          <v-btn icon x-small color="error" @click="confirmDelete(item)"><v-icon x-small>mdi-delete</v-icon></v-btn>
        </template>
        <template v-slot:no-data>
          <div class="text-center py-12">
            <v-icon size="48" color="grey lighten-2">mdi-briefcase</v-icon>
            <p class="grey--text mt-3">No deals found</p>
            <v-btn color="primary" to="/deals/new">Create Deal</v-btn>
          </div>
        </template>
      </v-data-table>
    </v-card>

    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title class="error--text"><v-icon left color="error">mdi-alert</v-icon>Delete Deal?</v-card-title>
        <v-card-text>Delete "{{ deleteTarget && deleteTarget.name }}"?</v-card-text>
        <v-card-actions><v-spacer /><v-btn text @click="deleteDialog = false">Cancel</v-btn><v-btn color="error" @click="deleteDeal" :loading="deleting">Delete</v-btn></v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import dealService from '../../services/dealService'
export default {
  name: 'DealList',
  data() {
    return {
      deals: [], total: 0, loading: false, search: '', searchTimer: null, stageFilter: null,
      options: { page: 1, itemsPerPage: 20, sortBy: ['created_at'], sortDesc: [true] },
      deleteDialog: false, deleteTarget: null, deleting: false,
      stages: ['Qualification', 'Needs Analysis', 'Proposal', 'Negotiation', 'Closed Won', 'Closed Lost'],
      headers: [
        { text: 'Deal Name', value: 'name' },
        { text: 'Amount', value: 'amount' },
        { text: 'Stage', value: 'stage' },
        { text: 'Probability', value: 'probability', sortable: false },
        { text: 'Account', value: 'account_name', sortable: false },
        { text: 'Close Date', value: 'expected_close_date' },
        { text: 'Actions', value: 'actions', sortable: false, align: 'center' },
      ]
    }
  },
  watch: { '$route.query': { handler: 'applyQuery', immediate: true } },
  methods: {
    applyQuery() { if (this.$route.query.stage) this.stageFilter = this.$route.query.stage },
    async load() {
      this.loading = true
      try {
        const { page, itemsPerPage, sortBy, sortDesc } = this.options
        const params = { page, page_size: itemsPerPage, sort: sortBy[0] || 'created_at', order: sortDesc[0] ? 'desc' : 'asc' }
        if (this.search) params.search = this.search
        if (this.stageFilter) params.stage = this.stageFilter
        const res = await dealService.list(params)
        this.deals = res.data.data || []; this.total = res.data.total || 0
      } finally { this.loading = false }
    },
    onSearchInput() { clearTimeout(this.searchTimer); this.searchTimer = setTimeout(() => { this.options.page = 1; this.load() }, 400) },
    confirmDelete(d) { this.deleteTarget = d; this.deleteDialog = true },
    async deleteDeal() {
      this.deleting = true
      try { await dealService.delete(this.deleteTarget.id); this.$store.dispatch('snackbar/success', 'Deal deleted'); this.deleteDialog = false; this.load() } finally { this.deleting = false }
    },
    stageColor(s) { const m = { Qualification: 'blue', 'Needs Analysis': 'cyan', Proposal: 'orange', Negotiation: 'purple', 'Closed Won': 'green', 'Closed Lost': 'red' }; return m[s] || 'grey' },
    probColor(p) { if (p >= 70) return 'green'; if (p >= 30) return 'orange'; return 'red' },
    isOverdue(d) { return d && new Date(d) < new Date() },
    formatDate(d) { return d ? new Date(d).toLocaleDateString('en-IN', { day: 'numeric', month: 'short', year: 'numeric' }) : '—' }
  }
}
</script>
<style scoped>.gap-2 { gap: 8px; }</style>
