<template>
  <div class="pa-4 pa-md-6">
    <!-- Header -->
    <div class="d-flex align-center mb-6">
      <v-btn icon @click="$router.back()"><v-icon>mdi-arrow-left</v-icon></v-btn>
      <div class="ml-2">
        <h1 class="text-h5 font-weight-bold">{{ isEdit ? 'Edit Lead' : 'New Lead' }}</h1>
        <p class="text-body-2 grey--text mb-0">{{ isEdit ? 'Update lead information' : 'Add a new lead to your CRM' }}</p>
      </div>
    </div>

    <v-form ref="form" v-model="valid" @submit.prevent="saveLead">
      <v-row>
        <!-- Main Form -->
        <v-col cols="12" md="8">
          <!-- Basic Info Card -->
          <v-card class="mb-4" outlined elevation="0">
            <v-card-title class="text-body-1 font-weight-bold">
              <v-icon left small>mdi-account</v-icon>Basic Information
            </v-card-title>
            <v-divider />
            <v-card-text>
              <v-row dense>
                <v-col cols="12" sm="6">
                  <v-text-field v-model="form.first_name" label="First Name *" outlined dense :rules="[required]" id="lead-first-name" />
                </v-col>
                <v-col cols="12" sm="6">
                  <v-text-field v-model="form.last_name" label="Last Name *" outlined dense :rules="[required]" id="lead-last-name" />
                </v-col>
                <v-col cols="12" sm="6">
                  <v-text-field v-model="form.email" label="Email" type="email" outlined dense :rules="emailRules" id="lead-email" />
                </v-col>
                <v-col cols="12" sm="6">
                  <v-text-field v-model="form.company" label="Company" outlined dense id="lead-company" />
                </v-col>
                <v-col cols="12" sm="6">
                  <v-text-field v-model="form.phone" label="Phone" outlined dense id="lead-phone" />
                </v-col>
                <v-col cols="12" sm="6">
                  <v-text-field v-model="form.mobile" label="Mobile" outlined dense />
                </v-col>
                <v-col cols="12" sm="6">
                  <v-text-field v-model="form.job_title" label="Job Title" outlined dense />
                </v-col>
                <v-col cols="12" sm="6">
                  <v-text-field v-model="form.website" label="Website" outlined dense />
                </v-col>
              </v-row>
            </v-card-text>
          </v-card>

          <!-- Lead Info Card -->
          <v-card class="mb-4" outlined elevation="0">
            <v-card-title class="text-body-1 font-weight-bold">
              <v-icon left small>mdi-chart-bar</v-icon>Lead Details
            </v-card-title>
            <v-divider />
            <v-card-text>
              <v-row dense>
                <v-col cols="12" sm="6">
                  <v-select v-model="form.lead_status" label="Lead Status *" :items="leadStatuses" outlined dense :rules="[required]" id="lead-status" />
                </v-col>
                <v-col cols="12" sm="6">
                  <v-select v-model="form.lead_source" label="Lead Source" :items="leadSources" outlined dense clearable />
                </v-col>
                <v-col cols="12" sm="6">
                  <v-select v-model="form.industry" label="Industry" :items="industries" outlined dense clearable />
                </v-col>
                <v-col cols="12" sm="6">
                  <v-select v-model="form.rating" label="Rating" :items="['Hot','Warm','Cold']" outlined dense clearable />
                </v-col>
                <v-col cols="12" sm="6">
                  <v-text-field v-model.number="form.annual_revenue" label="Annual Revenue (₹)" type="number" outlined dense min="0" />
                </v-col>
                <v-col cols="12" sm="6">
                  <v-text-field v-model.number="form.num_employees" label="Number of Employees" type="number" outlined dense min="0" />
                </v-col>
                <v-col cols="12" sm="6">
                  <v-select v-model="form.owner_id" label="Lead Owner" :items="users" item-text="name" item-value="id" outlined dense clearable />
                </v-col>
              </v-row>
            </v-card-text>
          </v-card>

          <!-- Address Card -->
          <v-card class="mb-4" outlined elevation="0">
            <v-card-title class="text-body-1 font-weight-bold">
              <v-icon left small>mdi-map-marker</v-icon>Address
            </v-card-title>
            <v-divider />
            <v-card-text>
              <v-row dense>
                <v-col cols="12">
                  <v-text-field v-model="form.address" label="Address" outlined dense />
                </v-col>
                <v-col cols="12" sm="4">
                  <v-text-field v-model="form.city" label="City" outlined dense />
                </v-col>
                <v-col cols="12" sm="4">
                  <v-text-field v-model="form.state" label="State" outlined dense />
                </v-col>
                <v-col cols="12" sm="4">
                  <v-text-field v-model="form.country" label="Country" outlined dense />
                </v-col>
              </v-row>
            </v-card-text>
          </v-card>

          <!-- Description -->
          <v-card outlined elevation="0">
            <v-card-title class="text-body-1 font-weight-bold">
              <v-icon left small>mdi-text</v-icon>Description
            </v-card-title>
            <v-divider />
            <v-card-text>
              <v-textarea v-model="form.description" outlined dense rows="3" placeholder="Notes about this lead..." />
            </v-card-text>
          </v-card>
        </v-col>

        <!-- Sidebar Actions -->
        <v-col cols="12" md="4">
          <v-card outlined elevation="0" class="sticky-sidebar">
            <v-card-text>
              <v-btn type="submit" color="primary" block large :loading="saving" :disabled="!valid || saving" class="mb-3" id="lead-save-btn">
                <v-icon left>{{ isEdit ? 'mdi-content-save' : 'mdi-plus' }}</v-icon>
                {{ isEdit ? 'Save Changes' : 'Create Lead' }}
              </v-btn>
              <v-btn outlined block @click="$router.back()" :disabled="saving">Cancel</v-btn>

              <v-divider class="my-4" />

              <!-- Status Info -->
              <div v-if="isEdit" class="text-caption grey--text">
                <div class="mb-1"><strong>Created:</strong> {{ formatDate(original && original.created_at) }}</div>
                <div><strong>Updated:</strong> {{ formatDate(original && original.updated_at) }}</div>
              </div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-form>
  </div>
</template>

<script>
import leadService from '../../services/leadService'
import userService from '../../services/userService'

export default {
  name: 'LeadForm',
  data() {
    return {
      valid: false,
      saving: false,
      users: [],
      original: null,
      form: {
        first_name: '', last_name: '', email: '', phone: '', mobile: '',
        company: '', website: '', lead_source: '', lead_status: 'New',
        industry: '', job_title: '', annual_revenue: null, num_employees: null,
        rating: '', address: '', city: '', state: '', country: '', description: '',
        owner_id: null,
      },
      required: v => !!v || 'This field is required',
      emailRules: [v => !v || /.+@.+\..+/.test(v) || 'Enter a valid email'],
      leadStatuses: ['New', 'Contacted', 'Qualified', 'Unqualified', 'Converted'],
      leadSources: ['Website', 'Referral', 'Advertisement', 'Cold Call', 'Email', 'Social Media', 'Campaign', 'Other'],
      industries: ['Technology', 'Banking', 'Finance', 'Manufacturing', 'Healthcare', 'Education', 'Retail', 'Energy', 'FMCG', 'Automotive', 'Construction', 'Other'],
    }
  },
  computed: {
    isEdit() {
      return !!this.$route.params.id
    }
  },
  async mounted() {
    this.loadUsers()
    if (this.isEdit) {
      await this.loadLead()
    }
  },
  methods: {
    async loadUsers() {
      try {
        const res = await userService.listSimple()
        this.users = res.data.data || []
      } catch (e) { /* ignore */ }
    },
    async loadLead() {
      try {
        const res = await leadService.get(this.$route.params.id)
        this.original = res.data.data
        Object.assign(this.form, this.original)
      } catch (err) {
        this.$store.dispatch('snackbar/error', 'Failed to load lead')
        this.$router.back()
      }
    },
    async saveLead() {
      if (!this.$refs.form.validate()) return
      this.saving = true
      try {
        if (this.isEdit) {
          await leadService.update(this.$route.params.id, this.form)
          this.$store.dispatch('snackbar/success', 'Lead updated successfully')
          this.$router.push(`/leads/${this.$route.params.id}`)
        } else {
          const res = await leadService.create(this.form)
          this.$store.dispatch('snackbar/success', 'Lead created successfully')
          this.$router.push(`/leads/${res.data.data.id}`)
        }
      } catch (err) {
        this.$store.dispatch('snackbar/error', err.response?.data?.message || 'Failed to save lead')
      } finally {
        this.saving = false
      }
    },
    formatDate(d) {
      if (!d) return '—'
      return new Date(d).toLocaleDateString('en-IN')
    }
  }
}
</script>

<style scoped>
.sticky-sidebar { position: sticky; top: 80px; }
</style>
