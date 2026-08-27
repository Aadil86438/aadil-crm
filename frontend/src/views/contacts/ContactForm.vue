<template>
  <div class="pa-4 pa-md-6">
    <div class="d-flex align-center mb-6">
      <v-btn icon @click="$router.back()"><v-icon>mdi-arrow-left</v-icon></v-btn>
      <div class="ml-2">
        <h1 class="text-h5 font-weight-bold">{{ isEdit ? 'Edit Contact' : 'New Contact' }}</h1>
      </div>
    </div>

    <v-form ref="form" v-model="valid" @submit.prevent="save">
      <v-row>
        <v-col cols="12" md="8">
          <v-card outlined elevation="0" class="mb-4">
            <v-card-title class="text-body-1 font-weight-bold"><v-icon left small>mdi-account</v-icon>Contact Information</v-card-title>
            <v-divider />
            <v-card-text>
              <v-row dense>
                <v-col cols="12" sm="6"><v-text-field v-model="form.first_name" label="First Name *" outlined dense :rules="[v => !!v || 'Required']" id="contact-first-name" /></v-col>
                <v-col cols="12" sm="6"><v-text-field v-model="form.last_name" label="Last Name *" outlined dense :rules="[v => !!v || 'Required']" id="contact-last-name" /></v-col>
                <v-col cols="12" sm="6"><v-text-field v-model="form.email" label="Email" type="email" outlined dense id="contact-email" /></v-col>
                <v-col cols="12" sm="6"><v-text-field v-model="form.phone" label="Phone" outlined dense /></v-col>
                <v-col cols="12" sm="6"><v-text-field v-model="form.mobile" label="Mobile" outlined dense /></v-col>
                <v-col cols="12" sm="6"><v-text-field v-model="form.job_title" label="Job Title" outlined dense /></v-col>
                <v-col cols="12" sm="6"><v-text-field v-model="form.department" label="Department" outlined dense /></v-col>
                <v-col cols="12" sm="6"><v-select v-model="form.account_id" :items="accounts" item-text="name" item-value="id" label="Account" outlined dense clearable /></v-col>
                <v-col cols="12" sm="6"><v-select v-model="form.owner_id" :items="users" item-text="name" item-value="id" label="Owner" outlined dense clearable /></v-col>
                <v-col cols="12"><v-text-field v-model="form.address" label="Address" outlined dense /></v-col>
                <v-col cols="12" sm="4"><v-text-field v-model="form.city" label="City" outlined dense /></v-col>
                <v-col cols="12" sm="4"><v-text-field v-model="form.state" label="State" outlined dense /></v-col>
                <v-col cols="12" sm="4"><v-text-field v-model="form.country" label="Country" outlined dense /></v-col>
                <v-col cols="12"><v-textarea v-model="form.description" label="Description" outlined dense rows="3" /></v-col>
              </v-row>
            </v-card-text>
          </v-card>
        </v-col>
        <v-col cols="12" md="4">
          <v-card outlined elevation="0" class="sticky-sidebar">
            <v-card-text>
              <v-btn type="submit" color="primary" block large :loading="saving" :disabled="!valid" class="mb-3" id="contact-save-btn">
                <v-icon left>mdi-content-save</v-icon>{{ isEdit ? 'Save' : 'Create' }}
              </v-btn>
              <v-btn outlined block @click="$router.back()">Cancel</v-btn>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-form>
  </div>
</template>

<script>
import contactService from '../../services/contactService'
import accountService from '../../services/accountService'
import userService from '../../services/userService'

export default {
  name: 'ContactForm',
  data() {
    return {
      valid: false, saving: false, accounts: [], users: [],
      form: { first_name: '', last_name: '', email: '', phone: '', mobile: '', job_title: '', department: '', account_id: null, owner_id: null, address: '', city: '', state: '', country: '', description: '' }
    }
  },
  computed: { isEdit() { return !!this.$route.params.id } },
  async mounted() {
    const [accounts, users] = await Promise.all([accountService.listSimple(), userService.listSimple()])
    this.accounts = accounts.data.data || []
    this.users = users.data.data || []
    if (this.isEdit) {
      const res = await contactService.get(this.$route.params.id)
      Object.assign(this.form, res.data.data)
    }
  },
  methods: {
    async save() {
      if (!this.$refs.form.validate()) return
      this.saving = true
      try {
        if (this.isEdit) {
          await contactService.update(this.$route.params.id, this.form)
          this.$store.dispatch('snackbar/success', 'Contact updated')
          this.$router.push(`/contacts/${this.$route.params.id}`)
        } else {
          const res = await contactService.create(this.form)
          this.$store.dispatch('snackbar/success', 'Contact created')
          this.$router.push(`/contacts/${res.data.data.id}`)
        }
      } catch (err) {
        this.$store.dispatch('snackbar/error', err.response?.data?.message || 'Failed to save')
      } finally { this.saving = false }
    }
  }
}
</script>
<style scoped>
.sticky-sidebar { position: sticky; top: 80px; }
</style>
