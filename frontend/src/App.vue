<script setup>
import HelloWorld from './components/HelloWorld.vue'
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <header class="bg-white shadow-sm border-b border-gray-200">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between items-center h-16">
          <div class="flex items-center space-x-4">
            <img alt="OgLogStream logo" class="h-8 w-8" src="./assets/logo.svg" />
            <h1 class="text-xl font-semibold text-gray-900">OgLogStream</h1>
          </div>
          <nav class="flex space-x-8">
            <router-link to="/" class="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium transition-colors">
              Dashboard
            </router-link>
            <router-link to="/live" class="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium transition-colors">
              Live Tail
            </router-link>
          </nav>
        </div>
      </div>
    </header>

    <main class="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
      <router-view />
    </main>
  </div>
</template>

<style scoped>
/* Tailwind CSS handles all styling */
</style>
