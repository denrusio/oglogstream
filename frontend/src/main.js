import './assets/tailwind.css'

import { createApp } from 'vue'
import App from './App.vue'
import { createRouter, createWebHistory } from 'vue-router'

import Live from './pages/Live.vue'
import Dashboard from './pages/Dashboard.vue'

const routes = [
  { path: '/', component: Dashboard },
  { path: '/live', component: Live },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

createApp(App).use(router).mount('#app')
