<template>
  <div class="pa-4 pa-md-6">
    <div class="d-flex align-center mb-6">
      <div>
        <h1 class="text-h5 font-weight-bold">Calendar</h1>
        <p class="text-body-2 grey--text mb-0">Scheduled activities & follow-ups</p>
      </div>
    </div>

    <v-card elevation="0" outlined class="pa-4">
      <div v-if="loading" class="text-center py-16"><v-progress-circular indeterminate color="primary" /></div>
      <v-sheet height="600" v-else>
        <v-calendar
          ref="calendar"
          v-model="focus"
          :events="events"
          color="primary"
          type="month"
        />
      </v-sheet>
    </v-card>
  </div>
</template>

<script>
import activityService from '../../services/activityService'

export default {
  name: 'CalendarView',
  data() {
    return {
      focus: '',
      loading: true,
      activities: []
    }
  },
  computed: {
    events() {
      return this.activities.map(a => ({
        name: a.subject,
        start: a.due_date ? a.due_date.split('T')[0] : a.created_at.split('T')[0],
        color: this.typeColor(a.type)
      }))
    }
  },
  async mounted() {
    try {
      const res = await activityService.getCalendar()
      this.activities = res.data.data || []
    } finally {
      this.loading = false
    }
  },
  methods: {
    typeColor(t) { const m = { call: 'blue', meeting: 'purple', email: 'teal', note: 'orange', task: 'green' }; return m[t] || 'grey' }
  }
}
</script>
