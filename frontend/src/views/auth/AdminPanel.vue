<template>
  <div class="admin-bg fill-height py-8 px-4">
    <v-container fluid style="max-width: 1200px">
      <!-- Code Entry (not authenticated) -->
      <v-card v-if="!authenticated" class="code-card mx-auto" style="max-width: 500px" elevation="0">
        <div class="code-header pa-6 text-center">
          <v-icon size="44" color="white" class="mb-3">mdi-shield-lock</v-icon>
          <h2 class="white--text text-h5 font-weight-bold">Admin Access</h2>
          <p class="white--text text-body-2 mt-1" style="opacity: 0.7">Enter 4-digit admin code to access member management</p>
        </div>
        <v-card-text class="pa-6">
          <v-form @submit.prevent="verifyCode" ref="codeForm">
            <v-text-field
              v-model="adminCode"
              label="Admin Code"
              placeholder="Enter 4-digit admin code"
              prepend-inner-icon="mdi-key"
              :type="showCode ? 'text' : 'password'"
              :append-icon="showCode ? 'mdi-eye' : 'mdi-eye-off'"
              @click:append="showCode = !showCode"
              outlined dense color="primary"
              :rules="[v => !!v || 'Code is required']"
              maxlength="4"
              class="mb-3 code-input"
              id="admin-code-input"
              autofocus
            />
            <v-alert v-if="codeError" type="error" dense text class="mb-3">{{ codeError }}</v-alert>
            <v-btn
              type="submit" color="primary" block large
              :loading="verifying"
              class="verify-btn"
              id="admin-verify-btn"
            >
              <v-icon left>mdi-login</v-icon>
              Verify & Access Panel
            </v-btn>
          </v-form>
          <div class="text-center mt-4">
            <v-btn text small color="grey" @click="$router.push('/register')">
              <v-icon left small>mdi-arrow-left</v-icon>
              Back to Home
            </v-btn>
          </div>
        </v-card-text>
      </v-card>

      <!-- Admin Dashboard (authenticated) -->
      <div v-else>
        <!-- Top App Bar -->
        <v-card class="mb-6 pa-5 banner-card" elevation="0">
          <div class="d-flex align-center justify-space-between flex-wrap" style="gap: 16px">
            <div class="d-flex align-center">
              <v-avatar color="primary" size="48" class="mr-3 elevation-2">
                <v-icon color="white">mdi-shield-account</v-icon>
              </v-avatar>
              <div>
                <div class="d-flex align-center">
                  <span class="text-h5 font-weight-black text--primary mr-2">Proprietor CRM</span>
                  <v-chip color="primary" small label class="font-weight-bold">ADMIN CONSOLE</v-chip>
                </div>
                <div class="text-body-2 grey--text text--darken-1">Manage all pending, approved, and rejected member registrations</div>
              </div>
            </div>
            <div class="d-flex align-center" style="gap: 12px">
              <v-btn color="primary" outlined small @click="loadData" :loading="loading">
                <v-icon left small>mdi-refresh</v-icon>
                Refresh Data
              </v-btn>
              <v-btn color="error" text small @click="logout">
                <v-icon left small>mdi-logout</v-icon>
                Exit
              </v-btn>
            </div>
          </div>
        </v-card>

        <!-- Section Navigation Tabs -->
        <v-card class="mb-6 pa-2" elevation="0" style="border-radius: 12px; border: 1px solid #E2E8F0; background: white">
          <v-tabs v-model="currentView" color="primary" class="admin-main-tabs" active-class="font-weight-bold">
            <v-tab value="members">
              <v-icon left>mdi-account-group</v-icon>
              Member Registrations
            </v-tab>
            <v-tab value="redis">
              <v-icon left color="red">mdi-database</v-icon>
              Redis Cache Inspector
              <v-chip size="x-small" color="red lighten-5" class="ml-2 font-weight-bold text-caption red--text" label v-if="redisData.connected">
                {{ redisData.key_count }} keys
              </v-chip>
            </v-tab>
          </v-tabs>
        </v-card>

        <!-- VIEW 1: MEMBER REGISTRATIONS -->
        <div v-if="currentView === 0">
          <!-- KPI Metric Cards -->
          <v-row class="mb-6" dense>
            <v-col cols="12" sm="3">
              <v-card class="kpi-card pa-4" elevation="0" @click="statusFilter = 'all'" :class="{ 'active-kpi': statusFilter === 'all' }">
                <div class="d-flex align-center justify-space-between mb-2">
                  <span class="text-caption grey--text text--darken-1 font-weight-bold">TOTAL MEMBERS</span>
                  <v-avatar color="blue lighten-5" size="36">
                    <v-icon color="primary" small>mdi-account-group</v-icon>
                  </v-avatar>
                </div>
                <div class="text-h4 font-weight-black text--primary">{{ allRequests.length }}</div>
                <div class="text-caption grey--text mt-1">All registrations</div>
              </v-card>
            </v-col>
            <v-col cols="12" sm="3">
              <v-card class="kpi-card pa-4" elevation="0" @click="statusFilter = 'pending'" :class="{ 'active-kpi': statusFilter === 'pending' }">
                <div class="d-flex align-center justify-space-between mb-2">
                  <span class="text-caption warning--text text--darken-2 font-weight-bold">PENDING APPROVAL</span>
                  <v-avatar color="amber lighten-5" size="36">
                    <v-icon color="warning" small>mdi-clock-outline</v-icon>
                  </v-avatar>
                </div>
                <div class="text-h4 font-weight-black warning--text text--darken-2">{{ pendingCount }}</div>
                <div class="text-caption grey--text mt-1">Awaiting action</div>
              </v-card>
            </v-col>
            <v-col cols="12" sm="3">
              <v-card class="kpi-card pa-4" elevation="0" @click="statusFilter = 'approved'" :class="{ 'active-kpi': statusFilter === 'approved' }">
                <div class="d-flex align-center justify-space-between mb-2">
                  <span class="text-caption success--text font-weight-bold">APPROVED MEMBERS</span>
                  <v-avatar color="green lighten-5" size="36">
                    <v-icon color="success" small>mdi-check-circle-outline</v-icon>
                  </v-avatar>
                </div>
                <div class="text-h4 font-weight-black success--text">{{ approvedCount }}</div>
                <div class="text-caption grey--text mt-1">Active CRM users</div>
              </v-card>
            </v-col>
            <v-col cols="12" sm="3">
              <v-card class="kpi-card pa-4" elevation="0" @click="statusFilter = 'rejected'" :class="{ 'active-kpi': statusFilter === 'rejected' }">
                <div class="d-flex align-center justify-space-between mb-2">
                  <span class="text-caption error--text font-weight-bold">REJECTED REQUESTS</span>
                  <v-avatar color="red lighten-5" size="36">
                    <v-icon color="error" small>mdi-close-circle-outline</v-icon>
                  </v-avatar>
                </div>
                <div class="text-h4 font-weight-black error--text">{{ rejectedCount }}</div>
                <div class="text-caption grey--text mt-1">Declined access</div>
              </v-card>
            </v-col>
          </v-row>

          <!-- Premium Data Table Container -->
          <v-card class="table-container-card" elevation="0">
            <v-toolbar flat color="transparent" class="px-2 pt-2">
              <v-tabs v-model="activeTab" color="primary" active-class="font-weight-bold">
                <v-tab value="all" @click="statusFilter = 'all'">All Members ({{ allRequests.length }})</v-tab>
                <v-tab value="pending" @click="statusFilter = 'pending'">Pending ({{ pendingCount }})</v-tab>
                <v-tab value="approved" @click="statusFilter = 'approved'">Approved ({{ approvedCount }})</v-tab>
                <v-tab value="rejected" @click="statusFilter = 'rejected'">Rejected ({{ rejectedCount }})</v-tab>
              </v-tabs>
              <v-spacer></v-spacer>
              <v-text-field
                v-model="search"
                prepend-inner-icon="mdi-magnify"
                label="Search member name, email, company..."
                single-line hide-details outlined dense style="max-width: 380px" class="search-input"
              ></v-text-field>
            </v-toolbar>

            <v-divider></v-divider>

            <v-data-table
              :headers="headers"
              :items="filteredRequests"
              :search="search"
              :loading="loading"
              loading-text="Loading member registration records..."
              class="premium-table"
              no-data-text="No member registration records found"
              :items-per-page="10"
            >
              <!-- Member Column -->
              <template v-slot:item.name="{ item }">
                <div class="d-flex align-center py-2">
                  <v-avatar size="38" :color="avatarColor(item.approval_status)" class="mr-3 white--text font-weight-bold">
                    {{ initials(item.name) }}
                  </v-avatar>
                  <div>
                    <div class="font-weight-bold text-subtitle-2 text--primary">{{ item.name }}</div>
                    <div class="caption grey--text">{{ item.email }}</div>
                  </div>
                </div>
              </template>

              <!-- Company Column -->
              <template v-slot:item.company_name="{ item }">
                <div class="d-flex align-center">
                  <v-icon small color="grey" class="mr-1">mdi-office-building</v-icon>
                  <span class="text-body-2 font-weight-medium">{{ item.company_name }}</span>
                </div>
              </template>

              <!-- Transaction ID Column -->
              <template v-slot:item.transaction_id="{ item }">
                <v-chip v-if="item.transaction_id" small outlined color="indigo font-weight-bold">
                  <v-icon left x-small>mdi-receipt</v-icon>
                  {{ item.transaction_id }}
                </v-chip>
                <span v-else class="caption grey--text italic">Not provided</span>
              </template>

              <!-- Status Column -->
              <template v-slot:item.approval_status="{ item }">
                <v-chip small :color="statusChipColor(item.approval_status)" dark class="font-weight-bold px-3">
                  <v-icon left x-small>{{ statusIcon(item.approval_status) }}</v-icon>
                  {{ item.approval_status.toUpperCase() }}
                </v-chip>
              </template>

              <!-- Created At Column -->
              <template v-slot:item.created_at="{ item }">
                <span class="caption grey--text text--darken-2 font-weight-medium">{{ formatDate(item.created_at) }}</span>
              </template>

              <!-- Actions Column -->
              <template v-slot:item.actions="{ item }">
                <div class="d-flex align-center justify-end" style="gap: 8px">
                  <v-btn
                    v-if="item.approval_status === 'pending' || item.approval_status === 'rejected'"
                    color="success" x-small class="px-3 rounded-lg font-weight-bold"
                    :loading="approving === item.id" @click="approve(item.id)"
                  >
                    <v-icon left x-small>mdi-check</v-icon> Approve
                  </v-btn>
                  <v-btn
                    v-if="item.approval_status === 'pending' || item.approval_status === 'approved'"
                    color="error" outlined x-small class="px-3 rounded-lg font-weight-bold"
                    :loading="rejecting === item.id" @click="reject(item.id)"
                  >
                    <v-icon left x-small>mdi-close</v-icon Reject
                  </v-btn>
                </div>
              </template>
            </v-data-table>
          </v-card>
        </div>

        <!-- VIEW 2: REDIS CACHE INSPECTOR -->
        <div v-else-if="currentView === 1">
          <!-- Connection Status Card -->
          <v-card class="mb-6 pa-5 banner-card" elevation="0">
            <div class="d-flex align-center justify-space-between flex-wrap" style="gap: 16px">
              <div class="d-flex align-center">
                <v-avatar :color="redisData.connected ? 'green lighten-5' : 'red lighten-5'" size="48" class="mr-3">
                  <v-icon :color="redisData.connected ? 'success' : 'error'">
                    {{ redisData.connected ? 'mdi-database-check' : 'mdi-database-off' }}
                  </v-icon>
                </v-avatar>
                <div>
                  <div class="d-flex align-center">
                    <span class="text-h6 font-weight-bold text--primary mr-2">Redis In-Memory Database</span>
                    <v-chip :color="redisData.connected ? 'success' : 'error'" small label class="font-weight-bold text-white">
                      {{ redisData.connected ? 'CONNECTED & RUNNING' : 'DISCONNECTED' }}
                    </v-chip>
                  </div>
                  <div class="text-body-2 grey--text">Inspect keys, dynamic cache data, TTL, and values stored inside Redis</div>
                </div>
              </div>

              <div class="d-flex align-center" style="gap: 12px">
                <v-btn color="primary" outlined small @click="fetchRedisData" :loading="loadingRedis">
                  <v-icon left small>mdi-refresh</v-icon>
                  Refresh Cache
                </v-btn>
              </div>
            </div>
          </v-card>

          <!-- Redis Table Card -->
          <v-card class="table-container-card" elevation="0">
            <v-toolbar flat color="transparent" class="px-4 pt-2">
              <span class="text-subtitle-1 font-weight-bold text--primary">
                Stored Keys ({{ redisData.keys ? redisData.keys.length : 0 }})
              </span>
              <v-spacer></v-spacer>
              <v-text-field
                v-model="redisSearch"
                prepend-inner-icon="mdi-magnify"
                label="Search cache key..."
                single-line hide-details outlined dense style="max-width: 320px" class="search-input"
              ></v-text-field>
            </v-toolbar>

            <v-divider></v-divider>

            <v-data-table
              :headers="redisHeaders"
              :items="redisData.keys || []"
              :search="redisSearch"
              :loading="loadingRedis"
              loading-text="Reading Redis cache data..."
              class="premium-table"
              no-data-text="No keys found in Redis cache"
              :items-per-page="10"
            >
              <!-- Key Column -->
              <template v-slot:item.key="{ item }">
                <div class="d-flex align-center font-weight-bold text-subtitle-2 primary--text py-2">
                  <v-icon left small color="primary">mdi-key-variant</v-icon>
                  <code>{{ item.key }}</code>
                </div>
              </template>

              <!-- Data Type Column -->
              <template v-slot:item.type="{ item }">
                <v-chip small color="purple lighten-5" class="purple--text text--darken-2 font-weight-bold" label>
                  {{ item.type.toUpperCase() }}
                </v-chip>
              </template>

              <!-- TTL Column -->
              <template v-slot:item.ttl="{ item }">
                <v-chip small :color="item.ttl > 0 ? 'amber lighten-5' : 'grey lighten-3'" :class="item.ttl > 0 ? 'warning--text font-weight-bold' : 'grey--text'" label>
                  <v-icon left x-small>{{ item.ttl > 0 ? 'mdi-clock-outline' : 'mdi-infinity' }}</v-icon>
                  {{ item.ttl > 0 ? item.ttl + ' seconds' : (item.ttl === -1 ? 'No Expiration' : 'Expired') }}
                </v-chip>
              </template>

              <!-- Value Preview Column -->
              <template v-slot:item.value="{ item }">
                <div class="text-caption text-truncate grey--text text--darken-3" style="max-width: 350px">
                  <code>{{ JSON.stringify(item.value) }}</code>
                </div>
              </template>

              <!-- Actions Column -->
              <template v-slot:item.actions="{ item }">
                <div class="d-flex align-center justify-end" style="gap: 8px">
                  <v-btn color="primary" x-small text class="font-weight-bold" @click="viewKeyDetails(item)">
                    <v-icon left x-small>mdi-eye</v-icon> View Data
                  </v-btn>
                  <v-btn color="error" icon x-small :loading="deletingKey === item.key" @click="deleteKey(item.key)">
                    <v-icon small>mdi-delete</v-icon>
                  </v-btn>
                </div>
              </template>
            </v-data-table>
          </v-card>

          <!-- Key Content Viewer Dialog -->
          <v-dialog v-model="keyDialog" max-width="650px">
            <v-card style="border-radius: 16px" v-if="selectedKey">
              <v-card-title class="d-flex align-center justify-space-between primary white--text pa-4">
                <div class="d-flex align-center">
                  <v-icon color="white" class="mr-2">mdi-code-json</v-icon>
                  <span>Key Value Viewer</span>
                </div>
                <v-btn icon dark small @click="keyDialog = false"><v-icon>mdi-close</v-icon></v-btn>
              </v-card-title>
              <v-card-text class="pa-5">
                <div class="mb-3">
                  <span class="caption grey--text font-weight-bold">KEY NAME:</span>
                  <div class="text-subtitle-1 font-weight-bold primary--text"><code>{{ selectedKey.key }}</code></div>
                </div>
                <div class="d-flex mb-4" style="gap: 16px">
                  <div>
                    <span class="caption grey--text font-weight-bold">TYPE:</span>
                    <div><v-chip x-small color="purple" dark>{{ selectedKey.type }}</v-chip></div>
                  </div>
                  <div>
                    <span class="caption grey--text font-weight-bold">TTL:</span>
                    <div><v-chip x-small color="warning" dark>{{ selectedKey.ttl > 0 ? selectedKey.ttl + 's' : 'No expiry' }}</v-chip></div>
                  </div>
                </div>
                <span class="caption grey--text font-weight-bold mb-1 d-block">STORED DATA (JSON):</span>
                <div class="pa-4 grey lighten-4 rounded-lg" style="max-height: 300px; overflow-y: auto;">
                  <pre class="caption" style="white-space: pre-wrap; font-family: monospace;">{{ JSON.stringify(selectedKey.value, null, 2) }}</pre>
                </div>
              </v-card-text>
              <v-card-actions class="pa-4 bg-grey-lighten-4 d-flex justify-end">
                <v-btn color="grey" text @click="keyDialog = false">Close</v-btn>
              </v-card-actions>
            </v-card>
          </v-dialog>
        </div>
      </div>
    </v-container>
  </div>
</template>

<script>
import registrationService from '../../services/registrationService'

export default {
  name: 'AdminPanelView',
  data() {
    return {
      authenticated: false,
      adminCode: '',
      showCode: false,
      verifying: false,
      codeError: null,
      adminToken: null,
      allRequests: [],
      loading: false,
      approving: null,
      rejecting: null,
      search: '',
      activeTab: 'all',
      statusFilter: 'all',
      currentView: 0,
      redisData: { connected: false, key_count: 0, keys: [] },
      loadingRedis: false,
      redisSearch: '',
      deletingKey: null,
      selectedKey: null,
      keyDialog: false,
      redisHeaders: [
        { text: 'KEY NAME', value: 'key', sortable: true },
        { text: 'DATA TYPE', value: 'type', sortable: true },
        { text: 'TIME TO LIVE (TTL)', value: 'ttl', sortable: true },
        { text: 'STORED VALUE PREVIEW', value: 'value', sortable: false },
        { text: 'ACTIONS', value: 'actions', sortable: false, align: 'end' }
      ],
      headers: [
        { text: 'MEMBER DETAILS', value: 'name', sortable: true },
        { text: 'COMPANY / ORG', value: 'company_name', sortable: true },
        { text: 'TRANSACTION ID', value: 'transaction_id', sortable: true },
        { text: 'REGISTRATION DATE', value: 'created_at', sortable: true },
        { text: 'STATUS', value: 'approval_status', sortable: true, align: 'center' },
        { text: 'ACTIONS', value: 'actions', sortable: false, align: 'end' }
      ]
    }
  },
  watch: {
    currentView(val) {
      if (val === 1) {
        this.fetchRedisData()
      }
    }
  },
  computed: {
    filteredRequests() {
      if (this.statusFilter === 'all') return this.allRequests
      return this.allRequests.filter(r => r.approval_status === this.statusFilter)
    },
    pendingCount() {
      return this.allRequests.filter(r => r.approval_status === 'pending').length
    },
    approvedCount() {
      return this.allRequests.filter(r => r.approval_status === 'approved').length
    },
    rejectedCount() {
      return this.allRequests.filter(r => r.approval_status === 'rejected').length
    }
  },
  methods: {
    async verifyCode() {
      this.verifying = true
      this.codeError = null
      try {
        const res = await registrationService.verifyAdminCode(this.adminCode)
        this.adminToken = res.data.data.token
        this.authenticated = true
        this.loadData()
        this.fetchRedisData()
      } catch (err) {
        this.codeError = err.response?.data?.message || 'Invalid admin code'
      } finally {
        this.verifying = false
      }
    },
    async loadData() {
      this.loading = true
      try {
        const res = await registrationService.getAll(this.adminToken)
        this.allRequests = res.data.data || []
      } catch (e) {
        if (e.response?.status === 401) {
          this.authenticated = false
          this.adminToken = null
        }
      } finally {
        this.loading = false
      }
    },
    async fetchRedisData() {
      this.loadingRedis = true
      try {
        const res = await registrationService.getRedisData(this.adminToken)
        this.redisData = res.data.data || { connected: false, key_count: 0, keys: [] }
      } catch (e) {
        console.error('Failed to fetch Redis data', e)
      } finally {
        this.loadingRedis = false
      }
    },
    viewKeyDetails(item) {
      this.selectedKey = item
      this.keyDialog = true
    },
    async deleteKey(key) {
      if (!confirm(`Are you sure you want to delete key "${key}" from Redis?`)) return
      this.deletingKey = key
      try {
        await registrationService.deleteRedisKey(key, this.adminToken)
        this.fetchRedisData()
      } catch (e) {
        alert(e.response?.data?.message || 'Failed to delete key')
      } finally {
        this.deletingKey = null
      }
    },
    async approve(id) {
      this.approving = id
      try {
        await registrationService.approveRequest(id, this.adminToken)
        const req = this.allRequests.find(r => r.id === id)
        if (req) req.approval_status = 'approved'
      } catch (e) {
        alert(e.response?.data?.message || 'Failed to approve request')
      } finally {
        this.approving = null
      }
    },
    async reject(id) {
      this.rejecting = id
      try {
        await registrationService.rejectRequest(id, this.adminToken)
        const req = this.allRequests.find(r => r.id === id)
        if (req) req.approval_status = 'rejected'
      } catch (e) {
        alert(e.response?.data?.message || 'Failed to reject request')
      } finally {
        this.rejecting = null
      }
    },
    logout() {
      this.authenticated = false
      this.adminToken = null
      this.$router.push('/register')
    },
    initials(name) {
      return name ? name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2) : '?'
    },
    avatarColor(status) {
      if (status === 'approved') return 'success'
      if (status === 'rejected') return 'error'
      return 'warning'
    },
    statusChipColor(status) {
      if (status === 'approved') return '#10B981'
      if (status === 'rejected') return '#EF4444'
      return '#F59E0B'
    },
    statusIcon(status) {
      if (status === 'approved') return 'mdi-check-circle'
      if (status === 'rejected') return 'mdi-close-circle'
      return 'mdi-clock-outline'
    },
    formatDate(d) {
      if (!d) return '-'
      return new Date(d).toLocaleDateString('en-IN', { day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' })
    }
  }
}
</script>

<style scoped>
.admin-bg {
  background: linear-gradient(135deg, #F1F5F9 0%, #E2E8F0 100%);
  min-height: 100vh;
}
.code-card {
  border-radius: 20px !important;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0,0,0,0.12) !important;
}
.code-header {
  background: linear-gradient(135deg, #0F172A, #1E293B);
}
.code-input >>> input {
  font-size: 20px;
  letter-spacing: 8px;
  text-align: center;
  font-weight: 700;
}
.verify-btn {
  border-radius: 12px !important;
  text-transform: none;
  font-weight: 600;
}
.banner-card {
  border-radius: 16px !important;
  background: white;
  border: 1px solid #E2E8F0;
  box-shadow: 0 4px 20px rgba(0,0,0,0.03) !important;
}
.kpi-card {
  border-radius: 14px !important;
  background: white;
  border: 1px solid #E2E8F0;
  cursor: pointer;
  transition: all 0.2s ease-in-out;
}
.kpi-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0,0,0,0.06) !important;
}
.active-kpi {
  border-color: #3B82F6 !important;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2) !important;
}
.table-container-card {
  border-radius: 16px !important;
  background: white;
  border: 1px solid #E2E8F0;
  overflow: hidden;
  box-shadow: 0 8px 30px rgba(0,0,0,0.04) !important;
}
.search-input >>> .v-input__control {
  border-radius: 10px !important;
}
.premium-table >>> th {
  font-weight: 700 !important;
  color: #475569 !important;
  letter-spacing: 0.5px;
  background-color: #F8FAFC !important;
}
.premium-table >>> td {
  border-bottom: 1px solid #F1F5F9 !important;
}
</style>
