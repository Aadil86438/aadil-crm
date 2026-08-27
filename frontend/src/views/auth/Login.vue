<template>
  <div class="auth-bg fill-height d-flex align-center justify-center py-12">
    <v-container fluid>
        <v-row align="center" justify="center">
          <v-col cols="12" sm="8" md="5" lg="4">
            <v-card class="login-card" elevation="0">
              <!-- Logo & Title -->
              <div class="login-header pa-8 pb-6">
                <div class="logo-wrapper mb-4">
                  <v-icon size="40" color="white">mdi-handshake</v-icon>
                </div>
                <h1 class="white--text text-h5 font-weight-bold">Welcome Back</h1>
                <p class="white--text text-body-2 mt-1 mb-0 opacity-75">Sign in to your CRM account</p>
              </div>

              <v-card-text class="pa-8 pt-6">
                <v-form ref="form" @submit.prevent="handleLogin" v-model="valid">
                  <v-text-field
                    v-model="email"
                    label="Email Address"
                    type="email"
                    prepend-inner-icon="mdi-email-outline"
                    :rules="emailRules"
                    required
                    outlined
                    dense
                    color="primary"
                    background-color="white"
                    class="mb-2"
                    autofocus
                    id="login-email"
                  />
                  <v-text-field
                    v-model="password"
                    label="Password"
                    :type="showPassword ? 'text' : 'password'"
                    prepend-inner-icon="mdi-lock-outline"
                    :append-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
                    @click:append="showPassword = !showPassword"
                    :rules="passwordRules"
                    required
                    outlined
                    dense
                    color="primary"
                    background-color="white"
                    class="mb-4"
                    id="login-password"
                  />

                  <v-alert v-if="error" type="error" dense text class="mb-4">
                    {{ error }}
                  </v-alert>

                  <v-btn
                    type="submit"
                    color="primary"
                    block
                    large
                    :loading="loading"
                    :disabled="!valid || loading"
                    class="login-btn"
                    id="login-submit"
                  >
                    <v-icon left>mdi-login</v-icon>
                    Sign In
                  </v-btn>
                </v-form>

                <div class="mt-6 text-center">
                  <p class="text-caption grey--text">Demo Credentials:</p>
                  <div class="credentials-grid">
                    <v-chip small outlined class="ma-1" @click="fillCredentials('admin@crm.local', 'Admin@123')">
                      <v-icon left x-small>mdi-shield-crown</v-icon>Admin
                    </v-chip>
                    <v-chip small outlined class="ma-1" @click="fillCredentials('manager@crm.local', 'Manager@123')">
                      <v-icon left x-small>mdi-account-tie</v-icon>Manager
                    </v-chip>
                    <v-chip small outlined class="ma-1" @click="fillCredentials('sales@crm.local', 'Sales@123')">
                      <v-icon left x-small>mdi-account</v-icon>Sales
                    </v-chip>
                  </div>
                </div>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </v-main>
  </v-app>
</template>

<script>
import { mapActions } from 'vuex'

export default {
  name: 'LoginView',
  data() {
    return {
      valid: false,
      email: '',
      password: '',
      showPassword: false,
      loading: false,
      error: null,
      emailRules: [
        v => !!v || 'Email is required',
        v => /.+@.+\..+/.test(v) || 'Enter a valid email'
      ],
      passwordRules: [
        v => !!v || 'Password is required'
      ]
    }
  },
  methods: {
    ...mapActions('auth', ['login']),
    async handleLogin() {
      if (!this.$refs.form.validate()) return
      this.loading = true
      this.error = null
      try {
        await this.login({ email: this.email, password: this.password })
        this.$router.push('/')
      } catch (err) {
        this.error = err.response?.data?.message || 'Invalid email or password'
      } finally {
        this.loading = false
      }
    },
    fillCredentials(email, password) {
      this.email = email
      this.password = password
    }
  }
}
</script>

<style scoped>
.auth-bg {
  background: linear-gradient(135deg, #0D47A1 0%, #1565C0 40%, #1976D2 70%, #0288D1 100%);
  min-height: 100vh;
}
.login-card {
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 25px 60px rgba(0,0,0,0.3) !important;
}
.login-header {
  background: linear-gradient(135deg, #1565C0, #0288D1);
  text-align: center;
}
.logo-wrapper {
  width: 70px;
  height: 70px;
  background: rgba(255,255,255,0.2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto;
  backdrop-filter: blur(10px);
}
.login-btn {
  font-weight: 600;
  letter-spacing: 0.5px;
  border-radius: 8px !important;
}
.opacity-75 { opacity: 0.75; }
.credentials-grid { display: flex; flex-wrap: wrap; justify-content: center; }
</style>
