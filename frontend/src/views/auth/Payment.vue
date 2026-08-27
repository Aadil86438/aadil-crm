<template>
  <div class="payment-bg fill-height d-flex align-center justify-center py-6">
    <v-container fluid style="max-width: 520px">
      <!-- Invoice Card -->
      <v-card class="invoice-card" elevation="0">
        <!-- Invoice Header -->
        <div class="invoice-header pa-6 pb-4">
          <div class="d-flex align-center justify-space-between mb-3">
            <div class="d-flex align-center">
              <v-icon color="white" size="28" class="mr-2">mdi-home-city</v-icon>
              <div>
                <div class="white--text text-h6 font-weight-bold" style="line-height: 1.2">Propertier</div>
                <div class="white--text text-caption" style="opacity: 0.7">by MOHAMMED AADIL</div>
              </div>
            </div>
            <div class="text-right">
              <div class="white--text text-overline" style="opacity: 0.7">INVOICE</div>
              <div class="white--text text-caption" style="opacity: 0.6">#INV-{{ invoiceNumber }}</div>
            </div>
          </div>
          <v-divider dark style="opacity: 0.2" />
          <div class="mt-3 d-flex justify-space-between">
            <div>
              <div class="white--text text-caption" style="opacity: 0.6">Bill To</div>
              <div class="white--text text-body-2 font-weight-medium">{{ customerName }}</div>
              <div class="white--text text-caption" style="opacity: 0.6">{{ companyName }}</div>
            </div>
            <div class="text-right">
              <div class="white--text text-caption" style="opacity: 0.6">Date</div>
              <div class="white--text text-body-2">{{ formattedDate }}</div>
            </div>
          </div>
        </div>

        <!-- Bill Items -->
        <v-card-text class="pa-6 pt-5">
          <div class="text-overline font-weight-bold grey--text text--darken-1 mb-3">PLAN DETAILS</div>

          <div class="bill-item pa-3 mb-2 rounded" v-for="item in billItems" :key="item.name">
            <div class="d-flex align-center justify-space-between">
              <div class="d-flex align-center">
                <v-icon small color="primary" class="mr-2">{{ item.icon }}</v-icon>
                <div>
                  <div class="text-body-2 font-weight-medium">{{ item.name }}</div>
                  <div class="text-caption grey--text">{{ item.desc }}</div>
                </div>
              </div>
              <div class="text-body-2 font-weight-medium">{{ item.included ? '✓' : item.price }}</div>
            </div>
          </div>

          <v-divider class="my-4" />

          <!-- Pricing Summary -->
          <div class="pricing-summary">
            <div class="d-flex justify-space-between mb-1">
              <span class="text-body-2 grey--text">Subtotal</span>
              <span class="text-body-2">₹599.00</span>
            </div>
            <div class="d-flex justify-space-between mb-1">
              <span class="text-body-2 grey--text">Launch Discount</span>
              <span class="text-body-2 success--text">- ₹100.00</span>
            </div>
            <div class="d-flex justify-space-between mb-1">
              <span class="text-body-2 grey--text">GST (Inclusive)</span>
              <span class="text-body-2">₹0.00</span>
            </div>
            <v-divider class="my-3" />
            <div class="d-flex justify-space-between align-center">
              <span class="text-subtitle-1 font-weight-bold">Total Amount</span>
              <span class="text-h5 font-weight-black primary--text">₹499</span>
            </div>
          </div>

          <v-divider class="my-4" />

          <!-- Payment Method -->
          <div class="text-overline font-weight-bold grey--text text--darken-1 mb-3">PAYMENT METHOD — UPI</div>

          <!-- QR Code -->
          <div class="qr-section text-center pa-4 rounded mb-4">
            <div class="text-body-2 font-weight-medium mb-2 grey--text text--darken-2">Scan & Pay using any UPI app</div>
            <div class="qr-wrapper mx-auto mb-3">
              <img src="/img/payment-qr.png" alt="Payment QR Code" class="qr-image" />
            </div>
            <div class="text-body-2 font-weight-bold mb-1">A MOHAMMED AADIL</div>
            <div class="text-caption grey--text">UPI: 8643839796@yapl</div>
            <div class="d-flex justify-center mt-3" style="gap: 8px">
              <v-chip x-small outlined>GPay</v-chip>
              <v-chip x-small outlined>PhonePe</v-chip>
              <v-chip x-small outlined>Paytm</v-chip>
              <v-chip x-small outlined>BHIM</v-chip>
            </div>
          </div>

          <!-- Transaction ID Input -->
          <v-text-field
            v-model="transactionId"
            label="Transaction / UTR ID"
            placeholder="Enter your UPI transaction ID"
            prepend-inner-icon="mdi-receipt-text-outline"
            outlined dense color="primary"
            :rules="[v => !!v || 'Transaction ID is required']"
            class="mb-2" id="transaction-id"
          />

          <v-checkbox
            v-model="confirmed"
            label="I confirm that I have made the payment of ₹499"
            color="primary" dense hide-details
            class="mb-4 mt-0" id="confirm-payment"
          />

          <v-alert v-if="error" type="error" dense text class="mb-3">{{ error }}</v-alert>
          <v-alert v-if="success" type="success" dense text class="mb-3">{{ success }}</v-alert>

          <v-btn
            color="primary" block large
            :loading="loading"
            :disabled="!transactionId || !confirmed || loading"
            @click="submitPayment"
            class="submit-btn" id="submit-payment"
          >
            <v-icon left>mdi-check-circle</v-icon>
            I Have Paid — Submit for Verification
          </v-btn>

          <div class="text-center mt-4">
            <v-btn text small color="grey" @click="$router.push('/register')">
              <v-icon left small>mdi-arrow-left</v-icon>
              Back to Registration
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
  name: 'PaymentView',
  data() {
    return {
      transactionId: '',
      confirmed: false,
      loading: false,
      error: null,
      success: null,
      billItems: [
        { name: 'CRM Pro License', desc: 'Full access — Leads, Deals, Contacts, Accounts', icon: 'mdi-briefcase-check', included: true },
        { name: 'Task & Activity Manager', desc: 'Track tasks, calls, meetings & emails', icon: 'mdi-checkbox-marked-circle', included: true },
        { name: 'Deal Pipeline (Kanban)', desc: 'Visual sales pipeline with drag & drop', icon: 'mdi-view-column', included: true },
        { name: 'Reports & Analytics', desc: 'Sales reports, lead analytics, forecasting', icon: 'mdi-chart-bar', included: true },
        { name: 'Calendar & Scheduling', desc: 'Integrated calendar for team activities', icon: 'mdi-calendar-check', included: true },
        { name: 'Unlimited Users', desc: 'Add your entire team — no extra cost', icon: 'mdi-account-group', included: true },
      ]
    }
  },
  computed: {
    registrationId() {
      return this.$route.params.id
    },
    customerName() {
      return this.$route.query.name || 'Customer'
    },
    companyName() {
      return this.$route.query.company || 'Organization'
    },
    invoiceNumber() {
      const id = this.registrationId || ''
      return id.substring(0, 8).toUpperCase()
    },
    formattedDate() {
      return new Date().toLocaleDateString('en-IN', { day: '2-digit', month: 'short', year: 'numeric' })
    }
  },
  methods: {
    async submitPayment() {
      if (!this.transactionId || !this.confirmed) return
      this.loading = true
      this.error = null
      this.success = null
      try {
        await registrationService.submitPayment({
          registration_id: this.registrationId,
          transaction_id: this.transactionId
        })
        this.success = 'Payment submitted! Redirecting...'
        setTimeout(() => {
          this.$router.push({ name: 'PendingApproval', params: { id: this.registrationId } })
        }, 1500)
      } catch (err) {
        this.error = err.response?.data?.message || 'Failed to submit payment. Please try again.'
      } finally {
        this.loading = false
      }
    }
  },
  mounted() {
    if (!this.registrationId) {
      this.$router.push('/register')
    }
  }
}
</script>

<style scoped>
.payment-bg {
  background: linear-gradient(135deg, #F5F7FA 0%, #E8EDF2 100%);
  min-height: 100vh;
}
.invoice-card {
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0,0,0,0.12) !important;
  border: 1px solid rgba(0,0,0,0.05);
}
.invoice-header {
  background: linear-gradient(135deg, #1565C0, #0D47A1);
}
.bill-item {
  background: #F8FAFC;
  border: 1px solid #E2E8F0;
  transition: all 0.2s;
}
.bill-item:hover {
  background: #EFF6FF;
  border-color: #BFDBFE;
}
.pricing-summary {
  background: #FAFBFC;
  border-radius: 12px;
  padding: 16px;
  border: 1px solid #E2E8F0;
}
.qr-section {
  background: linear-gradient(135deg, #FFF9E6, #FFF3CD);
  border: 2px dashed #F59E0B;
}
.qr-wrapper {
  max-width: 320px;
  width: 100%;
  background: white;
  border-radius: 12px;
  padding: 8px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}
.qr-image {
  width: 100%;
  height: auto;
  display: block;
  border-radius: 8px;
}
.submit-btn {
  font-weight: 600;
  letter-spacing: 0.3px;
  border-radius: 10px !important;
  text-transform: none;
}
</style>
