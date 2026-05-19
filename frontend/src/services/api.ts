import axios from 'axios'
import type {
  Device,
  DeviceWithProxy,
  SessionStatus,
  BulkSendRequest,
  BulkSendStatus,
  CheckUserResponse,
  HealthStats,
  ValidationCacheStats,
  Message,
} from '@/types'

const api = axios.create({
  baseURL: '/api',
  timeout: 120000, // 2 minutes for bulk operations
  headers: {
    'Content-Type': 'application/json',
  },
})

// Add auth token to requests
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token')
    console.log('[API] Request to:', config.url, 'Token present:', !!token)
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Session Management
export const sessionApi = {
  getAll: () => api.get<Device[]>('/devices'),

  getOne: (jid: string) => api.get<Device>(`/devices/${jid}`),

  getStatus: (sender: string) =>
    api.get<SessionStatus>('/status', { params: { sender } }),

  connect: (sender: string) =>
    api.get('/connect', { params: { sender } }),

  connectBulk: (senders: string[]) =>
    api.post('/connect-bulk', { senders }),

  disconnect: (sender: string) =>
    api.get('/disconnect', { params: { sender } }),

  disconnectBulk: (senders: string[]) =>
    api.post('/disconnect-bulk', { senders }),

  logout: (sender: string) =>
    api.post('/logout', {}, { params: { sender } }),

  getQR: (sender: string) =>
    api.get('/qr', { params: { sender }, responseType: 'blob' }),

  getQRJson: (sender: string) =>
    api.get<{ data: string }>('/qr-json', { params: { sender } }),

  getPairCode: (sender: string) =>
    api.get<{ pair_code: string }>('/pair-code', { params: { sender } }),

  autoLogin: () => api.get('/autologin'),

  autoDisconnect: () => api.get('/auto-disconnect'),

  getDeviceProxies: () => api.get<DeviceWithProxy[]>('/device-proxies'),
}

// Messaging
export const messageApi = {
  send: (sender: string, recipient: string, message: string) =>
    api.post<{ message: string; id_pesan: string }>(
      '/send',
      { recipient, message },
      { params: { sender } }
    ),

  sendBulk: (sender: string, data: BulkSendRequest) =>
    api.post<{ message: string; recipients: number; note: string }>(
      '/send-bulk',
      data,
      { params: { sender } }
    ),

  sendPresence: (sender: string) =>
    api.post('/presence', {}, { params: { sender } }),

  deleteMessages: (sender: string, recipient: string, messageIDs: string[]) =>
    api.delete('/message', {
      params: { sender },
      data: { recipient, message_ids: messageIDs },
    }),

  getBulkSendStatus: (sender: string) =>
    api.get<BulkSendStatus>('/bulk-send-status', { params: { sender } }),
}

// User Validation
export const userApi = {
  checkUser: (sender: string, recipients: string[]) =>
    api.post<CheckUserResponse[]>(
      '/check-user',
      { recipients },
      { params: { sender } }
    ),

  checkUserSingle: (sender: string, recipient: string) =>
    api.post<CheckUserResponse>(
      '/check-user-single',
      { recipient },
      { params: { sender } }
    ),
}

// File Upload
export const uploadApi = {
  upload: (sender: string, formData: FormData) =>
    api.post<Message[]>('/upload', formData, {
      params: { sender },
      headers: { 'Content-Type': 'multipart/form-data' },
    }),

  uploadSingle: (sender: string, file: string, recipient: string, caption?: string) =>
    api.post<Message>(
      '/upload-single',
      { file, recipient, caption },
      { params: { sender } }
    ),
}

// Health & Monitoring
export const healthApi = {
  check: () => api.get<{ message: string }>('/health-check'),

  getSessionHealth: (sender: string) =>
    api.get<HealthStats>(`/health-status`, { params: { sender } }),

  getValidationStats: () =>
    api.get<ValidationCacheStats>('/validation-stats'),
}

// Error handling interceptor
let isRefreshing = false
let failedQueue: Array<{
  resolve: (value?: unknown) => void
  reject: (reason?: unknown) => void
}> = []

const processQueue = (error: Error | null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error)
    } else {
      prom.resolve()
    }
  })
  failedQueue = []
}

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config

    if (error.response) {
      // Handle 401 Unauthorized - try to refresh token
      if (error.response.status === 401 && !originalRequest._retry) {
        console.log('[API] 401 Unauthorized on:', originalRequest.url)
        if (isRefreshing) {
          console.log('[API] Already refreshing, queuing request')
          return new Promise((resolve, reject) => {
            failedQueue.push({ resolve, reject })
          })
            .then(() => {
              return api(originalRequest)
            })
            .catch((err) => {
              return Promise.reject(err)
            })
        }

        originalRequest._retry = true
        isRefreshing = true

        const refreshToken = localStorage.getItem('auth_refresh_token')
        console.log('[API] Refresh token present:', !!refreshToken)
        
        if (!refreshToken) {
          console.log('[API] No refresh token, redirecting to login')
          localStorage.removeItem('auth_token')
          localStorage.removeItem('auth_username')
          localStorage.removeItem('auth_refresh_token')
          window.location.href = '/login'
          return Promise.reject(error)
        }

        try {
          console.log('[API] Attempting token refresh...')
          const response = await axios.post('/api/refresh-token', {
            refresh_token: refreshToken,
          })

          const { token, refresh_token } = response.data
          console.log('[API] Token refresh successful')
          localStorage.setItem('auth_token', token)
          localStorage.setItem('auth_refresh_token', refresh_token)
          
          originalRequest.headers.Authorization = `Bearer ${token}`
          processQueue(null)
          
          return api(originalRequest)
        } catch (refreshError) {
          console.error('[API] Token refresh failed:', refreshError)
          processQueue(refreshError as Error)
          localStorage.removeItem('auth_token')
          localStorage.removeItem('auth_username')
          localStorage.removeItem('auth_refresh_token')
          console.log('[API] Redirecting to login due to refresh failure')
          window.location.href = '/login'
          return Promise.reject(refreshError)
        } finally {
          isRefreshing = false
        }
      }

      // Server responded with error
      const message = error.response.data?.message || error.response.data?.error || 'An error occurred'
      return Promise.reject(new Error(message))
    } else if (error.request) {
      // Request made but no response
      return Promise.reject(new Error('No response from server. Please check your connection.'))
    } else {
      // Something else happened
      return Promise.reject(error)
    }
  }
)

export default api
