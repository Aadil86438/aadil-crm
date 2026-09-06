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
          <v-tabs v-model="currentView" color="primary" class="admin-main-tabs" active-class="font-weight-bold" show-arrows center-active>
            <v-tab value="members">
              <v-icon left small>mdi-account-group</v-icon>
              Member Registrations
            </v-tab>
            <v-tab value="redis">
              <v-icon left small color="red">mdi-database</v-icon>
              Redis Inspector
              <v-chip size="x-small" color="red lighten-5" class="ml-2 font-weight-bold text-caption red--text" label v-if="redisData.connected">
                {{ redisData.key_count }} keys
              </v-chip>
            </v-tab>
            <v-tab value="k8s">
              <v-icon left small color="blue">mdi-kubernetes</v-icon>
              Kubernetes Cluster
              <v-chip size="x-small" color="blue lighten-5" class="ml-2 font-weight-bold text-caption blue--text" label v-if="k8sData.connected">
                {{ k8sData.pod_count }} pods
              </v-chip>
            </v-tab>
          </v-tabs>
        </v-card>

        <!-- VIEW 1: MEMBER REGISTRATIONS -->
        <div v-if="currentView === 0">
          <!-- KPI Metric Cards (2x2 on mobile, 4x1 on desktop) -->
          <v-row class="mb-6" dense>
            <v-col cols="6" sm="3">
              <v-card class="kpi-card pa-3 pa-sm-4" elevation="0" @click="statusFilter = 'all'" :class="{ 'active-kpi': statusFilter === 'all' }">
                <div class="d-flex align-center justify-space-between mb-1 mb-sm-2">
                  <span class="text-caption grey--text text--darken-1 font-weight-bold">TOTAL</span>
                  <v-avatar color="blue lighten-5" size="32">
                    <v-icon color="primary" small>mdi-account-group</v-icon>
                  </v-avatar>
                </div>
                <div class="text-h5 text-sm-h4 font-weight-black text--primary">{{ allRequests.length }}</div>
                <div class="text-caption grey--text mt-1 hidden-xs-only">All registrations</div>
              </v-card>
            </v-col>
            <v-col cols="6" sm="3">
              <v-card class="kpi-card pa-3 pa-sm-4" elevation="0" @click="statusFilter = 'pending'" :class="{ 'active-kpi': statusFilter === 'pending' }">
                <div class="d-flex align-center justify-space-between mb-1 mb-sm-2">
                  <span class="text-caption warning--text text--darken-2 font-weight-bold">PENDING</span>
                  <v-avatar color="amber lighten-5" size="32">
                    <v-icon color="warning" small>mdi-clock-outline</v-icon>
                  </v-avatar>
                </div>
                <div class="text-h5 text-sm-h4 font-weight-black warning--text text--darken-2">{{ pendingCount }}</div>
                <div class="text-caption grey--text mt-1 hidden-xs-only">Awaiting action</div>
              </v-card>
            </v-col>
            <v-col cols="6" sm="3">
              <v-card class="kpi-card pa-3 pa-sm-4" elevation="0" @click="statusFilter = 'approved'" :class="{ 'active-kpi': statusFilter === 'approved' }">
                <div class="d-flex align-center justify-space-between mb-1 mb-sm-2">
                  <span class="text-caption success--text font-weight-bold">APPROVED</span>
                  <v-avatar color="green lighten-5" size="32">
                    <v-icon color="success" small>mdi-check-circle-outline</v-icon>
                  </v-avatar>
                </div>
                <div class="text-h5 text-sm-h4 font-weight-black success--text">{{ approvedCount }}</div>
                <div class="text-caption grey--text mt-1 hidden-xs-only">Active CRM users</div>
              </v-card>
            </v-col>
            <v-col cols="6" sm="3">
              <v-card class="kpi-card pa-3 pa-sm-4" elevation="0" @click="statusFilter = 'rejected'" :class="{ 'active-kpi': statusFilter === 'rejected' }">
                <div class="d-flex align-center justify-space-between mb-1 mb-sm-2">
                  <span class="text-caption error--text font-weight-bold">REJECTED</span>
                  <v-avatar color="red lighten-5" size="32">
                    <v-icon color="error" small>mdi-close-circle-outline</v-icon>
                  </v-avatar>
                </div>
                <div class="text-h5 text-sm-h4 font-weight-black error--text">{{ rejectedCount }}</div>
                <div class="text-caption grey--text mt-1 hidden-xs-only">Declined access</div>
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
                    <v-icon left x-small>mdi-close</v-icon> Reject
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

        <!-- VIEW 3: KUBERNETES CLUSTER VISUALIZER & CRASH SIMULATOR -->
        <div v-else-if="currentView === 2">
          <!-- Connection Status & Controls Bar -->
          <v-card class="mb-6 pa-5 banner-card" elevation="0">
            <div class="d-flex align-center justify-space-between flex-wrap" style="gap: 16px">
              <div class="d-flex align-center">
                <v-avatar :color="k8sData.connected ? 'blue lighten-5' : 'amber lighten-5'" size="52" class="mr-3 elevation-1">
                  <v-icon :color="k8sData.connected ? 'primary' : 'warning'" size="30">
                    mdi-kubernetes
                  </v-icon>
                </v-avatar>
                <div>
                  <div class="d-flex align-center">
                    <span class="text-h6 font-weight-black text--primary mr-2">Kubernetes Engine Visualizer</span>
                    <v-chip :color="k8sData.connected ? 'primary' : 'warning'" small label class="font-weight-bold text-white">
                      {{ k8sData.connected ? 'CLUSTER ONLINE' : 'SIMULATED MODE' }}
                    </v-chip>
                  </div>
                  <div class="text-body-2 grey--text">
                    Node: <strong class="black--text">{{ k8sNodeName }}</strong> |
                    Healthy Pods: <strong class="primary--text font-weight-bold">{{ k8sData.pod_count }} / 4</strong> |
                    Cluster Engine: <strong class="success--text">Kubelet Active</strong>
                  </div>
                </div>
              </div>

              <div class="d-flex align-center" style="gap: 12px">
                <v-btn color="primary" outlined small @click="fetchK8sData" :loading="loadingK8s">
                  <v-icon left small>mdi-refresh</v-icon>
                  Poll Cluster Live
                </v-btn>
              </div>
            </div>
          </v-card>

          <!-- 🌌 BIG CLUSTER ARCHITECTURE FLOWCHART & TOPOLOGY MAP -->
          <v-card class="mb-6 pa-6 k8s-architecture-card elevation-0" style="background: linear-gradient(135deg, #0F172A 0%, #1E293B 100%); border-radius: 20px; border: 1px solid #334155;">
            <div class="d-flex align-center justify-space-between mb-4 flex-wrap" style="gap: 12px">
              <div class="d-flex align-center">
                <v-icon color="cyan lighten-2" class="mr-2" size="28">mdi-sitemap</v-icon>
                <div>
                  <h3 class="text-h6 font-weight-bold white--text">Kubernetes Cluster Architecture & Topology Map</h3>
                  <div class="caption cyan--text text--lighten-3">Node: <strong>{{ k8sNodeName }}</strong> | Services Router & Ingress Layer</div>
                </div>
              </div>
              <v-chip color="cyan darken-3" dark small label class="font-weight-bold">
                <v-icon left x-small color="cyan lighten-3">mdi-shield-check</v-icon>
                REPLICASET CONTROLLER ACTIVE
              </v-chip>
            </div>

            <!-- Visual Container Box: KUBERNETES NODE -->
            <div class="k8s-node-box pa-4 rounded-xl" style="background: rgba(15, 23, 42, 0.7); border: 2px dashed #475569;">
              <div class="d-flex justify-space-between align-center mb-4 pb-3" style="border-bottom: 1px solid #334155;">
                <div class="d-flex align-center">
                  <v-icon color="cyan lighten-3" small class="mr-2">mdi-server</v-icon>
                  <span class="text-subtitle-2 font-weight-bold cyan--text text--lighten-3">KUBERNETES NODE: {{ k8sNodeName }} (Linux x86_64)</span>
                </div>
                <div class="caption grey--text text--lighten-1">
                  Cluster Subnet: <code class="cyan--text">10.244.0.0/16</code> | Control Plane: <span class="green--text text--lighten-2 font-weight-bold">● HEALTHY</span>
                </div>
              </div>

              <!-- Component Topology Cards Row -->
              <v-row dense>
                <v-col v-for="comp in topologyComponents" :key="comp.component" cols="12" sm="6" md="3">
                  <div
                    class="topology-card pa-3 rounded-lg text-center"
                    :class="{ 'dying-border': isPodDying(comp.component), 'healthy-border': !isPodDying(comp.component) }"
                    style="background: #1E293B; border: 1px solid #334155; position: relative; transition: all 0.3s ease;"
                  >
                    <!-- Ingress / Service Line Header -->
                    <div class="caption text-uppercase font-weight-bold mb-1" :style="{ color: comp.color }">
                      <v-icon x-small :color="comp.color" class="mr-1">mdi-swap-horizontal</v-icon>
                      {{ comp.serviceName }}
                    </div>
                    <v-avatar size="36" :color="comp.bg" class="my-1">
                      <v-icon :color="comp.color" small>{{ comp.icon }}</v-icon>
                    </v-avatar>

                    <!-- Live Pod Name in Topology -->
                    <div class="caption white--text font-weight-bold text-truncate mt-1">
                      {{ getPodName(comp.component) }}
                    </div>

                    <!-- Status Pill -->
                    <div class="d-flex justify-center align-center mt-2" style="gap: 4px">
                      <v-chip x-small dark :color="isPodDying(comp.component) ? 'error' : 'success'" class="font-weight-bold">
                        <v-icon left x-small>{{ isPodDying(comp.component) ? 'mdi-alert-circle' : 'mdi-check-circle' }}</v-icon>
                        {{ isPodDying(comp.component) ? 'CRASHED / DEAD' : '1/1 READY' }}
                      </v-chip>
                      <span class="caption grey--text text--lighten-1">:{{ comp.port }}</span>
                    </div>
                  </div>
                </v-col>
              </v-row>
            </div>
          </v-card>

          <!-- 🎓 Educational Banner -->
          <v-alert type="info" outlined dense class="mb-6 rounded-lg" icon="mdi-lightbulb-on-outline" style="background: white">
            <strong>🎓 How Kubernetes Auto-Healing Works:</strong> Click <strong>💥 SIMULATE CRASH / KILL POD</strong> on any pod below.
            You will observe:
            <ol class="mt-1 pl-4 caption">
              <li><strong>1. Crash / Termination:</strong> Pod status turns <span class="error--text font-weight-bold">TERMINATING / CRASHED</span> (Red glowing border).</li>
              <li><strong>2. Detection:</strong> ReplicaSet controller detects missing pod within milliseconds.</li>
              <li><strong>3. Auto-Healing:</strong> Kubernetes automatically spins up a brand-new replacement Pod (Green <span class="success--text font-weight-bold">1/1 READY</span>).</li>
              <li><strong>4. Cleanup:</strong> The terminated dead pod disappears, leaving your app 100% healthy!</li>
            </ol>
          </v-alert>

          <!-- 📦 PODS VISUAL CARDS GRID WITH ANIMATED LIFECYCLE -->
          <v-row class="mb-6">
            <v-col v-for="pod in displayPods" :key="pod.name" cols="12" md="6">
              <v-card
                class="pa-4 rounded-xl elevation-0 pod-visual-card"
                :class="{
                  'dying-pod-card': pod.isDying,
                  'creating-pod-card': pod.isCreating,
                  'healthy-pod-card': !pod.isDying && !pod.isCreating
                }"
              >
                <!-- Pod Header -->
                <div class="d-flex align-center justify-space-between mb-3">
                  <div class="d-flex align-center">
                    <v-avatar size="42" :color="getCompColorBg(pod.component)" class="mr-3">
                      <v-icon :color="getCompColor(pod.component)" small>
                        {{ getCompIcon(pod.component) }}
                      </v-icon>
                    </v-avatar>
                    <div>
                      <div class="font-weight-bold text-subtitle-2 text--primary d-flex align-center">
                        <code>{{ pod.name }}</code>
                      </div>
                      <div class="caption grey--text">
                        Component: <strong class="black--text">{{ pod.component.toUpperCase() }}</strong>
                      </div>
                    </div>
                  </div>

                  <!-- Status Chip -->
                  <v-chip
                    small
                    :color="pod.isDying ? 'error' : (pod.isCreating ? 'warning' : 'success')"
                    dark
                    class="font-weight-bold px-3"
                    :class="{ 'pulse-anim': pod.isDying }"
                  >
                    <v-progress-circular v-if="pod.isDying || pod.isCreating" indeterminate size="12" width="2" class="mr-1"></v-progress-circular>
                    <v-icon v-else left x-small>mdi-check-circle</v-icon>
                    {{ pod.isDying ? 'TERMINATING / DEAD' : (pod.isCreating ? 'CONTAINER CREATING' : 'RUNNING (1/1)') }}
                  </v-chip>
                </div>

                <v-divider class="mb-3"></v-divider>

                <!-- Detailed Specs Grid -->
                <div class="d-flex justify-space-between caption grey--text text--darken-2 mb-3 flex-wrap" style="gap: 8px">
                  <div>Node: <strong class="black--text">{{ pod.node }}</strong></div>
                  <div>Pod IP: <strong class="black--text">{{ pod.ip || '10.244.0.x' }}</strong></div>
                  <div>Ready: <strong :class="pod.isDying ? 'error--text' : 'success--text'">{{ pod.isDying ? '0/1' : pod.ready }}</strong></div>
                  <div>Restarts: <strong class="error--text font-weight-bold">{{ pod.restarts }}</strong></div>
                </div>

                <!-- Crash Progress Bar / Status Banner -->
                <div v-if="pod.isDying" class="pa-3 red lighten-5 rounded-lg mb-3 border-error">
                  <div class="caption error--text font-weight-bold d-flex align-center">
                    <v-icon color="error" x-small class="mr-1">mdi-alert</v-icon>
                    CRASH DETECTED! Kubernetes force-killing container...
                  </div>
                  <v-progress-linear indeterminate color="error" height="4" class="mt-2 rounded"></v-progress-linear>
                </div>

                <div v-else-if="pod.isCreating" class="pa-3 amber lighten-5 rounded-lg mb-3 border-warning">
                  <div class="caption warning--text text--darken-3 font-weight-bold d-flex align-center">
                    <v-icon color="warning" x-small class="mr-1">mdi-cog-sync</v-icon>
                    AUTO-HEAL: Pulling container image & starting replacement...
                  </div>
                  <v-progress-linear indeterminate color="warning" height="4" class="mt-2 rounded"></v-progress-linear>
                </div>

                <!-- Kill Action Button -->
                <v-btn
                  v-else
                  color="error"
                  block
                  outlined
                  class="rounded-lg font-weight-bold kill-btn"
                  :loading="killingPod === pod.name"
                  @click="killPodWithAnimation(pod)"
                >
                  <v-icon left small>mdi-bomb</v-icon>
                  💥 SIMULATE CRASH / KILL POD
                </v-btn>
              </v-card>
            </col>
          </v-row>

          <!-- 💻 REAL-TIME KUBERNETES EVENT CONSOLE -->
          <v-card class="pa-5 console-card elevation-0 mb-6" style="background: #0F172A; border-radius: 16px; border: 1px solid #1E293B;">
            <div class="d-flex align-center justify-space-between mb-3">
              <div class="d-flex align-center">
                <v-icon color="green lighten-2" small class="mr-2">mdi-console-line</v-icon>
                <span class="text-subtitle-2 font-weight-bold white--text">Kubernetes ReplicaSet Live Event Stream</span>
              </div>
              <v-btn text x-small color="grey lighten-1" @click="k8sLogs = []">Clear Stream</v-btn>
            </div>
            <div class="console-body pa-3 rounded-lg" style="background: #020617; max-height: 180px; overflow-y: auto; font-family: monospace; font-size: 12px;">
              <div v-for="(log, idx) in k8sLogs" :key="idx" class="console-line py-1" :class="log.type">
                <span class="grey--text">[{{ log.time }}]</span>
                <span class="ml-2 font-weight-bold">{{ log.text }}</span>
              </div>
              <div v-if="k8sLogs.length === 0" class="grey--text italic">Awaiting cluster events... Click "SIMULATE CRASH" above to see live ReplicaSet events.</div>
            </div>
          </v-card>
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
      k8sData: { connected: false, pod_count: 0, pods: [] },
      loadingK8s: false,
      killingPod: null,
      dyingPodNames: [],
      creatingPodNames: [],
      k8sLogs: [
        { time: new Date().toLocaleTimeString(), type: 'info', text: '🌌 K8S CLUSTER VISUALIZER: Connected to cluster node.' },
        { time: new Date().toLocaleTimeString(), type: 'info', text: '✅ REPLICASET CONTROLLER: 4/4 Desired pods active and healthy.' }
      ],
      topologyComponents: [
        { component: 'postgres', serviceName: 'postgres-svc', port: '5432', icon: 'mdi-database', color: '#60A5FA', bg: 'blue lighten-5' },
        { component: 'redis', serviceName: 'redis-svc', port: '6379', icon: 'mdi-database-clock', color: '#F87171', bg: 'red lighten-5' },
        { component: 'backend', serviceName: 'crm-backend-svc', port: '8080', icon: 'mdi-cog-sync', color: '#34D399', bg: 'green lighten-5' },
        { component: 'crm-frontend', serviceName: 'crm-frontend-svc', port: '80', icon: 'mdi-web', color: '#A78BFA', bg: 'purple lighten-5' }
      ],
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
      } else if (val === 2) {
        this.fetchK8sData()
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
    },
    k8sNodeName() {
      return (this.k8sData.pods && this.k8sData.pods[0] && this.k8sData.pods[0].node) || 'docker-desktop'
    },
    displayPods() {
      if (!this.k8sData.pods) return []
      return this.k8sData.pods.map(pod => {
        const isDying = this.dyingPodNames.includes(pod.name)
        const isCreating = this.creatingPodNames.includes(pod.name)
        return {
          ...pod,
          isDying,
          isCreating
        }
      })
    }
  },
  methods: {
    addK8sLog(text, type = 'info') {
      const time = new Date().toLocaleTimeString()
      this.k8sLogs.unshift({ time, text, type })
      if (this.k8sLogs.length > 30) this.k8sLogs.pop()
    },
    getPodName(comp) {
      if (!this.k8sData.pods) return 'crm-' + comp + '-pod'
      const found = this.k8sData.pods.find(p => p.component === comp)
      return found ? found.name : 'crm-' + comp + '-pod'
    },
    isPodDying(comp) {
      if (!this.k8sData.pods) return false
      const found = this.k8sData.pods.find(p => p.component === comp)
      return found ? this.dyingPodNames.includes(found.name) : false
    },
    getCompColor(comp) {
      if (comp === 'postgres') return 'blue lighten-1'
      if (comp === 'redis') return 'red lighten-1'
      if (comp === 'crm-frontend') return 'purple lighten-1'
      return 'green lighten-1'
    },
    getCompColorBg(comp) {
      if (comp === 'postgres') return 'blue lighten-5'
      if (comp === 'redis') return 'red lighten-5'
      if (comp === 'crm-frontend') return 'purple lighten-5'
      return 'green lighten-5'
    },
    getCompIcon(comp) {
      if (comp === 'postgres') return 'mdi-database'
      if (comp === 'redis') return 'mdi-database-clock'
      if (comp === 'crm-frontend') return 'mdi-web'
      return 'mdi-cog-sync'
    },
    async killPodWithAnimation(pod) {
      const podName = pod.name
      const compName = pod.component.toUpperCase()
      this.killingPod = podName

      // 1. Instantly trigger visual crash state (Red border, pulsing, status TERMINATING)
      this.dyingPodNames.push(podName)
      this.addK8sLog(`💥 COMMAND: Force-deleting Pod [${podName}] (SIGKILL sent)`, 'error')
      this.addK8sLog(`🚨 EVENT: Pod ${podName} entered TERMINATING state. ReplicaSet desired=1, actual=0.`, 'error')

      try {
        await registrationService.killK8sPod(podName, this.adminToken)

        // 2. Step 2 (1.2s): Show Container Creating log
        setTimeout(() => {
          this.addK8sLog(`⚡ K8S SCHEDULER: Triggered auto-healing for ${compName}. Assigning node ${this.k8sNodeName}...`, 'warn')
          this.addK8sLog(`📦 KUBELET: Pulling container image & provisioning replacement Pod...`, 'info')
        }, 1200)

        // 3. Step 3 (2.5s): Auto-healing complete! Fetch fresh pods, clear dying state, show success
        setTimeout(async () => {
          this.dyingPodNames = this.dyingPodNames.filter(n => n !== podName)
          await this.fetchK8sData()
          this.addK8sLog(`✅ AUTO-HEAL COMPLETE: New replacement Pod for ${compName} is 1/1 READY! Dead pod removed.`, 'success')
        }, 2600)
      } catch (e) {
        this.dyingPodNames = this.dyingPodNames.filter(n => n !== podName)
        alert(e.response?.data?.message || 'Failed to kill pod')
      } finally {
        this.killingPod = null
      }
    },
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
    async fetchK8sData() {
      this.loadingK8s = true
      try {
        const res = await registrationService.getK8sStatus(this.adminToken)
        this.k8sData = res.data.data || { connected: false, pod_count: 0, pods: [] }
      } catch (e) {
        console.error('Failed to fetch K8s cluster status', e)
      } finally {
        this.loadingK8s = false
      }
    },
    async killPod(podName) {
      this.killingPod = podName
      try {
        await registrationService.killK8sPod(podName, this.adminToken)
        setTimeout(() => this.fetchK8sData(), 800)
        setTimeout(() => this.fetchK8sData(), 2500)
      } catch (e) {
        alert(e.response?.data?.message || 'Failed to kill pod')
      } finally {
        this.killingPod = null
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
.k8s-architecture-card {
  box-shadow: 0 10px 40px rgba(15, 23, 42, 0.15) !important;
}
.topology-card {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
.healthy-border {
  border-color: #334155 !important;
}
.dying-border {
  border-color: #EF4444 !important;
  box-shadow: 0 0 15px rgba(239, 68, 68, 0.4) !important;
  animation: pulse-red 1s infinite alternate;
}
.pod-visual-card {
  background: white;
  border: 1px solid #E2E8F0;
  transition: all 0.3s ease;
}
.healthy-pod-card:hover {
  border-color: #3B82F6 !important;
  box-shadow: 0 8px 24px rgba(59, 130, 246, 0.08) !important;
}
.dying-pod-card {
  border: 2px solid #EF4444 !important;
  background: #FEF2F2 !important;
  animation: pulse-red 0.8s infinite alternate;
}
.creating-pod-card {
  border: 2px solid #F59E0B !important;
  background: #FFFBEB !important;
}
@keyframes pulse-red {
  0% { transform: scale(1); box-shadow: 0 0 0 rgba(239, 68, 68, 0.4); }
  100% { transform: scale(1.01); box-shadow: 0 0 16px rgba(239, 68, 68, 0.6); }
}
.console-line.info { color: #38BDF8; }
.console-line.warn { color: #FBBF24; }
.console-line.error { color: #F87171; }
.console-line.success { color: #4ADE80; }

@media (max-width: 600px) {
  .admin-bg {
    padding: 12px 6px !important;
  }
  .banner-card {
    padding: 14px !important;
  }
  .k8s-architecture-card {
    padding: 14px !important;
    border-radius: 14px !important;
  }
  .k8s-node-box {
    padding: 10px !important;
  }
  .topology-card {
    padding: 10px !important;
  }
  .code-input >>> input {
    font-size: 16px !important;
    letter-spacing: 4px !important;
  }
}
</style>
