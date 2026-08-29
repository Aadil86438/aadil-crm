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
                <div class="white--text text-h6 font-weight-bold" style="line-height: 1.2">Proprietor</div>
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

          <!-- Payment Methods -->
          <div class="text-overline font-weight-bold grey--text text--darken-1 mb-3">PAYMENT METHOD</div>

          <!-- Razorpay Secure Payment -->
          <div class="razorpay-section text-center pa-4 rounded mb-4">
            <v-icon color="primary" size="40" class="mb-2">mdi-shield-check</v-icon>
            <div class="text-body-1 font-weight-bold mb-1">Secure Razorpay Checkout</div>
            <div class="text-caption grey--text mb-3">Pay securely via UPI, Google Pay, PhonePe, Credit/Debit Card, or Net Banking</div>
            <div class="d-flex justify-center" style="gap: 8px">
              <v-chip x-small outlined color="primary">Google Pay</v-chip>
              <v-chip x-small outlined color="primary">PhonePe</v-chip>
              <v-chip x-small outlined color="primary">UPI</v-chip>
              <v-chip x-small outlined color="primary">Cards</v-chip>
              <v-chip x-small outlined color="primary">NetBanking</v-chip>
            </div>
          </div>

          <v-alert v-if="error" type="error" dense text class="mb-3">{{ error }}</v-alert>
          <v-alert v-if="success" type="success" dense text class="mb-3">{{ success }}</v-alert>

          <v-btn
            color="primary" block large
            :loading="loading"
            :disabled="loading"
            @click="initiatePayment"
            class="submit-btn" id="pay-razorpay"
          >
            <v-icon left>mdi-flash</v-icon>
            Pay ₹499 via Razorpay
          </v-btn>

          <div class="text-center mt-3">
            <div class="d-flex align-center justify-center" style="gap: 6px">
              <v-icon x-small color="success">mdi-lock</v-icon>
              <span class="text-caption grey--text">256-bit SSL Encrypted • PCI DSS Compliant</span>
            </div>
          </div>

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
    async initiatePayment() {
      this.loading = true
      this.error = null
      this.success = null

      try {
        // Step 1: Create Razorpay Order on backend
        const { data } = await registrationService.createPaymentOrder({
          registration_id: this.registrationId
        })

        const orderData = data.data

        // Step 2: Open Razorpay Checkout popup
        const options = {
          key: orderData.razorpay_key_id,
          amount: orderData.amount,
          currency: orderData.currency,
          name: 'Proprietor by MOHAMMED AADIL',
          description: 'CRM Pro License — ₹499',
          order_id: orderData.order_id,
          prefill: {
            name: orderData.prefill?.name || this.customerName,
            email: orderData.prefill?.email || '',
          },
          theme: {
            color: '#1565C0'
          },
          handler: async (response) => {
            // Step 3: Verify payment on backend
            await this.verifyPayment(response)
          },
          modal: {
            ondismiss: () => {
              this.loading = false
              this.error = 'Payment was cancelled. Please try again.'
            }
          }
        }

        // eslint-disable-next-line no-undef
        const rzp = new Razorpay(options)
        rzp.on('payment.failed', (response) => {
          this.loading = false
          this.error = response.error?.description || 'Payment failed. Please try again.'
        })
        rzp.open()

      } catch (err) {
        this.loading = false
        this.error = err.response?.data?.message || 'Failed to initiate payment. Please try again.'
      }
    },

    async verifyPayment(razorpayResponse) {
      try {
        await registrationService.verifyPayment({
          registration_id: this.registrationId,
          razorpay_order_id: razorpayResponse.razorpay_order_id,
          razorpay_payment_id: razorpayResponse.razorpay_payment_id,
          razorpay_signature: razorpayResponse.razorpay_signature
        })

        this.success = 'Payment verified successfully! Redirecting...'
        this.loading = false

        setTimeout(() => {
          this.$router.push({ name: 'PendingApproval', params: { id: this.registrationId } })
        }, 1500)

      } catch (err) {
        this.loading = false
        this.error = err.response?.data?.message || 'Payment verification failed. Contact support.'
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
.razorpay-section {
  background: linear-gradient(135deg, #EFF6FF, #DBEAFE);
  border: 2px solid #93C5FD;
  border-radius: 12px;
}
.submit-btn {
  font-weight: 600;
  letter-spacing: 0.3px;
  border-radius: 10px !important;
  text-transform: none;
  font-size: 16px !important;
}
</style>
