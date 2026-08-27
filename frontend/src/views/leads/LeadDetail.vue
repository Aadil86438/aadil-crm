<template>
  <div class="pa-4 pa-md-6">
    <!-- Loading -->
    <div v-if="loading" class="text-center py-16">
      <v-progress-circular indeterminate color="primary" />
    </div>

    <template v-else-if="lead">
      <!-- Header -->
      <div class="d-flex flex-wrap align-center mb-4 gap-2">
        <v-btn icon @click="$router.back()"><v-icon>mdi-arrow-left</v-icon></v-btn>
        <div class="ml-2">
          <div class="d-flex align-center gap-2">
            <h1 class="text-h5 font-weight-bold">{{ lead.first_name }} {{ lead.last_name }}</h1>
            <v-chip small :color="statusColors[lead.lead_status]" dark>{{ lead.lead_status }}</v-chip>
            <v-chip small color="success" dark v-if="lead.is_converted"><v-icon left x-small>mdi-check</v-icon>Converted</v-chip>
          </div>
          <p class="text-body-2 grey--text mb-0">{{ lead.company }} • {{ lead.job_title }}</p>
        </div>
        <v-spacer />
        <v-btn outlined color="primary" :to="`/leads/${lead.id}/edit`">
          <v-icon left small>mdi-pencil</v-icon>Edit
        </v-btn>
        <v-btn color="success" :to="`/leads/${lead.id}/convert`" v-if="!lead.is_converted">
          <v-icon left small>mdi-account-convert</v-icon>Convert
        </v-btn>
        <v-btn color="error" outlined @click="deleteDialog = true">
          <v-icon left small>mdi-delete</v-icon>Delete
        </v-btn>
      </div>

      <v-row>
        <!-- Main Details -->
        <v-col cols="12" md="8">
          <!-- Contact Info -->
          <v-card class="mb-4" outlined elevation="0">
            <v-card-title class="text-body-1 font-weight-bold">
              <v-icon left small>mdi-information</v-icon>Lead Information
            </v-card-title>
            <v-divider />
            <v-card-text>
              <v-row dense>
                <v-col cols="12" sm="6" v-for="field in leadFields" :key="field.label">
                  <div class="field-row">
                    <div class="field-label">{{ field.label }}</div>
                    <div class="field-value">{{ field.value || '—' }}</div>
                  </div>
                </v-col>
              </v-row>
            </v-card-text>
          </v-card>

          <!-- Notes Section -->
          <v-card class="mb-4" outlined elevation="0">
            <v-card-title class="text-body-1 font-weight-bold d-flex align-center">
              <v-icon left small>mdi-note-text</v-icon>Notes
              <v-spacer />
              <v-btn small text color="primary" @click="noteDialog = true">+ Add Note</v-btn>
            </v-card-title>
            <v-divider />
            <v-card-text>
              <div v-for="note in notes" :key="note.id" class="note-item mb-3 pa-3">
                <div class="d-flex align-center mb-1">
                  <strong class="text-body-2">{{ note.title || 'Note' }}</strong>
                  <v-spacer />
                  <span class="text-caption grey--text">{{ formatDate(note.created_at) }}</span>
                  <v-btn icon x-small class="ml-1" @click="deleteNote(note.id)">
                    <v-icon x-small color="error">mdi-delete</v-icon>
                  </v-btn>
                </div>
                <p class="text-body-2 mb-0 grey--text text--darken-2">{{ note.body }}</p>
              </div>
              <div v-if="notes.length === 0" class="text-center grey--text py-4 text-caption">No notes yet</div>
            </v-card-text>
          </v-card>

          <!-- Activities Timeline -->
          <v-card outlined elevation="0">
            <v-card-title class="text-body-1 font-weight-bold d-flex align-center">
              <v-icon left small>mdi-timeline</v-icon>Activity Timeline
              <v-spacer />
              <v-btn small text color="primary" @click="activityDialog = true">+ Log Activity</v-btn>
            </v-card-title>
            <v-divider />
            <v-card-text>
              <v-timeline dense align-top v-if="activities.length">
                <v-timeline-item
                  v-for="act in activities"
                  :key="act.id"
                  :color="activityColor(act.type)"
                  small
                >
                  <template v-slot:icon>
                    <v-icon x-small color="white">{{ activityIcon(act.type) }}</v-icon>
                  </template>
                  <div class="d-flex align-start">
                    <div class="flex-grow-1">
                      <strong class="text-body-2">{{ act.subject }}</strong>
                      <p class="text-caption grey--text mb-0">{{ act.description }}</p>
                    </div>
                    <span class="text-caption grey--text ml-3">{{ formatDate(act.created_at) }}</span>
                  </div>
                </v-timeline-item>
              </v-timeline>
              <div v-else class="text-center grey--text py-4 text-caption">No activities yet</div>
            </v-card-text>
          </v-card>
        </v-col>

        <!-- Sidebar Info -->
        <v-col cols="12" md="4">
          <v-card class="mb-4" outlined elevation="0">
            <v-card-text>
              <div class="field-row mb-2"><span class="field-label">Owner</span><span class="field-value">{{ lead.owner_name || '—' }}</span></div>
              <div class="field-row mb-2"><span class="field-label">Source</span><span class="field-value">{{ lead.lead_source || '—' }}</span></div>
              <div class="field-row mb-2"><span class="field-label">Rating</span><span class="field-value">{{ lead.rating || '—' }}</span></div>
              <div class="field-row mb-2"><span class="field-label">Industry</span><span class="field-value">{{ lead.industry || '—' }}</span></div>
              <v-divider class="my-3" />
              <div class="field-row mb-2"><span class="field-label">Created</span><span class="field-value">{{ formatDate(lead.created_at) }}</span></div>
              <div class="field-row"><span class="field-label">Updated</span><span class="field-value">{{ formatDate(lead.updated_at) }}</span></div>
            </v-card-text>
          </v-card>

          <!-- Quick Tasks -->
          <v-card outlined elevation="0">
            <v-card-title class="text-body-2 font-weight-bold d-flex align-center">
              Tasks
              <v-spacer />
              <v-btn x-small text color="primary" to="/tasks">View All</v-btn>
            </v-card-title>
            <v-divider />
            <v-card-text class="pa-3">
              <p class="text-caption grey--text text-center py-2">Tasks linked to this lead appear here</p>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <!-- Add Note Dialog -->
      <v-dialog v-model="noteDialog" max-width="500">
        <v-card>
          <v-card-title>Add Note</v-card-title>
          <v-card-text>
            <v-text-field v-model="newNote.title" label="Title (optional)" outlined dense class="mb-3" />
            <v-textarea v-model="newNote.body" label="Note *" outlined dense rows="4" />
          </v-card-text>
          <v-card-actions>
            <v-spacer />
            <v-btn text @click="noteDialog = false">Cancel</v-btn>
            <v-btn color="primary" @click="saveNote" :loading="savingNote">Save</v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>

      <!-- Log Activity Dialog -->
      <v-dialog v-model="activityDialog" max-width="500">
        <v-card>
          <v-card-title>Log Activity</v-card-title>
          <v-card-text>
            <v-select v-model="newActivity.type" :items="activityTypes" label="Type *" outlined dense class="mb-3" />
            <v-text-field v-model="newActivity.subject" label="Subject *" outlined dense class="mb-3" />
            <v-textarea v-model="newActivity.description" label="Description" outlined dense rows="3" />
          </v-card-text>
          <v-card-actions>
            <v-spacer />
            <v-btn text @click="activityDialog = false">Cancel</v-btn>
            <v-btn color="primary" @click="saveActivity" :loading="savingActivity">Log</v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>

      <!-- Delete Dialog -->
      <v-dialog v-model="deleteDialog" max-width="420">
        <v-card>
          <v-card-title class="error--text"><v-icon left color="error">mdi-alert</v-icon>Delete Lead?</v-card-title>
          <v-card-text>This will permanently delete {{ lead.first_name }} {{ lead.last_name }}.</v-card-text>
          <v-card-actions>
            <v-spacer />
            <v-btn text @click="deleteDialog = false">Cancel</v-btn>
            <v-btn color="error" @click="deleteLead" :loading="deleting">Delete</v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
    </template>

    <!-- Not Found -->
    <div v-else class="text-center py-16">
      <v-icon size="64" color="grey">mdi-account-question</v-icon>
      <h3 class="mt-4">Lead not found</h3>
      <v-btn class="mt-4" color="primary" to="/leads">Back to Leads</v-btn>
    </div>
  </div>
</template>

<script>
import leadService from '../../services/leadService'
import noteService from '../../services/noteService'
import activityService from '../../services/activityService'

export default {
  name: 'LeadDetail',
  data() {
    return {
      lead: null,
      notes: [],
      activities: [],
      loading: true,
      deleteDialog: false,
      deleting: false,
      noteDialog: false,
      savingNote: false,
      activityDialog: false,
      savingActivity: false,
      newNote: { title: '', body: '' },
      newActivity: { type: 'call', subject: '', description: '' },
      activityTypes: ['call', 'meeting', 'email', 'note', 'task'],
      statusColors: { New: 'blue', Contacted: 'orange', Qualified: 'green', Unqualified: 'grey', Converted: 'purple' },
    }
  },
  computed: {
    leadFields() {
      if (!this.lead) return []
      return [
        { label: 'Email', value: this.lead.email },
        { label: 'Phone', value: this.lead.phone },
        { label: 'Mobile', value: this.lead.mobile },
        { label: 'Website', value: this.lead.website },
        { label: 'Address', value: this.lead.address },
        { label: 'City', value: this.lead.city },
        { label: 'State', value: this.lead.state },
        { label: 'Country', value: this.lead.country },
        { label: 'Annual Revenue', value: this.lead.annual_revenue ? '₹' + this.lead.annual_revenue.toLocaleString() : null },
        { label: 'Employees', value: this.lead.num_employees },
        { label: 'Description', value: this.lead.description },
      ].filter(f => f.value)
    }
  },
  mounted() {
    this.loadAll()
  },
  methods: {
    async loadAll() {
      this.loading = true
      try {
        const [leadRes, notesRes, activitiesRes] = await Promise.all([
          leadService.get(this.$route.params.id),
          noteService.list({ entity_type: 'lead', entity_id: this.$route.params.id }),
          activityService.list({ related_lead_id: this.$route.params.id }),
        ])
        this.lead = leadRes.data.data
        this.notes = notesRes.data.data || []
        this.activities = activitiesRes.data.data || []
      } catch (err) {
        this.$store.dispatch('snackbar/error', 'Failed to load lead')
      } finally {
        this.loading = false
      }
    },
    async deleteLead() {
      this.deleting = true
      try {
        await leadService.delete(this.lead.id)
        this.$store.dispatch('snackbar/success', 'Lead deleted')
        this.$router.push('/leads')
      } catch (err) {
        this.$store.dispatch('snackbar/error', 'Failed to delete')
      } finally {
        this.deleting = false
      }
    },
    async saveNote() {
      if (!this.newNote.body) return
      this.savingNote = true
      try {
        await noteService.create({ ...this.newNote, related_lead_id: this.lead.id })
        this.newNote = { title: '', body: '' }
        this.noteDialog = false
        const res = await noteService.list({ entity_type: 'lead', entity_id: this.lead.id })
        this.notes = res.data.data || []
        this.$store.dispatch('snackbar/success', 'Note added')
      } catch (err) {
        this.$store.dispatch('snackbar/error', 'Failed to add note')
      } finally {
        this.savingNote = false
      }
    },
    async deleteNote(id) {
      await noteService.delete(id)
      this.notes = this.notes.filter(n => n.id !== id)
    },
    async saveActivity() {
      if (!this.newActivity.subject) return
      this.savingActivity = true
      try {
        await activityService.create({ ...this.newActivity, related_lead_id: this.lead.id })
        this.newActivity = { type: 'call', subject: '', description: '' }
        this.activityDialog = false
        const res = await activityService.list({ related_lead_id: this.lead.id })
        this.activities = res.data.data || []
        this.$store.dispatch('snackbar/success', 'Activity logged')
      } catch (err) {
        this.$store.dispatch('snackbar/error', 'Failed to log activity')
      } finally {
        this.savingActivity = false
      }
    },
    formatDate(d) {
      if (!d) return '—'
      return new Date(d).toLocaleDateString('en-IN', { day: 'numeric', month: 'short', year: 'numeric' })
    },
    activityColor(type) {
      const m = { call: 'blue', meeting: 'purple', email: 'teal', note: 'orange', task: 'green' }
      return m[type] || 'grey'
    },
    activityIcon(type) {
      const m = { call: 'mdi-phone', meeting: 'mdi-account-group', email: 'mdi-email', note: 'mdi-note-text', task: 'mdi-checkbox-marked-circle' }
      return m[type] || 'mdi-circle'
    }
  }
}
</script>

<style scoped>
.gap-2 { gap: 8px; }
.field-row { display: flex; justify-content: space-between; padding: 4px 0; }
.field-label { color: #757575; font-size: 0.75rem; min-width: 100px; }
.field-value { font-size: 0.875rem; font-weight: 500; text-align: right; }
.note-item { background: #F9FAFB; border-radius: 8px; border-left: 3px solid #1565C0; }
</style>
