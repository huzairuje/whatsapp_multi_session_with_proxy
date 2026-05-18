import axios from 'axios'
import type {
  Device,
  DeviceWithProxy,
  SessionStatus,
  BulkSendRequest,
  BulkSendResult,
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
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response) {
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
