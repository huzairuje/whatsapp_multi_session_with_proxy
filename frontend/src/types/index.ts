// API Response Types
export interface ApiResponse<T> {
  message?: string
  data?: T
  error?: string
}

// Session/Device Types
export interface Device {
  pushName: string
  platform: string
  user: string
  server: string
  isLoggedIn: boolean
}

export interface DeviceWithProxy extends Device {
  proxyURL: string | null
}

export interface SessionStatus {
  id: string
  pushName: string
  isLogin: boolean
}

// Bulk Send Types
export interface BulkSendRequest {
  recipients: string[]
  message: string
  variables?: Record<string, string>
}

export interface BulkSendResult {
  recipient: string
  message_id?: string
  success: boolean
  error?: string
}

export interface BulkSendStatus {
  sender: string
  daily_count: number
  daily_limit: number
  remaining: number
}

// Message Types
export interface Message {
  messageID: string
  jid: string
  type: string
  body: string
  sent: boolean
  fileName?: string
}

// Recipient Types
export interface Recipient {
  phone: string
  name?: string
  variables?: Record<string, string>
}

export interface CheckUserResponse {
  Query: string
  IsIn: boolean
  JID: string
  VerifiedName?: string
}

// Template Types
export interface MessageTemplate {
  id: string
  name: string
  message: string
  variables: string[]
  createdAt: string
}

// Health Types
export interface HealthStats {
  total_sends: number
  failed_sends: number
  error_rate: number
  consecutive_errors: number
  is_healthy: boolean
  last_connect_time: string
  last_error_time: string
}

// Validation Cache Types
export interface ValidationCacheStats {
  total_entries: number
  valid_entries: number
  invalid_entries: number
  expired_entries: number
}

// Config Types
export interface BulkSendConfig {
  minDelay: number
  maxDelay: number
  batchSize: number
  batchPauseMin: number
  batchPauseMax: number
  dailyLimit: number
  typingDelayMin: number
  typingDelayMax: number
  enablePresenceSimulation: boolean
  allowedHourStart: number
  allowedHourEnd: number
  timezone: string
  enableTimeRestrictions: boolean
  errorBackoffMinutes: number
  enableRecipientValidation: boolean
  validationCacheDuration: number
  enableHealthCheck: boolean
  maxErrorRate: number
}

// Stats Types
export interface DashboardStats {
  totalSessions: number
  activeSessions: number
  messagesToday: number
  successRate: number
}

export interface SessionStats {
  user: string
  messagesSent: number
  messagesSuccess: number
  messagesFailed: number
  dailyLimit: number
  healthStatus: 'healthy' | 'warning' | 'error'
}

// UI State Types
export interface ToastMessage {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  message: string
  duration?: number
}

export interface ModalState {
  isOpen: boolean
  type: 'qr' | 'confirm' | 'info' | null
  data?: any
}

// Form Types
export interface ConnectDeviceForm {
  sender: string
}

export interface SendMessageForm {
  sender: string
  recipient: string
  message: string
}

export interface BulkSendForm {
  sender: string
  recipients: string
  message: string
  variables: Record<string, string>
}
