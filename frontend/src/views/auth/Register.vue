<template>
  <div class="auth-bg fill-height">
    <v-container fluid class="fill-height pa-0">
      <v-row no-gutters class="fill-height">
        <!-- Left: Hero Section -->
        <v-col cols="12" md="6" class="d-none d-md-flex hero-section align-center justify-center">
          <div class="hero-content text-center px-8">
            <div class="hero-logo mb-6">
              <v-icon size="56" color="white">mdi-home-city</v-icon>
            </div>
            <h1 class="white--text text-h3 font-weight-black mb-3" style="letter-spacing: -0.5px">Proprietor</h1>
            <p class="white--text text-subtitle-1 mb-1" style="opacity: 0.8">by <strong>MOHAMMED AADIL</strong></p>
            <v-divider dark class="my-5 mx-auto" style="max-width: 60px; opacity: 0.4" />
            <p class="white--text text-body-1" style="opacity: 0.7; max-width: 400px; margin: 0 auto; line-height: 1.7">
              Premium CRM solution for modern businesses. Manage leads, deals, contacts & more — all in one place.
            </p>
            <div class="mt-8 d-flex justify-center flex-wrap" style="gap: 20px">
              <div v-for="stat in heroStats" :key="stat.label" class="text-center">
                <div class="white--text text-h5 font-weight-bold">{{ stat.value }}</div>
                <div class="white--text text-caption" style="opacity: 0.6">{{ stat.label }}</div>
              </div>
            </div>
          </div>
        </v-col>

        <!-- Right: Registration Form -->
        <v-col cols="12" md="6" class="d-flex align-center justify-center form-section">
          <div class="form-container">
            <!-- Mobile Logo -->
            <div class="d-flex d-md-none align-center justify-center mb-6">
              <v-icon color="primary" size="32" class="mr-2">mdi-home-city</v-icon>
              <span class="text-h5 font-weight-black primary--text">Proprietor</span>
            </div>

            <div class="mb-6">
              <h2 class="text-h5 font-weight-bold grey--text text--darken-3">Create Account</h2>
              <p class="text-body-2 grey--text mt-1 mb-0">Get started with your CRM in seconds</p>
            </div>

            <v-form ref="form" @submit.prevent="handleRegister" v-model="valid">
              <v-text-field
                v-model="form.name"
                label="Full Name"
                prepend-inner-icon="mdi-account-outline"
                :rules="[v => !!v || 'Full name is required']"
                required outlined dense color="primary"
                class="mb-1" autofocus id="register-name"
              />
              <v-text-field
                v-model="form.email"
                label="Email Address"
                type="email"
                prepend-inner-icon="mdi-email-outline"
                :rules="emailRules"
                required outlined dense color="primary"
                class="mb-1" id="register-email"
              />
              <v-text-field
                v-model="form.company_name"
                label="Company / Organization"
                prepend-inner-icon="mdi-office-building-outline"
                :rules="[v => !!v || 'Company name is required']"
                required outlined dense color="primary"
                class="mb-1" id="register-company"
              />
              <v-text-field
                v-model="form.password"
                label="Password"
                :type="showPass ? 'text' : 'password'"
                prepend-inner-icon="mdi-lock-outline"
                :append-icon="showPass ? 'mdi-eye' : 'mdi-eye-off'"
                @click:append="showPass = !showPass"
                :rules="passwordRules"
                required outlined dense color="primary"
                class="mb-1" id="register-password"
              />
              <v-text-field
                v-model="confirmPassword"
                label="Confirm Password"
                :type="showConfirm ? 'text' : 'password'"
                prepend-inner-icon="mdi-lock-check-outline"
                :append-icon="showConfirm ? 'mdi-eye' : 'mdi-eye-off'"
                @click:append="showConfirm = !showConfirm"
                :rules="confirmRules"
                required outlined dense color="primary"
                class="mb-3" id="register-confirm"
              />

              <v-alert v-if="error" type="error" dense text class="mb-3">{{ error }}</v-alert>

              <v-btn
                type="submit" color="primary" block large
                :loading="loading" :disabled="!valid || loading"
                class="register-btn mb-4" id="register-submit"
              >
                <v-icon left>mdi-account-plus</v-icon>
                Create Account
              </v-btn>
            </v-form>

            <div class="text-center">
              <p class="text-body-2 grey--text">
                Already have an account?
                <router-link to="/login" class="font-weight-medium primary--text text-decoration-none">Sign In</router-link>
              </p>
            </div>

            <!-- Admin Access -->
            <v-divider class="my-4" />
            <div class="text-center">
              <v-btn text small color="grey" @click="$router.push('/admin-panel')" id="admin-access-btn">
                <v-icon left small>mdi-shield-lock</v-icon>
                Admin Access
              </v-btn>
            </div>
          </div>
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<script>
import registrationService from '../../services/registrationService'

export default {
  name: 'RegisterView',
  data() {
    return {
      valid: false,
      form: { name: '', email: '', password: '', company_name: '' },
      confirmPassword: '',
      showPass: false,
      showConfirm: false,
      loading: false,
      error: null,
      emailRules: [
        v => !!v || 'Email is required',
        v => /.+@.+\..+/.test(v) || 'Enter a valid email'
      ],
      passwordRules: [
        v => !!v || 'Password is required',
        v => (v && v.length >= 8) || 'Minimum 8 characters'
      ],
      heroStats: [
        { value: '10K+', label: 'Users' },
        { value: '99.9%', label: 'Uptime' },
        { value: '24/7', label: 'Support' }
      ]
    }
  },
  computed: {
    confirmRules() {
      return [
        v => !!v || 'Confirm your password',
        v => v === this.form.password || 'Passwords do not match'
      ]
    }
  },
  methods: {
    async handleRegister() {
      if (!this.$refs.form.validate()) return
      this.loading = true
      this.error = null
      try {
        const res = await registrationService.register(this.form)
        const regId = res.data.data.id
        this.$router.push({ name: 'Payment', params: { id: regId }, query: { name: this.form.name, company: this.form.company_name } })
      } catch (err) {
        this.error = err.response?.data?.message || 'Registration failed. Please try again.'
      } finally {
        this.loading = false
      }
    }
  }
}
</script>

<style scoped>
.auth-bg {
  background: #F5F7FA;
  min-height: 100vh;
}
.hero-section {
  background: linear-gradient(135deg, #0D47A1 0%, #1565C0 30%, #1976D2 60%, #0288D1 100%);
  position: relative;
  overflow: hidden;
}
.hero-section::before {
  content: '';
  position: absolute;
  top: -50%;
  right: -50%;
  width: 100%;
  height: 100%;
  background: radial-gradient(circle, rgba(255,255,255,0.08) 0%, transparent 70%);
}
.hero-logo {
  width: 100px;
  height: 100px;
  background: rgba(255,255,255,0.15);
  border-radius: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto;
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255,255,255,0.2);
}
.hero-content { position: relative; z-index: 1; }
.form-section { background: white; }
.form-container {
  width: 100%;
  max-width: 420px;
  padding: 40px 32px;
}
.register-btn {
  font-weight: 600;
  letter-spacing: 0.5px;
  border-radius: 10px !important;
  text-transform: none;
  font-size: 15px;
}
@media (max-width: 959px) {
  .form-container { padding: 24px 20px; }
}
</style>
