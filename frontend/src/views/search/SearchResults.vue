<template>
  <div class="pa-4 pa-md-6">
    <div class="mb-6">
      <h1 class="text-h5 font-weight-bold">Search Results</h1>
      <p class="text-body-2 grey--text mb-0">Results for "{{ query }}"</p>
    </div>

    <div v-if="loading" class="text-center py-16"><v-progress-circular indeterminate color="primary" /></div>

    <template v-else-if="results">
      <v-row>
        <!-- Leads -->
        <v-col cols="12" md="6" v-if="results.leads && results.leads.length">
          <v-card outlined elevation="0" class="mb-4">
            <v-card-title class="text-body-1 font-weight-bold"><v-icon left color="orange">mdi-account-arrow-right</v-icon>Leads</v-card-title>
            <v-divider />
            <v-list dense>
              <v-list-item v-for="item in results.leads" :key="item.id" :to="`/leads/${item.id}`">
                <v-list-item-content>
                  <v-list-item-title class="font-weight-medium">{{ item.title }}</v-list-item-title>
                  <v-list-item-subtitle>{{ item.subtitle }}</v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </v-card>
        </v-col>

        <!-- Contacts -->
        <v-col cols="12" md="6" v-if="results.contacts && results.contacts.length">
          <v-card outlined elevation="0" class="mb-4">
            <v-card-title class="text-body-1 font-weight-bold"><v-icon left color="blue">mdi-contacts</v-icon>Contacts</v-card-title>
            <v-divider />
            <v-list dense>
              <v-list-item v-for="item in results.contacts" :key="item.id" :to="`/contacts/${item.id}`">
                <v-list-item-content>
                  <v-list-item-title class="font-weight-medium">{{ item.title }}</v-list-item-title>
                  <v-list-item-subtitle>{{ item.subtitle }}</v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </v-card>
        </v-col>

        <!-- Accounts -->
        <v-col cols="12" md="6" v-if="results.accounts && results.accounts.length">
          <v-card outlined elevation="0" class="mb-4">
            <v-card-title class="text-body-1 font-weight-bold"><v-icon left color="teal">mdi-office-building</v-icon>Accounts</v-card-title>
            <v-divider />
            <v-list dense>
              <v-list-item v-for="item in results.accounts" :key="item.id" :to="`/accounts/${item.id}`">
                <v-list-item-content>
                  <v-list-item-title class="font-weight-medium">{{ item.title }}</v-list-item-title>
                  <v-list-item-subtitle>{{ item.subtitle }}</v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </v-card>
        </v-col>

        <!-- Deals -->
        <v-col cols="12" md="6" v-if="results.deals && results.deals.length">
          <v-card outlined elevation="0" class="mb-4">
            <v-card-title class="text-body-1 font-weight-bold"><v-icon left color="purple">mdi-briefcase</v-icon>Deals</v-card-title>
            <v-divider />
            <v-list dense>
              <v-list-item v-for="item in results.deals" :key="item.id" :to="`/deals/${item.id}`">
                <v-list-item-content>
                  <v-list-item-title class="font-weight-medium">{{ item.title }}</v-list-item-title>
                  <v-list-item-subtitle>{{ item.subtitle }}</v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </v-card>
        </v-col>
      </v-row>

      <div v-if="isEmpty" class="text-center py-16 grey--text">
        <v-icon size="64" color="grey lighten-2">mdi-magnify</v-icon>
        <p class="text-h6 mt-4">No matching records found</p>
      </div>
    </template>
  </div>
</template>

<script>
import searchService from '../../services/searchService'

export default {
  name: 'SearchResults',
  data() {
    return {
      query: '',
      loading: false,
      results: null,
    }
  },
  computed: {
    isEmpty() {
      if (!this.results) return true
      return !((this.results.leads && this.results.leads.length) ||
        (this.results.contacts && this.results.contacts.length) ||
        (this.results.accounts && this.results.accounts.length) ||
        (this.results.deals && this.results.deals.length))
    }
  },
  watch: {
    '$route.query.q': { handler: 'performSearch', immediate: true }
  },
  methods: {
    async performSearch(q) {
      this.query = q || ''
      if (!this.query || this.query.length < 2) return
      this.loading = true
      try {
        const res = await searchService.search(this.query)
        this.results = res.data.data
      } finally {
        this.loading = false
      }
    }
  }
}
</script>
