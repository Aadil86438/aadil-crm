<template>
  <div class="pa-4 pa-md-6">
    <div class="d-flex align-center mb-6">
      <v-btn icon @click="$router.back()"><v-icon>mdi-arrow-left</v-icon></v-btn>
      <div class="ml-2">
        <h1 class="text-h5 font-weight-bold">{{ isEdit ? 'Edit Account' : 'New Account' }}</h1>
      </div>
    </div>
    <v-form ref="form" v-model="valid" @submit.prevent="save">
      <v-row>
        <v-col cols="12" md="8">
          <v-card outlined elevation="0">
            <v-card-title class="text-body-1 font-weight-bold"><v-icon left small>mdi-office-building</v-icon>Account Information</v-card-title>
            <v-divider />
            <v-card-text>
              <v-row dense>
                <v-col cols="12" sm="6"><v-text-field v-model="form.name" label="Account Name *" outlined dense :rules="[v => !!v || 'Required']" id="account-name" /></v-col>
                <v-col cols="12" sm="6"><v-select v-model="form.account_type" :items="accountTypes" label="Account Type" outlined dense /></v-col>
                <v-col cols="12" sm="6"><v-text-field v-model="form.website" label="Website" outlined dense /></v-col>
                <v-col cols="12" sm="6"><v-text-field v-model="form.industry" label="Industry" outlined dense /></v-col>
                <v-col cols="12" sm="6"><v-text-field v-model="form.phone" label="Phone" outlined dense /></v-col>
                <v-col cols="12" sm="6"><v-text-field v-model="form.email" label="Email" outlined dense /></v-col>
                <v-col cols="12" sm="6"><v-text-field v-model.number="form.num_employees" label="Employees" type="number" outlined dense /></v-col>
                <v-col cols="12" sm="6"><v-text-field v-model.number="form.annual_revenue" label="Annual Revenue (₹)" type="number" outlined dense /></v-col>
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
          <v-card outlined elevation="0" style="position: sticky; top: 80px">
            <v-card-text>
              <v-btn type="submit" color="primary" block large :loading="saving" :disabled="!valid" class="mb-3" id="account-save-btn">
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
import accountService from '../../services/accountService'
import userService from '../../services/userService'
export default {
  name: 'AccountForm',
  data() {
    return {
      valid: false, saving: false, users: [],
      accountTypes: ['Customer', 'Prospect', 'Partner', 'Reseller', 'Other'],
      form: { name: '', website: '', industry: '', phone: '', email: '', num_employees: null, annual_revenue: null, account_type: 'Prospect', owner_id: null, address: '', city: '', state: '', country: '', description: '' }
    }
  },
  computed: { isEdit() { return !!this.$route.params.id } },
  async mounted() {
    const users = await userService.listSimple(); this.users = users.data.data || []
    if (this.isEdit) { const res = await accountService.get(this.$route.params.id); Object.assign(this.form, res.data.data) }
  },
  methods: {
    async save() {
      if (!this.$refs.form.validate()) return
      this.saving = true
      try {
        if (this.isEdit) { await accountService.update(this.$route.params.id, this.form); this.$store.dispatch('snackbar/success', 'Account updated'); this.$router.push(`/accounts/${this.$route.params.id}`) }
        else { const res = await accountService.create(this.form); this.$store.dispatch('snackbar/success', 'Account created'); this.$router.push(`/accounts/${res.data.data.id}`) }
      } catch (err) { this.$store.dispatch('snackbar/error', 'Failed to save') } finally { this.saving = false }
    }
  }
}
</script>
