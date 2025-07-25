<template>
  <div class="space-y-6">
    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <div v-for="stat in stats" :key="stat.level" class="bg-white rounded-lg shadow p-6">
        <div class="flex items-center">
          <div class="flex-shrink-0">
            <div :class="getStatColor(stat.level)" class="h-8 w-8 rounded-md flex items-center justify-center">
              <svg class="h-5 w-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
              </svg>
            </div>
          </div>
          <div class="ml-5 w-0 flex-1">
            <dl>
              <dt class="text-sm font-medium text-gray-500 truncate">{{ stat.level.toUpperCase() }}</dt>
              <dd class="text-lg font-medium text-gray-900">{{ stat.count }}</dd>
            </dl>
          </div>
        </div>
      </div>
    </div>

    <!-- Logs Table -->
    <div class="bg-white rounded-lg shadow">
      <div class="px-4 py-5 sm:p-6">
        <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Recent Logs</h3>
        
        <!-- Filters -->
        <div class="mb-4 grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label for="level-filter" class="block text-sm font-medium text-gray-700">Level</label>
            <select
              id="level-filter"
              v-model="levelFilter"
              @change="fetchLogs"
              class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm rounded-md"
            >
              <option value="">All levels</option>
              <option value="error">Error</option>
              <option value="warn">Warning</option>
              <option value="info">Info</option>
              <option value="debug">Debug</option>
            </select>
          </div>
          <div>
            <label for="service-filter" class="block text-sm font-medium text-gray-700">Service</label>
            <input
              id="service-filter"
              v-model="serviceFilter"
              @input="fetchLogs"
              type="text"
              placeholder="Filter by service..."
              class="mt-1 focus:ring-blue-500 focus:border-blue-500 block w-full shadow-sm sm:text-sm border-gray-300 rounded-md"
            />
          </div>
        </div>

        <!-- Table -->
        <div class="overflow-hidden">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Time</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Level</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Service</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Message</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="log in logs" :key="log.timestamp + log.message">
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {{ formatTimestamp(log.timestamp) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                  <span :class="getLevelColor(log.level)" class="inline-flex px-2 py-1 text-xs font-semibold rounded-full">
                    {{ log.level?.toUpperCase() }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-blue-600">
                  {{ log.service }}
                </td>
                <td class="px-6 py-4 text-sm text-gray-900 font-mono break-all">
                  {{ log.message }}
                </td>
              </tr>
            </tbody>
          </table>
          
          <div v-if="logs.length === 0" class="text-center py-12">
            <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <h3 class="mt-2 text-sm font-medium text-gray-900">No logs found</h3>
            <p class="mt-1 text-sm text-gray-500">Try adjusting your filters or check back later.</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { config } from '../config.js'

const logs = ref([])
const stats = ref([])
const levelFilter = ref('')
const serviceFilter = ref('')

function getLevelColor(level) {
  const colors = {
    error: 'bg-red-100 text-red-800',
    warn: 'bg-yellow-100 text-yellow-800',
    info: 'bg-blue-100 text-blue-800',
    debug: 'bg-gray-100 text-gray-800',
    fatal: 'bg-purple-100 text-purple-800'
  }
  return colors[level?.toLowerCase()] || 'bg-gray-100 text-gray-800'
}

function getStatColor(level) {
  const colors = {
    error: 'bg-red-500',
    warn: 'bg-yellow-500',
    info: 'bg-blue-500',
    debug: 'bg-gray-500',
    fatal: 'bg-purple-500'
  }
  return colors[level?.toLowerCase()] || 'bg-gray-500'
}

function formatTimestamp(timestamp) {
  if (!timestamp) return ''
  return new Date(timestamp).toLocaleString()
}

async function fetchLogs() {
  try {
    const params = new URLSearchParams()
    if (levelFilter.value) params.append('level', levelFilter.value)
    if (serviceFilter.value) params.append('service', serviceFilter.value)
    
    const response = await fetch(`${config.apiBaseUrl}/api/logs?${params}`)
    logs.value = await response.json()
  } catch (error) {
    console.error('Failed to fetch logs:', error)
  }
}

async function fetchStats() {
  try {
    const response = await fetch(`${config.apiBaseUrl}/api/stats`)
    stats.value = await response.json()
  } catch (error) {
    console.error('Failed to fetch stats:', error)
  }
}

onMounted(() => {
  fetchLogs()
  fetchStats()
})
</script> 