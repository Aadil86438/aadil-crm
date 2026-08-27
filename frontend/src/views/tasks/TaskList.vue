<template>
  <div class="pa-4 pa-md-6">
    <div class="d-flex flex-wrap align-center mb-4 gap-2">
      <div>
        <h1 class="text-h5 font-weight-bold">Tasks</h1>
        <p class="text-body-2 grey--text mb-0">{{ total }} total tasks</p>
      </div>
      <v-spacer />
      <v-btn color="primary" @click="openDialog()" id="create-task-btn">
        <v-icon left>mdi-plus</v-icon>New Task
      </v-btn>
    </div>

    <!-- Filters -->
    <v-card class="mb-4 pa-3" elevation="0" outlined>
      <v-row dense>
        <v-col cols="12" sm="5" md="4">
          <v-text-field v-model="search" placeholder="Search tasks..." prepend-inner-icon="mdi-magnify" dense outlined hide-details clearable @input="onSearchInput" id="task-search" />
        </v-col>
        <v-col cols="6" sm="3" md="2">
          <v-select v-model="statusFilter" :items="statuses" label="Status" dense outlined hide-details clearable @change="load" />
        </v-col>
        <v-col cols="6" sm="3" md="2">
          <v-select v-model="priorityFilter" :items="priorities" label="Priority" dense outlined hide-details clearable @change="load" />
        </v-col>
      </v-row>
    </v-card>

    <v-card elevation="0" outlined>
      <v-data-table :headers="headers" :items="tasks" :loading="loading" :server-items-length="total" :options.sync="options" :footer-props="{ itemsPerPageOptions: [10, 20, 50] }" @update:options="load" id="tasks-table">
        <template v-slot:item.subject="{ item }">
          <span class="font-weight-medium text-body-2">{{ item.subject }}</span>
        </template>
        <template v-slot:item.status="{ item }">
          <v-chip x-small :color="statusColor(item.status)" dark>{{ item.status }}</v-chip>
        </template>
        <template v-slot:item.priority="{ item }">
          <v-chip x-small outlined :color="priorityColor(item.priority)">{{ item.priority }}</v-chip>
        </template>
        <template v-slot:item.due_date="{ item }">
          <span class="text-caption" :class="isOverdue(item.due_date, item.status) ? 'error--text font-weight-bold' : ''">{{ formatDate(item.due_date) }}</span>
        </template>
        <template v-slot:item.actions="{ item }">
          <v-btn icon x-small @click="openDialog(item)"><v-icon x-small>mdi-pencil</v-icon></v-btn>
          <v-btn icon x-small color="error" @click="confirmDelete(item)"><v-icon x-small>mdi-delete</v-icon></v-btn>
        </template>
        <template v-slot:no-data>
          <div class="text-center py-12">
            <v-icon size="48" color="grey lighten-2">mdi-checkbox-marked-circle</v-icon>
            <p class="grey--text mt-3">No tasks found</p>
            <v-btn color="primary" @click="openDialog()">Create Task</v-btn>
          </div>
        </template>
      </v-data-table>
    </v-card>

    <!-- Task Form Dialog -->
    <v-dialog v-model="dialog" max-width="600">
      <v-card>
        <v-card-title>{{ editItem ? 'Edit Task' : 'New Task' }}</v-card-title>
        <v-card-text>
          <v-form ref="form" v-model="valid">
            <v-text-field v-model="form.subject" label="Subject *" outlined dense :rules="[v => !!v || 'Required']" id="task-subject" class="mb-3" />
            <v-row dense>
              <v-col cols="6"><v-select v-model="form.status" :items="statuses" label="Status" outlined dense id="task-status" /></v-col>
              <v-col cols="6"><v-select v-model="form.priority" :items="priorities" label="Priority" outlined dense id="task-priority" /></v-col>
            </v-row>
            <v-menu v-model="dateMenu" :close-on-content-click="false" transition="scale-transition" offset-y min-width="auto" class="mb-3">
              <template v-slot:activator="{ on, attrs }">
                <v-text-field v-model="form.due_date" label="Due Date" prepend-icon="mdi-calendar" readonly v-bind="attrs" v-on="on" outlined dense />
              </template>
              <v-date-picker v-model="form.due_date" @input="dateMenu = false" />
            </v-menu>
            <v-textarea v-model="form.description" label="Description" outlined dense rows="3" />
          </v-form>
        </v-card-text>
        <v-card-actions>
          <v-spacer /><v-btn text @click="dialog = false">Cancel</v-btn>
          <v-btn color="primary" @click="save" :loading="saving" id="task-save-btn">Save</v-btn>
        </v-card-actions>
      </v-card>
    </dialog>

    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title class="error--text"><v-icon left color="error">mdi-alert</v-icon>Delete Task?</v-card-title>
        <v-card-text>Delete "{{ deleteTarget && deleteTarget.subject }}"?</v-card-text>
        <v-card-actions><v-spacer /><v-btn text @click="deleteDialog = false">Cancel</v-btn><v-btn color="error" @click="deleteTask" :loading="deleting">Delete</v-btn></v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import taskService from '../../services/taskService'

export default {
  name: 'TaskList',
  data() {
    return {
      tasks: [], total: 0, loading: false, search: '', searchTimer: null,
      statusFilter: null, priorityFilter: null,
      options: { page: 1, itemsPerPage: 20, sortBy: ['due_date'], sortDesc: [false] },
      dialog: false, valid: false, saving: false, editItem: null, dateMenu: false,
      deleteDialog: false, deleteTarget: null, deleting: false,
      statuses: ['Not Started', 'In Progress', 'Completed', 'Deferred'],
      priorities: ['High', 'Medium', 'Low'],
      form: { subject: '', status: 'Not Started', priority: 'Medium', due_date: null, description: '' },
      headers: [
        { text: 'Subject', value: 'subject' },
        { text: 'Status', value: 'status' },
        { text: 'Priority', value: 'priority' },
        { text: 'Due Date', value: 'due_date' },
        { text: 'Owner', value: 'owner_name', sortable: false },
        { text: 'Actions', value: 'actions', sortable: false, align: 'center' }
      ]
    }
  },
  methods: {
    async load() {
      this.loading = true
      try {
        const { page, itemsPerPage, sortDesc } = this.options
        const params = { page, page_size: itemsPerPage, order: sortDesc[0] ? 'desc' : 'asc' }
        if (this.search) params.search = this.search
        if (this.statusFilter) params.status = this.statusFilter
        if (this.priorityFilter) params.priority = this.priorityFilter
        const res = await taskService.list(params)
        this.tasks = res.data.data || []
        this.total = res.data.total || 0
      } finally { this.loading = false }
    },
    onSearchInput() { clearTimeout(this.searchTimer); this.searchTimer = setTimeout(() => { this.options.page = 1; this.load() }, 400) },
    openDialog(item = null) {
      this.editItem = item
      if (item) {
        this.form = {
          subject: item.subject, status: item.status, priority: item.priority,
          due_date: item.due_date ? item.due_date.split('T')[0] : null, description: item.description || ''
        }
      } else {
        this.form = { subject: '', status: 'Not Started', priority: 'Medium', due_date: null, description: '' }
      }
      this.dialog = true
    },
    async save() {
      if (!this.$refs.form.validate()) return
      this.saving = true
      try {
        if (this.editItem) {
          await taskService.update(this.editItem.id, this.form)
          this.$store.dispatch('snackbar/success', 'Task updated')
        } else {
          await taskService.create(this.form)
          this.$store.dispatch('snackbar/success', 'Task created')
        }
        this.dialog = false
        this.load()
      } finally { this.saving = false }
    },
    confirmDelete(t) { this.deleteTarget = t; this.deleteDialog = true },
    async deleteTask() {
      this.deleting = true
      try { await taskService.delete(this.deleteTarget.id); this.$store.dispatch('snackbar/success', 'Task deleted'); this.deleteDialog = false; this.load() } finally { this.deleting = false }
    },
    statusColor(s) { const m = { 'Not Started': 'blue-grey', 'In Progress': 'blue', Completed: 'green', Deferred: 'orange' }; return m[s] || 'grey' },
    priorityColor(p) { const m = { High: 'red', Medium: 'orange', Low: 'blue' }; return m[p] || 'grey' },
    isOverdue(d, s) { return d && s !== 'Completed' && new Date(d) < new Date() },
    formatDate(d) { return d ? new Date(d).toLocaleDateString('en-IN', { day: 'numeric', month: 'short', year: 'numeric' }) : '—' }
  }
}
</script>
<style scoped>.gap-2 { gap: 8px; }</style>
