<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="bg-white rounded-lg shadow p-6">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-2xl font-bold text-gray-900">Live Tail & Testing</h2>
          <p class="text-gray-600">Send test logs and watch them in real-time</p>
        </div>
        <div class="flex items-center space-x-4">
          <!-- Connection Status -->
          <div class="flex items-center space-x-2">
            <div :class="['h-2 w-2 rounded-full', isConnected ? 'bg-green-500 animate-pulse' : 'bg-red-500']"></div>
            <span class="text-sm text-gray-600">{{ isConnected ? 'Connected' : 'Disconnected' }}</span>
          </div>
          <!-- Last Send Status -->
          <div class="flex items-center space-x-2">
            <div :class="['h-2 w-2 rounded-full', lastStatus === 'success' ? 'bg-blue-500' : lastStatus === 'error' ? 'bg-red-500' : 'bg-gray-400']"></div>
            <span class="text-sm text-gray-600">{{ getStatusText() }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Split Layout: Form + Live Logs -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      
      <!-- Left: Test Form -->
      <div class="space-y-4">
        <!-- Quick Send Form -->
        <div class="bg-white rounded-lg shadow p-6">
          <h3 class="text-lg font-medium text-gray-900 mb-4">Send Test Log</h3>
          
          <form @submit.prevent="sendLog" class="space-y-4">
            <div class="grid grid-cols-2 gap-3">
              <!-- Level -->
              <div>
                <label for="level" class="block text-sm font-medium text-gray-700 mb-1">Level</label>
                <select 
                  v-model="form.level" 
                  id="level"
                  class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
                >
                  <option value="info">Info</option>
                  <option value="warn">Warning</option>
                  <option value="error">Error</option>
                  <option value="fatal">Fatal</option>
                </select>
              </div>

              <!-- Service -->
              <div>
                <label for="service" class="block text-sm font-medium text-gray-700 mb-1">Service</label>
                <input 
                  v-model="form.service"
                  type="text" 
                  id="service"
                  placeholder="auth-service"
                  class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
                />
              </div>
            </div>

            <!-- Message -->
            <div>
              <label for="message" class="block text-sm font-medium text-gray-700 mb-1">Message</label>
              <textarea 
                v-model="form.message"
                id="message"
                rows="3"
                placeholder="Enter log message..."
                class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
              ></textarea>
            </div>

            <!-- Buttons -->
            <div class="flex space-x-3">
              <button 
                type="submit"
                :disabled="isLoading || !form.message.trim() || !form.service.trim()"
                class="flex-1 px-4 py-2 bg-blue-600 text-white font-medium rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors text-sm"
              >
                {{ isLoading ? 'Sending...' : 'Send Log' }}
              </button>

              <button 
                type="button"
                @click="sendRandomLog"
                :disabled="isLoading"
                class="px-4 py-2 bg-gray-600 text-white font-medium rounded-md hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors text-sm"
              >
                Random
              </button>
            </div>
          </form>
        </div>

        <!-- Quick Presets -->
        <div class="bg-white rounded-lg shadow p-6">
          <h3 class="text-lg font-medium text-gray-900 mb-4">Quick Tests</h3>
          
          <div class="grid grid-cols-1 gap-2">
            <button 
              v-for="preset in presetLogs" 
              :key="preset.name"
              @click="sendPresetLog(preset)"
              :disabled="isLoading"
              class="p-3 text-left border border-gray-200 rounded-lg hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              <div class="flex items-center justify-between mb-1">
                <span class="text-sm font-medium text-gray-900">{{ preset.name }}</span>
                <span 
                  :class="getLevelColor(preset.level)"
                  class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium"
                >
                  {{ preset.level.toUpperCase() }}
                </span>
              </div>
              <div class="text-xs text-gray-500">{{ preset.service }}</div>
              <div class="text-xs text-gray-600 mt-1">{{ preset.message.substring(0, 60) }}...</div>
            </button>
          </div>
        </div>

        <!-- Send History -->
        <div v-if="responses.length > 0" class="bg-white rounded-lg shadow p-6">
          <h3 class="text-lg font-medium text-gray-900 mb-4">Send History</h3>
          
          <div class="space-y-2 max-h-40 overflow-y-auto">
            <div 
              v-for="(response, idx) in responses.slice(0, 5)" 
              :key="idx"
              class="flex items-center space-x-2 p-2 rounded border border-gray-100 text-sm"
              :class="response.success ? 'bg-green-50' : 'bg-red-50'"
            >
              <svg v-if="response.success" class="h-4 w-4 text-green-500 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
              </svg>
              <svg v-else class="h-4 w-4 text-red-500 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
              </svg>
              <div class="flex-1 min-w-0">
                <div class="text-xs text-gray-500">{{ response.timestamp }}</div>
                <div class="font-medium truncate" :class="response.success ? 'text-green-800' : 'text-red-800'">
                  {{ response.message }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Right: Live Stream -->
      <div class="bg-white rounded-lg shadow">
        <div class="p-4 border-b border-gray-200 flex items-center justify-between">
          <h3 class="text-lg font-medium text-gray-900">Live Stream</h3>
          <div class="flex items-center space-x-2">
            <button 
              @click="clearLogs"
              class="px-3 py-1 text-xs bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 transition-colors"
            >
              Clear
            </button>
            <span class="text-xs text-gray-500">{{ logs.length }} logs</span>
          </div>
        </div>
        
        <div class="p-4">
          <div v-if="logs.length === 0" class="text-center py-12">
            <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <h3 class="mt-2 text-sm font-medium text-gray-900">No logs yet</h3>
            <p class="mt-1 text-sm text-gray-500">Send a test log to see it appear here</p>
          </div>
          
          <div v-else class="space-y-2 max-h-96 overflow-y-auto" ref="logsContainer">
            <div
              v-for="(log, idx) in logs"
              :key="idx"
              class="flex items-start space-x-3 p-3 rounded-lg border border-gray-100 hover:bg-gray-50 transition-colors"
              :class="{ 'ring-2 ring-blue-200 bg-blue-50': log.isNew }"
            >
              <div class="flex-shrink-0">
                <span
                  :class="getLevelColor(log.level)"
                  class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
                >
                  {{ log.level?.toUpperCase() }}
                </span>
              </div>
              <div class="flex-1 min-w-0">
                <div class="flex items-center space-x-2 text-sm text-gray-500 mb-1">
                  <span>{{ formatTimestamp(log.timestamp) }}</span>
                  <span>•</span>
                  <span class="font-medium text-blue-600">{{ log.service }}</span>
                </div>
                <p class="text-gray-900 font-mono text-sm break-all">{{ log.message }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue'
import { config } from '../config.js'

const logs = ref([])
const isConnected = ref(false)
const isLoading = ref(false)
const lastStatus = ref('idle') // 'idle', 'success', 'error'
const responses = ref([])
const logsContainer = ref(null)

let ws

const form = reactive({
  level: 'info',
  service: 'test-service',
  message: ''
})

const presetLogs = [
  {
    name: 'User Login',
    level: 'info',
    service: 'auth-service',
    message: 'User john@example.com successfully logged in from IP 192.168.1.100'
  },
  {
    name: 'Payment Error',
    level: 'error',
    service: 'payment-api',
    message: 'Payment processing failed for order #12345: Insufficient funds'
  },
  {
    name: 'Database Warning',
    level: 'warn',
    service: 'database-svc',
    message: 'Connection pool utilization is high (85%). Consider scaling up.'
  },
  {
    name: 'System Critical',
    level: 'fatal',
    service: 'core-system',
    message: 'Critical system failure detected. Automatic failover initiated.'
  }
]

function getLevelColor(level) {
  const colors = {
    info: 'bg-blue-100 text-blue-800',
    warn: 'bg-yellow-100 text-yellow-800',
    error: 'bg-red-100 text-red-800',
    fatal: 'bg-purple-100 text-purple-800'
  }
  return colors[level] || 'bg-gray-100 text-gray-800'
}

function formatTimestamp(timestamp) {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  return date.toLocaleTimeString()
}

function getStatusText() {
  if (lastStatus.value === 'success') return 'Last: Success'
  if (lastStatus.value === 'error') return 'Last: Failed'
  return 'Ready'
}

function addResponse(success, message) {
  responses.value.unshift({
    success,
    message,
    timestamp: new Date().toLocaleTimeString()
  })
  
  // Keep only last 10 responses
  if (responses.value.length > 10) {
    responses.value.pop()
  }
}

function clearLogs() {
  logs.value = []
}

async function scrollToBottom() {
  await nextTick()
  if (logsContainer.value) {
    logsContainer.value.scrollTop = logsContainer.value.scrollHeight
  }
}

function handleMessage(event) {
  try {
    const log = JSON.parse(event.data)
    log.isNew = true
    logs.value.unshift(log)
    
    // Remove highlight after 3 seconds
    setTimeout(() => {
      log.isNew = false
    }, 3000)
    
    // Keep only last 100 logs
    if (logs.value.length > 100) {
      logs.value.pop()
    }
    
    scrollToBottom()
  } catch (error) {
    console.error('Failed to parse log:', error)
  }
}

async function sendLog() {
  if (!form.message.trim() || !form.service.trim()) return

  isLoading.value = true
  lastStatus.value = 'idle'

  try {
    const logEntry = {
      level: form.level,
      service: form.service.trim(),
      message: form.message.trim(),
      timestamp: new Date().toISOString()
    }

    const response = await fetch(`${config.apiBaseUrl}/log`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(logEntry)
    })

    if (response.ok) {
      lastStatus.value = 'success'
      addResponse(true, `${form.level.toUpperCase()} from ${form.service}`)
      
      // Clear form
      form.message = ''
    } else {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`)
    }
  } catch (error) {
    lastStatus.value = 'error'
    addResponse(false, `Failed: ${error.message}`)
    console.error('Failed to send log:', error)
  } finally {
    isLoading.value = false
  }
}

async function sendPresetLog(preset) {
  form.level = preset.level
  form.service = preset.service
  form.message = preset.message
  
  await sendLog()
}

function sendRandomLog() {
  const randomPreset = presetLogs[Math.floor(Math.random() * presetLogs.length)]
  const timestamp = new Date().toLocaleTimeString()
  
  form.level = randomPreset.level
  form.service = randomPreset.service
  form.message = `${randomPreset.message} (sent at ${timestamp})`
  
  sendLog()
}

onMounted(() => {
  ws = new WebSocket(`${config.wsBaseUrl}/ws/live`)
  
  ws.onmessage = handleMessage
  
  ws.onopen = () => {
    console.log('WebSocket connected')
    isConnected.value = true
  }
  
  ws.onerror = (error) => {
    console.error('WebSocket error:', error)
    isConnected.value = false
  }
  
  ws.onclose = () => {
    console.log('WebSocket disconnected')
    isConnected.value = false
  }
})

onUnmounted(() => {
  if (ws) {
    ws.close()
  }
})
</script> 