<template>
  <div class="pa-4 pa-md-6">
    <div class="d-flex align-center mb-6">
      <v-btn icon @click="$router.back()"><v-icon>mdi-arrow-left</v-icon></v-btn>
      <div class="ml-2">
        <h1 class="text-h5 font-weight-bold">{{ isEdit ? 'Edit Deal' : 'New Deal' }}</h1>
      </div>
    </div>
    <v-form ref="form" v-model="valid" @submit.prevent="save">
      <v-row>
        <v-col cols="12" md="8">
          <v-card outlined elevation="0">
            <v-card-title class="text-body-1 font-weight-bold"><v-icon left small>mdi-briefcase</v-icon>Deal Details</v-card-title>
            <v-divider />
            <v-card-text>
              <v-row dense>
                <v-col cols="12" sm="6"><v-text-field v-model="form.name" label="Deal Name *" outlined dense :rules="[v => !!v || 'Required']" id="deal-name" /></v-col>
                <v-col cols="12" sm="6"><v-text-field v-model.number="form.amount" label="Amount (₹)" type="number" outlined dense min="0" id="deal-amount" /></v-col>
                <v-col cols="12" sm="6"><v-select v-model="form.stage" :items="stages" label="Stage *" outlined dense :rules="[v => !!v || 'Required']" id="deal-stage" /></v-col>
                <v-col cols="12" sm="6"><v-text-field v-model.number="form.probability" label="Probability (%)" type="number" outlined dense min="0" max="100" /></v-col>
                <v-col cols="12" sm="6"><v-menu v-model="dateMenu" :close-on-content-click="false" transition="scale-transition" offset-y min-width="auto"><template v-slot:activator="{ on, attrs }"><v-text-field v-model="form.expected_close_date" label="Expected Close Date" prepend-icon="mdi-calendar" readonly v-bind="attrs" v-on="on" outlined dense /></template><v-date-picker v-model="form.expected_close_date" @input="dateMenu = false" /></v-menu></v-col>
                <v-col cols="12" sm="6"><v-select v-model="form.lead_source" :items="leadSources" label="Lead Source" outlined dense clearable /></v-col>
                <v-col cols="12" sm="6"><v-select v-model="form.account_id" :items="accounts" item-text="name" item-value="id" label="Account" outlined dense clearable /></v-col>
                <v-col cols="12" sm="6"><v-select v-model="form.contact_id" :items="contacts" item-text="name" item-value="id" label="Contact" outlined dense clearable /></v-col>
                <v-col cols="12" sm="6"><v-select v-model="form.owner_id" :items="users" item-text="name" item-value="id" label="Owner" outlined dense clearable /></v-col>
                <v-col cols="12"><v-textarea v-model="form.description" label="Description" outlined dense rows="3" /></v-col>
              </v-row>
            </v-card-text>
          </v-card>
        </v-col>
        <v-col cols="12" md="4">
          <v-card outlined elevation="0" style="position: sticky; top: 80px">
            <v-card-text>
              <v-btn type="submit" color="primary" block large :loading="saving" :disabled="!valid" class="mb-3" id="deal-save-btn">
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
import dealService from '../../services/dealService'
import accountService from '../../services/accountService'
import contactService from '../../services/contactService'
import userService from '../../services/userService'

export default {
  name: 'DealForm',
  data() {
    return {
      valid: false, saving: false, accounts: [], contacts: [], users: [], dateMenu: false,
      stages: ['Qualification', 'Needs Analysis', 'Proposal', 'Negotiation', 'Closed Won', 'Closed Lost'],
      leadSources: ['Website', 'Referral', 'Advertisement', 'Cold Call', 'Email', 'Social Media', 'Campaign', 'Other'],
      form: { name: '', amount: null, stage: 'Qualification', probability: 10, expected_close_date: null, lead_source: '', account_id: null, contact_id: null, owner_id: null, description: '' }
    }
  },
  computed: { isEdit() { return !!this.$route.params.id } },
  async mounted() {
    const [accs, cons, usrs] = await Promise.all([accountService.listSimple(), contactService.listSimple(), userService.listSimple()])
    this.accounts = accs.data.data || []
    this.contacts = (cons.data.data || []).map(c => ({ id: c.id, name: `${c.first_name} ${c.last_name}` }))
    this.users = usrs.data.data || []
    if (this.isEdit) {
      const res = await dealService.get(this.$route.params.id)
      const data = res.data.data
      if (data.expected_close_date) data.expected_close_date = data.expected_close_date.split('T')[0]
      Object.assign(this.form, data)
    }
  },
  methods: {
    async save() {
      if (!this.$refs.form.validate()) return
      this.saving = true
      try {
        if (this.isEdit) {
          await dealService.update(this.$route.params.id, this.form)
          this.$store.dispatch('snackbar/success', 'Deal updated')
          this.$router.push(`/deals/${this.$route.params.id}`)
        } else {
          const res = await dealService.create(this.form)
          this.$store.dispatch('snackbar/success', 'Deal created')
          this.$router.push(`/deals/${res.data.data.id}`)
        }
      } catch (err) {
        this.$store.dispatch('snackbar/error', 'Failed to save deal')
      } finally {
        this.saving = false
      }
    }
  }
}
</script>
