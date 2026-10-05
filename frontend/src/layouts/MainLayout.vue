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
        <span v-if="!miniVariant || $vuetify.breakpoint.mobile" class="white--text text-h6 font-weight-bold">Proprietor</span>
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

export default {
  name: 'MainLayout',
  data() {
    return {
      drawer: true,
      miniVariant: false,
      navItems: [
        { title: 'Dashboard', icon: 'mdi-view-dashboard', to: '/' },
        { title: 'Employee Directory', icon: 'mdi-account-group', to: '/users' },
        { title: 'Employee Tasks', icon: 'mdi-checkbox-marked-circle', to: '/tasks' },
        { title: 'Work Activities', icon: 'mdi-timeline', to: '/activities' },
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
    }
  },
  methods: {
    ...mapActions('auth', ['logout']),
    async logout() {
      await this.$store.dispatch('auth/logout')
    }
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
