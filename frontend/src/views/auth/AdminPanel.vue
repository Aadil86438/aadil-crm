<template>
  <div class="admin-bg fill-height d-flex align-center justify-center py-8">
    <v-container fluid style="max-width: 700px">
      <!-- Code Entry (not authenticated) -->
      <v-card v-if="!authenticated" class="code-card" elevation="0">
        <div class="code-header pa-6 text-center">
          <v-icon size="44" color="white" class="mb-3">mdi-shield-lock</v-icon>
          <h2 class="white--text text-h5 font-weight-bold">Admin Access</h2>
          <p class="white--text text-body-2 mt-1" style="opacity: 0.7">Enter the admin code to manage registrations</p>
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
              Verify & Access
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
        <div class="d-flex align-center justify-space-between mb-5">
          <div>
            <div class="d-flex align-center mb-1">
              <v-icon color="primary" class="mr-2">mdi-home-city</v-icon>
              <span class="text-h6 font-weight-bold primary--text">Propertier</span>
              <v-chip x-small color="primary" dark class="ml-2">ADMIN</v-chip>
            </div>
            <p class="text-body-2 grey--text mb-0">Manage registration requests</p>
          </div>
          <v-btn outlined color="grey" small @click="logout">
            <v-icon left small>mdi-logout</v-icon>
            Exit
          </v-btn>
        </div>

        <!-- Stats -->
        <v-row dense class="mb-4">
          <v-col cols="4">
            <v-card class="stat-card pa-4 text-center" elevation="0">
              <div class="text-h4 font-weight-black primary--text">{{ pendingRequests.length }}</div>
              <div class="text-caption grey--text font-weight-medium">PENDING</div>
            </v-card>
          </v-col>
          <v-col cols="4">
            <v-card class="stat-card pa-4 text-center" elevation="0">
              <div class="text-h4 font-weight-black success--text">{{ approvedCount }}</div>
              <div class="text-caption grey--text font-weight-medium">APPROVED</div>
            </v-card>
          </v-col>
          <v-col cols="4">
            <v-card class="stat-card pa-4 text-center" elevation="0">
              <div class="text-h4 font-weight-black error--text">{{ rejectedCount }}</div>
              <div class="text-caption grey--text font-weight-medium">REJECTED</div>
            </v-card>
          </v-col>
        </v-row>

        <!-- Requests List -->
        <v-card v-if="pendingRequests.length === 0" class="empty-card pa-8 text-center" elevation="0">
          <v-icon size="60" color="grey lighten-1" class="mb-3">mdi-inbox-outline</v-icon>
          <h3 class="text-h6 grey--text text--lighten-1">No Pending Requests</h3>
          <p class="text-body-2 grey--text">All registration requests have been processed.</p>
          <v-btn outlined color="primary" small @click="loadPending" class="mt-2">
            <v-icon left small>mdi-refresh</v-icon>
            Refresh
          </v-btn>
        </v-card>

        <v-card
          v-for="req in pendingRequests"
          :key="req.id"
          class="request-card mb-3 pa-4"
          elevation="0"
        >
          <div class="d-flex align-start">
            <v-avatar size="44" color="primary" class="mr-4">
              <span class="white--text text-body-1 font-weight-bold">{{ initials(req.name) }}</span>
            </v-avatar>
            <div class="flex-grow-1">
              <div class="d-flex align-center flex-wrap mb-1">
                <span class="text-subtitle-1 font-weight-bold mr-2">{{ req.name }}</span>
                <v-chip x-small outlined color="orange" class="mr-2">PENDING</v-chip>
              </div>
              <div class="text-body-2 grey--text mb-1">{{ req.email }}</div>
              <div class="d-flex flex-wrap" style="gap: 12px">
                <div class="d-flex align-center">
                  <v-icon x-small color="grey" class="mr-1">mdi-office-building</v-icon>
                  <span class="text-caption grey--text text--darken-1">{{ req.company_name }}</span>
                </div>
                <div class="d-flex align-center">
                  <v-icon x-small color="grey" class="mr-1">mdi-receipt-text</v-icon>
                  <span class="text-caption grey--text text--darken-1">Txn: <strong>{{ req.transaction_id }}</strong></span>
                </div>
                <div class="d-flex align-center">
                  <v-icon x-small color="grey" class="mr-1">mdi-calendar</v-icon>
                  <span class="text-caption grey--text text--darken-1">{{ formatDate(req.created_at) }}</span>
                </div>
              </div>
            </div>
            <div class="d-flex align-center ml-3" style="gap: 8px">
              <v-btn
                color="success" small
                :loading="approving === req.id"
                @click="approve(req.id)"
                class="action-btn"
              >
                <v-icon left small>mdi-check</v-icon>
                Approve
              </v-btn>
              <v-btn
                color="error" small outlined
                :loading="rejecting === req.id"
                @click="reject(req.id)"
                class="action-btn"
              >
                <v-icon left small>mdi-close</v-icon>
                Reject
              </v-btn>
            </div>
          </div>
        </v-card>
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
      pendingRequests: [],
      approvedCount: 0,
      rejectedCount: 0,
      loading: false,
      approving: null,
      rejecting: null
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
        this.loadPending()
      } catch (err) {
        this.codeError = err.response?.data?.message || 'Invalid admin code'
      } finally {
        this.verifying = false
      }
    },
    async loadPending() {
      this.loading = true
      try {
        const res = await registrationService.getPending(this.adminToken)
        this.pendingRequests = res.data.data || []
      } catch (e) {
        if (e.response?.status === 401) {
          this.authenticated = false
          this.adminToken = null
        }
      } finally {
        this.loading = false
      }
    },
    async approve(id) {
      this.approving = id
      try {
        await registrationService.approveRequest(id, this.adminToken)
        this.approvedCount++
        this.pendingRequests = this.pendingRequests.filter(r => r.id !== id)
      } catch (e) {
        alert(e.response?.data?.message || 'Failed to approve')
      } finally {
        this.approving = null
      }
    },
    async reject(id) {
      this.rejecting = id
      try {
        await registrationService.rejectRequest(id, this.adminToken)
        this.rejectedCount++
        this.pendingRequests = this.pendingRequests.filter(r => r.id !== id)
      } catch (e) {
        alert(e.response?.data?.message || 'Failed to reject')
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
    formatDate(d) {
      return new Date(d).toLocaleDateString('en-IN', { day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' })
    }
  }
}
</script>

<style scoped>
.admin-bg {
  background: linear-gradient(135deg, #F0F4FF 0%, #E8EDF5 100%);
  min-height: 100vh;
}
.code-card {
  border-radius: 20px !important;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0,0,0,0.1) !important;
}
.code-header {
  background: linear-gradient(135deg, #1E293B, #334155);
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
.stat-card {
  border-radius: 14px !important;
  border: 1px solid #E2E8F0;
  background: white;
}
.empty-card {
  border-radius: 16px !important;
  border: 2px dashed #E2E8F0;
  background: white;
}
.request-card {
  border-radius: 14px !important;
  border: 1px solid #E2E8F0;
  background: white;
  transition: all 0.2s;
}
.request-card:hover {
  border-color: #BFDBFE;
  box-shadow: 0 4px 16px rgba(0,0,0,0.06) !important;
}
.action-btn {
  border-radius: 8px !important;
  text-transform: none;
  font-weight: 600;
}
</style>
