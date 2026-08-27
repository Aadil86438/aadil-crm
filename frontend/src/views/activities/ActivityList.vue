<template>
  <div class="pa-4 pa-md-6">
    <div class="d-flex flex-wrap align-center mb-4 gap-2">
      <div>
        <h1 class="text-h5 font-weight-bold">Activities</h1>
        <p class="text-body-2 grey--text mb-0">{{ total }} logged activities</p>
      </div>
      <v-spacer />
      <v-btn color="primary" @click="dialog = true" id="log-activity-btn">
        <v-icon left>mdi-plus</v-icon>Log Activity
      </v-btn>
    </div>

    <!-- Filters -->
    <v-card class="mb-4 pa-3" elevation="0" outlined>
      <v-row dense>
        <v-col cols="6" sm="3" md="2">
          <v-select v-model="typeFilter" :items="activityTypes" label="Activity Type" dense outlined hide-details clearable @change="load" />
        </v-col>
      </v-row>
    </v-card>

    <!-- Timeline list -->
    <v-card elevation="0" outlined class="pa-4">
      <div v-if="loading" class="text-center py-8"><v-progress-circular indeterminate color="primary" /></div>
      <v-timeline dense align-top v-else-if="activities.length">
        <v-timeline-item v-for="act in activities" :key="act.id" :color="typeColor(act.type)" small>
          <template v-slot:icon><v-icon x-small color="white">{{ typeIcon(act.type) }}</v-icon></template>
          <v-card class="pa-3 elevation-1">
            <div class="d-flex align-start justify-space-between">
              <div>
                <v-chip x-small :color="typeColor(act.type)" dark class="mr-2">{{ act.type.toUpperCase() }}</v-chip>
                <span class="font-weight-bold text-body-2">{{ act.subject }}</span>
              </div>
              <span class="text-caption grey--text">{{ formatDate(act.created_at) }}</span>
            </div>
            <p class="text-body-2 grey--text text--darken-2 mt-2 mb-0">{{ act.description || 'No description provided' }}</p>
            <div class="mt-2 text-caption grey--text" v-if="act.owner_name">
              Logged by: <strong>{{ act.owner_name }}</strong>
            </div>
          </v-card>
        </v-timeline-item>
      </v-timeline>
      <div v-else class="text-center py-12 grey--text">No activities logged yet</div>
    </v-card>

    <!-- Log Activity Dialog -->
    <v-dialog v-model="dialog" max-width="500">
      <v-card>
        <v-card-title>Log Activity</v-card-title>
        <v-card-text>
          <v-form ref="form">
            <v-select v-model="form.type" :items="activityTypes" label="Type *" outlined dense class="mb-3" />
            <v-text-field v-model="form.subject" label="Subject *" outlined dense class="mb-3" id="activity-subject" />
            <v-textarea v-model="form.description" label="Description" outlined dense rows="3" />
          </v-form>
        </v-card-text>
        <v-card-actions>
          <v-spacer /><v-btn text @click="dialog = false">Cancel</v-btn>
          <v-btn color="primary" @click="save" :loading="saving" id="activity-save-btn">Log</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import activityService from '../../services/activityService'

export default {
  name: 'ActivityList',
  data() {
    return {
      activities: [], total: 0, loading: false, dialog: false, saving: false, typeFilter: null,
      activityTypes: ['call', 'meeting', 'email', 'note', 'task'],
      form: { type: 'call', subject: '', description: '' }
    }
  },
  mounted() { this.load() },
  methods: {
    async load() {
      this.loading = true
      try {
        const params = { page: 1, page_size: 50 }
        if (this.typeFilter) params.type = this.typeFilter
        const res = await activityService.list(params)
        this.activities = res.data.data || []
        this.total = res.data.total || 0
      } finally { this.loading = false }
    },
    async save() {
      if (!this.form.subject) return
      this.saving = true
      try {
        await activityService.create(this.form)
        this.$store.dispatch('snackbar/success', 'Activity logged')
        this.dialog = false
        this.form = { type: 'call', subject: '', description: '' }
        this.load()
      } finally { this.saving = false }
    },
    formatDate(d) { return d ? new Date(d).toLocaleString('en-IN') : '—' },
    typeColor(t) { const m = { call: 'blue', meeting: 'purple', email: 'teal', note: 'orange', task: 'green' }; return m[t] || 'grey' },
    typeIcon(t) { const m = { call: 'mdi-phone', meeting: 'mdi-account-group', email: 'mdi-email', note: 'mdi-note', task: 'mdi-checkbox-marked-circle' }; return m[t] || 'mdi-circle' }
  }
}
</script>
<style scoped>.gap-2 { gap: 8px; }</style>
