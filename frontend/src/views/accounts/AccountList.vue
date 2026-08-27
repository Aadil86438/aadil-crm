<template>
  <div class="pa-4 pa-md-6">
    <div class="d-flex flex-wrap align-center mb-4 gap-2">
      <div>
        <h1 class="text-h5 font-weight-bold">Accounts</h1>
        <p class="text-body-2 grey--text mb-0">{{ total }} total accounts</p>
      </div>
      <v-spacer />
      <v-btn color="primary" to="/accounts/new" id="create-account-btn">
        <v-icon left>mdi-plus</v-icon><span class="d-none d-sm-inline">New Account</span>
      </v-btn>
    </div>

    <v-card class="mb-4 pa-3" elevation="0" outlined>
      <v-row dense>
        <v-col cols="12" sm="6" md="5">
          <v-text-field v-model="search" placeholder="Search accounts..." prepend-inner-icon="mdi-magnify" dense outlined hide-details clearable @input="onSearchInput" id="account-search" />
        </v-col>
        <v-col cols="6" sm="3" md="2">
          <v-select v-model="typeFilter" :items="accountTypes" label="Type" dense outlined hide-details clearable @change="loadAccounts" />
        </v-col>
      </v-row>
    </v-card>

    <v-card elevation="0" outlined>
      <v-data-table :headers="headers" :items="accounts" :loading="loading" :server-items-length="total" :options.sync="options" :footer-props="{ itemsPerPageOptions: [10, 20, 50] }" @update:options="loadAccounts" id="accounts-table">
        <template v-slot:item.name="{ item }">
          <div class="d-flex align-center py-1">
            <v-avatar size="32" color="teal" class="mr-3"><v-icon small color="white">mdi-office-building</v-icon></v-avatar>
            <div>
              <router-link :to="`/accounts/${item.id}`" class="text-body-2 font-weight-medium text-decoration-none primary--text">{{ item.name }}</router-link>
              <div class="text-caption grey--text">{{ item.industry }}</div>
            </div>
          </div>
        </template>
        <template v-slot:item.account_type="{ item }">
          <v-chip x-small :color="typeColor(item.account_type)" dark>{{ item.account_type }}</v-chip>
        </template>
        <template v-slot:item.annual_revenue="{ item }">
          <span class="text-body-2">{{ item.annual_revenue ? '₹' + (item.annual_revenue / 100000).toFixed(0) + 'L' : '—' }}</span>
        </template>
        <template v-slot:item.created_at="{ item }"><span class="text-caption">{{ formatDate(item.created_at) }}</span></template>
        <template v-slot:item.actions="{ item }">
          <v-btn icon x-small :to="`/accounts/${item.id}`"><v-icon x-small>mdi-eye</v-icon></v-btn>
          <v-btn icon x-small :to="`/accounts/${item.id}/edit`"><v-icon x-small>mdi-pencil</v-icon></v-btn>
          <v-btn icon x-small color="error" @click="confirmDelete(item)"><v-icon x-small>mdi-delete</v-icon></v-btn>
        </template>
        <template v-slot:no-data>
          <div class="text-center py-12">
            <v-icon size="48" color="grey lighten-2">mdi-office-building</v-icon>
            <p class="grey--text mt-3">No accounts found</p>
            <v-btn color="primary" to="/accounts/new">Create Account</v-btn>
          </div>
        </template>
      </v-data-table>
    </v-card>

    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title class="error--text"><v-icon left color="error">mdi-alert</v-icon>Delete Account?</v-card-title>
        <v-card-text>Delete "{{ deleteTarget && deleteTarget.name }}"?</v-card-text>
        <v-card-actions><v-spacer /><v-btn text @click="deleteDialog = false">Cancel</v-btn><v-btn color="error" @click="deleteAccount" :loading="deleting">Delete</v-btn></v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import accountService from '../../services/accountService'
export default {
  name: 'AccountList',
  data() {
    return {
      accounts: [], total: 0, loading: false, search: '', searchTimer: null, typeFilter: null,
      options: { page: 1, itemsPerPage: 20, sortBy: ['created_at'], sortDesc: [true] },
      deleteDialog: false, deleteTarget: null, deleting: false,
      accountTypes: ['Customer', 'Prospect', 'Partner', 'Reseller', 'Other'],
      headers: [
        { text: 'Account', value: 'name', sortable: false },
        { text: 'Type', value: 'account_type' },
        { text: 'Phone', value: 'phone', sortable: false },
        { text: 'Revenue', value: 'annual_revenue', sortable: false },
        { text: 'Owner', value: 'owner_name', sortable: false },
        { text: 'Created', value: 'created_at' },
        { text: 'Actions', value: 'actions', sortable: false, align: 'center' },
      ]
    }
  },
  methods: {
    async loadAccounts() {
      this.loading = true
      try {
        const { page, itemsPerPage, sortBy, sortDesc } = this.options
        const params = { page, page_size: itemsPerPage, sort: sortBy[0] || 'created_at', order: sortDesc[0] ? 'desc' : 'asc' }
        if (this.search) params.search = this.search
        if (this.typeFilter) params.account_type = this.typeFilter
        const res = await accountService.list(params)
        this.accounts = res.data.data || []; this.total = res.data.total || 0
      } finally { this.loading = false }
    },
    onSearchInput() { clearTimeout(this.searchTimer); this.searchTimer = setTimeout(() => { this.options.page = 1; this.loadAccounts() }, 400) },
    confirmDelete(a) { this.deleteTarget = a; this.deleteDialog = true },
    async deleteAccount() {
      this.deleting = true
      try { await accountService.delete(this.deleteTarget.id); this.$store.dispatch('snackbar/success', 'Account deleted'); this.deleteDialog = false; this.loadAccounts() } finally { this.deleting = false }
    },
    formatDate(d) { return d ? new Date(d).toLocaleDateString('en-IN', { day: 'numeric', month: 'short', year: 'numeric' }) : '—' },
    typeColor(t) { const m = { Customer: 'green', Prospect: 'blue', Partner: 'purple', Reseller: 'orange', Other: 'grey' }; return m[t] || 'grey' }
  }
}
</script>
<style scoped>.gap-2 { gap: 8px; }</style>
