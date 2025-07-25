// API Configuration
export const config = {
  // Определяем базовый URL для API
  apiBaseUrl: getApiBaseUrl(),
  wsBaseUrl: getWsBaseUrl()
}

function getApiBaseUrl() {
  // В продакшн режиме (когда фронтенд на :3000) используем HAProxy
  if (window.location.port === '3000') {
    return `${window.location.protocol}//${window.location.hostname}:80`
  }
  
  // В dev режиме используем текущий хост (nginx проксирует)
  return ''
}

function getWsBaseUrl() {
  // В продакшн режиме используем HAProxy
  if (window.location.port === '3000') {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    return `${protocol}//${window.location.hostname}:80`
  }
  
  // В dev режиме используем текущий хост
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${protocol}//${window.location.host}`
} 