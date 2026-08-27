<template>
  <div class="pa-4 pa-md-6">
    <div v-if="loading" class="text-center py-16"><v-progress-circular indeterminate color="primary" /></div>
    <template v-else-if="deal">
      <div class="d-flex flex-wrap align-center mb-6 gap-2">
        <v-btn icon @click="$router.back()"><v-icon>mdi-arrow-left</v-icon></v-btn>
        <v-avatar size="48" color="purple" class="ml-2"><v-icon color="white">mdi-briefcase</v-icon></v-avatar>
        <div class="ml-3">
          <div class="d-flex align-center gap-2">
            <h1 class="text-h5 font-weight-bold">{{ deal.name }}</h1>
            <v-chip small :color="stageColor(deal.stage)" dark>{{ deal.stage }}</v-chip>
          </div>
          <p class="text-subtitle-1 font-weight-bold green--text text--darken-2 mb-0">
            {{ deal.amount ? '₹' + Number(deal.amount).toLocaleString() : '₹0' }}
          </p>
        </div>
        <v-spacer />
        <v-btn outlined color="primary" :to="`/deals/${deal.id}/edit`"><v-icon left small>mdi-pencil</v-icon>Edit</v-btn>
        <v-btn color="error" outlined @click="deleteDialog = true"><v-icon left small>mdi-delete</v-icon>Delete</v-btn>
      </div>

      <v-row>
        <v-col cols="12" md="8">
          <v-card outlined elevation="0" class="mb-4">
            <v-card-title class="text-body-1 font-weight-bold">Deal Information</v-card-title>
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
              <div v-if="deal.account_id" class="mb-3">
                <div class="text-caption grey--text mb-1">Account</div>
                <v-chip small color="teal" outlined :to="`/accounts/${deal.account_id}`">
                  <v-icon left x-small>mdi-office-building</v-icon>{{ deal.account_name }}
                </v-chip>
              </div>
              <div v-if="deal.contact_id" class="mb-3">
                <div class="text-caption grey--text mb-1">Contact</div>
                <v-chip small color="cyan" outlined :to="`/contacts/${deal.contact_id}`">
                  <v-icon left x-small>mdi-contacts</v-icon>{{ deal.contact_name }}
                </v-chip>
              </div>
              <div class="field-row mb-1"><span class="text-caption grey--text">Owner</span><span class="text-body-2">{{ deal.owner_name || '—' }}</span></div>
              <v-divider class="my-2" />
              <div class="field-row mb-1"><span class="text-caption grey--text">Created</span><span class="text-body-2">{{ formatDate(deal.created_at) }}</span></div>
              <div class="field-row"><span class="text-caption grey--text">Updated</span><span class="text-body-2">{{ formatDate(deal.updated_at) }}</span></div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-dialog v-model="deleteDialog" max-width="400">
        <v-card>
          <v-card-title class="error--text"><v-icon left color="error">mdi-alert</v-icon>Delete Deal?</v-card-title>
          <v-card-text>Delete deal "{{ deal.name }}"?</v-card-text>
          <v-card-actions><v-spacer /><v-btn text @click="deleteDialog = false">Cancel</v-btn><v-btn color="error" @click="deleteDeal" :loading="deleting">Delete</v-btn></v-card-actions>
        </v-card>
      </v-dialog>
    </template>
  </div>
</template>

<script>
import dealService from '../../services/dealService'

export default {
  name: 'DealDetail',
  data() { return { deal: null, loading: true, deleteDialog: false, deleting: false } },
  computed: {
    fields() {
      if (!this.deal) return []
      return [
        { label: 'Stage', value: this.deal.stage },
        { label: 'Probability', value: this.deal.probability ? `${this.deal.probability}%` : null },
        { label: 'Expected Close', value: this.formatDate(this.deal.expected_close_date) },
        { label: 'Lead Source', value: this.deal.lead_source },
        { label: 'Description', value: this.deal.description }
      ]
    }
  },
  async mounted() {
    this.loading = true
    try {
      const res = await dealService.get(this.$route.params.id)
      this.deal = res.data.data
    } finally {
      this.loading = false
    }
  },
  methods: {
    async deleteDeal() {
      this.deleting = true
      try {
        await dealService.delete(this.deal.id)
        this.$store.dispatch('snackbar/success', 'Deal deleted')
        this.$router.push('/deals')
      } finally {
        this.deleting = false
      }
    },
    formatDate(d) { return d ? new Date(d).toLocaleDateString('en-IN') : '—' },
    stageColor(s) { const m = { Qualification: 'blue', 'Needs Analysis': 'cyan', Proposal: 'orange', Negotiation: 'purple', 'Closed Won': 'green', 'Closed Lost': 'red' }; return m[s] || 'grey' }
  }
}
</script>
<style scoped>
.gap-2 { gap: 8px; }
.field-row { display: flex; justify-content: space-between; }
.border-b { border-bottom: 1px solid #f0f0f0; }
</style>
