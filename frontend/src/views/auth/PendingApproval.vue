<template>
  <div class="pending-bg fill-height d-flex align-center justify-center py-12">
    <v-container fluid style="max-width: 480px">
      <v-card class="pending-card text-center" elevation="0">
        <v-card-text class="pa-8">
          <!-- Status Animation -->
          <div class="status-circle mx-auto mb-6" :class="statusClass">
            <v-icon size="48" :color="statusColor">{{ statusIcon }}</v-icon>
          </div>

          <h2 class="text-h5 font-weight-bold grey--text text--darken-3 mb-2">{{ statusTitle }}</h2>
          <p class="text-body-1 grey--text mb-6">{{ statusMessage }}</p>

          <!-- Progress Stepper -->
          <div class="stepper-container mb-6">
            <div class="stepper-step" :class="{ active: true, completed: true }">
              <div class="stepper-dot"><v-icon x-small color="white">mdi-check</v-icon></div>
              <div class="stepper-label">Registered</div>
            </div>
            <div class="stepper-line completed" />
            <div class="stepper-step" :class="{ active: true, completed: true }">
              <div class="stepper-dot"><v-icon x-small color="white">mdi-check</v-icon></div>
              <div class="stepper-label">Payment Sent</div>
            </div>
            <div class="stepper-line" :class="{ completed: status === 'approved' }" />
            <div class="stepper-step" :class="{ active: status === 'approved', pending: status === 'pending' }">
              <div class="stepper-dot">
                <v-icon x-small :color="status === 'approved' ? 'white' : 'grey'">
                  {{ status === 'approved' ? 'mdi-check' : 'mdi-clock-outline' }}
                </v-icon>
              </div>
              <div class="stepper-label">{{ status === 'approved' ? 'Approved!' : 'Verification' }}</div>
            </div>
          </div>

          <v-alert v-if="status === 'approved'" type="success" text prominent class="mb-4">
            <div class="font-weight-medium">Your account has been activated!</div>
            <div class="text-caption mt-1">You can now sign in and start using Proprietor CRM.</div>
          </v-alert>

          <v-alert v-if="status === 'rejected'" type="error" text prominent class="mb-4">
            <div class="font-weight-medium">Registration was not approved</div>
            <div class="text-caption mt-1">Please contact support or try registering again.</div>
          </v-alert>

          <v-btn
            v-if="status === 'approved'"
            color="primary" large block
            @click="$router.push('/login')"
            class="action-btn mb-3"
          >
            <v-icon left>mdi-login</v-icon>
            Sign In Now
          </v-btn>

          <v-btn
            v-if="status === 'rejected'"
            color="primary" large block outlined
            @click="$router.push('/register')"
            class="action-btn mb-3"
          >
            <v-icon left>mdi-refresh</v-icon>
            Register Again
          </v-btn>

          <div v-if="status === 'pending'" class="mt-2">
            <v-progress-linear indeterminate color="primary" rounded height="3" class="mb-3" />
            <p class="text-caption grey--text">Checking status automatically...</p>
          </div>

          <div class="mt-4">
            <v-btn text small color="grey" @click="$router.push('/register')">
              <v-icon left small>mdi-arrow-left</v-icon>
              Back to Home
            </v-btn>
          </div>
        </v-card-text>
      </v-card>
    </v-container>
  </div>
</template>

<script>
import registrationService from '../../services/registrationService'

export default {
  name: 'PendingApprovalView',
  data() {
    return {
      status: 'pending',
      pollTimer: null
    }
  },
  computed: {
    registrationId() {
      return this.$route.params.id
    },
    statusClass() {
      return {
        'status-pending': this.status === 'pending',
        'status-approved': this.status === 'approved',
        'status-rejected': this.status === 'rejected'
      }
    },
    statusColor() {
      const map = { pending: 'orange', approved: 'success', rejected: 'error' }
      return map[this.status] || 'grey'
    },
    statusIcon() {
      const map = { pending: 'mdi-clock-outline', approved: 'mdi-check-circle', rejected: 'mdi-close-circle' }
      return map[this.status] || 'mdi-help-circle'
    },
    statusTitle() {
      const map = { pending: 'Verification in Progress', approved: 'Account Activated!', rejected: 'Registration Declined' }
      return map[this.status] || 'Processing...'
    },
    statusMessage() {
      const map = {
        pending: 'Your payment is being verified by the admin. This usually takes a few minutes.',
        approved: 'Congratulations! Your account is ready.',
        rejected: 'Unfortunately, your registration was not approved.'
      }
      return map[this.status] || ''
    }
  },
  methods: {
    async checkStatus() {
      if (!this.registrationId) return
      try {
        const res = await registrationService.checkStatus(this.registrationId)
        const data = res.data.data
        this.status = data.approval_status
        if (this.status !== 'pending' && this.pollTimer) {
          clearInterval(this.pollTimer)
        }
      } catch (e) { /* ignore polling errors */ }
    },
    startPolling() {
      this.checkStatus()
      this.pollTimer = setInterval(() => this.checkStatus(), 5000)
    }
  },
  mounted() {
    if (!this.registrationId) {
      this.$router.push('/register')
      return
    }
    this.startPolling()
  },
  beforeDestroy() {
    if (this.pollTimer) clearInterval(this.pollTimer)
  }
}
</script>

<style scoped>
.pending-bg {
  background: linear-gradient(135deg, #F0F4FF 0%, #E8EDF5 100%);
  min-height: 100vh;
}
.pending-card {
  border-radius: 20px !important;
  box-shadow: 0 20px 60px rgba(0,0,0,0.08) !important;
  border: 1px solid rgba(0,0,0,0.04);
}
.status-circle {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.5s ease;
}
.status-pending {
  background: linear-gradient(135deg, #FFF3CD, #FFE69C);
  animation: pulse-pending 2s ease-in-out infinite;
}
.status-approved {
  background: linear-gradient(135deg, #D1FAE5, #A7F3D0);
}
.status-rejected {
  background: linear-gradient(135deg, #FEE2E2, #FECACA);
}
@keyframes pulse-pending {
  0%, 100% { transform: scale(1); box-shadow: 0 0 0 0 rgba(245,158,11,0.3); }
  50% { transform: scale(1.05); box-shadow: 0 0 0 16px rgba(245,158,11,0); }
}

/* Stepper */
.stepper-container {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  gap: 0;
}
.stepper-step {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 80px;
}
.stepper-dot {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #E5E7EB;
  margin-bottom: 6px;
  transition: all 0.3s;
}
.stepper-step.completed .stepper-dot { background: #10B981; }
.stepper-step.active .stepper-dot { background: #10B981; }
.stepper-step.pending .stepper-dot { background: #F59E0B; }
.stepper-label { font-size: 11px; color: #6B7280; font-weight: 500; }
.stepper-line {
  height: 2px;
  flex: 1;
  background: #E5E7EB;
  margin-top: 13px;
  min-width: 30px;
  transition: all 0.3s;
}
.stepper-line.completed { background: #10B981; }
.action-btn {
  border-radius: 12px !important;
  text-transform: none;
  font-weight: 600;
}
</style>
