<template>
  <div class="pa-4 pa-md-6">
    <div class="d-flex flex-wrap align-center mb-4 gap-2">
      <div>
        <h1 class="text-h5 font-weight-bold">Pipeline Kanban</h1>
        <p class="text-body-2 grey--text mb-0">Drag & drop or move deals across stages</p>
      </div>
      <v-spacer />
      <v-btn outlined color="primary" to="/deals" class="mr-2">
        <v-icon left small>mdi-format-list-bulleted</v-icon>List View
      </v-btn>
      <v-btn color="primary" to="/deals/new">
        <v-icon left>mdi-plus</v-icon>New Deal
      </v-btn>
    </div>

    <!-- Pipeline Summary Header -->
    <v-row dense class="mb-4" v-if="summary">
      <v-col cols="6" sm="3">
        <v-card outlined dense class="pa-2 text-center">
          <div class="text-caption grey--text">Total Deals</div>
          <div class="text-h6 font-weight-bold">{{ summary.total_deals }}</div>
        </v-card>
      </v-col>
      <v-col cols="6" sm="3">
        <v-card outlined dense class="pa-2 text-center">
          <div class="text-caption grey--text">Pipeline Value</div>
          <div class="text-h6 font-weight-bold primary--text">{{ formatCurrency(summary.total_value) }}</div>
        </v-card>
      </v-col>
      <v-col cols="6" sm="3">
        <v-card outlined dense class="pa-2 text-center">
          <div class="text-caption grey--text">Won Value</div>
          <div class="text-h6 font-weight-bold success--text">{{ formatCurrency(summary.won_revenue) }}</div>
        </v-card>
      </v-col>
      <v-col cols="6" sm="3">
        <v-card outlined dense class="pa-2 text-center">
          <div class="text-caption grey--text">Avg Deal Size</div>
          <div class="text-h6 font-weight-bold">{{ formatCurrency(summary.avg_deal_size) }}</div>
        </v-card>
      </v-col>
    </v-row>

    <!-- Loading State -->
    <div v-if="loading" class="text-center py-16">
      <v-progress-circular indeterminate color="primary" size="56" />
    </div>

    <!-- Kanban Board -->
    <div v-else class="kanban-board">
      <div v-for="stage in stages" :key="stage" class="kanban-column">
        <div class="column-header d-flex align-center justify-space-between pa-3" :style="{ borderTopColor: stageColors[stage] }">
          <div>
            <span class="text-subtitle-2 font-weight-bold">{{ stage }}</span>
            <span class="text-caption grey--text ml-2">({{ getStageDeals(stage).length }})</span>
          </div>
          <div class="text-caption font-weight-bold">{{ formatCurrency(getStageTotal(stage)) }}</div>
        </div>

        <div class="column-body pa-2">
          <v-card
            v-for="deal in getStageDeals(stage)"
            :key="deal.id"
            class="kanban-card mb-3 pa-3"
            elevation="1"
            :to="`/deals/${deal.id}`"
          >
            <div class="d-flex align-start justify-space-between">
              <div class="font-weight-semibold text-body-2 mb-1 primary--text">{{ deal.name }}</div>
              <v-menu offset-y left>
                <template v-slot:activator="{ on, attrs }">
                  <v-btn icon x-small v-bind="attrs" v-on="on" @click.prevent>
                    <v-icon x-small>mdi-dots-vertical</v-icon>
                  </v-btn>
                </template>
                <v-list dense>
                  <v-subheader class="text-caption">Move to stage</v-subheader>
                  <v-list-item
                    v-for="targetStage in stages"
                    :key="targetStage"
                    v-show="targetStage !== deal.stage"
                    @click.prevent="moveDealStage(deal, targetStage)"
                  >
                    <v-list-item-title class="text-caption">{{ targetStage }}</v-list-item-title>
                  </v-list-item>
                </v-list>
              </v-menu>
            </div>

            <div class="text-subtitle-2 font-weight-bold green--text text--darken-2 mb-2">
              {{ deal.amount ? formatCurrency(deal.amount) : '₹0' }}
            </div>

            <div class="d-flex align-center justify-space-between text-caption grey--text">
              <span>{{ deal.account_name || 'No Account' }}</span>
              <span>{{ deal.probability }}%</span>
            </div>
          </v-card>

          <div v-if="getStageDeals(stage).length === 0" class="text-center grey--text text-caption py-8 border-dashed rounded">
            No deals
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import dealService from '../../services/dealService'

export default {
  name: 'DealKanban',
  data() {
    return {
      loading: true,
      pipeline: {},
      summary: null,
      stages: ['Qualification', 'Needs Analysis', 'Proposal', 'Negotiation', 'Closed Won', 'Closed Lost'],
      stageColors: {
        Qualification: '#1976D2',
        'Needs Analysis': '#00ACC1',
        Proposal: '#FB8C00',
        Negotiation: '#8E24AA',
        'Closed Won': '#43A047',
        'Closed Lost': '#E53935'
      }
    }
  },
  async mounted() {
    await this.loadPipeline()
  },
  methods: {
    async loadPipeline() {
      this.loading = true
      try {
        const res = await dealService.getPipeline()
        this.pipeline = res.data.data.pipeline || {}
        this.summary = res.data.data.summary || null
      } catch (err) {
        this.$store.dispatch('snackbar/error', 'Failed to load pipeline')
      } finally {
        this.loading = false
      }
    },
    getStageDeals(stage) {
      return this.pipeline[stage] || []
    },
    getStageTotal(stage) {
      const deals = this.getStageDeals(stage)
      return deals.reduce((sum, d) => sum + (d.amount || 0), 0)
    },
    async moveDealStage(deal, newStage) {
      try {
        await dealService.updateStage(deal.id, newStage)
        this.$store.dispatch('snackbar/success', `Moved to ${newStage}`)
        await this.loadPipeline()
      } catch (err) {
        this.$store.dispatch('snackbar/error', 'Failed to update deal stage')
      }
    },
    formatCurrency(val) {
      if (!val && val !== 0) return '₹0'
      if (val >= 10000000) return '₹' + (val / 10000000).toFixed(1) + 'Cr'
      if (val >= 100000) return '₹' + (val / 100000).toFixed(1) + 'L'
      if (val >= 1000) return '₹' + (val / 1000).toFixed(0) + 'K'
      return '₹' + Number(val).toLocaleString()
    }
  }
}
</script>

<style scoped>
.gap-2 { gap: 8px; }
.kanban-board {
  display: flex;
  gap: 16px;
  overflow-x: auto;
  padding-bottom: 16px;
  min-height: calc(100vh - 240px);
}
.kanban-column {
  flex: 0 0 280px;
  background: #EFF2F5;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  max-height: calc(100vh - 240px);
}
.column-header {
  background: #FFFFFF;
  border-top: 4px solid #1976D2;
  border-radius: 8px 8px 0 0;
}
.column-body {
  overflow-y: auto;
  flex: 1;
}
.kanban-card {
  border-radius: 8px !important;
  transition: transform 0.2s, box-shadow 0.2s;
  background: #FFFFFF !important;
}
.kanban-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.1) !important;
}
.border-dashed {
  border: 1px dashed #CBD5E1;
}
</style>
