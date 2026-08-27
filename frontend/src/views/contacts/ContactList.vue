<template>
  <div class="pa-4 pa-md-6">
    <div class="d-flex flex-wrap align-center mb-4 gap-2">
      <div>
        <h1 class="text-h5 font-weight-bold">Contacts</h1>
        <p class="text-body-2 grey--text mb-0">{{ total }} total contacts</p>
      </div>
      <v-spacer />
      <v-btn color="primary" to="/contacts/new" id="create-contact-btn">
        <v-icon left>mdi-plus</v-icon><span class="d-none d-sm-inline">New Contact</span>
      </v-btn>
    </div>

    <!-- Filters -->
    <v-card class="mb-4 pa-3" elevation="0" outlined>
      <v-row dense>
        <v-col cols="12" sm="6" md="5">
          <v-text-field v-model="search" placeholder="Search by name, email..." prepend-inner-icon="mdi-magnify" dense outlined hide-details clearable @input="onSearchInput" id="contact-search" />
        </v-col>
      </v-row>
    </v-card>

    <v-card elevation="0" outlined>
      <v-data-table
        :headers="headers"
        :items="contacts"
        :loading="loading"
        :server-items-length="total"
        :options.sync="tableOptions"
        :footer-props="{ itemsPerPageOptions: [10, 20, 50] }"
        @update:options="loadContacts"
        id="contacts-table"
      >
        <template v-slot:item.name="{ item }">
          <div class="d-flex align-center py-1">
            <v-avatar size="32" color="cyan" class="mr-3">
              <span class="white--text text-caption">{{ (item.first_name || '')[0] }}{{ (item.last_name || '')[0] }}</span>
            </v-avatar>
            <div>
              <router-link :to="`/contacts/${item.id}`" class="text-body-2 font-weight-medium text-decoration-none primary--text">{{ item.first_name }} {{ item.last_name }}</router-link>
              <div class="text-caption grey--text">{{ item.job_title }}</div>
            </div>
          </div>
        </template>
        <template v-slot:item.account_name="{ item }">
          <router-link v-if="item.account_id" :to="`/accounts/${item.account_id}`" class="text-decoration-none primary--text text-body-2">{{ item.account_name }}</router-link>
          <span v-else class="grey--text">—</span>
        </template>
        <template v-slot:item.created_at="{ item }">
          <span class="text-caption">{{ formatDate(item.created_at) }}</span>
        </template>
        <template v-slot:item.actions="{ item }">
          <v-btn icon x-small :to="`/contacts/${item.id}`"><v-icon x-small>mdi-eye</v-icon></v-btn>
          <v-btn icon x-small :to="`/contacts/${item.id}/edit`"><v-icon x-small>mdi-pencil</v-icon></v-btn>
          <v-btn icon x-small color="error" @click="confirmDelete(item)"><v-icon x-small>mdi-delete</v-icon></v-btn>
        </template>
        <template v-slot:no-data>
          <div class="text-center py-12">
            <v-icon size="48" color="grey lighten-2">mdi-contacts</v-icon>
            <p class="grey--text mt-3">No contacts found</p>
            <v-btn color="primary" to="/contacts/new">Create Contact</v-btn>
          </div>
        </template>
      </v-data-table>
    </v-card>

    <v-dialog v-model="deleteDialog" max-width="420">
      <v-card>
        <v-card-title class="error--text"><v-icon left color="error">mdi-alert</v-icon>Delete Contact?</v-card-title>
        <v-card-text>Delete {{ deleteTarget && deleteTarget.first_name }} {{ deleteTarget && deleteTarget.last_name }}?</v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn text @click="deleteDialog = false">Cancel</v-btn>
          <v-btn color="error" @click="deleteContact" :loading="deleting">Delete</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import contactService from '../../services/contactService'
export default {
  name: 'ContactList',
  data() {
    return {
      contacts: [], total: 0, loading: false, search: '', searchTimer: null,
      tableOptions: { page: 1, itemsPerPage: 20, sortBy: ['created_at'], sortDesc: [true] },
      deleteDialog: false, deleteTarget: null, deleting: false,
      headers: [
        { text: 'Name', value: 'name', sortable: false },
        { text: 'Email', value: 'email' },
        { text: 'Phone', value: 'phone', sortable: false },
        { text: 'Account', value: 'account_name', sortable: false },
        { text: 'Owner', value: 'owner_name', sortable: false },
        { text: 'Created', value: 'created_at' },
        { text: 'Actions', value: 'actions', sortable: false, align: 'center' },
      ]
    }
  },
  methods: {
    async loadContacts() {
      this.loading = true
      try {
        const { page, itemsPerPage, sortBy, sortDesc } = this.tableOptions
        const params = { page, page_size: itemsPerPage, sort: sortBy[0] || 'created_at', order: sortDesc[0] ? 'desc' : 'asc' }
        if (this.search) params.search = this.search
        const res = await contactService.list(params)
        this.contacts = res.data.data || []
        this.total = res.data.total || 0
      } catch (err) {
        this.$store.dispatch('snackbar/error', 'Failed to load contacts')
      } finally {
        this.loading = false
      }
    },
    onSearchInput() {
      clearTimeout(this.searchTimer)
      this.searchTimer = setTimeout(() => { this.tableOptions.page = 1; this.loadContacts() }, 400)
    },
    confirmDelete(contact) { this.deleteTarget = contact; this.deleteDialog = true },
    async deleteContact() {
      this.deleting = true
      try {
        await contactService.delete(this.deleteTarget.id)
        this.$store.dispatch('snackbar/success', 'Contact deleted')
        this.deleteDialog = false; this.loadContacts()
      } catch (err) { this.$store.dispatch('snackbar/error', 'Failed to delete') } finally { this.deleting = false }
    },
    formatDate(d) { return d ? new Date(d).toLocaleDateString('en-IN', { day: 'numeric', month: 'short', year: 'numeric' }) : '—' }
  }
}
</script>
