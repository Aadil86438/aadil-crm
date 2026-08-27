<template>
  <v-app>
    <!-- Navigation Drawer (Sidebar) -->
    <v-navigation-drawer
      v-model="drawer"
      app
      :mini-variant="miniVariant && !$vuetify.breakpoint.mobile"
      :permanent="!$vuetify.breakpoint.mobile"
      :temporary="$vuetify.breakpoint.mobile"
      color="sidebar"
      dark
      width="260"
      mini-variant-width="64"
    >
      <!-- Logo -->
      <div class="sidebar-logo d-flex align-center pa-4" :class="miniVariant && !$vuetify.breakpoint.mobile ? 'justify-center' : ''">
        <v-icon color="white" size="28" class="mr-2" :class="miniVariant && !$vuetify.breakpoint.mobile ? 'mr-0' : ''">mdi-home-city</v-icon>
        <span v-if="!miniVariant || $vuetify.breakpoint.mobile" class="white--text text-h6 font-weight-bold">Propertier</span>
        <v-spacer v-if="!miniVariant || $vuetify.breakpoint.mobile" />
        <v-btn icon small @click="miniVariant = !miniVariant" v-if="!$vuetify.breakpoint.mobile">
          <v-icon color="white" small>{{ miniVariant ? 'mdi-chevron-right' : 'mdi-chevron-left' }}</v-icon>
        </v-btn>
      </div>

      <v-divider dark class="opacity-20" />

      <!-- Navigation Items -->
      <v-list nav dense class="pa-2 mt-2">
        <v-list-item
          v-for="item in navItems"
          :key="item.title"
          :to="item.to"
          exact-path
          active-class="sidebar-active"
          class="sidebar-item mb-1"
          :disabled="item.hidden && !isAdmin"
          v-show="!item.hidden || isAdmin"
        >
          <v-list-item-icon class="mr-3">
            <v-icon size="20">{{ item.icon }}</v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title class="text-body-2 font-weight-medium">{{ item.title }}</v-list-item-title>
          </v-list-item-content>
          <v-chip v-if="item.badge" x-small color="accent" class="ml-1">{{ item.badge }}</v-chip>
        </v-list-item>
      </v-list>

      <!-- Bottom User Section -->
      <template v-slot:append>
        <v-divider dark class="opacity-20" />
        <div class="pa-3">
          <v-list-item class="px-1">
            <v-list-item-avatar size="32" color="primary">
              <span class="white--text text-caption font-weight-bold">{{ userInitials }}</span>
            </v-list-item-avatar>
            <v-list-item-content v-if="!miniVariant || $vuetify.breakpoint.mobile">
              <v-list-item-title class="white--text text-body-2 font-weight-medium">{{ currentUser.name }}</v-list-item-title>
              <v-list-item-subtitle class="text-caption" style="color: rgba(255,255,255,0.6)">{{ roleLabel }}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
        </div>
      </template>
    </v-navigation-drawer>

    <!-- Top App Bar -->
    <v-app-bar
      app
      color="white"
      elevation="1"
      height="64"
    >
      <!-- Mobile menu toggle -->
      <v-app-bar-nav-icon
        v-if="$vuetify.breakpoint.mobile"
        @click="drawer = !drawer"
      />

      <!-- Page Title / Breadcrumbs -->
      <v-toolbar-title class="d-flex align-center">
        <span class="text-subtitle-1 font-weight-semibold grey--text text--darken-2">{{ $route.name }}</span>
      </v-toolbar-title>

      <v-spacer />

      <!-- Global Search Bar -->
      <div class="search-container mr-3" :class="$vuetify.breakpoint.xs ? 'd-none' : ''">
        <v-text-field
          v-model="searchQuery"
          placeholder="Search leads, contacts, deals..."
          prepend-inner-icon="mdi-magnify"
          dense
          outlined
          hide-details
          class="search-field"
          style="width: 300px;"
          @keyup.enter="goSearch"
          @input="onSearchInput"
          clearable
          id="global-search"
        >
          <!-- Search Results Dropdown -->
          <template v-slot:append v-if="searchResults && searchQuery">
            <v-menu v-model="searchMenu" :close-on-click="true" offset-y>
              <template v-slot:activator="{ on }"></template>
            </v-menu>
          </template>
        </v-text-field>

        <!-- Search dropdown -->
        <v-card v-if="searchMenu && searchResults" class="search-dropdown elevation-8" width="400">
          <v-list dense>
            <template v-for="(items, category) in filteredResults">
              <template v-if="items && items.length > 0">
                <v-subheader :key="'header-'+category" class="text-caption font-weight-bold">{{ categoryLabel(category) }}</v-subheader>
                <v-list-item
                  v-for="item in items"
                  :key="item.id"
                  @click="navigateToResult(item)"
                  class="py-0"
                >
                  <v-list-item-icon class="mr-3">
                    <v-icon small :color="typeColor(item.type)">{{ typeIcon(item.type) }}</v-icon>
                  </v-list-item-icon>
                  <v-list-item-content>
                    <v-list-item-title class="text-body-2">{{ item.title }}</v-list-item-title>
                    <v-list-item-subtitle class="text-caption">{{ item.subtitle }}</v-list-item-subtitle>
                  </v-list-item-content>
                </v-list-item>
              </template>
            </template>
            <v-list-item @click="goSearch" class="py-2">
              <v-list-item-content>
                <v-list-item-title class="text-caption primary--text text-center">View all results →</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-card>
      </div>

      <!-- Mobile Search -->
      <v-btn icon v-if="$vuetify.breakpoint.xs" @click="$router.push('/search')">
        <v-icon>mdi-magnify</v-icon>
      </v-btn>

      <!-- User Menu -->
      <v-menu offset-y left>
        <template v-slot:activator="{ on, attrs }">
          <v-btn icon v-bind="attrs" v-on="on" class="ml-1" id="user-menu-btn">
            <v-avatar size="36" color="primary">
              <span class="white--text text-caption font-weight-bold">{{ userInitials }}</span>
            </v-avatar>
          </v-btn>
        </template>
        <v-list dense min-width="200">
          <v-list-item class="py-2">
            <v-list-item-content>
              <v-list-item-title class="font-weight-medium">{{ currentUser.name }}</v-list-item-title>
              <v-list-item-subtitle>{{ currentUser.email }}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
          <v-divider />
          <v-list-item :to="isAdmin ? '/users' : '/'">
            <v-list-item-icon><v-icon small>mdi-cog</v-icon></v-list-item-icon>
            <v-list-item-content><v-list-item-title>Settings</v-list-item-title></v-list-item-content>
          </v-list-item>
          <v-divider />
          <v-list-item @click="logout" id="logout-btn">
            <v-list-item-icon><v-icon small color="error">mdi-logout</v-icon></v-list-item-icon>
            <v-list-item-content><v-list-item-title class="error--text">Logout</v-list-item-title></v-list-item-content>
          </v-list-item>
        </v-list>
      </v-menu>
    </v-app-bar>

    <!-- Main Content -->
    <v-main class="main-content">
      <router-view />
    </v-main>
  </v-app>
</template>

<script>
import { mapGetters, mapActions } from 'vuex'
import searchService from '../services/searchService'

export default {
  name: 'MainLayout',
  data() {
    return {
      drawer: true,
      miniVariant: false,
      searchQuery: '',
      searchResults: null,
      searchMenu: false,
      searchTimer: null,
      navItems: [
        { title: 'Dashboard', icon: 'mdi-view-dashboard', to: '/' },
        { title: 'Leads', icon: 'mdi-account-arrow-right', to: '/leads' },
        { title: 'Contacts', icon: 'mdi-contacts', to: '/contacts' },
        { title: 'Accounts', icon: 'mdi-office-building', to: '/accounts' },
        { title: 'Deals', icon: 'mdi-briefcase', to: '/deals' },
        { title: 'Tasks', icon: 'mdi-checkbox-marked-circle', to: '/tasks' },
        { title: 'Activities', icon: 'mdi-timeline', to: '/activities' },
        { title: 'Calendar', icon: 'mdi-calendar', to: '/calendar' },
        { title: 'Reports', icon: 'mdi-chart-bar', to: '/reports' },
        { title: 'Users', icon: 'mdi-account-group', to: '/users' },
        { title: 'Audit Log', icon: 'mdi-shield-check', to: '/audit-log', hidden: true },
      ]
    }
  },
  computed: {
    ...mapGetters('auth', ['user', 'isAdmin']),
    currentUser() {
      return this.user || { name: 'User', email: '', role: '' }
    },
    userInitials() {
      if (!this.currentUser.name) return 'U'
      return this.currentUser.name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2)
    },
    roleLabel() {
      const map = { admin: 'Admin', manager: 'Manager', sales_user: 'Team Member' }
      return map[this.currentUser.role] || 'User'
    },
    filteredResults() {
      if (!this.searchResults) return {}
      return {
        leads: this.searchResults.leads || [],
        contacts: this.searchResults.contacts || [],
        accounts: this.searchResults.accounts || [],
        deals: this.searchResults.deals || [],
      }
    }
  },
  methods: {
    ...mapActions('auth', ['logout']),
    async logout() {
      await this.$store.dispatch('auth/logout')
    },
    onSearchInput() {
      clearTimeout(this.searchTimer)
      if (!this.searchQuery || this.searchQuery.length < 2) {
        this.searchMenu = false
        this.searchResults = null
        return
      }
      this.searchTimer = setTimeout(async () => {
        try {
          const res = await searchService.search(this.searchQuery)
          this.searchResults = res.data.data
          this.searchMenu = true
        } catch (e) { /* ignore */ }
      }, 300)
    },
    goSearch() {
      this.searchMenu = false
      if (this.searchQuery) {
        this.$router.push({ name: 'Search', query: { q: this.searchQuery } })
      }
    },
    navigateToResult(item) {
      this.searchMenu = false
      this.searchQuery = ''
      const routes = {
        lead: `/leads/${item.id}`,
        contact: `/contacts/${item.id}`,
        account: `/accounts/${item.id}`,
        deal: `/deals/${item.id}`,
      }
      if (routes[item.type]) this.$router.push(routes[item.type])
    },
    categoryLabel(cat) {
      const map = { leads: 'LEADS', contacts: 'CONTACTS', accounts: 'ACCOUNTS', deals: 'DEALS', tasks: 'TASKS' }
      return map[cat] || cat.toUpperCase()
    },
    typeIcon(type) {
      const icons = { lead: 'mdi-account-arrow-right', contact: 'mdi-contacts', account: 'mdi-office-building', deal: 'mdi-briefcase', task: 'mdi-checkbox-marked-circle' }
      return icons[type] || 'mdi-circle'
    },
    typeColor(type) {
      const colors = { lead: 'orange', contact: 'blue', account: 'green', deal: 'purple', task: 'teal' }
      return colors[type] || 'grey'
    }
  },
  mounted() {
    // Close search dropdown when clicking outside
    document.addEventListener('click', e => {
      if (!this.$el.contains(e.target)) {
        this.searchMenu = false
      }
    })
  }
}
</script>

<style scoped>
.sidebar-logo {
  height: 64px;
  background: rgba(0,0,0,0.1);
}
.sidebar-item {
  border-radius: 8px !important;
  color: rgba(255,255,255,0.8) !important;
  transition: all 0.2s ease;
}
.sidebar-item:hover {
  background: rgba(255,255,255,0.1) !important;
  color: white !important;
}
.sidebar-active {
  background: rgba(255,255,255,0.15) !important;
  color: white !important;
  font-weight: 600 !important;
}
.opacity-20 { opacity: 0.2; }
.main-content { background-color: #F5F7FA; }
.search-container { position: relative; }
.search-field >>> .v-input__slot { border-radius: 8px !important; background: #F5F7FA !important; }
.search-dropdown {
  position: absolute;
  top: 100%;
  right: 0;
  z-index: 200;
  max-height: 400px;
  overflow-y: auto;
  border-radius: 12px !important;
}
</style>
