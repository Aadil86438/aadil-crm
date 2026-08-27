<template>
  <v-app>
    <router-view />
    <!-- Global Snackbar -->
    <v-snackbar
      v-model="snackbar.show"
      :color="snackbar.color"
      :timeout="snackbar.timeout"
      bottom
      right
      multi-line
    >
      <v-icon left small>
        {{ snackbar.color === 'success' ? 'mdi-check-circle' : snackbar.color === 'error' ? 'mdi-alert-circle' : 'mdi-information' }}
      </v-icon>
      {{ snackbar.message }}
      <template v-slot:action="{ attrs }">
        <v-btn text v-bind="attrs" @click="hideSnackbar">Close</v-btn>
      </template>
    </v-snackbar>
  </v-app>
</template>

<script>
import { mapState, mapMutations } from 'vuex'

export default {
  name: 'App',
  computed: {
    ...mapState('snackbar', ['show', 'message', 'color', 'timeout']),
    snackbar() {
      return this.$store.state.snackbar
    }
  },
  methods: {
    ...mapMutations('snackbar', { hideSnackbar: 'HIDE' })
  }
}
</script>

<style>
* { font-family: 'Inter', 'Roboto', sans-serif; }
::-webkit-scrollbar { width: 6px; height: 6px; }
::-webkit-scrollbar-track { background: #f1f1f1; }
::-webkit-scrollbar-thumb { background: #c0c0c0; border-radius: 3px; }
::-webkit-scrollbar-thumb:hover { background: #999; }
.v-application { background-color: #F5F7FA !important; }
</style>
