<template>
  <div class="pa-4 pa-md-6">
    <!-- Page Header -->
    <div class="d-flex flex-wrap align-center mb-4 gap-2">
      <div>
        <h1 class="text-h5 font-weight-bold">Leads</h1>
        <p class="text-body-2 grey--text mb-0">{{ total }} total leads</p>
      </div>
      <v-spacer />
      <v-btn color="primary" @click="$router.push('/leads/new')" id="create-lead-btn">
        <v-icon left>mdi-plus</v-icon>
        <span class="d-none d-sm-inline">New Lead</span>
      </v-btn>
    </div>

    <!-- Filters -->
    <v-card class="mb-4 pa-3" elevation="0" outlined>
      <v-row dense>
        <v-col cols="12" sm="5" md="4">
          <v-text-field
            v-model="search"
            placeholder="Search by name, company, email..."
            prepend-inner-icon="mdi-magnify"
            dense
            outlined
            hide-details
            clearable
            @input="onSearchInput"
            id="lead-search"
          />
        </v-col>
        <v-col cols="6" sm="3" md="2">
          <v-select
            v-model="filters.status"
            :items="leadStatuses"
            label="Status"
            dense
            outlined
            hide-details
            clearable
            @change="loadLeads"
          />
        </v-col>
        <v-col cols="6" sm="3" md="2">
          <v-select
            v-model="filters.source"
            :items="leadSources"
            label="Source"
            dense
            outlined
            hide-details
            clearable
            @change="loadLeads"
          />
        </v-col>
        <v-col cols="12" sm="1" class="d-flex align-center">
          <v-btn text small @click="clearFilters" color="grey">
            <v-icon small>mdi-filter-remove</v-icon>
          </v-btn>
        </v-col>
      </v-row>
    </v-card>

    <!-- Data Table -->
    <v-card elevation="0" outlined>
      <v-data-table
        v-model="selected"
        :headers="headers"
        :items="leads"
        :loading="loading"
        :server-items-length="total"
        :options.sync="tableOptions"
        :footer-props="{ itemsPerPageOptions: [10, 20, 50], showCurrentPage: true }"
        show-select
        must-sort
        @update:options="loadLeads"
        class="leads-table"
        id="leads-table"
      >
        <!-- Name Cell -->
        <template v-slot:item.name="{ item }">
          <div class="d-flex align-center py-1">
            <v-avatar size="32" :color="statusColors[item.lead_status] || 'grey'" class="mr-3">
              <span class="white--text text-caption font-weight-bold">{{ initials(item) }}</span>
            </v-avatar>
            <div>
              <router-link :to="`/leads/${item.id}`" class="text-body-2 font-weight-medium text-decoration-none primary--text">
                {{ item.first_name }} {{ item.last_name }}
              </router-link>
              <div class="text-caption grey--text">{{ item.company }}</div>
            </div>
          </div>
        </template>

        <!-- Status Cell -->
        <template v-slot:item.lead_status="{ item }">
          <v-chip small :color="statusColors[item.lead_status]" dark class="font-weight-medium">
            {{ item.lead_status }}
          </v-chip>
        </template>

        <!-- Source Cell -->
        <template v-slot:item.lead_source="{ item }">
          <v-chip x-small outlined>{{ item.lead_source || '—' }}</v-chip>
        </template>

        <!-- Owner Cell -->
        <template v-slot:item.owner_name="{ item }">
          <span class="text-body-2">{{ item.owner_name || '—' }}</span>
        </template>

        <!-- Date Cell -->
        <template v-slot:item.created_at="{ item }">
          <span class="text-caption">{{ formatDate(item.created_at) }}</span>
        </template>

        <!-- Actions Cell -->
        <template v-slot:item.actions="{ item }">
          <div class="d-flex">
            <v-tooltip bottom>
              <template v-slot:activator="{ on }">
                <v-btn icon x-small v-on="on" :to="`/leads/${item.id}`">
                  <v-icon x-small>mdi-eye</v-icon>
                </v-btn>
              </template>
              <span>View</span>
            </v-tooltip>
            <v-tooltip bottom>
              <template v-slot:activator="{ on }">
                <v-btn icon x-small v-on="on" :to="`/leads/${item.id}/edit`">
                  <v-icon x-small>mdi-pencil</v-icon>
                </v-btn>
              </template>
              <span>Edit</span>
            </v-tooltip>
            <v-tooltip bottom v-if="!item.is_converted && item.lead_status === 'Qualified'">
              <template v-slot:activator="{ on }">
                <v-btn icon x-small v-on="on" color="success" :to="`/leads/${item.id}/convert`">
                  <v-icon x-small>mdi-account-convert</v-icon>
                </v-btn>
              </template>
              <span>Convert</span>
            </v-tooltip>
            <v-tooltip bottom>
              <template v-slot:activator="{ on }">
                <v-btn icon x-small color="error" v-on="on" @click="confirmDelete(item)">
                  <v-icon x-small>mdi-delete</v-icon>
                </v-btn>
              </template>
              <span>Delete</span>
            </v-tooltip>
          </div>
        </template>

        <!-- Empty State -->
        <template v-slot:no-data>
          <div class="text-center py-16">
            <v-icon size="64" color="grey lighten-2">mdi-account-arrow-right</v-icon>
            <p class="text-h6 grey--text mt-4">No leads found</p>
            <p class="text-body-2 grey--text mb-4">{{ search ? 'Try a different search term' : 'Start by creating your first lead' }}</p>
            <v-btn color="primary" @click="$router.push('/leads/new')">
              <v-icon left>mdi-plus</v-icon>Create Lead
            </v-btn>
          </div>
        </template>

        <!-- Loading Overlay -->
        <template v-slot:loading>
          <div class="text-center py-8">
            <v-progress-circular indeterminate color="primary" />
          </div>
        </template>
      </v-data-table>

      <!-- Bulk Actions Bar -->
      <v-expand-transition>
        <div v-if="selected.length > 0" class="bulk-bar pa-3 d-flex align-center">
          <span class="text-body-2 mr-4">{{ selected.length }} selected</span>
          <v-btn small color="error" outlined @click="bulkDelete">
            <v-icon left small>mdi-delete</v-icon>Delete Selected
          </v-btn>
          <v-btn small class="ml-2" outlined @click="selected = []">Clear</v-btn>
        </div>
      </v-expand-transition>
    </v-card>

    <!-- Confirm Delete Dialog -->
    <v-dialog v-model="deleteDialog" max-width="420">
      <v-card>
        <v-card-title class="error--text">
          <v-icon left color="error">mdi-alert</v-icon>
          Delete Lead?
        </v-card-title>
        <v-card-text>
          Are you sure you want to delete <strong>{{ deleteTarget && deleteTarget.first_name }} {{ deleteTarget && deleteTarget.last_name }}</strong>?
          This action cannot be undone.
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn text @click="deleteDialog = false">Cancel</v-btn>
          <v-btn color="error" @click="deleteLead" :loading="deleting">Delete</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import leadService from '../../services/leadService'

export default {
  name: 'LeadList',
  data() {
    return {
      leads: [],
      total: 0,
      loading: false,
      search: '',
      selected: [],
      searchTimer: null,
      filters: { status: null, source: null },
      tableOptions: { page: 1, itemsPerPage: 20, sortBy: ['created_at'], sortDesc: [true] },
      deleteDialog: false,
      deleteTarget: null,
      deleting: false,
      headers: [
        { text: 'Name', value: 'name', sortable: false },
        { text: 'Email', value: 'email' },
        { text: 'Phone', value: 'phone', sortable: false },
        { text: 'Status', value: 'lead_status' },
        { text: 'Source', value: 'lead_source', sortable: false },
        { text: 'Owner', value: 'owner_name', sortable: false },
        { text: 'Created', value: 'created_at' },
        { text: 'Actions', value: 'actions', sortable: false, align: 'center', width: 120 },
      ],
      leadStatuses: ['New', 'Contacted', 'Qualified', 'Unqualified', 'Converted'],
      leadSources: ['Website', 'Referral', 'Advertisement', 'Cold Call', 'Email', 'Social Media', 'Campaign', 'Other'],
      statusColors: {
        New: 'blue', Contacted: 'orange', Qualified: 'green', Unqualified: 'grey', Converted: 'purple'
      }
    }
  },
  watch: {
    '$route.query': { handler: 'applyQueryFilters', immediate: true }
  },
  methods: {
    applyQueryFilters() {
      if (this.$route.query.status) this.filters.status = this.$route.query.status
    },
    async loadLeads() {
      this.loading = true
      try {
        const { page, itemsPerPage, sortBy, sortDesc } = this.tableOptions
        const params = {
          page, page_size: itemsPerPage,
          sort: sortBy[0] || 'created_at',
          order: sortDesc[0] ? 'desc' : 'asc',
        }
        if (this.search) params.search = this.search
        if (this.filters.status) params.status = this.filters.status
        if (this.filters.source) params.source = this.filters.source

        const res = await leadService.list(params)
        this.leads = res.data.data || []
        this.total = res.data.total || 0
      } catch (err) {
        this.$store.dispatch('snackbar/error', 'Failed to load leads')
      } finally {
        this.loading = false
      }
    },
    onSearchInput() {
      clearTimeout(this.searchTimer)
      this.searchTimer = setTimeout(() => {
        this.tableOptions.page = 1
        this.loadLeads()
      }, 400)
    },
    clearFilters() {
      this.search = ''
      this.filters = { status: null, source: null }
      this.loadLeads()
    },
    initials(lead) {
      return ((lead.first_name || '')[0] || '') + ((lead.last_name || '')[0] || '')
    },
    formatDate(d) {
      if (!d) return '—'
      return new Date(d).toLocaleDateString('en-IN', { day: 'numeric', month: 'short', year: 'numeric' })
    },
    confirmDelete(lead) {
      this.deleteTarget = lead
      this.deleteDialog = true
    },
    async deleteLead() {
      if (!this.deleteTarget) return
      this.deleting = true
      try {
        await leadService.delete(this.deleteTarget.id)
        this.$store.dispatch('snackbar/success', 'Lead deleted successfully')
        this.deleteDialog = false
        this.loadLeads()
      } catch (err) {
        this.$store.dispatch('snackbar/error', 'Failed to delete lead')
      } finally {
        this.deleting = false
      }
    },
    async bulkDelete() {
      if (!confirm(`Delete ${this.selected.length} leads?`)) return
      try {
        await Promise.all(this.selected.map(l => leadService.delete(l.id)))
        this.$store.dispatch('snackbar/success', `${this.selected.length} leads deleted`)
        this.selected = []
        this.loadLeads()
      } catch (err) {
        this.$store.dispatch('snackbar/error', 'Bulk delete failed')
      }
    }
  }
}
</script>

<style scoped>
.leads-table >>> .v-data-table__wrapper { overflow-x: auto; }
.bulk-bar { background: #E3F2FD; border-top: 1px solid #BBDEFB; }
.gap-2 { gap: 8px; }
</style>
