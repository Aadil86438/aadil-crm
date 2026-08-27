<template>
  <div class="pa-4 pa-md-6">
    <div class="d-flex flex-wrap align-center mb-4 gap-2">
      <div>
        <h1 class="text-h5 font-weight-bold">User Management</h1>
        <p class="text-body-2 grey--text mb-0">Manage system users, roles, and password resets</p>
      </div>
      <v-spacer />
      <v-btn color="primary" @click="openDialog()" v-if="isAdmin" id="create-user-btn">
        <v-icon left>mdi-account-plus</v-icon>New User
      </v-btn>
    </div>

    <v-card elevation="0" outlined>
      <v-data-table :headers="headers" :items="users" :loading="loading" id="users-table">
        <template v-slot:item.name="{ item }">
          <div class="d-flex align-center py-1">
            <v-avatar size="32" color="primary" class="mr-3">
              <span class="white--text text-caption font-weight-bold">{{ initials(item.name) }}</span>
            </v-avatar>
            <span class="font-weight-medium text-body-2">{{ item.name }}</span>
          </div>
        </template>

        <template v-slot:item.role="{ item }">
          <v-chip x-small :color="roleColor(item.role)" dark class="text-capitalize">{{ item.role }}</v-chip>
        </template>

        <template v-slot:item.status="{ item }">
          <v-chip x-small :color="item.status === 'active' ? 'success' : 'grey'" dark>{{ item.status }}</v-chip>
        </template>

        <template v-slot:item.actions="{ item }">
          <v-btn icon x-small @click="openDialog(item)" v-if="isAdmin"><v-icon x-small>mdi-pencil</v-icon></v-btn>
          <v-btn icon x-small color="warning" @click="openResetDialog(item)" v-if="isAdmin"><v-icon x-small>mdi-key</v-icon></v-btn>
          <v-btn icon x-small color="error" @click="confirmDelete(item)" v-if="isAdmin && item.id !== currentUser.id"><v-icon x-small>mdi-delete</v-icon></v-btn>
        </template>
      </v-data-table>
    </v-card>

    <!-- Create/Edit User Dialog -->
    <v-dialog v-model="dialog" max-width="500">
      <v-card>
        <v-card-title>{{ editItem ? 'Edit User' : 'New User' }}</v-card-title>
        <v-card-text>
          <v-form ref="form" v-model="valid">
            <v-text-field v-model="form.name" label="Full Name *" outlined dense :rules="[v => !!v || 'Required']" id="user-name" class="mb-3" />
            <v-text-field v-model="form.email" label="Email Address *" type="email" outlined dense :rules="emailRules" id="user-email" class="mb-3" :disabled="!!editItem" />
            <v-text-field v-if="!editItem" v-model="form.password" label="Password *" type="password" outlined dense :rules="[v => (v && v.length >= 8) || 'Min 8 chars']" id="user-password" class="mb-3" />
            <v-select v-model="form.role" :items="roles" label="Role *" outlined dense :rules="[v => !!v || 'Required']" id="user-role" class="mb-3" />
            <v-select v-if="editItem" v-model="form.status" :items="['active', 'disabled']" label="Status" outlined dense />
          </v-form>
        </v-card-text>
        <v-card-actions>
          <v-spacer /><v-btn text @click="dialog = false">Cancel</v-btn>
          <v-btn color="primary" @click="save" :loading="saving" id="user-save-btn">Save</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Reset Password Dialog -->
    <v-dialog v-model="resetDialog" max-width="400">
      <v-card>
        <v-card-title>Reset Password</v-card-title>
        <v-card-text>
          <v-text-field v-model="newPassword" label="New Password (min 8 chars)" type="password" outlined dense />
        </v-card-text>
        <v-card-actions>
          <v-spacer /><v-btn text @click="resetDialog = false">Cancel</v-btn>
          <v-btn color="warning" @click="resetPassword" :loading="saving">Update Password</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Delete Dialog -->
    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title class="error--text"><v-icon left color="error">mdi-alert</v-icon>Delete User?</v-card-title>
        <v-card-text>Delete user "{{ deleteTarget && deleteTarget.name }}"?</v-card-text>
        <v-card-actions><v-spacer /><v-btn text @click="deleteDialog = false">Cancel</v-btn><v-btn color="error" @click="deleteUser" :loading="deleting">Delete</v-btn></v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import userService from '../../services/userService'

export default {
  name: 'UserList',
  data() {
    return {
      users: [], loading: false, dialog: false, valid: false, saving: false, editItem: null,
      resetDialog: false, newPassword: '', resetTarget: null,
      deleteDialog: false, deleteTarget: null, deleting: false,
      roles: ['admin', 'manager', 'sales_user'],
      form: { name: '', email: '', password: '', role: 'sales_user', status: 'active' },
      emailRules: [v => !!v || 'Required', v => /.+@.+\..+/.test(v) || 'Invalid email'],
      headers: [
        { text: 'Name', value: 'name' },
        { text: 'Email', value: 'email' },
        { text: 'Role', value: 'role' },
        { text: 'Status', value: 'status' },
        { text: 'Actions', value: 'actions', sortable: false, align: 'center' }
      ]
    }
  },
  computed: {
    ...mapGetters('auth', ['user', 'isAdmin']),
    currentUser() { return this.user || {} }
  },
  mounted() { this.load() },
  methods: {
    async load() {
      this.loading = true
      try {
        const res = await userService.list()
        this.users = res.data.data || []
      } finally { this.loading = false }
    },
    initials(name) { return name ? name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2) : 'U' },
    roleColor(r) { const m = { admin: 'purple', manager: 'indigo', sales_user: 'blue' }; return m[r] || 'grey' },
    openDialog(item = null) {
      this.editItem = item
      if (item) {
        this.form = { name: item.name, email: item.email, role: item.role, status: item.status }
      } else {
        this.form = { name: '', email: '', password: '', role: 'sales_user', status: 'active' }
      }
      this.dialog = true
    },
    async save() {
      if (!this.$refs.form.validate()) return
      this.saving = true
      try {
        if (this.editItem) {
          await userService.update(this.editItem.id, this.form)
          this.$store.dispatch('snackbar/success', 'User updated')
        } else {
          await userService.create(this.form)
          this.$store.dispatch('snackbar/success', 'User created')
        }
        this.dialog = false
        this.load()
      } catch (err) {
        this.$store.dispatch('snackbar/error', err.response?.data?.message || 'Action failed')
      } finally { this.saving = false }
    },
    openResetDialog(item) { this.resetTarget = item; this.newPassword = ''; this.resetDialog = true },
    async resetPassword() {
      if (!this.newPassword || this.newPassword.length < 8) return
      this.saving = true
      try {
        await userService.resetPassword(this.resetTarget.id, this.newPassword)
        this.$store.dispatch('snackbar/success', 'Password reset successfully')
        this.resetDialog = false
      } catch (err) {
        this.$store.dispatch('snackbar/error', 'Failed to reset password')
      } finally { this.saving = false }
    },
    confirmDelete(u) { this.deleteTarget = u; this.deleteDialog = true },
    async deleteUser() {
      this.deleting = true
      try {
        await userService.delete(this.deleteTarget.id)
        this.$store.dispatch('snackbar/success', 'User deleted')
        this.deleteDialog = false
        this.load()
      } catch (err) {
        this.$store.dispatch('snackbar/error', err.response?.data?.message || 'Failed to delete user')
      } finally { this.deleting = false }
    }
  }
}
</script>
<style scoped>.gap-2 { gap: 8px; }</style>
