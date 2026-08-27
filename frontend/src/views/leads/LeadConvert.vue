<template>
  <div class="pa-4 pa-md-6">
    <div class="d-flex align-center mb-6">
      <v-btn icon @click="$router.back()"><v-icon>mdi-arrow-left</v-icon></v-btn>
      <div class="ml-2">
        <h1 class="text-h5 font-weight-bold">Convert Lead</h1>
        <p class="text-body-2 grey--text mb-0">Convert this lead to contact, account, and/or deal</p>
      </div>
    </div>

    <div v-if="loading" class="text-center py-16">
      <v-progress-circular indeterminate color="primary" />
    </div>

    <template v-else-if="lead">
      <!-- Lead Summary -->
      <v-alert type="info" outlined class="mb-4">
        <strong>{{ lead.first_name }} {{ lead.last_name }}</strong> from {{ lead.company }}
        <v-chip small class="ml-2" :color="lead.lead_status === 'Qualified' ? 'green' : 'orange'" dark>{{ lead.lead_status }}</v-chip>
      </v-alert>

      <v-alert type="error" v-if="lead.is_converted" outlined class="mb-4">
        This lead has already been converted.
        <v-btn small text :to="`/leads/${lead.id}`">View Lead</v-btn>
      </v-alert>

      <template v-if="!lead.is_converted">
        <v-row>
          <v-col cols="12" md="8">
            <!-- Create Contact -->
            <v-card class="mb-4" outlined elevation="0">
              <v-card-title>
                <v-checkbox v-model="form.create_contact" color="primary" hide-details class="mt-0 pt-0 mr-2" />
                <v-icon left color="cyan">mdi-contacts</v-icon>
                Create Contact
              </v-card-title>
              <v-expand-transition>
                <div v-if="form.create_contact">
                  <v-divider />
                  <v-card-text>
                    <p class="text-body-2 grey--text">A contact will be created with the following data from this lead:</p>
                    <v-list dense>
                      <v-list-item><v-list-item-content><v-list-item-title>{{ lead.first_name }} {{ lead.last_name }}</v-list-item-title><v-list-item-subtitle>Name</v-list-item-subtitle></v-list-item-content></v-list-item>
                      <v-list-item><v-list-item-content><v-list-item-title>{{ lead.email || '—' }}</v-list-item-title><v-list-item-subtitle>Email</v-list-item-subtitle></v-list-item-content></v-list-item>
                      <v-list-item><v-list-item-content><v-list-item-title>{{ lead.job_title || '—' }}</v-list-item-title><v-list-item-subtitle>Job Title</v-list-item-subtitle></v-list-item-content></v-list-item>
                    </v-list>
                  </v-card-text>
                </div>
              </v-expand-transition>
            </v-card>

            <!-- Create Account -->
            <v-card class="mb-4" outlined elevation="0">
              <v-card-title>
                <v-checkbox v-model="form.create_account" color="primary" hide-details class="mt-0 pt-0 mr-2" />
                <v-icon left color="teal">mdi-office-building</v-icon>
                Create Account
              </v-card-title>
              <v-expand-transition>
                <div v-if="form.create_account">
                  <v-divider />
                  <v-card-text>
                    <p class="text-body-2 grey--text">Account will be created for: <strong>{{ lead.company }}</strong></p>
                  </v-card-text>
                </div>
              </v-expand-transition>
              <v-expand-transition>
                <div v-if="!form.create_account">
                  <v-divider />
                  <v-card-text>
                    <p class="text-body-2 grey--text">Or link to existing account:</p>
                    <v-select v-model="form.account_id" :items="accounts" item-text="name" item-value="id" label="Select Account" outlined dense clearable />
                  </v-card-text>
                </div>
              </v-expand-transition>
            </v-card>

            <!-- Create Deal -->
            <v-card class="mb-4" outlined elevation="0">
              <v-card-title>
                <v-checkbox v-model="form.create_deal" color="primary" hide-details class="mt-0 pt-0 mr-2" />
                <v-icon left color="purple">mdi-briefcase</v-icon>
                Create Deal
              </v-card-title>
              <v-expand-transition>
                <div v-if="form.create_deal">
                  <v-divider />
                  <v-card-text>
                    <v-text-field v-model="form.deal_name" label="Deal Name *" outlined dense :rules="[v => !form.create_deal || !!v || 'Deal name is required']" id="deal-name" />
                    <v-text-field v-model.number="form.deal_amount" label="Deal Amount (₹)" type="number" outlined dense min="0" id="deal-amount" />
                  </v-card-text>
                </div>
              </v-expand-transition>
            </v-card>

            <!-- Convert Button -->
            <v-btn color="success" large :loading="converting" @click="convertLead" :disabled="converting" block id="convert-lead-btn">
              <v-icon left>mdi-account-convert</v-icon>
              Convert Lead
            </v-btn>
          </v-col>

          <!-- Summary Panel -->
          <v-col cols="12" md="4">
            <v-card outlined elevation="0">
              <v-card-title class="text-body-1 font-weight-bold">What will be created</v-card-title>
              <v-divider />
              <v-card-text>
                <v-list dense>
                  <v-list-item>
                    <v-list-item-icon>
                      <v-icon :color="form.create_contact ? 'success' : 'grey'">{{ form.create_contact ? 'mdi-check-circle' : 'mdi-circle-outline' }}</v-icon>
                    </v-list-item-icon>
                    <v-list-item-content><v-list-item-title>Contact</v-list-item-title></v-list-item-content>
                  </v-list-item>
                  <v-list-item>
                    <v-list-item-icon>
                      <v-icon :color="form.create_account ? 'success' : (form.account_id ? 'blue' : 'grey')">{{ form.create_account ? 'mdi-check-circle' : (form.account_id ? 'mdi-link' : 'mdi-circle-outline') }}</v-icon>
                    </v-list-item-icon>
                    <v-list-item-content><v-list-item-title>Account</v-list-item-title></v-list-item-content>
                  </v-list-item>
                  <v-list-item>
                    <v-list-item-icon>
                      <v-icon :color="form.create_deal ? 'success' : 'grey'">{{ form.create_deal ? 'mdi-check-circle' : 'mdi-circle-outline' }}</v-icon>
                    </v-list-item-icon>
                    <v-list-item-content><v-list-item-title>Deal</v-list-item-title></v-list-item-content>
                  </v-list-item>
                </v-list>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>
      </template>
    </template>

    <!-- Success Dialog -->
    <v-dialog v-model="successDialog" max-width="500" persistent>
      <v-card>
        <v-card-title class="success--text"><v-icon left color="success">mdi-check-circle</v-icon>Lead Converted!</v-card-title>
        <v-card-text>
          <p>The lead has been successfully converted.</p>
          <v-list dense>
            <v-list-item v-if="result && result.contact_id">
              <v-list-item-icon><v-icon color="cyan">mdi-contacts</v-icon></v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>Contact created</v-list-item-title>
              </v-list-item-content>
              <v-list-item-action>
                <v-btn small text color="primary" :to="`/contacts/${result.contact_id}`">View</v-btn>
              </v-list-item-action>
            </v-list-item>
            <v-list-item v-if="result && result.account_id">
              <v-list-item-icon><v-icon color="teal">mdi-office-building</v-icon></v-list-item-icon>
              <v-list-item-content><v-list-item-title>Account created</v-list-item-title></v-list-item-content>
              <v-list-item-action>
                <v-btn small text color="primary" :to="`/accounts/${result.account_id}`">View</v-btn>
              </v-list-item-action>
            </v-list-item>
            <v-list-item v-if="result && result.deal_id">
              <v-list-item-icon><v-icon color="purple">mdi-briefcase</v-icon></v-list-item-icon>
              <v-list-item-content><v-list-item-title>Deal created</v-list-item-title></v-list-item-content>
              <v-list-item-action>
                <v-btn small text color="primary" :to="`/deals/${result.deal_id}`">View</v-btn>
              </v-list-item-action>
            </v-list-item>
          </v-list>
        </v-card-text>
        <v-card-actions>
          <v-btn text @click="$router.push('/leads')">Back to Leads</v-btn>
          <v-spacer />
          <v-btn color="primary" @click="successDialog = false">Done</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import leadService from '../../services/leadService'
import accountService from '../../services/accountService'

export default {
  name: 'LeadConvert',
  data() {
    return {
      lead: null,
      accounts: [],
      loading: true,
      converting: false,
      successDialog: false,
      result: null,
      form: {
        create_contact: true,
        create_account: false,
        create_deal: false,
        account_id: null,
        deal_name: '',
        deal_amount: null,
      }
    }
  },
  async mounted() {
    this.loading = true
    try {
      const [leadRes, accountsRes] = await Promise.all([
        leadService.get(this.$route.params.id),
        accountService.listSimple()
      ])
      this.lead = leadRes.data.data
      this.accounts = accountsRes.data.data || []
      if (this.lead.company) {
        this.form.deal_name = this.lead.company + ' - ' + new Date().getFullYear()
      }
    } catch (err) {
      this.$store.dispatch('snackbar/error', 'Failed to load lead')
    } finally {
      this.loading = false
    }
  },
  methods: {
    async convertLead() {
      if (this.form.create_deal && !this.form.deal_name) {
        this.$store.dispatch('snackbar/error', 'Please enter a deal name')
        return
      }
      this.converting = true
      try {
        const payload = {
          create_contact: this.form.create_contact,
          create_account: this.form.create_account,
          create_deal: this.form.create_deal,
          account_id: this.form.account_id,
          deal_name: this.form.deal_name,
          deal_amount: this.form.deal_amount,
        }
        const res = await leadService.convert(this.lead.id, payload)
        this.result = res.data.data
        this.successDialog = true
        this.$store.dispatch('snackbar/success', 'Lead converted successfully!')
      } catch (err) {
        this.$store.dispatch('snackbar/error', err.response?.data?.message || 'Conversion failed')
      } finally {
        this.converting = false
      }
    }
  }
}
</script>
