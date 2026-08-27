<template>
  <div class="pa-4 pa-md-6">
    <div v-if="loading" class="text-center py-16"><v-progress-circular indeterminate color="primary" /></div>
    <template v-else-if="contact">
      <div class="d-flex flex-wrap align-center mb-6 gap-2">
        <v-btn icon @click="$router.back()"><v-icon>mdi-arrow-left</v-icon></v-btn>
        <v-avatar size="48" color="cyan" class="ml-2"><span class="white--text text-h6">{{ (contact.first_name || '')[0] }}{{ (contact.last_name || '')[0] }}</span></v-avatar>
        <div class="ml-3">
          <h1 class="text-h5 font-weight-bold">{{ contact.first_name }} {{ contact.last_name }}</h1>
          <p class="text-body-2 grey--text mb-0">{{ contact.job_title }} <span v-if="contact.account_name">• {{ contact.account_name }}</span></p>
        </div>
        <v-spacer />
        <v-btn outlined color="primary" :to="`/contacts/${contact.id}/edit`"><v-icon left small>mdi-pencil</v-icon>Edit</v-btn>
        <v-btn color="error" outlined @click="deleteDialog = true"><v-icon left small>mdi-delete</v-icon>Delete</v-btn>
      </div>

      <v-row>
        <v-col cols="12" md="8">
          <v-card outlined elevation="0" class="mb-4">
            <v-card-title class="text-body-1 font-weight-bold"><v-icon left small>mdi-information</v-icon>Contact Details</v-card-title>
            <v-divider />
            <v-card-text>
              <v-row dense>
                <v-col cols="12" sm="6" v-for="f in fields" :key="f.label">
                  <div class="d-flex justify-space-between py-1 border-bottom">
                    <span class="text-caption grey--text">{{ f.label }}</span>
                    <span class="text-body-2 font-weight-medium">{{ f.value || '—' }}</span>
                  </div>
                </v-col>
              </v-row>
            </v-card-text>
          </v-card>

          <!-- Activity Timeline -->
          <v-card outlined elevation="0">
            <v-card-title class="text-body-1 font-weight-bold d-flex align-center">
              <v-icon left small>mdi-timeline</v-icon>Activities
              <v-spacer />
              <v-btn small text color="primary" @click="activityDialog = true">+ Log</v-btn>
            </v-card-title>
            <v-divider />
            <v-card-text>
              <v-timeline dense v-if="activities.length">
                <v-timeline-item v-for="a in activities" :key="a.id" :color="typeColor(a.type)" small>
                  <template v-slot:icon><v-icon x-small color="white">{{ typeIcon(a.type) }}</v-icon></template>
                  <div>
                    <strong class="text-body-2">{{ a.subject }}</strong>
                    <p class="text-caption grey--text mb-0">{{ formatDate(a.created_at) }}</p>
                  </div>
                </v-timeline-item>
              </v-timeline>
              <p v-else class="text-caption grey--text text-center py-4">No activities yet</p>
            </v-card-text>
          </v-card>
        </v-col>

        <v-col cols="12" md="4">
          <v-card outlined elevation="0">
            <v-card-text>
              <div v-if="contact.account_id" class="mb-3">
                <div class="text-caption grey--text mb-1">Account</div>
                <v-chip small color="teal" outlined :to="`/accounts/${contact.account_id}`">
                  <v-icon left x-small>mdi-office-building</v-icon>{{ contact.account_name }}
                </v-chip>
              </div>
              <div class="field-row mb-1"><span class="text-caption grey--text">Owner</span><span class="text-body-2">{{ contact.owner_name || '—' }}</span></div>
              <v-divider class="my-2" />
              <div class="field-row mb-1"><span class="text-caption grey--text">Created</span><span class="text-body-2">{{ formatDate(contact.created_at) }}</span></div>
              <div class="field-row"><span class="text-caption grey--text">Updated</span><span class="text-body-2">{{ formatDate(contact.updated_at) }}</span></div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <!-- Activity Dialog -->
      <v-dialog v-model="activityDialog" max-width="500">
        <v-card>
          <v-card-title>Log Activity</v-card-title>
          <v-card-text>
            <v-select v-model="newActivity.type" :items="['call','meeting','email','note','task']" label="Type" outlined dense class="mb-3" />
            <v-text-field v-model="newActivity.subject" label="Subject *" outlined dense class="mb-3" />
            <v-textarea v-model="newActivity.description" label="Description" outlined dense rows="3" />
          </v-card-text>
          <v-card-actions>
            <v-spacer /><v-btn text @click="activityDialog = false">Cancel</v-btn>
            <v-btn color="primary" @click="saveActivity" :loading="savingActivity">Log</v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>

      <!-- Delete Dialog -->
      <v-dialog v-model="deleteDialog" max-width="400">
        <v-card>
          <v-card-title class="error--text"><v-icon left color="error">mdi-alert</v-icon>Delete Contact?</v-card-title>
          <v-card-text>Delete {{ contact.first_name }} {{ contact.last_name }}? This cannot be undone.</v-card-text>
          <v-card-actions><v-spacer /><v-btn text @click="deleteDialog = false">Cancel</v-btn><v-btn color="error" @click="deleteContact" :loading="deleting">Delete</v-btn></v-card-actions>
        </v-card>
      </v-dialog>
    </template>
  </div>
</template>

<script>
import contactService from '../../services/contactService'
import activityService from '../../services/activityService'

export default {
  name: 'ContactDetail',
  data() {
    return {
      contact: null, activities: [], loading: true,
      deleteDialog: false, deleting: false,
      activityDialog: false, savingActivity: false,
      newActivity: { type: 'call', subject: '', description: '' }
    }
  },
  computed: {
    fields() {
      if (!this.contact) return []
      return [
        { label: 'Email', value: this.contact.email }, { label: 'Phone', value: this.contact.phone },
        { label: 'Mobile', value: this.contact.mobile }, { label: 'Department', value: this.contact.department },
        { label: 'City', value: this.contact.city }, { label: 'Country', value: this.contact.country },
        { label: 'Description', value: this.contact.description }
      ]
    }
  },
  async mounted() {
    this.loading = true
    try {
      const [cRes, aRes] = await Promise.all([contactService.get(this.$route.params.id), activityService.list({ related_contact_id: this.$route.params.id })])
      this.contact = cRes.data.data
      this.activities = aRes.data.data || []
    } finally { this.loading = false }
  },
  methods: {
    async deleteContact() {
      this.deleting = true
      try { await contactService.delete(this.contact.id); this.$store.dispatch('snackbar/success', 'Contact deleted'); this.$router.push('/contacts') } finally { this.deleting = false }
    },
    async saveActivity() {
      if (!this.newActivity.subject) return
      this.savingActivity = true
      try {
        await activityService.create({ ...this.newActivity, related_contact_id: this.contact.id })
        this.activityDialog = false; this.newActivity = { type: 'call', subject: '', description: '' }
        const res = await activityService.list({ related_contact_id: this.contact.id })
        this.activities = res.data.data || []
        this.$store.dispatch('snackbar/success', 'Activity logged')
      } finally { this.savingActivity = false }
    },
    formatDate(d) { return d ? new Date(d).toLocaleDateString('en-IN') : '—' },
    typeColor(t) { const m = { call: 'blue', meeting: 'purple', email: 'teal', note: 'orange', task: 'green' }; return m[t] || 'grey' },
    typeIcon(t) { const m = { call: 'mdi-phone', meeting: 'mdi-account-group', email: 'mdi-email', note: 'mdi-note', task: 'mdi-checkbox-marked-circle' }; return m[t] || 'mdi-circle' }
  }
}
</script>
<style scoped>
.gap-2 { gap: 8px; }
.field-row { display: flex; justify-content: space-between; }
.border-bottom { border-bottom: 1px solid #f0f0f0; }
</style>
