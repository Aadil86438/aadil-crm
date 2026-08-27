<template>
  <div class="pa-4 pa-md-6">
    <div v-if="loading" class="text-center py-16"><v-progress-circular indeterminate color="primary" /></div>
    <template v-else-if="account">
      <div class="d-flex flex-wrap align-center mb-6 gap-2">
        <v-btn icon @click="$router.back()"><v-icon>mdi-arrow-left</v-icon></v-btn>
        <v-avatar size="48" color="teal" class="ml-2"><v-icon color="white">mdi-office-building</v-icon></v-avatar>
        <div class="ml-3">
          <div class="d-flex align-center gap-2">
            <h1 class="text-h5 font-weight-bold">{{ account.name }}</h1>
            <v-chip small :color="typeColor(account.account_type)" dark>{{ account.account_type }}</v-chip>
          </div>
          <p class="text-body-2 grey--text mb-0">{{ account.industry }}</p>
        </div>
        <v-spacer />
        <v-btn outlined color="primary" :to="`/accounts/${account.id}/edit`"><v-icon left small>mdi-pencil</v-icon>Edit</v-btn>
        <v-btn color="error" outlined @click="deleteDialog = true"><v-icon left small>mdi-delete</v-icon>Delete</v-btn>
      </div>

      <v-row>
        <v-col cols="12" md="8">
          <v-card outlined elevation="0" class="mb-4">
            <v-card-title class="text-body-1 font-weight-bold">Account Details</v-card-title>
            <v-divider />
            <v-card-text>
              <v-row dense>
                <v-col cols="12" sm="6" v-for="f in fields" :key="f.label">
                  <div class="d-flex justify-space-between py-1 border-b">
                    <span class="text-caption grey--text">{{ f.label }}</span>
                    <span class="text-body-2 font-weight-medium">{{ f.value || '—' }}</span>
                  </div>
                </v-col>
              </v-row>
            </v-card-text>
          </v-card>
        </v-col>

        <v-col cols="12" md="4">
          <v-card outlined elevation="0" class="mb-3">
            <v-card-text>
              <div class="field-row mb-1"><span class="text-caption grey--text">Owner</span><span class="text-body-2">{{ account.owner_name || '—' }}</span></div>
              <v-divider class="my-2" />
              <div class="field-row mb-1"><span class="text-caption grey--text">Created</span><span class="text-body-2">{{ formatDate(account.created_at) }}</span></div>
              <div class="field-row"><span class="text-caption grey--text">Updated</span><span class="text-body-2">{{ formatDate(account.updated_at) }}</span></div>
            </v-card-text>
          </v-card>

          <!-- Quick Stats -->
          <v-card outlined elevation="0">
            <v-card-title class="text-body-2 font-weight-bold">Quick Links</v-card-title>
            <v-divider />
            <v-card-text>
              <v-btn block text small color="primary" :to="`/contacts?account_id=${account.id}`" class="justify-start"><v-icon left small>mdi-contacts</v-icon>View Contacts</v-btn>
              <v-btn block text small color="primary" :to="`/deals?account_id=${account.id}`" class="justify-start"><v-icon left small>mdi-briefcase</v-icon>View Deals</v-btn>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-dialog v-model="deleteDialog" max-width="400">
        <v-card>
          <v-card-title class="error--text"><v-icon left color="error">mdi-alert</v-icon>Delete Account?</v-card-title>
          <v-card-text>Delete "{{ account.name }}"?</v-card-text>
          <v-card-actions><v-spacer /><v-btn text @click="deleteDialog = false">Cancel</v-btn><v-btn color="error" @click="deleteAccount" :loading="deleting">Delete</v-btn></v-card-actions>
        </v-card>
      </v-dialog>
    </template>
  </div>
</template>

<script>
import accountService from '../../services/accountService'
export default {
  name: 'AccountDetail',
  data() { return { account: null, loading: true, deleteDialog: false, deleting: false } },
  computed: {
    fields() {
      if (!this.account) return []
      return [
        { label: 'Website', value: this.account.website }, { label: 'Phone', value: this.account.phone },
        { label: 'Email', value: this.account.email }, { label: 'Industry', value: this.account.industry },
        { label: 'Employees', value: this.account.num_employees }, { label: 'Annual Revenue', value: this.account.annual_revenue ? '₹' + this.account.annual_revenue.toLocaleString() : null },
        { label: 'City', value: this.account.city }, { label: 'Country', value: this.account.country },
        { label: 'Description', value: this.account.description }
      ]
    }
  },
  async mounted() {
    this.loading = true
    try { const res = await accountService.get(this.$route.params.id); this.account = res.data.data } finally { this.loading = false }
  },
  methods: {
    async deleteAccount() {
      this.deleting = true
      try { await accountService.delete(this.account.id); this.$store.dispatch('snackbar/success', 'Account deleted'); this.$router.push('/accounts') } finally { this.deleting = false }
    },
    formatDate(d) { return d ? new Date(d).toLocaleDateString('en-IN') : '—' },
    typeColor(t) { const m = { Customer: 'green', Prospect: 'blue', Partner: 'purple', Reseller: 'orange' }; return m[t] || 'grey' }
  }
}
</script>
<style scoped>
.gap-2 { gap: 8px; }
.field-row { display: flex; justify-content: space-between; }
.border-b { border-bottom: 1px solid #f0f0f0; }
</style>
